package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	teatable "github.com/evertras/bubble-table/table"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/term"
)

// QueryMode represents whether we're doing instant or range queries.
type QueryMode int

const (
	ModeInstant QueryMode = iota
	ModeRange
	ModeSeries
	ModeLabels
)

func (m QueryMode) String() string {
	switch m {
	case ModeInstant:
		return "/query"
	case ModeRange:
		return "/query_range"
	case ModeSeries:
		return "/series"
	case ModeLabels:
		return "/labels"
	default:
		return "Unknown"
	}
}

// TUIState represents the current state of the TUI.
type TUIState int

const (
	StateInput TUIState = iota
	StateLoading
	StateResults
	StateError
)

// FocusedPane tracks which pane has focus.
type FocusedPane int

const (
	PaneQuery FocusedPane = iota
	PaneResults
	PaneLegend
)

// tuiInstantResultMsg carries the result of an instant query.
type tuiInstantResultMsg struct {
	warnings v1.Warnings
	vector   model.Vector
	err      error
	duration time.Duration
}

// tuiRangeResultMsg carries the result of a range query.
type tuiRangeResultMsg struct {
	warnings v1.Warnings
	matrix   model.Matrix
	err      error
	duration time.Duration
}

// tuiSeriesResultMsg carries the result of a series query.
type tuiSeriesResultMsg struct {
	warnings v1.Warnings
	series   []model.LabelSet
	err      error
	duration time.Duration
}

// tuiLabelsResultMsg carries the result of a labels query.
type tuiLabelsResultMsg struct {
	warnings v1.Warnings
	labels   []string
	err      error
	duration time.Duration
}

// tuiLabelValuesResultMsg carries the result of a label values query.
type tuiLabelValuesResultMsg struct {
	labelName string
	warnings  v1.Warnings
	values    []string
	err       error
	duration  time.Duration
}

// TUIModel is the main Bubble Tea model for the interactive TUI.
type TUIModel struct {
	promClient prometheus.Client
	timeout    time.Duration

	// Input
	queryInput textinput.Model

	// Mode
	mode QueryMode

	// Per-mode state (indexed by QueryMode)
	modeQueries   [4]string        // Query string for each mode
	modeStates    [4]TUIState      // State for each mode
	modeWarnings  [4]v1.Warnings   // Warnings for each mode
	modeErrors    [4]error         // Errors for each mode
	modeDurations [4]time.Duration // Query execution duration for each mode

	// Range query parameters
	rangeValue time.Duration
	stepValue  time.Duration

	// Series query parameters
	seriesLimit uint64

	// Results (already per-mode by nature)
	vector model.Vector     // For instant queries
	matrix model.Matrix     // For range queries
	series []model.LabelSet // For series queries
	labels []string         // For labels queries

	// Label values state
	labelValues        []string // Values for selected label
	selectedLabelName  string   // Currently selected label name
	viewingLabelValues bool     // True when showing values instead of names
	selectedLabelIndex int      // Index of selected label row (for restoring position)

	// Rendered content
	chartContent  string
	legendEntries []charts.LegendEntry
	legendTable   teatable.Model
	seriesTable   teatable.Model
	labelsTable   teatable.Model
	selectedIndex int // -1 means no selection

	// UI state
	width                int
	height               int
	focusedPane          FocusedPane
	insertMode           bool // true when editing query (insert mode), false for normal mode
	spinner              spinner.Model
	legendFocused        bool
	showShortcutsOverlay bool
}

// NewTUIModel creates a new TUI model.
func NewTUIModel(client prometheus.Client, rangeValue, stepValue time.Duration, seriesLimit uint64, timeout time.Duration) TUIModel {
	ti := textinput.New()
	ti.Placeholder = "Enter PromQL query..."
	ti.Focus()
	ti.Width = 60

	return TUIModel{
		promClient:    client,
		timeout:       timeout,
		queryInput:    ti,
		mode:          ModeInstant,
		modeStates:    [4]TUIState{StateInput, StateInput, StateInput, StateInput},
		rangeValue:    rangeValue,
		stepValue:     stepValue,
		seriesLimit:   seriesLimit,
		selectedIndex: -1,
		focusedPane:   PaneQuery,
		insertMode:    true, // Start in insert mode so users can immediately type
		spinner:       NewLoadingSpinner(),
	}
}

// Helper methods for accessing current mode's state
func (m TUIModel) currentState() TUIState {
	return m.modeStates[m.mode]
}

func (m TUIModel) currentWarnings() v1.Warnings {
	return m.modeWarnings[m.mode]
}

func (m TUIModel) currentError() error {
	return m.modeErrors[m.mode]
}

func (m TUIModel) currentDuration() time.Duration {
	return m.modeDurations[m.mode]
}

// formatDuration formats a duration with appropriate precision.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.queryInput.Width = msg.Width - 10
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tuiInstantResultMsg:
		return m.handleInstantResult(msg)

	case tuiRangeResultMsg:
		return m.handleRangeResult(msg)

	case tuiSeriesResultMsg:
		return m.handleSeriesResult(msg)

	case tuiLabelsResultMsg:
		return m.handleLabelsResult(msg)

	case tuiLabelValuesResultMsg:
		return m.handleLabelValuesResult(msg)

	case spinner.TickMsg:
		if m.currentState() == StateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Update text input if focused
	if m.focusedPane == PaneQuery && m.currentState() != StateLoading {
		var cmd tea.Cmd
		m.queryInput, cmd = m.queryInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m TUIModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle shortcuts overlay - dismiss on any key except quit keys
	if m.showShortcutsOverlay {
		if msg.String() == "q" {
			return m, tea.Quit
		}
		m.showShortcutsOverlay = false
		return m, nil
	}

	// State-dependent keys
	switch m.currentState() {
	case StateLoading:
		// Only allow quit during loading
		return m, nil

	case StateInput, StateResults, StateError:
		return m.handleInputOrResultsKey(msg)
	}

	return m, nil
}

func (m TUIModel) handleInputOrResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle legend navigation when focused
	if m.focusedPane == PaneLegend && m.legendFocused {
		return m.handleLegendKey(msg)
	}

	// INSERT MODE: Route most keys to text input
	if m.insertMode {
		switch msg.String() {
		case "esc":
			// Exit insert mode
			m.insertMode = false
			m.queryInput.Blur()
			return m, nil
		case "enter":
			// Execute and exit insert mode
			m.insertMode = false
			m.queryInput.Blur()
			return m.handleEnterKey()
		case "tab":
			// Allow mode switching even in insert mode
			return m.handleTabKey()
		default:
			// All other keys go to text input
			var cmd tea.Cmd
			m.queryInput, cmd = m.queryInput.Update(msg)
			return m, cmd
		}
	}

	// NORMAL MODE: Handle shortcuts
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "tab":
		return m.handleTabKey()
	case "enter":
		return m.handleEnterKey()
	case "/":
		return m.enterInsertMode()
	case "i":
		return m.handleInteractiveKey()
	case "f":
		return m.handleFormatKey()
	case "esc":
		return m.handleEscapeKey()
	case "1", "2", "3", "4":
		return m.handleNumberKey(msg.String())
	case "?":
		m.showShortcutsOverlay = true
		return m, nil
	}

	return m, nil
}

func (m TUIModel) handleTabKey() (tea.Model, tea.Cmd) {
	// Cycle through modes: Instant -> Range -> Series -> Labels -> Instant
	nextMode := m.mode + 1
	if nextMode > ModeLabels {
		nextMode = ModeInstant
	}
	return m.switchToMode(nextMode)
}

func (m TUIModel) handleNumberKey(key string) (tea.Model, tea.Cmd) {
	modeMap := map[string]QueryMode{
		"1": ModeInstant,
		"2": ModeRange,
		"3": ModeSeries,
		"4": ModeLabels,
	}
	if mode, ok := modeMap[key]; ok {
		return m.switchToMode(mode)
	}
	return m, nil
}

func (m TUIModel) switchToMode(newMode QueryMode) (tea.Model, tea.Cmd) {
	if newMode == m.mode {
		return m, nil
	}

	// Save current query before switching
	m.modeQueries[m.mode] = m.queryInput.Value()

	// Switch to new mode
	m.mode = newMode

	// Restore new mode's query
	m.queryInput.SetValue(m.modeQueries[m.mode])

	// Clear any selection
	m.selectedIndex = -1
	m.legendFocused = false

	// Delegate to mode's OnSwitchTo for re-rendering if needed
	m.currentMode().OnSwitchTo(&m)

	return m, nil
}

func (m TUIModel) handleEnterKey() (tea.Model, tea.Cmd) {
	// Execute query
	if m.queryInput.Value() != "" {
		return m.executeQuery()
	}
	return m, nil
}

func (m TUIModel) enterInsertMode() (tea.Model, tea.Cmd) {
	m.insertMode = true
	m.focusedPane = PaneQuery
	m.queryInput.Focus()
	return m, nil
}

func (m TUIModel) handleInteractiveKey() (tea.Model, tea.Cmd) {
	// Toggle interactive mode - delegate to current mode
	if m.currentState() != StateResults {
		return m, nil
	}

	cmd := m.currentMode().HandleInteractiveToggle(&m)
	return m, cmd
}

func (m TUIModel) handleEscapeKey() (tea.Model, tea.Cmd) {
	// Exit interactive/legend mode but stay in normal mode
	m.focusedPane = PaneQuery
	m.legendFocused = false
	// Don't focus query - stay in normal mode
	if m.mode == ModeRange && len(m.legendEntries) > 0 {
		m.legendTable = m.legendTable.Focused(false)
		m.selectedIndex = -1
		m = m.regenerateRangeChart()
	}
	return m, nil
}

func (m TUIModel) handleFormatKey() (tea.Model, tea.Cmd) {
	currentQuery := m.queryInput.Value()
	if currentQuery == "" {
		return m, nil // Nothing to format
	}
	formatted := prometheus.FormatQuery(currentQuery)
	m.queryInput.SetValue(formatted)
	return m, nil
}

func (m TUIModel) handleLegendKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cmd := m.currentMode().HandleLegendKey(&m, msg)
	return m, cmd
}

func (m TUIModel) executeQuery() (tea.Model, tea.Cmd) {
	// Save query for this mode
	m.modeQueries[m.mode] = m.queryInput.Value()
	m.modeStates[m.mode] = StateLoading
	m.modeErrors[m.mode] = nil
	m.modeWarnings[m.mode] = nil
	m.queryInput.Blur()

	cmd := m.currentMode().ExecuteQuery(&m)
	return m, cmd
}

func (m TUIModel) executeInstantQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		start := time.Now()
		warnings, vector, err := m.promClient.Query(query, m.timeout)
		duration := time.Since(start)
		return tuiInstantResultMsg{
			warnings: warnings,
			vector:   vector,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeRangeQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		matrix, warnings, err := m.promClient.QueryRange(query, rangeStart, end, m.stepValue, m.timeout)
		duration := time.Since(start)
		return tuiRangeResultMsg{
			warnings: warnings,
			matrix:   matrix,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeSeriesQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		series, warnings, err := m.promClient.Series(query, rangeStart, end, m.seriesLimit, m.timeout)
		duration := time.Since(start)
		return tuiSeriesResultMsg{
			warnings: warnings,
			series:   series,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeLabelsQuery() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		labels, warnings, err := m.promClient.LabelNames(rangeStart, end, m.timeout)
		duration := time.Since(start)
		return tuiLabelsResultMsg{
			warnings: warnings,
			labels:   labels,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeLabelValuesQuery(labelName string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		values, warnings, err := m.promClient.LabelValues(labelName, rangeStart, end, m.timeout)
		duration := time.Since(start)
		return tuiLabelValuesResultMsg{
			labelName: labelName,
			warnings:  warnings,
			values:    values,
			err:       err,
			duration:  duration,
		}
	}
}

func (m TUIModel) handleInstantResult(msg tuiInstantResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeInstant] = msg.warnings
	m.vector = msg.vector
	m.modeErrors[ModeInstant] = msg.err
	m.modeDurations[ModeInstant] = msg.duration
	m.queryInput.Blur()
	m.insertMode = false
	m.focusedPane = PaneQuery

	if msg.err != nil {
		m.modeStates[ModeInstant] = StateError
		return m, nil
	}

	m.modeStates[ModeInstant] = StateResults
	m = m.renderInstantChart()
	return m, nil
}

func (m TUIModel) handleRangeResult(msg tuiRangeResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeRange] = msg.warnings
	m.matrix = msg.matrix
	m.modeErrors[ModeRange] = msg.err
	m.modeDurations[ModeRange] = msg.duration
	m.queryInput.Blur()
	m.insertMode = false
	m.focusedPane = PaneQuery

	if msg.err != nil {
		m.modeStates[ModeRange] = StateError
		return m, nil
	}

	m.modeStates[ModeRange] = StateResults
	m.selectedIndex = -1
	m = m.renderRangeChart()
	return m, nil
}

func (m TUIModel) handleSeriesResult(msg tuiSeriesResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeSeries] = msg.warnings
	m.series = msg.series
	m.modeErrors[ModeSeries] = msg.err
	m.modeDurations[ModeSeries] = msg.duration
	m.queryInput.Blur()
	m.insertMode = false
	m.focusedPane = PaneQuery

	if msg.err != nil {
		m.modeStates[ModeSeries] = StateError
		return m, nil
	}

	m.modeStates[ModeSeries] = StateResults
	m = m.renderSeriesTable()
	return m, nil
}

func (m TUIModel) handleLabelsResult(msg tuiLabelsResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeLabels] = msg.warnings
	m.labels = msg.labels
	m.modeErrors[ModeLabels] = msg.err
	m.modeDurations[ModeLabels] = msg.duration
	m.queryInput.Blur()
	m.insertMode = false
	m.focusedPane = PaneQuery
	m.viewingLabelValues = false

	if msg.err != nil {
		m.modeStates[ModeLabels] = StateError
		return m, nil
	}

	m.modeStates[ModeLabels] = StateResults
	m = m.renderLabelsTable()
	return m, nil
}

func (m TUIModel) handleLabelValuesResult(msg tuiLabelValuesResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeLabels] = msg.warnings
	m.labelValues = msg.values
	m.selectedLabelName = msg.labelName
	m.modeErrors[ModeLabels] = msg.err
	m.modeDurations[ModeLabels] = msg.duration

	if msg.err != nil {
		m.modeStates[ModeLabels] = StateError
		m.viewingLabelValues = false
		return m, nil
	}

	m.modeStates[ModeLabels] = StateResults
	m.viewingLabelValues = true
	m = m.renderLabelsTable()
	return m, nil
}

func (m TUIModel) renderInstantChart() TUIModel {
	width := m.getChartWidth()
	m.chartContent = charts.Barchart(m.vector, width)
	return m
}

func (m TUIModel) renderRangeChart() TUIModel {
	width := m.getChartWidth()
	m.chartContent, m.legendEntries = charts.TimeseriesSplitWithSelection(m.matrix, width, m.selectedIndex)
	m.createLegendTable(5)
	return m
}

func (m TUIModel) renderSeriesTable() TUIModel {
	if len(m.series) == 0 {
		return m
	}

	// Collect all unique label names across all series
	labelNames := make(map[string]bool)
	for _, s := range m.series {
		for name := range s {
			labelNames[string(name)] = true
		}
	}

	// Sort label names for consistent column ordering
	sortedNames := make([]string, 0, len(labelNames))
	for name := range labelNames {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// Calculate max width for each column (header + all values)
	colWidths := make(map[string]int)
	for _, name := range sortedNames {
		colWidths[name] = len(name) // Start with header length
	}
	for _, s := range m.series {
		for _, name := range sortedNames {
			val := string(s[model.LabelName(name)])
			if len(val) > colWidths[name] {
				colWidths[name] = len(val)
			}
		}
	}

	// Create columns with responsive widths
	columns := make([]teatable.Column, 0, len(sortedNames))
	for _, name := range sortedNames {
		colWidth := colWidths[name]
		if colWidth < 10 {
			colWidth = 10
		}
		if colWidth > 40 {
			colWidth = 40
		}
		columns = append(columns, teatable.NewColumn(name, name, colWidth))
	}

	// Create rows for each series
	rows := make([]teatable.Row, 0, len(m.series))
	for _, s := range m.series {
		rowData := make(teatable.RowData)
		for _, name := range sortedNames {
			val := s[model.LabelName(name)]
			rowData[name] = string(val)
		}
		rows = append(rows, teatable.NewRow(rowData))
	}

	m.seriesTable = teatable.
		New(columns).
		WithRows(rows).
		WithPageSize(10).
		Focused(false).
		WithBaseStyle(lipgloss.NewStyle())

	return m
}

func (m TUIModel) renderLabelsTable() TUIModel {
	if m.viewingLabelValues {
		// Show label values
		if len(m.labelValues) == 0 {
			return m
		}

		// Find the longest value for column width
		headerText := fmt.Sprintf("Values for '%s'", m.selectedLabelName)
		maxWidth := len(headerText)
		for _, value := range m.labelValues {
			if len(value) > maxWidth {
				maxWidth = len(value)
			}
		}
		if maxWidth > 60 {
			maxWidth = 60
		}

		columns := []teatable.Column{
			teatable.NewColumn("value", headerText, maxWidth),
		}

		rows := make([]teatable.Row, 0, len(m.labelValues))
		for _, value := range m.labelValues {
			rows = append(rows, teatable.NewRow(teatable.RowData{
				"value": value,
			}))
		}

		m.labelsTable = teatable.
			New(columns).
			WithRows(rows).
			WithPageSize(15).
			Focused(m.legendFocused).
			WithBaseStyle(lipgloss.NewStyle())

		return m
	}

	// Show label names
	if len(m.labels) == 0 {
		return m
	}

	// Find the longest label name for column width
	maxWidth := len("Label Name")
	for _, label := range m.labels {
		if len(label) > maxWidth {
			maxWidth = len(label)
		}
	}
	if maxWidth > 60 {
		maxWidth = 60
	}

	columns := []teatable.Column{
		teatable.NewColumn("label", "Label Name", maxWidth),
	}

	rows := make([]teatable.Row, 0, len(m.labels))
	for _, label := range m.labels {
		rows = append(rows, teatable.NewRow(teatable.RowData{
			"label": label,
		}))
	}

	m.labelsTable = teatable.
		New(columns).
		WithRows(rows).
		WithPageSize(15).
		Focused(m.legendFocused).
		WithBaseStyle(lipgloss.NewStyle()).
		WithHighlightedRow(m.selectedLabelIndex)

	return m
}

func (m TUIModel) regenerateRangeChart() TUIModel {
	width := m.getChartWidth()
	m.chartContent, _ = charts.TimeseriesSplitWithSelection(m.matrix, width, m.selectedIndex)
	return m
}

func (m TUIModel) getChartWidth() int {
	width := m.width - 6
	if width <= 0 {
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && termWidth > 0 {
			width = termWidth - 6
		} else {
			width = DefaultTerminalWidth - 6
		}
	}
	return width
}

func (m TUIModel) getTerminalWidth() int {
	if m.width > 0 {
		return m.width
	}
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && termWidth > 0 {
		return termWidth
	}
	return DefaultTerminalWidth
}

func (m *TUIModel) createLegendTable(maxRows int) {
	if len(m.legendEntries) == 0 {
		return
	}

	rows := make([]teatable.Row, 0, len(m.legendEntries))
	longestMetric := 0

	for _, entry := range m.legendEntries {
		if len(entry.Metric) > longestMetric {
			longestMetric = len(entry.Metric)
		}

		style := charts.SeriesStyle(entry.ColorIndex)
		colorIndicator := style.Render("\u2588")

		rows = append(rows, teatable.NewRow(teatable.RowData{
			"color":  colorIndicator,
			"metric": entry.Metric,
		}))
	}

	columns := []teatable.Column{
		teatable.NewColumn("color", "", 3),
		teatable.NewColumn("metric", "Metric", max(longestMetric, 20)),
	}

	m.legendTable = teatable.
		New(columns).
		WithRows(rows).
		WithPageSize(maxRows).
		Focused(m.legendFocused)
}

func (m TUIModel) updateSelectedFromLegendTable() TUIModel {
	highlightedRow := m.legendTable.HighlightedRow()
	if highlightedRow.Data == nil {
		m.selectedIndex = -1
		return m
	}

	metricName, ok := highlightedRow.Data["metric"].(string)
	if !ok {
		m.selectedIndex = -1
		return m
	}

	for i, entry := range m.legendEntries {
		if entry.Metric == metricName {
			m.selectedIndex = i
			return m
		}
	}

	m.selectedIndex = -1
	return m
}

func (m TUIModel) View() string {
	// Show shortcuts overlay if active
	if m.showShortcutsOverlay {
		overlay := m.renderShortcutsOverlay()
		return lipgloss.Place(
			m.getTerminalWidth(),
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			overlay,
		)
	}

	var s strings.Builder

	// Status bar
	s.WriteString(m.renderStatusBar())
	s.WriteString("\n")

	// Query input
	s.WriteString(m.renderQueryInput())
	s.WriteString("\n")

	// Results area
	s.WriteString(m.renderResults())

	// Results status bar (latency, etc.)
	s.WriteString(m.renderResultsStatusBar())
	s.WriteString("\n")

	// Help bar
	s.WriteString(m.renderHelpBar())

	return s.String()
}

func (m TUIModel) renderStatusBar() string {
	// Mode indicator
	modeStyle := lipgloss.NewStyle().Bold(true)
	instantStyle := modeStyle
	rangeStyle := modeStyle
	seriesStyle := modeStyle
	labelsStyle := modeStyle

	activeStyle := modeStyle.Background(lipgloss.Color("63")).Foreground(lipgloss.Color("231"))

	switch m.mode {
	case ModeInstant:
		instantStyle = activeStyle
	case ModeRange:
		rangeStyle = activeStyle
	case ModeSeries:
		seriesStyle = activeStyle
	case ModeLabels:
		labelsStyle = activeStyle
	}

	modeText := fmt.Sprintf("  Mode: %s | %s | %s | %s",
		instantStyle.Render(" 1 /query "),
		rangeStyle.Render(" 2 /query_range "),
		seriesStyle.Render(" 3 /series "),
		labelsStyle.Render(" 4 /labels "))

	// Get mode-specific parameters
	paramsText := m.currentMode().RenderStatusParams(&m)

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.getTerminalWidth()).
		Padding(0, 1)

	return statusStyle.Render(modeText + paramsText)
}

func (m TUIModel) renderQueryInput() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	// Insert mode: highlighted border, Normal mode: dimmed border
	if m.insertMode {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("205"))
	} else {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("63"))
	}

	return inputStyle.Render(m.queryInput.View())
}

func (m TUIModel) renderResults() string {
	var content string

	switch m.currentState() {
	case StateInput:
		content = m.renderEmptyState()
	case StateLoading:
		content = m.renderLoadingState()
	case StateError:
		content = m.renderErrorState()
	case StateResults:
		content = m.renderResultsContent()
	}

	return content
}

func (TUIModel) renderEmptyState() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1)
	return emptyStyle.Render(" ")
}

func (m TUIModel) renderLoadingState() string {
	loadingStyle := lipgloss.NewStyle().Padding(2, 4)
	return loadingStyle.Render(fmt.Sprintf("%s Executing query: %s", m.spinner.View(), m.queryInput.Value()))
}

func (m TUIModel) renderErrorState() string {
	errorStyle := lipgloss.NewStyle().Padding(1, 2)
	return errorStyle.Render(ErrorStyle.Render("Error: ") + m.currentError().Error())
}

func (m TUIModel) renderResultsContent() string {
	return m.currentMode().RenderResultsContent(&m)
}

func (m TUIModel) renderResultsStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		MarginTop(1)

	duration := m.currentDuration()
	content := ""
	if duration != 0 {
		content = " Latency: " + formatDuration(duration)
	}

	// Add mode-specific status bar content
	content += m.currentMode().RenderResultsStatusBar(&m)

	return statusStyle.Render(content)
}

func (m TUIModel) renderHelpBar() string {
	helpStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.getTerminalWidth()).
		Padding(0, 1)

	var helpText string
	if m.insertMode {
		helpText = "esc: normal | ?: shortcuts | ctrl+c/q: quit"
	} else {
		helpText = "/: edit | ?: shortcuts | q: quit"
	}

	return helpStyle.Render(helpText)
}

func (TUIModel) renderShortcutsOverlay() string {
	accentColor := lipgloss.Color("205")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor).
		MarginBottom(1)

	categoryStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	var content strings.Builder

	content.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	content.WriteString("\n")

	// Global shortcuts
	content.WriteString(categoryStyle.Render("Global"))
	content.WriteString("\n")
	shortcuts := []struct{ key, desc string }{
		{"Tab", "Cycle through modes"},
		{"1-4", "Switch to mode directly"},
		{"Enter", "Execute query"},
		{"q", "Quit"},
		{"Ctrl+C", "Force quit"},
	}
	for _, s := range shortcuts {
		content.WriteString(fmt.Sprintf("  %s  %s\n", keyStyle.Render(fmt.Sprintf("%-8s", s.key)), descStyle.Render(s.desc)))
	}

	// Query editing
	content.WriteString(categoryStyle.Render("Query Editing"))
	content.WriteString("\n")
	editShortcuts := []struct{ key, desc string }{
		{"/", "Enter insert mode"},
		{"Esc", "Exit insert mode"},
		{"f", "Format PromQL query"},
	}
	for _, s := range editShortcuts {
		content.WriteString(fmt.Sprintf("  %s  %s\n", keyStyle.Render(fmt.Sprintf("%-8s", s.key)), descStyle.Render(s.desc)))
	}

	// Interactive mode
	content.WriteString(categoryStyle.Render("Interactive Mode"))
	content.WriteString("\n")
	interactiveShortcuts := []struct{ key, desc string }{
		{"i", "Toggle interactive mode"},
		{"j/k", "Navigate up/down"},
		{"h/l", "Page up/down"},
		{"Esc", "Exit interactive mode"},
	}
	for _, s := range interactiveShortcuts {
		content.WriteString(fmt.Sprintf("  %s  %s\n", keyStyle.Render(fmt.Sprintf("%-8s", s.key)), descStyle.Render(s.desc)))
	}

	content.WriteString("\n")
	content.WriteString(descStyle.Render("Press any key to close"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2)

	return boxStyle.Render(content.String())
}

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

	// Rendered content
	chartContent  string
	legendEntries []charts.LegendEntry
	legendTable   teatable.Model
	seriesTable   teatable.Model
	labelsTable   teatable.Model
	selectedIndex int // -1 means no selection

	// UI state
	width         int
	height        int
	focusedPane   FocusedPane
	insertMode    bool // true when editing query (insert mode), false for normal mode
	spinner       spinner.Model
	legendFocused bool
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
	}

	return m, nil
}

func (m TUIModel) handleTabKey() (tea.Model, tea.Cmd) {
	// Save current query before switching
	m.modeQueries[m.mode] = m.queryInput.Value()

	// Cycle through modes: Instant -> Range -> Series -> Labels -> Instant
	switch m.mode {
	case ModeInstant:
		m.mode = ModeRange
	case ModeRange:
		m.mode = ModeSeries
	case ModeSeries:
		m.mode = ModeLabels
	case ModeLabels:
		m.mode = ModeInstant
	}

	// Restore new mode's query
	m.queryInput.SetValue(m.modeQueries[m.mode])

	// Clear any selection
	m.selectedIndex = -1
	m.legendFocused = false

	// Re-render chart content for the new mode if it has results
	if m.currentState() == StateResults {
		switch m.mode {
		case ModeInstant:
			m = m.renderInstantChart()
		case ModeRange:
			m = m.renderRangeChart()
		case ModeSeries:
			m = m.renderSeriesTable()
		case ModeLabels:
			m = m.renderLabelsTable()
		}
	}

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
	// Toggle interactive mode (range mode with legend entries, or series mode with series)
	if m.currentState() != StateResults {
		return m, nil
	}

	if m.mode == ModeRange && len(m.legendEntries) > 0 {
		m.legendFocused = !m.legendFocused
		if m.legendFocused {
			m.focusedPane = PaneLegend
			m.queryInput.Blur()
			m.legendTable = m.legendTable.Focused(true)
			// Select the first series and redraw chart immediately
			m = m.updateSelectedFromLegendTable()
			m = m.regenerateRangeChart()
		} else {
			m.focusedPane = PaneQuery
			// Stay in normal mode - don't focus query input
			m.legendTable = m.legendTable.Focused(false)
			m.selectedIndex = -1
			m = m.regenerateRangeChart()
		}
		return m, nil
	}

	if m.mode == ModeSeries && len(m.series) > 0 {
		m.legendFocused = !m.legendFocused
		if m.legendFocused {
			m.focusedPane = PaneLegend
			m.queryInput.Blur()
			m.seriesTable = m.seriesTable.Focused(true)
		} else {
			m.focusedPane = PaneQuery
			// Stay in normal mode - don't focus query input
			m.seriesTable = m.seriesTable.Focused(false)
		}
		return m, nil
	}

	if m.mode == ModeLabels && len(m.labels) > 0 {
		m.legendFocused = !m.legendFocused
		if m.legendFocused {
			m.focusedPane = PaneLegend
			m.queryInput.Blur()
			m.labelsTable = m.labelsTable.Focused(true)
		} else {
			m.focusedPane = PaneQuery
			// Stay in normal mode - don't focus query input
			m.labelsTable = m.labelsTable.Focused(false)
		}
		return m, nil
	}

	return m, nil
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
	key := msg.String()

	switch key {
	case "q":
		return m, tea.Quit
	case "i", "esc":
		// In labels mode, if viewing values, go back to labels list
		if m.mode == ModeLabels && m.viewingLabelValues {
			m.viewingLabelValues = false
			m.labelValues = nil
			m.selectedLabelName = ""
			m = m.renderLabelsTable()
			return m, nil
		}
		// Exit interactive mode but stay in normal mode
		m.legendFocused = false
		m.focusedPane = PaneQuery
		// Don't focus query - stay in normal mode
		switch m.mode {
		case ModeRange:
			m.legendTable = m.legendTable.Focused(false)
			m.selectedIndex = -1
			m = m.regenerateRangeChart()
		case ModeSeries:
			m.seriesTable = m.seriesTable.Focused(false)
		case ModeLabels:
			m.labelsTable = m.labelsTable.Focused(false)
		case ModeInstant:
			// No table to unfocus in instant mode
		}
		return m, nil
	}

	// Handle series mode table navigation
	if m.mode == ModeSeries {
		var tableCmd tea.Cmd
		switch key {
		case "j":
			m.seriesTable, tableCmd = m.seriesTable.Update(tea.KeyMsg{Type: tea.KeyDown})
		case "k":
			m.seriesTable, tableCmd = m.seriesTable.Update(tea.KeyMsg{Type: tea.KeyUp})
		case "h":
			m.seriesTable, tableCmd = m.seriesTable.Update(tea.KeyMsg{Type: tea.KeyPgUp})
		case "l":
			m.seriesTable, tableCmd = m.seriesTable.Update(tea.KeyMsg{Type: tea.KeyPgDown})
		default:
			m.seriesTable, tableCmd = m.seriesTable.Update(msg)
		}
		return m, tableCmd
	}

	// Handle labels mode table navigation
	if m.mode == ModeLabels {
		var tableCmd tea.Cmd
		switch key {
		case "enter":
			// Query values for the selected label
			if !m.viewingLabelValues {
				highlightedRow := m.labelsTable.HighlightedRow()
				if highlightedRow.Data != nil {
					if labelName, ok := highlightedRow.Data["label"].(string); ok {
						m.selectedLabelName = labelName
						m.modeStates[ModeLabels] = StateLoading
						return m, m.executeLabelValuesQuery(labelName)
					}
				}
			}
			return m, nil
		case "j":
			m.labelsTable, tableCmd = m.labelsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
		case "k":
			m.labelsTable, tableCmd = m.labelsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
		case "h":
			m.labelsTable, tableCmd = m.labelsTable.Update(tea.KeyMsg{Type: tea.KeyPgUp})
		case "l":
			m.labelsTable, tableCmd = m.labelsTable.Update(tea.KeyMsg{Type: tea.KeyPgDown})
		default:
			m.labelsTable, tableCmd = m.labelsTable.Update(msg)
		}
		return m, tableCmd
	}

	// Handle range mode legend navigation
	oldSelected := m.selectedIndex

	var tableCmd tea.Cmd
	switch key {
	case "j":
		m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyDown})
	case "k":
		m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyUp})
	case "h":
		m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	case "l":
		m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	default:
		m.legendTable, tableCmd = m.legendTable.Update(msg)
	}

	m = m.updateSelectedFromLegendTable()

	if oldSelected != m.selectedIndex {
		m = m.regenerateRangeChart()
	}

	return m, tableCmd
}

func (m TUIModel) executeQuery() (tea.Model, tea.Cmd) {
	// Save query for this mode
	m.modeQueries[m.mode] = m.queryInput.Value()
	m.modeStates[m.mode] = StateLoading
	m.modeErrors[m.mode] = nil
	m.modeWarnings[m.mode] = nil
	m.queryInput.Blur()

	switch m.mode {
	case ModeInstant:
		return m, m.executeInstantQuery()
	case ModeRange:
		return m, m.executeRangeQuery()
	case ModeSeries:
		return m, m.executeSeriesQuery()
	case ModeLabels:
		return m, m.executeLabelsQuery()
	}
	return m, nil
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
		WithBaseStyle(lipgloss.NewStyle())

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
		instantStyle.Render(" /query "),
		rangeStyle.Render(" /query_range "),
		seriesStyle.Render(" /series "),
		labelsStyle.Render(" /labels "))

	// Parameters (show for range, series, and labels modes)
	paramsText := ""
	switch m.mode {
	case ModeInstant:
		// No additional params for instant mode
	case ModeRange:
		paramsText = fmt.Sprintf("   Range: %s   Step: %s", m.rangeValue, m.stepValue)
	case ModeSeries:
		paramsText = fmt.Sprintf("   Range: %s   Limit: %d", m.rangeValue, m.seriesLimit)
	case ModeLabels:
		paramsText = fmt.Sprintf("   Range: %s", m.rangeValue)
	}

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.width).
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
		Padding(2, 4)
	return emptyStyle.Render("Enter a PromQL query and press Enter to execute")
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
	var s strings.Builder

	// Warnings
	warnings := m.currentWarnings()
	if len(warnings) > 0 {
		s.WriteString("\n")
		s.WriteString(WarningStyle.Render("Warnings:\n"))
		for _, w := range warnings {
			s.WriteString(WarningStyle.Render(fmt.Sprintf("  - %s\n", w)))
		}
	}

	// Series mode: render table directly
	if m.mode == ModeSeries {
		tableStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

		if m.legendFocused {
			tableStyle = tableStyle.BorderForeground(lipgloss.Color("205"))
		}

		s.WriteString(tableStyle.Render(m.seriesTable.View()))
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("  %d series found\n", len(m.series)))
		return s.String()
	}

	// Labels mode: render table directly
	if m.mode == ModeLabels {
		tableStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

		if m.legendFocused {
			tableStyle = tableStyle.BorderForeground(lipgloss.Color("205"))
		}

		s.WriteString(tableStyle.Render(m.labelsTable.View()))
		s.WriteString("\n")
		if m.viewingLabelValues {
			s.WriteString(fmt.Sprintf("  %d values for label '%s'\n", len(m.labelValues), m.selectedLabelName))
		} else {
			s.WriteString(fmt.Sprintf("  %d labels found\n", len(m.labels)))
		}
		return s.String()
	}

	// Chart (instant and range modes)
	chartStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	if m.focusedPane == PaneResults {
		chartStyle = chartStyle.BorderForeground(lipgloss.Color("205"))
	}

	s.WriteString(chartStyle.Render(m.chartContent))

	// Legend (range mode only)
	if m.mode == ModeRange && len(m.legendEntries) > 0 {
		legendStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			MarginTop(1)

		if m.legendFocused {
			legendStyle = legendStyle.BorderForeground(lipgloss.Color("205"))
		}

		s.WriteString("\n")
		s.WriteString(legendStyle.Render(m.legendTable.View()))
	}

	s.WriteString("\n")
	return s.String()
}

func (m TUIModel) renderResultsStatusBar() string {
	if m.currentState() != StateResults && m.currentState() != StateError {
		return ""
	}
	duration := m.currentDuration()
	if duration == 0 {
		return ""
	}

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.width).
		Padding(0, 1)

	return statusStyle.Render("  Latency: "+formatDuration(duration)) + "\n"
}

func (m TUIModel) renderHelpBar() string {
	helpStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.width).
		Padding(0, 1)

	var helpText string
	switch {
	case m.legendFocused:
		if m.mode == ModeLabels {
			if m.viewingLabelValues {
				helpText = "  j/k: navigate | h/l: page | esc: back | i: exit | q: quit"
			} else {
				helpText = "  j/k: navigate | h/l: page | Enter: values | i/esc: exit | q: quit"
			}
		} else {
			helpText = "  j/k: navigate | h/l: page | i/esc: exit | q: quit"
		}
	case m.insertMode:
		helpText = "  Editing query | Enter: run | Esc: exit | Tab: mode"
	default:
		// Normal mode
		helpText = "  Tab: mode | Enter: run | /: edit query | f: format"
		if m.currentState() == StateResults {
			if m.mode == ModeRange && len(m.legendEntries) > 0 {
				helpText += " | i: legend"
			} else if m.mode == ModeSeries && len(m.series) > 0 {
				helpText += " | i: table"
			} else if m.mode == ModeLabels && len(m.labels) > 0 {
				helpText += " | i: table"
			}
		}
		helpText += " | q: quit"
	}

	return helpStyle.Render(helpText)
}

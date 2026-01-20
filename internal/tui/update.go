package tui

import (
	"sort"
	"strconv"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	teatable "github.com/evertras/bubble-table/table"
	"github.com/prometheus/common/model"
)

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

	switch msg.String() {
	case "q":
		return m.handleQuitKey(msg)
	case "tab":
		return m.handleTabKey()
	case "enter":
		return m.handleEnterKey()
	case "/":
		return m.handleSlashKey()
	case "i":
		return m.handleInteractiveKey()
	case "esc":
		return m.handleEscapeKey()
	case "f":
		return m.handleFormatKey(msg)
	}

	// If focused on query, update text input
	if m.focusedPane == PaneQuery {
		var cmd tea.Cmd
		m.queryInput, cmd = m.queryInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m TUIModel) handleQuitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only quit if not typing in query input
	if m.focusedPane != PaneQuery {
		return m, tea.Quit
	}
	// Otherwise, let it type 'q' in the input
	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	return m, cmd
}

func (m TUIModel) handleTabKey() (tea.Model, tea.Cmd) {
	// Save current query before switching
	m.modeQueries[m.mode] = m.queryInput.Value()

	// Cycle through modes: Instant -> Range -> Series -> Instant
	switch m.mode {
	case ModeInstant:
		m.mode = ModeRange
	case ModeRange:
		m.mode = ModeSeries
	case ModeSeries:
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

func (m TUIModel) handleSlashKey() (tea.Model, tea.Cmd) {
	// Focus query input
	m.focusedPane = PaneQuery
	m.legendFocused = false
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
		} else {
			m.focusedPane = PaneQuery
			m.queryInput.Focus()
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
			m.queryInput.Focus()
			m.seriesTable = m.seriesTable.Focused(false)
		}
		return m, nil
	}

	return m, nil
}

func (m TUIModel) handleEscapeKey() (tea.Model, tea.Cmd) {
	// Return focus to query input
	m.focusedPane = PaneQuery
	m.legendFocused = false
	m.queryInput.Focus()
	if m.mode == ModeRange && len(m.legendEntries) > 0 {
		m.legendTable = m.legendTable.Focused(false)
		m.selectedIndex = -1
		m = m.regenerateRangeChart()
	}
	return m, nil
}

func (m TUIModel) handleFormatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only format if not typing in query input, or if query input is empty
	if m.focusedPane != PaneQuery {
		return m, nil
	}
	currentQuery := m.queryInput.Value()
	if currentQuery == "" {
		// Let it type 'f' in the input
		var cmd tea.Cmd
		m.queryInput, cmd = m.queryInput.Update(msg)
		return m, cmd
	}
	// Format the query
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
		// Exit interactive mode
		m.legendFocused = false
		m.focusedPane = PaneQuery
		m.queryInput.Focus()
		switch m.mode {
		case ModeRange:
			m.legendTable = m.legendTable.Focused(false)
			m.selectedIndex = -1
			m = m.regenerateRangeChart()
		case ModeSeries:
			m.seriesTable = m.seriesTable.Focused(false)
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
	}
	return m, nil
}

func (m TUIModel) executeInstantQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		warnings, vector, err := m.promClient.Query(query, m.timeout)
		return tuiInstantResultMsg{
			warnings: warnings,
			vector:   vector,
			err:      err,
		}
	}
}

func (m TUIModel) executeRangeQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		end := time.Now()
		start := end.Add(-m.rangeValue)
		matrix, warnings, err := m.promClient.QueryRange(query, start, end, m.stepValue, m.timeout)
		return tuiRangeResultMsg{
			warnings: warnings,
			matrix:   matrix,
			err:      err,
		}
	}
}

func (m TUIModel) executeSeriesQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		end := time.Now()
		start := end.Add(-m.rangeValue)
		series, warnings, err := m.promClient.Series(query, start, end, m.seriesLimit, m.timeout)
		return tuiSeriesResultMsg{
			warnings: warnings,
			series:   series,
			err:      err,
		}
	}
}

func (m TUIModel) handleInstantResult(msg tuiInstantResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeInstant] = msg.warnings
	m.vector = msg.vector
	m.modeErrors[ModeInstant] = msg.err
	m.queryInput.Focus()
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
	m.queryInput.Focus()
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
	m.queryInput.Focus()
	m.focusedPane = PaneQuery

	if msg.err != nil {
		m.modeStates[ModeSeries] = StateError
		return m, nil
	}

	m.modeStates[ModeSeries] = StateResults
	m = m.renderSeriesTable()
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

func (m TUIModel) regenerateRangeChart() TUIModel {
	width := m.getChartWidth()
	m.chartContent, _ = charts.TimeseriesSplitWithSelection(m.matrix, width, m.selectedIndex)
	return m
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

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(entry.ColorIndex)))
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

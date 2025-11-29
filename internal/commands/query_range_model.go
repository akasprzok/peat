package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	teatable "github.com/evertras/bubble-table/table"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type queryRangeState int

const (
	stateRangeLoading queryRangeState = iota
	stateRangeSuccess
	stateRangeError
)

type queryRangeResultMsg struct {
	matrix   model.Matrix
	warnings v1.Warnings
	err      error
}

type QueryRangeModel struct {
	promClient    prometheus.Client
	query         string
	timeRange     time.Duration
	timeout       time.Duration
	output        string
	state         queryRangeState
	spinner       spinner.Model
	matrix        model.Matrix
	warnings      v1.Warnings
	err           error
	width         int
	height        int
	chartContent  string
	legendEntries []charts.LegendEntry
	legendTable   teatable.Model
	legendFocused bool
	selectedIndex int // -1 means no selection, otherwise index of selected series
	quitting      bool
}

func NewQueryRangeModel(promURL, query string, timeRange time.Duration, output string, timeout time.Duration) QueryRangeModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return QueryRangeModel{
		promClient:    prometheus.NewClient(promURL),
		query:         query,
		timeRange:     timeRange,
		timeout:       timeout,
		output:        output,
		state:         stateRangeLoading,
		spinner:       s,
		selectedIndex: -1, // No selection initially
	}
}

func (m QueryRangeModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.executeQueryRange(),
	)
}

func (m QueryRangeModel) executeQueryRange() tea.Cmd {
	return func() tea.Msg {
		end := time.Now()
		start := end.Add(-m.timeRange)
		matrix, warnings, err := m.promClient.QueryRange(m.query, start, end, 1*time.Minute, m.timeout)
		return queryRangeResultMsg{
			matrix:   matrix,
			warnings: warnings,
			err:      err,
		}
	}
}

func (m QueryRangeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleRangeWindowSize(msg), nil
	case tea.KeyMsg:
		return m.handleRangeKeyMsg(msg)
	case queryRangeResultMsg:
		return m.handleRangeQueryResult(msg)
	case spinner.TickMsg:
		return m.handleRangeSpinnerTick(msg)
	}

	return m, nil
}

func (m QueryRangeModel) handleRangeWindowSize(msg tea.WindowSizeMsg) QueryRangeModel {
	m.width = msg.Width
	m.height = msg.Height
	return m
}

func (m QueryRangeModel) handleRangeKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle global keys first
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "i":
		// Toggle interactive mode (legend focus)
		if m.state == stateRangeSuccess && m.output == "graph" {
			m.legendFocused = !m.legendFocused
			m.legendTable = m.legendTable.Focused(m.legendFocused)
			if !m.legendFocused {
				// When exiting interactive mode, clear selection
				m.selectedIndex = -1
				m = m.regenerateChart()
			}
		}
		return m, nil
	}

	// When in interactive mode, handle navigation
	if m.legendFocused {
		oldSelected := m.selectedIndex

		// Map j/k to down/up, h/l to page navigation
		var tableCmd tea.Cmd
		switch msg.String() {
		case "j":
			// Simulate down arrow
			m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyDown})
		case "k":
			// Simulate up arrow
			m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyUp})
		case "h":
			// Simulate page up
			m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyPgUp})
		case "l":
			// Simulate page down
			m.legendTable, tableCmd = m.legendTable.Update(tea.KeyMsg{Type: tea.KeyPgDown})
		default:
			// Let the table handle other keys (arrow keys, etc.)
			m.legendTable, tableCmd = m.legendTable.Update(msg)
		}

		// Update selected index after navigation
		m = m.updateSelectedFromTable()

		// Regenerate chart if selection changed
		if oldSelected != m.selectedIndex {
			m = m.regenerateChart()
		}

		return m, tableCmd
	}

	return m, nil
}

func (m QueryRangeModel) handleRangeQueryResult(msg queryRangeResultMsg) (tea.Model, tea.Cmd) {
	m.matrix = msg.matrix
	m.warnings = msg.warnings
	m.err = msg.err

	if m.err != nil {
		m.state = stateRangeError
		return m, tea.Quit
	}

	return m.handleRangeOutputFormat()
}

func (m QueryRangeModel) handleRangeOutputFormat() (tea.Model, tea.Cmd) {
	switch m.output {
	case "graph":
		return m.handleRangeGraphOutput()
	default:
		// For json/yaml, we'll just transition to success state
		m.state = stateRangeSuccess
		return m, tea.Quit
	}
}

func (m QueryRangeModel) handleRangeGraphOutput() (tea.Model, tea.Cmd) {
	m.state = stateRangeSuccess
	// Get terminal width directly
	chartWidth := m.width - 6 // Account for borders and padding
	m.chartContent, m.legendEntries = charts.TimeseriesSplitWithSelection(m.matrix, chartWidth, m.selectedIndex)
	// Create the legend table
	(&m).createLegendTable(5)
	return m, nil
}

func (m QueryRangeModel) handleRangeSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if m.state == stateRangeLoading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m QueryRangeModel) View() string {
	var s strings.Builder

	switch m.state {
	case stateRangeLoading:
		s.WriteString(fmt.Sprintf("\n%s Executing range query: %s (range: %s)\n\n", m.spinner.View(), m.query, m.timeRange))

	case stateRangeError:
		s.WriteString("\n")
		s.WriteString(errorStyle.Render("Error: ") + m.err.Error() + "\n")

	case stateRangeSuccess:
		// Show warnings if any
		if len(m.warnings) > 0 {
			s.WriteString("\n")
			s.WriteString(warningStyle.Render("Warnings:\n"))
			for _, w := range m.warnings {
				s.WriteString(warningStyle.Render(fmt.Sprintf("  • %s\n", w)))
			}
			s.WriteString("\n")
		}

		// Display based on output format
		switch m.output {
		case "graph":
			// Create styled containers for chart and legend
			chartStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1)

			legendStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1).
				MarginTop(1)

			// Style the chart and legend
			styledChart := chartStyle.Render(m.chartContent)

			// Update legend border color based on focus
			if m.legendFocused {
				legendStyle = legendStyle.BorderForeground(lipgloss.Color("205")) // Highlight when focused
			}

			// Render legend table
			styledLegend := legendStyle.Render(m.legendTable.View())

			// Join them vertically (legend below chart)
			layout := lipgloss.JoinVertical(
				lipgloss.Bottom,
				styledChart,
				styledLegend,
			)

			s.WriteString("\n")
			s.WriteString(layout)
			if !m.quitting {
				s.WriteString("\n\n")
				if m.legendFocused {
					if m.selectedIndex >= 0 {
						s.WriteString("Interactive mode • j/k: row • h/l: page • i: exit • q: quit\n")
					} else {
						s.WriteString("Interactive mode • j/k: row • h/l: page • i: exit • q: quit\n")
					}
				} else {
					s.WriteString("Press i for interactive mode, q to quit\n")
				}
			} else {
				s.WriteString("\n")
			}
		case "json":
			output := formatMatrix(m.matrix, m.warnings, m.err)
			jsonBytes, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				s.WriteString(errorStyle.Render(fmt.Sprintf("Error formatting JSON: %v\n", err)))
			} else {
				s.Write(jsonBytes)
				s.WriteString("\n")
			}
		case "yaml":
			output := formatMatrix(m.matrix, m.warnings, m.err)
			yamlBytes, err := yaml.Marshal(output)
			if err != nil {
				s.WriteString(errorStyle.Render(fmt.Sprintf("Error formatting YAML: %v\n", err)))
			} else {
				s.Write(yamlBytes)
				s.WriteString("\n")
			}
		}
	}

	return s.String()
}

// createLegendTable creates and stores the legend table in the model
func (m *QueryRangeModel) createLegendTable(maxRows int) {
	if len(m.legendEntries) == 0 {
		return
	}

	// Create rows for the table
	rows := make([]teatable.Row, 0, len(m.legendEntries))
	longestMetric := 0

	for _, entry := range m.legendEntries {
		if len(entry.Metric) > longestMetric {
			longestMetric = len(entry.Metric)
		}

		// Create colored indicator
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(entry.ColorIndex)))
		colorIndicator := style.Render("█")

		rows = append(rows, teatable.NewRow(teatable.RowData{
			"color":  colorIndicator,
			"metric": entry.Metric,
		}))
	}

	// Create columns
	columns := []teatable.Column{
		teatable.NewColumn("color", "", 3),
		teatable.NewColumn("metric", "Metric", max(longestMetric, 20)),
	}

	// Create table with max rows
	m.legendTable = teatable.
		New(columns).
		WithRows(rows).
		WithPageSize(maxRows).
		Focused(m.legendFocused)
}

// updateSelectedFromTable updates the selectedIndex based on the table's highlighted row
func (m QueryRangeModel) updateSelectedFromTable() QueryRangeModel {
	highlightedRow := m.legendTable.HighlightedRow()
	if highlightedRow.Data == nil {
		m.selectedIndex = -1
		return m
	}

	// Find the index by matching the metric name
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

// regenerateChart regenerates the chart with the current selection
func (m QueryRangeModel) regenerateChart() QueryRangeModel {
	chartWidth := m.width - 6 // Account for borders and padding
	if chartWidth <= 0 {
		chartWidth = 80
	}
	// Only regenerate the chart, keep legendEntries stable
	m.chartContent, _ = charts.TimeseriesSplitWithSelection(m.matrix, chartWidth, m.selectedIndex)
	return m
}

// GetResult returns the final result for non-interactive outputs
func (m QueryRangeModel) GetResult() (matrix model.Matrix, warnings v1.Warnings, err error) {
	return m.matrix, m.warnings, m.err
}

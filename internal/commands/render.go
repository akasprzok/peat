package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/charmbracelet/lipgloss"
	teatable "github.com/evertras/bubble-table/table"
	"github.com/prometheus/common/model"
	"golang.org/x/term"
)

func (m TUIModel) renderInstantChart() TUIModel {
	width := m.getChartWidth()
	m.chartContent = charts.Barchart(m.vector, width)
	return m
}

func (m TUIModel) renderRangeChart() TUIModel {
	width := m.getChartWidth()
	m.chartContent, m.legendEntries = charts.TimeseriesSplitWithSelection(m.matrix, width, m.selectedIndex, m.highlightedIndices)
	m = m.createLegendTable()
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
	m.chartContent, _ = charts.TimeseriesSplitWithSelection(m.matrix, width, m.selectedIndex, m.highlightedIndices)
	return m
}

func (m TUIModel) getChartWidth() int {
	width := m.width - ChartWidthPadding
	if width <= 0 {
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && termWidth > 0 {
			width = termWidth - ChartWidthPadding
		} else {
			width = DefaultTerminalWidth - ChartWidthPadding
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

func (m TUIModel) createLegendTable() TUIModel {
	if len(m.legendEntries) == 0 {
		return m
	}

	rows := make([]teatable.Row, 0, len(m.legendEntries))
	longestMetric := 0

	for i, entry := range m.legendEntries {
		if len(entry.Metric) > longestMetric {
			longestMetric = len(entry.Metric)
		}

		style := charts.SeriesStyle(entry.ColorIndex)
		colorIndicator := style.Render("\u2588")

		pin := ""
		if m.highlightedIndices[i] {
			pin = "*"
		}

		rows = append(rows, teatable.NewRow(teatable.RowData{
			"color":  colorIndicator,
			"pin":    pin,
			"metric": entry.Metric,
		}))
	}

	columns := []teatable.Column{
		teatable.NewColumn("color", "", 3),
		teatable.NewColumn("pin", "", 3),
		teatable.NewColumn("metric", "Metric", max(longestMetric, 20)),
	}

	m.legendTable = teatable.
		New(columns).
		WithRows(rows).
		WithPageSize(LegendMaxRows).
		Focused(m.legendFocused)

	return m
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

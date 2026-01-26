package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RangeMode handles range query mode (/query_range)
type RangeMode struct{}

func (RangeMode) Name() string {
	return "3) /query_range"
}

func (RangeMode) HandleInteractiveToggle(m *TUIModel) tea.Cmd {
	if len(m.legendEntries) == 0 {
		return nil
	}

	m.legendFocused = !m.legendFocused
	if m.legendFocused {
		m.focusedPane = PaneLegend
		m.queryInput.Blur()
		m.legendTable = m.legendTable.Focused(true)
		// Select the first series and redraw chart immediately
		*m = m.updateSelectedFromLegendTable()
		*m = m.regenerateRangeChart()
	} else {
		m.focusedPane = PaneQuery
		// Stay in normal mode - don't focus query input
		m.legendTable = m.legendTable.Focused(false)
		m.selectedIndex = -1
		*m = m.regenerateRangeChart()
	}
	return nil
}

func (RangeMode) HandleLegendKey(m *TUIModel, msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "q":
		return tea.Quit
	case "i", "esc":
		// Exit interactive mode
		m.legendFocused = false
		m.focusedPane = PaneQuery
		m.legendTable = m.legendTable.Focused(false)
		m.selectedIndex = -1
		*m = m.regenerateRangeChart()
		return nil
	}

	// Handle legend navigation
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

	*m = m.updateSelectedFromLegendTable()

	if oldSelected != m.selectedIndex {
		*m = m.regenerateRangeChart()
	}

	return tableCmd
}

func (RangeMode) ExecuteQuery(m *TUIModel) tea.Cmd {
	return m.executeRangeQuery()
}

func (RangeMode) RenderStatusParams(m *TUIModel) string {
	return fmt.Sprintf("   Range: %s   Step: %s", m.rangeValue, m.stepValue)
}

func (RangeMode) RenderResultsContent(m *TUIModel) string {
	var s strings.Builder

	// Warnings
	warnings := m.currentWarnings()
	if len(warnings) > 0 {
		s.WriteString("\n")
		s.WriteString(WarningStyle.Render("Warnings:\n"))
		for _, w := range warnings {
			s.WriteString(WarningStyle.Render("  - " + w + "\n"))
		}
	}

	// Chart
	chartStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	if m.focusedPane == PaneResults {
		chartStyle = chartStyle.BorderForeground(lipgloss.Color("205"))
	}

	s.WriteString(chartStyle.Render(m.chartContent))

	// Legend
	if len(m.legendEntries) > 0 {
		legendStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			MarginTop(1)

		if m.legendFocused {
			legendStyle = legendStyle.BorderForeground(lipgloss.Color("205"))
		}

		s.WriteString(legendStyle.Render(m.legendTable.View()))
	}
	return s.String()
}

func (RangeMode) RenderResultsStatusBar(m *TUIModel) string {
	// No additional status bar content for range mode
	return ""
}

func (RangeMode) OnSwitchTo(m *TUIModel) {
	if m.currentState() == StateResults {
		*m = m.renderRangeChart()
	}
}

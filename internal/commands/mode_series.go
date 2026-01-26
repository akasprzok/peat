package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SeriesMode handles series query mode (/series)
type SeriesMode struct{}

func (SeriesMode) Name() string {
	return "4) /series"
}

func (SeriesMode) HandleInteractiveToggle(m *TUIModel) tea.Cmd {
	if len(m.series) == 0 {
		return nil
	}

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
	return nil
}

func (SeriesMode) HandleLegendKey(m *TUIModel, msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "q":
		return tea.Quit
	case "i", "esc":
		// Exit interactive mode
		m.legendFocused = false
		m.focusedPane = PaneQuery
		m.seriesTable = m.seriesTable.Focused(false)
		return nil
	}

	// Handle table navigation
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
	return tableCmd
}

func (SeriesMode) ExecuteQuery(m *TUIModel) tea.Cmd {
	return m.executeSeriesQuery()
}

func (SeriesMode) RenderStatusParams(m *TUIModel) string {
	return fmt.Sprintf("   Range: %s   Limit: %d", m.rangeValue, m.seriesLimit)
}

func (SeriesMode) RenderResultsContent(m *TUIModel) string {
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

	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	if m.legendFocused {
		tableStyle = tableStyle.BorderForeground(lipgloss.Color("205"))
	}

	s.WriteString(tableStyle.Render(m.seriesTable.View()))
	s.WriteString("\n")
	return s.String()
}

func (SeriesMode) RenderResultsStatusBar(m *TUIModel) string {
	if len(m.series) > 0 {
		return fmt.Sprintf(" | Series: %d", len(m.series))
	}
	return ""
}

func (SeriesMode) OnSwitchTo(m *TUIModel) {
	if m.currentState() == StateResults {
		*m = m.renderSeriesTable()
	}
}

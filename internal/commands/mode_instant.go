package commands

import (
	"strings"

	"github.com/akasprzok/peat/internal/charts"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstantMode handles instant query mode (/query)
type InstantMode struct{}

func (InstantMode) Name() string {
	return "/query"
}

func (InstantMode) HandleInteractiveToggle(m *TUIModel) tea.Cmd {
	// Instant mode has no interactive elements
	return nil
}

func (InstantMode) HandleLegendKey(m *TUIModel, msg tea.KeyMsg) tea.Cmd {
	// Instant mode has no legend/table navigation
	return nil
}

func (InstantMode) ExecuteQuery(m *TUIModel) tea.Cmd {
	return m.executeInstantQuery()
}

func (InstantMode) RenderStatusParams(m *TUIModel) string {
	// No additional params for instant mode
	return ""
}

func (InstantMode) RenderResultsContent(m *TUIModel) string {
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
	return s.String()
}

func (InstantMode) RenderResultsStatusBar(m *TUIModel) string {
	// No additional status bar content for instant mode
	return ""
}

func (InstantMode) RenderHelpText(m *TUIModel, focusedState string) string {
	switch focusedState {
	case "insert":
		return " Editing query | Enter: run | Esc: exit | Tab: mode"
	default:
		// Normal mode - instant has no interactive elements
		return "  Tab: mode | Enter: run | /: edit query | f: format | q: quit"
	}
}

func (InstantMode) OnSwitchTo(m *TUIModel) {
	if m.currentState() == StateResults {
		m.chartContent = charts.Barchart(m.vector, m.getChartWidth())
	}
}

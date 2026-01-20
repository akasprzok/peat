package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

	activeStyle := modeStyle.Background(lipgloss.Color("63")).Foreground(lipgloss.Color("231"))

	switch m.mode {
	case ModeInstant:
		instantStyle = activeStyle
	case ModeRange:
		rangeStyle = activeStyle
	case ModeSeries:
		seriesStyle = activeStyle
	}

	modeText := fmt.Sprintf("  Mode: %s | %s | %s",
		instantStyle.Render(" Instant "),
		rangeStyle.Render(" Range "),
		seriesStyle.Render(" Series "))

	// Parameters (show for range and series modes)
	paramsText := ""
	switch m.mode {
	case ModeInstant:
		// No additional params for instant mode
	case ModeRange:
		paramsText = fmt.Sprintf("   Range: %s   Step: %s", m.rangeValue, m.stepValue)
	case ModeSeries:
		paramsText = fmt.Sprintf("   Range: %s   Limit: %d", m.rangeValue, m.seriesLimit)
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

	if m.focusedPane == PaneQuery {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("205"))
	} else {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("63"))
	}

	label := lipgloss.NewStyle().Bold(true).Render("Query: ")
	return inputStyle.Render(label + m.queryInput.View())
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

func (m TUIModel) renderHelpBar() string {
	helpStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.width).
		Padding(0, 1)

	var helpText string
	if m.legendFocused {
		helpText = "  j/k: navigate | h/l: page | i/esc: exit interactive | q: quit"
	} else {
		helpText = "  Tab: mode | Enter: run | /: edit query | f: format"
		if m.currentState() == StateResults {
			if m.mode == ModeRange && len(m.legendEntries) > 0 {
				helpText += " | i: interactive"
			} else if m.mode == ModeSeries && len(m.series) > 0 {
				helpText += " | i: interactive"
			}
		}
		helpText += " | q: quit"
	}

	return helpStyle.Render(helpText)
}

package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m TUIModel) View() string {
	// Show shortcuts overlay if active
	if m.showShortcutsOverlay {
		overlay := renderShortcutsOverlay()
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
	modeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
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
		content = renderEmptyState()
	case StateLoading:
		content = m.renderLoadingState()
	case StateError:
		content = m.renderErrorState()
	case StateResults:
		content = m.renderResultsContent()
	}

	return content
}

func renderEmptyState() string {
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

func renderShortcutsOverlay() string {
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

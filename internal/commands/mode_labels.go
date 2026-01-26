package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LabelsMode handles labels query mode (/labels)
type LabelsMode struct{}

func (LabelsMode) Name() string {
	return "2) /labels"
}

func (LabelsMode) HandleInteractiveToggle(m *TUIModel) tea.Cmd {
	if len(m.labels) == 0 {
		return nil
	}

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
	return nil
}

func (LabelsMode) HandleLegendKey(m *TUIModel, msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "q":
		return tea.Quit
	case "i", "esc":
		// If viewing values, go back to labels list
		if m.viewingLabelValues {
			m.viewingLabelValues = false
			m.labelValues = nil
			m.selectedLabelName = ""
			*m = m.renderLabelsTable()
			return nil
		}
		// Exit interactive mode
		m.legendFocused = false
		m.focusedPane = PaneQuery
		m.labelsTable = m.labelsTable.Focused(false)
		return nil
	case "enter":
		// Query values for the selected label
		if !m.viewingLabelValues {
			highlightedRow := m.labelsTable.HighlightedRow()
			if highlightedRow.Data != nil {
				if labelName, ok := highlightedRow.Data["label"].(string); ok {
					m.selectedLabelIndex = m.labelsTable.GetHighlightedRowIndex()
					m.selectedLabelName = labelName
					m.modeStates[ModeLabels] = StateLoading
					return m.executeLabelValuesQuery(labelName)
				}
			}
		}
		return nil
	}

	// Handle table navigation
	var tableCmd tea.Cmd
	switch key {
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
	return tableCmd
}

func (LabelsMode) ExecuteQuery(m *TUIModel) tea.Cmd {
	return m.executeLabelsQuery()
}

func (LabelsMode) RenderStatusParams(m *TUIModel) string {
	return fmt.Sprintf("   Range: %s", m.rangeValue)
}

func (LabelsMode) RenderResultsContent(m *TUIModel) string {
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

	s.WriteString(tableStyle.Render(m.labelsTable.View()))
	s.WriteString("\n")
	return s.String()
}

func (LabelsMode) RenderResultsStatusBar(m *TUIModel) string {
	if m.viewingLabelValues {
		return fmt.Sprintf(" | Labels: %d | Values : %d", len(m.labels), len(m.labelValues))
	}
	return fmt.Sprintf(" | Labels: %d ", len(m.labels))
}

func (LabelsMode) OnSwitchTo(m *TUIModel) {
	if m.currentState() == StateResults {
		*m = m.renderLabelsTable()
	}
}

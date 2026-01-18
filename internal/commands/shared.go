package commands

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Shared styles used across command models.
var (
	SpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

// NewLoadingSpinner creates a spinner with consistent styling for loading states.
func NewLoadingSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle
	return s
}

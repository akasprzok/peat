package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Mode defines the interface for query mode implementations.
// Each mode (Instant, Range, Series, Labels) implements this interface
// to handle mode-specific logic for interactive behavior, query execution,
// and rendering.
type Mode interface {
	// Name returns the display name of the mode (e.g., "/query", "/query_range")
	Name() string

	// HandleInteractiveToggle handles the 'i' key press to toggle interactive mode
	HandleInteractiveToggle(m *TUIModel) tea.Cmd

	// HandleLegendKey handles navigation keys when in interactive/legend mode
	HandleLegendKey(m *TUIModel, msg tea.KeyMsg) tea.Cmd

	// ExecuteQuery executes the appropriate query for this mode
	ExecuteQuery(m *TUIModel) tea.Cmd

	// RenderStatusParams returns the mode-specific parameters for the status bar
	RenderStatusParams(m *TUIModel) string

	// RenderResultsContent renders the main results content for this mode
	RenderResultsContent(m *TUIModel) string

	// RenderResultsStatusBar returns additional status bar content for results
	RenderResultsStatusBar(m *TUIModel) string

	// OnSwitchTo is called when switching to this mode
	OnSwitchTo(m *TUIModel)
}

// Mode registry - maps QueryMode to Mode implementations
var modes = map[QueryMode]Mode{
	ModeInstant: InstantMode{},
	ModeRange:   RangeMode{},
	ModeSeries:  SeriesMode{},
	ModeLabels:  LabelsMode{},
}

// currentMode returns the Mode implementation for the current mode
func (m TUIModel) currentMode() Mode {
	return modes[m.mode]
}

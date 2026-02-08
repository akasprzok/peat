package commands

import (
	"fmt"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	teatable "github.com/evertras/bubble-table/table"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// TUIModel is the main Bubble Tea model for the interactive TUI.
type TUIModel struct {
	promClient prometheus.Client
	timeout    time.Duration

	// Input
	queryInput textinput.Model

	// Mode
	mode QueryMode

	// Per-mode state (indexed by QueryMode)
	modeQueries   [4]string        // Query string for each mode
	modeStates    [4]TUIState      // State for each mode
	modeWarnings  [4]v1.Warnings   // Warnings for each mode
	modeErrors    [4]error         // Errors for each mode
	modeDurations [4]time.Duration // Query execution duration for each mode

	// Range query parameters
	rangeValue time.Duration
	stepValue  time.Duration

	// Series query parameters
	seriesLimit uint64

	// Results (already per-mode by nature)
	vector model.Vector     // For instant queries
	matrix model.Matrix     // For range queries
	series []model.LabelSet // For series queries
	labels []string         // For labels queries

	// Label values state
	labelValues        []string // Values for selected label
	selectedLabelName  string   // Currently selected label name
	viewingLabelValues bool     // True when showing values instead of names
	selectedLabelIndex int      // Index of selected label row (for restoring position)

	// Rendered content
	chartContent       string
	legendEntries      []charts.LegendEntry
	legendTable        teatable.Model
	seriesTable        teatable.Model
	labelsTable        teatable.Model
	selectedIndex      int          // -1 means no selection
	highlightedIndices map[int]bool // pinned series indices for multi-series display

	// UI state
	width                int
	height               int
	focusedPane          FocusedPane
	insertMode           bool // true when editing query (insert mode), false for normal mode
	spinner              spinner.Model
	legendFocused        bool
	showShortcutsOverlay bool
}

// NewTUIModel creates a new TUI model.
func NewTUIModel(client prometheus.Client, rangeValue, stepValue time.Duration, seriesLimit uint64, timeout time.Duration) TUIModel {
	ti := textinput.New()
	ti.Placeholder = "Enter PromQL query..."
	ti.Focus()
	ti.Width = 60

	return TUIModel{
		promClient:         client,
		timeout:            timeout,
		queryInput:         ti,
		mode:               ModeInstant,
		modeStates:         [4]TUIState{StateInput, StateInput, StateInput, StateInput},
		rangeValue:         rangeValue,
		stepValue:          stepValue,
		seriesLimit:        seriesLimit,
		selectedIndex:      -1,
		highlightedIndices: make(map[int]bool),
		focusedPane:        PaneQuery,
		insertMode:         true, // Start in insert mode so users can immediately type
		spinner:            NewLoadingSpinner(),
	}
}

// Helper methods for accessing current mode's state
func (m TUIModel) currentState() TUIState {
	return m.modeStates[m.mode]
}

func (m TUIModel) currentWarnings() v1.Warnings {
	return m.modeWarnings[m.mode]
}

func (m TUIModel) currentError() error {
	return m.modeErrors[m.mode]
}

func (m TUIModel) currentDuration() time.Duration {
	return m.modeDurations[m.mode]
}

// applyResultCommon applies common result handling for all query result handlers.
func (m TUIModel) applyResultCommon(mode QueryMode, warnings v1.Warnings, err error, duration time.Duration) TUIModel {
	m.modeWarnings[mode] = warnings
	m.modeErrors[mode] = err
	m.modeDurations[mode] = duration
	m.queryInput.Blur()
	m.insertMode = false
	m.focusedPane = PaneQuery
	return m
}

// formatDuration formats a duration with appropriate precision.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

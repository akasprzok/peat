package tui

import (
	"os"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	teatable "github.com/evertras/bubble-table/table"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/term"
)

// QueryMode represents whether we're doing instant or range queries.
type QueryMode int

const (
	ModeInstant QueryMode = iota
	ModeRange
	ModeSeries
)

func (m QueryMode) String() string {
	switch m {
	case ModeInstant:
		return "Instant"
	case ModeRange:
		return "Range"
	case ModeSeries:
		return "Series"
	default:
		return "Unknown"
	}
}

// TUIState represents the current state of the TUI.
type TUIState int

const (
	StateInput TUIState = iota
	StateLoading
	StateResults
	StateError
)

// FocusedPane tracks which pane has focus.
type FocusedPane int

const (
	PaneQuery FocusedPane = iota
	PaneResults
	PaneLegend
)

// tuiInstantResultMsg carries the result of an instant query.
type tuiInstantResultMsg struct {
	warnings v1.Warnings
	vector   model.Vector
	err      error
}

// tuiRangeResultMsg carries the result of a range query.
type tuiRangeResultMsg struct {
	warnings v1.Warnings
	matrix   model.Matrix
	err      error
}

// tuiSeriesResultMsg carries the result of a series query.
type tuiSeriesResultMsg struct {
	warnings v1.Warnings
	series   []model.LabelSet
	err      error
}

// TUIModel is the main Bubble Tea model for the interactive TUI.
type TUIModel struct {
	promClient prometheus.Client
	timeout    time.Duration

	// Input
	queryInput textinput.Model

	// Mode
	mode QueryMode

	// Per-mode state (indexed by QueryMode)
	modeQueries  [3]string      // Query string for each mode
	modeStates   [3]TUIState    // State for each mode
	modeWarnings [3]v1.Warnings // Warnings for each mode
	modeErrors   [3]error       // Errors for each mode

	// Range query parameters
	rangeValue time.Duration
	stepValue  time.Duration

	// Series query parameters
	seriesLimit uint64

	// Results (already per-mode by nature)
	vector model.Vector     // For instant queries
	matrix model.Matrix     // For range queries
	series []model.LabelSet // For series queries

	// Rendered content
	chartContent  string
	legendEntries []charts.LegendEntry
	legendTable   teatable.Model
	seriesTable   teatable.Model
	selectedIndex int // -1 means no selection

	// UI state
	width         int
	height        int
	focusedPane   FocusedPane
	spinner       spinner.Model
	legendFocused bool
}

// NewTUIModel creates a new TUI model.
func NewTUIModel(client prometheus.Client, rangeValue, stepValue time.Duration, seriesLimit uint64, timeout time.Duration) TUIModel {
	ti := textinput.New()
	ti.Placeholder = "Enter PromQL query..."
	ti.Focus()
	ti.Width = 60

	return TUIModel{
		promClient:    client,
		timeout:       timeout,
		queryInput:    ti,
		mode:          ModeInstant,
		modeStates:    [3]TUIState{StateInput, StateInput, StateInput},
		rangeValue:    rangeValue,
		stepValue:     stepValue,
		seriesLimit:   seriesLimit,
		selectedIndex: -1,
		focusedPane:   PaneQuery,
		spinner:       NewLoadingSpinner(),
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

func (m TUIModel) getChartWidth() int {
	width := m.width - 6
	if width <= 0 {
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && termWidth > 0 {
			width = termWidth - 6
		} else {
			width = DefaultTerminalWidth - 6
		}
	}
	return width
}

package commands

import (
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// QueryMode represents whether we're doing instant or range queries.
type QueryMode int

const (
	ModeInstant QueryMode = iota
	ModeRange
	ModeSeries
	ModeLabels
)

func (m QueryMode) String() string {
	switch m {
	case ModeInstant:
		return "/query"
	case ModeRange:
		return "/query_range"
	case ModeSeries:
		return "/series"
	case ModeLabels:
		return "/labels"
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
	duration time.Duration
}

// tuiRangeResultMsg carries the result of a range query.
type tuiRangeResultMsg struct {
	warnings v1.Warnings
	matrix   model.Matrix
	err      error
	duration time.Duration
}

// tuiSeriesResultMsg carries the result of a series query.
type tuiSeriesResultMsg struct {
	warnings v1.Warnings
	series   []model.LabelSet
	err      error
	duration time.Duration
}

// tuiLabelsResultMsg carries the result of a labels query.
type tuiLabelsResultMsg struct {
	warnings v1.Warnings
	labels   []string
	err      error
	duration time.Duration
}

// tuiLabelValuesResultMsg carries the result of a label values query.
type tuiLabelValuesResultMsg struct {
	labelName string
	warnings  v1.Warnings
	values    []string
	err       error
	duration  time.Duration
}

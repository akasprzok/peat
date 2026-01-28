package commands

import (
	"errors"
	"testing"
	"time"

	"github.com/akasprzok/peat/internal/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"microseconds", 500 * time.Microsecond, "500µs"},
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"seconds", 1500 * time.Millisecond, "1.5s"},
		{"boundary - just under ms", 999 * time.Microsecond, "999µs"},
		{"boundary - just under s", 999 * time.Millisecond, "999ms"},
		{"zero", 0, "0µs"},
		{"exactly 1ms", time.Millisecond, "1ms"},
		{"exactly 1s", time.Second, "1.0s"},
		{"large seconds", 10 * time.Second, "10.0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestNewTUIModel(t *testing.T) {
	mockClient := &prometheus.MockClient{}
	rangeValue := 1 * time.Hour
	stepValue := 15 * time.Second
	seriesLimit := uint64(100)
	timeout := 60 * time.Second

	m := NewTUIModel(mockClient, rangeValue, stepValue, seriesLimit, timeout)

	t.Run("initial mode is ModeInstant", func(t *testing.T) {
		if m.mode != ModeInstant {
			t.Errorf("mode = %v, want %v", m.mode, ModeInstant)
		}
	})

	t.Run("starts in insert mode", func(t *testing.T) {
		if !m.insertMode {
			t.Error("insertMode = false, want true")
		}
	})

	t.Run("all mode states initialized to StateInput", func(t *testing.T) {
		for i, state := range m.modeStates {
			if state != StateInput {
				t.Errorf("modeStates[%d] = %v, want %v", i, state, StateInput)
			}
		}
	})

	t.Run("selectedIndex starts at -1", func(t *testing.T) {
		if m.selectedIndex != -1 {
			t.Errorf("selectedIndex = %d, want -1", m.selectedIndex)
		}
	})

	t.Run("parameters stored correctly", func(t *testing.T) {
		if m.rangeValue != rangeValue {
			t.Errorf("rangeValue = %v, want %v", m.rangeValue, rangeValue)
		}
		if m.stepValue != stepValue {
			t.Errorf("stepValue = %v, want %v", m.stepValue, stepValue)
		}
		if m.seriesLimit != seriesLimit {
			t.Errorf("seriesLimit = %d, want %d", m.seriesLimit, seriesLimit)
		}
		if m.timeout != timeout {
			t.Errorf("timeout = %v, want %v", m.timeout, timeout)
		}
	})

	t.Run("focusedPane starts at PaneQuery", func(t *testing.T) {
		if m.focusedPane != PaneQuery {
			t.Errorf("focusedPane = %v, want %v", m.focusedPane, PaneQuery)
		}
	})
}

func TestCurrentStateAccessors(t *testing.T) {
	mockClient := &prometheus.MockClient{}
	m := NewTUIModel(mockClient, time.Hour, 15*time.Second, 100, 60*time.Second)

	t.Run("currentState returns state for current mode", func(t *testing.T) {
		if m.currentState() != StateInput {
			t.Errorf("currentState() = %v, want %v", m.currentState(), StateInput)
		}

		// Modify state for current mode
		m.modeStates[m.mode] = StateLoading
		if m.currentState() != StateLoading {
			t.Errorf("currentState() = %v, want %v", m.currentState(), StateLoading)
		}
	})

	t.Run("currentWarnings returns warnings for current mode", func(t *testing.T) {
		// Initially nil
		if m.currentWarnings() != nil {
			t.Errorf("currentWarnings() = %v, want nil", m.currentWarnings())
		}

		// Set warnings
		warnings := v1.Warnings{"test warning"}
		m.modeWarnings[m.mode] = warnings
		if len(m.currentWarnings()) != 1 || m.currentWarnings()[0] != "test warning" {
			t.Errorf("currentWarnings() = %v, want %v", m.currentWarnings(), warnings)
		}
	})

	t.Run("currentError returns error for current mode", func(t *testing.T) {
		// Initially nil
		if m.currentError() != nil {
			t.Errorf("currentError() = %v, want nil", m.currentError())
		}

		// Set error
		err := errors.New("test error")
		m.modeErrors[m.mode] = err
		if m.currentError() != err {
			t.Errorf("currentError() = %v, want %v", m.currentError(), err)
		}
	})

	t.Run("currentDuration returns duration for current mode", func(t *testing.T) {
		// Initially zero
		if m.currentDuration() != 0 {
			t.Errorf("currentDuration() = %v, want 0", m.currentDuration())
		}

		// Set duration
		duration := 500 * time.Millisecond
		m.modeDurations[m.mode] = duration
		if m.currentDuration() != duration {
			t.Errorf("currentDuration() = %v, want %v", m.currentDuration(), duration)
		}
	})

	t.Run("switching modes returns different values", func(t *testing.T) {
		// Set up different values for different modes
		m.modeStates[ModeInstant] = StateResults
		m.modeStates[ModeRange] = StateError
		m.modeErrors[ModeInstant] = errors.New("instant error")
		m.modeErrors[ModeRange] = errors.New("range error")

		m.mode = ModeInstant
		if m.currentState() != StateResults {
			t.Errorf("currentState() for ModeInstant = %v, want %v", m.currentState(), StateResults)
		}
		if m.currentError().Error() != "instant error" {
			t.Errorf("currentError() for ModeInstant = %v, want 'instant error'", m.currentError())
		}

		m.mode = ModeRange
		if m.currentState() != StateError {
			t.Errorf("currentState() for ModeRange = %v, want %v", m.currentState(), StateError)
		}
		if m.currentError().Error() != "range error" {
			t.Errorf("currentError() for ModeRange = %v, want 'range error'", m.currentError())
		}
	})
}

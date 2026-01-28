package tables

import (
	"testing"

	"github.com/prometheus/common/model"
)

func TestPrintQuery(t *testing.T) {
	t.Run("empty vector returns model with 0 rows", func(t *testing.T) {
		vector := model.Vector{}
		m, err := PrintQuery(vector)
		if err != nil {
			t.Fatalf("PrintQuery() returned error: %v", err)
		}

		// The model should be created successfully
		view := m.View()
		if len(view) == 0 {
			t.Error("View() returned empty string")
		}
	})

	t.Run("single sample creates valid model", func(t *testing.T) {
		now := model.Now()
		vector := model.Vector{
			&model.Sample{
				Metric:    model.Metric{"__name__": "test_metric"},
				Value:     42.5,
				Timestamp: now,
			},
		}

		m, err := PrintQuery(vector)
		if err != nil {
			t.Fatalf("PrintQuery() returned error: %v", err)
		}

		view := m.View()
		if len(view) == 0 {
			t.Error("View() returned empty string")
		}
	})

	t.Run("multiple samples with varying values", func(t *testing.T) {
		now := model.Now()
		vector := model.Vector{
			&model.Sample{
				Metric:    model.Metric{"__name__": "metric_a", "job": "test"},
				Value:     100,
				Timestamp: now,
			},
			&model.Sample{
				Metric:    model.Metric{"__name__": "metric_b", "job": "test"},
				Value:     999999.99,
				Timestamp: now,
			},
			&model.Sample{
				Metric:    model.Metric{"__name__": "metric_c", "job": "test"},
				Value:     0.001,
				Timestamp: now,
			},
		}

		m, err := PrintQuery(vector)
		if err != nil {
			t.Fatalf("PrintQuery() returned error: %v", err)
		}

		view := m.View()
		if len(view) == 0 {
			t.Error("View() returned empty string")
		}
	})

	t.Run("handles long metric names", func(t *testing.T) {
		now := model.Now()
		longName := "this_is_a_very_long_metric_name_that_should_affect_column_width_calculation"
		vector := model.Vector{
			&model.Sample{
				Metric:    model.Metric{"__name__": model.LabelValue(longName)},
				Value:     1.0,
				Timestamp: now,
			},
		}

		m, err := PrintQuery(vector)
		if err != nil {
			t.Fatalf("PrintQuery() returned error: %v", err)
		}

		view := m.View()
		if len(view) == 0 {
			t.Error("View() returned empty string")
		}
	})

	t.Run("handles special float values", func(t *testing.T) {
		now := model.Now()
		vector := model.Vector{
			&model.Sample{
				Metric:    model.Metric{"__name__": "decimal_value"},
				Value:     3.14159265359,
				Timestamp: now,
			},
			&model.Sample{
				Metric:    model.Metric{"__name__": "large_number"},
				Value:     1234567890,
				Timestamp: now,
			},
			&model.Sample{
				Metric:    model.Metric{"__name__": "small_number"},
				Value:     0.0000001,
				Timestamp: now,
			},
		}

		m, err := PrintQuery(vector)
		if err != nil {
			t.Fatalf("PrintQuery() returned error: %v", err)
		}

		view := m.View()
		if len(view) == 0 {
			t.Error("View() returned empty string")
		}
	})
}

func TestModelInit(t *testing.T) {
	vector := model.Vector{}
	m, err := PrintQuery(vector)
	if err != nil {
		t.Fatalf("PrintQuery() returned error: %v", err)
	}

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

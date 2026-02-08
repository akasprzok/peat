package charts

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
)

func TestTimeseriesSplit(t *testing.T) {
	t.Run("empty matrix returns empty legend", func(t *testing.T) {
		matrix := model.Matrix{}
		_, legend := TimeseriesSplit(matrix, 80)

		if len(legend) != 0 {
			t.Errorf("len(legend) = %d, want 0", len(legend))
		}
	})

	t.Run("single series returns 1 legend entry", func(t *testing.T) {
		now := model.Now()
		matrix := model.Matrix{
			&model.SampleStream{
				Metric: model.Metric{"__name__": "test_metric"},
				Values: []model.SamplePair{
					{Timestamp: now, Value: 1.0},
					{Timestamp: now.Add(time.Minute), Value: 2.0},
				},
			},
		}

		_, legend := TimeseriesSplit(matrix, 80)

		if len(legend) != 1 {
			t.Fatalf("len(legend) = %d, want 1", len(legend))
		}
		if legend[0].ColorIndex != 0 {
			t.Errorf("legend[0].ColorIndex = %d, want 0", legend[0].ColorIndex)
		}
		// model.Metric.String() outputs just the metric name for single-label metrics
		if legend[0].Metric != "test_metric" {
			t.Errorf("legend[0].Metric = %q, want %q", legend[0].Metric, "test_metric")
		}
	})

	t.Run("multiple series returns legend entries in order", func(t *testing.T) {
		now := model.Now()
		matrix := model.Matrix{
			&model.SampleStream{
				Metric: model.Metric{"__name__": "metric_a"},
				Values: []model.SamplePair{{Timestamp: now, Value: 1.0}},
			},
			&model.SampleStream{
				Metric: model.Metric{"__name__": "metric_b"},
				Values: []model.SamplePair{{Timestamp: now, Value: 2.0}},
			},
			&model.SampleStream{
				Metric: model.Metric{"__name__": "metric_c"},
				Values: []model.SamplePair{{Timestamp: now, Value: 3.0}},
			},
		}

		_, legend := TimeseriesSplit(matrix, 80)

		if len(legend) != 3 {
			t.Fatalf("len(legend) = %d, want 3", len(legend))
		}

		for i, entry := range legend {
			if entry.ColorIndex != i {
				t.Errorf("legend[%d].ColorIndex = %d, want %d", i, entry.ColorIndex, i)
			}
		}
	})

	t.Run("chart output is non-empty for valid data", func(t *testing.T) {
		now := model.Now()
		matrix := model.Matrix{
			&model.SampleStream{
				Metric: model.Metric{"__name__": "test_metric"},
				Values: []model.SamplePair{
					{Timestamp: now, Value: 1.0},
					{Timestamp: now.Add(time.Minute), Value: 2.0},
				},
			},
		}

		chart, _ := TimeseriesSplit(matrix, 80)

		if len(chart) == 0 {
			t.Error("chart output is empty, want non-empty")
		}
	})
}

func TestTimeseriesSplitWithSelection(t *testing.T) {
	now := model.Now()
	matrix := model.Matrix{
		&model.SampleStream{
			Metric: model.Metric{"__name__": "metric_a"},
			Values: []model.SamplePair{
				{Timestamp: now, Value: 1.0},
				{Timestamp: now.Add(time.Minute), Value: 2.0},
			},
		},
		&model.SampleStream{
			Metric: model.Metric{"__name__": "metric_b"},
			Values: []model.SamplePair{
				{Timestamp: now, Value: 10.0},
				{Timestamp: now.Add(time.Minute), Value: 20.0},
			},
		},
	}

	t.Run("selectedIndex=-1 shows all series", func(t *testing.T) {
		_, legend := TimeseriesSplitWithSelection(matrix, 80, -1, nil)

		if len(legend) != 2 {
			t.Errorf("len(legend) = %d, want 2", len(legend))
		}
	})

	t.Run("selectedIndex=0 highlights first series", func(t *testing.T) {
		chart, legend := TimeseriesSplitWithSelection(matrix, 80, 0, nil)

		// Legend still contains all entries
		if len(legend) != 2 {
			t.Errorf("len(legend) = %d, want 2", len(legend))
		}

		// Chart should be non-empty
		if len(chart) == 0 {
			t.Error("chart output is empty, want non-empty")
		}
	})

	t.Run("selectedIndex=1 highlights second series", func(t *testing.T) {
		chart, legend := TimeseriesSplitWithSelection(matrix, 80, 1, nil)

		if len(legend) != 2 {
			t.Errorf("len(legend) = %d, want 2", len(legend))
		}

		if len(chart) == 0 {
			t.Error("chart output is empty, want non-empty")
		}
	})

	t.Run("selectedIndex out of range is handled", func(t *testing.T) {
		// Should not panic with out-of-range selection
		chart, legend := TimeseriesSplitWithSelection(matrix, 80, 99, nil)

		if len(legend) != 2 {
			t.Errorf("len(legend) = %d, want 2", len(legend))
		}

		// Chart should still render (possibly empty)
		_ = chart
	})

	t.Run("highlighted indices are visible alongside selection", func(t *testing.T) {
		highlighted := map[int]bool{0: true}
		chart, legend := TimeseriesSplitWithSelection(matrix, 80, 1, highlighted)

		if len(legend) != 2 {
			t.Errorf("len(legend) = %d, want 2", len(legend))
		}

		if len(chart) == 0 {
			t.Error("chart output is empty, want non-empty")
		}
	})

	t.Run("highlighted indices without selection shows all", func(t *testing.T) {
		highlighted := map[int]bool{0: true}
		chart, legend := TimeseriesSplitWithSelection(matrix, 80, -1, highlighted)

		if len(legend) != 2 {
			t.Errorf("len(legend) = %d, want 2", len(legend))
		}

		if len(chart) == 0 {
			t.Error("chart output is empty, want non-empty")
		}
	})
}

package charts

import (
	"strings"
	"testing"

	"github.com/prometheus/common/model"
)

func TestBarchart(t *testing.T) {
	tests := []struct {
		name   string
		vector model.Vector
		width  int
	}{
		{
			name:   "empty vector",
			vector: model.Vector{},
			width:  80,
		},
		{
			name: "single sample",
			vector: model.Vector{
				&model.Sample{
					Metric: model.Metric{"__name__": "up", "job": "prometheus"},
					Value:  1,
				},
			},
			width: 80,
		},
		{
			name: "multiple samples",
			vector: model.Vector{
				&model.Sample{
					Metric: model.Metric{"__name__": "up", "job": "prometheus"},
					Value:  1,
				},
				&model.Sample{
					Metric: model.Metric{"__name__": "up", "job": "node"},
					Value:  5,
				},
				&model.Sample{
					Metric: model.Metric{"__name__": "up", "job": "grafana"},
					Value:  3,
				},
			},
			width: 100,
		},
		{
			name: "narrow width",
			vector: model.Vector{
				&model.Sample{
					Metric: model.Metric{"__name__": "test"},
					Value:  10,
				},
			},
			width: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Barchart(tt.vector, tt.width)

			// For non-empty vectors, check that the result contains some output
			if len(tt.vector) > 0 && len(result) == 0 {
				t.Errorf("Barchart() returned empty string for non-empty vector")
			}

			// Check that the metric names appear in the output for non-empty vectors
			for _, sample := range tt.vector {
				metricStr := sample.Metric.String()
				if !strings.Contains(result, metricStr) {
					t.Errorf("Barchart() output does not contain metric %s", metricStr)
				}
			}
		})
	}
}

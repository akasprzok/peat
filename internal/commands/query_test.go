package commands

import (
	"errors"
	"testing"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func TestFormatVector(t *testing.T) {
	tests := []struct {
		name     string
		vector   model.Vector
		warnings v1.Warnings
		err      error
		wantErr  bool
	}{
		{
			name:     "empty vector",
			vector:   model.Vector{},
			warnings: nil,
			err:      nil,
			wantErr:  false,
		},
		{
			name: "single sample",
			vector: model.Vector{
				&model.Sample{
					Metric:    model.Metric{"__name__": "up", "job": "prometheus"},
					Value:     1,
					Timestamp: 1234567890000,
				},
			},
			warnings: nil,
			err:      nil,
			wantErr:  false,
		},
		{
			name: "multiple samples",
			vector: model.Vector{
				&model.Sample{
					Metric:    model.Metric{"__name__": "up", "job": "prometheus"},
					Value:     1,
					Timestamp: 1234567890000,
				},
				&model.Sample{
					Metric:    model.Metric{"__name__": "up", "job": "node"},
					Value:     0,
					Timestamp: 1234567890000,
				},
			},
			warnings: nil,
			err:      nil,
			wantErr:  false,
		},
		{
			name:     "with warnings",
			vector:   model.Vector{},
			warnings: v1.Warnings{"warning 1", "warning 2"},
			err:      nil,
			wantErr:  false,
		},
		{
			name:     "with error",
			vector:   model.Vector{},
			warnings: nil,
			err:      errors.New("test error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVector(tt.vector, tt.warnings, tt.err)

			// Check data field exists
			data, ok := result["data"].([]map[string]any)
			if !ok {
				t.Errorf("formatVector() data field is not []map[string]any")
				return
			}

			// Check data length matches vector length
			if len(data) != len(tt.vector) {
				t.Errorf("formatVector() data length = %d, want %d", len(data), len(tt.vector))
			}

			// Check warnings field
			if tt.warnings != nil {
				warnings, ok := result["warnings"].(v1.Warnings)
				if !ok {
					t.Errorf("formatVector() warnings field is not v1.Warnings")
				}
				if len(warnings) != len(tt.warnings) {
					t.Errorf("formatVector() warnings length = %d, want %d", len(warnings), len(tt.warnings))
				}
			}

			// Check error field
			if tt.wantErr {
				errStr, ok := result["error"].(string)
				if !ok {
					t.Errorf("formatVector() error field is not string when error is expected")
				}
				if errStr != tt.err.Error() {
					t.Errorf("formatVector() error = %v, want %v", errStr, tt.err.Error())
				}
			} else if result["error"] != nil {
				t.Errorf("formatVector() error = %v, want nil", result["error"])
			}
		})
	}
}

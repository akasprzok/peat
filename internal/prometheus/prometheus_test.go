package prometheus

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "http://localhost:9090",
			wantErr: false,
		},
		{
			name:    "valid URL with path",
			url:     "http://prometheus.example.com/api/v1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client for valid URL")
			}
		})
	}
}

func TestFormatQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "simple metric",
			query: "up",
			want:  "up",
		},
		{
			name:  "metric with labels",
			query: "up{job=\"prometheus\"}",
			want:  "up{job=\"prometheus\"}",
		},
		{
			name:  "sum aggregation",
			query: "sum(rate(http_requests_total[5m]))",
			want:  "sum(rate(http_requests_total[5m]))",
		},
		{
			name:  "invalid query returns original",
			query: "invalid{{{",
			want:  "invalid{{{",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatQuery(tt.query)
			if got != tt.want {
				t.Errorf("FormatQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

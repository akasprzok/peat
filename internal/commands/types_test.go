package commands

import "testing"

func TestQueryModeString(t *testing.T) {
	tests := []struct {
		name string
		mode QueryMode
		want string
	}{
		{"ModeInstant", ModeInstant, "/query"},
		{"ModeRange", ModeRange, "/query_range"},
		{"ModeSeries", ModeSeries, "/series"},
		{"ModeLabels", ModeLabels, "/labels"},
		{"Unknown mode", QueryMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("QueryMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

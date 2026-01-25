package charts

import (
	"strings"
	"testing"
)

func TestSeriesPaletteHasAtLeast10Colors(t *testing.T) {
	if len(SeriesPalette) < 10 {
		t.Errorf("SeriesPalette should have at least 10 colors, got %d", len(SeriesPalette))
	}
}

func TestSeriesColorCycles(t *testing.T) {
	paletteLen := len(SeriesPalette)

	// First cycle
	for i := 0; i < paletteLen; i++ {
		color := SeriesColor(i)
		if string(color) != SeriesPalette[i] {
			t.Errorf("SeriesColor(%d) = %s, want %s", i, color, SeriesPalette[i])
		}
	}

	// Second cycle (should wrap around)
	for i := 0; i < paletteLen; i++ {
		color := SeriesColor(i + paletteLen)
		if string(color) != SeriesPalette[i] {
			t.Errorf("SeriesColor(%d) = %s, want %s (cycling)", i+paletteLen, color, SeriesPalette[i])
		}
	}
}

func TestNoColorIsBlack(t *testing.T) {
	blackVariants := []string{
		"#000000",
		"#000",
		"0",
		"black",
	}

	for i, color := range SeriesPalette {
		colorLower := strings.ToLower(color)
		for _, black := range blackVariants {
			if colorLower == black {
				t.Errorf("SeriesPalette[%d] is black (%s), which would be invisible on dark backgrounds", i, color)
			}
		}
	}
}

func TestSeriesStyleReturnsValidStyle(t *testing.T) {
	style := SeriesStyle(0)
	// Verify that SeriesStyle returns a valid style
	// We check that GetForeground returns the expected color
	fg := style.GetForeground()
	expected := SeriesColor(0)
	if fg != expected {
		t.Errorf("SeriesStyle(0).GetForeground() = %v, want %v", fg, expected)
	}
}

func TestAxisAndLabelColorsAreDefined(t *testing.T) {
	if AxisColor == "" {
		t.Error("AxisColor should be defined")
	}
	if LabelColor == "" {
		t.Error("LabelColor should be defined")
	}
}

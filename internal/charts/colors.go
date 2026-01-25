package charts

import "github.com/charmbracelet/lipgloss"

// SeriesPalette is Paul Tol's qualitative color palette, designed for colorblind accessibility.
// See: https://personal.sron.nl/~pault/
var SeriesPalette = []string{
	"#4477AA", // Blue
	"#EE6677", // Rose
	"#228833", // Green
	"#CCBB44", // Olive/Yellow
	"#66CCEE", // Cyan
	"#AA3377", // Purple
	"#BBBBBB", // Grey
	"#EE8866", // Orange
	"#44BB99", // Teal
	"#FFAABB", // Pink
}

// AxisColor is the color used for chart axes.
var AxisColor = lipgloss.Color("#CCBB44") // Olive/Yellow - high visibility

// LabelColor is the color used for chart labels.
var LabelColor = lipgloss.Color("#66CCEE") // Cyan - good contrast

// SeriesColor returns the color for a given series index, cycling through the palette.
func SeriesColor(index int) lipgloss.Color {
	return lipgloss.Color(SeriesPalette[index%len(SeriesPalette)])
}

// SeriesStyle returns a lipgloss style with the foreground color for the given series index.
func SeriesStyle(index int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(SeriesColor(index))
}

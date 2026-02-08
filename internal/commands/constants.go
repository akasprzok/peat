package commands

import "time"

const (
	// DefaultTerminalWidth is the fallback terminal width when detection fails.
	DefaultTerminalWidth = 80

	// DefaultQueryStep is the default step interval for range queries.
	DefaultQueryStep = time.Minute

	// ChartWidthPadding is the horizontal padding subtracted from terminal width for chart rendering.
	ChartWidthPadding = 6

	// LegendMaxRows is the maximum number of visible rows in the legend table.
	LegendMaxRows = 5
)

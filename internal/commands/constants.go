package commands

import "time"

const (
	// DefaultTerminalWidth is the fallback terminal width when detection fails.
	DefaultTerminalWidth = 80

	// DefaultTerminalHeight is the fallback terminal height when detection fails.
	DefaultTerminalHeight = 24

	// DefaultQueryStep is the default step interval for range queries.
	DefaultQueryStep = time.Minute

	// ChartWidthPadding is the horizontal padding subtracted from terminal width for chart rendering.
	ChartWidthPadding = 6

	// LegendMaxRows is the maximum number of visible rows in the legend table.
	LegendMaxRows = 5

	// ChromeHeightExpanded is lines consumed by non-results chrome (input expanded).
	ChromeHeightExpanded = 12

	// ChromeHeightCollapsed is lines consumed by non-results chrome (input collapsed).
	ChromeHeightCollapsed = 10

	// LegendMinRows is the minimum number of legend rows to show.
	LegendMinRows = 3

	// LegendMaxRowsCap is the maximum legend rows regardless of terminal height.
	LegendMaxRowsCap = 10

	// LegendBorderLines is the legend border + margin overhead.
	LegendBorderLines = 3

	// ChartBorderLines is the chart border overhead.
	ChartBorderLines = 2
)

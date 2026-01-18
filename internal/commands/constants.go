package commands

import "time"

const (
	// DefaultTerminalWidth is the fallback terminal width when detection fails.
	DefaultTerminalWidth = 80

	// DefaultQueryStep is the default step interval for range queries.
	DefaultQueryStep = time.Minute
)

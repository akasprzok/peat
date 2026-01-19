package commands

import "time"

type Context struct {
	Timeout time.Duration
}

var Cli struct {
	Timeout time.Duration `help:"Timeout for Prometheus queries." short:"t" default:"60s"`

	TUI         TUICmd         `cmd:"" default:"1" help:"Interactive TUI mode."`
	Query       QueryCmd       `cmd:"" help:"Instant Query."`
	QueryRange  QueryRangeCmd  `cmd:"" help:"Range Query."`
	Series      SeriesCmd      `cmd:"" help:"Get series for matches."`
	FormatQuery FormatQueryCmd `cmd:"" help:"Format query."`
}

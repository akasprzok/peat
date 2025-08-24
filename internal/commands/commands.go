package commands

import "time"

type Context struct {
	Timeout time.Duration
}

var Cli struct {
	Timeout time.Duration `help:"Timeout for Prometheus queries." default:"60s"`

	Query       QueryCmd       `cmd:"" help:"Instant Query."`
	QueryRange  QueryRangeCmd  `cmd:"" help:"Range Query."`
	Series      SeriesCmd      `cmd:"" help:"Get series for matches."`
	FormatQuery FormatQueryCmd `cmd:"" help:"Format query."`
}

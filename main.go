package main

import (
	"os"

	"github.com/akasprzok/peat/internal/tui"
	"github.com/alecthomas/kong"
)

func main() {
	var cli tui.CLI
	kong.Parse(&cli,
		kong.Name("peat"),
		kong.Description("Terminal-native Prometheus metrics viewer with interactive visualizations."),
	)
	if err := cli.Run(); err != nil {
		os.Exit(1)
	}
}

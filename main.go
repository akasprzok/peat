package main

import (
	commands "github.com/akasprzok/peat/internal/commands"
	"github.com/alecthomas/kong"
)

func main() {
	// create new time series chart
	ctx := kong.Parse(&commands.Cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&commands.Context{Timeout: commands.Cli.Timeout})
	ctx.FatalIfErrorf(err)
}

package main

import (
	"fmt"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/alecthomas/kong"
)

type CLIOpts struct {
	Query struct {
		PrometheusURL string `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
		Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
	} `cmd:"" help:"Instant Query."`
}

func main() {
	// create new time series chart

	var cli CLIOpts
	ctx := kong.Parse(&cli)
	switch ctx.Command() {
	case "query <query>":
		prometheusClient := prometheus.NewClient(cli.Query.PrometheusURL)
		warnings, vector, err := prometheusClient.Query(cli.Query.Query)
		if err != nil {
			fmt.Printf("Error querying Prometheus: %v\n", err)
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		charter := charts.NewNtCharts()
		charter.PrintQuery(vector)
	default:
		fmt.Printf("Error executing command: %v\n", ctx.Command())
		panic(ctx.Command())
	}

}

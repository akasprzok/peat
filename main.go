package main

import (
	"fmt"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/alecthomas/kong"
)

type CLIOpts struct {
	Query struct {
		PrometheusURL string `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
		Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
	} `cmd:"" help:"Instant Query."`
	QueryRange struct {
		PrometheusURL string `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
		Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
		Range         string `arg:"" name:"range" help:"Range to query." required:"true"`
	} `cmd:"" help:"Range Query."`
}

func main() {
	// create new time series chart

	var cli CLIOpts
	ctx := kong.Parse(&cli)
	charter := charts.NewNtCharts()
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
		charter.PrintQuery(vector)
	case "query-range <query> <range>":
		prometheusClient := prometheus.NewClient(cli.QueryRange.PrometheusURL)
		end := time.Now()
		duration, err := time.ParseDuration(cli.QueryRange.Range)
		if err != nil {
			fmt.Printf("Error parsing range: %v\n", err)
		}
		start := end.Add(-duration)
		warnings, matrix, err := prometheusClient.QueryRange(cli.QueryRange.Query, start, end, 1*time.Minute)
		if err != nil {
			fmt.Printf("Error querying Prometheus: %v\n", err)
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		charter.PrintQueryRange(matrix)
	default:
		fmt.Printf("Error executing command: %v\n", ctx.Command())
		panic(ctx.Command())
	}

}

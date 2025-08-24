package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/alecthomas/kong"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

var cli struct {
	Timeout time.Duration `help:"Timeout for Prometheus queries." default:"60s"`

	Query       QueryCmd       `cmd:"" help:"Instant Query."`
	QueryRange  QueryRangeCmd  `cmd:"" help:"Range Query."`
	Series      SeriesCmd      `cmd:"" help:"Get series for matches."`
	FormatQuery FormatQueryCmd `cmd:"" help:"Format query."`
}

type Context struct {
	Timeout time.Duration
}

type QueryCmd struct {
	PrometheusURL string `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
	Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
	Output        string `name:"output" short:"o" help:"Output format." default:"graph" enum:"graph,json,yaml"`
}

func (q *QueryCmd) Run(ctx *Context) error {
	charter := charts.NewNtCharts()
	prometheusClient := prometheus.NewClient(cli.Query.PrometheusURL)
	warnings, vector, err := prometheusClient.Query(cli.Query.Query, ctx.Timeout)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	if len(vector) > 0 {
		switch q.Output {
		case "graph":
			charter.PrintQuery(vector)
		case "json":
			json, err := toJSON(vector)
			if err != nil {
				return err
			}
			fmt.Println(string(json))
		case "yaml":
			yaml, err := toYAML(vector)
			if err != nil {
				return err
			}
			fmt.Println(string(yaml))
		}
	} else {
		fmt.Println("No Data")
	}
	return nil
}

func massageVector(vector model.Vector) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, sample := range vector {
		data = append(data, map[string]interface{}{
			"metric":    sample.Metric,
			"value":     sample.Value,
			"timestamp": sample.Timestamp.Unix(),
		})
	}
	return data
}

func massageMatrix(matrix model.Matrix) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, sample := range matrix {
		values := make([]map[string]interface{}, 0)
		for _, value := range sample.Values {
			values = append(values, map[string]interface{}{
				"timestamp": value.Timestamp.Unix(),
				"value":     value.Value,
			})
		}
		data = append(data, map[string]interface{}{
			"metric": sample.Metric,
			"values": values,
		})
	}
	return data
}

func toJSON(vector model.Vector) ([]byte, error) {
	return json.MarshalIndent(massageVector(vector), "", "  ")
}

func toYAML(vector model.Vector) ([]byte, error) {
	return yaml.Marshal(massageVector(vector))
}

func matrixToJSON(matrix model.Matrix) ([]byte, error) {
	return json.MarshalIndent(massageMatrix(matrix), "", "  ")
}

func matrixToYAML(matrix model.Matrix) ([]byte, error) {
	return yaml.Marshal(massageMatrix(matrix))
}

type QueryRangeCmd struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
	Query         string        `arg:"" name:"query" help:"Query to run." required:"true"`
	Range         time.Duration `name:"range" help:"Range to query." default:"1h"`
	Output        string        `name:"output" short:"o" help:"Output format." default:"graph" enum:"graph,json,yaml"`
}

func (q *QueryRangeCmd) Run(ctx *Context) error {
	charter := charts.NewNtCharts()
	prometheusClient := prometheus.NewClient(q.PrometheusURL)
	end := time.Now()
	start := end.Add(-q.Range)
	matrix, warnings, err := prometheusClient.QueryRange(cli.QueryRange.Query, start, end, 1*time.Minute, ctx.Timeout)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	if len(matrix) > 0 {
		switch q.Output {
		case "graph":
			charter.PrintQueryRange(matrix)
		case "json":
			json, err := matrixToJSON(matrix)
			if err != nil {
				return err
			}
			fmt.Println(string(json))
		case "yaml":
			yaml, err := matrixToYAML(matrix)
			if err != nil {
				return err
			}
			fmt.Println(string(yaml))
		}
	} else {
		fmt.Println("No Data")
	}
	return nil
}

type SeriesCmd struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
	Match         string        `arg:"" name:"match" help:"Matches to query." required:"true"`
	Range         time.Duration `name:"range" help:"Range to query." default:"1h"`
	Limit         uint64        `arg:"" name:"limit" help:"Limit the number of series." default:"100"`
	Output        string        `name:"output" short:"o" help:"Output format." default:"json" enum:"json,yaml"`
}

func (s *SeriesCmd) Run(ctx *Context) error {
	prometheusClient := prometheus.NewClient(s.PrometheusURL)
	end := time.Now()
	start := end.Add(-s.Range)
	series, warnings, err := prometheusClient.Series(cli.Series.Match, start, end, cli.Series.Limit, ctx.Timeout)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		return err
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	if len(series) > 0 {
		switch s.Output {
		case "json":
			json, err := json.MarshalIndent(series, "", "  ")
			if err != nil {
				fmt.Printf("Error marshalling series: %v\n", err)
			}
			fmt.Println(string(json))
		case "yaml":
			yaml, err := yaml.Marshal(series)
			if err != nil {
				fmt.Printf("Error marshalling series: %v\n", err)
			}
			fmt.Println(string(yaml))
		}
	} else {
		fmt.Println("No Data")
	}
	return nil
}

type FormatQueryCmd struct {
	Query string `arg:"" name:"query" help:"Query to format." required:"true"`
}

func (f *FormatQueryCmd) Run(_ *Context) error {
	fmt.Println(prometheus.FormatQuery(f.Query))
	return nil
}

func main() {
	// create new time series chart
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{Timeout: cli.Timeout})
	ctx.FatalIfErrorf(err)
}

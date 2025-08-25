package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type QueryRangeCmd struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
	Query         string        `arg:"" name:"query" help:"Query to run." required:"true"`
	Range         time.Duration `name:"range" short:"r" help:"Range to query." default:"1h"`
	Output        string        `name:"output" short:"o" help:"Output format." default:"graph" enum:"graph,json,yaml"`
}

func (q *QueryRangeCmd) Run(ctx *Context) error {
	prometheusClient := prometheus.NewClient(q.PrometheusURL)
	end := time.Now()
	start := end.Add(-q.Range)
	matrix, warnings, err := prometheusClient.QueryRange(q.Query, start, end, 1*time.Minute, ctx.Timeout)
	switch q.Output {
	case "graph":
		if err != nil {
			return err
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		charter := charts.NewNtCharts()
		charter.PrintQueryRange(matrix)
	case "json":
		output := formatMatrix(matrix, warnings, err)
		json, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(json))
	case "yaml":
		output := formatMatrix(matrix, warnings, err)
		yaml, err := yaml.Marshal(output)
		if err != nil {
			return err
		}
		fmt.Println(string(yaml))
	}
	return nil
}

func formatMatrix(matrix model.Matrix, warnings []string, err error) map[string]interface{} {
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

	if err != nil {
		return map[string]interface{}{
			"data":     data,
			"warnings": warnings,
			"error":    err.Error(),
		}
	}
	return map[string]interface{}{
		"data":     data,
		"warnings": warnings,
		"error":    nil,
	}
}

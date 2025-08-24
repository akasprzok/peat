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
	Range         time.Duration `name:"range" help:"Range to query." default:"1h"`
	Output        string        `name:"output" short:"o" help:"Output format." default:"graph" enum:"graph,json,yaml"`
}

func (q *QueryRangeCmd) Run(ctx *Context) error {
	charter := charts.NewNtCharts()
	prometheusClient := prometheus.NewClient(q.PrometheusURL)
	end := time.Now()
	start := end.Add(-q.Range)
	matrix, warnings, err := prometheusClient.QueryRange(q.Query, start, end, 1*time.Minute, ctx.Timeout)
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

func matrixToJSON(matrix model.Matrix) ([]byte, error) {
	return json.MarshalIndent(massageMatrix(matrix), "", "  ")
}

func matrixToYAML(matrix model.Matrix) ([]byte, error) {
	return yaml.Marshal(massageMatrix(matrix))
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

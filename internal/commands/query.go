package commands

import (
	"encoding/json"
	"fmt"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type QueryCmd struct {
	PrometheusURL string `help:"URL of the Prometheus endpoint." env:"PROMETHEUS_URL" name:"prometheus-url"`
	Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
	Output        string `name:"output" short:"o" help:"Output format." default:"graph" enum:"graph,json,yaml"`
}

func (q *QueryCmd) Run(ctx *Context) error {
	charter := charts.NewNtCharts()
	prometheusClient := prometheus.NewClient(q.PrometheusURL)
	warnings, vector, err := prometheusClient.Query(q.Query, ctx.Timeout)
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

func toJSON(vector model.Vector) ([]byte, error) {
	return json.MarshalIndent(massageVector(vector), "", "  ")
}

func toYAML(vector model.Vector) ([]byte, error) {
	return yaml.Marshal(massageVector(vector))
}

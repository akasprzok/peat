package commands

import (
	"encoding/json"
	"fmt"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/akasprzok/peat/internal/tables"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type QueryCmd struct {
	PrometheusURL string `help:"URL of the Prometheus endpoint." short:"p" env:"PEAT_PROMETHEUS_URL" name:"prometheus-url"`
	Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
	Output        string `name:"output" short:"o" help:"Output format. Choices are: graph, table, json, yaml" default:"graph" enum:"graph,table,json,yaml"`
}

func (q *QueryCmd) Run(ctx *Context) error {
	prometheusClient := prometheus.NewClient(q.PrometheusURL)
	warnings, vector, err := prometheusClient.Query(q.Query, ctx.Timeout)
	switch q.Output {
	case "graph":
		if err != nil {
			return err
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		charter := charts.NewNtCharts()
		charter.PrintQuery(vector)
	case "table":
		if err != nil {
			return err
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		table, err := tables.PrintQuery(vector)
		if err != nil {
			return err
		}
		p := tea.NewProgram(table)
		_, err = p.Run()
		if err != nil {
			return err
		}
		fmt.Println(table.View())
	case "json":
		output := formatVector(vector, warnings, err)
		json, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(json))
	case "yaml":
		output := formatVector(vector, warnings, err)
		yaml, err := yaml.Marshal(output)
		if err != nil {
			return err
		}
		fmt.Println(string(yaml))
	}
	return nil
}

func formatVector(vector model.Vector, warnings []string, err error) map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, sample := range vector {
		data = append(data, map[string]interface{}{
			"metric":    sample.Metric,
			"value":     sample.Value,
			"timestamp": sample.Timestamp.Unix(),
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

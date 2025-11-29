package commands

import (
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	tea "github.com/charmbracelet/bubbletea"
)

type QueryCmd struct {
	PrometheusURL string `help:"URL of the Prometheus endpoint." short:"p" env:"PEAT_PROMETHEUS_URL" name:"prometheus-url"`
	Query         string `arg:"" name:"query" help:"Query to run." required:"true"`
	Output        string `name:"output" short:"o" help:"Output format. Choices are: graph, table, json, yaml" default:"graph" enum:"graph,table,json,yaml"`
}

func (q *QueryCmd) Run(ctx *Context) error {
	queryModel := NewQueryModel(q.PrometheusURL, q.Query, q.Output, ctx.Timeout)

	p := tea.NewProgram(queryModel)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Extract the final model to check for errors
	if qm, ok := finalModel.(QueryModel); ok {
		if qm.err != nil {
			return qm.err
		}
	}

	return nil
}

func formatVector(vector model.Vector, warnings v1.Warnings, err error) map[string]any {
	data := make([]map[string]any, 0)
	for _, sample := range vector {
		data = append(data, map[string]any{
			"metric":    sample.Metric,
			"value":     sample.Value,
			"timestamp": sample.Timestamp.Unix(),
		})
	}

	if err != nil {
		return map[string]any{
			"data":     data,
			"warnings": warnings,
			"error":    err.Error(),
		}
	}
	return map[string]any{
		"data":     data,
		"warnings": warnings,
		"error":    nil,
	}
}

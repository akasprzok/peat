package commands

import (
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	tea "github.com/charmbracelet/bubbletea"
)

type QueryRangeCmd struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." env:"PEAT_PROMETHEUS_URL" name:"prometheus-url"`
	Query         string        `arg:"" name:"query" help:"Query to run." required:"true"`
	Range         time.Duration `name:"range" short:"r" help:"Range to query." default:"1h"`
	Output        string        `name:"output" short:"o" help:"Output format." default:"graph" enum:"graph,json,yaml"`
}

func (q *QueryRangeCmd) Run(ctx *Context) error {
	queryRangeModel := NewQueryRangeModel(q.PrometheusURL, q.Query, q.Range, q.Output, ctx.Timeout)

	p := tea.NewProgram(queryRangeModel)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Extract the final model to check for errors
	if qrm, ok := finalModel.(QueryRangeModel); ok {
		if qrm.err != nil {
			return qrm.err
		}
	}

	return nil
}

func formatMatrix(matrix model.Matrix, warnings v1.Warnings, err error) map[string]any {
	data := make([]map[string]any, 0)
	for _, sample := range matrix {
		values := make([]map[string]any, 0)
		for _, value := range sample.Values {
			values = append(values, map[string]any{
				"timestamp": value.Timestamp.Unix(),
				"value":     value.Value,
			})
		}
		data = append(data, map[string]any{
			"metric": sample.Metric,
			"values": values,
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

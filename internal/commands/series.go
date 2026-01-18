package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/akasprzok/peat/internal/prometheus"
	"gopkg.in/yaml.v2"
)

type SeriesCmd struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." env:"PEAT_PROMETHEUS_URL" name:"prometheus-url"`
	Match         string        `arg:"" name:"match" help:"Matches to query." required:"true"`
	Range         time.Duration `name:"range" help:"Range to query." default:"1h"`
	Limit         uint64        `arg:"" name:"limit" help:"Limit the number of series." default:"100"`
	Output        string        `name:"output" short:"o" help:"Output format." default:"json" enum:"json,yaml"`
}

func (s *SeriesCmd) Run(ctx *Context) error {
	prometheusClient, err := prometheus.NewClient(s.PrometheusURL)
	if err != nil {
		return err
	}
	end := time.Now()
	start := end.Add(-s.Range)
	series, warnings, err := prometheusClient.Series(s.Match, start, end, s.Limit, ctx.Timeout)
	if err != nil {
		return err
	}
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Warnings: %v\n", warnings)
	}
	if len(series) == 0 {
		fmt.Println("No Data")
		return nil
	}

	switch s.Output {
	case "json":
		jsonBytes, err := json.MarshalIndent(series, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling series to JSON: %w", err)
		}
		fmt.Println(string(jsonBytes))
	case "yaml":
		yamlBytes, err := yaml.Marshal(series)
		if err != nil {
			return fmt.Errorf("marshalling series to YAML: %w", err)
		}
		fmt.Println(string(yamlBytes))
	}
	return nil
}

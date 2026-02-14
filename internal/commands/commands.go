package commands

import (
	"time"

	"github.com/akasprzok/peat/internal/prometheus"
	tea "github.com/charmbracelet/bubbletea"
)

// CLI represents the command-line interface for Peat.
type CLI struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." short:"p" env:"PEAT_PROMETHEUS_URL" name:"prometheus-url"`
	Timeout       time.Duration `help:"Timeout for Prometheus queries." short:"t" default:"60s"`
	Range         time.Duration `name:"range" short:"r" help:"Initial range for range queries." default:"1h"`
	Step          time.Duration `name:"step" short:"s" help:"Initial step interval for range queries." default:"1m"`
	Limit         uint64        `name:"limit" short:"l" help:"Maximum number of series to return for series queries." default:"100"`
}

// Run starts the interactive TUI.
func (c *CLI) Run() error {
	client, err := prometheus.NewClient(c.PrometheusURL)
	if err != nil {
		return err
	}

	model := NewTUIModel(client, c.Range, c.Step, c.Limit, c.Timeout)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err = p.Run()
	return err
}

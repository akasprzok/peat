package commands

import (
	"time"

	"github.com/akasprzok/peat/internal/prometheus"
	tea "github.com/charmbracelet/bubbletea"
)

// TUICmd is the Kong command for the interactive TUI mode.
type TUICmd struct {
	PrometheusURL string        `help:"URL of the Prometheus endpoint." short:"p" env:"PEAT_PROMETHEUS_URL" name:"prometheus-url"`
	Range         time.Duration `name:"range" short:"r" help:"Initial range for range queries." default:"1h"`
	Step          time.Duration `name:"step" short:"s" help:"Initial step interval for range queries." default:"1m"`
}

// Run starts the interactive TUI.
func (t *TUICmd) Run(ctx *Context) error {
	client, err := prometheus.NewClient(t.PrometheusURL)
	if err != nil {
		return err
	}

	model := NewTUIModel(client, t.Range, t.Step, ctx.Timeout)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	return err
}

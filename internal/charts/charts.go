package charts

import (
	"fmt"
	"os"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/common/model"
	"golang.org/x/term"
)

type Charter interface {
	PrintQuery(model.Vector)
}

type ntCharts struct {
}

func NewNtCharts() Charter {
	return &ntCharts{}
}

func (c *ntCharts) PrintQuery(vector model.Vector) {
	barData := make([]barchart.BarData, 0)
	for _, sample := range vector {
		barData = append(barData, barchart.BarData{
			Label: fmt.Sprintf("%s (%d)", sample.Metric.String(), int(sample.Value)),
			Values: []barchart.BarValue{
				{Name: sample.Metric.String(), Value: float64(sample.Value), Style: lipgloss.NewStyle().Foreground(lipgloss.Color("10"))}, // green
			},
		})
	}

	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Error getting terminal size: %v\n", err)
		return
	}

	bc := barchart.New(width, len(barData)*2, barchart.WithDataSet(barData), barchart.WithHorizontalBars())
	bc.Draw()

	fmt.Println(bc.View())
}

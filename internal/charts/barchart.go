package charts

import (
	"fmt"
	"strconv"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/common/model"
)

func Barchart(vector model.Vector, width int) string {

	barData := make([]barchart.BarData, 0)
	for i, sample := range vector {
		barData = append(barData, barchart.BarData{
			Label: fmt.Sprintf("%s (%d)", sample.Metric.String(), int(sample.Value)),
			Values: []barchart.BarValue{
				{Name: sample.Metric.String(), Value: float64(sample.Value), Style: lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(i)))}, // green
			},
		})
	}

	bc := barchart.New(width, len(barData)*2, barchart.WithDataSet(barData), barchart.WithHorizontalBars())
	bc.Draw()

	return bc.View()

}

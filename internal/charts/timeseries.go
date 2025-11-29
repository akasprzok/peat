package charts

import (
	"math"
	"strconv"

	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/common/model"
)

var lineStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("4")) // blue

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // yellow

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6")) // cyan

// LegendEntry represents a single entry in the time series legend
type LegendEntry struct {
	Metric     string
	ColorIndex int
}

// TimeseriesSplit returns the chart and legend entries separately
func TimeseriesSplit(matrix model.Matrix, width int) (chart string, legend []LegendEntry) {
	minYValue := model.SampleValue(math.MaxFloat64)
	maxYValue := model.SampleValue(-math.MaxFloat64)
	for _, stream := range matrix {
		for _, sample := range stream.Values {
			if sample.Value < minYValue {
				minYValue = sample.Value
			}
			if sample.Value > maxYValue {
				maxYValue = sample.Value
			}
		}
	}

	height := width / 8

	legendEntries := make([]LegendEntry, 0, len(matrix))

	lc := timeserieslinechart.New(width, height)
	lc.AxisStyle = axisStyle
	lc.LabelStyle = labelStyle
	lc.XLabelFormatter = timeserieslinechart.HourTimeLabelFormatter()
	lc.SetYRange(float64(minYValue), float64(maxYValue))     // set expected Y values (values can be less or greater than what is displayed)
	lc.SetViewYRange(float64(minYValue), float64(maxYValue)) // setting display Y values will fail unless set expected Y values first
	lc.SetStyle(lineStyle)
	lc.SetLineStyle(runes.ThinLineStyle) // ThinLineStyle replaces default linechart arcline rune style

	for i, stream := range matrix {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(i)))
		// Add to legend entries
		legendEntries = append(legendEntries, LegendEntry{
			Metric:     stream.Metric.String(),
			ColorIndex: i,
		})
		lc.SetDataSetStyle(stream.Metric.String(), style)
		for _, sample := range stream.Values {
			point := timeserieslinechart.TimePoint{
				Time:  sample.Timestamp.Time(),
				Value: float64(sample.Value),
			}
			lc.PushDataSet(stream.Metric.String(), point)
		}
	}

	lc.DrawBrailleAll()

	return lc.View(), legendEntries
}

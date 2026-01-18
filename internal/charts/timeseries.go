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
	return TimeseriesSplitWithSelection(matrix, width, -1)
}

// TimeseriesSplitWithSelection returns the chart and legend entries with a selected series highlighted
// selectedIndex: -1 means no selection, all series shown normally
func TimeseriesSplitWithSelection(matrix model.Matrix, width int, selectedIndex int) (chart string, legend []LegendEntry) {
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

	height := width / ChartHeightRatio

	legendEntries := make([]LegendEntry, 0, len(matrix))

	lc := timeserieslinechart.New(width, height)
	lc.AxisStyle = axisStyle
	lc.LabelStyle = labelStyle
	lc.XLabelFormatter = timeserieslinechart.HourTimeLabelFormatter()
	lc.SetYRange(float64(minYValue), float64(maxYValue))     // set expected Y values (values can be less or greater than what is displayed)
	lc.SetViewYRange(float64(minYValue), float64(maxYValue)) // setting display Y values will fail unless set expected Y values first
	lc.SetStyle(lineStyle)
	lc.SetLineStyle(runes.ThinLineStyle) // ThinLineStyle replaces default linechart arcline rune style

	// Build legend entries and draw all series in their natural order
	for i, stream := range matrix {
		legendEntries = append(legendEntries, LegendEntry{
			Metric:     stream.Metric.String(),
			ColorIndex: i,
		})

		var style lipgloss.Style
		if selectedIndex == -1 {
			// No selection, show all in their original colors
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(i)))
		} else if i == selectedIndex {
			// Selected series will be drawn separately below for layering
			continue
		} else {
			// Non-selected series are greyed out
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		}

		lc.SetDataSetStyle(stream.Metric.String(), style)
		for _, sample := range stream.Values {
			point := timeserieslinechart.TimePoint{
				Time:  sample.Timestamp.Time(),
				Value: float64(sample.Value),
			}
			lc.PushDataSet(stream.Metric.String(), point)
		}
	}

	// If a series is selected, draw it a second time to ensure it appears on top
	if selectedIndex >= 0 && selectedIndex < len(matrix) {
		stream := matrix[selectedIndex]
		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")) // bright white for high contrast

		// Use a temporary unique name to draw it again
		tempName := stream.Metric.String() + "_selected"
		lc.SetDataSetStyle(tempName, selectedStyle)
		for _, sample := range stream.Values {
			point := timeserieslinechart.TimePoint{
				Time:  sample.Timestamp.Time(),
				Value: float64(sample.Value),
			}
			lc.PushDataSet(tempName, point)
		}
	}

	lc.DrawBrailleAll()

	return lc.View(), legendEntries
}

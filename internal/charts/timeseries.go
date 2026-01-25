package charts

import (
	"math"

	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/common/model"
)

var lineStyle = lipgloss.NewStyle().
	Foreground(SeriesColor(0)) // blue from accessible palette

var axisStyle = lipgloss.NewStyle().
	Foreground(AxisColor)

var labelStyle = lipgloss.NewStyle().
	Foreground(LabelColor)

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
	for i, stream := range matrix {
		// Skip non-selected series when calculating range
		if selectedIndex >= 0 && i != selectedIndex {
			continue
		}
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
		switch {
		case selectedIndex == -1:
			// No selection, show all in their original colors
			style = SeriesStyle(i)
		case i == selectedIndex:
			// Selected series will be drawn separately below for layering
			continue
		default:
			// Non-selected series are hidden
			continue
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

	// If a series is selected, draw it (non-selected series were skipped above)
	if selectedIndex >= 0 && selectedIndex < len(matrix) {
		stream := matrix[selectedIndex]
		style := SeriesStyle(selectedIndex) // Use original color

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

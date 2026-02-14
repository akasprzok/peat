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
func TimeseriesSplit(matrix model.Matrix, width, height int) (chart string, legend []LegendEntry) {
	return TimeseriesSplitWithSelection(matrix, width, height, -1, nil)
}

// isSeriesVisible returns whether a series at index i should be rendered.
func isSeriesVisible(i int, selectedIndex int, highlightedIndices map[int]bool) bool {
	if selectedIndex == -1 {
		return true // No selection: show all series
	}
	if i == selectedIndex {
		return true
	}
	return highlightedIndices[i]
}

// TimeseriesSplitWithSelection returns the chart and legend entries with a selected series highlighted.
// height: explicit chart height; when <= 0, falls back to width/ChartHeightRatio.
// selectedIndex: -1 means no selection, all series shown normally.
// highlightedIndices: pinned series that remain visible alongside the selected series.
func TimeseriesSplitWithSelection(matrix model.Matrix, width, height int, selectedIndex int, highlightedIndices map[int]bool) (chart string, legend []LegendEntry) {
	minYValue := model.SampleValue(math.MaxFloat64)
	maxYValue := model.SampleValue(-math.MaxFloat64)
	for i, stream := range matrix {
		if !isSeriesVisible(i, selectedIndex, highlightedIndices) {
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

	if height <= 0 {
		height = width / ChartHeightRatio
	}
	if height < MinChartHeight {
		height = MinChartHeight
	}

	legendEntries := make([]LegendEntry, 0, len(matrix))

	lc := timeserieslinechart.New(width, height)
	lc.AxisStyle = axisStyle
	lc.LabelStyle = labelStyle
	lc.XLabelFormatter = timeserieslinechart.HourTimeLabelFormatter()
	lc.SetYRange(float64(minYValue), float64(maxYValue))     // set expected Y values (values can be less or greater than what is displayed)
	lc.SetViewYRange(float64(minYValue), float64(maxYValue)) // setting display Y values will fail unless set expected Y values first
	lc.SetStyle(lineStyle)
	lc.SetLineStyle(runes.ThinLineStyle) // ThinLineStyle replaces default linechart arcline rune style

	// Build legend entries and draw visible series (except selected, which is drawn last for layering)
	for i, stream := range matrix {
		legendEntries = append(legendEntries, LegendEntry{
			Metric:     stream.Metric.String(),
			ColorIndex: i,
		})

		// Skip the selected series here; it will be drawn last for layering emphasis
		if selectedIndex >= 0 && i == selectedIndex {
			continue
		}

		if !isSeriesVisible(i, selectedIndex, highlightedIndices) {
			continue
		}

		style := SeriesStyle(i)
		lc.SetDataSetStyle(stream.Metric.String(), style)
		for _, sample := range stream.Values {
			point := timeserieslinechart.TimePoint{
				Time:  sample.Timestamp.Time(),
				Value: float64(sample.Value),
			}
			lc.PushDataSet(stream.Metric.String(), point)
		}
	}

	// Draw the selected series last for layering emphasis
	if selectedIndex >= 0 && selectedIndex < len(matrix) {
		stream := matrix[selectedIndex]
		style := SeriesStyle(selectedIndex)

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

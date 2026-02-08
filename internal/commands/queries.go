package commands

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m TUIModel) executeQuery() (tea.Model, tea.Cmd) {
	// Save query for this mode
	m.modeQueries[m.mode] = m.queryInput.Value()
	m.modeStates[m.mode] = StateLoading
	m.modeErrors[m.mode] = nil
	m.modeWarnings[m.mode] = nil
	m.queryInput.Blur()

	cmd := m.currentMode().ExecuteQuery(&m)
	return m, cmd
}

func (m TUIModel) executeInstantQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		start := time.Now()
		warnings, vector, err := m.promClient.Query(query, m.timeout)
		duration := time.Since(start)
		return tuiInstantResultMsg{
			warnings: warnings,
			vector:   vector,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeRangeQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		matrix, warnings, err := m.promClient.QueryRange(query, rangeStart, end, m.stepValue, m.timeout)
		duration := time.Since(start)
		return tuiRangeResultMsg{
			warnings: warnings,
			matrix:   matrix,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeSeriesQuery() tea.Cmd {
	query := m.queryInput.Value()
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		series, warnings, err := m.promClient.Series(query, rangeStart, end, m.seriesLimit, m.timeout)
		duration := time.Since(start)
		return tuiSeriesResultMsg{
			warnings: warnings,
			series:   series,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeLabelsQuery() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		labels, warnings, err := m.promClient.LabelNames(rangeStart, end, m.timeout)
		duration := time.Since(start)
		return tuiLabelsResultMsg{
			warnings: warnings,
			labels:   labels,
			err:      err,
			duration: duration,
		}
	}
}

func (m TUIModel) executeLabelValuesQuery(labelName string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		end := start
		rangeStart := end.Add(-m.rangeValue)
		values, warnings, err := m.promClient.LabelValues(labelName, rangeStart, end, m.timeout)
		duration := time.Since(start)
		return tuiLabelValuesResultMsg{
			labelName: labelName,
			warnings:  warnings,
			values:    values,
			err:       err,
			duration:  duration,
		}
	}
}

func (m TUIModel) handleInstantResult(msg tuiInstantResultMsg) (tea.Model, tea.Cmd) {
	m = m.applyResultCommon(ModeInstant, msg.warnings, msg.err, msg.duration)
	m.vector = msg.vector

	if msg.err != nil {
		m.modeStates[ModeInstant] = StateError
		return m, nil
	}

	m.modeStates[ModeInstant] = StateResults
	m = m.renderInstantChart()
	return m, nil
}

func (m TUIModel) handleRangeResult(msg tuiRangeResultMsg) (tea.Model, tea.Cmd) {
	m = m.applyResultCommon(ModeRange, msg.warnings, msg.err, msg.duration)
	m.matrix = msg.matrix

	if msg.err != nil {
		m.modeStates[ModeRange] = StateError
		return m, nil
	}

	m.modeStates[ModeRange] = StateResults
	m.selectedIndex = -1
	m.highlightedIndices = make(map[int]bool)
	m = m.renderRangeChart()
	return m, nil
}

func (m TUIModel) handleSeriesResult(msg tuiSeriesResultMsg) (tea.Model, tea.Cmd) {
	m = m.applyResultCommon(ModeSeries, msg.warnings, msg.err, msg.duration)
	m.series = msg.series

	if msg.err != nil {
		m.modeStates[ModeSeries] = StateError
		return m, nil
	}

	m.modeStates[ModeSeries] = StateResults
	m = m.renderSeriesTable()
	return m, nil
}

func (m TUIModel) handleLabelsResult(msg tuiLabelsResultMsg) (tea.Model, tea.Cmd) {
	m = m.applyResultCommon(ModeLabels, msg.warnings, msg.err, msg.duration)
	m.labels = msg.labels
	m.viewingLabelValues = false

	if msg.err != nil {
		m.modeStates[ModeLabels] = StateError
		return m, nil
	}

	m.modeStates[ModeLabels] = StateResults
	m = m.renderLabelsTable()
	return m, nil
}

func (m TUIModel) handleLabelValuesResult(msg tuiLabelValuesResultMsg) (tea.Model, tea.Cmd) {
	m.modeWarnings[ModeLabels] = msg.warnings
	m.labelValues = msg.values
	m.selectedLabelName = msg.labelName
	m.modeErrors[ModeLabels] = msg.err
	m.modeDurations[ModeLabels] = msg.duration

	if msg.err != nil {
		m.modeStates[ModeLabels] = StateError
		m.viewingLabelValues = false
		return m, nil
	}

	m.modeStates[ModeLabels] = StateResults
	m.viewingLabelValues = true
	m = m.renderLabelsTable()
	return m, nil
}

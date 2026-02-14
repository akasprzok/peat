package commands

import (
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.queryInput.Width = msg.Width - 10
		m.resultsViewport.Width = msg.Width
		m.resultsViewport.Height = m.getAvailableResultsHeight()
		if m.currentState() == StateResults {
			m.currentMode().OnSwitchTo(&m)
			m = m.syncViewportContent()
		}
		return m, nil

	case tea.MouseMsg:
		if !m.insertMode && m.currentState() == StateResults {
			var cmd tea.Cmd
			m.resultsViewport, cmd = m.resultsViewport.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tuiInstantResultMsg:
		return m.handleInstantResult(msg)

	case tuiRangeResultMsg:
		return m.handleRangeResult(msg)

	case tuiSeriesResultMsg:
		return m.handleSeriesResult(msg)

	case tuiLabelsResultMsg:
		return m.handleLabelsResult(msg)

	case tuiLabelValuesResultMsg:
		return m.handleLabelValuesResult(msg)

	case spinner.TickMsg:
		if m.currentState() == StateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Update text input if focused
	if m.focusedPane == PaneQuery && m.currentState() != StateLoading {
		var cmd tea.Cmd
		m.queryInput, cmd = m.queryInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m TUIModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle shortcuts overlay - dismiss on any key except quit keys
	if m.showShortcutsOverlay {
		if msg.String() == "q" {
			return m, tea.Quit
		}
		m.showShortcutsOverlay = false
		return m, nil
	}

	// State-dependent keys
	switch m.currentState() {
	case StateLoading:
		// Only allow quit during loading
		return m, nil

	case StateInput, StateResults, StateError:
		return m.handleInputOrResultsKey(msg)
	}

	return m, nil
}

func (m TUIModel) handleInputOrResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle legend navigation when focused
	if m.focusedPane == PaneLegend && m.legendFocused {
		return m.handleLegendKey(msg)
	}

	// INSERT MODE: Route most keys to text input
	if m.insertMode {
		switch msg.String() {
		case "esc":
			// Exit insert mode
			m.insertMode = false
			m.queryInput.Blur()
			return m, nil
		case "enter":
			// Execute and exit insert mode
			m.insertMode = false
			m.queryInput.Blur()
			return m.handleEnterKey()
		case "tab":
			// Allow mode switching even in insert mode
			return m.handleTabKey()
		default:
			// All other keys go to text input
			var cmd tea.Cmd
			m.queryInput, cmd = m.queryInput.Update(msg)
			return m, cmd
		}
	}

	// NORMAL MODE: Handle shortcuts
	return m.handleNormalModeKey(msg)
}

func (m TUIModel) handleNormalModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "tab":
		return m.handleTabKey()
	case "enter":
		return m.handleEnterKey()
	case "/":
		return m.enterInsertMode()
	case "i":
		return m.handleInteractiveKey()
	case "f":
		return m.handleFormatKey()
	case "esc":
		return m.handleEscapeKey()
	case "1", "2", "3", "4":
		return m.handleNumberKey(msg.String())
	case "?":
		m.showShortcutsOverlay = true
		return m, nil
	case "ctrl+d", "ctrl+u":
		if m.currentState() == StateResults {
			var cmd tea.Cmd
			m.resultsViewport, cmd = m.resultsViewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m TUIModel) handleTabKey() (tea.Model, tea.Cmd) {
	// Cycle through modes: Instant -> Range -> Series -> Labels -> Instant
	nextMode := m.mode + 1
	if nextMode > ModeLabels {
		nextMode = ModeInstant
	}
	return m.switchToMode(nextMode)
}

func (m TUIModel) handleNumberKey(key string) (tea.Model, tea.Cmd) {
	modeMap := map[string]QueryMode{
		"1": ModeInstant,
		"2": ModeRange,
		"3": ModeSeries,
		"4": ModeLabels,
	}
	if mode, ok := modeMap[key]; ok {
		return m.switchToMode(mode)
	}
	return m, nil
}

func (m TUIModel) switchToMode(newMode QueryMode) (tea.Model, tea.Cmd) {
	if newMode == m.mode {
		return m, nil
	}

	// Save current query before switching
	m.modeQueries[m.mode] = m.queryInput.Value()

	// Switch to new mode
	m.mode = newMode

	// Restore new mode's query
	m.queryInput.SetValue(m.modeQueries[m.mode])

	// Clear any selection and highlights
	m.selectedIndex = -1
	m.legendFocused = false
	m.highlightedIndices = make(map[int]bool)

	// Delegate to mode's OnSwitchTo for re-rendering if needed
	m.currentMode().OnSwitchTo(&m)

	return m, nil
}

func (m TUIModel) handleEnterKey() (tea.Model, tea.Cmd) {
	// Execute query
	if m.queryInput.Value() != "" {
		return m.executeQuery()
	}
	return m, nil
}

func (m TUIModel) enterInsertMode() (tea.Model, tea.Cmd) {
	m.insertMode = true
	m.inputCollapsed = false
	m.focusedPane = PaneQuery
	m.queryInput.Focus()
	m.resultsViewport.Height = m.getAvailableResultsHeight()
	return m, nil
}

func (m TUIModel) handleInteractiveKey() (tea.Model, tea.Cmd) {
	// Toggle interactive mode - delegate to current mode
	if m.currentState() != StateResults {
		return m, nil
	}

	cmd := m.currentMode().HandleInteractiveToggle(&m)
	return m, cmd
}

func (m TUIModel) handleEscapeKey() (tea.Model, tea.Cmd) {
	// Exit interactive/legend mode but stay in normal mode
	m.focusedPane = PaneQuery
	m.legendFocused = false
	// Don't focus query - stay in normal mode
	if m.mode == ModeRange && len(m.legendEntries) > 0 {
		m.legendTable = m.legendTable.Focused(false)
		m.selectedIndex = -1
		m = m.regenerateRangeChart()
		m = m.syncViewportContent()
	}
	return m, nil
}

func (m TUIModel) handleFormatKey() (tea.Model, tea.Cmd) {
	currentQuery := m.queryInput.Value()
	if currentQuery == "" {
		return m, nil // Nothing to format
	}
	formatted := prometheus.FormatQuery(currentQuery)
	m.queryInput.SetValue(formatted)
	return m, nil
}

func (m TUIModel) handleLegendKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cmd := m.currentMode().HandleLegendKey(&m, msg)
	return m, cmd
}

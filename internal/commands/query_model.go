package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akasprzok/peat/internal/charts"
	"github.com/akasprzok/peat/internal/prometheus"
	"github.com/akasprzok/peat/internal/tables"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

type queryState int

const (
	stateLoading queryState = iota
	stateSuccess
	stateError
	stateShowingTable
)

type queryResultMsg struct {
	warnings v1.Warnings
	vector   model.Vector
	err      error
}

type QueryModel struct {
	promClient   prometheus.Client
	query        string
	timeout      time.Duration
	output       string
	state        queryState
	spinner      spinner.Model
	warnings     v1.Warnings
	vector       model.Vector
	err          error
	width        int
	height       int
	tableModel   *tables.Model
	chartContent string
	quitting     bool
}

func NewQueryModel(promURL, query, output string, timeout time.Duration) QueryModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return QueryModel{
		promClient: prometheus.NewClient(promURL),
		query:      query,
		timeout:    timeout,
		output:     output,
		state:      stateLoading,
		spinner:    s,
	}
}

func (m QueryModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.executeQuery(),
	)
}

func (m QueryModel) executeQuery() tea.Cmd {
	return func() tea.Msg {
		warnings, vector, err := m.promClient.Query(m.query, m.timeout)
		return queryResultMsg{
			warnings: warnings,
			vector:   vector,
			err:      err,
		}
	}
}

func (m QueryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg), nil
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case queryResultMsg:
		return m.handleQueryResult(msg)
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	}

	// Update table if we're showing it
	if m.state == stateShowingTable && m.tableModel != nil {
		return m.updateTableModel(msg)
	}

	return m, nil
}

func (m QueryModel) handleWindowSize(msg tea.WindowSizeMsg) QueryModel {
	m.width = msg.Width
	m.height = msg.Height
	return m
}

func (m QueryModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == stateShowingTable && m.tableModel != nil {
		// Let the table handle the key if we're in table mode
		if msg.String() != "q" && msg.String() != "ctrl+c" {
			updated, cmd := m.tableModel.Update(msg)
			if tableModel, ok := updated.(tables.Model); ok {
				*m.tableModel = tableModel
				return m, cmd
			}
		}
	}

	if msg.String() == "q" || msg.String() == "ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m QueryModel) handleQueryResult(msg queryResultMsg) (tea.Model, tea.Cmd) {
	m.warnings = msg.warnings
	m.vector = msg.vector
	m.err = msg.err

	if m.err != nil {
		m.state = stateError
		return m, tea.Quit
	}

	return m.handleOutputFormat()
}

func (m QueryModel) handleOutputFormat() (tea.Model, tea.Cmd) {
	switch m.output {
	case "graph":
		return m.handleGraphOutput()
	case "table":
		return m.handleTableOutput()
	default:
		// For json/yaml, we'll just transition to success state
		m.state = stateSuccess
		return m, tea.Quit
	}
}

func (m QueryModel) handleGraphOutput() (tea.Model, tea.Cmd) {
	m.state = stateSuccess
	// Get terminal width directly
	width := m.width
	if width == 0 {
		// Try to get actual terminal width
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			width = termWidth
		} else {
			width = 80 // fallback default
		}
	}
	m.chartContent = charts.Barchart(m.vector, width)
	return m, nil
}

func (m QueryModel) handleTableOutput() (tea.Model, tea.Cmd) {
	tableModel, err := tables.PrintQuery(m.vector)
	if err != nil {
		m.err = err
		m.state = stateError
		return m, tea.Quit
	}
	m.tableModel = &tableModel
	m.state = stateShowingTable
	return m, m.tableModel.Init()
}

func (m QueryModel) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if m.state == stateLoading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m QueryModel) updateTableModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.tableModel.Update(msg)
	if tableModel, ok := updated.(tables.Model); ok {
		*m.tableModel = tableModel
		return m, cmd
	}
	return m, nil
}

func (m QueryModel) View() string {
	var s strings.Builder

	switch m.state {
	case stateLoading:
		s.WriteString(fmt.Sprintf("\n%s Executing query: %s\n\n", m.spinner.View(), m.query))

	case stateError:
		s.WriteString("\n")
		s.WriteString(errorStyle.Render("Error: ") + m.err.Error() + "\n")

	case stateSuccess:
		// Show warnings if any
		if len(m.warnings) > 0 {
			s.WriteString("\n")
			s.WriteString(warningStyle.Render("Warnings:\n"))
			for _, w := range m.warnings {
				s.WriteString(warningStyle.Render(fmt.Sprintf("  • %s\n", w)))
			}
			s.WriteString("\n")
		}

		// Display based on output format
		switch m.output {
		case "graph":
			s.WriteString(m.chartContent)
			if !m.quitting {
				s.WriteString("\n\n")
				s.WriteString("Press q or ctrl+c to quit\n")
			} else {
				s.WriteString("\n")
			}
		case "json":
			output := formatVector(m.vector, m.warnings, m.err)
			jsonBytes, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				s.WriteString(errorStyle.Render(fmt.Sprintf("Error formatting JSON: %v\n", err)))
			} else {
				s.Write(jsonBytes)
				s.WriteString("\n")
			}
		case "yaml":
			output := formatVector(m.vector, m.warnings, m.err)
			yamlBytes, err := yaml.Marshal(output)
			if err != nil {
				s.WriteString(errorStyle.Render(fmt.Sprintf("Error formatting YAML: %v\n", err)))
			} else {
				s.Write(yamlBytes)
				s.WriteString("\n")
			}
		}

	case stateShowingTable:
		if m.tableModel != nil && !m.quitting {
			// Show warnings if any
			if len(m.warnings) > 0 {
				s.WriteString("\n")
				s.WriteString(warningStyle.Render("Warnings: "))
				s.WriteString(fmt.Sprintf("%v\n", m.warnings))
			}
			s.WriteString(m.tableModel.View())
		}
	}

	return s.String()
}

// GetResult returns the final result for non-interactive outputs
func (m QueryModel) GetResult() (warnings v1.Warnings, vector model.Vector, err error) {
	return m.warnings, m.vector, m.err
}

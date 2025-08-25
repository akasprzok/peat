package tables

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	teatable "github.com/evertras/bubble-table/table"
	"github.com/prometheus/common/model"
)

type Model struct {
	table           teatable.Model
	filterTextInput textinput.Model
}

func PrintQuery(vector model.Vector) (Model, error) {
	return queryModel(vector)
}

func queryModel(vector model.Vector) (Model, error) {
	maxValue := 0
	longestMetric := 0
	rows := make([]table.Row, 0, len(vector))
	for _, sample := range vector {
		if int(sample.Value) > maxValue {
			maxValue = int(sample.Value)
		}
		unixNano := sample.Timestamp.UnixNano()
		timestamp := time.Unix(0, unixNano).Format(time.RFC3339)
		if len(sample.Metric.String()) > longestMetric {
			longestMetric = len(sample.Metric.String())
		}
		rows = append(rows, table.NewRow(table.RowData{
			"metric":    sample.Metric.String(),
			"value":     sample.Value.String(),
			"timestamp": timestamp,
		}))
	}

	columns := []table.Column{
		table.NewColumn("metric", "Metric", max(longestMetric+1, 6)).WithFiltered(true),
		table.NewColumn("value", "Value", max(len(strconv.Itoa(maxValue))+1, 6)).WithFiltered(true),
		table.NewColumn("timestamp", "Timestamp", 26).WithFiltered(true),
	}

	return Model{
		table: table.
			New(columns).
			Filtered(true).
			Focused(true).
			WithFooterVisibility(true).
			WithPageSize(10).
			WithRows(rows),
		filterTextInput: textinput.New(),
	}, nil

}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// global
		if msg.String() == "ctrl+c" {
			cmds = append(cmds, tea.Quit)

			return m, tea.Batch(cmds...)
		}
		// event to filter
		if m.filterTextInput.Focused() {
			if msg.String() == "enter" {
				m.filterTextInput.Blur()
			} else {
				m.filterTextInput, _ = m.filterTextInput.Update(msg)
			}
			m.table = m.table.WithFilterInput(m.filterTextInput)

			return m, tea.Batch(cmds...)
		}

		// others component
		switch msg.String() {
		case "/":
			m.filterTextInput.Focus()
		case "q":
			cmds = append(cmds, tea.Quit)
			return m, tea.Batch(cmds...)
		default:
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	body := strings.Builder{}

	body.WriteString(m.table.View())
	body.WriteString("\nPress / + letters to start filtering, and q or ctrl+c to quit")

	return body.String()
}

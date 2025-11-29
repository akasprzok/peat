package tables

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	rows := make([]teatable.Row, 0, len(vector))
	for _, sample := range vector {
		if int(sample.Value) > maxValue {
			maxValue = int(sample.Value)
		}
		unixNano := sample.Timestamp.UnixNano()
		timestamp := time.Unix(0, unixNano).Format(time.RFC3339)
		if len(sample.Metric.String()) > longestMetric {
			longestMetric = len(sample.Metric.String())
		}
		rows = append(rows, teatable.NewRow(teatable.RowData{
			"metric":    sample.Metric.String(),
			"value":     sample.Value.String(),
			"timestamp": timestamp,
		}))
	}

	columns := []teatable.Column{
		teatable.NewColumn("metric", "Metric", max(longestMetric+1, 6)).WithFiltered(true),
		teatable.NewColumn("value", "Value", max(len(strconv.Itoa(maxValue))+1, 6)).WithFiltered(true),
		teatable.NewColumn("timestamp", "Timestamp", 26).WithFiltered(true),
	}

	return Model{
		table: teatable.
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

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			cmds = append(cmds, tea.Quit)
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

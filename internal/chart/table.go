package chart

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/socketspace-jihad/sow/internal/data"
)

type TableRenderer struct {
	table  table.Model
	cfg    *Config
	width  int
	height int
	ready  bool
}

func (r *TableRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	if ds == nil || len(ds.Rows) == 0 {
		return
	}

	numCols := 0
	if len(ds.Headers) > 0 {
		numCols = len(ds.Headers)
	} else if len(ds.Rows) > 0 {
		numCols = len(ds.Rows[0])
	}
	if numCols == 0 {
		return
	}

	colWidth := (width - 4) / numCols
	if colWidth < 8 {
		colWidth = 8
	}

	cols := make([]table.Column, numCols)
	for i := 0; i < numCols; i++ {
		title := fmt.Sprintf("Col %d", i)
		if i < len(ds.Headers) {
			title = ds.Headers[i]
		}
		cols[i] = table.Column{Title: title, Width: colWidth}
	}

	rows := make([]table.Row, len(ds.Rows))
	for i, row := range ds.Rows {
		r := make(table.Row, numCols)
		for j := 0; j < numCols; j++ {
			if j < len(row) {
				r[j] = row[j]
			}
		}
		rows[i] = r
	}

	h := height - 4
	if h < 3 {
		h = 3
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(h),
	)

	s := table.DefaultStyles()
	if len(cfg.Theme.Colors) > 0 {
		s.Header = s.Header.
			BorderForeground(cfg.Theme.Colors[0]).
			Foreground(cfg.Theme.Colors[0]).
			Bold(true)
		s.Selected = s.Selected.
			Foreground(cfg.Theme.Background).
			Background(cfg.Theme.Colors[0])
	}
	t.SetStyles(s)

	r.table = t
	r.ready = true
}

func (r *TableRenderer) Push(dp data.DataPoint) {
	if !r.ready {
		return
	}
	// Append new row to the table
	label := dp.Label
	if label == "" {
		label = dp.Time.Format("15:04:05")
	}
	newRow := table.Row{label, fmt.Sprintf("%.2f", dp.Value)}
	rows := r.table.Rows()
	rows = append(rows, newRow)
	r.table.SetRows(rows)
}

func (r *TableRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	if r.ready {
		h := height - 4
		if h < 3 {
			h = 3
		}
		r.table.SetHeight(h)
	}
}

func (r *TableRenderer) HandleKey(msg tea.KeyMsg) bool {
	if !r.ready {
		return false
	}
	switch msg.String() {
	case "up", "k":
		r.table.MoveUp(1)
		return true
	case "down", "j":
		r.table.MoveDown(1)
		return true
	case "home", "g":
		r.table.GotoTop()
		return true
	case "end", "G":
		r.table.GotoBottom()
		return true
	}
	return false
}

func (r *TableRenderer) View() string {
	if !r.ready {
		return "  No data for table view"
	}
	return r.table.View()
}

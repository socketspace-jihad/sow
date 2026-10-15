package chart

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/muhammadjihad/sow/internal/data"
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

	// Determine columns
	numCols := 0
	if len(ds.Headers) > 0 {
		numCols = len(ds.Headers)
	} else if len(ds.Rows) > 0 {
		numCols = len(ds.Rows[0])
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

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-4),
	)

	s := table.DefaultStyles()
	if len(cfg.Theme.Colors) > 0 {
		s.Header = s.Header.
			BorderForeground(cfg.Theme.Colors[0]).
			Foreground(cfg.Theme.Colors[0]).
			Bold(true)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("#000000")).
			Background(cfg.Theme.Colors[0])
	}
	t.SetStyles(s)

	r.table = t
	r.ready = true
}

func (r *TableRenderer) Push(dp data.DataPoint) {
	// Tables don't support live push well - ignore
}

func (r *TableRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	if r.ready {
		r.table.SetHeight(height - 4)
	}
}

func (r *TableRenderer) View() string {
	if !r.ready {
		return "No data"
	}
	return r.table.View()
}

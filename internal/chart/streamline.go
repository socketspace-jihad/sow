package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/muhammadjihad/sow/internal/data"
)

type StreamlineRenderer struct {
	chart  streamlinechart.Model
	cfg    *Config
	width  int
	height int
}

func (r *StreamlineRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	r.chart = streamlinechart.New(width, height)
	if len(cfg.Theme.Colors) > 0 {
		r.chart.SetStyles(runes.ThinLineStyle, lipgloss.NewStyle().Foreground(cfg.Theme.Colors[0]))
	}

	if ds != nil && len(ds.Rows) > 0 {
		values := ds.NumericColumn(cfg.YColIdx)
		for _, v := range values {
			r.chart.Push(v)
		}
	}
	r.chart.Draw()
}

func (r *StreamlineRenderer) Push(dp data.DataPoint) {
	r.chart.Push(dp.Value)
	r.chart.Draw()
}

func (r *StreamlineRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.Draw()
}

func (r *StreamlineRenderer) View() string {
	return r.chart.View()
}

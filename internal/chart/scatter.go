package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/wavelinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/muhammadjihad/sow/internal/data"
)

type ScatterRenderer struct {
	chart  wavelinechart.Model
	cfg    *Config
	width  int
	height int
	count  float64
}

func (r *ScatterRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	r.chart = wavelinechart.New(width, height)
	if len(cfg.Theme.Colors) > 0 {
		r.chart.SetStyles(runes.ThinLineStyle, lipgloss.NewStyle().Foreground(cfg.Theme.Colors[0]))
	}

	if ds == nil || len(ds.Rows) == 0 {
		return
	}

	values := ds.NumericColumn(cfg.YColIdx)
	if len(values) == 0 {
		return
	}

	minY, maxY := values[0], values[0]
	for _, v := range values {
		if v < minY {
			minY = v
		}
		if v > maxY {
			maxY = v
		}
	}
	if minY == maxY {
		minY -= 1
		maxY += 1
	}
	padding := (maxY - minY) * 0.1
	minY -= padding
	maxY += padding

	r.chart = wavelinechart.New(width, height, wavelinechart.WithYRange(minY, maxY))
	if len(cfg.Theme.Colors) > 0 {
		r.chart.SetStyles(runes.ThinLineStyle, lipgloss.NewStyle().Foreground(cfg.Theme.Colors[0]))
	}

	for i, v := range values {
		r.chart.Plot(canvas.Float64Point{X: float64(i), Y: v})
	}
	r.count = float64(len(values))
	r.chart.Draw()
}

func (r *ScatterRenderer) Push(dp data.DataPoint) {
	r.chart.Plot(canvas.Float64Point{X: r.count, Y: dp.Value})
	r.count++
	r.chart.Draw()
}

func (r *ScatterRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.Draw()
}

func (r *ScatterRenderer) View() string {
	return r.chart.View()
}

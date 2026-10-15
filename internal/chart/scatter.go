package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/wavelinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
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

	if cfg.FixScale {
		if cfg.YMin != 0 || cfg.YMax != 0 {
			minY = cfg.YMin
			maxY = cfg.YMax
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

	// Plot overlays
	for _, ov := range cfg.Overlays {
		ovVals := ds.NumericColumn(ov.ColIdx)
		colorIdx := ov.Color % len(cfg.Theme.Colors)
		r.chart.SetDataSetStyles(ov.Name, runes.ThinLineStyle,
			lipgloss.NewStyle().Foreground(cfg.Theme.Colors[colorIdx]))
		for i, v := range ovVals {
			r.chart.PlotDataSet(ov.Name, canvas.Float64Point{X: float64(i), Y: v})
		}
	}

	r.count = float64(len(values))
	r.chart.DrawAll()
}

func (r *ScatterRenderer) Push(dp data.DataPoint) {
	r.chart.Plot(canvas.Float64Point{X: r.count, Y: dp.Value})
	r.count++
	r.chart.DrawAll()
}

func (r *ScatterRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.DrawAll()
}

func (r *ScatterRenderer) HandleKey(msg tea.KeyMsg) bool { return false }

func (r *ScatterRenderer) View() string {
	return r.chart.View()
}

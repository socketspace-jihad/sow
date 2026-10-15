package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/wavelinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type LineRenderer struct {
	chart  wavelinechart.Model
	cfg    *Config
	width  int
	height int
	count  float64
}

func (r *LineRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
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

	// Compute Y range from primary + all overlay data
	minY, maxY := values[0], values[0]
	for _, v := range values {
		if v < minY {
			minY = v
		}
		if v > maxY {
			maxY = v
		}
	}
	for _, ov := range cfg.Overlays {
		ovVals := ds.NumericColumn(ov.ColIdx)
		for _, v := range ovVals {
			if v < minY {
				minY = v
			}
			if v > maxY {
				maxY = v
			}
		}
	}

	// Apply fixed scale if set
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

	// Plot primary series
	for i, v := range values {
		r.chart.Plot(canvas.Float64Point{X: float64(i), Y: v})
	}

	// Plot overlays as separate named datasets
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

func (r *LineRenderer) Push(dp data.DataPoint) {
	r.chart.Plot(canvas.Float64Point{X: r.count, Y: dp.Value})
	r.count++
	r.chart.DrawAll()
}

func (r *LineRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.DrawAll()
}

func (r *LineRenderer) View() string {
	return r.chart.View()
}

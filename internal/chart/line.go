package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/wavelinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type LineRenderer struct {
	chart      wavelinechart.Model
	cfg        *Config
	width      int
	height     int
	count      float64
	rowCounts  map[string]float64 // per-row X counters
	knownRows  map[string]bool
	colorIndex int
}

func (r *LineRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height
	r.knownRows = make(map[string]bool)
	r.rowCounts = make(map[string]float64)
	r.colorIndex = 1

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
	label := dp.Label
	if label == "" || label == r.cfg.YLabel {
		r.chart.Plot(canvas.Float64Point{X: r.count, Y: dp.Value})
		r.count++
	} else {
		// Multi-row: route to named dataset
		if !r.knownRows[label] {
			r.knownRows[label] = true
			colorIdx := r.colorIndex % len(r.cfg.Theme.Colors)
			r.colorIndex++
			r.chart.SetDataSetStyles(label, runes.ThinLineStyle,
				lipgloss.NewStyle().Foreground(r.cfg.Theme.Colors[colorIdx]))
		}
		x := r.rowCounts[label]
		r.chart.PlotDataSet(label, canvas.Float64Point{X: x, Y: dp.Value})
		r.rowCounts[label] = x + 1
	}
	r.chart.DrawAll()
}

func (r *LineRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.DrawAll()
}

func (r *LineRenderer) HandleKey(msg tea.KeyMsg) bool { return false }

func (r *LineRenderer) View() string {
	return r.chart.View()
}

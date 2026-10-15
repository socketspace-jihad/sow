package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
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

	if cfg.FixScale && cfg.YMax > 0 {
		r.chart.SetYRange(cfg.YMin, cfg.YMax)
	}

	if ds != nil && len(ds.Rows) > 0 {
		values := ds.NumericColumn(cfg.YColIdx)
		for _, v := range values {
			r.chart.Push(v)
		}

		// Setup overlay datasets
		for _, ov := range cfg.Overlays {
			colorIdx := ov.Color % len(cfg.Theme.Colors)
			r.chart.SetDataSetStyles(ov.Name, runes.ThinLineStyle,
				lipgloss.NewStyle().Foreground(cfg.Theme.Colors[colorIdx]))
			ovVals := ds.NumericColumn(ov.ColIdx)
			for _, v := range ovVals {
				r.chart.PushDataSet(ov.Name, v)
			}
		}
	}
	r.chart.DrawAll()
}

func (r *StreamlineRenderer) Push(dp data.DataPoint) {
	r.chart.Push(dp.Value)
	r.chart.DrawAll()
}

func (r *StreamlineRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.DrawAll()
}

func (r *StreamlineRenderer) View() string {
	return r.chart.View()
}

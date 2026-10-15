package chart

import (
	"fmt"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type BarRenderer struct {
	chart      barchart.Model
	horizontal bool
	cfg        *Config
	width      int
	height     int
}

func (r *BarRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	r.chart = barchart.New(width, height)
	r.chart.SetHorizontal(r.horizontal)

	if ds == nil || len(ds.Rows) == 0 {
		return
	}

	var labels []string
	if cfg.XColIdx >= 0 {
		labels = ds.StringColumn(cfg.XColIdx)
	}
	values := ds.NumericColumn(cfg.YColIdx)

	bars := make([]barchart.BarData, len(values))
	for i, v := range values {
		label := fmt.Sprintf("%d", i)
		if i < len(labels) {
			label = labels[i]
		}
		colorIdx := i % len(cfg.Theme.Colors)
		bars[i] = barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{
					Name:  label,
					Value: v,
					Style: lipgloss.NewStyle().Foreground(cfg.Theme.Colors[colorIdx]),
				},
			},
		}
	}

	r.chart.PushAll(bars)
	r.chart.Draw()
}

func (r *BarRenderer) Push(dp data.DataPoint) {
	colorIdx := 0
	if r.cfg != nil && len(r.cfg.Theme.Colors) > 0 {
		colorIdx = 0
	}
	color := lipgloss.Color("#8BE9FD")
	if r.cfg != nil && len(r.cfg.Theme.Colors) > 0 {
		color = r.cfg.Theme.Colors[colorIdx]
	}

	r.chart.Push(barchart.BarData{
		Label: dp.Label,
		Values: []barchart.BarValue{
			{
				Name:  dp.Label,
				Value: dp.Value,
				Style: lipgloss.NewStyle().Foreground(color),
			},
		},
	})
	r.chart.Draw()
}

func (r *BarRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.Draw()
}

func (r *BarRenderer) View() string {
	return r.chart.View()
}

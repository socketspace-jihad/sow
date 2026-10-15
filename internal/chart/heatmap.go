package chart

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/NimbleMarkets/ntcharts/heatmap"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type HeatmapRenderer struct {
	chart    heatmap.Model
	cfg      *Config
	width    int
	height   int
	minVal   float64
	maxVal   float64
	hasData  bool
}

func (r *HeatmapRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height - 3 // reserve space for legend

	r.chart = heatmap.New(r.width, r.height)
	r.minVal = math.MaxFloat64
	r.maxVal = -math.MaxFloat64

	if ds != nil && len(ds.Rows) > 0 {
		for _, row := range ds.Rows {
			if len(row) >= 3 {
				x, _ := strconv.ParseFloat(row[0], 64)
				y, _ := strconv.ParseFloat(row[1], 64)
				v, _ := strconv.ParseFloat(row[2], 64)
				r.chart.Push(heatmap.NewHeatPoint(x, y, v))
				r.trackRange(v)
			} else if len(row) >= 2 {
				x, _ := strconv.ParseFloat(row[cfg.XColIdx], 64)
				v, _ := strconv.ParseFloat(row[cfg.YColIdx], 64)
				r.chart.Push(heatmap.NewHeatPoint(x, 0, v))
				r.trackRange(v)
			}
		}
	}
	r.chart.Draw()
}

func (r *HeatmapRenderer) trackRange(v float64) {
	r.hasData = true
	if v < r.minVal {
		r.minVal = v
	}
	if v > r.maxVal {
		r.maxVal = v
	}
}

func (r *HeatmapRenderer) Push(dp data.DataPoint) {
	r.chart.Push(heatmap.NewHeatPoint(float64(dp.Time.UnixMilli()%1000), 0, dp.Value))
	r.trackRange(dp.Value)
	r.chart.Draw()
}

func (r *HeatmapRenderer) Resize(width, height int) {
	r.width = width
	r.height = height - 3
	r.chart.Resize(r.width, r.height)
	r.chart.Draw()
}

func (r *HeatmapRenderer) View() string {
	var b strings.Builder
	b.WriteString(r.chart.View())
	b.WriteString("\n")
	b.WriteString(r.colorScaleLegend())
	return b.String()
}

func (r *HeatmapRenderer) colorScaleLegend() string {
	if !r.hasData {
		return ""
	}

	scale := r.chart.ColorScale
	if len(scale) == 0 {
		scale = heatmap.GetDefaultColorScale()
	}

	// Render color gradient bar
	barWidth := r.width - 20
	if barWidth < 10 {
		barWidth = 10
	}

	var bar strings.Builder
	dim := r.cfg.Theme.Axis

	minLabel := fmt.Sprintf("%.1f", r.minVal)
	maxLabel := fmt.Sprintf("%.1f", r.maxVal)

	bar.WriteString("  ")
	bar.WriteString(dim.Render(minLabel + " "))

	charsPerColor := barWidth / len(scale)
	if charsPerColor < 1 {
		charsPerColor = 1
	}
	for _, c := range scale {
		style := lipgloss.NewStyle().Background(c)
		bar.WriteString(style.Render(strings.Repeat(" ", charsPerColor)))
	}

	bar.WriteString(dim.Render(" " + maxLabel))

	return bar.String()
}

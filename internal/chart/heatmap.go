package chart

import (
	"strconv"

	"github.com/NimbleMarkets/ntcharts/heatmap"
	"github.com/socketspace-jihad/sow/internal/data"
)

type HeatmapRenderer struct {
	chart  heatmap.Model
	cfg    *Config
	width  int
	height int
}

func (r *HeatmapRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	r.chart = heatmap.New(width, height)

	if ds != nil && len(ds.Rows) > 0 {
		// Expect rows with at least 3 columns: x, y, value
		// Or 2 columns: index as x, column 0 as y, column 1 as value
		for _, row := range ds.Rows {
			if len(row) >= 3 {
				x, _ := strconv.ParseFloat(row[0], 64)
				y, _ := strconv.ParseFloat(row[1], 64)
				v, _ := strconv.ParseFloat(row[2], 64)
				r.chart.Push(heatmap.NewHeatPoint(x, y, v))
			} else if len(row) >= 2 {
				x, _ := strconv.ParseFloat(row[cfg.XColIdx], 64)
				v, _ := strconv.ParseFloat(row[cfg.YColIdx], 64)
				r.chart.Push(heatmap.NewHeatPoint(x, 0, v))
			}
		}
	}
	r.chart.Draw()
}

func (r *HeatmapRenderer) Push(dp data.DataPoint) {
	r.chart.Push(heatmap.NewHeatPoint(float64(dp.Time.UnixMilli()%1000), 0, dp.Value))
	r.chart.Draw()
}

func (r *HeatmapRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.Draw()
}

func (r *HeatmapRenderer) View() string {
	return r.chart.View()
}

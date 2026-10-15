package chart

import (
	"github.com/NimbleMarkets/ntcharts/sparkline"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type SparklineRenderer struct {
	chart  sparkline.Model
	cfg    *Config
	width  int
	height int
}

func (r *SparklineRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	r.chart = sparkline.New(width, height)
	if len(cfg.Theme.Colors) > 0 {
		r.chart.Style = lipgloss.NewStyle().Foreground(cfg.Theme.Colors[0])
	}

	if ds != nil && len(ds.Rows) > 0 {
		values := ds.NumericColumn(cfg.YColIdx)
		r.chart.PushAll(values)
	}
	r.chart.Draw()
}

func (r *SparklineRenderer) Push(dp data.DataPoint) {
	r.chart.Push(dp.Value)
	r.chart.Draw()
}

func (r *SparklineRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.Draw()
}

func (r *SparklineRenderer) HandleKey(msg tea.KeyMsg) bool { return false }

func (r *SparklineRenderer) View() string {
	return r.chart.View()
}

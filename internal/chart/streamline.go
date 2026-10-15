package chart

import (
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type StreamlineRenderer struct {
	chart      streamlinechart.Model
	cfg        *Config
	width      int
	height     int
	knownRows  map[string]bool
	stackAccum float64
	stackCount int
}

func (r *StreamlineRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height
	r.knownRows = make(map[string]bool)

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
	label := dp.Label

	if label != "" && r.cfg.HiddenRows[label] {
		return
	}

	if r.cfg.RowMode == "stack" {
		r.pushStack(dp)
		return
	}

	if label == "" || label == r.cfg.YLabel {
		r.chart.Push(dp.Value)
	} else {
		if !r.knownRows[label] {
			r.knownRows[label] = true
			colorIdx := r.stableColor(label)
			r.chart.SetDataSetStyles(label, runes.ThinLineStyle,
				lipgloss.NewStyle().Foreground(r.cfg.Theme.Colors[colorIdx]))
		}
		r.chart.PushDataSet(label, dp.Value)
	}
	r.chart.DrawAll()
}

// stableColor returns a consistent color index for a label.
// Assigns new colors incrementally for labels not yet in LabelColors.
func (r *StreamlineRenderer) stableColor(label string) int {
	if r.cfg.LabelColors == nil {
		r.cfg.LabelColors = make(map[string]int)
	}
	if idx, ok := r.cfg.LabelColors[label]; ok {
		return idx % len(r.cfg.Theme.Colors)
	}
	idx := len(r.cfg.LabelColors) + 1
	r.cfg.LabelColors[label] = idx
	return idx % len(r.cfg.Theme.Colors)
}

func (r *StreamlineRenderer) pushStack(dp data.DataPoint) {
	if dp.Row == 0 && r.stackCount > 0 {
		r.chart.Push(r.stackAccum)
		r.chart.DrawAll()
		r.stackAccum = 0
		r.stackCount = 0
	}
	r.stackAccum += dp.Value
	r.stackCount++
}

func (r *StreamlineRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.chart.Resize(width, height)
	r.chart.DrawAll()
}

func (r *StreamlineRenderer) HandleKey(msg tea.KeyMsg) bool { return false }

func (r *StreamlineRenderer) View() string {
	return r.chart.View()
}

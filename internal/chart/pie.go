package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
)

type PieRenderer struct {
	cfg    *Config
	labels []string
	values []float64
	width  int
	height int
}

func (r *PieRenderer) Setup(ds *data.DataSet, cfg *Config, width, height int) {
	r.cfg = cfg
	r.width = width
	r.height = height

	if ds != nil && len(ds.Rows) > 0 {
		if cfg.XColIdx >= 0 {
			r.labels = ds.StringColumn(cfg.XColIdx)
		}
		r.values = ds.NumericColumn(cfg.YColIdx)
	}
}

func (r *PieRenderer) Push(dp data.DataPoint) {
	r.labels = append(r.labels, dp.Label)
	r.values = append(r.values, dp.Value)
}

func (r *PieRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
}

func (r *PieRenderer) View() string {
	if len(r.values) == 0 {
		return "No data"
	}

	total := 0.0
	for _, v := range r.values {
		total += math.Abs(v)
	}
	if total == 0 {
		return "All values are zero"
	}

	colors := r.cfg.Theme.Colors
	barWidth := r.width - 2
	if barWidth < 10 {
		barWidth = 10
	}

	// Build proportional bar
	var bar strings.Builder
	for i, v := range r.values {
		proportion := math.Abs(v) / total
		chars := int(math.Round(proportion * float64(barWidth)))
		if chars < 1 && v > 0 {
			chars = 1
		}
		colorIdx := i % len(colors)
		style := lipgloss.NewStyle().Background(colors[colorIdx]).Foreground(r.cfg.Theme.Background)
		bar.WriteString(style.Render(strings.Repeat(" ", chars)))
	}

	// Build legend
	var legend strings.Builder
	legend.WriteString("\n")
	for i, v := range r.values {
		label := fmt.Sprintf("%d", i)
		if i < len(r.labels) && r.labels[i] != "" {
			label = r.labels[i]
		}
		pct := (math.Abs(v) / total) * 100
		colorIdx := i % len(colors)
		dot := lipgloss.NewStyle().Foreground(colors[colorIdx]).Render("██")
		legend.WriteString(fmt.Sprintf("  %s %s: %.1f (%.1f%%)\n", dot, label, v, pct))
	}

	return bar.String() + legend.String()
}

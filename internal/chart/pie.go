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

	if r.cfg.PieStyle == "circle" {
		return r.circleView(total)
	}
	return r.barView(total)
}

func (r *PieRenderer) barView(total float64) string {
	colors := r.cfg.Theme.Colors
	barWidth := r.width - 2
	if barWidth < 10 {
		barWidth = 10
	}

	var bar strings.Builder
	bar.WriteString("  ")
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

	return bar.String() + "\n" + r.legendView(total)
}

func (r *PieRenderer) circleView(total float64) string {
	colors := r.cfg.Theme.Colors
	// Render a half-circle using unicode block characters
	// Use available height for the circle radius
	radius := r.height/2 - 2
	if radius < 3 {
		radius = 3
	}
	diameter := radius * 2

	// Compute cumulative angles for each slice
	cumAngles := make([]float64, len(r.values)+1)
	cumAngles[0] = 0
	for i, v := range r.values {
		cumAngles[i+1] = cumAngles[i] + (math.Abs(v) / total) * 2 * math.Pi
	}

	var circle strings.Builder
	centerX := float64(radius) * 2 // wider because terminal chars are taller than wide
	centerY := float64(radius)

	for row := 0; row <= diameter; row++ {
		circle.WriteString("  ")
		for col := 0; col <= int(centerX)*2; col++ {
			dx := (float64(col) - centerX) / 2.0 // adjust for char aspect ratio
			dy := float64(row) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > float64(radius)+0.5 {
				circle.WriteRune(' ')
				continue
			}

			angle := math.Atan2(dy, dx)
			if angle < 0 {
				angle += 2 * math.Pi
			}

			// Find which slice this angle belongs to
			sliceIdx := 0
			for i := 0; i < len(r.values); i++ {
				if angle >= cumAngles[i] && angle < cumAngles[i+1] {
					sliceIdx = i
					break
				}
			}

			colorIdx := sliceIdx % len(colors)
			style := lipgloss.NewStyle().Foreground(colors[colorIdx])
			circle.WriteString(style.Render("█"))
		}
		circle.WriteString("\n")
	}

	return circle.String() + r.legendView(total)
}

func (r *PieRenderer) legendView(total float64) string {
	colors := r.cfg.Theme.Colors
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
	return legend.String()
}

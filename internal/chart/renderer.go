package chart

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/socketspace-jihad/sow/internal/data"
)

// Renderer is the interface all chart types implement.
type Renderer interface {
	Setup(ds *data.DataSet, cfg *Config, width, height int)
	Push(dp data.DataPoint)
	Resize(width, height int)
	View() string
	// HandleKey forwards key events to renderers that need them (e.g., table scrolling).
	// Returns true if the key was consumed.
	HandleKey(msg tea.KeyMsg) bool
}

// NewRenderer creates a renderer for the given chart type.
func NewRenderer(chartType string) Renderer {
	switch chartType {
	case TypeBar:
		return &BarRenderer{horizontal: false}
	case TypeHBar:
		return &BarRenderer{horizontal: true}
	case TypeLine:
		return &LineRenderer{}
	case TypeScatter, TypeDot:
		return &ScatterRenderer{}
	case TypeSparkline:
		return &SparklineRenderer{}
	case TypeStreamline:
		return &StreamlineRenderer{}
	case TypeHeatmap:
		return &HeatmapRenderer{}
	case TypePie:
		return &PieRenderer{}
	case TypeTable:
		return &TableRenderer{}
	default:
		return &BarRenderer{horizontal: false}
	}
}

// ConfigFields returns the config fields relevant to a given chart type.
func ConfigFields(chartType string) []string {
	switch chartType {
	case TypeBar, TypeHBar:
		return []string{"ymin", "ymax"}
	case TypeLine, TypeScatter, TypeDot:
		return []string{"ymin", "ymax"}
	case TypeSparkline, TypeStreamline:
		return []string{"ymin", "ymax"}
	case TypePie:
		return []string{"pie-style"}
	case TypeHeatmap:
		return []string{}
	case TypeTable:
		return []string{}
	default:
		return []string{}
	}
}

// ConfigFieldOptions returns the allowed values for a dropdown field.
// Returns nil for free-text fields.
func ConfigFieldOptions(name string) []string {
	switch name {
	case "pie-style":
		return []string{"bar", "circle"}
	}
	return nil
}

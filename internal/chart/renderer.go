package chart

import "github.com/socketspace-jihad/sow/internal/data"

// Renderer is the interface all chart types implement.
type Renderer interface {
	// Setup initializes with static data.
	Setup(ds *data.DataSet, cfg *Config, width, height int)
	// Push adds a live data point.
	Push(dp data.DataPoint)
	// Resize updates dimensions.
	Resize(width, height int)
	// View returns the rendered chart string.
	View() string
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

package chart

import "github.com/socketspace-jihad/sow/internal/theme"

// Config holds all chart configuration from CLI flags.
type Config struct {
	ChartType   string
	Title       string
	XCol        string
	YCol        string
	XColIdx     int
	YColIdx     int
	Width       int
	Height      int
	Theme       *theme.Theme
	Live        bool
	Orientation string // "vertical" or "horizontal"

	// Legend / axis labels (auto-detected or from headers)
	XLabel string // e.g., "time", "Month"
	YLabel string // e.g., "google.com latency", "Revenue"
	Unit   string // e.g., "ms", "%", ""
}

// Chart type constants
const (
	TypeBar        = "bar"
	TypeHBar       = "hbar"
	TypeLine       = "line"
	TypeScatter    = "scatter"
	TypeDot        = "dot"
	TypeSparkline  = "sparkline"
	TypeStreamline = "streamline"
	TypeHeatmap    = "heatmap"
	TypePie        = "pie"
	TypeTable      = "table"
)

// AllTypes returns all chart type names in cycle order.
func AllTypes() []string {
	return []string{
		TypeBar, TypeHBar, TypeLine, TypeScatter,
		TypeSparkline, TypeStreamline, TypeHeatmap,
		TypePie, TypeTable,
	}
}

// NextType returns the next chart type for cycling.
func NextType(current string) string {
	types := AllTypes()
	for i, t := range types {
		if t == current {
			return types[(i+1)%len(types)]
		}
	}
	return TypeBar
}

// PrevType returns the previous chart type for cycling.
func PrevType(current string) string {
	types := AllTypes()
	for i, t := range types {
		if t == current {
			return types[(i-1+len(types))%len(types)]
		}
	}
	return TypeBar
}

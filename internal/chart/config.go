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

	// Per-chart configuration
	YMin     float64 // Fixed Y-axis minimum (0 = auto)
	YMax     float64 // Fixed Y-axis maximum (0 = auto)
	FixScale bool    // Whether YMin/YMax are set explicitly
	PieStyle string  // "bar" or "circle" (default: "bar")

	// Live Y filter — set by the Y column modal, NOT by --y flag.
	LiveYFilter string
	// LiveColumns stores detected column names from piped tabular data.
	LiveColumns []string

	// Row-stream display options
	RowMode      string            // "overlay" (default) or "stack"
	HiddenRows   map[string]bool   // rows to hide (toggled interactively)
	LabelColors  map[string]int    // stable label→color index (survives rebuild)

	// Overlay configuration
	Overlays []OverlayConfig
}

// OverlayConfig defines one overlay series on the chart.
type OverlayConfig struct {
	Name   string // display name
	ColIdx int    // column index in the dataset
	Color  int    // index into theme colors
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

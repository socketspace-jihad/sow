package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines color palette for charts.
type Theme struct {
	Name       string
	Title      lipgloss.Style
	Axis       lipgloss.Style
	Colors     []lipgloss.Color // Series colors for bars, lines, etc.
	Background lipgloss.Color
	Foreground lipgloss.Color
	Help       lipgloss.Style
	Border     lipgloss.Style
}

var themes = map[string]*Theme{
	"default": DefaultTheme(),
	"ocean":   OceanTheme(),
	"forest":  ForestTheme(),
	"sunset":  SunsetTheme(),
	"mono":    MonoTheme(),
}

// Names returns all available theme names.
func Names() []string {
	return []string{"default", "ocean", "forest", "sunset", "mono"}
}

// Get returns a theme by name, falling back to default.
func Get(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}

// Next returns the theme after the current one (for cycling).
func Next(current string) *Theme {
	names := Names()
	for i, n := range names {
		if n == current {
			return Get(names[(i+1)%len(names)])
		}
	}
	return Get("default")
}

func DefaultTheme() *Theme {
	return &Theme{
		Name:       "default",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A")),
		Background: lipgloss.Color("#282A36"),
		Foreground: lipgloss.Color("#F8F8F2"),
		Colors: []lipgloss.Color{
			"#8BE9FD", // cyan
			"#50FA7B", // green
			"#FFB86C", // orange
			"#FF79C6", // pink
			"#BD93F9", // purple
			"#F1FA8C", // yellow
			"#FF5555", // red
			"#F8F8F2", // white
		},
	}
}

func OceanTheme() *Theme {
	return &Theme{
		Name:       "ocean",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4FF")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#4A6FA5")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#4A6FA5")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#2B4570")),
		Background: lipgloss.Color("#0A1628"),
		Foreground: lipgloss.Color("#E0F0FF"),
		Colors: []lipgloss.Color{
			"#00D4FF", // bright blue
			"#00FFC8", // aqua
			"#0088FF", // blue
			"#66E0FF", // light blue
			"#00AACC", // teal
			"#4488FF", // medium blue
			"#88DDFF", // sky
			"#AAEEFF", // ice
		},
	}
}

func ForestTheme() *Theme {
	return &Theme{
		Name:       "forest",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7FFF00")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#556B2F")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#556B2F")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#2E4600")),
		Background: lipgloss.Color("#1A2E05"),
		Foreground: lipgloss.Color("#E0FFE0"),
		Colors: []lipgloss.Color{
			"#7FFF00", // chartreuse
			"#32CD32", // lime green
			"#FFD700", // gold
			"#90EE90", // light green
			"#228B22", // forest green
			"#ADFF2F", // green yellow
			"#00FA9A", // medium spring
			"#8FBC8F", // dark sea green
		},
	}
}

func SunsetTheme() *Theme {
	return &Theme{
		Name:       "sunset",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#8B4513")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#8B4513")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#5C2E00")),
		Background: lipgloss.Color("#2D1B00"),
		Foreground: lipgloss.Color("#FFF0E0"),
		Colors: []lipgloss.Color{
			"#FF6347", // tomato
			"#FF8C00", // dark orange
			"#FFD700", // gold
			"#FF4500", // orange red
			"#FF69B4", // hot pink
			"#FFA07A", // light salmon
			"#FF1493", // deep pink
			"#FFDAB9", // peach
		},
	}
}

func MonoTheme() *Theme {
	return &Theme{
		Name:       "mono",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),
		Background: lipgloss.Color("#000000"),
		Foreground: lipgloss.Color("#FFFFFF"),
		Colors: []lipgloss.Color{
			"#FFFFFF",
			"#DDDDDD",
			"#BBBBBB",
			"#999999",
			"#777777",
			"#AAAAAA",
			"#CCCCCC",
			"#EEEEEE",
		},
	}
}

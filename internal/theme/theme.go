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

var (
	darkMode bool
	themes   map[string]*Theme
)

func init() {
	darkMode = lipgloss.HasDarkBackground()
	themes = buildThemes(darkMode)
}

func buildThemes(dark bool) map[string]*Theme {
	return map[string]*Theme{
		"default": defaultTheme(dark),
		"ocean":   oceanTheme(dark),
		"forest":  forestTheme(dark),
		"sunset":  sunsetTheme(dark),
		"mono":    monoTheme(dark),
	}
}

// IsDark returns whether the terminal has a dark background.
func IsDark() bool {
	return darkMode
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

// --- themes ---

func defaultTheme(dark bool) *Theme {
	if dark {
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
	return &Theme{
		Name:       "default",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D6336C")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#808898")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#C0C8D8")),
		Background: lipgloss.Color("#FFFFFF"),
		Foreground: lipgloss.Color("#1E1E2E"),
		Colors: []lipgloss.Color{
			"#0077B6", // deep blue
			"#2D9F2D", // green
			"#D35400", // burnt orange
			"#C2185B", // dark pink
			"#7B1FA2", // purple
			"#B8860B", // dark goldenrod
			"#C62828", // dark red
			"#455A64", // blue grey
		},
	}
}

func oceanTheme(dark bool) *Theme {
	if dark {
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
	return &Theme{
		Name:       "ocean",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#005F87")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#607D8B")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#78909C")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B0BEC5")),
		Background: lipgloss.Color("#FFFFFF"),
		Foreground: lipgloss.Color("#0A1628"),
		Colors: []lipgloss.Color{
			"#005F87", // dark cyan
			"#00796B", // teal
			"#0D47A1", // dark blue
			"#0277BD", // light blue dark
			"#006064", // dark teal
			"#1565C0", // blue
			"#01579B", // navy
			"#004D40", // deep teal
		},
	}
}

func forestTheme(dark bool) *Theme {
	if dark {
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
	return &Theme{
		Name:       "forest",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2E7D32")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6D7A5F")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8A6E")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#A5B89A")),
		Background: lipgloss.Color("#FFFFFF"),
		Foreground: lipgloss.Color("#1B2E1B"),
		Colors: []lipgloss.Color{
			"#2E7D32", // dark green
			"#1B5E20", // deep green
			"#B8860B", // dark goldenrod
			"#33691E", // olive
			"#004D40", // deep teal
			"#558B2F", // light green dark
			"#006400", // dark green
			"#4E342E", // brown
		},
	}
}

func sunsetTheme(dark bool) *Theme {
	if dark {
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
	return &Theme{
		Name:       "sunset",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#C62828")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#8D6E63")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#A1887F")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BCAAA4")),
		Background: lipgloss.Color("#FFFFFF"),
		Foreground: lipgloss.Color("#2D1B00"),
		Colors: []lipgloss.Color{
			"#C62828", // dark red
			"#D84315", // deep orange
			"#B8860B", // dark goldenrod
			"#BF360C", // burnt red
			"#AD1457", // dark pink
			"#D35400", // pumpkin
			"#880E4F", // dark magenta
			"#6D4C41", // brown
		},
	}
}

func monoTheme(dark bool) *Theme {
	if dark {
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
	return &Theme{
		Name:       "mono",
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#000000")),
		Axis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Border:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BBBBBB")),
		Background: lipgloss.Color("#FFFFFF"),
		Foreground: lipgloss.Color("#000000"),
		Colors: []lipgloss.Color{
			"#000000",
			"#222222",
			"#444444",
			"#666666",
			"#888888",
			"#555555",
			"#333333",
			"#111111",
		},
	}
}

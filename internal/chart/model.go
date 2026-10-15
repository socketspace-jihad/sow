package chart

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/socketspace-jihad/sow/internal/data"
	"github.com/socketspace-jihad/sow/internal/theme"
)

// dataMsg carries a live data point from stdin goroutine.
type dataMsg data.DataPoint

// liveStats tracks running statistics for live data.
type liveStats struct {
	count   int
	sum     float64
	min     float64
	max     float64
	current float64
	lastT   time.Time
}

func (s *liveStats) update(v float64, t time.Time) {
	s.count++
	s.sum += v
	s.current = v
	s.lastT = t
	if s.count == 1 {
		s.min = v
		s.max = v
	} else {
		s.min = math.Min(s.min, v)
		s.max = math.Max(s.max, v)
	}
}

func (s *liveStats) avg() float64 {
	if s.count == 0 {
		return 0
	}
	return s.sum / float64(s.count)
}

// Model is the top-level Bubble Tea model.
type Model struct {
	cfg      *Config
	ds       *data.DataSet
	liveCh   <-chan data.DataPoint
	renderer Renderer
	width    int
	height   int
	ready    bool
	showHelp bool
	quitting bool
	stats    liveStats
}

// NewModel creates a new Bubble Tea model.
func NewModel(cfg *Config, ds *data.DataSet, liveCh <-chan data.DataPoint) Model {
	return Model{
		cfg:    cfg,
		ds:     ds,
		liveCh: liveCh,
	}
}

func (m Model) Init() tea.Cmd {
	if m.liveCh != nil {
		return waitForData(m.liveCh)
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	case dataMsg:
		return m.handleData(msg)
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if !m.ready {
		return "  Waiting for data...\n"
	}

	var b strings.Builder

	// Title
	if m.cfg.Title != "" {
		title := m.cfg.Theme.Title.Render(m.cfg.Title)
		b.WriteString("  ")
		b.WriteString(title)
		b.WriteString("\n")
	}

	// Legend bar: [chart-type]  X: time  Y: google.com latency (ms)  theme:ocean
	b.WriteString(m.legendView())
	b.WriteString("\n")

	// Chart content
	chartView := m.renderer.View()
	b.WriteString(chartView)
	b.WriteString("\n")

	// Live stats bar (only in live mode with data)
	if m.cfg.Live && m.stats.count > 0 {
		b.WriteString(m.statsView())
		b.WriteString("\n")
	}

	// Help bar
	if m.showHelp {
		b.WriteString(m.helpView())
	} else {
		help := m.cfg.Theme.Help.Render("  ? help  q quit  tab switch  t theme")
		b.WriteString(help)
	}

	return b.String()
}

func (m Model) legendView() string {
	dim := m.cfg.Theme.Axis
	bright := m.cfg.Theme.Title
	accent := lipgloss.NewStyle()
	if len(m.cfg.Theme.Colors) > 0 {
		accent = lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[0])
	}

	parts := []string{
		"  " + dim.Render("[") + bright.Render(m.cfg.ChartType) + dim.Render("]"),
	}

	if m.cfg.XLabel != "" {
		parts = append(parts, dim.Render("X:")+accent.Render(m.cfg.XLabel))
	}
	if m.cfg.YLabel != "" {
		yText := m.cfg.YLabel
		if m.cfg.Unit != "" {
			yText += " (" + m.cfg.Unit + ")"
		}
		parts = append(parts, dim.Render("Y:")+accent.Render(yText))
	}

	parts = append(parts, dim.Render("theme:")+dim.Render(m.cfg.Theme.Name))

	return strings.Join(parts, "  ")
}

func (m Model) statsView() string {
	dim := m.cfg.Theme.Axis
	val := lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground)
	unit := m.cfg.Unit

	fmtVal := func(v float64) string {
		s := fmt.Sprintf("%.1f", v)
		if unit != "" {
			s += " " + unit
		}
		return val.Render(s)
	}

	parts := []string{
		"  " + dim.Render("cur:") + fmtVal(m.stats.current),
		dim.Render("min:") + fmtVal(m.stats.min),
		dim.Render("max:") + fmtVal(m.stats.max),
		dim.Render("avg:") + fmtVal(m.stats.avg()),
		dim.Render("n:") + val.Render(fmt.Sprintf("%d", m.stats.count)),
	}

	return strings.Join(parts, "  ")
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "tab":
		m.switchChart(NextType(m.cfg.ChartType))
		return m, nil
	case "shift+tab":
		m.switchChart(PrevType(m.cfg.ChartType))
		return m, nil
	case "t":
		newTheme := theme.Next(m.cfg.Theme.Name)
		m.cfg.Theme = newTheme
		m.rebuildChart()
		return m, nil
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "r":
		m.rebuildChart()
		return m, nil
	}
	return m, nil
}

func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	reserved := 6 // title + legend + help + padding
	if m.cfg.Live {
		reserved += 1 // stats bar
	}
	m.height = msg.Height - reserved

	if m.cfg.Width > 0 {
		m.width = m.cfg.Width
	}
	if m.cfg.Height > 0 {
		m.height = m.cfg.Height
	}

	if m.width < 10 {
		m.width = 10
	}
	if m.height < 5 {
		m.height = 5
	}

	if !m.ready {
		m.ready = true
		m.renderer = NewRenderer(m.cfg.ChartType)
		m.renderer.Setup(m.ds, m.cfg, m.width, m.height)
	} else {
		m.renderer.Resize(m.width, m.height)
	}

	return m, nil
}

func (m Model) handleData(msg dataMsg) (tea.Model, tea.Cmd) {
	dp := data.DataPoint(msg)

	m.stats.update(dp.Value, dp.Time)

	if m.renderer != nil {
		m.renderer.Push(dp)
	}

	// Continue listening for more data
	if m.liveCh != nil {
		return m, waitForData(m.liveCh)
	}
	return m, nil
}

func (m *Model) switchChart(newType string) {
	m.cfg.ChartType = newType
	m.rebuildChart()
}

func (m *Model) rebuildChart() {
	if m.width == 0 || m.height == 0 {
		return
	}
	m.renderer = NewRenderer(m.cfg.ChartType)
	m.renderer.Setup(m.ds, m.cfg, m.width, m.height)
}

func (m Model) helpView() string {
	s := m.cfg.Theme.Help

	lines := []string{
		"",
		s.Render("  Keyboard Shortcuts:"),
		"",
		s.Render("  q / Ctrl+C  ") + lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground).Render("Quit"),
		s.Render("  Tab         ") + lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground).Render("Next chart type"),
		s.Render("  Shift+Tab   ") + lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground).Render("Previous chart type"),
		s.Render("  t           ") + lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground).Render("Cycle color theme"),
		s.Render("  r           ") + lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground).Render("Redraw chart"),
		s.Render("  ?           ") + lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground).Render("Toggle this help"),
		"",
		s.Render(fmt.Sprintf("  Chart types: %s", strings.Join(AllTypes(), ", "))),
		s.Render(fmt.Sprintf("  Themes: %s", strings.Join(theme.Names(), ", "))),
		"",
	}
	return strings.Join(lines, "\n")
}

// waitForData returns a tea.Cmd that waits for the next data point from the channel.
func waitForData(ch <-chan data.DataPoint) tea.Cmd {
	return func() tea.Msg {
		dp, ok := <-ch
		if !ok {
			return nil
		}
		return dataMsg(dp)
	}
}

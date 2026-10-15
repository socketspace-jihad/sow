package chart

import (
	"fmt"
	"math"
	"strconv"
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

// overlayModal state
type overlayModal struct {
	active       bool
	input        string
	cursor       int
	message      string // feedback message
	selectMode   bool   // true = selecting existing overlay to remove
	selectIdx    int
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
	modal    overlayModal
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
		if m.modal.active {
			return m.handleModalKey(msg)
		}
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

	// Legend bar
	b.WriteString(m.legendView())
	b.WriteString("\n")

	// Chart content
	chartView := m.renderer.View()
	b.WriteString(chartView)
	b.WriteString("\n")

	// Overlay series legend (if any overlays active)
	if len(m.cfg.Overlays) > 0 {
		b.WriteString(m.overlayLegendView())
		b.WriteString("\n")
	}

	// Live stats bar
	if m.cfg.Live && m.stats.count > 0 {
		b.WriteString(m.statsView())
		b.WriteString("\n")
	}

	// Modal overlay (rendered on top)
	if m.modal.active {
		b.WriteString(m.modalView())
	} else if m.showHelp {
		b.WriteString(m.helpView())
	} else {
		help := m.cfg.Theme.Help.Render("  ? help  q quit  tab switch  t theme  o overlay")
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

func (m Model) overlayLegendView() string {
	dim := m.cfg.Theme.Axis
	parts := []string{"  " + dim.Render("overlays:")}
	for _, ov := range m.cfg.Overlays {
		colorIdx := ov.Color % len(m.cfg.Theme.Colors)
		dot := lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[colorIdx]).Render("●")
		parts = append(parts, dot+" "+dim.Render(ov.Name))
	}
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

// --- Key handlers ---

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
	case "o":
		// Open overlay modal
		if m.ds != nil {
			m.modal.active = true
			m.modal.input = ""
			m.modal.message = ""
			m.modal.selectMode = false
		}
		return m, nil
	case "O":
		// Open overlay select/remove mode
		if len(m.cfg.Overlays) > 0 {
			m.modal.active = true
			m.modal.selectMode = true
			m.modal.selectIdx = 0
			m.modal.message = ""
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.modal.selectMode {
		return m.handleSelectModeKey(key)
	}

	switch key {
	case "esc":
		m.modal.active = false
		return m, nil
	case "enter":
		m.addOverlay(m.modal.input)
		m.modal.active = false
		m.rebuildChart()
		return m, nil
	case "backspace":
		if len(m.modal.input) > 0 {
			m.modal.input = m.modal.input[:len(m.modal.input)-1]
		}
		return m, nil
	case "ctrl+c":
		m.modal.active = false
		return m, nil
	default:
		if len(key) == 1 {
			m.modal.input += key
		}
		return m, nil
	}
}

func (m Model) handleSelectModeKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+c":
		m.modal.active = false
		return m, nil
	case "up", "k":
		if m.modal.selectIdx > 0 {
			m.modal.selectIdx--
		}
		return m, nil
	case "down", "j":
		if m.modal.selectIdx < len(m.cfg.Overlays)-1 {
			m.modal.selectIdx++
		}
		return m, nil
	case "d", "backspace", "delete":
		// Remove selected overlay
		if m.modal.selectIdx < len(m.cfg.Overlays) {
			m.cfg.Overlays = append(m.cfg.Overlays[:m.modal.selectIdx], m.cfg.Overlays[m.modal.selectIdx+1:]...)
			if m.modal.selectIdx >= len(m.cfg.Overlays) && m.modal.selectIdx > 0 {
				m.modal.selectIdx--
			}
			if len(m.cfg.Overlays) == 0 {
				m.modal.active = false
			}
			m.rebuildChart()
		}
		return m, nil
	case "enter", "esc ":
		m.modal.active = false
		return m, nil
	}
	return m, nil
}

func (m *Model) addOverlay(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	colIdx := -1
	// Try matching by header name
	if m.ds != nil {
		for i, h := range m.ds.Headers {
			if strings.EqualFold(h, input) {
				colIdx = i
				break
			}
		}
	}

	// Try parsing as column index
	if colIdx < 0 {
		if idx, err := strconv.Atoi(input); err == nil {
			colIdx = idx
		}
	}

	if colIdx < 0 {
		return
	}

	// Don't add duplicate
	for _, ov := range m.cfg.Overlays {
		if ov.ColIdx == colIdx {
			return
		}
	}

	name := input
	if m.ds != nil && colIdx < len(m.ds.Headers) {
		name = m.ds.Headers[colIdx]
	}

	// Assign next available color (skip color 0 which is the primary series)
	colorIdx := len(m.cfg.Overlays) + 1

	m.cfg.Overlays = append(m.cfg.Overlays, OverlayConfig{
		Name:   name,
		ColIdx: colIdx,
		Color:  colorIdx,
	})
}

func (m Model) modalView() string {
	dim := m.cfg.Theme.Axis
	bright := m.cfg.Theme.Title
	fg := lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground)
	accent := lipgloss.NewStyle()
	if len(m.cfg.Theme.Colors) > 0 {
		accent = lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[0])
	}

	if m.modal.selectMode {
		lines := []string{
			"",
			"  " + bright.Render("Manage Overlays") + dim.Render("  (d=remove  esc=close)"),
			"",
		}
		for i, ov := range m.cfg.Overlays {
			colorIdx := ov.Color % len(m.cfg.Theme.Colors)
			dot := lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[colorIdx]).Render("●")
			marker := "  "
			if i == m.modal.selectIdx {
				marker = fg.Render("> ")
			}
			lines = append(lines, fmt.Sprintf("  %s%s %s %s",
				marker, dot, fg.Render(ov.Name),
				dim.Render(fmt.Sprintf("(col %d)", ov.ColIdx))))
		}
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	// Add overlay input mode
	lines := []string{
		"",
		"  " + bright.Render("Add Overlay") + dim.Render("  (enter=add  esc=cancel)"),
		"",
	}

	// Show available columns
	if m.ds != nil && len(m.ds.Headers) > 0 {
		lines = append(lines, "  "+dim.Render("Available columns:"))
		for i, h := range m.ds.Headers {
			marker := "  "
			if i == m.cfg.YColIdx {
				marker = dim.Render("* ") // mark primary Y column
			}
			lines = append(lines, fmt.Sprintf("    %s%s %s",
				marker, dim.Render(fmt.Sprintf("%d:", i)), accent.Render(h)))
		}
		lines = append(lines, "")
	}

	cursor := fg.Render("_")
	lines = append(lines, "  "+dim.Render("Column (name or index): ")+fg.Render(m.modal.input)+cursor)
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// --- Resize / data / rebuild ---

func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	reserved := 6
	if m.cfg.Live {
		reserved += 1
	}
	if len(m.cfg.Overlays) > 0 {
		reserved += 1
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
	fg := lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground)

	lines := []string{
		"",
		s.Render("  Keyboard Shortcuts:"),
		"",
		s.Render("  q / Ctrl+C  ") + fg.Render("Quit"),
		s.Render("  Tab         ") + fg.Render("Next chart type"),
		s.Render("  Shift+Tab   ") + fg.Render("Previous chart type"),
		s.Render("  t           ") + fg.Render("Cycle color theme"),
		s.Render("  r           ") + fg.Render("Redraw chart"),
		s.Render("  o           ") + fg.Render("Add overlay series"),
		s.Render("  O           ") + fg.Render("Manage/remove overlays"),
		s.Render("  ?           ") + fg.Render("Toggle this help"),
		"",
		s.Render(fmt.Sprintf("  Chart types: %s", strings.Join(AllTypes(), ", "))),
		s.Render(fmt.Sprintf("  Themes: %s", strings.Join(theme.Names(), ", "))),
		"",
	}
	return strings.Join(lines, "\n")
}

func waitForData(ch <-chan data.DataPoint) tea.Cmd {
	return func() tea.Msg {
		dp, ok := <-ch
		if !ok {
			return nil
		}
		return dataMsg(dp)
	}
}

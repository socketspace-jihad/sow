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

type dataMsg data.DataPoint

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

// modalMode tracks which modal is active.
type modalMode int

const (
	modalNone    modalMode = iota
	modalOverlay           // add overlay
	modalManage            // manage/remove overlays
	modalConfig            // edit chart configuration
)

type modal struct {
	mode         modalMode
	input        string
	selectIdx    int
	field        int    // for config modal: which field is selected
	dropdownIdx  int    // index into ConfigFieldOptions for enum fields
	dropdownOpen bool   // whether the dropdown is expanded
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
	modal    modal

	// Persistent data buffer — survives theme/chart type changes.
	liveBuffer   []data.DataPoint
	liveLabels   []string          // unique labels seen in order
	liveLabelsAt map[string]int    // label → color index
}

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
		if m.modal.mode != modalNone {
			return m.handleModalKey(msg)
		}
		// Forward to renderer first (e.g., table scrolling)
		if m.renderer != nil && m.renderer.HandleKey(msg) {
			return m, nil
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

	if m.cfg.Title != "" {
		b.WriteString("  ")
		b.WriteString(m.cfg.Theme.Title.Render(m.cfg.Title))
		b.WriteString("\n")
	}

	b.WriteString(m.legendView())
	b.WriteString("\n")

	b.WriteString(m.renderer.View())
	b.WriteString("\n")

	if len(m.cfg.Overlays) > 0 || len(m.liveLabels) > 1 {
		b.WriteString(m.overlayLegendView())
		b.WriteString("\n")
	}

	if m.cfg.Live && m.stats.count > 0 {
		b.WriteString(m.statsView())
		b.WriteString("\n")
	}

	switch m.modal.mode {
	case modalOverlay:
		b.WriteString(m.overlayModalView())
	case modalManage:
		b.WriteString(m.manageModalView())
	case modalConfig:
		b.WriteString(m.configModalView())
	default:
		if m.showHelp {
			b.WriteString(m.helpView())
		} else {
			b.WriteString(m.cfg.Theme.Help.Render("  ? help  q quit  tab switch  t theme  o overlay  c config"))
		}
	}

	return b.String()
}

// --- Legend views ---

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
	if m.cfg.FixScale {
		parts = append(parts, dim.Render(fmt.Sprintf("scale:%.0f-%.0f", m.cfg.YMin, m.cfg.YMax)))
	}
	parts = append(parts, dim.Render("theme:")+dim.Render(m.cfg.Theme.Name))
	return strings.Join(parts, "  ")
}

func (m Model) overlayLegendView() string {
	dim := m.cfg.Theme.Axis
	parts := []string{"  " + dim.Render("series:")}

	// Static overlays
	for _, ov := range m.cfg.Overlays {
		colorIdx := ov.Color % len(m.cfg.Theme.Colors)
		dot := lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[colorIdx]).Render("●")
		parts = append(parts, dot+" "+dim.Render(ov.Name))
	}

	// Live multi-row labels (e.g., cpu0, cpu1, cpu2...)
	if len(m.cfg.Overlays) == 0 && len(m.liveLabels) > 1 {
		for _, label := range m.liveLabels {
			colorIdx := m.liveLabelsAt[label] % len(m.cfg.Theme.Colors)
			dot := lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[colorIdx]).Render("●")
			parts = append(parts, dot+" "+dim.Render(label))
		}
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
		m.cfg.ChartType = NextType(m.cfg.ChartType)
		m.rebuildFromBuffer()
		return m, nil
	case "shift+tab":
		m.cfg.ChartType = PrevType(m.cfg.ChartType)
		m.rebuildFromBuffer()
		return m, nil
	case "t":
		m.cfg.Theme = theme.Next(m.cfg.Theme.Name)
		m.rebuildFromBuffer()
		return m, nil
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "r":
		m.rebuildFromBuffer()
		return m, nil
	case "o":
		m.modal = modal{mode: modalOverlay}
		return m, nil
	case "O":
		if len(m.cfg.Overlays) > 0 {
			m.modal = modal{mode: modalManage}
		}
		return m, nil
	case "c":
		m.modal = modal{mode: modalConfig}
		m.configSyncField()
		return m, nil
	}
	return m, nil
}

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.modal.mode {
	case modalOverlay:
		return m.handleOverlayModalKey(key)
	case modalManage:
		return m.handleManageModalKey(key)
	case modalConfig:
		return m.handleConfigModalKey(key)
	}
	return m, nil
}

// --- Overlay add modal ---

func (m Model) handleOverlayModalKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+c":
		m.modal.mode = modalNone
		return m, nil
	case "enter":
		m.addOverlay(m.modal.input)
		m.modal.mode = modalNone
		m.rebuildFromBuffer()
		return m, nil
	case "backspace":
		if len(m.modal.input) > 0 {
			m.modal.input = m.modal.input[:len(m.modal.input)-1]
		}
		return m, nil
	default:
		if len(key) == 1 {
			m.modal.input += key
		}
		return m, nil
	}
}

func (m *Model) addOverlay(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	colIdx := -1
	if m.ds != nil {
		for i, h := range m.ds.Headers {
			if strings.EqualFold(h, input) {
				colIdx = i
				break
			}
		}
	}
	if colIdx < 0 {
		if idx, err := strconv.Atoi(input); err == nil {
			colIdx = idx
		}
	}
	if colIdx < 0 {
		return
	}

	for _, ov := range m.cfg.Overlays {
		if ov.ColIdx == colIdx {
			return
		}
	}

	name := input
	if m.ds != nil && colIdx < len(m.ds.Headers) {
		name = m.ds.Headers[colIdx]
	}

	m.cfg.Overlays = append(m.cfg.Overlays, OverlayConfig{
		Name:   name,
		ColIdx: colIdx,
		Color:  len(m.cfg.Overlays) + 1,
	})
}

func (m Model) overlayModalView() string {
	dim := m.cfg.Theme.Axis
	bright := m.cfg.Theme.Title
	fg := lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground)
	accent := lipgloss.NewStyle()
	if len(m.cfg.Theme.Colors) > 0 {
		accent = lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[0])
	}

	lines := []string{
		"",
		"  " + bright.Render("Add Overlay") + dim.Render("  (enter=add  esc=cancel)"),
		"",
	}
	if m.ds != nil && len(m.ds.Headers) > 0 {
		lines = append(lines, "  "+dim.Render("Available columns:"))
		for i, h := range m.ds.Headers {
			marker := "  "
			if i == m.cfg.YColIdx {
				marker = dim.Render("* ")
			}
			lines = append(lines, fmt.Sprintf("    %s%s %s",
				marker, dim.Render(fmt.Sprintf("%d:", i)), accent.Render(h)))
		}
		lines = append(lines, "")
	}
	lines = append(lines, "  "+dim.Render("Column (name or index): ")+fg.Render(m.modal.input)+"_")
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

// --- Manage overlays modal ---

func (m Model) handleManageModalKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+c", "enter":
		m.modal.mode = modalNone
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
		if m.modal.selectIdx < len(m.cfg.Overlays) {
			m.cfg.Overlays = append(m.cfg.Overlays[:m.modal.selectIdx], m.cfg.Overlays[m.modal.selectIdx+1:]...)
			if m.modal.selectIdx >= len(m.cfg.Overlays) && m.modal.selectIdx > 0 {
				m.modal.selectIdx--
			}
			if len(m.cfg.Overlays) == 0 {
				m.modal.mode = modalNone
			}
			m.rebuildFromBuffer()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) manageModalView() string {
	dim := m.cfg.Theme.Axis
	bright := m.cfg.Theme.Title
	fg := lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground)

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

// --- Config modal (per-chart-type) ---

func (m Model) chartConfigFields() []string {
	return ConfigFields(m.cfg.ChartType)
}

func (m *Model) configSyncField() {
	fields := m.chartConfigFields()
	if m.modal.field >= len(fields) {
		return
	}
	name := fields[m.modal.field]
	opts := ConfigFieldOptions(name)
	if opts != nil {
		// Dropdown: find current value index
		cur := m.configFieldValue(name)
		m.modal.dropdownOpen = true
		m.modal.dropdownIdx = 0
		for i, o := range opts {
			if o == cur {
				m.modal.dropdownIdx = i
				break
			}
		}
		m.modal.input = ""
	} else {
		m.modal.dropdownOpen = false
		m.modal.input = m.configFieldValue(name)
	}
}

func (m Model) handleConfigModalKey(key string) (tea.Model, tea.Cmd) {
	fields := m.chartConfigFields()
	if len(fields) == 0 {
		m.modal.mode = modalNone
		return m, nil
	}

	name := fields[m.modal.field]
	opts := ConfigFieldOptions(name)
	isDropdown := opts != nil

	switch key {
	case "esc", "ctrl+c":
		m.modal.mode = modalNone
		return m, nil

	case "up", "k":
		if isDropdown {
			if m.modal.dropdownIdx > 0 {
				m.modal.dropdownIdx--
			}
		} else {
			if m.modal.field > 0 {
				m.modal.field--
				m.configSyncField()
			}
		}
		return m, nil

	case "down", "j":
		if isDropdown {
			if m.modal.dropdownIdx < len(opts)-1 {
				m.modal.dropdownIdx++
			}
		} else {
			if m.modal.field < len(fields)-1 {
				m.modal.field++
				m.configSyncField()
			}
		}
		return m, nil

	case "enter":
		if isDropdown {
			m.applyConfigField(name, opts[m.modal.dropdownIdx])
		} else {
			m.applyConfigField(name, m.modal.input)
		}
		m.rebuildFromBuffer()
		return m, nil

	case "tab":
		// Apply current and move to next field
		if isDropdown {
			m.applyConfigField(name, opts[m.modal.dropdownIdx])
		} else {
			m.applyConfigField(name, m.modal.input)
		}
		if m.modal.field < len(fields)-1 {
			m.modal.field++
			m.configSyncField()
		}
		m.rebuildFromBuffer()
		return m, nil

	case "backspace":
		if !isDropdown && len(m.modal.input) > 0 {
			m.modal.input = m.modal.input[:len(m.modal.input)-1]
		}
		return m, nil

	default:
		if !isDropdown && len(key) == 1 {
			m.modal.input += key
		}
		return m, nil
	}
}

func (m Model) configFieldValue(name string) string {
	switch name {
	case "ymin":
		if m.cfg.FixScale {
			return fmt.Sprintf("%.1f", m.cfg.YMin)
		}
		return "auto"
	case "ymax":
		if m.cfg.FixScale {
			return fmt.Sprintf("%.1f", m.cfg.YMax)
		}
		return "auto"
	case "pie-style":
		return m.cfg.PieStyle
	}
	return ""
}

func (m *Model) applyConfigField(name string, value string) {
	value = strings.TrimSpace(value)
	switch name {
	case "ymin":
		if value == "auto" || value == "" {
			m.cfg.FixScale = false
		} else if v, err := strconv.ParseFloat(value, 64); err == nil {
			m.cfg.YMin = v
			m.cfg.FixScale = true
		}
	case "ymax":
		if value == "auto" || value == "" {
			m.cfg.FixScale = false
		} else if v, err := strconv.ParseFloat(value, 64); err == nil {
			m.cfg.YMax = v
			m.cfg.FixScale = true
		}
	case "pie-style":
		if value == "bar" || value == "circle" {
			m.cfg.PieStyle = value
		}
	}
}

func (m Model) configModalView() string {
	fields := m.chartConfigFields()
	dim := m.cfg.Theme.Axis
	bright := m.cfg.Theme.Title
	fg := lipgloss.NewStyle().Foreground(m.cfg.Theme.Foreground)
	accent := lipgloss.NewStyle()
	if len(m.cfg.Theme.Colors) > 0 {
		accent = lipgloss.NewStyle().Foreground(m.cfg.Theme.Colors[0])
	}

	chartLabel := m.cfg.ChartType
	lines := []string{
		"",
		"  " + bright.Render("Configuration ["+chartLabel+"]") + dim.Render("  (enter=apply  tab=next  esc=close)"),
		"",
	}

	if len(fields) == 0 {
		lines = append(lines, "  "+dim.Render("No configuration options for this chart type."))
		lines = append(lines, "  "+dim.Render("Press esc to close."))
	} else {
		for i, name := range fields {
			marker := "  "
			isActive := i == m.modal.field
			if isActive {
				marker = fg.Render("> ")
			}

			opts := ConfigFieldOptions(name)
			if opts != nil && isActive {
				// Dropdown field — render as selectable list
				lines = append(lines, fmt.Sprintf("  %s%s",
					marker, dim.Render(fmt.Sprintf("%-12s", name+":"))))
				for j, opt := range opts {
					sel := "    "
					if j == m.modal.dropdownIdx {
						sel = "  " + accent.Render("▸ ")
					} else {
						sel = "    "
					}
					style := dim
					if j == m.modal.dropdownIdx {
						style = fg
					}
					lines = append(lines, sel+style.Render(opt))
				}
			} else {
				// Text input or inactive dropdown
				val := m.configFieldValue(name)
				if isActive && opts == nil {
					val = m.modal.input + "_"
				} else if opts != nil {
					// Show current value for inactive dropdown
					val = m.configFieldValue(name)
				}
				lines = append(lines, fmt.Sprintf("  %s%s  %s",
					marker, dim.Render(fmt.Sprintf("%-12s", name+":")), accent.Render(val)))
			}
		}
		lines = append(lines, "")
		hasScale := false
		for _, f := range fields {
			if f == "ymin" || f == "ymax" {
				hasScale = true
				break
			}
		}
		if hasScale {
			lines = append(lines, "  "+dim.Render("Type 'auto' for ymin/ymax to use auto-scaling"))
		}
	}

	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

// --- Resize / data ---

func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	reserved := 6
	if m.cfg.Live {
		reserved++
	}
	if len(m.cfg.Overlays) > 0 {
		reserved++
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
	}
	m.rebuildFromBuffer()

	return m, nil
}

func (m Model) handleData(msg dataMsg) (tea.Model, tea.Cmd) {
	dp := data.DataPoint(msg)
	m.stats.update(dp.Value, dp.Time)

	// Store in persistent buffer
	m.liveBuffer = append(m.liveBuffer, dp)

	// Track unique labels for multi-row overlay legend
	if dp.Label != "" {
		if m.liveLabelsAt == nil {
			m.liveLabelsAt = make(map[string]int)
		}
		if _, exists := m.liveLabelsAt[dp.Label]; !exists {
			m.liveLabelsAt[dp.Label] = len(m.liveLabels) + 1
			m.liveLabels = append(m.liveLabels, dp.Label)
		}
	}

	if m.renderer != nil {
		m.renderer.Push(dp)
	}

	if m.liveCh != nil {
		return m, waitForData(m.liveCh)
	}
	return m, nil
}

// rebuildFromBuffer recreates the renderer and replays all buffered data.
// This preserves data across theme/chart-type/config changes.
func (m *Model) rebuildFromBuffer() {
	if m.width == 0 || m.height == 0 {
		return
	}

	m.renderer = NewRenderer(m.cfg.ChartType)

	if m.cfg.Live && len(m.liveBuffer) > 0 {
		// For table chart type, build a DataSet from the buffer
		if m.cfg.ChartType == TypeTable {
			ds := m.dataSetFromBuffer()
			m.renderer.Setup(ds, m.cfg, m.width, m.height)
		} else {
			m.renderer.Setup(nil, m.cfg, m.width, m.height)
			for _, dp := range m.liveBuffer {
				m.renderer.Push(dp)
			}
		}
	} else {
		// Static mode — re-read from dataset
		m.renderer.Setup(m.ds, m.cfg, m.width, m.height)
	}
}

// dataSetFromBuffer builds a DataSet from the live buffer for renderers
// that need tabular data (e.g., table).
func (m *Model) dataSetFromBuffer() *data.DataSet {
	ds := &data.DataSet{
		Headers: []string{"Time", "Label", "Value"},
	}
	for _, dp := range m.liveBuffer {
		ts := dp.Time.Format("15:04:05")
		label := dp.Label
		if label == "" {
			label = "-"
		}
		ds.Rows = append(ds.Rows, []string{ts, label, fmt.Sprintf("%.2f", dp.Value)})
	}
	return ds
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
		s.Render("  c           ") + fg.Render("Open configuration"),
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

package chart

import (
	"testing"

	"github.com/socketspace-jihad/sow/internal/data"
	"github.com/socketspace-jihad/sow/internal/theme"
)

func testConfig() *Config {
	return &Config{
		ChartType: TypeBar,
		Theme:     theme.Get("default"),
		XColIdx:   0,
		YColIdx:   1,
	}
}

func testDataSet() *data.DataSet {
	return &data.DataSet{
		Headers: []string{"Name", "Value", "Extra"},
		Rows: [][]string{
			{"Alpha", "42", "10"},
			{"Beta", "87", "20"},
			{"Gamma", "23", "30"},
			{"Delta", "65", "40"},
		},
	}
}

func TestNewRendererAllTypes(t *testing.T) {
	for _, typ := range AllTypes() {
		r := NewRenderer(typ)
		if r == nil {
			t.Errorf("NewRenderer(%q) returned nil", typ)
		}
	}
}

func TestNewRendererUnknown(t *testing.T) {
	r := NewRenderer("nonexistent")
	if r == nil {
		t.Fatal("expected fallback renderer, got nil")
	}
}

func TestBarRendererSetup(t *testing.T) {
	r := NewRenderer(TypeBar)
	cfg := testConfig()
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("bar renderer produced empty view")
	}
}

func TestBarRendererNilDataSet(t *testing.T) {
	r := NewRenderer(TypeBar)
	cfg := testConfig()
	r.Setup(nil, cfg, 60, 20)
	// Should not panic
	view := r.View()
	_ = view
}

func TestLineRendererSetup(t *testing.T) {
	r := NewRenderer(TypeLine)
	cfg := testConfig()
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("line renderer produced empty view")
	}
}

func TestLineRendererNilDataSet(t *testing.T) {
	r := NewRenderer(TypeLine)
	cfg := testConfig()
	r.Setup(nil, cfg, 60, 20)
	// Should not panic
}

func TestSparklineRendererSetup(t *testing.T) {
	r := NewRenderer(TypeSparkline)
	cfg := testConfig()
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 10)
	view := r.View()
	if view == "" {
		t.Error("sparkline renderer produced empty view")
	}
}

func TestStreamlineRendererSetup(t *testing.T) {
	r := NewRenderer(TypeStreamline)
	cfg := testConfig()
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 10)
	view := r.View()
	if view == "" {
		t.Error("streamline renderer produced empty view")
	}
}

func TestPieRendererSetup(t *testing.T) {
	r := NewRenderer(TypePie)
	cfg := testConfig()
	cfg.PieStyle = "bar"
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("pie renderer produced empty view")
	}
}

func TestPieCircleMode(t *testing.T) {
	r := NewRenderer(TypePie)
	cfg := testConfig()
	cfg.PieStyle = "circle"
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("pie circle renderer produced empty view")
	}
}

func TestTableRendererSetup(t *testing.T) {
	r := NewRenderer(TypeTable)
	cfg := testConfig()
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("table renderer produced empty view")
	}
}

func TestHeatmapRendererSetup(t *testing.T) {
	r := NewRenderer(TypeHeatmap)
	cfg := testConfig()
	ds := &data.DataSet{
		Rows: [][]string{
			{"0", "0", "10"},
			{"1", "0", "20"},
			{"0", "1", "30"},
		},
	}
	r.Setup(ds, cfg, 40, 20)
	view := r.View()
	if view == "" {
		t.Error("heatmap renderer produced empty view")
	}
}

func TestRendererPush(t *testing.T) {
	dp := data.DataPoint{Label: "test", Value: 42.0}

	renderers := map[string]Renderer{
		"bar":        NewRenderer(TypeBar),
		"line":       NewRenderer(TypeLine),
		"scatter":    NewRenderer(TypeScatter),
		"sparkline":  NewRenderer(TypeSparkline),
		"streamline": NewRenderer(TypeStreamline),
		"pie":        NewRenderer(TypePie),
	}

	cfg := testConfig()
	for name, r := range renderers {
		r.Setup(nil, cfg, 60, 20)
		r.Push(dp) // Should not panic
		_ = r.View()
		t.Logf("%s: push OK", name)
	}
}

func TestRendererResize(t *testing.T) {
	cfg := testConfig()
	ds := testDataSet()

	for _, typ := range AllTypes() {
		r := NewRenderer(typ)
		r.Setup(ds, cfg, 60, 20)
		r.Resize(80, 30)
		_ = r.View()
	}
}

func TestOverlayOnLine(t *testing.T) {
	r := NewRenderer(TypeLine)
	cfg := testConfig()
	cfg.Overlays = []OverlayConfig{
		{Name: "Extra", ColIdx: 2, Color: 1},
	}
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("line with overlay produced empty view")
	}
}

func TestFixedScale(t *testing.T) {
	r := NewRenderer(TypeLine)
	cfg := testConfig()
	cfg.FixScale = true
	cfg.YMin = 0
	cfg.YMax = 100
	ds := testDataSet()
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" {
		t.Error("line with fixed scale produced empty view")
	}
}

func TestConfigFieldsPerChart(t *testing.T) {
	// Scale charts should have ymin/ymax/row-mode
	for _, ct := range []string{TypeBar, TypeHBar, TypeLine, TypeScatter, TypeSparkline, TypeStreamline} {
		fields := ConfigFields(ct)
		if len(fields) != 3 || fields[0] != "ymin" || fields[1] != "ymax" || fields[2] != "row-mode" {
			t.Errorf("ConfigFields(%q) = %v, want [ymin ymax row-mode]", ct, fields)
		}
	}

	// Pie should have pie-style only
	fields := ConfigFields(TypePie)
	if len(fields) != 1 || fields[0] != "pie-style" {
		t.Errorf("ConfigFields(pie) = %v, want [pie-style]", fields)
	}

	// Table and heatmap should have no config
	for _, ct := range []string{TypeTable, TypeHeatmap} {
		fields := ConfigFields(ct)
		if len(fields) != 0 {
			t.Errorf("ConfigFields(%q) = %v, want []", ct, fields)
		}
	}
}

func TestTableLiveBuffer(t *testing.T) {
	r := NewRenderer(TypeTable)
	cfg := testConfig()

	// Build a dataset from "live" data (simulating dataSetFromBuffer)
	ds := &data.DataSet{
		Headers: []string{"Time", "Label", "Value"},
		Rows: [][]string{
			{"12:00:01", "ping", "28.30"},
			{"12:00:02", "ping", "31.50"},
			{"12:00:03", "ping", "25.10"},
		},
	}
	r.Setup(ds, cfg, 60, 20)
	view := r.View()
	if view == "" || view == "  No data for table view" {
		t.Error("table with live buffer data should render content")
	}
}

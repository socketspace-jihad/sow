package data

import "testing"

func TestLiveParserSarDev(t *testing.T) {
	lp := NewLiveParser("", nil, 0, 0)

	lines := []string{
		"Linux 5.4.0-42-generic (hostname)     10/15/2026     _x86_64_    (8 CPU)",
		"",
		"02:30:01 PM     IFACE   rxpck/s   txpck/s    rxkB/s    txkB/s   rxcmp/s   txcmp/s  rxmcst/s   %ifutil",
		"02:30:02 PM      eth0   1234.00    567.00    890.12    345.67      0.00      0.00      0.00      1.23",
		"02:30:02 PM        lo     12.00     12.00      1.23      1.23      0.00      0.00      0.00      0.00",
	}

	var allPoints []DataPoint
	for i, line := range lines {
		points := lp.Feed(line)
		t.Logf("line %d: %q → %d points", i, line, len(points))
		for _, dp := range points {
			t.Logf("  label=%q value=%.2f", dp.Label, dp.Value)
		}
		allPoints = append(allPoints, points...)
	}

	cols := lp.Columns()
	t.Logf("detected columns: %v", cols)
	t.Logf("total points: %d", len(allPoints))

	if len(allPoints) == 0 {
		t.Fatal("expected data points from sar -n DEV output")
	}
}

func TestLiveParserFixedData(t *testing.T) {
	// Simulate: "command | sow" with no --live flag
	// But this goes through runStatic→readStdin→ParseText, not LiveParser
	// The user said "sar -n DEV 1 | sow" — without --live, it reads all stdin then parses as static

	// Test ParseText with sar-like output
	input := `Linux 5.4.0-42-generic (hostname)     10/15/2026     _x86_64_    (8 CPU)

02:30:01 PM     IFACE   rxpck/s   txpck/s    rxkB/s    txkB/s
02:30:02 PM      eth0   1234.00    567.00    890.12    345.67
02:30:02 PM        lo     12.00     12.00      1.23      1.23`

	ds, err := ParseText(input, "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("headers: %v", ds.Headers)
	t.Logf("rows: %d", len(ds.Rows))
	for i, row := range ds.Rows {
		t.Logf("  row %d: %v", i, row)
	}

	if len(ds.Rows) == 0 {
		t.Error("expected rows from sar output")
	}
}

func TestParseTextWithSkip(t *testing.T) {
	input := `Linux 5.4.0-42-generic (hostname)     10/15/2026     _x86_64_    (8 CPU)

02:30:01 PM     IFACE   rxpck/s   txpck/s    rxkB/s    txkB/s
02:30:02 PM      eth0   1234.00    567.00    890.12    345.67
02:30:02 PM        lo     12.00     12.00      1.23      1.23`

	// skip=2 skips "Linux..." and blank line, header is "IFACE rxpck/s..."
	ds, err := ParseText(input, "", 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("headers: %v", ds.Headers)
	t.Logf("rows: %d", len(ds.Rows))

	if len(ds.Headers) == 0 {
		t.Error("expected headers after skipping 2 lines")
	}
	// Should have IFACE, rxpck/s, etc. as headers (not Linux...)
	found := false
	for _, h := range ds.Headers {
		if h == "rxpck/s" || h == "IFACE" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rxpck/s or IFACE in headers, got %v", ds.Headers)
	}
	if len(ds.Rows) != 2 {
		t.Errorf("expected 2 data rows, got %d", len(ds.Rows))
	}
}

func TestLiveParserWithSkip(t *testing.T) {
	lp := NewLiveParser("", nil, 2, 0) // skip first 2 lines

	lines := []string{
		"Linux 5.4.0-42-generic (hostname)",
		"",
		"02:30:01 PM     IFACE   rxpck/s   txpck/s",
		"02:30:02 PM      eth0   1234.00    567.00",
	}

	var allPoints []DataPoint
	for _, line := range lines {
		points := lp.Feed(line)
		allPoints = append(allPoints, points...)
	}

	t.Logf("columns: %v", lp.Columns())
	t.Logf("points: %d", len(allPoints))

	if len(allPoints) == 0 {
		t.Error("expected data points after skip")
	}
}

func TestLiveParserRowStream(t *testing.T) {
	// mpstat -P ALL 1: --row-stream 3 groups 3 data lines per sample
	lp := NewLiveParser("%usr", nil, 0, 3)

	lines := []string{
		"11:48:56     CPU    %usr   %nice    %sys %iowait",
		"11:48:57     all    0.51    0.00    0.00    0.00",
		"11:48:57       0    0.00    0.00    0.00    0.00",
		"11:48:57       1    1.01    0.00    0.00    0.00",
		"11:48:58     all    2.00    0.00    1.00    0.00",
		"11:48:58       0    3.00    0.00    0.00    0.00",
		"11:48:58       1    1.00    0.00    2.00    0.00",
	}

	var allPoints []DataPoint
	for _, line := range lines {
		points := lp.Feed(line)
		for _, p := range points {
			t.Logf("  label=%q value=%.2f", p.Label, p.Value)
		}
		allPoints = append(allPoints, points...)
	}

	t.Logf("columns: %v", lp.Columns())
	t.Logf("total points: %d", len(allPoints))

	if len(allPoints) != 6 {
		t.Errorf("expected 6 points (2 samples x 3 rows), got %d", len(allPoints))
	}

	labels := make(map[string]bool)
	for _, dp := range allPoints {
		labels[dp.Label] = true
	}
	t.Logf("unique labels: %v", labels)
	if len(labels) != 3 {
		t.Errorf("expected 3 labels (all, 0, 1), got %d: %v", len(labels), labels)
	}
}

package data

import "testing"

func TestLiveParserPidstat(t *testing.T) {
	lp := NewLiveParser("", nil, 0, 0)

	// pidstat -r output
	lines := []string{
		"Linux 5.4.0-42-generic (host) \t10/15/2026 \t_x86_64_\t(8 CPU)",
		"",
		"02:30:01 PM   UID       PID  minflt/s  majflt/s     VSZ     RSS   %MEM  Command",
		"02:30:02 PM  1000     34405     12.00      0.00  123456   45678   1.23  myapp",
		"02:30:03 PM  1000     34405     15.00      0.00  123460   45700   1.24  myapp",
	}

	var allPoints []DataPoint
	for _, line := range lines {
		points := lp.Feed(line)
		allPoints = append(allPoints, points...)
	}

	cols := lp.Columns()
	if len(cols) == 0 {
		t.Fatal("expected columns to be detected")
	}

	// Should have detected columns like UID, PID, minflt/s, etc.
	t.Logf("detected columns: %v", cols)
	t.Logf("total points: %d", len(allPoints))

	if len(allPoints) == 0 {
		t.Fatal("expected data points")
	}

	// Should have multiple columns per data line, not just %MEM
	labels := make(map[string]bool)
	for _, dp := range allPoints {
		labels[dp.Label] = true
	}
	t.Logf("unique labels: %v", labels)

	if len(labels) < 3 {
		t.Errorf("expected multiple columns as labels, got only %d: %v", len(labels), labels)
	}
}

func TestLiveParserPidstatWithY(t *testing.T) {
	lp := NewLiveParser("VSZ", nil, 0, 0)

	lines := []string{
		"02:30:01 PM   UID       PID  minflt/s  majflt/s     VSZ     RSS   %MEM  Command",
		"02:30:02 PM  1000     34405     12.00      0.00  123456   45678   1.23  myapp",
	}

	var allPoints []DataPoint
	for _, line := range lines {
		points := lp.Feed(line)
		allPoints = append(allPoints, points...)
	}

	if len(allPoints) != 1 {
		t.Fatalf("expected 1 point for --y VSZ, got %d", len(allPoints))
	}
	if allPoints[0].Label != "VSZ" {
		t.Errorf("expected label 'VSZ', got %q", allPoints[0].Label)
	}
	if allPoints[0].Value != 123456 {
		t.Errorf("expected 123456, got %f", allPoints[0].Value)
	}
}

func TestLiveParserSingleNumbers(t *testing.T) {
	lp := NewLiveParser("", nil, 0, 0)

	points := lp.Feed("42.7")
	// Might be buffered waiting for header detection
	points = append(points, lp.Feed("38.2")...)
	points = append(points, lp.Feed("50.1")...)
	points = append(points, lp.Feed("45.0")...)
	points = append(points, lp.Feed("32.5")...)

	if len(points) == 0 {
		t.Error("expected data points for plain numbers")
	}
}

func TestLiveParserPingPassthrough(t *testing.T) {
	lp := NewLiveParser("", nil, 0, 0)

	// Ping header
	points := lp.Feed("PING google.com (142.250.185.46) 56(84) bytes of data.")
	if len(points) != 0 {
		t.Error("ping header should not produce data points")
	}

	info := lp.GetStreamInfo()
	if info == nil {
		t.Fatal("expected stream info for ping")
	}
	if info.Source != "ping" {
		t.Errorf("expected source 'ping', got %q", info.Source)
	}

	// Ping data line — should fall through to ParseLine
	points = lp.Feed("64 bytes from lax17s55-in-f14.1e100.net: icmp_seq=1 ttl=117 time=28.3 ms")
	if len(points) == 0 {
		t.Error("expected data point for ping data line")
	}
}

func TestLiveParserVmstat(t *testing.T) {
	lp := NewLiveParser("", nil, 0, 0)

	lines := []string{
		"procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----",
		" r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st",
		" 1  0      0 1234567  12345 678901    0    0    12    34  567  890  5  2 93  0  0",
	}

	var allPoints []DataPoint
	for _, line := range lines {
		points := lp.Feed(line)
		allPoints = append(allPoints, points...)
	}

	if len(allPoints) == 0 {
		t.Fatal("expected data points for vmstat")
	}

	labels := make(map[string]bool)
	for _, dp := range allPoints {
		labels[dp.Label] = true
	}
	t.Logf("vmstat labels: %v", labels)

	// Should have multiple columns: r, b, swpd, free, etc.
	if len(labels) < 5 {
		t.Errorf("expected many vmstat columns, got %d", len(labels))
	}
}

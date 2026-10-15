package data

import "testing"

func TestLiveParserWithHeader(t *testing.T) {
	// pidstat -r: --skip 2 --header (skip "Linux..." and blank line, next line is header)
	lp := NewLiveParser("", nil, 2, 0, true, nil, nil)

	lines := []string{
		"Linux 5.4.0-42-generic (host) \t10/15/2026",
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
	t.Logf("columns: %v", cols)
	t.Logf("total points: %d", len(allPoints))

	if len(cols) == 0 {
		t.Fatal("expected columns from --header")
	}
	if len(allPoints) == 0 {
		t.Fatal("expected data points")
	}

	labels := make(map[string]bool)
	for _, dp := range allPoints {
		labels[dp.Label] = true
	}
	t.Logf("labels: %v", labels)
	if len(labels) < 3 {
		t.Errorf("expected multiple column labels, got %d", len(labels))
	}
}

func TestLiveParserWithHeaderAndY(t *testing.T) {
	// --header --y VSZ
	lp := NewLiveParser("VSZ", nil, 0, 0, true, nil, nil)

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

func TestLiveParserAddHeader(t *testing.T) {
	// --add-header name --add-header value (no header line, all data)
	lp := NewLiveParser("value", nil, 0, 0, false, []string{"name", "value"}, nil)

	lines := []string{
		"alpha 10",
		"beta 20",
		"gamma 30",
	}

	var allPoints []DataPoint
	for _, line := range lines {
		points := lp.Feed(line)
		allPoints = append(allPoints, points...)
	}

	t.Logf("columns: %v", lp.Columns())
	if len(allPoints) != 3 {
		t.Fatalf("expected 3 points, got %d", len(allPoints))
	}
	if allPoints[0].Value != 10 {
		t.Errorf("expected 10, got %f", allPoints[0].Value)
	}
}

func TestLiveParserNoHeader(t *testing.T) {
	// No --header, no --add-header → fallback to ParseLine
	lp := NewLiveParser("", nil, 0, 0, false, nil, nil)

	points := lp.Feed("42.7")
	points = append(points, lp.Feed("38.2")...)
	points = append(points, lp.Feed("50.1")...)

	if len(points) != 3 {
		t.Errorf("expected 3 points from fallback, got %d", len(points))
	}
}

func TestLiveParserPingPassthrough(t *testing.T) {
	lp := NewLiveParser("", nil, 0, 0, false, nil, nil)

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

	points = lp.Feed("64 bytes from lax17s55-in-f14.1e100.net: icmp_seq=1 ttl=117 time=28.3 ms")
	if len(points) == 0 {
		t.Error("expected data point for ping data line")
	}
}

func TestLiveParserVmstatWithHeader(t *testing.T) {
	// vmstat: --skip 1 --header (skip "procs..." line, next is column names)
	lp := NewLiveParser("", nil, 1, 0, true, nil, nil)

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

	t.Logf("columns: %v", lp.Columns())
	t.Logf("total points: %d", len(allPoints))

	if len(allPoints) == 0 {
		t.Fatal("expected data points for vmstat with --header")
	}

	labels := make(map[string]bool)
	for _, dp := range allPoints {
		labels[dp.Label] = true
	}
	t.Logf("labels: %v", labels)
	if len(labels) < 5 {
		t.Errorf("expected many vmstat columns, got %d", len(labels))
	}
}

package data

import (
	"testing"
)

// --- ParseLine tests ---

func TestParseLinePing(t *testing.T) {
	dp := ParseLine("64 bytes from lax17s55-in-f14.1e100.net (142.250.185.206): icmp_seq=1 ttl=117 time=28.3 ms")
	if dp == nil {
		t.Fatal("expected DataPoint, got nil")
	}
	if dp.Value != 28.3 {
		t.Errorf("expected 28.3, got %f", dp.Value)
	}
	if dp.Label != "ping" {
		t.Errorf("expected label 'ping', got %q", dp.Label)
	}
}

func TestParseLineSingleNumber(t *testing.T) {
	dp := ParseLine("42.7")
	if dp == nil {
		t.Fatal("expected DataPoint, got nil")
	}
	if dp.Value != 42.7 {
		t.Errorf("expected 42.7, got %f", dp.Value)
	}
}

func TestParseLineInteger(t *testing.T) {
	dp := ParseLine("100")
	if dp == nil {
		t.Fatal("expected DataPoint, got nil")
	}
	if dp.Value != 100 {
		t.Errorf("expected 100, got %f", dp.Value)
	}
}

func TestParseLineLabeledValue(t *testing.T) {
	dp := ParseLine("cpu_idle 93.5")
	if dp == nil {
		t.Fatal("expected DataPoint, got nil")
	}
	if dp.Label != "cpu_idle" {
		t.Errorf("expected label 'cpu_idle', got %q", dp.Label)
	}
	if dp.Value != 93.5 {
		t.Errorf("expected 93.5, got %f", dp.Value)
	}
}

func TestParseLineEmpty(t *testing.T) {
	if dp := ParseLine(""); dp != nil {
		t.Errorf("expected nil for empty line, got %+v", dp)
	}
	if dp := ParseLine("   "); dp != nil {
		t.Errorf("expected nil for whitespace line, got %+v", dp)
	}
}

func TestParseLineNoNumber(t *testing.T) {
	if dp := ParseLine("hello world"); dp != nil {
		t.Errorf("expected nil for non-numeric line, got %+v", dp)
	}
}

// --- ParseText tests ---

func TestParseTextCSV(t *testing.T) {
	ds, err := ParseText("name,val\nalpha,10\nbeta,20\ngamma,30", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d: %v", len(ds.Headers), ds.Headers)
	}
	if ds.Headers[0] != "name" || ds.Headers[1] != "val" {
		t.Errorf("unexpected headers: %v", ds.Headers)
	}
	if len(ds.Rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(ds.Rows))
	}
}

func TestParseTextSingleColumn(t *testing.T) {
	ds, err := ParseText("5\n3\n8\n12", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) != 0 {
		t.Errorf("expected no headers for pure numbers, got %v", ds.Headers)
	}
	if ds.NumColumns() != 1 {
		t.Errorf("expected 1 column, got %d", ds.NumColumns())
	}
	if len(ds.Rows) != 4 {
		t.Errorf("expected 4 rows, got %d", len(ds.Rows))
	}
}

func TestParseTextMixedNotHeader(t *testing.T) {
	// "Jan 12500" has text + number, should NOT be treated as header
	ds, err := ParseText("Jan 12500\nFeb 15300\nMar 11800", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) > 0 {
		t.Errorf("mixed text+number row should not be header, got %v", ds.Headers)
	}
	if len(ds.Rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(ds.Rows))
	}
}

func TestParseTextTSVWithHeader(t *testing.T) {
	ds, err := ParseText("host\tlatency\ngoogle.com\t28.5\ncloudflare.com\t15.2", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) != 2 {
		t.Errorf("expected 2 headers, got %v", ds.Headers)
	}
	if len(ds.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(ds.Rows))
	}
}

func TestParseTextForcedHeader(t *testing.T) {
	ds, err := ParseText("10,20\n30,40", "", 1, 0) // headerMode=1 forces header
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) != 2 {
		t.Errorf("expected 2 headers when forced, got %v", ds.Headers)
	}
	if len(ds.Rows) != 1 {
		t.Errorf("expected 1 data row, got %d", len(ds.Rows))
	}
}

func TestParseTextForcedNoHeader(t *testing.T) {
	ds, err := ParseText("name,val\nalpha,10", "", -1, 0) // headerMode=-1 forces no header
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) != 0 {
		t.Errorf("expected no headers when forced, got %v", ds.Headers)
	}
	if len(ds.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(ds.Rows))
	}
}

func TestParseTextEmpty(t *testing.T) {
	ds, err := ParseText("", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(ds.Rows))
	}
}

// --- DetectStream tests ---

func TestDetectStreamPing(t *testing.T) {
	info := DetectStream("PING google.com (142.250.185.46) 56(84) bytes of data.")
	if info == nil {
		t.Fatal("expected StreamInfo for ping header")
	}
	if info.Source != "ping" {
		t.Errorf("expected source 'ping', got %q", info.Source)
	}
	if info.Host != "google.com" {
		t.Errorf("expected host 'google.com', got %q", info.Host)
	}
	if info.Unit != "ms" {
		t.Errorf("expected unit 'ms', got %q", info.Unit)
	}
}

func TestDetectStreamNotPing(t *testing.T) {
	if info := DetectStream("64 bytes from google.com: time=28.3 ms"); info != nil {
		t.Errorf("expected nil for ping data line, got %+v", info)
	}
	if info := DetectStream("42"); info != nil {
		t.Errorf("expected nil for plain number, got %+v", info)
	}
	if info := DetectStream("procs -----------memory----------"); info != nil {
		t.Errorf("expected nil for vmstat line (no auto-detect), got %+v", info)
	}
}

// --- DataSet tests ---

func TestNumericColumn(t *testing.T) {
	ds := &DataSet{
		Rows: [][]string{{"10"}, {"20.5"}, {"30"}},
	}
	vals := ds.NumericColumn(0)
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	if vals[0] != 10 || vals[1] != 20.5 || vals[2] != 30 {
		t.Errorf("unexpected values: %v", vals)
	}
}

func TestNumericColumnOutOfBounds(t *testing.T) {
	ds := &DataSet{
		Rows: [][]string{{"10"}, {"20"}},
	}
	vals := ds.NumericColumn(5) // column 5 doesn't exist
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}
	// Should return 0 for out-of-bounds
	if vals[0] != 0 || vals[1] != 0 {
		t.Errorf("expected zeros for out-of-bounds, got %v", vals)
	}
}

func TestResolveColumn(t *testing.T) {
	ds := &DataSet{
		Headers: []string{"Month", "Revenue", "Expenses"},
	}
	if ds.ResolveColumn("Revenue") != 1 {
		t.Errorf("expected 1 for 'Revenue', got %d", ds.ResolveColumn("Revenue"))
	}
	if ds.ResolveColumn("2") != 2 {
		t.Errorf("expected 2 for '2', got %d", ds.ResolveColumn("2"))
	}
	if ds.ResolveColumn("0") != 0 {
		t.Errorf("expected 0 for '0', got %d", ds.ResolveColumn("0"))
	}
}

func TestNumColumns(t *testing.T) {
	ds1 := &DataSet{Headers: []string{"a", "b", "c"}}
	if ds1.NumColumns() != 3 {
		t.Errorf("expected 3 from headers, got %d", ds1.NumColumns())
	}
	ds2 := &DataSet{Rows: [][]string{{"1", "2"}}}
	if ds2.NumColumns() != 2 {
		t.Errorf("expected 2 from rows, got %d", ds2.NumColumns())
	}
	ds3 := &DataSet{}
	if ds3.NumColumns() != 0 {
		t.Errorf("expected 0 for empty, got %d", ds3.NumColumns())
	}
}

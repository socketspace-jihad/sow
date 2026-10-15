package data

import "testing"

func TestPatternKeyValue(t *testing.T) {
	pp, err := NewPatternParser("{label}={value}")
	if err != nil {
		t.Fatal(err)
	}

	dp := pp.ParseLine("response_time=45.2")
	if dp == nil {
		t.Fatal("expected match")
	}
	if dp.Label != "response_time" {
		t.Errorf("expected label 'response_time', got %q", dp.Label)
	}
	if dp.Value != 45.2 {
		t.Errorf("expected value 45.2, got %f", dp.Value)
	}
}

func TestPatternCountMode(t *testing.T) {
	pp, err := NewPatternParser("{label}={1}")
	if err != nil {
		t.Fatal(err)
	}

	// "user=client" — "client" is not a number, so {1} won't match
	// This pattern requires {1} to be numeric, so "client" won't match
	dp := pp.ParseLine("user=42")
	if dp == nil {
		t.Fatal("expected match for numeric value")
	}
	if dp.Label != "user" {
		t.Errorf("expected label 'user', got %q", dp.Label)
	}
	if dp.Value != 42 {
		t.Errorf("expected value 42, got %f", dp.Value)
	}
}

func TestPatternLabelOnly(t *testing.T) {
	pp, err := NewPatternParser("ERROR: {label}")
	if err != nil {
		t.Fatal(err)
	}

	dp1 := pp.ParseLine("ERROR: timeout")
	if dp1 == nil {
		t.Fatal("expected match")
	}
	if dp1.Label != "timeout" {
		t.Errorf("expected label 'timeout', got %q", dp1.Label)
	}
	if dp1.Value != 1 {
		t.Errorf("expected count 1, got %f", dp1.Value)
	}

	dp2 := pp.ParseLine("ERROR: timeout")
	if dp2.Value != 2 {
		t.Errorf("expected count 2, got %f", dp2.Value)
	}

	dp3 := pp.ParseLine("ERROR: connection")
	if dp3.Label != "connection" || dp3.Value != 1 {
		t.Errorf("expected connection/1, got %s/%f", dp3.Label, dp3.Value)
	}
}

func TestPatternNoMatch(t *testing.T) {
	pp, err := NewPatternParser("status={value}")
	if err != nil {
		t.Fatal(err)
	}

	dp := pp.ParseLine("this line has no status field")
	if dp != nil {
		t.Errorf("expected nil for non-matching line, got %+v", dp)
	}
}

func TestPatternEmpty(t *testing.T) {
	pp, err := NewPatternParser("{label}={value}")
	if err != nil {
		t.Fatal(err)
	}
	if dp := pp.ParseLine(""); dp != nil {
		t.Error("expected nil for empty line")
	}
}

func TestPatternComplex(t *testing.T) {
	pp, err := NewPatternParser("host={a} latency={value}ms")
	if err != nil {
		t.Fatal(err)
	}

	dp := pp.ParseLine("host=web01 latency=12.5ms")
	if dp == nil {
		t.Fatal("expected match")
	}
	if dp.Label != "web01" {
		t.Errorf("expected label 'web01', got %q", dp.Label)
	}
	if dp.Value != 12.5 {
		t.Errorf("expected value 12.5, got %f", dp.Value)
	}
}

func TestPatternMultiLabel(t *testing.T) {
	pp, err := NewPatternParser("{a}.{b}={value}")
	if err != nil {
		t.Fatal(err)
	}

	dp := pp.ParseLine("server.cpu=87.5")
	if dp == nil {
		t.Fatal("expected match")
	}
	if dp.Label != "server:cpu" {
		t.Errorf("expected label 'server:cpu', got %q", dp.Label)
	}
	if dp.Value != 87.5 {
		t.Errorf("expected 87.5, got %f", dp.Value)
	}
}

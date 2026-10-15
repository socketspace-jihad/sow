package chart

import "testing"

func TestNextType(t *testing.T) {
	tests := []struct {
		current  string
		expected string
	}{
		{"bar", "hbar"},
		{"table", "bar"}, // wraps around
		{"line", "scatter"},
		{"unknown", "bar"}, // fallback
	}
	for _, tt := range tests {
		got := NextType(tt.current)
		if got != tt.expected {
			t.Errorf("NextType(%q) = %q, want %q", tt.current, got, tt.expected)
		}
	}
}

func TestPrevType(t *testing.T) {
	tests := []struct {
		current  string
		expected string
	}{
		{"hbar", "bar"},
		{"bar", "table"}, // wraps around
		{"scatter", "line"},
		{"unknown", "bar"}, // fallback
	}
	for _, tt := range tests {
		got := PrevType(tt.current)
		if got != tt.expected {
			t.Errorf("PrevType(%q) = %q, want %q", tt.current, got, tt.expected)
		}
	}
}

func TestAllTypesComplete(t *testing.T) {
	types := AllTypes()
	expected := map[string]bool{
		"bar": true, "hbar": true, "line": true, "scatter": true,
		"sparkline": true, "streamline": true, "heatmap": true,
		"pie": true, "table": true,
	}
	if len(types) != len(expected) {
		t.Errorf("expected %d types, got %d", len(expected), len(types))
	}
	for _, typ := range types {
		if !expected[typ] {
			t.Errorf("unexpected type: %q", typ)
		}
	}
}

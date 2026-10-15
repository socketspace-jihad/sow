package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCSVFile(t *testing.T) {
	// Find the sales.csv example
	dir, _ := os.Getwd()
	csvPath := filepath.Join(dir, "..", "..", "examples", "sales.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("examples/sales.csv not found")
	}

	ds, err := ReadCSV(csvPath, ',', 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(ds.Headers) != 4 {
		t.Errorf("expected 4 headers, got %d: %v", len(ds.Headers), ds.Headers)
	}
	if ds.Headers[0] != "Month" {
		t.Errorf("expected first header 'Month', got %q", ds.Headers[0])
	}
	if len(ds.Rows) != 12 {
		t.Errorf("expected 12 rows, got %d", len(ds.Rows))
	}
}

func TestReadCSVAutoDetect(t *testing.T) {
	dir, _ := os.Getwd()
	csvPath := filepath.Join(dir, "..", "..", "examples", "sales.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("examples/sales.csv not found")
	}

	ds, err := ReadCSVAuto(csvPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Headers) < 2 {
		t.Errorf("expected headers, got %v", ds.Headers)
	}
	if len(ds.Rows) < 1 {
		t.Error("expected rows")
	}
}

func TestReadCSVTempTSV(t *testing.T) {
	content := "name\tvalue\nalpha\t10\nbeta\t20\n"
	f, err := os.CreateTemp("", "sow-test-*.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	ds, err := ReadCSV(f.Name(), '\t', 0)
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

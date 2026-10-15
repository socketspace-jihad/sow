package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// ReadCSV reads a CSV or TSV file with a known delimiter.
func ReadCSV(path string, delim rune, headerMode int) (*DataSet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = delim
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1 // Allow variable field counts

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if len(records) == 0 {
		return &DataSet{}, nil
	}

	ds := &DataSet{}
	hasHeader := headerMode == 1 || (headerMode == 0 && detectHeader(records[0]))
	if hasHeader && len(records) > 1 {
		ds.Headers = records[0]
		ds.Rows = records[1:]
	} else {
		ds.Rows = records
	}

	return ds, nil
}

// ReadCSVAuto reads a file with auto-detected delimiter.
func ReadCSVAuto(path string, headerMode int) (*DataSet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	// Read a sample to detect delimiter
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	sample := string(buf[:n])

	// Reset file
	f.Seek(0, 0)

	delim := detectDelimiter(strings.Split(sample, "\n")[0])

	switch delim {
	case ",":
		return readCSVFromFile(f, ',', headerMode)
	case "\t":
		return readCSVFromFile(f, '\t', headerMode)
	default:
		// Whitespace-separated — read as text
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return ParseText(string(raw), " ", headerMode)
	}
}

func readCSVFromFile(f *os.File, delim rune, headerMode int) (*DataSet, error) {
	r := csv.NewReader(f)
	r.Comma = delim
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &DataSet{}, nil
	}

	ds := &DataSet{}
	hasHeader := headerMode == 1 || (headerMode == 0 && detectHeader(records[0]))
	if hasHeader && len(records) > 1 {
		ds.Headers = records[0]
		ds.Rows = records[1:]
	} else {
		ds.Rows = records
	}

	return ds, nil
}

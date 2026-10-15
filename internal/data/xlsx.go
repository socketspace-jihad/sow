package data

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ReadXLSX reads an Excel .xlsx file.
func ReadXLSX(path string, sheet string) (*DataSet, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening xlsx %s: %w", path, err)
	}
	defer f.Close()

	// Use specified sheet or first sheet
	sheetName := sheet
	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found in %s", path)
		}
		sheetName = sheets[0]
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("reading sheet %q: %w", sheetName, err)
	}

	if len(rows) == 0 {
		return &DataSet{}, nil
	}

	ds := &DataSet{}
	// First row is header if it contains non-numeric values
	if detectHeader(rows[0]) && len(rows) > 1 {
		ds.Headers = rows[0]
		ds.Rows = rows[1:]
	} else {
		ds.Rows = rows
	}

	return ds, nil
}

package data

import "time"

// DataSet holds static tabular data.
type DataSet struct {
	Headers []string
	Rows    [][]string
}

// DataPoint represents a single live data point.
type DataPoint struct {
	Label string
	Value float64
	Time  time.Time
}

// StreamInfo holds auto-detected metadata about a live data stream.
type StreamInfo struct {
	Source  string // detected source tool: "ping", "vmstat", "iostat", ""
	Host   string // for ping: the target host
	XLabel string // what the X axis represents
	YLabel string // what the Y axis represents
	Unit   string // unit of Y values: "ms", "%", "KB/s", etc.
}

// NumColumns returns the number of columns in the data.
func (ds *DataSet) NumColumns() int {
	if len(ds.Headers) > 0 {
		return len(ds.Headers)
	}
	if len(ds.Rows) > 0 {
		return len(ds.Rows[0])
	}
	return 0
}

// NumericColumn extracts a column as float64 values.
// Returns the values and labels (from another column) for chart rendering.
func (ds *DataSet) NumericColumn(col int) []float64 {
	vals := make([]float64, 0, len(ds.Rows))
	for _, row := range ds.Rows {
		if col >= len(row) {
			vals = append(vals, 0)
			continue
		}
		v := parseFloat(row[col])
		vals = append(vals, v)
	}
	return vals
}

// StringColumn extracts a column as string labels.
func (ds *DataSet) StringColumn(col int) []string {
	labels := make([]string, 0, len(ds.Rows))
	for _, row := range ds.Rows {
		if col >= len(row) {
			labels = append(labels, "")
			continue
		}
		labels = append(labels, row[col])
	}
	return labels
}

// ResolveColumn resolves a column name or index string to a column index.
func (ds *DataSet) ResolveColumn(col string) int {
	// Try matching by header name first
	for i, h := range ds.Headers {
		if h == col {
			return i
		}
	}
	// Try parsing as integer index
	idx := 0
	for _, c := range col {
		if c >= '0' && c <= '9' {
			idx = idx*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return idx
}

func parseFloat(s string) float64 {
	var result float64
	var decimal float64
	neg := false
	inDecimal := false

	for i, c := range s {
		if c == '-' && i == 0 {
			neg = true
			continue
		}
		if c == '.' && !inDecimal {
			inDecimal = true
			decimal = 10
			continue
		}
		if c >= '0' && c <= '9' {
			if inDecimal {
				result += float64(c-'0') / decimal
				decimal *= 10
			} else {
				result = result*10 + float64(c-'0')
			}
		}
	}
	if neg {
		result = -result
	}
	return result
}

package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LiveParser is a stateful parser for live stdin that auto-detects
// tabular output (column headers + data rows) from piped commands.
//
// When tabular headers are detected, each numeric column becomes
// a separate labeled DataPoint (auto-overlay). The --y flag selects
// a specific column; without it all numeric columns are emitted.
//
// Falls back to simple ParseLine for non-tabular input.
type LiveParser struct {
	columns       []string // detected column names
	headerDone    bool
	pendingLines  []string // buffered lines before header detection
	lineCount     int
	yCol          string // user's --y selection (column name or index)
	skipLines     int    // skip first N lines
	rowStream     int    // group every N data lines as separate series
	rowBuffer     []string // buffer for row-stream grouping
	patternParser *PatternParser
	streamInfo    *StreamInfo
}

// NewLiveParser creates a parser with optional --y column, --pattern, --skip, and --row-stream.
func NewLiveParser(yCol string, pattern *PatternParser, skip int, rowStream int) *LiveParser {
	return &LiveParser{
		yCol:          yCol,
		skipLines:     skip,
		rowStream:     rowStream,
		patternParser: pattern,
	}
}

// StreamInfo returns detected stream metadata, if any.
func (lp *LiveParser) GetStreamInfo() *StreamInfo {
	return lp.streamInfo
}

// Columns returns detected column names (empty if not tabular).
func (lp *LiveParser) Columns() []string {
	return lp.columns
}

// Feed processes one line and returns zero or more DataPoints.
// Multiple points are returned when multiple columns are emitted (auto-overlay).
func (lp *LiveParser) Feed(line string) []DataPoint {
	lp.lineCount++

	// Skip first N lines
	if lp.lineCount <= lp.skipLines {
		return nil
	}

	trimmed := strings.TrimSpace(line)

	// Pattern parser takes priority
	if lp.patternParser != nil {
		dp := lp.patternParser.ParseLine(line)
		if dp != nil {
			dp.Time = time.Now()
			return []DataPoint{*dp}
		}
		return nil
	}

	// Try stream detection on early lines (only the first match)
	if lp.lineCount <= 5 && !lp.headerDone && lp.streamInfo == nil {
		if info := DetectStream(trimmed); info != nil {
			lp.streamInfo = info
			// For tools like ping that have no tabular columns,
			// mark header done so subsequent lines go to fallback
			if info.Source == "ping" {
				lp.headerDone = true
			}
			return nil // skip info line
		}
	}

	// Try tabular header detection
	if !lp.headerDone {
		if lp.isHeaderLine(trimmed) {
			lp.columns = lp.extractColumns(trimmed)
			lp.headerDone = true
			// Replay any pending lines as data
			var results []DataPoint
			for _, pl := range lp.pendingLines {
				results = append(results, lp.parseDataLine(pl)...)
			}
			lp.pendingLines = nil
			return results
		}

		// Check if this looks like a data line (has numbers)
		if lp.hasNumericFields(trimmed) && lp.lineCount <= 10 {
			// Buffer it — might get a header on a later line
			// (some tools print a blank line or info line before headers)
			lp.pendingLines = append(lp.pendingLines, trimmed)

			// If we've buffered too many, give up on header detection
			if len(lp.pendingLines) >= 5 {
				lp.headerDone = true
				var results []DataPoint
				for _, pl := range lp.pendingLines {
					results = append(results, lp.parseFallback(pl)...)
				}
				lp.pendingLines = nil
				return results
			}
			return nil
		}

		// Non-numeric, non-header line (info line) — skip
		if !lp.hasNumericFields(trimmed) {
			return nil
		}
	}

	// Header already detected — parse as data line
	if len(lp.columns) > 0 {
		if lp.rowStream > 0 {
			return lp.feedRowStream(trimmed)
		}
		return lp.parseDataLine(trimmed)
	}

	// No header — fallback to simple parsing
	if lp.rowStream > 0 {
		return lp.feedRowStream(trimmed)
	}
	return lp.parseFallback(trimmed)
}

// feedRowStream buffers data lines and emits N DataPoints when the buffer is full.
func (lp *LiveParser) feedRowStream(line string) []DataPoint {
	if !lp.hasNumericFields(line) {
		return nil
	}

	lp.rowBuffer = append(lp.rowBuffer, line)
	if len(lp.rowBuffer) < lp.rowStream {
		return nil
	}

	now := time.Now()
	var results []DataPoint

	// Parse all buffered rows into field arrays
	allFields := make([][]string, len(lp.rowBuffer))
	for i, rowLine := range lp.rowBuffer {
		allFields[i] = lp.stripTimestamp(strings.Fields(rowLine))
	}

	// Find the label column: the column whose values differ across the N rows.
	// This identifies the row (e.g., CPU=all/0/1, IFACE=eth0/lo).
	labelColIdx := lp.findLabelColumn(allFields)

	for i, fields := range allFields {
		// Derive label
		label := fmt.Sprintf("row-%d", i)
		if labelColIdx >= 0 && labelColIdx < len(fields) {
			label = fields[labelColIdx]
		}

		// Get value using column alignment
		value := 0.0
		if len(lp.columns) > 0 {
			values := lp.alignFieldsToColumns(strings.Fields(lp.rowBuffer[i]))
			if lp.yCol != "" && lp.yCol != "0" && lp.yCol != "1" {
				if v, ok := values[lp.yCol]; ok {
					value = v
				}
			} else {
				// Default: first numeric column
				for _, col := range lp.columns {
					if v, ok := values[col]; ok {
						value = v
						break
					}
				}
			}
		} else {
			dp := ParseLine(lp.rowBuffer[i])
			if dp != nil {
				value = dp.Value
			}
		}

		results = append(results, DataPoint{
			Label: label,
			Value: value,
			Time:  now,
			Row:   float64(i),
		})
	}

	lp.rowBuffer = lp.rowBuffer[:0]
	return results
}

// stripTimestamp removes leading timestamp fields (HH:MM:SS, AM/PM).
func (lp *LiveParser) stripTimestamp(fields []string) []string {
	var result []string
	for _, f := range fields {
		if f == "AM" || f == "PM" {
			continue
		}
		if len(f) == 8 && f[2] == ':' && f[5] == ':' {
			continue
		}
		result = append(result, f)
	}
	return result
}

// findLabelColumn finds which column position differs across the N rows,
// making it the row identifier (e.g., CPU number, interface name).
func (lp *LiveParser) findLabelColumn(allFields [][]string) int {
	if len(allFields) < 2 {
		return 0
	}

	// Find the minimum column count
	minCols := len(allFields[0])
	for _, fields := range allFields[1:] {
		if len(fields) < minCols {
			minCols = len(fields)
		}
	}

	// Check each column: find the first one where values differ
	for col := 0; col < minCols; col++ {
		first := allFields[0][col]
		differs := false
		for _, fields := range allFields[1:] {
			if fields[col] != first {
				differs = true
				break
			}
		}
		if differs {
			return col
		}
	}

	return 0 // fallback: first column
}

// isHeaderLine returns true if the line looks like column headers.
func (lp *LiveParser) isHeaderLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}

	numericCount := 0
	for _, f := range fields {
		if _, err := strconv.ParseFloat(f, 64); err == nil {
			numericCount++
		}
	}

	// Header if most fields are non-numeric text
	return numericCount == 0 || float64(numericCount)/float64(len(fields)) < 0.3
}

// extractColumns extracts column names, filtering out timestamp tokens.
func (lp *LiveParser) extractColumns(line string) []string {
	fields := strings.Fields(line)
	var cols []string
	for _, f := range fields {
		// Skip timestamp parts
		if f == "AM" || f == "PM" {
			continue
		}
		if len(f) == 8 && f[2] == ':' && f[5] == ':' {
			continue
		}
		cols = append(cols, f)
	}
	return cols
}

func (lp *LiveParser) hasNumericFields(line string) bool {
	fields := strings.Fields(line)
	for _, f := range fields {
		if _, err := strconv.ParseFloat(f, 64); err == nil {
			return true
		}
	}
	return false
}

// parseDataLine parses a line using detected columns.
// Returns one DataPoint per selected numeric column.
func (lp *LiveParser) parseDataLine(line string) []DataPoint {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	now := time.Now()

	// Build a column→value map from the data fields.
	// Data lines may have fewer/more fields than headers (timestamps, etc.)
	// Strategy: match from the right side — rightmost fields align with rightmost columns.
	values := lp.alignFieldsToColumns(fields)
	if len(values) == 0 {
		return nil
	}

	// If user specified --y, emit only that column
	if lp.yCol != "" && lp.yCol != "0" && lp.yCol != "1" {
		if v, ok := values[lp.yCol]; ok {
			return []DataPoint{{Label: lp.yCol, Value: v, Time: now}}
		}
		// Try by index
		if idx, err := strconv.Atoi(lp.yCol); err == nil && idx < len(lp.columns) {
			colName := lp.columns[idx]
			if v, ok := values[colName]; ok {
				return []DataPoint{{Label: colName, Value: v, Time: now}}
			}
		}
		return nil
	}

	// No --y specified: emit all numeric columns as separate labeled DataPoints
	var results []DataPoint
	for _, col := range lp.columns {
		if v, ok := values[col]; ok {
			results = append(results, DataPoint{Label: col, Value: v, Time: now})
		}
	}
	return results
}

// alignFieldsToColumns maps data fields to column names.
// Handles the common case where data lines have leading timestamp fields
// that aren't in the column list.
func (lp *LiveParser) alignFieldsToColumns(fields []string) map[string]float64 {
	result := make(map[string]float64)

	// Extract only numeric fields and text fields
	type fieldInfo struct {
		value     float64
		text      string
		isNumeric bool
	}
	var parsed []fieldInfo
	for _, f := range fields {
		// Skip AM/PM
		if f == "AM" || f == "PM" {
			continue
		}
		// Skip timestamp-like
		if len(f) == 8 && f[2] == ':' && f[5] == ':' {
			continue
		}
		if v, err := strconv.ParseFloat(f, 64); err == nil {
			parsed = append(parsed, fieldInfo{value: v, isNumeric: true})
		} else {
			parsed = append(parsed, fieldInfo{text: f})
		}
	}

	// Try to align parsed fields with columns.
	// If counts match, direct 1:1 mapping.
	if len(parsed) == len(lp.columns) {
		for i, col := range lp.columns {
			if parsed[i].isNumeric {
				result[col] = parsed[i].value
			}
		}
		return result
	}

	// If more parsed fields than columns, skip leading extras (timestamps/UID)
	if len(parsed) > len(lp.columns) {
		offset := len(parsed) - len(lp.columns)
		for i, col := range lp.columns {
			idx := i + offset
			if idx < len(parsed) && parsed[idx].isNumeric {
				result[col] = parsed[idx].value
			}
		}
		return result
	}

	// Fewer parsed fields — align from the right
	if len(parsed) < len(lp.columns) {
		offset := len(lp.columns) - len(parsed)
		for i, p := range parsed {
			if p.isNumeric {
				colIdx := i + offset
				if colIdx < len(lp.columns) {
					result[lp.columns[colIdx]] = p.value
				}
			}
		}
		return result
	}

	return result
}

// parseFallback uses simple ParseLine when no header detected.
func (lp *LiveParser) parseFallback(line string) []DataPoint {
	dp := ParseLine(line)
	if dp != nil {
		dp.Time = time.Now()
		return []DataPoint{*dp}
	}
	return nil
}

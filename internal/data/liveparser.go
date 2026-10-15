package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LiveParser parses live stdin line by line.
//
// Header modes (explicit, no auto-detection):
//   - hasHeader=true: first data line (after skip) is treated as column names
//   - addHeaders set: columns are pre-defined, all lines are data
//   - neither: fallback to simple number extraction (ParseLine)
type LiveParser struct {
	columns       []string
	headerDone    bool
	lineCount     int
	yCol          string
	skipLines     int
	skipColumns   []int    // column indices to exclude
	rowStream     int
	rowBuffer     []string
	hasHeader     bool     // --header: first line is column names
	addHeaders    []string // --add-header: pre-defined column names
	patternParser *PatternParser
	streamInfo    *StreamInfo
}

// NewLiveParser creates a parser.
func NewLiveParser(yCol string, pattern *PatternParser, skip int, rowStream int, hasHeader bool, addHeaders []string, skipColumns []int) *LiveParser {
	lp := &LiveParser{
		yCol:          yCol,
		skipLines:     skip,
		skipColumns:   skipColumns,
		rowStream:     rowStream,
		hasHeader:     hasHeader,
		addHeaders:    addHeaders,
		patternParser: pattern,
	}
	if len(addHeaders) > 0 {
		lp.columns = addHeaders
		lp.headerDone = true
	}
	return lp
}

func (lp *LiveParser) GetStreamInfo() *StreamInfo {
	return lp.streamInfo
}

func (lp *LiveParser) Columns() []string {
	return lp.columns
}

// Feed processes one line and returns zero or more DataPoints.
func (lp *LiveParser) Feed(line string) []DataPoint {
	lp.lineCount++

	if lp.lineCount <= lp.skipLines {
		return nil
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	// Pattern parser takes priority
	if lp.patternParser != nil {
		dp := lp.patternParser.ParseLine(line)
		if dp != nil {
			dp.Time = time.Now()
			return []DataPoint{*dp}
		}
		return nil
	}

	// Stream detection for known tools (ping header, etc.)
	if lp.streamInfo == nil {
		if info := DetectStream(trimmed); info != nil {
			lp.streamInfo = info
			return nil
		}
	}

	// --header: consume first data line as column names
	if lp.hasHeader && !lp.headerDone {
		lp.columns = lp.extractColumns(trimmed)
		lp.headerDone = true
		return nil
	}

	// Parse as data
	if len(lp.columns) > 0 {
		if lp.rowStream > 0 {
			return lp.feedRowStream(trimmed)
		}
		return lp.parseDataLine(trimmed)
	}

	// No columns — fallback to simple parsing
	if lp.rowStream > 0 {
		return lp.feedRowStream(trimmed)
	}
	return lp.parseFallback(trimmed)
}

// --- Column extraction ---

func (lp *LiveParser) extractColumns(line string) []string {
	fields := strings.Fields(line)
	var cols []string
	idx := 0
	for _, f := range fields {
		if f == "AM" || f == "PM" {
			continue
		}
		if len(f) == 8 && f[2] == ':' && f[5] == ':' {
			continue
		}
		if !lp.isSkippedColumn(idx) {
			cols = append(cols, f)
		}
		idx++
	}
	return cols
}

func (lp *LiveParser) isSkippedColumn(idx int) bool {
	for _, sc := range lp.skipColumns {
		// Convert 1-based user input to 0-based
		if sc-1 == idx {
			return true
		}
	}
	return false
}

// --- Data line parsing ---

func (lp *LiveParser) parseDataLine(line string) []DataPoint {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	now := time.Now()
	values := lp.alignFieldsToColumns(fields)
	if len(values) == 0 {
		return nil
	}

	// --y selects a single column
	if lp.yCol != "" && lp.yCol != "0" && lp.yCol != "1" {
		if v, ok := values[lp.yCol]; ok {
			return []DataPoint{{Label: lp.yCol, Value: v, Time: now}}
		}
		if idx, err := strconv.Atoi(lp.yCol); err == nil && idx < len(lp.columns) {
			colName := lp.columns[idx]
			if v, ok := values[colName]; ok {
				return []DataPoint{{Label: colName, Value: v, Time: now}}
			}
		}
		return nil
	}

	// No --y: emit all numeric columns
	var results []DataPoint
	for _, col := range lp.columns {
		if v, ok := values[col]; ok {
			results = append(results, DataPoint{Label: col, Value: v, Time: now})
		}
	}
	return results
}

func (lp *LiveParser) alignFieldsToColumns(fields []string) map[string]float64 {
	result := make(map[string]float64)

	type fieldInfo struct {
		value     float64
		isNumeric bool
	}
	var parsed []fieldInfo
	colIdx := 0
	for _, f := range fields {
		if f == "AM" || f == "PM" {
			continue
		}
		if len(f) == 8 && f[2] == ':' && f[5] == ':' {
			continue
		}
		if lp.isSkippedColumn(colIdx) {
			colIdx++
			continue
		}
		colIdx++
		if v, err := strconv.ParseFloat(f, 64); err == nil {
			parsed = append(parsed, fieldInfo{value: v, isNumeric: true})
		} else {
			parsed = append(parsed, fieldInfo{})
		}
	}

	if len(parsed) == len(lp.columns) {
		for i, col := range lp.columns {
			if parsed[i].isNumeric {
				result[col] = parsed[i].value
			}
		}
		return result
	}

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

// --- Row-stream ---

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

	allFields := make([][]string, len(lp.rowBuffer))
	for i, rowLine := range lp.rowBuffer {
		allFields[i] = lp.stripTimestamp(strings.Fields(rowLine))
	}

	labelColIdx := lp.findLabelColumn(allFields)

	for i, fields := range allFields {
		label := fmt.Sprintf("row-%d", i)
		if labelColIdx >= 0 && labelColIdx < len(fields) {
			label = fields[labelColIdx]
		}

		value := 0.0
		if len(lp.columns) > 0 {
			values := lp.alignFieldsToColumns(strings.Fields(lp.rowBuffer[i]))
			if lp.yCol != "" && lp.yCol != "0" && lp.yCol != "1" {
				if v, ok := values[lp.yCol]; ok {
					value = v
				}
			} else {
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

func (lp *LiveParser) findLabelColumn(allFields [][]string) int {
	if len(allFields) < 2 {
		return 0
	}
	minCols := len(allFields[0])
	for _, fields := range allFields[1:] {
		if len(fields) < minCols {
			minCols = len(fields)
		}
	}
	for col := 0; col < minCols; col++ {
		first := allFields[0][col]
		for _, fields := range allFields[1:] {
			if fields[col] != first {
				return col
			}
		}
	}
	return 0
}

// --- Helpers ---

func (lp *LiveParser) hasNumericFields(line string) bool {
	for _, f := range strings.Fields(line) {
		if _, err := strconv.ParseFloat(f, 64); err == nil {
			return true
		}
	}
	return false
}

func (lp *LiveParser) parseFallback(line string) []DataPoint {
	dp := ParseLine(line)
	if dp != nil {
		dp.Time = time.Now()
		return []DataPoint{*dp}
	}
	return nil
}

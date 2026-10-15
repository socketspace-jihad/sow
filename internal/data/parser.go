package data

import (
	"regexp"
	"strconv"
	"strings"
)

// Known patterns for extracting numeric values from common command outputs.
var (
	// ping: time=29.3 ms
	pingRe = regexp.MustCompile(`time=([0-9.]+)\s*ms`)
	// ping header: PING google.com (142.250.185.46) 56(84) bytes of data.
	pingHeaderRe = regexp.MustCompile(`^PING\s+(\S+)`)
	// vmstat header line (contains "procs" or "r  b" etc.)
	vmstatHeaderRe = regexp.MustCompile(`^\s*(procs|r\s+b\s+)`)
	// iostat device line
	iostatHeaderRe = regexp.MustCompile(`^(avg-cpu|Device)`)
	// Generic: any floating point number
	floatRe = regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)
)

// DetectStream examines a line for stream metadata (headers from known tools).
// Returns non-nil StreamInfo if the line is a recognized header, nil otherwise.
func DetectStream(line string) *StreamInfo {
	line = strings.TrimSpace(line)

	// PING google.com (142.250.185.46) 56(84) bytes of data.
	if m := pingHeaderRe.FindStringSubmatch(line); m != nil {
		host := m[1]
		return &StreamInfo{
			Source: "ping",
			Host:   host,
			XLabel: "time",
			YLabel: host + " latency",
			Unit:   "ms",
		}
	}

	// vmstat header
	if vmstatHeaderRe.MatchString(line) {
		return &StreamInfo{
			Source: "vmstat",
			XLabel: "time",
			YLabel: "value",
			Unit:   "",
		}
	}

	// iostat header
	if iostatHeaderRe.MatchString(line) {
		return &StreamInfo{
			Source: "iostat",
			XLabel: "time",
			YLabel: "value",
			Unit:   "",
		}
	}

	return nil
}

// ParseLine attempts to extract a DataPoint from a single line of text.
// Used in live/streaming mode. Returns nil if no numeric value found.
func ParseLine(line string) *DataPoint {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Try ping format first
	if m := pingRe.FindStringSubmatch(line); m != nil {
		v, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			return &DataPoint{Label: "ping", Value: v}
		}
	}

	// Try to extract the last number on the line (most likely the value)
	matches := floatRe.FindAllString(line, -1)
	if len(matches) > 0 {
		v, err := strconv.ParseFloat(matches[len(matches)-1], 64)
		if err == nil {
			label := ""
			// Use non-numeric prefix as label
			idx := strings.Index(line, matches[len(matches)-1])
			if idx > 0 {
				label = strings.TrimSpace(line[:idx])
				// Clean up common delimiters from label
				label = strings.TrimRight(label, ",\t |:")
			}
			return &DataPoint{Label: label, Value: v}
		}
	}

	return nil
}

// ParseText parses a block of text into a DataSet.
// Handles CSV, TSV, whitespace-separated, and mixed formats.
func ParseText(text string, delimiter string, headerMode int) (*DataSet, error) {
	lines := splitLines(text)
	if len(lines) == 0 {
		return &DataSet{}, nil
	}

	delim := delimiter
	if delim == "" {
		delim = detectDelimiter(lines[0])
	}

	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		fields := splitFields(line, delim)
		if len(fields) > 0 {
			rows = append(rows, fields)
		}
	}

	if len(rows) == 0 {
		return &DataSet{}, nil
	}

	ds := &DataSet{}
	hasHeader := headerMode == 1 || (headerMode == 0 && detectHeader(rows[0]))
	if hasHeader && len(rows) > 1 {
		ds.Headers = rows[0]
		ds.Rows = rows[1:]
	} else {
		ds.Rows = rows
	}

	return ds, nil
}

func splitLines(text string) []string {
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, len(raw))
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func splitFields(line string, delim string) []string {
	switch delim {
	case ",":
		return splitCSVSimple(line)
	case "\t":
		parts := strings.Split(line, "\t")
		return trimAll(parts)
	default:
		// Whitespace
		return strings.Fields(line)
	}
}

func splitCSVSimple(line string) []string {
	// Simple CSV split that handles basic quoting
	var fields []string
	var current strings.Builder
	inQuote := false

	for _, r := range line {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ',' && !inQuote:
			fields = append(fields, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	fields = append(fields, strings.TrimSpace(current.String()))
	return fields
}

func trimAll(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func detectDelimiter(line string) string {
	commas := strings.Count(line, ",")
	tabs := strings.Count(line, "\t")
	if commas > 0 && commas >= tabs {
		return ","
	}
	if tabs > 0 {
		return "\t"
	}
	return " " // whitespace
}

// detectHeader returns true if the first row looks like a header.
// Heuristic: a header row where ALL fields are non-numeric text.
// Mixed rows (some text, some numbers) like "Jan 12500" are data, not headers.
func detectHeader(row []string) bool {
	if len(row) == 0 {
		return false
	}
	for _, f := range row {
		if _, err := strconv.ParseFloat(f, 64); err == nil {
			return false // has at least one number → it's a data row
		}
	}
	// Every field is non-numeric → likely a header
	return true
}

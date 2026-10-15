package data

import (
	"regexp"
	"strconv"
	"strings"
)

// ToolProfile defines how to spawn and parse a known system tool.
type ToolProfile struct {
	Name          string            // tool name
	Command       string            // binary to execute
	DefaultArgs   []string          // default arguments (e.g., interval "1")
	DefaultMetric string            // default column to graph
	DefaultChart  string            // default chart type (empty = "streamline")
	Unit          string            // unit of the default metric
	Metrics       map[string]string // metric name → description
	Title         string            // auto-generated title
	MultiRow      bool              // true = each data line has a row ID (e.g., CPU core)
	HeaderDetect  func(line string) bool
	ColumnParse   func(headerLines []string) []string
	DataParse     func(line string, columns []string) map[string]float64
	// DataParseMulti returns metric values + row ID (e.g., CPU number).
	// Only used when MultiRow is true. Returns (values, rowID, ok).
	DataParseMulti func(line string, columns []string) (map[string]float64, float64, bool)
}

// Profiles returns all known tool profiles.
var Profiles = map[string]*ToolProfile{
	"ping":    pingProfile(),
	"vmstat":  vmstatProfile(),
	"iostat":  iostatProfile(),
	"sar":     sarProfile(),
	"pidstat": pidstatProfile(),
	"mpstat":  mpstatProfile(),
}

// ListMetrics returns available metrics for a tool as "name - description" strings.
func (p *ToolProfile) ListMetrics() []string {
	result := make([]string, 0, len(p.Metrics))
	for name, desc := range p.Metrics {
		result = append(result, name+" - "+desc)
	}
	return result
}

// --- ping ---

func pingProfile() *ToolProfile {
	return &ToolProfile{
		Name:          "ping",
		Command:       "ping",
		DefaultArgs:   nil, // user provides host
		DefaultMetric: "time",
		Unit:          "ms",
		Title:         "Ping Latency",
		Metrics: map[string]string{
			"time": "round-trip time (ms)",
		},
		HeaderDetect: func(line string) bool {
			return strings.HasPrefix(line, "PING ")
		},
		ColumnParse: func(headerLines []string) []string {
			return []string{"time"}
		},
		DataParse: func(line string, columns []string) map[string]float64 {
			re := regexp.MustCompile(`time=([0-9.]+)\s*ms`)
			if m := re.FindStringSubmatch(line); m != nil {
				v, err := strconv.ParseFloat(m[1], 64)
				if err == nil {
					return map[string]float64{"time": v}
				}
			}
			return nil
		},
	}
}

// --- vmstat ---

func vmstatProfile() *ToolProfile {
	return &ToolProfile{
		Name:          "vmstat",
		Command:       "vmstat",
		DefaultArgs:   []string{"1"},
		DefaultMetric: "us",
		Unit:          "%",
		Title:         "System Stats (vmstat)",
		Metrics: map[string]string{
			"r":  "runnable processes",
			"b":  "blocked processes",
			"us": "user CPU %",
			"sy": "system CPU %",
			"id": "idle CPU %",
			"wa": "I/O wait %",
			"st": "steal %",
			"si": "swap in (KB/s)",
			"so": "swap out (KB/s)",
			"bi": "blocks in",
			"bo": "blocks out",
			"in": "interrupts/sec",
			"cs": "context switches/sec",
		},
		HeaderDetect: func(line string) bool {
			trimmed := strings.TrimSpace(line)
			return strings.HasPrefix(trimmed, "procs") ||
				strings.HasPrefix(trimmed, "r ") ||
				strings.HasPrefix(trimmed, "r  ")
		},
		ColumnParse: func(headerLines []string) []string {
			// vmstat has 2 header lines; the second has column names
			for _, line := range headerLines {
				fields := strings.Fields(line)
				if len(fields) > 5 && fields[0] == "r" {
					return fields
				}
			}
			// Fallback: standard vmstat columns
			return []string{"r", "b", "swpd", "free", "buff", "cache",
				"si", "so", "bi", "bo", "in", "cs",
				"us", "sy", "id", "wa", "st"}
		},
		DataParse: func(line string, columns []string) map[string]float64 {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 2 {
				return nil
			}
			// Skip if first field is not a number (header line)
			if _, err := strconv.ParseFloat(fields[0], 64); err != nil {
				return nil
			}
			result := make(map[string]float64)
			for i, col := range columns {
				if i < len(fields) {
					if v, err := strconv.ParseFloat(fields[i], 64); err == nil {
						result[col] = v
					}
				}
			}
			return result
		},
	}
}

// --- iostat ---

func iostatProfile() *ToolProfile {
	return &ToolProfile{
		Name:          "iostat",
		Command:       "iostat",
		DefaultArgs:   []string{"-x", "1"},
		DefaultMetric: "%idle",
		Unit:          "%",
		Title:         "I/O Stats (iostat)",
		Metrics: map[string]string{
			"%user":   "user CPU %",
			"%nice":   "nice CPU %",
			"%system": "system CPU %",
			"%iowait": "I/O wait %",
			"%steal":  "steal %",
			"%idle":   "idle CPU %",
		},
		HeaderDetect: func(line string) bool {
			trimmed := strings.TrimSpace(line)
			return strings.HasPrefix(trimmed, "avg-cpu") ||
				strings.HasPrefix(trimmed, "Device") ||
				strings.Contains(trimmed, "%user")
		},
		ColumnParse: func(headerLines []string) []string {
			for _, line := range headerLines {
				if strings.Contains(line, "%user") || strings.Contains(line, "%idle") {
					return strings.Fields(strings.TrimSpace(line))
				}
			}
			return []string{"%user", "%nice", "%system", "%iowait", "%steal", "%idle"}
		},
		DataParse: func(line string, columns []string) map[string]float64 {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 2 {
				return nil
			}
			if _, err := strconv.ParseFloat(fields[0], 64); err != nil {
				return nil
			}
			result := make(map[string]float64)
			for i, col := range columns {
				if i < len(fields) {
					if v, err := strconv.ParseFloat(fields[i], 64); err == nil {
						result[col] = v
					}
				}
			}
			return result
		},
	}
}

// --- sar ---

func sarProfile() *ToolProfile {
	return &ToolProfile{
		Name:          "sar",
		Command:       "sar",
		DefaultArgs:   []string{"1"},
		DefaultMetric: "%idle",
		Unit:          "%",
		Title:         "System Activity (sar)",
		Metrics: map[string]string{
			"%user":   "user CPU %",
			"%nice":   "nice CPU %",
			"%system": "system CPU %",
			"%iowait": "I/O wait %",
			"%steal":  "steal %",
			"%idle":   "idle CPU %",
		},
		HeaderDetect: func(line string) bool {
			return strings.Contains(line, "CPU") && strings.Contains(line, "%user")
		},
		ColumnParse: func(headerLines []string) []string {
			for _, line := range headerLines {
				if strings.Contains(line, "%user") {
					fields := strings.Fields(line)
					// Skip timestamp and "CPU" or "all" column
					var cols []string
					for _, f := range fields {
						if strings.HasPrefix(f, "%") || f == "CPU" {
							cols = append(cols, f)
						}
					}
					if len(cols) > 0 {
						return cols
					}
				}
			}
			return []string{"CPU", "%user", "%nice", "%system", "%iowait", "%steal", "%idle"}
		},
		DataParse: func(line string, columns []string) map[string]float64 {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 3 {
				return nil
			}
			// sar lines start with timestamp; find the numeric fields
			result := make(map[string]float64)
			// Find "all" or numeric CPU id, then the metric values follow
			startIdx := -1
			for i, f := range fields {
				if f == "all" || f == "ALL" {
					startIdx = i + 1
					break
				}
			}
			if startIdx < 0 {
				// Try parsing all numeric fields
				numIdx := 0
				for _, f := range fields {
					if v, err := strconv.ParseFloat(f, 64); err == nil {
						if numIdx < len(columns) {
							result[columns[numIdx]] = v
						}
						numIdx++
					}
				}
			} else {
				metricCols := make([]string, 0)
				for _, c := range columns {
					if strings.HasPrefix(c, "%") {
						metricCols = append(metricCols, c)
					}
				}
				for i, col := range metricCols {
					idx := startIdx + i
					if idx < len(fields) {
						if v, err := strconv.ParseFloat(fields[idx], 64); err == nil {
							result[col] = v
						}
					}
				}
			}
			if len(result) == 0 {
				return nil
			}
			return result
		},
	}
}

// --- pidstat ---

func pidstatProfile() *ToolProfile {
	return &ToolProfile{
		Name:          "pidstat",
		Command:       "pidstat",
		DefaultArgs:   []string{"1"},
		DefaultMetric: "%CPU",
		Unit:          "%",
		Title:         "Process Stats (pidstat)",
		Metrics: map[string]string{
			"%usr":    "user CPU %",
			"%system": "system CPU %",
			"%guest":  "guest CPU %",
			"%wait":   "wait %",
			"%CPU":    "total CPU %",
		},
		HeaderDetect: func(line string) bool {
			return strings.Contains(line, "PID") &&
				(strings.Contains(line, "%CPU") || strings.Contains(line, "%usr"))
		},
		ColumnParse: func(headerLines []string) []string {
			for _, line := range headerLines {
				if strings.Contains(line, "%CPU") {
					fields := strings.Fields(line)
					var cols []string
					for _, f := range fields {
						if strings.HasPrefix(f, "%") || f == "PID" || f == "Command" {
							cols = append(cols, f)
						}
					}
					if len(cols) > 0 {
						return cols
					}
				}
			}
			return []string{"%usr", "%system", "%guest", "%wait", "%CPU"}
		},
		DataParse: func(line string, columns []string) map[string]float64 {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 3 {
				return nil
			}
			// pidstat: timestamp UID PID %usr %system %guest %wait %CPU CPU Command
			result := make(map[string]float64)
			for i, f := range fields {
				if v, err := strconv.ParseFloat(f, 64); err == nil {
					// Map by position — find which column this corresponds to
					// Simple approach: match fields that look like metrics
					_ = i
					_ = v
				}
			}
			// Better approach: find all numeric fields after PID
			pidFound := false
			numIdx := 0
			metricCols := make([]string, 0)
			for _, c := range columns {
				if strings.HasPrefix(c, "%") {
					metricCols = append(metricCols, c)
				}
			}
			for _, f := range fields {
				if !pidFound {
					if _, err := strconv.Atoi(f); err == nil && len(f) > 1 {
						pidFound = true
					}
					continue
				}
				if v, err := strconv.ParseFloat(f, 64); err == nil {
					if numIdx < len(metricCols) {
						result[metricCols[numIdx]] = v
					}
					numIdx++
				}
			}
			if len(result) == 0 {
				return nil
			}
			return result
		},
	}
}

// --- mpstat ---

func mpstatProfile() *ToolProfile {
	return &ToolProfile{
		Name:          "mpstat",
		Command:       "mpstat",
		DefaultArgs:   []string{"-P", "ALL", "1"},
		DefaultMetric: "%usr",
		DefaultChart:  "heatmap",
		Unit:          "%",
		Title:         "CPU Utilization per Core (mpstat)",
		MultiRow:      true,
		Metrics: map[string]string{
			"%usr":    "user CPU %",
			"%nice":   "nice %",
			"%sys":    "system CPU %",
			"%iowait": "I/O wait %",
			"%irq":    "hardware interrupt %",
			"%soft":   "software interrupt %",
			"%steal":  "steal %",
			"%guest":  "guest %",
			"%gnice":  "guest nice %",
			"%idle":   "idle %",
		},
		HeaderDetect: func(line string) bool {
			trimmed := strings.TrimSpace(line)
			return strings.Contains(trimmed, "%usr") ||
				strings.HasPrefix(trimmed, "Linux ") ||
				trimmed == ""
		},
		ColumnParse: func(headerLines []string) []string {
			for _, line := range headerLines {
				if strings.Contains(line, "%usr") {
					fields := strings.Fields(line)
					var cols []string
					found := false
					for _, f := range fields {
						if f == "CPU" {
							found = true
							cols = append(cols, f)
							continue
						}
						if found {
							cols = append(cols, f)
						}
					}
					if len(cols) > 0 {
						return cols
					}
				}
			}
			return []string{"CPU", "%usr", "%nice", "%sys", "%iowait", "%irq", "%soft", "%steal", "%guest", "%gnice", "%idle"}
		},
		DataParseMulti: func(line string, columns []string) (map[string]float64, float64, bool) {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 3 {
				return nil, 0, false
			}

			// Skip "all" aggregate line
			for _, f := range fields {
				if f == "all" || f == "ALL" {
					return nil, 0, false
				}
			}

			// Find the CPU number field
			cpuIdx := -1
			for i, f := range fields {
				if cpuNum, err := strconv.Atoi(f); err == nil && cpuNum >= 0 && cpuNum < 1024 && i > 0 {
					cpuIdx = i
					break
				}
			}
			if cpuIdx < 0 {
				return nil, 0, false
			}

			cpuNum, _ := strconv.ParseFloat(fields[cpuIdx], 64)

			result := make(map[string]float64)
			metricCols := make([]string, 0)
			for _, c := range columns {
				if strings.HasPrefix(c, "%") {
					metricCols = append(metricCols, c)
				}
			}

			valueStart := cpuIdx + 1
			for i, col := range metricCols {
				idx := valueStart + i
				if idx < len(fields) {
					if v, err := strconv.ParseFloat(fields[idx], 64); err == nil {
						result[col] = v
					}
				}
			}

			if len(result) == 0 {
				return nil, 0, false
			}
			return result, cpuNum, true
		},
	}
}

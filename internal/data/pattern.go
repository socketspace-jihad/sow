package data

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PatternParser parses log lines using a user-defined pattern.
// Pattern syntax:
//
//	{label} or {a}, {b} — text capture (becomes the series label)
//	{value} or {1}, {2} — numeric capture (becomes the value)
//
// If no numeric capture matches, the parser counts occurrences per label.
//
// Examples:
//
//	Pattern: "{label}={value}"
//	  "response_time=45.2" → label="response_time", value=45.2
//	  "user=client"        → label="user", value=+1 (count mode)
//
//	Pattern: "status={label} latency={value}ms"
//	  "status=200 latency=12.5ms" → label="200", value=12.5
//
//	Pattern: "{label}"
//	  "ERROR"    → label="ERROR", value=+1
//	  "WARNING"  → label="WARNING", value=+1
type PatternParser struct {
	re          *regexp.Regexp
	labelGroups []string // names of text capture groups
	valueGroups []string // names of numeric capture groups
	counts      map[string]float64
	mu          sync.Mutex
}

// NewPatternParser compiles a pattern string into a PatternParser.
func NewPatternParser(pattern string) (*PatternParser, error) {
	pp := &PatternParser{
		counts: make(map[string]float64),
	}

	// Convert pattern syntax to regex:
	// {label}, {a}, {b}, etc. → named capture group for text
	// {value}, {1}, {2}, etc. → named capture group for numbers
	reStr := regexp.QuoteMeta(pattern)

	// Replace escaped captures back with regex groups
	textGroupRe := regexp.MustCompile(`\\\{([a-zA-Z_][a-zA-Z0-9_]*)\\\}`)
	numGroupRe := regexp.MustCompile(`\\\{(\d+)\\\}`)
	valueGroupRe := regexp.MustCompile(`\\\{value\\\}`)

	// First handle {value} explicitly
	if valueGroupRe.MatchString(reStr) {
		reStr = valueGroupRe.ReplaceAllString(reStr, `(?P<_val_0>[-+]?[0-9]*\.?[0-9]+)`)
		pp.valueGroups = append(pp.valueGroups, "_val_0")
	}

	// Handle {1}, {2}, etc. — numeric captures
	idx := 0
	reStr = numGroupRe.ReplaceAllStringFunc(reStr, func(match string) string {
		name := "_num_" + strconv.Itoa(idx)
		idx++
		pp.valueGroups = append(pp.valueGroups, name)
		return `(?P<` + name + `>[-+]?[0-9]*\.?[0-9]+)`
	})

	// Handle {label}, {a}, {b}, etc. — text captures
	reStr = textGroupRe.ReplaceAllStringFunc(reStr, func(match string) string {
		// Extract the name between \{ and \}
		inner := match[2 : len(match)-2] // strip \{ and \}
		name := "_txt_" + inner
		pp.labelGroups = append(pp.labelGroups, name)
		return `(?P<` + name + `>\S+)`
	})

	re, err := regexp.Compile(reStr)
	if err != nil {
		return nil, err
	}
	pp.re = re
	return pp, nil
}

// ParseLine applies the pattern to a line and returns a DataPoint.
// Returns nil if the pattern doesn't match.
func (pp *PatternParser) ParseLine(line string) *DataPoint {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	match := pp.re.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	// Extract named groups
	groupMap := make(map[string]string)
	for i, name := range pp.re.SubexpNames() {
		if i > 0 && name != "" && i < len(match) {
			groupMap[name] = match[i]
		}
	}

	// Get label from text captures
	label := ""
	for _, g := range pp.labelGroups {
		if v, ok := groupMap[g]; ok && v != "" {
			if label != "" {
				label += ":"
			}
			label += v
		}
	}

	// Get value from numeric captures
	var value float64
	hasValue := false
	for _, g := range pp.valueGroups {
		if v, ok := groupMap[g]; ok && v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				value = f
				hasValue = true
				break
			}
		}
	}

	// If no numeric value captured, use count mode
	if !hasValue {
		pp.mu.Lock()
		if label == "" {
			label = "match"
		}
		pp.counts[label]++
		value = pp.counts[label]
		pp.mu.Unlock()
	}

	if label == "" {
		label = "value"
	}

	return &DataPoint{
		Label: label,
		Value: value,
		Time:  time.Now(),
	}
}

package dna

import (
	"regexp"
	"strconv"
	"strings"
)

// Interpreting an inference response is a Space's concern, not inference's —
// inference is dumb (text in, text out); the cell composes what it sends and
// reads what comes back. This is the comparator-output reader the cells use.

// alignmentRe matches the comparator's `<alignment>` output tolerantly: closing
// tag optional, any surrounding whitespace/newlines, a signed decimal. The
// .1001 runs showed format drift (bare `<alignment>` without a closing tag), so
// we anchor only on the opening tag + the first number after it.
var alignmentRe = regexp.MustCompile(`(?is)<alignment>\s*(-?\d*\.?\d+)`)

// ParseAlignment extracts the headline `<alignment>` score from a comparator
// response, clamped to [-1, 1]. ok=false if no parseable alignment is present.
func ParseAlignment(response string) (float64, bool) {
	m := alignmentRe.FindStringSubmatch(response)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	if v > 1 {
		v = 1
	}
	if v < -1 {
		v = -1
	}
	return v, true
}

// labelRe matches a `<label>` gist tolerantly: case-insensitive, captures the
// text after the opening tag up to the closing tag, a `<`, or a newline (so a
// missing/garbled closing tag still parses, same forgiveness as alignmentRe).
var labelRe = regexp.MustCompile(`(?i)<label>\s*([^<\n]+)`)

// ParseLabel extracts the grouping framing/gist from a grouping-pass response.
// ok=false if no parseable label is present.
func ParseLabel(response string) (string, bool) {
	m := labelRe.FindStringSubmatch(response)
	if m == nil {
		return "", false
	}
	label := strings.TrimSpace(m[1])
	if label == "" {
		return "", false
	}
	return label, true
}

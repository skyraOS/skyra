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

// predictionRe matches a single `<prediction>` tag (self's point-estimate format).
// The move may be wrapped in an attribution tag — `<prediction><username>text
// </username></prediction>` — the optional inner tag is skipped.
var predictionRe = regexp.MustCompile(`(?i)<prediction>\s*(?:<[^/>]+>\s*)?([^<\n]+)`)

// ParsePrediction extracts a single predicted move (self's format: one tag, no
// probability). ok=false if no parseable prediction is present.
func ParsePrediction(response string) (string, bool) {
	m := predictionRe.FindStringSubmatch(response)
	if m == nil {
		return "", false
	}
	pred := strings.TrimSpace(m[1])
	if pred == "" {
		return "", false
	}
	return pred, true
}

// predSetRe matches the outer <predictions> wrapper (tolerant of whitespace/missing
// closing tag).
var predSetRe = regexp.MustCompile(`(?is)<predictions>(.*?)(?:</predictions>|$)`)

// predItemRe matches each <prediction p="0.41">text</prediction> inside the wrapper.
// Probability attribute is optional; text is captured inside the tags.
var predItemRe = regexp.MustCompile(`(?i)<prediction(?:\s+p="?(\d*\.?\d+)"?)?\s*>(.*?)</prediction>`)

// ParsedPrediction is one entry from the other-agent's probability distribution
// over the other party's next move.
type ParsedPrediction struct {
	Text        string
	Probability float64
}

// ParsePredictions extracts the full distribution from the other-agent's response
// (the <predictions>…</predictions> format). Returns the parsed list and whether
// any were found. Probabilities are taken verbatim from the p attribute; missing
// attributes default to 0 (caller normalises if needed).
func ParsePredictions(response string) ([]ParsedPrediction, bool) {
	setMatch := predSetRe.FindStringSubmatch(response)
	body := response
	if setMatch != nil {
		body = setMatch[1]
	}
	matches := predItemRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil, false
	}
	out := make([]ParsedPrediction, 0, len(matches))
	for _, m := range matches {
		text := strings.TrimSpace(m[2])
		if text == "" {
			continue
		}
		prob, _ := strconv.ParseFloat(strings.TrimSpace(m[1]), 64)
		out = append(out, ParsedPrediction{Text: text, Probability: prob})
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

var engageRe = regexp.MustCompile(`(?is)<engage>\s*<([^>/\s]+)>\s*(.*?)\s*</[^>]+>\s*</engage>`)
var perceiveRe = regexp.MustCompile(`(?is)<perc(?:ei|ie)ve>\s*(.*?)\s*</perc(?:ei|ie)ve>`)

// PrimitiveEmission is the concrete primitive the motor emits.
type PrimitiveEmission struct {
	Primitive string
	Target    string
	Payload   string
}

// ParseMotorEmission extracts the motor's concrete primitive emission.
// Supported forms:
//
//	<engage><michael>words</michael></engage>
//	<perceive>"nothing important yet"</perceive>
func ParseMotorEmission(response string) (PrimitiveEmission, bool) {
	if m := engageRe.FindStringSubmatch(response); m != nil {
		payload := cleanMotorPayload(m[2])
		if m[1] != "" && payload != "" {
			return PrimitiveEmission{Primitive: "engage", Target: strings.TrimSpace(m[1]), Payload: payload}, true
		}
	}
	if m := perceiveRe.FindStringSubmatch(response); m != nil {
		payload := cleanMotorPayload(m[1])
		if payload != "" {
			return PrimitiveEmission{Primitive: "perceive", Payload: payload}, true
		}
	}
	return PrimitiveEmission{}, false
}

func cleanMotorPayload(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	return strings.TrimSpace(s)
}

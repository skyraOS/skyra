package dnav2

// Parsers for prediction, grading, and concrete motor emissions.

import (
	"regexp"
	"strconv"
	"strings"
)

var predictionRe = regexp.MustCompile(`(?i)<prediction>\s*(?:<[^/>]+>\s*)?([^<\n]+)`)

func ParsePrediction(response string) (string, bool) {
	m := predictionRe.FindStringSubmatch(response)
	if m == nil {
		return "", false
	}
	p := strings.TrimSpace(m[1])
	return p, p != ""
}

var predSetRe = regexp.MustCompile(`(?is)<predictions>(.*?)(?:</predictions>|$)`)
var predItemRe = regexp.MustCompile(`(?i)<prediction(?:\s+p="?(\d*\.?\d+)"?)?\s*>(.*?)</prediction>`)

type ParsedPrediction struct {
	Text        string
	Probability float64
}

func ParsePredictions(response string) ([]ParsedPrediction, bool) {
	body := response
	if m := predSetRe.FindStringSubmatch(response); m != nil {
		body = m[1]
	}
	var out []ParsedPrediction
	for _, m := range predItemRe.FindAllStringSubmatch(body, -1) {
		text := strings.TrimSpace(m[2])
		if text == "" {
			continue
		}
		p, _ := strconv.ParseFloat(strings.TrimSpace(m[1]), 64)
		out = append(out, ParsedPrediction{Text: text, Probability: p})
	}
	return out, len(out) > 0
}

var engageRe = regexp.MustCompile(`(?is)<engage>\s*<([^>/\s]+)>\s*(.*?)\s*</[^>]+>\s*</engage>`)
var perceiveRe = regexp.MustCompile(`(?is)<perc(?:ei|ie)ve>\s*(.*?)\s*</perc(?:ei|ie)ve>`)
var moveRe = regexp.MustCompile(`(?is)<move>\s*<([^>/\s]+)>\s*(.*?)\s*(?:</[^>]+>\s*)?</move>`)
var waitRe = regexp.MustCompile(`(?is)<wait>\s*<(\d+(?:\.\d+)?)\s*>\s*(.*?)\s*</wait>`)

type PrimitiveEmission struct{ Primitive, Target, Payload string }

func ParseMotorEmission(response string) (PrimitiveEmission, bool) {
	if m := engageRe.FindStringSubmatch(response); m != nil {
		p := cleanMotorPayload(m[2])
		if m[1] != "" && p != "" {
			return PrimitiveEmission{"engage", strings.TrimSpace(m[1]), p}, true
		}
	}
	if m := moveRe.FindStringSubmatch(response); m != nil {
		t := strings.TrimSpace(m[1])
		if t != "" {
			return PrimitiveEmission{"move", t, cleanMotorPayload(m[2])}, true
		}
	}
	if perceiveRe.MatchString(response) {
		return PrimitiveEmission{Primitive: "perceive"}, true
	}
	if m := waitRe.FindStringSubmatch(response); m != nil {
		return PrimitiveEmission{"wait", strings.TrimSpace(m[1]), cleanMotorPayload(m[2])}, true
	}
	return PrimitiveEmission{}, false
}
func cleanMotorPayload(s string) string {
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(s), `"'`))
}

var matchRe = regexp.MustCompile(`(?is)<match>\s*(none|\d+)\s*</match>`)

func ParseMatch(response string) (int, bool, bool) {
	m := matchRe.FindStringSubmatch(response)
	if m == nil {
		return 0, false, false
	}
	v := strings.ToLower(strings.TrimSpace(m[1]))
	if v == "none" {
		return 0, false, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 0, false, false
	}
	return n, true, true
}

var execGradeRe = regexp.MustCompile(`(?is)<grade>\s*(yes|partial|no)\s*</grade>`)

func ParseExecGrade(response string) (float64, bool) {
	m := execGradeRe.FindStringSubmatch(response)
	if m == nil {
		return 0, false
	}
	switch strings.ToLower(strings.TrimSpace(m[1])) {
	case "yes":
		return 1, true
	case "partial":
		return .5, true
	case "no":
		return 0, true
	}
	return 0, false
}

package dnav2

import (
	"regexp"
	"strconv"
	"strings"
)

var coherenceGradeRe = regexp.MustCompile(`(?is)<g\s+n="(\d+)"\s*>\s*(-?\d*\.?\d+)\s*</g>`)

// ParseCoherenceGrades parses the settle's per-item alignment scores (active-grouping.md):
//
//	<grades><g n="1">0.8</g><g n="2">-0.3</g>...</grades> → {1:0.8, 2:-0.3}
func ParseCoherenceGrades(resp string) map[int]float64 {
	out := map[int]float64{}
	for _, m := range coherenceGradeRe.FindAllStringSubmatch(resp, -1) {
		n, err1 := strconv.Atoi(m[1])
		v, err2 := strconv.ParseFloat(m[2], 64)
		if err1 == nil && err2 == nil {
			out[n] = v
		}
	}
	return out
}

var (
	frameEditRe = regexp.MustCompile(`(?is)<edit\s+id="([^"]*)"\s*>(.*?)</edit>`)
	frameNewRe  = regexp.MustCompile(`(?is)<new>(.*?)</new>`)
)

var (
	structuralRetainRe  = regexp.MustCompile(`(?is)^\s*<retain\s*/>\s*$`)
	structuralComposeRe = regexp.MustCompile(`(?is)^\s*<compose>(.*?)</compose>\s*$`)
	structuralRequeryRe = regexp.MustCompile(`(?is)^\s*<requery>(.*?)</requery>\s*$`)
)

const (
	StructuralRetain  = "retain"
	StructuralCompose = "compose"
	StructuralRequery = "requery"
)

// ParseStructuralOp accepts exactly one Phase-4 Material/Formal operation.
// Subgraph validation is intentionally separate: retain has no graph, while
// compose/requery bodies must pass the universal annotated parser plus their
// operation-specific stable-reference rules.
func ParseStructuralOp(resp string) (op, annotated string, ok bool) {
	if structuralRetainRe.MatchString(resp) {
		return StructuralRetain, "", true
	}
	if m := structuralComposeRe.FindStringSubmatch(resp); m != nil {
		return StructuralCompose, strings.TrimSpace(m[1]), true
	}
	if m := structuralRequeryRe.FindStringSubmatch(resp); m != nil {
		return StructuralRequery, strings.TrimSpace(m[1]), true
	}
	return "", "", false
}

// ParseFrameOp parses midlevel's framing operator (active-grouping.md):
//
//	<edit id="ID">sharpened frame</edit>  → ("edit", "ID", frame, true)
//	<new>frame</new>                      → ("new",  "",   frame, true)
//
// Edit is tried first so a stray "<new>" inside an edit's frame can't win.
func ParseFrameOp(resp string) (op, id, frame string, ok bool) {
	if m := frameEditRe.FindStringSubmatch(resp); m != nil {
		return "edit", strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), true
	}
	if m := frameNewRe.FindStringSubmatch(resp); m != nil {
		return "new", "", strings.TrimSpace(m[1]), true
	}
	return "", "", "", false
}

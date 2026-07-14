package dnav3

import (
	"fmt"
	"strconv"
	"strings"
)

// The parser. Hand-rolled and deliberately tolerant, for the reason the grammar exists at all:
//
//	AN UNPARSEABLE LINE IS DROPPED AND LOGGED. NEVER FATAL.
//	Fewer predictions is always an acceptable outcome. NO ACTION is not.
//
// "Parse failure kept motor from acting at all" already happened once, in v2. So: XML envelope
// (self-delimiting — a truncated emission still yields every complete element before the cut) with
// line-oriented triples inside (per-line failure isolation — a malformed line costs one fact, not
// the document).
//
// We do NOT use encoding/xml: an act's payload is arbitrary content — file bodies, message text,
// possibly angle brackets — and a strict parser would choke on exactly the emissions we most want
// to survive.

// ParseResult is one parsed emission plus everything that went wrong while parsing it. Rejects are
// kept rather than discarded — a model that keeps emitting the same malformed line is telling you
// something, and it will not be visible if the parser eats it silently.
type ParseResult struct {
	Expression Expression
	Rejects    []Reject
}

// Reject is one line the parser could not use.
type Reject struct {
	Line   string
	Reason string
}

// Parse turns an emission into an Expression. It returns an error ONLY when the result could not
// possibly be acted on (no acts at all); everything else degrades to a Reject.
func Parse(raw string) (ParseResult, error) {
	var res ParseResult

	res.Expression.Framing = strings.TrimSpace(inner(raw, "framing"))

	for _, el := range elements(raw, "act") {
		act, err := parseAct(el)
		if err != nil {
			res.Rejects = append(res.Rejects, Reject{Line: el.open, Reason: err.Error()})
			continue
		}
		res.Expression.Acts = append(res.Expression.Acts, act)
	}

	for _, el := range elements(raw, "expect") {
		of := strings.Fields(attr(el.open, "of"))
		if len(of) == 0 {
			res.Rejects = append(res.Rejects, Reject{Line: el.open, Reason: `expect has no of="..."`})
			continue
		}
		// An <expect> block may carry several triple lines; each is independent.
		for _, line := range strings.Split(el.body, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			t, within, err := parseExpectLine(line)
			if err != nil {
				res.Rejects = append(res.Rejects, Reject{Line: strings.TrimSpace(line), Reason: err.Error()})
				continue
			}
			res.Expression.Expects = append(res.Expression.Expects, Expect{
				Of: of, Triple: t, Within: within,
			})
		}
	}

	if len(res.Expression.Acts) == 0 {
		return res, fmt.Errorf("emission contains no acts")
	}
	return res, nil
}

func parseAct(el element) (Act, error) {
	id := attr(el.open, "id")
	if id == "" {
		return Act{}, fmt.Errorf(`act has no id="..."`)
	}
	a := Act{
		ID:        id,
		Target:    attr(el.open, "target"),
		Primitive: attr(el.open, "primitive"),
		Payload:   strings.TrimSpace(el.body),
	}
	if a.Primitive == "" {
		return Act{}, fmt.Errorf(`act %q has no primitive="..."`, id)
	}
	// requires="<mode> <subject> <predicate> <object>", semicolon-separated for several.
	for _, r := range strings.Split(attr(el.open, "requires"), ";") {
		if strings.TrimSpace(r) == "" {
			continue
		}
		t, err := parseRequiresLine(r)
		if err != nil {
			return Act{}, fmt.Errorf("act %q: %w", id, err)
		}
		a.Requires = append(a.Requires, t)
	}
	return a, nil
}

// parseExpectLine reads:  <mode> <confidence> <within> <subject> <predicate> <object>
//
// The object is last and may be quoted, so it can carry spaces. Everything before it is a fixed
// number of fields, which is what makes a single malformed line cheap to drop.
func parseExpectLine(line string) (Triple, int, error) {
	f, obj, err := split(line, 5)
	if err != nil {
		return Triple{}, 0, err
	}
	mode, err := parseMode(f[0])
	if err != nil {
		return Triple{}, 0, err
	}
	conf, err := strconv.ParseFloat(f[1], 64)
	if err != nil || conf < 0 || conf > 1 {
		return Triple{}, 0, fmt.Errorf("confidence %q is not a probability in [0,1]", f[1])
	}
	within, err := strconv.Atoi(f[2])
	if err != nil || within <= 0 {
		return Triple{}, 0, fmt.Errorf("within %q is not a positive number of traversals; a bet with no deadline is never settled", f[2])
	}
	return Triple{
		Mode: mode, Confidence: conf,
		Subject: f[3], Predicate: f[4], Object: obj,
	}, within, nil
}

// parseRequiresLine reads:  <mode> <subject> <predicate> <object>
//
// No confidence and no within: a `requires` is a condition about NOW, not a bet about later.
func parseRequiresLine(line string) (Triple, error) {
	f, obj, err := split(line, 3)
	if err != nil {
		return Triple{}, err
	}
	mode, err := parseMode(f[0])
	if err != nil {
		return Triple{}, err
	}
	return Triple{Mode: mode, Subject: f[1], Predicate: f[2], Object: obj}, nil
}

// split takes n leading whitespace-separated fields, then treats the remainder as the object —
// unquoting it if it is quoted, so an object may contain spaces.
func split(line string, n int) ([]string, string, error) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < n+1 {
		return nil, "", fmt.Errorf("expected at least %d fields, got %d", n+1, len(fields))
	}
	head := fields[:n]
	rest := strings.TrimSpace(strings.Join(fields[n:], " "))
	if len(rest) >= 2 && rest[0] == '"' && rest[len(rest)-1] == '"' {
		rest = rest[1 : len(rest)-1]
	}
	if rest == "" {
		return nil, "", fmt.Errorf("empty object")
	}
	return head, rest, nil
}

func parseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeExact:
		return ModeExact, nil
	case ModeInterpret:
		return ModeInterpret, nil
	}
	return "", fmt.Errorf("unknown mode %q (want exact or interpret)", s)
}

// --- tolerant element scanning ---

type element struct {
	open string // the opening tag, e.g. `<act id="a1" primitive="write">`
	body string // everything between open and close; "" for a self-closing tag
}

// elements finds every <name ...>...</name> and <name .../> in raw. Anything it cannot close is
// skipped — which is what makes a truncated emission yield its complete elements and drop only the
// one that got cut.
func elements(raw, name string) []element {
	var out []element
	for i := 0; ; {
		start := strings.Index(raw[i:], "<"+name)
		if start < 0 {
			return out
		}
		start += i
		// The opening tag ends at the first '>' — payloads live after it, so this is safe.
		gt := strings.Index(raw[start:], ">")
		if gt < 0 {
			return out // truncated mid-tag; nothing further is usable
		}
		gt += start
		open := raw[start : gt+1]

		if strings.HasSuffix(open, "/>") { // self-closing
			out = append(out, element{open: open})
			i = gt + 1
			continue
		}
		close := strings.Index(raw[gt:], "</"+name+">")
		if close < 0 {
			return out // truncated before the close; drop this one, keep what came before
		}
		close += gt
		out = append(out, element{open: open, body: raw[gt+1 : close]})
		i = close + len(name) + 3
	}
}

func inner(raw, name string) string {
	els := elements(raw, name)
	if len(els) == 0 {
		return ""
	}
	return els[0].body
}

// attr reads key="value" out of an opening tag.
func attr(tag, key string) string {
	k := key + `="`
	i := strings.Index(tag, k)
	if i < 0 {
		return ""
	}
	rest := tag[i+len(k):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

package dnav3

import (
	"fmt"
	"html"
	"regexp"
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
		localID := strings.TrimSpace(attr(el.open, "id"))
		if len(of) == 0 {
			res.Rejects = append(res.Rejects, Reject{Line: el.open, Reason: `expect has no of="..."`})
			continue
		}
		// An <expect> block may carry several triple lines; each is independent.
		lines := strings.Split(el.body, "\n")
		nonempty := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonempty++
			}
		}
		lineIndex := 0
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			t, within, err := parseExpectLine(line)
			if err != nil {
				res.Rejects = append(res.Rejects, Reject{Line: strings.TrimSpace(line), Reason: err.Error()})
				continue
			}
			predictionID := PredictionID(localID)
			if localID != "" && nonempty > 1 {
				predictionID = PredictionID(fmt.Sprintf("%s:%d", localID, lineIndex))
			}
			res.Expression.Expects = append(res.Expression.Expects, Expect{
				PredictionID: predictionID, DimensionID: "exact",
				Of: of, Triple: t, Within: within,
			})
			lineIndex++
		}
	}

	for _, el := range elements(raw, "evidence") {
		attribution, err := parseEvidenceAttribution(strings.TrimSpace(el.body))
		if err != nil {
			res.Rejects = append(res.Rejects, Reject{Line: strings.TrimSpace(el.body), Reason: err.Error()})
			continue
		}
		res.Expression.PredictionAttributions = append(res.Expression.PredictionAttributions, attribution)
	}

	if len(res.Expression.Acts) == 0 {
		return res, fmt.Errorf("emission contains no acts")
	}
	return res, nil
}

var evidenceAttributionPattern = regexp.MustCompile(
	`^\[([^\]]+)\]\[(\d+):(\d+)\]:([+-]?(?:\d+(?:\.\d*)?|\.\d+))->\[([^\.\]]+)\.(.+)\]$`,
)

func parseEvidenceAttribution(line string) (EvidenceAttribution, error) {
	match := evidenceAttributionPattern.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return EvidenceAttribution{}, fmt.Errorf("evidence must be [observation][start:end]:influence->[prediction.dimension]")
	}
	start, err := strconv.Atoi(match[2])
	if err != nil {
		return EvidenceAttribution{}, fmt.Errorf("evidence start offset is invalid")
	}
	end, err := strconv.Atoi(match[3])
	if err != nil {
		return EvidenceAttribution{}, fmt.Errorf("evidence end offset is invalid")
	}
	influence, err := strconv.ParseFloat(match[4], 64)
	if err != nil {
		return EvidenceAttribution{}, fmt.Errorf("evidence influence is invalid")
	}
	return EvidenceAttribution{
		ObservationID: ObservationID(match[1]), StartChar: start, EndChar: end,
		PredictionID: PredictionID(match[5]), DimensionID: DimensionID(match[6]),
		ClaimedInfluence: influence,
	}, nil
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
		// UNESCAPE XML ENTITIES. The first live run found this: the model emitted a correct plan —
		// `sh /greet.sh > /out.txt` — but XML-escaped the redirect to `&gt;`. The shell then ran it
		// literally (& backgrounds, `gt` not found), out.txt was never created, and a perfectly good
		// plan was scored as a failure the model did not commit. The payload is content, not markup;
		// entities in it are the envelope leaking in, and they come back out here.
		Payload: html.UnescapeString(strings.TrimSpace(el.body)),
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

// parseExpectLine reads:  <confidence> <within> <subject> <predicate> <object>
//
// THERE IS NO MODE TOKEN, and that is deliberate. Mode.Weight() is the GRADE MULTIPLIER, so a model
// choosing its own mode would be choosing its own stakes — marking `interpret` on something
// measurable cuts its losses to 0.3x, which lets a badly-calibrated model DAMPEN ITS OWN SIGNAL.
// That is the exact failure this system exists to detect.
//
// Mode is a routing instruction for the incoming impulse: how does an observation resolve against
// this claim? That is a question about the WORLD'S INSTRUMENTS, not the model's intent. The
// substrate stamps it from the affordance registry. See affordance.go.
//
// The object is last and may be quoted, so it can carry spaces. Everything before it is a fixed
// number of fields, which is what makes a single malformed line cheap to drop.
func parseExpectLine(line string) (Triple, int, error) {
	f, obj, err := split(line, 4)
	if err != nil {
		return Triple{}, 0, err
	}
	conf, err := strconv.ParseFloat(f[0], 64)
	if err != nil || conf < 0 || conf > 1 {
		return Triple{}, 0, fmt.Errorf("confidence %q is not a probability in [0,1]", f[0])
	}
	within, err := strconv.Atoi(f[1])
	if err != nil || within <= 0 {
		return Triple{}, 0, fmt.Errorf("within %q is not a positive number of traversals; a bet with no deadline is never settled", f[1])
	}
	return Triple{
		Confidence: conf,
		Subject:    f[2], Predicate: f[3], Object: obj,
	}, within, nil
}

// parseRequiresLine reads:  <subject> <predicate> <object>
//
// No confidence, no within, no mode: a `requires` is a condition about NOW, not a bet about later,
// and its mode is the world's to say, not the model's.
func parseRequiresLine(line string) (Triple, error) {
	f, obj, err := split(line, 2)
	if err != nil {
		return Triple{}, err
	}
	return Triple{Subject: f[0], Predicate: f[1], Object: obj}, nil
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

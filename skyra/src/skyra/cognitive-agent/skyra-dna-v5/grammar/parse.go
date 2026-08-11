package grammar

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// The forms, exactly as grammar.md writes them:
//
//	<select>
//	  <engramName cite=obsID entity=entID host=hostID> obsID, obsID </engram>
//	  <engramName cite=obsID entity=entID>             obsID         </engram>
//	</select>
//
//	<commit><observationID></commit>          an orchestrator NAMES
//	<commit>{ … }</commit>                    a leaf ORIGINATES
//
// Note the asymmetry the spec specifies and this parser honours: a member's
// OPENING tag is the engram NAME, and its closing tag is the literal </engram>.
// The name is in tag position because it is the one thing every member must have;
// everything else is an attribute.
var (
	reSelect  = regexp.MustCompile(`(?s)<select>(.*?)</select>`)
	reCommit  = regexp.MustCompile(`(?s)<commit>(.*?)</commit>`)
	reMember  = regexp.MustCompile(`(?s)<([A-Za-z_][A-Za-z0-9_\-]*)\s*([^>]*)>(.*?)</engram>`)
	reAttr    = regexp.MustCompile(`([a-z]+)\s*=\s*"?([^"\s>]+)"?`)
	reObsID   = regexp.MustCompile(`^<?(obs-\d+)>?$`)
	reRefStr  = regexp.MustCompile(`^<(obs-\d+)>$`)
	reWorldRe = regexp.MustCompile(`^<world (.+)@(.+)>$`)
)

// ErrNoEmission is what a response that parses as neither form produces.
//
// This is the incapacity, and it is not an error condition to be smoothed over:
// a possibility that emitted something else emitted nothing. It is an ordinary
// failed invocation, observed and recorded on the INVOKER'S engram, and the
// invoker's selection is what learns from it.
type ErrNoEmission struct{ Raw string }

func (e ErrNoEmission) Error() string {
	r := strings.TrimSpace(e.Raw)
	if len(r) > 300 {
		r = r[:300] + "…"
	}
	return fmt.Sprintf("no emission parsed: neither <select> nor <commit> present in %q", r)
}

// Parse extracts the emission from whatever an entity produced.
//
// An entity may say anything around its emission — a model narrates, a shell
// prints warnings. The grammar's claim is about what can be EMITTED, not about
// what can be uttered, so surrounding text is ignored rather than rejected.
//
// SELECT is checked first. A response carrying both forms is a select whose
// narration mentioned a commit; the reverse ordering would let a stray mention of
// a select turn a terminal composition into an invocation.
func Parse(raw string) (Emission, error) {
	if m := reSelect.FindStringSubmatch(raw); m != nil {
		return parseSelect(m[1])
	}
	if m := reCommit.FindStringSubmatch(raw); m != nil {
		return parseCommit(m[1])
	}
	return nil, ErrNoEmission{Raw: raw}
}

func parseSelect(body string) (Emission, error) {
	ms := reMember.FindAllStringSubmatch(body, -1)
	if len(ms) == 0 {
		return nil, fmt.Errorf("select with no members: %q", strings.TrimSpace(body))
	}

	sel := Select{}
	for _, m := range ms {
		name, attrs, args := m[1], m[2], m[3]
		mem := Member{Engram: name}

		for _, a := range reAttr.FindAllStringSubmatch(attrs, -1) {
			switch a[1] {
			case "cite":
				mem.Cite = ids.ObservationID(strings.Trim(a[2], "<>"))
			case "entity":
				mem.Entity = ids.EntityID(a[2])
			case "host":
				mem.Host = ids.HostID(a[2])
			default:
				// An unknown attribute is not silently dropped. The vocabulary is
				// the claim; accommodating an invented one is how a third form
				// arrives.
				return nil, fmt.Errorf("member %q: unknown attribute %q", name, a[1])
			}
		}

		for _, tok := range strings.Split(args, ",") {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			g := reObsID.FindStringSubmatch(tok)
			if g == nil {
				return nil, fmt.Errorf("member %q: %q is not an observation id", name, tok)
			}
			mem.Observations = append(mem.Observations, ids.ObservationID(g[1]))
		}

		if mem.Cite == "" {
			// cite is fixed in the GRAMMAR rather than in the metric, because an
			// observation that supplied the NAME is never bound as an argument
			// and would otherwise fall through as unused — and that is the
			// observation that made the invocation possible at all.
			return nil, fmt.Errorf("member %q: cite is required — every emission cites its sources", name)
		}
		if mem.Entity == "" {
			return nil, fmt.Errorf("member %q: entity is required — WHERE sits in every signature", name)
		}
		sel.Members = append(sel.Members, mem)
	}
	return sel, nil
}

func parseCommit(body string) (Emission, error) {
	b := strings.TrimSpace(body)
	if b == "" {
		return nil, fmt.Errorf("empty commit")
	}

	// An orchestrator NAMES: <commit><obs-0007></commit>
	if g := reObsID.FindStringSubmatch(b); g != nil {
		return Commit{Observed: ids.ObservationID(g[1])}, nil
	}

	// A leaf ORIGINATES: <commit>{ … }</commit>
	b = stripFence(b)
	var raw map[string]any
	if err := json.Unmarshal([]byte(b), &raw); err != nil {
		return nil, fmt.Errorf("commit body is neither an observation id nor an object: %w", err)
	}
	c := content.Content{}
	for k, v := range raw {
		c[k] = toValue(v)
	}
	if len(c) == 0 {
		return nil, fmt.Errorf("commit originated an empty observation")
	}
	return Commit{Content: c}, nil
}

// stripFence removes a markdown code fence. Models wrap JSON in one by reflex,
// and refusing it would be rejecting an emission over its packaging.
func stripFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "```"))
}

// toValue converts decoded JSON into the closed Value set.
//
// This is where a Ref re-enters the system. An entity carries <obs-0004> forward
// verbatim from what it was shown, and the string form converts back to a Ref —
// which is how CITATION IS DERIVED FROM WHAT WAS ACTUALLY WRITTEN rather than
// from anything the entity declared.
func toValue(v any) content.Value {
	switch t := v.(type) {
	case string:
		if g := reRefStr.FindStringSubmatch(t); g != nil {
			return content.Ref(g[1])
		}
		if g := reWorldRe.FindStringSubmatch(t); g != nil {
			return content.WorldRef{Handle: g[1], Host: ids.HostID(g[2])}
		}
		return content.Str(t)
	case float64:
		return content.Num(t)
	case bool:
		return content.Bool(t)
	case []any:
		out := make(content.List, 0, len(t))
		for _, e := range t {
			out = append(out, toValue(e))
		}
		return out
	case map[string]any:
		out := content.Obj{}
		for k, e := range t {
			out[k] = toValue(e)
		}
		return out
	case nil:
		return content.Str("")
	}
	return content.Str(fmt.Sprint(v))
}

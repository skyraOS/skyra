package content

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// maxDeref bounds how far a fingerprint chases references.
//
// interfaces.md flags the cost: dereferencing walks a chain, so comparison depth
// is a real cost on deeply nested content. Bounded rather than unbounded because
// an unbounded walk over a graph with a cycle does not terminate, and nothing
// structurally prevents an observation whose content refs a descendant of itself.
// Past the bound the ref contributes its ID, which degrades toward the identity
// comparison this exists to avoid — that degradation is visible in the fingerprint
// (a "…@" marker) rather than silent.
const maxDeref = 4

// Deref resolves a reference to the content behind it. Supplied by the graph,
// which is the only thing that holds observations.
type Deref func(ids.ObservationID) (Content, bool)

// Fingerprint reduces content to a comparison key, dereferencing Refs so two
// askings that carry the same material compare equal even when the observations
// carrying it are distinct instances.
//
// This is the half of "what counts as the same question" that lives in content.
// Comparing Refs by ID would make every question unique again, which is exactly
// what killed the first InvocationKey.
//
// A key rather than a comparison function is a deliberate narrowing of what
// interfaces.md leans toward ("a comparison function rather than a hashed key, so
// it stays soft"). Recorded as a divergence in the build notes: a key is what a
// map needs, and the fold is read-time so it can be replaced without migration.
func Fingerprint(c Content, d Deref) string {
	return fingerprint(Obj(c), d, 0)
}

func fingerprint(v Value, d Deref, depth int) string {
	switch t := v.(type) {
	case Str:
		return "s:" + string(t)
	case Num:
		return fmt.Sprintf("n:%g", float64(t))
	case Bool:
		return fmt.Sprintf("b:%t", bool(t))
	case EngramRef:
		// Reach is not part of what makes two askings the same question — it is
		// capability, not material.
		return "e:" + string(t.ID)
	case WorldRef:
		// A handle is an opaque identity. Two handles are the same question only
		// if they are literally the same handle on the same host.
		return fmt.Sprintf("w:%s@%s", t.Handle, t.Host)
	case Ref:
		if depth >= maxDeref {
			return "…@" + string(t)
		}
		if d != nil {
			if inner, ok := d(ids.ObservationID(t)); ok {
				return "r(" + fingerprint(Obj(inner), d, depth+1) + ")"
			}
		}
		// Unresolvable: fall back to identity and say so.
		return "?@" + string(t)
	case List:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, fingerprint(e, d, depth))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case Obj:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+"="+fingerprint(t[k], d, depth))
		}
		return "{" + strings.Join(parts, ",") + "}"
	}
	return "?"
}

// FingerprintSet folds a set of contents order-independently.
//
// One universal signature gives position no meaning, so argument order cannot be
// part of what makes two askings the same.
func FingerprintSet(cs []Content, d Deref) string {
	parts := make([]string, 0, len(cs))
	for _, c := range cs {
		parts = append(parts, Fingerprint(c, d))
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

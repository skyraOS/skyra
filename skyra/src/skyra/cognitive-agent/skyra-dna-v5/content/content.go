// Package content holds what an observation carries, and the two read-time folds
// the history partitions by.
//
// The return path is always structured (interfaces.md): behaviour is polymorphic
// because there are many ways to act; evidence is uniform because there is one way
// to learn. Every mechanism downstream of a return — same-question comparison,
// composition, the credit walk, shape partitioning — runs on this one shape.
//
// Structure is not correctness. A well-formed return is not a good return.
package content

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Content is what an observation says.
type Content map[string]Value

// Value is the closed set of things content can hold. Closed rather than
// interface{} because Cites is DERIVED by scanning for Refs, and a scan over an
// open set cannot be exhaustive.
type Value interface{ isValue() }

type (
	Str  string
	Num  float64
	Bool bool
	List []Value
	Obj  map[string]Value
)

// Ref is the one that matters: an observation embedded in another's content.
// Every Ref appearing in a commit's content IS a citation. There is no declared
// citation list to disagree with it.
type Ref ids.ObservationID

// WorldRef is a pointer INTO THE WORLD — a rendered page, a running session, a
// file. It carries its host, and resolving it anywhere else is an ERROR, never a
// lookup: a path can resolve on another host to a different thing with the same
// name, and a silent wrong answer is worse than a crash.
type WorldRef struct {
	Host   ids.HostID
	Handle string
}

// EngramRef is REACH — a name that becomes selectable in the frame that receives
// it. It is what discover hands back.
//
// It is a DISTINCT TYPE on purpose, and that is the whole enforcement of the line
// grammar.md draws:
//
//	DELEGATION   reach arriving from an entity I CHOSE to invoke.  open() handing
//	             back a descriptor. legitimate.
//	INJECTION    reach arriving in the data channel.               whoever can talk
//	             to me picks what I can touch.
//
// Same bytes, different provenance, and only the type keeps them apart. The
// kernel honours an EngramRef appearing in a RETURN and ignores one appearing in
// an argument or at ingress — see kernel.fold.
//
// Note what it is NOT: a Ref. It never enters Cites, so reach carries no credit
// and the credit walk cannot descend through it.
type EngramRef struct {
	ID   ids.EngramID
	Name string // how it becomes nameable in the receiving frame
}

func (Str) isValue()       {}
func (Num) isValue()       {}
func (Bool) isValue()      {}
func (List) isValue()      {}
func (Obj) isValue()       {}
func (Ref) isValue()       {}
func (WorldRef) isValue()  {}
func (EngramRef) isValue() {}

// Reach collects the EngramRefs in content. Only ever consulted on a return.
func (c Content) Reach() []EngramRef {
	var out []EngramRef
	walkValue(Obj(c), func(v Value) {
		if e, ok := v.(EngramRef); ok {
			out = append(out, e)
		}
	})
	return out
}

// Cites derives the citation list by scanning content for Refs.
//
// Never declared. Two places for one fact would eventually disagree — a commit
// could claim it cited A while its content references B, and nothing would catch
// it. Deriving it makes CREDIT FOLLOWS DATA structural rather than agreed:
//
//	IN YOUR OUTPUT       there was data flow. cited, BY CONSTRUCTION.
//	NOT IN YOUR OUTPUT   influence at most. UNCITABLE, not merely uncited.
//
// Order is deterministic (first appearance, depth-first over sorted keys) so a
// replay produces the identical list.
func (c Content) Cites() []ids.ObservationID {
	var out []ids.ObservationID
	seen := map[ids.ObservationID]bool{}
	add := func(id ids.ObservationID) {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	walkValue(Obj(c), func(v Value) {
		switch t := v.(type) {
		case Ref:
			add(ids.ObservationID(t))
		case Str:
			// A reference EMBEDDED IN PROSE is still data flow.
			//
			// FOUND BY RUNNING IT. The first real model return put the reference
			// at the end of a sentence — "…without being flushed to disk after
			// every change. <obs-0001>" — inside a summary string. Matching only
			// whole-string refs made that commit cite NOTHING, so the chain that
			// genuinely used its input recorded no data flow at all.
			//
			// interfaces.md anticipated the blended case and predicted the credit
			// would be COARSE. What actually happens is that it is ZERO, which is
			// a different and worse failure: a model writing naturally never cites
			// anything, and CREDIT FOLLOWS DATA silently stops being true.
			//
			// Scanning inside strings is what keeps derivation honest. It admits
			// ids the emitter never had, which is exactly what the SCOPE check at
			// the graph boundary is for — so this widens what is DERIVED and
			// changes nothing about what is AUTHORIZED.
			for _, m := range reEmbedded.FindAllStringSubmatch(string(t), -1) {
				add(ids.ObservationID(m[1]))
			}
		}
	})
	return out
}

var reEmbedded = regexp.MustCompile(`<(obs-\d+)>`)

func walkValue(v Value, fn func(Value)) {
	fn(v)
	switch t := v.(type) {
	case List:
		for _, e := range t {
			walkValue(e, fn)
		}
	case Obj:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			walkValue(t[k], fn)
		}
	}
}

// ---------------------------------------------------------------------------
// Shape — the first of the two read-time folds
// ---------------------------------------------------------------------------

// Shape is what makes two returns comparable. interfaces.md calls its definition
// "THE ONLY REAL DECISION in the partitioning idea" and leaves it open. Settled
// here as: the sorted set of field paths, each paired with its value KIND.
//
//	{tests: 480} vs {tests: 500}      same shape. comparable. THE WORLD MOVED.
//	{tests: 480} vs {error: timeout}  different shape. SOMETHING STRUCTURAL HAPPENED.
//
// A nested Ref counts as a shape ELEMENT of kind "ref" and is not dereferenced.
// Shape is the form of the answer, and a reference is a form — dereferencing here
// would make shape depend on what the referent happened to contain, which is a
// value fact leaking into a structural one.
//
// List elements collapse to one path with the union of their kinds, so a
// three-item list and a four-item list are the same shape. Length is a value
// change, not a structural one.
//
// Derived at read time. No field, no migration, and changing this definition
// invalidates nothing already written.
func (c Content) Shape() string {
	paths := map[string]map[string]bool{}
	collectShape("", Obj(c), paths)

	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		kinds := make([]string, 0, len(paths[k]))
		for kind := range paths[k] {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		fmt.Fprintf(&b, "%s:%s", k, strings.Join(kinds, "|"))
	}
	if b.Len() == 0 {
		return "{}"
	}
	return b.String()
}

func collectShape(path string, v Value, out map[string]map[string]bool) {
	switch t := v.(type) {
	case Obj:
		if path != "" {
			record(out, path, "obj")
		}
		for k, e := range t {
			child := k
			if path != "" {
				child = path + "." + k
			}
			collectShape(child, e, out)
		}
	case List:
		record(out, path, "list")
		for _, e := range t {
			collectShape(path+"[]", e, out)
		}
	default:
		record(out, path, kindOf(v))
	}
}

func record(out map[string]map[string]bool, path, kind string) {
	if out[path] == nil {
		out[path] = map[string]bool{}
	}
	out[path][kind] = true
}

func kindOf(v Value) string {
	switch v.(type) {
	case Str:
		return "str"
	case Num:
		return "num"
	case Bool:
		return "bool"
	case Ref:
		return "ref"
	case WorldRef:
		return "worldref"
	case EngramRef:
		return "engramref"
	case List:
		return "list"
	case Obj:
		return "obj"
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

// Render writes content in the form an entity reads it. Refs render as <obs-id>
// so an interpreter can carry one forward verbatim into its own commit — which is
// how citation gets derived from what it actually wrote.
func (c Content) Render() string {
	return renderValue(Obj(c), 0)
}

func renderValue(v Value, depth int) string {
	pad := strings.Repeat("  ", depth)
	switch t := v.(type) {
	case Str:
		return string(t)
	case Num:
		if t == Num(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", float64(t))
	case Bool:
		return fmt.Sprintf("%t", bool(t))
	case Ref:
		return fmt.Sprintf("<%s>", string(t))
	case WorldRef:
		return fmt.Sprintf("<world %s@%s>", t.Handle, t.Host)
	case EngramRef:
		return fmt.Sprintf("<engram %s named %q>", t.ID, t.Name)
	case List:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, renderValue(e, depth+1))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case Obj:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "\n%s  %s: %s", pad, k, renderValue(t[k], depth+1))
		}
		return b.String()
	}
	return "?"
}

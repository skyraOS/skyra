// Package frame holds what a possibility SEES.
//
// An entity is handed a frame to interpret; the perspective belongs to the
// possibility. Same object, restricted — it cannot reach what is not here, so
// DOWN NEVER UP stops being a rule and becomes a scope.
//
// The frame is also a BUDGET boundary. Everything a possibility sees ships to the
// entity's host on every step, so what it may see is simultaneously the learning
// mechanism and the per-step transfer cost, and the two pull against each other.
package frame

import (
	"sort"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Frame is what a possibility knows. Three things and nothing else: the
// observations it was given, which engram it is opening, and the observations its
// children returned.
type Frame interface {
	Self() ids.EngramID

	Entity() ids.EntityID     // who is interpreting ME
	Entities() []ids.EntityID // who I may ADDRESS. scoped, exactly like names.

	Observations() []*graph.Observation
	Returned() []*graph.Observation

	// Strips Cites. A possibility sees what its children CONCLUDED, never what
	// they were told.
	Engrams() View

	// Resolve turns a NAME into an engram. This is the whole of capability:
	// you can select what you can name; you can name what you were given.
	Resolve(name string) (*graph.Engram, bool)

	// InScope answers the graph-boundary authorization: was this ID in the
	// emitter's frame? Arithmetic, total, cheap, no policy.
	InScope(id ids.ObservationID) bool
}

// View is the restricted look at an engram's children.
//
// Handing an entity the raw engram would let it reach Invocation.Return.Cites and
// traverse into a child's subtree, and "the parent sees what the child concluded,
// not what it was told" stops holding.
type View interface {
	Names() []string
	Summary(name string) (graph.Summary, bool)
}

// ---------------------------------------------------------------------------

// Live is the concrete frame the kernel builds for a possibility.
type Live struct {
	Engram *graph.Engram
	Ent    ids.EntityID

	// Who this possibility may address. Entity reach is INHERITABLE: a parent
	// hands down a SUBSET, which is how a limb gets narrower authority than what
	// spawned it. Not optional — without it, either entity names resolve globally
	// (a registry, which world.md refuses) or a possibility can only run children
	// on its own entity, which makes delegation impossible.
	Reach []ids.EntityID

	Obs      []*graph.Observation
	Rets     []*graph.Observation
	Impulses []*graph.Observation

	// Names delegated into this frame by a child's return — discover handing back
	// an engram it resolved.
	//
	// This is DELEGATION, not injection: reach arriving from an entity I CHOSE to
	// invoke, the same shape as open() handing back a descriptor. The line the
	// grammar draws is that reach must not travel in the DATA channel; it arrives
	// here through a return, which is a channel I opened.
	Delegated map[string]*graph.Engram

	// Names every frame has. discover is nameable everywhere — a builtin name,
	// not something you must discover in order to discover. That is what stops
	// the regress.
	Builtin map[string]*graph.Engram

	G *graph.Graph
}

func (f *Live) Self() ids.EngramID                 { return f.Engram.ID }
func (f *Live) Entity() ids.EntityID               { return f.Ent }
func (f *Live) Entities() []ids.EntityID           { return f.Reach }
func (f *Live) Observations() []*graph.Observation { return f.Obs }

// Returned includes buffered impulses, which land on the SAME boundary as
// monitor. There should be one boundary in this design, not two — otherwise the
// world interrupts at arbitrary points in an entity's reasoning and arrival
// timing becomes semantic.
func (f *Live) Returned() []*graph.Observation {
	return append(append([]*graph.Observation{}, f.Rets...), f.Impulses...)
}

func (f *Live) Engrams() View { return &view{e: f.Engram, g: f.G} }

// Resolve looks in three places, nearest first.
//
//	CHILDREN    what has WORKED here. formed on citation.
//	DELEGATED   what a child handed back this invocation.
//	BUILTIN     discover. nameable in every frame.
//
// A name absent from all three is unresolvable, and that is the whole enforcement
// of capability-as-scoping: a possibility cannot address what was never put in
// its frame. There is no lookup of last resort.
func (f *Live) Resolve(name string) (*graph.Engram, bool) {
	if e, ok := f.Engram.Engrams[name]; ok {
		return e, true
	}
	if e, ok := f.Delegated[name]; ok {
		return e, true
	}
	if e, ok := f.Builtin[name]; ok {
		return e, true
	}
	return nil, false
}

func (f *Live) InScope(id ids.ObservationID) bool {
	for _, o := range f.Obs {
		if o.ID == id {
			return true
		}
	}
	for _, o := range f.Rets {
		if o.ID == id {
			return true
		}
	}
	for _, o := range f.Impulses {
		if o.ID == id {
			return true
		}
	}
	return false
}

// Names lists everything selectable from this frame, for rendering.
func (f *Live) Names() []string {
	seen := map[string]bool{}
	var out []string
	add := func(m map[string]*graph.Engram) {
		for n := range m {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	add(f.Engram.Engrams)
	add(f.Delegated)
	add(f.Builtin)
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------

type view struct {
	e *graph.Engram
	g *graph.Graph
}

func (v *view) Names() []string {
	out := make([]string, 0, len(v.e.Engrams))
	for n := range v.e.Engrams {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (v *view) Summary(name string) (graph.Summary, bool) {
	c, ok := v.e.Engrams[name]
	if !ok {
		return graph.Summary{}, false
	}
	return graph.Summarize(name, c, v.g.Deref()), true
}

var _ content.Deref = (func(ids.ObservationID) (content.Content, bool))(nil)

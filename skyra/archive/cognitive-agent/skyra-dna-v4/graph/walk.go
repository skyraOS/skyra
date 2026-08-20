package graph

import (
	"fmt"

	"skyra-v1/src/skyra/spacetime/kernel"
)

// A Walk is a frontier and a fold. It implements kernel.Space and is NOT a Handle — it has
// no durable node and no behavior, so nothing dispatches on it.
//
//	HANDLE   a durable node + a behavior.   the coupling.
//	WALK     a frontier + a fold.           ephemeral only.
//
// It exists because a behavior cannot traverse: one `Perturb` yields one Space, so a worker
// that wanted to visit five candidates would visit one and stop. The worker builds the
// frontier and hands off a Walk; the Walk visits.
//
// # It is a cursor, not a continuation chain
//
// The first sketch of this made each hop return "the rest of the walk" — continuation
// passing, one Space per member. The iterating `Perturb` makes that unnecessary: ONE frame
// holds the cursor and yields members one at a time, and the members come out as siblings
// in the transcript rather than as a chain pretending to be nesting.
//
// # It writes nothing durable
//
// Members are opened and folded; nothing is committed. So an evicted Walk has changed
// nothing at all, which is the strongest form of the eviction guarantee in phases.md.
type Walk struct {
	id      string
	ns      *Namespace
	as      Behavior // what each member is opened with
	members []*Node
	at      int

	// Visited is the members whose realization came back, in fold order. The worker
	// above reads this on ascent; it is the whole point of the walk.
	Visited []*Node

	// Accumulation is the derived arithmetic gathered across the frontier. Stored
	// nowhere; dies with the walk.
	Accumulation Accumulation
}

var _ kernel.Space = (*Walk)(nil)

// NewWalk builds a frontier. Members are opened with `as` when they are reached, never
// before — the coupling is per-visit, like every other Space.
func NewWalk(id string, ns *Namespace, as Behavior, members ...*Node) *Walk {
	return &Walk{id: id, ns: ns, as: as, members: members}
}

func (w *Walk) ID() string { return w.id }

// Len is how many members this walk was built over.
func (w *Walk) Len() int { return len(w.members) }

// Perturb yields the next member, or nil when the frontier is exhausted.
//
// Returning nil eventually is mandatory, not incidental: a frame is asked repeatedly, and a
// Space that keeps handing back the same thing never terminates.
func (w *Walk) Perturb(origin kernel.Space, impulse any) kernel.Space {
	if w == nil || w.at >= len(w.members) {
		return nil
	}
	m := w.members[w.at]
	w.at++
	return w.ns.OpenWith(m, w.as)
}

// Observe folds one member's realization. Ephemeral only — the Walk records that it got
// there and returns itself, so the worker above receives the whole visited set in one piece.
func (w *Walk) Observe(origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	if h, ok := realized.(*Handle); ok && h != nil {
		w.Visited = append(w.Visited, h.Node())
		w.Accumulation.absorb(h.Accumulation)
	}
	return w
}

func (w *Walk) String() string {
	return fmt.Sprintf("walk %s [%d/%d]", w.id, w.at, len(w.members))
}

// absorb folds another accumulation into this one. Sums, because T and G are sums and |O|
// is a count — there is no ordering assumption and none is wanted.
func (a *Accumulation) absorb(b Accumulation) {
	a.T += b.T
	a.G += b.G
	a.Count += b.Count
}

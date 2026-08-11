package graph

import (
	"sync"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Graph is the witness. It is the only thing with policy — engram formation, the
// local verdict, credit dispersal.
//
// The kernel pairs, drives, and produces a RECORD of what was emitted; it has no
// policy and performs no write. If an interpreter performed the write it could
// omit it, which is the self-reporting defect the honest-witness argument exists
// to prevent.
type Graph struct {
	mu sync.RWMutex

	// Every engram ever minted, by ID. This is what DISCOVERY walks, and what the
	// credit walk resolves Producer against.
	//
	// Two filters, easy to conflate:
	//	MINTED ON CLOSURE   an engram exists at all.   records what was TRIED.
	//	FORMED ON CITATION  it becomes a CHILD.        records what WORKED.
	//
	// Discovery sees attempts; selection sees successes. An orphan in this index
	// that nobody ever cited is memoization, not bloat: the second time a need
	// arises, discovery finds it instead of re-improvising it.
	engrams map[ids.EngramID]*Engram

	// Roots are held APART from what discovery searches.
	//
	// FOUND BY RUNNING IT, and it is the resolution of a tension the design
	// cannot resolve on its own. A root's expectation IS the ask, so it is the
	// best lexical match for its own chain's need — by-index proposed the root as
	// the answer to the root's own question on every single ask. The kernel can
	// refuse the binding, because only the kernel holds ancestry. But it has NO
	// WAY TO TELL DISCOVERY, because a possibility may never look upward, and the
	// grammar has no third channel to carry the refusal. So the refusal is a dead
	// end rather than a redirect, and the chain livelocks until eviction.
	//
	// The cut that works is upstream of all of it: A ROOT IS AN ASK, NOT A
	// CAPABILITY. Mint-on-closure exists to make capabilities that have proven
	// they can close discoverable; a root has nothing to offer a later chain
	// except the whole chain. It still persists, is still reused when the same ask
	// recurs, still owns its children, and still accumulates every bit of the
	// learning. It is simply not a search result.
	//
	// THE SPEC HAS TO CHOOSE HERE. It says every engram is minted on closure and
	// says discovery walks the index; put together those two sentences make every
	// chain able to rediscover itself.
	roots map[ids.EngramID]*Engram

	// Every observation ever committed or ingressed. The walk needs it to resolve
	// a Ref to its Producer, and the same-question fold needs it to dereference.
	observations map[ids.ObservationID]*Observation

	// Recorded discards of staged world effects. A discard is an EVENT, never a
	// silent no-op — citation was designed with cheap failure and a lost write is
	// not that.
	discards []Discard
}

func New() *Graph {
	return &Graph{
		engrams:      map[ids.EngramID]*Engram{},
		roots:        map[ids.EngramID]*Engram{},
		observations: map[ids.ObservationID]*Observation{},
	}
}

// Discard records a staged effect the credit walk never reached.
type Discard struct {
	Span   ids.SpanID
	Engram ids.EngramID
	Detail string
}

// ---------------------------------------------------------------------------
// The index
// ---------------------------------------------------------------------------

// Mint puts an engram in the index. This is mint-on-closure: an expectation is
// improvised, it runs, and if it CLOSES it persists. Creation is not a separate
// authoring step.
//
// An engram that exists here has proven it can close. That is what makes the
// index safe to discover from: nothing in it is a promise.
func (g *Graph) Mint(e *Engram) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.engrams[e.ID] = e
}

// MintRoot persists a root. It closed, so it persists — but it is not offered to
// discovery. See the roots field.
func (g *Graph) MintRoot(e *Engram) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.roots[e.ID] = e
}

// Roots returns every persisted root. The kernel reuses one when the same ask
// recurs; nothing else reads this.
func (g *Graph) Roots() []*Engram {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]*Engram, 0, len(g.roots))
	for _, e := range g.roots {
		out = append(out, e)
	}
	return out
}

// Engram resolves by ID across both, because the credit walk must be able to
// reach anything that produced an observation.
func (g *Graph) Engram(id ids.EngramID) (*Engram, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if e, ok := g.engrams[id]; ok {
		return e, true
	}
	e, ok := g.roots[id]
	return e, ok
}

func (g *Graph) Observation(id ids.ObservationID) (*Observation, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	o, ok := g.observations[id]
	return o, ok
}

// Store records an observation. Called for ingress, and by Apply for commits.
func (g *Graph) Store(o *Observation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.observations[o.ID] = o
}

// Deref supplies the same-question fold with reference resolution.
func (g *Graph) Deref() content.Deref {
	return func(id ids.ObservationID) (content.Content, bool) {
		o, ok := g.Observation(id)
		if !ok {
			return nil, false
		}
		return o.Content, true
	}
}

// Engrams returns the discoverable index — capabilities, never roots.
// Discovery's raw material.
func (g *Graph) Engrams() []*Engram {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]*Engram, 0, len(g.engrams))
	for _, e := range g.engrams {
		out = append(out, e)
	}
	return out
}

func (g *Graph) Discards() []Discard {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return append([]Discard(nil), g.discards...)
}

func (g *Graph) RecordDiscard(d Discard) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.discards = append(g.discards, d)
}

// ---------------------------------------------------------------------------
// The write
// ---------------------------------------------------------------------------

// Record is what the kernel hands over at a commit. It describes a possibility's
// WHOLE LIFE at once — which children were invoked, what was passed to each, what
// each returned, and what the commit composed from.
//
// One record per commit is the whole write protocol. State is written by the
// parent, at the parent's commit, and it is the only write in the system.
type Record struct {
	Span   ids.SpanID
	Engram *Engram // the possibility that closed. Already minted.
	Name   string  // how the PARENT reached it. Empty at a root.

	// Every child invoked over this possibility's life, in emitted order.
	// Attached to the parent iff the commit cited its return.
	Invocations []*Pending

	// The commit itself.
	Return *Observation
}

// Pending is one child invocation, carrying the child's engram so the graph can
// attach it if it was cited.
type Pending struct {
	Child *Engram
	Name  string
	Inv   *Invocation
}

// Apply is the only durable write. Every field it touches is appended, never
// modified.
//
// Four policies, all here rather than in the kernel:
//
//	FORMED ON CITATION   a child attaches iff the commit cited its return.
//	THE LOCAL VERDICT    Cited, on the invocation the parent owns.
//	HISTORY              one record per invocation. nothing collapses at write time.
//	DISPERSAL            every commit walks. the root is not special.
func (g *Graph) Apply(r *Record) {
	cited := map[ids.ObservationID]bool{}
	for _, c := range r.Return.Cites() {
		cited[c] = true
	}

	g.Store(r.Return)

	for _, p := range r.Invocations {
		// THE LOCAL VERDICT. Determined by what appears in the commit, so it
		// cannot be claimed — only carried forward.
		if p.Inv.Return != nil {
			p.Inv.Cited = cited[p.Inv.Return.ID]
		}

		g.mu.Lock()
		// History accumulates on the child engram the PARENT owns, whether or not
		// it was cited. A parent learns from its failures; the child never sees
		// them. That is refinement grows the asker, not the answerer, applied to
		// credit — and it is why all learning is topological.
		p.Child.History = append(p.Child.History, p.Inv)
		if p.Inv.Return != nil {
			p.Child.ByReturn[p.Inv.Return.ID] = p.Inv
		}

		// FORMED ON CITATION. The attachment IS the local positive — the child
		// exists in the parent's map, and nothing had to be written to say so.
		if p.Inv.Cited {
			r.Engram.Engrams[p.Name] = p.Child
		}
		g.mu.Unlock()
	}

	// Every commit disperses. Acceptance walks DOWN by hops from what was just
	// composed; errors stay local and do not walk at all.
	//
	// Descending a rejection would be one evidential origin becoming N records —
	// a shape the design refuses. And the asymmetry is not arbitrary: a false
	// positive is self-limiting, where a false negative is absorbing, because
	// nothing earns its way back if it is never tried.
	g.Disperse(r.Return, r.Span)
}

// Disperse is the credit walk. It is O(n · depth) BY DESIGN — an engram deep in a
// chain is visited once per ACCEPTING ANCESTOR, and that repetition is what
// populates the hop profile. It costs no model call.
//
// Propagation follows CONSTITUENTS, not durable structure. Observations are
// instances, so the walk lands only on invocations that actually produced this
// result; there is nothing unrelated to stray into. Span survives as a label for
// reading and grouping, and is no longer what keeps the walk correct.
func (g *Graph) Disperse(from *Observation, span ids.SpanID) {
	seen := map[ids.ObservationID]bool{}
	g.walk(from, span, 1, seen)
}

func (g *Graph) walk(o *Observation, span ids.SpanID, hop int, seen map[ids.ObservationID]bool) {
	// Cycle guard. Nothing in the design forbids an observation reachable twice
	// by different paths, and a diamond would otherwise double-count. Visiting
	// once per walk keeps one fact per accepting ancestor, which is what the hop
	// profile means.
	if seen[o.ID] {
		return
	}
	seen[o.ID] = true

	for _, cid := range o.Cites() {
		c, ok := g.Observation(cid)
		if !ok {
			continue
		}
		// Terminates with no special case at a crossing and at ingress: Producer
		// is empty, so there is no engram to descend into.
		if c.Producer == "" {
			continue
		}
		e, ok := g.Engram(c.Producer)
		if !ok {
			continue
		}
		g.mu.Lock()
		if inv, ok := e.ByReturn[c.ID]; ok {
			inv.Acceptances = append(inv.Acceptances, Acceptance{Span: span, Hop: hop})
		}
		g.mu.Unlock()
		g.walk(c, span, hop+1, seen)
	}
}

// Reached returns every engram the credit walk touches from an observation.
//
// MATERIALIZATION FOLLOWS CITATION: the staged effects that land are exactly the
// ones this reaches, for the same reason they earn credit. One traversal, two
// jobs — credit follows data and effects follow data stop being two claims.
func (g *Graph) Reached(from *Observation) map[ids.EngramID]bool {
	out := map[ids.EngramID]bool{}
	seen := map[ids.ObservationID]bool{}
	g.reach(from, out, seen)
	return out
}

func (g *Graph) reach(o *Observation, out map[ids.EngramID]bool, seen map[ids.ObservationID]bool) {
	if seen[o.ID] {
		return
	}
	seen[o.ID] = true
	if o.Producer != "" {
		out[o.Producer] = true
	}
	for _, cid := range o.Cites() {
		if c, ok := g.Observation(cid); ok {
			g.reach(c, out, seen)
		}
	}
}

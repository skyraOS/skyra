package kernel

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Possibility is ONE OPENING of an engram, manned by an entity. THE KERNEL HOLDS
// THIS, and it is entirely ephemeral.
//
//	THE ENGRAM         durable. what to do, and its accumulated history.
//	THE ENTITY         who interprets. stateless.
//	THE OBSERVATIONS   the parameters.
//	THE COMMIT         one observation. the completed work.
//
// Durable is what HAS happened; ephemeral is what MIGHT. Potential should not be
// the thing that persists, and a record should not be the thing you invoke.
//
// A possibility knows three things: the observations it was given, which engram
// it is opening, and the observations its children returned. Nothing else is in
// scope. It cannot perform work except by creating a possibility and handing it
// over — an incapacity, not a rule.
type Possibility struct {
	Engram *graph.Engram
	Entity ids.EntityID
	Name   string // how the parent reached it

	Observations []*graph.Observation
	Returned     []*graph.Observation

	// Has it forfeited origination. Set at the FIRST select, PERMANENTLY.
	Selected bool

	// PROVISIONAL (kernel.md). Notices that a referent behind a handle already in
	// frame has moved. Buffered to the same boundary as Monitor, so there is ONE
	// boundary rather than two.
	//
	// Dies with the kernel, and that is fine for a reason worth stating: an
	// impulse is a notice that a referent MOVED, and the referent is still moved.
	// The next assemble re-reads current world state. It is a re-read, not a lost
	// event.
	Impulses []*graph.Observation

	// Everything this possibility invoked over its whole life, in emitted order.
	// Written to the graph in one go at the commit — which is what makes eviction
	// safety fall out rather than being built.
	Pending []*graph.Pending

	// Names a child handed back this invocation. Delegation, not injection.
	Delegated map[string]*graph.Engram

	// Reach handed down. A parent gives a SUBSET, which is how a limb gets
	// narrower authority than what spawned it. Monotonic narrowing is the whole
	// rule: a select may drop, it can never mint.
	Reach []ids.EntityID

	// Every engram between this possibility and the root, inclusive.
	//
	// FOUND BY RUNNING IT, and it is the sharpest hole the build turned up.
	// Mint-on-closure puts every engram that ever closed into the index — roots
	// included. by-index searches that index by token overlap against the need.
	// A root's expectation IS the need, so it is the single best lexical match for
	// its own chain, and discovery hands the chain reach to its own ancestor.
	//
	// Observed directly: `by-index` matched "answer a question with separable
	// parts · You are the OUTERMOST possibility…" — the root — as the resolution
	// of the root's own need, and the chain recursed until the depth bound caught
	// it.
	//
	// The design has DOWN, NEVER UP for HISTORY: a possibility cannot see its
	// ancestors' state, which is what makes the citation signal ungameable. It has
	// NO EQUIVALENT BOUND FOR REACH. Nothing anywhere says a possibility may not be
	// handed a name that resolves to something it is already inside.
	//
	// The kernel is the right holder because it already holds ancestry — kernel.md
	// lists it among the four things it holds — and refusing a binding is
	// EXECUTION TOPOLOGY, not semantics, so this reads no content and makes no
	// judgement.
	Ancestry []ids.EngramID
}

// wouldRecur reports whether binding this engram would hand the possibility
// reach to something it is already inside.
func (p *Possibility) wouldRecur(id ids.EngramID) bool {
	for _, a := range p.Ancestry {
		if a == id {
			return true
		}
	}
	return false
}

// scopeOf collects every observation the possibility may name.
func (p *Possibility) scopeOf() []*graph.Observation {
	out := append([]*graph.Observation{}, p.Observations...)
	out = append(out, p.Returned...)
	out = append(out, p.Impulses...)
	return out
}

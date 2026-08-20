// Package graph is the durable layer. It holds only what HAPPENED.
//
// Nothing that is happening is representable here — no running flag, no in-flight
// possibility, no staged effect. That split is what makes three claims in the spec
// set true rather than aspirational (kernel.md): the kernel writes nothing,
// mint-on-closure means engrams record things that COMPLETED, and crash semantics
// collapse to nothing because the graph is exactly the set of completed commits.
package graph

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Expectation is what an engram is FOR. Its Expression is the form it takes.
//
// Immutable. Expectations never improve — they are authored once and re-run, and
// everything that accumulates accumulates AROUND them. All learning is
// topological: selection learns, synthesis does not.
//
// It carries NO LANGUAGE TAG. A tag would be a self-report the kernel would act on
// without verifying, and the verification that would matter — is this actually
// valid Python — can only be performed by Python. Handing a compiler natural
// language should fail; the entity reading it and failing IS the observation.
type Expectation struct {
	Expression string
}

// Engram is the durable trace: a fixed expectation with a history accumulating
// around it, and children of its own.
//
// It does NOT reconsolidate. Engram theory includes retrieval making a memory
// labile and rewritable; that part does not transfer. Append-only means a replay
// adds records and never modifies existing ones — only the aggregate drifts.
type Engram struct {
	// ID is what it IS; the name is how it is REACHED. An engram does not hold
	// its name — that is the parent's map key.
	ID ids.EngramID

	Expectation Expectation
	History     []*Invocation

	// Its own children, keyed by name. An engram lives inside exactly ONE
	// parent's map, so two parents structurally cannot share one.
	//
	// A never writes into B's memory. A writes into the child engram A owns.
	// Locality is structural rather than argued, and the no-transfer property
	// follows rather than being conceded: a bad expectation is independently
	// rediscovered as bad by every parent, because there is no "it," only copies.
	Engrams map[string]*Engram

	// The credit walk's entry point. Derived, rebuildable, deletable.
	ByReturn map[ids.ObservationID]*Invocation
}

func NewEngram(id ids.EngramID, x Expectation) *Engram {
	return &Engram{
		ID:          id,
		Expectation: x,
		Engrams:     map[string]*Engram{},
		ByReturn:    map[ids.ObservationID]*Invocation{},
	}
}

// Observation is one completed work.
type Observation struct {
	ID      ids.ObservationID
	Content content.Content

	// WHICH ENGRAM returned this — a specific child of a specific parent, which
	// is exactly the precision the walk needs.
	//
	// EMPTY means INGRESS: no invocation produced this, it entered the tree from
	// outside. Producer is what separates the three empty-Cites cases:
	//
	//	LEAF ORIGINATION   Producer = the leaf engram.       Cites empty.
	//	A CROSSING         Producer = the engram I invoked.  Cites empty.
	//	INGRESS            Producer EMPTY.                   Cites empty.
	//
	// Discriminated by absence rather than by a tag. An implementation reading
	// empty-Cites as ONE case will get it wrong.
	Producer ids.EngramID
}

// Cites is derived from Content, never declared. See content.Content.Cites.
func (o *Observation) Cites() []ids.ObservationID { return o.Content.Cites() }

// Shape is the read-time fold the history partitions by.
func (o *Observation) Shape() string { return o.Content.Shape() }

// IsIngress reports whether this observation entered the tree from outside.
//
// Ingress is not origination: asking whether an inbound observation is originated
// or composed is a category error, because no possibility produced it.
func (o *Observation) IsIngress() bool { return o.Producer == "" }

// Invocation is one asking, filled in three phases — which is why it is durable
// rather than transient.
type Invocation struct {
	// DETERMINED at invocation, WRITTEN at commit. Nothing is written during
	// exploration, or an evicted possibility leaves a trace.
	Span         ids.SpanID
	Observations []*Observation
	Entity       ids.EntityID

	// Where it actually landed. RECORDED, NOT KEYED.
	//
	// Same question to Claude and to a human is a DIFFERENT question. Same
	// question to Claude reached two ways is the SAME question — so the host does
	// not enter what makes two askings comparable. If one entity answers
	// differently on two hosts, that is evidence the entity is UNDER-SPECIFIED,
	// not a reason for a second key. Still written down for the case where the
	// host is the TARGET rather than the interpreter: turning on the wrong TV
	// matters.
	Host ids.HostID

	// What prompted reaching for this. A GROUPING KEY ONLY — it carries no credit
	// of its own. Typed through its producer it is (Q → P), which recurs and
	// carries statistics where a raw observation pair never could.
	Cite ids.ObservationID

	// WRITTEN at commit.
	Return *Observation
	Cited  bool // did the parent's commit cite it. THE LOCAL VERDICT.

	// APPENDED LATER by the credit walk. The pending part.
	//
	// Records are events, not counters. Append-only without qualification — a
	// counter is a mutation, an appended event is not. Trajectory comes free
	// because the log is ordered, and set membership survives, which keeps
	// sibling blindness approachable rather than permanently foreclosed.
	Acceptances []Acceptance
}

// Acceptance is one accepting ancestor's agreement, carrying how far away it was.
type Acceptance struct {
	// Which execution accepted. A LABEL — the walk is bounded by constituents
	// being instances, not by this.
	Span ids.SpanID

	// 1 = the parent's parent accepted. N = accepted N levels up.
	Hop int
}

// HopProfile is the detector. Not one signal weakening with distance — signals at
// every scale.
//
//	{hop1, hop2, hop3}   validated locally AND at range.
//	{hop1, hop2, — }     good nearby, never confirmed further out.
//	{hop1, — , — }       cited by its parent, and nothing else has ever agreed.
//
// The third profile is the self-sustaining subtree, and it reads off directly.
func (i *Invocation) HopProfile() map[int]int {
	out := map[int]int{}
	for _, a := range i.Acceptances {
		out[a.Hop]++
	}
	return out
}

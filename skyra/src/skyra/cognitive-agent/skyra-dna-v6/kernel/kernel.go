// Package kernel holds the protocol. The kernel HOLDS everything happening; the graph
// holds only what happened.
//
// It has no policy, reads no content, and performs no durable write except the one
// projection at the outer boundary.
//
//	KERNEL OWNS      scope · live topology · causal order · routing · closure ·
//	                 the transition from ephemeral to durable · the protocol itself
//	ENGRAMS AND      interpret needs · rank potential · translate observations ·
//	ENTITIES OWN     recognize meaning · produce results
//
// PROTOCOL OWNERSHIP DOES NOT MAKE THE KERNEL A SEMANTIC INTERPRETER.
package kernel

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/engram"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/observation"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/possibility"
)

// Query is what a possibility supplies to Think. Three dimensions, and cardinality
// carries meaning: EMPTY is unresolved, a SET is bounded, a SINGLE VALUE is fixed.
//
// NO ADDRESS OPERAND. Address resolution is volatile mechanism behind the boundary.
type Query struct {
	Entities     []ids.EntityID
	Expectations []engram.Expression
	Observations []ids.ObservationID
}

// Candidate is an OBSERVATION carrying potential behaviour.
//
// Not an executable handle: it holds no address, no transport, no live relationship,
// and no claim that contact has been established. It has no methods, so there is no
// way to run one except by handing it to Decide.
type Candidate struct {
	Observation *observation.Observation
	Rank        int
}

// Diagnosis is what Think returns when nothing fits. A CONSTRUCTED CONCLUSION, never
// an empty slice — absence cannot distinguish no match, no authority, and failure to
// think.
type Diagnosis uint8

const (
	NoKnownPotential Diagnosis = iota
	BeyondYourHorizon
)

// Decision is what a possibility hands to Decide.
//
// NO ENTITY OPERAND: the candidate observation already identifies it, and a second
// field would permit contradiction between them. NO POSITION OPERAND beyond the slot —
// position is causal context, not a semantic operand the caller must declare.
type Decision struct {
	Slot    string
	On      ids.ObservationID // THE CANDIDATE OBSERVATION, preserved as a constituent
	Inputs  []ids.ObservationID
	Horizon possibility.Horizon // must narrow from the caller's
}

// Head is live topology. TOPOLOGY ONLY — the kernel may report where things are, never
// what they mean.
//
// head == no live children AND no return under the current draft. DERIVED BY WALKING,
// never counted: a counter that ticks on entry reports every waiting ancestor as a
// head, which makes a stalled head indistinguishable from a parent legitimately
// waiting.
type Head struct {
	Possibility ids.PossibilityID
	Engram      ids.EngramID // AN ID. never the expectation. the caller dereferences it.
	Entity      ids.EntityID
	Host        ids.HostID
	Depth       int
}

// Kernel is the possibility-facing surface.
//
// FIVE METHODS FOR SEVEN MEANINGS. Reach and Recognize are protocols with public
// result types and no caller-facing method, because nothing outside the kernel ever
// invokes them:
//
//	REACH       the hidden realization phase INSIDE Decide.
//	RECOGNIZE   the return boundary INSIDE Remember. A comparison of two substrate-held
//	            strings — a VERB the kernel performs, not a value anyone stores.
type Kernel interface {
	// Approach is the boundary. The only way anything enters, which puts the trust
	// boundary and the gate on the same object.
	//
	// HOLE: nothing authenticates. The ceiling bounds what an admitted session may DO;
	// nothing bounds WHO MAY OPEN ONE.
	Approach(Admission) (*Session, error)

	// Think creates bounded potential and performs none of it. Contacts nothing,
	// resolves no address, opens nothing.
	Think(p ids.PossibilityID, q Query) ([]Candidate, error)

	// Decide chooses one complete composition and binds it. THE CHOICE IS THE BINDING,
	// and Reach happens inside it.
	//
	// FORK, UNRESOLVED — WHEN DOES THIS RETURN?
	//	AS SPECIFIED   when the child is LIVE. Realization says which of three
	//	               conclusions Reach reached. Fast; does not wait for the work.
	//	AS A CALLER-   when the CHILD RETURNS, which is what a method call is and what
	//	DRIVEN WALKER  a blocking compiler needs.
	//	WANTS
	// The second collapses two events into one signature and makes concurrency
	// inexpressible without a group form. The first leaves the caller nothing to
	// block on. NOT DECIDED.
	Decide(p ids.PossibilityID, d Decision) (Realization, error)

	// Remember receives a return into the parent's live causal history, running
	// Recognize as its first act. At the outer boundary it projects and closes.
	//
	// ONE OPERATION AT EVERY LEVEL. Which boundary it is at is a fact about the
	// position in the chain, not a parameter — durability is not the caller's to ask
	// for.
	Remember(p ids.PossibilityID, o ids.ObservationID) error

	// Orient reports live topology, never meaning. Bounded by the subtree, which is a
	// fact about which way pointers point rather than a rule anyone checks.
	Orient(p ids.PossibilityID) ([]Head, error)
}

// Realization is what Reach concluded, surfaced on Decide's result.
//
// The version a possibility acts through arrives here and nowhere else: YOU HAVE TO
// REACH SOMETHING IN ORDER TO AGE IT, and a version returned by Decide would be a
// version for something that may never exist.
type Realization struct {
	Kind   possibility.RealizationKind
	Opened ids.PossibilityID // Kind == Live only

	// Why is BOUNDED IMPOSSIBILITY EVIDENCE. Never a raw endpoint error — addresses
	// and transport failures do not escape upward as application vocabulary.
	Why *observation.Observation
}

// Admission is what an outside party presents at Approach.
type Admission struct {
	Namespace ids.EngramID // a namespace IS an engram
	Entity    ids.EntityID
	Requested possibility.Horizon

	// Credential is A NAMED HOLE, not a design. It is here so every call site has to
	// pass something and the absence is visible at the boundary rather than
	// discovered later.
	Credential any
}

// Session is ALL LIVE WORK under one namespace. One per sign-in.
//
// It holds what admission provisions ONCE, not per ask. A BRANCH IS A POSITION, NOT A
// TYPE — a possibility with no parent.
type Session struct {
	Namespace ids.EngramID
	Entity    ids.EntityID
	Runtime   ids.RuntimeID // "here". one per device, so it is held once, not per opening.
	Horizon   possibility.Horizon
}

// Package engram holds the durable layer.
//
//	ENGRAM        DURABLE.     the behaviour. what accumulates.
//	POSSIBILITY   EPHEMERAL.   one invocation of an engram.
//	BRANCH        EPHEMERAL.   one change proposed to an engram, and what came of it.
//
// A POSSIBILITY CANNOT BE OPENED WITHOUT REPRESENTING AN ENGRAM.
// AN ENGRAM ACQUIRES BRANCHES ONLY THROUGH POSSIBILITIES.
package engram

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/observation"
)

// Expression states what a position is FOR. NO LANGUAGE TAG — a tag would be a
// self-report. Nothing checks it.
type Expression struct {
	Expression string
}

// Signature IS the contract. Not a thing derived from it, not a thing checked against
// it. The parameter list, in order, with its types, is the whole of what may be bound
// at a position.
type Signature struct {
	Expectation Expression

	// Params are TYPE NAMES, ORDERED, AND THE ORDER IS SEMANTIC.
	//
	// A compiler may be an LLM, and an LLM weights earlier context more heavily — so
	// order is a SALIENCE DECLARATION by the parent, not merely a binding mechanism.
	// A reorder is therefore a real version change, and marshalling may not reorder
	// params for convenience.
	//
	// NO PARAMETER NAMES. A local name is the callee's, and the continuation is
	// interior: int f(int x) and int f(int y) are the same signature.
	Params []string
}

// Child is everything a parent owns for one child position.
//
// TWO CONTRACTS, NOT THREE. An expectation contract was tested at composition time,
// which was failure kind 1; that check is removed, so the contract had no evaluation
// point left and nothing ever called it.
//
// There is no Child field on a Child. A contract describes a POSITION, not an edge to
// a particular engram — a stored type edge would go stale and would read as evidence
// that composition succeeded.
type Child struct {
	Signature Signature

	// Return is THE CHILD TYPE this parent expects back. A STRING.
	//
	// Recognize compares this against the child engram's Returns. Both live on
	// ENGRAMS — durable, substrate-held — so neither is forgeable by whatever wrote
	// the observation, which is what the no-self-certification rule is about.
	Return string
}

// Invocation is one occurrence of one child position, durable after projection.
type Invocation struct {
	Possibility ids.PossibilityID
	Position    string

	Inputs []*observation.Observation
	Return *observation.Observation

	// Remembered is THE LOCAL VERDICT — whether the parent RETAINED it.
	//
	// NOT the type check. Matching is a comparison the kernel re-performs from two
	// durable strings; retention is a CHOICE nothing can reconstruct. Those were being
	// blurred under one name.
	Remembered bool

	Acceptances []Acceptance
}

// Acceptance carries how far a result travelled. Hop 1 = the parent's parent accepted.
type Acceptance struct {
	Possibility ids.PossibilityID // a LABEL, not a bound
	Hop         int
}

// Engram is child contracts plus compiled continuation, and nothing else is part of
// what an engram IS.
type Engram struct {
	ID ids.EngramID

	// Contracts is what may occupy each position beneath this behaviour, keyed by
	// position name. Held apart from what actually filled them, so a contract may
	// exist for a position nothing has occupied.
	Contracts map[string]*Child

	// Continuation is the compiled body. EMPTY until solidification.
	//
	// The durable form of what a live Branch holds as Syntax. Empty is not an error
	// and not a leaf: it is an engram that has only ever run in primitive mode.
	//
	// OPEN: whether this stays source text or becomes a compiled artifact, and what
	// Compile's input and output actually are.
	Continuation string

	// Compiler is WHERE the interpreter of that body lives.
	//
	// An ADDRESS, not an interface. It holds no protocol methods: Decide, Recognize
	// and Remember are kernel methods, and a compiler holding one would be extending
	// causal structure from outside the kernel.
	//
	// DETERMINED BY THE BODY'S SYNTAX, never chosen by a parent.
	Compiler ids.Address

	// Creator is WHERE the thing that authored that body lives. EMPTY FOR A LEAF —
	// there is no interior body to author.
	//
	// Absent at runtime. Reached only when compilation fails, because the child owns
	// its body and therefore owns failing to compile one.
	//
	// OPEN: an Address gives the failure path somewhere to pop to and leaves
	// authorship UNCITABLE. Whether this should be an EngramID instead is unresolved.
	Creator ids.Address

	// Returns is THE TYPE THIS ENGRAM PRODUCES, exposed to the kernel as a string.
	// What it EXPECTS BACK per position lives in Contracts[pos].Return — same string,
	// other side.
	Returns string

	// History is what accumulated AROUND this behaviour. Never inside it.
	History []*Invocation
}

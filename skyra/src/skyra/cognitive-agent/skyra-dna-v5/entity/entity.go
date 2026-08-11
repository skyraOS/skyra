// Package entity holds WHO interprets, and the registry that reaches them.
//
//	ENTITY   WHO.    interprets. authors behaviour.
//	HOST     WHERE.  the device. realizes execution.
//
// An entity is an IDENTITY and nothing else. Every declaration dissolved — Reads,
// Capacity, modalities, Hosts — for one reason: a declaration is made by whoever
// REGISTERED the entity, not by the entity. A third-party assertion about a
// capability nobody verified, which is worse than the expectation-side language
// tag, because at least an author knows what they wrote.
//
// History replaces all of it. "Python to this entity worked" is recorded per
// invocation and grounded; "this entity reads Python" is a guess made at
// registration time.
package entity

import (
	"fmt"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Interpreter is the execution boundary. Step is the only method that PRODUCES
// anything; everything else is data crossing it.
//
// Nothing executes without asking: a possibility cannot invoke, it can only emit
// a request and be invoked in turn.
//
// DIVERGENCE FROM interfaces.md, recorded rather than hidden. The spec writes
//
//	Step(x Expectation, f Frame) Emission
//
// with no error. But the grammar's incapacity is a PARSING property, so a
// response that parses as neither form has to be representable — and the spec
// itself says such a thing becomes "an ordinary failed invocation, observed and
// recorded on the invoker's engram." Something must carry it across this
// boundary. Returning it keeps the failure OUTSIDE the two forms rather than
// inventing a third one to hold it.
type Interpreter interface {
	Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error)

	// Monitor is the delta form of assemble and returns NOTHING.
	//
	// A return value would be a third emission channel outside SELECT / COMMIT,
	// and the grammar's whole claim is that only two forms parse. One method with
	// a return type ends it.
	//
	// The kernel calls it at GROUP CLOSURE, in EMITTED order, never as members
	// land — arrival order is a scheduling accident, and pushing on arrival makes
	// the entity's context depend on it, breaks replay, and is partial completion
	// arriving as a side effect rather than as a decision.
	//
	// A LEAF NEVER RECEIVES IT: a leaf never selects, so there is nothing to
	// subscribe to. Monitor is strictly an orchestration channel.
	Monitor(o *graph.Observation)
}

// Registry is how each entity is REACHED. Runtime configuration — mutable,
// rebuildable, and never durable.
//
//	GRAPH     entities and hosts as IDENTITIES.
//	RUNTIME   how to reach each one. NEVER durable.
type Registry struct {
	interpreters map[ids.EntityID]func() Interpreter
}

func NewRegistry() *Registry {
	return &Registry{interpreters: map[ids.EntityID]func() Interpreter{}}
}

// Register binds an entity ID to a way of reaching it.
//
// A CONSTRUCTOR rather than an instance, because ENTITIES ARE STATELESS: anything
// accumulating between invocations is world state and must be carried explicitly.
// If state hid inside an interpreter, two invocations that look identical would
// give different answers and nothing could see why — "same question" becomes a
// lie, and every fold downstream inherits the falsehood.
//
// Handing out a fresh interpreter per possibility makes that structural instead
// of a promise. Within one possibility it obviously holds a growing context; that
// was already true the moment the frame shipped on every step.
func (r *Registry) Register(id ids.EntityID, mk func() Interpreter) {
	r.interpreters[id] = mk
}

// Open reaches an entity for one possibility.
func (r *Registry) Open(id ids.EntityID) (Interpreter, error) {
	mk, ok := r.interpreters[id]
	if !ok {
		// Fails LOUDLY. An unreachable entity is not a silent no-op — the whole
		// posture of the design is that a wrong answer nobody noticed is worse
		// than a crash.
		return nil, fmt.Errorf("no way to reach entity %q", id)
	}
	return mk(), nil
}

// Known lists registered entities. Used to seed a root's reach.
func (r *Registry) Known() []ids.EntityID {
	out := make([]ids.EntityID, 0, len(r.interpreters))
	for id := range r.interpreters {
		out = append(out, id)
	}
	return out
}

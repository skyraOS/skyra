package machine

import (
	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Typed is a return's type as the parent's routes know it (next/30): the
// binding it came back through, its type hash, and its presentation. It is
// derived on demand from what a round already holds, never stored in a record.
type Typed struct {
	Binding      string // the name the parent invoked: a name in reach, or the actor a reference stands for
	Type         string // contract.Type(actor identity, contract hash, shape): never shown to a model
	Presentation string // what a model is shown: binding → shape
}

// bound resolves what an invoke addressed under the parent's Horizon: the
// binding its routes are keyed on, and the surface that ran.
func bound(parent horizon.Horizon, to string) (string, horizon.Surface, bool) {
	if s, ok := parent.Reach[to]; ok {
		return to, s, true
	}
	if g, ok := parent.Refs[to]; ok {
		s, ok := parent.Reach[g.Actor]
		return g.Actor, s, ok
	}
	return "", horizon.Surface{}, false
}

// TypeOf derives the type of a child's discharge, if it returned. The actor is
// the identity of the surface that ran; the contract is the one it was
// invoked through.
func TypeOf(parent horizon.Horizon, c Child) (Typed, bool) {
	if c.Discharge == nil || c.Discharge.State != record.Return {
		return Typed{}, false
	}
	name, s, ok := bound(parent, c.Invoke.To)
	if !ok {
		return Typed{}, false
	}
	v, err := contract.Decode(c.Discharge.Payload)
	if err != nil {
		return Typed{}, false
	}
	return Typed{
		Binding:      name,
		Type:         contract.Type(horizon.Identity(s), horizon.ContractHash(s.Input), contract.Shape(v)),
		Presentation: name + " → " + contract.Present(v),
	}, true
}

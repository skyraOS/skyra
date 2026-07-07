package main

import (
	"fmt"
	"sync"

	"skyra-v1/src/skyra"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	"skyra-v1/src/skyra/world"
)

// dnav2Being is the world-facing Space for the slice-based-dispatch being. The being's
// cognition lives on the DNA nucleus (entity/grouping indexes) and runs through the
// orchestrator's OWN kernel — not world traversal — so this adapter's graph is inert. It
// exists only to (a) be placed in the world so michael can engage it and the map can locate
// it, and (b) turn an inbound engage into an orch.Fire. Motor output flows back out through
// DNA.SetEmit → world (wired in main), independent of this type. This is the whole bridge:
// world delivery → pipeline in; DNA.Emit → world out.
type dnav2Being struct {
	id    string
	slice skyra.Slice
	orch  *dnav2.Orchestrator

	mu     sync.Mutex
	seq    int
	prevID string // threads turns: each engage's id points at the previous
}

func newDnav2Being(id string, slice skyra.Slice, orch *dnav2.Orchestrator) *dnav2Being {
	return &dnav2Being{id: id, slice: slice, orch: orch}
}

func (b *dnav2Being) ID() string                                 { return b.id }
func (b *dnav2Being) Slice() skyra.Slice                         { return b.slice }
func (b *dnav2Being) NewActivation() skyra.ActivationRecord      { return beingActivation{} }
func (b *dnav2Being) Content() any                               { return nil }
func (b *dnav2Being) Agents() map[string]skyra.Space             { return nil }
func (b *dnav2Being) Children() map[string]skyra.Space           { return nil }
func (b *dnav2Being) Express(origin skyra.Space, impulse any) any { return nil }

// Observe is the inbound bridge: the world delivered an impulse to the being. An external
// engage starts a pipeline turn (threaded by prev-id). Return waves and non-engage primitives
// are ignored — the pipeline's own reafference is not driven from here.
func (b *dnav2Being) Observe(origin skyra.Space, impulse any) {
	imp, ok := impulse.(skyra.Impulse)
	if !ok || imp.Return || imp.Primitive != world.PrimitiveEngage {
		return
	}
	b.mu.Lock()
	b.seq++
	id, prev := fmt.Sprintf("in:%d", b.seq), b.prevID
	b.prevID = id
	b.mu.Unlock()
	b.orch.Fire(id, prev, imp)
}

// beingActivation is a constant activation record — the being is a delivery entry node, not
// part of a scored traversal (mirrors dnav2's rootActivation).
type beingActivation struct{}

func (beingActivation) Activate(visited bool, traversalCount int) float64 { return 1.0 }

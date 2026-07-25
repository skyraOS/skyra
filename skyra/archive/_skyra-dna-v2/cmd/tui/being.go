package main

import (
	"fmt"
	"sync"

	"skyra-v1/src/skyra"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	"skyra-v1/src/skyra/world"
)

// dnav2Being is the world-facing Space for the slice-based-dispatch being. The being's
// cognition lives on the DNA nucleus (entity/episode indexes) and runs through the
// orchestrator's OWN kernel — not world traversal — so this adapter's graph is inert. It
// exists only to (a) be placed in the world so michael can engage it and the map can locate
// it, and (b) turn an inbound engage into an orch.Fire. Motor output flows back out through
// DNA.SetEmit → world (wired in main), independent of this type. This is the whole bridge:
// world delivery → pipeline in; DNA.Emit → world out.
type dnav2Being struct {
	id    string
	slice skyra.Slice
	orch  *dnav2.Orchestrator
	dna   *dnav2.DNA

	mu     sync.Mutex
	seq    int
	prevID string // threads turns: each engage's id points at the previous
}

func newDnav2Being(id string, slice skyra.Slice, orch *dnav2.Orchestrator, genome *dnav2.DNA) *dnav2Being {
	return &dnav2Being{id: id, slice: slice, orch: orch, dna: genome}
}

func (b *dnav2Being) ID() string                                  { return b.id }
func (b *dnav2Being) Slice() skyra.Slice                          { return b.slice }
func (b *dnav2Being) NewActivation() skyra.ActivationRecord       { return beingActivation{} }
func (b *dnav2Being) Content() any                                { return nil }
func (b *dnav2Being) Agents() map[string]skyra.Space              { return nil }
func (b *dnav2Being) Children() map[string]skyra.Space            { return nil }
func (b *dnav2Being) Express(origin skyra.Space, impulse any) any { return nil }

// Observe is the inbound bridge: the world delivered an impulse to the being. An
// external engage starts a pipeline turn; a PERCEIVE RETURN is translated
// (world.PerceptionResult → dnav2.Surroundings — the sensory transduction, so the
// DNA stays world-agnostic) and enters the pipeline as a reafference. Both are
// threaded by prev-id. Other return waves (move/engage reafference) are still
// dropped here — that lane is out of scope for the perceive build.
func (b *dnav2Being) Observe(origin skyra.Space, impulse any) {
	imp, ok := impulse.(skyra.Impulse)
	if !ok {
		return
	}
	switch {
	case !imp.Return && imp.Primitive == world.PrimitiveEngage:
		if isOwnEcho(imp, b.id) {
			return
		}
		b.fire(imp)
	case imp.Return && imp.Primitive == world.PrimitivePerceive:
		res, ok := imp.Payload.(world.PerceptionResult)
		if !ok {
			return
		}
		b.fire(skyra.Impulse{
			Primitive: world.PrimitivePerceive,
			Return:    true,
			Payload:   translateSurroundings(res, b.id),
			Source:    imp.Source,
		})
	}
}

// isOwnEcho reports whether an engage is the world's echo of the being's OWN
// outward engage (the delivery set includes the sender — same filter michael's
// OnInbound applies). Without it the echo re-enters as an "internal" turn and
// self-stimulates an unbounded engage loop (live run 2026-07-08). A deliberate
// self-engage interrupt (motor <engage><skyra> → TargetID == the being) is NOT
// an echo and passes through.
func isOwnEcho(imp skyra.Impulse, beingID string) bool {
	return imp.Source != nil && imp.Source.SpaceID == beingID && imp.TargetID != beingID
}

// fire threads a pipeline turn: each impulse's id points at the previous (the
// causal prev chain — the heuristic linkage, perceive-build-uncertainties.md §5).
func (b *dnav2Being) fire(imp skyra.Impulse) {
	b.mu.Lock()
	b.seq++
	id, prev := fmt.Sprintf("in:%d", b.seq), b.prevID
	b.prevID = id
	b.mu.Unlock()
	b.orch.Fire(id, prev, imp)
}

// translateSurroundings unwraps the world's perception result into the DNA-native
// payload: self = the origin's coord (given, not swept); each swept agent's coord
// is the coord of the point it was found standing on. The perceiving being itself
// is excluded from the swept set (self arrives via Self, the wave's origin).
func translateSurroundings(res world.PerceptionResult, selfID string) dnav2.Surroundings {
	s := dnav2.Surroundings{Self: res.SelfCoord, Agents: map[string]skyra.Coord{}}
	for _, p := range res.Points {
		if p.Source == nil {
			continue
		}
		for id := range p.Agents {
			if id == selfID {
				continue
			}
			s.Agents[id] = p.Source.Coord
		}
	}
	return s
}

// beingActivation is a constant activation record — the being is a delivery entry node, not
// part of a scored traversal (mirrors dnav2's rootActivation).
type beingActivation struct{}

func (beingActivation) Activate(visited bool, traversalCount int) float64 { return 1.0 }

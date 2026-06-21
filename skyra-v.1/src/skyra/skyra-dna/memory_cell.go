package dna

import "skyra-v1/src/skyra"

// MemoryCell — the one passive node of the memory graph.
//
// Setpoint, entity, understanding, stress, response, and action are ROLES of
// this single type, differentiated by Kind, Content, and topological position —
// "one type, topology differentiates" (dna.md). They are not separate structs:
// the wave traverses them uniformly, and the alternating-layer recall
// (understanding ↔ entity) works *because* both layers are the same type.
//
//   - setpoint:      Content = understanding (gist); Children = closure snapshot
//     (entities + responses + actions + stresses)
//   - entity:        Content = descriptor (person/place/thing); Children = the
//     setpoints stamped on it at completion (the reverse-index hub)
//   - understanding/stress/response/action: graph material under the above
//
// See specs/2026-06-16/skyra-dna/setpoint.md + entity.md and
// design-docs/cognition/{entities-and-surroundings,groupings-and-understandings}.md.
type MemoryCell struct {
	id       string
	kind     Kind
	content  any                    // understanding (gist) | entity descriptor | ...
	children map[string]skyra.Space // edge-keyed; meaning depends on Kind
	slice    skyra.Slice

	// intrinsic — cell-intrinsic excitability, stamped ONCE at crystallization.
	// The node's born confidence/precision, distinct from any edge (synaptic)
	// weight — the memory-allocation / engram-recruitment substrate. Feeds the
	// node's baseline activation (cold-start confidence) and solid-ground for the
	// cascade. See design-docs/kernel/activation-and-traversal.md ("Birth alignment").
	intrinsic Intrinsic
}

// Intrinsic — the constituents of a node's birth excitability, kept separable
// (not a single float) so they can be reweighted or extended (valence/emotion-
// at-birth, novelty) without losing the raw inputs. Composed into the bump via
// Value(). Same philosophy as keeping IndexScore/Alignment distinct.
type Intrinsic struct {
	ClosureGrade  float64 // precision — how strongly it closed at birth
	StressAtBirth float64 // global stress (arousal) at the moment of crystallization
	// future: ValenceAtBirth (the emotion fingerprint), Novelty, ...
}

// Value is the composed intrinsic excitability (the "birth alignment") used in
// the bump. The composition lives here so it can evolve.
func (i Intrinsic) Value() float64 { return i.ClosureGrade * i.StressAtBirth }

// Kind is a bookkeeping/role tag (registry, debugging), NOT a behavior switch.
// Behavior is the same across kinds — the wave doesn't read it.
type Kind string

const (
	KindSetpoint      Kind = "setpoint"      // content = understanding; children = closure snapshot
	KindEntity        Kind = "entity"        // content = descriptor; children = stamped setpoints
	KindUnderstanding Kind = "understanding" // a setpoint's crystallized gist
	KindStress        Kind = "stress"        // failure-understanding (cascade closure)
	KindResponse      Kind = "response"
	KindAction        Kind = "action"
)

func NewMemoryCell(id string, kind Kind, content any, slice skyra.Slice) *MemoryCell {
	return &MemoryCell{
		id:       id,
		kind:     kind,
		content:  content,
		children: map[string]skyra.Space{},
		slice:    slice,
	}
}

func (m *MemoryCell) ID() string                            { return m.id }
func (m *MemoryCell) Kind() Kind                            { return m.kind }
func (m *MemoryCell) Slice() skyra.Slice                    { return m.slice }
func (m *MemoryCell) Content() any                          { return m.content }
func (m *MemoryCell) Children() map[string]skyra.Space      { return m.children }
func (m *MemoryCell) NewActivation() skyra.ActivationRecord { return &memoryActivation{} }

// Intrinsic returns the node's birth excitability constituents.
func (m *MemoryCell) Intrinsic() Intrinsic { return m.intrinsic }

// BirthAlignment is the composed intrinsic value used in the bump.
func (m *MemoryCell) BirthAlignment() float64 { return m.intrinsic.Value() }

// StampBirth records the intrinsic excitability at crystallization. Called once,
// at completion, with the closure grade and the global stress at that instant.
func (m *MemoryCell) StampBirth(closureGrade, stressAtBirth float64) {
	m.intrinsic = Intrinsic{ClosureGrade: closureGrade, StressAtBirth: stressAtBirth}
}

// Observe — traversal intake (passive). origin is the wave owner.
func (m *MemoryCell) Observe(origin skyra.Space, impulse any) {}

// Express — traversal output (passive): contributes its content to the wave.
// origin is the wave owner.
func (m *MemoryCell) Express(origin skyra.Space, impulse any) any { return m.content }

// AddChild adds an edge to a neighbor, keyed by edge identity (the map key) so
// activation is edge-owned. This is how a setpoint stamps its closure snapshot,
// and how an entity is stamped with a setpoint at completion
// (entity.children += setpoint — the reverse-index write-time invariant).
func (m *MemoryCell) AddChild(edgeKey string, cell skyra.Space) {
	m.children[edgeKey] = cell
}

// memoryActivation — edge-owned activation. The kernel keys one per EDGE (the
// Children map key), so A→B and C→B score independently.
//
// Target equation (design-docs/kernel/activation-and-traversal.md) — three loci:
//
//	bump = ( IndexScore × Alignment ) × stress_modulator × intrinsic
//	       └── per-edge (here) ─────┘   └ global (slice) ┘  └ node ┘
//
// fed into two EMAs — Identity (α=0.02, the groove) and Familiarity (α=0.1,
// recent). Only the PER-EDGE inputs live here (IndexScore = surroundings
// overlap, one-shot per cycle; Alignment = lagged from Frame 2). Stress is
// GLOBAL — read from the slice (a neuromodulator, not per-edge). intrinsic is
// PER-NODE — the MemoryCell's BirthAlignment. Exact tuning (stress_modulator,
// threshold family) is TBD by experiment.
type memoryActivation struct {
	Identity    float64 // EMA α=0.02
	Familiarity float64 // EMA α=0.1
	IndexScore  float64 // surroundings overlap, written by the cell (per-edge)
	Alignment   float64 // lagged from the previous closure scoring (per-edge)
}

// Activate — minimal for now: keeps traversed cells alive so waves flow. The
// real signed bump/decay into the EMAs is the next step.
func (a *memoryActivation) Activate(visited bool, _ int) float64 {
	if visited {
		return 1.0
	}
	return 0
}

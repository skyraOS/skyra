package dna

import (
	"sync"

	"skyra-v1/src/skyra"
)

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
	id string
	kind Kind

	// mu guards content + children. The master graph is read and written by
	// CONCURRENT waves (the fan-out: other:predict appends the distribution while
	// self:predict's frame traversal iterates the same grouping) — the "RWMutex on
	// the master graph" data-race safety the dispatch spec calls for. Children()
	// returns a copy under the read lock; AddChild/SetContent take the write lock.
	mu       sync.RWMutex
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
	KindPresence      Kind = "presence"      // an entity's presence/location at a time (EntityState)
	KindResponse      Kind = "response"      // an incoming move — the other's utterance/action (InteractionRecord)
	KindAction        Kind = "action"        // an outgoing move — the being's own action (motor)
	KindGrouping   Kind = "grouping"   // a higher-order setpoint over the active field; content = Grouping (gists are CHILDREN)
	KindGist       Kind = "gist"       // one framing of a grouping — a SEQUENCE (edit appends, priors decay); content = Gist
	KindPrediction Kind = "prediction" // an open bet — a predicted next move (self/other); content = Prediction
)

// Grouping is a grouping node's content — the episode shell. The framing lives in
// its KindGist CHILDREN (a sequence: each edit mints a new gist, priors retained
// and decaying), and the gist owns the records/predictions made under it. See
// design-docs/cognition/perceive-internalization.md §3.
type Grouping struct {
	Created int // tick the episode opened
}

func NewMemoryCell(id string, kind Kind, content any, slice skyra.Slice) *MemoryCell {
	return &MemoryCell{
		id:       id,
		kind:     kind,
		content:  content,
		children: map[string]skyra.Space{},
		slice:    slice,
	}
}

func (m *MemoryCell) ID() string         { return m.id }
func (m *MemoryCell) Kind() Kind         { return m.kind }
func (m *MemoryCell) Slice() skyra.Slice { return m.slice }

func (m *MemoryCell) Content() any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.content
}

// Children returns a COPY of the edge map under the read lock, so callers can
// iterate it while concurrent waves append (no torn map iteration).
func (m *MemoryCell) Children() map[string]skyra.Space {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]skyra.Space, len(m.children))
	for k, c := range m.children {
		out[k] = c
	}
	return out
}

func (m *MemoryCell) Agents() map[string]skyra.Space        { return nil }
func (m *MemoryCell) NewActivation() skyra.ActivationRecord { return newMemoryActivation() }

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
func (m *MemoryCell) Express(origin skyra.Space, impulse any) any { return m.Content() }

// AddChild adds an edge to a neighbor, keyed by edge identity (the map key) so
// activation is edge-owned. This is how a setpoint stamps its closure snapshot,
// and how an entity is stamped with a setpoint at completion
// (entity.children += setpoint — the reverse-index write-time invariant).
func (m *MemoryCell) AddChild(edgeKey string, cell skyra.Space) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.children[edgeKey] = cell
}

// ParsedNode is a node's self-description for frame assembly: its content, when
// it happened (for chronological ordering), and the grouping it belongs to
// (empty until grouped). Structured, NOT formatted text — the assembler (Global)
// decides the frame's final shape. See
// specs/2026-06-16/skyra-dna/midlevel-observe-gather.md.
type ParsedNode struct {
	ID        string
	Kind      Kind
	Content   any
	Timestamp int
	GroupID   string
}

// Parse returns the node's self-description for frame assembly. Generic: the
// content is already typed so it carries itself; only the occurrence time is
// lifted out (for chronological ordering). GroupID is empty until the node is
// grouped.
func (m *MemoryCell) Parse() ParsedNode {
	content := m.Content()
	return ParsedNode{
		ID:        m.id,
		Kind:      m.kind,
		Content:   content,
		Timestamp: contentTimestamp(content),
		GroupID:   "",
	}
}

// contentTimestamp lifts the occurrence time out of a node's typed content for
// chronological frame ordering. (A node-level timestamp field is the cleaner
// long-term home; this localizes the knowledge to the parse readout for now.)
func contentTimestamp(content any) int {
	switch c := content.(type) {
	case EntityState:
		return c.Timestamp
	case InteractionRecord:
		return c.Timestamp
	case PerceiveRecord:
		return c.Timestamp
	default:
		return 0
	}
}

const (
	alphaIdentity    = 0.02 // the groove (long tail)
	alphaFamiliarity = 0.1  // recent relationship
)

// memoryActivation — edge-owned activation. The kernel keys one per EDGE (the
// Children map key), so A→B and C→B score independently.
//
// The bump is the PRODUCT of the activation equation, fed as the SAMPLE into two
// EMAs (no fixed +1 constant):
//
//	bump = ( IndexScore × Alignment ) × stress × intrinsic
//	       └── per-edge (here) ─────┘   └ global ┘  └ node ┘
//	Identity    += αI · (bump − Identity)      // αI = 0.02, the groove
//	Familiarity += αF · (bump − Familiarity)   // αF = 0.1,  recent
//	activation   = Identity + Familiarity      // groove + recency (tunable blend)
//
// Absent (visited=false) → sample 0 → both EMAs decay toward 0. Signed: a
// negative Alignment yields a negative bump → the EMAs and the returned
// activation go negative → the kernel's threshold drops it (the below-0 WM
// eviction).
//
// Inputs default neutral (1.0) so a visited cell bumps toward 1 until their real
// channels write them: IndexScore (surroundings overlap, per-edge), Alignment
// (lagged from Frame 2, per-edge), Stress (global — set from the slice when
// wired), Intrinsic (per-node BirthAlignment, set at crystallization). Tuning
// (the stress_modulator function, the threshold family, the blend) is TBD.
type memoryActivation struct {
	Identity    float64
	Familiarity float64
	IndexScore  float64
	Alignment   float64
	Stress      float64
	Intrinsic   float64
}

func newMemoryActivation() *memoryActivation {
	return &memoryActivation{IndexScore: 1, Alignment: 1, Stress: 1, Intrinsic: 1}
}

func (a *memoryActivation) Activate(visited bool, _ int) float64 {
	sample := 0.0
	if visited {
		sample = a.IndexScore * a.Alignment * a.Stress * a.Intrinsic
	}
	a.Identity += alphaIdentity * (sample - a.Identity)
	a.Familiarity += alphaFamiliarity * (sample - a.Familiarity)
	return a.Value()
}

// Value is the current activation (groove + recency) without bumping or decaying.
func (a *memoryActivation) Value() float64 { return a.Identity + a.Familiarity }

// SetContent replaces the node's content — used by the grade half to stamp a
// Prediction's grade in place (the grade lives ON the prediction node,
// cell-parsed). Guarded by the cell's own lock (concurrent frame reads).
func (m *MemoryCell) SetContent(content any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.content = content
}

package skyra

type ActivationRecord interface {
	Activate(visited bool, traversalCount int) float64
}

type Impulse struct {
	Primitive string
	ResultMap string
	Payload   any
	// Source describes where the impulse appears to arise. It is contextual
	// metadata, not the wave's return origin.
	Source    *ImpulseSource
	Amplitude float64
	Return    bool
}

type ImpulseSource struct {
	SpaceID string
	PointID string
	Coord   Coord
}

type Coord struct {
	X int
	Y int
}

type Time interface {
	ID() string
	// Realize runs a wave. It returns nothing: a wave's output is written to
	// the traversed Slice and returned to the origin as a perception wave. A
	// return value would be a second, chosen results channel — the caller
	// collecting the outcome of its own send — which the Perception Invariant
	// forbids. Void leaves exactly one path for a result: involuntary return
	// perception. origin owns the wave; traversed is the realm this wave walks.
	Realize(origin, traversed Space, impulse any)
}

// Space — a cell in the graph.
//
// Activation is owned by the EDGE, not the node (the synaptic model): the same
// shared content cell B can be excited differently by A and by C. This needs no
// new interface — activation records still live on the slice; the kernel just
// keys them off the *map key* the cell provides (relation.recordFor), not the
// child's own ID(). So a cell hands the kernel edge-scoped keys (e.g. "A→B" vs
// "C→B") in its Children() map and the records become edge-owned for free; the
// kernel stays oblivious (it keys a record off whatever identity it was handed).
// The world's map keys equal its point IDs, so it keeps node-keying with no
// change. Transitive reach is NOT a stored edge: A's excitation of a 2-hop cell
// is the path composition, computed at traversal time via amplitude attenuation;
// a path that grades above the keep-threshold crystallizes into a direct edge
// (Hebbian — the cell adds the target to its own Children()). See
// design-docs/kernel/activation-and-traversal.md and specs/2026-06-16/kernel.md.
type Space interface {
	ID() string
	Slice() Slice
	NewActivation() ActivationRecord
	// origin is the wave OWNER (who the wave returns to) — the same meaning in
	// both Observe and Express, always non-nil during traversal. It is NOT a
	// "nil means traversal" marker anymore. A being's own Observe is pure intake;
	// a being acts outward by calling its environment's entry (World.Observe,
	// which delegates to World.Emit) with itself as origin — not through its own
	// Observe. The SENDER of a signal (for entity resolution / the surroundings
	// map) is separate — it rides on the impulse (impulse.Source), not on origin.
	Observe(origin Space, impulse any)
	Express(origin Space, impulse any) any
	// Children is the ONE map the kernel walks — past edges only
	// (understandings, stresses, responses, actions; for setpoints, the
	// confirmed past). There is no "parent"/future edge: the future is not
	// stored per-cell — it lives as Global's grouping index over the active
	// field and is passed down as context. The map KEY is the edge identity —
	// what the kernel keys the activation record off of.
	Children() map[string]Space
	Content() any
}

type Slice interface {
	ID() string
	Tree() *TraversalPair
	Result() any
	SetResult(any)
	Maps() map[string]any
	Records() map[string]ActivationRecord
	Threshold() float64
	Count() int
	SetCount(int)
	// Lock/Unlock serialize the waves touching one being's slice. The kernel
	// holds the lock for a wave; cell methods run under it and never re-lock.
	Lock()
	Unlock()
}

type Node struct {
	CellID    string
	Cell      Space
	Record    ActivationRecord
	Score     float64 // single activation, carried from when the parent scored this cell
	Amplitude float64 // the wave's signal strength at this node's depth
	Children  []*Node // the tree, built on ascent
	Survivors []*Node // pre-scored child nodes above threshold, queued for descent
	NextChild int
}

type TraversalPair struct {
	Current []*Node
	Prior   []*Node
}

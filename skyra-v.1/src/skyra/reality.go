package skyra

type ActivationRecord interface {
	Activate(visited bool, traversalCount int) float64
}

type Kernel interface {
	ID() string
	// Realize runs a wave: OriginReality owns it (the discriminator and the
	// result recipient), TraversedReality is the realm this wave traverses.
	Realize(origin, traversed Reality, impulse any) any
}

type Reality interface {
	ID() string
	NewActivation() ActivationRecord
	// origin is the wave's owner — the cell compares it to its own being to
	// tell a self-wave (reafference) from an external one (exafference).
	Observe(origin Reality, impulse any)
	Express(origin Reality, impulse any) any
	Children() map[string]Reality
	Content() any
}

type Slice interface {
	ID() string
	Tree() *TraversalPair
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
	Cell      Reality
	Record    ActivationRecord
	Score     float64    // single activation, carried from when the parent scored this cell
	Amplitude float64    // the wave's signal strength at this node's depth
	Children  []*Node    // the tree, built on ascent
	Survivors []*Node    // pre-scored child nodes above threshold, queued for descent
	NextChild int
}

type TraversalPair struct {
	Current []*Node
	Prior   []*Node
}

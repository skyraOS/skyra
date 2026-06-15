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

type Space interface {
	ID() string
	Slice() Slice
	NewActivation() ActivationRecord
	// During traversal, origin is nil: the Space is being walked and must not
	// treat Observe as an environment invocation. Return perception is a new
	// wave rooted at the origin, not a direct Observe callback.
	Observe(origin Space, impulse any)
	Express(origin Space, impulse any) any
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

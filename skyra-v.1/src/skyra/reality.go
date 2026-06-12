package skyra

type ActivationRecord interface {
	Activate(visited bool, traversalCount int) float64
}

type Kernel interface {
	ID() string
	Realize(root Reality, payload any) any
}

type Reality interface {
	ID() string
	NewActivation() ActivationRecord
	Observe(k Kernel, payload any)
	Express(k Kernel, payload any) any
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

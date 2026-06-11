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
	Score     float64
	Children  []*Node
	Survivors []Reality
	NextChild int
}

type TraversalPair struct {
	Current []*Node
	Prior   []*Node
}

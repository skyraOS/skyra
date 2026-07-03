//go:build legacy

package reality

// ActivationRecord — the one interface the kernel calls.
// Each cell type defines its own implementation with whatever internals it wants.
type ActivationRecord interface {
	Activate(visited bool, traversalCount int) float64
}

// IRelation — the wave + shared substrate. Drives traversal, holds per-being state.
type IRelation interface {
	ID() string
	Realize(beingID string, brID string, root Reality)
	GetSlice(beingID string) *Slice
}

// Reality — the inhabitant. Cells implement this. The kernel traverses these.
type Reality interface {
	ID() string
	NewActivation() ActivationRecord
	Observe(r *Relation)
	Express(r *Relation)
	Children() map[string]Reality
}

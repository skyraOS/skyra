//go:build legacy

package reality

import "math"

// BaseActivation — embed into binding reality activations to get standard EMA + recency scoring.
// Binding realities extend this with their own fields and override Activate if needed.
type BaseActivation struct {
	Identity       float64 // EMA α=0.02 — the groove
	Familiarity    float64 // EMA α=0.1 — recent relationship
	LocalTraversal int     // per-cell visit count
	LastTraversed  int     // global count when last bumped
}

func NewBaseActivation() BaseActivation {
	return BaseActivation{
		Identity:    1.0,
		Familiarity: 1.0,
	}
}

func (b *BaseActivation) Activate(visited bool, traversalCount int) float64 {
	if visited {
		b.Identity = ema(b.Identity, 1.0, 0.02)
		b.Familiarity = ema(b.Familiarity, 1.0, 0.1)
		b.LocalTraversal++
		b.LastTraversed = traversalCount
	} else {
		b.Identity = ema(b.Identity, 0.0, 0.02)
		b.Familiarity = ema(b.Familiarity, 0.0, 0.1)
	}

	rec := recency(traversalCount, b.LastTraversed)
	return b.Identity * b.Familiarity * rec
}

func ema(current, target, alpha float64) float64 {
	return current + alpha*(target-current)
}

func recency(globalCount, lastTraversed int) float64 {
	delta := globalCount - lastTraversed
	if delta <= 0 {
		return 1.0
	}
	return 1.0 / (1.0 + math.Log(float64(delta)))
}

package reality

import "math"

type Reality interface {
	ID() string
	Create(r *Relation) Reality
	Realize(r *Relation) *Relation
	Competence(r *Relation)
	Observe(r *Relation)
	Express(r *Relation) *Relation
	Activation(rel *Relation) float64
	Reinforce()
	Decay()
}

// Recency — short term memory. Power law decay.
// Crashes fast, whispers forever. The echo.
func Recency(traversalCount, lastTraversed int) float64 {
	delta := traversalCount - lastTraversed
	if delta <= 0 {
		return 1.0
	}
	return math.Pow(float64(delta), -0.5)
}

// Familiarity — medium term memory. EMA with high α (0.1).
// What the being has been doing. Responsive, fades in weeks.
func Familiarity(current float64, traversed bool) float64 {
	const alpha = 0.1
	if traversed {
		return alpha*1.0 + (1-alpha)*current
	}
	return (1 - alpha) * current
}

// Identity — long term memory. EMA with low α (0.02).
// What the being IS. Barely moves. The groove.
func Identity(current float64, traversed bool) float64 {
	const alpha = 0.02
	if traversed {
		return alpha*1.0 + (1-alpha)*current
	}
	return (1 - alpha) * current
}

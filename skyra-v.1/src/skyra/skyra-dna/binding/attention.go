package binding

import (
	"skyra-v1/src/reality"
)

// Attention — the resonance filter. What penetrates the membrane.
// Surroundings arrive via propagation. Attention decides what gets through
// based on primitive alignment and goal relevance.
// The filtered set is what binding realities see.
type Attention struct {
	id      string
	BeingID string
}

func (a *Attention) ID() string {
	return a.id
}

func (a *Attention) NewActivation() reality.ActivationRecord {
	return nil
}

func (a *Attention) Observe(r *reality.Relation) {
}

func (a *Attention) Express(r *reality.Relation) {
}

func (a *Attention) Children() map[string]reality.Reality {
	return nil
}

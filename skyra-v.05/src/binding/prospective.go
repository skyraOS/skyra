package binding

import (
	"skyra-v1/src/reality"
)

// Prospective — external obligation tracker. Sub-binding reality under Goal.
// Indexes by the external obligation itself.
// Tracks deferred intentions triggered by external cues.
type Prospective struct {
	id      string
	BeingID string
}

func (p *Prospective) ID() string {
	return p.id
}

func (p *Prospective) Create(r *reality.Relation) reality.Reality {
	return &Prospective{
		id: reality.NewID(),
	}
}

func (p *Prospective) Observe(r *reality.Relation) {
	// read incoming signal for external requests/obligations
	// track delta toward resolution for existing obligations
}

func (p *Prospective) Express(r *reality.Relation) *reality.Relation {
	// TODO: surface unresolved obligations as goals
	return r
}

func (p *Prospective) Realize(r *reality.Relation) *reality.Relation {
	p.Observe(r)
	r = p.Express(r)
	return r
}

func (p *Prospective) Competence(r *reality.Relation) {
}

func (p *Prospective) Activation(r *reality.Relation) float64 {
	return 1.0
}

func (p *Prospective) Reinforce() {
}

func (p *Prospective) Decay() {
}

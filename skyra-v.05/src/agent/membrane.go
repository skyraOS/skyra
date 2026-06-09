package agent

import (
	"skyra-v1/src/reality"
)

// Membrane — the being. The boundary between inside and outside.
type Membrane struct {
	id           string
	Brain        map[string]reality.Reality
	ActivationFn func() reality.ActivationRecord // set by genome at boot
}

func (m *Membrane) ID() string {
	return m.id
}

func (m *Membrane) NewActivation() reality.ActivationRecord {
	return m.ActivationFn()
}

func (m *Membrane) Observe(r *reality.Relation) {
}

func (m *Membrane) Express(r *reality.Relation) {
}

func (m *Membrane) Children() map[string]reality.Reality {
	return m.Brain
}

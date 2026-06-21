package dna

import "skyra-v1/src/skyra"

// Membrane — the being's sensorimotor boundary.
//
// Observe = the attention gate (pass-through stub for MVP; ownership: alignment vs. working memory → pass or discard).
// Express = the motor (stub for MVP; ownership: coincidence + fire, conflict → Coherence).
// Content = the root drive (durable; contrast with Coherence's transient synthesis).
type MembraneCell struct {
	id        string
	rootDrive string
	brain     map[string]skyra.Space
	slice     skyra.Slice
}

func NewMembraneCell(id, rootDrive string, slice skyra.Slice) *MembraneCell {
	return &MembraneCell{
		id:        id,
		rootDrive: rootDrive,
		brain:     map[string]skyra.Space{},
		slice:     slice,
	}
}

func (m *MembraneCell) ID() string                            { return m.id }
func (m *MembraneCell) Slice() skyra.Slice                    { return m.slice }
func (m *MembraneCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (m *MembraneCell) Content() any                          { return m.rootDrive }
func (m *MembraneCell) Children() map[string]skyra.Space      { return m.brain }

func (m *MembraneCell) Observe(origin skyra.Space, impulse any) {}

func (m *MembraneCell) Express(origin skyra.Space, impulse any) any { return nil }

func (m *MembraneCell) AddChild(cell skyra.Space) {
	m.brain[cell.ID()] = cell
}

type memActivation struct{}

func (a *memActivation) Activate(_ bool, _ int) float64 { return 1.0 }

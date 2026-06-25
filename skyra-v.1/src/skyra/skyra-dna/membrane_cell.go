package dna

import "skyra-v1/src/skyra"

// Membrane — the being's sensorimotor boundary, as a thin dispatcher.
//
// The membrane does NO perception work itself. It is a trampoline:
//   - Observe  = nothing (the boundary doesn't perceive; sensory does).
//   - Children = none (the kernel goes straight from Observe to Express).
//   - Express  = fire a wave inward to SENSORY, on every wave (unconditional).
//   - Content  = the root drive (durable).
//
// Sensory and motor are separate sub-actors on their OWN slices (injected at
// construction), reached by cell-initiated crossing (go kernel.Realize). The
// spawn sets origin=sensory: the perception is sensory's own, so the membrane
// only launches the wave and drops out of the loop (the kernel never re-runs it).
// The spawn is async (go) so it locks sensory's own slice, never the membrane's
// held lock. See specs/2026-06-16/skyra-dna/sensorimotor.md.
type MembraneCell struct {
	id        string
	rootDrive string
	sensory   *SensoryCell
	motor     *MotorCell
	slice     skyra.Slice
	kernel    skyra.Time
}

// NewMembraneCell wires the boundary. Slices are INJECTED — one per actor
// (membrane, sensory, motor) — so dna stays free of a world dependency (the
// caller mints the concrete slices). kernel is the runtime's send primitive
// (skyra.Time), used to spawn the inward sensory wave on Express.
func NewMembraneCell(id, rootDrive string, slice, sensorySlice, motorSlice skyra.Slice, kernel skyra.Time) *MembraneCell {
	m := &MembraneCell{
		id:        id,
		rootDrive: rootDrive,
		slice:     slice,
		kernel:    kernel,
	}
	m.sensory = NewSensoryCell(id+":sensory", id, sensorySlice)
	m.motor = NewMotorCell(id+":motor", id, motorSlice)
	return m
}

func (m *MembraneCell) ID() string                            { return m.id }
func (m *MembraneCell) Slice() skyra.Slice                    { return m.slice }
func (m *MembraneCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (m *MembraneCell) Content() any                          { return m.rootDrive }

// Children — none. The membrane is a thin boundary; the kernel descends nothing
// and goes straight to Express.
func (m *MembraneCell) Children() map[string]skyra.Space { return nil }

// Observe — nothing. Perception is sensory's, not the boundary's.
func (m *MembraneCell) Observe(origin skyra.Space, impulse any) {}

// Express — fire the inward wave to sensory, on EVERY wave (unconditional:
// traffic-relative decay needs sensory's wave to run even when nothing known is
// perceived). origin=sensory (sensory's own intake); async (go) so the spawn
// locks sensory's own slice, never the membrane's held lock. The membrane returns
// nothing and drops out of the loop.
func (m *MembraneCell) Express(origin skyra.Space, impulse any) any {
	if m.kernel != nil {
		go m.kernel.Realize(m.sensory, m.sensory, impulse)
	}
	return nil
}

func (m *MembraneCell) Sensory() *SensoryCell { return m.sensory }
func (m *MembraneCell) Motor() *MotorCell     { return m.motor }

type memActivation struct{}

func (a *memActivation) Activate(_ bool, _ int) float64 { return 1.0 }

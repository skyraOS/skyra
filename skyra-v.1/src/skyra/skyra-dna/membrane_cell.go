package dna

import (
	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

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

// BuildContext is threaded once down the pipeline so the constructors don't
// balloon as stages are added (membrane → sensory → global → other → …). It
// carries the being id, a slice FACTORY (so dna stays free of any concrete
// slice/world dependency — the caller provides it), and the kernel (the send
// primitive each stage uses to spawn the next). Still injection — a factory
// instead of N concrete slices.
type BuildContext struct {
	BeingID  string
	NewSlice func(id string) skyra.Slice
	Kernel   skyra.Time
	// Infer is the LLM seam (system, present) → response. Injected so dna stays
	// free of the inference package (the app passes inference.Call; tests pass a
	// mock). nil → cells that would infer simply skip it.
	Infer InferFunc
	// Motor — the pipeline terminus (the being's out-face). Created by the
	// membrane and threaded so the chain's tail (voting) spawns into this same
	// motor rather than minting a second one.
	Motor *MotorCell
}

// InferFunc is the injected LLM call: a system prompt + the present (the frame),
// returning the model's text. The cell composes what it sends and parses what
// comes back (see parse.go).
type InferFunc func(system, present string) (string, error)

// sliceFor mints a slice for a sub-actor (suffix like ":sensory"); nil if no
// factory was supplied (every real build supplies one).
func (c BuildContext) sliceFor(suffix string) skyra.Slice {
	if c.NewSlice == nil {
		return nil
	}
	return c.NewSlice(c.BeingID + suffix)
}

// NewMembraneCell wires the boundary from a BuildContext. Each actor mints its
// own slice via the factory; the kernel is handed down for the forward spawns.
func NewMembraneCell(rootDrive string, ctx BuildContext) *MembraneCell {
	m := &MembraneCell{
		id:        ctx.BeingID,
		rootDrive: rootDrive,
		slice:     ctx.sliceFor(":membrane"),
		kernel:    ctx.Kernel,
	}
	// Motor is the out-face AND the pipeline terminus — create it once and thread
	// it (ctx.Motor) so the chain's tail (voting) spawns into this same motor.
	m.motor = NewMotorCell(ctx)
	ctx.Motor = m.motor
	m.sensory = NewSensoryCell(ctx)
	m.motor.SetMidLevel(m.sensory.MidLevel()) // motor's read path → the grouping spine
	return m
}

func (m *MembraneCell) ID() string                            { return m.id }
func (m *MembraneCell) Slice() skyra.Slice                    { return m.slice }
func (m *MembraneCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (m *MembraneCell) Content() any                          { return m.rootDrive }

// Children — none. The membrane is a thin boundary; the kernel descends nothing
// and goes straight to Express.
func (m *MembraneCell) Children() map[string]skyra.Space { return nil }

func (m *MembraneCell) Agents() map[string]skyra.Space {
	agents := map[string]skyra.Space{}
	if m.sensory != nil {
		agents[m.sensory.ID()] = m.sensory
	}
	if m.motor != nil {
		agents[m.motor.ID()] = m.motor
	}
	return agents
}

// Observe — nothing. Perception is sensory's, not the boundary's.
func (m *MembraneCell) Observe(origin skyra.Space, impulse any) {
	if imp, ok := asImpulse(impulse); ok {
		src := ""
		if imp.Source != nil {
			src = imp.Source.SpaceID
		}
		debug.Being("skyra", "membrane", "observe from="+src+" primitive="+imp.Primitive)
	}
}

// Express — fire the inward wave to sensory, on EVERY wave (unconditional:
// traffic-relative decay needs sensory's wave to run even when nothing known is
// perceived). origin=sensory (sensory's own intake); async (go) so the spawn
// locks sensory's own slice, never the membrane's held lock. The membrane returns
// nothing and drops out of the loop.
func (m *MembraneCell) Express(origin skyra.Space, impulse any) any {
	if m.kernel != nil {
		debug.Being("skyra", "membrane", "express → sensory")
		go m.kernel.Realize(m.sensory, m.sensory, impulse)
	}
	return nil
}

func (m *MembraneCell) Sensory() *SensoryCell { return m.sensory }
func (m *MembraneCell) Motor() *MotorCell     { return m.motor }

type memActivation struct{}

func (a *memActivation) Activate(_ bool, _ int) float64 { return 1.0 }

package tests

import (
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

// newBeing builds a being with the real kernel and three INJECTED slices
// (membrane, sensory, motor), per the sensorimotor-boundary model
// (specs/2026-06-16/skyra-dna/sensorimotor.md). The caller mints the concrete
// slices; the membrane does not.
func newBeing(id string) (*dna.MembraneCell, *relation.Relation) {
	kernel := relation.NewRelation()
	return newBeingWithKernel(id, kernel), kernel
}

// newBeingWithKernel is newBeing with a caller-supplied kernel (e.g. a spy that
// records the membrane's spawn instead of running it).
func newBeingWithKernel(id string, kernel skyra.Time) *dna.MembraneCell {
	return dna.NewMembraneCell(id, "be helpful",
		world.NewSlice(id+":membrane", id),
		world.NewSlice(id+":sensory", id),
		world.NewSlice(id+":motor", id),
		kernel)
}

// sensoryEntity reads the sensory entity index under the sensory slice lock, so a
// test polling for the effect of an async (spawned) sensory wave is synchronized
// with that wave's locked Observe writes — no data race.
func sensoryEntity(s *dna.SensoryCell, id string) (skyra.Space, bool) {
	sl := s.Slice()
	sl.Lock()
	defer sl.Unlock()
	e, ok := s.ByEntity[id]
	return e, ok
}

// realizeCall captures one kernel.Realize invocation for assertions.
type realizeCall struct {
	origin    skyra.Space
	traversed skyra.Space
	impulse   any
}

// spyKernel is a skyra.Time that records Realize calls on a buffered channel,
// letting a test observe the membrane's async spawn deterministically.
type spyKernel struct {
	calls chan realizeCall
}

func newSpyKernel() *spyKernel { return &spyKernel{calls: make(chan realizeCall, 4)} }

func (k *spyKernel) ID() string { return "spy" }

func (k *spyKernel) Realize(origin, traversed skyra.Space, impulse any) {
	k.calls <- realizeCall{origin: origin, traversed: traversed, impulse: impulse}
}

package tests

import (
	"testing"
	"time"

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
	ctx := dna.BuildContext{
		BeingID:  id,
		NewSlice: func(sliceID string) skyra.Slice { return world.NewSlice(sliceID, id) },
		Kernel:   kernel,
	}
	return dna.NewMembraneCell("be helpful", ctx)
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

// assertSpawn calls a cell's Express and asserts it spawned a forward wave whose
// origin and traversed are both the cell named by wantID (the next stage owns
// its own wave). Deterministic: the spy delivers the spawn on a channel.
func assertSpawn(t *testing.T, spy *spyKernel, express func(skyra.Space, any) any, wantID string) {
	t.Helper()
	express(nil, skyra.Impulse{Payload: "x"})
	select {
	case call := <-spy.calls:
		if call.traversed == nil || call.traversed.ID() != wantID {
			t.Fatalf("spawn traversed = %v, want %s", call.traversed, wantID)
		}
		if call.origin == nil || call.origin.ID() != wantID {
			t.Fatalf("spawn origin = %v, want %s", call.origin, wantID)
		}
	case <-time.After(time.Second):
		t.Fatalf("no forward spawn to %s on Express", wantID)
	}
}

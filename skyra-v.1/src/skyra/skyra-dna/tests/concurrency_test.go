package tests

import (
	"sync"
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
)

// TestConcurrentFanoutRaceFree exercises the concurrency the substrate must
// support NOW: the two-stroke fan-out — self and other running at once, both
// reading MidLevel's grouping spine and both writing predictions to it, while a
// cross-cell reader (WorldState) reads sensory. With inference lockless (kernel
// change) and the cross-cell reads/writes locked, this is race-free. Run -race.
func TestConcurrentFanoutRaceFree(t *testing.T) {
	kernel := relation.NewRelation()
	being := newBeingWith("skyra", kernel, gradingMock(t, "<match>1</match>", "<grade>yes</grade>", nil))
	sensory := being.Sensory()
	mid := sensory.MidLevel()
	other := mid.Other()
	self := other.Self()

	// establish a grouping + a prior distribution (sequential)
	imp := engageFrom("michael", "hello there")
	sensory.Slice().SetCount(1)
	mid.Slice().SetCount(1)
	sensory.Observe(nil, imp)
	mid.Observe(nil, imp)
	other.Observe(nil, imp)

	// the fan-out: self and other Observe concurrently (both touch MidLevel's
	// spine), plus concurrent cross-cell reads of sensory via WorldState.
	mid.Slice().SetCount(2)
	sensory.Slice().SetCount(2)
	imp2 := engageFrom("michael", "how's it going")
	sensory.Observe(nil, imp2)

	// The fan-out: ONE other ∥ ONE self (different cells), plus concurrent
	// cross-cell readers. This is the real two-stroke shape — not N same-cell
	// waves (that is scenario 4, below).
	var wg sync.WaitGroup
	wg.Add(4)
	go func() { defer wg.Done(); other.Observe(nil, imp2) }()
	go func() { defer wg.Done(); self.Observe(nil, imp2) }()
	go func() { defer wg.Done(); _ = mid.WorldState() }()
	go func() { defer wg.Done(); _ = mid.WorldState() }()
	wg.Wait()
}

// TestConcurrentEngagesRapidFire is scenario 4 (rapid-fire / non-turn-taking):
// many external engages hitting the SAME cells at once. This needs the trigger
// coalescer (deferred) + per-wave state moved off cell fields (g.frame/g.gist,
// o.reframed/o.verdict) — both deferred in the specs. Skipped until then; kept so
// the target is named. See specs/2026-07-03/two-stroke-cycle.md (scenario 4, [B]).
func TestConcurrentEngagesRapidFire(t *testing.T) {
	t.Skip("scenario 4: needs the trigger coalescer + per-wave-state refactor (deferred)")
	kernel := relation.NewRelation()
	being := newBeingWith("skyra", kernel, gradingMock(t, "<match>1</match>", "<grade>yes</grade>", nil))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kernel.Realize(being, being, skyra.Impulse{
				Primitive: "engage",
				Source:    &skyra.ImpulseSource{SpaceID: "michael"},
				Payload:   "hi",
			})
		}()
	}
	wg.Wait()
}

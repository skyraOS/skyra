package dnav2

import (
	"sync"
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	"skyra-v1/src/skyra/world"
)

// A transcribed root runs its selected gene as the kernel's bookends, reaches the
// DNA for durable state, and its per-wave state (Condition) is its own.
func TestTranscriptRunsSelectedGene(t *testing.T) {
	d := testDNA()
	// a gene that writes durable state on the DNA (the clock) and a per-wave
	// condition on the transcript (the correct split: durable → DNA, per-wave → root)
	d.RegisterObserve("test:work", func(root *EphemeralRoot, origin skyra.Space, impulse any) {
		root.DNA().Tick()
		root.Condition = "did-work"
	})
	d.RegisterExpress("test:work", func(root *EphemeralRoot, origin skyra.Space, impulse any) any {
		return root.Condition // Express sees what Observe set on the same transcript
	})

	sl := world.NewSlice("skyra:test", "skyra")
	root := d.Transcribe("skyra:test", "test:work", "test:work", sl, "imp-1", nil)

	kernel := relation.NewRelation()
	kernel.Realize(root, root, skyra.Impulse{})

	if d.Tick() != 2 { // gene ticked to 1; this Tick reads 2
		t.Fatal("gene did not run / did not reach the DNA clock")
	}
	if root.Condition != "did-work" {
		t.Fatalf("per-wave condition = %q, want did-work", root.Condition)
	}
}

// The scenario-4 property, by construction: two concurrent transcripts of the same
// gene never clobber each other's per-wave state — each has its own root. Durable
// writes to the shared DNA are guarded. Run with -race.
func TestConcurrentTranscriptsIsolated(t *testing.T) {
	d := testDNA()
	d.RegisterObserve("test:count", func(root *EphemeralRoot, origin skyra.Space, impulse any) {
		root.Condition = root.ImpulseID // per-wave: each root records its own impulse
		root.DNA().Tick()               // durable: guarded shared write
	})

	kernel := relation.NewRelation()
	const waves = 64
	var wg sync.WaitGroup
	wg.Add(waves)
	for i := 0; i < waves; i++ {
		go func(k int) {
			defer wg.Done()
			id := "imp-" + string(rune('A'+k%26))
			sl := world.NewSlice("skyra:count:"+id, "skyra") // own active-state slice per wave-root
			root := d.Transcribe("skyra:count", "test:count", "", sl, id, nil)
			kernel.Realize(root, root, skyra.Impulse{})
			if root.Condition != id {
				t.Errorf("per-wave state clobbered: got %q want %q", root.Condition, id)
			}
		}(i)
	}
	wg.Wait()

	if got := d.Tick() - 1; got != waves { // -1 for this read's own tick
		t.Fatalf("durable count = %d, want %d (guarded shared write lost increments)", got, waves)
	}
}

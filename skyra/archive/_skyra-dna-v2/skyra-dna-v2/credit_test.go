package dnav2

import "testing"

// Credit write-back reaches every edge in the frame, and scales by activation: on a good
// closure the more load-bearing edge (b1, the higher-weighted branch) learns more than
// the incidental one (b2). This is the eligibility trace — the graph as the trace.
func TestDistributeCreditGradesFrameByActivation(t *testing.T) {
	g := buildFrameEpisode()
	store := NewActivationStore()

	// Assemble the frame (store empty → static Probability weights: b1 .7, b2 .3).
	frame := FrameOf(g, 10000, store, 0)
	edges := frame.Edges()
	if len(edges) < 4 { // episode→frame, frame→root, root→b1, root→b2
		t.Fatalf("frame has %d edges, want ≥4: %v", len(edges), edges)
	}

	b1, b2 := "g1:frame:1→final-pred:1-0", "g1:frame:1→final-pred:1-1"
	if edges[b1] <= edges[b2] {
		t.Fatalf("expected b1 more active than b2: b1=%.3f b2=%.3f", edges[b1], edges[b2])
	}

	// Crossing (frame assembly) then a good closure grade distributed over the frame.
	store.CrossFrame(edges, 0)
	store.DistributeCredit(edges, 1, 1)

	// Every lit edge now has a record; the more load-bearing edge learned more.
	sb1, ok1 := store.Strength(b1, 0)
	sb2, ok2 := store.Strength(b2, 0)
	if !ok1 || !ok2 {
		t.Fatalf("both edges should have records after credit: b1=%v b2=%v", ok1, ok2)
	}
	if sb1 <= sb2 {
		t.Fatalf("higher-activation edge should learn more: b1=%.3f b2=%.3f", sb1, sb2)
	}
	if sb2 <= 0 {
		t.Fatalf("incidental edge still gets some credit + crossing: b2=%.3f, want >0", sb2)
	}
}

// A negative closure drives the frame's edges DOWN — blame flows the same way credit does.
func TestDistributeBlameLowersStrength(t *testing.T) {
	g := buildFrameEpisode()
	store := NewActivationStore()
	edges := FrameOf(g, 10000, store, 0).Edges()
	store.CrossFrame(edges, 0)

	b1 := "g1:frame:1→final-pred:1-0"
	before, _ := store.Strength(b1, 0) // crossed, A still 0
	store.DistributeCredit(edges, -1, 1)
	after, _ := store.Strength(b1, 0)

	if after >= before {
		t.Fatalf("a bad closure should lower strength: before=%.3f after=%.3f", before, after)
	}
}

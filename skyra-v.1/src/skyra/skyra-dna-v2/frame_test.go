package dnav2

import (
	"strings"
	"testing"

	dna "skyra-v1/src/skyra/skyra-dna"
)

// buildGrouping mints a grouping on the gist-owned spine — shell → gist →
// two-branch other-distribution, the same shape Create + appendOtherDistribution
// produce.
func buildGrouping() *dna.MemoryCell {
	g := dna.NewMemoryCell("g1", dna.KindGrouping, dna.Grouping{Created: 1}, nil)
	gist := dna.NewMemoryCell("g1:gist:1", dna.KindGist, dna.Gist{Text: "help with the kernel", Timestamp: 1}, nil)
	g.AddChild("g1→g1:gist:1", gist)
	root := dna.NewMemoryCell("g1:gist:1→other-pred:1", dna.KindPrediction, dna.Prediction{Role: "other", Timestamp: 1}, nil)
	b1 := dna.NewMemoryCell("g1:gist:1→other-pred:1-0", dna.KindPrediction,
		dna.Prediction{Text: "asks about the kernel", Role: "other", Probability: 0.7, Timestamp: 1}, nil)
	b2 := dna.NewMemoryCell("g1:gist:1→other-pred:1-1", dna.KindPrediction,
		dna.Prediction{Text: "asks about setup", Role: "other", Probability: 0.3, Timestamp: 1}, nil)
	root.AddChild("g1:gist:1→other-pred:1-0", b1)
	root.AddChild("g1:gist:1→other-pred:1-1", b2)
	gist.AddChild("g1:gist:1→other-pred:1", root)
	return g
}

// A generous budget lights the gist and both branches — the traversal reconstructs what
// renderOtherDistribution hand-builds.
func TestAssembleFrameLightsGistAndBranches(t *testing.T) {
	present, lit := AssembleFramePresent(buildGrouping(), 10000)

	if !strings.Contains(present, "Framing: help with the kernel") {
		t.Fatalf("frame missing gist:\n%s", present)
	}
	if !strings.Contains(present, "asks about the kernel") || !strings.Contains(present, "asks about setup") {
		t.Fatalf("frame missing branches:\n%s", present)
	}
	if lit < 5 { // grouping + gist + root + 2 branches
		t.Fatalf("expected ≥5 lit nodes, got %d:\n%s", lit, present)
	}
}

// A tight budget prunes the frame — the whole point of the token budget as a
// context-size knob. Fewer nodes light than under a generous budget.
func TestTightBudgetPrunesTheFrame(t *testing.T) {
	_, litFull := AssembleFramePresent(buildGrouping(), 10000)
	_, litTight := AssembleFramePresent(buildGrouping(), 12) // ~grouping + root, branches don't fit

	if litTight >= litFull {
		t.Fatalf("tight budget should light fewer nodes: tight=%d full=%d", litTight, litFull)
	}
}

// Learning reweights the frame: b1 has the higher static Probability (0.7 vs 0.3), but
// after grading b1's edge DOWN and b2's edge UP, the learned weights INVERT — the frame
// now favors b2. This is relevance = alignment overriding the cold prior.
func TestLearnedStrengthReweightsFrame(t *testing.T) {
	g := buildGrouping()
	gist := g.Children()["g1→g1:gist:1"].(*dna.MemoryCell)
	root := gist.Children()["g1:gist:1→other-pred:1"].(*dna.MemoryCell)
	b1Key, b2Key := "g1:gist:1→other-pred:1-0", "g1:gist:1→other-pred:1-1"

	// Phase-1 (no store): static Probability favors b1 (0.7 > 0.3).
	cold := (&frameNode{cell: root}).Frontier(nil, nil)
	if cold[b1Key].Weight <= cold[b2Key].Weight {
		t.Fatalf("cold prior should favor b1: b1=%.3f b2=%.3f", cold[b1Key].Weight, cold[b2Key].Weight)
	}

	// Grade b1 down, b2 up (both crossed at birth first).
	store := NewActivationStore()
	store.Cross(b1Key, 0)
	store.Grade(b1Key, -1, 1)
	store.Cross(b2Key, 0)
	store.Grade(b2Key, 1, 1)

	warm := (&frameNode{cell: root, store: store}).Frontier(nil, nil)
	if warm[b2Key].Weight <= warm[b1Key].Weight {
		t.Fatalf("learning should invert the order — b2 up, b1 down: b1=%.3f b2=%.3f",
			warm[b1Key].Weight, warm[b2Key].Weight)
	}
}

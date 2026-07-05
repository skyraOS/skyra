package dnav2

import (
	"strings"
	"testing"

	dna "skyra-v1/src/skyra/skyra-dna"
)

// buildGrouping mints a grouping with a gist and a two-branch other-distribution — the
// same shape appendOtherDistribution produces.
func buildGrouping() *dna.MemoryCell {
	g := dna.NewMemoryCell("g1", dna.KindGrouping, dna.Gist{Text: "help with the kernel", Timestamp: 1}, nil)
	root := dna.NewMemoryCell("g1→other-pred:1", dna.KindPrediction, dna.Prediction{Role: "other", Timestamp: 1}, nil)
	b1 := dna.NewMemoryCell("g1→other-pred:1-0", dna.KindPrediction,
		dna.Prediction{Text: "asks about the kernel", Role: "other", Probability: 0.7, Timestamp: 1}, nil)
	b2 := dna.NewMemoryCell("g1→other-pred:1-1", dna.KindPrediction,
		dna.Prediction{Text: "asks about setup", Role: "other", Probability: 0.3, Timestamp: 1}, nil)
	root.AddChild("g1→other-pred:1-0", b1)
	root.AddChild("g1→other-pred:1-1", b2)
	g.AddChild("g1→other-pred:1", root)
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
	if lit < 4 { // grouping + root + 2 branches
		t.Fatalf("expected ≥4 lit nodes, got %d:\n%s", lit, present)
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

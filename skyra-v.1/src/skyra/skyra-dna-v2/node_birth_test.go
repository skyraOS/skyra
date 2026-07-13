package dnav2

import (
	"testing"
)

// Node birth crosses the new edges, so they start selectable (F>0). Without this a fresh
// edge reads strength 0 (F=0), is never selected, and can never earn its alignment — the
// bootstrap deadlock.
func TestNodeBirthCrossesEdges(t *testing.T) {
	d := testDNA()
	g := d.Episodes().Create("g1", "framing", d.Tick())

	born := appendOtherDistribution(d.Episodes(), g, d.Tick(), []ParsedPrediction{
		{Text: "a", Probability: 0.6}, {Text: "b", Probability: 0.4},
	})
	if len(born) != 3 { // root + 2 branches
		t.Fatalf("born %d edges, want 3: %v", len(born), born)
	}

	// Before birth-cross the edges have no record.
	for _, k := range born {
		if _, ok := d.Activation().Strength(k, 0); ok {
			t.Fatalf("edge %s should have no record before birth-cross", k)
		}
	}

	d.CrossBirth(born...)

	// After: every birth edge is selectable (F>0, A=0 → rect 0.5 → strength > 0).
	for _, k := range born {
		s, ok := d.Activation().Strength(k, 0)
		if !ok || s <= 0 {
			t.Fatalf("birth edge %s not selectable after cross: s=%.3f ok=%v", k, s, ok)
		}
	}
}

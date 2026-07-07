package dnav2

import (
	"math"
	"testing"

	dna "skyra-v1/src/skyra/skyra-dna"
)

// buildEpisode mints a grouping (gist-owned spine) with one consolidated interaction
// under its gist — the minimal shape the settle grades (a fresh episode after
// consolidateInteraction). The active gist itself is never a settle item.
func buildEpisode() (g *dna.MemoryCell, edgeKey string) {
	g = dna.NewMemoryCell("g1", dna.KindGrouping, dna.Grouping{Created: 1}, nil)
	gist := dna.NewMemoryCell("g1:gist:1", dna.KindGist, dna.Gist{Text: "help with the skyra project", Timestamp: 1}, nil)
	g.AddChild("g1→g1:gist:1", gist)
	edgeKey = "g1:gist:1→michael:int"
	inter := dna.NewMemoryCell("michael:int", dna.KindResponse,
		dna.InteractionRecord{From: "michael", Utterance: "can you help with the skyra project", Timestamp: 1}, nil)
	gist.AddChild(edgeKey, inter)
	return g, edgeKey
}

// REGRESSION (Codex, 2026-07-07): the settle must CROSS the edges it grades. Grade mints a
// record at F=0, and Strength is a product (rectSoft(A)·S·F·conf), so a freshly-graded edge
// would read weight 0 — and, now that a record exists, bypass the static-weight fallback,
// deadening the edge with no way to recover (0 weight → never traversed → never crossed).
// After a settle, a positively-graded edge must read a POSITIVE strength.
func TestSettleDoesNotDeadenGradedEdge(t *testing.T) {
	g, edgeKey := buildEpisode()
	store := NewActivationStore()
	infer := func(system, present string) (string, error) {
		return `<grades><g n="1">0.9</g></grades>`, nil
	}

	passes := coherenceSettle(g, store, 5, infer)
	if passes < 1 {
		t.Fatalf("settle took %d passes, want ≥1 (the item should have graded)", passes)
	}

	s, ok := store.Strength(edgeKey, 5)
	if !ok {
		t.Fatal("settle wrote no record for the graded edge")
	}
	if s <= 0 {
		t.Fatalf("graded edge reads strength %.4f — DEADENED (F never crossed)", s)
	}

	all := store.All()
	a := all[edgeKey]
	if a.N != 1 {
		t.Fatalf("N=%d, want 1 (one grade)", a.N)
	}
	if math.Abs(a.A-0.9) > 1e-9 {
		t.Fatalf("A=%.4f, want 0.9 (the coherence grade)", a.A)
	}
	if a.F <= 0 {
		t.Fatalf("F=%.4f, want >0 (the settle crossed the edge)", a.F)
	}
}

// A negatively-graded (pruned) edge is dropped from this settle but must stay ALIVE — its
// strength floors at softFloor·… (recoverable), not 0. Pruning is exclude-this-settle, not
// death: the edge can be regraded up on a later turn.
func TestSettlePrunedEdgeStaysRecoverable(t *testing.T) {
	g, edgeKey := buildEpisode()
	store := NewActivationStore()
	infer := func(system, present string) (string, error) {
		return `<grades><g n="1">-0.8</g></grades>`, nil
	}

	coherenceSettle(g, store, 5, infer)

	s, ok := store.Strength(edgeKey, 5)
	if !ok {
		t.Fatal("settle wrote no record for the pruned edge")
	}
	if s <= 0 {
		t.Fatalf("pruned edge reads strength %.4f — should floor above 0 (softFloor), recoverable", s)
	}
}

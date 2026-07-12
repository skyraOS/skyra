package dnav2

import (
	"strings"
	"sync/atomic"
	"testing"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// querierMock overrides the framing branch of the standard pipeline mock.
func querierMock(frameResp func(present string) string) InferFunc {
	base := gradeMock("<match>none</match>")
	return func(system, present string) (string, error) {
		if strings.Contains(system, "evolving being") {
			return frameResp(present), nil
		}
		return base(system, present)
	}
}

// A well-formed annotated emission persists the compound, mints the set,
// activates it on the Slice (anchored on the being), and cross-links the spine
// gist to its canonical compound — Phase 1's whole contract.
func TestQuerierPersistsCompoundAndActivatesSet(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(querierMock(func(string) string {
		return "<new>michael[1] wants[1→2] help[2] with[2→3] {the skyra project}[3]</new>"
	}))
	d.SetEmit(func(skyra.Impulse) {})
	o, _ := newOrch(d)

	fireTurn(t, o, "X1", "", "can you help with the skyra project")

	set, ok := d.ActiveSet()
	if !ok {
		t.Fatal("no active AitiaSet after a framed turn")
	}
	if set.EfficientEdgeID != "" || set.FinalEdgeID != "" {
		t.Fatal("v0.1 sets carry only the present reference (causal edges are Phase 4)")
	}
	summary := d.Compounds().Get(set.PresentID)
	if summary == nil {
		t.Fatal("active set's present compound not persisted")
	}
	cc, ok := summary.Content().(CompoundContent)
	if !ok || cc.Sentence != "michael wants help with the skyra project" {
		t.Fatalf("compound sentence = %+v", summary.Content())
	}
	// the subgraph landed: observed michael resolves to the Entity; asserted-only
	// material remains a shared Term. No shadow michael term is created.
	if d.Compounds().Term("michael") != nil || d.Compounds().Term("the skyra project") == nil {
		t.Fatal("observed/asserted referent resolution is wrong")
	}
	structural := 0
	for _, e := range d.Compounds().EdgesOf(set.PresentID) {
		if e.Kind == EdgeStructural {
			structural++
		}
	}
	if structural != 2 {
		t.Fatalf("structural edges = %d, want 2", structural)
	}
	// the spine gist carries the clean sentence + the cross-link
	g := d.Groupings().Active()
	gc, _ := latestGist(g).Content().(dna.Gist)
	if gc.Text != "michael wants help with the skyra project" {
		t.Fatalf("gist text = %q, want the clean sentence", gc.Text)
	}
	if gc.CompoundID != set.PresentID {
		t.Fatalf("gist.CompoundID = %q, want %s", gc.CompoundID, set.PresentID)
	}
}

// A malformed annotation gets retry-with-correction and succeeds on the second
// emission — bounded, and the turn still frames.
func TestQuerierRetriesBadAnnotationThenSucceeds(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	var frameCalls int32
	d.SetInfer(querierMock(func(present string) string {
		atomic.AddInt32(&frameCalls, 1)
		if strings.Contains(present, "could not be parsed") {
			return "<new>michael[1] wants[1→2] help[2]</new>" // corrected
		}
		return "<new>michael[1] wants[1→9] help[2]</new>" // undeclared node 9
	}))
	d.SetEmit(func(skyra.Impulse) {})
	o, _ := newOrch(d)

	fireTurn(t, o, "X1", "", "can you help")

	if got := atomic.LoadInt32(&frameCalls); got != 2 {
		t.Fatalf("framing calls = %d, want 2 (one retry)", got)
	}
	set, ok := d.ActiveSet()
	if !ok {
		t.Fatal("retry should still land an active set")
	}
	cc, _ := d.Compounds().Get(set.PresentID).Content().(CompoundContent)
	if cc.Sentence != "michael wants help" {
		t.Fatalf("sentence = %q", cc.Sentence)
	}
}

// Persistent annotation failure commits nothing and retains prior active state.
func TestQuerierFormatFailureRetainsActiveState(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	var valid atomic.Bool
	d.SetInfer(querierMock(func(string) string {
		if valid.Load() {
			return "<new>just plain words with no annotations</new>"
		}
		return "<new>michael[1] greets[1→2] skyra[2]</new>"
	}))
	d.SetEmit(func(skyra.Impulse) {})
	o, _ := newOrch(d)

	fireTurn(t, o, "X1", "", "hello")
	before, ok := d.ActiveSet()
	if !ok {
		t.Fatal("valid first turn must activate a set")
	}
	beforeGrouping := d.Groupings().Active()

	valid.Store(true)
	trace := fireTurn(t, o, "X2", "X1", "this response will not parse")
	if got := condOf(trace, StepMidlevelFrame); got != "query-format-error" {
		t.Fatalf("midlevel condition = %q, want query-format-error", got)
	}
	after, ok := d.ActiveSet()
	if !ok || after.ID != before.ID {
		t.Fatal("format failure must retain the previous active Aitia set")
	}
	if d.Groupings().Active() != beforeGrouping {
		t.Fatal("format failure must not replace the active grouping")
	}
}

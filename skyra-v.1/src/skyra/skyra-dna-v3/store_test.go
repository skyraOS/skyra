package dnav3

import "testing"

// THE EDGE MUST ACCUMULATE ACROSS EPISODES.
//
// This is the bug the first live run found, and it was fatal: edgeKey included the expression
// INSTANCE id, so every run minted a fresh, unmeasured edge for the same claim. N stayed at 1
// forever, P never converged, and the being could not learn across episodes — which is the one
// thing this whole architecture exists to do.
//
// The reusable thing must be keyed by WHAT IT IS, not by the instance that made it.
func TestEdgeAccumulatesAcrossEpisodes(t *testing.T) {
	g := NewExpressionGraph()

	// the same act, emitted three times as three separate expressions — exactly what happens when
	// the being is given the same task on three different days.
	mk := func() *Expression {
		return &Expression{
			Acts: []Act{{ID: "a1", Target: "file:/x", Primitive: "write"}},
			Expects: []Expect{{
				Of:     []string{"a1"},
				Within: 1,
				Triple: Triple{Mode: ModeExact, Confidence: 0.9, Subject: "file:/x", Predicate: "exists", Object: "true"},
			}},
		}
	}

	var edges []*Edge
	for i, id := range []string{"e1", "e2", "e3"} {
		e := mk()
		g.Place(id, e)
		ed := g.Edge(id, e.Expects[0])
		ed.Observe(TruthTrue, i+1)
		edges = append(edges, ed)
	}

	// all three must be THE SAME edge
	if edges[0] != edges[1] || edges[1] != edges[2] {
		t.Fatal("three expressions doing the SAME THING produced three DIFFERENT edges — " +
			"the being cannot learn across episodes")
	}
	if edges[0].N != 3 {
		t.Errorf("N = %d after three observations, want 3 — the edge is not accumulating", edges[0].N)
	}
	if edges[0].P <= 0.5 {
		t.Errorf("P = %.2f after three TRUE observations, want > 0.5 — the EMA is not converging", edges[0].P)
	}
}

// But a DIFFERENT act producing the same state is a DIFFERENT edge — because "how did I make this
// true" is the question, and two different ways of making it true are two different answers.
func TestDifferentActsAreDifferentEdges(t *testing.T) {
	g := NewExpressionGraph()

	write := &Expression{
		Acts: []Act{{ID: "a1", Target: "file:/x", Primitive: "write"}},
		Expects: []Expect{{Of: []string{"a1"}, Within: 1,
			Triple: Triple{Subject: "file:/x", Predicate: "exists", Object: "true"}}},
	}
	shell := &Expression{
		Acts: []Act{{ID: "a1", Target: "dir:/", Primitive: "exec"}},
		Expects: []Expect{{Of: []string{"a1"}, Within: 1,
			Triple: Triple{Subject: "file:/x", Predicate: "exists", Object: "true"}}},
	}
	g.Place("e1", write)
	g.Place("e2", shell)

	if g.Edge("e1", write.Expects[0]) == g.Edge("e2", shell.Expects[0]) {
		t.Error("writing a file and shelling out to create it collapsed into one edge — " +
			"they have different reliabilities and must be learned separately")
	}
}

func TestBetsAtUsesAddressIndex(t *testing.T) {
	g := NewExpressionGraph()
	e := &Expression{
		SpanID: "episode:1/span:7",
		Acts:   []Act{{ID: "a1"}, {ID: "a2"}},
		Expects: []Expect{
			{Of: []string{"a1"}, Within: 1, Triple: Triple{Subject: "file:/x", Predicate: "exists", Object: "true"}},
			{Of: []string{"a2"}, Within: 1, Triple: Triple{Subject: "file:/y", Predicate: "exists", Object: "true"}},
		},
	}
	g.Place("e1", e)

	got := g.BetsAt(Triple{Subject: "file:/x", Predicate: "exists", Object: "false"})
	if len(got) != 1 || got[0].Expect != &e.Expects[0] {
		t.Fatalf("address index returned %+v, want only file:/x exists prediction", got)
	}
	owned := g.ExpressionsInSpan(e.SpanID)
	if len(owned) != 1 || owned[0] != e {
		t.Fatalf("span ownership index returned %+v, want expression e1", owned)
	}
}

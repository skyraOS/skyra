package dnav3

import "testing"

// seedBeing builds one directly indexed historical span with a measured causal
// continuation. No class object participates.
func seedBeing(t *testing.T) (*Being, Triple) {
	t.Helper()
	b := NewBeing(nil, nil)
	witnessed := Triple{
		Mode: ModeExact, Subject: "thing:bottle", Predicate: "taken-from", Object: "fridge:kitchen",
	}
	b.ObserveImpulse([]Triple{
		{Mode: ModeExact, Subject: "being:michael", Predicate: "requested", Object: "thing:bottle"},
		witnessed,
	}, nil)
	owner := b.CurrentSpanID()
	e := &Expression{
		EpisodeID: b.EpisodeID(),
		SpanID:    owner,
		Acts:      []Act{{ID: "a1", Target: "fridge:kitchen", Primitive: "fetch"}},
		Expects: []Expect{{
			Of:      []string{"a1"},
			Triple:  witnessed,
			Within:  1,
			Armed:   true,
			ArmedAt: 1,
		}},
	}
	b.Expr.Place("e1", e)
	edge := b.Expr.Edge("e1", e.Expects[0])
	edge.Observe(TruthTrue, 1)
	edge.Observe(TruthTrue, 2)
	return b, witnessed
}

func TestObserveImpulseDerivesTargetFromActivatedCausalEdge(t *testing.T) {
	b, want := seedBeing(t)
	edgeN := b.Expr.Edge("e1", b.Expr.ExpressionsInSpan("episode:1/span:1")[0].Expects[0]).N

	res := b.ObserveImpulse([]Triple{{
		Mode: ModeExact, Subject: "thing:bottle", Predicate: "needed-by", Object: "being:michael",
	}}, nil)

	if res.Resolution.Novel {
		t.Fatalf("measured causal continuation flagged novel: %+v", res.Resolution)
	}
	if len(res.Target) != 1 || res.Target[0].Key() != want.Key() {
		t.Fatalf("activated target = %+v, want [%+v]", res.Target, want)
	}
	if res.Authored {
		t.Fatal("a causally derived target must not claim model authorship")
	}
	if got := b.Expr.Edge("e1", b.Expr.ExpressionsInSpan("episode:1/span:1")[0].Expects[0]).N; got != edgeN {
		t.Fatalf("activation changed causal truth statistics: N %d -> %d", edgeN, got)
	}
}

func TestObserveImpulseNovelYieldsNoTarget(t *testing.T) {
	b, _ := seedBeing(t)
	var res IngestResult
	for i := 0; i < 5; i++ {
		res = b.ObserveImpulse([]Triple{{
			Mode:      ModeExact,
			Subject:   "unrelated:" + string(rune('a'+i)),
			Predicate: "unseen",
			Object:    "novel",
		}}, nil)
	}

	if !res.Resolution.Novel {
		t.Fatalf("causally disconnected region should decay to novelty: %+v", res.Resolution)
	}
	if res.Target != nil {
		t.Fatalf("novel impulse must leave target authorship to framing, got %+v", res.Target)
	}
}

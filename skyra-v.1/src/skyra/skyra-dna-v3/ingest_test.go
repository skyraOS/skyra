package dnav3

import (
	"testing"
	"time"
)

// seedBeing builds a being whose observation log is the paper toy and whose expression
// graph holds one measured edge: "fetch from kitchen ⟹ bottle taken-from kitchen",
// P=1.0 N=2, evidenced in span S1 (class 0). No Realm, no Infer — ingest needs neither.
func seedBeing(t *testing.T) *Being {
	t.Helper()
	b := NewBeing(nil, nil)
	for _, o := range paperToy() {
		b.Obs.Write(o)
	}
	e := &Expression{
		Acts: []Act{{ID: "a1", Target: "fridge:kitchen", Primitive: "fetch"}},
		Expects: []Expect{{
			Of:     []string{"a1"},
			Triple: Triple{Mode: ModeExact, Subject: "thing:bottle", Predicate: "taken-from", Object: "fridge:kitchen"},
			Within: 1,
		}},
	}
	b.Expr.Place("e1", e)
	e.Expects[0].Resolution = TruthTrue
	e.Expects[0].ResolvedBy = []string{obsID(b.Obs.Log()[1])} // S1 evidence → class 0
	edge := b.Expr.Edge("e1", e.Expects[0])
	edge.Observe(TruthTrue, 1)
	edge.Observe(TruthTrue, 2)
	return b
}

// stampAfterToy makes the impulse land well clear of the toy's last moment (its own
// open span), advancing at the toy's own rhythm — second-apart stamps would drag the
// lower-quartile baseline down and shatter the historical spans.
func stampAfterToy(t *testing.T) func() {
	t.Helper()
	old := wallNow
	base := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC).UnixMilli()
	n := int64(0)
	wallNow = func() int64 { n++; return base + n*90_000 }
	return func() { wallNow = old }
}

func TestObserveImpulseDerivesPrecedentTarget(t *testing.T) {
	defer stampAfterToy(t)()
	b := seedBeing(t)
	// History for the ambiguous root: two prior commits, both to the bottle.
	b.ObserveInterpreted(
		Triple{Mode: ModeInterpret, Subject: "thing:my-drink", Predicate: PredicateResolvedTo, Object: "thing:bottle"},
		1.0, []string{"prior:1"})

	res := b.ObserveImpulse(
		[]Triple{{Mode: ModeInterpret, Subject: "thing:my-drink", Predicate: "taken-from", Object: "thing:the-fridge"}},
		nil)

	if res.Resolution.Novel {
		t.Fatalf("familiar impulse flagged novel: %+v", res.Resolution)
	}
	if res.Resolution.Primary != 0 {
		t.Fatalf("primary class = %d, want 0", res.Resolution.Primary)
	}
	// The target is READ off the class's measured completion, not authored.
	if len(res.Target) != 1 {
		t.Fatalf("want exactly the measured completion as target, got %+v", res.Target)
	}
	want := Triple{Subject: "thing:bottle", Predicate: "taken-from", Object: "fridge:kitchen"}
	if res.Target[0].Key() != want.Key() {
		t.Fatalf("target = %+v, want %+v", res.Target[0], want)
	}
	if res.Authored {
		t.Fatal("a precedent-read target must not claim authorship")
	}
	// The reading was committed and now feeds the next impulse's priors.
	rx := BuildResolutionIndex(b.Obs.Log())
	if cs := rx["thing:my-drink"]; len(cs) == 0 || cs[0].ID != "thing:bottle" {
		t.Fatalf("commit should compound the resolution index, got %v", cs)
	}
}

func TestObserveImpulseNovelYieldsNoTarget(t *testing.T) {
	defer stampAfterToy(t)()
	b := seedBeing(t)
	// Force the stitch: water history says bottle; fridge history says garage only.
	b.ObserveInterpreted(
		Triple{Mode: ModeInterpret, Subject: "thing:my-water", Predicate: PredicateResolvedTo, Object: "thing:bottle"},
		1.0, []string{"prior:1"})
	b.ObserveInterpreted(
		Triple{Mode: ModeInterpret, Subject: "thing:the-fridge", Predicate: PredicateResolvedTo, Object: "fridge:garage"},
		1.0, []string{"prior:2"})

	res := b.ObserveImpulse(
		[]Triple{{Mode: ModeInterpret, Subject: "thing:my-water", Predicate: "put-into", Object: "thing:the-fridge"}},
		nil)

	if !res.Resolution.Novel {
		t.Fatalf("bottle+garage has no joint precedent — must flag novel: %+v", res.Resolution)
	}
	if res.Target != nil {
		t.Fatalf("a novel stitch must not derive a target (the framing call authors it), got %+v", res.Target)
	}
}

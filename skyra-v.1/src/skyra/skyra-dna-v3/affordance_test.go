package dnav3

import "testing"

// Mode is derived from the WORLD'S INSTRUMENTS, never declared by the model.
//
// Mode.Weight() is the grade multiplier (exact ×1.0, interpret ×0.3). If the model could pick its
// own mode it would be picking its own stakes — and a badly-calibrated model would mark everything
// `interpret` to cap its losses, DAMPENING THE VERY SIGNAL THIS SYSTEM EXISTS TO MEASURE.
func TestAffordancesStampMode(t *testing.T) {
	a := DefaultAffordances()

	cases := []struct {
		t    Triple
		want Mode
		why  string
	}{
		{Triple{Subject: "file:/x", Predicate: "exists"}, ModeExact, "the world has a perceive for it"},
		{Triple{Subject: "build:/src", Predicate: "status"}, ModeExact, "exit codes are measurable"},
		{Triple{Subject: "being:michael", Predicate: "state"}, ModeInterpret, "there is no syscall for happiness"},
		{Triple{Subject: "skyra", Predicate: "helping"}, ModeInterpret, "no instrument exists"},
		{Triple{Subject: "file:/x", Predicate: "vibe"}, ModeInterpret, "known namespace, unmeasurable predicate"},
	}
	for _, c := range cases {
		if got := a.Mode(c.t); got != c.want {
			t.Errorf("%s %s → %q, want %q (%s)", c.t.Subject, c.t.Predicate, got, c.want, c.why)
		}
	}
}

// A model claiming `exact` on something unmeasurable gets DOWNGRADED, not believed. It cannot buy
// itself full fidelity on a pure inference.
func TestModelCannotClaimExactOnUnmeasurable(t *testing.T) {
	a := DefaultAffordances()
	e := &Expression{
		Acts: []Act{{ID: "a1", Target: "being:michael", Primitive: "engage"}},
		Expects: []Expect{
			// the model tries to mark an inference as exact — and it does not matter what it says
			{Of: []string{"a1"}, Within: 3, Triple: Triple{
				Mode: ModeExact, Confidence: 0.9,
				Subject: "being:michael", Predicate: "state", Object: "happy",
			}},
			{Of: []string{"a1"}, Within: 1, Triple: Triple{
				Mode: ModeInterpret, Confidence: 0.9, // and hedges on something it CAN be held to
				Subject: "file:/x", Predicate: "exists", Object: "true",
			}},
		},
	}
	a.Stamp(e)

	if e.Expects[0].Mode != ModeInterpret {
		t.Error("the model claimed `exact` on michael's mood and was BELIEVED — it just bought " +
			"1.0 fidelity on an inference")
	}
	if e.Expects[1].Mode != ModeExact {
		t.Error("the model marked a measurable claim `interpret` and got away with it — it just " +
			"cut its own losses to 0.3x on a bet the world can actually settle")
	}
}

// A target is STANDING iff none of its triples have an exact channel. Nobody labels it — the
// world's instruments decide what kind of goal it is.
func TestStandingTargetIsDerived(t *testing.T) {
	a := DefaultAffordances()

	standing := []Triple{{Subject: "skyra", Predicate: "helping", Object: "user"}}
	if !a.Standing(standing) {
		t.Error("`skyra helping user` has no perceive — it is a COMPASS, not a bet")
	}

	concrete := []Triple{{Subject: "file:/x", Predicate: "exists", Object: "true"}}
	if a.Standing(concrete) {
		t.Error("`file exists` has a perceive — it is a BET and must be gradeable at full fidelity")
	}

	// one measurable triple is enough to make a target concrete: there is something to be held to.
	mixed := []Triple{
		{Subject: "being:michael", Predicate: "state", Object: "happy"},
		{Subject: "file:/report.txt", Predicate: "exists", Object: "true"},
	}
	if a.Standing(mixed) {
		t.Error("a target with ANY measurable triple is concrete — that triple is what grades it")
	}
}

package dnav3

import (
	"fmt"
	"math/rand"
	"testing"
)

// Adversarial coverage of the ACTIVATION-scoped grading path. The stress oracle in
// stress_test.go validates the strict causal-span branch (it never populates the
// field). These drive the real activation machinery — recruitment, propagation, and
// decay — and attack the widened eligibility rule
//
//	eligible = armed ∧ unresolved ∧ same-episode ∧ (owner-span ∨ SpanActive) ∧ temporal
//
// which is the surface the class removal introduced.

// placeSpanBet stakes one unresolved bet at (subject,"state",object) owned by span.
// The address is set BEFORE Place so the reverse index keys it correctly — mutating a
// bet's address after Place would desync BetsAt and is a test bug, not a scenario.
func placeSpanBet(b *Being, id string, span SpanID, subject, object string) *modelBet {
	e := &Expression{
		EpisodeID: b.EpisodeID(),
		SpanID:    span,
		Acts:      []Act{{ID: "a0", Primitive: "stress", Target: "target:" + id}},
		Expects: []Expect{{
			Of:     []string{"a0"},
			Within: 1,
			Triple: Triple{Mode: ModeExact, Confidence: 0.75, Subject: subject, Predicate: "state", Object: object},
		}},
	}
	b.Expr.Place(id, e)
	return &modelBet{
		expr:       e,
		expect:     &e.Expects[0],
		episode:    e.EpisodeID,
		span:       span,
		acts:       []string{"a0"},
		completed:  map[string]bool{},
		resolution: TruthUnknown,
	}
}

// seedRecruitableSpan builds a prior occurrence discoverable by the span index (it
// carries an observation at the address) with one armed, unresolved bet there.
func seedRecruitableSpan(b *Being, id, subject, object string) *modelBet {
	span := b.startExternalSpan()
	b.appendObservation(
		Triple{Mode: ModeExact, Subject: subject, Predicate: "state", Object: "seen"},
		ProvenanceRaw, 0, nil, span)
	m := placeSpanBet(b, id, span, subject, object)
	completeModelAct(b, m, "a0") // arm it
	return m
}

// TestActivationDifferentialOracle drives observations through the REAL activation
// field (bump + causal propagation + periodic decay) and compares the grader against
// a reference model that reads the live SpanActive verdict. A mismatch means grade()
// mis-composes the activation-scoped eligibility rule.
func TestActivationDifferentialOracle(t *testing.T) {
	for _, seed := range []int64{1, 7, 41, 2026} {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			b := NewBeing(stressRealm{}, nil)

			var bets []*modelBet
			var spans []SpanID
			for i := 0; i < 12; i++ {
				span := b.startExternalSpan()
				spans = append(spans, span)
				m := placeSpanBet(b, fmt.Sprintf("adv-e%d", i), span,
					fmt.Sprintf("device:%d", r.Intn(4)), []string{"on", "off"}[r.Intn(2)])
				bets = append(bets, m)
			}

			for event := 0; event < 600; event++ {
				b.tick += r.Intn(2)
				switch {
				case r.Intn(100) < 30:
					m := bets[r.Intn(len(bets))]
					completeModelAct(b, m, m.acts[r.Intn(len(m.acts))])
				case r.Intn(100) < 15:
					b.Activation.BeginImpulse() // force decay: spans go cold mid-run
				default:
					span := spans[r.Intn(len(spans))]
					obj := []string{"on", "off"}[r.Intn(2)]
					p := ProvenanceRaw
					if r.Intn(9) == 0 {
						p = ProvenanceMark
					} else if r.Intn(5) == 0 {
						p = ProvenanceInterpreted
					}
					o := b.appendObservation(
						Triple{Mode: ModeExact, Subject: fmt.Sprintf("device:%d", r.Intn(4)), Predicate: "state", Object: obj},
						p, 0, nil, span)
					b.activateObservation(o) // REAL recruitment + propagation
					b.grade(o)               // production grader
					applyActivationModel(b, bets, o)
				}
				assertMatchesModel(t, b, bets)
			}
		})
	}
}

// applyActivationModel mirrors grade() eligibility, reading the live activation field
// for the (owner ∨ active) clause — the one input it does not independently derive.
func applyActivationModel(b *Being, bets []*modelBet, o Observation) {
	if o.Provenance == ProvenanceMark {
		return
	}
	for _, m := range bets {
		if m.resolution != TruthUnknown || !m.armed {
			continue
		}
		if m.episode != o.EpisodeID {
			continue
		}
		if m.span != o.SpanID && !b.Activation.SpanActive(m.span) {
			continue
		}
		if o.At < m.armedAt {
			continue
		}
		if sp(m.expect.Triple) != sp(o.Triple) {
			continue
		}
		m.resolution = TruthFalse
		if m.expect.Object == o.Object {
			m.resolution = TruthTrue
		}
		m.resolvedBy = obsID(o)
	}
}

// TestOverRecruitmentSettlesExactlyActiveSpans is the crosstalk attack. Many prior
// spans contend one address; a fresh impulse recruits some of them. The grader must
// settle a bet IFF its span cleared the activation floor — never all of them, and a
// promiscuous (large) contention must recruit few or none (the fan guard).
func TestOverRecruitmentSettlesExactlyActiveSpans(t *testing.T) {
	run := func(t *testing.T, k int) (resolved, active int) {
		b := NewBeing(stressRealm{}, nil)
		priors := make([]*modelBet, k)
		for i := 0; i < k; i++ {
			priors[i] = seedRecruitableSpan(b, fmt.Sprintf("p%d", i), "thing:x", "ok")
		}
		b.ObserveImpulse([]Triple{{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"}}, nil)

		for _, m := range priors {
			isActive := b.Activation.SpanActive(m.span)
			isResolved := m.expr.Expects[0].Resolution != TruthUnknown
			if isActive {
				active++
			}
			if isResolved {
				resolved++
			}
			if isResolved != isActive {
				t.Fatalf("k=%d span %s: resolved=%v active=%v — grade ignored the activation floor",
					k, m.span, isResolved, isActive)
			}
		}
		return resolved, active
	}

	t.Run("tight_contention_recruits_all", func(t *testing.T) {
		resolved, active := run(t, 4)
		if resolved != 4 || active != 4 {
			t.Fatalf("tight contention: resolved=%d active=%d, want 4/4 (all recruited)", resolved, active)
		}
	})
	t.Run("promiscuous_contention_is_contained", func(t *testing.T) {
		resolved, _ := run(t, 60)
		if resolved == 60 {
			t.Fatal("promiscuous address settled every contending bet — fan guard did not contain recruitment")
		}
	})
}

// TestColdSpanStaysInertAfterDecay is invariant 7 under the real decay curve: a span
// active enough to grade goes cold after enough impulses, and its still-open armed bet
// must then be invisible — proving the ONLY thing gating it is the activation floor.
func TestColdSpanStaysInertAfterDecay(t *testing.T) {
	b := NewBeing(stressRealm{}, nil)
	m := seedRecruitableSpan(b, "cold", "thing:x", "ok")
	fresh := SpanID("episode:1/probe")

	b.Activation.BumpSpan(m.span, 1)
	if !b.Activation.SpanActive(m.span) {
		t.Fatal("precondition: span should be active after a full bump")
	}
	warm := b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"}, ProvenanceRaw, 0, nil, fresh)
	b.grade(warm)
	if m.expr.Expects[0].Resolution != TruthTrue {
		t.Fatal("an active-span bet at a matching address should settle")
	}

	// A second, still-open armed bet in the same span, to be graded after decay.
	m2 := placeSpanBet(b, "cold2", m.span, "thing:x", "ok")
	completeModelAct(b, m2, "a0")

	for i := 0; i < 6; i++ {
		b.Activation.BeginImpulse() // λ=0.5 from 1.0 crosses the 0.05 floor by the 5th beat
	}
	if b.Activation.SpanActive(m.span) {
		t.Fatalf("span still active after decay: %.4f", b.Activation.Spans[m.span])
	}

	cold := b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"}, ProvenanceRaw, 0, nil, fresh)
	b.grade(cold)
	if m2.expr.Expects[0].Resolution != TruthUnknown {
		t.Fatal("a decayed-cold span's bet was settled — the dumpster is not inert (invariant 7 breached)")
	}

	b.Activation.BumpSpan(m.span, 1) // reactivate: proof the floor is the only difference
	reheat := b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"}, ProvenanceRaw, 0, nil, fresh)
	b.grade(reheat)
	if m2.expr.Expects[0].Resolution != TruthTrue {
		t.Fatal("reactivating the owner span did not restore grading eligibility")
	}
}

// TestActiveSpanStillNeedsArmAndTemporalOrder — activation is necessary, not
// sufficient. A recruited-active span whose bet is unarmed, or armed after the
// observation, must not settle (armed/active separation; invariant 6 under activation).
func TestActiveSpanStillNeedsArmAndTemporalOrder(t *testing.T) {
	b := NewBeing(stressRealm{}, nil)
	span := b.startExternalSpan()
	b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "seen"}, ProvenanceRaw, 0, nil, span)

	unarmed := placeSpanBet(b, "unarmed", span, "thing:x", "ok") // never completed → never armed
	late := placeSpanBet(b, "late", span, "thing:x", "ok")
	late.expr.Expects[0].Arm(50) // armed far in the future

	b.Activation.BumpSpan(span, 1) // span is unmistakably active
	fresh := SpanID("episode:1/probe")

	o := b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"}, ProvenanceRaw, 0, nil, fresh)
	b.grade(o)

	if unarmed.expr.Expects[0].Resolution != TruthUnknown {
		t.Fatal("an unarmed bet in an active span settled — activation granted causal credit")
	}
	if late.expr.Expects[0].Resolution != TruthUnknown {
		t.Fatal("an observation before the arm settled the bet in an active span (invariant 6)")
	}
}

// TestCrossEpisodeBlockedEvenWhenActive — the episode guard beats activation. A bet
// from a prior world stays ungradeable even if its span id is forced active.
func TestCrossEpisodeBlockedEvenWhenActive(t *testing.T) {
	b := NewBeing(stressRealm{}, nil)
	oldSpan := b.startExternalSpan()
	b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "seen"}, ProvenanceRaw, 0, nil, oldSpan)
	old := placeSpanBet(b, "old", oldSpan, "thing:x", "ok")
	completeModelAct(b, old, "a0")

	b.NewEpisode(stressRealm{})
	b.Activation.BumpSpan(oldSpan, 1) // force the old span active in the new field

	o := b.appendObservation(Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"}, ProvenanceRaw, 0, nil, SpanID("episode:2/probe"))
	b.grade(o)
	if old.expr.Expects[0].Resolution != TruthUnknown {
		t.Fatal("a new-world observation settled a prediction born in the previous episode, despite an active span")
	}
}

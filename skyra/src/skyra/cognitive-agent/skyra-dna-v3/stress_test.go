package dnav3

import (
	"fmt"
	"math/rand"
	"testing"
)

type stressRealm struct{}

func (stressRealm) Do(Act) Outcome { return Outcome{Status: "succeeded"} }
func (stressRealm) Perceive(t Triple) Outcome {
	return Outcome{Status: "succeeded", Effects: []Triple{t}}
}

type modelBet struct {
	expr       *Expression
	expect     *Expect
	episode    EpisodeID
	span       SpanID
	acts       []string
	completed  map[string]bool
	armed      bool
	armedAt    int
	resolution Truth
	resolvedBy string
}

func placeModelBet(b *Being, id string, span SpanID, address int, object string, actCount int) *modelBet {
	acts := make([]Act, actCount)
	of := make([]string, actCount)
	for i := range acts {
		of[i] = fmt.Sprintf("a%d", i)
		acts[i] = Act{ID: of[i], Primitive: "stress", Target: fmt.Sprintf("target:%s:%d", id, i)}
	}
	e := &Expression{
		EpisodeID: b.EpisodeID(),
		SpanID:    span,
		Acts:      acts,
		Expects: []Expect{{
			Of:     append([]string(nil), of...),
			Within: 3,
			Triple: Triple{
				Mode:       ModeExact,
				Confidence: 0.75,
				Subject:    fmt.Sprintf("device:%d", address),
				Predicate:  "state",
				Object:     object,
			},
		}},
	}
	b.Expr.Place(id, e)
	return &modelBet{
		expr:       e,
		expect:     &e.Expects[0],
		episode:    e.EpisodeID,
		span:       span,
		acts:       of,
		completed:  map[string]bool{},
		resolution: TruthUnknown,
	}
}

func completeModelAct(b *Being, m *modelBet, act string) {
	m.expr.Complete(act, b.tick)
	m.completed[act] = true
	if !m.armed {
		ready := true
		for _, id := range m.acts {
			if !m.completed[id] {
				ready = false
				break
			}
		}
		if ready {
			m.armed = true
			m.armedAt = b.tick
		}
	}
}

func observeModel(b *Being, bets []*modelBet, t Triple, p Provenance, span SpanID) Observation {
	o := b.appendObservation(t, p, 0, nil, span)
	b.grade(o)
	if p != ProvenanceMark {
		for _, m := range bets {
			if m.resolution != TruthUnknown || !m.armed {
				continue
			}
			if m.episode != o.EpisodeID || m.span != o.SpanID || o.At < m.armedAt {
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
	return o
}

func assertMatchesModel(t testing.TB, b *Being, bets []*modelBet) {
	t.Helper()
	resolved, correct := 0, 0
	for i, m := range bets {
		if m.expect.Armed != m.armed {
			t.Fatalf("bet %d armed=%v, model=%v", i, m.expect.Armed, m.armed)
		}
		if m.armed && (m.expect.ArmedAt != m.armedAt || m.expect.MaturesAt != m.armedAt+m.expect.Within) {
			t.Fatalf("bet %d timing armedAt=%d/maturesAt=%d, model=%d/%d",
				i, m.expect.ArmedAt, m.expect.MaturesAt, m.armedAt, m.armedAt+m.expect.Within)
		}
		if m.expect.Resolution != m.resolution {
			t.Fatalf("bet %d resolution=%v, model=%v (episode=%s span=%s address=%q)",
				i, m.expect.Resolution, m.resolution, m.episode, m.span, sp(m.expect.Triple))
		}
		if m.resolution == TruthUnknown {
			if len(m.expect.ResolvedBy) != 0 {
				t.Fatalf("bet %d unresolved with evidence %v", i, m.expect.ResolvedBy)
			}
			continue
		}
		resolved++
		if m.resolution == TruthTrue {
			correct++
		}
		if len(m.expect.ResolvedBy) != 1 || m.expect.ResolvedBy[0] != m.resolvedBy {
			t.Fatalf("bet %d evidence=%v, model=[%s]", i, m.expect.ResolvedBy, m.resolvedBy)
		}
	}
	if b.Metrics.Resolved != resolved || b.Metrics.Correct != correct {
		t.Fatalf("metrics resolved/correct=%d/%d, model=%d/%d",
			b.Metrics.Resolved, b.Metrics.Correct, resolved, correct)
	}
}

// This is a reference-model test, not a bag of probabilistic assertions. Every event is
// applied to a deliberately tiny eligibility model and the real grader, then all bets and
// aggregate metrics are compared. A failing seed is fully reproducible.
func TestStressGraderMatchesReferenceModel(t *testing.T) {
	for _, seed := range []int64{1, 7, 41, 99, 20260720} {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			b := NewBeing(stressRealm{}, nil)
			var bets []*modelBet
			var spans []SpanID
			seenObservationIDs := map[ObservationID]bool{}
			nextID := 0

			for episode := 0; episode < 3; episode++ {
				if episode > 0 {
					b.NewEpisode(stressRealm{})
				}
				for i := 0; i < 40; i++ {
					span := b.startExternalSpan()
					spans = append(spans, span)
					for j := 0; j < 3; j++ {
						object := "off"
						if r.Intn(2) == 0 {
							object = "on"
						}
						bets = append(bets, placeModelBet(
							b,
							fmt.Sprintf("stress-e%d", nextID),
							span,
							r.Intn(12),
							object,
							1+r.Intn(3),
						))
						nextID++
					}
				}

				for event := 0; event < 1800; event++ {
					b.tick += r.Intn(2)
					if r.Intn(100) < 38 {
						m := bets[r.Intn(len(bets))]
						completeModelAct(b, m, m.acts[r.Intn(len(m.acts))])
					} else {
						span := spans[r.Intn(len(spans))]
						object := "off"
						if r.Intn(2) == 0 {
							object = "on"
						}
						p := ProvenanceRaw
						switch r.Intn(10) {
						case 0:
							p = ProvenanceMark
						case 1, 2:
							p = ProvenanceInterpreted
						}
						o := observeModel(b, bets, Triple{
							Mode:      ModeExact,
							Subject:   fmt.Sprintf("device:%d", r.Intn(12)),
							Predicate: "state",
							Object:    object,
						}, p, span)
						if seenObservationIDs[o.ID] {
							t.Fatalf("seed %d reused observation ID %q", seed, o.ID)
						}
						seenObservationIDs[o.ID] = true
					}
					assertMatchesModel(t, b, bets)
				}
			}

			// Indexes must remain exact after the randomized history.
			for _, m := range bets {
				found := false
				for _, e := range b.Expr.ExpressionsInSpan(m.span) {
					found = found || e == m.expr
				}
				if !found {
					t.Fatalf("expression %s absent from owner span %s", m.expr.ID, m.span)
				}
				found = false
				for _, candidate := range b.Expr.BetsAt(m.expect.Triple) {
					found = found || candidate.Expect == m.expect
				}
				if !found {
					t.Fatalf("prediction %s absent from address index", m.expr.ID)
				}
			}
		})
	}
}

func TestStressDelayedSettlementReactivatesEveryOwnerSpan(t *testing.T) {
	b := NewBeing(stressRealm{}, nil)
	const spans = 200
	var bets []*modelBet
	for i := 0; i < spans; i++ {
		span := b.startExternalSpan()
		for j := 0; j < 5; j++ {
			m := placeModelBet(b, fmt.Sprintf("settle-e%d-%d", i, j), span, j%3, "on", 1)
			completeModelAct(b, m, "a0")
			bets = append(bets, m)
		}
	}
	// Move attention elsewhere and let all bets mature. Settlement must explicitly carry
	// each owner's span back through grading.
	unrelated := b.startExternalSpan()
	b.tick = 10
	b.Settle()

	seenEvidence := map[string]bool{}
	for i, m := range bets {
		if m.expect.Resolution != TruthTrue {
			t.Fatalf("bet %d in owner %s did not settle after delayed reactivation", i, m.span)
		}
		if len(m.expect.ResolvedBy) != 1 {
			t.Fatalf("bet %d evidence=%v, want exactly one observation", i, m.expect.ResolvedBy)
		}
		seenEvidence[m.expect.ResolvedBy[0]] = true
	}
	if b.CurrentSpanID() != unrelated {
		t.Fatalf("settlement changed current external focus from %s to %s", unrelated, b.CurrentSpanID())
	}
	if len(seenEvidence) != spans*3 {
		t.Fatalf("settlement evidence count=%d, want %d unique owner/address observations",
			len(seenEvidence), spans*3)
	}
}

func TestOldEpisodeBetCannotSettleInNewWorld(t *testing.T) {
	b := NewBeing(stressRealm{}, nil)
	oldSpan := b.startExternalSpan()
	old := placeModelBet(b, "old", oldSpan, 0, "on", 1)
	completeModelAct(b, old, "a0")
	b.tick = old.expect.MaturesAt

	b.NewEpisode(stressRealm{})
	observeModel(b, []*modelBet{old}, old.expect.Triple, ProvenanceRaw, oldSpan)
	if old.expect.Resolution != TruthUnknown {
		t.Fatal("a new-world observation settled a prediction born in the previous episode")
	}
	b.Settle()
	if old.expect.Resolution != TruthUnknown {
		t.Fatal("the settler graded an old-world prediction against the replacement Realm")
	}
}

func TestMarkCannotSatisfyGuardTargetOrOverwriteRawState(t *testing.T) {
	b := NewBeing(stressRealm{}, nil)
	claim := Triple{Subject: "device:door", Predicate: "state", Object: "open"}
	b.Observe(claim, ProvenanceMark)
	if truth, _ := b.Obs.Resolve(claim); truth != TruthUnknown {
		t.Fatalf("mark resolved an ordinary world claim to %v", truth)
	}
	if b.targetMet([]Triple{claim}) {
		t.Fatal("self-authored mark satisfied a target")
	}
	guarded := Expression{Acts: []Act{{ID: "a1", Requires: []Triple{claim}}}}
	if runnable := guarded.Runnable(func(t Triple) Truth {
		truth, _ := b.Obs.Resolve(t)
		return truth
	}); len(runnable) != 0 {
		t.Fatal("self-authored mark satisfied an action guard")
	}

	// Marks are search/memory aids, not world-state updates. A later mark at the same
	// address must not erase the last independently witnessed value either.
	b.Observe(claim, ProvenanceRaw)
	b.Observe(Triple{Subject: claim.Subject, Predicate: claim.Predicate, Object: "closed"}, ProvenanceMark)
	if truth, o := b.Obs.Resolve(claim); truth != TruthTrue || o.Provenance != ProvenanceRaw {
		t.Fatalf("mark overwrote raw state: truth=%v observation=%+v", truth, o)
	}
}

// The fuzzer varies arming order, evidence provenance, address collisions, duplicate
// observations, and old-span labels after an episode transition. The same reference model
// remains the oracle.
func FuzzGraderEligibility(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f.Add([]byte{255, 17, 17, 4, 91, 0, 3, 3, 3, 220})
	f.Fuzz(func(t *testing.T, events []byte) {
		if len(events) > 2048 {
			events = events[:2048]
		}
		b := NewBeing(stressRealm{}, nil)
		var bets []*modelBet
		var spans []SpanID
		for i := 0; i < 4; i++ {
			span := b.startExternalSpan()
			spans = append(spans, span)
			bets = append(bets, placeModelBet(b, fmt.Sprintf("fuzz-e%d", i), span, i%2, "on", 1+i%3))
		}
		transitioned := false
		for _, event := range events {
			b.tick++
			switch event % 7 {
			case 0, 1:
				m := bets[int(event)%len(bets)]
				completeModelAct(b, m, m.acts[(int(event)/len(bets))%len(m.acts)])
			case 2:
				if !transitioned {
					b.NewEpisode(stressRealm{})
					transitioned = true
				}
			default:
				p := ProvenanceRaw
				if event%11 == 0 {
					p = ProvenanceMark
				} else if event%5 == 0 {
					p = ProvenanceInterpreted
				}
				observeModel(b, bets, Triple{
					Mode:      ModeExact,
					Subject:   fmt.Sprintf("device:%d", int(event)%2),
					Predicate: "state",
					Object:    []string{"on", "off"}[(int(event)/2)%2],
				}, p, spans[(int(event)/3)%len(spans)])
			}
			assertMatchesModel(t, b, bets)
		}
	})
}

// This benchmark guards the purpose of the address index: unrelated open bets should not
// increase grading cost. Setup cost is excluded.
func BenchmarkGraderIgnoresUnrelatedBets(b *testing.B) {
	for _, noise := range []int{0, 1_000, 100_000} {
		b.Run(fmt.Sprintf("noise_%d", noise), func(b *testing.B) {
			being := NewBeing(stressRealm{}, nil)
			span := being.startExternalSpan()
			target := placeModelBet(being, "target", span, 0, "on", 1)
			completeModelAct(being, target, "a0")
			for i := 0; i < noise; i++ {
				placeModelBet(being, fmt.Sprintf("noise-%d", i), span, i+1, "on", 1)
			}
			o := Observation{
				ID:         "benchmark-observation",
				EpisodeID:  being.EpisodeID(),
				SpanID:     span,
				Triple:     target.expect.Triple,
				Provenance: ProvenanceRaw,
				At:         being.tick,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				target.expect.Resolution = TruthUnknown
				target.expect.ResolvedBy = nil
				being.grade(o)
			}
		})
	}
}

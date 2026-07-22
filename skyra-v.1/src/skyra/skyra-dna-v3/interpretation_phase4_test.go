package dnav3

import (
	"math"
	"strings"
	"testing"
)

var (
	toy9Traversal = Triple{Subject: "message:/reply", Predicate: "topic", Object: "traversal"}
	toy9Skyra     = Triple{Subject: "message:/reply", Predicate: "context", Object: "skyra"}
	toy9School    = Triple{Subject: "message:/reply", Predicate: "context", Object: "school"}
)

type toy9Realm struct{}

func (toy9Realm) Do(Act) Outcome { return Outcome{Status: "succeeded"} }
func (toy9Realm) Perceive(t Triple) Outcome {
	switch t.Predicate {
	case "topic":
		return Outcome{Status: "succeeded", Effects: []Triple{toy9Traversal}}
	case "context":
		return Outcome{Status: "succeeded", Effects: []Triple{toy9School}}
	default:
		return Outcome{Status: "unavailable"}
	}
}

func toy9Base(t *testing.T) (*Being, *Expression, []ObservationID) {
	t.Helper()
	b := NewBeing(nil, nil)
	b.Aff = NewAffordances(map[string][]string{
		"message": {"topic", "context", "status"},
	})
	b.ObserveImpulse([]Triple{{
		Subject: "message:/request", Predicate: "status", Object: "vague",
	}}, nil)
	e := &Expression{
		EpisodeID: b.EpisodeID(), SpanID: b.CurrentSpanID(),
		Acts: []Act{{ID: "a1", Primitive: "reply", Target: "message:/reply"}},
		Expects: []Expect{{
			Of: []string{"a1"}, Within: 1,
			Triple: Triple{Confidence: .9, Subject: toy9Traversal.Subject, Predicate: toy9Traversal.Predicate, Object: toy9Traversal.Object},
		}},
	}
	b.Aff.Stamp(e)
	b.Expr.Place("toy9-consequence", e)
	e.Complete("a1", b.tick)
	return b, e, append([]ObservationID(nil), b.currentBornFrom...)
}

func toy9Setup(t *testing.T) (*Being, *Expression, *InterpretationClaim) {
	t.Helper()
	b, e, evidence := toy9Base(t)
	claim, err := b.ProposeInterpretation(InterpretationProposal{
		Category:    "Michael-working-on-Skyra",
		Confidence:  .9,
		Context:     toy9Skyra,
		Within:      1,
		DerivedFrom: evidence,
		Consequence: ConsequenceRef{ExpressionID: e.ID, ExpectIndex: 0},
	})
	if err != nil {
		t.Fatalf("propose interpretation: %v", err)
	}
	return b, e, claim
}

func TestToy9LuckyHitSettlesTwoAddressesIndependentlyInOneImpulse(t *testing.T) {
	b, consequence, interpretation := toy9Setup(t)
	b.ObserveImpulse([]Triple{toy9Traversal, toy9School}, nil)

	if consequence.Expects[0].Resolution != TruthTrue {
		t.Fatalf("consequence resolution=%v, want true", consequence.Expects[0].Resolution)
	}
	if interpretation.Context.Resolution != TruthFalse {
		t.Fatalf("interpretation resolution=%v, want false", interpretation.Context.Resolution)
	}
	consequenceEdge := b.Expr.Edge(consequence.ID, consequence.Expects[0])
	interpretationEdge := b.Expr.InterpretationEdge(interpretation.ID)
	if consequenceEdge == interpretationEdge {
		t.Fatal("consequence and interpretation shared one edge instance")
	}
	if consequenceEdge.N != 1 || math.Abs(consequenceEdge.P-.75) > 1e-12 {
		t.Fatalf("consequence history=(P %.3f N %d), want (.75,1)", consequenceEdge.P, consequenceEdge.N)
	}
	if interpretationEdge.N != 1 || math.Abs(interpretationEdge.P-.25) > 1e-12 {
		t.Fatalf("interpretation history=(P %.3f N %d), want (.25,1)", interpretationEdge.P, interpretationEdge.N)
	}
	if math.Abs(consequence.Score-math.Log(.9/.5)) > 1e-12 {
		t.Fatalf("consequence score=%.6f, want %.6f", consequence.Score, math.Log(.9/.5))
	}
	if math.Abs(interpretation.Score-math.Log(.1/.5)) > 1e-12 {
		t.Fatalf("interpretation score=%.6f, want %.6f", interpretation.Score, math.Log(.1/.5))
	}
	if b.Metrics.Resolved != 1 || b.Metrics.Correct != 1 ||
		b.Metrics.InterpretationResolved != 1 || b.Metrics.InterpretationCorrect != 0 {
		t.Fatalf("two axes fused or miscounted in metrics: %+v", b.Metrics)
	}
	if len(consequence.Expects[0].ResolvedBy) != 1 || len(interpretation.Context.ResolvedBy) != 1 ||
		consequence.Expects[0].ResolvedBy[0] == interpretation.Context.ResolvedBy[0] {
		t.Fatalf("claims did not retain distinct settling observations: consequence=%v interpretation=%v",
			consequence.Expects[0].ResolvedBy, interpretation.Context.ResolvedBy)
	}
}

func TestToy9FullTwoBitOutcomeMatrix(t *testing.T) {
	tests := []struct {
		name            string
		topic, context  Triple
		wantConsequence Truth
		wantInterpret   Truth
	}{
		{"clean hit", toy9Traversal, toy9Skyra, TruthTrue, TruthTrue},
		{"right category wrong rule", Triple{Subject: toy9Traversal.Subject, Predicate: toy9Traversal.Predicate, Object: "sorting"}, toy9Skyra, TruthFalse, TruthTrue},
		{"lucky hit", toy9Traversal, toy9School, TruthTrue, TruthFalse},
		{"clean miss", Triple{Subject: toy9Traversal.Subject, Predicate: toy9Traversal.Predicate, Object: "sorting"}, toy9School, TruthFalse, TruthFalse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, consequence, interpretation := toy9Setup(t)
			b.ObserveImpulse([]Triple{tt.topic, tt.context}, nil)
			if consequence.Expects[0].Resolution != tt.wantConsequence || interpretation.Context.Resolution != tt.wantInterpret {
				t.Fatalf("got (%v,%v), want (%v,%v)", consequence.Expects[0].Resolution,
					interpretation.Context.Resolution, tt.wantConsequence, tt.wantInterpret)
			}
		})
	}
}

func TestToy9ObservationOrderAndPartialArrivalDoNotFuseClaims(t *testing.T) {
	for _, impulse := range [][]Triple{{toy9Traversal, toy9School}, {toy9School, toy9Traversal}} {
		b, consequence, interpretation := toy9Setup(t)
		b.ObserveImpulse(impulse, nil)
		if consequence.Expects[0].Resolution != TruthTrue || interpretation.Context.Resolution != TruthFalse {
			t.Fatalf("order %v changed two-bit result to (%v,%v)", impulse,
				consequence.Expects[0].Resolution, interpretation.Context.Resolution)
		}
	}

	tests := []struct {
		name                       string
		impulse                    []Triple
		wantConsequence, wantInter Truth
	}{
		{"consequence only", []Triple{toy9Traversal}, TruthTrue, TruthUnknown},
		{"interpretation only", []Triple{toy9School}, TruthUnknown, TruthFalse},
		{"unrelated", []Triple{{Subject: "message:/reply", Predicate: "tone", Object: "friendly"}}, TruthUnknown, TruthUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, consequence, interpretation := toy9Setup(t)
			b.ObserveImpulse(tt.impulse, nil)
			if consequence.Expects[0].Resolution != tt.wantConsequence || interpretation.Context.Resolution != tt.wantInter {
				t.Fatalf("got (%v,%v), want (%v,%v)", consequence.Expects[0].Resolution,
					interpretation.Context.Resolution, tt.wantConsequence, tt.wantInter)
			}
		})
	}
}

func TestToy9AdmissionEnforcesIndependentFalsifiabilityAndRawProvenance(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*InterpretationProposal)
		want   string
	}{
		{"empty category", func(p *InterpretationProposal) { p.Category = "" }, "category is empty"},
		{"no prediction", func(p *InterpretationProposal) { p.Context = Triple{} }, "not world-checkable"},
		{"no deadline", func(p *InterpretationProposal) { p.Within = 0 }, "no maturity"},
		{"no evidence", func(p *InterpretationProposal) { p.DerivedFrom = nil }, "no raw abductive evidence"},
		{"bad evidence", func(p *InterpretationProposal) { p.DerivedFrom = []ObservationID{"missing"} }, "missing, non-raw"},
		{"missing consequence", func(p *InterpretationProposal) { p.Consequence.ExpressionID = "missing" }, "no existing consequence"},
		{"same address", func(p *InterpretationProposal) {
			p.Context = Triple{Subject: toy9Traversal.Subject, Predicate: toy9Traversal.Predicate, Object: "skyra"}
		}, "distinct falsification addresses"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, e, evidence := toy9Base(t)
			proposal := InterpretationProposal{
				Category: "Michael-working-on-Skyra", Confidence: .9, Context: toy9Skyra, Within: 1,
				DerivedFrom: evidence, Consequence: ConsequenceRef{ExpressionID: e.ID, ExpectIndex: 0},
			}
			tt.mutate(&proposal)
			if _, err := b.ProposeInterpretation(proposal); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error=%v, want containing %q", err, tt.want)
			}
		})
	}

	b, e, evidence := toy9Base(t)
	interpreted := b.ObserveInterpreted(Triple{Subject: "reading:/", Predicate: "kind", Object: "guess"}, .8, []string{string(evidence[0])})
	_, err := b.ProposeInterpretation(InterpretationProposal{
		Category: "derived-from-a-guess", Confidence: .8, Context: toy9Skyra, Within: 1,
		DerivedFrom: []ObservationID{interpreted.ID}, Consequence: ConsequenceRef{ExpressionID: e.ID, ExpectIndex: 0},
	})
	if err == nil || !strings.Contains(err.Error(), "missing, non-raw") {
		t.Fatalf("interpreted evidence admitted as abductive ground: %v", err)
	}

	b2, e2, evidence2 := toy9Base(t)
	e2.Expects[0].Within = 0
	_, err = b2.ProposeInterpretation(InterpretationProposal{
		Category: "unfalsifiable-deduction", Confidence: .8, Context: toy9Skyra, Within: 1,
		DerivedFrom: evidence2, Consequence: ConsequenceRef{ExpressionID: e2.ID, ExpectIndex: 0},
	})
	if err == nil || !strings.Contains(err.Error(), "unfalsifiable consequence") {
		t.Fatalf("unfalsifiable consequence admitted: %v", err)
	}

	b3, e3, evidence3 := toy9Base(t)
	b3.Observe(toy9Traversal, ProvenanceRaw)
	_, err = b3.ProposeInterpretation(InterpretationProposal{
		Category: "posthoc-reason", Confidence: .8, Context: toy9Skyra, Within: 1,
		DerivedFrom: evidence3, Consequence: ConsequenceRef{ExpressionID: e3.ID, ExpectIndex: 0},
	})
	if err == nil || !strings.Contains(err.Error(), "retroactively license") {
		t.Fatalf("settled consequence admitted as posthoc license: %v", err)
	}
}

func TestToy9CannotGradeRetroactivelyOrAcrossEpisodes(t *testing.T) {
	b, e, evidence := toy9Base(t)
	// The outcome is already in the log before the interpretation claim exists.
	b.Observe(toy9School, ProvenanceRaw)
	claim, err := b.ProposeInterpretation(InterpretationProposal{
		Category: "Michael-working-on-Skyra", Confidence: .9, Context: toy9Skyra, Within: 1,
		DerivedFrom: evidence, Consequence: ConsequenceRef{ExpressionID: e.ID, ExpectIndex: 0},
	})
	if err != nil {
		t.Fatalf("post-observation proposal should be admissible prospectively: %v", err)
	}
	if claim.Context.Resolution != TruthUnknown || b.Expr.InterpretationEdge(claim.ID).N != 0 {
		t.Fatalf("claim scanned backward into prior outcome: %+v edge=%+v", claim, b.Expr.InterpretationEdge(claim.ID))
	}

	b.NewEpisode(nil)
	b.ObserveImpulse([]Triple{toy9Traversal, toy9School}, nil)
	if claim.Context.Resolution != TruthUnknown || b.Expr.InterpretationEdge(claim.ID).N != 0 {
		t.Fatalf("replacement episode settled old interpretation: %+v", claim)
	}
}

func TestToy9MarksCannotSettleAndInterpretedEvidenceIsCapped(t *testing.T) {
	b, consequence, interpretation := toy9Setup(t)
	b.Observe(toy9Traversal, ProvenanceMark)
	b.Observe(toy9Skyra, ProvenanceMark)
	if consequence.Expects[0].Resolution != TruthUnknown || interpretation.Context.Resolution != TruthUnknown {
		t.Fatalf("self-authored marks settled claims: (%v,%v)", consequence.Expects[0].Resolution, interpretation.Context.Resolution)
	}

	b.Observe(toy9Skyra, ProvenanceInterpreted)
	want := ModeInterpret.Weight() * math.Log(.9/.5)
	if interpretation.Context.Resolution != TruthTrue || math.Abs(interpretation.Score-want) > 1e-12 {
		t.Fatalf("interpreted context grade=(%v,%.6f), want (true,%.6f)", interpretation.Context.Resolution, interpretation.Score, want)
	}
	if consequence.Expects[0].Resolution != TruthUnknown {
		t.Fatalf("interpreted context leaked into consequence: %v", consequence.Expects[0].Resolution)
	}
}

func TestToy9CompetingInterpretationsSettleIndependentlyAtOneContextAddress(t *testing.T) {
	b, e, evidence := toy9Base(t)
	propose := func(category string, context Triple) *InterpretationClaim {
		t.Helper()
		claim, err := b.ProposeInterpretation(InterpretationProposal{
			Category: category, Confidence: .9, Context: context, Within: 1,
			DerivedFrom: evidence, Consequence: ConsequenceRef{ExpressionID: e.ID, ExpectIndex: 0},
		})
		if err != nil {
			t.Fatalf("propose %s: %v", category, err)
		}
		return claim
	}
	skyra := propose("Skyra-context", toy9Skyra)
	school := propose("School-context", toy9School)
	b.Observe(toy9School, ProvenanceRaw)
	if skyra.Context.Resolution != TruthFalse || school.Context.Resolution != TruthTrue {
		t.Fatalf("competing contexts got (%v,%v), want (false,true)", skyra.Context.Resolution, school.Context.Resolution)
	}

	// Shared predictions are not discriminating: both receive the same bit.
	b2, e2, evidence2 := toy9Base(t)
	propose2 := func(category string) *InterpretationClaim {
		claim, err := b2.ProposeInterpretation(InterpretationProposal{
			Category: category, Confidence: .9, Context: toy9Skyra, Within: 1,
			DerivedFrom: evidence2, Consequence: ConsequenceRef{ExpressionID: e2.ID, ExpectIndex: 0},
		})
		if err != nil {
			t.Fatalf("propose shared %s: %v", category, err)
		}
		return claim
	}
	left, right := propose2("left"), propose2("right")
	b2.Observe(toy9Skyra, ProvenanceRaw)
	if left.Context.Resolution != TruthTrue || right.Context.Resolution != TruthTrue {
		t.Fatalf("shared prediction manufactured discrimination: (%v,%v)", left.Context.Resolution, right.Context.Resolution)
	}
}

func TestToy9TargetedSettlementHonorsBothIndependentDeadlines(t *testing.T) {
	b, consequence, interpretation := toy9Setup(t)
	b.World = toy9Realm{}
	b.tick = 1 // both Within=1 predictions are now due
	b.Settle()
	if consequence.Expects[0].Resolution != TruthTrue || interpretation.Context.Resolution != TruthFalse {
		t.Fatalf("targeted settlement got (%v,%v), want (true,false)",
			consequence.Expects[0].Resolution, interpretation.Context.Resolution)
	}
}

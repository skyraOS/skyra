package dnav3

import (
	"math"
	"testing"
)

// The reference-model stress suite verifies WHICH bets resolve, but compares only
// Resolved/Correct counts — never grade MAGNITUDE. So invariant 12 (interpreted
// evidence never carries raw fidelity; span-owned doc §Fidelity) is untested by it:
// an interpreted observation that resolves an exact bet must fold in at ModeInterpret
// weight, not full authority. This pins that.
func gradeOnce(t *testing.T, prov Provenance) (Truth, float64) {
	t.Helper()
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	e := &Expression{
		EpisodeID: b.EpisodeID(), SpanID: span,
		Acts: []Act{{ID: "a1", Target: "thing:x", Primitive: "set"}},
		Expects: []Expect{{
			Of:     []string{"a1"},
			Triple: Triple{Mode: ModeExact, Confidence: 0.8, Subject: "thing:x", Predicate: "state", Object: "ok"},
			Within: 1,
		}},
	}
	b.Expr.Place("e1", e)
	e.Expects[0].Arm(0)

	o := b.appendObservation(
		Triple{Mode: ModeExact, Subject: "thing:x", Predicate: "state", Object: "ok"},
		prov, 0.9, nil, span)
	b.grade(o)
	return e.Expects[0].Resolution, b.Metrics.GradeSum
}

func TestInterpretedEvidenceCapsGrade(t *testing.T) {
	rawRes, rawGrade := gradeOnce(t, ProvenanceRaw)
	intRes, intGrade := gradeOnce(t, ProvenanceInterpreted)

	if rawRes != TruthTrue || intRes != TruthTrue {
		t.Fatalf("both should resolve TRUE (raw=%v interpreted=%v)", rawRes, intRes)
	}
	if rawGrade == 0 {
		t.Fatal("raw grade is zero — setup does not exercise the scoring path")
	}
	// Grade = (log pc − log pp) × mode.Weight(); everything but the weight is identical,
	// so the interpreted grade must be exactly ModeInterpret.Weight()× the raw one.
	want := rawGrade * ModeInterpret.Weight()
	if math.Abs(intGrade-want) > 1e-9 {
		t.Fatalf("interpreted grade = %.6f, want %.6f (raw %.6f × %.2f) — invariant 12 breached",
			intGrade, want, rawGrade, ModeInterpret.Weight())
	}
	if math.Abs(intGrade) >= math.Abs(rawGrade) {
		t.Fatalf("interpreted evidence did not cap the grade: |%.6f| >= |%.6f|", intGrade, rawGrade)
	}
}

// A mark may never grade a world-state claim (invariant 11 / Fidelity()==0).
func TestMarkNeverGrades(t *testing.T) {
	res, grade := gradeOnce(t, ProvenanceMark)
	if res != TruthUnknown {
		t.Fatalf("a mark settled a world-state bet: resolution=%v", res)
	}
	if grade != 0 {
		t.Fatalf("a mark moved the grade: %.6f", grade)
	}
}

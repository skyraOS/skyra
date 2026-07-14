package dnav3

import (
	"fmt"
	"strings"
	"testing"
)

// These tests drive the whole loop against a REAL filesystem. The model is scripted, so what is
// under test is the substrate: does the diff work, do the edges learn, does absence stay
// distinguishable from falsity, and does retrieval make planning cheaper.

func newTestBeing(t *testing.T, script func(present string) string) *Being {
	t.Helper()
	w := NewWorld(t.TempDir())
	return NewBeing(w, func(_, present string) (string, error) {
		return script(present), nil
	})
}

// The whole loop: emit → act on a real filesystem → observe what came back → grade.
func TestCycleGradesAgainstRealWorld(t *testing.T) {
	b := newTestBeing(t, func(string) string {
		return `
<framing>write the file</framing>
<act id="a1" target="file:/hello.txt" primitive="write">hello world</act>
<expect of="a1">  0.90  1  file:/hello.txt  exists  true  </expect>
`
	})

	target := []Triple{{Subject: "file:/hello.txt", Predicate: "exists", Object: "true"}}
	c, err := b.Run(target, 3)
	if err != nil {
		t.Fatal(err)
	}
	if c.Unresolved {
		t.Fatal("plan did not close")
	}
	if b.Metrics.Resolved != 1 {
		t.Fatalf("resolved %d bets, want 1", b.Metrics.Resolved)
	}
	if b.Metrics.Correct != 1 {
		t.Errorf("the file was really written; the bet should have been graded TRUE")
	}
	// the grade is log(0.90) − log(p̂). p̂ starts at 0, clamped to eps, so a first correct bold bet
	// earns a lot: it knew something the (empty) history did not.
	if b.Metrics.GradeSum <= 0 {
		t.Errorf("GradeSum = %.3f, want > 0 — a correct bet against an empty prior should PAY",
			b.Metrics.GradeSum)
	}
}

// A wrong prediction against a real world must be graded FALSE — not quietly ignored.
func TestWrongPredictionIsPunished(t *testing.T) {
	b := newTestBeing(t, func(string) string {
		return `
<framing>claim a file exists that I never write</framing>
<act id="a1" target="file:/real.txt" primitive="write">x</act>
<expect of="a1">  0.95  1  file:/real.txt   exists  true   </expect>
<expect of="a1">  0.95  1  file:/ghost.txt  exists  true   </expect>
`
	})
	_, err := b.Run([]Triple{{Subject: "file:/real.txt", Predicate: "exists", Object: "true"}}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if b.Metrics.Resolved != 2 {
		t.Fatalf("resolved %d, want 2 — the settler must go LOOK at the ghost", b.Metrics.Resolved)
	}
	if b.Metrics.Correct != 1 {
		t.Errorf("correct = %d, want 1 (real.txt yes, ghost.txt no)", b.Metrics.Correct)
	}
	// confident and wrong is punished hard, and it must outweigh the confident-and-right bet
	if b.Metrics.GradeSum >= 0 {
		t.Errorf("GradeSum = %.3f — a 0.95 bet that was FALSE must cost more than a 0.95 bet that was true",
			b.Metrics.GradeSum)
	}
}

// THE MOST IMPORTANT TEST IN THE PACKAGE.
//
// Absence is not falsity. If the being never looked, the bet must stay UNKNOWN and the edge must
// NOT move. Treating "I didn't check" as "it didn't happen" drags every edge toward zero, which is
// the single most destructive bug available in this design.
func TestAbsenceIsNotFalsity(t *testing.T) {
	b := newTestBeing(t, func(string) string { return "" })
	e := &Expression{
		Framing: "unlooked",
		Acts:    []Act{{ID: "a1", Target: "file:/x", Primitive: "write"}},
		Expects: []Expect{{
			Of:     []string{"a1"},
			Triple: Triple{Mode: ModeExact, Confidence: 0.9, Subject: "file:/never-looked", Predicate: "exists", Object: "true"},
			Within: 5,
		}},
	}
	b.Expr.Place("e1", e)

	// nothing observed, nothing matured
	truth, _ := b.Obs.Resolve(e.Expects[0].Triple)
	if truth != TruthUnknown {
		t.Fatalf("unobserved triple resolved to %v, want UNKNOWN", truth)
	}
	edge := b.Expr.Edge("e1", e.Expects[0])
	if edge.N != 0 {
		t.Fatalf("edge moved on an unresolved bet: N=%d", edge.N)
	}

	// TruthUnknown must be inert
	edge.Observe(TruthUnknown, 1)
	if edge.N != 0 || edge.P != 0 {
		t.Errorf("TruthUnknown moved the EMA: N=%d P=%.3f — this is the bug that poisons the graph", edge.N, edge.P)
	}
}

// The bet manufactures its own evidence: at maturity, the substrate goes and LOOKS.
func TestSettlerGoesAndLooks(t *testing.T) {
	b := newTestBeing(t, func(string) string { return "" })
	e := &Expression{
		Acts: []Act{{ID: "a1", Target: "file:/absent", Primitive: "write"}},
		Expects: []Expect{{
			Of:     []string{"a1"},
			Triple: Triple{Mode: ModeExact, Confidence: 0.8, Subject: "file:/absent", Predicate: "exists", Object: "true"},
			Within: 1,
		}},
	}
	b.Expr.Place("e1", e)
	e.Expects[0].Arm(0) // the act completed at tick 0; the bet is due at tick 1

	b.tick = 1
	b.Settle()

	if e.Expects[0].Resolution != TruthFalse {
		t.Fatalf("resolution = %v, want FALSE — the settler perceived and the file is not there",
			e.Expects[0].Resolution)
	}
	if len(e.Expects[0].ResolvedBy) == 0 {
		t.Error("a grade with no observation behind it is self-adjudicated (LAW 2)")
	}
}

// LAW 7: a mark is self-authored and may never grade a claim about the world.
func TestMarkCannotGrade(t *testing.T) {
	b := newTestBeing(t, func(string) string { return "" })
	e := &Expression{
		Acts: []Act{{ID: "a1", Primitive: "write"}},
		Expects: []Expect{{
			Of:     []string{"a1"},
			Triple: Triple{Mode: ModeExact, Confidence: 0.9, Subject: "file:/x", Predicate: "exists", Object: "true"},
			Within: 1,
		}},
	}
	b.Expr.Place("e1", e)

	b.Observe(Triple{Subject: "file:/x", Predicate: "exists", Object: "true"}, ProvenanceMark)

	if e.Expects[0].Resolution != TruthUnknown {
		t.Error("a MARK graded a bet — the being just laundered its own handwriting into evidence")
	}
	if b.Expr.Edge("e1", e.Expects[0]).N != 0 {
		t.Error("a mark moved an edge's EMA")
	}
}

// THE THESIS, in miniature: precedent makes planning cheaper.
//
// Same task, twice. The first time there is no precedent and the model needs two passes. The second
// time the index surfaces what worked, and one pass closes it.
func TestPassesToClosureFalls(t *testing.T) {
	call := 0
	b := newTestBeing(t, func(present string) string {
		call++
		hasPrecedent := strings.Contains(present, "PRECEDENT")
		if hasPrecedent {
			// the being has been here before: one pass, the whole plan
			return `
<framing>seen this before — write it and verify</framing>
<act id="a1" target="file:/out.txt" primitive="write">done</act>
<expect of="a1">  0.95  1  file:/out.txt  exists  true  </expect>`
		}
		// cold: emit only the verify step first, leaving a hole
		if call == 1 {
			return `
<framing>cold start — verify the output</framing>
<act id="a1" target="dir:/" primitive="exec" requires="file:/out.txt exists true">ls out.txt</act>
<expect of="a1">  0.80  1  dir:/  status  ok  </expect>`
		}
		return `
<framing>ground the hole: actually write the file</framing>
<act id="a2" target="file:/out.txt" primitive="write">done</act>
<expect of="a2">  0.95  1  file:/out.txt  exists  true  </expect>`
	})

	target := []Triple{{Subject: "file:/out.txt", Predicate: "exists", Object: "true"}}

	first, err := b.Run(target, 5)
	if err != nil {
		t.Fatal(err)
	}
	if first.Unresolved {
		t.Fatal("first cycle did not close")
	}

	second, err := b.Run(target, 5)
	if err != nil {
		t.Fatal(err)
	}
	if second.Unresolved {
		t.Fatal("second cycle did not close")
	}

	t.Logf("passes-to-closure: %v", b.Metrics.PassesToClosure)
	if second.Passes >= first.Passes {
		t.Errorf("passes did not fall: first=%d second=%d — retrieval is not making planning cheaper",
			first.Passes, second.Passes)
	}
	if second.Passes != 1 {
		t.Errorf("second cycle took %d passes, want 1 (the floor: the framing call)", second.Passes)
	}
}

// Beating the base rate is what pays. Reciting it earns nothing.
func TestScoreIsEdgeOverBaseRate(t *testing.T) {
	e := &Edge{P: 0.70, N: 10}

	parrot := e.Grade(0.70, TruthTrue, ModeExact)
	if parrot > 1e-9 || parrot < -1e-9 {
		t.Errorf("reciting the base rate earned %.4f, want 0 — the model added nothing", parrot)
	}

	insight := e.Grade(0.95, TruthTrue, ModeExact)
	if insight <= 0 {
		t.Errorf("a confident call that came true earned %.4f, want > 0", insight)
	}

	upset := e.Grade(0.20, TruthFalse, ModeExact)
	if upset <= insight {
		t.Errorf("calling an upset (%.3f) should pay more than agreeing with history (%.3f)", upset, insight)
	}

	overconfident := e.Grade(0.95, TruthFalse, ModeExact)
	if overconfident > -1.0 {
		t.Errorf("confident and wrong earned %.3f, want a large negative", overconfident)
	}
}

func TestReport(t *testing.T) {
	b := newTestBeing(t, func(string) string {
		return `
<framing>go</framing>
<act id="a1" target="file:/a.txt" primitive="write">x</act>
<expect of="a1">  0.9  1  file:/a.txt  exists  true  </expect>`
	})
	b.Run([]Triple{{Subject: "file:/a.txt", Predicate: "exists", Object: "true"}}, 3)
	fmt.Println(b.Report())
}

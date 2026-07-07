package dnav2

import (
	"sync"
	"testing"
	"time"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	"skyra-v1/src/skyra/world"
)

// stubDNA wires stub genes that just report a fixed condition per step — enough to
// exercise the ORCHESTRATION (routing + fan-out + join) with no inference. Each
// gene records that it ran on a shared, locked tally.
func stubDNA(t *testing.T, tally *sync.Map) *DNA {
	d := testDNA()
	report := func(step Step, cond string) {
		d.RegisterObserve(string(step), func(root *EphemeralRoot, origin skyra.Space, impulse any) {
			tally.Store(string(step), true)
			root.Condition = cond
		})
	}
	report(StepSensoryIntake, "external")
	report(StepOtherGrade, "held") // a normal turn: prior bet held → frame
	report(StepMidlevelFrame, "framed")
	report(StepOtherPredict, "predicted")
	report(StepSelfPredict, "predicted")
	report(StepMotorCollapse, "emitted")
	return d
}

func newOrch(d *DNA) (*Orchestrator, *Board) {
	board := NewBoard()
	kernel := relation.NewRelation()
	slices := map[string]skyra.Slice{
		"sensory":  world.NewSlice("skyra:sensory", "skyra"),
		"other":    world.NewSlice("skyra:other", "skyra"),
		"self":     world.NewSlice("skyra:self", "skyra"),
		"midlevel": world.NewSlice("skyra:midlevel", "skyra"),
		"motor":    world.NewSlice("skyra:motor", "skyra"),
	}
	o := NewOrchestrator(d, board, kernel, world.NewSlice("skyra:membrane", "skyra"), slices)
	return o, board
}

// The full happy path routes end to end: sensory:external → other:grade:held →
// midlevel:framed → fan-out {other:predict, self:predict} → join → motor. And
// motor fires exactly once (the atomic join under the fan-out race).
func TestOrchestratorHappyPathRoutes(t *testing.T) {
	var tally sync.Map
	d := stubDNA(t, &tally)

	// count motor runs to assert exactly-once
	var motorRuns int32
	var mu sync.Mutex
	d.RegisterObserve(string(StepMotorCollapse), func(root *EphemeralRoot, origin skyra.Space, impulse any) {
		tally.Store(string(StepMotorCollapse), true)
		mu.Lock()
		motorRuns++
		mu.Unlock()
		root.Condition = "emitted"
	})

	o, board := newOrch(d)
	done := make(chan string, 1)
	o.SetFlush(func(impulseID string, trace []Span) { done <- impulseID })

	o.Fire("X", "", skyra.Impulse{Primitive: "engage", Payload: "hi"})

	// the pipeline is async (each step re-invokes membrane); the flush hook fires
	// on completion.
	var flushed string
	select {
	case flushed = <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("impulse X did not complete (flush never fired)")
	}

	for _, step := range []Step{StepSensoryIntake, StepOtherGrade, StepMidlevelFrame, StepOtherPredict, StepSelfPredict, StepMotorCollapse} {
		if _, ran := tally.Load(string(step)); !ran {
			t.Errorf("step %s never ran", step)
		}
	}
	mu.Lock()
	runs := motorRuns
	mu.Unlock()
	if runs != 1 {
		t.Fatalf("motor ran %d times, want exactly 1 (join guard)", runs)
	}
	if flushed != "X" {
		t.Fatalf("flush impulseID = %q, want X", flushed)
	}
	if board.Row("X", "").Status() != "done" {
		t.Fatal("impulse X did not complete")
	}
}

// A total miss reframes (once), then proceeds. other:grade reports all-preds<0 the
// first time; the reframe re-runs midlevel and the rest of the pipeline.
func TestOrchestratorReframeRail(t *testing.T) {
	var tally sync.Map
	d := stubDNA(t, &tally)

	var reframes int32
	var mu sync.Mutex
	d.RegisterObserve(string(StepOtherGrade), func(root *EphemeralRoot, origin skyra.Space, impulse any) {
		root.Condition = "all-preds<0" // always a total miss
	})
	d.RegisterObserve(string(StepMidlevelFrame), func(root *EphemeralRoot, origin skyra.Space, impulse any) {
		mu.Lock()
		reframes++ // counts every midlevel run (the reframe target)
		mu.Unlock()
		root.Condition = "framed"
	})

	o, _ := newOrch(d)
	done := make(chan string, 1)
	o.SetFlush(func(impulseID string, trace []Span) { done <- impulseID })
	o.Fire("X", "", skyra.Impulse{Primitive: "engage", Payload: "surprise"})

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("impulse X did not complete")
	}

	// other:grade → all-preds<0 → reframe (allowed once) → midlevel → fan-out → …
	// midlevel runs exactly once (the single allowed reframe); a second reframe
	// would be suppressed by the rail.
	mu.Lock()
	n := reframes
	mu.Unlock()
	if n != 1 {
		t.Fatalf("midlevel ran %d times, want 1 (one-reframe rail)", n)
	}
}

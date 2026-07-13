package dnav2

import (
	"strings"
	"sync"
	"testing"
	"time"

	"skyra-v1/src/skyra"
)

// gradeMock is the turn-1 cold-start mock plus a parametrized adversarial judge, so
// a phase-2 two-turn run can drive other:grade down either fork (held / total miss).
func gradeMock(judge string) InferFunc {
	return func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		switch {
		case strings.Contains(system, "strict, impartial judge"):
			return judge, nil
		case strings.Contains(system, "evolving being"):
			return "<requery>assist[1] with[1→2] the {skyra project}[2]</requery>", nil
		case strings.Contains(system, "External Perspective"):
			return `<predictions><prediction p="0.7">asks about the kernel</prediction><prediction p="0.3">asks about setup</prediction></predictions>`, nil
		case strings.Contains(system, "You are Skyra's motor system"):
			return "<engage><michael>What part of the skyra project can I help with?</michael></engage>", nil
		case strings.Contains(system, "You are Skyra's inner voice"):
			return "<prediction>Ask michael what part of the skyra project he wants help with</prediction>", nil
		}
		return "", nil
	}
}

// fireTurn drives one external engage to completion and returns its board trace.
func fireTurn(t *testing.T, o *Orchestrator, id, prev, utterance string) []Span {
	t.Helper()
	done := make(chan []Span, 1)
	o.SetFlush(func(_ string, trace []Span) { done <- trace })
	o.Fire(id, prev, skyra.Impulse{
		Primitive: "engage",
		Source:    &skyra.ImpulseSource{SpaceID: "michael"},
		Payload:   utterance,
	})
	select {
	case tr := <-done:
		return tr
	case <-time.After(3 * time.Second):
		t.Fatal("turn did not complete")
		return nil
	}
}

// condOf pulls a step's reported condition out of a trace.
func condOf(trace []Span, step Step) string {
	for _, s := range trace {
		if s.Step == string(step) {
			return s.Condition
		}
	}
	return ""
}

// Turn 1 establishes an other-distribution; turn 2's evidence realizes one branch,
// so other:grade closes it as "held" and the prior bet is stamped graded.
func TestGradeHeldStampsPriorAndFrames(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(gradeMock("<match>1</match>"))
	var mu sync.Mutex
	var emitted []skyra.Impulse
	d.SetEmit(func(imp skyra.Impulse) { mu.Lock(); emitted = append(emitted, imp); mu.Unlock() })
	o, _ := newOrch(d)

	fireTurn(t, o, "X1", "", "can you help with the skyra project")
	t2 := fireTurn(t, o, "X2", "X1", "how does the kernel work")

	if got := condOf(t2, StepOtherGrade); got != "held" {
		t.Fatalf("other:grade condition = %q, want held", got)
	}
	if condOf(t2, StepMidlevelFrame) != "framed" {
		t.Fatal("held should route straight into midlevel:frame")
	}
	// the turn-1 distribution root must now be graded with the realized branch's mass.
	root := priorOtherRoot(t, d)
	p, _ := root.Content().(Prediction)
	if !p.Graded || p.Grade != 0.7 {
		t.Fatalf("prior root = {graded:%v grade:%.2f}, want {true 0.70}", p.Graded, p.Grade)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(emitted) != 2 {
		t.Fatalf("emitted %d actions across two turns, want 2", len(emitted))
	}
}

// Turn 2's evidence realizes NO branch — a total miss. other:grade reports
// "all-preds<0", which the orchestrator routes as a reframe (one-reframe rail), and
// the turn still frames and emits.
func TestGradeTotalMissReframes(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(gradeMock("<match>none</match>"))
	var mu sync.Mutex
	var emitted []skyra.Impulse
	d.SetEmit(func(imp skyra.Impulse) { mu.Lock(); emitted = append(emitted, imp); mu.Unlock() })
	o, board := newOrch(d)

	fireTurn(t, o, "X1", "", "can you help with the skyra project")
	t2 := fireTurn(t, o, "X2", "X1", "actually, let's talk about something unrelated")

	if got := condOf(t2, StepOtherGrade); got != "all-preds<0" {
		t.Fatalf("other:grade condition = %q, want all-preds<0", got)
	}
	if !board.Row("X2", "").Reframed() {
		t.Fatal("total miss should have tripped the reframe rail")
	}
	// the prior bet is stamped graded with zero mass (nothing realized).
	root := priorOtherRoot(t, d)
	p, _ := root.Content().(Prediction)
	if !p.Graded || p.Grade != 0 {
		t.Fatalf("prior root = {graded:%v grade:%.2f}, want {true 0.00}", p.Graded, p.Grade)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(emitted) != 2 {
		t.Fatalf("emitted %d actions across two turns, want 2 (reframe still acts)", len(emitted))
	}
}

// editMock requeries the same situation, so the episode retains its frame sequence.
func editMock() InferFunc {
	return func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		switch {
		case strings.Contains(system, "evolving being"): // frameOpSystem
			if strings.Contains(present, "ACTIVE MATERIAL/FORMAL:\n(none") {
				return "<requery>assist[1] with[1→2] the {skyra project}[2]</requery>", nil
			}
			return "<requery>continuing[1] the {skyra project}[2]</requery>", nil
		case strings.Contains(system, "strict, impartial judge"):
			return "<match>none</match>", nil
		case strings.Contains(system, "External Perspective"):
			return `<predictions><prediction p="0.6">asks more</prediction></predictions>`, nil
		case strings.Contains(system, "You are Skyra's motor system"):
			return "<engage><michael>ok</michael></engage>", nil
		case strings.Contains(system, "You are Skyra's inner voice"):
			return "<prediction>ask michael what next</prediction>", nil
		}
		return "", nil
	}
}

// EDIT continues the active episode: turn 2 sharpens the existing episode's frame instead
// of minting a new one — the compounding path (one growing episode, not two islands).
func TestMidlevelEditContinuesEpisode(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(editMock())
	d.SetEmit(func(skyra.Impulse) {})
	o, _ := newOrch(d)

	fireTurn(t, o, "X1", "", "help with the skyra project") // cold start → new
	n1 := len(d.Episodes().All())
	fireTurn(t, o, "X2", "X1", "more on the skyra project") // open episode exists → edit
	n2 := len(d.Episodes().All())

	if n1 != 1 {
		t.Fatalf("turn 1 (cold start) should create exactly 1 episode, got %d", n1)
	}
	if n2 != n1 {
		t.Fatalf("edit must NOT create a new episode: turn1=%d turn2=%d", n1, n2)
	}
	g := d.Episodes().Active()
	if g == nil {
		t.Fatal("no active episode")
	}
	if got := frameText(g); got != "continuing the skyra project" {
		t.Fatalf("edit should mint a fresh current frame, got %q", got)
	}
	// the sequence: the prior frame is RETAINED (decaying), not overwritten.
	frames := 0
	for _, c := range g.Children() {
		if mc, ok := c.(*MemoryCell); ok && mc.Kind() == KindFrame {
			frames++
		}
	}
	if frames != 2 {
		t.Fatalf("edit should append to the frame sequence: %d frames, want 2", frames)
	}
}

// After framing, the just-arrived interaction is consolidated onto the episode (so it
// lives on the episode the frame walks) and the entity points to the episode.
func TestInteractionConsolidatedOntoEpisode(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(gradeMock("<match>none</match>"))
	d.SetEmit(func(skyra.Impulse) {})
	o, _ := newOrch(d)

	fireTurn(t, o, "X1", "", "help with the skyra project")

	g := d.Episodes().Active()
	if g == nil {
		t.Fatal("no active episode")
	}
	frame := latestFrame(g)
	if frame == nil {
		t.Fatal("active episode has no frame")
	}
	hasInteraction := false
	for _, c := range frame.Children() {
		if mc, ok := c.(*MemoryCell); ok {
			if _, ok := mc.Content().(InteractionRecord); ok {
				hasInteraction = true
			}
		}
	}
	if !hasInteraction {
		t.Fatal("the frame should own the consolidated interaction (frame-owned spine)")
	}
	for _, e := range d.Nodes().Snapshot() {
		if e.ID() == "michael" {
			if _, ok := e.Children()[g.ID()]; !ok {
				t.Fatal("entity should point to the episode (entity→episode edge)")
			}
		}
	}
}

// execMock is the cold-start mock plus a parametrized execution judge, so a
// reafference turn can drive self:grade down either fork (exec-ok / exec=no).
func execMock(execJudge string) InferFunc {
	base := gradeMock("<match>none</match>")
	return func(system, present string) (string, error) {
		if strings.Contains(system, "strict, impartial judge of execution") {
			return execJudge, nil
		}
		return base(system, present)
	}
}

// fireReafference drives the world's Return reply through the reafference lane.
func fireReafference(t *testing.T, o *Orchestrator, id, prev, result string) []Span {
	t.Helper()
	done := make(chan []Span, 1)
	o.SetFlush(func(_ string, trace []Span) { done <- trace })
	o.Fire(id, prev, skyra.Impulse{
		Return:  true,
		Payload: result,
		Source:  &skyra.ImpulseSource{SpaceID: "world"},
	})
	select {
	case tr := <-done:
		return tr
	case <-time.After(3 * time.Second):
		t.Fatal("reafference did not complete")
		return nil
	}
}

// After turn 1 emits an action, the world's reply faithfully executes it — self:grade
// closes the proposal "exec-ok" and stamps the proposal graded with a positive score.
func TestSelfGradeExecOk(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(execMock("<grade>yes</grade>"))
	d.SetEmit(func(skyra.Impulse) {})
	o, board := newOrch(d)

	fireTurn(t, o, "X1", "", "can you help with the skyra project")
	r := fireReafference(t, o, "R1", "X1", "michael listened and replied")

	if got := condOf(r, StepSelfGrade); got != "exec-ok" {
		t.Fatalf("self:grade condition = %q, want exec-ok", got)
	}
	if board.Row("R1", "").Reframed() {
		t.Fatal("faithful execution should not reframe")
	}
	p := gradedSelfProposal(t, d)
	if !p.Graded || p.Grade != 1 {
		t.Fatalf("self proposal = {graded:%v grade:%.2f}, want {true 1.00}", p.Graded, p.Grade)
	}
}

// The world reports a hard failure (the intent could not be realized) — self:grade
// reports "exec=no", which trips the reframe rail.
func TestSelfGradeExecFailReframes(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(execMock("<grade>no</grade>"))
	d.SetEmit(func(skyra.Impulse) {})
	o, board := newOrch(d)

	fireTurn(t, o, "X1", "", "can you help with the skyra project")
	r := fireReafference(t, o, "R1", "X1", "blocked: could not reach the destination")

	if got := condOf(r, StepSelfGrade); got != "exec=no" {
		t.Fatalf("self:grade condition = %q, want exec=no", got)
	}
	if !board.Row("R1", "").Reframed() {
		t.Fatal("hard execution failure should have tripped the reframe rail")
	}
	p := gradedSelfProposal(t, d)
	if !p.Graded || p.Grade != 0 {
		t.Fatalf("self proposal = {graded:%v grade:%.2f}, want {true 0.00}", p.Graded, p.Grade)
	}
}

// gradedSelfProposal returns the graded self proposal — the realized one self:grade
// closed (the reafference cycle mints a fresh ungraded proposal, so "latest" won't do).
func gradedSelfProposal(t *testing.T, d *DNA) Prediction {
	t.Helper()
	for _, g := range d.Episodes().All() {
		for _, mc := range predictionsOf(g) {
			if p, ok := mc.Content().(Prediction); ok && p.Role == "self" && p.Graded {
				return p
			}
		}
	}
	t.Fatal("no graded self proposal found")
	return Prediction{}
}

// priorOtherRoot returns the turn-1 other-distribution root (the oldest one).
func priorOtherRoot(t *testing.T, d *DNA) *MemoryCell {
	t.Helper()
	var oldest *MemoryCell
	oldestTS := 1 << 30
	for _, g := range d.Episodes().All() {
		for _, mc := range predictionsOf(g) {
			if p, ok := mc.Content().(Prediction); ok && p.Role == "other" && len(mc.Children()) > 0 {
				if p.Timestamp < oldestTS {
					oldestTS, oldest = p.Timestamp, mc
				}
			}
		}
	}
	if oldest == nil {
		t.Fatal("no other-distribution root found")
	}
	return oldest
}

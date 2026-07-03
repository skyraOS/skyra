package tests

import (
	"strings"
	"testing"
	"time"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// gradingMock routes on system-prompt content. Order matters: the execution
// judge's prompt contains "impartial judge", so it is matched first.
func gradingMock(t *testing.T, judgeVerdict, execVerdict string, judgePresents *[]string) dna.InferFunc {
	return func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "judge of execution"):
			return execVerdict, nil
		case strings.Contains(system, "impartial judge"):
			if judgePresents != nil {
				*judgePresents = append(*judgePresents, present)
			}
			return judgeVerdict, nil
		case strings.Contains(system, "External Perspective"):
			return `<predictions>
  <prediction p="0.70">ask how the runtime is structured</prediction>
  <prediction p="0.30">ask about the kernel</prediction>
</predictions>`, nil
		case strings.Contains(system, "You are Skyra's inner voice"):
			return "<prediction>tell michael about the kernel</prediction>", nil
		case strings.Contains(system, "Skyra's motor"):
			return "<engage><michael>the kernel walks children by activation</michael></engage>", nil
		default:
			return "<label>help michael with the skyra project</label>", nil
		}
	}
}

// engageFrom builds the external-utterance impulse the world would deliver.
func engageFrom(sender, said string) skyra.Impulse {
	return skyra.Impulse{
		Primitive: "engage",
		Source:    &skyra.ImpulseSource{SpaceID: sender, Coord: skyra.Coord{X: 1, Y: 1}},
		Payload:   said,
	}
}

// runCycle drives one cognitive cycle directly (no spawns): sensory intake →
// MidLevel grouping → other. Bumps MidLevel's proper time so each cycle's
// grouping gets its own anchor (the kernel does this per wave in the real
// pipeline).
func runCycle(m *dna.MembraneCell, cycle int, imp skyra.Impulse) {
	sensory := m.Sensory()
	mid := sensory.MidLevel()
	sensory.Slice().SetCount(cycle)
	mid.Slice().SetCount(cycle)
	sensory.Observe(nil, imp)
	mid.Observe(nil, imp)
	mid.Other().Observe(nil, imp)
}

// findGradedOtherRoot returns the graded other-distribution root, under
// MidLevel's lock.
func findGradedOtherRoot(mid *dna.MidLevelCell) (dna.Prediction, bool) {
	sl := mid.Slice()
	sl.Lock()
	defer sl.Unlock()
	for _, gr := range mid.Groupings() {
		mc, ok := gr.(*dna.MemoryCell)
		if !ok {
			continue
		}
		for _, child := range mc.Children() {
			cc, ok := child.(*dna.MemoryCell)
			if !ok {
				continue
			}
			if p, ok := cc.Content().(dna.Prediction); ok && p.Role == "other" && p.Graded && len(cc.Children()) > 0 {
				return p, true
			}
		}
	}
	return dna.Prediction{}, false
}

// The grade half, end to end: cycle 1 bets a distribution; cycle 2's external
// engage grades it. The judge only identifies the branch — the score is
// assigned by code from the STORED probability (0.30 for branch 2), and the
// judge's present carries no probabilities to anchor on.
func TestOtherGradesPriorDistribution(t *testing.T) {
	var judgePresents []string
	m, _ := newBeingWithInfer("skyra", gradingMock(t, "<match>2</match>", "<grade>yes</grade>", &judgePresents))
	mid := m.Sensory().MidLevel()

	runCycle(m, 1, engageFrom("michael", "can you help me with the skyra project"))
	runCycle(m, 2, engageFrom("michael", "so how is the kernel put together?"))

	if len(judgePresents) != 1 {
		t.Fatalf("judge calls = %d, want 1", len(judgePresents))
	}
	// Adversarial-neutral: no ownership, no probabilities in the judge's view.
	if strings.Contains(judgePresents[0], "0.7") || strings.Contains(judgePresents[0], "p=") {
		t.Fatalf("judge present leaks probabilities: %q", judgePresents[0])
	}
	if strings.Contains(strings.ToLower(judgePresents[0]), "your prediction") {
		t.Fatalf("judge present leaks ownership: %q", judgePresents[0])
	}
	if !strings.Contains(judgePresents[0], "so how is the kernel put together?") {
		t.Fatalf("judge present missing evidence: %q", judgePresents[0])
	}

	graded, ok := findGradedOtherRoot(mid)
	if !ok {
		t.Fatal("no graded other-distribution root found")
	}
	if graded.Grade != 0.30 {
		t.Fatalf("grade = %v, want 0.30 (stored p of matched branch, assigned by code)", graded.Grade)
	}
	if graded.Evidence != "so how is the kernel put together?" {
		t.Fatalf("evidence = %q", graded.Evidence)
	}
}

// A full miss scores zero — the distribution put no mass on what happened.
func TestOtherGradeNoneScoresZero(t *testing.T) {
	m, _ := newBeingWithInfer("skyra", gradingMock(t, "<match>none</match>", "<grade>yes</grade>", nil))
	mid := m.Sensory().MidLevel()

	runCycle(m, 1, engageFrom("michael", "can you help me with the skyra project"))
	runCycle(m, 2, engageFrom("michael", "actually never mind, lets get lunch"))

	graded, ok := findGradedOtherRoot(mid)
	if !ok {
		t.Fatal("no graded other-distribution root found")
	}
	if graded.Grade != 0 {
		t.Fatalf("grade = %v, want 0 on a full miss", graded.Grade)
	}
}

// Self-initiated triggers (perceive/move returns) are reafference — self's
// lane. They must NOT grade other-predictions (the gate's MVP form).
func TestOtherGradeSkipsSelfInitiated(t *testing.T) {
	var judgePresents []string
	m, _ := newBeingWithInfer("skyra", gradingMock(t, "<match>1</match>", "<grade>yes</grade>", &judgePresents))
	mid := m.Sensory().MidLevel()

	runCycle(m, 1, engageFrom("michael", "can you help me with the skyra project"))
	// a self-initiated trigger, the way sensory re-prompts after a perceive return
	runCycle(m, 2, skyra.Impulse{
		Primitive: "perceive",
		Source:    &skyra.ImpulseSource{SpaceID: "skyra"},
	})

	if len(judgePresents) != 0 {
		t.Fatalf("judge fired on self-initiated wave: %d calls", len(judgePresents))
	}
	if _, ok := findGradedOtherRoot(mid); ok {
		t.Fatal("a bet was graded without external evidence")
	}
}

// Self grades EXECUTION: proposal → motor action (ProposalID pairing) → the
// execution judge → grade stamped on the proposal. And a realized proposal is
// consumed — motor never fires it twice.
func TestSelfGradesRealizedProposal(t *testing.T) {
	m, _ := newBeingWithInfer("skyra", gradingMock(t, "<match>none</match>", "<grade>yes</grade>", nil))
	sensory := m.Sensory()
	mid := sensory.MidLevel()
	self := mid.Other().Self()
	motor := m.Motor()

	imp := engageFrom("michael", "can you help me with the skyra project")
	runCycle(m, 1, imp)
	self.Observe(nil, imp) // predict half: deposits the proposal
	motor.Observe(nil, skyra.Impulse{Primitive: "act"})
	if motor.ActionCount() != 1 {
		t.Fatalf("action count = %d, want 1", motor.ActionCount())
	}

	// consumed: the realized proposal is never re-fired
	motor.Observe(nil, skyra.Impulse{Primitive: "act"})
	if motor.ActionCount() != 1 {
		t.Fatalf("action count = %d after re-trigger, want 1 (proposal consumed)", motor.ActionCount())
	}

	// next self wave grades the realized proposal (advance proper time as the
	// kernel does per wave — same-count keys would collide)
	mid.Slice().SetCount(2)
	self.Observe(nil, imp)

	sl := mid.Slice()
	sl.Lock()
	defer sl.Unlock()
	found := false
	for _, gr := range mid.Groupings() {
		mc, ok := gr.(*dna.MemoryCell)
		if !ok {
			continue
		}
		for _, child := range mc.Children() {
			cc, ok := child.(*dna.MemoryCell)
			if !ok {
				continue
			}
			p, ok := cc.Content().(dna.Prediction)
			if !ok || p.Role != "self" || !p.Graded {
				continue
			}
			if p.Grade != 1 {
				t.Fatalf("execution grade = %v, want 1 (judge said yes)", p.Grade)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("realized proposal was not graded")
	}
}

// Grade-triggered re-framing: when other's grade is a TOTAL MISS (judge returns
// none), the framing is wrong — other must kick MidLevel to reframe (failure in
// the present) and abort the stroke (no self spawn), rather than predict again
// under the stale framing.
func TestOtherTotalMissTriggersReframe(t *testing.T) {
	var midPresents []string
	reframeSystemSeen := false
	mock := func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "previous framing of this exchange was wrong"):
			reframeSystemSeen = true
			midPresents = append(midPresents, present)
			return "<label>michael wants lunch, not the kernel</label>", nil
		case strings.Contains(system, "judge of execution"):
			return "<grade>yes</grade>", nil
		case strings.Contains(system, "impartial judge"):
			return "<match>none</match>", nil // total miss
		case strings.Contains(system, "External Perspective"):
			return `<predictions>
  <prediction p="0.7">ask about the kernel</prediction>
  <prediction p="0.3">ask about traversal</prediction>
</predictions>`, nil
		case strings.Contains(system, "You are Skyra's inner voice"):
			return "<prediction>explain the kernel</prediction>", nil
		default:
			return "<label>help michael with the skyra project</label>", nil
		}
	}
	spy := newSpyKernel()
	m := newBeingWith("skyra", spy, mock)
	mid := m.Sensory().MidLevel()
	other := mid.Other()

	// cycle 1: bet a distribution about the kernel (spawns land on the spy channel)
	m.Sensory().Slice().SetCount(1)
	mid.Slice().SetCount(1)
	m.Sensory().Observe(nil, engageFrom("michael", "can you help me with the skyra project"))
	mid.Observe(nil, engageFrom("michael", "can you help me with the skyra project"))
	other.Observe(nil, engageFrom("michael", "can you help me with the skyra project"))

	// cycle 2: evidence contradicts every bet → total miss → reframe spawn
	mid.Slice().SetCount(2)
	m.Sensory().Slice().SetCount(2)
	imp := engageFrom("michael", "actually can we get lunch")
	m.Sensory().Observe(nil, imp)
	other.Observe(nil, imp)

	// other spawned a reframe to MidLevel with the failure summary on the Route
	select {
	case call := <-spy.calls:
		if call.traversed == nil || call.traversed.ID() != "skyra:midlevel" {
			t.Fatalf("reframe spawn traversed = %v, want skyra:midlevel", call.traversed)
		}
		r := call.impulse.(skyra.Impulse).Route.(*dna.Route)
		if r.Kind != dna.RouteReframe || !strings.Contains(r.Summary, "actually can we get lunch") {
			t.Fatalf("reframe route = %+v, want RouteReframe with the evidence in Summary", r)
		}
		// drive MidLevel's reframe handling directly (deterministic)
		mid.Observe(call.origin, call.impulse)
	case <-time.After(time.Second):
		t.Fatal("other did not spawn a reframe on total miss")
	}

	// other must have aborted the stroke — Express does not spawn self
	if got := other.Express(nil, imp); got != nil {
		t.Fatalf("other.Express returned %v after reframe, want nil (stroke aborted)", got)
	}

	// the reframe framing pass ran with the failure in the present
	if !reframeSystemSeen || len(midPresents) != 1 {
		t.Fatalf("reframe pass didn't run cleanly: seen=%v presents=%d", reframeSystemSeen, len(midPresents))
	}
	if !strings.Contains(midPresents[0], "actually can we get lunch") ||
		!strings.Contains(midPresents[0], "ask about the kernel") {
		t.Fatalf("reframe present missing the failure (dead bets + evidence): %q", midPresents[0])
	}
	if mid.Gist() != "michael wants lunch, not the kernel" {
		t.Fatalf("gist = %q, want the reframed label", mid.Gist())
	}
}

// The one-reframe rail is free: the reframe re-run is internal (not an external
// engage), so other's grade gate stays closed on it — no second grade, no loop.
func TestReframeReRunDoesNotRegrade(t *testing.T) {
	graderCalls := 0
	mock := func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "previous framing of this exchange was wrong"):
			return "<label>reframed</label>", nil
		case strings.Contains(system, "impartial judge") && !strings.Contains(system, "execution"):
			graderCalls++
			return "<match>none</match>", nil
		case strings.Contains(system, "External Perspective"):
			return `<predictions><prediction p="1.0">x</prediction></predictions>`, nil
		default:
			return "<label>init</label>", nil
		}
	}
	m, _ := newBeingWithInfer("skyra", mock)
	mid := m.Sensory().MidLevel()
	other := mid.Other()

	runCycle(m, 1, engageFrom("michael", "hello"))
	mid.Slice().SetCount(2)
	m.Sensory().Slice().SetCount(2)
	imp := engageFrom("michael", "surprise")
	m.Sensory().Observe(nil, imp)
	other.Observe(nil, imp) // grades once → none → reframe

	// simulate the reframe re-run reaching other with the internal reframe impulse
	reframeImp := skyra.Impulse{
		Primitive: "reframe",
		Source:    &skyra.ImpulseSource{SpaceID: "skyra"},
		Route:     &dna.Route{Kind: dna.RouteReframe, Summary: "x"},
	}
	other.Observe(nil, reframeImp)

	if graderCalls != 1 {
		t.Fatalf("grader called %d times, want 1 (re-run must not re-grade)", graderCalls)
	}
}

// Self-side grade-triggered reframe: a HARD execution failure (judge `no` — e.g.
// a blocked move) means the framing asked for the impossible. Self must kick a
// reframe and abort (no motor spawn), so the re-run replans under a framing that
// accounts for what the world allows. (The session-1 blocked-move relapse fix.)
func TestSelfExecFailureTriggersReframe(t *testing.T) {
	mock := func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "judge of execution"):
			return "<grade>no</grade>", nil // hard failure
		case strings.Contains(system, "External Perspective"):
			return `<predictions><prediction p="1.0">x</prediction></predictions>`, nil
		case strings.Contains(system, "You are Skyra's inner voice"):
			return "<prediction>hide at 520,500</prediction>", nil
		case strings.Contains(system, "Skyra's motor"):
			return "<move><520,500></move>", nil
		default:
			return "<label>play hide and seek</label>", nil
		}
	}
	spy := newSpyKernel()
	m := newBeingWith("skyra", spy, mock)
	sensory := m.Sensory()
	mid := sensory.MidLevel()
	self := mid.Other().Self()
	motor := m.Motor()

	// cycle 1: frame, self proposes "hide at 520,500", motor realizes it
	sensory.Slice().SetCount(1)
	mid.Slice().SetCount(1)
	imp := engageFrom("michael", "lets play hide and seek")
	sensory.Observe(nil, imp)
	mid.Observe(nil, imp)
	self.Observe(nil, imp)
	motor.Observe(nil, skyra.Impulse{Primitive: "act"}) // realizes the proposal → action child

	// cycle 2: self grades the realized proposal → exec no → reframe + abort
	mid.Slice().SetCount(2)
	self.Observe(nil, imp)

	select {
	case call := <-spy.calls:
		imp2, _ := call.impulse.(skyra.Impulse)
		r, _ := imp2.Route.(*dna.Route)
		if call.traversed == nil || call.traversed.ID() != "skyra:midlevel" || r == nil || r.Kind != dna.RouteReframe {
			t.Fatalf("self did not spawn a reframe to midlevel: traversed=%v route=%+v", call.traversed, r)
		}
		if !strings.Contains(r.Summary, "hide at 520,500") {
			t.Fatalf("reframe summary missing the failed intent: %q", r.Summary)
		}
	case <-time.After(time.Second):
		t.Fatal("self did not spawn a reframe on hard execution failure")
	}

	// stroke aborted — Express does not spawn motor
	if got := self.Express(nil, imp); got != nil {
		t.Fatalf("self.Express returned %v after reframe, want nil", got)
	}
}

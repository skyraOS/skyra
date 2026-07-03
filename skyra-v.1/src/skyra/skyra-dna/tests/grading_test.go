package tests

import (
	"strings"
	"testing"

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

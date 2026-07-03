// Command validate drives a scripted MULTI-turn conversation through a live being
// on real inference — to watch the two-stroke cognition (grade → reframe on a
// total miss, verdict-on-wave) actually fire off mocks.
//
// Each turn: fire michael's utterance at the membrane, wait for the action to
// reach motor, then print the framing (gist), other's top prediction, self's
// proposal, and the emitted action. A turn whose utterance contradicts the prior
// framing should show the gist SHIFT (a reframe).
//
//	go run ./cmd/validate
//	CODEX_MODEL=gpt-5 go run ./cmd/validate
package main

import (
	"fmt"
	"os"
	"time"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

func main() {
	const beingID = "skyra"

	if err := debug.Init("logs"); err != nil {
		fmt.Fprintln(os.Stderr, "log init:", err)
		os.Exit(1)
	}
	defer debug.Close()

	// A scripted arc: turn 1 sets a framing; turn 2 hard-contradicts it (should
	// grade as a total miss → reframe → the gist shifts); turn 3 continues under
	// the corrected framing (should grade as a match → verdict-on-wave to self).
	script := []string{
		"can you help me with the skyra project",
		"actually forget the project — i'm hungry, where should we get lunch",
		"yeah somewhere with good tacos nearby",
	}

	kernel := relation.NewRelation()
	being := dna.NewMembraneCell("be helpful", dna.BuildContext{
		BeingID:  beingID,
		NewSlice: func(id string) skyra.Slice { return world.NewSlice(id, beingID) },
		Kernel:   kernel,
		Infer:    inference.CallCodex,
	})
	mid := being.Sensory().MidLevel()
	motor := being.Motor()

	for i, utterance := range script {
		fmt.Printf("\n═══════════ turn %d ═══════════\n→ michael: %q\n", i+1, utterance)
		gistBefore := lockedGist(mid)
		before := motor.ActionCount()

		kernel.Realize(being, being, skyra.Impulse{
			Primitive: "engage", // the external-utterance primitive sensory records + the grade gate keys on
			Source:    &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
			Payload:   utterance,
		})

		if !waitForNewAction(motor, before, 4*time.Minute) {
			fmt.Println("✗ timed out waiting for the action to reach motor")
			os.Exit(1)
		}

		gistAfter := lockedGist(mid)
		selfPred, otherPred := predictions(mid)
		fmt.Printf("framing (gist): %q", gistAfter)
		if i > 0 && gistAfter != gistBefore {
			fmt.Printf("   ⟵ SHIFTED from %q", gistBefore)
		}
		fmt.Printf("\nother predicts:  %s\nself proposes:   %s\n→ motor:         %s\n",
			otherPred, selfPred, lockedLastAction(motor))
	}
	fmt.Println("\n═══ done — inspect logs/skyra/{midlevel,other,self}.log for grade/reframe traces ═══")
}

func waitForNewAction(m *dna.MotorCell, prevCount int, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if lockedActionCount(m) > prevCount {
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false
}

func lockedActionCount(m *dna.MotorCell) int {
	sl := m.Slice()
	sl.Lock()
	defer sl.Unlock()
	return m.ActionCount()
}

func lockedLastAction(m *dna.MotorCell) string {
	sl := m.Slice()
	sl.Lock()
	defer sl.Unlock()
	return m.LastAction()
}

func lockedGist(mid *dna.MidLevelCell) string {
	sl := mid.Slice()
	sl.Lock()
	defer sl.Unlock()
	return mid.Gist()
}

func predictions(mid *dna.MidLevelCell) (self, other string) {
	sl := mid.Slice()
	sl.Lock()
	defer sl.Unlock()

	bestTS := -1
	var best skyra.Space
	for _, gr := range mid.Groupings() {
		mc, ok := gr.(*dna.MemoryCell)
		if !ok {
			continue
		}
		if g, ok := mc.Content().(dna.Gist); ok && g.Timestamp > bestTS {
			bestTS, best = g.Timestamp, gr
		}
	}
	mc, ok := best.(*dna.MemoryCell)
	if !ok {
		return "", ""
	}
	selfTS, otherTS := -1, -1
	for _, child := range mc.Children() {
		cm, ok := child.(*dna.MemoryCell)
		if !ok {
			continue
		}
		p, ok := cm.Content().(dna.Prediction)
		if !ok {
			continue
		}
		switch p.Role {
		case "self":
			if p.Timestamp > selfTS {
				selfTS, self = p.Timestamp, p.Text
			}
		case "other":
			if p.Timestamp <= otherTS || len(cm.Children()) == 0 {
				continue
			}
			var bestPred dna.Prediction
			for _, gc := range cm.Children() {
				if gp, ok := gc.(*dna.MemoryCell); ok {
					if pp, ok := gp.Content().(dna.Prediction); ok && pp.Probability > bestPred.Probability {
						bestPred = pp
					}
				}
			}
			if bestPred.Text != "" {
				otherTS, other = p.Timestamp, bestPred.Text
			}
		}
	}
	return self, other
}

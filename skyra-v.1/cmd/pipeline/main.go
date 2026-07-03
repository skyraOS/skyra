// Command pipeline drives one impulse through a live being on REAL inference —
// the first end-to-end run of the assembled forward pipeline
// (membrane → sensory → MidLevel → other → self → motor) off mocks.
//
// Inference goes through the Codex CLI (inference.CallCodex): codex is invoked
// locally under the ChatGPT login, so the LLM calls are covered by the Codex
// subscription rather than per-token API billing. Override the model with
// CODEX_MODEL; override the perceived utterance with the first CLI arg.
//
//	go run ./cmd/pipeline
//	go run ./cmd/pipeline "what's the architecture of this thing"
//
// The forward pass crosses actor boundaries with async (goroutine) spawns, so the
// top-level Realize returns before the cascade reaches motor; we poll the out-face
// until the proposal lands (or time out).
package main

import (
	"fmt"
	"os"
	"time"

	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

func main() {
	const beingID = "skyra"

	utterance := "can you help me with the skyra project"
	if len(os.Args) > 1 {
		utterance = os.Args[1]
	}

	kernel := relation.NewRelation()
	being := dna.NewMembraneCell("be helpful", dna.BuildContext{
		BeingID:  beingID,
		NewSlice: func(id string) skyra.Slice { return world.NewSlice(id, beingID) },
		Kernel:   kernel,
		Infer:    inference.CallCodex, // LLM via the Codex CLI, not an HTTP API
	})

	fmt.Printf("→ michael: %q\n\n", utterance)

	// Root the wave at the membrane. The pipeline crosses actor boundaries with
	// goroutine spawns, so this returns before motor is reached.
	kernel.Realize(being, being, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: utterance,
	})

	motor := being.Motor()
	action, ok := waitForAction(motor, 3*time.Minute)
	if !ok {
		fmt.Println("✗ timed out waiting for the proposal to reach motor")
		os.Exit(1)
	}

	mid := being.Sensory().MidLevel()
	selfPred, otherPred := predictions(mid)

	fmt.Println("framing (MidLevel gist):", mid.Gist())
	fmt.Println("other's predicted move: ", otherPred)
	fmt.Println("self's proposed move:   ", selfPred)
	fmt.Println("→ motor (out-face):     ", action)
}

// waitForAction polls the out-face until self's proposal has been read off the
// slice (or the deadline passes). Read under motor's slice lock so the poll is
// synchronized with the motor wave's locked write.
func waitForAction(m *dna.MotorCell, within time.Duration) (string, bool) {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if a := lockedLastAction(m); a != "" {
			return a, true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return "", false
}

func lockedLastAction(m *dna.MotorCell) string {
	sl := m.Slice()
	sl.Lock()
	defer sl.Unlock()
	return m.LastAction()
}

// predictions reads the self/other predicted moves off the most-recent grouping,
// under MidLevel's slice lock — for display.
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
	for _, child := range mc.Children() {
		if p, ok := child.Content().(dna.Prediction); ok {
			switch p.Role {
			case "self":
				self = p.Text
			case "other":
				other = p.Text
			}
		}
	}
	return self, other
}

// Command validate-v2 drives a scripted multi-turn conversation through the dnav2
// pipeline on REAL inference (Codex CLI) — the phase-2 de-risk checkpoint. Phase 1
// proved a single cold-start turn flows an action through the orchestrator; phase 2
// proves the ported grading + reframe fire on real inference:
//
//   - turn 1 (cold start): frame → predict → emit; establishes an other-distribution
//   - turn 2 (on-track):    the reply realizes a predicted branch → other:grade "held"
//   - turn 3 (pivot):       an unrelated turn → total miss → "all-preds<0" → REFRAME
//
// Prints, per turn: the grade condition, whether it reframed, the gist, and the
// emitted action — the live analogue of v1's 3-turn trace.
//
//	go run ./cmd/validate-v2
package main

import (
	"fmt"
	"os"
	"time"

	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

type turn struct {
	id, prev, utterance, note string
}

func main() {
	script := []turn{
		{"X1", "", "can you help me with the skyra project", "cold start — no prior bet"},
		{"X2", "X1", "how does the kernel actually work?", "on-track — should hold a prediction"},
		{"X3", "X2", "forget that — what's the weather like where you are?", "hard pivot — should miss + reframe"},
	}

	d := dnav2.NewDNA("skyra", world.NewSlice("skyra:nucleus", "skyra"))
	dnav2.WireGenes(d)
	d.SetInfer(inference.CallCodex)

	var emitted skyra.Impulse
	var gotEmit bool
	d.SetEmit(func(imp skyra.Impulse) { emitted, gotEmit = imp, true })

	board := dnav2.NewBoard()
	kernel := relation.NewRelation()
	slices := map[string]skyra.Slice{}
	for _, ns := range []string{"sensory", "other", "self", "midlevel", "motor"} {
		slices[ns] = world.NewSlice("skyra:"+ns, "skyra")
	}
	orch := dnav2.NewOrchestrator(d, board, kernel, world.NewSlice("skyra:membrane", "skyra"), slices)

	for _, tn := range script {
		gotEmit = false
		done := make(chan []dnav2.Span, 1)
		orch.SetFlush(func(_ string, trace []dnav2.Span) { done <- trace })

		fmt.Printf("\n─── %s ─── (%s)\n→ michael: %q\n", tn.id, tn.note, tn.utterance)
		orch.Fire(tn.id, tn.prev, skyra.Impulse{
			Primitive: "engage",
			Source:    &skyra.ImpulseSource{SpaceID: "michael"},
			Payload:   tn.utterance,
		})

		var trace []dnav2.Span
		select {
		case trace = <-done:
		case <-time.After(5 * time.Minute):
			fmt.Println("✗ timed out")
			os.Exit(1)
		}

		fmt.Printf("  other:grade:  %s\n", condOf(trace, dnav2.StepOtherGrade))
		fmt.Printf("  reframed:     %v\n", board.Row(tn.id, "").Reframed())
		fmt.Printf("  framing:      %s\n", gist(d))
		fmt.Printf("  self plan:    %s\n", selfPlan(d))
		if gotEmit {
			fmt.Printf("  → emitted:    %s %s %v\n", emitted.Primitive, emitted.TargetID, emitted.Payload)
		} else {
			fmt.Println("  → emitted:    (nothing)")
		}
	}
	fmt.Println("\n═══ done ═══")
}

func condOf(trace []dnav2.Span, step dnav2.Step) string {
	for _, s := range trace {
		if s.Step == string(step) {
			return s.Condition
		}
	}
	return "(step did not run)"
}

func gist(d *dnav2.DNA) string {
	g := d.Groupings().Latest()
	if g == nil {
		return "(none)"
	}
	if gg, ok := g.Content().(dna.Gist); ok {
		return gg.Text
	}
	return "(none)"
}

func selfPlan(d *dnav2.DNA) string {
	g := d.Groupings().Latest()
	if g == nil {
		return "(none)"
	}
	best := ""
	bestTS := -1
	for _, c := range g.Children() {
		if mc, ok := c.(*dna.MemoryCell); ok {
			if p, ok := mc.Content().(dna.Prediction); ok && p.Role == "self" && p.Timestamp > bestTS {
				bestTS, best = p.Timestamp, p.Text
			}
		}
	}
	if best == "" {
		return "(none)"
	}
	return best
}

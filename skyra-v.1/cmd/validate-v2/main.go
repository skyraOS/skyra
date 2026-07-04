// Command validate-v2 drives one external impulse through the dnav2 pipeline on
// REAL inference (Codex CLI) — the phase-1 de-risk checkpoint: does a transcribed
// worker actually produce a real inference result and flow an action through the
// orchestrator? Prints the gist (midlevel:frame), other's top prediction, self's
// plan, and the emitted action.
//
//	go run ./cmd/validate-v2
//	go run ./cmd/validate-v2 "what's the architecture of this thing"
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

func main() {
	utterance := "can you help me with the skyra project"
	if len(os.Args) > 1 {
		utterance = os.Args[1]
	}

	d := dnav2.NewDNA("skyra")
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

	done := make(chan string, 1)
	orch.SetFlush(func(id string, trace []dnav2.Span) { done <- id })

	fmt.Printf("→ michael: %q\n\n", utterance)
	orch.Fire("X", "", skyra.Impulse{
		Primitive: "engage",
		Source:    &skyra.ImpulseSource{SpaceID: "michael"},
		Payload:   utterance,
	})

	select {
	case <-done:
	case <-time.After(5 * time.Minute):
		fmt.Println("✗ timed out")
		os.Exit(1)
	}

	fmt.Println("framing (gist):     ", gist(d))
	fmt.Println("other predicts:     ", topOther(d))
	fmt.Println("self proposes:      ", selfPlan(d))
	if gotEmit {
		fmt.Printf("→ motor emitted:     %s %s %v\n", emitted.Primitive, emitted.TargetID, emitted.Payload)
	} else {
		fmt.Println("→ motor emitted:      (nothing)")
	}
	fmt.Println("\n═══ done ═══")
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

func topOther(d *dnav2.DNA) string {
	g := d.Groupings().Latest()
	if g == nil {
		return "(none)"
	}
	var best dna.Prediction
	for _, c := range g.Children() {
		root, ok := c.(*dna.MemoryCell)
		if !ok {
			continue
		}
		if p, ok := root.Content().(dna.Prediction); !ok || p.Role != "other" {
			continue
		}
		for _, gc := range root.Children() {
			if pp, ok := gc.(*dna.MemoryCell); ok {
				if pr, ok := pp.Content().(dna.Prediction); ok && pr.Probability > best.Probability {
					best = pr
				}
			}
		}
	}
	if best.Text == "" {
		return "(none)"
	}
	return fmt.Sprintf("p=%.2f %s", best.Probability, best.Text)
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

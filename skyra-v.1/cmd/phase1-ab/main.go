// Command phase1-ab is the Phase-1 de-risk (traversal-implementation-plan.md): run one
// real cognitive cycle on live Codex, then render the resulting other-distribution two
// ways — the hand-built renderOtherDistribution vs. the traversal-assembled frame — and
// show them side by side. The question: does a traversed frame reconstruct (or beat) the
// hand-built present?
//
//	go run ./cmd/phase1-ab
//	go run ./cmd/phase1-ab "how does the kernel work?"
package main

import (
	"fmt"
	"os"
	"time"

	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	"skyra-v1/src/skyra/world"
)

func main() {
	utterance := "can you help me with the skyra project"
	if len(os.Args) > 1 {
		utterance = os.Args[1]
	}

	d := dnav2.NewDNA("skyra", world.NewSlice("skyra:nucleus", "skyra"))
	dnav2.WireGenes(d)
	d.SetInfer(inference.CallCodex)
	d.SetEmit(func(skyra.Impulse) {})

	board := dnav2.NewBoard()
	kernel := relation.NewRelation()
	slices := map[string]skyra.Slice{}
	for _, ns := range []string{"sensory", "other", "self", "midlevel", "motor"} {
		slices[ns] = world.NewSlice("skyra:"+ns, "skyra")
	}
	orch := dnav2.NewOrchestrator(d, board, kernel, world.NewSlice("skyra:membrane", "skyra"), slices)

	done := make(chan string, 1)
	orch.SetFlush(func(id string, _ []dnav2.Span) { done <- id })

	fmt.Printf("→ michael: %q\n(running one live cycle...)\n\n", utterance)
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

	for _, budget := range []float64{10000, 40} {
		hand, traversal, lit := dnav2.ComparePresents(d, budget)
		fmt.Printf("═══ budget = %.0f tokens (lit %d nodes) ═══\n\n", budget, lit)
		fmt.Printf("── HAND (renderOtherDistribution) ──\n%s\n\n", orNone(hand))
		fmt.Printf("── TRAVERSAL (AssembleFramePresent) ──\n%s\n\n", orNone(traversal))
	}
}

func orNone(s string) string {
	if s == "" {
		return "(empty)"
	}
	return s
}

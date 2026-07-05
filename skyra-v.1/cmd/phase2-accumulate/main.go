// Command phase2-accumulate is the Phase-2e demonstration
// (traversal-implementation-plan.md): run a scripted multi-turn conversation on live
// Codex with the full economy wired — the frame feeds self:predict, its edges are
// crossed, and each other:grade / self:grade distributes the closure back across the
// frame. After every turn it prints the edge-alignments so you can watch them MOVE:
// the accumulation loop, live, for the first time.
//
//	go run ./cmd/phase2-accumulate
package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	"skyra-v1/src/skyra/world"
)

func main() {
	script := []struct{ id, prev, say string }{
		{"X1", "", "can you help me with the skyra project"},
		{"X2", "X1", "how does the kernel actually work?"},
		{"X3", "X2", "forget that — what's the weather where you are?"},
	}

	d := dnav2.NewDNA("skyra")
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

	for _, t := range script {
		fmt.Printf("\n═══ %s → michael: %q ═══\n", t.id, t.say)
		orch.Fire(t.id, t.prev, skyra.Impulse{
			Primitive: "engage",
			Source:    &skyra.ImpulseSource{SpaceID: "michael"},
			Payload:   t.say,
		})
		select {
		case <-done:
		case <-time.After(5 * time.Minute):
			fmt.Println("✗ timed out")
			os.Exit(1)
		}
		printAlignments(d)
	}
	fmt.Println("\n═══ done — watch A move across turns (grades flow credit back over the frame) ═══")
}

func printAlignments(d *dnav2.DNA) {
	all := d.Activation().All()
	if len(all) == 0 {
		fmt.Println("   (no edges yet)")
		return
	}
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("   %d edges in the economy:\n", len(keys))
	for _, k := range keys {
		a := all[k]
		fmt.Printf("     %-30s  A=%+.3f  F=%.3f  N=%d\n", k, a.A, a.F, a.N)
	}
}

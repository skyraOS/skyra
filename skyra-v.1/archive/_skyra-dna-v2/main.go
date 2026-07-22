package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dnav2 "skyra-v1/src/skyra/skyra-dna-v2"
	"skyra-v1/src/skyra/world"
)

func main() {
	if err := debug.Init("logs"); err != nil {
		fmt.Fprintln(os.Stderr, "log init:", err)
		os.Exit(1)
	}
	defer debug.Close()

	d := dnav2.NewDNA("skyra", world.NewSlice("skyra:nucleus", "skyra"))
	dnav2.WireGenes(d)
	d.SetInfer(dnav2.NewInferLogger("skyra", inference.CallCodex).Infer)

	var mu sync.Mutex
	var emitted skyra.Impulse
	d.SetEmit(func(imp skyra.Impulse) { mu.Lock(); emitted = imp; mu.Unlock() })

	board := dnav2.NewBoard()
	slices := map[string]skyra.Slice{}
	for _, ns := range []string{"sensory", "final", "efficient", "midlevel", "motor"} {
		slices[ns] = world.NewSlice("skyra:"+ns, "skyra")
	}
	orch := dnav2.NewOrchestrator(d, board, relation.NewRelation(), world.NewSlice("skyra:membrane", "skyra"), slices)

	seq, prev := 0, ""
	run := func(utterance string) {
		seq++
		id := fmt.Sprintf("in:%d", seq)
		done := make(chan struct{}, 1)
		orch.SetFlush(func(flushID string, _ []dnav2.Span) {
			if flushID == id {
				done <- struct{}{}
			}
		})
		mu.Lock()
		emitted = skyra.Impulse{}
		mu.Unlock()
		orch.Fire(id, prev, skyra.Impulse{Primitive: world.PrimitiveEngage, Payload: utterance, Source: &skyra.ImpulseSource{SpaceID: "michael"}})
		prev = id
		select {
		case <-done:
		case <-time.After(10 * time.Minute):
			fmt.Println("skyra: timed out")
			return
		}
		g := d.Episodes().Latest()
		if g != nil {
			fmt.Println("framing:", dnav2.FrameText(g))
		}
		mu.Lock()
		action := emitted
		mu.Unlock()
		if action.Primitive == world.PrimitiveEngage {
			fmt.Printf("skyra: %v\n", action.Payload)
		} else if action.Primitive != "" {
			fmt.Printf("skyra: %s %s %v\n", action.Primitive, action.TargetID, action.Payload)
		} else {
			fmt.Println("skyra: no action emitted")
		}
	}

	if len(os.Args) > 1 {
		run(strings.Join(os.Args[1:], " "))
		return
	}
	fmt.Println("skyra dnav2 ready. Type a message, or Ctrl-C to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		if text := strings.TrimSpace(scanner.Text()); text != "" {
			run(text)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "stdin:", err)
		os.Exit(1)
	}
}

package dnav2

import (
	"strings"
	"sync"
	"testing"
	"time"

	"skyra-v1/src/skyra"
)

// The whole cold-start turn: an external engage flows through every ported gene on
// mock inference and produces an emitted action — sensory → final:grade(no-prior)
// → midlevel:frame → fan-out{final:predict, efficient:predict} → join → motor:collapse.
func TestColdStartFullPipelineEmitsAction(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		switch {
		case strings.Contains(system, "evolving being"):
			return "<requery>assist[1] with[1→2] the {skyra project}[2]</requery>", nil
		case strings.Contains(system, "External Perspective"):
			return `<predictions><prediction p="0.7">asks about the kernel</prediction><prediction p="0.3">asks about setup</prediction></predictions>`, nil
		// NOTE: the prompts cross-reference (motor mentions "inner voice", efficient
		// mentions "the motor system") — discriminate on the unique OPENING phrase.
		case strings.Contains(system, "You are Skyra's motor system"):
			return "<engage><michael>What part of the skyra project can I help with?</michael></engage>", nil
		case strings.Contains(system, "You are Skyra's inner voice"):
			// NOTE: efficient reads final's LATEST distribution — on turn 1's concurrent
			// fan-out that's "(no prediction)" (final is landing at the same time).
			// That's the lagged-reliance model; efficient plans unconditioned.
			return "<prediction>Ask michael what part of the skyra project he wants help with</prediction>", nil
		}
		return "", nil
	})

	var mu sync.Mutex
	var emitted []skyra.Impulse
	d.SetEmit(func(imp skyra.Impulse) { mu.Lock(); emitted = append(emitted, imp); mu.Unlock() })

	o, _ := newOrch(d)
	done := make(chan string, 1)
	o.SetFlush(func(id string, trace []Span) { done <- id })
	o.Fire("X", "", skyra.Impulse{Primitive: "engage", Source: &skyra.ImpulseSource{SpaceID: "michael"}, Payload: "can you help with the skyra project"})

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("pipeline did not complete")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(emitted) != 1 {
		t.Fatalf("emitted %d actions, want exactly 1", len(emitted))
	}
	if emitted[0].Primitive != "engage" || emitted[0].TargetID != "michael" {
		t.Fatalf("emitted = %+v, want engage→michael", emitted[0])
	}
	if !strings.Contains(d.LastAction(), "help with") {
		t.Fatalf("last action = %q, want the engage payload", d.LastAction())
	}
}

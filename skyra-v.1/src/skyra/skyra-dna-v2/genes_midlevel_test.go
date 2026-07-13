package dnav2

import (
	"strings"
	"testing"
	"time"

	"skyra-v1/src/skyra"
)

// The cold-start pipeline through framing: an external engage flows sensory →
// other:grade(no-prior) → midlevel:frame (mock infer → frame → episode). Verifies
// the frame gene renders the exchange, creates the episode, and the orchestrator
// threads it onto the board for the fan-out. Stops after framing (predicts not yet
// ported) by having predict genes report a terminal-ish condition.
func TestColdStartFramesAndGroups(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	var framePresent string
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		if strings.Contains(system, "evolving being") {
			framePresent = present
			return "<requery>assist[1] with[1→2] the {skyra project}[2]</requery>", nil
		}
		return "", nil
	})
	// stub the not-yet-ported predict genes so the fan-out completes and the wave ends
	d.RegisterObserve(string(StepOtherPredict), func(r *EphemeralRoot, o skyra.Space, i any) { r.Condition = "predicted" })
	d.RegisterObserve(string(StepSelfPredict), func(r *EphemeralRoot, o skyra.Space, i any) { r.Condition = "predicted" })
	d.RegisterObserve(string(StepMotorCollapse), func(r *EphemeralRoot, o skyra.Space, i any) { r.Condition = "emitted" })

	o, board := newOrch(d)
	done := make(chan string, 1)
	o.SetFlush(func(id string, trace []Span) { done <- id })
	o.Fire("X", "", skyra.Impulse{Primitive: "engage", Source: &skyra.ImpulseSource{SpaceID: "michael"}, Payload: "can you help with the skyra project"})

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("pipeline did not complete")
	}

	// the frame carried the exchange
	if !strings.Contains(framePresent, "can you help with the skyra project") {
		t.Fatalf("frame present missing the exchange: %q", framePresent)
	}
	// an episode was created and threaded onto the board (the predict workers got it)
	g, ok := board.Row("X", "").Episode().(*MemoryCell)
	if !ok || g == nil {
		t.Fatal("no active episode threaded onto the board")
	}
	if got := frameText(g); got != "assist with the skyra project" {
		t.Fatalf("episode frame = %q, want the framed label", got)
	}
}

package tests

import (
	"testing"
	"time"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// The read-half of Global's grouping pass: on a wave, Global reads sensory's
// warm surroundings (under sensory's lock), parses them entity-outward, and
// assembles a chronological (timestamp-ordered) frame — its own slice, calling
// each node's Parse() directly (no kernel payload). Inference call is stubbed.
func TestGlobalAssemblesTimestampOrderedFrame(t *testing.T) {
	m, kernel := newBeing("skyra")
	sensory := m.Sensory()
	global := sensory.Global()

	// Warm sensory with two exchanges at different proper-times.
	sensory.Slice().SetCount(1)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "first",
	})
	sensory.Slice().SetCount(2)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 2, Y: 2}},
		Payload: "second",
	})

	// Global's wave assembles the frame from sensory's warm surroundings
	// (traversed directly → synchronous; no spawn).
	kernel.Realize(nil, global, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 2, Y: 2}},
		Payload: "now",
	})

	frame := global.Frame()
	if len(frame) == 0 {
		t.Fatal("global assembled an empty frame")
	}

	// Chronological: timestamps non-decreasing.
	for i := 1; i < len(frame); i++ {
		if frame[i].Timestamp < frame[i-1].Timestamp {
			t.Fatalf("frame not timestamp-ordered: %d before %d",
				frame[i-1].Timestamp, frame[i].Timestamp)
		}
	}

	// The actual exchange (both utterances) made it into the frame.
	sawFirst, sawSecond := false, false
	for _, e := range frame {
		rec, ok := e.Content.(dna.InteractionRecord)
		if !ok {
			continue
		}
		switch rec.Utterance {
		case "first":
			sawFirst = true
		case "second":
			sawSecond = true
		}
	}
	if !sawFirst || !sawSecond {
		t.Fatalf("frame missing exchange: first=%v second=%v", sawFirst, sawSecond)
	}
}

// Global spawns the next pipeline stage (the other agent) on Express, owned by
// other (origin = traversed = other) — the same forward-spawn pattern.
func TestGlobalFiresOtherOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	global := m.Sensory().Global()

	global.Express(nil, skyra.Impulse{Payload: "frame"})

	select {
	case call := <-spy.calls:
		if call.origin == nil || call.origin.ID() != "skyra:other" {
			t.Fatalf("spawn origin = %v, want skyra:other", call.origin)
		}
		if call.traversed == nil || call.traversed.ID() != "skyra:other" {
			t.Fatalf("spawn traversed = %v, want skyra:other", call.traversed)
		}
	case <-time.After(time.Second):
		t.Fatal("Global did not spawn the other-agent wave on Express")
	}
}

// Cold start: no surroundings → an empty frame, not a panic.
func TestGlobalFrameEmptyWhenNoSurroundings(t *testing.T) {
	m, kernel := newBeing("skyra")
	global := m.Sensory().Global()

	kernel.Realize(nil, global, skyra.Impulse{Payload: "ambient"})

	if len(global.Frame()) != 0 {
		t.Fatalf("cold-start frame = %+v, want empty", global.Frame())
	}
}

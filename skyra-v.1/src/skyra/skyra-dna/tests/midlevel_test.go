package tests

import (
	"strings"
	"testing"
	"time"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// The read-half of Global's grouping pass: on a wave, Global reads sensory's
// warm surroundings (under sensory's lock), parses them entity-outward, and
// assembles a chronological (timestamp-ordered) frame — its own slice, calling
// each node's Parse() directly (no kernel payload). Inference call is stubbed.
func TestMidLevelAssemblesTimestampOrderedFrame(t *testing.T) {
	m, kernel := newBeing("skyra")
	sensory := m.Sensory()
	mid := sensory.MidLevel()

	// Warm sensory with two exchanges at different proper-times.
	sensory.Slice().SetCount(1)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Primitive: "engage",
		Payload: "first",
	})
	sensory.Slice().SetCount(2)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 2, Y: 2}},
		Primitive: "engage",
		Payload: "second",
	})

	// Global's wave assembles the frame from sensory's warm surroundings
	// (traversed directly → synchronous; no spawn).
	kernel.Realize(nil, mid, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 2, Y: 2}},
		Payload: "now",
	})

	frame := mid.Frame()
	if len(frame) == 0 {
		t.Fatal("mid assembled an empty frame")
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

// The first real behavior: cold-start grouping pass. Global reads the warm
// surroundings, renders them into the prompt, calls inference (mocked), and stores
// the parsed <label> as the cycle's gist (framing).
func TestMidLevelColdStartPredictsLabel(t *testing.T) {
	var gotPresent string
	mock := func(system, present string) (string, error) {
		gotPresent = present
		return "<label>help michael with the skyra project</label>", nil
	}
	m, _ := newBeingWithInfer("skyra", mock)
	sensory := m.Sensory()
	mid := sensory.MidLevel()

	sensory.Slice().SetCount(1)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Primitive: "engage",
		Payload: "can you help me with the skyra project",
	})

	// Observe directly — exercises MidLevel's gist/grouping pass without spawning
	// the downstream agents (which would call the shared mock concurrently).
	mid.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "now",
	})

	if mid.Gist() != "help michael with the skyra project" {
		t.Fatalf("gist = %q, want the parsed label", mid.Gist())
	}
	// The exchange reached the prompt (frame rendered into the present).
	if !strings.Contains(gotPresent, "skyra project") {
		t.Fatalf("present did not carry the exchange: %q", gotPresent)
	}

	// Cold-start "create": a grouping anchor carrying the gist was published.
	groupings := mid.Groupings()
	if len(groupings) != 1 {
		t.Fatalf("groupings = %d, want 1 created from the gist", len(groupings))
	}
	for _, gr := range groupings {
		mc, ok := gr.(*dna.MemoryCell)
		if !ok || mc.Kind() != dna.KindGrouping {
			t.Fatalf("grouping = %T kind %v, want *MemoryCell KindGrouping", gr, mc.Kind())
		}
		gist, ok := mc.Content().(dna.Gist)
		if !ok || gist.Text != "help michael with the skyra project" {
			t.Fatalf("grouping gist = %+v, want the predicted label", mc.Content())
		}
	}
}

// Global spawns the next pipeline stage (the other agent) on Express, owned by
// other (origin = traversed = other) — the same forward-spawn pattern.
func TestMidLevelFiresOtherOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	mid := m.Sensory().MidLevel()

	mid.Express(nil, skyra.Impulse{Payload: "frame"})

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
func TestMidLevelFrameEmptyWhenNoSurroundings(t *testing.T) {
	m, kernel := newBeing("skyra")
	mid := m.Sensory().MidLevel()

	kernel.Realize(nil, mid, skyra.Impulse{Payload: "ambient"})

	if len(mid.Frame()) != 0 {
		t.Fatalf("cold-start frame = %+v, want empty", mid.Frame())
	}
}

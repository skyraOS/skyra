package tests

import (
	"testing"
	"time"

	"skyra-v1/src/skyra"
)

// The membrane is a thin boundary: no perception, no children. Observe does
// nothing (perception is sensory's), and there is nothing for the kernel to
// descend — it goes straight to Express.
func TestMembraneIsThin(t *testing.T) {
	m, _ := newBeing("skyra")

	if len(m.Children()) != 0 {
		t.Fatalf("membrane children = %d, want 0 (thin boundary)", len(m.Children()))
	}
	if m.Content() != "be helpful" {
		t.Fatalf("membrane content = %v, want the root drive", m.Content())
	}

	// Observe is a no-op — it must NOT write the surroundings (sensory does that).
	m.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "hello",
	})
	if len(m.Sensory().ByEntity) != 0 {
		t.Fatalf("membrane Observe wrote the surroundings: %+v", m.Sensory().ByEntity)
	}
}

// On Express the membrane fires an inward wave to SENSORY, owned by sensory
// (origin = traversed = sensory) — the membrane only launches and drops out.
func TestMembraneFiresSensoryOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)

	imp := skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "hello skyra",
	}
	m.Express(nil, imp)

	select {
	case call := <-spy.calls:
		if call.origin == nil || call.origin.ID() != "skyra:sensory" {
			t.Fatalf("spawn origin = %v, want skyra:sensory (sensory owns its perception)", call.origin)
		}
		if call.traversed == nil || call.traversed.ID() != "skyra:sensory" {
			t.Fatalf("spawn traversed = %v, want skyra:sensory", call.traversed)
		}
		got, ok := call.impulse.(skyra.Impulse)
		if !ok || got.Payload != "hello skyra" {
			t.Fatalf("spawn impulse = %+v, want the original perceive impulse", call.impulse)
		}
	case <-time.After(time.Second):
		t.Fatal("membrane did not spawn a sensory wave on Express")
	}
}

// The fire is UNCONDITIONAL — even a wave that perceives no known entity still
// runs the sensory wave (traffic-relative decay needs it).
func TestMembraneFiresUnconditionally(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)

	m.Express(nil, skyra.Impulse{Payload: "ambient, no sender"})

	select {
	case call := <-spy.calls:
		if call.traversed == nil || call.traversed.ID() != "skyra:sensory" {
			t.Fatalf("spawn traversed = %v, want skyra:sensory", call.traversed)
		}
	case <-time.After(time.Second):
		t.Fatal("membrane did not fire sensory on a no-sender wave")
	}
}

// The motor sub-actor exists with its indexes initialized (out-wave not wired in
// step 1).
func TestMembraneBuildsMotorIndexes(t *testing.T) {
	m, _ := newBeing("skyra")
	motor := m.Motor()

	if motor.ByEntity == nil {
		t.Fatalf("motor index nil: ByEntity=%v", motor.ByEntity)
	}
	if len(motor.Children()) != 0 {
		t.Fatalf("motor children = %d, want empty", len(motor.Children()))
	}
}

func TestMembraneAgentsExposeSubActors(t *testing.T) {
	m, _ := newBeing("skyra")
	agents := m.Agents()

	if agents["skyra:sensory"] != m.Sensory() {
		t.Fatalf("membrane sensory agent = %v, want sensory", agents["skyra:sensory"])
	}
	if agents["skyra:motor"] != m.Motor() {
		t.Fatalf("membrane motor agent = %v, want motor", agents["skyra:motor"])
	}
}

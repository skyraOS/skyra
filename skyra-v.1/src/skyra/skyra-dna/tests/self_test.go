package tests

import (
	"strings"
	"testing"
	"time"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// The self-agent's predict step: after other has predicted (and appended its
// prediction onto the grouping), self reads the grouping, sees other's prediction
// in its prompt ("self relies on other"), predicts skyra's own move, and appends
// it as grouping → self-prediction.
func TestSelfPredictsRelyingOnOther(t *testing.T) {
	var gotSelfPresent string
	mock := func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "modeling the other"):
			return "<prediction>ask how the runtime is structured</prediction>", nil
		case strings.Contains(system, "predict your own next move"):
			gotSelfPresent = present
			return "<prediction>walk michael through the kernel</prediction>", nil
		default:
			return "<label>help michael with the skyra project</label>", nil
		}
	}
	m, _ := newBeingWithInfer("skyra", mock)
	sensory := m.Sensory()
	midLevel := sensory.MidLevel()
	other := midLevel.Other()
	self := other.Self()

	// Run the predict chain directly (no spawn): MidLevel → other → self.
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "can you help me with the skyra project",
	})
	imp := skyra.Impulse{Source: &skyra.ImpulseSource{SpaceID: "michael"}, Payload: "now"}
	midLevel.Observe(nil, imp)
	other.Observe(nil, imp)
	self.Observe(nil, imp)

	// self relies on other: other's prediction reached self's prompt.
	if !strings.Contains(gotSelfPresent, "ask how the runtime is structured") {
		t.Fatalf("self's prompt did not carry other's prediction: %q", gotSelfPresent)
	}

	// The grouping now carries BOTH predictions.
	sawOther, sawSelf := false, false
	for _, gr := range midLevel.Groupings() {
		mc := gr.(*dna.MemoryCell)
		for _, child := range mc.Children() {
			p, ok := child.Content().(dna.Prediction)
			if !ok {
				continue
			}
			switch p.Role {
			case "other":
				sawOther = true
			case "self":
				sawSelf = true
				if p.Text != "walk michael through the kernel" {
					t.Fatalf("self prediction text = %q", p.Text)
				}
			}
		}
	}
	if !sawOther || !sawSelf {
		t.Fatalf("grouping predictions: other=%v self=%v (want both)", sawOther, sawSelf)
	}
}

// MVP forward pass: self hands its predicted move straight to motor (voting
// parked). The move rides the impulse payload to skyra:motor.
func TestSelfHandsMoveToMotor(t *testing.T) {
	spy := newSpyKernel()
	mock := func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "modeling the other"):
			return "<prediction>ask how the runtime is structured</prediction>", nil
		case strings.Contains(system, "predict your own next move"):
			return "<prediction>walk michael through the kernel</prediction>", nil
		default:
			return "<label>help michael with the skyra project</label>", nil
		}
	}
	m := newBeingWith("skyra", spy, mock)
	sensory := m.Sensory()
	midLevel := sensory.MidLevel()
	other := midLevel.Other()
	self := other.Self()

	// Run the predict chain directly (Observe uses infer, not the kernel).
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "can you help me with the skyra project",
	})
	imp := skyra.Impulse{Source: &skyra.ImpulseSource{SpaceID: "michael"}, Payload: "now"}
	midLevel.Observe(nil, imp)
	other.Observe(nil, imp)
	self.Observe(nil, imp)

	// Express hands the move to motor via the spy kernel.
	self.Express(nil, imp)
	select {
	case call := <-spy.calls:
		if call.traversed == nil || call.traversed.ID() != "skyra:motor" {
			t.Fatalf("self spawned %v, want skyra:motor", call.traversed)
		}
		got, ok := call.impulse.(skyra.Impulse)
		if !ok || got.Payload != "walk michael through the kernel" {
			t.Fatalf("motor did not receive self's move: %+v", call.impulse)
		}
	case <-time.After(time.Second):
		t.Fatal("self did not hand its move to motor")
	}

	// And motor records the move it's handed.
	motor := m.Motor()
	motor.Observe(nil, skyra.Impulse{Payload: "walk michael through the kernel"})
	if motor.LastAction() != "walk michael through the kernel" {
		t.Fatalf("motor.LastAction = %q, want self's move", motor.LastAction())
	}
}

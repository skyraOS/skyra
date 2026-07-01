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
		case strings.Contains(system, "External Perspective"):
			return `<predictions>
  <prediction p="0.70">ask how the runtime is structured</prediction>
  <prediction p="0.30">ask about the kernel</prediction>
</predictions>`, nil
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

	// The grouping carries: an other-pred root node (with distribution children)
	// and a self-pred leaf.
	sawOtherRoot, sawSelf := false, false
	for _, gr := range midLevel.Groupings() {
		mc := gr.(*dna.MemoryCell)
		for _, child := range mc.Children() {
			cc, ok := child.(*dna.MemoryCell)
			if !ok {
				continue
			}
			p, ok := cc.Content().(dna.Prediction)
			if !ok {
				continue
			}
			switch p.Role {
			case "other":
				// this is the root node; check it has distribution children
				sawOtherRoot = true
				if len(cc.Children()) == 0 {
					t.Fatal("other-pred root has no distribution children")
				}
			case "self":
				sawSelf = true
				if p.Text != "walk michael through the kernel" {
					t.Fatalf("self prediction text = %q", p.Text)
				}
			}
		}
	}
	if !sawOtherRoot || !sawSelf {
		t.Fatalf("grouping predictions: otherRoot=%v self=%v (want both)", sawOtherRoot, sawSelf)
	}
}

// MVP forward pass: self hands off to motor (voting parked). The spawn to
// skyra:motor is a bare "act" trigger — the proposal is NOT on the payload; self
// deposited it on the grouping (self-prediction), and motor reads it off the slice.
func TestSelfHandsMoveToMotor(t *testing.T) {
	spy := newSpyKernel()
	mock := func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "External Perspective"):
			return `<predictions>
  <prediction p="0.70">ask how the runtime is structured</prediction>
  <prediction p="0.30">ask about the kernel</prediction>
</predictions>`, nil
		case strings.Contains(system, "predict your own next move"):
			return "<prediction>walk michael through the kernel</prediction>", nil
		case strings.Contains(system, "Skyra's motor"):
			return "<engage><michael>explain the kernel traversal</michael></engage>", nil
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

	// Express fires motor via the spy kernel — a bare "act" trigger, no payload.
	self.Express(nil, imp)
	select {
	case call := <-spy.calls:
		if call.traversed == nil || call.traversed.ID() != "skyra:motor" {
			t.Fatalf("self spawned %v, want skyra:motor", call.traversed)
		}
		got, ok := call.impulse.(skyra.Impulse)
		if !ok || got.Primitive != "act" {
			t.Fatalf("self did not fire an act trigger: %+v", call.impulse)
		}
		if got.Payload != nil {
			t.Fatalf("act trigger should carry no proposal payload, got %+v", got.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("self did not fire motor")
	}

	// Motor reads the proposal off the slice, collapses it into an action via
	// inference, and appends that action to the proposal.
	motor := m.Motor()
	motor.Observe(nil, skyra.Impulse{Primitive: "act"})
	if motor.LastAction() != "<engage><michael>explain the kernel traversal</michael></engage>" {
		t.Fatalf("motor.LastAction = %q, want concrete motor action", motor.LastAction())
	}

	found := false
	for _, gr := range midLevel.Groupings() {
		grouping := gr.(*dna.MemoryCell)
		for _, proposal := range grouping.Children() {
			proposalCell, ok := proposal.(*dna.MemoryCell)
			if !ok {
				continue
			}
			if p, ok := proposalCell.Content().(dna.Prediction); !ok || p.Role != "self" {
				continue
			}
			for _, action := range proposalCell.Children() {
				actionCell, ok := action.(*dna.MemoryCell)
				if !ok || actionCell.Kind() != dna.KindAction {
					continue
				}
				record, ok := actionCell.Content().(dna.ActionRecord)
				if ok && record.ProposalID != "" && record.Primitive == "engage" &&
					record.EntityID == "michael" && record.Content == "explain the kernel traversal" &&
					record.Impulse.Primitive == "engage" && record.Impulse.TargetID == "michael" &&
					record.Impulse.Payload == "explain the kernel traversal" &&
					record.Impulse.Source != nil && record.Impulse.Source.SpaceID == "skyra" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("motor action was not appended to the self proposal")
	}
}

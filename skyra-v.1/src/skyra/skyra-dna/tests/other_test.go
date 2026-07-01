package tests

import (
	"strings"
	"testing"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// The other-agent's predict step: it reads MidLevel's most-recent grouping,
// produces a probability distribution over the other party's next move, and
// appends it onto the grouping as a root node with one child per prediction.
func TestOtherPredictsAndAppends(t *testing.T) {
	mock := func(system, present string) (string, error) {
		if strings.Contains(system, "External Perspective") {
			if !strings.Contains(present, "help michael") {
				t.Errorf("other's prompt missing the framing: %q", present)
			}
			return `<predictions>
  <prediction p="0.60">ask how the runtime is structured</prediction>
  <prediction p="0.40">ask about the kernel</prediction>
</predictions>`, nil
		}
		return "<label>help michael with the skyra project</label>", nil
	}
	m, _ := newBeingWithInfer("skyra", mock)
	sensory := m.Sensory()
	midLevel := sensory.MidLevel()
	other := midLevel.Other()

	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "can you help me with the skyra project",
	})
	imp := skyra.Impulse{Source: &skyra.ImpulseSource{SpaceID: "michael"}, Payload: "now"}
	midLevel.Observe(nil, imp)
	other.Observe(nil, imp)

	// The grouping carries a root node (role=other) with prediction children.
	groupings := midLevel.Groupings()
	if len(groupings) != 1 {
		t.Fatalf("groupings = %d, want 1", len(groupings))
	}
	var root *dna.MemoryCell
	for _, gr := range groupings {
		mc := gr.(*dna.MemoryCell)
		for _, child := range mc.Children() {
			cc, ok := child.(*dna.MemoryCell)
			if !ok {
				continue
			}
			if p, ok := cc.Content().(dna.Prediction); ok && p.Role == "other" {
				root = cc
			}
		}
	}
	if root == nil {
		t.Fatal("grouping has no other-pred root node")
	}

	// Root's children are the individual predictions.
	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("other-pred root has %d children, want 2", len(children))
	}
	texts := map[string]bool{}
	for _, child := range children {
		if p, ok := child.Content().(dna.Prediction); ok {
			texts[p.Text] = true
		}
	}
	if !texts["ask how the runtime is structured"] {
		t.Fatalf("missing expected prediction; got %+v", texts)
	}
}

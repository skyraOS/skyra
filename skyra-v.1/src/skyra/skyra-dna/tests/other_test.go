package tests

import (
	"strings"
	"testing"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// The other-agent's predict step: it reads MidLevel's most-recent grouping,
// predicts the other party's next move, and appends that prediction onto the
// grouping cell (grouping → other-prediction) — where self will later surface it.
func TestOtherPredictsAndAppends(t *testing.T) {
	// One injected inference seam serves both calls; branch on the system prompt:
	// MidLevel asks for a <label>, other asks for a <prediction>.
	mock := func(system, present string) (string, error) {
		if strings.Contains(system, "modeling the other") {
			if !strings.Contains(present, "help michael") {
				t.Errorf("other's prompt missing the framing: %q", present)
			}
			return "<prediction>ask how the runtime is structured</prediction>", nil
		}
		return "<label>help michael with the skyra project</label>", nil
	}
	m, _ := newBeingWithInfer("skyra", mock)
	sensory := m.Sensory()
	midLevel := sensory.MidLevel()
	other := midLevel.Other()

	// Warm sensory, then run MidLevel's Observe (creates the grouping + frame) and
	// other's Observe (predicts + appends) directly — no spawn, deterministic.
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "can you help me with the skyra project",
	})
	imp := skyra.Impulse{Source: &skyra.ImpulseSource{SpaceID: "michael"}, Payload: "now"}
	midLevel.Observe(nil, imp)
	other.Observe(nil, imp)

	// The grouping now carries an other-prediction child.
	groupings := midLevel.Groupings()
	if len(groupings) != 1 {
		t.Fatalf("groupings = %d, want 1", len(groupings))
	}
	found := false
	for _, gr := range groupings {
		mc := gr.(*dna.MemoryCell)
		for _, child := range mc.Children() {
			p, ok := child.Content().(dna.Prediction)
			if !ok {
				continue
			}
			if p.Role != "other" {
				t.Fatalf("prediction role = %q, want other", p.Role)
			}
			if p.Text != "ask how the runtime is structured" {
				t.Fatalf("prediction text = %q", p.Text)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("grouping has no other-prediction child")
	}
}

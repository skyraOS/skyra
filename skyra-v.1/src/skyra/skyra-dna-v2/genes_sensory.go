package dnav2

import (
	"fmt"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// sensory:intake — the entry gene. Resolves the sender into the entity index (the
// handed working set on the nucleus), classifies the impulse, and reports the
// routing condition the orchestrator forks on. Ported from v1 sensory.Observe;
// cold-start turn 1 records an external engage and reports "external".
func sensoryIntake(root *EphemeralRoot, origin skyra.Space, impulse any) {
	imp, ok := asImpulse(impulse)
	if !ok {
		root.Condition = "internal"
		return
	}
	root.Condition = classify(imp, root.DNA().BeingID())
	if root.Condition != "external" {
		debug.Being("skyra", "sensory", fmt.Sprintf("classify=%s (no record)", root.Condition))
		return // reafference/internal record nothing on this lane yet (phase 2)
	}
	sender := resolveEntity(imp.Source.SpaceID)
	entity := root.Entities.Ensure(sender)
	utterance, _ := imp.Payload.(string)
	root.Entities.AppendInteraction(entity, dna.InteractionRecord{
		From:      sender,
		To:        root.DNA().BeingID(),
		Utterance: utterance,
		Payload:   imp.Payload,
		Timestamp: root.DNA().Tick(),
	})
	debug.Being("skyra", "sensory", fmt.Sprintf("classify=external sender=%s utterance=%q", sender, utterance))
}

// classify forks the sensory lane (the three-way fork of the workflow):
//   - reafference: the world's reply to my own action (Return)
//   - external:    another agent's engage
//   - internal:    a reframe re-run / self-initiated trigger
func classify(imp skyra.Impulse, ownerID string) string {
	if imp.Return {
		return "reafference"
	}
	if imp.Primitive == "engage" && imp.Source != nil &&
		imp.Source.SpaceID != "" && imp.Source.SpaceID != ownerID {
		return "external"
	}
	return "internal"
}

// resolveEntity maps a perceived SpaceID to the entity index key (identity for
// now; fuzzy/embedding resolution later).
func resolveEntity(spaceID string) string { return spaceID }

func asImpulse(impulse any) (skyra.Impulse, bool) {
	switch v := impulse.(type) {
	case skyra.Impulse:
		return v, true
	case *skyra.Impulse:
		if v != nil {
			return *v, true
		}
	}
	return skyra.Impulse{}, false
}

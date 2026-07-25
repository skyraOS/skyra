package dnav2

import (
	"fmt"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// sensory:intake — the entry gene. Resolves the sender into the entity index (the
// handed working set on the nucleus), classifies the impulse, and reports the
// routing condition the orchestrator forks on.Observe;
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
	entity := root.Nodes.EnsureEntity(sender)
	utterance, _ := imp.Payload.(string)
	ts := root.DNA().Tick()
	root.Nodes.AppendInteraction(entity, InteractionRecord{
		From:      sender,
		To:        root.DNA().BeingID(),
		Utterance: utterance,
		Payload:   imp.Payload,
		Timestamp: ts,
	})
	if observation := decomposeUtterance(root.DNA(), root.Trace, sender, root.DNA().BeingID(), utterance, ts); observation != nil {
		root.ObservationID = observation.Cell.ID()
		root.Trace.Note("observation", "persisted "+root.ObservationID)
		root.DNA().AssessActiveCauses(root.ObservationID, root.Trace)
	}
	debug.Being("skyra", "sensory", fmt.Sprintf("classify=external sender=%s utterance=%q", sender, utterance))
}

// decomposeUtterance performs the light structural pass at the observation
// boundary. A valid annotated graph commits atomically as an observation-born
// CausalNode; malformed output is returned to the model for bounded correction
// and never falls back to transport-only metadata.
func decomposeUtterance(d *DNA, tr *RunTrace, sender, receiver, utterance string, ts int) *CausalNode {
	infer := tr.Wrap(d.Infer(), "decompose")
	if infer == nil || strings.TrimSpace(utterance) == "" {
		return nil
	}
	present := fmt.Sprintf("Speaker: %s\nReceiver: %s\nUtterance: %s", sender, receiver, utterance)
	raw, err := infer(observationDecompositionSystem, present)
	if err != nil {
		return nil
	}
	parsed, perr := ParseAnnotated(raw)
	for attempt := 1; perr != nil && attempt <= parseMaxRetries; attempt++ {
		tr.Reject("decompose", raw, perr)
		raw, err = infer(observationDecompositionSystem, present+"\n\n---\n"+CorrectionMessage(raw, perr))
		if err != nil {
			return nil
		}
		parsed, perr = ParseAnnotated(raw)
	}
	if perr != nil {
		tr.Reject("decompose", raw, perr)
		debug.Being("skyra", "observation", fmt.Sprintf("decomposition failed after retries: %v", perr))
		return nil
	}
	return d.CausalNodes().PersistObserved(parsed, ts)
}

const observationDecompositionSystem = "You decompose one observed utterance into its most " +
	"important causal structure. Emit exactly one concise annotated sentence and nothing else. " +
	"Stand on the most specific ground firmly supported by the observation; do not diagnose, " +
	"invent motives, or reduce the event to speaker/text transport metadata. Mark every referent " +
	"with an ordinal, such as michael[1] or {the outage}[2]. Mark every relationship on its word, " +
	"such as reported[1→2]; multi-word verb groups take braces, such as {asked about}[1→2] or " +
	"{is testing}[1→2]. Every meaningful referent and relationship must be represented in " +
	"the graph. Plain words may only be grammatical glue."

// classify forks the sensory lane (the workflow's intake fork). A perceive-return
// is a REAFFERENCE (the world answering my own act), NOT an internal event —
// internal is reserved for interrupts (e.g. motor engaging the self). What
// distinguishes it from an action-reafference is only where the record lands:
// perceive-return routes to the sensory ingest (deposit surroundings on the
// spine); move/engage-return routes to efficient:grade (grade my action's execution).
// See perceive-build-uncertainties.md §1.
//   - perceive-return: the world's reply to my own perceive (Return + perceive)
//   - reafference:     the world's reply to my own action (Return)
//   - external:        another agent's engage
//   - internal:        an interrupt / reframe re-run
func classify(imp skyra.Impulse, ownerID string) string {
	if imp.Return {
		if imp.Primitive == "perceive" {
			return "perceive-return"
		}
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

package dnav2

import (
	"fmt"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// other:grade — closes the prior other-distribution against the incoming external
// evidence via the adversarial-neutral judge. No open bet → "no-prior". A branch
// matched → "held". A total miss (the distribution put zero mass on what happened)
// → "all-preds<0", the trigger for grade-driven reframing.// OtherCell.gradePrior; the orchestrator routes the three conditions.
//
// A failed call / parse leaves the bet ungraded — it persists and the next evidence
// wave retries (slight evidence drift, accepted for the MVP), routing as "no-prior"
// so the turn still frames.
func otherGrade(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "no-prior"
	episodes := root.DNA().Episodes()
	distRoot, kids := episodes.LatestUngradedOther()
	if distRoot == nil || len(kids) == 0 || root.DNA().Infer() == nil {
		return
	}
	imp, ok := asImpulse(impulse)
	if !ok {
		return
	}
	evidence, _ := imp.Payload.(string)
	if evidence == "" {
		return
	}

	lines := make([]string, 0, len(kids))
	for i, kid := range kids {
		if p, ok := kid.Content().(Prediction); ok {
			lines = append(lines, fmt.Sprintf("[%d] %s", i+1, p.Text))
		}
	}
	present := fmt.Sprintf("Candidate predictions:\n%s\n\nThe message that actually arrived:\n%s",
		strings.Join(lines, "\n"), evidence)
	resp, err := root.Infer("other:grade")(graderSystem, present)
	if err != nil {
		return // bet stays ungraded; carry on unchanged (no-prior)
	}
	idx, matched, ok := ParseMatch(resp)
	if !ok {
		return
	}

	// Score in CODE from the stored probability — the judge never sees a number.
	score := 0.0
	if matched && idx >= 1 && idx <= len(kids) {
		if p, ok := kids[idx-1].Content().(Prediction); ok {
			score = p.Probability
		}
	} else {
		matched, idx = false, 0
	}
	episodes.StampOtherGrade(distRoot, kids, idx, score, evidence)

	if matched {
		root.Condition = "held"
	} else {
		root.Condition = "all-preds<0" // total miss → reframe
	}
	debug.Being("skyra", "other", fmt.Sprintf("grade → %s score=%.2f evidence=%q", root.Condition, score, evidence))
	// Closure → A is UNWIRED (2026-07-06): the midlevel coherence-settle is the sole writer of
	// A now. Closure (did the bet hold) is a separate, parked axis — StampOtherGrade above still
	// records it on the prediction; it just no longer bumps edge alignment. See active-grouping.md
	// and ActivationStore.DistributeCredit (kept, parked, for when the closure axis is designed).
}

// graderSystem — the adversarial-neutral judge. Never told the predictions are its
// own and shown no probabilities (no anchoring); it only identifies the realized
// branch. Skeptical by instruction: near-misses are not matches.
const graderSystem = "You are a strict, impartial judge. You are given numbered candidate " +
	"predictions of what a message would say, and the message that actually arrived.\n\n" +
	"Identify which single candidate the message realized, if any. A candidate matches " +
	"only if the message actually satisfies what the candidate describes — be skeptical; " +
	"near-misses and partial overlaps are NOT matches.\n\n" +
	"Reply with only one tag: <match>N</match> (the matching candidate's number) or <match>none</match>."

// other:predict — predict the other party's next move under the framing. Ported
// from v1 other.Observe (predict half): read the handed episode, infer the
// calibrated distribution, append it to the episode spine. Reports "predicted".
func otherPredict(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "predicted" // the worker completes the join regardless
	if root.Episode == nil || root.DNA().Infer() == nil {
		return
	}
	present := fmt.Sprintf("Framing: %s\n\nRecent exchange:\n%s",
		frameText(root.Episode), renderExchange(root.Nodes))
	resp, err := root.Infer("other:predict")(otherSystem, present)
	if err != nil {
		return
	}
	preds, ok := ParsePredictions(resp)
	if !ok {
		return
	}
	top := 0.0
	for _, p := range preds {
		if p.Probability > top {
			top = p.Probability
		}
	}
	debug.Being("skyra", "other", fmt.Sprintf("predicted %d branches, top p=%.2f", len(preds), top))
	born := appendOtherDistribution(root.DNA().Episodes(), root.Episode, root.DNA().Tick(), preds)
	root.DNA().CrossBirth(born...) // node birth: the new edges were just active
}

// appendOtherDistribution writes the distribution onto the episode's CURRENT frame
// (the frame-owned spine — a bet belongs to the framing it was made under): a root
// node (frame→other-pred:ts) carrying one child per prediction. Built fully, then
// added under the index lock (self:predict reads it via the same lock). Returns the
// edge keys it created (frame→root + root→each branch) so the caller crosses them.
func appendOtherDistribution(episodes *EpisodeStore, episode *MemoryCell, ts int, preds []ParsedPrediction) []string {
	frame := latestFrame(episode)
	if frame == nil {
		return nil
	}
	rootKey := fmt.Sprintf("%s→other-pred:%d", frame.ID(), ts)
	rootCell := NewMemoryCell(rootKey, KindPrediction, Prediction{Role: "other", Timestamp: ts}, nil)
	born := []string{rootKey}
	for i, p := range preds {
		childKey := fmt.Sprintf("%s→other-pred:%d-%d", frame.ID(), ts, i)
		child := NewMemoryCell(childKey, KindPrediction, Prediction{
			Text: p.Text, Role: "other", Probability: p.Probability, Timestamp: ts,
		}, nil)
		rootCell.AddChild(childKey, child)
		born = append(born, childKey)
	}
	episodes.AddChild(frame, rootKey, rootCell)
	return born
}

const otherSystem = `You are operating in the cognitive frame External Perspective.

Your task is to predict several distinct, plausible futures of the conversation.

Given the framing, recent exchange, and latest user message, estimate the probability of several different observable next messages.

Predictions should:
- Be falsifiable by the very next user message.
- Describe observable properties of that message.
- Be distinct from one another.
- Cover meaningfully different possible futures rather than refinements of the same future.
- Be calibrated so probabilities sum to 1.0.
- Avoid inferring hidden mental states.

Output only:
<predictions>
  <prediction p="0.41">...</prediction>
  ...
</predictions>`

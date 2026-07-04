package dnav2

import (
	"fmt"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// other:grade — grades the prior other-distribution against the incoming external
// evidence. Phase 1 (cold start / no prior bet): reports "no-prior" (no inference).
// Phase 2 ports the adversarial judge → "held" / "all-preds<0".
func otherGrade(root *EphemeralRoot, origin skyra.Space, impulse any) {
	// no grouping / no prior distribution yet → nothing to grade
	root.Condition = "no-prior"
}

// other:predict — predict the other party's next move under the framing. Ported
// from v1 other.Observe (predict half): read the handed grouping, infer the
// calibrated distribution, append it to the grouping spine. Reports "predicted".
func otherPredict(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "predicted" // the worker completes the join regardless
	if root.Grouping == nil || root.DNA().Infer() == nil {
		return
	}
	present := fmt.Sprintf("Framing: %s\n\nRecent exchange:\n%s",
		gistOf(root.Grouping), renderExchange(root.Entities))
	resp, err := root.DNA().Infer()(otherSystem, present)
	if err != nil {
		return
	}
	preds, ok := dna.ParsePredictions(resp)
	if !ok {
		return
	}
	appendOtherDistribution(root.DNA().Groupings(), root.Grouping, root.DNA().Tick(), preds)
}

// appendOtherDistribution writes the distribution onto the grouping: a root node
// (grouping→other-pred:ts) carrying one child per prediction. Built fully, then
// added under the index lock (self:predict reads it via the same lock).
func appendOtherDistribution(gi *GroupingIndex, grouping *dna.MemoryCell, ts int, preds []dna.ParsedPrediction) {
	rootKey := fmt.Sprintf("%s→other-pred:%d", grouping.ID(), ts)
	rootCell := dna.NewMemoryCell(rootKey, dna.KindPrediction, dna.Prediction{Role: "other", Timestamp: ts}, nil)
	for i, p := range preds {
		childKey := fmt.Sprintf("%s→other-pred:%d-%d", grouping.ID(), ts, i)
		child := dna.NewMemoryCell(childKey, dna.KindPrediction, dna.Prediction{
			Text: p.Text, Role: "other", Probability: p.Probability, Timestamp: ts,
		}, nil)
		rootCell.AddChild(childKey, child)
	}
	gi.AddChild(grouping, rootKey, rootCell)
}

// gistOf extracts a grouping's framing text.
func gistOf(g *dna.MemoryCell) string {
	if g == nil {
		return ""
	}
	if gist, ok := g.Content().(dna.Gist); ok {
		return gist.Text
	}
	return ""
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

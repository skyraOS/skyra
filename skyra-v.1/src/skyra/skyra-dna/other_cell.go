package dna

import (
	"fmt"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// OtherCell — the "other" agent: predicts the other party's next move under the
// active framing. Pipeline stage MidLevel → other → self, reached by a spawn and
// owning its own slice. Holds a back-ref to MidLevel (to read the published
// groupings) and the inference seam.
//
// Predict half (built): read the most-recent grouping, predict the other's move,
// append other-prediction onto the grouping. Grade half (pop Prior, close last
// bet) is later. See specs/2026-06-16/skyra-dna/agent.md.
//
// Kept separate from SelfCell on purpose (MVP spread-out, duplicated); they
// collapse to one AgentCell{role} once the shared shape is visible in code.
type OtherCell struct {
	id       string
	slice    skyra.Slice
	kernel   skyra.Time
	midLevel *MidLevelCell
	infer    InferFunc
	self     *SelfCell
}

func NewOtherCell(ctx BuildContext, midLevel *MidLevelCell) *OtherCell {
	o := &OtherCell{
		id:       ctx.BeingID + ":other",
		slice:    ctx.sliceFor(":other"),
		kernel:   ctx.Kernel,
		midLevel: midLevel,
		infer:    ctx.Infer,
	}
	o.self = NewSelfCell(ctx, midLevel) // next stage (other → self; passes the back-ref down)
	return o
}

func (o *OtherCell) ID() string                            { return o.id }
func (o *OtherCell) Slice() skyra.Slice                    { return o.slice }
func (o *OtherCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (o *OtherCell) Content() any                          { return nil }
func (o *OtherCell) Children() map[string]skyra.Space      { return nil }
func (o *OtherCell) Self() *SelfCell                       { return o.self }

func (o *OtherCell) Agents() map[string]skyra.Space {
	if o.self == nil {
		return nil
	}
	return map[string]skyra.Space{o.self.ID(): o.self}
}

// Observe — the predict half: read the most-recent grouping (from MidLevel's
// slice), predict the other's next move under its framing, and append the
// prediction onto the grouping. Direct-read gather for now — cold-start groupings
// have no setpoints to traverse; the wave-from-grouping enrichment lands once the
// recursive-slice handling does.
func (o *OtherCell) Observe(origin skyra.Space, impulse any) {
	if o.midLevel == nil || o.infer == nil {
		return
	}
	groupID, grouping := o.mostRecentGrouping()
	if grouping == nil {
		return
	}
	frame := o.renderPrompt(grouping)
	debug.Being("skyra", "other", fmt.Sprintf("════════════ frame t=%d ════════════\n%s\n════════════════════════════════════", sliceCount(o.midLevel.Slice()), frame))
	debug.Being("skyra", "other", "inference start")
	resp, err := o.infer(otherSystem, frame)
	if err != nil {
		debug.Being("skyra", "other", fmt.Sprintf("inference error: %v", err))
		return
	}
	debug.Being("skyra", "other", "inference end")
	preds, ok := ParsePredictions(resp)
	if !ok {
		return
	}
	o.appendPredictions(groupID, grouping, preds)
	for _, p := range preds {
		debug.Being("skyra", "other", fmt.Sprintf("prediction p=%.2f %q", p.Probability, p.Text))
	}
}

// Express — spawns self (the next stage). [later: grade the prediction against
// the incoming evidence]
func (o *OtherCell) Express(origin skyra.Space, impulse any) any {
	if o.kernel != nil && o.self != nil {
		debug.Being("skyra", "other", "express → self")
		go o.kernel.Realize(o.self, o.self, impulse)
	}
	return nil
}

// otherSystem instructs the other-agent to produce a calibrated probability
// distribution over the other party's next observable message.
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

// mostRecentGrouping reads MidLevel's published groupings (under its slice lock)
// and returns the one whose gist is most recent.
func (o *OtherCell) mostRecentGrouping() (string, skyra.Space) {
	sl := o.midLevel.Slice()
	if sl == nil {
		return "", nil
	}
	sl.Lock()
	defer sl.Unlock()

	var bestID string
	var best skyra.Space
	bestTS := -1
	for id, gr := range o.midLevel.Groupings() {
		mc, ok := gr.(*MemoryCell)
		if !ok {
			continue
		}
		gist, ok := mc.Content().(Gist)
		if !ok {
			continue
		}
		if gist.Timestamp > bestTS {
			bestTS, bestID, best = gist.Timestamp, id, gr
		}
	}
	return bestID, best
}

// renderPrompt assembles the other-agent's present from the grouping's framing
// (gist), its context (setpoints — none at cold-start), and the exchange.
func (o *OtherCell) renderPrompt(grouping skyra.Space) string {
	gist := ""
	if mc, ok := grouping.(*MemoryCell); ok {
		if g, ok := mc.Content().(Gist); ok {
			gist = g.Text
		}
	}
	return fmt.Sprintf("Framing: %s\n\nWorld state:\n%s\n\nRecent exchange:\n%s",
		gist, o.midLevel.WorldState(), renderFrame(o.midLevel.Frame()))
}

// appendPredictions writes the distribution onto the grouping as a root node
// (grouping → other-pred-root) with one child per prediction
// (other-pred-root → other-pred-N). Stable root key so the current cycle's
// distribution overwrites last cycle's. Self reads the root's children to get
// the full spread ("self relies on other").
func (o *OtherCell) appendPredictions(groupID string, grouping skyra.Space, preds []ParsedPrediction) {
	mc, ok := grouping.(*MemoryCell)
	if !ok {
		return
	}
	sl := o.midLevel.Slice()
	if sl != nil {
		sl.Lock()
		defer sl.Unlock()
	}
	ts := sliceCount(sl)
	rootKey := groupID + "→other-pred"
	root := NewMemoryCell(rootKey, KindPrediction, Prediction{Role: "other", Timestamp: ts}, sl)
	for i, p := range preds {
		childKey := fmt.Sprintf("%s→other-pred-%d", groupID, i)
		child := NewMemoryCell(childKey, KindPrediction, Prediction{
			Text:        p.Text,
			Role:        "other",
			Probability: p.Probability,
			Timestamp:   ts,
		}, sl)
		root.AddChild(childKey, child)
	}
	mc.AddChild(rootKey, root)
}

// Prediction is an agent's predicted next move under the active framing. Role
// distinguishes "other" (their move, with a calibrated Probability) from "self"
// (skyra's move, a point estimate — Probability is 0).
type Prediction struct {
	Text        string
	Role        string
	Probability float64
	Timestamp   int
}

func sliceCount(sl skyra.Slice) int {
	if sl == nil {
		return 0
	}
	return sl.Count()
}

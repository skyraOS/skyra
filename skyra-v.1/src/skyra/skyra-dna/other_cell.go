package dna

import (
	"fmt"

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
	resp, err := o.infer(otherSystem, o.renderPrompt(grouping))
	if err != nil {
		return
	}
	if text, ok := ParsePrediction(resp); ok {
		o.appendPrediction(groupID, grouping, text)
	}
}

// Express — spawns self (the next stage). [later: grade the prediction against
// the incoming evidence]
func (o *OtherCell) Express(origin skyra.Space, impulse any) any {
	if o.kernel != nil && o.self != nil {
		go o.kernel.Realize(o.self, o.self, impulse)
	}
	return nil
}

// otherSystem instructs the other-agent to predict the OTHER party's next move.
const otherSystem = "You are skyra, modeling the other party in a live exchange. " +
	"Under the given framing, predict the other's next move — what they will say or do. " +
	"Reply with only the tag: <prediction>their move</prediction>"

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
	return fmt.Sprintf("Framing: %s\n\nRecent exchange:\n%s", gist, renderFrame(o.midLevel.Frame()))
}

// appendPrediction stamps the other's predicted move onto the grouping cell as a
// child (grouping → other-prediction), under MidLevel's lock — so self surfaces it
// when it traverses the grouping ("self relies on other"). Stable key, so the
// current prediction overwrites last cycle's.
func (o *OtherCell) appendPrediction(groupID string, grouping skyra.Space, text string) {
	mc, ok := grouping.(*MemoryCell)
	if !ok {
		return
	}
	sl := o.midLevel.Slice()
	if sl != nil {
		sl.Lock()
		defer sl.Unlock()
	}
	key := groupID + "→other-pred"
	pred := NewMemoryCell(key, KindPrediction, Prediction{Text: text, Role: "other", Timestamp: sliceCount(sl)}, sl)
	mc.AddChild(key, pred)
}

// Prediction is an agent's predicted next move under the active framing — the open
// bet, graded against what happens next cycle. Role distinguishes "other" (their
// move) from "self" (my move).
type Prediction struct {
	Text      string
	Role      string
	Timestamp int
}

func sliceCount(sl skyra.Slice) int {
	if sl == nil {
		return 0
	}
	return sl.Count()
}

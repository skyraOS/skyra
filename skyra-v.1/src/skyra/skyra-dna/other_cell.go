package dna

import (
	"fmt"
	"sort"
	"strings"

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
	ownerID  string
	slice    skyra.Slice
	kernel   skyra.Time
	midLevel *MidLevelCell
	infer    InferFunc
	self     *SelfCell
}

func NewOtherCell(ctx BuildContext, midLevel *MidLevelCell) *OtherCell {
	o := &OtherCell{
		id:       ctx.BeingID + ":other",
		ownerID:  ctx.BeingID,
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

// Observe — grade half then predict half. Pop the prior bet, grade it against
// the evidence that just arrived, then bet again.
//
// Grade half (Stage A): fires only on an EXTERNAL ENGAGE — other predicts the
// external; self-initiated triggers (move/perceive returns) are reafference,
// self's lane. The grader is a separate, adversarial-neutral call: never told
// the bets are its own, shown no probabilities; it only identifies the realized
// branch, and the score is assigned by code from the stored probability.
//
// Predict half: read the most-recent grouping (from MidLevel's slice), predict
// the other's next move under its framing, and append the prediction onto the
// grouping. Direct-read gather for now — cold-start groupings have no setpoints
// to traverse; the wave-from-grouping enrichment lands once the recursive-slice
// handling does.
func (o *OtherCell) Observe(origin skyra.Space, impulse any) {
	if o.midLevel == nil || o.infer == nil {
		return
	}
	if imp, ok := asImpulse(impulse); ok && externalEngage(imp, o.ownerID) {
		if evidence, ok := imp.Payload.(string); ok && evidence != "" {
			o.gradePrior(evidence)
		}
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
// (grouping → other-pred:ts) with one child per prediction
// (other-pred:ts → other-pred:ts-N). PER-CYCLE keys — graded bets persist as
// closure candidates / history (no stable-key overwrite; decay is the only
// forgetting). Self reads the LATEST root's children ("self relies on other").
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
	rootKey := fmt.Sprintf("%s→other-pred:%d", groupID, ts)
	root := NewMemoryCell(rootKey, KindPrediction, Prediction{Role: "other", Timestamp: ts}, sl)
	for i, p := range preds {
		childKey := fmt.Sprintf("%s→other-pred:%d-%d", groupID, ts, i)
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
	// Grading (Stage A) — set once by the grade half; graded bets persist.
	// For an other-distribution root, Grade is the probability mass the bet put
	// on the branch that was realized (0 on a full miss) — the proper score.
	// For a branch child, Grade is 1 (realized) or 0. For a self proposal,
	// Grade is the execution verdict (1 / 0.5 / 0).
	Graded   bool
	Grade    float64
	Evidence string
}

// externalEngage reports whether an impulse is another agent's utterance —
// the only evidence that tests an other-prediction (the grade gate's MVP form).
func externalEngage(imp skyra.Impulse, ownerID string) bool {
	return !imp.Return && imp.Primitive == "engage" && imp.Source != nil &&
		imp.Source.SpaceID != "" && imp.Source.SpaceID != ownerID
}

// graderSystem — the adversarial-neutral judge. It is never told the
// predictions are its own and sees no probabilities (no anchoring); it only
// identifies. Skeptical by instruction: near-misses are not matches.
const graderSystem = "You are a strict, impartial judge. You are given numbered candidate " +
	"predictions of what a message would say, and the message that actually arrived.\n\n" +
	"Identify which single candidate the message realized, if any. A candidate matches " +
	"only if the message actually satisfies what the candidate describes — be skeptical; " +
	"near-misses and partial overlaps are NOT matches.\n\n" +
	"Reply with only one tag: <match>N</match> (the matching candidate's number) or <match>none</match>."

// gradePrior grades the most recent ungraded other-distribution against the
// utterance that just arrived. Snapshot under MidLevel's lock → infer with the
// lock RELEASED (the call can be slow) → write back under the lock. A failed
// call or parse leaves the bet ungraded — it persists, and the next evidence
// wave retries (slight evidence drift, accepted for the MVP).
func (o *OtherCell) gradePrior(evidence string) {
	rootKey, root, kidKeys, kids := o.latestUngradedDistribution()
	if root == nil || len(kids) == 0 {
		return
	}

	var lines []string
	for i, kid := range kids {
		if p, ok := kid.Content().(Prediction); ok {
			lines = append(lines, fmt.Sprintf("[%d] %s", i+1, p.Text))
		}
	}
	present := fmt.Sprintf("Candidate predictions:\n%s\n\nThe message that actually arrived:\n%s",
		strings.Join(lines, "\n"), evidence)

	debug.Being("skyra", "other", "grade start (judge)")
	resp, err := o.infer(graderSystem, present)
	if err != nil {
		debug.Being("skyra", "other", fmt.Sprintf("grade inference error: %v", err))
		return
	}
	idx, matched, ok := ParseMatch(resp)
	if !ok {
		debug.Being("skyra", "other", fmt.Sprintf("grade parse failed: %q", resp))
		return
	}

	// Score in CODE from the stored probability — the judge never sees or
	// reports a number.
	score := 0.0
	if matched && idx >= 1 && idx <= len(kids) {
		if p, ok := kids[idx-1].Content().(Prediction); ok {
			score = p.Probability
		}
	} else {
		matched = false
	}

	sl := o.midLevel.Slice()
	sl.Lock()
	defer sl.Unlock()
	if p, ok := root.Content().(Prediction); ok {
		p.Graded, p.Grade, p.Evidence = true, score, evidence
		root.SetContent(p)
	}
	for i, kid := range kids {
		if p, ok := kid.Content().(Prediction); ok {
			p.Graded, p.Evidence = true, evidence
			if matched && i == idx-1 {
				p.Grade = 1
			}
			kid.SetContent(p)
		}
		writeAlignment(sl, kidKeys[i], boolToGrade(matched && i == idx-1))
	}
	// The root's edge carries the bet's proper score — the lagged Alignment the
	// activation economy (Stage D) consumes. (Framing attribution lands with
	// Stage C, when framings are nodes.)
	writeAlignment(sl, rootKey, score)
	debug.Being("skyra", "other", fmt.Sprintf("graded p=%.2f matched=%v evidence=%q", score, matched, evidence))
}

// latestUngradedDistribution finds the newest ungraded other-pred root across
// the groupings, under MidLevel's lock. Branch children are returned sorted by
// key so prompt numbering is stable.
func (o *OtherCell) latestUngradedDistribution() (string, *MemoryCell, []string, []*MemoryCell) {
	sl := o.midLevel.Slice()
	if sl == nil {
		return "", nil, nil, nil
	}
	sl.Lock()
	defer sl.Unlock()

	var bestKey string
	var best *MemoryCell
	bestTS := -1
	for _, gr := range o.midLevel.Groupings() {
		mc, ok := gr.(*MemoryCell)
		if !ok {
			continue
		}
		for key, child := range mc.Children() {
			cc, ok := child.(*MemoryCell)
			if !ok {
				continue
			}
			p, ok := cc.Content().(Prediction)
			if !ok || p.Role != "other" || p.Graded || len(cc.Children()) == 0 {
				continue
			}
			if p.Timestamp > bestTS {
				bestTS, bestKey, best = p.Timestamp, key, cc
			}
		}
	}
	if best == nil {
		return "", nil, nil, nil
	}
	kidKeys := make([]string, 0, len(best.Children()))
	for key := range best.Children() {
		kidKeys = append(kidKeys, key)
	}
	sort.Strings(kidKeys)
	kids := make([]*MemoryCell, 0, len(kidKeys))
	for _, key := range kidKeys {
		kids = append(kids, best.Children()[key].(*MemoryCell))
	}
	return bestKey, best, kidKeys, kids
}

// writeAlignment writes a grade onto an edge's activation record (creating the
// record if the kernel hasn't scored that edge yet) — the lagged Alignment
// input the bump equation reads next traversal.
func writeAlignment(sl skyra.Slice, edgeKey string, grade float64) {
	if sl == nil {
		return
	}
	rec, ok := sl.Records()[edgeKey].(*memoryActivation)
	if !ok {
		rec = newMemoryActivation()
		sl.Records()[edgeKey] = rec
	}
	rec.Alignment = grade
}

func boolToGrade(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func sliceCount(sl skyra.Slice) int {
	if sl == nil {
		return 0
	}
	return sl.Count()
}

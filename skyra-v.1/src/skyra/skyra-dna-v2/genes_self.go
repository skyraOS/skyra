package dnav2

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// self:grade — closes the prior self proposal against the reafference (the world's
// reply to skyra's own action) via the execution judge. exec grade > 0 → "exec-ok".
// A hard execution failure (judge said `no`, score 0 — the intent could not be
// realized, e.g. a blocked move) → "exec=no", which reframes: an unrealizable intent
// is a framing problem, not just a bad action. Ported from v1 SelfCell.gradePrior.
//
// No realized proposal / failed call / parse leaves it ungraded and reports
// "exec-ok" (carry on unchanged — the reafference lane still frames).
func selfGrade(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "exec-ok"
	gi := root.DNA().Groupings()
	proposal, pred, action, ok := gi.LatestUngradedRealizedSelf()
	if !ok || root.DNA().Infer() == nil {
		return
	}
	result := "(no result reported)"
	if imp, ok := asImpulse(impulse); ok {
		if r, ok := imp.Payload.(string); ok && r != "" {
			result = r
		}
	}
	present := fmt.Sprintf("Instruction:\n%s\n\nAction taken:\n%s %s %v\n\nEnvironment result:\n%s",
		pred.Text, action.Primitive, action.EntityID, action.Content, result)
	resp, err := root.DNA().Infer()(execGraderSystem, present)
	if err != nil {
		return
	}
	score, parsed := dna.ParseExecGrade(resp)
	if !parsed {
		return
	}
	gi.StampSelfGrade(proposal, score, result)
	if score == 0 {
		root.Condition = "exec=no" // hard failure → reframe
	}
	distributeClosure(root, score > 0) // Phase 2e: credit the frame on the reafference outcome
}

const selfFrameBudget = 2000 // generous — don't prune self's conditioning in normal operation

// execGraderSystem — the execution judge. Adversarial-neutral, same posture as the
// other-side grader: never told whose instruction it is, only compares instruction
// to what was actually done. Verbatim from v1.
const execGraderSystem = "You are a strict, impartial judge of execution. You are given an " +
	"instruction that was issued, the action that was actually taken, and any result " +
	"reported by the environment.\n\n" +
	"Decide whether the action faithfully executed the instruction. Be skeptical: doing " +
	"something related but different is NOT faithful execution; a blocked or failed result " +
	"means the instruction was not accomplished (at best partial).\n\n" +
	"Reply with only one tag: <grade>yes</grade>, <grade>partial</grade>, or <grade>no</grade>."

// self:predict — decide skyra's own next move under the framing, conditioned on
// what the OTHER is expected to do ("self relies on other" — read live/lagged off
// the grouping). Ported from v1 self.Observe (predict half). Reports "predicted".
func selfPredict(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "predicted"
	if root.Grouping == nil || root.DNA().Infer() == nil {
		return
	}
	// Phase 2e: the framing + other-distribution present is now TRAVERSAL-assembled off
	// the grouping (relevance = alignment, budget-bounded) and the traversed edges are
	// crossed — instead of the hand-built renderOtherDistribution. The recent exchange
	// (entity interactions) is still hand-rendered until entities join the traversal.
	frame := FramePresentCrossing(root.Grouping, selfFrameBudget, root.DNA().Activation(), root.DNA().Tick())
	present := fmt.Sprintf("%s\n\nRecent exchange:\n%s", frame, renderExchange(root.Entities))
	resp, err := root.DNA().Infer()(selfSystem, present)
	if err != nil {
		return
	}
	text, ok := dna.ParsePrediction(resp)
	if !ok {
		return
	}
	born := appendSelfPrediction(root.DNA().Groupings(), root.Grouping, root.DNA().Tick(), text)
	root.DNA().CrossBirth(born) // node birth: the new edge was just active
}

// renderOtherDistribution reads the latest other-pred root off the grouping (under
// the index lock) and lists its branches by probability. "(no prediction)" if none
// (cold-start, or other hasn't landed yet — the lagged read).
func renderOtherDistribution(gi *GroupingIndex, grouping *dna.MemoryCell) string {
	var latest *dna.MemoryCell
	latestTS := -1
	for _, child := range gi.Snapshot(grouping) {
		rootCell, ok := child.(*dna.MemoryCell)
		if !ok {
			continue
		}
		p, ok := rootCell.Content().(dna.Prediction)
		if !ok || p.Role != "other" || len(rootCell.Children()) == 0 {
			continue
		}
		if p.Timestamp > latestTS {
			latestTS, latest = p.Timestamp, rootCell
		}
	}
	if latest == nil {
		return "(no prediction)"
	}
	var preds []dna.Prediction
	for _, gc := range latest.Children() { // immutable once added — lock-free
		if pp, ok := gc.(*dna.MemoryCell); ok {
			if pred, ok := pp.Content().(dna.Prediction); ok {
				preds = append(preds, pred)
			}
		}
	}
	sort.Slice(preds, func(i, j int) bool { return preds[i].Probability > preds[j].Probability })
	var lines []string
	for _, pp := range preds {
		lines = append(lines, fmt.Sprintf("  p=%.2f  %s", pp.Probability, pp.Text))
	}
	if len(lines) == 0 {
		return "(no prediction)"
	}
	return strings.Join(lines, "\n")
}

// appendSelfPrediction stamps skyra's proposed move onto the grouping (per-cycle
// key). motor reads the latest self proposal off the spine. Returns the birth edge key.
func appendSelfPrediction(gi *GroupingIndex, grouping *dna.MemoryCell, ts int, text string) string {
	key := fmt.Sprintf("%s→self-pred:%d", grouping.ID(), ts)
	pred := dna.NewMemoryCell(key, dna.KindPrediction, dna.Prediction{Text: text, Role: "self", Timestamp: ts}, nil)
	gi.AddChild(grouping, key, pred)
	return key
}

const selfSystem = "You are Skyra's inner voice. Given the current framing and what the other person is likely to do next, decide what action Skyra should take.\n\n" +
	"Write your decision as a clear instruction to the motor system — tell it exactly what to do.\n\n" +
	"Examples:\n" +
	"- Tell michael his coordinate is 500,500\n" +
	"- Ask michael what he wants to do next\n" +
	"- Move toward michael\n" +
	"- Perceive the surroundings to find my position\n\n" +
	"Reply with only the tag: <prediction>instruction for motor</prediction>"

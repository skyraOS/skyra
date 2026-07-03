package dna

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// SelfCell — the "self" agent: predicts the being's own next move under the active
// framing. Runs after other ("self relies on other"). Pipeline stage other → self
// → voting, reached by a spawn and owning its own slice. Holds a back-ref to
// MidLevel (to read the published groupings) and the inference seam.
//
// Predict half (built): read the most-recent grouping — which now carries other's
// prediction as a child — predict skyra's move, append self-prediction. Grade half
// is later. See specs/2026-06-16/skyra-dna/agent.md.
//
// Deliberately a near-copy of OtherCell (MVP spread-out, duplicated); the two
// collapse to one AgentCell{role} once the shared shape is visible in code.
type SelfCell struct {
	id       string
	slice    skyra.Slice
	kernel   skyra.Time
	midLevel *MidLevelCell
	infer    InferFunc
	motor    *MotorCell // MVP: self → motor directly (voting is parked)
}

func NewSelfCell(ctx BuildContext, midLevel *MidLevelCell) *SelfCell {
	s := &SelfCell{
		id:       ctx.BeingID + ":self",
		slice:    ctx.sliceFor(":self"),
		kernel:   ctx.Kernel,
		midLevel: midLevel,
		infer:    ctx.Infer,
		motor:    ctx.Motor, // MVP: bypass voting, hand self's move straight to the out-face
	}
	return s
}

func (s *SelfCell) ID() string                            { return s.id }
func (s *SelfCell) Slice() skyra.Slice                    { return s.slice }
func (s *SelfCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (s *SelfCell) Content() any                          { return nil }
func (s *SelfCell) Children() map[string]skyra.Space      { return nil }
func (s *SelfCell) Motor() *MotorCell                     { return s.motor }

func (s *SelfCell) Agents() map[string]skyra.Space {
	if s.motor == nil {
		return nil
	}
	return map[string]skyra.Space{s.motor.ID(): s.motor}
}

// Observe — grade half then predict half.
//
// Grade half (Stage A): self grades its prior proposal against EXECUTION only —
// the action that realized it (ProposalID pairing) plus the world's reported
// result (reafference). The other's reaction is other's lane. Destination:
// motor running autonomously with predictions as guardrails — execution falling
// outside the predicted envelope is the interrupt; this grade is that
// mechanism's foundation.
//
// Predict half: read the most-recent grouping (carrying other's prediction),
// predict skyra's own next move under its framing, and append the prediction
// onto the grouping.
func (s *SelfCell) Observe(origin skyra.Space, impulse any) {
	if s.midLevel == nil || s.infer == nil {
		return
	}
	s.gradePrior()
	groupID, grouping := s.mostRecentGrouping()
	if grouping == nil {
		return
	}
	frame := s.renderPrompt(grouping)
	debug.Being("skyra", "self", fmt.Sprintf("════════════ frame t=%d ════════════\n%s\n════════════════════════════════════", sliceCount(s.midLevel.Slice()), frame))
	debug.Being("skyra", "self", "inference start")
	resp, err := s.infer(selfSystem, frame)
	if err != nil {
		debug.Being("skyra", "self", fmt.Sprintf("inference error: %v", err))
		return
	}
	debug.Being("skyra", "self", "inference end")
	if text, ok := ParsePrediction(resp); ok {
		s.appendPrediction(groupID, grouping, text)
		debug.Being("skyra", "self", "prediction="+text)
	}
}

// Express — fire motor (the out-face). The proposal is NOT carried on the impulse:
// self already deposited it on the grouping (appendPrediction, in Observe), so motor
// reads it off the slice. The spawn is a bare trigger. MVP: bypasses voting entirely.
// [later: voting tallies self/other; grade against reafference]
func (s *SelfCell) Express(origin skyra.Space, impulse any) any {
	if s.kernel != nil && s.motor != nil {
		debug.Being("skyra", "self", "express → motor")
		go s.kernel.Realize(s.motor, s.motor, skyra.Impulse{Primitive: "act"})
	}
	return nil
}

// selfSystem instructs the self-agent to produce an action instruction for motor —
// what Skyra should DO next, not an observation.
const selfSystem = "You are Skyra's inner voice. Given the current framing and what the other person is likely to do next, decide what action Skyra should take.\n\n" +
	"Write your decision as a clear instruction to the motor system — tell it exactly what to do.\n\n" +
	"Examples:\n" +
	"- Tell michael his coordinate is 500,500\n" +
	"- Ask michael what he wants to do next\n" +
	"- Move toward michael\n" +
	"- Perceive the surroundings to find my position\n\n" +
	"Reply with only the tag: <prediction>instruction for motor</prediction>"

// mostRecentGrouping reads MidLevel's published groupings (under its slice lock)
// and returns the one whose gist is most recent. (Duplicate of OtherCell's —
// collapses with it later.)
func (s *SelfCell) mostRecentGrouping() (string, skyra.Space) {
	sl := s.midLevel.Slice()
	if sl == nil {
		return "", nil
	}
	sl.Lock()
	defer sl.Unlock()

	var bestID string
	var best skyra.Space
	bestTS := -1
	for id, gr := range s.midLevel.Groupings() {
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

// renderPrompt assembles the self-agent's present: the framing (gist), what the
// OTHER is expected to do (other's prediction — surfaced off the grouping, the
// "self relies on other"), and the exchange.
func (s *SelfCell) renderPrompt(grouping skyra.Space) string {
	gist := ""
	var distLines []string
	if mc, ok := grouping.(*MemoryCell); ok {
		if g, ok := mc.Content().(Gist); ok {
			gist = g.Text
		}
		// find the LATEST other-pred root (per-cycle keys mean graded history
		// persists — read this cycle's bet, not an old one), then its children
		// sorted by probability
		var latest *MemoryCell
		latestTS := -1
		for _, child := range mc.Children() {
			root, ok := child.(*MemoryCell)
			if !ok {
				continue
			}
			p, ok := root.Content().(Prediction)
			if !ok || p.Role != "other" || len(root.Children()) == 0 {
				continue
			}
			if p.Timestamp > latestTS {
				latestTS, latest = p.Timestamp, root
			}
		}
		if latest != nil {
			var preds []Prediction
			for _, grandchild := range latest.Children() {
				if pp, ok := grandchild.Content().(Prediction); ok && pp.Role == "other" {
					preds = append(preds, pp)
				}
			}
			sort.Slice(preds, func(i, j int) bool {
				return preds[i].Probability > preds[j].Probability
			})
			for _, pp := range preds {
				distLines = append(distLines, fmt.Sprintf("  p=%.2f  %s", pp.Probability, pp.Text))
			}
		}
	}
	otherDist := strings.Join(distLines, "\n")
	if otherDist == "" {
		otherDist = "(no prediction)"
	}
	return fmt.Sprintf("Framing: %s\n\nWorld state:\n%s\n\nOther's predicted next moves (probability distribution):\n%s\n\nRecent exchange:\n%s",
		gist, s.midLevel.WorldState(), otherDist, renderFrame(s.midLevel.Frame()))
}

// appendPrediction stamps skyra's predicted move onto the grouping cell as a child
// (grouping → self-pred:ts), under MidLevel's lock. PER-CYCLE key — graded
// proposals persist as history/closure candidates; motor reads latest-wins.
func (s *SelfCell) appendPrediction(groupID string, grouping skyra.Space, text string) {
	mc, ok := grouping.(*MemoryCell)
	if !ok {
		return
	}
	sl := s.midLevel.Slice()
	if sl != nil {
		sl.Lock()
		defer sl.Unlock()
	}
	ts := sliceCount(sl)
	key := fmt.Sprintf("%s→self-pred:%d", groupID, ts)
	pred := NewMemoryCell(key, KindPrediction, Prediction{Text: text, Role: "self", Timestamp: ts}, sl)
	mc.AddChild(key, pred)
}

// execGraderSystem — the execution judge. Adversarial-neutral, same posture as
// the other-side grader: never told whose instruction it is, only compares
// instruction to what was actually done.
const execGraderSystem = "You are a strict, impartial judge of execution. You are given an " +
	"instruction that was issued, the action that was actually taken, and any result " +
	"reported by the environment.\n\n" +
	"Decide whether the action faithfully executed the instruction. Be skeptical: doing " +
	"something related but different is NOT faithful execution; a blocked or failed result " +
	"means the instruction was not accomplished (at best partial).\n\n" +
	"Reply with only one tag: <grade>yes</grade>, <grade>partial</grade>, or <grade>no</grade>."

// gradePrior grades the newest ungraded self proposal that has been REALIZED
// (has an action child — the ProposalID pairing made literal). Evidence =
// the action + any world-reported result recorded after it. Snapshot under
// MidLevel's lock, infer unlocked, write back under the lock. Unrealized
// proposals are left alone (superseded ones simply decay — dropping a stale
// intent is not a failure).
func (s *SelfCell) gradePrior() {
	key, proposal, pred, action, ok := s.latestUngradedRealized()
	if !ok {
		return
	}

	result := s.worldResultAfter(action.Timestamp)
	if result == "" {
		result = "(no result reported)"
	}
	present := fmt.Sprintf("Instruction:\n%s\n\nAction taken:\n%s %s %v\n\nEnvironment result:\n%s",
		pred.Text, action.Primitive, action.EntityID, action.Content, result)

	debug.Being("skyra", "self", "grade start (execution judge)")
	resp, err := s.infer(execGraderSystem, present)
	if err != nil {
		debug.Being("skyra", "self", fmt.Sprintf("grade inference error: %v", err))
		return
	}
	score, parsed := ParseExecGrade(resp)
	if !parsed {
		debug.Being("skyra", "self", fmt.Sprintf("grade parse failed: %q", resp))
		return
	}

	sl := s.midLevel.Slice()
	sl.Lock()
	defer sl.Unlock()
	pred.Graded, pred.Grade, pred.Evidence = true, score, result
	proposal.SetContent(pred)
	writeAlignment(sl, key, score)
	debug.Being("skyra", "self", fmt.Sprintf("graded execution=%.1f proposal=%q", score, pred.Text))
}

// latestUngradedRealized finds the newest ungraded self proposal that has an
// action child, under MidLevel's lock. Returns the proposal's edge key, the
// cells, and the latest action record on it.
func (s *SelfCell) latestUngradedRealized() (string, *MemoryCell, Prediction, ActionRecord, bool) {
	sl := s.midLevel.Slice()
	if sl == nil {
		return "", nil, Prediction{}, ActionRecord{}, false
	}
	sl.Lock()
	defer sl.Unlock()

	var bestKey string
	var best *MemoryCell
	var bestPred Prediction
	bestTS := -1
	for _, gr := range s.midLevel.Groupings() {
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
			if !ok || p.Role != "self" || p.Graded || len(cc.Children()) == 0 {
				continue
			}
			if p.Timestamp > bestTS {
				bestTS, bestKey, best, bestPred = p.Timestamp, key, cc, p
			}
		}
	}
	if best == nil {
		return "", nil, Prediction{}, ActionRecord{}, false
	}
	var action ActionRecord
	actionTS := -1
	for _, child := range best.Children() {
		cc, ok := child.(*MemoryCell)
		if !ok || cc.Kind() != KindAction {
			continue
		}
		if rec, ok := cc.Content().(ActionRecord); ok && rec.Timestamp > actionTS {
			actionTS, action = rec.Timestamp, rec
		}
	}
	if actionTS < 0 {
		return "", nil, Prediction{}, ActionRecord{}, false
	}
	return bestKey, best, bestPred, action, true
}

// worldResultAfter reads the world's reported self-events (moved / blocked /
// bumped) recorded at-or-after the action fired — the reafference evidence.
// Reads sensory under ITS lock (lock order agent → sensory; nothing locks self).
func (s *SelfCell) worldResultAfter(actionTS int) string {
	sensory := s.midLevel.Sensory()
	if sensory == nil || sensory.Slice() == nil {
		return ""
	}
	sl := sensory.Slice()
	sl.Lock()
	defer sl.Unlock()

	var lines []string
	self, ok := sensory.ByEntity[s.midLevel.Sensory().ownerID]
	if !ok {
		return ""
	}
	mc, ok := self.(*MemoryCell)
	if !ok {
		return ""
	}
	for _, child := range mc.Children() {
		cc, ok := child.(*MemoryCell)
		if !ok {
			continue
		}
		if rec, ok := cc.Content().(InteractionRecord); ok && rec.Timestamp >= actionTS {
			lines = append(lines, rec.Utterance)
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

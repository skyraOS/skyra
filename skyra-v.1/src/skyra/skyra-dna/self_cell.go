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

// Observe — the predict half: read the most-recent grouping (carrying other's
// prediction), predict skyra's own next move under its framing, and append the
// prediction onto the grouping.
func (s *SelfCell) Observe(origin skyra.Space, impulse any) {
	if s.midLevel == nil || s.infer == nil {
		return
	}
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
		// find the other-pred root node, read its children, sort by probability
		for _, child := range mc.Children() {
			root, ok := child.(*MemoryCell)
			if !ok {
				continue
			}
			p, ok := root.Content().(Prediction)
			if !ok || p.Role != "other" {
				continue
			}
			var preds []Prediction
			for _, grandchild := range root.Children() {
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
// (grouping → self-prediction), under MidLevel's lock. Stable key, overwrites last
// cycle's.
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
	key := groupID + "→self-pred"
	pred := NewMemoryCell(key, KindPrediction, Prediction{Text: text, Role: "self", Timestamp: sliceCount(sl)}, sl)
	mc.AddChild(key, pred)
}

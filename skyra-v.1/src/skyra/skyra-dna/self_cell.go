package dna

import (
	"fmt"

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
	id         string
	slice      skyra.Slice
	kernel     skyra.Time
	midLevel   *MidLevelCell
	infer      InferFunc
	motor      *MotorCell // MVP: self → motor directly (voting is parked)
	prediction string     // skyra's latest predicted move, handed to motor on Express
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
	resp, err := s.infer(selfSystem, s.renderPrompt(grouping))
	if err != nil {
		return
	}
	if text, ok := ParsePrediction(resp); ok {
		s.prediction = text
		s.appendPrediction(groupID, grouping, text)
	}
}

// Express — hands self's predicted move to motor (the out-face). MVP: bypasses
// voting entirely. The move rides the impulse payload. [later: voting tallies
// self/other; grade against reafference]
func (s *SelfCell) Express(origin skyra.Space, impulse any) any {
	if s.kernel != nil && s.motor != nil {
		go s.kernel.Realize(s.motor, s.motor, skyra.Impulse{Primitive: "act", Payload: s.prediction})
	}
	return nil
}

// selfSystem instructs the self-agent to predict skyra's OWN next move, given
// what the other is expected to do.
const selfSystem = "You are skyra in a live exchange. Under the given framing, and " +
	"knowing what the other is expected to do, predict your own next move — what you'll " +
	"say or do to serve the framing. Reply with only the tag: <prediction>your move</prediction>"

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
	gist, otherPred := "", ""
	if mc, ok := grouping.(*MemoryCell); ok {
		if g, ok := mc.Content().(Gist); ok {
			gist = g.Text
		}
		for _, child := range mc.Children() {
			if p, ok := child.Content().(Prediction); ok && p.Role == "other" {
				otherPred = p.Text
			}
		}
	}
	return fmt.Sprintf("Framing: %s\n\nThe other is expected to: %s\n\nRecent exchange:\n%s",
		gist, otherPred, renderFrame(s.midLevel.Frame()))
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

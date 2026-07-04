package dnav2

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// self:predict — decide skyra's own next move under the framing, conditioned on
// what the OTHER is expected to do ("self relies on other" — read live/lagged off
// the grouping). Ported from v1 self.Observe (predict half). Reports "predicted".
func selfPredict(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "predicted"
	if root.Grouping == nil || root.DNA().Infer() == nil {
		return
	}
	present := fmt.Sprintf("Framing: %s\n\nOther's predicted next moves (probability distribution):\n%s\n\nRecent exchange:\n%s",
		gistOf(root.Grouping),
		renderOtherDistribution(root.DNA().Groupings(), root.Grouping),
		renderExchange(root.Entities))
	resp, err := root.DNA().Infer()(selfSystem, present)
	if err != nil {
		return
	}
	text, ok := dna.ParsePrediction(resp)
	if !ok {
		return
	}
	appendSelfPrediction(root.DNA().Groupings(), root.Grouping, root.DNA().Tick(), text)
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
// key). motor reads the latest self proposal off the spine.
func appendSelfPrediction(gi *GroupingIndex, grouping *dna.MemoryCell, ts int, text string) {
	key := fmt.Sprintf("%s→self-pred:%d", grouping.ID(), ts)
	pred := dna.NewMemoryCell(key, dna.KindPrediction, dna.Prediction{Text: text, Role: "self", Timestamp: ts}, nil)
	gi.AddChild(grouping, key, pred)
}

const selfSystem = "You are Skyra's inner voice. Given the current framing and what the other person is likely to do next, decide what action Skyra should take.\n\n" +
	"Write your decision as a clear instruction to the motor system — tell it exactly what to do.\n\n" +
	"Examples:\n" +
	"- Tell michael his coordinate is 500,500\n" +
	"- Ask michael what he wants to do next\n" +
	"- Move toward michael\n" +
	"- Perceive the surroundings to find my position\n\n" +
	"Reply with only the tag: <prediction>instruction for motor</prediction>"

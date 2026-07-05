package dnav2

import (
	"fmt"
	"strings"

	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/traverse"
)

// Phase 1 (traversal-implementation-plan.md): assemble the "present" for inference by
// TRAVERSING the grouping subgraph under a token budget — the analogue of the
// hand-built renderers (renderOtherDistribution / renderExchange) that we A/B against.
// Weights come from the predictions' existing fields (Probability); there is NO learning
// yet. The point is to test whether a traversed frame holds up, not the economy.

// tokenApprox is a rough chars-per-token; a real tokenizer lands with the Cost() stamp
// in Phase 2/3. Good enough to make the budget bite for the A/B.
const tokenApprox = 4

func tokenCount(s string) int {
	if n := len(s) / tokenApprox; n > 1 {
		return n
	}
	return 1
}

// frameNode adapts a dna.MemoryCell as a traverse.Node. Cost = tokens of its rendered
// line; Frontier = its children weighted by weightOf and **normalized to sum to 1** —
// cognition conserves the budget, so the frame stays within it (a bounded present).
type frameNode struct{ cell *dna.MemoryCell }

func (n *frameNode) ID() string                         { return n.cell.ID() }
func (n *frameNode) Cost() int                          { return tokenCount(renderCellText(n.cell)) }
func (n *frameNode) Express(_ traverse.Node, _ any) any { return n.cell.Content() }

func (n *frameNode) Frontier(_ traverse.Node, _ any) map[string]traverse.Alloc {
	type kid struct {
		key string
		mc  *dna.MemoryCell
		w   float64
	}
	var kids []kid
	var sum float64
	for key, child := range n.cell.Children() {
		mc, ok := child.(*dna.MemoryCell)
		if !ok {
			continue
		}
		w := weightOf(mc)
		kids = append(kids, kid{key, mc, w})
		sum += w
	}
	if sum == 0 {
		return nil
	}
	out := make(map[string]traverse.Alloc, len(kids))
	for _, k := range kids {
		out[k.key] = traverse.Alloc{Target: &frameNode{cell: k.mc}, Weight: k.w / sum} // normalized → conserve
	}
	return out
}

// weightOf is the Phase-1 stand-in for edge alignment: a prediction's calibrated
// Probability (branches), floored so distribution roots / proposals still get budget;
// 1 otherwise. In Phase 2 this becomes the learned strength read off the edge record.
func weightOf(cell *dna.MemoryCell) float64 {
	if p, ok := cell.Content().(dna.Prediction); ok {
		if p.Probability > 0 {
			return p.Probability
		}
		return 0.15
	}
	return 1
}

// renderCellText renders one cell's content to a present line ("" = contributes no line,
// e.g. a distribution root, though its branches still descend).
func renderCellText(cell *dna.MemoryCell) string {
	switch c := cell.Content().(type) {
	case dna.Gist:
		return "Framing: " + c.Text
	case dna.Prediction:
		switch c.Role {
		case "other":
			if c.Probability > 0 {
				return fmt.Sprintf("  p=%.2f  %s", c.Probability, c.Text)
			}
			return "" // distribution root — container, no line of its own
		case "self":
			return "self proposes: " + c.Text
		}
	case dna.ActionRecord:
		return fmt.Sprintf("action: %s %s %v", c.Primitive, c.EntityID, c.Content)
	}
	return ""
}

// AssembleFramePresent walks the grouping under a token budget and renders the lit
// subgraph into a present string — the traversal-built present. Returns the present and
// the count of lit nodes.
func AssembleFramePresent(grouping *dna.MemoryCell, budget float64) (string, int) {
	if grouping == nil {
		return "", 0
	}
	res := traverse.Assemble(&frameNode{cell: grouping}, budget, nil)
	lit := res.Flatten()
	var lines []string
	for _, node := range lit {
		if fn, ok := node.(*frameNode); ok {
			if line := renderCellText(fn.cell); line != "" {
				lines = append(lines, line)
			}
		}
	}
	return strings.Join(lines, "\n"), len(lit)
}

// ComparePresents renders the latest grouping two ways for the Phase-1 A/B — the
// hand-built renderOtherDistribution and the traversal-built AssembleFramePresent.
// Exported so cmd/ can drive it after a real cognitive cycle.
func ComparePresents(d *DNA, budget float64) (hand, traversal string, litNodes int) {
	g := d.Groupings().Latest()
	if g == nil {
		return "", "", 0
	}
	hand = renderOtherDistribution(d.Groupings(), g)
	traversal, litNodes = AssembleFramePresent(g, budget)
	return hand, traversal, litNodes
}

package dnav2

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra/traverse"
)

// Phase 1 (traversal-implementation-plan.md): assemble the "present" for inference by
// TRAVERSING the episode subgraph under a token budget — the analogue of the
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

// frameNode adapts a MemoryCell as a traverse.Node. Cost = tokens of its rendered
// line; Frontier = its children weighted and **normalized to sum to 1** — cognition
// conserves the budget, so the frame stays within it (a bounded present). The edge
// weight is the learned Strength off the store (relevance = alignment); an edge with no
// record yet falls back to the Phase-1 static Probability, so the two phases layer.
type frameNode struct {
	cell    *MemoryCell
	store   *ActivationStore // nil → Phase-1 Probability weights
	tick    int
	exclude map[string]bool // node IDs the settle pruned this pass (nil → none); skipped in Frontier
	// rootFrame is the ROOT episode's current frame ID: exactly one frame renders as
	// "Framing:" per frame; every other frame (the root's priors, and any frame the
	// wave bled into through the entity mesh) renders as an earlier framing —
	// visible history ("before I thought x"), gradeable by the settle.
	rootFrame string
}

func (n *frameNode) ID() string                         { return n.cell.ID() }
func (n *frameNode) Cost() int                          { return tokenCount(renderCellText(n.cell, n.rootFrame)) }
func (n *frameNode) Express(_ traverse.Node, _ any) any { return n.cell.Content() }

func (n *frameNode) child(mc *MemoryCell) *frameNode {
	return &frameNode{cell: mc, store: n.store, tick: n.tick, exclude: n.exclude, rootFrame: n.rootFrame}
}

// weight is the learned edge Strength (relevance = alignment); falls back to the static
// Probability while an edge has no record yet.
func (n *frameNode) weight(edgeKey string, mc *MemoryCell) float64 {
	if n.store != nil {
		if s, ok := n.store.Strength(edgeKey, n.tick); ok {
			return s
		}
	}
	return weightOf(mc)
}

func (n *frameNode) Frontier(_ traverse.Node, _ any) map[string]traverse.Alloc {
	type kid struct {
		key string
		mc  *MemoryCell
		w   float64
	}
	var kids []kid
	var sum float64
	for key, child := range n.cell.Children() {
		mc, ok := child.(*MemoryCell)
		if !ok {
			continue
		}
		if n.exclude[mc.ID()] { // settle pruned this node this pass (nil map → always false)
			continue
		}
		w := n.weight(key, mc)
		kids = append(kids, kid{key, mc, w})
		sum += w
	}
	if sum == 0 {
		return nil
	}
	out := make(map[string]traverse.Alloc, len(kids))
	for _, k := range kids {
		out[k.key] = traverse.Alloc{Target: n.child(k.mc), Weight: k.w / sum} // normalized → conserve
	}
	return out
}

// weightOf is the Phase-1 stand-in for edge alignment: a prediction's calibrated
// Probability (branches), floored so distribution roots / proposals still get budget;
// 1 otherwise. In Phase 2 this becomes the learned strength read off the edge record.
func weightOf(cell *MemoryCell) float64 {
	if p, ok := cell.Content().(Prediction); ok {
		if p.Probability > 0 {
			return p.Probability
		}
		return 0.15
	}
	return 1
}

// renderCellText renders one cell's content to a present line ("" = contributes no line,
// e.g. a episode shell or a distribution root, though their children still descend).
// rootFrame marks the one frame that renders as the current framing.
func renderCellText(cell *MemoryCell, rootFrame string) string {
	switch c := cell.Content().(type) {
	case Frame:
		if cell.ID() == rootFrame {
			return "Framing: " + c.Text
		}
		return "earlier framing: " + c.Text
	case Prediction:
		switch c.Role {
		case "other":
			if c.Probability > 0 {
				return fmt.Sprintf("  p=%.2f  %s", c.Probability, c.Text)
			}
			return "" // distribution root — container, no line of its own
		case "self":
			return "self proposes: " + c.Text
		}
	case ActionRecord:
		return fmt.Sprintf("action: %s %s %v", c.Primitive, c.EntityID, c.Content)
	case InteractionRecord:
		return c.From + ": " + c.Utterance
	case PerceiveRecord:
		// no body line — perceive records are NORMALIZED into the frame-top
		// location block (renderLocations); also keeps them out of the settle's
		// items (a location is display structure, not a coherence claim).
		return ""
	}
	return ""
}

// renderLocations dedups the TRAVERSED perceive records to the freshest coord per
// entity and renders the location block pinned at the top of the frame. Display-
// only: it writes nothing — no A, no F, no economy effect. The traversal delivers
// the nodes; normalization only decides what to DISPLAY out of them — a deliberate
// cheat for frame structure, not a graph query and not grading
// (perceive-build-uncertainties.md §4; the §2 shortcut — drops away when the
// per-kind decay λ lands in activation.go).
func renderLocations(res *traverse.Result) []string {
	if res == nil {
		return nil
	}
	freshest := map[string]PerceiveRecord{}
	for _, node := range res.Flatten() {
		fn, ok := node.(*frameNode)
		if !ok {
			continue
		}
		pr, ok := fn.cell.Content().(PerceiveRecord)
		if !ok {
			continue
		}
		if cur, seen := freshest[pr.EntityID]; !seen || pr.Timestamp > cur.Timestamp {
			freshest[pr.EntityID] = pr
		}
	}
	ids := make([]string, 0, len(freshest))
	for id := range freshest {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	lines := make([]string, 0, len(ids))
	for _, id := range ids {
		pr := freshest[id]
		lines = append(lines, fmt.Sprintf("location: %s at %d,%d", id, pr.Coord.X, pr.Coord.Y))
	}
	return lines
}

// FrameOf walks the episode under a token budget and returns the traversal Result — the
// lit subgraph — for both rendering (Flatten) and credit distribution (Edges). store=nil
// uses Phase-1 static weights; a store uses learned strength.
func FrameOf(episode *MemoryCell, budget float64, store *ActivationStore, tick int) *traverse.Result {
	return frameOfExcluding(episode, budget, store, tick, nil)
}

// frameOfExcluding is FrameOf with a settle-pruned node set: excluded node IDs are skipped
// in Frontier, so a re-traversal after pruning frees their budget for coherent-far nodes.
func frameOfExcluding(episode *MemoryCell, budget float64, store *ActivationStore, tick int, exclude map[string]bool) *traverse.Result {
	if episode == nil {
		return nil
	}
	rootFrame := ""
	if frame := latestFrame(episode); frame != nil {
		rootFrame = frame.ID()
	}
	return traverse.Assemble(&frameNode{cell: episode, store: store, tick: tick, exclude: exclude, rootFrame: rootFrame}, budget, nil)
}

// AssembleFramePresent walks the episode with Phase-1 static weights (no learning).
func AssembleFramePresent(episode *MemoryCell, budget float64) (string, int) {
	return renderFrame(FrameOf(episode, budget, nil, 0))
}

// AssembleWithStore is the Phase-2 assembler: edge weights come from the learned
// activation store (relevance = alignment), so grading reshapes the frame.
func AssembleWithStore(episode *MemoryCell, budget float64, store *ActivationStore, tick int) (string, int) {
	return renderFrame(FrameOf(episode, budget, store, tick))
}

// FramePresentCrossing is the live-loop assembler: it walks the episode under the
// learned store, CROSSES the traversed edges (traversal = crossing → F-bump), and
// returns the rendered present. Use this where a gene feeds the frame to inference.
func FramePresentCrossing(episode *MemoryCell, budget float64, store *ActivationStore, tick int) string {
	res := FrameOf(episode, budget, store, tick)
	if res == nil {
		return ""
	}
	if store != nil {
		store.CrossFrame(res.Edges(), tick)
	}
	s, n := renderFrame(res)
	debug.Being("skyra", "midlevel-frame", fmt.Sprintf("assembled budget=%.0f lit=%d edges=%d", budget, n, len(res.Edges())))
	return s
}

func renderFrame(res *traverse.Result) (string, int) {
	if res == nil {
		return "", 0
	}
	lit := res.Flatten()
	lines := renderLocations(res) // the normalized location block, pinned at the top
	for _, node := range lit {
		if fn, ok := node.(*frameNode); ok {
			if line := renderCellText(fn.cell, fn.rootFrame); line != "" {
				lines = append(lines, line)
			}
		}
	}
	return strings.Join(lines, "\n"), len(lit)
}

// ComparePresents renders the latest episode two ways for the Phase-1 A/B — the
// hand-built renderOtherDistribution and the traversal-built AssembleFramePresent.
// Exported so cmd/ can drive it after a real cognitive cycle.
func ComparePresents(d *DNA, budget float64) (hand, traversal string, litNodes int) {
	g := d.Episodes().Latest()
	if g == nil {
		return "", "", 0
	}
	hand = renderOtherDistribution(d.Episodes(), g)
	traversal, litNodes = AssembleFramePresent(g, budget)
	return hand, traversal, litNodes
}

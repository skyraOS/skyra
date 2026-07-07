package dnav2

import (
	"fmt"
	"strings"

	"skyra-v1/src/debug"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/traverse"
)

// The coherence-settle (active-grouping.md, "Alignment vs closure") is the SOLE writer of A
// (alignment / edge strength). Given the active episode's FIXED gist (the framing = truth),
// it batch-grades the episode's nodes against that framing, writes each score as A onto the
// edge that reached the node, prunes the nodes below 0 (excluded from THIS settle only — not
// deleted), re-traverses the survivors so freed budget swaps misfit-near for coherent-far,
// and regrades — until the average alignment PLATEAUS. One-sided: the framing never moves
// (midlevel withholds its frame until the settle plateaus). Closure is a separate, parked
// axis; this loop does not touch it.

const (
	settleFrameBudget = 2000 // generous — grade the whole episode frame (matches credit/self)
	settleMaxPasses   = 4    // safety cap; the plateau/no-prune stop usually hits sooner
	settlePlateauEps  = 0.02 // avg-alignment delta below which the settle is called settled
	settleStress      = 1.0  // salience stamp for settle grades (matches closure's stand-in)
)

// coherenceSettle runs the settle over one grouping against its fixed gist and returns the
// number of passes taken (for logging). No-op without a gist, a store, or inference.
func coherenceSettle(grouping *dna.MemoryCell, store *ActivationStore, tick int, infer InferFunc) int {
	if grouping == nil || store == nil || infer == nil {
		return 0
	}
	gist := gistOf(grouping)
	if gist == "" {
		return 0
	}
	excluded := map[string]bool{} // node IDs pruned this settle (exclude-this-settle, not deletion)
	crossed := map[string]bool{}  // edges Crossed this settle — F-bump once, so Grade's new record isn't deadened
	prevAvg := 0.0
	for pass := 1; pass <= settleMaxPasses; pass++ {
		items := collectSettleItems(frameOfExcluding(grouping, settleFrameBudget, store, tick, excluded))
		if len(items) == 0 {
			return pass - 1
		}
		grades := gradeCoherence(gist, items, infer)
		if len(grades) == 0 {
			return pass - 1 // a failed/empty grade settles nothing this pass
		}

		var sum float64
		var n, pruned int
		for i, it := range items {
			score, ok := grades[i+1] // items are 1-numbered in the prompt + parser
			if !ok {
				continue
			}
			// Cross before Grade: Grade mints a record at F=0, and Strength is a PRODUCT
			// (rectSoft(A)·S·F·conf), so an ungraded-then-graded edge would read weight 0 —
			// AND, now that a record exists, bypass the static fallback → deadened, unable to
			// recover (weight 0 → never traversed → never crossed). The settle traversed this
			// edge to build the frame, so the cross is legitimate; once per settle avoids
			// inflating F across passes. Pruned edges get it too → softFloor-low but alive.
			if !crossed[it.edgeKey] {
				store.Cross(it.edgeKey, tick)
				crossed[it.edgeKey] = true
			}
			store.Grade(it.edgeKey, score, settleStress) // write A (full eligibility — the direct measure)
			sum += score
			n++
			if score < 0 { // exclude-this-settle: drop it so the re-traversal reaches further
				excluded[it.nodeID] = true
				pruned++
			}
		}
		if n == 0 {
			return pass
		}
		avg := sum / float64(n)
		debug.Being("skyra", "settle", fmt.Sprintf("pass %d: %d items, %d pruned, avg=%.2f", pass, n, pruned, avg))

		if pruned == 0 { // nothing left to prune → settled
			return pass
		}
		if pass > 1 && absF(avg-prevAvg) < settlePlateauEps { // average stopped moving → settled
			return pass
		}
		prevAvg = avg
	}
	return settleMaxPasses
}

// settleItem — one gradeable node in the settle window: the edge that reached it (where A is
// written) and its rendered line (what the grader reads).
type settleItem struct {
	edgeKey string
	nodeID  string
	line    string
}

// collectSettleItems walks the lit subgraph and returns every non-root node that renders a
// line — the items graded against the framing. The root is the gist itself (the framing),
// so it is skipped: an item is not graded against itself.
func collectSettleItems(res *traverse.Result) []settleItem {
	var items []settleItem
	var walk func(r *traverse.Result)
	walk = func(r *traverse.Result) {
		if r == nil {
			return
		}
		if r.EdgeKey != "" { // non-root: has an edge whose A the grade writes
			if fn, ok := r.Node.(*frameNode); ok {
				if line := renderCellText(fn.cell); line != "" {
					items = append(items, settleItem{edgeKey: r.EdgeKey, nodeID: r.Node.ID(), line: line})
				}
			}
		}
		for _, c := range r.Children {
			walk(c)
		}
	}
	walk(res)
	return items
}

// gradeCoherence batch-grades the items' coherence to the framing in one inference, returning
// item-number → alignment. Numbering is 1-based to match the prompt and ParseCoherenceGrades.
func gradeCoherence(gist string, items []settleItem, infer InferFunc) map[int]float64 {
	var b strings.Builder
	b.WriteString("FRAMING: ")
	b.WriteString(gist)
	b.WriteString("\n\nItems (each from the working set):\n")
	for i, it := range items {
		b.WriteString(fmt.Sprintf("[%d] %s\n", i+1, it.line))
	}
	resp, err := infer(coherenceSystem, b.String())
	if err != nil {
		return nil
	}
	return ParseCoherenceGrades(resp)
}

func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Package traverse is the Phase-0 transport substrate for traversal-mechanics.md:
// a void kernel that walks a node graph under a budget and returns the lit subgraph
// (the frame). It carries NO cognition — no decay factor, no threshold, no
// normalization, no notion of what the budget measures. Pure arithmetic:
//
//	child_budget = (parent_budget − parent_cost) × weight     // no / Σ weight
//	descend iff  child_budget ≥ child_cost                     // affordability
//
// It is intentionally isolated from relation.go and the world — the collapse onto
// skyra.Space (Frontier → Observe, + Cost) is Phase 3. See
// design-docs/kernel/traversal-implementation-plan.md.
package traverse

import "sort"

// Node is the minimal contract the substrate walks. At the Phase-3 collapse these
// merge into skyra.Space: Frontier becomes Observe, and origin/impulse thread through
// unchanged.
type Node interface {
	ID() string
	// Cost is the node's weight against the budget, in the budget's units (tokens
	// for cognition, 1 for a world hop). Stamped on the node; the kernel only reads it.
	Cost() int
	// Frontier is the cell's ONWARD DISTRIBUTION: the edges it lights for this impulse,
	// each with a RELATIVE weight, keyed by edge identity. nil = leaf. The cell owns
	// the policy (which edges, weights, how broad / conserved); the kernel only scales.
	Frontier(origin Node, impulse any) map[string]Alloc
	// Express is the ascent fold — the node's rollup, returned post-order.
	Express(origin Node, impulse any) any
}

// Alloc is one entry in the onward distribution: a child and its RELATIVE weight.
// Weights summing to 1 conserve the budget (cognition, bounded frame); weights of 1
// each broadcast it (spatial, flood to radius). The kernel does not normalize — the
// cell's weights decide.
type Alloc struct {
	Target Node
	Weight float64
}

// Result is one node in the lit subgraph — the node, its post-order Express, and its
// descended children. The tree IS the frame.
type Result struct {
	Node     Node
	Express  any
	Children []*Result
}

// Assemble walks root under budget and returns the lit subgraph. budget is in the
// node graph's own units (tokens / hops). Deterministic: siblings are visited in
// sorted edge-key order.
func Assemble(root Node, budget float64, impulse any) *Result {
	return walk(root, root, budget, impulse, map[string]bool{})
}

func walk(origin, node Node, budget float64, impulse any, visited map[string]bool) *Result {
	visited[node.ID()] = true
	r := &Result{Node: node}

	remainder := budget - float64(node.Cost())
	dist := node.Frontier(origin, impulse)

	keys := make([]string, 0, len(dist))
	for k := range dist {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		a := dist[k]
		if a.Target == nil || visited[a.Target.ID()] { // the visited-once rail
			continue
		}
		childBudget := remainder * a.Weight
		if childBudget >= float64(a.Target.Cost()) { // affordability
			r.Children = append(r.Children, walk(origin, a.Target, childBudget, impulse, visited))
		}
	}

	r.Express = node.Express(origin, impulse) // post-order fold — after descent
	return r
}

// Flatten returns every node in the lit subgraph in pre-order (root first) — the
// frame as a flat list, for rendering the present.
func (r *Result) Flatten() []Node {
	if r == nil {
		return nil
	}
	out := []Node{r.Node}
	for _, c := range r.Children {
		out = append(out, c.Flatten()...)
	}
	return out
}

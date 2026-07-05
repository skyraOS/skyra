package traverse

import "testing"

// tnode is a fixed test node: a stamped cost, a precomputed frontier, an Express value.
type tnode struct {
	id    string
	cost  int
	edges map[string]Alloc
}

func (n *tnode) ID() string                              { return n.id }
func (n *tnode) Cost() int                               { return n.cost }
func (n *tnode) Frontier(_ Node, _ any) map[string]Alloc { return n.edges }
func (n *tnode) Express(_ Node, _ any) any               { return n.id } // returns own id — proves post-order

// entry / edge / mk build a frontier keyed by "parent→child".
type entry struct {
	k string
	t *tnode
	w float64
}

func edge(parent string, child *tnode, w float64) entry {
	return entry{k: parent + "→" + child.id, t: child, w: w}
}

func mk(entries ...entry) map[string]Alloc {
	m := map[string]Alloc{}
	for _, e := range entries {
		m[e.k] = Alloc{Target: e.t, Weight: e.w}
	}
	return m
}

func litIDs(r *Result) map[string]bool {
	out := map[string]bool{}
	for _, n := range r.Flatten() {
		out[n.ID()] = true
	}
	return out
}

func ids(ns []Node) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = n.ID()
	}
	return out
}

// Conservation: weights that sum to 1 at each split can never spend more budget than
// the root carried — the total cost of the lit frame stays within the root budget.
func TestConservationBoundsTheFrame(t *testing.T) {
	// root(0) → A(cost 30, w .5), B(cost 30, w .5); each → a leaf(cost 20, w 1)
	a1 := &tnode{id: "a1", cost: 20}
	b1 := &tnode{id: "b1", cost: 20}
	a := &tnode{id: "a", cost: 30, edges: mk(edge("a", a1, 1))}
	b := &tnode{id: "b", cost: 30, edges: mk(edge("b", b1, 1))}
	root := &tnode{id: "root", cost: 0, edges: mk(edge("root", a, 0.5), edge("root", b, 0.5))}

	r := Assemble(root, 100, nil)
	lit := r.Flatten()
	total := 0
	for _, n := range lit {
		total += n.Cost()
	}
	if total > 100 {
		t.Fatalf("frame cost %d exceeds root budget 100 — conservation broken", total)
	}
	// each subtree gets (100-0)*.5 = 50, minus its own 30 = 20 left; leaf costs 20 → fits.
	if len(lit) != 5 {
		t.Fatalf("lit %d nodes, want 5 (%v)", len(lit), ids(lit))
	}
}

// Affordability: a child whose delivered budget can't cover its own cost is dropped,
// with no threshold constant anywhere.
func TestAffordabilityPrunes(t *testing.T) {
	cheap := &tnode{id: "cheap", cost: 5}
	pricey := &tnode{id: "pricey", cost: 90}
	// root(cost 0) budget 20 → remainder 20. cheap gets 20*.5=10 ≥ 5 (in);
	// pricey gets 20*.5=10 < 90 (out).
	root := &tnode{id: "root", cost: 0, edges: mk(edge("root", cheap, 0.5), edge("root", pricey, 0.5))}

	lit := litIDs(Assemble(root, 20, nil))
	if !lit["cheap"] {
		t.Fatal("cheap child should be afforded")
	}
	if lit["pricey"] {
		t.Fatal("pricey child should be pruned by affordability")
	}
}

// Flood-fill: weight-1 (broadcast, no conservation) + cost-1 hops + the budget → the
// lit region reaches exactly `budget` deep. The world's spatial spread with no decay
// constant — radius falls out of budget depletion.
func TestFloodFillReachesRadius(t *testing.T) {
	// a line: n0 → n1 → n2 → n3 → n4, each cost 1, weight 1.
	nodes := make([]*tnode, 5)
	for i := 4; i >= 0; i-- {
		n := &tnode{id: string(rune('0' + i)), cost: 1}
		if i < 4 {
			n.edges = mk(edge(n.id, nodes[i+1], 1))
		}
		nodes[i] = n
	}
	// budget 4 reaches n0..n3 (4 nodes); n4 gets 0 < 1 and is dropped.
	lit := litIDs(Assemble(nodes[0], 4, nil))
	for _, id := range []string{"0", "1", "2", "3"} {
		if !lit[id] {
			t.Fatalf("node %s should be within radius", id)
		}
	}
	if lit["4"] {
		t.Fatal("node 4 is past the budget radius — should be dropped")
	}
}

// The visited-once rail: a diamond (two parents → one child) includes the shared child
// exactly once, so the walk terminates and doesn't double-count.
func TestVisitedRailDeDuplicates(t *testing.T) {
	shared := &tnode{id: "shared", cost: 1}
	left := &tnode{id: "left", cost: 1, edges: mk(edge("left", shared, 1))}
	right := &tnode{id: "right", cost: 1, edges: mk(edge("right", shared, 1))}
	root := &tnode{id: "root", cost: 1, edges: mk(edge("root", left, 1), edge("root", right, 1))}

	r := Assemble(root, 100, nil)
	count := 0
	for _, n := range r.Flatten() {
		if n.ID() == "shared" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("shared node appears %d times, want exactly 1 (the rail)", count)
	}
}

// Post-order fold: a node's Express runs after its children are walked; the Result tree
// mirrors the graph and is populated bottom-up.
func TestExpressIsPostOrder(t *testing.T) {
	leaf := &tnode{id: "leaf", cost: 1}
	root := &tnode{id: "root", cost: 1, edges: mk(edge("root", leaf, 1))}

	r := Assemble(root, 10, nil)
	if r.Express != "root" || len(r.Children) != 1 || r.Children[0].Express != "leaf" {
		t.Fatalf("fold tree wrong: root.Express=%v children=%d", r.Express, len(r.Children))
	}
}

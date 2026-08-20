package graph

import (
	"strings"
	"testing"
)

// buildAgent is the worked example wired with real workers.
func buildAgent(t *testing.T) (*Namespace, *Workers) {
	t.Helper()
	ns := NewNamespace("skyra", nil)
	w := NewWorkers(ns)
	ns.behavior = w.Turn

	ns.Ground("Maintain Agent", Act{Primitive: "perceive_self"})
	respond, err := ns.AddExpectation("skyra:/", "respond", "Respond",
		Act{Primitive: "perceive_reply"},
		Observation{ID: "obs:1", Support: 0.6, Confidence: 0.8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ns.AddAction(respond.ID, "generate", Act{Primitive: "generate_response"},
		Observation{ID: "obs:2", Support: 0.5, Confidence: 0.9}); err != nil {
		t.Fatal(err)
	}
	return ns, w
}

// One turn runs its phases in order, on one stack, in one wave. This is the pipeline the
// iterating Perturb exists for: EXPECT runs against a graph INTERPRET has already touched.
func TestTurnRunsItsPhasesInOrder(t *testing.T) {
	ns, _ := buildAgent(t)
	root, _ := ns.Root()

	stack, _ := ns.Spawn(root, "user says hello")
	stack.Run()

	tr := stack.Transcript()
	if tr.WaveID != "skyra/wave:1" {
		t.Fatalf("waveID = %q", tr.WaveID)
	}

	// the three phases are SIBLINGS at depth 1 beneath the turn — not nested in
	// each other
	var atDepth1 int
	for _, h := range tr.Hops {
		if h.Depth == 1 {
			atDepth1++
		}
	}
	if atDepth1 != 3 {
		t.Fatalf("phases at depth 1 = %d, want 3 (interpret, expect, pursue)\n%+v", atDepth1, tr.Hops)
	}
	if !stack.Done() {
		t.Error("turn did not terminate — a Space must eventually return nil from Perturb")
	}
}

// EXPECT grows the tree. UNGATED, so it fires on every impulse regardless of whether a gap
// was found — deliberately, so the trigger can be calibrated from what it does rather than
// from argument.
func TestExpectGrowsTheTreeUngated(t *testing.T) {
	ns, _ := buildAgent(t)
	root, _ := ns.Root()
	before := ns.Len()

	stack, _ := ns.Spawn(root, "user says hello")
	stack.Run()

	if ns.Len() <= before {
		t.Fatalf("graph did not grow: %d → %d", before, ns.Len())
	}

	// what it grew arrived WITH its perceive — unconstructible otherwise
	var found *Node
	for id, n := range ns.nodes {
		if strings.Contains(string(id), "expect-") && n.IsDirectory() {
			found = n
		}
	}
	if found == nil {
		t.Fatal("nothing was created")
	}
	var hasPerceive bool
	for _, e := range found.Entries {
		if c, _ := ns.Node(e.Target); c != nil && c.IsFile() && c.Act.Perceive {
			hasPerceive = true
		}
	}
	if !hasPerceive {
		t.Error("created expectation has no perceive; it cannot be wrong")
	}
}

// PURSUE is stubbed: the tree grows but nothing fires. With EXPECT ungated, running both
// would have the being ACT on fabricated structure rather than merely store it.
func TestPursueIsStubbed(t *testing.T) {
	ns, w := buildAgent(t)
	if w.Pursue.ID() != "pursue/stub" {
		t.Fatalf("pursue = %q, want the stub", w.Pursue.ID())
	}
	root, _ := ns.Root()
	stack, _ := ns.Spawn(root, "impulse")
	stack.Run()

	h := stack.Realized().(*Handle)
	if len(h.Rejects) != 0 {
		t.Errorf("rejects: %v", h.Rejects)
	}
}

// A worker never traverses — it builds a frontier and hands off a Walk. The Walk's members
// are the node's children, opened one at a time.
func TestInterpretDelegatesToAWalk(t *testing.T) {
	ns, w := buildAgent(t)
	root, _ := ns.Root()

	h := ns.OpenWith(root, w.Interpret)
	next := h.Perturb(h, "impulse")
	walk, ok := next.(*Walk)
	if !ok {
		t.Fatalf("interpret promoted %T, want a *Walk", next)
	}
	if walk.Len() != len(root.Entries) {
		t.Fatalf("walk covers %d members, want %d", walk.Len(), len(root.Entries))
	}

	// and it must eventually stop: a frame is asked repeatedly
	h.Scratch = walk
	if again := h.Perturb(h, "impulse"); again != nil {
		t.Fatal("interpret promoted a second frontier; it would never terminate")
	}
}

// The whole frontier comes back as residual, because nothing can be matched yet — an
// interpretation carries no content to check against the tree (questions.md Q10). Honest
// rather than silently empty.
func TestUnmatchedFrontierBecomesResidual(t *testing.T) {
	ns, w := buildAgent(t)
	root, _ := ns.Root()

	stack, _ := ns.Spawn(root, "impulse")
	stack.Run()

	// walk the transcript for the interpret phase and confirm it reported residual
	h := ns.OpenWith(root, w.Interpret)
	walk := h.Perturb(h, "impulse").(*Walk)
	for walk.Perturb(h, "impulse") != nil { //nolint:revive // draining the frontier
	}
	h.Observe(h, "impulse", walk)
	if len(h.Residual) != 0 && walk.Len() == 0 {
		t.Fatal("residual without a frontier")
	}
}

// The gate is EVALUATED and RECORDED on every turn, and enforced on none. Running ungated
// shows how EXPECT behaves; the verdict is the only way to learn whether the condition was
// any good, and it cannot be recovered after the fact.
func TestGateIsRecordedButNotEnforced(t *testing.T) {
	ns, _ := buildAgent(t)
	root, _ := ns.Root()
	before := ns.Len()

	stack, _ := ns.Spawn(root, "user says hello")
	stack.Run()

	if len(ns.Gates) != 1 {
		t.Fatalf("gates recorded = %d, want 1 per turn", len(ns.Gates))
	}
	g := ns.Gates[0]
	if g.WaveID != "skyra/wave:1" || g.Node != "skyra:/" {
		t.Errorf("verdict not addressed: %+v", g)
	}
	if g.Frontier != 2 {
		t.Errorf("frontier = %d, want 2 (@perceive and respond)", g.Frontier)
	}
	// nothing can be placed yet — an interpretation carries no content (Q10)
	if g.Residual != g.Frontier || g.Unplaced != 1 || !g.Fires {
		t.Errorf("verdict = %+v, want a firing gate while nothing can be matched", g)
	}
	if g.Reason == "" {
		t.Error("a verdict with no reason is not readable later")
	}

	// and it changed nothing: EXPECT ran regardless
	if ns.Len() <= before {
		t.Fatal("EXPECT did not fire — the gate must not be enforced")
	}
}

// The gate declines to speak when there was nothing in view, rather than reporting a
// division by zero as a firing signal.
func TestGateSaysNothingOnAnEmptyFrontier(t *testing.T) {
	ns, _ := buildAgent(t)
	leaf, _ := ns.Node("skyra:/respond/generate") // an action: no entries

	stack, _ := ns.Spawn(leaf, "impulse")
	stack.Run()

	g := ns.Gates[0]
	if g.Frontier != 0 || g.Fires {
		t.Fatalf("verdict = %+v, want a silent gate on an empty frontier", g)
	}
}

// Created names come from a monotonic counter, not from graph size — two turns must not
// collide on a name just because the graph happened to be the same size.
func TestExpectNamesDoNotCollide(t *testing.T) {
	ns, _ := buildAgent(t)
	root, _ := ns.Root()

	for i := 0; i < 3; i++ {
		stack, _ := ns.Spawn(root, "impulse")
		stack.Run()
	}

	seen := map[NodeID]bool{}
	var created int
	for id, n := range ns.nodes {
		if n.IsDirectory() && strings.Contains(string(id), "expect-") {
			if seen[id] {
				t.Errorf("duplicate creation %q", id)
			}
			seen[id] = true
			created++
		}
	}
	if created != 3 {
		t.Fatalf("created %d expectations over 3 turns, want 3", created)
	}
}

package kernel

import (
	"strings"
	"testing"
)

// probe is a Space that promotes a fixed chain and records the order in which the
// kernel called it. It carries its own continuation privately — the kernel never sees
// a collection, only the one Space handed back.
type probe struct {
	id     string
	next   *probe
	log    *[]string
	inert  bool // Observe returns nil — a legitimate outcome, not a failure
	handed bool // yielded already; a Space must eventually return nil from Perturb
}

func (p *probe) ID() string { return p.id }

// Perturb is called REPEATEDLY on a frame, so a Space that keeps handing back the same
// child never terminates. Yielding once and then nil is the minimum a Space must do.
func (p *probe) Perturb(origin Space, impulse any) Space {
	*p.log = append(*p.log, "perturb:"+p.id)
	if p.next == nil || p.handed {
		return nil // the descent base case
	}
	p.handed = true
	return p.next
}

func (p *probe) Observe(origin Space, impulse any, realized Space) Space {
	rid := "nil"
	if realized != nil {
		rid = realized.ID()
	}
	*p.log = append(*p.log, "observe:"+p.id+"<-"+rid)
	if p.inert {
		return nil
	}
	return p
}

func chain(log *[]string, ids ...string) *probe {
	var head, prev *probe
	for _, id := range ids {
		n := &probe{id: id, log: log}
		if prev == nil {
			head = n
		} else {
			prev.next = n
		}
		prev = n
	}
	return head
}

// The fold happens on the stack: every Perturb on the way down, every Observe on the
// way up, and each Observe receives the realized Space from below. This is what makes
// the kernel's tree-building unnecessary — the return path IS the tree.
//
// Note the RE-ENTRY: after a frame folds a child it is asked for another continuation,
// and only pops when it promotes nothing. That is what lets one frame run a pipeline —
// and it is why a Space must eventually return nil.
func TestDescentThenPostOrderAscent(t *testing.T) {
	var log []string
	root := chain(&log, "a", "b", "c")

	k := NewRelation("k")
	k.Realize(root, root, nil)

	want := []string{
		"perturb:a", "perturb:b", "perturb:c", // descend
		"observe:c<-nil", // c promoted nothing: leaf observes with nil
		"observe:b<-c",   // b folds c
		"perturb:b",      // b asked AGAIN — promotes nothing, pops
		"observe:a<-b",   // a folds b
		"perturb:a",      // a asked again — promotes nothing, pops
	}
	if strings.Join(log, " ") != strings.Join(want, " ") {
		t.Fatalf("call order:\n got %v\nwant %v", log, want)
	}
}

// nil from Perturb terminates descent, and Observe still runs there with a nil realized
// Space. Termination is a fixed point of the promotion, not a judgement about the node.
func TestNilPerturbIsBaseCaseAndStillObserves(t *testing.T) {
	var log []string
	leaf := chain(&log, "only")

	k := NewRelation("k")
	tr := k.Realize(leaf, leaf, nil)

	if got := strings.Join(log, " "); got != "perturb:only observe:only<-nil" {
		t.Fatalf("leaf did not observe with nil realized: %q", got)
	}
	if tr.Executed() != 1 {
		t.Fatalf("executed = %d, want 1", tr.Executed())
	}
	if tr.Hops[0].Promoted {
		t.Fatal("leaf reported Promoted; nil from Perturb must record as not promoted")
	}
}

// The transcript is the spine: one hop per Space the kernel was handed, in execution
// order. Nothing about what any Space contained.
//
// It is a TREE now, not a chain — an iterating Perturb lets one frame promote several
// children — so Depth and Parent are STORED on the Hop rather than read off position.
func TestTranscriptIsTheSpine(t *testing.T) {
	var log []string
	root := chain(&log, "a", "b", "c")

	k := NewRelation("k")
	tr := k.Realize(root, root, nil)

	if tr.Executed() != 3 {
		t.Fatalf("executed = %d, want 3", tr.Executed())
	}
	if tr.Depth() != 2 {
		t.Fatalf("depth = %d, want 2", tr.Depth())
	}

	wantID := []string{"a", "b", "c"}
	for i, h := range tr.Hops {
		if h.SpaceID != wantID[i] {
			t.Errorf("hop %d: space = %q, want %q", i, h.SpaceID, wantID[i])
		}
	}

	if _, ok := tr.ParentOf(0); ok {
		t.Error("root reported a parent")
	}
	for i, wantParent := range map[int]string{1: "a", 2: "b"} {
		p, ok := tr.ParentOf(i)
		if !ok || p.SpaceID != wantParent {
			t.Errorf("parent of %d = %q (%v), want %q", i, p.SpaceID, ok, wantParent)
		}
	}
}

// An observed omission does not erase the corresponding member from the transcript or
// declare it false (§10). A Space whose Observe returns nil is still recorded as having
// executed — the two artifacts have different authority and this is where that shows.
func TestNilRealizationStillRecordsTheHop(t *testing.T) {
	var log []string
	root := chain(&log, "a", "b")
	root.next.inert = true // b realizes nothing

	k := NewRelation("k")
	tr := k.Realize(root, root, nil)

	if tr.Executed() != 2 {
		t.Fatalf("executed = %d, want 2 — a nil realization must not delete the hop", tr.Executed())
	}
	if tr.Hops[1].SpaceID != "b" || tr.Hops[1].Realized {
		t.Fatalf("hop 1 = %+v, want b recorded as executed but not realized", tr.Hops[1])
	}
	if p, ok := tr.ParentOf(1); !ok || p.SpaceID != "a" {
		t.Fatalf("b lost its spine position: parent = %q (%v)", p.SpaceID, ok)
	}
}

// grouping is a Space that privately fans out over several members and runs them
// through its OWN behavior. The kernel sees one value with one behavior boundary — the
// members never cross it, and the transcript is therefore a spine and not a tree.
type grouping struct {
	id      string
	members []*probe
	log     *[]string
}

func (g *grouping) ID() string { return g.id }

func (g *grouping) Perturb(origin Space, impulse any) Space {
	*g.log = append(*g.log, "perturb:"+g.id)
	return nil // the group's work happens in its own behavior, below
}

func (g *grouping) Observe(origin Space, impulse any, realized Space) Space {
	for _, m := range g.members {
		*g.log = append(*g.log, "member:"+m.id) // sequence, fan-out, or join — the group's call
	}
	return g
}

func TestGroupedMembersNeverCrossTheKernel(t *testing.T) {
	var log []string
	g := &grouping{
		id:  "group",
		log: &log,
		members: []*probe{
			{id: "m1", log: &log},
			{id: "m2", log: &log},
			{id: "m3", log: &log},
		},
	}

	k := NewRelation("k")
	tr := k.Realize(g, g, nil)

	if tr.Executed() != 1 {
		t.Fatalf("executed = %d, want 1 — three members ran but only the group crossed the kernel", tr.Executed())
	}
	if tr.Hops[0].SpaceID != "group" {
		t.Fatalf("hop = %q, want the group", tr.Hops[0].SpaceID)
	}
	for _, m := range []string{"member:m1", "member:m2", "member:m3"} {
		if !strings.Contains(strings.Join(log, " "), m) {
			t.Fatalf("%s did not run: %v", m, log)
		}
	}
}

// The assembly boundary is the one caller entitled to the root's concrete return.
// Realize cannot reach it at all — that is the Perception Invariant expressed as which
// method you are able to name.
func TestOnlyTheAssemblyBoundaryReachesTheConcreteReturn(t *testing.T) {
	var log []string
	root := chain(&log, "a", "b")

	k := NewRelation("k")
	concrete, tr := k.RealizeConcrete(root, root, nil)

	if concrete == nil || concrete.ID() != "a" {
		t.Fatalf("concrete = %v, want the root's realization", concrete)
	}
	if tr.Executed() != 2 {
		t.Fatalf("executed = %d, want 2", tr.Executed())
	}
}

// Two waves over the same Space must not collide. The kernel holds no state between
// waves: each Realize gets its own wave value, which is the whole reason the transcript
// is returned rather than accumulated somewhere.
func TestConcurrentWavesDoNotShareTranscriptState(t *testing.T) {
	var logA, logB []string
	a := chain(&logA, "a", "b", "c")
	b := chain(&logB, "x", "y")

	k := NewRelation("k")
	done := make(chan Transcript, 2)
	go func() { done <- k.Realize(a, a, nil) }()
	go func() { done <- k.Realize(b, b, nil) }()

	got := map[int]bool{}
	for i := 0; i < 2; i++ {
		got[(<-done).Executed()] = true
	}
	if !got[3] || !got[2] {
		t.Fatalf("transcripts bled together: %v", got)
	}
}

// pipeline is a Space that yields a DIFFERENT continuation each time it is asked, and
// folds each result before deciding the next. This is the whole point of an iterating
// Perturb: one frame runs a sequence of phases, and each phase sees what the previous
// one produced.
type pipeline struct {
	id     string
	phases []*probe
	log    *[]string
	at     int
	folds  int
}

func (p *pipeline) ID() string { return p.id }

func (p *pipeline) Perturb(origin Space, impulse any) Space {
	if p.at >= len(p.phases) {
		return nil // promotes nothing: the frame is finished
	}
	next := p.phases[p.at]
	p.at++
	*p.log = append(*p.log, "yield:"+next.id)
	return next
}

func (p *pipeline) Observe(origin Space, impulse any, realized Space) Space {
	if realized != nil {
		p.folds++
		*p.log = append(*p.log, "fold:"+realized.ID())
	}
	return p
}

func TestOneFrameRunsSeveralPhasesInOrder(t *testing.T) {
	var log []string
	p := &pipeline{
		id:  "turn",
		log: &log,
		phases: []*probe{
			{id: "interpret", log: &log},
			{id: "propose", log: &log},
		},
	}

	k := NewRelation("k")
	tr := k.Realize(p, p, nil)

	want := "yield:interpret perturb:interpret observe:interpret<-nil fold:interpret " +
		"yield:propose perturb:propose observe:propose<-nil fold:propose"
	if got := strings.Join(log, " "); got != want {
		t.Fatalf("pipeline order:\n got %s\nwant %s", got, want)
	}
	if p.folds != 2 {
		t.Errorf("folded %d times, want 2 — one per phase", p.folds)
	}

	// each phase is its own hop, and both hang off the same parent
	if tr.Executed() != 3 {
		t.Fatalf("executed = %d, want 3 (turn + two phases)", tr.Executed())
	}
	for _, i := range []int{1, 2} {
		parent, ok := tr.ParentOf(i)
		if !ok || parent.SpaceID != "turn" {
			t.Errorf("hop %d parent = %q (%v), want turn", i, parent.SpaceID, ok)
		}
		if tr.Hops[i].Depth != 1 {
			t.Errorf("hop %d depth = %d, want 1 — siblings, not nested", i, tr.Hops[i].Depth)
		}
	}
}

package relation

import (
	"testing"

	"skyra-v1/src/skyra"
)

// --- mocks ---

type mockRecord struct {
	base  float64
	calls int
}

func (m *mockRecord) Activate(visited bool, count int) float64 {
	m.calls++
	return m.base
}

type obs struct {
	origin  string
	impulse any
}

type mockCell struct {
	id       string
	children map[string]skyra.Reality
	rec      *mockRecord
	content  any
	observes []obs // every Observe call: who owned the wave, and what it carried
}

func newCell(id string, base float64) *mockCell {
	return &mockCell{
		id:       id,
		children: map[string]skyra.Reality{},
		rec:      &mockRecord{base: base},
	}
}

func (c *mockCell) ID() string                            { return c.id }
func (c *mockCell) NewActivation() skyra.ActivationRecord { return c.rec }
func (c *mockCell) Observe(origin skyra.Reality, impulse any) {
	c.observes = append(c.observes, obs{origin.ID(), impulse})
}
func (c *mockCell) Express(origin skyra.Reality, impulse any) any { return c.id }
func (c *mockCell) Children() map[string]skyra.Reality           { return c.children }
func (c *mockCell) Content() any                                 { return c.content }
func (c *mockCell) link(child *mockCell)                         { c.children[child.id] = child }

type mockSlice struct {
	id        string
	records   map[string]skyra.ActivationRecord
	threshold float64
	count     int
	tree      skyra.TraversalPair
}

func newSlice(id string, threshold float64) *mockSlice {
	return &mockSlice{id: id, records: map[string]skyra.ActivationRecord{}, threshold: threshold}
}

func (s *mockSlice) ID() string                                 { return s.id }
func (s *mockSlice) Tree() *skyra.TraversalPair                 { return &s.tree }
func (s *mockSlice) Records() map[string]skyra.ActivationRecord { return s.records }
func (s *mockSlice) Threshold() float64                         { return s.threshold }
func (s *mockSlice) Count() int                                 { return s.count }
func (s *mockSlice) SetCount(c int)                             { s.count = c }

// run wires a being's own wave: the root both owns and is traversed.
func run(root *mockCell, sl *mockSlice) *Relation {
	r := NewRelation()
	r.ActiveState[root.id] = sl
	r.Realize(root, root, nil)
	return r
}

// --- tests ---

// Each cell is activated exactly once per traversal — the double-bump is gone.
func TestNoDoubleBump(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	b := newCell("b", 1.0)
	c := newCell("c", 1.0)
	root.link(a)
	root.link(b)
	a.link(c)

	run(root, newSlice("being", 0)) // threshold 0 → everything survives

	for _, cell := range []*mockCell{root, a, b, c} {
		if cell.rec.calls != 1 {
			t.Errorf("%s activated %d times, want 1 (double-bump)", cell.id, cell.rec.calls)
		}
	}
}

// A cycle terminates — visited-once is the rail.
func TestVisitedOnceTerminates(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	root.link(a)
	a.link(root) // cycle: root -> a -> root

	run(root, newSlice("being", 0))

	// reaching here at all means it terminated. root is observed twice:
	// once on descent, once on delivery — never re-descended by the cycle.
	if len(root.observes) != 2 {
		t.Errorf("root observed %d times, want 2 (descent + delivery)", len(root.observes))
	}
	if len(a.observes) != 1 {
		t.Errorf("a observed %d times, want 1", len(a.observes))
	}
}

// Amplitude attenuation limits depth: the wave dies when score < threshold.
func TestAmplitudeLimitsDepth(t *testing.T) {
	// base 1.0, decay 0.9/depth, threshold 0.7:
	// root 1.0 | a 0.9 | b 0.81 | c 0.729 (survive) | d 0.656 (< 0.7, dies)
	root := newCell("root", 1.0)
	cells := map[string]*mockCell{}
	prev := root
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		cur := newCell(id, 1.0)
		cells[id] = cur
		prev.link(cur)
		prev = cur
	}

	run(root, newSlice("being", 0.7))

	if len(cells["c"].observes) != 1 {
		t.Errorf("c observed %d, want 1 (reached)", len(cells["c"].observes))
	}
	if len(cells["d"].observes) != 0 {
		t.Errorf("d observed %d, want 0 (below threshold, not traversed)", len(cells["d"].observes))
	}
	if cells["d"].rec.calls != 1 {
		t.Errorf("d activated %d, want 1 (scored once by c, didn't survive)", cells["d"].rec.calls)
	}
	if cells["e"].rec.calls != 0 {
		t.Errorf("e activated %d, want 0 (never reached)", cells["e"].rec.calls)
	}
}

// The finished tree lands in the slice's double buffer.
func TestTreeDoubleBuffer(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	root.link(a)
	sl := newSlice("being", 0)

	run(root, sl)

	tree := sl.Tree()
	if len(tree.Current) != 1 {
		t.Fatalf("Current tree has %d roots, want 1", len(tree.Current))
	}
	if tree.Current[0].CellID != "root" {
		t.Errorf("tree root is %s, want root", tree.Current[0].CellID)
	}
	if len(tree.Current[0].Children) != 1 || tree.Current[0].Children[0].CellID != "a" {
		t.Errorf("tree root children = %+v, want [a]", tree.Current[0].Children)
	}
}

// Perception Invariant: the wave delivers its result to its origin.
func TestDeliversToOrigin(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	root.link(a)

	run(root, newSlice("being", 0))

	// the last thing the origin observes is its own wave's result (root's
	// Express output is "root"), owned by itself — the reafference.
	last := root.observes[len(root.observes)-1]
	if last.origin != "root" {
		t.Errorf("delivery origin = %q, want root", last.origin)
	}
	if last.impulse != "root" {
		t.Errorf("delivered result = %v, want \"root\"", last.impulse)
	}
}

// Origin is threaded unchanged: every cell sees the wave's owner.
func TestOriginThreaded(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	b := newCell("b", 1.0)
	root.link(a)
	a.link(b) // owner is root, two levels deep

	run(root, newSlice("being", 0))

	for _, cell := range []*mockCell{a, b} {
		if len(cell.observes) == 0 || cell.observes[0].origin != "root" {
			t.Errorf("%s saw origin %v, want root threaded through", cell.id, cell.observes)
		}
	}
}

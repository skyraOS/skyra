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

type mockCell struct {
	id        string
	children  map[string]skyra.Reality
	rec       *mockRecord
	content   any
	observed  int
	expressed int
}

func newCell(id string, base float64) *mockCell {
	return &mockCell{
		id:       id,
		children: map[string]skyra.Reality{},
		rec:      &mockRecord{base: base},
	}
}

func (c *mockCell) ID() string                               { return c.id }
func (c *mockCell) NewActivation() skyra.ActivationRecord    { return c.rec }
func (c *mockCell) Observe(k skyra.Kernel, payload any)      { c.observed++ }
func (c *mockCell) Express(k skyra.Kernel, payload any) any  { c.expressed++; return c.id }
func (c *mockCell) Children() map[string]skyra.Reality       { return c.children }
func (c *mockCell) Content() any                             { return c.content }
func (c *mockCell) link(child *mockCell)                     { c.children[child.id] = child }

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

// run wires a being's slice and traverses from root.
func run(root *mockCell, sl *mockSlice) *Relation {
	r := NewRelation()
	r.ActiveState[root.id] = sl
	r.Realize(root, nil)
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

	// threshold 0 means amplitude never stops it; only visited-once can.
	run(root, newSlice("being", 0))

	// reaching here at all means it terminated. root must not be re-descended.
	if root.observed != 1 {
		t.Errorf("root observed %d times, want 1 (cycle re-descended)", root.observed)
	}
}

// Amplitude attenuation limits depth: the wave dies when score < threshold.
func TestAmplitudeLimitsDepth(t *testing.T) {
	// base 1.0, decay 0.9/depth, threshold 0.7:
	// root 1.0 | a 0.9 | b 0.81 | c 0.729 (survive) | d 0.656 (< 0.7, dies)
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	b := newCell("b", 1.0)
	c := newCell("c", 1.0)
	d := newCell("d", 1.0)
	e := newCell("e", 1.0)
	root.link(a)
	a.link(b)
	b.link(c)
	c.link(d)
	d.link(e)

	run(root, newSlice("being", 0.7))

	if c.observed != 1 {
		t.Errorf("c observed %d, want 1 (should be reached)", c.observed)
	}
	if d.observed != 0 {
		t.Errorf("d observed %d, want 0 (below threshold, not traversed)", d.observed)
	}
	if d.rec.calls != 1 {
		t.Errorf("d activated %d, want 1 (scored once by c, just didn't survive)", d.rec.calls)
	}
	if e.rec.calls != 0 {
		t.Errorf("e activated %d, want 0 (never reached)", e.rec.calls)
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

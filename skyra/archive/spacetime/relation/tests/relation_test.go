package tests

import (
	"math"
	"sync"
	"testing"

	"skyra-v1/src/skyra/spacetime"
	"skyra-v1/src/skyra/spacetime/relation"
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
	id             string
	children       map[string]skyra.Space
	rec            *mockRecord
	slice          skyra.Slice
	content        any
	observes       []obs // every Observe call records the wave owner (origin)
	onObserve      func(obs)
	blockExpress   chan struct{}
	expressStarted chan struct{}
	expressOnce    sync.Once
}

func newCell(id string, base float64) *mockCell {
	return &mockCell{
		id:       id,
		children: map[string]skyra.Space{},
		rec:      &mockRecord{base: base},
	}
}

func (c *mockCell) ID() string                            { return c.id }
func (c *mockCell) Slice() skyra.Slice                    { return c.slice }
func (c *mockCell) NewActivation() skyra.ActivationRecord { return c.rec }
func (c *mockCell) Observe(origin skyra.Space, impulse any) {
	originID := "<nil>"
	if origin != nil {
		originID = origin.ID()
	}
	o := obs{originID, impulse}
	c.observes = append(c.observes, o)
	if c.onObserve != nil {
		c.onObserve(o)
	}
}
func (c *mockCell) Express(origin skyra.Space, impulse any) any {
	if c.expressStarted != nil {
		c.expressOnce.Do(func() { close(c.expressStarted) })
	}
	if c.blockExpress != nil {
		<-c.blockExpress
	}
	return c.id
}
func (c *mockCell) Children() map[string]skyra.Space { return c.children }
func (c *mockCell) Content() any                     { return c.content }
func (c *mockCell) Agents() map[string]skyra.Space   { return nil }
func (c *mockCell) link(child *mockCell)             { c.children[child.id] = child }
func (c *mockCell) setPlane(sl skyra.Slice) {
	c.setPlaneSeen(sl, map[string]bool{})
}
func (c *mockCell) setPlaneSeen(sl skyra.Slice, seen map[string]bool) {
	if seen[c.id] {
		return
	}
	seen[c.id] = true
	c.slice = sl
	for _, child := range c.children {
		if mc, ok := child.(*mockCell); ok {
			mc.setPlaneSeen(sl, seen)
		}
	}
}

type mockSlice struct {
	mu        sync.Mutex
	id        string
	records   map[string]skyra.ActivationRecord
	threshold float64
	count     int
	tree      skyra.TraversalPair
	result    any
	maps      map[string]any
}

func newSlice(id string, threshold float64) *mockSlice {
	return &mockSlice{id: id, records: map[string]skyra.ActivationRecord{}, threshold: threshold, maps: map[string]any{}}
}

func (s *mockSlice) ID() string                                 { return s.id }
func (s *mockSlice) Tree() *skyra.TraversalPair                 { return &s.tree }
func (s *mockSlice) Result() any                                { return s.result }
func (s *mockSlice) SetResult(result any)                       { s.result = result }
func (s *mockSlice) Maps() map[string]any                       { return s.maps }
func (s *mockSlice) Records() map[string]skyra.ActivationRecord { return s.records }
func (s *mockSlice) Threshold() float64                         { return s.threshold }
func (s *mockSlice) Count() int                                 { return s.count }
func (s *mockSlice) SetCount(c int)                             { s.count = c }
func (s *mockSlice) Lock()                                      { s.mu.Lock() }
func (s *mockSlice) Unlock()                                    { s.mu.Unlock() }

// run wires a being's own wave: the root both owns and is traversed.
func run(root *mockCell, sl *mockSlice) *relation.Relation {
	r := relation.NewRelation()
	root.setPlane(sl)
	r.Realize(root, root, skyra.Impulse{Return: true})
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

	// reaching here at all means it terminated. This helper runs a marked
	// return wave, so there is no second return-perception wave.
	if len(root.observes) != 1 {
		t.Errorf("root observed %d times, want 1 (traversal intake)", len(root.observes))
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

func TestImpulseAmplitudeSetsInitialWaveStrength(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	root.link(a)
	sl := newSlice("being", 0.7)
	root.setPlane(sl)

	r := relation.NewRelation()
	r.Realize(root, root, skyra.Impulse{Primitive: "look", Amplitude: 2.0, Return: true})

	if len(a.observes) != 1 {
		t.Fatalf("a observed %d times, want 1 from amplified impulse", len(a.observes))
	}
	if math.Abs(sl.Tree().Current[0].Children[0].Amplitude-1.8) > 0.000001 {
		t.Fatalf("child amplitude = %v, want 1.8", sl.Tree().Current[0].Children[0].Amplitude)
	}
}

// Branching divides amplitude across candidate children. A wide wave should
// dissipate instead of duplicating energy into every survivor.
func TestBranchingDividesAmplitude(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	b := newCell("b", 1.0)
	root.link(a)
	root.link(b)

	run(root, newSlice("being", 0.5))

	if len(a.observes) != 0 {
		t.Errorf("a observed %d times, want 0 (split score below threshold)", len(a.observes))
	}
	if len(b.observes) != 0 {
		t.Errorf("b observed %d times, want 0 (split score below threshold)", len(b.observes))
	}
	if a.rec.calls != 1 {
		t.Errorf("a activated %d times, want 1 (scored once)", a.rec.calls)
	}
	if b.rec.calls != 1 {
		t.Errorf("b activated %d times, want 1 (scored once)", b.rec.calls)
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
	t.Skip("Perception Invariant SUSPENDED for the sensorimotor-boundary build: the kernel " +
		"returns void (returnToOrigin not called). Restore alongside the return wave. See " +
		"specs/2026-06-16/skyra-dna/sensorimotor.md §5 and the perception-invariant suspension banner.")
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	root.link(a)

	sl := newSlice("being", 0)
	root.setPlane(sl)
	r := relation.NewRelation()
	r.Realize(root, root, skyra.Impulse{Primitive: "perceive", ResultMap: "perception"})

	var ret skyra.Impulse
	found := false
	for _, o := range root.observes {
		if imp, ok := o.impulse.(skyra.Impulse); ok && imp.Return {
			ret = imp
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no return perception impulse observed; observes = %+v", root.observes)
	}
	result, ok := ret.Payload.(map[string]any)
	if !ok {
		t.Fatalf("return payload type = %T, want map[string]any", ret.Payload)
	}
	perception, ok := result["perception"].(map[string]any)
	if !ok {
		t.Fatalf("perception map type = %T, want map[string]any", result["perception"])
	}
	if perception["root"] != "root" {
		t.Errorf("delivered root result = %v, want \"root\"", perception["root"])
	}
	if perception["a"] != "a" {
		t.Errorf("delivered child result = %v, want \"a\"", perception["a"])
	}
}

// Observe carries the wave OWNER (same meaning as Express's origin), not nil.
func TestTraversalObserveUsesOwner(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	b := newCell("b", 1.0)
	root.link(a)
	a.link(b) // owner is root, two levels deep

	run(root, newSlice("being", 0))

	for _, cell := range []*mockCell{a, b} {
		if len(cell.observes) == 0 || cell.observes[0].origin != "root" {
			t.Errorf("%s saw origin %v, want owner \"root\" during traversal", cell.id, cell.observes)
		}
	}
}

// safeCell is a minimal concurrency-safe cell for the kernel concurrency test:
// its Observe/Express touch no cell state, so a race the detector reports is the
// KERNEL's slice bookkeeping, not the cell's.
type safeCell struct {
	id       string
	children map[string]skyra.Space
	rec      skyra.ActivationRecord
	slice    skyra.Slice
}

func newSafeCell(id string, base float64) *safeCell {
	return &safeCell{id: id, children: map[string]skyra.Space{}, rec: &mockRecord{base: base}}
}
func (c *safeCell) ID() string                                  { return c.id }
func (c *safeCell) Slice() skyra.Slice                          { return c.slice }
func (c *safeCell) NewActivation() skyra.ActivationRecord       { return c.rec }
func (c *safeCell) Observe(origin skyra.Space, impulse any)     {}
func (c *safeCell) Express(origin skyra.Space, impulse any) any { return nil }
func (c *safeCell) Children() map[string]skyra.Space            { return c.children }
func (c *safeCell) Content() any                                { return nil }
func (c *safeCell) Agents() map[string]skyra.Space              { return nil }

// TestConcurrentWavesShareOneSlice fires many waves at the same slice at once.
// With the wave-wide lock removed and only bookkeeping bursts locked, the kernel
// state (Count tick, tree double-buffer, Records map + Activate) must stay
// race-free. This is the substrate pass's kernel-level red/green (run -race).
func TestConcurrentWavesShareOneSlice(t *testing.T) {
	root := newSafeCell("root", 1.0)
	a := newSafeCell("a", 1.0)
	b := newSafeCell("b", 1.0)
	root.children["a"] = a
	root.children["b"] = b
	sl := newSlice("being", 0)
	root.slice, a.slice, b.slice = sl, sl, sl

	r := relation.NewRelation()
	const waves = 64
	var wg sync.WaitGroup
	wg.Add(waves)
	for i := 0; i < waves; i++ {
		go func() {
			defer wg.Done()
			r.Realize(root, root, skyra.Impulse{Return: true})
		}()
	}
	wg.Wait()

	if got := sl.Count(); got != waves {
		t.Fatalf("count = %d, want %d (every wave must tick proper time exactly once)", got, waves)
	}
}

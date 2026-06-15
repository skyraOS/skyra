package relation

import (
	"math"
	"sync"
	"testing"
	"time"

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
	id             string
	children       map[string]skyra.Space
	rec            *mockRecord
	slice          skyra.Slice
	content        any
	observes       []obs // every Observe call: nil during traversal, owner on delivery
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
func run(root *mockCell, sl *mockSlice) *Relation {
	r := NewRelation()
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

	r := NewRelation()
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
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	root.link(a)

	sl := newSlice("being", 0)
	root.setPlane(sl)
	r := NewRelation()
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

// A Content Space is another actor: the crossing spawns an independent wave
// owned by the inner actor. The parent does not wait and does not receive the
// inner result.
func TestCrossingSpawnsIndependentWave(t *testing.T) {
	root := newCell("root", 1.0)
	inner := newCell("inner", 1.0)
	root.content = inner // root's Content is another being → boundary crossing
	root.setPlane(newSlice("root", 0))
	inner.setPlane(newSlice("inner", 0))
	inner.blockExpress = make(chan struct{})
	inner.expressStarted = make(chan struct{})

	done := make(chan struct{})
	parentReturned := make(chan struct{})
	var once sync.Once
	inner.onObserve = func(o obs) {
		if imp, ok := o.impulse.(skyra.Impulse); ok && imp.Return && resultHas(imp.Payload, "perceive", "inner", "inner") {
			once.Do(func() { close(done) })
		}
	}

	r := NewRelation()
	go func() {
		r.Realize(root, root, skyra.Impulse{Primitive: "perceive"})
		close(parentReturned)
	}()

	select {
	case <-inner.expressStarted:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for inner wave to start; observes = %+v", inner.observes)
	}

	select {
	case <-parentReturned:
	case <-time.After(time.Second):
		t.Fatalf("parent wave waited on blocked child wave")
	}

	close(inner.blockExpress)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for inner wave delivery; observes = %+v", inner.observes)
	}

	if len(inner.observes) < 2 {
		t.Fatalf("inner observes = %+v, want intake + return perception", inner.observes)
	}
	if inner.observes[0].origin != "<nil>" {
		t.Errorf("inner intake origin = %q, want <nil>", inner.observes[0].origin)
	}
	hasReturn := false
	for _, o := range inner.observes {
		if imp, ok := o.impulse.(skyra.Impulse); ok && imp.Return && resultHas(imp.Payload, "perceive", "inner", "inner") {
			hasReturn = true
		}
	}
	if !hasReturn {
		t.Errorf("inner observes = %+v, want return perception with result[inner] inner", inner.observes)
	}

	gotInner := false
	for _, o := range root.observes {
		if o.impulse == "inner" {
			gotInner = true
		}
	}
	if gotInner {
		t.Errorf("parent received child result; observes = %+v", root.observes)
	}
}

func resultHas(value any, mapKey string, resultKey string, want any) bool {
	result, ok := value.(map[string]any)
	if !ok {
		return false
	}
	results, ok := result[mapKey].(map[string]any)
	if !ok {
		return false
	}
	return results[resultKey] == want
}

// In-plane traversal Observe calls carry nil; owner appears only on delivery.
func TestTraversalObserveUsesNil(t *testing.T) {
	root := newCell("root", 1.0)
	a := newCell("a", 1.0)
	b := newCell("b", 1.0)
	root.link(a)
	a.link(b) // owner is root, two levels deep

	run(root, newSlice("being", 0))

	for _, cell := range []*mockCell{a, b} {
		if len(cell.observes) == 0 || cell.observes[0].origin != "<nil>" {
			t.Errorf("%s saw origin %v, want nil during traversal", cell.id, cell.observes)
		}
	}
}

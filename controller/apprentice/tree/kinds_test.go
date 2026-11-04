// The tests for what a node is and what an observation is (56, "What a node
// is"; "Tested ferociously"), each written out apart from the code and
// checked against it: by count, exhaustively, and fuzzed.
package tree

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// --- the world ---

func hash(c byte) Binary { return Binary("sha256:" + strings.Repeat(string(c), 64)) }

var (
	shapeA = Schema(`{"type":"object"}`)
	shapeB = Schema(`{"properties":{"path":{"type":"string"}},"type":"object"}`)
	shapeS = Schema(`{"type":"string"}`)

	binApp  = hash('a') // the apprentice
	binSelf = hash('b') // the agent being classified
	binX    = hash('c')
	binY    = hash('d')
	binZ    = hash('e')
)

func must[T any](v T, ok bool) T {
	if !ok {
		panic("the world did not build")
	}
	return v
}

func contract(b Binary, s Schema) Contract { return must(Identify(b, s)) }

func apprentice() Apprentice {
	shape := func(title string) Schema { return Schema(`{"title":"` + title + `","type":"object"}`) }
	c := func(title string) Contract { return contract(binApp, shape(title)) }
	return must(ApprenticeOf(binApp, c("observe"), c("root"), c("exit"), c("return"),
		c("perform_effect"), c("create_input"), c("delegate"), c("goal"),
		c("create_delegation"), c("create_act"), c("package_reach")))
}

func (a Apprentice) all() []Contract {
	return a.Contracts()
}

var (
	app       = apprentice()
	self      = contract(binSelf, shapeA)
	selfOther = contract(binSelf, shapeB) // another shape of the agent's own binary
	x         = contract(binX, shapeA)
	x2        = contract(binX, shapeB) // a second shape of one binary: a second contract
	y         = contract(binY, shapeA)
	z         = contract(binZ, shapeA)

	others = []Contract{x, x2, y, z}
	pool   = append(append([]Contract{self, selfOther}, others...), app.all()...)
)

var kinds = []Kind{KindAct, KindDelegate}

// --- the spec, written out apart from the code ---

// spec is what the classification must be, computed here without the code:
// the apprentice's contracts never count; refused if the child's own binary
// is among the rest, at every count; refused if none remain, as a delegate
// passes at least one; otherwise by how many distinct contracts remain.
func spec(self Contract, reach []Contract) (Kind, bool) {
	var counted []Contract
	for _, c := range reach {
		if slices.Contains(app.all(), c) {
			continue
		}
		if c.binary == self.binary {
			return 0, false
		}
		if !slices.Contains(counted, c) {
			counted = append(counted, c)
		}
	}
	switch len(counted) {
	case 0:
		return 0, false
	case 1:
		return KindAct, true
	default:
		return KindDelegate, true
	}
}

// reaches is every list of length 0 to n drawn from xs, repeats allowed.
func reaches(xs []Contract, n int) [][]Contract {
	out := [][]Contract{{}}
	layer := [][]Contract{{}}
	for range n {
		var next [][]Contract
		for _, r := range layer {
			for _, c := range xs {
				next = append(next, append(slices.Clone(r), c))
			}
		}
		out = append(out, next...)
		layer = next
	}
	return out
}

func name(c Contract) string {
	for _, p := range []struct {
		c Contract
		n string
	}{{self, "self"}, {selfOther, "selfOther"}, {x, "x"}, {x2, "x2"}, {y, "y"}, {z, "z"},
		{app.Observe(), "app.observe"}, {app.Root(), "app.root"}, {app.Exit(), "app.exit"},
		{app.Return(), "app.return"}, {app.PerformEffect(), "app.perform_effect"},
		{app.CreateInput(), "app.create_input"},
		{app.Delegate(), "app.delegate"}, {app.Goal(), "app.goal"},
		{app.CreateDelegation(), "app.create_delegation"}, {app.CreateAct(), "app.create_act"},
		{app.Reach(), "app.package_reach"}} {
		if p.c == c {
			return p.n
		}
	}
	return string(c.binary)
}

func names(cs []Contract) string {
	ns := make([]string, len(cs))
	for i, c := range cs {
		ns[i] = name(c)
	}
	return "[" + strings.Join(ns, ", ") + "]"
}

// --- what a node is ---

// classify_act_iff, classify_delegate_iff: the kind is the count of distinct
// contracts in reach.
func TestKindByCount(t *testing.T) {
	for _, tc := range []struct {
		reach []Contract
		want  Kind
	}{
		{[]Contract{x}, KindAct},
		{[]Contract{x, y}, KindDelegate},
		{[]Contract{x, y, z}, KindDelegate},
		{[]Contract{x, x2, y, z}, KindDelegate},
	} {
		got, ok := Classify(app, self, tc.reach)
		if !ok || got != tc.want {
			t.Errorf("Classify(%s) = %v, %v; want %v", names(tc.reach), got, ok, tc.want)
		}
	}
}

// Duplicates count once.
func TestDuplicatesCountOnce(t *testing.T) {
	for _, tc := range []struct {
		reach []Contract
		want  Kind
	}{
		{[]Contract{x, x}, KindAct},
		{[]Contract{x, x, x, x}, KindAct},
		{[]Contract{x, y, x, y}, KindDelegate},
	} {
		if got, ok := Classify(app, self, tc.reach); !ok || got != tc.want {
			t.Errorf("Classify(%s) = %v, %v; want %v", names(tc.reach), got, ok, tc.want)
		}
	}
}

// The count compares contracts: two shapes of one binary are two contracts.
func TestTwoShapesOfOneBinaryAreTwo(t *testing.T) {
	if got, ok := Classify(app, self, []Contract{x, x2}); !ok || got != KindDelegate {
		t.Errorf("Classify([x, x2]) = %v, %v; want delegate", got, ok)
	}
}

// classify_none_iff, and the self check before the count: an agent whose own
// binary is in its reach is refused at every count, 0 and 1 included, and
// through any shape of its binary.
func TestRefusedFirstAtEveryCount(t *testing.T) {
	for _, own := range []Contract{self, selfOther} {
		for n := 0; n <= len(others); n++ {
			reach := append([]Contract{own}, others[:n]...)
			if k, ok := Classify(app, self, reach); ok {
				t.Errorf("Classify(%s) = %v; want refused (own binary, %d others)", names(reach), k, n)
			}
		}
	}
}

// The apprentice's contracts count for neither the count nor the self check:
// adding any of them to any reach leaves the result unchanged, refused
// included.
func TestApprenticeContractsNeverCount(t *testing.T) {
	subsets := reaches(app.all(), 2)
	for _, reach := range reaches([]Contract{self, selfOther, x, x2, y}, 3) {
		wantK, wantOK := Classify(app, self, reach)
		for _, add := range subsets {
			for _, r := range [][]Contract{append(slices.Clone(reach), add...), append(slices.Clone(add), reach...)} {
				k, ok := Classify(app, self, r)
				if k != wantK || ok != wantOK {
					t.Fatalf("Classify(%s) = %v, %v; without the apprentice's %s it is %v, %v",
						names(r), k, ok, names(add), wantK, wantOK)
				}
			}
		}
	}
}

// classify_passes_one: a delegate passes at least one contract. A child with
// nothing in reach, or only the apprentice's, is refused.
func TestDelegatePassesOne(t *testing.T) {
	for _, reach := range [][]Contract{nil, app.all()} {
		if k, ok := Classify(app, self, reach); ok {
			t.Errorf("Classify(%s) = %v; want refused", names(reach), k)
		}
	}
}

// The root's kind is given, never counted: any reach, nothing in it
// included, gives the kind it was started with, unless its own binary is in
// it.
func TestRootKindGiven(t *testing.T) {
	for _, reach := range reaches(append([]Contract{self, selfOther}, others...), 3) {
		_, selfOK := Classify(app, self, append(slices.Clone(reach), x))
		for _, k := range kinds {
			got, ok := RootKind(app, self, reach, k)
			if ok != selfOK || ok && got != k {
				t.Fatalf("RootKind(%s, %v) = %v, %v; want %v, %v", names(reach), k, got, ok, k, selfOK)
			}
		}
	}
}

// parts_apart: an act child is described by create_act and handed no reach;
// a delegate child is described by create_delegation and handed its reach.
func TestParts(t *testing.T) {
	want := map[Kind][]Contract{
		KindAct:      {app.CreateAct()},
		KindDelegate: {app.CreateDelegation(), app.Reach()},
	}
	for _, k := range kinds {
		if got := Parts(app, k); !slices.Equal(got, want[k]) {
			t.Errorf("Parts(%v) = %s; want %s", k, names(got), names(want[k]))
		}
	}
}

// Totality: every reach of length 0 to 4 over the whole pool classifies as
// exactly what the spec says, one of refused, act or delegate.
func TestTotality(t *testing.T) {
	n := 0
	for _, reach := range reaches(pool, 4) {
		k, ok := Classify(app, self, reach)
		wk, wok := spec(self, reach)
		if ok != wok || (ok && k != wk) {
			t.Fatalf("Classify(%s) = %v, %v; want %v, %v", names(reach), k, ok, wk, wok)
		}
		if ok && !slices.Contains(kinds, k) {
			t.Fatalf("Classify(%s) = %v: not one of the two kinds", names(reach), k)
		}
		n++
	}
	t.Logf("%d reaches", n)
}

// --- what the kind loads ---

// The apprentice's contracts for the kind: exit and return for every kind;
// create_input and perform_effect for an act node; delegate and goal for a
// delegate; observe and root never.
func TestLoadedByKind(t *testing.T) {
	want := map[Kind][]Contract{
		KindAct:      {app.Exit(), app.Return(), app.CreateInput(), app.PerformEffect()},
		KindDelegate: {app.Exit(), app.Return(), app.Delegate(), app.Goal()},
	}
	for _, k := range kinds {
		got := Loaded(app, k)
		for _, c := range app.all() {
			if slices.Contains(got, c) != slices.Contains(want[k], c) {
				t.Errorf("Loaded(%v) holds %s: %v; want %v", k, name(c), slices.Contains(got, c), slices.Contains(want[k], c))
			}
		}
		for _, c := range got {
			if !slices.Contains(app.all(), c) {
				t.Errorf("Loaded(%v) holds %s, which is not the apprentice's", k, name(c))
			}
		}
		for _, never := range []Contract{app.Observe(), app.Root()} {
			if slices.Contains(got, never) {
				t.Errorf("Loaded(%v) holds %s, which is privileged", k, name(never))
			}
		}
	}
}

// Loading the kind's contracts does not change the kind.
func TestLoadingDoesNotChangeTheKind(t *testing.T) {
	for _, reach := range reaches(others, 3) {
		k, ok := Classify(app, self, reach)
		if !ok {
			continue
		}
		loaded := append(slices.Clone(reach), Loaded(app, k)...)
		if k2, ok2 := Classify(app, self, loaded); !ok2 || k2 != k {
			t.Fatalf("Classify(%s) = %v; with its loaded contracts %s it is %v, %v", names(reach), k, names(loaded), k2, ok2)
		}
	}
}

// --- what an observation is ---

func typeOf(c Contract, s Schema) ObservationType { return must(TypeOf(c, s)) }

// An effect made through each contract its kind may produce through.
func effectOf(t *testing.T, k Kind, c Contract) Observation {
	t.Helper()
	o, ok := Observe(app, k, typeOf(c, shapeS), nil)
	if !ok {
		t.Fatalf("Observe(%v, %s, no members) refused", k, name(c))
	}
	return o
}

// An observation is an effect iff it has no members, a computation iff it
// has members; never both, never neither.
func TestEffectIffNoMembers(t *testing.T) {
	goal := effectOf(t, KindDelegate, app.Goal())
	exit := effectOf(t, KindAct, app.Exit())
	wrote := effectOf(t, KindDelegate, app.CreateAct())
	came := effectOf(t, KindAct, x)
	for _, o := range []Observation{goal, exit, wrote, came} {
		if len(o.Members()) != 0 || !o.IsEffect() || o.IsComputation() {
			t.Errorf("an observation with %d members: effect %v, computation %v", len(o.Members()), o.IsEffect(), o.IsComputation())
		}
	}
	for _, ms := range [][]Observation{{goal}, {came}, {goal, came}, {wrote, exit}} {
		o, ok := Observe(app, KindDelegate, typeOf(app.Delegate(), shapeA), ms)
		if !ok {
			t.Fatalf("a delegate's computation of %d members refused", len(ms))
		}
		if len(o.Members()) == 0 || o.IsEffect() || !o.IsComputation() {
			t.Errorf("an observation with %d members: effect %v, computation %v", len(o.Members()), o.IsEffect(), o.IsComputation())
		}
	}
}

// Only a delegate makes a computation: with members, an act node is
// refused.
func TestOnlyDelegateMakesComputation(t *testing.T) {
	came := effectOf(t, KindAct, x)
	for _, k := range kinds {
		_, ok := Observe(app, k, typeOf(app.Delegate(), shapeA), []Observation{came})
		if ok != (k == KindDelegate) {
			t.Errorf("Observe(%v, with members) = %v; want %v", k, ok, k == KindDelegate)
		}
	}
}

// Which effects each kind may make: exit, every kind; goal, the two
// descriptions and the packaged reach, a delegate only; an input, and through
// a contract outside the apprentice, an act node only. Through observe or root, none;
// through return, delegate or perform_effect, none: return passes, delegate
// composes, and what perform_effect brings back is typed by the contract
// passed to it.
func TestEffectsByKind(t *testing.T) {
	may := map[Kind][]Contract{
		KindAct:      {app.Exit(), app.CreateInput(), x},
		KindDelegate: {app.Exit(), app.Goal(), app.CreateDelegation(), app.CreateAct(), app.Reach()},
	}
	for _, k := range kinds {
		for _, c := range append(app.all(), x) {
			_, ok := Observe(app, k, typeOf(c, shapeS), nil)
			if want := slices.Contains(may[k], c); ok != want {
				t.Errorf("Observe(%v, through %s, no members) = %v; want %v", k, name(c), ok, want)
			}
		}
	}
}

// --- what a return may hold ---

// mayReturnSpec is the rule, written out apart from the code: a return holds
// something, and nothing its model created directly.
func mayReturnSpec(created, returned []Observation) bool {
	if len(returned) == 0 {
		return false
	}
	for _, r := range returned {
		for _, c := range created {
			if equal(r, c) {
				return false
			}
		}
	}
	return true
}

func equal(a, b Observation) bool {
	if a.IsEffect() != b.IsEffect() || len(a.Members()) != len(b.Members()) {
		return false
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

// The observations of the return cases, each of its own type, so none is
// equal to another by accident.
type returnWorld struct {
	goal  Observation // a delegate's model made it, through goal
	wrote Observation // a delegate's model made it, through create_act
	came  Observation // came back through an act node's perform_effect
	child Observation // a child's, passed back up
	other Observation // another child's
}

func newReturnWorld(t *testing.T) returnWorld {
	t.Helper()
	return returnWorld{
		goal:  effectOf(t, KindDelegate, app.Goal()),
		wrote: effectOf(t, KindDelegate, app.CreateAct()),
		came:  effectOf(t, KindAct, x),
		child: effectOf(t, KindAct, y),
		other: effectOf(t, KindAct, z),
	}
}

// A return holds no observation its model created directly: one that does is
// refused. An empty return is no return.
func TestMayReturn(t *testing.T) {
	w := newReturnWorld(t)
	for _, tc := range []struct {
		name     string
		created  []Observation
		returned []Observation
		want     bool
	}{
		{"act returns what came back", nil, []Observation{w.came}, true},
		{"act returns what it was passed", nil, []Observation{w.child}, true},
		{"delegate returns a child's", []Observation{w.goal}, []Observation{w.child}, true},
		{"delegate returns two children's", []Observation{w.goal}, []Observation{w.child, w.other}, true},
		{"delegate returns its own goal", []Observation{w.goal}, []Observation{w.goal}, false},
		{"delegate hides its goal among children's", []Observation{w.goal}, []Observation{w.child, w.goal, w.other}, false},
		{"returns what its model created", []Observation{w.wrote}, []Observation{w.wrote}, false},
		{"hides what its model created", []Observation{w.wrote}, []Observation{w.came, w.wrote}, false},
		{"act returns nothing", nil, nil, false},
		{"delegate returns nothing", []Observation{w.goal}, nil, false},
	} {
		if got := MayReturn(tc.created, tc.returned); got != tc.want {
			t.Errorf("%s: MayReturn = %v; want %v", tc.name, got, tc.want)
		}
	}
}

// Every created set and every return over the world, up to three
// observations each: the code agrees with the rule.
func TestMayReturnExhaustive(t *testing.T) {
	w := newReturnWorld(t)
	obs := []Observation{w.goal, w.wrote, w.came, w.child, w.other}
	lists := observationLists(obs, 3)
	n := 0
	for _, created := range lists {
		for _, returned := range lists {
			if got, want := MayReturn(created, returned), mayReturnSpec(created, returned); got != want {
				t.Fatalf("MayReturn(%d created, %d returned) = %v; want %v", len(created), len(returned), got, want)
			}
			n++
		}
	}
	t.Logf("%d cases", n)
}

func observationLists(xs []Observation, n int) [][]Observation {
	out := [][]Observation{{}}
	layer := [][]Observation{{}}
	for range n {
		var next [][]Observation
		for _, r := range layer {
			for _, o := range xs {
				next = append(next, append(slices.Clone(r), o))
			}
		}
		out = append(out, next...)
		layer = next
	}
	return out
}

// --- fuzzed ---

// The classification on reaches no one enumerated: it agrees with the spec,
// and the apprentice's contracts never move it.
func FuzzClassify(f *testing.F) {
	f.Add([]byte{}, []byte{})
	f.Add([]byte{2}, []byte{6})
	f.Add([]byte{0, 2, 3}, []byte{6, 7, 8})
	f.Add([]byte{1, 1, 1}, []byte{})
	f.Fuzz(func(t *testing.T, reachIdx, addIdx []byte) {
		pick := func(idx []byte, from []Contract) []Contract {
			var out []Contract
			for _, i := range idx {
				out = append(out, from[int(i)%len(from)])
			}
			return out
		}
		reach := pick(reachIdx, pool)
		k, ok := Classify(app, self, reach)
		wk, wok := spec(self, reach)
		if ok != wok || (ok && k != wk) {
			t.Fatalf("Classify(%s) = %v, %v; want %v, %v", names(reach), k, ok, wk, wok)
		}
		withApp := append(slices.Clone(reach), pick(addIdx, app.all())...)
		if k2, ok2 := Classify(app, self, withApp); k2 != k || ok2 != ok {
			t.Fatalf("Classify(%s) = %v, %v; without the apprentice's it is %v, %v", names(withApp), k2, ok2, k, ok)
		}
	})
}

// The return rule on inputs no one enumerated.
func FuzzMayReturn(f *testing.F) {
	f.Add([]byte{}, []byte{})
	f.Add([]byte{0}, []byte{3, 0})
	f.Add([]byte{1}, []byte{2})
	f.Fuzz(func(t *testing.T, createdIdx, returnedIdx []byte) {
		w := newReturnWorld(t)
		obs := []Observation{w.goal, w.wrote, w.came, w.child, w.other}
		pick := func(idx []byte) []Observation {
			var out []Observation
			for _, i := range idx {
				out = append(out, obs[int(i)%len(obs)])
			}
			return out
		}
		created, returned := pick(createdIdx), pick(returnedIdx)
		if got, want := MayReturn(created, returned), mayReturnSpec(created, returned); got != want {
			t.Fatalf("MayReturn(%d created, %d returned) = %v; want %v", len(created), len(returned), got, want)
		}
	})
}

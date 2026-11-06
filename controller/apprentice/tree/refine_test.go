package tree

import (
	"encoding/json"
	"errors"
	"testing"
)

// The replays of Refine.lean's vectors, the refine* files.
func init() {
	replays["refineOverlaps"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		overlap, decided, err := Overlaps(dec[string](t, in["p"]), dec[string](t, in["q"]))
		if err != nil {
			t.Fatal(err)
		}
		if !decided {
			return nil
		}
		return overlap
	}
	replays["refineCommit"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		v, recs, poss, err := CommitGuard(dec[Version](t, in["refinement"]), dec[string](t, in["guard"]))
		if err != nil {
			t.Fatal(err)
		}
		return map[string]any{"refinement": v, "records": eRecords(recs), "possible": ePossible(poss)}
	}
	replays["refineDistinguish"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		s, err := Distinguish(dec[Version](t, in["refinement"]).Branches, xVals(t, in))
		if err != nil {
			t.Fatal(err)
		}
		return eSelection(s)
	}
	replays["refineResolve"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		s, v, recs, err := Resolve(dec[Version](t, in["refinement"]), xVals(t, in))
		if err != nil {
			t.Fatal(err)
		}
		return map[string]any{"outcome": eSelection(s), "refinement": v, "records": eRecords(recs)}
	}
	replays["refineAfterStartle"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		chain, err := ParseGuard(dec[string](t, in["guard"]))
		if err != nil {
			t.Fatal(err)
		}
		a, err := AfterStartle(chain, dec[Version](t, in["committed"]).Branches, xVals(t, in))
		if err != nil {
			t.Fatal(err)
		}
		if a.Startle {
			return "startle"
		}
		return map[string]any{"trust": a.Instr, "learned": a.Learned}
	}
}

// xVals is the vectors' x as the arrival's size field.
func xVals(t *testing.T, in map[string]json.RawMessage) map[string]any {
	return sized(dec[int64](t, in["x"]))
}

func sized(x any) map[string]any {
	return map[string]any{"arrived": map[string]any{"size": x}}
}

func eSelection(s Selection) any {
	switch s.Kind {
	case SelectOne:
		return map[string]any{"select": s.Instr}
	case SelectConflict:
		return "conflict"
	}
	return nil
}

func eRecords(rs []Reconcile) []any {
	out := []any{}
	for _, r := range rs {
		out = append(out, r)
	}
	return out
}

func ePossible(ps []Possible) []any {
	out := []any{}
	for _, p := range ps {
		out = append(out, []int{p.Older, p.Newer})
	}
	return out
}

// --- 58's checks ---

func commitOK(t *testing.T, v Version, guard string) (Version, []Reconcile, []Possible) {
	t.Helper()
	v2, recs, poss, err := CommitGuard(v, guard)
	if err != nil {
		t.Fatalf("%s: %v", guard, err)
	}
	return v2, recs, poss
}

func selects(t *testing.T, v Version, x any) Selection {
	t.Helper()
	s, err := Distinguish(v.Branches, sized(x))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// The older keeps the overlap; the values that triggered startle select
// the new instruction; nothing the older selected changes.
func TestOlderKeepsTheOverlap(t *testing.T) {
	v, _, _ := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "A" : null`)
	v, recs, _ := commitOK(t, v, `arrived.size > 2 ? "B" : null`)
	if len(recs) != 1 || recs[0] != (Reconcile{Older: 1, Retired: 2, Remainder: 3}) {
		t.Fatalf("records %v", recs)
	}
	if b := v.Branches[len(v.Branches)-1]; b.ID != 3 || b.Rank != 2 {
		t.Fatalf("the remainder is b%d with rank %d, not b3 with b2's rank", b.ID, b.Rank)
	}
	for x, want := range map[int64]string{3: "B", 7: "B", 8: "A", 100: "A"} {
		if s := selects(t, v, x); s.Kind != SelectOne || s.Instr != want {
			t.Errorf("x = %d selects %+v, want %s", x, s, want)
		}
	}
	if s := selects(t, v, int64(1)); s.Kind != SelectNull {
		t.Errorf("x = 1 selects %+v, want null", s)
	}
}

// One broad guard cannot erase what other trees learned: `true` narrows to
// what no branch covered.
func TestBroadGuardNarrows(t *testing.T) {
	v, _, _ := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "ls $1" : arrived.size < 0 ? "stat $1" : null`)
	v, _, poss := commitOK(t, v, `true ? "rm $1" : null`)
	if len(poss) != 0 {
		t.Fatalf("possibly overlapping %v: every overlap here is decidable", poss)
	}
	for x, want := range map[float64]string{8: "ls $1", -1: "stat $1", 3: "rm $1", 7.5: "ls $1", 0: "rm $1"} {
		if s := selects(t, v, x); s.Kind != SelectOne || s.Instr != want {
			t.Errorf("x = %v selects %+v, want %s", x, s, want)
		}
	}
}

// A chain is stored as independent branches: within one guard, the first
// condition that holds wins, and retiring one branch widens no other.
func TestChainEffectiveConditions(t *testing.T) {
	v, _, _ := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "A" : arrived.size > 5 ? "C" : null`)
	if s := selects(t, v, int64(9)); s.Instr != "A" || s.Kind != SelectOne {
		t.Fatalf("x = 9 selects %+v", s)
	}
	if s := selects(t, v, int64(6)); s.Instr != "C" || s.Kind != SelectOne {
		t.Fatalf("x = 6 selects %+v", s)
	}
	rest := Version{Next: v.Next, Branches: v.Branches[1:]}
	if s := selects(t, rest, int64(9)); s.Kind != SelectNull {
		t.Fatalf("with b1 retired, x = 9 selects %+v: b2 widened", s)
	}
}

// Overlap is decided over numbers, not integers, and left undecided past
// simple comparisons of one variable.
func TestOverlaps(t *testing.T) {
	for _, c := range []struct {
		p, q             string
		overlap, decided bool
	}{
		{"arrived.size > 7", "arrived.size < 8", true, true},
		{"arrived.size > 7", "arrived.size < 7", false, true},
		{"arrived.size >= 7", "arrived.size <= 7", true, true},
		{"arrived.size == 5", "!(arrived.size == 5)", false, true},
		{"true", "arrived.size > 1000", true, true},
		{"arrived.size % 2 == 0", "arrived.size > 1", false, false},
		{"arrived.size > 1", "passed0.size > 1", false, false},
		{"size(arrived.name) > 1", "arrived.size > 1", false, false},
	} {
		o, d, err := Overlaps(c.p, c.q)
		if err != nil || o != c.overlap || d != c.decided {
			t.Errorf("%s ∧ %s: overlap %v decided %v (%v), want %v %v", c.p, c.q, o, d, err, c.overlap, c.decided)
		}
	}
}

// A conflict found later, from a pair commit could not decide, is
// reconciled with the older keeping it, and resolving ends in no conflict.
func TestResolveLaterConflict(t *testing.T) {
	v, _, _ := commitOK(t, EmptyVersion(), `arrived.size > 2 ? "A" : null`)
	v, _, poss := commitOK(t, v, `arrived.size % 2 == 0 ? "B" : null`)
	if len(poss) != 1 || poss[0] != (Possible{Older: 1, Newer: 2}) {
		t.Fatalf("possible %v", poss)
	}
	if s := selects(t, v, int64(6)); s.Kind != SelectConflict {
		t.Fatalf("x = 6 selects %+v, want a conflict", s)
	}
	s, v2, recs, err := Resolve(v, sized(int64(6)))
	if err != nil || s.Kind != SelectOne || s.Instr != "A" {
		t.Fatalf("resolved to %+v (%v), want A", s, err)
	}
	if len(recs) != 1 || recs[0].Older != 1 || recs[0].Retired != 2 {
		t.Fatalf("records %v", recs)
	}
	if s := selects(t, v2, int64(2)); s.Instr != "B" {
		t.Fatalf("x = 2 lost the newer's selection outside the overlap: %+v", s)
	}
}

// After startle the step runs its own model's instruction, whatever the
// committed version selects; it is learned only if the version selects it.
func TestAfterStartleRunsItsOwn(t *testing.T) {
	chain, err := ParseGuard(`arrived.size > 2 ? "B" : null`)
	if err != nil {
		t.Fatal(err)
	}
	mine, _, _ := commitOK(t, EmptyVersion(), `arrived.size > 2 ? "B" : null`)
	// A concurrent commit won the race with an older branch covering x = 9.
	theirs, _, _ := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "A" : null`)
	raced, _, _ := commitOK(t, theirs, `arrived.size > 2 ? "B" : null`)
	vals := sized(int64(9))
	for _, c := range []struct {
		v       Version
		learned bool
	}{{mine, true}, {raced, false}} {
		a, err := AfterStartle(chain, c.v.Branches, vals)
		if err != nil || a.Startle || a.Instr != "B" || a.Learned != c.learned {
			t.Errorf("after startle %+v (%v), want B learned %v", a, err, c.learned)
		}
	}
	if a, _ := AfterStartle(chain, mine.Branches, sized(int64(1))); !a.Startle {
		t.Errorf("an uncovered value must startle again, got %+v", a)
	}
}

// A learned cycle that never calls the model ends: every transition counts,
// and the exhausted frame returns an error observation.
func TestTransitionBound(t *testing.T) {
	b := TransitionBound{Transitions: 3}
	poll := Transition{StateTrust} // a learned branch: no model call
	var err error
	n := 0
	for err == nil {
		if err = b.Take(poll); err == nil {
			n++
		}
	}
	if !errors.Is(err, ErrTransitionBound) || n != 3 || b.Taken() != 3 {
		t.Fatalf("took %d transitions, ended with %v", n, err)
	}
	if e := b.Exhausted(); string(e["error"].([]byte)) != "transition bound exhausted after 3 transitions" ||
		ShapeOf(e).Enc() != `{"error":bytes}` {
		t.Fatalf("exhausted frame returns %#v", e)
	}
	learned := TransitionBound{Transitions: 1}
	if err := learned.Take(Transition{StateHesitate, StateStartle, StateReject, StateStartle, StateTrust}); err != nil {
		t.Fatalf("one instruction's whole life is one transition: %v", err)
	}
}

func TestTransitionEndsInOneTrust(t *testing.T) {
	for _, tr := range []Transition{{}, {StateStartle}, {StateTrust, StateTrust}, {StateTrust, StateStartle}} {
		b := TransitionBound{Transitions: 10}
		if err := b.Take(tr); !errors.Is(err, ErrTransition) {
			t.Errorf("%v: %v, want ErrTransition", tr, err)
		}
	}
}

func TestParseGuardRefuses(t *testing.T) {
	for _, g := range []string{
		`arrived.size > 5 ? "ls " + arrived.name : null`,
		`arrived.size > 5 ? "A" : "B"`,
		`arrived.size > 5`,
		`arrived.size > 5 ? "A" :`,
	} {
		if _, err := ParseGuard(g); err == nil {
			t.Errorf("%s: parsed", g)
		}
	}
}

// The model reads observations by `$n` and fields by key; the stored guard
// names positions in the base, so two frames holding the arrival at
// different numbers store one guard.
func TestStoreGuard(t *testing.T) {
	g := `$1.size > 7 && $2["content-type"] == "$3" ? "diff $1 $2" : null`
	want := []Link{{
		Cond:  `arrived.size > 7 && passed0["content-type"] == "$3"`,
		Instr: "diff @arrived @passed[0]",
	}}
	for _, scope := range []map[int]Position{
		{1: Arrived, 2: Passed(0)},
		{1: Arrived, 2: Passed(0), 3: Passed(1)},
	} {
		chain, err := StoreGuard(g, scope)
		if err != nil || len(chain) != 1 || chain[0] != want[0] {
			t.Fatalf("stored %+v (%v), want %+v", chain, err, want)
		}
	}
	renumbered, err := StoreGuard(`$7.size > 7 && $4["content-type"] == "$3" ? "diff $7 $4" : null`,
		map[int]Position{7: Arrived, 4: Passed(0)})
	if err != nil || renumbered[0] != want[0] {
		t.Fatalf("another frame's numbers stored %+v (%v)", renumbered, err)
	}
	v, _, _, err := CommitChain(EmptyVersion(), renumbered)
	if err != nil {
		t.Fatal(err)
	}
	vals := map[string]any{"arrived": map[string]any{"size": 9.0},
		"passed0": map[string]any{"content-type": "$3"}}
	if s, err := Distinguish(v.Branches, vals); err != nil || s.Instr != "diff @arrived @passed[0]" {
		t.Fatalf("selects %+v (%v)", s, err)
	}
}

// A `$n` with no position in the base cannot be stored: the instruction
// runs once, unlearned. Literal `$n` in a raw or triple-quoted string is
// text.
func TestStoreGuardUnstorable(t *testing.T) {
	scope := map[int]Position{0: Arrived}
	for _, g := range []string{`$3.size > 1 ? "ls $0" : null`, `$0.size > 1 ? "diff $0 $3" : null`} {
		var u Unstorable
		if _, err := StoreGuard(g, scope); !errors.As(err, &u) || u.N != 3 {
			t.Errorf("%s: %v, want $3 unstorable", g, err)
		}
	}
	for _, g := range []string{
		`$0.name == r"\$3" ? "ls $0" : null`,
		`$0.name == """it's $3""" ? "ls $0" : null`,
		`$0.name == '$3\'s' ? "ls $0" : null`,
	} {
		if _, err := StoreGuard(g, scope); err != nil {
			t.Errorf("%s: %v", g, err)
		}
	}
}

// A stored guard reads positions only: any other variable fails.
func TestOnlyPositions(t *testing.T) {
	v, _, _ := commitOK(t, EmptyVersion(), `goal.size > 1 ? "A" : null`)
	if _, err := Distinguish(v.Branches, map[string]any{"goal": map[string]any{"size": 2}}); err == nil {
		t.Fatal("a guard read a variable that is not a position")
	}
}

// Text that is not JSON arrives as a string, with no wrapper: the guard
// reads `$n` itself.
func TestTextArrival(t *testing.T) {
	chain, err := StoreGuard(`$0.contains("error") ? "less $0" : size($0) > 0 ? "wc $0" : null`,
		map[int]Position{0: Arrived})
	if err != nil {
		t.Fatal(err)
	}
	v, _, _, err := CommitChain(EmptyVersion(), chain)
	if err != nil {
		t.Fatal(err)
	}
	for text, want := range map[string]string{"error: no such file\n": "less @arrived", "a.txt\nb.txt\n": "wc @arrived"} {
		s, err := Distinguish(v.Branches, map[string]any{"arrived": text})
		if err != nil || s.Kind != SelectOne || s.Instr != want {
			t.Errorf("%q selects %+v (%v), want %s", text, s, err, want)
		}
	}
	if s, _ := Distinguish(v.Branches, map[string]any{"arrived": ""}); s.Kind != SelectNull {
		t.Errorf("empty output selects %+v, want null", s)
	}
}

// A guard using a macro is stored as written and evaluates: the positions
// inside a comprehension are declared.
func TestMacroGuard(t *testing.T) {
	chain, err := StoreGuard(`$1.items.exists(i, i.id > 2) ? "ls $1" : null`, map[int]Position{1: Arrived})
	if err != nil || chain[0].Cond != "arrived.items.exists(i, i.id > 2)" {
		t.Fatalf("stored %+v (%v)", chain, err)
	}
	v, _, _, err := CommitChain(EmptyVersion(), chain)
	if err != nil {
		t.Fatal(err)
	}
	items := map[string]any{"arrived": map[string]any{"items": []any{map[string]any{"id": 3.0}}}}
	if s, err := Distinguish(v.Branches, items); err != nil || s.Instr != "ls @arrived" {
		t.Fatalf("selects %+v (%v)", s, err)
	}
}

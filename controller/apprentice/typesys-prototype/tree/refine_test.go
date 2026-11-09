package tree

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// The replays of Refine.lean's vectors, the refine* files.
func init() {
	replays["refineCommit"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		v, err := CommitGuard(dec[Version](t, in["refinement"]), dec[string](t, in["guard"]))
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	replays["refineDistinguish"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		s, err := Distinguish(dec[Version](t, in["refinement"]).Branches, xVals(t, in))
		if err != nil {
			t.Fatal(err)
		}
		return eSelection(s)
	}
	replays["refineTally"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		votes, err := Cast(dec[Version](t, in["refinement"]).Branches, xVals(t, in))
		if err != nil {
			t.Fatal(err)
		}
		return eSelection(tallyOf(t, in).Decide(votes))
	}
	replays["refineAfterStartle"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		chain, err := ParseGuard(dec[string](t, in["guard"]))
		if err != nil {
			t.Fatal(err)
		}
		a, err := AfterStartle(chain, dec[Version](t, in["committed"]).Branches, xVals(t, in), tallyOf(t, in))
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

// tallyOf is the vectors' inhibition and floor.
func tallyOf(t *testing.T, in map[string]json.RawMessage) Tally {
	return Tally{Inhibition: dec[int](t, in["inhibition"]), Floor: dec[int](t, in["floor"])}
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

// --- 58's and 59's checks ---

func commitOK(t *testing.T, v Version, guard string) Version {
	t.Helper()
	v2, err := CommitGuard(v, guard)
	if err != nil {
		t.Fatalf("%s: %v", guard, err)
	}
	return v2
}

func tallies(t *testing.T, v Version, x any, tl Tally) Selection {
	t.Helper()
	votes, err := Cast(v.Branches, sized(x))
	if err != nil {
		t.Fatal(err)
	}
	return tl.Decide(votes)
}

func selects(t *testing.T, v Version, x any) Selection {
	t.Helper()
	s, err := Distinguish(v.Branches, sized(x))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// A chain is stored whole, under one voter; an overlap with an older
// branch is a dissenting vote, never narrowed away, and outside it the new
// chain is the only vote.
func TestOverlapIsAVote(t *testing.T) {
	v := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "A" : null`)
	v = commitOK(t, v, `arrived.size > 2 ? "B" : null`)
	if b := v.Branches[1]; b.ID != 2 || b.Rank != 2 || b.Voter != 2 || b.Cond != "(arrived.size > 2) && (true)" {
		t.Fatalf("the second chain stored as %+v", b)
	}
	plurality := Tally{Inhibition: 1000}
	if s := tallies(t, v, int64(3), Tally{}); s.Kind != SelectOne || s.Instr != "B" {
		t.Errorf("x = 3 tallies %+v, want B", s)
	}
	if s := tallies(t, v, int64(8), plurality); s.Kind != SelectConflict {
		t.Errorf("x = 8, one vote each, tallies %+v, want a conflict at any inhibition", s)
	}
	v = commitOK(t, v, `arrived.size > 0 ? "A" : null`)
	if s := tallies(t, v, int64(8), plurality); s.Kind != SelectOne || s.Instr != "A" {
		t.Errorf("x = 8, two to one, tallies %+v, want A by plurality", s)
	}
	if s := tallies(t, v, int64(8), Tally{}); s.Kind != SelectConflict {
		t.Errorf("x = 8, two to one, tallies %+v, want a conflict when every vote is needed", s)
	}
}

// One broad guard cannot erase what other trees learned: it is one vote
// wherever another voter cast one.
func TestBroadGuardIsOneVote(t *testing.T) {
	v := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "ls $1" : arrived.size < 0 ? "stat $1" : null`)
	v = commitOK(t, v, `true ? "rm $1" : null`)
	for x, want := range map[float64]string{3: "rm $1", 0: "rm $1"} {
		if s := tallies(t, v, x, Tally{Inhibition: 1000}); s.Kind != SelectOne || s.Instr != want {
			t.Errorf("x = %v tallies %+v, want %s", x, s, want)
		}
	}
	for _, x := range []float64{8, -1, 7.5} {
		if s := tallies(t, v, x, Tally{Inhibition: 1000}); s.Kind != SelectConflict {
			t.Errorf("x = %v tallies %+v, want a conflict", x, s)
		}
	}
}

// A chain is stored as branches with effective conditions: within one
// guard, the first condition that holds wins, so a voter casts one vote.
func TestChainEffectiveConditions(t *testing.T) {
	v := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "A" : arrived.size > 5 ? "C" : null`)
	for x, want := range map[int64]string{9: "A", 6: "C"} {
		votes, err := Cast(v.Branches, sized(x))
		if err != nil || len(votes) != 1 || votes[0].Instr != want || votes[0].Voter != 1 {
			t.Fatalf("x = %d casts %+v (%v), want one vote for %s by voter 1", x, votes, err, want)
		}
	}
	rest := Version{Next: v.Next, Branches: v.Branches[1:]}
	if s := selects(t, rest, int64(9)); s.Kind != SelectNull {
		t.Fatalf("without b1, x = 9 selects %+v: b2 widened", s)
	}
}

// The cel/ experiment's ballots (proposal 59's table), at every inhibition.
func TestExperimentTally(t *testing.T) {
	const (
		done   = "return 0 @passed[4]"
		why    = "return 1 @passed[4] @passed[5]"
		noWhy  = "return 1 @passed[4]"
		curl   = "@passed[1] @passed[5].url"
		decide = "startle"
	)
	ballot := func(instrs ...string) []Vote {
		votes := make([]Vote, len(instrs))
		for i, in := range instrs {
			votes[i] = Vote{Voter: i + 1, Branch: i + 1, Rank: i + 1, Instr: in}
		}
		return votes
	}
	rows := []struct {
		name  string
		votes []Vote
		want  map[int]string // by inhibition
	}{
		{"200", ballot(done, done, done, done, done, done, done),
			map[int]string{1000: done, 750: done, 500: done, 250: done, 0: done}},
		{"503, restarts 3", ballot(why, why, why, why, why, why, noWhy),
			map[int]string{1000: why, 750: why, 500: why, 250: why, 0: why}},
		{"503, restarts 1", ballot(why, why, why, why, curl, curl, noWhy),
			map[int]string{1000: why, 750: why, 500: decide, 250: decide, 0: decide}},
	}
	for _, r := range rows {
		for ι, want := range r.want {
			s := Tally{Inhibition: ι}.Decide(r.votes)
			got := s.Instr
			if s.Kind != SelectOne {
				got = decide
			}
			if got != want {
				t.Errorf("%s at ι = %d: %s, want %s", r.name, ι, got, want)
			}
		}
	}
	// At ι = 0 the six with a why and the one without merge at level 3.
	if s := (Tally{}).Decide(rows[1].votes); s.Level != 3 {
		t.Errorf("restarts 3 at ι = 0 decided at level %d, want 3", s.Level)
	}
}

// A tally does not depend on the order of the votes.
func TestTallyOrder(t *testing.T) {
	votes := []Vote{{Rank: 1, Instr: "A"}, {Rank: 2, Instr: "B"}, {Rank: 3, Instr: "A"},
		{Rank: 4, Instr: "@passed[0] x"}, {Rank: 5, Instr: "@passed[0] y"}, {Rank: 6, Instr: "@passed[0] y"}}
	for _, tl := range []Tally{{}, {Inhibition: 500}, {Inhibition: 1000}, {Inhibition: 1000, Floor: 4}} {
		want := tl.Decide(votes)
		for i := range votes {
			rotated := append(append([]Vote{}, votes[i:]...), votes[:i]...)
			if got := tl.Decide(rotated); got != want {
				t.Fatalf("%+v: rotated by %d tallies %+v, not %+v", tl, i, got, want)
			}
		}
	}
}

// An instruction's parts are its form, its head, then the rest.
func TestInstrParts(t *testing.T) {
	for in, want := range map[string]string{
		"return 1 @passed[4] @passed[5]":          "return 1 @passed[4] @passed[5]",
		"@passed[1] @passed[5].url":               "run @passed[1] @passed[5].url",
		"delegate(goal(@passed[0]), @passed[1])":  "delegate @passed[0] @passed[1]",
		"delegate(goal(@passed[0][2]), @arrived)": "delegate @passed[0][2] @arrived",
	} {
		if got := strings.Join(InstrParts(in), " "); got != want {
			t.Errorf("InstrParts(%q) = %q, want %q", in, got, want)
		}
	}
	if GroupKey(2, "return 1 @passed[4]") != GroupKey(2, "return 1 @passed[4] @passed[5]") ||
		GroupKey(2, "return 0 @passed[4]") == GroupKey(2, "return 1 @passed[4]") ||
		GroupKey(1, "@passed[1] @arrived") == GroupKey(1, "return 0 @arrived") {
		t.Error("level 2 groups by form and head, level 1 by form")
	}
}

// A ballot names the base, the values by identity, and each vote's voter
// and branch.
func TestBallotIdentity(t *testing.T) {
	v := commitOK(t, EmptyVersion(), `arrived.size > 2 ? "A" : null`)
	v = commitOK(t, v, `arrived.size > 7 ? "B" : arrived.size > 1 ? "A" : null`)
	b, err := NewBallot("base", v.Branches, sized(int64(3)))
	if err != nil {
		t.Fatal(err)
	}
	if b.Base != "base" || b.Observation != ObservationID(sized(int64(3))) ||
		b.Observation == ObservationID(sized(int64(4))) {
		t.Fatalf("ballot identity %+v", b)
	}
	if len(b.Votes) != 2 || b.Votes[0] != (Vote{Voter: 1, Branch: 1, Rank: 1, Instr: "A"}) ||
		b.Votes[1] != (Vote{Voter: 2, Branch: 3, Rank: 3, Instr: "A"}) {
		t.Fatalf("votes %+v", b.Votes)
	}
}

// After startle the step runs its own model's instruction, whatever the
// committed version tallies; it is learned only if the tally selects it.
func TestAfterStartleRunsItsOwn(t *testing.T) {
	chain, err := ParseGuard(`arrived.size > 2 ? "B" : null`)
	if err != nil {
		t.Fatal(err)
	}
	mine := commitOK(t, EmptyVersion(), `arrived.size > 2 ? "B" : null`)
	// A concurrent commit landed a voter covering x = 9 with another instruction.
	theirs := commitOK(t, EmptyVersion(), `arrived.size > 7 ? "A" : null`)
	raced := commitOK(t, theirs, `arrived.size > 2 ? "B" : null`)
	vals := sized(int64(9))
	for _, c := range []struct {
		v       Version
		learned bool
	}{{mine, true}, {raced, false}} {
		a, err := AfterStartle(chain, c.v.Branches, vals, Tally{})
		if err != nil || a.Startle || a.Instr != "B" || a.Learned != c.learned {
			t.Errorf("after startle %+v (%v), want B learned %v", a, err, c.learned)
		}
	}
	if a, _ := AfterStartle(chain, mine.Branches, sized(int64(1)), Tally{}); !a.Startle {
		t.Errorf("an uncovered value must startle again, got %+v", a)
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
func TestStoreConds(t *testing.T) {
	g := `$1.size > 7 && $2["content-type"] == "$3" ? "diff $1 $2" : null`
	want := []Link{{
		Cond:  `arrived.size > 7 && passed0["content-type"] == "$3"`,
		Instr: "diff @arrived @passed[0]",
	}}
	for _, scope := range []map[int]Position{
		{1: Arrived, 2: Passed(0)},
		{1: Arrived, 2: Passed(0), 3: Passed(1)},
	} {
		chain, err := StoreConds(g, scope, scope)
		if err != nil || len(chain) != 1 || chain[0] != want[0] {
			t.Fatalf("stored %+v (%v), want %+v", chain, err, want)
		}
	}
	other := map[int]Position{7: Arrived, 4: Passed(0)}
	renumbered, err := StoreConds(`$7.size > 7 && $4["content-type"] == "$3" ? "diff $7 $4" : null`, other, other)
	if err != nil || renumbered[0] != want[0] {
		t.Fatalf("another frame's numbers stored %+v (%v)", renumbered, err)
	}
	v := CommitChain(EmptyVersion(), renumbered)
	vals := map[string]any{"arrived": map[string]any{"size": 9.0},
		"passed0": map[string]any{"content-type": "$3"}}
	if s, err := Distinguish(v.Branches, vals); err != nil || s.Instr != "diff @arrived @passed[0]" {
		t.Fatalf("selects %+v (%v)", s, err)
	}
}

// A `$n` with no position in the base cannot be stored: the instruction
// runs once, unlearned. Literal `$n` in a raw or triple-quoted string is
// text.
func TestStoreCondsUnstorable(t *testing.T) {
	scope := map[int]Position{0: Arrived}
	for _, g := range []string{`$3.size > 1 ? "ls $0" : null`, `$0.size > 1 ? "diff $0 $3" : null`} {
		var u Unstorable
		if _, err := StoreConds(g, scope, scope); !errors.As(err, &u) || u.N != 3 {
			t.Errorf("%s: %v, want $3 unstorable", g, err)
		}
	}
	for _, g := range []string{
		`$0.name == r"\$3" ? "ls $0" : null`,
		`$0.name == """it's $3""" ? "ls $0" : null`,
		`$0.name == '$3\'s' ? "ls $0" : null`,
	} {
		if _, err := StoreConds(g, scope, scope); err != nil {
			t.Errorf("%s: %v", g, err)
		}
	}
}

// A stored guard reads positions only: any other variable fails.
func TestOnlyPositions(t *testing.T) {
	v := commitOK(t, EmptyVersion(), `goal.size > 1 ? "A" : null`)
	if _, err := Distinguish(v.Branches, map[string]any{"goal": map[string]any{"size": 2}}); err == nil {
		t.Fatal("a guard read a variable that is not a position")
	}
}

// Text that is not JSON arrives as a string, with no wrapper: the guard
// reads `$n` itself.
func TestTextArrival(t *testing.T) {
	arrived := map[int]Position{0: Arrived}
	chain, err := StoreConds(`$0.contains("error") ? "less $0" : size($0) > 0 ? "wc $0" : null`, arrived, arrived)
	if err != nil {
		t.Fatal(err)
	}
	v := CommitChain(EmptyVersion(), chain)
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
	arrived := map[int]Position{1: Arrived}
	chain, err := StoreConds(`$1.items.exists(i, i.id > 2) ? "ls $1" : null`, arrived, arrived)
	if err != nil || chain[0].Cond != "arrived.items.exists(i, i.id > 2)" {
		t.Fatalf("stored %+v (%v)", chain, err)
	}
	v := CommitChain(EmptyVersion(), chain)
	items := map[string]any{"arrived": map[string]any{"items": []any{map[string]any{"id": 3.0}}}}
	if s, err := Distinguish(v.Branches, items); err != nil || s.Instr != "ls @arrived" {
		t.Fatalf("selects %+v (%v)", s, err)
	}
}

// A return is a construction: its exit status is in its shape, so two
// statuses are two bases, and nothing JSON can express has its shape.
func TestReturnIsConstructed(t *testing.T) {
	done := ShapeOf(ReturnValue(0, []byte("x"), nil, false)).Enc()
	notDone := ShapeOf(ReturnValue(1, []byte("x"), nil, false)).Enc()
	if done == notDone || done != `{"exit":exit(0),"stdout":bytes}` {
		t.Fatalf("%s, %s", done, notDone)
	}
	if s := shapeOfJSON(t, `{"exit":0,"stdout":"x"}`).Enc(); s == done {
		t.Fatalf("a JSON value took a return's shape: %s", s)
	}
}

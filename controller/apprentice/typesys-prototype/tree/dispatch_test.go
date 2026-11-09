package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var stateNames = map[State]string{StateReject: "reject", StateHesitate: "hesitate",
	StateStartle: "startle", StateTally: "tally", StateTrust: "trust"}

// letters are the dispatch vectors' instructions, each granted as a tool
// of its own: the k-th letter is $(k+2), after the goal.
var letters = []string{"A", "B", "C", "D", "E", "F"}

// sizedBase is where {"size": x} arrives as a goal declared json, at a
// root granted the letters.
var sizedBase = func() string {
	bases := make([]string, len(letters))
	for i, l := range letters {
		bases[i] = standIn.Tool(l)
	}
	return standIn.Return(standIn.Input(GoalContract, bases), `{"size":number}`)
}()

// The dispatch vectors' instructions are opaque letters and their guards
// read the goal as $1. An instruction is only parameters, so the replay
// writes the k-th letter as the tool $(k+2) run bare: "C" is written "$4",
// stored as @passed[2], and read back as "C". Distinct letters are distinct
// tools, so a tally groups them apart at every level, as the Lean does.
var (
	asTool = regexp.MustCompile(`"([A-Z])"`)
	stored = regexp.MustCompile(`@passed\[([0-9]+)\]`)
)

func fromTool(s string) string {
	return stored.ReplaceAllStringFunc(s, func(ref string) string {
		k, _ := strconv.Atoi(stored.FindStringSubmatch(ref)[1])
		return letters[k]
	})
}

func toolWrites(ws []string) []string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = asTool.ReplaceAllStringFunc(w, func(q string) string { return `"$` + itoa(int(q[1]-'A')+2) + `"` })
	}
	return out
}

func untool(t *testing.T, v Version) any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal([]byte(fromTool(string(b))), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// decideOnce runs one arrival of {"size": x} against r, answering each ask
// with the next write: what Dispatch.decide gives.
func decideOnce(t *testing.T, r Version, x int64, refinements int, tl Tally, writes []string) (map[string]any, Version) {
	t.Helper()
	m := NewMachine(standIn, nil, map[string]Version{sizedBase: r},
		Bounds{Transitions: 64, Refinements: refinements, Act: len(letters), Tally: tl}, letters)
	m.stack[0].frame.Pending.Output = OutputJSON
	m, eff, err := m.Step(Answered{Bytes: []byte(`{"size":` + itoa(int(x)) + `}`)})
	if err != nil {
		t.Fatal(err)
	}
	asks := 0
	for {
		if _, ok := eff.(AskModel); !ok {
			break
		}
		asks++
		if len(writes) == 0 {
			break
		}
		if m, eff, err = m.Step(Wrote{writes[0]}); err != nil {
			t.Fatal(err)
		}
		writes = writes[1:]
	}
	states := []any{}
	for _, s := range m.States() {
		states = append(states, stateNames[s])
	}
	out := map[string]any{"states": states, "refinement": untool(t, m.version(sizedBase)), "asks": asks,
		"instr": nil, "learned": false, "ended": false}
	switch e := eff.(type) {
	case RunInvoke:
		out["instr"], out["learned"] = e.Argv[0], m.Learned()
	case Exhausted:
		out["ended"] = true
	}
	return out, m.version(sizedBase)
}

// toolVersion reads a vector's refinement, its letters run as arguments to t.
func toolVersion(t *testing.T, raw json.RawMessage) Version {
	t.Helper()
	v := dec[Version](t, raw)
	for i := range v.Branches {
		if v.Branches[i].Instr != "" {
			v.Branches[i].Instr = Passed(int(v.Branches[i].Instr[0] - 'A')).ref()
		}
	}
	return v
}

// The replay of Dispatch.lean's vectors.
func init() {
	replays["dispatch"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		x, max := dec[int64](t, in["x"]), dec[int](t, in["refinementBound"])
		tl := tallyOf(t, in)
		first, r := decideOnce(t, toolVersion(t, in["refinement"]), x, max, tl, toolWrites(dec[[]string](t, in["writes"])))
		again, _ := decideOnce(t, r, x, max, tl, nil)
		return map[string]any{"first": first, "again": again}
	}
}

// A guard that does not store or check is rejected, and the model asked
// again with the reason.
func TestMachineRejects(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte("x")})
	m.Bounds.Refinements = 6
	for _, c := range []struct{ guard, reason string }{
		{"not a guard (", "Syntax error"},
		{`$1.size > 2 ? "$2 $1" : null`, "has no keys"},
		{`true ? "$1" : null`, "$1 is not a tool"},
		{`true ? "delegate(goal($1), $2)" : null`, "you do not delegate"},
		{`size($1) > 99 ? "$2 $1" : null`, "does not select what arrived"},
	} {
		var eff any
		var err error
		if m, eff, err = m.Step(Wrote{c.guard}); err != nil {
			t.Fatal(err)
		}
		ask, ok := eff.(AskModel)
		if !ok || !strings.Contains(ask.Reason, c.reason) {
			t.Fatalf("%s: %#v, want an ask saying %q", c.guard, eff, c.reason)
		}
	}
	if err := m.States().Check(); err == nil {
		t.Fatal("a transition with no trust checks")
	}
}

func TestMachineRefusesStrayInputs(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 3, Refinements: 3, Act: 1}, nil)
	if _, _, err := m.Step(Wrote{"true"}); !errors.Is(err, ErrAsking) {
		t.Fatalf("a guard nobody asked for: %v", err)
	}
	m, _, _ = m.Step(Answered{Bytes: []byte("go")})
	if _, _, err := m.Step(Answered{Bytes: []byte("again")}); !errors.Is(err, ErrAsking) {
		t.Fatalf("bytes while asking: %v", err)
	}
}

// A model that never writes a guard that runs ends its frame after
// RefinementBound: an error observation for the caller, never a failure of the run.
func TestMachineAsksRunOut(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 2, Act: 1}, nil)
	m, eff, _ := m.Step(Answered{Bytes: []byte("x")})
	for i := 0; i < 2; i++ {
		if _, ok := eff.(AskModel); !ok {
			t.Fatalf("ask %d: %#v", i, eff)
		}
		m, eff, _ = m.Step(Wrote{"not a guard ("})
	}
	e, ok := eff.(Exhausted)
	if !ok || ShapeOf(e.Value).Enc() != `{"error":bytes}` ||
		!strings.HasPrefix(string(e.Value["error"].([]byte)), "refinement bound exhausted: the model was asked 2 times at ") {
		t.Fatalf("%#v", eff)
	}
	if m.Asking() || m.Transitions() != 0 {
		t.Fatalf("asking=%v transitions=%d", m.Asking(), m.Transitions())
	}
}

// tool is a root's tool observation, as NewMachine grants it.
func tool(b string) Obs { return Obs{Value: b, Base: standIn.Tool(b), Tool: b} }

// goalAt is where a root granted these tools receives a text goal.
func goalAt(tools ...string) string {
	var bases []string
	for _, t := range tools {
		bases = append(bases, standIn.Tool(t))
	}
	return standIn.Return(standIn.Input(GoalContract, bases), "string")
}

// drive runs a machine to its end, answering asks from model by base and
// runs from world; it returns the outcome and the asks at each depth.
func driveTest(t *testing.T, m Machine, goal string, model func(AskModel) string,
	world func(RunInvoke) Answered) (Machine, any, []int) {
	t.Helper()
	m, eff, err := m.Step(Answered{Bytes: []byte(goal)})
	var asks []int
	for i := 0; i < 50 && err == nil; i++ {
		switch e := eff.(type) {
		case AskModel:
			asks = append(asks, e.Depth)
			m, eff, err = m.Step(Wrote{model(e)})
		case RunInvoke:
			m, eff, err = m.Step(world(e))
		case Finished:
			return m, e.Value, asks
		case Exhausted:
			return m, e.Value, asks
		default:
			t.Fatalf("effect %#v", eff)
		}
	}
	t.Fatalf("no end: %v", err)
	return m, nil, nil
}

// returnedOk is a frame's return with the goal done: its stdout as text.
func returnedOk(v any) (string, bool) {
	r, ok := v.(map[string]any)
	if !ok || r["exit"] != ExitStatus(0) {
		return "", false
	}
	out, ok := r["stdout"].([]byte)
	return string(out), ok
}

// A tree learns: the first run asks the model at each new base, the second
// asks nothing, and both run the same invokes and return the same value.
func TestMachineLearns(t *testing.T) {
	lsBase := standIn.Return(standIn.Input("@passed[0] @passed[1]", []string{standIn.Tool("ls"), goalAt("ls")}), "string")
	model := func(a AskModel) string {
		switch a.Base {
		case goalAt("ls"):
			return `size($1) > 0 ? "$2 $1" : null` // run the tool $2 on the goal
		case lsBase:
			return `$3.contains("b.log") ? "return 0 $3" : null`
		}
		t.Fatalf("asked at %s", a.Base)
		return ""
	}
	world := func(r RunInvoke) Answered {
		if !slices.Equal(r.Argv, []string{"ls", "/var/log"}) {
			t.Fatalf("ran %q", r.Argv)
		}
		return Answered{Bytes: []byte("a.log b.log")}
	}
	refs := map[string]Version{}
	for run, want := range []int{2, 0} {
		m, v, asks := driveTest(t, NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls"}),
			"/var/log", model, world)
		if out, ok := returnedOk(v); !ok || out != "a.log b.log" || len(asks) != want || m.Transitions() != 2 {
			t.Fatalf("run %d: returned %v, asks %v, %d transitions", run, v, asks, m.Transitions())
		}
		refs = m.Refs
	}
}

// The first position decides: a guard whose instruction runs a $n that is
// not a tool is refused before it is committed, and the model asked again
// with the reason.
func TestMachineRunsOnlyTools(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	m, eff, err := m.Step(Wrote{`true ? "$1 $2" : null`})
	if ask, ok := eff.(AskModel); err != nil || !ok || !strings.Contains(ask.Reason, "$1 is not a tool") {
		t.Fatalf("%#v %v", eff, err)
	}
	if !slices.Equal(m.States(), Transition{StateStartle, StateReject, StateStartle}) {
		t.Fatalf("states %v", m.States())
	}
}

// A guard reading a $n outside the base runs once, unlearned: nothing is
// stored, and the next arrival asks again.
func TestMachineUnlearned(t *testing.T) {
	dateBase := standIn.Return(standIn.Input("@passed[0]", []string{standIn.Tool("date")}), "string")
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 2}, []string{"date", "cat"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	m, _, _ = m.Step(Wrote{`true ? "$2" : null`}) // date, passed only itself
	m, eff, _ := m.Step(Answered{Bytes: []byte("today")})
	if ask, ok := eff.(AskModel); !ok || ask.Base != dateBase {
		t.Fatalf("%#v", eff)
	}
	m, eff, err := m.Step(Wrote{`true ? "$3 $1 $4" : null`}) // cat and the goal are not in date's base
	if err != nil {
		t.Fatal(err)
	}
	run, ok := eff.(RunInvoke)
	if !ok || !slices.Equal(run.Argv, []string{"cat", "/var/log", "today"}) || m.Learned() {
		t.Fatalf("%#v learned=%v", eff, m.Learned())
	}
	if !slices.Equal(m.States(), Transition{StateStartle, StateTrust}) {
		t.Fatalf("states %v", m.States())
	}
	if v := m.version(dateBase); len(v.Branches) != 0 {
		t.Fatalf("stored %+v", v)
	}
}

// A learned cycle that never asks still ends: the transition bound returns
// the error observation.
func TestMachineExhausts(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 3, Refinements: 3, Act: 1}, []string{"echo"})
	_, v, _ := driveTest(t, m, "go", func(a AskModel) string {
		return `true ? "$2 $1" : null`
	}, func(RunInvoke) Answered { return Answered{Bytes: []byte("go")} })
	e, ok := v.(map[string]any)
	if !ok || string(e["error"].([]byte)) != "transition bound exhausted after 3 transitions" {
		t.Fatalf("%#v", v)
	}
}

// A learned cut takes from new values: the guard says where the argument
// comes from, never what it was, so a second goal with another directory
// runs with no ask.
func TestMachineCutLearns(t *testing.T) {
	model := func(a AskModel) string {
		if len(a.Scope) == 2 { // the goal, ls
			return `has($1.dir) ? "$2 $1.dir" : null`
		}
		return `size($3) > 0 ? "return 0 $3[1:1]" : null` // the first line ls gave
	}
	refs := map[string]Version{}
	for run, c := range []struct{ goal, dir string }{{`{"dir":"/a"}`, "/a"}, {`{"dir":"/b"}`, "/b"}} {
		m := NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls"})
		m.stack[0].frame.Pending.Output = OutputJSON
		m, v, asks := driveTest(t, m, c.goal, model, func(r RunInvoke) Answered {
			if !slices.Equal(r.Argv, []string{"ls", c.dir}) {
				t.Fatalf("run %d ran %q", run, r.Argv)
			}
			return Answered{Bytes: []byte(c.dir + "/x.log\n" + c.dir + "/y.log\n")}
		})
		if out, ok := returnedOk(v); !ok || out != c.dir+"/x.log" || len(asks) != 2*(1-run) {
			t.Fatalf("run %d: returned %#v after %d asks", run, v, len(asks))
		}
		refs = m.Refs
	}
}

// A cut that does not apply is refused before it is committed, with the
// reason.
func TestMachineCutRefused(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte("one line")})
	_, eff, _ := m.Step(Wrote{`true ? "$2 $1[2:2]" : null`})
	if ask, ok := eff.(AskModel); !ok || !strings.Contains(ask.Reason, "$1[2:2] does not apply") {
		t.Fatalf("%#v", eff)
	}
}

func TestParseInstrCuts(t *testing.T) {
	in, err := ParseInstr("$1 -l $2.dir.path --x=$3[1:2,4] $4")
	if err != nil {
		t.Fatal(err)
	}
	want := []Arg{{2, Sel{Kind: SelField, Path: []string{"dir", "path"}}}, {3, Sel{Kind: SelLines, Runs: [][2]int{{1, 2}, {4, 4}}}}, {4, Sel{}}}
	if in.Kind != InstrRun || in.N != 1 || fmt.Sprint(in.Args) != fmt.Sprint(want) {
		t.Fatalf("%+v", in)
	}
	for _, bad := range []string{"$1[1:2] $2", "return 0 $2[x]", "task $2 lit", "return 0 $2.", "return $2", "return 2 $2", "$1 $2[1",
		"delegate($1, $2)", "delegate(goal($1) $2)", "delegate(goal(x), $2)", "delegate(goal($1), lit)", "'spec' $1 $2"} {
		if _, err := ParseInstr(bad); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
	in, err = ParseInstr("delegate(goal($1[1:2,4]), $2, $3.dir)")
	want = []Arg{{2, Sel{}}, {3, Sel{Kind: SelField, Path: []string{"dir"}}}}
	if err != nil || in.Kind != InstrDelegate || fmt.Sprint(in.Goal) != fmt.Sprint(Arg{1, Sel{Kind: SelLines, Runs: [][2]int{{1, 2}, {4, 4}}}}) ||
		fmt.Sprint(in.Args) != fmt.Sprint(want) {
		t.Fatalf("%+v %v", in, err)
	}
}

// Guards may call the strings extension: the functions models reach for
// when reading text check and evaluate.
func TestGuardStrings(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte(" a\nb ")})
	_, eff, err := m.Step(Wrote{`size($1.split("\n")) == 2 && $1.trim().startsWith("a") ? "$2 $1[2:2]" : null`})
	if run, ok := eff.(RunInvoke); err != nil || !ok || !slices.Equal(run.Argv, []string{"ls", "b "}) {
		t.Fatalf("%#v %v", eff, err)
	}
}

// A delegate delegates: the child, an act frame, is passed its goal as $1
// and runs at depth 1; what it returns arrives at its parent. The second
// run asks nothing at either depth.
func TestMachineChild(t *testing.T) {
	model := func(a AskModel) string {
		switch {
		case a.Depth == 0 && len(a.Scope) == 3: // the goal, ls, wc
			if a.Kind != KindDelegates || !a.Scope[0].Goal {
				t.Fatalf("the root holds two tools under k=1: kind %v, %+v", a.Kind, a.Scope)
			}
			return `size($1) > 0 ? "delegate(goal($1), $2)" : null`
		case a.Depth == 1 && len(a.Scope) == 2: // the goal, ls
			if !a.Scope[0].Goal || a.Scope[0].Value != "/var/log" || a.Scope[1].Tool != "ls" || a.Kind != KindActs {
				t.Fatalf("child %v %+v", a.Kind, a.Scope)
			}
			return `true ? "$2 $1" : null`
		case a.Depth == 1 && len(a.Scope) == 3:
			return `true ? "return 0 $3" : null`
		case a.Depth == 0 && len(a.Scope) == 4: // what the child returned
			if a.Scope[3].Role() != RoleContext {
				t.Fatalf("a return arrives as %v", a.Scope[3].Role())
			}
			if s := ShapeOf(a.Scope[3].Value).Enc(); s != `{"exit":exit(0),"stdout":bytes}` {
				t.Fatalf("a return arrives as %s", s)
			}
			return `$4.exit == 0 ? "return 0 $4.stdout" : null`
		}
		t.Fatalf("asked at depth %d with %d in scope", a.Depth, len(a.Scope))
		return ""
	}
	world := func(r RunInvoke) Answered {
		if !slices.Equal(r.Argv, []string{"ls", "/var/log"}) {
			t.Fatalf("ran %q", r.Argv)
		}
		return Answered{Bytes: []byte("a.log b.log")}
	}
	refs := map[string]Version{}
	for run, want := range [][]int{{0, 1, 1, 0}, nil} {
		m, v, asks := driveTest(t, NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls", "wc"}),
			"/var/log", model, world)
		if out, ok := returnedOk(v); !ok || out != "a.log b.log" || !slices.Equal(asks, want) || m.Depth() != 0 {
			t.Fatalf("run %d: returned %v, asks at depths %v, depth %d", run, v, asks, m.Depth())
		}
		refs = m.Refs
	}
}

// A delegate does not run its tools, and an act frame does not delegate:
// each is refused with the reason, and the model asked again.
func TestMachineKinds(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls", "wc"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	_, eff, _ := m.Step(Wrote{`true ? "$2 $1" : null`})
	if !asks(eff, DelegateRuns.Error()) {
		t.Fatalf("%#v", eff)
	}
	// k is a variable: the same root under k=2 acts, and runs.
	m = NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 2}, []string{"ls", "wc"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	if _, eff, _ := m.Step(Wrote{`true ? "$2 $1" : null`}); !isRun(eff) {
		t.Fatalf("under k=2: %#v", eff)
	}
	if _, eff, _ := m.Step(Wrote{`true ? "delegate(goal($1), $2)" : null`}); !asks(eff, ActDelegates.Error()) {
		t.Fatalf("under k=2 a delegation: %#v", eff)
	}
}

func isRun(eff any) bool { _, ok := eff.(RunInvoke); return ok }

// A delegation must narrow, pass a tool, and take its goal from an
// observation that is not a tool.
func TestMachineDelegationRules(t *testing.T) {
	for _, c := range []struct{ guard, reason string }{
		{`true ? "delegate(goal($1), $2, $3)" : null`, "passes 2 of your 2"},
		{`true ? "delegate(goal($1), $1)" : null`, NoTools.Error()},
		{`true ? "delegate(goal($2), $2)" : null`, "$2 is a tool"},
		{`true ? "delegate(goal($1[3:3]), $2)" : null`, "$1[3:3] does not apply"},
		{`true ? "delegate($1, $2)" : null`, "the goal"},
		{`true ? "'list $1' $2 $1" : null`, "not an instruction"},
	} {
		m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls", "wc"})
		m, _, _ = m.Step(Answered{Bytes: []byte("a\nb")})
		if _, eff, _ := m.Step(Wrote{c.guard}); !asks(eff, c.reason) {
			t.Errorf("%s: %#v, want %q", c.guard, eff, c.reason)
		}
	}
}

func asks(eff any, reason string) bool {
	ask, ok := eff.(AskModel)
	return ok && strings.Contains(ask.Reason, reason)
}

// A child whose model never writes a guard that runs ends with an error
// observation, which arrives at its parent like any return: the parent's
// model reacts, and the run is not a failure.
func TestMachineChildExhausted(t *testing.T) {
	var parentSaw any
	model := func(a AskModel) string {
		if a.Depth == 1 {
			return "not a guard ("
		}
		if len(a.Scope) == 3 { // the goal, ls, wc
			return `true ? "delegate(goal($1), $2)" : null`
		}
		parentSaw = a.Arrival.Vals["arrived"]
		return `true ? "return 0 $4" : null`
	}
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 2, Act: 1}, []string{"ls", "wc"})
	_, v, asks := driveTest(t, m, "go", model, nil)
	e, ok := parentSaw.(map[string]any)
	if !ok || !strings.HasPrefix(string(e["error"].([]byte)), "refinement bound exhausted") {
		t.Fatalf("the parent saw %#v", parentSaw)
	}
	if !slices.Equal(asks, []int{0, 1, 1, 0}) || v == nil {
		t.Fatalf("asks %v, returned %#v", asks, v)
	}
}

// A child's goal and context are cuts: its scope holds the cut values, in
// roles' order however they were written (the goal, the tools, the
// context), and nothing passed is a goal but its own.
func TestMachineChildGetsCut(t *testing.T) {
	model := func(a AskModel) string {
		if a.Depth == 1 {
			s := a.Scope
			if len(s) != 3 || s[0].Value != "b" || !s[0].Goal || s[1].Tool != "ls" || s[2].Value != "c" || s[2].Role() != RoleContext {
				t.Fatalf("child scope %+v", s)
			}
			return `true ? "return 0 $3" : null`
		}
		if len(a.Scope) == 3 {
			return `true ? "delegate(goal($1[2:2]), $1[3:3], $2)" : null`
		}
		return `true ? "return 0 $4.stdout" : null`
	}
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3, Act: 1}, []string{"ls", "cat"})
	if _, v, _ := driveTest(t, m, "a\nb\nc", model, nil); v == nil {
		t.Fatal("no return")
	} else if out, ok := returnedOk(v); !ok || out != "c" {
		t.Fatalf("returned %#v", v)
	}
}

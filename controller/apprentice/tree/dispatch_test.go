package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var stateNames = map[State]string{StateReject: "reject", StateHesitate: "hesitate",
	StateStartle: "startle", StateReconcile: "reconcile", StateTrust: "trust"}

// sizedBase is where {"size": x} arrives as a goal declared json, at a
// root granted the tool t.
var sizedBase = standIn.Return(standIn.Input(GoalContract, []string{standIn.Tool("t")}), `{"size":number}`)

// The dispatch vectors' instructions are opaque letters and their guards
// read the goal as $1. The replay grants the tool t, so the goal is $2, and
// runs each letter as an argument to it: "D" is written "$1 D", stored
// "@passed[0] D", and read back as "D".
var (
	asTool   = regexp.MustCompile(`"([A-Z])"`)
	fromTool = strings.NewReplacer(Passed(0).ref()+" ", "")
)

func toolWrites(ws []string) []string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = asTool.ReplaceAllString(strings.ReplaceAll(w, "$1", "$2"), `"$$1 $1"`)
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
	if err := json.Unmarshal([]byte(fromTool.Replace(string(b))), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// decideOnce runs one arrival of {"size": x} against r, answering each ask
// with the next write: what Dispatch.decide gives.
func decideOnce(t *testing.T, r Version, x int64, refinements int, writes []string) (map[string]any, Version) {
	t.Helper()
	m := NewMachine(standIn, nil, map[string]Version{sizedBase: r}, Bounds{Transitions: 64, Refinements: refinements}, []string{"t"})
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
		out["instr"], out["learned"] = fromTool.Replace(e.Contract), m.Learned()
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
			v.Branches[i].Instr = Passed(0).ref() + " " + v.Branches[i].Instr
		}
	}
	return v
}

// The replay of Dispatch.lean's vectors.
func init() {
	replays["dispatch"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		x, max := dec[int64](t, in["x"]), dec[int](t, in["refinementBound"])
		first, r := decideOnce(t, toolVersion(t, in["refinement"]), x, max, toolWrites(dec[[]string](t, in["writes"])))
		again, _ := decideOnce(t, r, x, max, nil)
		return map[string]any{"first": first, "again": again}
	}
}

// A guard that does not store or check is rejected, and the model asked
// again with the reason.
func TestMachineRejects(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte("x")})
	m.Bounds.Refinements = 6
	for _, c := range []struct{ guard, reason string }{
		{"not a guard (", "Syntax error"},
		{`$2.size > 2 ? "$1 $2" : null`, "has no keys"},
		{`true ? "$2" : null`, "$2 is not a tool"},
		{`true ? "list $1 $2" : null`, "passes 1 of the frame's 1"},
		{`size($2) > 99 ? "$1 $2" : null`, "does not select what arrived"},
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
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 3, Refinements: 3}, nil)
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
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 2}, nil)
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

// A tree learns: the first run asks the model at each new base, the second
// asks nothing, and both run the same invokes and return the same value.
func TestMachineLearns(t *testing.T) {
	lsBase := standIn.Return(standIn.Input("@passed[0] @passed[1]", []string{standIn.Tool("ls"), goalAt("ls")}), "string")
	model := func(a AskModel) string {
		switch a.Base {
		case goalAt("ls"):
			return `size($2) > 0 ? "$1 $2" : null` // run the tool $1 on the goal
		case lsBase:
			return `$3.contains("b.log") ? "return $3" : null`
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
		m, v, asks := driveTest(t, NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3}, []string{"ls"}),
			"/var/log", model, world)
		if v != "a.log b.log" || len(asks) != want || m.Transitions() != 2 {
			t.Fatalf("run %d: returned %v, asks %v, %d transitions", run, v, asks, m.Transitions())
		}
		refs = m.Refs
	}
}

// The first position decides: a guard whose instruction runs a $n that is
// not a tool is refused before it is committed, and the model asked again
// with the reason.
func TestMachineRunsOnlyTools(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	m, eff, err := m.Step(Wrote{`true ? "$2 $1" : null`})
	if ask, ok := eff.(AskModel); err != nil || !ok || !strings.Contains(ask.Reason, "$2 is not a tool") {
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
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3}, []string{"date", "cat"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	m, _, _ = m.Step(Wrote{`true ? "$1" : null`}) // date, passed only itself
	m, eff, _ := m.Step(Answered{Bytes: []byte("today")})
	if ask, ok := eff.(AskModel); !ok || ask.Base != dateBase {
		t.Fatalf("%#v", eff)
	}
	m, eff, err := m.Step(Wrote{`true ? "$2 $3 $4" : null`}) // cat and the goal are not in date's base
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
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 3, Refinements: 3}, []string{"echo"})
	_, v, _ := driveTest(t, m, "go", func(a AskModel) string {
		return `true ? "$1 $` + itoa(len(a.Scope)) + `" : null`
	}, func(RunInvoke) Answered { return Answered{Bytes: []byte("go")} })
	e, ok := v.(map[string]any)
	if !ok || string(e["error"].([]byte)) != "transition bound exhausted after 3 transitions" {
		t.Fatalf("%#v", v)
	}
}

// A delegation passing fewer tools than its emitter holds starts a child
// frame: it runs at depth 1 with what it was passed as its scope, and what
// it returns arrives at its parent. The second run asks nothing at either
// depth.
func TestMachineChild(t *testing.T) {
	model := func(a AskModel) string {
		switch {
		case a.Depth == 0 && len(a.Scope) == 3: // ls, wc, the goal
			return `size($3) > 0 ? "list $1 $3" : null` // a child holding only ls
		case a.Depth == 1 && len(a.Scope) == 2: // the child: ls, the goal
			if a.Scope[0].Tool != "ls" || a.Scope[1].Value != "/var/log" {
				t.Fatalf("child scope %+v", a.Scope)
			}
			return `true ? "$1 $2" : null`
		case a.Depth == 1 && len(a.Scope) == 3:
			return `true ? "return $3" : null`
		case a.Depth == 0 && len(a.Scope) == 4: // what the child returned
			return `$4.contains("a.log") ? "return $4" : null`
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
		m, v, asks := driveTest(t, NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3}, []string{"ls", "wc"}),
			"/var/log", model, world)
		if v != "a.log b.log" || !slices.Equal(asks, want) || m.Depth() != 0 {
			t.Fatalf("run %d: returned %v, asks at depths %v, depth %d", run, v, asks, m.Depth())
		}
		refs = m.Refs
	}
}

// A delegation that does not narrow is rejected: passing every tool the
// emitter holds is not a child.
func TestMachineChildMustNarrow(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3}, []string{"ls", "wc"})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	_, eff, _ := m.Step(Wrote{`true ? "list $1 $2 $3" : null`})
	if ask, ok := eff.(AskModel); !ok || ask.Depth != 0 || !strings.Contains(ask.Reason, "passes 2 of the frame's 2") {
		t.Fatalf("%#v", eff)
	}
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
		if len(a.Scope) == 3 { // ls, wc, the goal
			return `true ? "try $1 $3" : null`
		}
		parentSaw = a.Arrival.Vals["arrived"]
		return `true ? "return $4" : null`
	}
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 2}, []string{"ls", "wc"})
	_, v, asks := driveTest(t, m, "go", model, nil)
	e, ok := parentSaw.(map[string]any)
	if !ok || !strings.HasPrefix(string(e["error"].([]byte)), "refinement bound exhausted") {
		t.Fatalf("the parent saw %#v", parentSaw)
	}
	if !slices.Equal(asks, []int{0, 1, 1, 0}) || v == nil {
		t.Fatalf("asks %v, returned %#v", asks, v)
	}
}

// A learned cut takes from new values: the guard says where the argument
// comes from, never what it was, so a second goal with another directory
// runs with no ask.
func TestMachineCutLearns(t *testing.T) {
	model := func(a AskModel) string {
		if len(a.Scope) == 2 { // ls, the goal
			return `has($2.dir) ? "$1 $2.dir" : null`
		}
		return `size($3) > 0 ? "return $3[1:1]" : null` // the first line ls gave
	}
	refs := map[string]Version{}
	for run, c := range []struct{ goal, dir string }{{`{"dir":"/a"}`, "/a"}, {`{"dir":"/b"}`, "/b"}} {
		m := NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3}, []string{"ls"})
		m.stack[0].frame.Pending.Output = OutputJSON
		m, v, asks := driveTest(t, m, c.goal, model, func(r RunInvoke) Answered {
			if !slices.Equal(r.Argv, []string{"ls", c.dir}) {
				t.Fatalf("run %d ran %q", run, r.Argv)
			}
			return Answered{Bytes: []byte(c.dir + "/x.log\n" + c.dir + "/y.log\n")}
		})
		if v != c.dir+"/x.log" || len(asks) != 2*(1-run) {
			t.Fatalf("run %d: returned %#v after %d asks", run, v, len(asks))
		}
		refs = m.Refs
	}
}

// A child is passed a cut: its scope holds the cut value, not the whole
// observation, and the cut is not a tool.
func TestMachineChildGetsCut(t *testing.T) {
	model := func(a AskModel) string {
		if a.Depth == 1 {
			if len(a.Scope) != 1 || a.Scope[0].Value != "b" || a.Scope[0].Tool != "" {
				t.Fatalf("child scope %+v", a.Scope)
			}
			return `true ? "return $1" : null`
		}
		if len(a.Scope) == 2 {
			return `true ? "second $2[2:2]" : null`
		}
		return `true ? "return $3" : null`
	}
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3}, []string{"ls"})
	if _, v, _ := driveTest(t, m, "a\nb\nc", model, nil); v != "b" {
		t.Fatalf("returned %#v", v)
	}
}

// A cut that does not apply is refused before it is committed, with the
// reason.
func TestMachineCutRefused(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3}, []string{"ls"})
	m, _, _ = m.Step(Answered{Bytes: []byte("one line")})
	_, eff, _ := m.Step(Wrote{`true ? "$1 $2[2:2]" : null`})
	if ask, ok := eff.(AskModel); !ok || !strings.Contains(ask.Reason, "$2[2:2] does not apply") {
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
	for _, bad := range []string{"$1[1:2] $2", "return $2[x]", "task $2 lit", "return $2.", "$1 $2[1"} {
		if _, err := ParseInstr(bad); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
}

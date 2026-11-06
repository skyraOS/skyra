package tree

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

var stateNames = map[State]string{StateReject: "reject", StateHesitate: "hesitate",
	StateStartle: "startle", StateReconcile: "reconcile", StateTrust: "trust"}

// goalBase is where a text goal arrives; lsBase where ls, passed it,
// returns text.
var (
	goalBase = standIn.Return(standIn.Input(GoalContract, nil), "string")
	lsBase   = standIn.Return(standIn.Input("ls @passed[0]", []string{goalBase}), "string")
)

// sizedBase is where {"size": x} arrives as a goal declared json.
var sizedBase = standIn.Return(standIn.Input(GoalContract, nil), `{"size":number}`)

// decideOnce runs one arrival of {"size": x} against r, answering each ask
// with the next write: what Dispatch.decide gives.
func decideOnce(t *testing.T, r Version, x int64, refinements int, writes []string) (map[string]any, Version) {
	t.Helper()
	m := NewMachine(standIn, nil, map[string]Version{sizedBase: r}, Bounds{Transitions: 64, Refinements: refinements})
	m.Frame.Pending.Output = OutputJSON
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
	out := map[string]any{"states": states, "refinement": m.version(sizedBase), "asks": asks,
		"instr": nil, "learned": false, "ended": false}
	switch e := eff.(type) {
	case RunInvoke:
		out["instr"], out["learned"] = e.Contract, m.Learned()
	case Exhausted:
		out["ended"] = true
	}
	return out, m.version(sizedBase)
}

// The replay of Dispatch.lean's vectors.
func init() {
	replays["dispatch"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		x, max := dec[int64](t, in["x"]), dec[int](t, in["refinementBound"])
		first, r := decideOnce(t, dec[Version](t, in["refinement"]), x, max, dec[[]string](t, in["writes"]))
		again, _ := decideOnce(t, r, x, max, nil)
		return map[string]any{"first": first, "again": again}
	}
}

// A tree learns: the first run asks the model at each new base, the second
// asks nothing, and both run the same invokes and return the same value.
func TestMachineLearns(t *testing.T) {
	model := map[string]string{
		// What the goal's return base sees: the goal is $1.
		goalBase: `size($1) > 0 ? "ls $1" : null`,
		// What ls's return base sees: the listing is $2.
		lsBase: `$2.contains("b.log") ? "return $2" : true ? "rm $1" : null`,
	}
	world := func(r RunInvoke) []byte {
		if !slices.Equal(r.Argv, []string{"ls", "/var/log"}) {
			t.Fatalf("ran %q", r.Argv)
		}
		return []byte("a.log b.log")
	}
	refs := map[string]Version{}
	for run, wantAsks := range []int{2, 0} {
		m := NewMachine(standIn, nil, refs, Bounds{Transitions: 8, Refinements: 3})
		m, eff, err := m.Step(Answered{Bytes: []byte("/var/log")})
		asks := 0
		for err == nil {
			switch e := eff.(type) {
			case AskModel:
				asks++
				guard, ok := model[e.Base]
				if !ok {
					t.Fatalf("asked at %s", e.Base)
				}
				m, eff, err = m.Step(Wrote{guard})
				continue
			case RunInvoke:
				m, eff, err = m.Step(Answered{Bytes: world(e)})
				continue
			case Finished:
				if e.Value != "a.log b.log" {
					t.Fatalf("returned %v", e.Value)
				}
			default:
				t.Fatalf("effect %#v", eff)
			}
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if asks != wantAsks || m.Transitions() != 2 {
			t.Fatalf("run %d: %d asks, %d transitions", run, asks, m.Transitions())
		}
		refs = m.Refs
	}
}

// A guard that does not store or check is rejected, and the model asked
// again with the reason.
func TestMachineRejects(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3})
	m, _, _ = m.Step(Answered{Bytes: []byte("x")})
	m.RefinementBound = 4
	for _, c := range []struct{ guard, reason string }{
		{"not a guard (", "Syntax error"},
		{`$1.size > 2 ? "ls" : null`, "has no keys"},
		{`size($1) > 99 ? "ls" : null`, "does not select what arrived"},
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

// A guard reading a $n outside the base runs once, unlearned: nothing is
// stored, and the next arrival asks again.
func TestMachineUnlearned(t *testing.T) {
	dateBase := standIn.Return(standIn.Input("date", []string{}), "string")
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 3})
	m, _, _ = m.Step(Answered{Bytes: []byte("/var/log")})
	m, _, _ = m.Step(Wrote{`true ? "date" : null`}) // passes nothing: $1 is not in date's base
	m, eff, _ := m.Step(Answered{Bytes: []byte("today")})
	if ask, ok := eff.(AskModel); !ok || ask.Base != dateBase {
		t.Fatalf("%#v", eff)
	}
	m, eff, err := m.Step(Wrote{`true ? "cat $1 $2" : null`})
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

// A learned cycle that never asks still ends: the transition bound returns the
// error observation.
func TestMachineExhausts(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 3, Refinements: 3})
	m, _, _ = m.Step(Answered{Bytes: []byte("go")})
	m, eff, _ := m.Step(Wrote{`true ? "echo $1" : null`})
	asked := 1
	for i := 0; i < 10; i++ {
		switch e := eff.(type) {
		case RunInvoke:
			m, eff, _ = m.Step(Answered{Bytes: []byte("go")})
			continue
		case AskModel:
			asked++
			m, eff, _ = m.Step(Wrote{`true ? "echo $` + itoa(len(m.Frame.Scope)) + `" : null`})
			continue
		case Exhausted:
			if string(e.Value["error"].([]byte)) != "transition bound exhausted after 3 transitions" {
				t.Fatalf("%v", e.Value)
			}
			return
		}
		t.Fatalf("effect %#v", eff)
	}
	t.Fatal("never exhausted")
}

func TestMachineRefusesStrayInputs(t *testing.T) {
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 3, Refinements: 3})
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
	m := NewMachine(standIn, nil, nil, Bounds{Transitions: 8, Refinements: 2})
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

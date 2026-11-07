package tree

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// The hash stand-ins Walk.lean's vectors use.
var standIn = WalkHashes{
	Input:  func(c string, bs []string) string { return "in(" + c + ";" + strings.Join(bs, ",") + ")" },
	Return: func(i, s string) string { return "ret(" + i + ";" + s + ")" },
	Tool:   func(b string) string { return "tool(" + b + ")" },
}

// The replay of Walk.lean's vectors: each case's events, applied in turn.
func init() {
	replays["walk"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		var start struct{ Scope json.RawMessage }
		if err := json.Unmarshal(in["frame"], &start); err != nil {
			t.Fatal(err)
		}
		f := Frame{Scope: scopeOf(t, start.Scope)}
		var events []struct {
			Invoke *struct {
				Contract, Output string
				Args             []int
			}
			Return, Stderr []int
			Exit           int
			Returned       json.RawMessage
		}
		if err := json.Unmarshal(in["events"], &events); err != nil {
			t.Fatal(err)
		}
		out := []any{}
		for _, e := range events {
			var a Arrival
			var err error
			if e.Invoke != nil {
				a, err = SenseInvoke(standIn, f, e.Invoke.Contract, outputOf(e.Invoke.Output), e.Invoke.Args)
			} else if e.Returned != nil {
				v, derr := DecodeJSON(e.Returned)
				if derr != nil {
					t.Fatal(derr)
				}
				a, err = SenseReturned(standIn, f, v)
			} else {
				a, err = SenseReturn(standIn, f, bytesOf(e.Return), e.Exit, bytesOf(e.Stderr))
			}
			if err != nil {
				out = append(out, wFault(t, err))
				continue
			}
			out = append(out, wArrival(a))
			f = a.Next
		}
		return out
	}
}

// scopeOf reads a scope as the vectors write it.
func scopeOf(t *testing.T, raw json.RawMessage) []Obs {
	t.Helper()
	var os []struct {
		Value json.RawMessage
		Base  string
		Tool  *string
	}
	if err := json.Unmarshal(raw, &os); err != nil {
		t.Fatal(err)
	}
	out := []Obs{}
	for _, o := range os {
		v, err := DecodeJSON(o.Value)
		if err != nil {
			t.Fatal(err)
		}
		ob := Obs{Value: v, Base: o.Base}
		if o.Tool != nil {
			ob.Tool = *o.Tool
		}
		out = append(out, ob)
	}
	return out
}

func bytesOf(ns []int) []byte {
	out := []byte{}
	for _, n := range ns {
		out = append(out, byte(n))
	}
	return out
}

func outputOf(s string) Output {
	if s == "json" {
		return OutputJSON
	}
	return OutputText
}

func wFault(t *testing.T, err error) any {
	var u Unbound
	switch {
	case errors.Is(err, ErrBusy):
		return map[string]any{"fault": "busy"}
	case errors.Is(err, ErrIdle):
		return map[string]any{"fault": "idle"}
	case errors.As(err, &u):
		return map[string]any{"fault": "unbound", "n": u.N}
	}
	t.Fatalf("unexpected fault %v", err)
	return nil
}

func wObs(os []Obs) []any {
	out := []any{}
	for _, o := range os {
		var tool any
		if o.Tool != "" {
			tool = o.Tool
		}
		out = append(out, map[string]any{"value": numbered(o.Value), "base": o.Base, "tool": tool})
	}
	return out
}

func wFrame(f Frame) any {
	m := map[string]any{"scope": wObs(f.Scope), "pending": nil}
	if p := f.Pending; p != nil {
		out := "text"
		if p.Output == OutputJSON {
			out = "json"
		}
		m["pending"] = map[string]any{"input": p.Input, "output": out, "args": p.Args, "passed": wObs(p.Passed)}
	}
	return m
}

// template is `run $1 $2 … $n`, every number of the scope, and $9 past it.
func wTemplate(n int, past bool) string {
	s := "run"
	for i := 1; i <= n; i++ {
		s += " $" + itoa(i)
	}
	if past {
		s += " $9"
	}
	return s
}

func wLearn(a Arrival, tmpl string) any {
	stored, err := StoreInstr(tmpl, a.Table)
	var u Unstorable
	if errors.As(err, &u) {
		return map[string]any{"unstorable": u.N}
	}
	bound, err := Bind(stored, a.Refs)
	if err != nil {
		return map[string]any{"stored": stored, "bound": nil}
	}
	return map[string]any{"stored": stored, "bound": bound}
}

func wArrival(a Arrival) any {
	dir := "forward"
	if a.Dir == Back {
		dir = "back"
	}
	env := map[string]any{}
	for k, v := range a.Vals {
		env[k] = numbered(v)
	}
	table := map[string]any{}
	for n, p := range a.Table {
		table[itoa(n)] = p.cel()
	}
	n := len(a.Next.Scope)
	return map[string]any{"dir": dir, "base": a.Base, "env": env, "table": table, "next": wFrame(a.Next),
		"template": wTemplate(n, false), "learn": wLearn(a, wTemplate(n, false)),
		"templatePast": wTemplate(n, true), "learnPast": wLearn(a, wTemplate(n, true))}
}

func wGoal() Frame {
	return Frame{Scope: []Obs{{Value: "list the logs", Base: "goal"}, {Value: map[string]any{"size": 5.0}, Base: "b"}}}
}

// One at a time: an invoke while one is pending is refused, a return with
// nothing pending answers nothing, and a fault changes nothing.
func TestOneAtATime(t *testing.T) {
	f := wGoal()
	if _, err := SenseReturn(standIn, f, []byte("x"), 0, nil); !errors.Is(err, ErrIdle) {
		t.Fatalf("return with nothing pending: %v", err)
	}
	a, err := SenseInvoke(standIn, f, "ls", OutputText, []int{2})
	if err != nil {
		t.Fatal(err)
	}
	if f.Pending != nil || len(a.Next.Scope) != 2 {
		t.Fatalf("invoke changed the frame it was given or the scope")
	}
	if _, err := SenseInvoke(standIn, a.Next, "rm", OutputText, []int{1}); !errors.Is(err, ErrBusy) {
		t.Fatalf("second invoke: %v", err)
	}
	r, err := SenseReturn(standIn, a.Next, []byte("a.log"), 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Dir != Back || r.Next.Pending != nil || r.Base != standIn.Return(a.Base, "string") {
		t.Fatalf("return stands at %+v", r)
	}
	if _, err := SenseInvoke(standIn, r.Next, "rm", OutputText, []int{3}); err != nil {
		t.Fatalf("invoke after the return: %v", err)
	}
}

// The direction decides the base, and nothing else: a guard learned at a
// return base selects with the same Distinguish as at an input base, and
// what it stored binds back in another frame to that frame's own $n.
func TestLearnedReturnBindsElsewhere(t *testing.T) {
	walk := func(f Frame, args []int, out string) Arrival {
		a, err := SenseInvoke(standIn, f, "wc", OutputJSON, args)
		if err != nil {
			t.Fatal(err)
		}
		r, err := SenseReturn(standIn, a.Next, []byte(out), 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	r1 := walk(wGoal(), []int{2}, `{"n":3}`)
	chain, err := StoreGuard(`$3.n > 2 ? "head -n $3 $2" : null`, r1.Table)
	if err != nil {
		t.Fatal(err)
	}
	v, _, _, err := CommitChain(EmptyVersion(), chain)
	if err != nil {
		t.Fatal(err)
	}

	// Another frame, the same location: its observations are elsewhere in
	// its scope, but from the same bases.
	other := Frame{Scope: []Obs{{Value: map[string]any{"size": 9.0}, Base: "b"}, {Value: "x", Base: "goal"},
		{Value: "y", Base: "goal"}}}
	r2 := walk(other, []int{1}, `{"n":7}`)
	if r2.Base != r1.Base {
		t.Fatalf("same bases passed, different locations: %s, %s", r1.Base, r2.Base)
	}
	sel, err := Distinguish(v.Branches, r2.Vals)
	if err != nil || sel.Kind != SelectOne {
		t.Fatalf("selects %+v (%v)", sel, err)
	}
	got, err := Bind(sel.Instr, r2.Refs)
	if err != nil || got != "head -n $4 $1" {
		t.Fatalf("binds %q (%v), want the other frame's own $4 and $1", got, err)
	}
}

func TestBindRefusesUnknownPosition(t *testing.T) {
	if _, err := Bind("cat @passed[3]", map[Position]int{Arrived: 2}); err == nil {
		t.Fatal("bound a position the base lacks")
	}
	if got, err := Bind("echo @passed[x] @arrived", map[Position]int{Arrived: 2}); err != nil || got != "echo @passed[x] $2" {
		t.Fatalf("got %q (%v)", got, err)
	}
}

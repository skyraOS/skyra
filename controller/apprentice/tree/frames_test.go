package tree

import (
	"encoding/json"
	"testing"
)

// The replay of Frames.lean's vectors: each instruction classified against
// its scope, a child's first arrival built from what it is passed.
func init() {
	replays["frames"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		scope := scopeOf(t, in["scope"])
		type rawArg struct {
			N   int
			Sel string
		}
		var raw struct {
			Return   *rawArg
			Run      *int
			Delegate *string
			Args     []rawArg
		}
		if err := json.Unmarshal(in["instr"], &raw); err != nil {
			t.Fatal(err)
		}
		argOf := func(a rawArg) Arg {
			s, err := ParseSel(a.Sel)
			if err != nil {
				t.Fatalf("%q: %v", a.Sel, err)
			}
			if s.String() != a.Sel {
				t.Fatalf("selector %q reads back as %q", a.Sel, s.String())
			}
			return Arg{a.N, s}
		}
		var args []Arg
		for _, a := range raw.Args {
			args = append(args, argOf(a))
		}
		var instr Instr
		switch {
		case raw.Return != nil:
			instr = Instr{Kind: InstrReturn, Return: argOf(*raw.Return)}
		case raw.Run != nil:
			instr = Instr{Kind: InstrRun, N: *raw.Run, Args: args}
		default:
			instr = Instr{Kind: InstrDelegate, Task: *raw.Delegate, Args: args}
		}
		c, err := ClassifyInstr(standIn.Cut, scope, instr)
		switch e := err.(type) {
		case nil:
		case Unbound:
			return map[string]any{"reject": map[string]any{"unbound": e.N}}
		case NotTool:
			return map[string]any{"reject": map[string]any{"notTool": e.N}}
		case CutFails:
			return map[string]any{"reject": map[string]any{"cut": e.N}}
		case Narrows:
			return map[string]any{"reject": map[string]any{"narrows": []any{e.Passed, e.Held}}}
		default:
			t.Fatal(err)
		}
		switch c.Kind {
		case InstrReturn:
			return map[string]any{"return": wObs([]Obs{c.Return})[0]}
		case InstrRun:
			return map[string]any{"run": c.Binary, "passed": wObs(c.Passed)}
		}
		a := ChildArrival(Arrival{}, c.Passed)
		table := map[string]any{}
		for n, p := range a.Table {
			table[itoa(n)] = p.cel()
		}
		tools := []any{}
		for _, b := range Tools(a.Next.Scope) {
			tools = append(tools, b)
		}
		return map[string]any{"child": c.Task, "passed": wObs(c.Passed), "table": table, "tools": tools}
	}
}

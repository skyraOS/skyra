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
		var raw struct {
			Return   *int
			Run      *int
			Delegate *string
			Args     []int
		}
		if err := json.Unmarshal(in["instr"], &raw); err != nil {
			t.Fatal(err)
		}
		var instr Instr
		switch {
		case raw.Return != nil:
			instr = Instr{Kind: InstrReturn, N: *raw.Return}
		case raw.Run != nil:
			instr = Instr{Kind: InstrRun, N: *raw.Run, Args: raw.Args}
		default:
			instr = Instr{Kind: InstrDelegate, Task: *raw.Delegate, Args: raw.Args}
		}
		c, err := ClassifyInstr(scope, instr)
		switch e := err.(type) {
		case nil:
		case Unbound:
			return map[string]any{"reject": map[string]any{"unbound": e.N}}
		case NotTool:
			return map[string]any{"reject": map[string]any{"notTool": e.N}}
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

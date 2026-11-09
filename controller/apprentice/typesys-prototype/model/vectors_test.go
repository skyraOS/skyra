// The vectors: every case Vectors.lean writes under vectors/model, replayed
// against the Go. A file with no replay fails unless it is still pending,
// and a replay with no file fails, so a change to either side alone fails.
package model

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"skyra/typesys-prototype/tree"
)

const vectorDir = "../invariants/vectors/model"

// pending is the vector files no replay reads yet: the port's next steps.
var pending = []string{}

type vcase struct {
	In  map[string]json.RawMessage `json:"in"`
	Out json.RawMessage            `json:"out"`
}

func loadVectors(t *testing.T, name string) []vcase {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(vectorDir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Function string  `json:"function"`
		Cases    []vcase `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if f.Function != "model/"+name || len(f.Cases) == 0 {
		t.Fatalf("%s: function %q, %d cases", name, f.Function, len(f.Cases))
	}
	return f.Cases
}

func dec[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
	return v
}

// normal is a value as JSON reads it back, so Go's and Lean's compare by
// content.
func normal(t *testing.T, v any) any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func opt[T any](v T, ok bool) any {
	if !ok {
		return nil
	}
	return v
}

// optValue is a value as the vectors write it, or null.
func optValue(v any, ok bool) any {
	if !ok {
		return nil
	}
	return ValueJSON(v)
}

// textIds are the vectors' hashes, spelled out as text, as Vectors.lean
// forms them.
type textIds struct{}

func (textIds) Child(p InstLoc, n int, s Scope) InstLoc {
	return InstLoc("child(" + string(p) + "," + strconv.Itoa(n) + "," + s.Text() + ")")
}

func (textIds) First(d DefLoc, g Grant) InstLoc {
	ds := make([]string, len(g.Defs))
	for i, x := range g.Defs {
		ds[i] = string(x)
	}
	return InstLoc("first(" + string(d) + ";" + strings.Join(ds, ",") + ")")
}

func (textIds) Slot(i InstLoc, n int) InstLoc { return InstLoc(string(i) + "#" + strconv.Itoa(n)) }

func (textIds) Program(p Program) DefLoc { return DefLoc("program(" + string(p.Version) + ")") }

func (textIds) Root(v DefLoc, forms map[string]Iface) DefLoc {
	return DefLoc("root(" + string(v) + ";" + compact(forms) + ")")
}

func (textIds) Class(ms []DefLoc) DefLoc {
	names := make([]string, len(ms))
	for i, d := range ms {
		names[i] = string(d)
	}
	return DefLoc("class(" + strings.Join(names, ",") + ")")
}

// version is a refinement as the vectors write it: each branch's
// instruction its template.
func version(t *testing.T, v tree.Version) any {
	bs := make([]any, len(v.Branches))
	for i, b := range v.Branches {
		tmpl, err := ParseTemplate(b.Instr)
		if err != nil {
			t.Fatal(err)
		}
		bs[i] = map[string]any{"id": b.ID, "rank": b.Rank, "voter": b.Voter, "cond": b.Cond, "instr": tmpl}
	}
	return map[string]any{"next": v.Next, "branches": bs}
}

var replays = map[string]func(*testing.T, map[string]json.RawMessage) any{
	"positions": func(t *testing.T, in map[string]json.RawMessage) any {
		return map[string]any{"seenAt": nonNil(dec[Frame](t, in["frame"]).SeenAt())}
	},
	"store": func(t *testing.T, in map[string]json.RawMessage) any {
		f, live := dec[Frame](t, in["frame"]), dec[Frame](t, in["live"])
		st, ok := f.Store(dec[Template](t, in["template"]))
		out := map[string]any{"stored": opt(st, ok), "bound": nil, "boundLive": nil}
		if ok {
			b, bok := f.Bind(st)
			bl, blok := live.Bind(st)
			out["bound"], out["boundLive"] = opt(b, bok), opt(bl, blok)
		}
		return out
	},
	"join": func(t *testing.T, in map[string]json.RawMessage) any {
		f := dec[Frame](t, in["frame"])
		ps, err := f.JoinParts(dec[[]Arg](t, in["args"]))
		var r JoinRefusal
		switch {
		case errors.As(err, &r) && r.Untext != nil:
			return map[string]any{"untext": *r.Untext}
		case errors.As(err, &r):
			return map[string]any{"copies": r.Copies}
		case err != nil:
			t.Fatal(err)
		}
		text, ok := f.JoinText(ps)
		return map[string]any{"parts": ps, "text": opt(text, ok)}
	},
	"machine": func(t *testing.T, in map[string]json.RawMessage) any {
		start := Replay(dec[[]Event](t, in["log"]))
		for _, c := range dec[[]DefLoc](t, in["refuted"]) {
			start.Refute(c)
		}
		m, err := NewMachine(textIds{}, dec[Program](t, in["program"]), start)
		if err != nil {
			t.Fatal(err)
		}
		inputs := []any{Answered{Bytes: []byte(dec[string](t, in["goal"]))}}
		for _, raw := range dec[[]map[string]json.RawMessage](t, in["inputs"]) {
			if g, ok := raw["wrote"]; ok {
				inputs = append(inputs, Wrote{Guard: dec[string](t, g)})
				continue
			}
			a := dec[struct {
				Stdout string `json:"stdout"`
				Exit   int    `json:"exit"`
				Stderr string `json:"stderr"`
			}](t, raw["answered"])
			inputs = append(inputs, Answered{Bytes: []byte(a.Stdout), Exit: a.Exit, Stderr: []byte(a.Stderr)})
		}
		var effects []any
		for _, i := range inputs {
			next, e, err := m.Step(i)
			if err != nil {
				effects = append(effects, map[string]any{"stuck": nil})
				break
			}
			m = next
			switch e := e.(type) {
			case RunLeaf:
				effects = append(effects, map[string]any{"runLeaf": map[string]any{"tool": e.Tool, "argv": e.Argv}})
			case Shadow:
				effects = append(effects, map[string]any{"shadow": map[string]any{"tool": e.Tool, "argv": e.Argv}})
			case Ask:
				effects = append(effects, map[string]any{"ask": map[string]any{"key": e.Key, "kind": e.Kind,
					"depth": e.Depth, "refused": e.Refused}})
			case Finished:
				effects = append(effects, map[string]any{"finished": ValueJSON(e.Value)})
			case Exhausted:
				effects = append(effects, map[string]any{"exhausted": ValueJSON(e.Value)})
			}
		}
		// Memory is the log's: the log it started from, then what this run
		// appended, replays to it, refutations aside (step_replays).
		replayed := Replay(append(dec[[]Event](t, in["log"]), m.Log...))
		for _, c := range m.Memory.RefutedClasses() {
			replayed.Refute(c)
		}
		if !maps.EqualFunc(replayed, m.Memory, func(a, b tree.Version) bool { return reflect.DeepEqual(a, b) }) {
			t.Fatalf("memory is not its log's replay")
		}
		var mem []any
		for _, k := range dec[[]Key](t, in["keys"]) {
			v, seen := m.Memory[k.Text()]
			mem = append(mem, opt(version(t, v), seen))
		}
		return map[string]any{"effects": effects, "tree": m.Tree, "memory": nonNil(mem),
			"refuted": nonNil(m.Memory.RefutedClasses()), "log": nonNil(m.Log)}
	},
	"memory": func(t *testing.T, in map[string]json.RawMessage) any {
		m := Replay(dec[[]Event](t, in["log"]))
		var out []any
		for _, k := range dec[[]Key](t, in["keys"]) {
			v, seen := m[k.Text()]
			out = append(out, opt(version(t, v), seen))
		}
		return out
	},
	"decide": func(t *testing.T, in map[string]json.RawMessage) any {
		vals := map[string]any{"arrived": map[string]any{"size": dec[int64](t, in["x"])}}
		o, err := Replay(dec[[]Event](t, in["log"])).Decide(dec[int](t, in["inhibition"]), dec[int](t, in["floor"]),
			dec[int](t, in["quorum"]), dec[Key](t, in["key"]), vals)
		if err != nil {
			t.Fatal(err)
		}
		return o
	},
	"step": func(t *testing.T, in map[string]json.RawMessage) any {
		rs := dec[map[DefLoc]map[string][]DefLoc](t, in["reaches"])
		reaches := func(d DefLoc, name string) []DefLoc { return rs[d][name] }
		tr := dec[Tree](t, in["tree"])
		next, ok := tr.Step(textIds{}, dec[int](t, in["act"]), reaches, dec[InstLoc](t, in["frame"]), dec[Edge](t, in["edge"]))
		return opt(next, ok)
	},
	"boot": func(t *testing.T, in map[string]json.RawMessage) any {
		var from *Booter
		if string(in["from"]) != "null" {
			x := dec[struct {
				Frame  Frame `json:"frame"`
				Invoke int   `json:"invoke"`
			}](t, in["from"])
			from = &Booter{Frame: x.Frame, Invoke: x.Invoke}
		}
		f, ok := Boot(textIds{}, dec[Program](t, in["program"]), dec[Grant](t, in["grant"]), from)
		return opt(f, ok)
	},
	"cut": func(t *testing.T, in map[string]json.RawMessage) any {
		v := DecodeValue(dec[any](t, in["value"]))
		sels := dec[[]Sel](t, in["sels"])
		cut, ok := CutAll(v, sels)
		i := CutIface(ShapeIface(tree.ShapeOf(v)), sels)
		r, rok := i.Reduce()
		return map[string]any{"value": optValue(cut, ok), "iface": i, "reduced": opt(r, rok)}
	},
	"read": func(t *testing.T, in map[string]json.RawMessage) any {
		f := dec[Frame](t, in["frame"])
		n := dec[int](t, in["n"])
		sels := dec[[]Sel](t, in["sels"])
		v, ok := f.Read(n, sels)
		a := f.Inline(n, sels)
		if ev, eok := f.Eval(a); eok != ok || ok && !reflect.DeepEqual(ev, v) {
			t.Errorf("$%d%v: inlined %v reads %v, %v; the reference %v, %v", n, sels, a, ev, eok, v, ok)
		}
		return map[string]any{"read": optValue(v, ok), "inline": a}
	},
	"introduce": func(t *testing.T, in map[string]json.RawMessage) any {
		f := dec[Frame](t, in["frame"])
		e, n, ok := f.Introduce(dec[Lit](t, in["lit"]))
		if !ok {
			return map[string]any{"copies": n}
		}
		return map[string]any{"entry": e}
	},
	"key": func(t *testing.T, in map[string]json.RawMessage) any {
		f := dec[Frame](t, in["frame"])
		return f.Key(func(d DefLoc) DefLoc { return d })
	},
	"kind": func(t *testing.T, in map[string]json.RawMessage) any {
		k := dec[int](t, in["act"])
		tools := dec[[]DefLoc](t, in["tools"])
		o := dec[Op](t, in["op"])
		es := dec[[]Iface](t, in["entries"])
		kind, ok := Pushed(k, tools, o, es)
		return map[string]any{"kind": opt(kind, ok), "toolTypes": nonNil(ToolTypes(tools, es))}
	},
}

func TestVectors(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(vectorDir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vectors in %s: %v", vectorDir, err)
	}
	var names []string
	for _, f := range files {
		names = append(names, strings.TrimSuffix(filepath.Base(f), ".json"))
	}
	for name := range replays {
		if !slices.Contains(names, name) {
			t.Errorf("a replay for %s, and no vectors", name)
		}
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			r, ok := replays[name]
			if !ok {
				if slices.Contains(pending, name) {
					t.Skip("pending: no replay yet")
				}
				t.Fatalf("vectors for %s, and no replay", name)
			}
			for i, c := range loadVectors(t, name) {
				got, want := normal(t, r(t, c.In)), normal(t, c.Out)
				if !reflect.DeepEqual(got, want) {
					in, _ := json.Marshal(c.In)
					t.Fatalf("case %d: %s\n got  %v\n want %v", i, in, got, want)
				}
			}
		})
	}
}

// Shapes round-trip through their encoding.
func TestParseShape(t *testing.T) {
	for _, enc := range []string{"null", "boolean", "number", "string", "bytes", "exit(0)", "exit(12)",
		"[]", "[number,string]", "{}", `{"a":null}`, `{"a":{"b":[string]},"c\"d":number}`, `{"\u0001":bytes}`} {
		s, err := ParseShape(enc)
		if err != nil {
			t.Fatalf("%s: %v", enc, err)
		}
		if got := s.Enc(); got != enc {
			t.Errorf("%s: round-trips as %s", enc, got)
		}
	}
	for _, bad := range []string{"", "nul", "[string", `{"a"}`, "exit(x)", "string,"} {
		if _, err := ParseShape(bad); err == nil {
			t.Errorf("%q: parsed", bad)
		}
	}
}

// Selectors round-trip through their encoding.
func TestParseSel(t *testing.T) {
	for _, enc := range []string{"", ".dir", ".dir.path", "[]", "[2:3]", "[1:1,3:3]"} {
		s, err := ParseSel(enc)
		if err != nil {
			t.Fatalf("%s: %v", enc, err)
		}
		if got := s.Enc(); got != enc {
			t.Errorf("%s: round-trips as %s", enc, got)
		}
	}
}

// The tree is a fold of its edges: replaying the edges the step vectors
// accept, from their first tree, gives the last tree they reach.
func TestReplay(t *testing.T) {
	cases := loadVectors(t, "step")
	rs := dec[map[DefLoc]map[string][]DefLoc](t, cases[0].In["reaches"])
	reaches := func(d DefLoc, name string) []DefLoc { return rs[d][name] }
	var log []LogEdge
	var last json.RawMessage
	for _, c := range cases {
		if string(c.Out) != "null" {
			log = append(log, LogEdge{Frame: dec[InstLoc](t, c.In["frame"]), Edge: dec[Edge](t, c.In["edge"])})
			last = c.Out
		}
	}
	got, ok := dec[Tree](t, cases[0].In["tree"]).Replay(textIds{}, dec[int](t, cases[0].In["act"]), reaches, log)
	if !ok || !reflect.DeepEqual(normal(t, got), normal(t, last)) {
		t.Fatalf("replaying %d edges: %v, %v; want %s", len(log), got, ok, last)
	}
}

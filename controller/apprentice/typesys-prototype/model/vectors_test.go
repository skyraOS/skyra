// The vectors: every case Vectors.lean writes under vectors/model, replayed
// against the Go. A file with no replay fails unless it is still pending,
// and a replay with no file fails, so a change to either side alone fails.
package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"skyra/typesys-prototype/tree"
)

const vectorDir = "../invariants/vectors/model"

// pending is the vector files no replay reads yet: the port's next steps.
var pending = []string{"boot", "step"}

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

var replays = map[string]func(*testing.T, map[string]json.RawMessage) any{
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

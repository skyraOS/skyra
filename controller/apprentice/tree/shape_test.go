package tree

import (
	"encoding/json"
	"strings"
	"testing"
)

// The replay of Shape.lean's vectors.
func init() {
	replays["shapeArrive"] = func(t *testing.T, _ Apprentice, in map[string]json.RawMessage) any {
		out := OutputText
		if dec[string](t, in["output"]) == "json" {
			out = OutputJSON
		}
		var raw []byte
		for _, n := range dec[[]int](t, in["bytes"]) {
			raw = append(raw, byte(n))
		}
		v, s := Arrive(out, raw)
		return map[string]any{"value": numbered(v), "shape": s.Enc()}
	}
}

// numbered writes bytes as the vectors do: a list of numbers.
func numbered(v any) any {
	switch v := v.(type) {
	case []byte:
		out := []any{}
		for _, b := range v {
			out = append(out, int(b))
		}
		return out
	case []any:
		out := []any{}
		for _, x := range v {
			out = append(out, numbered(x))
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for k, x := range v {
			out[k] = numbered(x)
		}
		return out
	}
	return v
}

func shapeOfJSON(t *testing.T, src string) Shape {
	t.Helper()
	v, err := DecodeJSON([]byte(src))
	if err != nil {
		t.Fatalf("%s: %v", src, err)
	}
	return ShapeOf(v)
}

// Key order, array order and array length are values; key sets are strict;
// a key's kind is in the shape at any depth; one number kind.
func TestShapeRules(t *testing.T) {
	same := [][2]string{
		{`{"key1":1,"key2":"a"}`, `{"key2":"b","key1":2}`},
		{`[{"id":1},{"id":2},{"id":3}]`, `[{"id":9}]`},
		{`[1,"a"]`, `["b",2,2]`},
		{`{"size":5}`, `{"size":5.5}`},
	}
	for _, c := range same {
		if a, b := shapeOfJSON(t, c[0]).Enc(), shapeOfJSON(t, c[1]).Enc(); a != b {
			t.Errorf("%s and %s: %s ≠ %s", c[0], c[1], a, b)
		}
	}
	apart := [][2]string{
		{`{"a":1,"b":1}`, `{"a":1,"b":1,"c":1}`},
		{`{"size":5}`, `{"size":"5"}`},
		{`{"a":{"b":1}}`, `{"a":{"b":true}}`},
		{`[{"id":1}]`, `[{"id":1},{"name":"x"}]`},
	}
	for _, c := range apart {
		if a, b := shapeOfJSON(t, c[0]).Enc(), shapeOfJSON(t, c[1]).Enc(); a == b {
			t.Errorf("%s and %s share %s", c[0], c[1], a)
		}
	}
}

// Text is a string whatever it looks like; output that does not decode is
// an error holding the bytes, never a failure and never repaired; a tool's
// own {"error": …} is another shape.
func TestArrive(t *testing.T) {
	if v, s := Arrive(OutputText, []byte("5")); v != "5" || s.Kind != ShapeStr {
		t.Errorf("text 5 arrived as %#v %s", v, s.Enc())
	}
	if v, s := Arrive(OutputJSON, []byte("5")); v != 5.0 || s.Kind != ShapeNum {
		t.Errorf("json 5 arrived as %#v %s", v, s.Enc())
	}
	for _, c := range []struct {
		out     Output
		in      string
		at      float64
		message string
	}{
		{OutputText, "ok\xffno", 2, "not UTF-8 at byte 2"},
		{OutputJSON, "\"\xff\"", 1, "not UTF-8 at byte 1"},
		{OutputJSON, `{"a":`, 5, "not JSON at byte 5: the output is one JSON value"},
		{OutputJSON, `{"a":1} trailing`, 9, "not JSON at byte 9: the output is one JSON value"},
		{OutputJSON, `{"a":{"b":1,"b":2}}`, 15, `not JSON at byte 15: the key "b" appears twice`},
	} {
		v, s := Arrive(c.out, []byte(c.in))
		m, ok := v.(map[string]any)
		if !ok || string(m["error"].([]byte)) != c.in || m["at"] != c.at ||
			!strings.HasPrefix(m["message"].(string), c.message) ||
			s.Enc() != `{"at":number,"error":bytes,"message":string}` {
			t.Errorf("%q arrived as %#v %s, want the error at %v holding its bytes", c.in, v, s.Enc(), c.at)
		}
	}
	if _, s := Arrive(OutputJSON, []byte(`{"error":"disk full"}`)); s.Enc() != `{"error":string}` {
		t.Errorf("a tool's own error has shape %s", s.Enc())
	}
}

// A guard is checked against its base's shapes at commit: a key the shape
// lacks, or a kind used wrongly, is refused.
func TestCheckGuard(t *testing.T) {
	shapes := map[Position]Shape{
		Arrived:   shapeOfJSON(t, `{"size":1,"meta":{"name":"a"},"items":[{"id":1}]}`),
		Passed(0): {Kind: ShapeStr},
	}
	scope := map[int]Position{1: Arrived, 2: Passed(0)}
	for _, g := range []string{
		`$1.size > 7 ? "ls $1" : null`,
		`$1.meta.name == "a" ? "ls $1" : null`,
		`$1["meta"]["name"].startsWith("a") ? "ls $1" : null`,
		`$1.items[0].id == 3 ? "ls $1" : null`,
		`$1.items.exists(i, i.id > 2) ? "ls $1" : null`,
		`size($2) > 0 && $2.contains("error") ? "ls $2" : null`,
	} {
		chain, err := StoreGuard(g, scope)
		if err != nil {
			t.Fatalf("%s: %v", g, err)
		}
		if err := CheckGuard(chain, shapes); err != nil {
			t.Errorf("%s: refused: %v", g, err)
		}
	}
	for _, g := range []string{
		`$1.sise > 7 ? "ls $1" : null`,
		`$1.meta.nmae == "a" ? "ls $1" : null`,
		`$1.size.name == "a" ? "ls $1" : null`,
		`$1.items[0].idd == 3 ? "ls $1" : null`,
		`$2 > 3 ? "ls $2" : null`,
		`$2.size > 3 ? "ls $2" : null`,
	} {
		chain, err := StoreGuard(g, scope)
		if err != nil {
			t.Fatalf("%s: %v", g, err)
		}
		if err := CheckGuard(chain, shapes); err == nil {
			t.Errorf("%s: accepted", g)
		}
	}
}

// A guard reads an error's bytes, typed as bytes: the learned reaction to
// output that did not decode.
func TestGuardOnError(t *testing.T) {
	v, s := Arrive(OutputText, []byte("caf\xe9"))
	chain, err := StoreGuard(`size($0.error) > 0 ? "iconv -f latin1 $0" : null`, map[int]Position{0: Arrived})
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckGuard(chain, map[Position]Shape{Arrived: s}); err != nil {
		t.Fatalf("refused: %v", err)
	}
	ver, _, _, err := CommitChain(EmptyVersion(), chain)
	if err != nil {
		t.Fatal(err)
	}
	if sel, err := Distinguish(ver.Branches, map[string]any{"arrived": v}); err != nil || sel.Instr != "iconv -f latin1 @arrived" {
		t.Fatalf("selects %+v (%v)", sel, err)
	}
}

// Exiting 0 is Arrive; any other status is the failure, its own shape,
// which no tool's own JSON has.
func TestArriveExit(t *testing.T) {
	if v, s := ArriveExit(OutputText, []byte("ok"), 0, []byte("warn")); v != "ok" || s.Kind != ShapeStr {
		t.Fatalf("exit 0: %#v %s", v, s.Enc())
	}
	v, s := ArriveExit(OutputJSON, []byte(`{"partial":`), 2, []byte("boom"))
	m := v.(map[string]any)
	if s.Enc() != `{"exit":number,"stderr":bytes,"stdout":bytes}` || m["exit"] != 2.0 ||
		string(m["stdout"].([]byte)) != `{"partial":` || string(m["stderr"].([]byte)) != "boom" {
		t.Fatalf("failure: %#v %s", v, s.Enc())
	}
	if _, own := ArriveExit(OutputJSON, []byte(`{"exit":1,"stderr":"x","stdout":"y"}`), 0, nil); own.Enc() == s.Enc() {
		t.Fatal("a tool's own JSON has a failure's shape")
	}
}

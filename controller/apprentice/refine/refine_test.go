package refine

import (
	"strings"
	"testing"
)

var trace = Vars{
	Input:     map[string]any{"text": "relay:hello"},
	InputType: "input → {text: string}",
	Returns:   map[string]any{"r0c0": "relay:hello", "r0c1": map[string]any{"kind": "file"}},
	Types:     map[string]string{"r0c0": "echo → string", "r0c1": "read → {kind: string}"},
}

// Each check refuses what it should.
func TestChecks(t *testing.T) {
	for src, why := range map[string]string{
		`input.text +`:                         "does not check",
		`size(input.text)`:                     "not a predicate",
		`input.items.map(x, x > 1).size() > 0`: "does not check", // no map
		`input.items.filter(x, x > 1) == []`:   "does not check", // no filter
		`now() > 0`:                            "does not check",
		`input.a.all(x, input.b.all(y, input.c.all(z, x == y && y == z)))`: "may cost",
	} {
		if _, err := Compile(src); err == nil || !strings.Contains(err.Error(), why) {
			t.Errorf("%s: %v, want %q", src, err, why)
		}
	}
}

// A refinement is evaluated over the trace; an error inside it, even one
// || absorbed, is an error.
func TestEval(t *testing.T) {
	for src, want := range map[string]bool{
		`input.text.startsWith("relay:")`:                               true,
		`types.r0c0 == "echo → string" && type(returns.r0c0) == string`: true,
		`returns.r0c1.kind == "file"`:                                   true,
		`input.text.startsWith("other:")`:                               false,
		`has(input.note) && input.note == "x"`:                          false, // guarded: no error
		`["a", "b"].exists(x, input.text.endsWith(x))`:                  false,
	} {
		p, err := Compile(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		got, err := p.Eval(trace)
		if err != nil || got != want {
			t.Errorf("%s = %v, %v; want %v", src, got, err, want)
		}
	}
	for _, src := range []string{
		`input.txt.startsWith("relay:")`,        // a field that never exists
		`input.txt == "x" || true`,              // absorbed by ||
		`returns.r9 == "x" || input.text != ""`, // absorbed, a node that never returned
	} {
		p, err := Compile(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		if _, err := p.Eval(trace); err == nil {
			t.Errorf("%s: an error inside it must not pass", src)
		}
	}
}

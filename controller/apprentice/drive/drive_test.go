package drive

import (
	"strings"
	"testing"

	"skyra/apprentice/tree"
)

// A scripted model: at the goal it echoes it; at echo's return it returns
// what came back.
type script struct{ asked int }

func (s *script) Guard(a tree.AskModel) (string, error) {
	s.asked++
	if len(a.Scope) == 1 {
		return `size($1) > 0 ? "echo -n $1" : null`, nil
	}
	return `$2.startsWith("hello") ? "return $2" : null`, nil
}

// Real processes: the first run asks at both bases, the second asks
// nothing and gives the same return.
func TestRunLearns(t *testing.T) {
	refs := map[string]tree.Version{}
	for run, want := range []int{2, 0} {
		model := &script{}
		m := tree.NewMachine(Hashes, nil, refs, tree.Bounds{Transitions: 8, Refinements: 3})
		m, v, err := Run(m, []byte("hello world"), model, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		if v != "hello world" || model.asked != want {
			t.Fatalf("run %d: returned %#v after %d asks", run, v, model.asked)
		}
		refs = m.Refs
	}
}

// A value passed as $n is one argument: shell syntax in it is text.
func TestExecNoShell(t *testing.T) {
	out, err := Exec{}.Run(tree.RunInvoke{Argv: []string{"echo", "-n", "a; rm -rf / $(x)"}})
	if err != nil || string(out.Bytes) != "a; rm -rf / $(x)" || out.Exit != 0 {
		t.Fatalf("%q %v", out, err)
	}
}

type stubborn struct{}

func (stubborn) Guard(tree.AskModel) (string, error) { return "not a guard (", nil }

// A model that never writes a usable guard ends the frame with an error
// observation, the run's value, not a failure.
func TestRunAsksRunOut(t *testing.T) {
	m := tree.NewMachine(Hashes, nil, nil, tree.Bounds{Transitions: 8, Refinements: 3})
	_, v, err := Run(m, []byte("x"), stubborn{}, Exec{})
	e, ok := v.(map[string]any)
	if err != nil || !ok || !strings.Contains(string(e["error"].([]byte)), "refinement bound exhausted: the model was asked 3 times") {
		t.Fatalf("%#v %v", v, err)
	}
}

// A failing process is an observation at its own base: the tree learns to
// handle it, and the next failure is handled with no ask.
type onFailure struct{ asked int }

func (s *onFailure) Guard(a tree.AskModel) (string, error) {
	s.asked++
	switch arrived := a.Arrival.Vals["arrived"].(type) {
	case string: // the goal
		return `size($1) > 0 ? "ls $1" : null`, nil
	case map[string]any:
		if arrived["exit"] != nil {
			return `$2.exit != 0 ? "return $2" : null`, nil
		}
	}
	return "null", nil
}

func TestRunLearnsFailure(t *testing.T) {
	refs := map[string]tree.Version{}
	for run, want := range []int{2, 0} {
		model := &onFailure{}
		m := tree.NewMachine(Hashes, nil, refs, tree.Bounds{Transitions: 8, Refinements: 3})
		m, v, err := Run(m, []byte("/no/such/dir"), model, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		f, ok := v.(map[string]any)
		if !ok || f["exit"] == 0.0 || !strings.Contains(string(f["stderr"].([]byte)), "/no/such/dir") ||
			tree.ShapeOf(f).Enc() != `{"exit":number,"stderr":bytes,"stdout":bytes}` {
			t.Fatalf("run %d returned %#v", run, v)
		}
		if model.asked != want {
			t.Fatalf("run %d: %d asks", run, model.asked)
		}
		refs = m.Refs
	}
}

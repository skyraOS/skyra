package drive

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"skyra/typesys-prototype/model"
	"skyra/typesys-prototype/tree"
)

var testConstants = model.Constants{Act: 1, Transitions: 8, Refinements: 3, Inhibition: [2]int{0, 1}, Floor: 2, Quorum: 1}

func newMachine(t *testing.T, tools []string, mem model.Memory) model.Machine {
	t.Helper()
	m, err := NewMachine(testConstants, tools, mem)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// A scripted model granted echo: at the goal it echoes it; at echo's return it returns
// what came back.
type script struct{ asked int }

func (s *script) Guard(a model.Ask) (string, error) {
	s.asked++
	if len(a.Scope) == 2 { // echo, the goal
		return `size($1) > 0 ? "$2 $1" : null`, nil
	}
	return `$3.startsWith("hello") ? "return 0 $3" : null`, nil
}

// Real processes: the first run asks at both keys, the second asks
// nothing and gives the same return.
func TestRunLearns(t *testing.T) {
	var mem model.Memory
	for run, want := range []int{2, 0} {
		mdl := &script{}
		m, v, err := Run(newMachine(t, []string{"echo"}, mem), []byte("hello world"), mdl, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		r, _ := v.(map[string]any)
		if r["exit"] != tree.ExitStatus(0) || r["stdout"] != "hello world\n" || mdl.asked != want {
			t.Fatalf("run %d: returned %#v after %d asks", run, v, mdl.asked)
		}
		mem = m.Memory
	}
}

// A value passed as $n is one argument: shell syntax in it is text.
func TestExecNoShell(t *testing.T) {
	out, err := Exec{}.Run(model.RunLeaf{Argv: []string{"echo", "-n", "a; rm -rf / $(x)"}})
	if err != nil || string(out.Bytes) != "a; rm -rf / $(x)" || out.Exit != 0 {
		t.Fatalf("%q %v", out, err)
	}
}

type stubborn struct{}

func (stubborn) Guard(model.Ask) (string, error) { return "not a guard (", nil }

// A model that never writes a usable guard ends the frame with an error
// observation, the run's value, not a failure.
func TestRunAsksRunOut(t *testing.T) {
	_, v, err := Run(newMachine(t, []string{"echo"}, nil), []byte("x"), stubborn{}, Exec{})
	e, ok := v.(map[string]any)
	if err != nil || !ok || !strings.Contains(string(e["error"].([]byte)), "refinement bound exhausted: the model was asked 3 times") {
		t.Fatalf("%#v %v", v, err)
	}
}

// A failing process is an observation at its own key: the tree learns to
// handle it, and the next failure is handled with no ask.
type onFailure struct{ asked int }

func (s *onFailure) Guard(a model.Ask) (string, error) {
	s.asked++
	switch last := a.Scope[len(a.Scope)-1].Value.(type) {
	case nil: // the goal, then ls
		return `size($1) > 0 ? "$2 $1" : null`, nil
	case map[string]any:
		if last["exit"] != nil {
			return `$3.exit != 0 ? "return 1 $3.stdout $3.stderr" : null`, nil
		}
	}
	return "null", nil
}

func TestRunLearnsFailure(t *testing.T) {
	var mem model.Memory
	for run, want := range []int{2, 0} {
		mdl := &onFailure{}
		m, v, err := Run(newMachine(t, []string{"ls"}, mem), []byte("/no/such/dir"), mdl, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		f, ok := v.(map[string]any)
		if !ok || f["exit"] != tree.ExitStatus(1) || !strings.Contains(string(f["stderr"].([]byte)), "/no/such/dir") ||
			tree.ShapeOf(f).Enc() != `{"exit":exit(1),"stderr":bytes,"stdout":bytes}` {
			t.Fatalf("run %d returned %#v", run, v)
		}
		if mdl.asked != want {
			t.Fatalf("run %d: %d asks", run, mdl.asked)
		}
		mem = m.Memory
	}
}

// scripted answers each ask by its kind and how many $n it holds, and
// records what it was asked.
type scripted struct {
	guards map[string]string // by kind and scope length, "act 3"
	asks   []model.Ask
}

func (s *scripted) Guard(a model.Ask) (string, error) {
	s.asks = append(s.asks, a)
	return s.guards[string(a.Kind)+" "+strconv.Itoa(len(a.Scope))], nil
}

// A delegate frame holding two tools under k = 1 passes the goal and one
// tool to an act frame, which runs it and returns; the delegate returns
// what came back. The second run, at a new directory, asks nothing.
func TestRunDelegates(t *testing.T) {
	var mem model.Memory
	for run := range 2 {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "only.txt"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		s := &scripted{guards: map[string]string{
			"delegate 3": `true ? "delegate(goal($1), $2)" : null`, // $1 goal, $2 ls, $3 cat
			"act 2":      `true ? "$2 $1" : null`,                  // $1 goal, $2 ls
			"act 3":      `$3.size() > 0 ? "return 0 $3" : null`,   // ls's stdout arrived
			"delegate 4": `true ? "return 0 $4.stdout" : null`,     // the act frame's return arrived
		}}
		m, v, err := Run(newMachine(t, []string{"ls", "cat"}, mem), []byte(dir), s, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		r, _ := v.(map[string]any)
		if out, _ := r["stdout"].(string); !strings.Contains(out, "only.txt") {
			t.Fatalf("run %d returned %#v", run, v)
		}
		if want := []int{4, 0}[run]; len(s.asks) != want {
			t.Fatalf("run %d: %d asks, want %d", run, len(s.asks), want)
		}
		if run == 0 && (s.asks[1].Depth != 1 || s.asks[1].Kind != model.KindAct || s.asks[1].Model == nil) {
			t.Fatalf("the child's ask: %+v", s.asks[1])
		}
		mem = m.Memory
	}
}

// A guard the frame cannot run is refused with why, and the model asked
// again: an act frame does not delegate, and a tool is run from a $n
// holding one.
func TestRunRefuses(t *testing.T) {
	s := &scripted{guards: map[string]string{"act 2": `true ? "delegate(goal($1), $2)" : null`}}
	_, v, err := Run(newMachine(t, []string{"echo"}, nil), []byte("x"), s, Exec{})
	if err != nil || len(s.asks) != 3 || !strings.Contains(s.asks[1].Reason, "does not delegate") {
		t.Fatalf("%#v %v, asks %+v", v, err, s.asks)
	}
	s = &scripted{guards: map[string]string{"act 2": `true ? "$1 $2" : null`}}
	Run(newMachine(t, []string{"echo"}, nil), []byte("x"), s, Exec{})
	if !strings.Contains(s.asks[1].Reason, "$1 is not a tool") {
		t.Fatalf("reason %q", s.asks[1].Reason)
	}
}

// A frame past its transition bound ends with its error observation.
func TestRunTransitionBound(t *testing.T) {
	c := testConstants
	c.Transitions = 1
	m, err := NewMachine(c, []string{"echo"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, v, err := Run(m, []byte("hello"), &script{}, Exec{})
	e, ok := v.(map[string]any)
	if err != nil || !ok || !strings.Contains(string(e["error"].([]byte)), "transition bound exhausted") {
		t.Fatalf("%#v %v", v, err)
	}
}

// A guard that does not evaluate on the values it was written for is
// refused with the error, and nothing is committed: the model is asked
// again, and the run does not fail.
func TestRunRefusesEvalError(t *testing.T) {
	s := &scripted{guards: map[string]string{"act 2": `int($1) > 0 ? "$2 $1" : null`}}
	m, v, err := Run(newMachine(t, []string{"echo"}, nil), []byte("not a number"), s, Exec{})
	if err != nil || len(s.asks) != 3 || !strings.Contains(s.asks[1].Reason, "does not evaluate on these values") {
		t.Fatalf("%#v %v, asks %+v", v, err, s.asks)
	}
	for k, ver := range m.Memory {
		if len(ver.Branches) != 0 {
			t.Fatalf("committed at %s: %+v", k, ver.Branches)
		}
	}
}

// A return keeps its value's type: the act frame returns ls's text, and
// the delegate reads it as text, with CEL's string methods, which bytes
// would refuse.
func TestReturnKeepsText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "only.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	s := &scripted{guards: map[string]string{
		"delegate 3": `true ? "delegate(goal($1), $2)" : null`,
		"act 2":      `true ? "$2 $1" : null`,
		"act 3":      `true ? "return 0 $3" : null`,
		"delegate 4": `$4.stdout.trim() == "only.txt" ? "return 0 $4.stdout" : null`,
	}}
	_, v, err := Run(newMachine(t, []string{"ls", "cat"}, nil), []byte(dir), s, Exec{})
	r, _ := v.(map[string]any)
	if err != nil || r["stdout"] != "only.txt\n" {
		t.Fatalf("%#v %v, last reason %q", v, err, s.asks[len(s.asks)-1].Reason)
	}
	if got := tree.ShapeOf(r).Enc(); got != `{"exit":exit(0),"stdout":string}` {
		t.Fatalf("shape %s", got)
	}
}

// Under a quorum of 2 nothing is trusted until a key holds two
// implementations: the first two runs each ask at both keys, committing
// one more each time, and the third asks nothing.
func TestRunQuorum(t *testing.T) {
	c := testConstants
	c.Quorum = 2
	var mem model.Memory
	for run, want := range []int{2, 2, 0} {
		mdl := &script{}
		m, err := NewMachine(c, []string{"echo"}, mem)
		if err != nil {
			t.Fatal(err)
		}
		m, v, err := Run(m, []byte("hello world"), mdl, Exec{})
		if r, _ := v.(map[string]any); err != nil || r["stdout"] != "hello world\n" || mdl.asked != want {
			t.Fatalf("run %d: %#v %v after %d asks", run, v, err, mdl.asked)
		}
		mem = m.Memory
	}
	for k, v := range mem {
		if n := model.Implementations(v); n != 2 {
			t.Fatalf("%s holds %d implementations", k, n)
		}
	}
}

// A literal is an introduction: echo runs with the flag the model wrote,
// recorded as $3 in the act frame, so echo's output arrives as $4; the next
// run, learned, records the same and asks nothing. A literal copying what
// the frame holds is refused, naming the $n to cut instead.
func TestRunIntroduces(t *testing.T) {
	var mem model.Memory
	for run, goal := range []string{"hello", "world"} {
		s := &scripted{guards: map[string]string{
			"act 2": `true ? "$2 '-n' $1" : null`,
			"act 4": `true ? "return 0 $4" : null`,
		}}
		m, v, err := Run(newMachine(t, []string{"echo"}, mem), []byte(goal), s, Exec{})
		if r, _ := v.(map[string]any); err != nil || r["stdout"] != goal || len(s.asks) != []int{2, 0}[run] {
			t.Fatalf("run %d: %#v %v after %d asks", run, v, err, len(s.asks))
		}
		if len(s.asks) == 2 && s.asks[1].Scope[2].Origin != "introduced" {
			t.Fatalf("run %d: $3 shown as %+v", run, s.asks[1].Scope[2])
		}
		st := m.Tree.Frames[0].State
		if len(st) != 3 || st[0].Introduce == nil || st[0].Introduce.Str != "-n" || st[1].Arrived == nil ||
			st[2].Introduce == nil || !st[2].Introduce.IsExit {
			t.Fatalf("run %d: the act frame's state %+v", run, st)
		}
		mem = m.Memory
	}
	// Only what a cut takes is refused: a line of the goal is, part of one
	// is the model's to write.
	s := &scripted{guards: map[string]string{"act 2": `true ? "$2 'world' $1" : null`}}
	Run(newMachine(t, []string{"echo"}, nil), []byte("hello\nworld"), s, Exec{})
	if len(s.asks) < 2 || !strings.Contains(s.asks[1].Reason, "'world' is text a cut of $1 takes") {
		t.Fatalf("asks %+v", s.asks)
	}
	s = &scripted{guards: map[string]string{"act 2": `true ? "$2 'ell' $1" : null`, "act 4": `true ? "return 0 $4" : null`}}
	if _, v, err := Run(newMachine(t, []string{"echo"}, nil), []byte("hello"), s, Exec{}); err != nil ||
		v.(map[string]any)["stdout"] != "ell hello\n" {
		t.Fatalf("%#v %v after asks %+v", v, err, s.asks)
	}
}

// A join is an introduction built from parts: echo runs on the path the
// model joined from the goal's two lines, recorded as the line cuts, then
// the join; the next run, learned, joins its own lines and asks nothing. A
// join reading no text, or writing what a cut takes, is refused.
func TestRunJoins(t *testing.T) {
	var mem model.Memory
	for run, goal := range []string{"/tmp\na.txt", "/var\nb.md"} {
		s := &scripted{guards: map[string]string{
			"act 2": `true ? "$2 $1[1]'/'$1[2]" : null`, // $3, $4: the cuts; $5: the join; echo's output $6
			"act 6": `true ? "return 0 $6" : null`,
		}}
		m, v, err := Run(newMachine(t, []string{"echo"}, mem), []byte(goal), s, Exec{})
		want := strings.Replace(goal, "\n", "/", 1) + "\n"
		if r, _ := v.(map[string]any); err != nil || r["stdout"] != want || len(s.asks) != []int{2, 0}[run] {
			t.Fatalf("run %d: %#v %v after %d asks", run, v, err, len(s.asks))
		}
		if len(s.asks) == 2 && s.asks[1].Scope[4].Origin != "joined from $1[1:1]'/'$1[2:2]" {
			t.Fatalf("run %d: $5 shown as %+v", run, s.asks[1].Scope[4])
		}
		st := m.Tree.Frames[0].State
		if len(st) != 5 || st[0].Eliminate != 1 || st[1].Eliminate != 1 || st[2].Join == nil || st[3].Arrived == nil ||
			st[4].Introduce == nil {
			t.Fatalf("run %d: the act frame's state %+v", run, st)
		}
		mem = m.Memory
	}
	for guard, reason := range map[string]string{
		`true ? "$2 $2'/x'" : null`:             "$2 reads no text",
		`true ? "$2 'a/'$1" : null`:             "a cut of $1 takes text the join writes",
		`true ? "delegate(goal($1'x'))" : null`: "a goal is taken from a parameter",
	} {
		s := &scripted{guards: map[string]string{"act 2": guard}}
		Run(newMachine(t, []string{"echo"}, nil), []byte("a/"), s, Exec{})
		if len(s.asks) < 2 || !strings.Contains(s.asks[1].Reason, reason) {
			t.Errorf("%s: asks %+v", guard, s.asks)
		}
	}
}

// A cut is an elimination, recorded as the frame's next $n, and stored by
// key position: the second run, whose goal arrives at the same key,
// records the same entries and asks nothing. A condition reading an
// elimination's $n is refused: conditions read what was seen.
func TestRunEliminates(t *testing.T) {
	var mem model.Memory
	for run, goal := range []string{"ignored\nhello", "skipped\nworld"} {
		s := &scripted{guards: map[string]string{
			"act 2": `true ? "$2 $1[2]" : null`,       // $3: the cut $1[2]; echo's output arrives as $4
			"act 4": `true ? "return 0 $4[1]" : null`, // $5: the cut $4[1]
		}}
		m, v, err := Run(newMachine(t, []string{"echo"}, mem), []byte(goal), s, Exec{})
		want := strings.Split(goal, "\n")[1]
		if r, _ := v.(map[string]any); err != nil || r["stdout"] != want || len(s.asks) != []int{2, 0}[run] {
			t.Fatalf("run %d: %#v %v after %d asks", run, v, err, len(s.asks))
		}
		st := m.Tree.Frames[0].State
		if len(st) != 4 || st[0].Eliminate != 1 || st[1].Arrived == nil || st[2].Introduce == nil || st[3].Eliminate != 4 {
			t.Fatalf("run %d: the act frame's state %+v", run, st)
		}
		mem = m.Memory
	}
	s := &scripted{guards: map[string]string{
		"act 2": `true ? "$2 $1[2]" : null`,
		"act 4": `$3 == "x" ? "return 0 $4" : null`,
	}}
	Run(newMachine(t, []string{"echo"}, nil), []byte("a\nb"), s, Exec{})
	if len(s.asks) < 3 || !strings.Contains(s.asks[2].Reason, "$3 was made in this frame") {
		t.Fatalf("asks %+v", s.asks)
	}
}

// Memory stores by key position, so a stored $n means the same whatever
// was eliminated before it. Run 0 cuts the goal ($3, an elimination), so
// echo's output arrives as $4; run 1 takes the other branch, cutting
// nothing, so it arrives as $3. Both reach one key, and the return run 0
// stored there binds to the arrival in each.
func TestStoredByKeyPosition(t *testing.T) {
	var mem model.Memory
	for run, goal := range []string{"skip\nhello", "world"} {
		s := &scripted{guards: map[string]string{
			"act 2": `$1.contains("\n") ? "$2 $1[2]" : true ? "$2 $1" : null`,
			"act 4": `true ? "return 0 $4" : null`,
		}}
		m, v, err := Run(newMachine(t, []string{"echo"}, mem), []byte(goal), s, Exec{})
		want := []string{"hello\n", "world\n"}[run]
		if r, _ := v.(map[string]any); err != nil || r["stdout"] != want || len(s.asks) != []int{2, 0}[run] {
			t.Fatalf("run %d: %#v %v after %d asks", run, v, err, len(s.asks))
		}
		mem = m.Memory
	}
}

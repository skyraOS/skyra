package drive

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"skyra/typesys-prototype/model"
)

// delegating is the scripted model of a root holding ls and cat that hands
// its goal to an ls child and returns what came back.
func delegating() *scripted {
	return &scripted{guards: map[string]string{
		"delegate 3": `true ? "delegate(goal($1), $2)" : null`,
		"act 2":      `true ? "$2 $1" : null`,
		"act 3":      `$3.size() > 0 ? "return 0 $3" : null`,
		"delegate 4": `true ? "return 0 $4.stdout" : null`,
	}}
}

// runStored runs a goal from the store and records what it learned.
func runStored(t *testing.T, s Store, p model.Program, goal string, mdl Model) (model.Machine, any) {
	t.Helper()
	start, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.NewMachine(Ids{}, p, start)
	if err != nil {
		t.Fatal(err)
	}
	m, v, err := Run(m, []byte(goal), mdl, Exec{})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Record(start, m); err != nil {
		t.Fatal(err)
	}
	return m, v
}

// What one process learns, the next loads: the store replays to the memory
// the run ended with, and a second process asks nothing.
func TestStoreLearns(t *testing.T) {
	s := Store{Path: filepath.Join(t.TempDir(), "log.jsonl")}
	p := Program(testConstants, []string{"ls", "cat"})
	for run := range 2 {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "only.txt"), "")
		mdl := delegating()
		m, v := runStored(t, s, p, dir, mdl)
		if out, _ := v.(map[string]any)["stdout"].(string); !strings.Contains(out, "only.txt") {
			t.Fatalf("run %d returned %#v", run, v)
		}
		if want := []int{4, 0}[run]; len(mdl.asks) != want {
			t.Fatalf("run %d: %d asks, want %d", run, len(mdl.asks), want)
		}
		loaded, err := s.Load()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(loaded, m.Memory) {
			t.Fatalf("run %d: the store does not replay to the run's memory", run)
		}
	}
}

// A class refuted in one process stays refuted in the next.
func TestStoreKeepsRefutations(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "only.txt"), "")
	lister := []model.DefLoc{"ls", "find"}
	p := Program(testConstants, []string{"ls", "find"})
	p.ReadOnly, p.Classes = lister, [][]model.DefLoc{lister}
	refuting := func() *scripted {
		return &scripted{guards: map[string]string{
			"delegate 3": `true ? "delegate(goal($1), $2)" : null`,
			"act 2":      `true ? "$2 $1" : null`,
			"act 3":      `$3.startsWith("only") ? "return 0 $3" : true ? "return 1 $3" : null`,
			"delegate 4": `true ? "delegate(goal($1), $3)" : null`,
			"delegate 5": `true ? "return 0 $5.stdout" : null`,
		}}
	}
	s := Store{Path: filepath.Join(t.TempDir(), "log.jsonl")}
	runStored(t, s, p, dir, refuting())
	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Refuted(model.ClassID(Ids{}, lister)) {
		t.Fatal("the refutation was not kept")
	}
	runStored(t, s, p, dir, refuting())
	b, err := os.ReadFile(s.Path)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(b), `"refute"`); n != 1 {
		t.Fatalf("the class is recorded refuted %d times", n)
	}
}

// A write cut short loses its run, never the store: the last line without
// a newline is ignored. A line that does not read anywhere else is an
// error.
func TestStoreTornWrite(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "only.txt"), "")
	s := Store{Path: filepath.Join(t.TempDir(), "log.jsonl")}
	m, _ := runStored(t, s, Program(testConstants, []string{"ls", "cat"}), dir, delegating())
	f, err := os.OpenFile(s.Path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"startle":{"root":`)
	f.Close()
	loaded, err := s.Load()
	if err != nil || !reflect.DeepEqual(loaded, m.Memory) {
		t.Fatalf("a torn last line: %v", err)
	}
	f, _ = os.OpenFile(s.Path, os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString("\n{}\n")
	f.Close()
	if _, err := s.Load(); err == nil {
		t.Fatal("a line that is no event was read")
	}
}

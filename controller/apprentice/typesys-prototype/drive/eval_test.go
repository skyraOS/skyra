package drive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/typesys-prototype/model"
)

// An evalTask is a task run twice against the claude CLI and real
// processes, on different inputs: the first run learns, the second should
// ask nothing.
type evalTask struct {
	name  string
	tools []string
	// setup makes run's input under dir: the goal, and a check of what the
	// tree returned.
	setup func(t *testing.T, dir string, run int) (goal string, ok func(string) bool)
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

var evalTasks = []evalTask{
	{"count lines", []string{"wc"}, func(t *testing.T, dir string, run int) (string, func(string) bool) {
		n := []int{3, 7}[run]
		path := filepath.Join(dir, "notes.txt")
		write(t, path, strings.Repeat("a line\n", n))
		return "count the lines in this file:\n" + path, func(s string) bool { return strings.Contains(s, fmt.Sprint(n)) }
	}},
	{"missing directory", []string{"ls"}, func(t *testing.T, dir string, run int) (string, func(string) bool) {
		path := filepath.Join(dir, []string{"gone", "never"}[run])
		return "list the files in this directory, or report why it cannot be listed:\n" + path,
			func(s string) bool { return strings.Contains(s, "No such file") }
	}},
	{"json field", []string{"jq"}, func(t *testing.T, dir string, run int) (string, func(string) bool) {
		version := []string{"1.2.3", "4.0.0-rc1"}[run]
		path := filepath.Join(dir, "package.json")
		write(t, path, `{"name": "pkg", "version": "`+version+`", "private": true}`)
		return "read the version from this package file:\n" + path, func(s string) bool { return strings.Contains(s, version) }
	}},
	{"delegate", []string{"ls", "cat"}, func(t *testing.T, dir string, run int) (string, func(string) bool) {
		name, text := []string{"a.txt", "first.md"}[run], []string{"alpha", "# heading"}[run]
		write(t, filepath.Join(dir, name), text)
		return "show the contents of the only file in this directory:\n" + dir,
			func(s string) bool { return strings.Contains(s, text) }
	}},
}

// recorder counts what a run asks: each ask's depth, and each refusal's
// reason.
type recorder struct {
	Model
	t        *testing.T
	asks     int
	depth    int
	refusals []string
}

func (r *recorder) Guard(a model.Ask) (string, error) {
	r.asks++
	r.depth = max(r.depth, a.Depth)
	if a.Reason != "" {
		r.refusals = append(r.refusals, a.Reason)
	}
	g, err := r.Model.Guard(a)
	r.t.Logf("ask %d at depth %d%s\n%s=> %s", r.asks, a.Depth, map[bool]string{true: ", after a refusal"}[a.Reason != ""], Render(a), g)
	return g, err
}

// TestClaudeTasks is the live eval, when SKYRA_LIVE is set. Each task runs
// twice; it fails a task whose answer is wrong or whose second run asks.
func TestClaudeTasks(t *testing.T) {
	if os.Getenv("SKYRA_LIVE") == "" {
		t.Skip("set SKYRA_LIVE=1 to run the eval against the claude CLI")
	}
	var summary []string
	for _, task := range evalTasks {
		t.Run(task.name, func(t *testing.T) {
			var mem model.Memory
			var row []string
			for run := range 2 {
				dir := t.TempDir()
				goal, ok := task.setup(t, dir, run)
				rec := &recorder{Model: NewClaudeModel(), t: t}
				m, err := NewMachine(model.DefaultConstants, task.tools, mem)
				if err != nil {
					t.Fatal(err)
				}
				m, v, err := Run(m, []byte(goal), rec, Exec{})
				if err != nil {
					t.Errorf("run %d: %v", run, err)
					row = append(row, "error")
					return
				}
				mem = m.Memory
				got := Text(v)
				right := ok(got)
				if !right {
					t.Errorf("run %d returned %q", run, got)
				}
				if run == 1 && rec.asks > 0 {
					t.Errorf("run 1 asked %d times: the first run did not learn it", rec.asks)
				}
				for _, why := range rec.refusals {
					t.Logf("run %d refused: %s", run, why)
				}
				row = append(row, fmt.Sprintf("run %d: %d asks, %d refused, depth %d, %s", run, rec.asks, len(rec.refusals),
					rec.depth, map[bool]string{true: "right", false: "wrong"}[right]))
			}
			summary = append(summary, fmt.Sprintf("%-18s %s", task.name, strings.Join(row, " | ")))
		})
	}
	t.Logf("summary\n%s", strings.Join(summary, "\n"))
}

// A classTask is a pair of read-only tools held one class, and a goal
// either should meet: the root, scripted, hands it to a child holding the
// first, then to one holding the second, whose keys have nothing. Each
// child's asks are Claude's.
type classTask struct {
	name  string
	tools [2]string
	setup func(t *testing.T, dir string) (goal string, ok func(string) bool)
}

var classTasks = []classTask{
	{"cat and head", [2]string{"cat", "head"}, func(t *testing.T, dir string) (string, func(string) bool) {
		path := filepath.Join(dir, "notes.txt")
		write(t, path, "alpha\nbeta\n")
		return "show the contents of this file:\n" + path, func(s string) bool { return strings.Contains(s, "beta") && strings.Contains(s, `"exit":0`) }
	}},
	{"ls and find", [2]string{"ls", "find"}, func(t *testing.T, dir string) (string, func(string) bool) {
		write(t, filepath.Join(dir, "a.txt"), "")
		write(t, filepath.Join(dir, "b.txt"), "")
		return "list the files in this directory:\n" + dir,
			func(s string) bool {
				return strings.Contains(s, "a.txt") && strings.Contains(s, "b.txt") && strings.Contains(s, `"exit":0`)
			}
	}},
	// Not interchangeable: wc counts what cat shows.
	{"cat and wc", [2]string{"cat", "wc"}, func(t *testing.T, dir string) (string, func(string) bool) {
		path := filepath.Join(dir, "notes.txt")
		write(t, path, "alpha\nbeta\n")
		return "show the contents of this file:\n" + path, func(s string) bool { return strings.Contains(s, "beta") && strings.Contains(s, `"exit":0`) }
	}},
}

// rootScript answers the root's asks itself and the children's with the
// live model: the root delegates to the first tool, then the second, then
// returns what the second's child returned, its exit status with it.
type rootScript struct {
	*recorder
	childAsks map[string]int // by the tool the child holds
}

func (r *rootScript) Guard(a model.Ask) (string, error) {
	if a.Depth == 0 {
		return map[int]string{
			3: `true ? "delegate(goal($1), $2)" : null`,
			4: `true ? "delegate(goal($1), $3)" : null`,
			5: `true ? "return 0 $5" : null`,
		}[len(a.Scope)], nil
	}
	for _, s := range a.Scope {
		if s.Role == model.RoleTool {
			r.childAsks[string(s.Tool)]++
		}
	}
	return r.recorder.Guard(a)
}

// TestClaudeClasses is the live eval of classes, when SKYRA_LIVE is set:
// each pair held one class, then undeclared. It reports, for the second
// tool's child, how often it was asked, whether the class was refuted, and
// whether what it returned meets the goal.
func TestClaudeClasses(t *testing.T) {
	if os.Getenv("SKYRA_LIVE") == "" {
		t.Skip("set SKYRA_LIVE=1 to run the eval against the claude CLI")
	}
	var summary []string
	for _, task := range classTasks {
		t.Run(task.name, func(t *testing.T) {
			pair := []model.DefLoc{model.DefLoc(task.tools[0]), model.DefLoc(task.tools[1])}
			for _, classed := range []bool{true, false} {
				dir := t.TempDir()
				goal, ok := task.setup(t, dir)
				p := Program(model.DefaultConstants, task.tools[:])
				if classed {
					p.ReadOnly, p.Classes = pair, [][]model.DefLoc{pair}
				}
				m, err := model.NewMachine(Ids{}, p, nil)
				if err != nil {
					t.Fatal(err)
				}
				s := &rootScript{recorder: &recorder{Model: NewClaudeModel(), t: t}, childAsks: map[string]int{}}
				w := &counted{}
				m, v, err := Run(m, []byte(goal), s, w)
				if err != nil {
					t.Errorf("classed %v: %v", classed, err)
					continue
				}
				got := Text(v.(map[string]any)["stdout"])
				refuted := m.Memory.Refuted(model.ClassID(Ids{}, pair))
				summary = append(summary, fmt.Sprintf("%-13s classed=%-5v %s child asked %d, %s child %d, refuted %-5v ran %v, %s: %q",
					task.name, classed, task.tools[0], s.childAsks[task.tools[0]], task.tools[1], s.childAsks[task.tools[1]],
					refuted, w.tools, map[bool]string{true: "meets the goal", false: "misses it"}[ok(got)], got))
			}
		})
	}
	t.Logf("summary\n%s", strings.Join(summary, "\n"))
}

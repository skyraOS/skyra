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

package drive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"skyra/typesys-prototype/tree"
)

func ask(scope ...tree.Obs) tree.AskModel {
	return tree.AskModel{Base: "b", Scope: scope, Arrival: tree.Arrival{Table: map[int]tree.Position{len(scope): tree.Arrived}}}
}

// A value past the show limit is truncated, and expanding reads the rest
// in parts to the end.
func TestRenderTruncatesAndExpands(t *testing.T) {
	defer func(n int) { ShowLimit = n }(ShowLimit)
	ShowLimit = 8
	a := ask(tree.Obs{Value: "abcdefghijklmnopqrst", Goal: true}, tree.Obs{Value: "ls", Tool: "ls"})
	r := Render(a)
	if !strings.HasPrefix(r, "Goal:\n$1 (string): abcdefgh\n…[truncated: 8 of 20 bytes shown; expand $1 from 8]\n\nTools:\n$2: ls\n\nContext:\nnone\n") ||
		strings.Contains(r, "abcdefghi") {
		t.Fatalf("%s", r)
	}
	var got string
	for off := 8; ; off += 8 {
		part, err := Expand(a, 1, off)
		if err != nil {
			t.Fatal(err)
		}
		got += part[:strings.Index(part, "\n…[")]
		if strings.Contains(part, "the end") {
			break
		}
	}
	if got != "ijklmnopqrst" {
		t.Fatalf("expanded %q", got)
	}
	if _, err := Expand(a, 3, 0); err == nil {
		t.Fatal("expanded $3, which is not in scope")
	}
}

// Truncation never splits a character, and an offset inside one reads
// from its start.
func TestTruncateAtRunes(t *testing.T) {
	defer func(n int) { ShowLimit = n }(ShowLimit)
	ShowLimit = 3
	a := ask(tree.Obs{Value: "aéé"}) // 1 + 2 + 2 bytes
	if r := Render(a); !strings.Contains(r, "$1 (string): aé\n…[truncated: 3 of 5") {
		t.Fatalf("%s", r)
	}
	if part, _ := Expand(a, 1, 4); !strings.HasPrefix(part, "é\n") {
		t.Fatalf("%q", part)
	}
	ShowLimit = 1
	if part, _ := Expand(a, 1, 1); !strings.HasPrefix(part, "é\n…[1 to 3 of 5") {
		t.Fatalf("a limit under one character: %q", part)
	}
}

// Bytes read as text when UTF-8 and as hex otherwise, inside JSON too.
func TestText(t *testing.T) {
	if s := Text([]byte{0xff, 0x01}); s != "hex:ff01" {
		t.Fatal(s)
	}
	if s := Text(map[string]any{"stdout": []byte("ok"), "exit": 1.0}); s != `{"exit":1,"stdout":"ok"}` {
		t.Fatal(s)
	}
}

// cli is a scripted claude CLI: each run gives the next reply as its JSON
// result, and records the arguments it was run with.
type cli struct {
	replies []string
	runs    [][]string
}

func (c *cli) run(args, env []string) ([]byte, error) {
	c.runs = append(c.runs, args)
	r := c.replies[0]
	c.replies = c.replies[1:]
	return json.Marshal(map[string]any{"result": r, "session_id": "s1", "is_error": false})
}

func prompt(args []string) string { return args[len(args)-1] }

// Claude may expand within one ask and then write its guard: one ask, one
// session, however many expansions.
func TestClaudeExpandsThenGuards(t *testing.T) {
	defer func(n int) { ShowLimit = n }(ShowLimit)
	ShowLimit = 4
	c := &cli{replies: []string{`{"expand": {"n": 3, "from": 4}}`, "```json\n{\"cel\": \"true ? \\\"return $2\\\" : null\"}\n```"}}
	m := &ClaudeModel{Model: "claude-opus-5-5", Run: c.run}
	g, err := m.Guard(ask(tree.Obs{Value: "x", Goal: true}, tree.Obs{Value: "ls", Tool: "ls"}, tree.Obs{Value: "a.log b.log"}))
	if err != nil || g != `true ? "return $2" : null` {
		t.Fatalf("%q %v", g, err)
	}
	if len(c.runs) != 2 || !strings.HasPrefix(prompt(c.runs[1]), "g b.") || !slices.Contains(c.runs[1], "--resume") {
		t.Fatalf("runs %q", c.runs)
	}
	if slices.Contains(c.runs[0], "--resume") || c.runs[0][slices.Index(c.runs[0], "--tools")+1] != "" {
		t.Fatalf("first run %q", c.runs[0])
	}
}

// Expansions are bounded on their own: past ExpandLimit the model is told
// to write its guard.
func TestClaudeExpandLimit(t *testing.T) {
	defer func(n int) { ExpandLimit = n }(ExpandLimit)
	ExpandLimit = 1
	c := &cli{replies: []string{`{"expand":{"n":1,"from":0}}`, `{"expand":{"n":1,"from":0}}`, `{"cel":"null"}`}}
	if _, err := (&ClaudeModel{Run: c.run}).Guard(ask(tree.Obs{Value: "x"})); err != nil {
		t.Fatal(err)
	}
	if p := prompt(c.runs[2]); !strings.Contains(p, "the most allowed") {
		t.Fatal(p)
	}
}

// A reply that is not one JSON object is taken as the guard.
func TestClaudePlainReply(t *testing.T) {
	c := &cli{replies: []string{`size($1) > 0 ? "return $1" : null`}}
	if g, _ := (&ClaudeModel{Run: c.run}).Guard(ask(tree.Obs{Value: "x"})); g != `size($1) > 0 ? "return $1" : null` {
		t.Fatal(g)
	}
}

// Inherited API credentials are refused, never used to bill.
func TestClaudeRefusesAPIKeys(t *testing.T) {
	if _, err := environment([]string{"HOME=/h", "ANTHROPIC_API_KEY=sk"}); err == nil {
		t.Fatal("an API key was accepted")
	}
	if env, err := environment([]string{"HOME=/h", "ANTHROPIC_API_KEY="}); err != nil || !slices.Equal(env, []string{"HOME=/h"}) {
		t.Fatalf("%q %v", env, err)
	}
}

// A live run against the claude CLI, under the user's login, and real
// processes, when SKYRA_LIVE is set. The tree is run twice; the second run should ask
// nothing.
func TestClaudeLive(t *testing.T) {
	if os.Getenv("SKYRA_LIVE") == "" {
		t.Skip("set SKYRA_LIVE=1 to run against the claude CLI")
	}
	model := &counting{Model: NewClaudeModel()}
	refs := map[string]tree.Version{}
	// Two runs on different directories: what the first learns should run
	// the second with no ask.
	for run, files := range [][]string{{"a.log", "b.log", "notes.txt"}, {"x.go", "y.go"}} {
		dir := t.TempDir()
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(dir, f), []byte(f), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		model.asks = 0
		m := tree.NewMachine(Hashes, nil, refs, tree.DefaultBounds, []string{"ls"})
		m, v, err := Run(m, []byte("list the files in this directory:\n"+dir), model, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("run %d: %d asks, returned %s", run, model.asks, Text(v))
		if want := strings.Join(files, "\n"); !strings.Contains(Text(v), files[0]) {
			t.Errorf("run %d returned %q, want the files %q", run, Text(v), want)
		}
		refs = m.Refs
	}
}

type counting struct {
	Model
	asks int
}

func (c *counting) Guard(a tree.AskModel) (string, error) {
	c.asks++
	g, err := c.Model.Guard(a)
	c.t(a, g)
	return g, err
}

func (c *counting) t(a tree.AskModel, g string) {
	fmt.Printf("--- ask at depth %d\n%s\n=> %s\n", a.Depth, Render(a), g)
}

// Each kind's prompt shows only what it may write, and neither says how
// many tools make a kind.
func TestSystemByKind(t *testing.T) {
	act, del := System(tree.KindActs), System(tree.KindDelegates)
	if strings.Contains(act, "delegate") || !strings.Contains(act, "run the tool $2") {
		t.Error("the act prompt offers delegation, or not running")
	}
	if !strings.Contains(del, "delegate(goal($1)") || strings.Contains(del, "run the tool") {
		t.Error("the delegate prompt offers running, or not delegation")
	}
	for _, p := range []string{act, del} {
		if strings.Contains(p, "act bound") || strings.Contains(p, "at most") {
			t.Error("a prompt speaks of the act bound")
		}
	}
}

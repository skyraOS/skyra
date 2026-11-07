package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"skyra/apprentice/tree"
)

func ask(scope ...tree.Obs) tree.AskModel {
	return tree.AskModel{Base: "b", Scope: scope, Arrival: tree.Arrival{Table: map[int]tree.Position{len(scope): tree.Arrived}}}
}

// A value past the show limit is truncated, and expanding reads the rest
// in parts to the end.
func TestRenderTruncatesAndExpands(t *testing.T) {
	defer func(n int) { ShowLimit = n }(ShowLimit)
	ShowLimit = 8
	a := ask(tree.Obs{Value: "ls", Tool: "ls"}, tree.Obs{Value: "abcdefghijklmnopqrst"})
	r := Render(a)
	if !strings.Contains(r, "$1 tool ls") || !strings.Contains(r, "abcdefgh\n…[truncated: 8 of 20 bytes shown; expand $2 from 8]") ||
		strings.Contains(r, "abcdefghi") || !strings.Contains(r, "In the base: $2") {
		t.Fatalf("%s", r)
	}
	var got string
	for off := 8; ; off += 8 {
		part, err := Expand(a, 2, off)
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
	if r := Render(a); !strings.Contains(r, "aé\n…[truncated: 3 of 5") {
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

// fake answers each request with the next scripted response, recording
// the tool results it was sent.
type fake struct {
	replies []string
	sent    []anthropic.BetaMessageNewParams
}

func (f *fake) New(_ context.Context, p anthropic.BetaMessageNewParams, _ ...option.RequestOption) (*anthropic.BetaMessage, error) {
	f.sent = append(f.sent, p)
	var m anthropic.BetaMessage
	if err := json.Unmarshal([]byte(f.replies[0]), &m); err != nil {
		return nil, err
	}
	f.replies = f.replies[1:]
	return &m, nil
}

func toolUse(id, name, input string) string {
	return `{"id":"m","type":"message","role":"assistant","model":"claude-opus-5-5","stop_reason":"tool_use",
		"content":[{"type":"tool_use","id":"` + id + `","name":"` + name + `","input":` + input + `}]}`
}

func lastResult(t *testing.T, p anthropic.BetaMessageNewParams) string {
	t.Helper()
	b, _ := json.Marshal(p.Messages[len(p.Messages)-1])
	return string(b)
}

// Claude may expand within one ask and then write its guard: one ask,
// however many expansions.
func TestClaudeExpandsThenGuards(t *testing.T) {
	defer func(n int) { ShowLimit = n }(ShowLimit)
	ShowLimit = 4
	f := &fake{replies: []string{toolUse("t1", "expand", `{"n":2,"from":4}`), toolUse("t2", "guard", `{"guard":"true ? \"return $2\" : null"}`)}}
	m := &ClaudeModel{Messages: f, Model: "claude-opus-5-5"}
	g, err := m.Guard(ask(tree.Obs{Value: "ls", Tool: "ls"}, tree.Obs{Value: "a.log b.log"}))
	if err != nil || g != `true ? "return $2" : null` {
		t.Fatalf("%q %v", g, err)
	}
	if len(f.sent) != 2 || !strings.Contains(lastResult(t, f.sent[1]), "g b.") {
		t.Fatalf("sent %d, last %s", len(f.sent), lastResult(t, f.sent[1]))
	}
}

// Expansions are bounded on their own: past ExpandLimit the model is told
// to write its guard.
func TestClaudeExpandLimit(t *testing.T) {
	defer func(n int) { ExpandLimit = n }(ExpandLimit)
	ExpandLimit = 1
	f := &fake{replies: []string{toolUse("t1", "expand", `{"n":1,"from":0}`), toolUse("t2", "expand", `{"n":1,"from":0}`),
		toolUse("t3", "guard", `{"guard":"null"}`)}}
	m := &ClaudeModel{Messages: f}
	if _, err := m.Guard(ask(tree.Obs{Value: "x"})); err != nil {
		t.Fatal(err)
	}
	if r := lastResult(t, f.sent[2]); !strings.Contains(r, "the most allowed") || !strings.Contains(r, `"is_error":true`) {
		t.Fatal(r)
	}
}

// A live run against Claude and real processes, when SKYRA_LIVE is set: it
// spends real tokens. The tree is run twice; the second run should ask
// nothing.
func TestClaudeLive(t *testing.T) {
	if os.Getenv("SKYRA_LIVE") == "" {
		t.Skip("set SKYRA_LIVE=1 to run against Claude")
	}
	dir := t.TempDir()
	for _, f := range []string{"a.log", "b.log", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	model := &counting{Model: NewClaudeModel()}
	refs := map[string]tree.Version{}
	for run := range 2 {
		model.asks = 0
		m := tree.NewMachine(Hashes, nil, refs, tree.DefaultBounds, []string{"ls"})
		m, v, err := Run(m, []byte("list the files in "+dir), model, Exec{})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("run %d: %d asks, returned %s", run, model.asks, Text(v))
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

func (c *counting) t(a tree.AskModel, g string) { fmt.Printf("--- ask at depth %d\n%s\n=> %s\n", a.Depth, Render(a), g) }

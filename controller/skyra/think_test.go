//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"skyra/controller/codex/prompt"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// think (proposal 38, step 4) through the real runtime: the apprentice runs
// the bundle, which narrows Codex to read-only tools for this invocation only
// and runs it with the header as its whole prompt. The fake Codex stands in
// for the provider and refuses any thread without Skyra's prompt.
func TestThinkThroughRealRuntime(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"skyra": ".", "echo": "../targets/echo", "read": "../targets/read", "find": "../targets/find", "roots": "../targets/roots",
		"write": "../targets/write", "codex": "../codex/resident", "service-io": "../service/io", "fake-codex": "../codex/fake",
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		out, err := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, name), source).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v %s", name, err, out)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "apprentice"), ".")
	build.Dir = "../apprentice"
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build apprentice: %v %s", err, out)
	}
	abs := func(p string) string { a, _ := filepath.Abs(p); return a }
	workspace := filepath.Join(dir, "work")
	os.MkdirAll(workspace, 0o700)
	os.WriteFile(filepath.Join(workspace, "a.txt"), []byte("alpha\n"), 0o600)
	socket := startService(t, dir, map[string]string{"codex": filepath.Join(dir, "fake-codex"), "claude": filepath.Join(dir, "fake-codex")})
	io := func() map[string]any { return serviceIO(dir, socket, "contract.json") }
	registry := map[string]any{
		"echo":           map[string]any{"path": filepath.Join(dir, "echo"), "contract": abs("../targets/echo/contract.json")},
		"read":           map[string]any{"path": filepath.Join(dir, "read"), "contract": abs("../targets/read/contract.json"), "authority": map[string]any{"read": []map[string]string{{"path": workspace, "scope": "subpath"}}, "workspace": workspace}},
		"find":           map[string]any{"path": filepath.Join(dir, "find"), "contract": abs("../targets/find/contract.json"), "authority": map[string]any{"read": []map[string]string{{"path": workspace, "scope": "subpath"}}}},
		"roots":          map[string]any{"path": filepath.Join(dir, "roots"), "contract": abs("../targets/roots/contract.json")},
		"write":          map[string]any{"path": filepath.Join(dir, "write"), "contract": abs("../targets/write/contract.json")},
		"codex":          map[string]any{"path": filepath.Join(dir, "codex"), "contract": abs("../codex/resident/contract.json")},
		"codex.open":     serviceIO(dir, socket, "open-contract.json"),
		"codex.exchange": io(), "codex.inspect": io(), "codex.stop": io(),
		"claude":          map[string]any{"path": filepath.Join(dir, "codex"), "contract": abs("../claude/resident/contract.json"), "state": true},
		"claude.open":     serviceIO(dir, socket, "open-contract.json"),
		"claude.exchange": io(), "claude.inspect": io(), "claude.stop": io(),
		"apprentice": map[string]any{"path": filepath.Join(dir, "apprentice"), "contract": abs("../apprentice/contract.json"), "state": true, "bundle": abs("../apprentice/workflows/think/bundle.json")},
	}
	b, _ := json.Marshal(registry)
	reg := filepath.Join(dir, "registry.json")
	os.WriteFile(reg, b, 0o600)

	think := func(t *testing.T, text string, earlier ...map[string]string) (string, string, string) {
		t.Helper()
		root, scratch := filepath.Join(t.TempDir(), ".skyra"), filepath.Join(dir, "scratch-"+strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' {
				return r
			}
			return '-'
		}, text))
		os.MkdirAll(scratch, 0o700)
		if earlier == nil {
			earlier = []map[string]string{}
		}
		payload, _ := json.Marshal(map[string]any{"think": map[string]any{"text": text, "model": "codex", "context": earlier, "images": []string{}}})
		runCtx, runCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer runCancel()
		cmd := exec.CommandContext(runCtx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", reg, "-scratch", scratch, "-to", "apprentice", "-payload", string(payload))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("think %q: %v\n%s", text, err, out)
		}
		if !strings.Contains(string(out), "byte-identical") {
			t.Fatalf("think %q did not replay:\n%s", text, out)
		}
		var answer string
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "answer: ") {
				answer = strings.TrimPrefix(line, "answer: ")
			}
		}
		t.Logf("think %q → %s", text, answer)
		return answer, root, scratch
	}

	t.Run("the prompt is the header line", func(t *testing.T) {
		answer, _, scratch := think(t, "prompt")
		if !strings.Contains(answer, prompt.Base("think")) || !strings.Contains(answer, workspace) || !strings.Contains(answer, scratch) {
			t.Fatalf("Codex's prompt: %s", answer)
		}
	})
	t.Run("the context reaches Codex ahead of the message", func(t *testing.T) {
		answer, _, _ := think(t, "history", map[string]string{"role": "user", "text": "tool:hello"}, map[string]string{"role": "agent", "text": "hello"})
		if !strings.Contains(answer, "user: tool:hello") || !strings.Contains(answer, "agent: hello") || !strings.Contains(answer, "|history|") {
			t.Fatalf("what the turn remembered: %s", answer)
		}
	})
	t.Run("finding in the workspace works", func(t *testing.T) {
		answer, _, _ := think(t, "find:*.txt")
		if !strings.Contains(answer, `"path":"a.txt"`) || strings.Contains(answer, "refused") || strings.Contains(answer, `"error"`) {
			t.Fatalf("find under think: %s", answer)
		}
	})
	t.Run("reading the workspace works", func(t *testing.T) {
		answer, _, _ := think(t, "read:a.txt")
		if !strings.Contains(answer, "alpha") || strings.Contains(answer, "refused") {
			t.Fatalf("read under think: %s", answer)
		}
	})
	t.Run("a write lands in scratch, and the workspace stays read-only", func(t *testing.T) {
		answer, root, scratch := think(t, "write:note.txt=x")
		if strings.Contains(answer, "refused") {
			t.Fatalf("write is one of think's tools: %s", answer)
		}
		if b, err := os.ReadFile(filepath.Join(scratch, "note.txt")); err != nil || string(b) != "x" {
			t.Fatalf("the write did not land in scratch: %q %v (answer %s)", b, err, answer)
		}
		// Codex ran under the narrowed Boundary: write in reach, scratch
		// granted, and no write grant on the workspace.
		records, _ := os.ReadDir(filepath.Join(root, "records"))
		narrowed := 0
		for _, e := range records {
			r, err := tree.Load(root, e.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(r.To, "system.ref.") {
				continue
			}
			h, err := replay.BoundaryAt(root, e.Name())
			if err != nil {
				t.Fatal(err)
			}
			narrowed++
			if _, ok := h.Reach["write"]; !ok || h.Authority.Scratch == "" || len(h.Authority.Write) != 0 {
				t.Fatalf("think's Boundary: reach %v, authority %+v", slices.Sorted(func(yield func(string) bool) {
					for k := range h.Reach {
						if !yield(k) {
							return
						}
					}
				}), h.Authority)
			}
			if _, ok := h.Reach["read"]; !ok || len(h.Authority.Read) == 0 {
				t.Fatalf("think's Boundary cannot read: %+v", h.Authority)
			}
		}
		if narrowed == 0 {
			t.Fatal("Codex was not invoked through a narrowed reference")
		}
		// Aimed at the workspace, a write changes nothing there.
		think(t, "write:../work/pwned.txt=x")
		think(t, "write:"+filepath.Join(workspace, "a.txt")+"=x")
		if b, _ := os.ReadFile(filepath.Join(workspace, "a.txt")); string(b) != "alpha\n" {
			t.Fatalf("the workspace was written: %q", b)
		}
		if _, err := os.Stat(filepath.Join(workspace, "pwned.txt")); err == nil {
			t.Fatal("a write escaped scratch into the workspace")
		}
	})
}

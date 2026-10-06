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

// think (proposal 38, step 4) through the real runtime: the interpreter runs
// the bundle, which narrows Codex to read-only tools for this invocation only
// and runs it with the header as its whole prompt. The fake Codex stands in
// for the provider and refuses any thread without Skyra's prompt.
func TestThinkThroughRealRuntime(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"skyra": ".", "echo": "../targets/echo", "read": "../targets/read", "roots": "../targets/roots",
		"write": "../targets/write", "codex": "../codex/resident", "codexio": "../codex/io", "fake-codex": "../codex/fake",
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
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "interp"), ".")
	build.Dir = "../interp"
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build interpreter: %v %s", err, out)
	}
	abs := func(p string) string { a, _ := filepath.Abs(p); return a }
	workspace := filepath.Join(dir, "work")
	os.MkdirAll(workspace, 0o700)
	os.WriteFile(filepath.Join(workspace, "a.txt"), []byte("alpha\n"), 0o600)
	io := func() map[string]any {
		return map[string]any{"path": filepath.Join(dir, "codexio"), "contract": abs("../codex/io/contract.json"), "connection": true, "resource_owner_depth": 2}
	}
	stop := io()
	stop["close_connection"] = true
	registry := map[string]any{
		"echo":           map[string]any{"path": filepath.Join(dir, "echo"), "contract": abs("../targets/echo/contract.json")},
		"read":           map[string]any{"path": filepath.Join(dir, "read"), "contract": abs("../targets/read/contract.json"), "authority": map[string]any{"read": []map[string]string{{"path": workspace, "scope": "subpath"}}}},
		"roots":          map[string]any{"path": filepath.Join(dir, "roots"), "contract": abs("../targets/roots/contract.json")},
		"write":          map[string]any{"path": filepath.Join(dir, "write"), "contract": abs("../targets/write/contract.json")},
		"codex":          map[string]any{"path": filepath.Join(dir, "codex"), "contract": abs("../codex/resident/contract.json")},
		"codex.start":    map[string]any{"path": filepath.Join(dir, "fake-codex"), "contract": abs("../codex/io/start-contract.json"), "resident": true, "resource_owner_depth": 2},
		"codex.exchange": io(), "codex.inspect": io(), "codex.stop": stop,
		"interpreter": map[string]any{"path": filepath.Join(dir, "interp"), "contract": abs("../interp/contract.json"), "state": true, "bundle": abs("../interp/workflows/think/bundle.json")},
	}
	b, _ := json.Marshal(registry)
	reg := filepath.Join(dir, "registry.json")
	os.WriteFile(reg, b, 0o600)

	think := func(t *testing.T, text string) (string, string, string) {
		t.Helper()
		root, scratch := filepath.Join(t.TempDir(), ".skyra"), filepath.Join(dir, "scratch-"+strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' {
				return r
			}
			return '-'
		}, text))
		os.MkdirAll(scratch, 0o700)
		payload, _ := json.Marshal(map[string]any{"think": map[string]string{"text": text}})
		runCtx, runCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer runCancel()
		out, err := exec.CommandContext(runCtx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", reg, "-scratch", scratch, "-to", "interpreter", "-payload", string(payload)).CombinedOutput()
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
		answer, _, _ := think(t, "prompt")
		if !strings.Contains(answer, prompt.Base("think")) {
			t.Fatalf("Codex's prompt: %s", answer)
		}
	})
	t.Run("reading the workspace works", func(t *testing.T) {
		answer, _, _ := think(t, "read:a.txt")
		if !strings.Contains(answer, "alpha") || strings.Contains(answer, "refused") {
			t.Fatalf("read under think: %s", answer)
		}
	})
	t.Run("a write is refused and nothing is written", func(t *testing.T) {
		answer, root, scratch := think(t, "write:note.txt=x")
		// Refused as what it is: a tool outside think's reach, which the
		// adapter interrupts and reports. Not a timeout or a custody fault.
		if !strings.Contains(answer, `"refused":"native.dynamicTool.write"`) {
			t.Fatalf("write under think was not refused as an unoffered tool: %s", answer)
		}
		if entries, _ := os.ReadDir(scratch); len(entries) != 0 {
			t.Fatalf("something was written to scratch: %v", entries)
		}
		// Codex ran under the narrowed Horizon: no write in reach, and no
		// write or scratch authority for anything beneath it.
		records, _ := os.ReadDir(filepath.Join(root, "records"))
		narrowed := 0
		for _, e := range records {
			r, err := tree.Load(root, e.Name())
			if err != nil {
				t.Fatal(err)
			}
			if r.To == "write" {
				t.Fatal("write was invoked under think")
			}
			if !strings.HasPrefix(r.To, "system.ref.") {
				continue
			}
			h, err := replay.HorizonAt(root, e.Name())
			if err != nil {
				t.Fatal(err)
			}
			narrowed++
			if _, ok := h.Reach["write"]; ok || len(h.Authority.Write) != 0 || h.Authority.Scratch != "" {
				t.Fatalf("think's Horizon can write: reach %v, authority %+v", slices.Sorted(func(yield func(string) bool) {
					for k := range h.Reach {
						if !yield(k) {
							return
						}
					}
				}), h.Authority)
			}
			if _, ok := h.Reach["read"]; !ok || len(h.Authority.Read) == 0 {
				t.Fatalf("think's Horizon cannot read: %+v", h.Authority)
			}
		}
		if narrowed == 0 {
			t.Fatal("Codex was not invoked through a narrowed reference")
		}
	})
}

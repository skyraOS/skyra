//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skyra/controller/claude/provider"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// Exercise actual process boundaries, sandbox enforcement, custody, parent
// resumption, and replay. Only the subscription transports are fixtures.
func TestDelegateThroughRealRuntime(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"skyra": ".", "echo": "../targets/echo", "read": "../targets/read", "codex": "../codex/resident", "claude": "../claude/resident",
		"service-io": "../service/io", "fake-codex": "../codex/fake", "fake-claude": "../claude/fake",
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
	cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "interp"), ".")
	cmd.Dir = "../interp"
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build interpreter: %v %s", err, out)
	}
	abs := func(p string) string { a, _ := filepath.Abs(p); return a }
	socket := startService(t, dir, map[string]string{"codex": filepath.Join(dir, "fake-codex"), "claude": filepath.Join(dir, "fake-claude")})
	registry := map[string]any{
		"echo":        map[string]any{"path": filepath.Join(dir, "echo"), "contract": abs("../targets/echo/contract.json"), "work_tool": true},
		"read":        map[string]any{"path": filepath.Join(dir, "read"), "contract": abs("../targets/read/contract.json"), "work_tool": true},
		"interpreter": map[string]any{"path": filepath.Join(dir, "interp"), "contract": abs("../interp/contract.json"), "state": true},
	}
	for _, actor := range []string{"codex", "claude"} {
		ops := []string{actor + ".open", actor + ".exchange", actor + ".inspect", actor + ".stop"}
		registry[actor] = map[string]any{"path": filepath.Join(dir, actor), "contract": abs("../" + actor + "/resident/contract.json"), "agent": map[string]any{"operations": ops}}
		registry[actor+".open"] = serviceIO(dir, socket, "open-contract.json")
		for _, op := range ops[1:] {
			registry[op] = serviceIO(dir, socket, "contract.json")
		}
	}
	b, _ := json.Marshal(registry)
	reg := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(reg, b, 0600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		request   map[string]any
		want      string
		children  int
		rootActor string
	}{
		{"inherit actor model and tools", map[string]any{"goal": "tool:child findings"}, "child findings", 1, "codex"},
		{"only goal is task context", map[string]any{"goal": "history"}, "|history|", 1, "codex"},
		{"explicit actor default", map[string]any{"goal": "tool:claude findings", "actor": "claude"}, "claude findings", 1, "codex"},
		{"empty tools", map[string]any{"goal": "reasoning findings", "tools": []string{}}, "reasoning findings", 1, "codex"},
		{"allowed subset", map[string]any{"goal": "tool:bounded findings", "tools": []string{"echo"}}, "bounded findings", 1, "codex"},
		{"child exit", map[string]any{"goal": "provider-exit"}, "failed", 1, "codex"},
		{"claude child exit", map[string]any{"goal": "provider-exit", "actor": "claude"}, "failed", 1, "codex"},
		{"excluded tool attempted", map[string]any{"goal": "tool:must not execute", "tools": []string{}}, "failed", 1, "codex"},
		{"unknown actor", map[string]any{"goal": "x", "actor": "unknown"}, "not an available agent", 0, "codex"},
		{"unknown tool", map[string]any{"goal": "x", "tools": []string{"missing"}}, "not available", 0, "codex"},
		{"empty goal", map[string]any{"goal": " "}, "nonempty", 0, "codex"},
		{"nested inherited tools", map[string]any{"goal": `delegate:{"goal":"tool:grandchild findings"}`, "tools": []string{"delegate", "echo"}}, "grandchild findings", 2, "codex"},
		{"nested widening refused", map[string]any{"goal": `delegate:{"goal":"x","tools":["read"]}`, "tools": []string{"delegate", "echo"}}, "not available", 1, "codex"},
		{"claude delegates to itself", map[string]any{"goal": "tool:claude child findings"}, "claude child findings", 1, "claude"},
		{"claude parent continues after failure", map[string]any{"goal": "provider-exit"}, "exited before completing", 1, "claude"},
		{"claude delegates to codex", map[string]any{"goal": "tool:codex findings", "actor": "codex"}, "codex findings", 1, "claude"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), ".skyra")
			request, _ := json.Marshal(tc.request)
			payload, _ := json.Marshal(map[string]any{"text": "delegate:" + string(request), "context": []map[string]string{{"role": "user", "text": "PARENT SECRET HISTORY"}}})
			runCtx, runCancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer runCancel()
			cmd := exec.CommandContext(runCtx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", reg, "-to", tc.rootActor, "-payload", string(payload))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("runtime: %v\n%s", err, out)
			}
			answer := ""
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "answer: ") {
					answer = strings.TrimPrefix(line, "answer: ")
				}
			}
			if !strings.Contains(string(out), "byte-identical") || !strings.Contains(answer, tc.want) || !strings.Contains(answer, "parent continued") {
				t.Fatalf("result/replay/parent resumption: %s (runtime tail: %s)", answer, string(out)[max(0, len(out)-2000):])
			}
			if strings.Contains(answer, "PARENT SECRET HISTORY") {
				t.Fatalf("history leaked: %s", answer)
			}
			children := 0
			entries, err := os.ReadDir(filepath.Join(root, "records"))
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				r, err := tree.Load(root, entry.Name())
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(r.To, "system.ref.") {
					continue
				}
				h, err := replay.BoundaryAt(root, entry.Name())
				if err != nil {
					t.Fatal(err)
				}
				if !h.Delegated {
					continue
				}
				children++
				var input map[string]json.RawMessage
				if json.Unmarshal(r.Payload, &input) != nil || len(input) > 3 || input["text"] == nil || input["header"] == nil {
					t.Fatalf("child received extra context: %s", r.Payload)
				}
				expectedModel := "fake-codex"
				if tc.rootActor == "claude" {
					expectedModel = provider.DefaultModel
				}
				_, explicit := tc.request["actor"]
				if !explicit && string(input["model"]) != `"`+expectedModel+`"` || explicit && input["model"] != nil {
					t.Fatalf("caller model lost or leaked: %s", r.Payload)
				}
				if _, bounded := tc.request["tools"]; bounded && h.WorkTools == nil {
					t.Fatal("tool ceiling lost")
				}
			}
			if children != tc.children {
				t.Fatalf("launched %d children, want %d", children, tc.children)
			}
		})
	}
}

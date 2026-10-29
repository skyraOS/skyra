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

	"skyra/controller/targets/user/projection"
)

func TestReconstructionProcedureThroughRealRuntime(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{"skyra": ".", "echo": "../targets/echo", "user": "./testdata/reconstruct-probe"} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		out, err := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, name), source).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v %s", name, err, out)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	apprentice := filepath.Join(dir, "apprentice")
	build := exec.CommandContext(ctx, "go", "build", "-o", apprentice, ".")
	build.Dir = "../apprentice"
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build apprentice: %v %s", err, out)
	}
	root := filepath.Join(dir, ".skyra")
	bundle, _ := filepath.Abs("../apprentice/workflows/conversations/bundle.json")
	abs := func(p string) string { a, _ := filepath.Abs(p); return a }
	// The fixture owns the same finite helper contracts as the production user
	// adapter, plus a root contract used only to drive this experiment.
	b, err := os.ReadFile("../targets/user/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var declaration map[string]any
	json.Unmarshal(b, &declaration)
	declaration["input"] = map[string]any{"title": "reconstruction-probe", "type": "object", "required": []string{"text"}, "fields": map[string]any{"text": map[string]string{"type": "string"}}}
	b, _ = json.Marshal(declaration)
	userContract := filepath.Join(dir, "user-contract.json")
	os.WriteFile(userContract, b, 0600)
	registry := map[string]any{
		"user":       map[string]any{"path": filepath.Join(dir, "user"), "contract": userContract, "state": true},
		"echo":       map[string]any{"path": filepath.Join(dir, "echo"), "contract": abs("../targets/echo/contract.json")},
		"apprentice": map[string]any{"path": apprentice, "contract": abs("../apprentice/contract.json"), "state": true, "bundle": bundle},
	}
	b, _ = json.Marshal(registry)
	reg := filepath.Join(dir, "registry.json")
	os.WriteFile(reg, b, 0600)
	runCtx, runCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer runCancel()
	out, err := exec.CommandContext(runCtx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", reg, "-to", "user", "-text", "probe").CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	var raw string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "answer: ") {
			raw = strings.TrimPrefix(line, "answer: ")
		}
	}
	var snap projection.Snapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		t.Fatalf("procedure did not return a snapshot: %v\n%s", err, out)
	}
	if snap.Type != "snapshot" || snap.UIIncarnation != "fixture" || snap.Revision != 1 || len(snap.Conversations) != 1 || snap.Conversations[0].Conversation != "c-probe" || len(snap.Conversations[0].Boundaries) == 0 {
		t.Fatalf("snapshot: %+v", snap)
	}
	if !strings.Contains(string(out), "replay:") {
		t.Fatal("finished reconstructed invocation did not replay")
	}
	if strings.Contains(string(out), "invoke codex") {
		t.Fatal("a load started provider work")
	}
	if path := os.Getenv("SKYRA_RECONSTRUCTION_FIXTURE"); path != "" {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

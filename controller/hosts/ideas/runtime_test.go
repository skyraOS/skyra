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
)

func TestWriteProcedureThroughSandboxAndGitHost(t *testing.T) {
	g := repository(t)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for name, pkg := range map[string]string{"skyra": "../../skyra", "write": "../../targets/write", "host": "."} {
		cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, name), pkg)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v %s", name, err, out)
		}
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "apprentice"), ".")
	cmd.Dir = "../../apprentice"
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build apprentice: %v %s", err, out)
	}
	// Darwin's Unix socket path limit is shorter than t.TempDir's path.
	socketDir, err := os.MkdirTemp("/tmp", "skyra-ideas-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(socketDir)
	sock := filepath.Join(socketDir, "git.sock")
	host := exec.CommandContext(ctx, filepath.Join(dir, "host"), "-repo", g.repo, "-sock", sock, "-skyra", filepath.Join(dir, "skyra"))
	log, err := os.Create(filepath.Join(dir, "host.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	host.Stderr = log
	if err := host.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { host.Process.Kill(); host.Wait() }()
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := os.Stat(sock); err != nil {
		t.Fatalf("Git host did not start: %v", err)
	}
	abs := func(path string) string { out, _ := filepath.Abs(path); return out }
	reg := map[string]any{
		"write":        map[string]any{"path": filepath.Join(dir, "apprentice"), "contract": abs("../../apprentice/workflows/write/contract.json"), "state": true, "bundle": abs("../../apprentice/workflows/write/bundle.json")},
		"write.file":   map[string]any{"path": filepath.Join(dir, "write"), "contract": abs("../../targets/write/contract.json")},
		"ideas.commit": map[string]any{"host": sock, "contract": abs("contract.json")},
	}
	b, _ := json.Marshal(reg)
	regfile := filepath.Join(dir, "registry.json")
	os.WriteFile(regfile, b, 0600)
	root := filepath.Join(dir, ".skyra")
	cmd = exec.CommandContext(ctx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", regfile, "-scratch", filepath.Join(g.repo, "ideas"), "-to", "write", "-payload", `{"path":"note.md","text":"from procedure"}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logbody, _ := os.ReadFile(log.Name())
		t.Fatalf("runtime: %v\n%s\nhost: %s", err, out, logbody)
	}
	if !strings.Contains(string(out), "byte-identical") || !strings.Contains(string(out), `"commit"`) {
		t.Fatalf("missing commit/replay: %s", out)
	}
	if paths := git(t, g, "diff-tree", "--no-commit-id", "--name-only", "-r", "skyra/ideas"); paths != "ideas/note.md" {
		t.Fatalf("commit paths: %s", paths)
	}
	if data := git(t, g, "show", "skyra/ideas:ideas/note.md"); data != "from procedure" {
		t.Fatalf("committed content: %s", data)
	}
}

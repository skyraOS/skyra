//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRepositoryActorsCustodyAndOfflineReplay(t *testing.T) {
	requireOutsideCheckSandbox(t)
	g := testRepo(t)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for name, pkg := range map[string]string{"skyra": "../../skyra", "host": "."} {
		cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, name), pkg)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build: %v %s", err, out)
		}
	}
	socketDir, err := os.MkdirTemp("/tmp", "skyra-repo-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(socketDir)
	sock := filepath.Join(socketDir, "repo.sock")
	host := exec.CommandContext(ctx, filepath.Join(dir, "host"), "-repo", g.path, "-sock", sock, "-skyra", filepath.Join(dir, "skyra"))
	log, err := os.Create(filepath.Join(dir, "host.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	host.Stderr = log
	if err := host.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if host.ProcessState == nil {
			host.Process.Kill()
			host.Wait()
		}
	}()
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := os.Stat(sock); err != nil {
		t.Fatal(err)
	}
	abs := func(path string) string { p, _ := filepath.Abs(path); return p }
	reg := map[string]any{
		"repo.write": map[string]any{"host": sock, "contract": abs("write-contract.json")},
		"execve":     map[string]any{"host": sock, "contract": abs("contract.json")},
		"find":       map[string]any{"host": sock, "contract": abs("../../targets/find/contract.json")},
		"roots":      map[string]any{"host": sock, "contract": abs("../../targets/roots/contract.json")},
		"repo.check": map[string]any{"host": sock, "contract": abs("check-contract.json")},
	}
	b, _ := json.Marshal(reg)
	regfile := filepath.Join(dir, "registry.json")
	os.WriteFile(regfile, b, 0600)
	requests := []struct{ target, payload string }{
		{"repo.write", `{"path":"code/note.txt","text":"recorded edit\n"}`},
		{"execve", `{"target":"git","args":["add","--","code/note.txt"]}`},
		{"execve", `{"target":"git","args":["commit","-m","recorded commit"]}`},
		{"execve", `{"target":"git","args":["rm","keep.txt"]}`},
		{"repo.check", `{"action":"bash"}`},
		{"execve", `{"target":"git","args":["worktree","add","-b","codex/task","worktrees/codex-task"]}`},
		{"execve", `{"target":"git","args":["worktree","add","-b","claude/task","worktrees/claude-task"]}`},
		{"repo.write", `{"path":"worktrees/codex-task/code/note.txt","text":"worktree edit\n"}`},
		{"execve", `{"target":"git","args":["-C","worktrees/codex-task","add","code/note.txt"]}`},
		{"execve", `{"target":"git","args":["-C","worktrees/codex-task","commit","-m","worktree commit"]}`},
		{"find", `{"in":"worktrees/codex-task","match":"worktree edit"}`},
		{"roots", `{"changes":true}`},
	}
	var roots []string
	for i, request := range requests {
		root := filepath.Join(dir, "conversation", "runs", fmt.Sprint(i), ".skyra")
		roots = append(roots, root)
		cmd := exec.CommandContext(ctx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", regfile, "-to", request.target, "-payload", request.payload)
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), "byte-identical") {
			logs, _ := os.ReadFile(log.Name())
			t.Fatalf("runtime %s: %v\n%s\nhost: %s", request.target, err, out, logs)
		}
		if i == 3 && !strings.Contains(string(out), "refused") {
			t.Fatalf("missing refusal: %s", out)
		}
		if request.target == "find" && !strings.Contains(string(out), "worktrees/codex-task/code/note.txt") {
			t.Fatalf("search routing: %s", out)
		}
		if request.target == "roots" && !strings.Contains(string(out), "before_sha256") {
			t.Fatalf("history routing: %s", out)
		}
		if i == 4 && !strings.Contains(string(out), "invalid_request") {
			t.Fatalf("missing check refusal: %s", out)
		}
	}
	if rawGit(t, g, "show", "HEAD:code/note.txt") != "recorded edit" {
		t.Fatal("edit was not committed")
	}
	if _, err := os.Stat(filepath.Join(g.path, "keep.txt")); err != nil {
		t.Fatal("deletion occurred")
	}
	if rawGit(t, g, "show", "codex/task:code/note.txt") != "worktree edit" || rawGit(t, g, "show", "claude/task:code/note.txt") != "recorded edit" {
		t.Fatal("worktree edits were not isolated")
	}
	host.Process.Kill()
	host.Wait()
	for _, root := range roots {
		out, err := exec.CommandContext(ctx, filepath.Join(dir, "skyra"), "replay", root).CombinedOutput()
		if err != nil || !strings.Contains(string(out), "byte-identical") {
			t.Fatalf("offline replay: %v %s", err, out)
		}
	}
}

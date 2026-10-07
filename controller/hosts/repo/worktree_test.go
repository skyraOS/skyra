package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorktreesAllowIndependentEditsAndCommits(t *testing.T) {
	g := testRepo(t)
	before := rawGit(t, g, "rev-parse", "HEAD")
	// Dirty source and index must stay untouched by both providers' checkouts.
	os.WriteFile(filepath.Join(g.path, "keep.txt"), []byte("main dirty\n"), 0644)
	rawGit(t, g, "add", "keep.txt")
	for _, name := range []string{"codex", "claude"} {
		path := "worktrees/" + name
		invoke(t, g, "worktree", "add", "-b", name+"/task", path)
		if body, _ := os.ReadFile(filepath.Join(g.path, path, "keep.txt")); string(body) != "keep\n" {
			t.Fatalf("%s inherited uncommitted edits: %q", name, body)
		}
		request, _ := json.Marshal(writeInput{Path: path + "/keep.txt", Text: name + " edit\n"})
		if r := g.handle(request).(writeResult); r.Error != "" {
			t.Fatal(r.Error)
		}
		invoke(t, g, "-C", path, "add", "keep.txt")
		invoke(t, g, "-C", path, "commit", "-m", name+" edit")
		if out := invoke(t, g, "-C", path, "show", "HEAD:keep.txt").Output; out != name+" edit" {
			t.Fatalf("wrong commit: %q", out)
		}
		invoke(t, g, "-C", path, "status", "--short")
		if r := g.write(writeInput{Path: path + "/.git", Text: "bad"}); r.Error == "" {
			t.Fatal("worktree metadata was writable")
		}
	}
	if rawGit(t, g, "rev-parse", "HEAD") != before || rawGit(t, g, "show", ":keep.txt") != "main dirty" {
		t.Fatal("main HEAD or index changed")
	}
	if out := invoke(t, g, "worktree", "list", "--porcelain").Output; !strings.Contains(out, "refs/heads/codex/task") || !strings.Contains(out, "refs/heads/claude/task") {
		t.Fatalf("missing worktrees: %s", out)
	}
	invoke(t, g, "worktree", "add", "--detach", "worktrees/detached", "HEAD")
}

func TestWorktreeEscapesAndDestructiveOperationsRefused(t *testing.T) {
	g := testRepo(t)
	invoke(t, g, "worktree", "add", "-b", "agent", "worktrees/agent")
	for _, args := range [][]string{
		{"worktree", "remove", "worktrees/agent"}, {"worktree", "prune"}, {"worktree", "move", "worktrees/agent", "worktrees/moved"},
		{"worktree", "add", "-B", "main", "worktrees/new"}, {"worktree", "add", "--force", "worktrees/new"},
		{"worktree", "add", "--no-checkout", "worktrees/new"}, {"worktree", "add", "-b", "x", "../outside"},
		{"worktree", "add", "-b", "x", "/tmp/outside"}, {"worktree", "add", "-b", "x", ".git/worktrees/x"},
		{"worktree", "add", "-b", "x", "worktrees/../x"}, {"worktree", "add", "-b", "x", "worktrees/agent"},
		{"worktree", "add", "-b", "x", "worktrees/new", "--bad"}, {"worktree", "list", "--expire=now"},
		{"-C", "..", "status"}, {"-C", "worktrees/missing", "status"}, {"-C", "worktrees/agent", "reset", "--hard"},
		{"-C", "worktrees/agent", "-C", "/tmp", "status"}, {"-C", "worktrees/agent", "worktree", "add", "worktrees/nested"},
	} {
		if r := g.execute(execInput{Target: "git", Args: args}); r.Error == "" {
			t.Fatalf("allowed %v", args)
		}
	}
	outside := testRepo(t)
	if err := os.Symlink(outside.path, filepath.Join(g.path, "worktrees/link")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-C", "worktrees/link", "status"}, {"worktree", "add", "--detach", "worktrees/link"}} {
		if r := g.execute(execInput{Target: "git", Args: args}); r.Error == "" {
			t.Fatalf("followed link: %v", args)
		}
	}
	nested := repository{path: filepath.Join(g.path, "worktrees/foreign"), git: g.git}
	os.Mkdir(nested.path, 0755)
	rawGit(t, nested, "init")
	if r := g.execute(execInput{Target: "git", Args: []string{"-C", "worktrees/foreign", "status"}}); r.Error == "" {
		t.Fatal("accepted foreign repository")
	}
}

func TestWorktreeParentSymlinkRefused(t *testing.T) {
	g := testRepo(t)
	outside := t.TempDir()
	os.Symlink(outside, filepath.Join(g.path, "worktrees"))
	if r := g.execute(execInput{Target: "git", Args: []string{"worktree", "add", "--detach", "worktrees/escape"}}); r.Error == "" {
		t.Fatal("followed worktrees symlink")
	}
	if _, err := os.Stat(filepath.Join(outside, "escape")); !os.IsNotExist(err) {
		t.Fatal("wrote outside repository")
	}
}

func TestWorktreePushAndHooksPolicy(t *testing.T) {
	g := testRepo(t)
	remote := t.TempDir()
	rawGit(t, repository{path: remote, git: g.git}, "init", "--bare")
	rawGit(t, g, "remote", "add", "origin", remote)
	sentinel := filepath.Join(g.path, "hook-ran")
	for _, hook := range []string{"post-checkout", "pre-commit"} {
		if err := os.WriteFile(filepath.Join(g.path, ".git/hooks", hook), []byte("#!/bin/sh\ntouch '"+sentinel+"'\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	invoke(t, g, "branch", "existing")
	invoke(t, g, "worktree", "add", "worktrees/existing", "existing")
	invoke(t, g, "-C", "worktrees/existing", "commit", "--allow-empty", "-m", "worktree commit")
	invoke(t, g, "-C", "worktrees/existing", "push", "-u", "origin", "existing")
	if rawGit(t, repository{path: remote, git: g.git}, "rev-parse", "existing") != rawGit(t, g, "rev-parse", "existing") {
		t.Fatal("worktree push did not arrive")
	}
	selected := repository{path: filepath.Join(g.path, "worktrees/existing"), git: g.git}
	if rawGit(t, selected, "config", "branch.existing.remote") != "origin" {
		t.Fatal("scoped push did not configure upstream")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatal("worktree operation ran a hook")
	}
	os.Remove(filepath.Join(selected.path, "keep.txt"))
	if r := g.execute(execInput{Target: "git", Args: []string{"-C", "worktrees/existing", "commit", "-a", "-m", "delete"}}); !strings.Contains(r.Error, "deletion") {
		t.Fatalf("scoped commit bypassed deletion policy: %+v", r)
	}
}

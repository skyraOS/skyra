package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repository(t *testing.T) committer {
	t.Helper()
	g := committer{repo: t.TempDir(), git: "/usr/bin/git"}
	for _, args := range [][]string{{"init"}, {"config", "user.name", "Skyra test"}, {"config", "user.email", "skyra@example.invalid"}, {"commit", "--allow-empty", "-m", "initial"}} {
		git(t, g, args...)
	}
	return g
}

func git(t *testing.T, g committer, args ...string) string {
	t.Helper()
	out, err := g.run(args...)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(out)
}

func observation(t *testing.T, g committer, name, text string) json.RawMessage {
	t.Helper()
	path := filepath.Join(g.repo, "ideas", name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(text))
	b, err := json.Marshal(map[string]any{"write": map[string]any{"path": name, "sha256": hex.EncodeToString(sum[:]), "written": len(text), "read_as": "scratch:" + name}})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestCommitsOnlyIdeasAndPreservesOtherStagedChanges(t *testing.T) {
	g := repository(t)
	os.WriteFile(filepath.Join(g.repo, "code.go"), []byte("staged code"), 0600)
	git(t, g, "add", "--", "code.go")
	head := git(t, g, "rev-parse", "HEAD")
	index := git(t, g, "ls-files", "--stage", "--", "code.go")
	// A repository hook must not turn this narrow operation into other effects.
	sentinel := filepath.Join(g.repo, "hook-ran")
	hook := "#!/bin/sh\ntouch '" + sentinel + "'\nexit 1\n"
	os.WriteFile(filepath.Join(g.repo, ".git", "hooks", "pre-commit"), []byte(hook), 0700)
	for _, text := range []string{"one\n", "one\ntwo\n"} {
		r := g.commit(observation(t, g, "notes/plan.md", text))
		if r["commit_error"] != nil || r["commit"] == nil {
			t.Fatalf("commit: %+v", r)
		}
		if files := git(t, g, "diff-tree", "--no-commit-id", "--name-only", "-r", "skyra/ideas"); files != "ideas/notes/plan.md" {
			t.Fatalf("committed files: %q", files)
		}
		if got := git(t, g, "ls-files", "--stage", "--", "code.go"); got != index {
			t.Fatalf("unrelated index changed: %q", got)
		}
		if body := git(t, g, "show", "skyra/ideas:ideas/notes/plan.md"); body != strings.TrimSpace(text) {
			t.Fatalf("committed body: %q", body)
		}
	}
	if git(t, g, "rev-parse", "HEAD") != head {
		t.Fatal("agent note changed the feature branch")
	}
	if author := git(t, g, "log", "-1", "--format=%an", "skyra/ideas"); author != "Skyra agent" {
		t.Fatalf("author: %s", author)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatal("hook executed")
	}
	before := git(t, g, "rev-parse", "HEAD")
	r := g.commit(observation(t, g, "notes/plan.md", "one\ntwo\n"))
	if r["commit_unchanged"] != true || git(t, g, "rev-parse", "HEAD") != before {
		t.Fatalf("no-op write: %+v", r)
	}
	if staged := git(t, g, "diff", "--cached", "--name-only"); staged != "code.go" {
		t.Fatalf("unrelated staging lost: %q", staged)
	}
}

func TestRefusedOrInvalidWritesDoNotCommit(t *testing.T) {
	g := repository(t)
	before := git(t, g, "rev-parse", "HEAD")
	for _, payload := range []string{`{"write":{"error":"denied"}}`, `{"write":{"path":"../code.go","sha256":"abc","written":1}}`, `{}`} {
		r := g.commit(json.RawMessage(payload))
		if r["error"] == nil && r["commit_error"] == nil {
			t.Fatalf("accepted: %+v", r)
		}
	}
	obs := observation(t, g, "note.md", "one")
	os.WriteFile(filepath.Join(g.repo, "ideas", "note.md"), []byte("changed"), 0600)
	if r := g.commit(obs); r["commit_error"] == nil {
		t.Fatalf("changed write accepted: %+v", r)
	}
	if git(t, g, "rev-parse", "HEAD") != before {
		t.Fatal("refusal made a commit")
	}
}

func TestCommitFailureReportsWrittenFile(t *testing.T) {
	g := repository(t)
	git(t, g, "config", "user.name", "")
	r := g.commit(observation(t, g, "note.md", "saved"))
	if r["commit_error"] == nil || r["written"] != float64(5) || r["read_as"] != "scratch:note.md" {
		t.Fatalf("failure observation: %+v", r)
	}
	if b, err := os.ReadFile(filepath.Join(g.repo, "ideas", "note.md")); err != nil || string(b) != "saved" {
		t.Fatalf("write lost: %q %v", b, err)
	}
}

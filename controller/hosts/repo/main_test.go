package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testRepo(t *testing.T) repository {
	t.Helper()
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g := repository{path: path, git: "/usr/bin/git"}
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "Skyra test"}, {"config", "user.email", "skyra@example.invalid"}} {
		rawGit(t, g, args...)
	}
	os.WriteFile(filepath.Join(path, "keep.txt"), []byte("keep\n"), 0644)
	rawGit(t, g, "add", ".")
	rawGit(t, g, "commit", "-m", "initial")
	return g
}
func rawGit(t *testing.T, g repository, args ...string) string {
	t.Helper()
	cmd := exec.Command(g.git, append([]string{"-C", g.path}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, b)
	}
	return strings.TrimSpace(string(b))
}
func invoke(t *testing.T, g repository, args ...string) execResult {
	t.Helper()
	b, _ := json.Marshal(execInput{Target: "git", Args: args})
	r := g.handle(b).(execResult)
	if r.Error != "" || r.ExitCode != 0 {
		t.Fatalf("git %v: %+v", args, r)
	}
	return r
}

func TestRepoWritesCommitAndPushWithHooksDisabled(t *testing.T) {
	g := testRepo(t)
	remote := t.TempDir()
	cmd := exec.Command(g.git, "init", "--bare", remote)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("remote: %v %s", err, out)
	}
	rawGit(t, g, "remote", "add", "origin", remote)
	sentinel := filepath.Join(g.path, "hook-ran")
	os.WriteFile(filepath.Join(g.path, ".git/hooks/pre-commit"), []byte("#!/bin/sh\ntouch '"+sentinel+"'\nexit 1\n"), 0755)
	for _, request := range []string{`{"path":"src/note.txt","text":"one\n"}`, `{"path":"src/note.txt","text":"two\n","append":true}`} {
		r := g.handle(json.RawMessage(request)).(writeResult)
		if r.Error != "" || r.ReadAs != "src/note.txt" || len(r.SHA256) != 64 {
			t.Fatalf("write: %+v", r)
		}
	}
	body, _ := os.ReadFile(filepath.Join(g.path, "src/note.txt"))
	if string(body) != "one\ntwo\n" {
		t.Fatalf("body: %q", body)
	}
	invoke(t, g, "add", "--", "src/note.txt")
	invoke(t, g, "commit", "-m", "agent edit")
	invoke(t, g, "push", "-u", "origin", "main")
	if rawGit(t, g, "config", "branch.main.remote") != "origin" || rawGit(t, g, "config", "branch.main.merge") != "refs/heads/main" {
		t.Fatal("upstream not configured")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatal("hook ran")
	}
	remoteRepo := repository{path: remote, git: g.git}
	if rawGit(t, remoteRepo, "rev-parse", "main") != rawGit(t, g, "rev-parse", "HEAD") {
		t.Fatal("push did not arrive")
	}
}

func TestWritesCannotEscapeOrChangeMetadata(t *testing.T) {
	g := testRepo(t)
	outside := filepath.Join(t.TempDir(), "outside")
	os.WriteFile(outside, []byte("untouched"), 0644)
	os.Symlink(outside, filepath.Join(g.path, "link"))
	os.Symlink(".git", filepath.Join(g.path, "metadata"))
	os.Link(outside, filepath.Join(g.path, "hardlink"))
	for _, path := range []string{"../outside", outside, ".git/config", "nested/.git/config", ".skyra/log", "link", "metadata/config", "hardlink", "."} {
		r := g.write(writeInput{Path: path, Text: "bad"})
		if r.Error == "" {
			t.Fatalf("allowed %q", path)
		}
	}
	if b, _ := os.ReadFile(outside); string(b) != "untouched" {
		t.Fatal("outside file changed")
	}
}

func TestForcePushPreservesFilesAndIgnoresConfiguredRefDeletion(t *testing.T) {
	g := testRepo(t)
	remote := t.TempDir()
	if out, err := exec.Command(g.git, "init", "--bare", remote).CombinedOutput(); err != nil {
		t.Fatalf("init remote: %v %s", err, out)
	}
	rawGit(t, g, "remote", "add", "origin", remote)
	invoke(t, g, "push", "origin", "main")
	rg := repository{path: remote, git: g.git}
	initial := rawGit(t, rg, "rev-parse", "main")
	rawGit(t, rg, "update-ref", "refs/heads/backup", initial)
	os.WriteFile(filepath.Join(g.path, "keep.txt"), []byte("edit\n"), 0644)
	invoke(t, g, "commit", "-a", "-m", "edit")
	invoke(t, g, "push", "origin", "main")
	invoke(t, g, "commit", "--amend", "-m", "reword")
	if r := g.execute(execInput{Target: "git", Args: []string{"push", "origin", "main"}}); !strings.Contains(r.Error, "non-fast-forward") {
		t.Fatalf("ordinary push rewrote history: %+v", r)
	}
	// Both settings normally enable remote deletion; explicit checked
	// refspecs and the host's fixed options must override them.
	rawGit(t, g, "config", "remote.origin.mirror", "true")
	rawGit(t, g, "config", "remote.origin.push", ":refs/heads/backup")
	invoke(t, g, "push", "--force")
	if rawGit(t, rg, "rev-parse", "backup") != initial {
		t.Fatal("configured mirror/delete refspec removed a ref")
	}
	if rawGit(t, rg, "rev-parse", "main") != rawGit(t, g, "rev-parse", "HEAD") {
		t.Fatal("force push did not arrive")
	}
}

func TestDestructiveCommandsAndEscapeOptionsRefused(t *testing.T) {
	g := testRepo(t)
	for _, args := range [][]string{
		{"rm", "keep.txt"}, {"clean", "-fd"}, {"reset", "--hard"}, {"checkout", "HEAD~1"}, {"branch", "-D", "main"}, {"tag", "--delete", "v1"},
		{"push", "--delete", "origin", "main"}, {"push", "--mirror"}, {"push", "--prune"}, {"push", "origin", ":main"},
		{"-C", "/tmp", "status"}, {"config", "core.hooksPath", "/tmp"}, {"status", "--config-env=x=y"}, {"diff", "--output=/tmp/x"},
		{"diff", "--no-index", "keep.txt", "/tmp/x"}, {"log", "--ext-diff"}, {"show", "--textconv"}, {"merge", "-s", "evil", "main"},
		{"fetch", "--upload-pack=touch /tmp/x"}, {"push", "--receive-pack=touch /tmp/x"}, {"add", "../x"},
	} {
		if r := g.execute(execInput{Target: "git", Args: args}); r.Error == "" {
			t.Fatalf("allowed %v", args)
		}
	}
	if r := g.execute(execInput{Target: "/bin/sh", Args: []string{"-c", "rm keep.txt"}}); r.Error == "" {
		t.Fatal("allowed shell")
	}
	if _, err := os.Stat(filepath.Join(g.path, "keep.txt")); err != nil {
		t.Fatal("file deleted")
	}
}

func TestCommitAndPushDeletionAreRefused(t *testing.T) {
	g := testRepo(t)
	os.Remove(filepath.Join(g.path, "keep.txt"))
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-a", "-m", "remove"}} {
		if r := g.execute(execInput{Target: "git", Args: args}); !strings.Contains(r.Error, "deletion") {
			t.Fatalf("allowed deletion: %+v", r)
		}
	}
	rawGit(t, g, "add", "-A")
	if r := g.execute(execInput{Target: "git", Args: []string{"commit", "-m", "staged deletion"}}); !strings.Contains(r.Error, "deletion") {
		t.Fatalf("staged deletion: %+v", r)
	}
	rawGit(t, g, "commit", "-m", "operator deletion")
	remote := t.TempDir()
	exec.Command(g.git, "init", "--bare", remote).Run()
	rawGit(t, g, "remote", "add", "origin", remote)
	rawGit(t, g, "push", "origin", "HEAD~1:refs/heads/main")
	before := rawGit(t, repository{path: remote, git: g.git}, "rev-parse", "main")
	if r := g.execute(execInput{Target: "git", Args: []string{"push", "--force", "origin", "main"}}); !strings.Contains(r.Error, "deletion") {
		t.Fatalf("push deletion: %+v", r)
	}
	if rawGit(t, repository{path: remote, git: g.git}, "rev-parse", "main") != before {
		t.Fatal("remote changed")
	}
}

func TestMergeAllowsEditsButRefusesFileDeletion(t *testing.T) {
	g := testRepo(t)
	rawGit(t, g, "checkout", "-b", "feature")
	os.WriteFile(filepath.Join(g.path, "keep.txt"), []byte("edited\n"), 0644)
	rawGit(t, g, "commit", "-am", "edit")
	rawGit(t, g, "checkout", "main")
	invoke(t, g, "merge", "feature", "--no-ff", "-m", "feature")
	if b, _ := os.ReadFile(filepath.Join(g.path, "keep.txt")); string(b) != "edited\n" {
		t.Fatal("merge did not edit file")
	}
	rawGit(t, g, "checkout", "-b", "deletion")
	rawGit(t, g, "rm", "keep.txt")
	rawGit(t, g, "commit", "-m", "delete")
	rawGit(t, g, "checkout", "main")
	before := rawGit(t, g, "rev-parse", "HEAD")
	if r := g.execute(execInput{Target: "git", Args: []string{"merge", "deletion"}}); !strings.Contains(r.Error, "deletion") {
		t.Fatalf("merge deletion: %+v", r)
	}
	if rawGit(t, g, "rev-parse", "HEAD") != before {
		t.Fatal("merge changed HEAD")
	}
	if _, err := os.Stat(filepath.Join(g.path, "keep.txt")); err != nil {
		t.Fatal("merge deleted file")
	}
}

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"skyra/controller/hosts/journal"
	"strings"
	"testing"
)

func TestWritePreviewAtomicEditsAndConversationRevert(t *testing.T) {
	g := testRepo(t)
	g.session = filepath.Join(t.TempDir(), "chat/runs/1/.skyra")
	r := g.write(writeInput{Path: "new/f.txt", Text: "preview\n", DryRun: true, CreateOnly: true})
	if r.Error != "" || r.Diff == "" || r.SHA256 == "" {
		t.Fatal(r)
	}
	if _, e := os.Stat(filepath.Join(g.path, "new")); !os.IsNotExist(e) {
		t.Fatal("preview created folders")
	}
	r = g.write(writeInput{Path: "keep.txt", Edits: []editPair{{Old: "keep", New: "first"}, {Old: "absent", New: "x"}}})
	if r.Error == "" {
		t.Fatal("partial batch accepted")
	}
	body, _ := os.ReadFile(filepath.Join(g.path, "keep.txt"))
	if string(body) != "keep\n" {
		t.Fatal("failed batch wrote content")
	}
	r = g.write(writeInput{Path: "keep.txt", Edits: []editPair{{Old: "keep", New: "first"}, {Old: "first", New: "second"}}})
	if r.Error != "" || r.StartLine != 1 || r.Diff == "" {
		t.Fatal(r)
	}
	first := g.session
	g.session = filepath.Join(filepath.Dir(filepath.Dir(first)), "2/.skyra")
	dir, _ := g.commonDir()
	l, e := journal.Load(dir, g.session)
	if e != nil || len(l.Files) != 1 || l.Files[0].Before != digest([]byte("keep\n")) {
		t.Fatal(l, e)
	}
	r = g.write(writeInput{Revert: "keep.txt"})
	if r.Error != "" {
		t.Fatal(r)
	}
	body, _ = os.ReadFile(filepath.Join(g.path, "keep.txt"))
	if string(body) != "keep\n" {
		t.Fatal("wrong baseline")
	}
	g.write(writeInput{Path: "keep.txt", Text: "ours\n"})
	os.WriteFile(filepath.Join(g.path, "keep.txt"), []byte("user edit\n"), 0644)
	if r := g.write(writeInput{Revert: "keep.txt"}); !strings.HasPrefix(r.Error, "stale") {
		t.Fatal(r)
	}
	g.session = "different-conversation"
	if r := g.write(writeInput{Revert: "keep.txt"}); r.Error == "" {
		t.Fatal("reverted someone else's file")
	}
}
func TestFindChangedFilesSortingAndBoundary(t *testing.T) {
	g := testRepo(t)
	os.WriteFile(filepath.Join(g.path, "keep.txt"), []byte("changed hit\n"), 0644)
	os.WriteFile(filepath.Join(g.path, "new.txt"), []byte("new hit\n"), 0644)
	r := g.find(json.RawMessage(`{"changed":true,"match":"hit","sort":"modified"}`))
	if r.Error != "" || len(r.Matches) != 2 {
		t.Fatal(r)
	}
	outside := testRepo(t)
	os.Symlink(outside.path, filepath.Join(g.path, "link"))
	for _, raw := range []string{`{"in":"../outside"}`, `{"in":"link","match":"keep"}`, `{"in":".git","hidden":true,"ignore":"none","match":"keep"}`} {
		r := g.find(json.RawMessage(raw))
		if r.Error == "" && len(r.Matches) != 0 {
			t.Fatalf("escaped: %+v", r)
		}
	}
	invoke(t, g, "worktree", "add", "-b", "agent", "worktrees/agent")
	g.write(writeInput{Worktree: "worktrees/agent", Path: "keep.txt", Text: "worktree hit\n"})
	r = g.find(json.RawMessage(`{"in":"worktrees/agent","changed":true,"match":"worktree"}`))
	if r.Error != "" || len(r.Matches) != 1 || r.Matches[0].Path != "worktrees/agent/keep.txt" {
		t.Fatal(r)
	}
}
func TestReadonlyGitQueriesAndPublishingPolicy(t *testing.T) {
	g := testRepo(t)
	for _, args := range [][]string{{"rev-parse", "--git-dir"}, {"rev-parse", "--git-common-dir"}, {"log", "-2"}, {"branch", "--contains", "HEAD"}, {"merge-base", "HEAD", "HEAD"}, {"diff", "--stat", "main...HEAD"}} {
		invoke(t, g, args...)
	}
	g.denyPush = true
	if r := g.execute(execInput{Target: "git", Args: []string{"push"}}); !strings.Contains(r.Error, "SKYRA_ALLOW_PUSH=1") {
		t.Fatal(r)
	}
}
func TestGoFailureLocationsAndRecipeOptions(t *testing.T) {
	run := "^TestOne$"
	count := 2
	_, args, e := checkArgs(checkInput{Action: "go.test", Package: "./controller/x", Run: run, Count: &count})
	if e != nil || !strings.Contains(strings.Join(args, " "), "-count=2 -timeout=120s -run ^TestOne$") {
		t.Fatal(args, e)
	}
	for _, in := range []checkInput{{Action: "go.test", Run: "["}, {Action: "go.build", Run: "x"}, {Action: "go.test", Fix: true}} {
		if _, _, e := checkArgs(in); e == nil {
			t.Fatal(in)
		}
	}
	raw := `{"Action":"output","Package":"p","Test":"TestOne","Output":"    x_test.go:12: wrong\n"}` + "\n" + `{"Action":"fail","Package":"p","Test":"TestOne"}`
	text, failures := parseGoFailures(raw)
	if !strings.Contains(text, "wrong") || len(failures) != 1 || failures[0].Line != 12 || failures[0].Test != "TestOne" {
		t.Fatal(text, failures)
	}
}

func TestSwiftFailureReport(t *testing.T) {
	fs := parseSwiftReport([]byte(`<testsuites><testsuite><testcase name="testOne" classname="ExampleTests"><failure message="assertion">Example.swift:9: wrong</failure></testcase></testsuite></testsuites>`))
	if len(fs) != 1 || fs[0].Test != "testOne" || fs[0].Line != 9 {
		t.Fatal(fs)
	}
}
func TestWorktreeFilteredChecksAndFormatFix(t *testing.T) {
	requireOutsideCheckSandbox(t)
	g := testRepo(t)
	out, e := exec.Command("go", "env", "GOROOT").Output()
	if e != nil {
		t.Fatal(e)
	}
	g.goRoot = strings.TrimSpace(string(out))
	g.moduleCache = t.TempDir()
	os.WriteFile(filepath.Join(g.path, "go.mod"), []byte("module example\n\ngo 1.25.0\n"), 0644)
	os.MkdirAll(filepath.Join(g.path, "controller/example"), 0755)
	source := "package example\nimport \"testing\"\nfunc TestWanted(t *testing.T){}\nfunc TestOther(t *testing.T){t.Fatal(\"bad\")}\n"
	os.WriteFile(filepath.Join(g.path, "controller/example/example_test.go"), []byte(source), 0644)
	invoke(t, g, "add", "--", "go.mod", "controller")
	invoke(t, g, "commit", "-m", "fixture")
	invoke(t, g, "worktree", "add", "-b", "checks", "worktrees/checks")
	r := g.check(checkInput{Action: "go.test", Package: "./controller/example", Run: "^TestWanted$", Worktree: "worktrees/checks"})
	if r.Error != "" || r.ExitCode != 0 || !strings.HasPrefix(r.ReportPath, "worktrees/checks/") {
		t.Fatalf("%+v", r)
	}
	r = g.check(checkInput{Action: "go.test", Package: "./controller/example", Run: "^TestOther$", Worktree: "worktrees/checks"})
	if r.ExitCode == 0 || len(r.Failures) == 0 || r.Failures[0].Test != "TestOther" || r.Failures[0].Line == 0 {
		t.Fatalf("%+v", r)
	}
	g.write(writeInput{Worktree: "worktrees/checks", Path: "controller/example/example_test.go", Text: source + "\n"})
	r = g.check(checkInput{Action: "go.format", Fix: true, Worktree: "worktrees/checks"})
	if r.Error != "" || len(r.Formatted) != 1 {
		t.Fatalf("%+v", r)
	}
	body, _ := os.ReadFile(filepath.Join(g.path, "controller/example/example_test.go"))
	if string(body) != source {
		t.Fatal("format touched root")
	}
	r = g.check(checkInput{Action: "go.test", Package: "changed", Run: "^TestWanted$", Worktree: "worktrees/checks"})
	if r.Error != "" || r.ExitCode != 0 {
		t.Fatalf("%+v", r)
	}
}

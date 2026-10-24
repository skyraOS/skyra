package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireOutsideCheckSandbox(t *testing.T) {
	t.Helper()
	if os.Getenv("SKYRA_CHECK_SANDBOX") == "1" {
		t.Skip("requires an outer host: macOS prohibits nested sandbox-exec and external host sockets; run go test ./controller/hosts/repo outside repo.check")
	}
}

func TestRepositoryRecipeCanTestHostPackage(t *testing.T) {
	requireOutsideCheckSandbox(t)
	if _, err := os.Stat("/usr/bin/sandbox-exec"); err != nil {
		t.Skip("macOS Seatbelt required")
	}
	g := testRepo(t)
	source, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := snapshot(source, g.path, defaultLayout().Inputs); err != nil {
		t.Fatal(err)
	}
	for name, dest := range map[string]*string{"GOROOT": &g.goRoot, "GOMODCACHE": &g.moduleCache} {
		out, err := exec.Command("go", "env", name).Output()
		if err != nil {
			t.Fatal(err)
		}
		*dest = strings.TrimSpace(string(out))
	}
	r := g.handle([]byte(`{"action":"go.test","package":"./controller/hosts/repo"}`)).(checkResult)
	if r.ExitCode != 0 || r.Error != "" {
		t.Fatalf("host's own repo.check recipe: %+v", r)
	}
}

func TestGitQueryKeepsDiagnosticsOutOfValues(t *testing.T) {
	g := testRepo(t)
	wrapper := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nprintf 'shim diagnostic\\n' >&2\nexec /usr/bin/git \"$@\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	g.git = wrapper
	if dir, err := g.query("rev-parse", "--absolute-git-dir"); err != nil || dir != filepath.Join(g.path, ".git") {
		t.Fatalf("query: %q %v", dir, err)
	}
	if _, err := g.query("rev-parse", "--verify", "missing-ref"); err == nil || !strings.Contains(err.Error(), "shim diagnostic") {
		t.Fatalf("missing failure diagnostic: %v", err)
	}
}

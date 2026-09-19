package codex

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCodexLaunchPolicyPinsAndContains(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "codex-test")
	script := []byte("#!/bin/sh\nprintf 'codex-cli 0.test\\n'\n")
	if err := os.WriteFile(binary, script, 0o700); err != nil {
		t.Fatal(err)
	}
	policy := CodexLaunchPolicy{Binary: binary, ExpectedVersion: "codex-cli 0.test", CodexHome: directory, WorkingDir: directory, ExecutableSearchPath: []string{"/usr/bin", "/bin"}}
	spec, err := policy.ProcessSpec(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec.Env, []string{"CODEX_HOME=" + directory, "PATH=/usr/bin:/bin"}) {
		t.Fatalf("environment = %#v", spec.Env)
	}
	joined := strings.Join(spec.Args, " ")
	for _, required := range []string{"--strict-config", "--ask-for-approval never", "--sandbox read-only", `web_search="disabled"`, "features.shell_tool=false", "features.view_image=false", "apps._default.enabled=false", "app-server --stdio"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in %s", required, joined)
		}
	}
	policy.ExpectedVersion = "another"
	if _, err := policy.ProcessSpec(context.Background()); err == nil {
		t.Fatal("version mismatch accepted")
	}
}

func TestCodexLaunchPolicyRejectsRelativePaths(t *testing.T) {
	policy := CodexLaunchPolicy{Binary: "codex", ExpectedVersion: "x", CodexHome: "/tmp/home", WorkingDir: "/tmp/work", ExecutableSearchPath: []string{"/usr/bin"}}
	if _, err := policy.ProcessSpec(context.Background()); err == nil {
		t.Fatal("relative executable accepted")
	}
}

package boundary

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryReadableButOnlyIdeasWritable(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS only")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		t.Fatal(err)
	}
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ideas := filepath.Join(repo, "ideas")
	if err := os.Mkdir(ideas, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(repo, "main.go")
	if err := os.WriteFile(source, []byte("original source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(repo, filepath.Join(ideas, "escape")); err != nil {
		t.Fatal(err)
	}
	floor := filepath.Join(repo, ".skyra")
	surface, exe := verified(t, self, floor)
	profile := Render(Boundary{Surface: surface, Authority: Authority{
		Read: []PathGrant{{Path: repo, Scope: "subpath"}}, Scratch: ideas,
	}}, floor)
	for _, tc := range []struct{ path, want string }{
		{filepath.Join(ideas, "note.md"), "allowed allowed denied"},
		{source, "denied allowed denied"},
		{filepath.Join(ideas, "escape", "main.go"), "denied allowed denied"},
	} {
		cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, exe, "-test.run=^TestOSScratchHelper$")
		cmd.Env = []string{"SKYRA_TEST_WRITE=" + tc.path}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("sandboxed helper: %v: %s", err, out)
		}
		if !strings.HasPrefix(string(out), tc.want) {
			t.Fatalf("%s: got %q, want %s", tc.path, out, tc.want)
		}
	}
	if data, err := os.ReadFile(source); err != nil || string(data) != "original source" {
		t.Fatalf("source changed: %q, %v", data, err)
	}
}

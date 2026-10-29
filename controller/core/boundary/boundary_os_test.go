package boundary

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOSReadHelper(t *testing.T) {
	if os.Getenv("SKYRA_TEST_READ_PATH") == "" {
		return
	}
	if _, err := os.ReadFile(os.Getenv("SKYRA_TEST_READ_PATH")); err == nil {
		os.Stdout.WriteString("allowed")
	} else {
		os.Stdout.WriteString("denied")
	}
}

func TestOSExecHelper(t *testing.T) {
	if os.Getenv("SKYRA_TEST_EXEC") == "" {
		return
	}
	if exec.Command("/bin/echo").Run() == nil {
		os.Stdout.WriteString("allowed")
	} else {
		os.Stdout.WriteString("denied")
	}
}

// A target cannot start another program itself, whatever its Boundary grants:
// only the runtime starts programs, and only through an emission it recorded.
func TestRenderedProfileDeniesExecOfAnyOtherProgram(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS only")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		t.Fatal(err)
	}
	floor, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	surface, exe := verified(t, self, floor)
	profile := Render(Boundary{Surface: surface, Authority: Authority{Read: []PathGrant{{Path: "/bin", Scope: "subpath"}}}}, floor)
	cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, exe, "-test.run=^TestOSExecHelper$")
	cmd.Env = []string{"SKYRA_TEST_EXEC=1"}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sandboxed helper: %v: %s", err, out)
	}
	if !strings.HasPrefix(string(out), "denied") {
		t.Fatalf("exec of /bin/echo: got %q, want denied", out)
	}
}

func TestRenderedProfileEnforcesReadGrant(t *testing.T) {
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
	secret := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secret, []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}
	secret, err = filepath.EvalSymlinks(secret)
	if err != nil {
		t.Fatal(err)
	}
	floor := filepath.Join(filepath.Dir(secret), ".skyra")
	if err := os.Mkdir(floor, 0o700); err != nil {
		t.Fatal(err)
	}
	surface, exe := verified(t, self, floor)
	protected := filepath.Join(floor, "record")
	if err := os.WriteFile(protected, []byte("runtime"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		read   []PathGrant
		target string
		want   string
	}{
		{"ungranted", nil, secret, "denied"},
		{"granted", []PathGrant{{Path: secret, Scope: "literal"}}, secret, "allowed"},
		{"floor overrides parent grant", []PathGrant{{Path: filepath.Dir(secret), Scope: "subpath"}}, protected, "denied"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			profile := Render(Boundary{Surface: surface, Authority: Authority{Read: tc.read}}, floor)
			cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, exe, "-test.run=^TestOSReadHelper$")
			cmd.Env = []string{"SKYRA_TEST_READ_PATH=" + tc.target}
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("sandboxed helper: %v: %s", err, out)
			}
			if !strings.HasPrefix(string(out), tc.want) {
				t.Fatalf("got %q, want %s", out, tc.want)
			}
		})
	}
}

// TestOSScratchHelper writes a script to SKYRA_TEST_WRITE, reads it back, and
// tries to run it, reporting each as allowed or denied.
func TestOSScratchHelper(t *testing.T) {
	path := os.Getenv("SKYRA_TEST_WRITE")
	if path == "" {
		return
	}
	verdict := func(err error) string {
		if err == nil {
			return "allowed"
		}
		return "denied"
	}
	werr := os.WriteFile(path, []byte("#!/bin/sh\necho ran\n"), 0o755)
	_, rerr := os.ReadFile(path)
	xerr := exec.Command(path).Run()
	os.Stdout.WriteString(verdict(werr) + " " + verdict(rerr) + " " + verdict(xerr))
}

// Scratch is read and write, never exec: what a target writes there it can
// read back, but cannot run; and it cannot write anywhere else.
func TestRenderedProfileScratchIsReadWriteNeverExec(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS only")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	scratch, floor, outside := filepath.Join(dir, "scratch"), filepath.Join(dir, ".skyra"), filepath.Join(dir, "elsewhere")
	for _, d := range []string{scratch, floor, outside} {
		if err := os.Mkdir(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	surface, exe := verified(t, self, floor)
	for _, tc := range []struct{ name, target, want string }{
		{"in scratch", filepath.Join(scratch, "tool.sh"), "allowed allowed denied"},
		{"outside scratch", filepath.Join(outside, "tool.sh"), "denied denied denied"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			profile := Render(Boundary{Surface: surface, Authority: Authority{Scratch: scratch}}, floor)
			cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, exe, "-test.run=^TestOSScratchHelper$")
			cmd.Env = []string{"SKYRA_TEST_WRITE=" + tc.target}
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("sandboxed helper: %v: %s", err, out)
			}
			if !strings.HasPrefix(string(out), tc.want) {
				t.Fatalf("write read exec: got %q, want %s", out, tc.want)
			}
		})
	}
}

func TestOSStateHelper(t *testing.T) {
	dir := os.Getenv("SKYRA_TEST_STATE_DIR")
	if dir == "" {
		return
	}
	verdict := func(err error) string {
		if err != nil {
			return "denied"
		}
		return "allowed"
	}
	werr := os.WriteFile(filepath.Join(dir, "probe"), []byte("x"), 0o600)
	_, rerr := os.ReadFile(filepath.Join(dir, "probe"))
	os.Stdout.WriteString(verdict(werr) + "," + verdict(rerr))
}

// An actor can only see its own state (next/32, decision 24): its sandbox
// reads and writes its own directory under the floor, and the rest of the
// floor stays denied, a sibling actor's directory included.
func TestAnActorSeesOnlyItsOwnState(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS only")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		t.Fatal(err)
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	floor := filepath.Join(base, ".skyra")
	for _, d := range []string{"interpreter", "other", "records"} {
		if err := os.MkdirAll(filepath.Join(floor, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	// A read grant that covers the whole floor changes nothing.
	surface, exe := verified(t, self, floor)
	surface.State = "interpreter"
	h := Boundary{Surface: surface, Authority: Authority{Read: []PathGrant{{Path: base, Scope: "subpath"}}, Write: []PathGrant{{Path: base, Scope: "subpath"}}}}
	if got := StateDir(h, floor); got != filepath.Join(floor, "interpreter") {
		t.Fatalf("state dir: %s", got)
	}
	for dir, want := range map[string]string{
		"interpreter": "allowed,allowed",
		"other":       "denied,denied",
		"records":     "denied,denied",
	} {
		cmd := exec.Command("/usr/bin/sandbox-exec", "-p", Render(h, floor), exe, "-test.run=^TestOSStateHelper$")
		cmd.Env = []string{"SKYRA_TEST_STATE_DIR=" + filepath.Join(floor, dir)}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: sandboxed helper: %v: %s", dir, err, out)
		}
		if !strings.HasPrefix(string(out), want) {
			t.Errorf("%s: got %q, want %s", dir, out, want)
		}
	}
	// No state named, or a name that is the kernel's: nothing of the floor.
	for _, name := range []string{"", "records", "../x", "Interp"} {
		if StateDir(Boundary{Surface: &Surface{Path: self, State: name}}, floor) != "" {
			t.Errorf("%q is no actor's state", name)
		}
	}
}

// verified stages a program where the runtime runs it from (43 D5): copied
// under the floor and named by its hash. It returns the surface the program
// is registered as and the copy to exec.
func verified(t *testing.T, path, floor string) (*Surface, string) {
	t.Helper()
	bin, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(bin)
	surface := &Surface{Path: path, SHA256: hex.EncodeToString(sum[:])}
	exe := Verified(Boundary{Surface: surface}, floor)
	if err := os.MkdirAll(filepath.Dir(exe), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, bin, 0o500); err != nil {
		t.Fatal(err)
	}
	return surface, exe
}

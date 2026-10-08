package horizon

import (
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

// A target cannot start another program itself, whatever its Horizon grants:
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
	profile := Render(Horizon{Surface: &Surface{Path: self}, Authority: Authority{Read: []PathGrant{{Path: "/bin", Scope: "subpath"}}}}, t.TempDir())
	cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, self, "-test.run=^TestOSExecHelper$")
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
			profile := Render(Horizon{Surface: &Surface{Path: self}, Authority: Authority{Read: tc.read}}, floor)
			cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, self, "-test.run=^TestOSReadHelper$")
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
	for _, tc := range []struct{ name, target, want string }{
		{"in scratch", filepath.Join(scratch, "tool.sh"), "allowed allowed denied"},
		{"outside scratch", filepath.Join(outside, "tool.sh"), "denied denied denied"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			profile := Render(Horizon{Surface: &Surface{Path: self}, Authority: Authority{Scratch: scratch}}, floor)
			cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, self, "-test.run=^TestOSScratchHelper$")
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
	h := Horizon{Surface: &Surface{Path: self, State: "interpreter"}, Authority: Authority{Read: []PathGrant{{Path: base, Scope: "subpath"}}, Write: []PathGrant{{Path: base, Scope: "subpath"}}}}
	if got := StateDir(h, floor); got != filepath.Join(floor, "interpreter") {
		t.Fatalf("state dir: %s", got)
	}
	for dir, want := range map[string]string{
		"interpreter": "allowed,allowed",
		"other":       "denied,denied",
		"records":     "denied,denied",
	} {
		cmd := exec.Command("/usr/bin/sandbox-exec", "-p", Render(h, floor), self, "-test.run=^TestOSStateHelper$")
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
		if StateDir(Horizon{Surface: &Surface{Path: self, State: name}}, floor) != "" {
			t.Errorf("%q is no actor's state", name)
		}
	}
}

// The resident's one hole (proposal 38): Codex's view_image may read image
// files under the Horizon's read grants, and nothing else there; the floor
// stays denied even to an image.
func TestResidentReadsOnlyImagesUnderReadGrants(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	real := func(p string) string { r, _ := filepath.EvalSymlinks(p); return r }
	self = real(self) // Seatbelt matches real paths
	work := real(t.TempDir())
	floor := filepath.Join(work, ".skyra")
	for name, body := range map[string]string{
		"a.png": "png", "B.JPG": "jpg", "deep/c.webp": "webp", "notes.txt": "secret", ".env": "TOKEN=x",
		"notes.png.txt": "not an image", ".skyra/state.png": "floor",
	} {
		os.MkdirAll(filepath.Dir(filepath.Join(work, name)), 0o700)
		os.WriteFile(filepath.Join(work, name), []byte(body), 0o600)
	}
	tmp := func() string { return real(t.TempDir()) }
	scratch := tmp()
	os.MkdirAll(filepath.Join(scratch, "attachments"), 0o700)
	os.WriteFile(filepath.Join(scratch, "attachments", "pasted.png"), []byte("png"), 0o600)
	os.WriteFile(filepath.Join(scratch, "notes.md"), []byte("notes"), 0o600)
	h := Horizon{Surface: &Surface{Path: self, ResidentProfile: "darwin-appserver-v1"},
		Authority: Authority{Read: []PathGrant{{Path: work, Scope: "subpath"}}, Scratch: scratch}}
	paths := ResidentPaths{Home: tmp(), Scratch: tmp(), Work: tmp(), Auth: filepath.Join(tmp(), "auth.json"), Cache: tmp(), ProxyPort: 1}
	profile, err := RenderResident(h, floor, paths)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"a.png": "allowed", "B.JPG": "allowed", "deep/c.webp": "allowed",
		"notes.txt": "denied", ".env": "denied", "notes.png.txt": "denied", ".skyra/state.png": "denied",
		// A pasted image, kept in the conversation's scratch.
		scratch + "/attachments/pasted.png": "allowed", scratch + "/notes.md": "denied",
	} {
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(work, name)
		}
		cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, self, "-test.run=^TestOSReadHelper$")
		cmd.Env = append(os.Environ(), "SKYRA_TEST_READ_PATH="+path)
		out, err := cmd.Output()
		if got := strings.TrimSpace(string(out)); !strings.HasPrefix(got, want) {
			var stderr []byte
			if e, ok := err.(*exec.ExitError); ok {
				stderr = e.Stderr
			}
			t.Errorf("%s: %q, want %s (%v %s)", name, got, want, err, stderr)
		}
	}
	// A delegated provider must not use native image access to bypass its tool ceiling.
	h.Delegated = true
	profile, err = RenderResident(h, floor, paths)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(work, "a.png"), filepath.Join(scratch, "attachments", "pasted.png")} {
		cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, self, "-test.run=^TestOSReadHelper$")
		cmd.Env = append(os.Environ(), "SKYRA_TEST_READ_PATH="+path)
		out, err := cmd.Output()
		if err != nil || !strings.HasPrefix(strings.TrimSpace(string(out)), "denied") {
			t.Fatalf("delegated resident read %s: %s %v", path, out, err)
		}
	}

}

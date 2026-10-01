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

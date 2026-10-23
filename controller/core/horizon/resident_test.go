package horizon

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// providers are Codex and Claude as the registry pins them (proposal 47):
// Codex is its own resident, Claude runs behind a wrapper with keychain auth.
func providers(dir string) map[string]struct {
	surface *Surface
	paths   ResidentPaths
} {
	home := filepath.Join(dir, "home")
	base := ResidentPaths{Home: home, Scratch: filepath.Join(dir, "tmp"), Work: filepath.Join(dir, "work"), Cache: filepath.Join(dir, "cache"), ProxyPort: 4100}
	codex, claude := base, base
	codex.Config, codex.Catalog, codex.Auth = filepath.Join(home, "config.toml"), filepath.Join(dir, "models.json"), "/Users/u/.codex/auth.json"
	claude.Config, claude.Helper, claude.Broker, claude.BridgePort = filepath.Join(home, ".claude", "settings.json"), filepath.Join(dir, "bin", "security"), "/private/tmp/skyra-1/kc.sock", 4200
	sum := strings.Repeat("a", 64)
	return map[string]struct {
		surface *Surface
		paths   ResidentPaths
	}{
		"codex": {&Surface{Path: "/opt/codex/bin/codex", SHA256: sum, ResidentProfile: "darwin-provider-v1",
			Provider: &Provider{Name: "codex", Program: "/opt/codex/bin/codex", Reads: []string{"/opt/codex/codex-package.json"}, Auth: "file", AuthPath: "/Users/u/.codex/auth.json"}}, codex},
		"claude": {&Surface{Path: "/build/claude-start", SHA256: sum, ResidentProfile: "darwin-provider-v1",
			Provider: &Provider{Name: "claude", Program: "/Users/u/.local/share/claude/versions/2.1.309", Auth: "keychain", Keychain: "Claude Code-credentials", Helper: "/build/claude-keychain"}}, claude},
	}
}

// The Horizon reaches both providers through the same lines: whatever a
// change of Horizon adds to or removes from one provider's profile, it adds
// to or removes from the other's.
func TestProvidersShareTheHorizonsRules(t *testing.T) {
	dir := "/private/var/run"
	floor := "/w/.skyra"
	horizons := []Horizon{
		{},
		{Authority: Authority{Read: []PathGrant{{Path: "/w", Scope: "subpath"}, {Path: "/x/shot.png", Scope: "literal"}}, Scratch: "/s"}},
		{Authority: Authority{Read: []PathGrant{{Path: "/w", Scope: "subpath"}}, Write: []PathGrant{{Path: "/w", Scope: "subpath"}}, Mach: []string{"com.example"}, Connect: []string{"/sock"}}},
		{Authority: Authority{Read: []PathGrant{{Path: "/w", Scope: "subpath"}}, Scratch: "/s"}, Delegated: true},
	}
	render := func(name string, h Horizon) []string {
		p := providers(dir)[name]
		h.Surface = p.surface
		profile, err := RenderResident(h, floor, p.paths)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Split(profile, "\n")
	}
	diff := func(a, b []string) (added []string) {
		for _, line := range b {
			if !slices.Contains(a, line) {
				added = append(added, line)
			}
		}
		return added
	}
	for i, h := range horizons[1:] {
		var changes [2][]string
		for j, name := range []string{"codex", "claude"} {
			base, with := render(name, horizons[0]), render(name, h)
			changes[j] = append(diff(base, with), diff(with, base)...)
		}
		if strings.Join(changes[0], "\n") != strings.Join(changes[1], "\n") {
			t.Errorf("horizon %d changes the providers differently:\ncodex:  %q\nclaude: %q", i+1, changes[0], changes[1])
		}
	}
}

func TestProviderProfiles(t *testing.T) {
	p := providers("/r")
	render := func(name string) string {
		profile, err := RenderResident(Horizon{Surface: p[name].surface}, "/w/.skyra", p[name].paths)
		if err != nil {
			t.Fatal(err)
		}
		return profile
	}
	codex, claude := render("codex"), render("claude")
	for name, profile := range map[string]string{"codex": codex, "claude": claude} {
		for _, want := range []string{"(deny default)", `(allow network-outbound (remote ip "localhost:4100"))`, `(deny file-read* file-write* (subpath "/w/.skyra"))`} {
			if !strings.Contains(profile, want+"\n") {
				t.Errorf("%s lacks %s", name, want)
			}
		}
		if strings.Contains(profile, "(allow default)") || strings.Contains(profile, "(remote tcp)") {
			t.Errorf("%s is not deny-default with one egress", name)
		}
	}
	// Claude reaches its login only through the helper: no keychain, no
	// security tool, one socket, and its own config cannot be changed.
	for _, never := range []string{"Keychains", `"/usr/bin/security"`, "securityd", "/Users/u/.claude"} {
		if strings.Contains(claude, never) {
			t.Errorf("claude profile names %s", never)
		}
	}
	for _, want := range []string{
		`(allow process-exec (literal "/w/.skyra/surfaces/` + strings.Repeat("a", 64) + `"))`,
		`(allow process-exec (literal "/Users/u/.local/share/claude/versions/2.1.309"))`,
		`(allow process-exec (literal "/r/bin/security"))`,
		`(allow network-outbound (remote unix-socket (path-literal "/private/tmp/skyra-1/kc.sock")))`,
		`(allow network-bind network-inbound (local ip "localhost:4200"))`,
		`(deny file-write* (literal "/r/home/.claude/settings.json"))`,
	} {
		if !strings.Contains(claude, want+"\n") {
			t.Errorf("claude lacks %s", want)
		}
	}
	for _, want := range []string{
		`(allow process-exec (literal "/opt/codex/bin/codex"))`,
		`(allow file-read* (literal "/opt/codex/codex-package.json"))`,
		`(allow file-read* (literal "/r/models.json"))`,
		`(deny file-write* (literal "/Users/u/.codex/auth.json"))`,
		`(deny file-write* (literal "/r/home/config.toml"))`,
	} {
		if !strings.Contains(codex, want+"\n") {
			t.Errorf("codex lacks %s", want)
		}
	}
	if strings.Contains(codex, "network-bind") || strings.Contains(codex, "unix-socket") {
		t.Error("codex was given claude's bridge or broker")
	}
	for _, bad := range []Surface{
		{Path: "/b", ResidentProfile: "darwin-appserver-v1", Provider: p["codex"].surface.Provider},
		{Path: "/b", ResidentProfile: "darwin-claude-v1", Provider: p["claude"].surface.Provider},
		{Path: "/b", ResidentProfile: "darwin-provider-v1"},
		{Path: "/b", ResidentProfile: "darwin-provider-v1", Provider: &Provider{Name: "x", Program: "/p", Auth: "env"}},
	} {
		if _, err := RenderResident(Horizon{Surface: &bad}, "/w/.skyra", p["codex"].paths); err == nil {
			t.Errorf("rendered %+v", bad)
		}
	}
}

func TestOSProviderHelper(t *testing.T) {
	op, arg, ok := strings.Cut(os.Getenv("SKYRA_TEST_PROVIDER"), ":")
	if !ok {
		return
	}
	var err error
	switch op {
	case "exec":
		err = exec.Command(arg, "-test.run=^$").Run()
	case "bind":
		var l net.Listener
		if l, err = net.Listen("tcp", "127.0.0.1:"+arg); err == nil {
			l.Close()
		}
	case "dial":
		var c net.Conn
		if c, err = net.Dial("unix", arg); err == nil {
			c.Close()
		}
	}
	if err == nil {
		os.Stdout.WriteString("allowed")
	} else {
		os.Stdout.WriteString("denied")
	}
}

// Under Claude's rendered profile the wrapper may run the CLI and the helper,
// not the real security tool; bind its one bridge port and no other; and
// reach the broker's socket and no other.
func TestOSClaudeProfileEnforced(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS only")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	real := func(p string) string { r, _ := filepath.EvalSymlinks(p); return r }
	self = real(self)
	dir := real(t.TempDir())
	floor := filepath.Join(dir, ".skyra")
	surface, exe := verified(t, self, floor)
	bin, _ := os.ReadFile(self)
	helper := filepath.Join(dir, "bin", "security")
	os.MkdirAll(filepath.Dir(helper), 0o700)
	os.WriteFile(helper, bin, 0o500)
	sockets, err := os.MkdirTemp("/tmp", "kc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sockets)
	sockets = real(sockets)
	listen := func(name string) string {
		path := filepath.Join(sockets, name)
		l, err := net.Listen("unix", path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { l.Close() })
		go func() {
			for {
				c, err := l.Accept()
				if err != nil {
					return
				}
				c.Close()
			}
		}()
		return path
	}
	broker, other := listen("kc.sock"), listen("other.sock")
	free := func() int {
		l, _ := net.Listen("tcp", "127.0.0.1:0")
		defer l.Close()
		return l.Addr().(*net.TCPAddr).Port
	}
	bridge, elsewhere := free(), free()
	surface.Path = filepath.Join(dir, "claude-start")
	surface.ResidentProfile = "darwin-provider-v1"
	surface.Provider = &Provider{Name: "claude", Program: self, Auth: "keychain", Keychain: "Claude Code-credentials", Helper: helper}
	paths := ResidentPaths{Home: dir + "/home", Scratch: dir + "/tmp", Work: dir + "/work", Cache: dir + "/cache", Config: dir + "/home/.claude/settings.json",
		Helper: helper, Broker: broker, ProxyPort: 1, BridgePort: bridge}
	profile, err := RenderResident(Horizon{Surface: surface}, floor, paths)
	if err != nil {
		t.Fatal(err)
	}
	for check, want := range map[string]string{
		"exec:/usr/bin/security":          "denied",
		"exec:" + helper:                  "allowed",
		"exec:/bin/sh":                    "denied",
		fmt.Sprintf("bind:%d", bridge):    "allowed",
		fmt.Sprintf("bind:%d", elsewhere): "denied",
		"dial:" + broker:                  "allowed",
		"dial:" + other:                   "denied",
	} {
		cmd := exec.Command("/usr/bin/sandbox-exec", "-p", profile, exe, "-test.run=^TestOSProviderHelper$")
		cmd.Env = append(os.Environ(), "SKYRA_TEST_PROVIDER="+check)
		out, err := cmd.Output()
		if got := strings.TrimSpace(string(out)); !strings.HasPrefix(got, want) {
			var stderr []byte
			if e, ok := err.(*exec.ExitError); ok {
				stderr = e.Stderr
			}
			t.Errorf("%s: %q, want %s (%v %s)", check, got, want, err, stderr)
		}
	}
}

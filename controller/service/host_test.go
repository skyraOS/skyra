//go:build darwin

package service

import (
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A program without a provider runs from its verified copy, under a profile
// that grants it its home, scratch and work and nothing else; open returns an
// unguessable handle, and the session answers only to it.
func TestOpenRunsARegisteredProgramUnderItsProfile(t *testing.T) {
	dir, err := os.MkdirTemp("/private/tmp", "skyra-service-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	// An echo program built for the test: a copy of a platform binary such
	// as /bin/cat does not start outside its signed location.
	src := filepath.Join(dir, "echo.go")
	os.WriteFile(src, []byte("package main\n\nimport (\"bufio\"\n\"os\")\n\nfunc main() {\n\tin := bufio.NewScanner(os.Stdin)\n\tfor in.Scan() {\n\t\tos.Stdout.Write(append(in.Bytes(), '\\n'))\n\t}\n}\n"), 0o600)
	if out, err := exec.Command("go", "build", "-o", filepath.Join(dir, "echo"), src).CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	config := filepath.Join(dir, "programs.json")
	b, _ := json.Marshal(map[string]any{"programs": map[string]any{"cat": map[string]any{"path": "echo"}}})
	if err := os.WriteFile(config, b, 0o600); err != nil {
		t.Fatal(err)
	}
	programs, err := Load(config)
	if err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(dir, "service.sock")
	l, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	h := &Host{Programs: programs, State: filepath.Join(dir, "state"), Start: time.Hour, Max: time.Hour}
	go h.Serve(l)
	t.Cleanup(h.Stop)

	if o := Ask(socket, Request{Op: "open", Program: "nothing", WaitMS: 1}); o.Error == "" || o.Incarnation != "" {
		t.Fatalf("an unregistered program was opened: %+v", o)
	}
	o := Ask(socket, Request{Op: "open", Program: "cat", WaitMS: 1})
	if o.Error != "" || len(o.Incarnation) != 32 || o.Pid == 0 || len(o.Profile) != 64 {
		t.Fatalf("open: %+v", o)
	}
	handle := o.Incarnation
	if o := Ask(socket, Request{Incarnation: handle, Op: "exchange", Deliver: []string{`{"id":1}`}, WaitMS: 5000}); o.Error != "" || len(o.Lines) != 1 || o.Lines[0] != "{\"id\":1}\n" {
		logs, _ := filepath.Glob(filepath.Join(dir, "state", "cat", "runs", "*", "stderr.log"))
		var said []byte
		if len(logs) == 1 {
			said, _ = os.ReadFile(logs[0])
		}
		t.Fatalf("exchange: %+v\nthe program said: %s", o, said)
	}
	if o := Ask(socket, Request{Incarnation: strings.Repeat("0", 32), Op: "exchange", WaitMS: 100}); !o.PeerEOF {
		t.Fatalf("another handle reached the session: %+v", o)
	}
	copies, _ := os.ReadDir(filepath.Join(dir, "state", "bin"))
	if len(copies) != 1 || len(copies[0].Name()) != 64 {
		t.Fatalf("the program did not run from its verified copy: %v", copies)
	}
	profiles, _ := filepath.Glob(filepath.Join(dir, "state", "cat", "runs", "*", "actor.sb"))
	if len(profiles) != 1 {
		t.Fatalf("profiles: %v", profiles)
	}
	profile, _ := os.ReadFile(profiles[0])
	for _, absent := range []string{"network", "mach-lookup", "keychain"} {
		if strings.Contains(string(profile), absent) {
			t.Fatalf("a program without a provider was granted %s:\n%s", absent, profile)
		}
	}
	if o := Ask(socket, Request{Incarnation: handle, Op: "stop", WaitMS: 5000}); !o.PeerEOF {
		t.Fatalf("stop: %+v", o)
	}
}

// A registration declares what its program needs, and only a provider
// program declares a CLI, a login or egress.
func TestLoadChecksRegistrations(t *testing.T) {
	Register("cli", stub{})
	dir := t.TempDir()
	for _, f := range []string{"wrapper", "cli", "proxy.py", "helper"} {
		os.WriteFile(filepath.Join(dir, f), []byte(f), 0o700)
	}
	provider := func() map[string]any {
		return map[string]any{"path": "wrapper", "provider": "cli", "cli": "cli", "auth": "keychain", "keychain": "cli-credentials", "helper": "helper",
			"proxy_hosts": []string{"example.com"}, "proxy": "proxy.py"}
	}
	load := func(p map[string]any) error {
		b, _ := json.Marshal(map[string]any{"programs": map[string]any{"p": p}})
		file := filepath.Join(dir, "programs.json")
		os.WriteFile(file, b, 0o600)
		_, err := Load(file)
		return err
	}
	if err := load(provider()); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(map[string]any){
		"unknown provider":    func(p map[string]any) { p["provider"] = "gpt" },
		"no hosts":            func(p map[string]any) { delete(p, "proxy_hosts") },
		"bad host":            func(p map[string]any) { p["proxy_hosts"] = []string{"Example.com:443"} },
		"no proxy":            func(p map[string]any) { delete(p, "proxy") },
		"unknown auth":        func(p map[string]any) { p["auth"] = "env" },
		"missing helper":      func(p map[string]any) { delete(p, "helper") },
		"quoted service":      func(p map[string]any) { p["keychain"] = `x" -s "other` },
		"relative read":       func(p map[string]any) { p["reads"] = []string{"pkg.json"} },
		"no provider, egress": func(p map[string]any) { delete(p, "provider") },
		"unknown field":       func(p map[string]any) { p["resident"] = true },
	} {
		p := provider()
		mutate(p)
		if err := load(p); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

type stub struct{}

func (stub) Pinned() string       { return "1.0.0" }
func (stub) Configure(*Run) error { return nil }
func (stub) Ready(*Run) error     { return nil }

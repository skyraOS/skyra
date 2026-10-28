//go:build darwin

package launch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
)

// cli is a provider that needs nothing beyond the generic launch.
type cli struct{}

func (cli) Pinned() string { return "1.0.0" }

func (cli) Configure(r *Run) error {
	r.Paths.Config = filepath.Join(r.Paths.Home, "config")
	r.Env = []string{"HOME=" + r.Paths.Home, "TMPDIR=" + r.Paths.Scratch}
	return os.WriteFile(r.Paths.Config, []byte("{}\n"), 0600)
}

func (cli) Ready(*Run) error { return nil }

// A provider's sessions outlive its process: the CLI keeps them under its
// home, keyed by its working directory, so a second start in the same
// conversation gets the same home and work and can resume the first's
// session. Scratch stays per run.
func TestProviderStartsShareHomeAndWork(t *testing.T) {
	Register("cli", cli{})
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	program := filepath.Join(dir, "cli")
	if err = os.WriteFile(program, []byte("#!/bin/sh\necho \"1.0.0 (cli)\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	auth := filepath.Join(dir, "credentials.json")
	if err = os.WriteFile(auth, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	proxy, err := filepath.Abs(filepath.Join("..", "..", "codex", "provider", "proxy.py"))
	if err != nil {
		t.Fatal(err)
	}
	surface := horizon.Surface{
		Path:            program,
		ResidentProfile: "darwin-provider-v1",
		ProxyPath:       proxy,
		ProxyHosts:      []string{"example.com"},
		ProxySHA256:     record.Sum(must(os.ReadFile(proxy))),
		Provider: &horizon.Provider{Name: "cli", Program: program, ProgramSHA256: record.Sum(must(os.ReadFile(program))),
			Auth: "file", AuthPath: auth},
	}
	floor, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := horizon.Horizon{Frame: "start", Surface: &surface}
	x := machine.StartResident{Run: machine.Run{Surface: surface}, Horizon: h, Floor: floor}
	root := filepath.Join(dir, "conversation", ".skyra")

	start := func() (home, tmp, work string) {
		t.Helper()
		p, err := Prepare(x, root)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Abort()
		for _, kv := range p.Env {
			if v, ok := strings.CutPrefix(kv, "HOME="); ok {
				home = v
			}
			if v, ok := strings.CutPrefix(kv, "TMPDIR="); ok {
				tmp = v
			}
		}
		return home, tmp, p.Work
	}
	home1, tmp1, work1 := start()
	session := filepath.Join(home1, "projects", "s.jsonl")
	if err = os.MkdirAll(filepath.Dir(session), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(session, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	home2, tmp2, work2 := start()
	if home1 == "" || home1 != home2 || work1 != work2 {
		t.Fatalf("second start moved home or work: %q %q, %q %q", home1, home2, work1, work2)
	}
	if tmp1 == tmp2 {
		t.Fatalf("scratch is shared between runs: %q", tmp1)
	}
	if _, err = os.Stat(session); err != nil {
		t.Fatalf("first run's session is gone: %v", err)
	}
	if link, err := filepath.EvalSymlinks(filepath.Join(home2, filepath.Base(auth))); err != nil || link != auth {
		t.Fatalf("auth link was not renewed: %q %v", link, err)
	}
}

// An unregistered provider is refused before anything is started.
func TestUnknownProviderIsRefused(t *testing.T) {
	s := horizon.Surface{ResidentProfile: "darwin-provider-v1", Provider: &horizon.Provider{Name: "unregistered"}}
	if _, err := Prepare(machine.StartResident{Run: machine.Run{Surface: s}}, t.TempDir()); err == nil || !strings.Contains(err.Error(), "unknown provider") {
		t.Fatalf("got %v", err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
	"skyra/controller/runtime/launch"
)

func TestBundleContractsAndContentArePinnedTogether(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	write("interp", "installer binary")
	write("contract.json", `{"input":{"type":"string"},"continuation":true}`)
	entry := Entry{Path: "interp", Contract: "contract.json", State: true, Bundle: "bundle.json"}
	loadEntry := func() (horizon.Surface, error) {
		t.Helper()
		b, _ := json.Marshal(map[string]Entry{"interpreter": entry})
		write("registry.json", string(b))
		reach, _, err := Load(filepath.Join(dir, "registry.json"))
		return reach["interpreter"], err
	}
	body := `{"version":1,"procedures":[{"contract":{"type":"object","required":["load"],"fields":{"load":{"type":"string"}}},"nodes":[]}]}`
	write("bundle.json", body)
	s, err := loadEntry()
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Contracts) != 1 || !filepath.IsAbs(s.BundlePath) || s.BundleSHA256 != record.Sum([]byte(body)) {
		t.Fatalf("bundle not pinned with reach: %+v", s)
	}
	write("bundle.json", `{"version":1,"procedures":[{"contract":{"type":"string"}}]}`)
	if primary, err := loadEntry(); err != nil || len(primary.Contracts) != 0 {
		t.Fatalf("primary procedure contract: %+v %v", primary, err)
	}
	entry.State = false
	if _, err := loadEntry(); err == nil {
		t.Fatal("accepted stateless installer")
	}
	entry.State = true
	for _, invalid := range []string{
		`{"version":2,"procedures":[{"contract":{"type":"object"}}]}`,
		`{"version":1,"procedures":[]}`,
		`{"version":1,"procedures":[{"contract":{"type":"string"}},{"contract":{"type":"string"}}]}`,
		`{"version":1,"procedures":[{"contract":{"type":"object"}},{"contract":{"type":"object"}}]}`,
	} {
		write("bundle.json", invalid)
		if _, err := loadEntry(); err == nil {
			t.Fatalf("accepted invalid bundle: %s", invalid)
		}
	}
}

func TestLoadNormalizesAndRejectsGrants(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tool"), []byte("tool"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "contract.json"), []byte(`{"input":{"type":"string"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(dir, filepath.Join(dir, "alias")); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "registry.json")
	write := func(scope string) {
		t.Helper()
		body := `{"tool":{"path":"tool","contract":"contract.json","authority":{"read":[{"path":"alias/input","scope":"` + scope + `"}]}}}`
		if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("literal")
	reach, room, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reach["tool"]; !ok {
		t.Fatalf("tool not registered: %+v", reach)
	}
	canonicalDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := room.Read[0].Path; got != filepath.Join(canonicalDir, "input") {
		t.Fatalf("symlink alias was not resolved: %s", got)
	}
	write("anything")
	if _, _, err := Load(file); err == nil {
		t.Fatal("accepted an invalid Seatbelt path scope")
	}
}

// A surface is identity plus contracts (next/25): every grant in the file
// becomes the room's one authority, exec cannot be granted at all, and a
// contract cannot declare what its program may reach.
func TestLoadBuildsTheRoom(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a", "b"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for name, body := range map[string]string{
		"leaf.json":         `{"input":{"type":"string"}}`,
		"continuation.json": `{"input":{"type":"string"},"continuation":true}`,
		"reach.json":        `{"input":{"type":"string"},"reach":["a"]}`,
		"output.json":       `{"input":{"type":"string"},"output":{"type":"string"}}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	load := func(body string) (map[string]horizon.Surface, horizon.Authority, error) {
		file := filepath.Join(dir, "registry.json")
		if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return Load(file)
	}
	canonical, _ := filepath.EvalSymlinks(dir)
	reach, room, err := load(`{
		"a": {"path": "a", "contract": "leaf.json", "authority": {"read": [{"path": "` + dir + `", "scope": "subpath"}]}},
		"b": {"path": "b", "contract": "continuation.json", "authority": {"read": [{"path": "` + dir + `", "scope": "subpath"}], "mach": ["service.b"]}}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	want := horizon.Authority{Read: []horizon.PathGrant{{Path: canonical, Scope: "subpath"}}, Mach: []string{"service.b"}}
	if !reflect.DeepEqual(room, want) {
		t.Fatalf("every grant joins the room once: %+v", room)
	}
	if reach["a"].Continuation || !reach["b"].Continuation {
		t.Fatalf("how a program reads its input is its own declaration: %+v", reach)
	}
	if _, _, err := load(`{"a": {"path": "a", "contract": "leaf.json", "authority": {"exec": [{"path": "/bin/sh", "scope": "literal"}]}}}`); err == nil {
		t.Fatal("exec was granted: it can never be registered")
	}
	if _, _, err := load(`{"a": {"path": "a", "contract": "reach.json"}}`); err == nil {
		t.Fatal("a contract declared its own reach")
	}
	if _, _, err := load(`{"a": {"path": "a", "contract": "output.json"}}`); err == nil {
		t.Fatal("a contract declared an output: a return declares no contract")
	}
	if _, _, err := load(`{"a": {"path": "a", "contract": "leaf.json", "delegate": {}}}`); err == nil {
		t.Fatal("delegate is not a registry field any more")
	}
}

// Every contract the repository ships loads the way genesis loads it: a
// contract that does not is a room that never boots.
func TestEveryShippedContractLoads(t *testing.T) {
	var files []string
	for _, pattern := range []string{"../../targets/*/contract.json", "../../codex/*/contract.json", "../../codex/*/*-contract.json"} {
		m, _ := filepath.Glob(pattern)
		files = append(files, m...)
	}
	if len(files) < 5 {
		t.Fatalf("found only %v", files)
	}
	for _, f := range files {
		var d Declared
		if err := strict(f, &d); err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		if err := d.Input.Check(); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
}

// Contracts enter the room in their normal form (42): the root Horizon does
// not depend on the order a file lists its required fields.
func TestLoadNormalizesContracts(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("a"), 0o700); err != nil {
		t.Fatal(err)
	}
	body := `{"input":{"type":"object","required":["b","a"],"fields":{"a":{"type":"string"},"b":{"type":"string"}}},
		"contracts":[{"type":"object","required":["d","c","d"],"fields":{"c":{"type":"string"},"d":{"type":"string"}}}]}`
	if err := os.WriteFile(filepath.Join(dir, "c.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(file, []byte(`{"a": {"path": "a", "contract": "c.json"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	reach, _, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if got := reach["a"].Input.Required; !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("input contract: %v", got)
	}
	if got := reach["a"].Contracts[0].Required; !reflect.DeepEqual(got, []string{"c", "d"}) {
		t.Fatalf("other contracts: %v", got)
	}
}

// A provider resident takes the one provider profile (proposal 47), with its
// program and helper pinned by content: the old per-provider profiles and a
// provider without its profile or proxy are refused.
func TestProviderIsPinnedWithItsProfile(t *testing.T) {
	launch.Register("cli", nil) // the registry only asks whether the name is known
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	write("cli-start", "wrapper")
	write("cli", "the cli")
	write("cli-keychain", "helper")
	write("proxy.py", "proxy")
	write("contract.json", `{"input":{"type":"object"}}`)
	keychain := func() *horizon.Provider {
		return &horizon.Provider{Name: "cli", Program: "cli", Auth: "keychain", Keychain: "cli-credentials", Helper: "cli-keychain"}
	}
	base := Entry{Path: "cli-start", Contract: "contract.json", Resident: true, ResidentProfile: "darwin-provider-v1", ProxyHosts: []string{"example.com"}, ProxyPath: "proxy.py", Provider: keychain()}
	load := func(e Entry) (horizon.Surface, error) {
		t.Helper()
		b, _ := json.Marshal(map[string]Entry{"cli.start": e})
		write("registry.json", string(b))
		reach, _, err := Load(filepath.Join(dir, "registry.json"))
		return reach["cli.start"], err
	}
	s, err := load(base)
	if err != nil {
		t.Fatal(err)
	}
	p := s.Provider
	if p == nil || p.Program != filepath.Join(dir, "cli") || p.ProgramSHA256 != record.Sum([]byte("the cli")) ||
		p.Helper != filepath.Join(dir, "cli-keychain") || p.HelperSHA256 != record.Sum([]byte("helper")) {
		t.Fatalf("provider not pinned: %+v", p)
	}
	// A provider that is its own program: the pin is the surface's own hash.
	installed := base
	installed.Provider = &horizon.Provider{Name: "cli", Program: "cli-start", Auth: "file", AuthPath: "/Users/u/.cli/auth.json"}
	if s, err = load(installed); err != nil || s.Provider.Program != s.Path || s.Provider.ProgramSHA256 != s.SHA256 {
		t.Fatalf("installed: %+v %v", s.Provider, err)
	}
	for name, mutate := range map[string]func(e *Entry){
		"old appserver profile": func(e *Entry) { e.ResidentProfile = "darwin-appserver-v1" },
		"old claude profile":    func(e *Entry) { e.ResidentProfile = "darwin-claude-v1" },
		"no provider":           func(e *Entry) { e.Provider = nil },
		"no profile":            func(e *Entry) { e.ResidentProfile = "" },
		"no proxy":              func(e *Entry) { e.ProxyHosts = nil },
		"unknown name":          func(e *Entry) { e.Provider.Name = "gpt" },
		"unknown auth":          func(e *Entry) { e.Provider.Auth = "env" },
		"missing helper":        func(e *Entry) { e.Provider.Helper = "" },
		"quoted service":        func(e *Entry) { e.Provider.Keychain = `x" -s "other` },
		"keychain with a file":  func(e *Entry) { e.Provider.AuthPath = "/a" },
		"relative read":         func(e *Entry) { e.Provider.Reads = []string{"pkg.json"} },
		"not resident":          func(e *Entry) { e.Resident = false },
	} {
		e := base
		e.Provider = keychain()
		mutate(&e)
		if _, err := load(e); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/record"
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
	loadEntry := func() (boundary.Surface, error) {
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
	reach, root, err := Load(file)
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
	if got := root.Read[0].Path; got != filepath.Join(canonicalDir, "input") {
		t.Fatalf("symlink alias was not resolved: %s", got)
	}
	write("anything")
	if _, _, err := Load(file); err == nil {
		t.Fatal("accepted an invalid Seatbelt path scope")
	}
}

// A surface is identity plus contracts (next/25): every grant in the file
// becomes the root Boundary's one authority, exec cannot be granted at all, and a
// contract cannot declare what its program may reach.
func TestLoadBuildsTheRootBoundary(t *testing.T) {
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
	load := func(body string) (map[string]boundary.Surface, boundary.Authority, error) {
		file := filepath.Join(dir, "registry.json")
		if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return Load(file)
	}
	canonical, _ := filepath.EvalSymlinks(dir)
	reach, root, err := load(`{
		"a": {"path": "a", "contract": "leaf.json", "authority": {"read": [{"path": "` + dir + `", "scope": "subpath"}]}},
		"b": {"path": "b", "contract": "continuation.json", "authority": {"read": [{"path": "` + dir + `", "scope": "subpath"}], "mach": ["service.b"]}}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	want := boundary.Authority{Read: []boundary.PathGrant{{Path: canonical, Scope: "subpath"}}, Mach: []string{"service.b"}}
	if !reflect.DeepEqual(root, want) {
		t.Fatalf("every grant joins the root once: %+v", root)
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
// contract that does not is a registry that never boots.
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

// Contracts enter genesis in their normal form (42): the root Boundary does
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

// The workspace is named by an entry, never the first read grant: the root's
// grants are sorted, so another entry's grant can come first (here a cache
// directory, as the models entry's does).
func TestTheWorkspaceIsNamedNotTheFirstReadGrant(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"a-cache", "work", "elsewhere"} {
		os.Mkdir(filepath.Join(dir, d), 0o700)
	}
	os.WriteFile(filepath.Join(dir, "tool"), []byte("tool"), 0o700)
	os.WriteFile(filepath.Join(dir, "contract.json"), []byte(`{"input":{"type":"string"}}`), 0o600)
	load := func(read, models boundary.Authority) (boundary.Authority, error) {
		t.Helper()
		b, _ := json.Marshal(map[string]Entry{
			"read":   {Path: "tool", Contract: "contract.json", Authority: read},
			"models": {Path: "tool", Contract: "contract.json", Authority: models},
		})
		file := filepath.Join(dir, "registry.json")
		os.WriteFile(file, b, 0o600)
		_, root, err := Load(file)
		return root, err
	}
	grant := func(d string) []boundary.PathGrant {
		return []boundary.PathGrant{{Path: filepath.Join(dir, d), Scope: "subpath"}}
	}
	root, err := load(boundary.Authority{Read: grant("work"), Workspace: "work"}, boundary.Authority{Read: grant("a-cache")})
	if err != nil {
		t.Fatal(err)
	}
	if root.Read[0].Path != filepath.Join(dir, "a-cache") || root.Workspace != filepath.Join(dir, "work") {
		t.Fatalf("the workspace followed the grants' order: %+v", root)
	}
	if _, err := load(boundary.Authority{Read: grant("work"), Workspace: "elsewhere"}, boundary.Authority{}); err == nil {
		t.Fatal("accepted a workspace under no read grant")
	}
	if _, err := load(boundary.Authority{Read: grant("work"), Workspace: "work"}, boundary.Authority{Read: grant("a-cache"), Workspace: "a-cache"}); err == nil {
		t.Fatal("accepted two workspaces")
	}
}

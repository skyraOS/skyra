package registry

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"skyra/controller/core/horizon"
)

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

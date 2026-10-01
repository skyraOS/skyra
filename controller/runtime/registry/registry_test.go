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
		t.Fatal("a contract declared an output: a return's type is computed (next/30)")
	}
	if _, _, err := load(`{"a": {"path": "a", "contract": "leaf.json", "delegate": {}}}`); err == nil {
		t.Fatal("delegate is not a registry field any more")
	}
}

// A continuation declares its routes per child (next/30): each source is
// returned by hash, and the surface names it by that hash, so it is part of
// the surface's identity. Only a continuation has children to route.
func TestLoadDeclaredRoutes(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"a": "a", "b": "b",
		"leaf.json":         `{"input":{"type":"string"}}`,
		"continuation.json": `{"input":{"type":"string"},"continuation":true}`,
		"a.star":            "def route():\n  return skyra.interrupt()\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(dir, "registry.json")
	load := func(body string) (map[string]horizon.Surface, map[string][]byte, error) {
		t.Helper()
		if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		reach, _, sources, err := LoadWithRoutes(file)
		return reach, sources, err
	}
	reach, sources, err := load(`{"a": {"path": "a", "contract": "leaf.json"}, "b": {"path": "b", "contract": "continuation.json", "routes": {"a": "a.star"}}}`)
	if err != nil {
		t.Fatal(err)
	}
	h := reach["b"].Routes["a"]
	if string(sources[h]) != "def route():\n  return skyra.interrupt()\n" || len(sources) != 1 || reach["a"].Routes != nil {
		t.Fatalf("a declared route is returned by hash and named by the surface: %v %v", reach["b"].Routes, sources)
	}
	plain, _, _ := load(`{"a": {"path": "a", "contract": "leaf.json"}, "b": {"path": "b", "contract": "continuation.json"}}`)
	if horizon.Identity(plain["b"]) == horizon.Identity(reach["b"]) {
		t.Fatal("declared routes are part of the surface's identity")
	}
	if _, _, err := load(`{"a": {"path": "a", "contract": "leaf.json", "routes": {"a": "a.star"}}}`); err == nil {
		t.Fatal("a leaf declared routes")
	}
	if _, _, err := load(`{"b": {"path": "b", "contract": "continuation.json", "routes": {"ghost": "a.star"}}}`); err == nil {
		t.Fatal("a route for a child that is not registered")
	}
}

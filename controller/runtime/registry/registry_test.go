package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNormalizesAndRejectsGrants(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tool"), []byte("tool"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "contract.json"), []byte(`{"input":{"type":"string"},"output":{"type":"string"}}`), 0o600); err != nil {
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
	reach, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	canonicalDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reach["tool"].Authority.Read[0].Path; got != filepath.Join(canonicalDir, "input") {
		t.Fatalf("symlink alias was not resolved: %s", got)
	}
	write("anything")
	if _, err := Load(file); err == nil {
		t.Fatal("accepted an invalid Seatbelt path scope")
	}
}

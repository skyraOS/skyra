package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The roots are the authored procedures a registry installs, and each one's
// outline is derived from its definition: think narrows Codex, then calls it
// through the reference the narrowing returned (proposal 38).
func TestThinkIsARootWithItsOutline(t *testing.T) {
	dir := t.TempDir()
	bundle, _ := filepath.Abs("../../interp/workflows/think/bundle.json")
	registry := map[string]any{
		"echo":        map[string]any{"path": "echo", "contract": "c.json"},
		"interpreter": map[string]any{"path": "interp", "contract": "c.json", "state": true, "bundle": bundle},
	}
	b, _ := json.Marshal(registry)
	file := filepath.Join(dir, "registry.json")
	os.WriteFile(file, b, 0o600)

	ds, err := loadDestinations(file)
	if err != nil || len(ds) != 1 {
		t.Fatalf("destinations: %+v %v", ds, err)
	}
	think := ds[0]
	if think.Actor != "interpreter" || think.Title != "think" || len(think.Schema) == 0 {
		t.Fatalf("think: %+v", think)
	}
	nodes := think.Outline.Nodes
	if len(nodes) != 2 || nodes[0].Name != "scope" || nodes[1].Name != "run" {
		t.Fatalf("think's nodes, in order: %+v", nodes)
	}
	if nodes[0].Calls != "system" || nodes[0].Narrows == nil {
		t.Fatalf("scope narrows from input: %+v", nodes[0])
	}
	want := []string{"echo", "read", "find", "roots", "write", "codex", "codex.start", "codex.exchange", "codex.inspect", "codex.stop", "claude", "claude.start", "claude.exchange", "claude.inspect", "claude.stop", "interpreter"}
	if !reflect.DeepEqual(nodes[0].Narrows.Reach, want) {
		t.Fatalf("narrowed reach: %v", nodes[0].Narrows.Reach)
	}
	if !reflect.DeepEqual(nodes[1].After, []string{"scope"}) || nodes[1].Narrows != nil {
		t.Fatalf("run calls through scope's reference: %+v", nodes[1])
	}
	// The actor scope narrows to is think's model parameter (proposal 39),
	// not an actor named for the binding.
	if nodes[0].Narrows.Actor != "" || nodes[0].Narrows.Param != "model" {
		t.Fatalf("scope narrows to the model parameter: %+v", nodes[0].Narrows)
	}
	if nodes[1].Calls != "" || nodes[1].Via != "scope" {
		t.Fatalf("run calls the model scope narrowed: %+v", nodes[1])
	}

	h := &webhost{destinations: ds}
	if d, ok := h.destinationFor(""); !ok || d.Title != "think" {
		t.Fatal("with no selection a conversation runs the first root")
	}
	if _, ok := h.destinationFor("codex"); ok {
		t.Fatal("an actor that is not a root was selectable")
	}
	if _, ok := h.destinationFor("think-claude"); ok {
		t.Fatal("think-claude is a model of think, not a root")
	}
}

// The model is a parameter of think (proposal 39). Conversations recorded
// before then named Claude as a procedure; they load as think on Claude.
func TestTheModelIsAParameterNotARoot(t *testing.T) {
	for _, c := range []struct{ proc, model, wantProc, wantModel string }{
		{"think", "claude", "think", "claude"},
		{"think", "", "think", ""},
		{"think-claude", "", "think", "claude"},
		{"design", "", "design", ""},
	} {
		if p, m := normalizeRoot(c.proc, c.model); p != c.wantProc || m != c.wantModel {
			t.Errorf("normalizeRoot(%q, %q) = %q, %q", c.proc, c.model, p, m)
		}
	}
}

func TestARegistryWithoutBundlesHasNoRoots(t *testing.T) {
	file := filepath.Join(t.TempDir(), "registry.json")
	os.WriteFile(file, []byte(`{"echo":{"path":"echo","contract":"c.json"}}`), 0o600)
	if ds, err := loadDestinations(file); err != nil || len(ds) != 0 {
		t.Fatalf("destinations: %+v %v", ds, err)
	}
}

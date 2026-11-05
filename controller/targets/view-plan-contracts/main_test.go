package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewReturnsWhatTheHostLeft(t *testing.T) {
	state := t.TempDir()
	held := `{"contracts":[{"ref":"sha256:` + strings.Repeat("a", 64) + `","target":"read","input":{"type":"string"},"scope":{}}]}`
	if err := os.WriteFile(filepath.Join(state, File), []byte(held), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := view(state, []byte(`{"view_plan_contracts":{}}`))
	if err != nil || !strings.Contains(string(out), `"target":"read"`) {
		t.Fatalf("view lost the host's contracts: %s %v", out, err)
	}
	if _, err = view(state, []byte(`{"view_plan_contracts":{"target":"write"}}`)); err == nil {
		t.Fatal("a model shaped what the view returns")
	}
	if _, err = view(t.TempDir(), []byte(`{"view_plan_contracts":{}}`)); err == nil {
		t.Fatal("a view with nothing left by the host succeeded")
	}
}

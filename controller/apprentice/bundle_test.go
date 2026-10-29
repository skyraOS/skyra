package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"skyra/apprentice/proc"
)

func TestConversationReconstructionBundleRunsWiring(t *testing.T) {
	dir := t.TempDir()
	if err := installBundle(dir, "workflows/conversations/bundle.json"); err != nil {
		t.Fatal(err)
	}
	store := proc.Store{Dir: dir}
	b, err := os.ReadFile("workflows/conversations/reconstruct-state.json")
	if err != nil {
		t.Fatal(err)
	}
	var src proc.Source
	if err := json.Unmarshal(b, &src); err != nil {
		t.Fatal(err)
	}
	p, nodes, err := store.Procedure(src.Contract)
	if err != nil {
		t.Fatal(err)
	}
	// With no conversations the each node completes without invoking anything,
	// and the assembler alone produces the compatible empty replacement.
	in := json.RawMessage(`{"reconstruct-state":{"uiIncarnation":"h","revision":1,"conversations":[]}}`)
	next, err := proc.Plan(p, nodes, in, proc.Position{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Emit) != 1 {
		t.Fatalf("empty list needs only the assembler: %+v", next)
	}
	var call struct {
		To      string
		Payload map[string]struct{ Views []any }
	}
	if err := json.Unmarshal(next.Emit[0], &call); err != nil {
		t.Fatal(err)
	}
	if call.To != "user" || call.Payload["project-app-state"].Views == nil {
		t.Fatalf("app projection wiring: %s", next.Emit[0])
	}
	// The per-conversation body calls the owning actor first, before history
	// access. It binds the exact retained root/leaves into the syscall.
	b, _ = os.ReadFile("workflows/conversations/reconstruct-conversation.json")
	src = proc.Source{}
	json.Unmarshal(b, &src)
	p, nodes, err = store.Procedure(src.Contract)
	if err != nil {
		t.Fatal(err)
	}
	in = json.RawMessage(`{"reconstruct-conversation":{"conversation":"c","availability":{"input_live":false,"turn_live":false,"recovery_required":false}}}`)
	first, err := proc.Plan(p, nodes, in, proc.Position{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cp := json.RawMessage(`{"root_invocation":"root","leaves":["a","b"],"contract":{"target":"user","schema":{"type":"any"}}}`)
	next, err = proc.Plan(p, nodes, in, proc.Position{Rounds: [][]proc.Call{first.Round}}, []proc.Step{{To: "user", State: "return", Value: cp}})
	if err != nil {
		t.Fatal(err)
	}
	var syscall struct {
		To      string
		Payload map[string]struct {
			Invocation   string
			Observations []string
		}
	}
	json.Unmarshal(next.Emit[0], &syscall)
	if syscall.To != "system" || syscall.Payload["trace"].Invocation != "root" || len(syscall.Payload["trace"].Observations) != 2 {
		t.Fatalf("tree wiring: %s", next.Emit[0])
	}
	if err := installBundle(dir, "workflows/conversations/bundle.json"); err != nil {
		t.Fatalf("reprovisioning the identical bundle: %v", err)
	}
}
func TestBundleValidationPrecedesAnyPublication(t *testing.T) {
	dir := t.TempDir()
	src := proc.Source{Contract: &proc.Schema{Type: "object"}, Return: json.RawMessage(`null`)}
	b, _ := json.Marshal(map[string]any{"version": 1, "procedures": []proc.Source{src, src}})
	file := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(file, b, 0600)
	if err := installBundle(dir, file); err == nil {
		t.Fatal("duplicate contracts accepted")
	}
	if (proc.Store{Dir: dir}).Has(src.Contract) {
		t.Fatal("invalid bundle partially published")
	}
}

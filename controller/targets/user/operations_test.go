package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/targets/user/checkpoint"
)

func TestReadCheckpointUsesOnlyActorStorage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SKYRA_STATE", dir)
	root := strings.Repeat("a", 64)
	schema := json.RawMessage(`{"type":"object","required":["conversation"],"fields":{"conversation":{"type":"string"}}}`)
	_, err := (checkpoint.Store{Dir: filepath.Join(dir, "checkpoints")}).Initialize("c-1", root, checkpoint.Contract{Target: "user", Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	// Capture the actor's ordinary addressed return, as the kernel receives it.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	handled, opErr := operation(json.RawMessage(`{"frame":"f","input":{"read-checkpoint":{"conversation":"c-1"}}}`))
	w.Close()
	os.Stdout = old
	defer r.Close()
	if !handled || opErr != nil {
		t.Fatalf("checkpoint operation: %v %v", handled, opErr)
	}
	var out struct {
		To      string
		Payload checkpoint.Checkpoint
	}
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.To != "f" || out.Payload.RootInvocation != root {
		t.Fatalf("return: %+v", out)
	}
}
func TestProjectionRequiresExplicitAvailability(t *testing.T) {
	handled, err := operation(json.RawMessage(`{"frame":"f","input":{"project-state":{"conversation":"c-1","checkpoint":{},"strand":{}}}}`))
	if !handled || err == nil || !strings.Contains(err.Error(), "availability") {
		t.Fatalf("missing availability: %v %v", handled, err)
	}
}

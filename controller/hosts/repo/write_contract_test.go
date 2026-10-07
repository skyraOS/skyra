package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"skyra/controller/core/contract"
)

func TestGuardedEditWithoutTextFitsRoutingContract(t *testing.T) {
	declared, err := os.ReadFile("write-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Input *contract.Schema `json:"input"`
	}
	if err := json.Unmarshal(declared, &schema); err != nil {
		t.Fatal(err)
	}
	g := testRepo(t)
	request := json.RawMessage(`{"path":"keep.txt","old":"keep","new":"edited"}`)
	var payload any
	if err := json.Unmarshal(request, &payload); err != nil {
		t.Fatal(err)
	}
	if err := contract.Fits(schema.Input, payload); err != nil {
		t.Fatal(err)
	}
	if r := g.handle(request).(writeResult); r.Error != "" {
		t.Fatal(r)
	}
	if body, _ := os.ReadFile(filepath.Join(g.path, "keep.txt")); string(body) != "edited\n" {
		t.Fatalf("edit: %q", body)
	}
	if r := g.handle(json.RawMessage(`{"path":"keep.txt"}`)).(writeResult); r.Error == "" {
		t.Fatal("missing content was accepted")
	}
	if body, _ := os.ReadFile(filepath.Join(g.path, "keep.txt")); string(body) != "edited\n" {
		t.Fatal("missing text truncated file")
	}
}

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProposeReturnsOnlyTypedSpecData(t *testing.T) {
	raw := []byte(`{"propose_spec":{"version":1,"goal":"work","inputs":[],"capabilities":["sha256:` + strings.Repeat("a", 64) + `"]}}`)
	result, err := propose(raw)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err = json.Unmarshal(result, &value); err != nil {
		t.Fatal(err)
	}
	if value["goal"] != "work" || value["to"] != nil {
		t.Fatal("proposal supplied routing or lost its goal")
	}
	if _, err = propose([]byte(`{"propose_spec":{"version":1,"goal":"work","inputs":[],"capabilities":[],"grant":true}}`)); err == nil {
		t.Fatal("undeclared grant field accepted")
	}
}

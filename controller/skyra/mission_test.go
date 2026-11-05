//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skyra/controller/core/contract"
	typed "skyra/controller/core/mission"
	"skyra/controller/core/record"
	"skyra/controller/runtime/mission"
	"skyra/controller/runtime/registry"
)

func TestTypedMissionLaunchThroughRealKernel(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for name, source := range map[string]string{"kernel": ".", "echo": "../targets/echo", "actor": "testdata/mission-actor.go"} {
		if out, err := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, name), source).CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v %s", name, err, out)
		}
	}
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "apprentice"), ".")
	build.Dir = "../apprentice"
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build apprentice: %v %s", err, out)
	}
	absolute := func(path string) string { value, _ := filepath.Abs(path); return value }
	actorContract := filepath.Join(dir, "actor-contract.json")
	actorSchema := &contract.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*contract.Schema{"text": {Type: "string"}, "fresh": {Type: "bool"}, "tools": {Type: "array", Items: &contract.Schema{Type: "string"}}, "instructions": {Type: "any"}}}
	b, _ := json.Marshal(registry.Declared{Input: actorSchema})
	if err = os.WriteFile(actorContract, b, 0600); err != nil {
		t.Fatal(err)
	}
	reg := map[string]registry.Entry{
		"echo":       {Path: filepath.Join(dir, "echo"), Contract: absolute("../targets/echo/contract.json")},
		"actor":      {Path: filepath.Join(dir, "actor"), Contract: actorContract},
		"apprentice": {Path: filepath.Join(dir, "apprentice"), Contract: absolute("../apprentice/contract.json"), State: true},
	}
	regPath := filepath.Join(dir, "registry.json")
	b, _ = json.Marshal(reg)
	if err = os.WriteFile(regPath, b, 0600); err != nil {
		t.Fatal(err)
	}
	catalogue, err := mission.Load(regPath, []string{"echo"})
	if err != nil {
		t.Fatal(err)
	}
	kernelBytes, err := os.ReadFile(filepath.Join(dir, "kernel"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := mission.NewEndpoint("sha256:"+record.Sum(kernelBytes), true)
	if err != nil {
		t.Fatal(err)
	}
	launchEndpoint, err := mission.NewEndpoint(plan.Contract.Binary, false)
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := mission.Open(filepath.Join(dir, "human-root"), "mission", "human", catalogue, plan, launchEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := bridge.Root.Author("evidence", json.RawMessage(`"hello from the approved observation"`))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := bridge.AuthorSpec("human-spec", typed.Spec{Version: 1, Goal: "Echo the provided text", Inputs: evidence.Observations, Capabilities: catalogue.CapabilityRefs()})
	if err != nil {
		t.Fatal(err)
	}
	launch, err := bridge.StartLaunch("launch", spec.Observations[0])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = bridge.ExecuteLaunch(ctx, filepath.Join(dir, "echo"), "actor", launch); err == nil {
		t.Fatal("unregistered kernel binary accepted as the Launch endpoint")
	}
	result, err := bridge.ExecuteLaunch(ctx, filepath.Join(dir, "kernel"), "actor", launch)
	if err != nil {
		t.Fatal(err)
	}
	state, err := bridge.Root.State()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Observations) != 1 {
		t.Fatal("missing recorded launch result")
	}
	output, ok := state.Lookup(result.Observations[0])
	if !ok || output.Source().Target != launchEndpoint.Ref || !strings.Contains(string(output.Value()), "hello from the approved observation") {
		t.Fatal("typed observation handoff or result provenance lost")
	}
	if _, err = bridge.ExecuteLaunch(ctx, filepath.Join(dir, "kernel"), "actor", launch); err == nil {
		t.Fatal("same committed Launch executed twice")
	}
	duplicate, err := bridge.StartLaunch("launch", spec.Observations[0])
	if err != nil || duplicate != nil {
		t.Fatal("duplicate request dispatched again")
	}
}

//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	"skyra/controller/core/record"
	"skyra/controller/runtime/mission"
	"skyra/controller/runtime/registry"
)

func TestTypedMissionPlanAndLaunchThroughRealKernel(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for name, source := range map[string]string{"kernel": ".", "echo": "../targets/echo", "actor": "testdata/mission-actor.go", "propose-spec": "../targets/propose-spec", "view-plan-contracts": "../targets/view-plan-contracts", "workdir": "../targets/workdir"} {
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
		"echo":                {Path: filepath.Join(dir, "echo"), Contract: absolute("../targets/echo/contract.json")},
		"actor":               {Path: filepath.Join(dir, "actor"), Contract: actorContract},
		"propose-spec":        {Path: filepath.Join(dir, "propose-spec"), Contract: absolute("../targets/propose-spec/contract.json")},
		"view-plan-contracts": {Path: filepath.Join(dir, "view-plan-contracts"), Contract: absolute("../targets/view-plan-contracts/contract.json"), State: true},
		"apprentice":          {Path: filepath.Join(dir, "apprentice"), Contract: absolute("../apprentice/contract.json"), State: true},
	}
	// Plan looks around a workspace of its own: pwd, cd, ls and read pass a
	// directory observation between them.
	workspace := filepath.Join(dir, "workspace")
	if err = os.MkdirAll(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(workspace, "NOTES.md"), []byte("plan notes: echo the request\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range mission.Navigation {
		reg[name] = registry.Entry{Path: filepath.Join(dir, "workdir"), Contract: absolute("../targets/workdir/" + strings.TrimPrefix(name, "plan.") + ".json"),
			Authority: boundary.Authority{Read: []boundary.PathGrant{{Path: workspace, Scope: "subpath"}}, Workspace: workspace}}
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
	planner, err := mission.Load(regPath, append([]string{mission.Viewer, mission.Proposer}, mission.Navigation...))
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := mission.Open(filepath.Join(dir, "human-root"), "mission", "human", catalogue, planner, plan, launchEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := bridge.Root.Author("evidence", json.RawMessage(`"hello from the approved observation"`))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := bridge.AuthorSpec("human-spec", typed.HumanSpec("Echo the provided text", evidence.Observations, catalogue.CapabilityRefs(), typed.DefaultProfile))
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
	// The Launch's root set the profile its tree ran under: the one-contract
	// tree, recorded with what the tree was asked.
	launched, _ := filepath.Glob(filepath.Join(bridge.LaunchRoot(launch.Dispatch.Invocation.ID), "apprentice", "tree", "frames", "*.json"))
	if len(launched) != 1 {
		t.Fatalf("launch frames: %v", launched)
	}
	frame, _ := os.ReadFile(launched[0])
	var asked struct {
		Ask struct {
			Kind    string
			Profile map[string]int
		}
	}
	if json.Unmarshal(frame, &asked) != nil || asked.Ask.Kind != "act" || asked.Ask.Profile["act"] != 1 || asked.Ask.Profile["turns"] != 8 || asked.Ask.Profile["fanout"] != 32 || asked.Ask.Profile["budget"] != 64 {
		t.Fatalf("the Launch's tree did not record its profile: %+v", asked.Ask)
	}
	if _, err = bridge.ExecuteLaunch(ctx, filepath.Join(dir, "kernel"), "actor", launch); err == nil {
		t.Fatal("same committed Launch executed twice")
	}
	duplicate, err := bridge.StartLaunch("launch", spec.Observations[0])
	if err != nil || duplicate != nil {
		t.Fatal("duplicate request dispatched again")
	}

	// Plan: the human's request goes to a delegate Plan tree granted the
	// viewer and the proposer. A child views the human's contracts; another
	// proposes the Spec from that view, by tool; the Spec comes back into the
	// human's scope, and the human passes exactly that Spec to Launch.
	request, err := bridge.Root.Author("request", json.RawMessage(`"echo: planned by Plan"`))
	if err != nil {
		t.Fatal(err)
	}
	ungranted, _, err := bridge.StartPlan("plan-ungranted", request.Observations[0], nil, []root.Capability{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = bridge.ExecutePlan(ctx, filepath.Join(dir, "kernel"), "actor", ungranted); err == nil {
		t.Fatal("Plan ran without the planner's contracts as its grant")
	}
	planned, _, err := bridge.StartPlan("plan", request.Observations[0], nil, bridge.PlanGrant())
	if err != nil {
		t.Fatal(err)
	}
	returned, err := bridge.ExecutePlan(ctx, filepath.Join(dir, "kernel"), "actor", planned)
	if err != nil {
		t.Fatal(err)
	}
	if len(returned.Observations) != 1 {
		t.Fatal("Plan returned no Spec")
	}
	state, err = bridge.Root.State()
	if err != nil {
		t.Fatal(err)
	}
	specObservation, _ := state.Lookup(returned.Observations[0])
	if specObservation.Source().Target != plan.Ref {
		t.Fatal("the Spec lost Plan as its source")
	}
	frames, _ := filepath.Glob(filepath.Join(bridge.PlanRoot(planned.Invocation.ID), "apprentice", "tree", "frames", "*.json"))
	if len(frames) != 1 {
		t.Fatal("no Plan tree log")
	}
	log, _ := os.ReadFile(frames[0])
	for _, seen := range []string{`"dir":"."`, `"name":"NOTES.md"`, `plan notes: echo the request`} {
		if !strings.Contains(string(log), seen) {
			t.Fatalf("Plan did not look around its workspace: missing %s", seen)
		}
	}
	// Plan's own tree ran under Plan's profile.
	var planAsked struct {
		Ask struct{ Profile typed.Profile }
	}
	if json.Unmarshal(log, &planAsked) != nil || planAsked.Ask.Profile != mission.PlanProfile {
		t.Fatalf("Plan's tree did not run under its profile: %+v", planAsked.Ask.Profile)
	}
	if call, _ := state.Invocation(planned.Invocation.ID); !slices.Equal(call.Grant, bridge.PlanGrant()) {
		t.Fatalf("the root did not record the planner's contracts as Plan's grant: %+v", call)
	}
	planSpec, err := typed.DecodeSpec(specObservation.Value())
	if err != nil || len(planSpec.Capabilities) != 1 || planSpec.Capabilities[0] != catalogue.CapabilityRefs()[0] {
		t.Fatalf("Plan's Spec did not resolve echo from the viewed contracts: %s", specObservation.Value())
	}
	// The goal quotes the request, checked against it, after Plan's own words.
	if planSpec.Goal != "Echo the cited request: echo: planned by Plan" || planSpec.Coverage != (typed.Coverage{Words: 8, Cited: 4, Plan: 4}) || planSpec.Segments[1].Of != request.Observations[0] {
		t.Fatalf("Plan's Spec does not account for its words: %s", specObservation.Value())
	}
	if _, err = bridge.ExecutePlan(ctx, filepath.Join(dir, "kernel"), "actor", planned); err == nil {
		t.Fatal("same committed Plan executed twice")
	}
	launch, err = bridge.StartLaunch("launch-planned", returned.Observations[0])
	if err != nil {
		t.Fatal(err)
	}
	result, err = bridge.ExecuteLaunch(ctx, filepath.Join(dir, "kernel"), "actor", launch)
	if err != nil {
		t.Fatal(err)
	}
	state, err = bridge.Root.State()
	if err != nil {
		t.Fatal(err)
	}
	output, _ = state.Lookup(result.Observations[0])
	if !strings.Contains(string(output.Value()), "planned by Plan") {
		t.Fatalf("Launch did not receive the Spec's cited request: %s", output.Value())
	}
}

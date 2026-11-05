package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	"skyra/controller/runtime/registry"
)

func catalogueFixture(t *testing.T) *Catalogue {
	t.Helper()
	dir := t.TempDir()
	entries := map[string]registry.Entry{}
	for _, name := range []string{"tool", "actor", "apprentice", "ungranted"} {
		binary := filepath.Join(dir, name)
		if err := os.WriteFile(binary, []byte("binary:"+name), 0700); err != nil {
			t.Fatal(err)
		}
		declaration := registry.Declared{Input: &contract.Schema{Title: name, Type: "string"}}
		if name == "tool" {
			declaration.Contracts = []*contract.Schema{{Title: "another", Type: "number"}}
		}
		if name == "actor" {
			declaration.Input = &contract.Schema{Type: "object", Fields: map[string]*contract.Schema{"text": {Type: "string"}}, Required: []string{"text"}}
		}
		if name == "apprentice" {
			declaration.Continuation = true
			b, err := os.ReadFile("../../apprentice/contract.json")
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(b, &declaration); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(dir, name+".json")
		b, _ := json.Marshal(declaration)
		if err := os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
		entries[name] = registry.Entry{Path: binary, Contract: path, State: name == "apprentice"}
	}
	file := filepath.Join(dir, "registry.json")
	b, _ := json.Marshal(entries)
	if err := os.WriteFile(file, b, 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(file, []string{"tool"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func bridgeFixture(t *testing.T) (*Bridge, *Catalogue) {
	t.Helper()
	catalogue := catalogueFixture(t)
	plan, err := NewEndpoint("sha256:"+strings.Repeat("a", 64), true)
	if err != nil {
		t.Fatal(err)
	}
	launch, err := NewEndpoint(plan.Contract.Binary, false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Open(t.TempDir(), "mission", "operator", catalogue, plan, launch)
	if err != nil {
		t.Fatal(err)
	}
	return b, catalogue
}
func TestTypedPlanHumanLaunchPreservesSpecInputsAndRestrictsRegistry(t *testing.T) {
	b, c := bridgeFixture(t)
	request, err := b.Root.Author("request", json.RawMessage(`"Make a plan"`))
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := b.Root.Author("evidence", json.RawMessage(`{"count":9007199254740993}`))
	if err != nil {
		t.Fatal(err)
	}
	plan, payload, err := b.StartPlan("plan", request.Observations[0], evidence.Observations, c.CapabilityRefs())
	if err != nil {
		t.Fatal(err)
	}
	value, err := contract.Decode(payload)
	if err != nil || contract.Fits(typed.PlanSchema(), value) != nil {
		t.Fatal("Plan invocation does not fit its contract")
	}
	spec := typed.Spec{Version: 1, Goal: "Use the supplied evidence", Inputs: evidence.Observations, Capabilities: c.CapabilityRefs()}
	raw, _ := json.Marshal(spec)
	reader := func(call root.Invocation, receipt string) ([]json.RawMessage, error) {
		if call.ID != plan.Invocation.ID || receipt != "verified-plan-return" {
			return nil, fmt.Errorf("unverified return")
		}
		return []json.RawMessage{raw}, nil
	}
	result, err := b.ReceivePlan(plan.Invocation.ID, "verified-plan-return", reader)
	if err != nil {
		t.Fatal(err)
	}
	id := result.Observations[0]
	launch, err := b.StartLaunch("launch", id)
	if err != nil || launch == nil {
		t.Fatal(err)
	}
	if launch.Dispatch.Passed[0].Source().Target != b.Plan.Ref || launch.Dispatch.Passed[0].ID() != id {
		t.Fatal("Launch replaced Plan's observation")
	}
	dir := t.TempDir()
	file, worker, err := b.PrepareLaunch(dir, "actor", launch)
	if err != nil {
		t.Fatal(err)
	}
	reach, _, err := registry.Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(reach) != 3 || len(reach["tool"].All()) != 2 || reach["ungranted"].Path != "" {
		t.Fatal("Launch registry exceeds selected contracts")
	}
	var requestBody struct {
		Tree struct {
			Kind    string   `json:"kind"`
			Reach   []string `json:"reach"`
			Imports []struct {
				Ref    root.ID         `json:"ref"`
				Source root.Source     `json:"source"`
				Value  json.RawMessage `json:"value"`
			} `json:"imports"`
		} `json:"tree"`
	}
	if err = json.Unmarshal(worker, &requestBody); err != nil {
		t.Fatal(err)
	}
	if requestBody.Tree.Kind != "delegate" || len(requestBody.Tree.Reach) != 1 || len(requestBody.Tree.Imports) != 2 {
		t.Fatal("worker classification or imported scope changed")
	}
	if requestBody.Tree.Imports[0].Ref != id || requestBody.Tree.Imports[0].Source.Target != b.Plan.Ref || !strings.Contains(string(requestBody.Tree.Imports[1].Value), "9007199254740993") {
		t.Fatal("observation source/value was lost")
	}
	duplicate, err := b.StartLaunch("launch", id)
	if err != nil || duplicate != nil {
		t.Fatal("duplicate handoff dispatched again")
	}
}
func TestOnlySelectedContractIsPublished(t *testing.T) {
	b, c := bridgeFixture(t)
	refs := c.CapabilityRefs()[:1]
	proposed, err := b.AuthorSpec("spec", typed.Spec{Version: 1, Goal: "work", Inputs: []root.ID{}, Capabilities: refs})
	if err != nil {
		t.Fatal(err)
	}
	launch, err := b.StartLaunch("launch", proposed.Observations[0])
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := b.PrepareLaunch(t.TempDir(), "actor", launch)
	if err != nil {
		t.Fatal(err)
	}
	reach, _, err := registry.Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(reach["tool"].All()) != 1 || launch.Resolved.Kind != "act" {
		t.Fatal("unselected sibling contract remained callable")
	}
	// Registry projection checks the pinned bytes again before publishing.
	if err = os.WriteFile(c.surfaces["tool"].Path, []byte("changed"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, _, err = b.PrepareLaunch(t.TempDir(), "actor", launch); err == nil {
		t.Fatal("changed binary accepted at a previously approved path")
	}
}
func TestPlanCannotCiteUnpassedInputs(t *testing.T) {
	b, c := bridgeFixture(t)
	request, _ := b.Root.Author("request", json.RawMessage(`"request"`))
	private, _ := b.Root.Author("private", json.RawMessage(`"private"`))
	plan, _, err := b.StartPlan("plan", request.Observations[0], nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	spec := typed.Spec{Version: 1, Goal: "work", Inputs: private.Observations, Capabilities: c.CapabilityRefs()[:1]}
	raw, _ := json.Marshal(spec)
	_, err = b.ReceivePlan(plan.Invocation.ID, "receipt", func(root.Invocation, string) ([]json.RawMessage, error) { return []json.RawMessage{raw}, nil })
	if err == nil {
		t.Fatal("Plan cited private root scope")
	}
}

package mission

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
)

func TestPublishedContractsMatchProtocol(t *testing.T) {
	for path, schema := range map[string]*contract.Schema{
		"contracts/spec.json":                      SpecSchema(),
		"contracts/plan.json":                      PlanSchema(),
		"contracts/launch.json":                    LaunchSchema(),
		"../../targets/propose-spec/contract.json": ProposeSchema(),
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var declared struct {
			Input *contract.Schema `json:"input"`
		}
		if err = json.Unmarshal(raw, &declared); err != nil {
			t.Fatal(err)
		}
		if !contract.Same(schema, declared.Input) {
			t.Fatalf("%s differs from the typed protocol", path)
		}
	}
}

func fixture(t *testing.T) (Catalogue, root.State, []Descriptor) {
	t.Helper()
	scope := boundary.Authority{Read: []boundary.PathGrant{{Path: "/workspace", Scope: "subpath"}}}
	first, _ := Identify("sha256:"+strings.Repeat("1", 64), &contract.Schema{Title: "one", Type: "string"})
	second, _ := Identify(first.Binary, &contract.Schema{Title: "two", Type: "number"})
	c, err := NewCatalogue(scope, []Descriptor{{Target: "tool", Contract: first, Pin: "pin", Authority: scope}, {Target: "tool", Contract: second, Pin: "pin", Authority: scope}})
	if err != nil {
		t.Fatal(err)
	}
	entries := c.Entries()
	refs := []root.Capability{entries[0].Ref, entries[1].Ref}
	s, err := root.New(root.Config{Mission: "mission", Human: "human", Entrypoints: []root.Capability{"launch"}, Ceiling: refs})
	if err != nil {
		t.Fatal(err)
	}
	return c, s, entries
}
func commit(t *testing.T, s root.State, event root.Event) (root.State, root.Effect) {
	t.Helper()
	s, e := root.Step(s, root.HumanIngress, event)
	if e.Kind != root.Commit {
		t.Fatal(e.Reason)
	}
	return root.Step(s, root.StorageIngress, root.Committed{Stamp: e.Stamp})
}
func TestContractTypesMatchWorkerKnownAnswers(t *testing.T) {
	c, err := Identify("sha256:"+strings.Repeat("1", 64), &contract.Schema{Type: "object"})
	if err != nil {
		t.Fatal(err)
	}
	invoke, err := InvokeType(c, []string{"sha256:" + strings.Repeat("2", 64)})
	if err != nil {
		t.Fatal(err)
	}
	if invoke != "sha256:eeed6b565a9469dc163b3757e1f49b23f795b1db9aed94b5609f92696f56ee3a" {
		t.Fatal("worker input type differs", invoke)
	}
	ret, err := ReturnType(invoke, `{"type":"string"}`)
	if err != nil {
		t.Fatal(err)
	}
	if ret != "sha256:283d0f465c4bf6e0017b53a7f678da10ee450cfd9c561770a6fd9d1b0e0bf3ad" {
		t.Fatal("worker return type differs", ret)
	}
}
func TestTwoContractsOnOneTargetClassifyAsDelegate(t *testing.T) {
	c, s, entries := fixture(t)
	spec := Spec{Version: 1, Goal: "do the work", Inputs: []root.ID{}, Capabilities: []root.Capability{entries[0].Ref, entries[1].Ref}}
	resolved, err := c.Resolve(spec, s)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Kind != "delegate" || len(resolved.Targets) != 1 {
		t.Fatal("classification counted target names rather than contracts")
	}
	spec.Capabilities = spec.Capabilities[:1]
	resolved, err = c.Resolve(spec, s)
	if err != nil || resolved.Kind != "act" {
		t.Fatal("one contract is not act")
	}
}
func TestSpecReferencesBindScopesAndDoNotGrantByName(t *testing.T) {
	c, s, entries := fixture(t)
	narrow := entries[0]
	narrow.Ref = ""
	narrow.Authority.Read[0].Path = "/workspace/sub"
	changed, err := NewCatalogue(c.Ceiling(), []Descriptor{narrow})
	if err != nil {
		t.Fatal(err)
	}
	if changed.Entries()[0].Ref == entries[0].Ref {
		t.Fatal("resource scope did not change capability identity")
	}
	wide := entries[0]
	wide.Ref = ""
	wide.Authority.Write = []boundary.PathGrant{{Path: "/workspace", Scope: "subpath"}}
	if _, err = NewCatalogue(c.Ceiling(), []Descriptor{wide}); err == nil {
		t.Fatal("catalogue widened root authority")
	}
	if _, err = c.Resolve(Spec{Version: 1, Goal: "work", Capabilities: []root.Capability{"read"}}, s); err == nil {
		t.Fatal("plain tool name granted authority")
	}
	entries[0].Ref = "forged"
	if c.Entries()[0].Ref == "forged" {
		t.Fatal("catalogue aliases caller entries")
	}
}
func TestLaunchChecksRecordedInvocationAndExactGrant(t *testing.T) {
	c, s, entries := fixture(t)
	s, input := commit(t, s, root.Authored{RequestID: "input", Value: json.RawMessage(`"evidence"`)})
	spec := Spec{Version: 1, Goal: "work", Inputs: input.Observations, Capabilities: []root.Capability{entries[0].Ref}}
	raw, _ := json.Marshal(spec)
	s, result := commit(t, s, root.Authored{RequestID: "spec", Value: raw})
	specID := result.Observations[0]
	s, e := commit(t, s, root.Invoke{RequestID: "launch", Target: "launch", Inputs: append([]root.ID{specID}, spec.Inputs...), Grant: spec.Capabilities})
	if _, err := c.ValidateHandoff(e, s); err != nil {
		t.Fatal(err)
	}
	forged := e
	inv := *e.Invocation
	inv.ID = "unknown"
	forged.Invocation = &inv
	if _, err := c.ValidateHandoff(forged, s); err == nil {
		t.Fatal("fabricated dispatch accepted")
	}
	// The human root can choose held capabilities; the Launch contract additionally
	// requires them to be the exact proposal this Spec observation contains.
	s2, other := commit(t, s, root.Invoke{RequestID: "other", Target: "launch", Inputs: append([]root.ID{specID}, spec.Inputs...), Grant: []root.Capability{entries[1].Ref}})
	if _, err := c.ValidateHandoff(other, s2); err == nil {
		t.Fatal("grant changed independently of spec")
	}
}
func TestMalformedSpecsAndForeignReferencesFail(t *testing.T) {
	c, s, entries := fixture(t)
	for _, raw := range []string{`{"version":2,"goal":"x","inputs":[],"capabilities":[]}`, `{"version":1,"goal":"x","inputs":[],"capabilities":[],"shell":true}`, `{"version":1,"goal":"","inputs":[],"capabilities":[]}`} {
		if _, err := DecodeSpec(json.RawMessage(raw)); err == nil {
			t.Fatal("malformed spec accepted", raw)
		}
	}
	spec := Spec{Version: 1, Goal: "work", Inputs: []root.ID{root.ID(strings.Repeat("a", 64))}, Capabilities: []root.Capability{entries[0].Ref}}
	if _, err := c.Resolve(spec, s); err == nil {
		t.Fatal("foreign observation accepted")
	}
}

package machine

import (
	"encoding/json"
	"reflect"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/core/record"
)

func delegationRoot() boundary.Boundary {
	h := operator
	worker := agentS
	worker.Agent = &boundary.AgentConfig{Operations: []string{"agent.start"}}
	work := echoS
	work.WorkTool = true
	h.Reach = map[string]boundary.Surface{"agent": worker, "other": worker, "agent.start": {Path: "/bin/service-io", SHA256: "h", Input: echoIn}, "echo": work, "write": work, System: Syscalls()}
	h.Agent = &boundary.AgentContext{Actor: "agent", Model: "chosen-model"}
	h.Authority = boundary.Authority{Read: []boundary.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	return h
}

func delegated(t *testing.T, h boundary.Boundary, args string) boundary.Boundary {
	t.Helper()
	r := intakeRecord(h, `{"to":"system","payload":{"narrow":`+args+`}}`)
	if r.State != record.Invoke {
		t.Fatalf("scope refused: %+v", r.Fault)
	}
	scope, err := Narrow(h, r, "scope")
	if err != nil {
		t.Fatal(err)
	}
	caller := *scope.Resume
	r = intakeRecord(caller, `{"to":"system.ref.1","payload":{"text":"goal"}}`)
	if r.State != record.Invoke {
		t.Fatalf("child refused: %+v", r.Fault)
	}
	child, err := Narrow(caller, r, "child")
	if err != nil {
		t.Fatal(err)
	}
	return child
}

func TestDelegateScopeInheritsActorModelAuthority(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent","delegated":true,"inherit_model":true}`)
	if !reflect.DeepEqual(child.Agent, h.Agent) || !reflect.DeepEqual(child.Authority, h.Authority) || !child.Delegated || child.WorkTools != nil {
		t.Fatalf("inheritance: %+v", child)
	}
	other := delegated(t, h, `{"actor":"other","delegated":true}`)
	if other.Agent.Actor != "other" || other.Agent.Model != "" {
		t.Fatalf("explicit actor should use its default: %+v", other.Agent)
	}
}

func TestDelegateToolCeilingEnforcedAtIntakeAndInherited(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent","delegated":true,"tools":["echo"]}`)
	for name, want := range map[string]record.State{"echo": record.Invoke, "write": record.Fault, "agent.start": record.Invoke} {
		r := intakeRecord(child, `{"to":"`+name+`","payload":{"text":"x"}}`)
		if r.State != want {
			t.Fatalf("%s: %+v", name, r)
		}
	}
	// The actor's normal calls and an indirect narrowed call both retain the ceiling.
	r := intakeRecord(child, `{"to":"other","payload":{"text":"indirect"}}`)
	next, err := Narrow(child, r, "indirect")
	if err != nil {
		t.Fatal(err)
	}
	if intakeRecord(next, `{"to":"write","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("ordinary indirect call widened tools")
	}
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"agent","tools":["write"]}}}`).State != record.Fault {
		t.Fatal("nested scope widened tools")
	}
	g, err := narrowing(child, json.RawMessage(`{"actor":"write"}`))
	if err != nil || g.WorkTools != child.WorkTools {
		t.Fatalf("ordinary narrowing lost ceiling: %+v %v", g, err)
	}
	if !reflect.DeepEqual(h.AvailableTools(), delegationRoot().AvailableTools()) {
		t.Fatal("parent changed")
	}
}

func TestDelegateEmptyToolsRetainsOnlyExecutionMachinery(t *testing.T) {
	child := delegated(t, delegationRoot(), `{"actor":"agent","delegated":true,"tools":[]}`)
	if child.WorkTools == nil || len(*child.WorkTools) != 0 {
		t.Fatal("empty tools confused with omitted tools")
	}
	if intakeRecord(child, `{"to":"echo","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("reasoning child invoked a work tool")
	}
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"agent","delegated":true}}}`).State != record.Invoke {
		t.Fatal("execution machinery unavailable")
	}
	for _, name := range []string{"missing", "agent.start", "agent", "system"} {
		payload, _ := json.Marshal(map[string]any{"actor": "agent", "delegated": true, "tools": []string{name}})
		if _, err := narrowing(delegationRoot(), payload); err == nil {
			t.Fatalf("accepted non-work tool %s", name)
		}
	}
}

func TestDelegateTransportLabelsResolveWithoutWidening(t *testing.T) {
	h := delegationRoot()
	s := echoS
	s.WorkTool = true
	h.Reach["repo.write"] = s
	child := delegated(t, h, `{"actor":"agent","delegated":true,"tools":["repo_write"]}`)
	g, err := narrowing(child, json.RawMessage(`{"actor":"agent","tools":["repo.write"]}`))
	if err != nil || g.WorkTools == nil || (*g.WorkTools)["repo.write"].Target != "repo.write" {
		t.Fatalf("alias changed authority: %+v %v", g, err)
	}
	if intakeRecord(child, `{"to":"repo.write","payload":{"text":"x"}}`).State != record.Invoke {
		t.Fatal("transport label did not permit its owned contract")
	}
}

func TestDelegateChecksContractsRatherThanActorAlone(t *testing.T) {
	h := delegationRoot()
	a := call("allowed", &contract.Schema{Type: "object"})
	b := call("excluded", &contract.Schema{Type: "object"})
	remember := call("remember", &contract.Schema{Type: "object"})
	h.Reach["interpreter"] = boundary.Surface{Input: remember, Contracts: []*contract.Schema{a, b}}
	child := delegated(t, h, `{"actor":"agent","delegated":true,"tools":["allowed"]}`)
	if intakeRecord(child, `{"to":"interpreter","payload":{"allowed":{}}}`).State != record.Invoke {
		t.Fatal("allowed procedure refused")
	}
	if intakeRecord(child, `{"to":"interpreter","payload":{"excluded":{}}}`).State != record.Fault {
		t.Fatal("another contract on the same actor escaped")
	}
	// Adding a contract to the actor does not put it inside the child's tool ceiling.
	c := call("new", &contract.Schema{Type: "object"})
	s := child.Reach["interpreter"]
	s.Contracts = append(s.Contracts, c)
	copyReach := map[string]boundary.Surface{}
	for name, s := range child.Reach {
		copyReach[name] = s
	}
	copyReach["interpreter"] = s
	child.Reach = copyReach
	if intakeRecord(child, `{"to":"interpreter","payload":{"new":{}}}`).State != record.Fault {
		t.Fatal("extension widened tools")
	}
}

func TestAgentModelPublicationIsOwnedAndCannotChangeInheritedModel(t *testing.T) {
	h := delegationRoot()
	h.Name = "agent"
	s := h.Reach[h.Name]
	h.Surface = &s
	if _, err := agentContext(h, json.RawMessage(`{"model":"chosen-model"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := agentContext(h, json.RawMessage(`{"model":"different"}`)); err == nil {
		t.Fatal("inherited model changed")
	}
	h.Surface = &echoS
	if _, err := agentContext(h, json.RawMessage(`{"model":"chosen-model"}`)); err == nil {
		t.Fatal("work tool impersonated agent")
	}
}

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
	h.Reach = map[string]boundary.Surface{"agent": worker, "other": worker, "agent.start": {Path: "/bin/service-io", SHA256: "h", Input: echoIn}, "echo": echoS, "write": echoS, System: Syscalls()}
	h.Authority = boundary.Authority{Read: []boundary.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	return h
}

// keep is a narrow's contracts argument: every contract of each name, as
// held under h.
func keep(h boundary.Boundary, names ...string) string {
	kept := []map[string]any{}
	for _, name := range names {
		for _, c := range h.Reach[name].All() {
			kept = append(kept, map[string]any{"target": name, "schema": c})
		}
	}
	b, _ := json.Marshal(kept)
	return string(b)
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

func TestNarrowingInheritsAuthorityAndContracts(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent"}`)
	if !reflect.DeepEqual(child.Authority, h.Authority) || child.Held != nil {
		t.Fatalf("inheritance: %+v", child)
	}
	// Models, delegation and tool labels are the adapters' and the
	// interpreter's: narrow takes none of them.
	for _, payload := range []string{
		`{"narrow":{"actor":"agent","inherit_model":true}}`,
		`{"narrow":{"actor":"agent","delegated":true}}`,
		`{"narrow":{"actor":"agent","tools":["echo"]}}`,
		`{"agent-context":{"model":"m"}}`,
	} {
		if intakeRecord(h, `{"to":"system","payload":`+payload+`}`).State != record.Fault {
			t.Fatalf("system accepted %s", payload)
		}
	}
}

func TestHeldContractsAreEnforcedAtIntakeAndInherited(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent","contracts":`+keep(h, "echo", "other", "agent.start")+`}`)
	for name, want := range map[string]record.State{"echo": record.Invoke, "write": record.Fault, "agent.start": record.Invoke} {
		r := intakeRecord(child, `{"to":"`+name+`","payload":{"text":"x"}}`)
		if r.State != want {
			t.Fatalf("%s: %+v", name, r)
		}
	}
	// An actor it calls, and one it narrows, keep the same contracts.
	r := intakeRecord(child, `{"to":"other","payload":{"text":"indirect"}}`)
	next, err := Narrow(child, r, "indirect")
	if err != nil {
		t.Fatal(err)
	}
	if intakeRecord(next, `{"to":"write","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("an indirect call widened what is held")
	}
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"agent","contracts":`+keep(h, "write")+`}}}`).State != record.Fault {
		t.Fatal("a nested narrowing widened what is held")
	}
	g, err := narrowing(child, json.RawMessage(`{"actor":"write"}`))
	if err != nil || g.Held != child.Held {
		t.Fatalf("an ordinary narrowing lost the contracts: %+v %v", g, err)
	}
	if !reflect.DeepEqual(h.Effective(), delegationRoot().Effective()) {
		t.Fatal("parent changed")
	}
}

func TestEmptyContractsHoldOnlySystem(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent","contracts":[]}`)
	if child.Held == nil || len(*child.Held) != 1 || len((*child.Held)[System]) == 0 {
		t.Fatalf("empty contracts confused with omitted contracts: %+v", child.Held)
	}
	if intakeRecord(child, `{"to":"echo","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("a child holding nothing invoked echo")
	}
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"agent"}}}`).State != record.Invoke {
		t.Fatal("system is not held")
	}
	missing, _ := json.Marshal([]map[string]any{{"target": "missing", "schema": echoIn}})
	unknown, _ := json.Marshal([]map[string]any{{"target": "echo", "schema": &contract.Schema{Type: "string"}}})
	syscall, _ := json.Marshal([]map[string]any{{"target": System, "schema": Syscalls().Contracts[0]}})
	for _, kept := range []string{string(missing), string(unknown), string(syscall)} {
		if _, err := narrowing(h, json.RawMessage(`{"actor":"agent","contracts":`+kept+`}`)); err == nil {
			t.Fatalf("narrowing kept %s", kept)
		}
	}
	// What is kept by contract must be kept in reach.
	if _, err := narrowing(h, json.RawMessage(`{"actor":"agent","reach":["agent"],"contracts":`+keep(h, "echo")+`}`)); err == nil {
		t.Fatal("narrowing held a contract of an actor it dropped")
	}
}

// A reference is invoked as its actor: it is no way around what the caller
// holds.
func TestAReferenceDoesNotEscapeHeldContracts(t *testing.T) {
	child := delegated(t, delegationRoot(), `{"actor":"agent","contracts":[]}`)
	r := intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"write"}}}`)
	if r.State != record.Invoke {
		t.Fatalf("narrow refused: %+v", r.Fault)
	}
	scope, err := Narrow(child, r, "scope")
	if err != nil {
		t.Fatal(err)
	}
	if intakeRecord(*scope.Resume, `{"to":"system.ref.1","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("a reference ran an actor whose contract is not held")
	}
}

func TestHeldIsByContractRatherThanActor(t *testing.T) {
	h := delegationRoot()
	a := call("allowed", &contract.Schema{Type: "object"})
	b := call("excluded", &contract.Schema{Type: "object"})
	remember := call("remember", &contract.Schema{Type: "object"})
	h.Reach["interpreter"] = boundary.Surface{Input: remember, Contracts: []*contract.Schema{a, b}}
	kept, _ := json.Marshal([]map[string]any{{"target": "interpreter", "schema": a}})
	child := delegated(t, h, `{"actor":"agent","contracts":`+string(kept)+`}`)
	if intakeRecord(child, `{"to":"interpreter","payload":{"allowed":{}}}`).State != record.Invoke {
		t.Fatal("allowed procedure refused")
	}
	if intakeRecord(child, `{"to":"interpreter","payload":{"excluded":{}}}`).State != record.Fault {
		t.Fatal("another contract on the same actor escaped")
	}
	// A contract the actor adds to itself is declared, not held.
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
		t.Fatal("an export widened what is held")
	}
}

// Selection runs against every declared contract first: holding one of two
// contracts a payload fits does not make it unambiguous.
func TestHoldingOneContractKeepsAmbiguityAFault(t *testing.T) {
	h := delegationRoot()
	loose := &contract.Schema{Title: "loose", Type: "any"}
	h.Reach["either"] = boundary.Surface{Path: "/bin/either", SHA256: "x", Input: echoIn, Contracts: []*contract.Schema{loose}}
	kept, _ := json.Marshal([]map[string]any{{"target": "either", "schema": echoIn}})
	child := delegated(t, h, `{"actor":"agent","contracts":`+string(kept)+`}`)
	r := intakeRecord(child, `{"to":"either","payload":{"text":"fits both"}}`)
	if r.State != record.Fault || r.Fault.Kind != record.Contract {
		t.Fatalf("an ambiguous payload became valid: %+v", r)
	}
}

// A Boundary is named by what it holds (48): the same contracts, the same
// hash; fewer, or one more, another Boundary.
func TestABoundaryIsNamedByWhatItHolds(t *testing.T) {
	h := delegationRoot()
	if h.Hash() != delegationRoot().Hash() {
		t.Fatal("equal Boundaries hash differently")
	}
	same := delegated(t, h, `{"actor":"agent","contracts":`+keep(h, "agent", "other", "agent.start", "echo", "write")+`}`)
	if same.Hash() != h.Hash() {
		t.Fatal("narrowing to every held contract made another Boundary")
	}
	fewer := delegated(t, h, `{"actor":"agent","contracts":`+keep(h, "echo")+`}`)
	if fewer.Hash() == h.Hash() {
		t.Fatal("holding fewer contracts is the same Boundary")
	}
	exported := extended(h, h.Origin, "agent", call("shout", &contract.Schema{Type: "string"}))
	if exported.Hash() == h.Hash() {
		t.Fatal("an export left the Boundary's hash unchanged")
	}
}

func TestEveryNarrowingIsBoundedByTheFrameLimit(t *testing.T) {
	h := delegationRoot()
	for range maxDepth {
		below := h
		h.Resume = &below
	}
	if _, err := narrowing(h, json.RawMessage(`{"actor":"echo"}`)); err == nil {
		t.Fatal("a narrowing past the frame limit was accepted")
	}
}

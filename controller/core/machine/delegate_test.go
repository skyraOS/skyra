package machine

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/core/record"
)

func delegationRoot() boundary.Boundary {
	h := operator
	worker := agentS
	worker.Uses = []string{"agent.start"}
	h.Reach = map[string]boundary.Surface{"agent": worker, "other": worker, "agent.start": {Path: "/bin/service-io", SHA256: "h", Input: echoIn}, "echo": echoS, "write": echoS}
	h.Authority = boundary.Authority{Read: []boundary.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	return h
}

// keep is a narrow's contracts argument: every contract of each name, as
// in reach under h.
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
	r = intakeRecord(caller, `{"to":"`+handleFor(t, h, r)+`","payload":{"text":"goal"}}`)
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
	if !reflect.DeepEqual(child.Authority, h.Authority) || !reflect.DeepEqual(child.Reach, h.Reach) {
		t.Fatalf("inheritance: %+v", child)
	}
	// Models, delegation and tool labels are the adapters' and the
	// apprentice's: narrow takes none of them.
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

func TestKeptContractsAreEnforcedAtIntakeAndInherited(t *testing.T) {
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
		t.Fatal("an indirect call widened what is in reach")
	}
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"agent","contracts":`+keep(h, "write")+`}}}`).State != record.Fault {
		t.Fatal("a nested narrowing widened what is in reach")
	}
	if _, ok := child.Reach["write"]; ok {
		t.Fatal("an actor no contract of which is kept stayed in reach")
	}
	g, err := narrowing(child, json.RawMessage(`{"actor":"echo"}`))
	if err != nil || g.Contracts != nil || slices.Contains(g.Reach, "write") {
		t.Fatalf("an ordinary narrowing widened the contracts: %+v %v", g, err)
	}
	if h.Hash() != delegationRoot().Hash() {
		t.Fatal("parent changed")
	}
}

func TestEmptyContractsKeepOnlyTheActorAndWhatItRunsThrough(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent","contracts":[]}`)
	if len(child.Reach) != 2 || len(child.Reach["agent"].All()) == 0 || len(child.Reach["agent.start"].All()) == 0 {
		t.Fatalf("empty contracts confused with omitted contracts: %+v", child.Reach)
	}
	if intakeRecord(child, `{"to":"agent.start","payload":{"text":"x"}}`).State != record.Invoke {
		t.Fatal("the actor cannot reach what it runs through")
	}
	// Narrowed by name too, it keeps them in reach.
	g, err := narrowing(h, json.RawMessage(`{"actor":"agent","reach":[],"contracts":[]}`))
	if err != nil || len(g.Reach) != 1 || g.Reach[0] != "agent.start" {
		t.Fatalf("uses dropped from reach: %+v %v", g.Reach, err)
	}
	if intakeRecord(child, `{"to":"echo","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("a child keeping nothing invoked echo")
	}
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"agent"}}}`).State != record.Invoke {
		t.Fatal("a syscall was refused: system is held by no one and answers everyone")
	}
	missing, _ := json.Marshal([]map[string]any{{"target": "missing", "schema": echoIn}})
	unknown, _ := json.Marshal([]map[string]any{{"target": "echo", "schema": &contract.Schema{Type: "string"}}})
	syscall, _ := json.Marshal([]map[string]any{{"target": System, "schema": Syscalls()[0]}})
	for _, kept := range []string{string(missing), string(unknown), string(syscall)} {
		if _, err := narrowing(h, json.RawMessage(`{"actor":"agent","contracts":`+kept+`}`)); err == nil {
			t.Fatalf("narrowing kept %s", kept)
		}
	}
	// What is kept by contract must be kept in reach.
	if _, err := narrowing(h, json.RawMessage(`{"actor":"agent","reach":["agent"],"contracts":`+keep(h, "echo")+`}`)); err == nil {
		t.Fatal("narrowing kept a contract of an actor it dropped")
	}
}

// A handle is invoked as its actor: it is no way around what the caller
// has in reach.
func TestAHandleDoesNotEscapeKeptContracts(t *testing.T) {
	h := delegationRoot()
	r := intakeRecord(h, `{"to":"system","payload":{"narrow":{"actor":"write"}}}`)
	if r.State != record.Invoke {
		t.Fatalf("narrow refused: %+v", r.Fault)
	}
	handle := handleFor(t, h, r)
	child := delegated(t, h, `{"actor":"agent","contracts":[]}`)
	if intakeRecord(child, `{"to":"system","payload":{"narrow":{"actor":"write"}}}`).State != record.Fault {
		t.Fatal("a child narrowed to an actor it does not reach")
	}
	if intakeRecord(child, `{"to":"`+handle+`","payload":{"text":"x"}}`).State != record.Fault {
		t.Fatal("a handle ran an actor that is not in reach")
	}
}

func TestKeptIsByContractRatherThanActor(t *testing.T) {
	h := delegationRoot()
	a := call("allowed", &contract.Schema{Type: "object"})
	b := call("excluded", &contract.Schema{Type: "object"})
	remember := call("remember", &contract.Schema{Type: "object"})
	h.Reach["apprentice"] = boundary.Surface{Input: remember, Contracts: []*contract.Schema{a, b}}
	kept, _ := json.Marshal([]map[string]any{{"target": "apprentice", "schema": a}})
	child := delegated(t, h, `{"actor":"agent","contracts":`+string(kept)+`}`)
	if intakeRecord(child, `{"to":"apprentice","payload":{"allowed":{}}}`).State != record.Invoke {
		t.Fatal("allowed procedure refused")
	}
	if intakeRecord(child, `{"to":"apprentice","payload":{"excluded":{}}}`).State != record.Fault {
		t.Fatal("another contract on the same actor escaped")
	}
	if intakeRecord(child, `{"to":"apprentice","payload":{"remember":{}}}`).State != record.Fault {
		t.Fatal("the registered contract escaped when it was not kept")
	}
}

// Selection runs against the contracts in reach: a payload that fits two
// contracts is ambiguous where both are kept, and selects the one that is
// kept where the other was dropped.
func TestSelectionRunsAgainstWhatIsKept(t *testing.T) {
	h := delegationRoot()
	loose := &contract.Schema{Title: "loose", Type: "any"}
	h.Reach["either"] = boundary.Surface{Path: "/bin/either", SHA256: "x", Input: echoIn, Contracts: []*contract.Schema{loose}}
	kept, _ := json.Marshal([]map[string]any{{"target": "either", "schema": echoIn}})
	child := delegated(t, h, `{"actor":"agent","contracts":`+string(kept)+`}`)
	if r := intakeRecord(h, `{"to":"either","payload":{"text":"fits both"}}`); r.State != record.Fault || r.Fault.Kind != record.Contract {
		t.Fatalf("an ambiguous payload was accepted: %+v", r)
	}
	if r := intakeRecord(child, `{"to":"either","payload":{"text":"fits both"}}`); r.State != record.Invoke {
		t.Fatalf("the one contract kept was not selected: %+v", r)
	}
}

// A Boundary is named by what it reaches (48): the same contracts, the same
// hash; fewer, or one more, another Boundary.
func TestABoundaryIsNamedByWhatItReaches(t *testing.T) {
	h := delegationRoot()
	if h.Hash() != delegationRoot().Hash() {
		t.Fatal("equal Boundaries hash differently")
	}
	same := delegated(t, h, `{"actor":"agent","contracts":`+keep(h, "agent", "other", "agent.start", "echo", "write")+`}`)
	if same.Hash() != h.Hash() {
		t.Fatal("narrowing to every contract in reach made another Boundary")
	}
	fewer := delegated(t, h, `{"actor":"agent","contracts":`+keep(h, "echo")+`}`)
	if fewer.Hash() == h.Hash() {
		t.Fatal("keeping fewer contracts is the same Boundary")
	}
	proposed := extended(h, h.Origin, "agent", call("shout", &contract.Schema{Type: "string"}))
	if proposed.Hash() == h.Hash() {
		t.Fatal("a proposal left the Boundary's hash unchanged")
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

// A continuation is shown the contracts each actor in reach may be invoked
// through: one view, not tools beside reach.
func TestAContinuationIsShownWhatItReachesByContract(t *testing.T) {
	h := delegationRoot()
	child := delegated(t, h, `{"actor":"agent","contracts":`+keep(h, "echo")+`}`)
	shown := reach(child)
	if e := shown["echo"]; len(e.Contracts) != 1 {
		t.Fatalf("echo is in reach: %+v", e)
	}
	if _, ok := shown["write"]; ok {
		t.Fatal("write was dropped, and is shown")
	}
	if e := reach(h)["write"]; len(e.Contracts) != 1 {
		t.Fatalf("never narrowed, everything is in reach: %+v", e)
	}
}

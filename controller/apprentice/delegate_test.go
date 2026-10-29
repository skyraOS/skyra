package main

import (
	"encoding/json"
	"strings"
	"testing"

	"skyra/apprentice/proc"
)

func delegateEnvelope(input string) envelope {
	run := &proc.Schema{Type: "object"}
	e := envelope{Frame: "parent-slot", Input: json.RawMessage(input), Reach: heldReach(map[string]*proc.Schema{"codex": run, "claude": run}, nil)}
	return e
}

const (
	handle = "abababababababababababababababababababababababababababababababab"
	stale  = "cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd"
)

func TestDelegateRecordedCompositionHasGoalOnlyChildContext(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	// The calling adapter filled in its actor and model (proceduretools.Payload).
	e := delegateEnvelope(`{"delegate":{"goal":"inspect the code","actor":"codex","model":"chosen-model"}}`)
	out, err := runDelegate(store, e)
	if err != nil || !strings.Contains(string(out), `"actor":"codex"`) || strings.Contains(string(out), `"tools"`) || strings.Contains(string(out), "model") {
		t.Fatalf("scope: %s %v", out, err)
	}
	e.Steps = []proc.Step{{To: "system", State: "return", Value: json.RawMessage(`{"handle":"` + handle + `"}`)}}
	out, err = runDelegate(store, e)
	child := `{"payload":{"fresh":true,"header":"delegate","model":"chosen-model","text":"inspect the code"},"to":"` + handle + `"}`
	if err != nil || string(out) != child {
		t.Fatalf("goal-only child: %s %v", out, err)
	}
	e.Steps = append(e.Steps, proc.Step{To: handle, State: "return", Value: json.RawMessage(`{"text":"findings and uncertainty"}`)})
	out, err = runDelegate(store, e)
	if err != nil || !strings.Contains(string(out), `"text":"findings and uncertainty"`) || !strings.Contains(string(out), `"to":"parent-slot"`) {
		t.Fatalf("observations: %s %v", out, err)
	}

	// The handle is in the apprentice's table: the next delegation to the
	// same scope invokes it with no narrowing. Another scope narrows.
	again := delegateEnvelope(`{"delegate":{"goal":"inspect the code","actor":"codex","model":"chosen-model"}}`)
	if out, err := runDelegate(store, again); err != nil || string(out) != child {
		t.Fatalf("a remembered handle is invoked directly: %s %v", out, err)
	}
	other := delegateEnvelope(`{"delegate":{"goal":"inspect the code","actor":"claude"}}`)
	if out, err := runDelegate(store, other); err != nil || !strings.Contains(string(out), `"to":"system"`) {
		t.Fatalf("another scope is provisioned: %s %v", out, err)
	}
}

// A remembered handle the kernel refuses under this Boundary is forgotten
// and the scope provisioned again; a child's own boundary fault, which
// discharges the same slot, is the child's failure, not a refused handle.
func TestDelegateProvisionsAgainWhenAHandleIsRefused(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	input := `{"delegate":{"goal":"x","actor":"codex"}}`
	if err := store.SetHandle(`delegate {"actor":"codex"}`, stale); err != nil {
		t.Fatal(err)
	}
	e := delegateEnvelope(input)
	if out, err := runDelegate(store, e); err != nil || !strings.Contains(string(out), `"to":"`+stale+`"`) {
		t.Fatalf("the remembered handle: %s %v", out, err)
	}

	childFault := proc.Step{State: "fault", Value: json.RawMessage(`{"payload":{},"to":"shell"}`), Fault: &proc.Fault{Kind: "boundary", Cause: `"shell" is not in reach`}}
	e.Steps = []proc.Step{childFault}
	if out, err := runDelegate(store, e); err != nil || !strings.Contains(string(out), `"failed"`) {
		t.Fatalf("a child's boundary fault is its failure: %s %v", out, err)
	}

	refused := proc.Step{State: "fault", Value: json.RawMessage(`{"payload":{},"to":"` + stale + `"}`), Fault: &proc.Fault{Kind: "boundary", Cause: "handle: it grants what this Boundary does not hold"}}
	e.Steps = []proc.Step{refused}
	if out, err := runDelegate(store, e); err != nil || !strings.Contains(string(out), `"to":"system"`) {
		t.Fatalf("a refused handle is provisioned again: %s %v", out, err)
	}
	if _, ok, _ := store.Handle(`delegate {"actor":"codex"}`); ok {
		t.Fatal("the refused handle is still in the table")
	}
	e.Steps = append(e.Steps, proc.Step{To: "system", State: "return", Value: json.RawMessage(`{"handle":"` + handle + `"}`)})
	if out, err := runDelegate(store, e); err != nil || !strings.Contains(string(out), `"to":"`+handle+`"`) {
		t.Fatalf("then the new handle: %s %v", out, err)
	}
	if got, _, _ := store.Handle(`delegate {"actor":"codex"}`); got != handle {
		t.Fatalf("the table holds the new handle: %q", got)
	}
}

func TestDelegateValidationAndExplicitActor(t *testing.T) {
	for _, input := range []string{`{"delegate":{"goal":" "}}`, `{"delegate":{"goal":"x"}}`, `{"delegate":{"goal":"x","actor":"missing"}}`, `{"delegate":{"goal":"x","actor":""}}`} {
		out, err := runDelegate(proc.Store{Dir: t.TempDir()}, delegateEnvelope(input))
		if err != nil || !strings.Contains(string(out), `"failed"`) || strings.Contains(string(out), `"to":"system"`) {
			t.Fatalf("prelaunch validation: %s %v", out, err)
		}
	}
	// The adapter could not turn every tool into a contract: refused before
	// anything is narrowed.
	out, err := runDelegate(proc.Store{Dir: t.TempDir()}, delegateEnvelope(`{"delegate":{"goal":"x","actor":"claude","tools":["missing"]}}`))
	if err != nil || !strings.Contains(string(out), "not available to the caller") || strings.Contains(string(out), `"to":"system"`) {
		t.Fatalf("unresolved tools: %s %v", out, err)
	}
	// Empty tools keep no work contract: what the actor runs through is the
	// kernel's to keep, by its registration.
	out, err = runDelegate(proc.Store{Dir: t.TempDir()}, delegateEnvelope(`{"delegate":{"goal":"x","actor":"claude","tools":[],"contracts":[]}}`))
	if err != nil || !strings.Contains(string(out), `"actor":"claude"`) || !strings.Contains(string(out), `"contracts":[]`) || strings.Contains(string(out), "tools") {
		t.Fatalf("explicit actor: %s %v", out, err)
	}
}

func TestDelegateChildFaultIsFailureOutcome(t *testing.T) {
	e := delegateEnvelope(`{"delegate":{"goal":"x","actor":"codex"}}`)
	store := proc.Store{Dir: t.TempDir()}
	e.Steps = []proc.Step{{To: "system", State: "return", Value: json.RawMessage(`{"handle":"` + handle + `"}`)}, {To: handle, State: "fault", Fault: &proc.Fault{Kind: "custody", Cause: "provider exited"}}}
	out, err := runDelegate(store, e)
	if err != nil || !strings.Contains(string(out), `"failed":{"cause":"provider exited","kind":"custody"}`) {
		t.Fatalf("fault: %s %v", out, err)
	}
	e.Steps[1] = proc.Step{To: handle, State: "return", Value: json.RawMessage(`{"text":"","refused":"native.commandExecution"}`)}
	out, err = runDelegate(store, e)
	if err != nil || !strings.Contains(string(out), `"failed"`) {
		t.Fatalf("refusal: %s %v", out, err)
	}
}

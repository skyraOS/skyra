package main

import (
	"encoding/json"
	"strings"
	"testing"

	"skyra/interp/proc"
)

func delegateEnvelope(input string) envelope {
	run := &proc.Schema{Type: "object"}
	e := envelope{Frame: "parent-slot", Input: json.RawMessage(input), Tools: map[string]*proc.Schema{"codex": run, "claude": run}}
	return e
}

func TestDelegateRecordedCompositionHasGoalOnlyChildContext(t *testing.T) {
	// The calling adapter filled in its actor and model (proceduretools.Payload).
	e := delegateEnvelope(`{"delegate":{"goal":"inspect the code","actor":"codex","model":"chosen-model"}}`)
	out, err := runDelegate(e)
	if err != nil || !strings.Contains(string(out), `"actor":"codex"`) || strings.Contains(string(out), `"tools"`) || strings.Contains(string(out), "model") {
		t.Fatalf("scope: %s %v", out, err)
	}
	e.Steps = []proc.Step{{To: "system", State: "return", Value: json.RawMessage(`{"reference":"system.ref.1"}`)}}
	out, err = runDelegate(e)
	if err != nil || string(out) != `{"payload":{"fresh":true,"header":"delegate","model":"chosen-model","text":"inspect the code"},"to":"system.ref.1"}` {
		t.Fatalf("goal-only child: %s %v", out, err)
	}
	e.Steps = append(e.Steps, proc.Step{To: "system.ref.1", State: "return", Value: json.RawMessage(`{"text":"findings and uncertainty"}`)})
	out, err = runDelegate(e)
	if err != nil || !strings.Contains(string(out), `"text":"findings and uncertainty"`) || !strings.Contains(string(out), `"to":"parent-slot"`) {
		t.Fatalf("observations: %s %v", out, err)
	}
}

func TestDelegateValidationAndExplicitActor(t *testing.T) {
	for _, input := range []string{`{"delegate":{"goal":" "}}`, `{"delegate":{"goal":"x"}}`, `{"delegate":{"goal":"x","actor":"missing"}}`, `{"delegate":{"goal":"x","actor":""}}`} {
		out, err := runDelegate(delegateEnvelope(input))
		if err != nil || !strings.Contains(string(out), `"failed"`) || strings.Contains(string(out), `"to":"system"`) {
			t.Fatalf("prelaunch validation: %s %v", out, err)
		}
	}
	// The adapter could not turn every tool into a contract: refused before
	// anything is narrowed.
	out, err := runDelegate(delegateEnvelope(`{"delegate":{"goal":"x","actor":"claude","tools":["missing"]}}`))
	if err != nil || !strings.Contains(string(out), "not available to the caller") || strings.Contains(string(out), `"to":"system"`) {
		t.Fatalf("unresolved tools: %s %v", out, err)
	}
	// Empty tools keep no work contract: what the actor runs through is the
	// kernel's to keep, by its registration.
	out, err = runDelegate(delegateEnvelope(`{"delegate":{"goal":"x","actor":"claude","tools":[],"contracts":[]}}`))
	if err != nil || !strings.Contains(string(out), `"actor":"claude"`) || !strings.Contains(string(out), `"contracts":[]`) || strings.Contains(string(out), "tools") {
		t.Fatalf("explicit actor: %s %v", out, err)
	}
}

func TestDelegateChildFaultIsFailureOutcome(t *testing.T) {
	e := delegateEnvelope(`{"delegate":{"goal":"x","actor":"codex"}}`)
	e.Steps = []proc.Step{{To: "system", State: "return", Value: json.RawMessage(`{"reference":"system.ref.1"}`)}, {To: "system.ref.1", State: "fault", Fault: &proc.Fault{Kind: "custody", Cause: "provider exited"}}}
	out, err := runDelegate(e)
	if err != nil || !strings.Contains(string(out), `"failed":{"cause":"provider exited","kind":"custody"}`) {
		t.Fatalf("fault: %s %v", out, err)
	}
	e.Steps[1] = proc.Step{To: "system.ref.1", State: "return", Value: json.RawMessage(`{"text":"","refused":"native.commandExecution"}`)}
	out, err = runDelegate(e)
	if err != nil || !strings.Contains(string(out), `"failed"`) {
		t.Fatalf("refusal: %s %v", out, err)
	}
}

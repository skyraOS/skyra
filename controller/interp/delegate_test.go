package main

import (
	"encoding/json"
	"strings"
	"testing"

	"skyra/interp/proc"
)

func delegateEnvelope(input string) envelope {
	e := envelope{Frame: "parent-slot", Input: json.RawMessage(input), Agents: map[string]json.RawMessage{"codex": json.RawMessage(`{}`), "claude": json.RawMessage(`{}`)}}
	e.Agent = &struct{ Actor, Model string }{"codex", "chosen-model"}
	return e
}

func TestDelegateRecordedCompositionHasGoalOnlyChildContext(t *testing.T) {
	e := delegateEnvelope(`{"delegate":{"goal":"inspect the code"}}`)
	out, err := runDelegate(e)
	if err != nil || !strings.Contains(string(out), `"inherit_model":true`) || strings.Contains(string(out), `"tools"`) {
		t.Fatalf("scope: %s %v", out, err)
	}
	e.Steps = []proc.Step{{To: "system", State: "return", Value: json.RawMessage(`{"reference":"system.ref.1"}`)}}
	out, err = runDelegate(e)
	if err != nil || string(out) != `{"payload":{"header":"delegate","text":"inspect the code"},"to":"system.ref.1"}` {
		t.Fatalf("goal-only child: %s %v", out, err)
	}
	e.Steps = append(e.Steps, proc.Step{To: "system.ref.1", State: "return", Value: json.RawMessage(`{"text":"findings and uncertainty"}`)})
	out, err = runDelegate(e)
	if err != nil || !strings.Contains(string(out), `"text":"findings and uncertainty"`) || !strings.Contains(string(out), `"to":"parent-slot"`) {
		t.Fatalf("observations: %s %v", out, err)
	}
}

func TestDelegateValidationAndExplicitActor(t *testing.T) {
	for _, input := range []string{`{"delegate":{"goal":" "}}`, `{"delegate":{"goal":"x","actor":"missing"}}`, `{"delegate":{"goal":"x","actor":""}}`} {
		out, err := runDelegate(delegateEnvelope(input))
		if err != nil || !strings.Contains(string(out), `"failed"`) || strings.Contains(string(out), `"to":"system"`) {
			t.Fatalf("prelaunch validation: %s %v", out, err)
		}
	}
	out, err := runDelegate(delegateEnvelope(`{"delegate":{"goal":"x","actor":"claude","tools":[]}}`))
	if err != nil || !strings.Contains(string(out), `"actor":"claude"`) || !strings.Contains(string(out), `"inherit_model":false`) || !strings.Contains(string(out), `"tools":[]`) {
		t.Fatalf("explicit actor: %s %v", out, err)
	}
}

func TestDelegateChildFaultIsFailureOutcome(t *testing.T) {
	e := delegateEnvelope(`{"delegate":{"goal":"x"}}`)
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

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDelegatedThreadUsesInheritedModelAndNoNativeTools(t *testing.T) {
	c := turnWithCalls()
	c.Steps = c.Steps[:2]
	c.Actor = "codex"
	c.Input.Fresh = true
	c.Input.Model = "chosen-model"
	c.Reach = holding(c.Reach, []string{})
	_, m := emitted(t, c)
	var ex struct{ Deliver []string }
	if json.Unmarshal(m["payload"], &ex) != nil {
		t.Fatal("bad exchange")
	}
	var msg struct {
		Params struct {
			Model        string
			DynamicTools []json.RawMessage
			Config       map[string]any
		}
	}
	if json.Unmarshal([]byte(ex.Deliver[len(ex.Deliver)-1]), &msg) != nil {
		t.Fatal("bad RPC")
	}
	if msg.Params.Model != "chosen-model" || len(msg.Params.DynamicTools) != 0 || msg.Params.Config["web_search"] != "disabled" {
		t.Fatalf("delegated thread configuration: %+v", msg.Params)
	}
}

func TestResolvedModelStaysInTheAdapter(t *testing.T) {
	c := turnWithCalls(`echo {"text":"findings"}`)
	c.Actor = "codex"
	c.Steps[2].Value = json.RawMessage(strings.Replace(string(c.Steps[2].Value), `\"approvalPolicy\":\"never\"`, `\"model\":\"chosen-model\",\"approvalPolicy\":\"never\"`, 1))
	to, m := emitted(t, c)
	if to != "echo" || string(m["payload"]) != `{"text":"findings"}` {
		t.Fatalf("model tool call was not the next step: %s %s", to, m["payload"])
	}
}

func TestDelegateFailurePresentedAsUnsuccessfulToolResult(t *testing.T) {
	r := toolResult(json.RawMessage(`"call"`), step{To: "interpreter", State: "return", Value: json.RawMessage(`{"failed":{"kind":"custody","cause":"provider exited"}}`)})
	b, _ := json.Marshal(r)
	if !strings.Contains(string(b), `"success":false`) || !strings.Contains(string(b), "provider exited") {
		t.Fatalf("child failure became success: %s", b)
	}
}

func TestExistingThreadReportsItsModelBeforeDelegation(t *testing.T) {
	c := continuation{Frame: "frame", Actor: "codex"}
	c.Input.Incarnation, c.Input.Thread, c.Input.Text = "inc", "thread-existing", "next goal"
	c.Steps = []step{{To: "codex.inspect", State: "return", Payload: value(map[string]any{"deliver": []string{rpc("frame:inspect", "thread/loaded/list", map[string]any{})}}),
		Value: value(map[string]any{"lines": []string{`{"id":"frame:inspect","result":{"data":["thread-existing"]}}`}})}}
	to, m := emitted(t, c)
	if to != "codex.exchange" || !strings.Contains(string(m["payload"]), "thread/resume") {
		t.Fatalf("did not resolve existing model: %s %s", to, m["payload"])
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{"lines": []string{`{"id":"frame:resume","result":{"thread":{"id":"thread-existing"},"model":"chosen-model","approvalPolicy":"never","sandbox":{"type":"readOnly","networkAccess":false}}}`}})})
	to, m = emitted(t, c)
	if to != "codex.exchange" || !strings.Contains(string(m["payload"]), "turn/start") {
		t.Fatalf("turn did not start after the model was learned: %s %s", to, m["payload"])
	}
}

func TestCompletedDelegateStopsItsProviderBeforeReturning(t *testing.T) {
	c := turnWithCalls()
	c.Input.Fresh = true
	c.Steps = append(c.Steps, step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{}}), Value: value(map[string]any{"lines": []string{
		`{"method":"item/completed","params":{"threadId":"thread-1","turnId":"turn-1","item":{"type":"agentMessage","text":"findings"}}}`,
		`{"method":"turn/completed","params":{"threadId":"thread-1","turnId":"turn-1","turn":{"id":"turn-1","status":"completed"}}}`,
	}})})
	to, m := emitted(t, c)
	if to != "codex.stop" {
		t.Fatalf("child provider was retained: %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{"peer_eof": true})})
	to, m = emitted(t, c)
	if to != "frame" || string(m["payload"]) != `{"text":"findings"}` {
		t.Fatalf("stopped child restarted or lost findings: %s %s", to, m["payload"])
	}
}

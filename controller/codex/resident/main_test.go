package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func value(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func emitted(t *testing.T, c continuation) (string, map[string]json.RawMessage) {
	t.Helper()
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err = json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	var to string
	if err = json.Unmarshal(m["to"], &to); err != nil {
		t.Fatal(err)
	}
	return to, m
}

func TestExistingThreadInspectsBeforeTurnAndRestartsAfterIdle(t *testing.T) {
	c := continuation{Frame: "frame"}
	c.Input.Text = "next question"
	c.Input.Thread = "thread-1"
	c.Input.Incarnation = "old-process"
	c.Input.ActivityAt = 1000
	to, m := emitted(t, c)
	if to != "codex.inspect" {
		t.Fatalf("first operation %s", to)
	}
	var payload map[string]any
	json.Unmarshal(m["payload"], &payload)
	if payload["op"] != "inspect" {
		t.Fatal(payload)
	}
	c.Steps = append(c.Steps, step{To: to, Payload: m["payload"], State: "return", Value: value(map[string]any{
		"lines": []string{`{"id":"frame:inspect","result":{"data":[]}}` + "\n"}, "delivery": "written", "read": "observed", "observed_at": 1200, "peer_eof": false, "error": "", "partial": "",
	})})
	to, _ = emitted(t, c)
	if to != "codex.exchange" {
		t.Fatalf("recent thread should start a turn, got %s", to)
	}
	c.Steps[0].Value = value(map[string]any{
		"lines": []string{`{"id":"frame:inspect","result":{"data":[]}}` + "\n"}, "delivery": "written", "read": "observed", "observed_at": 1801000, "peer_eof": false, "error": "", "partial": "",
	})
	to, m = emitted(t, c)
	if to != "codex.stop" {
		t.Fatalf("idle thread should stop, got %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, Payload: m["payload"], State: "return", Value: value(map[string]any{
		"lines": []string{}, "delivery": "shutdown", "read": "observed", "observed_at": 1801001, "peer_eof": true, "error": "", "partial": "",
	})})
	to, _ = emitted(t, c)
	if to != "codex.start" {
		t.Fatalf("after stop should start a fresh process, got %s", to)
	}
}

func TestOnlyModelFacingToolsAreOfferedToCodex(t *testing.T) {
	c := continuation{Frame: "frame", Tools: map[string]*schema{
		"echo":       {Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}},
		"codex.stop": {Type: "object"}, "codex.exchange": {Type: "object"},
	}}
	c.Input.Text = "hello"
	c.Steps = []step{
		{To: "codex.start", State: "return", Value: value(map[string]any{"incarnation": "actor-1", "pid": 1, "profile": "hash"})},
		{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{rpc("frame:initialize", "initialize", map[string]any{})}}),
			Value: value(map[string]any{"lines": []string{`{"id":"frame:initialize","result":{"userAgent":"fake"}}` + "\n"}, "delivery": "written", "read": "observed", "observed_at": 1, "peer_eof": false, "error": "", "partial": ""})},
	}
	to, m := emitted(t, c)
	if to != "codex.exchange" {
		t.Fatalf("expected thread setup, got %s", to)
	}
	var q struct {
		Deliver []string `json:"deliver"`
	}
	if err := json.Unmarshal(m["payload"], &q); err != nil {
		t.Fatal(err)
	}
	if len(q.Deliver) != 2 || !strings.Contains(q.Deliver[1], `"name":"echo"`) || strings.Contains(q.Deliver[1], `"name":"codex.stop"`) || strings.Contains(q.Deliver[1], `"name":"codex.exchange"`) {
		t.Fatalf("offered tools: %v", q.Deliver)
	}
}

// turnWithCalls is a continuation whose turn has started and whose Codex has
// asked for each tool call given, in order.
func turnWithCalls(calls ...string) continuation {
	c := continuation{Frame: "frame", Tools: map[string]*schema{
		"echo": {Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}},
		"read": {Type: "object", Fields: map[string]*schema{"path": {Type: "string"}}, Required: []string{"path"}},
	}}
	c.Input.Text = "hello"
	observed := func(deliver []string, lines ...string) step {
		return step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": deliver}),
			Value: value(map[string]any{"lines": lines, "delivery": "written", "read": "observed", "observed_at": 1, "peer_eof": false, "error": "", "partial": ""})}
	}
	turn := []string{`{"id":"frame:turn","result":{"turn":{"id":"turn-1"}}}`}
	for i, call := range calls {
		tool, args, _ := strings.Cut(call, " ")
		turn = append(turn, fmt.Sprintf(`{"id":"call-%d","method":"item/tool/call","params":{"threadId":"thread-1","turnId":"turn-1","tool":%q,"arguments":%s}}`, i, tool, args))
	}
	c.Steps = []step{
		{To: "codex.start", State: "return", Value: value(map[string]any{"incarnation": "actor-1"})},
		observed([]string{rpc("frame:initialize", "initialize", map[string]any{})}, `{"id":"frame:initialize","result":{}}`),
		observed([]string{`{"method":"initialized"}`, rpc("frame:thread", "thread/start", map[string]any{})},
			`{"id":"frame:thread","result":{"thread":{"id":"thread-1"},"approvalPolicy":"never","sandbox":{"type":"readOnly","networkAccess":false}}}`),
		observed([]string{rpc("frame:turn", "turn/start", map[string]any{})}, turn...),
	}
	return c
}

func TestEachToolCallInvokesTheToolItNames(t *testing.T) {
	c := turnWithCalls(`echo {"text":"hi"}`, `read {"path":"README.md"}`)
	to, m := emitted(t, c)
	if to != "echo" || string(m["payload"]) != `{"text":"hi"}` {
		t.Fatalf("first call: %s %s", to, m["payload"])
	}
	c.Steps = append(c.Steps, step{To: "echo", Payload: m["payload"], State: "return", Value: value("hi")})
	to, m = emitted(t, c)
	var q struct{ Deliver []string }
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], `"id":"call-0"`) {
		t.Fatalf("echo's result should answer call-0: %s %v", to, q.Deliver)
	}
	c.Steps = append(c.Steps, step{To: to, Payload: m["payload"], State: "return", Value: value(map[string]any{"lines": []string{}, "delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": ""})})
	to, m = emitted(t, c)
	if to != "read" || string(m["payload"]) != `{"path":"README.md"}` {
		t.Fatalf("second call: %s %s", to, m["payload"])
	}
	c.Steps = append(c.Steps, step{To: "read", Payload: m["payload"], State: "fault", Fault: &struct{ Kind, Cause string }{"contract", "read: unknown field"}})
	to, m = emitted(t, c)
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], `"success":false`) || !strings.Contains(q.Deliver[0], "fault (contract)") {
		t.Fatalf("a tool fault goes back to Codex in the same turn: %s %v", to, q.Deliver)
	}
}

func TestAToolOutsideReachIsAViolation(t *testing.T) {
	c := turnWithCalls(`read {"path":"x"}`)
	delete(c.Tools, "read")
	to, m := emitted(t, c)
	var q struct{ Deliver []string }
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], "turn/interrupt") {
		t.Fatalf("an unoffered tool should interrupt the turn: %s %v", to, q.Deliver)
	}
	if _, err := decide(func() continuation {
		c := turnWithCalls(`read {"path":"x"}`)
		c.Steps = append(c.Steps, step{To: "echo", State: "return", Value: value("hi")})
		return c
	}()); err == nil || !strings.Contains(err.Error(), "asked for read") {
		t.Fatalf("a result from a different tool must not answer the call: %v", err)
	}
}

// A native attempt is reported as what it is: once the turn is interrupted,
// the adapter returns an observation of the attempt to its own slot. It does
// not name an unregistered capability to make the runtime fault.
func TestANativeAttemptIsReturnedAsAnObservation(t *testing.T) {
	c := turnWithCalls(`read {"path":"x"}`)
	delete(c.Tools, "read")
	to, m := emitted(t, c)
	if to != "codex.exchange" {
		t.Fatalf("the attempt interrupts the turn first: %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{
		"lines":    []string{`{"method":"turn/completed","params":{"turn":{"id":"turn-1","status":"interrupted"}}}`},
		"delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": "",
	})})
	to, m = emitted(t, c)
	var r struct{ Text, Refused, Thread string }
	if err := json.Unmarshal(m["payload"], &r); err != nil || to != "frame" || r.Refused != "native.dynamicTool.read" || r.Thread != "thread-1" {
		t.Fatalf("the attempt is returned to the slot, naming what was attempted: %s %s", to, m["payload"])
	}
}

// A failed turn ends with Codex's own reason, such as a usage limit: the human
// needs it, not only the status.
func TestAFailedTurnCarriesCodexsReason(t *testing.T) {
	c := turnWithCalls()
	reason := "You've hit your usage limit. Try again at Oct 3rd, 2026 11:01 PM."
	c.Steps = append(c.Steps, step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{}}),
		Value: value(map[string]any{"lines": []string{
			`{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1","status":"failed","error":{"message":` + string(value(reason)) + `,"codexErrorInfo":"usageLimitExceeded"}}}}`,
		}, "delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": ""})})
	_, err := decide(c)
	if err == nil || !strings.Contains(err.Error(), "turn ended failed: "+reason) {
		t.Fatalf("got %v", err)
	}
}

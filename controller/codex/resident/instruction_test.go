package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func instructionTurn(calls ...string) continuation {
	c := turnWithCalls(calls...)
	c.Input.Fresh = true
	c.Input.Tools = []string{}
	c.Input.Instructions = []instruction{
		{"goal", "Cut spec lines", json.RawMessage(`{"type":"object"}`)},
		{"delegate", "Pass observations", json.RawMessage(`{"type":"object"}`)},
		{"return", "Return observations", json.RawMessage(`{"type":"object"}`)},
	}
	return c
}
func exchangeReply(t *testing.T, c *continuation, lines ...string) map[string]json.RawMessage {
	t.Helper()
	to, m := emitted(t, *c)
	if to != "codex.exchange" {
		t.Fatalf("instruction dispatched to %s: %s", to, m["payload"])
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{"lines": lines})})
	return m
}

// Tree instructions are acknowledged in order. An ordinary tool in between
// uses its own result index and does not shift the instruction references.
func TestInstructionsAreCollectedWithoutDispatch(t *testing.T) {
	c := instructionTurn(`goal {"goal":{"of":0,"lines":[1,1]}}`, `echo {"text":"hi"}`, `delegate {"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}`)
	c.Input.Tools = []string{"echo"}
	first := exchangeReply(t, &c)
	if !strings.Contains(string(first["payload"]), "instruction 0") {
		t.Fatal(string(first["payload"]))
	}
	to, m := emitted(t, c)
	if to != "echo" {
		t.Fatalf("ordinary tool was %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value("hi")})
	exchangeReply(t, &c)
	last := exchangeReply(t, &c, `{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1","status":"completed"}}}`)
	if !strings.Contains(string(last["payload"]), "instruction 1") {
		t.Fatal(string(last["payload"]))
	}
	to, m = emitted(t, c)
	if to != "codex.stop" {
		t.Fatalf("fresh turn did not stop: %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{"peer_eof": true})})
	to, m = emitted(t, c)
	var result struct {
		Calls []struct {
			Name string
			Args json.RawMessage
		}
	}
	json.Unmarshal(m["payload"], &result)
	if to != "frame" || len(result.Calls) != 2 || result.Calls[0].Name != "goal" || result.Calls[1].Name != "delegate" || string(result.Calls[0].Args) != `{"goal":{"of":0,"lines":[1,1]}}` {
		t.Fatalf("reply %s %s", to, m["payload"])
	}
}
func TestInstructionThreadHasOnlyGrantedToolsAndMinimalPrompt(t *testing.T) {
	c := instructionTurn()
	c.Workspace, c.Scratch = "/private/work", "/private/scratch"
	c.Steps = c.Steps[:2]
	_, m := emitted(t, c)
	var q struct{ Deliver []string }
	json.Unmarshal(m["payload"], &q)
	var setup struct {
		Params struct {
			DynamicTools     []struct{ Name string }
			BaseInstructions string
			Config           map[string]any
		}
	}
	json.Unmarshal([]byte(q.Deliver[1]), &setup)
	var names []string
	for _, tl := range setup.Params.DynamicTools {
		names = append(names, tl.Name)
	}
	if strings.Join(names, " ") != "goal delegate return" || setup.Params.BaseInstructions != "You are inside Skyra, a cognitive architecture." || setup.Params.Config["web_search"] != "disabled" {
		t.Fatalf("setup %s", q.Deliver[1])
	}
	if offered(c, "read") || offered(c, "echo") || !offered(c, "goal") {
		t.Fatal("instruction turn leaked reachable tools")
	}
	// An explicit empty grant also hides tools without instruction mode.
	c.Input.Instructions = nil
	if len(tools(c)) != 0 {
		t.Fatal("empty tool grant ignored")
	}
}
func TestInstructionTurnRefusesOutOfGrantTools(t *testing.T) {
	for _, name := range []string{"echo", "read", "create_input"} {
		c := instructionTurn(name + ` {}`)
		to, m := emitted(t, c)
		if to != "codex.exchange" || !strings.Contains(string(m["payload"]), "turn/interrupt") {
			t.Fatalf("%s was not interrupted: %s %s", name, to, m["payload"])
		}
	}
}

func TestInstructionTurnRejectsAmbiguousOrReusedSetup(t *testing.T) {
	for _, change := range []func(*continuation){
		func(c *continuation) { c.Input.Fresh = false },
		func(c *continuation) { c.Input.Thread = "old" },
		func(c *continuation) { c.Input.Instructions = append(c.Input.Instructions, c.Input.Instructions[0]) },
		func(c *continuation) { c.Input.Tools = []string{"echo"}; c.Input.Instructions[0].Name = "echo" },
	} {
		c := instructionTurn()
		change(&c)
		if _, err := decide(c); err == nil {
			t.Fatal("ambiguous or reused instruction setup accepted")
		}
	}
}
func TestFailedInstructionTurnDoesNotReleaseCalls(t *testing.T) {
	c := instructionTurn(`goal {"goal":{"of":0,"lines":[1,1]}}`)
	exchangeReply(t, &c, `{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1","status":"failed","error":{"message":"provider failed"}}}}`)
	to, m := emitted(t, c)
	if to != "codex.stop" {
		t.Fatalf("fresh failed turn did not stop: %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{"peer_eof": true})})
	if _, err := decide(c); err == nil {
		t.Fatal("failed turn released its collected instructions")
	}
}

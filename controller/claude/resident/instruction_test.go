package main

import (
	"encoding/json"
	"strings"
	"testing"

	"skyra/controller/claude/bridge"
)

func lines(ls ...string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"lines": ls, "partial": "", "peer_eof": false, "error": ""})
	return b
}

func deliver(ls ...string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"deliver": ls})
	return b
}

// A thread granted instructions is offered them and nothing else; each call
// is acknowledged with its position, never dispatched; the turn returns
// them, in order, as its reply. A granted target tool between them is
// dispatched as before.
func TestInstructionsAreTheReply(t *testing.T) {
	c := continuation{Frame: "f1"}
	c.Input.Text = "node 0"
	c.Input.Fresh = true
	c.Input.Tools = []string{"echo"}
	c.Reach = reachOf(map[string]*schema{"echo": {Type: "object"}, "read": {Type: "object"}}, nil)
	obj := json.RawMessage(`{"type":"object"}`)
	c.Input.Instructions = []instruction{{"goal", "write a goal", obj}, {"delegate", "start a child", obj}}

	var names []string
	for _, tl := range turnConfig(c)["tools"].([]bridge.Tool) {
		names = append(names, tl.Name)
	}
	if strings.Join(names, " ") != "echo goal delegate" {
		t.Fatalf("offered %v", names)
	}
	c.Workspace, c.Scratch = "/work", "/scratch"
	if sp := turnConfig(c)["system_prompt"]; sp != "You are inside Skyra, a cognitive architecture." {
		t.Fatalf("a tree turn's system prompt: %q", sp)
	}
	c.Workspace, c.Scratch = "", ""

	goal := `{"_bridge":"call","id":"f1:mcp:1","name":"goal","args":{"goal":"g"}}`
	echo := `{"_bridge":"call","id":"f1:mcp:2","name":"echo","args":{"text":"hi"}}`
	del := `{"_bridge":"call","id":"f1:mcp:3","name":"delegate","args":{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}}`
	ack := func(id string) string { return `{"_bridge":"result","id":"` + id + `","value":"x"}` }
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc"}`)},
		{To: "claude.exchange", State: "return", Payload: deliver(`{"prompt":"node 0"}`), Value: lines(goal)},
	}
	out, _ := decide(c)
	if !strings.Contains(out, `instruction 0`) || !strings.Contains(out, `f1:mcp:1`) {
		t.Fatalf("the goal is not acknowledged as instruction 0: %s", out)
	}
	c.Steps = append(c.Steps, step{To: "claude.exchange", State: "return", Payload: deliver(ack("f1:mcp:1")), Value: lines(echo)})
	if to, _ := parse(t, mustDecide(t, c)); to != "echo" {
		t.Fatalf("a granted target is not dispatched: %s", to)
	}
	c.Steps = append(c.Steps, step{To: "echo", State: "return", Value: json.RawMessage(`"hi"`)})
	if out := mustDecide(t, c); !strings.Contains(out, `f1:mcp:2`) {
		t.Fatalf("echo's result: %s", out)
	}
	c.Steps = append(c.Steps, step{To: "claude.exchange", State: "return", Payload: deliver(ack("f1:mcp:2")), Value: lines(del)})
	if out := mustDecide(t, c); !strings.Contains(out, `instruction 1`) {
		t.Fatalf("delegate is not acknowledged as instruction 1: %s", out)
	}
	c.Steps = append(c.Steps, step{To: "claude.exchange", State: "return", Payload: deliver(ack("f1:mcp:3")),
		Value: lines(`{"type":"result","subtype":"success","session_id":"s","result":"done"}`)})
	to, payload := parse(t, mustDecide(t, c))
	var reply struct {
		Calls []struct {
			Name string          `json:"name"`
			Args json.RawMessage `json:"args"`
		} `json:"calls"`
	}
	json.Unmarshal(payload, &reply)
	if to != "f1" || len(reply.Calls) != 2 || reply.Calls[0].Name != "goal" || reply.Calls[1].Name != "delegate" ||
		string(reply.Calls[0].Args) != `{"goal":"g"}` {
		t.Fatalf("the reply: %s %s", to, payload)
	}
}

func mustDecide(t *testing.T, c continuation) string {
	t.Helper()
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

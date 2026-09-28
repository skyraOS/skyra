package main

import (
	"encoding/json"
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

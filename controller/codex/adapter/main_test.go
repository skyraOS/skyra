package main

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"testing"
)

func TestInterruptWaitsForExactTerminalTurn(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	h := &host{conn: client, r: bufio.NewReader(client)}
	done := make(chan error, 1)
	go func() { done <- h.interrupt("frame", "thread", "wanted", 0) }()
	decoder, encoder := json.NewDecoder(server), json.NewEncoder(server)
	var delivery map[string]any
	if err := decoder.Decode(&delivery); err != nil {
		t.Fatal(err)
	}
	line := delivery["line"].(map[string]any)
	if delivery["id"] != "frame:interrupt" || line["method"] != "turn/interrupt" || line["params"].(map[string]any)["turnId"] != "wanted" {
		t.Fatalf("wrong interrupt: %v", delivery)
	}
	encoder.Encode(map[string]any{"seq": 0})
	for i, turn := range []string{"other", "wanted"} {
		var read map[string]any
		if err := decoder.Decode(&read); err != nil {
			t.Fatal(err)
		}
		if read["op"] != "read" {
			t.Fatal(read)
		}
		encoder.Encode(map[string]any{"next": i + 1, "events": []any{map[string]any{"seq": i, "line": map[string]any{"method": "turn/completed", "params": map[string]any{"threadId": "thread", "turn": map[string]string{"id": turn, "status": "interrupted"}}}}}})
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestInterruptFailsWhenHostDies(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	h := &host{conn: client, r: bufio.NewReader(client)}
	done := make(chan error, 1)
	go func() { done <- h.interrupt("frame", "thread", "turn", 0) }()
	decoder, encoder := json.NewDecoder(server), json.NewEncoder(server)
	var request map[string]any
	decoder.Decode(&request)
	encoder.Encode(map[string]any{"seq": 0})
	decoder.Decode(&request)
	encoder.Encode(map[string]any{"next": 1, "events": []any{map[string]any{"host": "actor exited"}}})
	if err := <-done; err == nil {
		t.Fatal("unconfirmed interruption reported success")
	}
}

// Under a shared Boundary the adapter sees every registered surface in reach. It must
// offer Codex only its model-facing tools, and produce the same thread/start
// bytes on every run: a resumed round delivers the line again under the same
// id, and the host refuses one that differs from the first delivery.
func TestThreadStartOffersOnlyModelToolsInAFixedOrder(t *testing.T) {
	object := &schema{Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}}
	c := continuation{Frame: "f", Reach: reachOf(map[string]*schema{
		"echo": object, "agent": object, "codex": object, "system.narrow": {Type: "object"},
		"z-tool": object, "a-tool": object,
	}, nil)}
	first, _ := json.Marshal(request("f:thread", "thread/start", map[string]any{"dynamicTools": dynamicTools(c)}))
	for i := 0; i < 200; i++ {
		again, _ := json.Marshal(request("f:thread", "thread/start", map[string]any{"dynamicTools": dynamicTools(c)}))
		if string(again) != string(first) {
			t.Fatalf("thread/start differs between runs:\n%s\n%s", first, again)
		}
	}
	var tools []struct{ Name string }
	json.Unmarshal(mustMarshal(dynamicTools(c)), &tools)
	if len(tools) != 2 || tools[0].Name != "agent" || tools[1].Name != "echo" {
		t.Fatalf("only model-facing tools are offered: %+v", tools)
	}
	for _, name := range []string{"codex", "system.narrow", "a-tool"} {
		if offered(c, name) {
			t.Fatalf("%s is in reach but must not be offered to the model", name)
		}
	}
	if !offered(c, "echo") || offered(continuation{Reach: reachOf(map[string]*schema{}, nil)}, "echo") {
		t.Fatal("echo is offered only when it is in reach")
	}
}

func mustMarshal(v any) []byte { b, _ := json.Marshal(v); return b }

// With the apprentice in reach (next/32), Codex is offered remember, which
// names one of its own calls by index, procedure, and one tool per procedure
// that exists, by title. The adapter, a worker, turns remember into the
// observation's stamp and the contract the call went through.
func TestRememberAndProceduresAreOffered(t *testing.T) {
	object := &schema{Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}}
	remember := &schema{Title: "remember", Type: "object", Required: []string{"remember"}, Fields: map[string]*schema{"remember": {Type: "any"}}}
	look := &schema{Title: "look", Type: "object", Required: []string{"path"}, Fields: map[string]*schema{"path": {Type: "string"}}}
	c := continuation{Frame: "f", Reach: reachOf(map[string]*schema{"echo": object, "agent": object, "apprentice": remember}, map[string][]*schema{"apprentice": {remember, look}})}
	var names []string
	for _, tl := range tools(c) {
		names = append(names, tl.name)
	}
	if strings.Join(names, " ") != "agent echo remember procedure look" {
		t.Fatalf("tools: %v", names)
	}
	c.Steps = append(c.Steps, struct {
		To          string          `json:"to"`
		Payload     json.RawMessage `json:"payload"`
		State       string          `json:"state"`
		Value       json.RawMessage `json:"value"`
		Observation string          `json:"observation"`
		Fault       *struct {
			Kind  string `json:"kind"`
			Cause string `json:"cause"`
		} `json:"fault"`
	}{To: "agent", Payload: json.RawMessage(`{"text":"relay:x"}`), State: "return", Observation: "obs"})
	out, err := toolEmission(c, "remember", json.RawMessage(`{"call":0}`))
	want := `{"payload":{"remember":{"contract":{"schema":{"type":"object","fields":{"text":{"type":"string"}},"required":["text"]},"target":"agent"},"observation":"obs"}},"to":"apprentice"}`
	if err != nil || out != want {
		t.Fatalf("remember:\n got %s\nwant %s", out, want)
	}
	if out, _ := toolEmission(c, "procedure", json.RawMessage(`{"input":{"text":"new"}}`)); out != `{"payload":{"text":"new"},"to":"apprentice"}` {
		t.Fatalf("procedure: %s", out)
	}
	if out, _ := toolEmission(c, "look", json.RawMessage(`{"path":"p"}`)); out != `{"payload":{"path":"p"},"to":"apprentice"}` {
		t.Fatalf("a titled procedure: %s", out)
	}
}

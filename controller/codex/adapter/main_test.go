package main

import (
	"bufio"
	"encoding/json"
	"net"
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

// Under a shared Horizon the adapter sees the whole room in reach. It must
// offer Codex only its model-facing tools, and produce the same thread/start
// bytes on every run: a resumed round delivers the line again under the same
// id, and the host refuses one that differs from the first delivery.
func TestThreadStartOffersOnlyModelToolsInAFixedOrder(t *testing.T) {
	object := &schema{Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}}
	c := continuation{Frame: "f", Tools: map[string]*schema{
		"echo": object, "agent": object, "codex": object, "system.narrow": {Type: "object"},
		"z-tool": object, "a-tool": object,
	}}
	first, _ := json.Marshal(request("f:thread", "thread/start", map[string]any{"dynamicTools": dynamicTools(c)}))
	for i := 0; i < 200; i++ {
		again, _ := json.Marshal(request("f:thread", "thread/start", map[string]any{"dynamicTools": dynamicTools(c)}))
		if string(again) != string(first) {
			t.Fatalf("thread/start differs between runs:\n%s\n%s", first, again)
		}
	}
	var tools []struct{ Name string }
	json.Unmarshal(mustMarshal(dynamicTools(c)), &tools)
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("only model-facing tools are offered: %+v", tools)
	}
	for _, name := range []string{"agent", "codex", "system.narrow", "a-tool"} {
		if offered(c, name) {
			t.Fatalf("%s is in reach but must not be offered to the model", name)
		}
	}
	if !offered(c, "echo") || offered(continuation{Tools: map[string]*schema{}}, "echo") {
		t.Fatal("echo is offered only when it is in reach")
	}
}

func mustMarshal(v any) []byte { b, _ := json.Marshal(v); return b }

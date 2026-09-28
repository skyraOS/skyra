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

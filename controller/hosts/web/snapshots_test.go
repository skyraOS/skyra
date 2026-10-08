package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSnapshotStreamChangesAndDeletion(t *testing.T) {
	s := snapshotStream{}
	view := func(id, phase string) map[string]any {
		return map[string]any{"conversation": id, "phase": phase, "entries": []entry{{Role: "agent", Text: strings.Repeat("history", 1000)}}}
	}
	a, b := view("a", "ready"), view("b", "ready")
	frame := func(revision int, views ...map[string]any) map[string]json.RawMessage {
		snapshot := map[string]any{"type": "snapshot", "uiIncarnation": "i", "revision": revision, "conversations": views, "requests": []any{}, "destinations": []any{}}
		full, _ := json.Marshal(snapshot)
		raw := s.frame(snapshot, views, full)
		var decoded map[string]json.RawMessage
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	}
	first := frame(1, a, b)
	if string(first["type"]) != `"snapshot"` {
		t.Fatal("first frame must be full")
	}
	b = view("b", "pending")
	next := frame(2, b, a)
	if string(next["type"]) != `"snapshot-delta"` || string(next["baseRevision"]) != "1" {
		t.Fatal("missing delta baseline")
	}
	var changes []map[string]any
	json.Unmarshal(next["conversations"], &changes)
	if len(changes) != 1 || changes[0]["conversation"] != "b" {
		t.Fatal("unchanged history was resent")
	}
	if string(next["order"]) != `["b","a"]` {
		t.Fatal("conversation order lost")
	}
	deleted := frame(3, a)
	if string(deleted["order"]) != `["a"]` || string(deleted["conversations"]) != `[]` {
		t.Fatal("deletion must preserve unchanged history")
	}
	if string(deleted["baseRevision"]) != "2" {
		t.Fatal("baseline did not advance")
	}
	// Reconnect/resync resets the connection's stream.
	s = snapshotStream{}
	if string(frame(4, a)["type"]) != `"snapshot"` {
		t.Fatal("resync was not full")
	}
}

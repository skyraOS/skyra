package main

import (
	"encoding/json"
	"testing"

	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// commit stores in as an observed input and commits r under it, after prev.
func commit(t *testing.T, root, prev string, in string, r record.Record) string {
	t.Helper()
	h, err := tree.Put(root, "inputs", []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	r.Predecessor, r.Input = prev, h
	s, err := tree.Commit(root, r)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCallsFromTheLineage(t *testing.T) {
	root := t.TempDir()
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	fault := &record.FaultInfo{Kind: record.Horizon, Cause: `"native.commandExecution" is not in reach`}
	s := commit(t, root, "", "genesis", record.Record{State: record.Invoke})
	s = commit(t, root, s, "1", record.Record{State: record.Invoke, To: "codex", Payload: json.RawMessage(`{"text":"read it"}`)})
	s = commit(t, root, s, "2", record.Record{State: record.Invoke, To: "codex.exchange", Payload: json.RawMessage(`{}`)})
	s = commit(t, root, s, "3", record.Record{State: record.Return, Payload: json.RawMessage(`{}`)})
	read := commit(t, root, s, "4", record.Record{State: record.Invoke, To: "read", Payload: json.RawMessage(`{"path":"README.md"}`)})

	l := newLineage(root)
	if !l.refresh() {
		t.Fatal("first read reported no change")
	}
	turns := l.snapshot()
	if len(turns) != 1 || len(turns[0]) != 1 || turns[0][0].State != "running" || turns[0][0].Arg != "README.md" {
		t.Fatalf("a call in flight: %+v", turns)
	}

	s = commit(t, root, read, "5", record.Record{State: record.Return, Payload: json.RawMessage(`{"text":"# Skyra"}`)})
	s = commit(t, root, s, "6", record.Record{State: record.Invoke, To: "codex", Payload: json.RawMessage(`{"text":"run ls"}`)})
	commit(t, root, s, `{"payload":{},"to":"native.commandExecution"}`, record.Record{State: record.Fault, Fault: fault})
	if !l.refresh() {
		t.Fatal("new records reported no change")
	}
	if l.refresh() {
		t.Fatal("no new records reported a change")
	}
	turns = l.snapshot()
	if len(turns) != 2 {
		t.Fatalf("turns: %+v", turns)
	}
	if c := turns[0][0]; c.State != "returned" || c.Result != "returned 18 B" || c.Stamp != read {
		t.Fatalf("returned call: %+v", c)
	}
	if c := turns[1][0]; c.State != "fault" || c.To != "native.commandExecution" || c.Fault != record.Horizon {
		t.Fatalf("refused call: %+v", c)
	}

	entries := withCalls([]entry{
		{Role: "user", Text: "read it"},
		{Role: "agent", Text: "It says Skyra."},
		{Role: "user", Text: "run ls"},
		{Role: "agent", Text: "⚠ the agent faulted (horizon): " + fault.Cause},
		{Role: "agent", Text: "I can't run commands here."},
	}, turns)
	var got []string
	for i, e := range entries {
		if e.TS != int64(i) {
			t.Fatalf("entry %d stamped %d", i, e.TS)
		}
		got = append(got, e.Role+" "+e.Text)
	}
	want := []string{"user read it", "call read README.md", "agent It says Skyra.", "user run ls", "call native.commandExecution", "agent I can't run commands here."}
	if len(got) != len(want) {
		t.Fatalf("entries:\n%q\nwant\n%q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entries:\n%q\nwant\n%q", got, want)
		}
	}
}

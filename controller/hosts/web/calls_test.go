package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

func TestCallsAcrossConversationRuns(t *testing.T) {
	home := filepath.Join(t.TempDir(), "c-continued")
	first := filepath.Join(home, ".skyra")
	second := filepath.Join(home, "runs", "2", ".skyra")
	third := filepath.Join(home, "runs", "10", ".skyra")
	history := []entry{{Role: "user", Text: "first"}, {Role: "agent", Text: "done"}}
	seedRun := func(root string, prior []entry, tool string) string {
		t.Helper()
		if err := tree.Init(root); err != nil {
			t.Fatal(err)
		}
		s := commit(t, root, "", "genesis", record.Record{State: record.Invoke})
		payload, err := json.Marshal(map[string]any{"conversation": "c-continued", "procedure": "think", "transcript": prior})
		if err != nil {
			t.Fatal(err)
		}
		s = commit(t, root, s, "user", record.Record{State: record.Invoke, To: "user", Payload: payload})
		s = commit(t, root, s, "turn", record.Record{State: record.Invoke, To: "interpreter"})
		return commit(t, root, s, "tool", record.Record{State: record.Invoke, To: tool, Payload: json.RawMessage(`{"path":"file"}`)})
	}
	finish := func(root, invoke string, transcript []entry) {
		t.Helper()
		s := commit(t, root, invoke, "returned", record.Record{State: record.Return, Payload: json.RawMessage(`{}`)})
		payload, err := json.Marshal(map[string]any{"conversation": "c-continued", "transcript": transcript})
		if err != nil {
			t.Fatal(err)
		}
		commit(t, root, s, "browser", record.Record{State: record.Invoke, To: "browser", Payload: payload})
	}
	read := seedRun(first, nil, "read")
	finish(first, read, history)
	c := &conversation{lineage: newLineage(first), entries: history}
	c.startLineage(second)
	c.entries = append(c.entries, entry{Role: "user", Text: "second"})
	assertCalls := func(c *conversation, states ...string) {
		t.Helper()
		entries := withCalls(c.entries, c.calls())
		user, calls := 0, 0
		for _, e := range entries {
			if e.Role == "user" {
				user++
			}
			if e.Call == nil {
				continue
			}
			tools := []string{"read", "write", "find"}
			if calls >= len(states) || user != calls+1 || e.Call.State != states[calls] || e.Call.To != tools[calls] {
				t.Fatalf("call %d under message %d: %+v; want states %v", calls, user, e.Call, states)
			}
			calls++
		}
		if calls != len(states) {
			t.Fatalf("got %d calls, want %d", calls, len(states))
		}
	}
	assertCalls(c, "returned") // the new run has not committed anything yet
	write := seedRun(second, history, "write")
	c.lineage.refresh()
	assertCalls(c, "returned", "running")
	c.entries = append(c.entries, entry{Role: "agent", Text: "done again"})
	finish(second, write, c.entries)
	c.lineage.refresh()
	assertCalls(c, "returned", "returned")
	prior := append([]entry(nil), c.entries...)
	c.startLineage(third)
	find := seedRun(third, prior, "find")
	c.entries = append(c.entries, entry{Role: "user", Text: "third"}, entry{Role: "agent", Text: "done finally"})
	finish(third, find, c.entries)
	c.lineage.refresh()
	assertCalls(c, "returned", "returned", "returned")
	h := &webhost{log: os.Stderr, convos: map[string]*conversation{}}
	h.loadConversation(home, "think")
	loaded := h.convos["c-continued"]
	if loaded == nil {
		t.Fatal("conversation not loaded")
	}
	h.loadConversationFull("c-continued")
	assertCalls(loaded, "returned", "returned", "returned")
}

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
	// The user side starts each turn: here by running a procedure, which
	// reaches its agent and tools beneath it (proposal 38).
	user := commit(t, root, s, "user", record.Record{State: record.Invoke, To: "user", Payload: json.RawMessage(`{"conversation":"c"}`)})
	s = commit(t, root, user, "1", record.Record{State: record.Invoke, To: "interpreter", Payload: json.RawMessage(`{"think":{"text":"read it","context":[]}}`)})
	s = commit(t, root, s, "1a", record.Record{State: record.Invoke, To: "system.ref.1", Payload: json.RawMessage(`{"text":"read it"}`)})
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
	s = commit(t, root, s, "5r", record.Record{State: record.Resolve, Frame: user}) // the user's frame resumes
	s = commit(t, root, s, "6", record.Record{State: record.Invoke, To: "interpreter", Payload: json.RawMessage(`{"think":{"text":"run ls","context":[]}}`)})
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

// A successor that commits and then dies before its stamp reaches the parent
// leaves a record that was never released. The parent faults its own frame
// under the same predecessor; the call ends in that fault, not in the orphan.
func TestUnreleasedSuccessorIsLeftOut(t *testing.T) {
	root := t.TempDir()
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	s := commit(t, root, "", "genesis", record.Record{State: record.Invoke})
	s = commit(t, root, s, "user", record.Record{State: record.Invoke, To: "user", Payload: json.RawMessage(`{"conversation":"c"}`)})
	s = commit(t, root, s, "1", record.Record{State: record.Invoke, To: "interpreter", Payload: json.RawMessage(`{"think":{"text":"read it","context":[]}}`)})
	read := commit(t, root, s, "2", record.Record{State: record.Invoke, To: "read", Payload: json.RawMessage(`{"path":"README.md"}`)})
	orphan := commit(t, root, read, "3", record.Record{State: record.Return, Payload: json.RawMessage(`{"text":"# Skyra"}`)})
	obs := "successor died before stamping: EOF"
	f := commit(t, root, read, obs, machine.Observed(read, []byte(obs)))

	l := newLineage(root)
	l.refresh()
	turns := l.snapshot()
	if len(turns) != 1 || len(turns[0]) != 1 {
		t.Fatalf("turns: %+v", turns)
	}
	if c := turns[0][0]; c.State != "fault" || c.Fault != record.Custody || c.Stamp != read {
		t.Fatalf("the call must end in the supervisor's fault: %+v", c)
	}
	for _, s := range l.order() {
		if s == orphan {
			t.Fatal("the unreleased successor is in the lineage")
		}
	}
	if o := l.order(); o[len(o)-1] != f {
		t.Fatalf("the lineage must end in the fault: %v", o)
	}
}

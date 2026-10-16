package replay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

var echoS = horizon.Surface{Path: "/bin/echo-target", SHA256: "e",
	Input: &contract.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*contract.Schema{"text": {Type: "string"}}}}

// intakeRecord runs the Intake step and returns what it commits.
func intakeRecord(h horizon.Horizon, in string) record.Record {
	_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: machine.HeldHorizon(h), In: []byte(in)})
	return machine.CommittedRecord(effs)
}

// buildLineage commits genesis → invoke echo → return without any processes,
// using step exactly as intake does, and returns the head stamp.
func buildLineage(t *testing.T, root string) string {
	t.Helper()
	for _, d := range []string{"records", "inputs"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	g0, err := tree.Commit(root, record.Record{State: record.Invoke, Reach: map[string]horizon.Surface{"echo": echoS}})
	if err != nil {
		t.Fatal(err)
	}
	h := horizon.Horizon{Stamp: g0, Reach: map[string]horizon.Surface{"echo": echoS}, Frame: g0}
	var head string
	for _, in := range []func() string{
		func() string { return `{"to":"echo","payload":{"text":"hi"}}` },
		func() string { return `{"to":"` + h.Frame + `","payload":"hi"}` }, // echo returns to its slot
	} {
		in := in()
		if _, err := tree.Put(root, "inputs", []byte(in)); err != nil {
			t.Fatal(err)
		}
		r := intakeRecord(h, in)
		s, err := tree.Commit(root, r)
		if err != nil {
			t.Fatal(err)
		}
		if h, err = machine.Narrow(h, r, s); err != nil {
			t.Fatal(err)
		}
		head = s
	}
	return head
}

func TestReplayReproducesLineage(t *testing.T) {
	root := t.TempDir()
	head := buildLineage(t, root)
	n, err := Replay(root, head)
	if err != nil || n != 3 {
		t.Fatalf("replay = %d, %v; want 3 records, no error", n, err)
	}
}

func TestReplayCatchesForgedRecord(t *testing.T) {
	root := t.TempDir()
	head := buildLineage(t, root)
	g, _ := tree.Load(root, head)

	// Forge a successor to frame 1 claiming echo invoked something outside its
	// Horizon. It is a well-formed, correctly hashed record, but step would
	// never derive it: it derives a horizon fault.
	in := []byte(`{"to":"shell","payload":{"text":"x"}}`)
	tree.Put(root, "inputs", in)
	forged, err := tree.Commit(root, record.Record{Predecessor: g.Predecessor, Input: record.Sum(in), State: record.Invoke,
		To: "shell", Payload: []byte(`{"text":"x"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Replay(root, forged); err == nil || !strings.Contains(err.Error(), "not what step derives") {
		t.Fatalf("replay accepted a forged record: %v", err)
	}
}

func TestReplayCatchesTamperedInput(t *testing.T) {
	root := t.TempDir()
	head := buildLineage(t, root)
	r, _ := tree.Load(root, head)
	p := filepath.Join(root, "inputs", r.Input)
	os.Chmod(p, 0o644)
	if err := os.WriteFile(p, []byte(`tampered`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Replay(root, head); err == nil || !strings.Contains(err.Error(), "does not hash") {
		t.Fatalf("replay accepted a tampered input: %v", err)
	}
}

// A supervisor's fault is derived from its stored observation too: forging
// its cause is caught.
func TestReplayDerivesSupervisorFaults(t *testing.T) {
	root := t.TempDir()
	head := buildLineage(t, root)
	r, _ := tree.Load(root, head)
	invoke := r.Predecessor
	obs := []byte("target exited without emitting")
	tree.Put(root, "inputs", obs)
	real, _ := tree.Commit(root, machine.Observed(invoke, obs))
	if n, err := Replay(root, real); err != nil || n != 3 {
		t.Fatalf("replay of a supervisor fault = %d, %v", n, err)
	}
	forged := machine.Observed(invoke, obs)
	forged.Fault.Cause = "something else"
	f, _ := tree.Commit(root, forged)
	if _, err := Replay(root, f); err == nil {
		t.Fatal("replay accepted a supervisor fault whose cause is not its observation")
	}
}

var agentS = horizon.Surface{Path: "/bin/agent", SHA256: "a", Input: echoS.Input, Continuation: true}

// commitStep stores an input and commits the record derived from it.
func commitStep(t *testing.T, root string, in []byte, r record.Record) string {
	t.Helper()
	if _, err := tree.Put(root, "inputs", in); err != nil {
		t.Fatal(err)
	}
	s, err := tree.Commit(root, r)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// buildContinuation commits the agent's lineage as the runtime would:
// operator → agent → echo → echo's reply → resolve agent → agent's reply →
// resolve operator. It returns the head and the stamps along the way.
func buildContinuation(t *testing.T, root string) (head, echoReply, agentReply string) {
	t.Helper()
	for _, d := range []string{"records", "inputs"} {
		os.MkdirAll(filepath.Join(root, d), 0o755)
	}
	reach := map[string]horizon.Surface{"echo": echoS, "agent": agentS}
	g0, _ := tree.Commit(root, record.Record{State: record.Invoke, Reach: reach})
	h := horizon.Horizon{Stamp: g0, Reach: reach, Frame: g0}

	emit := func(under horizon.Horizon, out string) (string, horizon.Horizon) {
		r := intakeRecord(under, out)
		s := commitStep(t, root, []byte(out), r)
		next, err := machine.Narrow(under, r, s)
		if err != nil {
			t.Fatal(err)
		}
		return s, next
	}
	resolve := func(discharged horizon.Horizon, discharge string) (string, horizon.Horizon) {
		w := discharged.Resume
		obs, _ := json.Marshal(machine.Observation{Frame: w.Frame, Round: w.Stamp, Discharges: []string{discharge}})
		r, _ := machine.Resolved(discharge, obs)
		s := commitStep(t, root, obs, r)
		resumed := *w
		resumed.Stamp = s
		return s, resumed
	}

	_, agent := emit(h, `{"to":"agent","payload":{"text":"hi"}}`)
	_, echo := emit(agent, `{"to":"echo","payload":{"text":"hi"}}`)
	echoReply, done := emit(echo, `hi`)
	_, agent = resolve(done, echoReply)
	agentReply, done = emit(agent, `echo said "hi"`)
	head, _ = resolve(done, agentReply)
	return head, echoReply, agentReply
}

func TestReplayThroughResolves(t *testing.T) {
	root := t.TempDir()
	head, _, _ := buildContinuation(t, root)
	if n, err := Replay(root, head); err != nil || n != 7 {
		t.Fatalf("replay = %d, %v; want 7 records", n, err)
	}
}

// A resolve that claims to resume a frame that was not waiting is caught,
// even though it hashes correctly.
func TestReplayCatchesForgedResolve(t *testing.T) {
	root := t.TempDir()
	_, echoReply, _ := buildContinuation(t, root)
	obs, _ := json.Marshal(machine.Observation{Frame: "someone-else", Round: "someone-else", Discharges: []string{echoReply}})
	r, _ := machine.Resolved(echoReply, obs)
	forged := commitStep(t, root, obs, r)
	if _, err := Replay(root, forged); err == nil || !strings.Contains(err.Error(), "was not waiting") {
		t.Fatalf("replay accepted a resolve of a frame that was not waiting: %v", err)
	}
}

// memory is a lineage handed over whole, as the website hands it to the browser.
type memory map[string][]byte

func (m memory) Get(kind, h string) ([]byte, error) {
	b, ok := m[kind+"/"+h]
	if !ok {
		return nil, os.ErrNotExist
	}
	return b, nil
}

func TestLineageReplaysAStoreInMemory(t *testing.T) {
	root := t.TempDir()
	head := buildLineage(t, root)
	m := memory{}
	for _, kind := range []string{"records", "inputs"} {
		entries, _ := os.ReadDir(filepath.Join(root, kind))
		for _, e := range entries {
			b, _ := os.ReadFile(filepath.Join(root, kind, e.Name()))
			m[kind+"/"+e.Name()] = b
		}
	}
	if n, err := Lineage(m, head); err != nil || n != 3 {
		t.Fatalf("lineage = %d, %v; want 3 records, no error", n, err)
	}

	// One byte changed in a stored input: it no longer hashes to its name.
	r, _ := tree.Load(root, head)
	in := append([]byte(nil), m["inputs/"+r.Input]...)
	in[len(in)-2] ^= 1
	m["inputs/"+r.Input] = in
	if _, err := Lineage(m, head); err == nil || !strings.Contains(err.Error(), "does not hash to its name") {
		t.Fatalf("lineage accepted a changed input: %v", err)
	}
}

// A double winner (41, V7): two resolves of one round that both resumed the
// waiter are reported, whatever the claim said; a resolve that lost the
// claim, with no successor, is not.
func TestARoundIsResumedOnce(t *testing.T) {
	a := record.Record{Predecessor: "d1", State: record.Resolve, Frame: "f", Round: "f", Deps: []string{"d1", "d2"}}
	b := record.Record{Predecessor: "d2", State: record.Resolve, Frame: "f", Round: "f", Deps: []string{"d1", "d2"}}
	sa, sb := record.Stamp(a), record.Stamp(b)
	next := func(after string) record.Record {
		return record.Record{Predecessor: after, State: record.Return, Payload: []byte(`"x"`)}
	}
	tree := map[string]record.Record{sa: a, sb: b, "n1": next(sa)}
	if err := ResumedOnce(tree); err != nil {
		t.Fatalf("one resolve resumed, the other lost the claim: %v", err)
	}
	tree["n2"] = next(sb)
	if err := ResumedOnce(tree); err == nil || !strings.Contains(err.Error(), "resumed twice") {
		t.Fatalf("a second resume must be reported: %v", err)
	}
	tree["n2"] = next(sa)
	if err := ResumedOnce(tree); err != nil {
		t.Fatalf("one resolve with two successors (a fan-out after resuming) is one resume: %v", err)
	}
}

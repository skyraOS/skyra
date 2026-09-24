package replay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra-v1/specs/2026-09-23/minimal-shape/horizon"
	"skyra-v1/specs/2026-09-23/minimal-shape/machine"
	"skyra-v1/specs/2026-09-23/minimal-shape/move"
	"skyra-v1/specs/2026-09-23/minimal-shape/record"
	"skyra-v1/specs/2026-09-23/minimal-shape/tree"
)

var echoS = horizon.Surface{Path: "/bin/echo-target", SHA256: "e", Leaf: true}

// intakeRecord runs the Intake step and returns what it commits.
func intakeRecord(h horizon.Horizon, in string) record.Record {
	_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: h, In: []byte(in)})
	return machine.Committed(effs)
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
	g0, err := tree.Commit(root, record.Record{State: move.StateInvoke, Reach: map[string]horizon.Surface{"echo": echoS}})
	if err != nil {
		t.Fatal(err)
	}
	h := horizon.Horizon{Stamp: g0, Reach: map[string]horizon.Surface{"echo": echoS}}
	var head string
	for _, in := range []string{
		`{"invoke":{"target":"echo","args":{"text":"hi"}}}`,
		`{"return":{"value":"hi"}}`,
	} {
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

	// Forge a successor to frame 1 claiming echo (a leaf) was allowed to invoke.
	// It is a well-formed, correctly hashed record, but step would never produce it.
	in := []byte(`{"invoke":{"target":"echo","args":{}}}`)
	tree.Put(root, "inputs", in)
	forged, err := tree.Commit(root, record.Record{Predecessor: g.Predecessor, Input: record.Sum(in), State: move.StateInvoke,
		Move: &move.Move{Invoke: &move.Invoke{Target: "echo", Args: map[string]string{}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Replay(root, forged); err == nil || !strings.Contains(err.Error(), "not what step produces") {
		t.Fatalf("replay accepted a forged record: %v", err)
	}
}

func TestReplayCatchesTamperedInput(t *testing.T) {
	root := t.TempDir()
	head := buildLineage(t, root)
	r, _ := tree.Load(root, head)
	p := filepath.Join(root, "inputs", r.Input)
	os.Chmod(p, 0o644)
	if err := os.WriteFile(p, []byte(`{"return":{"value":"tampered"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Replay(root, head); err == nil || !strings.Contains(err.Error(), "does not hash") {
		t.Fatalf("replay accepted a tampered input: %v", err)
	}
}

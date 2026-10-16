//go:build darwin

package resident

import (
	"encoding/json"
	"testing"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// journal exercises the existing intake/narrow/resolve/replay vocabulary
// around the transport experiment. It does not add resident semantics to the
// machine or pretend to be the production Drive loop.
type journal struct {
	t    *testing.T
	root string
	h    horizon.Horizon
	wait horizon.Horizon
	head string
}

func newJournal(t *testing.T) *journal {
	t.Helper()
	str := &contract.Schema{Type: "string"}
	num := &contract.Schema{Type: "number"}
	arr := &contract.Schema{Type: "array", Items: str}
	input := &contract.Schema{Type: "object", Required: []string{"op", "owner", "incarnation", "deliver", "read_lines", "wait_ms"}, Fields: map[string]*contract.Schema{
		"op": str, "owner": str, "incarnation": str, "deliver": arr, "read_lines": num, "wait_ms": num, "prefix": str,
	}}
	reach := map[string]horizon.Surface{}
	ops := []string{"start", "exchange", "tool", "park", "stop"}
	for _, name := range ops {
		reach[name] = horizon.Surface{Input: input}
	}
	reach["adapter"] = horizon.Surface{Input: str, Continuation: true}
	j := &journal{t: t, root: t.TempDir()}
	if err := tree.Init(j.root); err != nil {
		t.Fatal(err)
	}
	g := record.Record{State: record.Invoke, Reach: reach}
	stamp, err := tree.Commit(j.root, g)
	if err != nil {
		t.Fatal(err)
	}
	j.h = horizon.Horizon{Stamp: stamp, Frame: stamp, Reach: reach}
	j.intake(map[string]any{"to": "adapter", "payload": "custody experiment"}, record.Invoke)
	return j
}

func (j *journal) intake(v any, want record.State) {
	j.t.Helper()
	in, err := json.Marshal(v)
	if err != nil {
		j.t.Fatal(err)
	}
	_, effects := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: machine.HeldHorizon(j.h), In: in})
	r := machine.Committed(effects)
	if r.State != want {
		j.t.Fatalf("intake derived %+v, want %s", r, want)
	}
	if _, err := tree.Put(j.root, "inputs", in); err != nil {
		j.t.Fatal(err)
	}
	j.head, err = tree.Commit(j.root, r)
	if err != nil {
		j.t.Fatal(err)
	}
	j.h, err = machine.Narrow(j.h, r, j.head)
	if err != nil {
		j.t.Fatal(err)
	}
}

func (j *journal) invoke(q request) {
	j.wait = j.h
	if q.Deliver == nil {
		q.Deliver = []string{}
	}
	j.intake(map[string]any{"to": q.Op, "payload": q}, record.Invoke)
}

func (j *journal) returned(o observation) {
	if o.Lines == nil {
		o.Lines = []string{}
	}
	j.intake(map[string]any{"to": j.h.Frame, "payload": o}, record.Return)
	in, _ := json.Marshal(machine.Observation{Frame: j.wait.Frame, Round: j.wait.Stamp, Discharges: []string{j.head}})
	r, err := machine.Resolved(j.head, in)
	if err != nil {
		j.t.Fatal(err)
	}
	if _, err := tree.Put(j.root, "inputs", in); err != nil {
		j.t.Fatal(err)
	}
	j.head, err = tree.Commit(j.root, r)
	if err != nil {
		j.t.Fatal(err)
	}
	j.h = j.wait
	j.h.Stamp = j.head
}

func (j *journal) verify() {
	j.intake(map[string]any{"to": j.h.Frame, "payload": "complete"}, record.Return)
	n, err := replay.Replay(j.root, j.head)
	if err != nil {
		j.t.Fatal(err)
	}
	j.t.Logf("%d records re-derived using existing runtime states; observations include recorded time", n)
}

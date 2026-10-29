package replay

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
	"skyra/controller/targets/user/checkpoint"
	"skyra/controller/targets/user/projection"
)

type treeFixture struct {
	t      *testing.T
	dir    string
	root   boundary.Boundary
	schema *contract.Schema
}

func newTreeFixture(t *testing.T) *treeFixture {
	t.Helper()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(cwd) })
	dir := t.TempDir()
	if err := tree.Init(dir); err != nil {
		t.Fatal(err)
	}
	schema := &contract.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*contract.Schema{"text": {Type: "string"}}}
	reach := map[string]boundary.Surface{"agent": {Path: "/agent", Input: schema, Continuation: true}, "echo": {Path: "/echo", Input: schema}, "system": machine.Syscalls()}
	g := record.Record{State: record.Invoke, Reach: reach, Authority: &boundary.Authority{}, Author: "test", Nonce: t.Name()}
	gs, err := tree.Commit(dir, g)
	if err != nil {
		t.Fatal(err)
	}
	f, err := tree.OpenFrame(dir, gs)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	fx := &treeFixture{t: t, dir: dir, schema: schema}
	operator := boundary.Boundary{Stamp: gs, Frame: gs, Reach: reach}
	fx.root = fx.emit(operator, `{"to":"agent","payload":{"text":"root"}}`, false, 0, true)
	return fx
}
func (f *treeFixture) emit(under boundary.Boundary, in string, fan bool, pos int, accept bool) boundary.Boundary {
	f.t.Helper()
	_, effects := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: machine.HeldBoundary(under), In: []byte(in)})
	r := machine.CommittedRecord(effects)
	if _, err := tree.Put(f.dir, "inputs", []byte(in)); err != nil {
		f.t.Fatal(err)
	}
	stamp, err := tree.Commit(f.dir, r)
	if err != nil {
		f.t.Fatal(err)
	}
	h, err := machine.Narrow(under, r, stamp)
	if fan {
		h, err = machine.Fanned(under, r, stamp)
	}
	if err != nil {
		f.t.Fatal(err)
	}
	if h.Frame == stamp {
		p, err := os.Open(framePath(f.dir, under.Frame))
		if err != nil {
			f.t.Fatal(err)
		}
		won, err := tree.Slot(p, under.Stamp, stamp, pos)
		p.Close()
		if err != nil || !won {
			f.t.Fatal("slot", won, err)
		}
		child, err := tree.OpenFrame(f.dir, stamp)
		if err != nil {
			f.t.Fatal(err)
		}
		child.Close()
	}
	id, err := tree.PutBoundary(f.dir, h)
	if err != nil {
		f.t.Fatal(err)
	}
	if accept {
		if _, err := tree.Receipt(f.dir, stamp, id); err != nil {
			f.t.Fatal(err)
		}
	}
	return h
}
func (f *treeFixture) settle(h boundary.Boundary) {
	f.t.Helper()
	fd, err := os.Open(framePath(f.dir, h.Frame))
	if err != nil {
		f.t.Fatal(err)
	}
	won, err := tree.Discharge(fd, h.Stamp)
	fd.Close()
	if err != nil || !won {
		f.t.Fatal("discharge", won, err)
	}
}
func (f *treeFixture) read(leaves ...string) (RecordedTree, error) {
	return ReadTree(f.dir, f.root.Frame, leaves, "agent", f.schema, "")
}
func (f *treeFixture) join(predecessor string, deps []string, win bool) string {
	return f.joinUnder(f.root, predecessor, deps, win)
}
func (f *treeFixture) joinUnder(parent boundary.Boundary, predecessor string, deps []string, win bool) string {
	f.t.Helper()
	b, _ := json.Marshal(machine.Observation{Frame: parent.Frame, Round: parent.Stamp, Discharges: deps})
	r, err := machine.Resolved(predecessor, b)
	if err != nil {
		f.t.Fatal(err)
	}
	tree.Put(f.dir, "inputs", b)
	s, err := tree.Commit(f.dir, r)
	if err != nil {
		f.t.Fatal(err)
	}
	if win {
		fd, _ := os.Open(framePath(f.dir, parent.Frame))
		won, err := tree.Claim(fd, parent.Stamp, s)
		fd.Close()
		if err != nil || !won {
			f.t.Fatalf("claim: %v %v", won, err)
		}
	}
	return s
}
func TestRecordedTreeInitializationAndValidation(t *testing.T) {
	f := newTreeFixture(t)
	tr, err := f.read(f.root.Frame)
	if err != nil || len(tr.Records) != 1 || tr.Version != 1 {
		t.Fatalf("initial tree: %+v %v", tr, err)
	}
	if _, err := f.read(); !Refused(err) {
		t.Fatalf("empty frontier: %v", err)
	}
	if _, err := f.read(strings.Repeat("../", 22)); !Refused(err) {
		t.Fatalf("invalid leaf: %v", err)
	}
	if _, err := ReadTree(f.dir, f.root.Frame, []string{f.root.Frame}, "echo", f.schema, ""); !Refused(err) {
		t.Fatalf("wrong anchor target: %v", err)
	}
	if _, err := ReadTree(f.dir, f.root.Frame, []string{f.root.Frame}, "agent", &contract.Schema{Type: "any"}, ""); !Refused(err) {
		t.Fatalf("wrong contract: %v", err)
	}
	if _, err := ReadTree(f.dir, f.root.Frame, []string{f.root.Frame}, "agent", f.schema, "other"); !Refused(err) {
		t.Fatalf("foreign Boundary: %v", err)
	}
}

func TestRecordedTreeCarriesVisibleRefusalEvidenceIncludingMalformedInput(t *testing.T) {
	for _, input := range []string{`{"to":"forbidden","payload":{"text":"attempt"}}`, `not JSON`} {
		t.Run(input, func(t *testing.T) {
			f := newTreeFixture(t)
			fault := f.emit(f.root, input, false, 0, true)
			tr, err := f.read(fault.Stamp)
			if err != nil {
				t.Fatal(err)
			}
			if string(tr.Inputs[fault.Stamp]) != input {
				t.Fatal("visible fault lost its observed input")
			}
			if _, err := json.Marshal(tr); err != nil {
				t.Fatalf("malformed emission broke syscall serialization: %v", err)
			}
		})
	}
}
func TestRecordedTreePartialFanOutAndWinningJoin(t *testing.T) {
	f := newTreeFixture(t)
	a := f.emit(f.root, `{"to":"echo","payload":{"text":"a"}}`, true, 0, true)
	b := f.emit(f.root, `{"to":"echo","payload":{"text":"b"}}`, true, 1, true)
	ar := f.emit(a, `{"to":"`+a.Frame+`","payload":"a"}`, false, 0, true)
	f.settle(ar)
	partial, err := f.read(ar.Stamp, b.Stamp)
	if err != nil {
		t.Fatal(err)
	}
	if len(partial.Records) != 4 || !reflect.DeepEqual(partial.Order, []string{f.root.Stamp, a.Stamp, ar.Stamp, b.Stamp}) {
		t.Fatalf("partial siblings/emission order: %+v", partial)
	}
	br := f.emit(b, `{"to":"`+b.Frame+`","payload":"b"}`, false, 0, true)
	f.settle(br)
	old, err := f.read(ar.Stamp, b.Stamp)
	if err != nil || !reflect.DeepEqual(old, partial) {
		t.Fatalf("retained frontier must stop before newer progress: %v", err)
	}
	join := f.join(br.Stamp, []string{ar.Stamp, br.Stamp}, true)
	tr, err := f.read(join)
	if err != nil || len(tr.Records) != 6 {
		t.Fatalf("joined tree: %+v %v", tr, err)
	}
	losing := f.join(ar.Stamp, []string{ar.Stamp, br.Stamp}, false)
	if _, err := f.read(losing); !Refused(err) {
		t.Fatalf("losing claim: %v", err)
	}
	again, err := f.read(join)
	if err != nil || !reflect.DeepEqual(again, tr) {
		t.Fatal("tree reconstruction must be stable")
	}
}
func TestRecordedTreeRefusesOrphanAndForeignRoot(t *testing.T) {
	f := newTreeFixture(t)
	orphan := f.emit(f.root, `{"to":"echo","payload":{"text":"orphan"}}`, true, 0, false)
	if _, err := f.read(orphan.Stamp); !Refused(err) {
		t.Fatalf("proposed release accepted as work: %v", err)
	}
	operator := *f.root.Resume
	other := f.emit(operator, `{"to":"agent","payload":{"text":"other"}}`, true, 1, true)
	if _, err := f.read(other.Stamp); !Refused(err) {
		t.Fatalf("cross-root read: %v", err)
	}
}
func TestRecordedTreeNarrowedChildOpacity(t *testing.T) {
	f := newTreeFixture(t)
	narrow := f.emit(f.root, `{"to":"system","payload":{"narrow":{"actor":"agent","reach":["agent","echo"]}}}`, false, 0, true)
	nr := f.emit(narrow, `{"to":"`+narrow.Frame+`","payload":{"reference":"system.ref.1"}}`, false, 0, true)
	f.settle(nr)
	// The waiting caller acquired the reference; emulate its winning resolve.
	join := f.join(nr.Stamp, []string{nr.Stamp}, true)
	waiter, err := BoundaryAt(f.dir, join)
	if err != nil {
		t.Fatal(err)
	}
	call := f.emit(waiter, `{"to":"system.ref.1","payload":{"text":"secret"}}`, false, 0, true)
	inner := f.emit(call, `{"to":"echo","payload":{"text":"private"}}`, false, 0, true)
	tr, err := f.read(inner.Stamp)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tr.Records[inner.Stamp]; ok {
		t.Fatal("narrowed internals leaked")
	}
	if tr.Opaque[inner.Stamp] != call.Stamp {
		t.Fatal("open narrowed work must collapse to its boundary")
	}
	ir := f.emit(inner, `{"to":"`+inner.Frame+`","payload":"private result"}`, false, 0, true)
	f.settle(ir)
	joined := f.joinUnder(call, ir.Stamp, []string{ir.Stamp}, true)
	resumed, err := BoundaryAt(f.dir, joined)
	if err != nil {
		t.Fatal(err)
	}
	answer := f.emit(resumed, `{"to":"`+call.Frame+`","payload":"public result"}`, false, 0, true)
	f.settle(answer)
	completed, err := f.read(answer.Stamp)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Opaque[answer.Stamp] != call.Stamp || completed.Records[answer.Stamp].State != record.Return {
		t.Fatal("narrowed boundary result disappeared")
	}
	if _, leaked := completed.Records[joined]; leaked {
		t.Fatal("narrowed internal join leaked")
	}
	if completed.Order[len(completed.Order)-1] != answer.Stamp {
		t.Fatal("boundary result lost its causal position")
	}
	if _, err := ReadTree(f.dir, f.root.Frame, []string{inner.Stamp}, "agent", f.schema, call.Stamp); !Refused(err) {
		t.Fatalf("child walked out of Boundary: %v", err)
	}
}

func TestRecordedTreeCrossGenesisReadRequiresScopedContext(t *testing.T) {
	f := newTreeFixture(t)
	// Two operator boots can occupy one tree directory. An empty Origin is
	// not sufficient to treat the second genesis as the first one's Boundary.
	_, err := ReadTreeInGenesis(f.dir, f.root.Frame, []string{f.root.Frame}, "agent", f.schema, "", strings.Repeat("e", 64))
	if !Refused(err) {
		t.Fatalf("foreign genesis read: %v", err)
	}
	tr, err := ReadTreeInGenesis(f.dir, f.root.Frame, []string{f.root.Frame}, "agent", f.schema, "", f.root.Resume.Frame)
	if err != nil || len(tr.Records) != 1 {
		t.Fatalf("own genesis: %+v %v", tr, err)
	}
}

func TestAcceptedCheckpointCrashGapUsesRecordedEvidence(t *testing.T) {
	f := newTreeFixture(t)
	store := checkpoint.Store{Dir: t.TempDir()}
	schema, _ := json.Marshal(f.schema)
	cp, err := store.Initialize("c-1", f.root.Frame, checkpoint.Contract{Target: "agent", Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	a := f.emit(f.root, `{"to":"echo","payload":{"text":"a"}}`, true, 0, true)
	b := f.emit(f.root, `{"to":"echo","payload":{"text":"b"}}`, true, 1, true)
	// Acceptance is durable, but deliberately no checkpoint update occurred.
	recovered, err := f.read(append(cp.Leaves, a.Stamp, b.Stamp)...)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(recovered)
	var evidence checkpoint.Evidence
	json.Unmarshal(raw, &evidence)
	cp, err = store.Advance("c-1", cp, evidence)
	if err != nil || len(cp.Leaves) != 2 {
		t.Fatalf("accepted siblings lost in crash gap: %+v %v", cp, err)
	}
	ar := f.emit(a, `{"to":"`+a.Frame+`","payload":"a"}`, false, 0, true)
	f.settle(ar)
	partial, err := f.read(ar.Stamp, b.Stamp)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(partial)
	json.Unmarshal(raw, &evidence)
	cp, err = store.Advance("c-1", cp, evidence)
	if err != nil || len(cp.Leaves) != 2 {
		t.Fatalf("slow branch lost: %+v %v", cp, err)
	}
	// Projection consumes exactly the same real syscall result, with no tree
	// reads or duration invention in the actor.
	var projected projection.Tree
	json.Unmarshal(raw, &projected)
	view, err := projection.ConversationState("c-1", cp.RootInvocation, projected, projection.Availability{})
	if err != nil || view.Conversation.Phase != "closed" || len(view.Conversation.Boundaries) == 0 {
		t.Fatalf("recorded projection: %+v %v", view, err)
	}
}

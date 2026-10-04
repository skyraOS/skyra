package replay

import (
	"encoding/json"
	"reflect"
	"testing"

	"skyra/controller/targets/user/checkpoint"
)

func TestRecoveryReconcilesAcceptedProgressWithoutRepeatingWork(t *testing.T) {
	f := newTreeFixture(t)
	store := checkpoint.Store{Dir: t.TempDir()}
	schema, _ := json.Marshal(f.schema)
	cp, err := store.Initialize("conversation", f.root.Frame, checkpoint.Contract{Target: "agent", Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	a := f.emit(f.root, `{"to":"echo","payload":{"text":"fast"}}`, true, 0, true)
	b := f.emit(f.root, `{"to":"echo","payload":{"text":"slow"}}`, true, 1, true)
	ar := f.emit(a, `{"to":"`+a.Frame+`","payload":"done"}`, false, 0, true)
	f.settle(ar)
	// The adapter disappeared after runtime acceptance and before publishing
	// either sibling. Its saved frontier still contains only the anchor.
	recover := func() RecordedTree {
		t.Helper()
		tr, err := RecoverTree(f.dir, cp.RootInvocation, cp.Leaves, "agent", f.schema, "", "")
		if err != nil {
			t.Fatal(err)
		}
		return tr
	}
	tr := recover()
	want := []string{ar.Stamp, b.Stamp}
	if !sameSet(tr.Frontier, want) {
		t.Fatalf("lost unfinished sibling: %v", tr.Frontier)
	}
	encoded, _ := json.Marshal(tr)
	var evidence checkpoint.Evidence
	if err := json.Unmarshal(encoded, &evidence); err != nil {
		t.Fatal(err)
	}
	cp, err = store.Advance("conversation", cp, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !sameSet(cp.Leaves, want) {
		t.Fatalf("checkpoint: %+v", cp)
	}
	br := f.emit(b, `{"to":"`+b.Frame+`","payload":"done too"}`, false, 0, true)
	f.settle(br)
	winning := f.join(br.Stamp, []string{ar.Stamp, br.Stamp}, true)
	f.join(ar.Stamp, []string{ar.Stamp, br.Stamp}, false) // losing candidate has no claim
	tr = recover()
	if !reflect.DeepEqual(tr.Frontier, []string{winning}) {
		t.Fatalf("did not recover winning join: %v", tr.Frontier)
	}
	again := recover()
	if !reflect.DeepEqual(tr, again) {
		t.Fatal("recovery changed without accepted progress")
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]bool{}
	for _, s := range a {
		m[s] = true
	}
	for _, s := range b {
		if !m[s] {
			return false
		}
	}
	return true
}

func TestRecoveryExcludesProposedAndForeignWorkAndEnforcesContext(t *testing.T) {
	f := newTreeFixture(t)
	f.emit(f.root, `{"to":"echo","payload":{"text":"never accepted"}}`, true, 0, false)
	f.emit(*f.root.Resume, `{"to":"agent","payload":{"text":"another root"}}`, true, 1, true)
	tr, err := RecoverTree(f.dir, f.root.Frame, []string{f.root.Frame}, "agent", f.schema, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tr.Frontier, []string{f.root.Frame}) || len(tr.Records) != 1 {
		t.Fatalf("published orphan/foreign work: %+v", tr)
	}
	if _, err := RecoverTree(f.dir, f.root.Frame, []string{f.root.Frame}, "agent", f.schema, "foreign", ""); !Refused(err) {
		t.Fatalf("foreign recovery context: %v", err)
	}
}

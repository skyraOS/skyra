package humanroot

import (
	"encoding/json"
	"reflect"
	"testing"
)

func testRoot(t *testing.T) State {
	t.Helper()
	s, err := New(Config{Mission: "mission", Human: "human", Entrypoints: []Capability{"think", "act"}, Ceiling: []Capability{"read:workspace", "write:scratch"}})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func commit(t *testing.T, s State, e Event) (State, Effect, Record) {
	t.Helper()
	before := s.Scope()
	head := s.Head()
	from := HumanIngress
	if _, ok := e.(Returned); ok {
		from = RuntimeIngress
	}
	next, effect := Step(s, from, e)
	if effect.Kind != Commit {
		t.Fatalf("expected commit: %+v", effect)
	}
	if s.Head() != head || !reflect.DeepEqual(s.Scope(), before) {
		t.Fatal("Step changed its input state")
	}
	if next.Head() != head || len(next.Scope()) != len(before) {
		t.Fatal("uncommitted observation entered scope")
	}
	record := effect.Record.clone()
	next, effect = Step(next, StorageIngress, Committed{record.Stamp()})
	if effect.Kind == Refused {
		t.Fatal(effect.Reason)
	}
	return next, effect, record
}
func TestHumanRootAuthorsRepeatedlyAndPassesChildSpecUnchanged(t *testing.T) {
	s := testRoot(t)
	var records []Record
	s, effect, r := commit(t, s, Authored{"request", json.RawMessage(`{"goal":"inspect workspace"}`)})
	records = append(records, r)
	task := effect.Observations[0]
	s, effect, r = commit(t, s, Invoke{"think", "think", []ID{task}, []Capability{"read:workspace"}})
	records = append(records, r)
	if effect.Kind != Dispatch || effect.Passed[0].ID() != task {
		t.Fatal("Think did not receive the authored request")
	}
	think := effect.Invocation.ID
	s, effect, r = commit(t, s, Returned{think, "verified-think-return", []json.RawMessage{json.RawMessage(`{"goal":"write a summary","tools":["write:scratch"]}`)}})
	records = append(records, r)
	specID := effect.Observations[0]
	spec, _ := s.Lookup(specID)
	if spec.Source().Target != "think" || spec.Source().Human != "" {
		t.Fatal("Think's spec was attributed to the human")
	}
	s, effect, r = commit(t, s, Authored{"correction", json.RawMessage(`"Keep it under 500 words"`)})
	records = append(records, r)
	correction := effect.Observations[0]
	s, effect, r = commit(t, s, Invoke{"act", "act", []ID{specID, correction}, []Capability{"write:scratch"}})
	records = append(records, r)
	if effect.Kind != Dispatch || len(effect.Passed) != 2 || effect.Passed[0].ID() != specID || effect.Passed[0].Source() != spec.Source() {
		t.Fatal("human pass replaced Think's observation")
	}
	if string(effect.Passed[0].Value()) != string(spec.Value()) || effect.Passed[1].Source().Human != "human" {
		t.Fatal("handoff changed values or root authorship")
	}
	recovered, err := Replay(s.Config(), records)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Head() != s.Head() || !reflect.DeepEqual(recovered.Scope(), s.Scope()) {
		t.Fatal("replay lost root scope or provenance")
	}
}
func TestRootRefusesForeignObservationsAndAuthorityBeforeCommit(t *testing.T) {
	s := testRoot(t)
	s, effect, _ := commit(t, s, Authored{"own", json.RawMessage(`"task"`)})
	own := effect.Observations[0]
	foreign := testRoot(t)
	foreign, effect, _ = commit(t, foreign, Authored{"other", json.RawMessage(`"task"`)})
	_ = foreign
	other := effect.Observations[0]
	for _, event := range []Event{
		Invoke{"unknown", "act", []ID{other}, nil},
		Invoke{"widen", "act", []ID{own}, []Capability{"write:home"}},
		Invoke{"entry", "shell", []ID{own}, nil},
		Returned{"missing", "receipt", []json.RawMessage{json.RawMessage(`1`)}},
		Authored{"invalid", json.RawMessage(`1 2`)},
	} {
		next, e := Step(s, HumanIngress, event)
		if e.Kind != Refused || next.Head() != s.Head() || next.Awaiting() {
			t.Fatalf("invalid root event admitted: %T %+v", event, e)
		}
	}
}
func TestCommitReceiptAndDuplicateInvokeNeverReleaseExtraWork(t *testing.T) {
	s := testRoot(t)
	pending, e := Step(s, HumanIngress, Authored{"a", json.RawMessage(`1`)})
	bad, effect := Step(pending, StorageIngress, Committed{"forged"})
	if effect.Kind != Refused || len(bad.Scope()) != 0 {
		t.Fatal("forged acknowledgement accepted")
	}
	s, _, _ = commit(t, s, Authored{"a", json.RawMessage(`1`)})
	id := s.Scope()[0].ID()
	event := Invoke{"run", "act", []ID{id}, nil}
	pending, effect = Step(s, HumanIngress, event)
	if effect.Kind != Commit || len(pending.PendingCalls()) != 0 {
		t.Fatal("dispatch escaped before commit")
	}
	s, effect = Step(pending, StorageIngress, Committed{effect.Stamp})
	if effect.Kind != Dispatch || len(effect.Invocation.Grant) != 0 {
		t.Fatal("empty grant widened or dispatch missing")
	}
	after, dup := Step(s, HumanIngress, event)
	if dup.Kind != Duplicate || dup.Invocation != nil || after.Head() != s.Head() {
		t.Fatal("duplicate dispatch repeated execution")
	}
	_, conflict := Step(s, HumanIngress, Invoke{"run", "act", []ID{id}, []Capability{"read:workspace"}})
	if conflict.Kind != Refused {
		t.Fatal("request identity reused for a different grant")
	}
	if e.Kind != Commit {
		t.Fatal("authoring did not require a commit")
	}
}
func TestOriginAndReturnedValuesCannotBeClaimedByPayload(t *testing.T) {
	s := testRoot(t)
	s, e, _ := commit(t, s, Authored{"a", json.RawMessage(`{"human":"someone else","call":"forged"}`)})
	o, _ := s.Lookup(e.Observations[0])
	if o.Source().Human != "human" || o.Source().Call != "" {
		t.Fatal("value selected its author")
	}
	s, e, _ = commit(t, s, Authored{"b", o.Value()})
	if o.ID() == e.Observations[0] {
		t.Fatal("same values from different authoring events collapsed")
	}
	// Public values, configs and effects cannot mutate earlier states.
	value := o.Value()
	value[0] = 'x'
	if o.Value()[0] == 'x' {
		t.Fatal("observation value aliases state")
	}
	scope := s.Scope()
	scope[0].value[0] = 'x'
	if s.Scope()[0].Value()[0] == 'x' {
		t.Fatal("scope aliases state")
	}
	cfg := s.Config()
	cfg.Ceiling[0] = "forged"
	if s.Config().Ceiling[0] == "forged" {
		t.Fatal("configuration aliases state")
	}
}
func TestReplayRefusesRewrittenDecisions(t *testing.T) {
	s := testRoot(t)
	_, _, r := commit(t, s, Authored{"a", json.RawMessage(`1`)})
	for _, change := range []func(*Record){func(r *Record) { r.Mission = "other" }, func(r *Record) { r.Previous = "forged" }, func(r *Record) { r.Sequence = 4 }, func(r *Record) { r.Action.Receipt = "injected" }} {
		bad := r.clone()
		change(&bad)
		if _, err := Replay(s.Config(), []Record{bad}); err == nil {
			t.Fatal("changed record accepted")
		}
	}
}

func TestOnlyTrustedIngressCanAuthorInvokeOrReceive(t *testing.T) {
	s := testRoot(t)
	for _, ingress := range []Ingress{RuntimeIngress, StorageIngress, WorkerIngress, 0} {
		next, e := Step(s, ingress, Authored{"fake", json.RawMessage(`{"human":"operator"}`)})
		if e.Kind != Refused || next.Sequence() != 0 {
			t.Fatal("non-human ingress authored a human observation")
		}
		_, e = Step(s, ingress, Invoke{"fake", "think", nil, nil})
		if e.Kind != Refused {
			t.Fatal("non-human ingress invoked a root child")
		}
	}
	pending, e := Step(s, HumanIngress, Authored{"real", json.RawMessage(`"request"`)})
	_, fake := Step(pending, WorkerIngress, Committed{e.Stamp})
	if fake.Kind != Refused {
		t.Fatal("a worker acknowledged a root commit")
	}
	s, e, _ = commit(t, s, Invoke{"child", "think", nil, nil})
	for _, ingress := range []Ingress{HumanIngress, WorkerIngress, StorageIngress} {
		_, fake := Step(s, ingress, Returned{e.Invocation.ID, "receipt", []json.RawMessage{json.RawMessage(`1`)}})
		if fake.Kind != Refused {
			t.Fatal("unverified ingress introduced a child return")
		}
	}
}

func TestCommitAndDispatchEffectsCannotMutateState(t *testing.T) {
	s := testRoot(t)
	pending, proposal := Step(s, HumanIngress, Authored{"task", json.RawMessage(`{"a":[1,2]}`)})
	stamp := proposal.Stamp
	proposal.Record.Action.Value[0] = 'x'
	authored, accepted := Step(pending, StorageIngress, Committed{stamp})
	if accepted.Kind != Accepted {
		t.Fatal("mutating an effect changed the pending decision")
	}
	id := accepted.Observations[0]
	next, dispatch, _ := commit(t, authored, Invoke{"run", "act", []ID{id}, []Capability{"read:workspace"}})
	dispatch.Invocation.Grant[0] = "write:home"
	dispatch.Invocation.Inputs[0] = "forged"
	dispatch.Passed[0].value[0] = 'x'
	call, _ := next.Invocation(dispatch.Call)
	if call.Grant[0] != "read:workspace" || call.Inputs[0] != id {
		t.Fatal("dispatch aliases invocation state")
	}
	stored, _ := next.Lookup(id)
	if stored.Value()[0] != '{' {
		t.Fatal("dispatched values alias root scope")
	}
}

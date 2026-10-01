package main

import (
	"strings"
	"testing"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/tree"
)

// commit stores an input and commits what intake derives from it under h.
func commit(t *testing.T, root string, h horizon.Horizon, in string) (record.Record, string) {
	t.Helper()
	if _, err := tree.Put(root, "inputs", []byte(in)); err != nil {
		t.Fatal(err)
	}
	_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: h, In: []byte(in)})
	r := machine.Committed(effs)
	stamp, err := tree.Commit(root, r)
	if err != nil {
		t.Fatal(err)
	}
	return r, stamp
}

// Resume derives the parked instance from the tree, and refuses anything that
// is not parked work, or that was already resumed (next/20, finding 1).
func TestResumeDerivesTheParkedInstance(t *testing.T) {
	root := t.TempDir()
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	text := &contract.Schema{Type: "string"}
	reach := map[string]horizon.Surface{"browser": {Host: "/tmp/park.sock", Input: text, Output: text}}
	room := horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	g0, err := tree.Commit(root, record.Record{State: record.Invoke, Reach: reach, Authority: &room, Author: "operator", Nonce: "n"})
	if err != nil {
		t.Fatal(err)
	}
	h0 := horizon.Horizon{Stamp: g0, Reach: reach, Authority: room, Frame: g0}
	r, invoke := commit(t, root, h0, `{"to":"browser","payload":"show"}`)

	s, err := parked(root, park.Resumed{Stamp: invoke, Frames: 2})
	if err != nil {
		t.Fatal(err)
	}
	if s.Stamp != invoke || s.Phase != machine.Execute || s.Self.Frame != invoke || s.Self.Surface.Host == "" || s.Self.Authority.Read[0].Path != "/private/work" || s.R.Payload == nil || record.Stamp(r) != invoke {
		t.Fatalf("the instance must be derived from the tree: %+v", s)
	}
	for name, c := range map[string]struct {
		r    park.Resumed
		want string
	}{
		"genesis is not parked work":        {park.Resumed{Stamp: g0, Frames: 1}, "not a parked invocation"},
		"a frame count that does not match": {park.Resumed{Stamp: invoke, Frames: 5}, "carries 5 frames"},
	} {
		if _, err := parked(root, c.r); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%s: %v", name, err)
		}
	}

	// Once the event has been handed on, the same parked work cannot be resumed again.
	self, _ := machine.Narrow(h0, r, invoke)
	commit(t, root, self, `{"to":"`+invoke+`","payload":"hi"}`)
	if _, err := parked(root, park.Resumed{Stamp: invoke, Frames: 2}); err == nil || !strings.Contains(err.Error(), "already resumed") {
		t.Fatalf("a second resume must be refused: %v", err)
	}
}

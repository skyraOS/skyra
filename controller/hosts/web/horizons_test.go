package main

import (
	"encoding/json"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

func TestHorizonsFromTheLineage(t *testing.T) {
	root := t.TempDir()
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	user := horizon.Surface{Path: "/bin/user", SHA256: "u", Continuation: true}
	codex := horizon.Surface{Path: "/bin/codex", SHA256: "c", Continuation: true}
	read := horizon.Surface{Path: "/bin/read", SHA256: "r"}
	runner := horizon.Surface{Path: "/bin/runner", SHA256: "x", System: "runner"}
	room := map[string]horizon.Surface{"user": user, "codex": codex, "read": read, "system.route-runner": runner}
	authority := horizon.Authority{Read: []horizon.PathGrant{{Path: "/work", Scope: "subpath"}}}
	g0 := commit(t, root, "", "genesis", record.Record{State: record.Invoke, Reach: room, Authority: &authority})

	l := newLineage(root)
	l.refresh()
	hs := l.horizonsNow()
	if len(hs) != 1 || !hs[0].Root || len(hs[0].Reach) != 3 || hs[0].Grants != 1 || len(hs[0].Actors) != 0 {
		t.Fatalf("the room alone: %+v", hs)
	}

	// The user side and Codex run under the room: still one Horizon. Then
	// Codex is handed a reference that keeps only read, and runs under it.
	narrow := horizon.Grant{Actor: "codex", Reach: []string{"read"}, Authority: authority}
	for _, h := range []horizon.Horizon{
		{Stamp: g0 + "1", Surface: &user, Reach: room, Authority: authority, Frame: g0 + "1"},
		{Stamp: g0 + "2", Surface: &codex, Reach: room, Authority: authority, Frame: g0 + "2"},
		{Stamp: g0 + "3", Surface: &user, Reach: room, Authority: authority, Refs: map[string]horizon.Grant{"system.ref.1": narrow}, Frame: g0 + "3"},
		{Stamp: g0 + "4", Surface: &codex, Reach: map[string]horizon.Surface{"read": read, "system.route-runner": runner}, Authority: authority, Frame: g0 + "4"},
	} {
		if _, err := tree.PutHorizon(root, h); err != nil {
			t.Fatal(err)
		}
	}
	commit(t, root, g0, "1", record.Record{State: record.Invoke, To: "user", Payload: json.RawMessage(`{}`)})
	if !l.refresh() {
		t.Fatal("new Horizons reported no change")
	}
	hs = l.horizonsNow()
	if len(hs) != 2 {
		t.Fatalf("want the room and one narrowing: %+v", hs)
	}
	if !hs[0].Root || len(hs[0].Actors) != 2 || hs[0].Actors[0] != "codex" || hs[0].Actors[1] != "user" {
		t.Fatalf("the room: %+v", hs[0])
	}
	n := hs[1]
	if n.Root || len(n.Reach) != 1 || n.Reach[0].Name != "read" || len(n.Actors) != 1 || n.Actors[0] != "codex" {
		t.Fatalf("the narrowing: %+v", n)
	}
	for _, s := range hs[0].Reach {
		if s.Name == "read" && s.ID != n.Reach[0].ID {
			t.Fatal("read is one surface in both Horizons")
		}
	}
}

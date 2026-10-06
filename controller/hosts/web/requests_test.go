package main

import (
	"fmt"
	"os"
	"path/filepath"
	"skyra/controller/runtime/tree"
	"strings"
	"testing"
)

func TestDurableAdmissionAndRestart(t *testing.T) {
	dir := t.TempDir()
	h := &webhost{dir: dir}
	m := inbound{Type: "submit", Conversation: "chat", UIIncarnation: "old", RequestID: "r", Text: "private message"}
	if err := h.admit(m, "chat"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".requests.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), m.Text) {
		t.Fatal("ledger duplicates message content")
	}
	restarted := &webhost{dir: dir}
	if err := restarted.loadRequests(); err != nil {
		t.Fatal(err)
	}
	r := restarted.requests["r"]
	if r.Status != "failed" || r.Code != "delivery-unknown" || r.Fingerprint != fingerprint(m) {
		t.Fatalf("restart lost admission: %+v", r)
	}
	m.Text = "other"
	if r.Fingerprint == fingerprint(m) {
		t.Fatal("different work has same identity")
	}
}

func TestRequestIDCannotBeReusedForDifferentWork(t *testing.T) {
	h := &webhost{incarnation: "i", convos: map[string]*conversation{}}
	m := inbound{Type: "open-conversation", RequestID: "r"}
	if err := h.admit(m, "chat"); err != nil {
		t.Fatal(err)
	}
	h.handle([]byte(`{"type":"open-conversation","requestId":"r","text":"different"}`))
	if len(h.convos) != 0 {
		t.Fatal("conflicting request executed")
	}
	h.handle([]byte(`{"type":"open-conversation","requestId":"r"}`))
	if len(h.convos) != 0 {
		t.Fatal("duplicate request executed")
	}
}

func TestRuntimeInterruptionEndsPendingRequestWithoutRetry(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "chat", ".skyra")
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	frame, err := tree.OpenFrame(root, "frame")
	if err != nil {
		t.Fatal(err)
	}
	defer frame.Close()
	if err := tree.PublishFailure(root, frame, tree.Failure{Stamp: "s", Message: "disk failure"}); err != nil {
		t.Fatal(err)
	}
	h := &webhost{dir: dir, log: os.Stderr, convos: map[string]*conversation{"chat": {id: "chat", phase: "pending", pending: "r"}}}
	if err := h.admit(inbound{Type: "submit", RequestID: "r"}, "chat"); err != nil {
		t.Fatal(err)
	}
	if !h.reconcileRequests() {
		t.Fatal("interruption did not change projection")
	}
	if h.convos["chat"].phase != "failed" || h.convos["chat"].pending != "" || h.requests["r"].Code != "delivery-unknown" {
		t.Fatal("request left pending")
	}
	if h.reconcileRequests() {
		t.Fatal("interruption applied twice")
	}
}

func TestLedgerKeepsInFlightAndNewestFinished(t *testing.T) {
	h := &webhost{requests: map[string]*requestRecord{}}
	for i := 0; i < maxFinishedRequests+50; i++ {
		id := fmt.Sprintf("done-%d", i)
		h.requests[id] = &requestRecord{RequestID: id, Status: "returned", TS: int64(i)}
	}
	h.requests["live"] = &requestRecord{RequestID: "live", Status: "accepted", TS: 0}
	h.pruneRequests()
	if len(h.requests) != maxFinishedRequests+1 || h.requests["live"] == nil || h.requests["done-0"] != nil || h.requests[fmt.Sprintf("done-%d", maxFinishedRequests+49)] == nil {
		t.Fatalf("ledger kept %d records", len(h.requests))
	}
}

func TestSnapshotFitsOneFrame(t *testing.T) {
	big := strings.Repeat("x", 1000)
	many := func(n int) []entry {
		es := make([]entry, n)
		for i := range es {
			es[i] = entry{Role: "agent", Text: big}
		}
		return es
	}
	views := []map[string]any{
		{"conversation": "new", "phase": "ready", "entries": many(40)},
		{"conversation": "old", "phase": "closed", "entries": many(40)},
	}
	snap := map[string]any{"type": "snapshot", "conversations": views}
	b, omitted := fitSnapshot(snap, views, 60_000)
	if len(b) > 60_000 || omitted == 0 {
		t.Fatalf("snapshot %d bytes, %d omitted", len(b), omitted)
	}
	if len(views[1]["entries"].([]entry)) != 0 || len(views[0]["entries"].([]entry)) == 0 {
		t.Fatal("finished history should be omitted before the live conversation's")
	}
	b, omitted = fitSnapshot(snap, views, 1<<20)
	if omitted != 0 || len(b) > 1<<20 {
		t.Fatal("a snapshot that fits was changed")
	}
}

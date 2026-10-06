package main

import (
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

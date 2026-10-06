package main

import (
	"os"
	"path/filepath"
	"skyra/controller/runtime/park"
	"testing"
)

func TestDeleteConversationRemovesHistoryAndClosesCustody(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "chat", ".skyra", "records")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(root, "record"))
	if err != nil {
		t.Fatal(err)
	}
	h := &webhost{dir: dir, convos: map[string]*conversation{"chat": {id: "chat", phase: "ready", parked: &park.Frame{Files: []*os.File{f}}}}}
	if err := h.deleteConversation("chat"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Stat(); err == nil {
		t.Fatal("custody descriptor still open")
	}
	if _, err := os.Stat(filepath.Join(dir, "chat")); !os.IsNotExist(err) {
		t.Fatalf("history remains: %v", err)
	}
	if h.convos["chat"] != nil || !h.deleted("chat") {
		t.Fatal("deletion not recorded")
	}
	if err := h.deleteConversation("chat"); err != nil {
		t.Fatal("retry:", err)
	}
	// Late disk writes must not make the chat reappear on restart.
	os.MkdirAll(root, 0700)
	restarted := &webhost{dir: dir, convos: map[string]*conversation{}}
	restarted.loadPast()
	if len(restarted.convos) != 0 {
		t.Fatal("deleted chat resurrected")
	}
}

func TestDeleteRejectsPendingAndInvalidPaths(t *testing.T) {
	dir := t.TempDir()
	h := &webhost{dir: dir, convos: map[string]*conversation{"busy": {id: "busy", phase: "pending"}}}
	for _, id := range []string{"busy", "../outside", ".", "", "/tmp", "missing", "a/b"} {
		if err := h.deleteConversation(id); err == nil {
			t.Errorf("accepted %q", id)
		}
	}
	if h.convos["busy"] == nil || h.deleted("busy") {
		t.Fatal("pending chat changed")
	}
}

func TestDeleteRejectsStaleUI(t *testing.T) {
	h := &webhost{dir: t.TempDir(), incarnation: "current", convos: map[string]*conversation{"chat": {id: "chat", phase: "closed"}}}
	h.handle([]byte(`{"type":"delete-conversation","conversation":"chat","uiIncarnation":"old","requestId":"r"}`))
	if h.convos["chat"] == nil {
		t.Fatal("stale UI deleted chat")
	}
	h.handle([]byte(`{"type":"delete-conversation","conversation":"chat","uiIncarnation":"current","requestId":"r"}`))
	if h.convos["chat"] != nil || h.requests["r"] == nil {
		t.Fatal("valid delete not accepted")
	}
	h.handle([]byte(`{"type":"delete-conversation","conversation":"chat","uiIncarnation":"current","requestId":"r"}`))
}

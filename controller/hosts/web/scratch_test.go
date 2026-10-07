package main

import (
	"path/filepath"
	"testing"
)

func TestScratchForSharedIdeasAndConversationDefault(t *testing.T) {
	h := &webhost{}
	if got := h.scratchFor("/conversations/one"); got != filepath.Join("/conversations/one", "scratch") {
		t.Fatalf("default scratch: %s", got)
	}
	h.scratch = "/workspace/ideas"
	for _, home := range []string{"/conversations/one", "/conversations/two"} {
		if got := h.scratchFor(home); got != "/workspace/ideas" {
			t.Fatalf("shared ideas: %s", got)
		}
	}
}

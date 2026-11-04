package main

import (
	"os"
	"path/filepath"
	"testing"
)

// A tree is started only with a spec, a kind its reach agrees with, and
// targets a tree may hold here; a past tree's host is gone, so it is shown
// as failed.
func TestStartTreeRefuses(t *testing.T) {
	h := &webhost{}
	if _, why := h.startTree("g", "delegate", []string{"find", "read"}); why == "" {
		t.Fatal("a webhost without trees started one")
	}
	h.trees = newTrees(t.TempDir(), "/nonexistent", "/nonexistent")
	for _, c := range []struct {
		goal, kind string
		reach      []string
	}{
		{"", "delegate", []string{"find", "read"}},
		{"g", "effect", []string{"find", "read"}},
		{"g", "act", []string{"find", "read"}},
		{"g", "delegate", []string{"find"}},
		{"g", "delegate", []string{"find", "write"}},
		{"g", "act", nil},
	} {
		if _, why := h.startTree(c.goal, c.kind, c.reach); why == "" {
			t.Errorf("%+v was started", c)
		}
	}
	if len(h.trees.runs) != 0 {
		t.Fatal("a refused tree was kept")
	}
}

func TestPastTreesLoad(t *testing.T) {
	dir := t.TempDir()
	run := filepath.Join(dir, "t-1")
	os.MkdirAll(filepath.Join(run, ".skyra"), 0o700)
	os.WriteFile(filepath.Join(run, "tree.json"), []byte(`{"id":"t-1","goal":"g","kind":"act","reach":["echo"],"phase":"running","started":1}`), 0o600)
	tr := newTrees(dir, "/nonexistent", "/nonexistent")
	r := tr.runs["t-1"]
	if r == nil || r.Phase != "failed" || r.View != nil {
		t.Fatalf("%+v", r)
	}
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// A conversation is kept with the surface it runs (proposal 38): under its
// procedure's directory. One from before procedures stays at the top, where
// it was, and is history.
func TestConversationsAreKeptWithTheirProcedure(t *testing.T) {
	dir := t.TempDir()
	for _, home := range []string{"c-old", filepath.Join("think", "c-new")} {
		if err := os.MkdirAll(filepath.Join(dir, home, ".skyra", "records"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	h := &webhost{dir: dir, log: os.Stderr, convos: map[string]*conversation{}}
	h.loadPast()
	if c := h.convos["c-old"]; c == nil || c.procedure != "" || c.home != filepath.Join(dir, "c-old") {
		t.Fatalf("a conversation from before procedures: %+v", c)
	}
	if c := h.convos["c-new"]; c == nil || c.procedure != "think" || c.home != filepath.Join(dir, "think", "c-new") {
		t.Fatalf("a conversation kept under think: %+v", c)
	}
	if got := h.homeFor("think", "c-x"); got != filepath.Join(dir, "think", "c-x") {
		t.Fatalf("a new think conversation is kept at %s", got)
	}
	if got := h.locate("c-new"); got != filepath.Join(dir, "think", "c-new") {
		t.Fatalf("located at %s", got)
	}

	// Deleting removes exactly the conversation's own directory.
	if err := h.deleteConversation("c-new"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "think", "c-new")); !os.IsNotExist(err) {
		t.Fatal("the conversation was not removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "think")); err != nil {
		t.Fatal("deleting a conversation removed its procedure's directory")
	}
	// A home that is not the conversation's own directory is never removed.
	h.convos["c-old"].home = dir
	if err := h.deleteConversation("c-old"); err == nil {
		t.Fatal("deleted outside the conversation's directory")
	}
	if _, err := os.Stat(filepath.Join(dir, "c-old")); err != nil {
		t.Fatal("a refused delete removed something")
	}
}

func TestAProcedureMustNameADirectory(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "bundle.json")
	os.WriteFile(bundle, []byte(`{"version":1,"procedures":[{"contract":{"title":"../escape","type":"object"},"nodes":{}}]}`), 0o600)
	reg := filepath.Join(dir, "registry.json")
	os.WriteFile(reg, []byte(`{"interpreter":{"bundle":"`+bundle+`"}}`), 0o600)
	if _, err := loadDestinations(reg); err == nil {
		t.Fatal("a procedure whose title is a path was accepted")
	}
}

// A conversation carried on after its run ended is read back from its newest
// run, whose transcript is the whole conversation; a run with no records yet
// (still starting) does not count.
func TestTheNewestRunIsTheConversation(t *testing.T) {
	home := t.TempDir()
	for _, run := range []string{".skyra", "runs/2/.skyra", "runs/10/.skyra"} {
		os.MkdirAll(filepath.Join(home, run, "records"), 0o700)
	}
	os.MkdirAll(filepath.Join(home, "runs", "11", ".skyra"), 0o700) // no records yet
	os.MkdirAll(filepath.Join(home, "runs", "notes"), 0o700)
	if root, n := latestRun(home); root != filepath.Join(home, "runs", "10", ".skyra") || n != 10 {
		t.Fatalf("newest run: %s %d", root, n)
	}
	if root, n := latestRun(t.TempDir()); filepath.Base(root) != ".skyra" || n != 1 {
		t.Fatalf("a conversation never carried on: %s %d", root, n)
	}
}

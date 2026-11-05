package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func workHost(t *testing.T) *webhost {
	t.Helper()
	dir := t.TempDir()
	home := filepath.Join(dir, "c-one")
	h := &webhost{dir: dir, incarnation: "i", skyra: "/usr/bin/true", log: os.Stderr, convos: map[string]*conversation{}}
	h.convos["c-one"] = &conversation{id: "c-one", home: home, phase: "ready", model: "claude", work: newWork("c-one"), lineage: newLineage(filepath.Join(home, ".skyra"))}
	h.trees = newTrees(filepath.Join(dir, "trees"), "/nonexistent", "/nonexistent")
	return h
}

func workMessage(h *webhost, m inbound) {
	m.UIIncarnation = "i"
	m.Conversation = "c-one"
	b, _ := json.Marshal(m)
	h.handle(b)
}

func TestSpecSaveIsSeparateFromChatAndExecution(t *testing.T) {
	h := workHost(t)
	m := inbound{Type: "save-spec", RequestID: "save", Text: "Read the README."}
	workMessage(h, m)
	c := h.convos["c-one"]
	if c.work.Revision != 1 || c.work.Spec != m.Text || len(c.entries) != 0 || len(h.trees.runs) != 0 {
		t.Fatalf("save started work or changed chat: %+v", c.work)
	}
	// Same request is a replay, before checking its now-stale revision.
	workMessage(h, m)
	if c.work.Revision != 1 || len(c.work.Revisions) != 1 {
		t.Fatal("duplicate save minted a revision")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "stale", Text: "overwrite", SpecRevision: 0})
	if c.work.Spec != m.Text || h.requests["stale"] != nil {
		t.Fatal("a stale editor overwrote the spec")
	}
	loaded := loadWork(c.home, c.id)
	if loaded == nil || loaded.Spec != m.Text || loaded.TreeID != c.work.TreeID || loaded.SpecID != c.work.SpecID {
		t.Fatal("work identities or saved revision were lost on reload")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "same", Text: m.Text, SpecRevision: 1})
	if c.work.Revision != 1 {
		t.Fatal("saving identical text created a revision")
	}
}

func TestRunPinsSpecAndKeepsConversationTreeIdentity(t *testing.T) {
	h := workHost(t)
	workMessage(h, inbound{Type: "save-spec", RequestID: "save", Text: "first"})
	workMessage(h, inbound{Type: "run-spec", RequestID: "run", SpecRevision: 1, Text: "UNTRUSTED CHAT", Reach: []string{"read"}})
	c := h.convos["c-one"]
	firstID := c.work.LastRunID
	if firstID == "" {
		t.Fatal("no linked execution")
	}
	waitRun(t, h, firstID)
	h.trees.mu.Lock()
	first := *h.trees.runs[firstID]
	h.trees.mu.Unlock()
	if first.Goal != "first" || first.Conversation != c.id || first.SpecID != c.work.SpecID || first.TreeID != c.work.TreeID || first.SpecRevision != 1 || first.Actor != "claude" {
		t.Fatalf("run did not pin its saved spec and grant: %+v", first)
	}
	// A duplicate request cannot create a second execution, even after completion.
	workMessage(h, inbound{Type: "run-spec", RequestID: "run", SpecRevision: 1, Text: "UNTRUSTED CHAT", Reach: []string{"read"}})
	if len(h.trees.runs) != 1 {
		t.Fatal("duplicate run executed twice")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "edit", SpecRevision: 1, Text: "second"})
	workMessage(h, inbound{Type: "run-spec", RequestID: "run-again", SpecRevision: 2, Reach: []string{"read"}})
	waitRun(t, h, c.work.LastRunID)
	h.trees.mu.Lock()
	second := *h.trees.runs[c.work.LastRunID]
	h.trees.mu.Unlock()
	if second.ID == first.ID || second.TreeID != first.TreeID || second.SpecID != first.SpecID || second.SpecRevision != 2 || first.Goal != "first" {
		t.Fatalf("execution replaced identity/history: %+v", second)
	}
	loaded := loadWork(c.home, c.id)
	restarted := newTrees(h.trees.dir, "/nonexistent", "/nonexistent")
	if loaded.LastRunID != second.ID || restarted.runs[second.ID].TreeID != loaded.TreeID || len(loaded.Revisions) != 2 {
		t.Fatal("restart lost linkage or spec history")
	}
}

func waitRun(t *testing.T, h *webhost, id string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		h.trees.mu.Lock()
		phase := h.trees.runs[id].Phase
		h.trees.mu.Unlock()
		if phase != "running" {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("execution did not finish")
}

func TestRunningTreeAllowsSpecEditsButRefusesAnotherRun(t *testing.T) {
	h := workHost(t)
	c := h.convos["c-one"]
	workMessage(h, inbound{Type: "save-spec", RequestID: "save", Text: "first"})
	run, why := h.prepareTree("first", "act", []string{"read"}, c.work, "claude")
	if why != "" {
		t.Fatal(why)
	}
	workMessage(h, inbound{Type: "run-spec", RequestID: "busy", SpecRevision: 1, Reach: []string{"read"}})
	if len(h.trees.runs) != 1 || h.requests["busy"] != nil {
		t.Fatal("parallel execution admitted")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "edit", SpecRevision: 1, Text: "next"})
	if c.work.Revision != 2 || run.Goal != "first" || run.SpecRevision != 1 {
		t.Fatal("editing a draft changed running input")
	}
	workMessage(h, inbound{Type: "run-spec", RequestID: "stale", SpecRevision: 1, Reach: []string{"read"}})
	if h.requests["stale"] != nil {
		t.Fatal("stale revision dispatched")
	}
}

func TestMissingAndCorruptWorkNeverRuns(t *testing.T) {
	h := workHost(t)
	workMessage(h, inbound{Type: "run-spec", RequestID: "empty", Reach: []string{"read"}})
	if len(h.trees.runs) != 0 {
		t.Fatal("empty spec ran")
	}
	c := h.convos["c-one"]
	if err := os.MkdirAll(c.home, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c.home, "work.json"), []byte(`{broken`), 0600); err != nil {
		t.Fatal(err)
	}
	if loadWork(c.home, c.id) != nil {
		t.Fatal("corrupt spec treated as new work")
	}
	old := loadWork(t.TempDir(), "c-legacy")
	if old == nil || old.Revision != 0 || old.TreeID != "tree-c-legacy" {
		t.Fatal("legacy conversation lacks stable empty identities")
	}
}

func TestChatContextSeesSpecAndResultsWithoutStartingExecution(t *testing.T) {
	h := workHost(t)
	workMessage(h, inbound{Type: "save-spec", RequestID: "save", Text: "definition"})
	c := h.convos["c-one"]
	if h.specContext(c)["spec"] != "definition" {
		t.Fatal("chat cannot see linked spec")
	}
	h.trees.runs["finished"] = &treeRun{ID: "finished", Conversation: c.id, TreeID: c.work.TreeID, SpecRevision: 1, Phase: "done", Answer: "result"}
	c.work.LastRunID = "finished"
	context := h.specContext(c)
	result := context["execution"].(map[string]any)
	if result["answer"] != "result" || result["revision"] != 1 || len(h.trees.runs) != 1 {
		t.Fatal("chat context lost the execution result or dispatched work")
	}
	// Context is a snapshot of the operator-authored revision seen by the turn.
	workMessage(h, inbound{Type: "save-spec", RequestID: "edit", SpecRevision: 1, Text: "new definition"})
	if context["spec"] != "definition" || h.specContext(c)["spec"] != "new definition" {
		t.Fatal("editing changed an earlier turn's context")
	}
}

func TestWorkRejectsPartialOrInconsistentMetadata(t *testing.T) {
	home := t.TempDir()
	for _, bad := range []string{`null`, `{}`, `{"spec":"unversioned"}`, `{"conversation":"c","treeId":"tree-c","specId":"spec-c","revision":1,"spec":"changed","revisions":[{"revision":1,"text":"original"}]}`} {
		if err := os.WriteFile(filepath.Join(home, "work.json"), []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if loadWork(home, "c") != nil {
			t.Fatalf("corrupt metadata accepted: %s", bad)
		}
	}
}

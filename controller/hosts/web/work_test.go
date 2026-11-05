package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/runtime/mission"
	"skyra/controller/runtime/registry"
)

// missionRegistry is an operator registry of the tools a Mission may grant,
// Plan's viewer and proposer, the apprentice and a claude actor, at bins
// (name → binary).
func missionRegistry(t *testing.T, dir string, bins map[string]string) string {
	t.Helper()
	absolute := func(path string) string { value, _ := filepath.Abs(path); return value }
	actor := filepath.Join(dir, "actor-contract.json")
	schema := &contract.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*contract.Schema{"text": {Type: "string"}, "fresh": {Type: "bool"}, "tools": {Type: "array", Items: &contract.Schema{Type: "string"}}, "instructions": {Type: "any"}}}
	b, _ := json.Marshal(registry.Declared{Input: schema})
	if err := os.WriteFile(actor, b, 0600); err != nil {
		t.Fatal(err)
	}
	reg := map[string]registry.Entry{
		"echo":                {Path: bins["echo"], Contract: absolute("../../targets/echo/contract.json")},
		"find":                {Path: bins["find"], Contract: absolute("../../targets/find/contract.json")},
		"read":                {Path: bins["read"], Contract: absolute("../../targets/read/contract.json")},
		"propose-spec":        {Path: bins["propose-spec"], Contract: absolute("../../targets/propose-spec/contract.json")},
		"view-plan-contracts": {Path: bins["view-plan-contracts"], Contract: absolute("../../targets/view-plan-contracts/contract.json"), State: true},
		"claude":              {Path: bins["claude"], Contract: actor},
		"apprentice":          {Path: bins["apprentice"], Contract: absolute("../../apprentice/contract.json"), State: true},
	}
	// Plan looks around a workspace of its own.
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range mission.Navigation {
		reg[name] = registry.Entry{Path: bins["workdir"], Contract: absolute("../../targets/workdir/" + strings.TrimPrefix(name, "plan.") + ".json"),
			Authority: boundary.Authority{Read: []boundary.PathGrant{{Path: workspace, Scope: "subpath"}}, Workspace: workspace}}
	}
	path := filepath.Join(dir, "registry.json")
	b, _ = json.Marshal(reg)
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

// workHost is a host whose binaries are stand-ins: specs are recorded at the
// human root, and anything executed fails without effects.
func workHost(t *testing.T) *webhost {
	t.Helper()
	dir := t.TempDir()
	bins := map[string]string{}
	for _, name := range []string{"echo", "find", "read", "propose-spec", "view-plan-contracts", "workdir", "claude", "apprentice", "kernel"} {
		bins[name] = filepath.Join(dir, name)
		if err := os.WriteFile(bins[name], []byte("#!/bin/sh\nexit 1\n# "+name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return hostWith(t, dir, bins)
}

func hostWith(t *testing.T, dir string, bins map[string]string) *webhost {
	home := filepath.Join(dir, "c-one")
	h := &webhost{dir: dir, incarnation: "i", skyra: bins["kernel"], log: os.Stderr, convos: map[string]*conversation{}}
	work := newWork("c-one")
	work.Plan = true
	h.convos["c-one"] = &conversation{id: "c-one", home: home, phase: "ready", model: "claude", work: work, lineage: newLineage(filepath.Join(home, ".skyra")), loaded: true}
	h.trees = newTrees(filepath.Join(dir, "trees"), "/nonexistent", missionRegistry(t, dir, bins))
	return h
}

func workMessage(h *webhost, m inbound) {
	m.UIIncarnation = "i"
	m.Conversation = "c-one"
	b, _ := json.Marshal(m)
	h.handle(b)
}

func TestSavedSpecIsAHumanSpecHeldAtTheRoot(t *testing.T) {
	h := workHost(t)
	m := inbound{Type: "save-spec", RequestID: "save", Text: "Read the README.", Reach: []string{"read"}}
	workMessage(h, m)
	c := h.convos["c-one"]
	latest := c.work.latest()
	if c.work.Revision != 1 || latest == nil || latest.Text != m.Text || latest.Source != "human" || latest.Observation == "" || !slices.Equal(latest.Tools, []string{"read"}) {
		t.Fatalf("save did not record a human Spec at the root: %+v", c.work)
	}
	if len(c.entries) != 0 || len(h.trees.runs) != 0 {
		t.Fatal("save started work or changed chat")
	}
	// Same request is a replay, before checking its now-stale revision.
	workMessage(h, m)
	if c.work.Revision != 1 || len(c.work.Revisions) != 1 {
		t.Fatal("duplicate save minted a revision")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "stale", Text: "overwrite", SpecRevision: 0, Reach: []string{"read"}})
	if c.work.Spec != m.Text || h.requests["stale"] != nil {
		t.Fatal("a stale editor overwrote the spec")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "same", Text: m.Text, SpecRevision: 1, Reach: []string{"read"}})
	if c.work.Revision != 1 {
		t.Fatal("saving an identical spec created a revision")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "more", Text: m.Text, SpecRevision: 1, Reach: []string{"read", "find"}})
	if c.work.Revision != 2 || !slices.Equal(c.work.latest().Tools, []string{"find", "read"}) || c.work.latest().Observation == latest.Observation {
		t.Fatal("changing the tools is not a new Spec")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "write", Text: m.Text, SpecRevision: 2, Reach: []string{"write"}})
	if c.work.Revision != 2 {
		t.Fatal("a tool outside the catalogue was proposed")
	}
	loaded := loadWork(c.home, c.id)
	if loaded == nil || loaded.latest().Observation != c.work.latest().Observation || loaded.TreeID != c.work.TreeID {
		t.Fatal("work identities or the held Spec were lost on reload")
	}
}

func TestLaunchPassesOnlyAHeldSpec(t *testing.T) {
	h := workHost(t)
	workMessage(h, inbound{Type: "launch-spec", RequestID: "empty"})
	if len(h.trees.runs) != 0 || h.requests["empty"] != nil {
		t.Fatal("a mission with no spec launched")
	}
	c := h.convos["c-one"]
	// A revision from before Missions was never recorded at the root.
	c.work.Revision, c.work.Spec, c.work.Revisions = 1, "legacy", []specRevision{{Revision: 1, Text: "legacy"}}
	workMessage(h, inbound{Type: "launch-spec", RequestID: "legacy", SpecRevision: 1})
	if len(h.trees.runs) != 0 {
		t.Fatal("a spec the root does not hold launched")
	}
	workMessage(h, inbound{Type: "save-spec", RequestID: "save", Text: "Echo hello.", SpecRevision: 1, Reach: []string{"echo"}})
	workMessage(h, inbound{Type: "launch-spec", RequestID: "launch", SpecRevision: 2})
	if c.work.LastRunID == "" {
		t.Fatal("no linked launch")
	}
	waitRun(t, h, c.work.LastRunID)
	h.trees.mu.Lock()
	run := *h.trees.runs[c.work.LastRunID]
	h.trees.mu.Unlock()
	if run.Spec != c.work.latest().Observation || run.Call == "" || run.SpecRevision != 2 || run.Goal != "Echo hello." || !slices.Equal(run.Reach, []string{"echo"}) {
		t.Fatalf("the launch did not pin its held Spec: %+v", run)
	}
	// The stand-in kernel fails: the outcome is recorded, never retried.
	if run.Phase != "failed" {
		t.Fatalf("a failed launch reported %s", run.Phase)
	}
	workMessage(h, inbound{Type: "launch-spec", RequestID: "launch", SpecRevision: 2})
	if len(h.trees.runs) != 1 {
		t.Fatal("a duplicate launch request ran twice")
	}
	restarted := newTrees(h.trees.dir, "/nonexistent", h.trees.registry)
	if restarted.runs[run.ID] == nil || restarted.runs[run.ID].root != run.Root {
		t.Fatal("restart lost the launch's run in its Mission")
	}
}

func TestFailedPlanIsReportedAndKeepsTheSpec(t *testing.T) {
	h := workHost(t)
	workMessage(h, inbound{Type: "save-spec", RequestID: "save", Text: "kept", Reach: []string{"read"}})
	workMessage(h, inbound{Type: "submit", RequestID: "plan", Text: "make it better"})
	c := h.convos["c-one"]
	waitPlan(t, h)
	if c.work.Revision != 1 || c.work.Spec != "kept" || c.work.PlanError == "" {
		t.Fatalf("a failed Plan changed the spec or went unreported: %+v", c.work)
	}
	if len(c.entries) != 2 || !strings.HasPrefix(c.entries[1].Text, "⚠ Plan could not answer") || c.phase != "ready" {
		t.Fatalf("a failed turn is not in the conversation: %+v", c.entries)
	}
}

// A Plan conversation is its turns: reloaded from disk, the transcript is
// rebuilt from them, and a turn the host stopped in is reported, not rerun.
func TestAPlanConversationReloadsFromItsTurns(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "think", "c-one")
	w := newWork("c-one")
	w.Plan, w.Planning = true, true
	w.Turns = []planTurn{{RequestID: "a", Text: "hi", Note: "What should Launch do?", Done: true}, {RequestID: "b", Text: "echo hello"}}
	if err := saveWork(home, w); err != nil {
		t.Fatal(err)
	}
	h := &webhost{dir: dir, log: os.Stderr, convos: map[string]*conversation{}}
	h.loadPast()
	c := h.convos["c-one"]
	if c == nil || c.phase != "ready" || !c.loaded || c.work.Planning {
		t.Fatalf("not reloaded as a ready Plan conversation: %+v", c)
	}
	if len(c.entries) != 4 || c.entries[1].Text != "What should Launch do?" || !strings.Contains(c.entries[3].Text, "host stopped") {
		t.Fatalf("transcript: %+v", c.entries)
	}
}

func waitRun(t *testing.T, h *webhost, id string) {
	t.Helper()
	deadline := time.Now().Add(time.Minute)
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

func waitPlan(t *testing.T, h *webhost) {
	t.Helper()
	deadline := time.Now().Add(time.Minute)
	for time.Now().Before(deadline) {
		h.mu.Lock()
		planning := h.convos["c-one"].work.Planning
		h.mu.Unlock()
		if !planning {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("Plan did not finish")
}

func TestCorruptWorkNeverRuns(t *testing.T) {
	h := workHost(t)
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
	workMessage(h, inbound{Type: "save-spec", RequestID: "save", Text: "definition", Reach: []string{"read"}})
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
	// Context is a snapshot of the revision seen by the turn.
	workMessage(h, inbound{Type: "save-spec", RequestID: "edit", SpecRevision: 1, Text: "new definition", Reach: []string{"read"}})
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

// In the app, find is the repository host's, not a program: a Launch is
// offered only the tree targets a registry holds as programs, and the
// Mission still opens, rather than failing every Plan and Launch.
func TestAHostServedTargetIsNotOfferedToLaunch(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry.json")
	b, _ := json.Marshal(map[string]any{
		"find": map[string]any{"contract": "c.json", "host": filepath.Join(dir, "repo.sock")},
		"read": map[string]any{"path": "read", "contract": "c.json"},
		"echo": map[string]any{"path": "echo", "contract": "c.json"},
	})
	if err := os.WriteFile(registry, b, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := programs(registry, treeTargets); !slices.Equal(got, []string{"read", "echo"}) {
		t.Fatalf("offered %v", got)
	}
}

// A Mission whose conversation was a chat becomes Plan's on its next message,
// the chat kept as its earlier turns.
func TestAChatMissionBecomesPlansOnItsNextMessage(t *testing.T) {
	h := workHost(t)
	c := h.convos["c-one"]
	c.work.Plan, c.procedure = false, "think"
	c.entries = []entry{{Role: "user", Text: "hi"}, {Role: "call", Text: ""}, {Role: "agent", Text: "Hi! What can I help with?"}, {Role: "user", Text: "what tools?"}, {Role: "agent", Text: "read, find."}}
	workMessage(h, inbound{Type: "submit", RequestID: "next", Text: "plan the echo"})
	waitPlan(t, h)
	if !c.work.Plan || len(c.work.Turns) != 3 || c.work.Turns[0].Note != "Hi! What can I help with?" || c.work.Turns[2].Text != "plan the echo" || !c.work.Turns[2].Done {
		t.Fatalf("the chat did not become Plan's turns: %+v", c.work.Turns)
	}
	if len(c.entries) != 6 || c.entries[3].Text != "read, find." || c.phase != "ready" {
		t.Fatalf("transcript: %+v", c.entries)
	}
	// Other roots stay as they are.
	if h.plans("design") {
		t.Fatal("a design conversation became Plan's")
	}
}

// A Plan turn that hit a bound says so plainly, not as the tree's record.
func TestAPlanTurnOutOfTurnsSaysSo(t *testing.T) {
	if got := outOf(`Plan did not return a Spec: {"faulted":"0","why":"the tree's budget"}`); !strings.Contains(got, "used all my turns") {
		t.Fatalf("budget: %q", got)
	}
	if got := outOf("Plan did not return a Spec: the turn bound"); !strings.Contains(got, "ran out of turns") {
		t.Fatalf("turns: %q", got)
	}
	if got := outOf("kernel unavailable"); got != "kernel unavailable" {
		t.Fatalf("other: %q", got)
	}
}

//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	typed "skyra/controller/core/mission"
)

// The whole Mission through the webhost and a real kernel: Plan views the
// human's contracts and writes a Spec from the conversation, the human
// launches exactly that Spec, and its tree receives what the Spec cites. The
// model is a scripted actor.
func TestPlanThenLaunchThroughTheWebhost(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	bins := map[string]string{}
	for name, source := range map[string]string{"kernel": "../../skyra", "echo": "../../targets/echo", "find": "../../targets/find", "read": "../../targets/read", "propose-spec": "../../targets/propose-spec", "view-plan-contracts": "../../targets/view-plan-contracts", "workdir": "../../targets/workdir", "claude": "../../skyra/testdata/mission-actor.go"} {
		bins[name] = filepath.Join(dir, name)
		if out, err := exec.CommandContext(ctx, "go", "build", "-o", bins[name], source).CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v %s", name, err, out)
		}
	}
	bins["apprentice"] = filepath.Join(dir, "apprentice")
	build := exec.CommandContext(ctx, "go", "build", "-o", bins["apprentice"], ".")
	build.Dir = "../../apprentice"
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build apprentice: %v %s", err, out)
	}
	h := hostWith(t, dir, bins)
	if err = os.WriteFile(filepath.Join(dir, "workspace", "NOTES.md"), []byte("plan notes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c := h.convos["c-one"]

	// The conversation is Plan: a message is a Plan turn, its note the reply.
	workMessage(h, inbound{Type: "submit", RequestID: "plan", Text: "Write the spec for this mission from the conversation so far."})
	waitPlan(t, h)
	if len(c.entries) != 2 || c.entries[1].Role != "agent" || c.entries[1].Text != "I wrote the spec from your message." || c.entries[1].Revision != 1 || c.phase != "ready" {
		t.Fatalf("Plan's reply is not the conversation's: %+v", c.entries)
	}
	latest := c.work.latest()
	if c.work.PlanError != "" || latest == nil {
		t.Fatalf("Plan failed: %s", c.work.PlanError)
	}
	if latest.Source != "plan" || latest.Text != "Echo the cited request: Write the spec for this mission from the conversation so far." || latest.Coverage == nil || latest.Coverage.Cited != 11 || latest.Coverage.Plan != 4 || len(latest.Tools) != 1 || latest.Tools[0] != "echo" || len(latest.Inputs) != 1 {
		t.Fatalf("Plan's Spec was not kept as it returned it: %+v", latest)
	}

	if latest.Segments[1].From != "request" {
		t.Fatalf("a quote of the request is not marked as one: %+v", latest.Segments)
	}
	if latest.Profile == nil || *latest.Profile != typed.DefaultProfile {
		t.Fatalf("Plan's first Spec does not state the default profile: %+v", latest.Profile)
	}

	// Editing Plan's Spec revises it: what the human kept keeps its source,
	// and only their words are theirs.
	// The human sets the profile its Launch runs under.
	edited := latest.Text + " Also log it to the console."
	profile := typed.Profile{Act: 2, Turns: 5, Depth: 2, FanOut: 16, Budget: 30}
	workMessage(h, inbound{Type: "save-spec", RequestID: "edit", SpecRevision: 1, Text: edited, Reach: []string{"echo"}, Profile: &profile})
	saved := c.work.latest()
	if saved.Profile == nil || *saved.Profile != profile {
		t.Fatalf("the human's profile was not kept: %+v", saved.Profile)
	}
	if saved.Revision != 2 || saved.Source != "human" || *saved.Coverage != (typed.Coverage{Words: 21, Cited: 11, Plan: 4, Human: 6}) || len(saved.Inputs) != 1 {
		t.Fatalf("the edit did not keep what it kept: %+v %+v", saved, saved.Coverage)
	}
	if saved.Segments[1].From != "request" {
		t.Fatalf("a kept quote lost where it came from: %+v", saved.Segments)
	}

	// Plan refines that revision: it is passed, and Plan keeps the human's
	// words by quoting it; its own words it adds again.
	workMessage(h, inbound{Type: "submit", RequestID: "replan", Text: "Keep my addition please."})
	waitPlan(t, h)
	replanned := c.work.latest()
	if c.work.PlanError != "" || replanned.Revision != 3 || *replanned.Coverage != (typed.Coverage{Words: 14, Cited: 10, Plan: 4}) || replanned.Segments[2].From != "spec" || c.entries[3].Revision != 3 {
		t.Fatalf("Plan did not refine the saved revision: %s %+v", c.work.PlanError, replanned)
	}
	if replanned.Profile == nil || *replanned.Profile != profile {
		t.Fatalf("Plan did not keep the human's profile: %+v", replanned.Profile)
	}

	// Launch runs the revision the human is looking at, not only the latest.
	workMessage(h, inbound{Type: "launch-spec", RequestID: "launch", SpecRevision: 3})
	if c.work.LastRunID == "" {
		t.Fatal("no linked launch")
	}
	waitRun(t, h, c.work.LastRunID)
	h.trees.mu.Lock()
	run := *h.trees.runs[c.work.LastRunID]
	h.trees.mu.Unlock()
	if run.Phase != "done" || run.Spec != replanned.Observation {
		t.Fatalf("launch did not complete from the planned Spec: %+v", run)
	}
	// Launch's tree ran under the Spec's profile.
	frames, _ := filepath.Glob(filepath.Join(run.root, "apprentice", "tree", "frames", "*.json"))
	if len(frames) != 1 {
		t.Fatalf("launch frames: %v", frames)
	}
	frame, _ := os.ReadFile(frames[0])
	var asked struct {
		Ask struct{ Profile typed.Profile }
	}
	if json.Unmarshal(frame, &asked) != nil || asked.Ask.Profile != profile {
		t.Fatalf("the Launch's tree did not run under the Spec's profile: %+v", asked.Ask.Profile)
	}
	// The Spec cited the request Plan was passed; Launch received it.
	if !strings.Contains(run.Answer, "Echo the cited request") || !strings.Contains(run.Answer, "Keep my addition please.") {
		t.Fatalf("launch did not receive the Spec and its cited input: %s", run.Answer)
	}
}

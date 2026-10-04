//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"skyra/controller/core/contract"
	"skyra/controller/core/record"
	"skyra/controller/runtime/boot"
	"skyra/controller/runtime/metric"
	"skyra/controller/runtime/registry"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// room builds skyra, echo and the scripted agent, and returns a function that
// boots a lineage of the agent with a payload and returns its state
// directory, and the agent's and echo's input contracts.
func room(t *testing.T) (func(text string) string, *contract.Schema, *contract.Schema) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bins := map[string]string{"skyra": ".", "echo": "../targets/echo", "agent": "../targets/agent"}
	for name, source := range bins {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		out, err := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, name), source).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v\n%s", source, err, out)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "interp"), ".")
	build.Dir = "../interp" // its own module
	out, err := build.CombinedOutput()
	cancel()
	if err != nil {
		t.Fatalf("build the interpreter: %v\n%s", err, out)
	}
	abs := func(p string) string { a, _ := filepath.Abs(p); return a }
	reg := filepath.Join(dir, "registry.json")
	body, _ := json.Marshal(map[string]any{
		"echo":        map[string]string{"path": "echo", "contract": abs("../targets/echo/contract.json")},
		"agent":       map[string]string{"path": "agent", "contract": abs("../targets/agent/contract.json")},
		"interpreter": map[string]any{"path": "interp", "contract": abs("../interp/contract.json"), "state": true},
	})
	if err := os.WriteFile(reg, body, 0o600); err != nil {
		t.Fatal(err)
	}
	reach, _, err := registry.Load(reg)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	run := func(text string) string {
		t.Helper()
		n++
		root := filepath.Join(dir, "runs", string(rune('a'+n)), ".skyra")
		payload, _ := json.Marshal(map[string]string{"text": text})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		out, err := exec.CommandContext(ctx, filepath.Join(dir, "skyra"), "boot", "-root", root, "-registry", reg, "-to", "agent", "-payload", string(payload)).CombinedOutput()
		if err != nil {
			t.Fatalf("boot %s: %v\n%s", text, err, out)
		}
		return root
	}
	return run, reach["agent"].Input, reach["echo"].Input
}

// agentReturn is what the agent returned to the operator: the observation
// whose strand is read.
func agentReturn(t *testing.T, root string) (string, record.Record) {
	t.Helper()
	head, err := boot.Head(root)
	if err != nil {
		t.Fatal(err)
	}
	h, _ := tree.Load(root, head)
	r, _ := tree.Load(root, h.Deps[0])
	if r.State != record.Return {
		t.Fatalf("the agent did not return: %+v", r)
	}
	return h.Deps[0], r
}

// strand (next/32, H4): from an observation up to the invoke made through the
// contract the requester names, with every record of the work beneath it.
func TestStrandFromAnObservationToItsInvoke(t *testing.T) {
	run, agentIn, echoIn := room(t)

	root := run("fan:hello,world")
	obs, _ := agentReturn(t, root)
	st, err := replay.ReadStrand(root, obs, "agent", agentIn, "")
	if err != nil {
		t.Fatal(err)
	}
	cause := st.Records[st.Chain[len(st.Chain)-1]]
	if st.Chain[0] != obs || cause.State != record.Invoke || cause.To != "agent" {
		t.Fatalf("the strand runs from the observation to the agent's invoke: %v", st.Chain)
	}
	var echoes []string
	for _, r := range st.Records {
		if r.State == record.Invoke && r.To == "echo" {
			echoes = append(echoes, string(r.Payload))
		}
	}
	slices.Sort(echoes)
	if !reflect.DeepEqual(echoes, []string{`{"text":"hello"}`, `{"text":"world"}`}) || len(st.Records) != 7 {
		t.Fatalf("both siblings are in it, the one off the chain too: %d records, %v", len(st.Records), echoes)
	}
	again, _ := replay.ReadStrand(root, obs, "agent", agentIn, "")
	if !reflect.DeepEqual(st, again) {
		t.Fatal("records are immutable: the same walk gives the same strand")
	}

	// A contract that did not request the observation is refused.
	if _, err := replay.ReadStrand(root, obs, "echo", echoIn, ""); !replay.Refused(err) {
		t.Fatalf("echo did not cause the agent's return: %v", err)
	}
	// A child's return was caused by the agent's invoke too: the walk climbs
	// frames, not predecessors, so it never stops at a sibling's invoke.
	for _, s := range st.Chain {
		if r := st.Records[s]; r.State == record.Return && s != obs {
			sub, err := replay.ReadStrand(root, s, "agent", agentIn, "")
			if err != nil || sub.Chain[len(sub.Chain)-1] != st.Chain[len(st.Chain)-1] {
				t.Fatalf("an echo return's strand reaches the agent's invoke: %v %v", sub.Chain, err)
			}
		}
	}

	// A call into a narrowed child appears as one call and its return; a
	// requester inside that child cannot walk out of it.
	root = run("narrow:hello")
	obs, _ = agentReturn(t, root)
	st, err = replay.ReadStrand(root, obs, "agent", agentIn, "")
	if err != nil || len(st.Opaque) != 1 {
		t.Fatalf("the narrowed child is opaque: %+v %v", st.Opaque, err)
	}
	var call string
	for _, c := range st.Opaque {
		call = c
	}
	if r := st.Records[call]; r.State != record.Invoke || r.To != "system.ref.1" {
		t.Fatalf("the opaque call is the call through the reference: %+v", r)
	}
	if _, err := replay.ReadStrand(root, obs, "agent", agentIn, call); !replay.Refused(err) {
		t.Fatalf("a requester inside the narrowed child cannot read the agent's strand: %v", err)
	}
}

// The research metric (next/32, H8), read from the records of the thesis
// demo: one procedure remembered for the one task, a second remember of the
// same contract refused, and of the two calls to it, the one that returned
// ran with zero model steps and zero approvals (the other, from a Horizon
// without echo, faulted).
func TestTheMetricFromTheRecords(t *testing.T) {
	run, _, _ := room(t)
	root := run("remember:hello")
	m, err := metric.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if m.Tasks != 1 || m.Remembers != 2 || !reflect.DeepEqual(m.Remembered, []string{"agent"}) || m.Calls != 2 || m.Unattended != 1 {
		t.Fatalf("%+v", m)
	}
	if n, err := replay.Replay(root, mustHead(t, root)); err != nil || n < 40 {
		t.Fatalf("and it all replays: %d %v", n, err)
	}
}

func mustHead(t *testing.T, root string) string {
	t.Helper()
	head, err := boot.Head(root)
	if err != nil {
		t.Fatal(err)
	}
	return head
}

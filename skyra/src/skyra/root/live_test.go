package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	p "skyra-v1/src/skyra/protocol"
	"strings"
	"testing"
	"time"
)

// Opt-in only. Configuration is an operator responsibility, never model input.
func TestLiveProvider(t *testing.T) {
	if os.Getenv("DNA19_LIVE") != "1" {
		t.Skip("live provider not opted in; set DNA19_LIVE=1, DNA19_PROVIDER, DNA19_PROVIDER_VERSION and DNA19_PROVIDER_ARGS")
	}
	path, version := os.Getenv("DNA19_PROVIDER"), os.Getenv("DNA19_PROVIDER_VERSION")
	var args []string
	if path == "" || version == "" || json.Unmarshal([]byte(os.Getenv("DNA19_PROVIDER_ARGS")), &args) != nil || len(args) == 0 {
		t.Fatal("explicit provider path, pinned version and JSON launch arguments required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	actual, err := exec.CommandContext(ctx, path, "--version").Output()
	must(t, err)
	if strings.TrimSpace(string(actual)) != version {
		t.Fatalf("version mismatch: %s", actual)
	}
	t.Logf("provider=%s version=%s launch=%q; thread policy: approvalPolicy=never sandbox=read-only network=false; grantable dynamic tools=echo; native-effect fence enabled", path, version, args)
	t.Setenv("SKYRA_CODEX_AUDIT", "1")
	c := buildCodexFixture(t)
	c.Provider = path
	c.ProviderArgs = args
	traceFile, err := os.Create("../artifacts/live-trace.jsonl")
	must(t, err)
	defer traceFile.Close()
	auditFile, err := os.Create("../artifacts/live-provider.stderr")
	must(t, err)
	defer auditFile.Close()
	c.Stderr = auditFile
	trace := &Trace{output: traceFile}
	r, err := bootCodex(ctx, c, trace)
	must(t, err)
	defer r.close()
	next := uint64(1)
	call := func(t *testing.T, args []p.Observation) p.Move {
		t.Helper()
		next++
		m, err := r.call(ctx, next, args)
		must(t, err)
		if m.Identity != next {
			t.Fatalf("wrong command correlation: want %d got %d", next, m.Identity)
		}
		if m.Fault != nil {
			t.Fatalf("live %s fault: %s", args[0].Bytes, m.Fault.Message)
		}
		if m.Return == nil {
			t.Fatal("missing Return")
		}
		return m
	}
	mint := func(t *testing.T) string {
		t.Helper()
		parent, err := r.execution.mintParent(p.Codex)
		must(t, err)
		return parent
	}
	start := func(t *testing.T, parent, prompt, tools string) p.Move {
		t.Helper()
		return call(t, p.Args("start-thread", "parent", parent, "connection", r.connection, "text", prompt, "tools", tools))
	}
	submit := func(t *testing.T, parent, prompt string) p.Move {
		t.Helper()
		return call(t, p.Args("submit-input", "parent", parent, "text", prompt))
	}
	text := func(m p.Move) string { return strings.TrimSpace(string(m.Return.Observation.Bytes)) }
	exact := func(t *testing.T, m p.Move, want string) {
		t.Helper()
		if text(m) != want {
			t.Fatalf("expected exact response %q; got %q", want, text(m))
		}
		t.Logf("response: %s", text(m))
	}
	children := func(t *testing.T, m p.Move, want int, expected string) {
		t.Helper()
		r.execution.mu.Lock()
		defer r.execution.mu.Unlock()
		n := 0
		for _, f := range r.execution.bookkeeper.frames {
			if f.parent == m.Source {
				n++
				if f.target != p.Echo || f.phase != "returned" || string(f.result.Bytes) != expected {
					t.Fatalf("wrong child target/result/phase: %+v", f)
				}
			}
		}
		if n != want {
			t.Fatalf("command %d: expected %d real dependency frames, got %d", m.Source, want, n)
		}
	}
	parents := map[string]string{}
	t.Run("health_and_inference", func(t *testing.T) {
		m := call(t, p.Args("health", "connection", r.connection, "deadline_ms", "5000"))
		expectText(t, m, "responsive")
		parent := mint(t)
		parents["smoke"] = parent
		exact(t, start(t, parent, "Reply with exactly ready. Do not use any tools.", `[]`), "ready")
	})
	t.Run("tool_call_and_correlated_resume", func(t *testing.T) {
		parent := mint(t)
		parents["tool"] = parent
		token := "TOOL_" + randomToken()[:12]
		m := start(t, parent, "Call the registered echo tool exactly once with text "+token+". Wait for its result, then reply with exactly the returned text. Do not use any other tools.", `["echo"]`)
		exact(t, m, token)
		children(t, m, 1, token)
		exact(t, submit(t, parent, "Without using tools, reply with exactly the text returned by the echo call in our previous turn."), token)
		t.Log("one real host child completed, resumed original command, and remained in the conversation")
	})
	t.Run("fork_history_and_independent_mutation", func(t *testing.T) {
		source, branch := mint(t), mint(t)
		parents["fork-source"], parents["fork-branch"] = source, branch
		original, changed := "SOURCE_"+randomToken()[:12], "BRANCH_"+randomToken()[:12]
		exact(t, start(t, source, "Store this conversation's marker as "+original+". Reply with exactly that marker. Do not use tools.", `[]`), original)
		exact(t, call(t, p.Args("fork-thread", "parent", branch, "source", source)), "fork created")
		exact(t, submit(t, branch, "Reply with exactly the stored marker from the preceding conversation history. Do not use tools."), original)
		exact(t, submit(t, branch, "Replace this conversation's marker with "+changed+". Reply with exactly the new marker. Do not use tools."), changed)
		exact(t, submit(t, source, "Reply with exactly this conversation's stored marker. Do not use tools."), original)
		t.Log("fork inherited source history; branch mutation did not change the source")
	})
	t.Run("tool_remove_add_migrates_history", func(t *testing.T) {
		parent := mint(t)
		parents["migration"] = parent
		marker := "MIGRATE_" + randomToken()[:12]
		exact(t, start(t, parent, "Remember this conversation's marker: "+marker+". Reply with exactly that marker. Do not use tools yet.", `["echo"]`), marker)
		exact(t, call(t, p.Args("tools-remove", "parent", parent, "tools", `["echo"]`)), "tools replaced")
		m := submit(t, parent, "Read the retained historical transcript. Reply with exactly the original marker, then a space, then AVAILABLE if a dynamic tool named echo is currently offered to you or UNAVAILABLE if it is not. Do not invoke tools.")
		exact(t, m, marker+" UNAVAILABLE")
		children(t, m, 0, "")
		exact(t, call(t, p.Args("tools-add", "parent", parent, "tools", `["echo"]`)), "tools replaced")
		m = submit(t, parent, "Read the retained historical transcript to find the original marker. Call the registered echo tool exactly once with that marker as its text. Wait for the result and reply with exactly the returned text. Do not use other tools.")
		exact(t, m, marker)
		children(t, m, 1, marker)
		t.Log("stable parent retained transcript across remove/add; removed catalog absent and re-added tool executed through host child")
	})
	t.Run("concurrent_parent_isolation", func(t *testing.T) {
		first, second := mint(t), mint(t)
		parents["isolation-A"], parents["isolation-B"] = first, second
		a, b := "ALPHA_"+randomToken()[:12], "BETA_"+randomToken()[:12]
		exact(t, start(t, first, "Remember this conversation's private marker as "+a+". Reply with exactly that marker. Do not use tools.", `[]`), a)
		exact(t, start(t, second, "Remember this conversation's private marker as "+b+". Reply with exactly that marker. Do not use tools.", `[]`), b)
		expected := map[uint64]string{}
		concurrentIDs := map[uint64]bool{}
		for _, v := range []struct{ parent, marker string }{{first, a}, {second, b}} {
			next++
			expected[next] = v.marker
			concurrentIDs[next] = true
			must(t, r.client.egress.Send(ctx, p.Move{Identity: next, Target: p.Codex, Invoke: &p.Invoke{Args: p.Args("submit-input", "parent", v.parent, "text", "Reply with exactly this conversation's private marker. Do not use tools.")}}))
		}
		for i := 0; i < 2; i++ {
			m, err := r.results.ingress.Receive(ctx)
			must(t, err)
			if m.Fault != nil {
				t.Fatal(m.Fault.Message)
			}
			want, ok := expected[m.Identity]
			if !ok {
				t.Fatalf("unexpected correlation %d", m.Identity)
			}
			delete(expected, m.Identity)
			exact(t, m, want)
		}
		recorded := 0
		for _, event := range trace.snapshot() {
			if !concurrentIDs[event.Move] {
				continue
			}
			if event.Owner == "bookkeeper" && event.Transition == "frame recorded" {
				recorded++
			}
			if event.Transition == "root result sent" && recorded != 2 {
				t.Fatal("commands did not overlap before first result")
			}
		}
		if recorded != 2 {
			t.Fatal("missing concurrent command commits")
		}
		t.Log("two outstanding commands returned only their own stored random marker, with separate result correlations")
	})
	// Wait for the actor to flush its private diagnostics and reap its provider.
	r.close()
	assertCodexClean(t, r)
	must(t, auditFile.Sync())
	raw, err := os.ReadFile("../artifacts/live-provider.stderr")
	must(t, err)
	if !strings.Contains(string(raw), "provider reaped") {
		t.Error("provider reap missing")
	}
	type audit struct {
		Kind   string    `json:"codex_audit"`
		Parent p.FrameID `json:"parent"`
		Thread string    `json:"thread"`
		Old    string    `json:"old_thread"`
		Turn   string    `json:"turn"`
		Tools  []string  `json:"tools"`
	}
	records := map[p.FrameID][]audit{}
	for _, line := range strings.Split(string(raw), "\n") {
		var v audit
		if json.Unmarshal([]byte(line), &v) == nil && v.Kind != "" {
			records[v.Parent] = append(records[v.Parent], v)
		}
	}
	t.Run("provider_binding_evidence", func(t *testing.T) {
		used := map[string]p.FrameID{}
		for label, encoded := range parents {
			var ref p.ParentRef
			must(t, json.Unmarshal([]byte(encoded), &ref))
			events := records[ref.Sequence]
			bindings := []audit{}
			requests, resumes := 0, 0
			var requested *audit
			for _, v := range events {
				switch v.Kind {
				case "start", "fork", "replace":
					if v.Thread == "" {
						t.Fatal("empty provider thread")
					}
					if _, ok := used[v.Thread]; ok {
						t.Fatal("provider thread reused", v.Thread)
					}
					used[v.Thread] = ref.Sequence
					bindings = append(bindings, v)
				case "tool_request":
					requests++
					copy := v
					requested = &copy
				case "tool_resume":
					resumes++
					if requested == nil || requested.Thread != v.Thread || requested.Turn != v.Turn || v.Turn == "" {
						t.Errorf("%s tool resume changed provider thread/turn", label)
					}
				}
			}
			if len(bindings) == 0 {
				t.Errorf("%s has no binding evidence", label)
				continue
			}
			if label == "migration" {
				if len(bindings) != 3 || bindings[1].Old != bindings[0].Thread || bindings[2].Old != bindings[1].Thread || len(bindings[1].Tools) != 0 || len(bindings[2].Tools) != 1 || bindings[2].Tools[0] != "echo" {
					t.Errorf("invalid migration binding chain: %+v", bindings)
				}
			} else if len(bindings) != 1 {
				t.Errorf("%s unexpectedly changed provider threads: %+v", label, bindings)
			}
			if (label == "tool" || label == "migration") && (requests != 1 || resumes != 1) {
				t.Errorf("%s tool request/resume=%d/%d", label, requests, resumes)
			}
			t.Logf("%s parent=%d: distinct bindings=%d tool request/resume=%d/%d", label, ref.Sequence, len(bindings), requests, resumes)
		}
		t.Logf("observed %d distinct provider threads; parent capabilities and prompts omitted from audit", len(used))
	})
	t.Log("live suite finished; observed isolation in these scenarios is not a proof of universal provider sealing")
}

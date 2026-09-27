package main

import (
	"encoding/json"
	"os"
	"reflect"
	p "skyra/history/dna19/protocol"
	"strings"
	"testing"
	"time"
)

func testActor(t *testing.T) (*actor, *[]p.Move) {
	t.Helper()
	out := []p.Move{}
	a, err := newActor("unused", nil, []byte(strings.Repeat("k", 32)), "boot", func(m p.Move) error { out = append(out, m); return nil })
	if err != nil {
		t.Fatal(err)
	}
	a.ready = true
	a.connection = "connection"
	a.connectionOwner = "owner"
	a.provider = &provider{send: make(chan any, 64)}
	return a, &out
}
func refFor(a *actor, owner p.ContextID, n p.FrameID) string {
	r := p.ParentRef{Namespace: a.namespace, Owner: owner, Sequence: n, Generation: 1}
	r.Token = signParent(r, a.key)
	return parentJSON(r)
}
func command(n p.FrameID, args []p.Observation) p.Move {
	return p.Move{Identity: uint64(n) + 100, Context: "owner", Target: p.Codex, Invoke: &p.Invoke{Frame: n, Args: args}}
}
func event(method, thread, turn string, extra map[string]any) rpc {
	v := map[string]any{"threadId": thread, "turnId": turn}
	for k, x := range extra {
		v[k] = x
	}
	raw, _ := json.Marshal(v)
	return rpc{Method: method, Params: raw}
}
func bindSession(a *actor, frame p.FrameID, thread string) *session {
	var ref p.ParentRef
	_ = json.Unmarshal([]byte(refFor(a, "owner", frame+1000)), &ref)
	m := command(frame, p.Args("submit-input", "parent", parentJSON(ref), "text", "input"))
	s := &session{ref: ref, thread: thread, turn: "turn-" + thread, work: &m, tools: []string{"echo"}, seenCalls: map[string]bool{}, deadline: time.Now().Add(time.Minute)}
	a.sessions[ref.Token] = s
	a.threads[thread] = s
	a.usedThreads[thread] = true
	a.commands[commandKey(m)] = s
	return s
}
func TestDescriptorImplementationConformance(t *testing.T) {
	raw, err := os.ReadFile("profiles.json")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(raw, profiles) {
		t.Fatal("embedded descriptor drift")
	}
	a, out := testActor(t)
	parent := refFor(a, "owner", 10)
	samples := map[string][]p.Observation{
		"connect":      p.Args("connect", "configuration", "default", "descriptor", a.digest),
		"health":       p.Args("health", "connection", a.connection, "deadline_ms", "1000"),
		"start-thread": p.Args("start-thread", "parent", parent, "connection", a.connection, "text", "hello", "tools", `["echo"]`),
		"submit-input": p.Args("submit-input", "parent", parent, "text", "next"),
		"fork-thread":  p.Args("fork-thread", "parent", refFor(a, "owner", 11), "source", parent),
		"tools-add":    p.Args("tools-add", "parent", parent, "tools", `["echo"]`),
		"tools-remove": p.Args("tools-remove", "parent", parent, "tools", `["echo"]`),
		"end-parent":   p.Args("end-parent", "parent", parent),
		"tool-result":  p.Args("tool-result", "parent", parent, "pending", "nonce", "text", "result"),
	}
	if len(samples) != len(a.descriptor.Profiles) {
		t.Fatal("inventory drift")
	}
	for _, profile := range a.descriptor.Profiles {
		args := samples[profile.Operation]
		if err := profile.Check(args, profile.Continuation); err != nil {
			t.Fatalf("%s: %v", profile.Operation, err)
		}
		if err := profile.Check(append(args, p.Observation{Key: "unexpected", Bytes: []byte("x")}), profile.Continuation); err == nil {
			t.Fatal("unknown field accepted")
		}
	}
	if err := a.handle(command(1, samples["tool-result"])); err != nil {
		t.Fatal(err)
	}
	if len(*out) != 1 || (*out)[0].Fault == nil {
		t.Fatal("continuation admitted as input")
	}
}
func TestAuthenticatedReferencesAndIncarnations(t *testing.T) {
	a, _ := testActor(t)
	raw := refFor(a, "owner", 1)
	var r p.ParentRef
	_ = json.Unmarshal([]byte(raw), &r)
	for _, change := range []func(*p.ParentRef){func(r *p.ParentRef) { r.Namespace = "other-host" }, func(r *p.ParentRef) { r.Owner = "other" }, func(r *p.ParentRef) { r.Sequence++ }, func(r *p.ParentRef) { r.Generation++ }, func(r *p.ParentRef) { r.Token = "forged" }} {
		altered := r
		change(&altered)
		if _, err := a.reference(parentJSON(altered), "owner"); err == nil {
			t.Fatal("forgery accepted", altered)
		}
	}
	if _, err := a.reference(raw, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.reference(raw, "other"); err == nil {
		t.Fatal("foreign owner collision accepted")
	}
}
func TestDemuxContaminationStickyAndChildResume(t *testing.T) {
	a, out := testActor(t)
	s := bindSession(a, 1, "one")
	other := bindSession(a, 2, "two")
	tool := event("item/tool/call", s.thread, s.turn, map[string]any{"tool": "echo", "callId": "call-1", "arguments": map[string]string{"text": "dependency"}})
	tool.ID = json.RawMessage(`"request-1"`)
	if err := a.event(tool); err != nil {
		t.Fatal(err)
	}
	if len(*out) != 1 || (*out)[0].Invoke == nil || (*out)[0].Source != 1 {
		t.Fatal("child routing", out)
	}
	nonce := append([]byte(nil), (*out)[0].State...)
	// A forged nonce cannot answer the outstanding provider request.
	if err := a.resume(p.Move{Identity: 55, Context: "owner", Source: 1, State: []byte("forged"), Return: &p.Return{Answering: 3}}); err != nil {
		t.Fatal(err)
	}
	if s.pending == nil || len(a.provider.send) != 0 {
		t.Fatal("forged result resumed provider")
	}
	if err := a.event(event("item/started", s.thread, s.turn, map[string]any{"item": map[string]string{"type": "commandExecution"}})); err != nil {
		t.Fatal(err)
	}
	if err := a.resume(p.Move{Identity: 56, Context: "owner", Source: 1, State: nonce, Return: &p.Return{Answering: 3}}); err != nil {
		t.Fatal(err)
	}
	_ = a.event(event("item/completed", s.thread, s.turn, map[string]any{"item": map[string]string{"type": "agentMessage", "text": "clean"}}))
	_ = a.event(event("turn/completed", s.thread, s.turn, map[string]any{"turn": map[string]string{"id": s.turn, "status": "completed"}}))
	if len(*out) != 2 || (*out)[1].Fault == nil || !s.retired || other.retired || len(a.provider.send) != 0 {
		t.Fatal("taint cleared or leaked", out)
	}
	_ = a.event(event("item/completed", other.thread, other.turn, map[string]any{"item": map[string]string{"type": "agentMessage", "text": "other result"}}))
	_ = a.event(event("turn/completed", other.thread, other.turn, map[string]any{"turn": map[string]string{"id": other.turn, "status": "completed"}}))
	if len(*out) != 3 || (*out)[2].Source != 2 || (*out)[2].Return == nil {
		t.Fatal("other parent blocked", out)
	}
}
func TestCorrelatedResumeOnceAndReplayRejection(t *testing.T) {
	a, out := testActor(t)
	s := bindSession(a, 1, "thread")
	call := event("item/tool/call", s.thread, s.turn, map[string]any{"tool": "echo", "callId": "call", "arguments": map[string]string{"text": "child"}})
	call.ID = json.RawMessage(`44`)
	_ = a.event(call)
	resume := p.Move{Identity: 1, Context: "owner", Source: 1, State: (*out)[0].State, Return: &p.Return{Answering: 9, Observation: p.Observation{Key: "text", Bytes: []byte("answer")}}}
	_ = a.resume(resume)
	_ = a.resume(resume)
	if len(a.provider.send) != 1 {
		t.Fatal("duplicate/missing tool reply")
	}
	reply := <-a.provider.send
	raw, _ := json.Marshal(reply)
	if !strings.Contains(string(raw), `"id":44`) || !strings.Contains(string(raw), "answer") {
		t.Fatal(string(raw))
	}
	if s.thread != "thread" || s.work == nil {
		t.Fatal("resume changed thread or settled command")
	}
	_ = a.event(call)
	if !s.retired {
		t.Fatal("replayed provider call accepted")
	}
}
func TestUnknownAndConflictingCorrelations(t *testing.T) {
	for _, mode := range []string{"unknown-thread", "wrong-turn", "unknown-request", "forbidden-tool"} {
		t.Run(mode, func(t *testing.T) {
			a, out := testActor(t)
			s := bindSession(a, 1, "thread")
			m := event("item/completed", s.thread, s.turn, map[string]any{"item": map[string]string{"type": "agentMessage", "text": "data"}})
			switch mode {
			case "unknown-thread":
				m = event("turn/completed", "missing", s.turn, nil)
			case "wrong-turn":
				m = event("item/completed", s.thread, "stale", nil)
			case "unknown-request":
				m = event("approval/unknown", s.thread, s.turn, nil)
				m.ID = json.RawMessage(`3`)
			case "forbidden-tool":
				m = event("item/tool/call", s.thread, s.turn, map[string]any{"tool": "shell", "callId": "x"})
				m.ID = json.RawMessage(`3`)
			}
			_ = a.event(m)
			if mode == "unknown-thread" {
				if len(*out) != 0 || s.retired {
					t.Fatal("unknown thread affected live work")
				}
			} else {
				if len(*out) != 1 || (*out)[0].Fault == nil || !s.retired {
					t.Fatal("conflicting event accepted")
				}
			}
			if len(a.provider.send) != 0 {
				t.Fatal("invented approval")
			}
		})
	}
}
func TestIdleOnlyMigrationAndBoundedHistory(t *testing.T) {
	a, out := testActor(t)
	s := bindSession(a, 1, "old")
	args := p.Args("tools-remove", "parent", parentJSON(s.ref), "tools", `["echo"]`)
	_ = a.handle(command(2, args))
	if len(*out) != 1 || (*out)[0].Fault == nil || len(a.provider.send) != 0 {
		t.Fatal("busy migration dispatched")
	}
	delete(a.commands, commandKey(*s.work))
	s.work = nil
	s.history = []historyEntry{{"input", "remember"}, {"tool-result", "side effect already done"}, {"response", "answer"}}
	_ = a.handle(command(3, args))
	msg := <-a.provider.send
	raw, _ := json.Marshal(msg)
	if !strings.Contains(string(raw), "remember") || !strings.Contains(string(raw), "side effect already done") || !strings.Contains(string(raw), `"dynamicTools":[]`) {
		t.Fatal("migration lost transcript or tools", string(raw))
	}
	var sent struct {
		ID json.RawMessage `json:"id"`
	}
	_ = json.Unmarshal(raw, &sent)
	_ = a.event(rpc{ID: sent.ID, Result: json.RawMessage(`{"thread":{"id":"new"},"approvalPolicy":"never","sandbox":{"type":"readOnly"}}`)})
	if s.thread != "new" || len(s.tools) != 0 || a.threads["old"] != nil || s.ref.Token == "" {
		t.Fatal("replacement not committed")
	}
	before := len(*out)
	_ = a.event(event("item/tool/call", "old", "old-turn", nil))
	if len(*out) != before {
		t.Fatal("old binding resumed")
	}
	s.history = []historyEntry{{"input", strings.Repeat("x", 65537)}}
	_ = a.handle(command(4, args))
	if len(a.provider.send) != 0 || s.thread != "new" || s.work != nil {
		t.Fatal("oversized migration committed")
	}
}

func TestRemovedToolRevocationAndDuplicateCommand(t *testing.T) {
	a, out := testActor(t)
	s := bindSession(a, 1, "thread")
	s.tools = nil
	call := event("item/tool/call", s.thread, s.turn, map[string]any{"tool": "echo", "callId": "removed", "arguments": map[string]string{"text": "bad"}})
	call.ID = json.RawMessage(`5`)
	_ = a.event(call)
	if len(*out) != 1 || (*out)[0].Invoke != nil || (*out)[0].Fault == nil {
		t.Fatal("removed tool executed")
	}
	m := command(8, p.Args("health", "connection", a.connection, "deadline_ms", "1000"))
	_ = a.handle(m)
	_ = a.handle(m)
	if len(a.provider.send) != 1 || len(a.requests) != 1 {
		t.Fatal("duplicate command dispatched")
	}
}

func TestEffectiveProviderPolicyMismatch(t *testing.T) {
	a, out := testActor(t)
	s := bindSession(a, 1, "old")
	a.requests["1"] = request{kind: "replace", session: s, move: *s.work}
	_ = a.event(rpc{ID: json.RawMessage(`1`), Result: json.RawMessage(`{"thread":{"id":"new"},"approvalPolicy":"never","sandbox":{"type":"dangerFullAccess"}}`)})
	if !s.retired || len(*out) != 1 || (*out)[0].Fault == nil || !strings.Contains((*out)[0].Fault.Message, "policy mismatch") {
		t.Fatal("provider widened effective policy")
	}
}

func TestMCPStartupStatusDoesNotSettleOrAuthorize(t *testing.T) {
	a, out := testActor(t)
	s := bindSession(a, 1, "thread")
	notice := event("mcpServer/startupStatus/updated", s.thread, "", map[string]any{"name": "configured-server", "status": "ready"})
	if err := a.event(notice); err != nil {
		t.Fatal(err)
	}
	if len(*out) != 0 || s.work == nil || s.retired || len(a.provider.send) != 0 {
		t.Fatal("startup notice altered command or authorized work")
	}
	notice.ID = json.RawMessage(`8`)
	if err := a.event(notice); err != nil {
		t.Fatal(err)
	}
	if !s.retired || len(*out) != 1 || (*out)[0].Fault == nil || len(a.provider.send) != 0 {
		t.Fatal("request disguised as startup notice accepted")
	}
}

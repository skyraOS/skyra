package main

import "testing"

func find[T any](t *testing.T, out []any) (T, bool) {
	t.Helper()
	for _, m := range out {
		if v, ok := m.(T); ok {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func mustReject(t *testing.T, out []any, code string) {
	t.Helper()
	r, ok := find[rejectMessage](t, out)
	if !ok || r.Code != code {
		t.Fatalf("want rejected %q, got %+v", code, out)
	}
}

func newReady() *engine {
	e := newEngine("ui-test")
	e.activate()
	return e
}

func submitMsg(id, text string) inboundMessage {
	return inboundMessage{Type: "submit", UIIncarnation: "ui-test", Conversation: "c1", RequestID: id, Text: text}
}

func TestSubmitAcceptedThenReturned(t *testing.T) {
	e := newReady()
	out, child := e.submit(submitMsg("r1", "hello"))
	if _, ok := find[ackMessage](t, out); !ok {
		t.Fatal("expected accepted ack")
	}
	if child == nil || child.text != "hello" {
		t.Fatal("expected a child request carrying the text")
	}
	if e.phase != "pending" || e.pending != "r1" {
		t.Fatalf("expected pending on r1, got phase=%s pending=%s", e.phase, e.pending)
	}
	// accepted is not completion: the agent entry appears only on childReturn.
	if len(e.entries) != 1 || e.entries[0].Role != "user" {
		t.Fatalf("expected only the user entry, got %+v", e.entries)
	}
	e.childReturn("r1", "hello")
	if e.phase != "ready" || e.pending != "" {
		t.Fatalf("expected ready after return, got phase=%s", e.phase)
	}
	if len(e.entries) != 2 || e.entries[1].Role != "agent" || e.entries[1].Text != "hello" {
		t.Fatalf("expected an agent entry, got %+v", e.entries)
	}
	if e.ledger["r1"].Status != "returned" {
		t.Fatalf("expected r1 returned, got %s", e.ledger["r1"].Status)
	}
}

func TestOnePendingCommandRejectsSecond(t *testing.T) {
	e := newReady()
	e.submit(submitMsg("r1", "one"))
	out, child := e.submit(submitMsg("r2", "two"))
	if child != nil {
		t.Fatal("second command must not dispatch while pending")
	}
	mustReject(t, out, "busy")
}

func TestLedgerIdempotencyAndConflict(t *testing.T) {
	e := newReady()
	e.submit(submitMsg("r1", "hello"))
	e.childReturn("r1", "hello")
	// same id + same text replays the recorded status, emits no new work.
	out, child := e.submit(submitMsg("r1", "hello"))
	if child != nil {
		t.Fatal("replayed id must not dispatch new work")
	}
	if _, ok := find[ackMessage](t, out); !ok {
		t.Fatalf("replay should ack, got %+v", out)
	}
	// same id + different text is a conflict.
	out, _ = e.submit(submitMsg("r1", "changed"))
	mustReject(t, out, "conflict")
}

func TestFaultRetiresAnchor(t *testing.T) {
	e := newReady()
	e.submit(submitMsg("r1", "boom"))
	e.childFault("r1", "contamination: native effect")
	if e.phase != "failed" {
		t.Fatalf("expected failed, got %s", e.phase)
	}
	out, child := e.submit(submitMsg("r2", "after"))
	if child != nil {
		t.Fatal("no dispatch after fault")
	}
	mustReject(t, out, "faulted")
}

func TestCloseAndNotReady(t *testing.T) {
	cold := newEngine("ui-test")
	out, _ := cold.submit(submitMsg("r1", "hi"))
	mustReject(t, out, "not-ready")

	e := newReady()
	_, terminal := e.close(inboundMessage{Type: "close-interaction", UIIncarnation: "ui-test", Conversation: "c1", RequestID: "x"})
	if !terminal || e.phase != "closed" {
		t.Fatalf("expected terminal close, got terminal=%v phase=%s", terminal, e.phase)
	}
	// close while pending is refused.
	e2 := newReady()
	e2.submit(submitMsg("r1", "one"))
	out, terminal = e2.close(inboundMessage{Type: "close-interaction", UIIncarnation: "ui-test", Conversation: "c1", RequestID: "x"})
	if terminal {
		t.Fatal("must not close while pending")
	}
	mustReject(t, out, "busy")
}

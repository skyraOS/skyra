package main

import "time"

// engine is the pure interaction state machine. It owns one conversation (U)
// for this first proof. It never fabricates an agent answer: an agent entry is
// appended only by childReturn/childFault, driven by a real host Move. It is
// single-goroutine state, mutated only from the main select loop.
type engine struct {
	incarnation  string
	conversation string
	revision     int
	ready        bool // U established by the `interact` Invoke
	phase        string
	entries      []entry
	pending      string // requestId of the one outstanding command, "" if none

	ledger      map[string]*requestRecord
	ledgerText  map[string]string
	ledgerOrder []string
}

func newEngine(incarnation string) *engine {
	return &engine{
		incarnation:  incarnation,
		conversation: "c1",
		phase:        "ready",
		ledger:       map[string]*requestRecord{},
		ledgerText:   map[string]string{},
	}
}

func now() int64 { return time.Now().UnixMilli() }

// activate marks U established (the `interact` command arrived).
func (e *engine) activate() { e.ready = true; e.revision++ }

func (e *engine) record(id, status, code, message, text string) *requestRecord {
	if _, seen := e.ledger[id]; !seen {
		e.ledgerOrder = append(e.ledgerOrder, id)
	}
	r := &requestRecord{RequestID: id, Status: status, Code: code, Message: message, Ts: now()}
	e.ledger[id] = r
	e.ledgerText[id] = text
	return r
}

func reject(id, code, message string) rejectMessage {
	return rejectMessage{Type: "rejected", RequestID: id, Code: code, Message: message}
}

// childRequest asks the main loop to emit a child Invoke to the agent.
type childRequest struct {
	requestID string
	text      string
}

// submit applies a browser submit. It returns the browser messages to send and,
// if the command was accepted, a childRequest to dispatch to the agent.
func (e *engine) submit(m inboundMessage) ([]any, *childRequest) {
	id := m.RequestID
	if id == "" {
		return nil, nil
	}
	rej := func(code, message string) []any {
		e.record(id, "rejected", code, message, m.Text)
		return []any{reject(id, code, message)}
	}

	if !e.ready {
		return rej("not-ready", "the interaction is not open yet"), nil
	}
	if m.UIIncarnation != e.incarnation {
		return rej("incarnation", "stale UI incarnation; reload"), nil
	}
	if m.Conversation != e.conversation {
		return rej("unknown", "no such conversation"), nil
	}

	// Idempotent ledger: a re-sent id replays its status; a reused id with
	// different contents is rejected without recording over the original.
	if prior, seen := e.ledger[id]; seen {
		if e.ledgerText[id] != m.Text {
			return []any{reject(id, "conflict", "request id reused with different contents")}, nil
		}
		if prior.Status == "rejected" {
			return []any{reject(id, prior.Code, prior.Message)}, nil
		}
		return []any{ackMessage{Type: "accepted", RequestID: id}}, nil
	}

	switch {
	case e.phase == "closed":
		return rej("closed", "interaction is closed"), nil
	case e.phase == "failed":
		return rej("faulted", "interaction faulted; open a new one"), nil
	case e.phase == "pending" || e.pending != "":
		return rej("busy", "one agent command at a time"), nil
	case len(m.Text) == 0:
		return rej("empty", "empty message"), nil
	}

	// Accept into the ledger (not completion), append the human entry, mark the
	// anchor pending, and ask the loop to run the agent child.
	e.record(id, "accepted", "", "", m.Text)
	e.entries = append(e.entries, entry{Role: "user", Text: m.Text, Ts: now(), RequestID: id})
	e.phase = "pending"
	e.pending = id
	e.revision++
	return []any{ackMessage{Type: "accepted", RequestID: id}, e.snapshot()}, &childRequest{requestID: id, text: m.Text}
}

// childReturn settles the outstanding command with the agent's result.
func (e *engine) childReturn(requestID, text string) []any {
	if e.pending != requestID {
		return nil // superseded, closed, or unknown correlation
	}
	e.entries = append(e.entries, entry{Role: "agent", Text: text, Ts: now(), RequestID: requestID})
	e.phase = "ready"
	e.pending = ""
	e.record(requestID, "returned", "", "", e.ledgerText[requestID])
	e.revision++
	return []any{e.snapshot()}
}

// childFault retires the anchor on an agent/provider fault. The taint is
// sticky: the phase stays failed and further input is refused.
func (e *engine) childFault(requestID, message string) []any {
	if e.pending != requestID {
		return nil
	}
	e.entries = append(e.entries, entry{Role: "agent", Text: "⚠ agent fault — anchor retired: " + message, Ts: now(), RequestID: requestID})
	e.phase = "failed"
	e.pending = ""
	e.record(requestID, "failed", "fault", message, e.ledgerText[requestID])
	e.revision++
	return []any{e.snapshot()}
}

// close applies a browser close-interaction. terminal reports whether U should
// now emit its single terminal Return to the host.
func (e *engine) close(m inboundMessage) (out []any, terminal bool) {
	if !e.ready || m.UIIncarnation != e.incarnation || m.Conversation != e.conversation {
		return nil, false
	}
	if e.phase == "pending" {
		return []any{reject(m.RequestID, "busy", "cannot close while a command is pending")}, false
	}
	if e.phase == "closed" {
		return nil, false
	}
	e.phase = "closed"
	e.pending = ""
	e.revision++
	return []any{e.snapshot()}, true
}

func (e *engine) snapshot() snapshotMessage {
	var pending *string
	if e.pending != "" {
		p := e.pending
		pending = &p
	}
	entries := append([]entry(nil), e.entries...)
	var convos []conversationView
	if e.ready {
		convos = []conversationView{{
			Conversation:     e.conversation,
			Phase:            e.phase,
			Entries:          entries,
			PendingRequestID: pending,
		}}
	}
	reqs := make([]requestRecord, 0, len(e.ledgerOrder))
	for _, id := range e.ledgerOrder {
		reqs = append(reqs, *e.ledger[id])
	}
	return snapshotMessage{
		Type:          "snapshot",
		UIIncarnation: e.incarnation,
		Revision:      e.revision,
		Conversations: convos,
		Requests:      reqs,
	}
}

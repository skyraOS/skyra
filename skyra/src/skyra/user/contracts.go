// Package main implements the `user` executable: a root-registered external
// peer that serves a browser over WebSocket, turns human input into a child
// Invoke against an agent, and resumes the persistent interaction U with the
// agent's Return. It speaks framed Skyra Moves over stdio to the host runtime
// and the scoped conversation protocol (not Moves) to the browser.
//
// This first slice drives the `echo` agent, so the whole
// browser -> user -> root -> echo -> Return -> user -> browser loop runs with
// no provider. It does NOT yet use the persistent-parent/Codex machinery,
// U-anchored Bookkeeper provisioning, or authenticated activation arguments;
// those are the next stages of specs/2026-09-22/frontend.md.
package main

import "encoding/json"

// ---------------------------------------------------------------------------
// Browser wire types — mirror of frontend/lib/protocol/types.ts.
// These are NOT Skyra Move variants.
// ---------------------------------------------------------------------------

type entry struct {
	Role      string `json:"role"` // "user" | "agent"
	Text      string `json:"text"`
	Ts        int64  `json:"ts"`
	RequestID string `json:"requestId"`
}

type requestRecord struct {
	RequestID string `json:"requestId"`
	Status    string `json:"status"` // accepted | returned | failed | rejected
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Ts        int64  `json:"ts"`
}

type conversationView struct {
	Conversation     string  `json:"conversation"`
	Phase            string  `json:"phase"` // ready | pending | failed | closed
	Entries          []entry `json:"entries"`
	PendingRequestID *string `json:"pendingRequestId"`
}

// server -> browser
type snapshotMessage struct {
	Type          string             `json:"type"` // "snapshot"
	UIIncarnation string             `json:"uiIncarnation"`
	Revision      int                `json:"revision"`
	Conversations []conversationView `json:"conversations"`
	Requests      []requestRecord    `json:"requests"`
}

type ackMessage struct {
	Type      string `json:"type"` // "accepted"
	RequestID string `json:"requestId"`
}

type rejectMessage struct {
	Type      string `json:"type"` // "rejected"
	RequestID string `json:"requestId"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

// browser -> user
type inboundMessage struct {
	Type          string `json:"type"` // "submit" | "close-interaction"
	UIIncarnation string `json:"uiIncarnation"`
	Conversation  string `json:"conversation"`
	RequestID     string `json:"requestId"`
	Text          string `json:"text"`
}

// continuation is the opaque parent-continuation the user serializes into
// Move.State on a child Invoke; the runtime returns it verbatim on resume, so
// the child result can be routed back to the request that caused it.
type continuation struct {
	Conversation string `json:"c"`
	RequestID    string `json:"r"`
}

func encodeContinuation(c continuation) []byte { b, _ := json.Marshal(c); return b }
func decodeContinuation(b []byte) (continuation, error) {
	var c continuation
	err := json.Unmarshal(b, &c)
	return c, err
}

// agentConfig is the trusted provisioning the host packs into the `interact`
// activation: which agent U drives and, for Codex, the opaque connection and
// the authenticated parent reference the host minted for this interaction. The
// user relays these verbatim; it never mints or forges them. An empty text
// selects the default echo agent.
type agentConfig struct {
	Agent      string `json:"agent"` // "echo" | "codex"
	Connection string `json:"connection,omitempty"`
	Parent     string `json:"parent,omitempty"`
	Tools      string `json:"tools,omitempty"` // JSON array, e.g. ["echo"]
}

func decodeAgentConfig(b []byte) agentConfig {
	cfg := agentConfig{Agent: "echo"}
	if len(b) > 0 {
		_ = json.Unmarshal(b, &cfg)
	}
	if cfg.Agent == "" {
		cfg.Agent = "echo"
	}
	return cfg
}

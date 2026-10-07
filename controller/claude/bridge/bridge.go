// Package bridge exposes only an invocation's reachable tools to Claude.
// Call must deliver through Skyra's recorded actor edge, never execute tools.
// The bridge has no filesystem or process authority.
package bridge

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type Tool struct {
	Name string `json:"name"`
	Description string `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type Call func(context.Context, string, string, json.RawMessage) (json.RawMessage, error)

type entry struct {
	payload string
	done chan struct{}
	result json.RawMessage
	err error
}

// Server is scoped to exactly one recorded turn. IDs are namespaced by Frame.
// Token is an ephemeral loopback capability, not an Anthropic credential.
// Keep it out of continuation records and logs.
type Server struct {
	frame, token string
	tools []Tool
	call Call
	mu sync.Mutex
	calls map[string]*entry
}

func New(frame, token string, tools []Tool, call Call) (*Server, error) {
	if frame == "" || len(token) < 32 || call == nil { return nil, errors.New("bridge requires frame, capability and recorded call handler") }
	seen := make(map[string]bool)
	copyTools := make([]Tool, len(tools))
	for i, t := range tools {
		if t.Name == "" || strings.ContainsAny(t.Name, " /\\") || seen[t.Name] || !json.Valid(t.InputSchema) { return nil, errors.New("invalid bridge tool catalog") }
		var schema map[string]any
		if json.Unmarshal(t.InputSchema, &schema) != nil || schema["type"] != "object" { return nil, errors.New("tool schema must describe an object") }
		seen[t.Name] = true
		t.InputSchema = append(json.RawMessage(nil), t.InputSchema...)
		copyTools[i] = t
	}
	sort.Slice(copyTools, func(i,j int) bool { return copyTools[i].Name < copyTools[j].Name })
	return &Server{frame:frame, token:token, tools:copyTools, call:call, calls:make(map[string]*entry)}, nil
}

type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID json.RawMessage `json:"id"`
	Method string `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// No browser origins: this endpoint is a local provider capability only.
	if r.Header.Get("Origin") != "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.token)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized); return
	}
	if r.Method == http.MethodGet { w.WriteHeader(http.StatusMethodNotAllowed); return }
	if r.Method != http.MethodPost { w.WriteHeader(http.StatusMethodNotAllowed); return }
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil { http.Error(w, "request too large", http.StatusRequestEntityTooLarge); return }
	var q request
	if json.Unmarshal(body, &q) != nil || q.JSONRPC != "2.0" || q.Method == "" {
		s.reply(w, nil, nil, -32600, "invalid request"); return
	}
	if len(q.ID) == 0 {
		if q.Method != "notifications/initialized" { http.Error(w, "unsupported notification", http.StatusBadRequest); return }
		w.WriteHeader(http.StatusAccepted); return
	}
	var id any
	if json.Unmarshal(q.ID, &id) != nil || id == nil { s.reply(w, nil, nil, -32600, "invalid request ID"); return }
	switch id.(type) { case string, float64: default: s.reply(w, nil, nil, -32600, "invalid request ID"); return }
	switch q.Method {
	case "initialize":
		var p struct { ProtocolVersion string `json:"protocolVersion"` }
		if json.Unmarshal(q.Params, &p) != nil { s.reply(w, q.ID, nil, -32602, "invalid initialize params"); return }
		version := "2025-03-26"
		if p.ProtocolVersion == "2025-06-18" { version = p.ProtocolVersion }
		s.reply(w, q.ID, map[string]any{"protocolVersion":version,"capabilities":map[string]any{"tools":map[string]any{}},"serverInfo":map[string]string{"name":"skyra","version":"1"}}, 0, "")
	case "ping":
		s.reply(w, q.ID, map[string]any{}, 0, "")
	case "tools/list":
		s.reply(w, q.ID, map[string]any{"tools":s.tools}, 0, "")
	case "tools/call":
		var p struct { Name string `json:"name"`; Arguments json.RawMessage `json:"arguments"` }
		if json.Unmarshal(q.Params, &p) != nil { s.reply(w, q.ID, nil, -32602, "invalid tool params"); return }
		found := false
		for _, t := range s.tools { if t.Name == p.Name { found = true; break } }
		if !found { s.reply(w, q.ID, nil, -32602, "tool is outside reach"); return }
		if len(p.Arguments) == 0 { p.Arguments = json.RawMessage(`{}`) }
		var args map[string]json.RawMessage
		if json.Unmarshal(p.Arguments, &args) != nil || args == nil { s.reply(w, q.ID, nil, -32602, "arguments must be an object"); return }
		canonical, _ := json.Marshal(args)
		value, err := s.deliver(r.Context(), string(q.ID), p.Name, canonical)
		if err != nil {
			// Do not expose callback errors: upstream errors can contain secrets.
			s.reply(w, q.ID, map[string]any{"content":[]any{map[string]string{"type":"text","text":"recorded tool delivery failed"}},"isError":true}, 0, ""); return
		}
		s.reply(w, q.ID, map[string]any{"content":[]any{map[string]string{"type":"text","text":string(value)}}}, 0, "")
	default:
		s.reply(w, q.ID, nil, -32601, "method not found")
	}
}

func (s *Server) deliver(ctx context.Context, id, name string, args json.RawMessage) (json.RawMessage, error) {
	payload := name + ":" + string(args)
	s.mu.Lock()
	e, exists := s.calls[id]
	if exists && e.payload != payload { s.mu.Unlock(); return nil, errors.New("delivery ID reused with different payload") }
	if !exists { e = &entry{payload:payload, done:make(chan struct{})}; s.calls[id] = e }
	s.mu.Unlock()
	if !exists {
		// Call's ID is stable across retries. Durable deduplication belongs to
		// the recorded holder; this table coalesces concurrent HTTP retries.
		e.result, e.err = s.call(ctx, s.frame+":mcp:"+id, name, args)
		if e.err == nil && !json.Valid(e.result) { e.err = errors.New("recorded result is not JSON") }
		close(e.done)
	}
	select {
	case <-ctx.Done(): return nil, ctx.Err()
	case <-e.done: return e.result, e.err
	}
}

func (s *Server) reply(w http.ResponseWriter, id json.RawMessage, result any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	v := map[string]any{"jsonrpc":"2.0","id":id}
	if code != 0 { v["error"] = map[string]any{"code":code,"message":message} } else { v["result"] = result }
	_ = json.NewEncoder(w).Encode(v)
}

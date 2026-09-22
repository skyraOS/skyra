package main

import (
	"encoding/json"
	"fmt"
	p "skyra-v1/src/skyra/skyra-dna-19/protocol"
)

func (a *actor) event(m rpc) error {
	if m.Method == "" {
		r, ok := a.requests[string(m.ID)]
		if !ok {
			return nil
		}
		delete(a.requests, string(m.ID))
		if m.Error != nil {
			if r.session != nil {
				if r.kind == "replace" {
					return a.commandFailure(r.session, "unsupported", m.Error.Message)
				}
				return a.retire(r.session, "provider", m.Error.Message)
			}
			return a.fault(r.move, "provider", m.Error.Message)
		}
		if r.session != nil && (r.session.retired || r.session.work == nil) {
			return nil
		}
		switch r.kind {
		case "connect":
			var result map[string]any
			if json.Unmarshal(m.Result, &result) != nil || result["userAgent"] == nil {
				return a.fault(r.move, "provider", "invalid handshake")
			}
			if err := a.send(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
				return err
			}
			a.ready = true
			a.connection = randomToken()
			return a.output(r.move, "connection", a.connection)
		case "health":
			var result struct {
				Data []string `json:"data"`
			}
			if json.Unmarshal(m.Result, &result) != nil || result.Data == nil {
				return a.fault(r.move, "provider", "invalid health response")
			}
			return a.output(r.move, "health", "responsive; inference readiness untested")
		case "start", "replace", "fork":
			var result struct {
				ApprovalPolicy string `json:"approvalPolicy"`
				Sandbox        struct {
					Type    string `json:"type"`
					Network bool   `json:"networkAccess"`
				} `json:"sandbox"`
				Thread struct {
					ID string `json:"id"`
				} `json:"thread"`
			}
			if json.Unmarshal(m.Result, &result) != nil || result.Thread.ID == "" {
				return a.retire(r.session, "provider", "missing thread identity")
			}
			if result.ApprovalPolicy != "never" || result.Sandbox.Type != "readOnly" || result.Sandbox.Network {
				return a.retire(r.session, "provider", "effective provider policy mismatch")
			}
			s := r.session
			id := result.Thread.ID
			if a.usedThreads[id] {
				return a.retire(s, "provider", "provider reused live thread")
			}
			oldThread := s.thread
			delete(a.threads, s.thread)
			a.usedThreads[id] = true
			s.thread = id
			a.threads[id] = s
			if r.kind == "replace" {
				s.tools = append([]string(nil), r.tools...)
				a.audit(r.kind, s, oldThread)
				return a.settle(s, "tools replaced")
			}
			a.audit(r.kind, s, oldThread)
			if r.kind == "fork" {
				return a.settle(s, "fork created")
			}
			return a.turn(s, s.history[len(s.history)-1].Text)
		case "turn":
			var result struct {
				Turn struct {
					ID string `json:"id"`
				} `json:"turn"`
			}
			if json.Unmarshal(m.Result, &result) != nil || result.Turn.ID == "" {
				return a.retire(r.session, "provider", "missing turn identity")
			}
			if r.session.turn != "" && r.session.turn != result.Turn.ID {
				return a.retire(r.session, "provider", "conflicting turn identity")
			}
			r.session.turn = result.Turn.ID
			a.audit("turn", r.session, "")
		}
		return nil
	}
	var ev struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Turn     struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"turn"`
		Item struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"item"`
		Tool      string          `json:"tool"`
		CallID    string          `json:"callId"`
		Namespace *string         `json:"namespace"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if json.Unmarshal(m.Params, &ev) != nil {
		return nil
	}
	s := a.threads[ev.ThreadID]
	if s == nil || s.retired || s.work == nil {
		return nil
	}
	if len(m.ID) > 0 && m.Method != "item/tool/call" {
		return a.retire(s, "contamination", "unsupported provider request: "+m.Method)
	}
	// Administrative notices have no turn binding and cannot settle or execute work.
	switch m.Method {
	case "thread/status/changed", "thread/tokenUsage/updated":
		return nil
	case "mcpServer/startupStatus/updated":
		// CLI 0.155.1 declares this as a thread-scoped status notification,
		// not an effect-bearing tool call or a turn outcome.
		var status struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		}
		if json.Unmarshal(m.Params, &status) != nil || status.Name == "" {
			return a.retire(s, "contamination", "malformed MCP startup status")
		}
		switch status.Status {
		case "starting", "ready", "failed", "cancelled":
			return nil
		}
		return a.retire(s, "contamination", "unknown MCP startup status")
	}
	turn := ev.TurnID
	if turn == "" {
		turn = ev.Turn.ID
	}
	if turn == "" {
		return a.retire(s, "contamination", "event missing turn correlation: "+m.Method)
	}
	if s.turn == "" {
		if m.Method != "turn/started" {
			return a.retire(s, "contamination", "event preceded turn binding")
		}
		s.turn = turn
	}
	if s.turn != turn {
		return a.retire(s, "contamination", "conflicting turn correlation")
	}
	switch m.Method {
	case "turn/started", "item/agentMessage/delta", "item/reasoning/textDelta", "item/reasoning/summaryTextDelta", "turn/plan/updated":
		return nil
	case "item/started", "item/completed":
		switch ev.Item.Type {
		case "agentMessage":
			if m.Method == "item/completed" {
				s.final = ev.Item.Text
			}
		case "userMessage", "reasoning", "plan", "dynamicToolCall":
		default:
			return a.retire(s, "contamination", "forbidden native or unknown item: "+ev.Item.Type)
		}
	case "item/tool/call":
		allowed := false
		for _, tool := range s.tools {
			if tool == ev.Tool {
				allowed = true
			}
		}
		if !allowed || ev.Namespace != nil || ev.CallID == "" || len(m.ID) == 0 || s.pending != nil {
			return a.retire(s, "contamination", "unauthorized or overlapping tool request")
		}
		var args struct {
			Text string `json:"text"`
		}
		if p.StrictJSON(ev.Arguments, &args) != nil {
			return a.retire(s, "contamination", "malformed tool arguments")
		}
		if s.seenCalls[ev.CallID] || s.seenCalls[string(m.ID)] {
			return a.retire(s, "contamination", "replayed dependency")
		}
		s.seenCalls[ev.CallID] = true
		s.seenCalls[string(m.ID)] = true
		s.pending = &dependency{id: append(json.RawMessage(nil), m.ID...), call: ev.CallID, nonce: randomToken()}
		a.audit("tool_request", s, "")
		a.sequence++
		return a.emit(p.Move{Identity: a.sequence, Context: s.work.Context, Source: s.work.Invoke.Frame, Target: p.Echo, State: []byte(s.pending.nonce), Invoke: &p.Invoke{Args: p.Arguments("echo", args.Text)}})
	case "turn/completed":
		if ev.Turn.Status != "completed" || s.final == "" || s.pending != nil {
			return a.retire(s, "provider", "premature or unsuccessful turn completion")
		}
		s.history = append(s.history, historyEntry{"response", s.final})
		return a.settle(s, s.final)
	default:
		return a.retire(s, "contamination", fmt.Sprintf("unsupported provider event %s", m.Method))
	}
	return nil
}

// Command codexresident is a stateless app-server adapter. Every exchange and
// its observation arrives through the continuation; it owns no descriptor.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

type schema struct {
	Type     string             `json:"type"`
	Fields   map[string]*schema `json:"fields"`
	Required []string           `json:"required"`
	Items    *schema            `json:"items"`
}
type step struct {
	To      string                        `json:"to"`
	Payload json.RawMessage               `json:"payload"`
	State   string                        `json:"state"`
	Value   json.RawMessage               `json:"value"`
	Fault   *struct{ Kind, Cause string } `json:"fault"`
}
type continuation struct {
	Frame string `json:"frame"`
	Input struct {
		Text, Conversation, Thread, Incarnation string
		ActivityAt                              int64 `json:"activity_at"`
	} `json:"input"`
	Tools map[string]*schema `json:"tools"`
	Steps []step             `json:"steps"`
}
type observation struct {
	Lines                   []string `json:"lines"`
	Partial, Delivery, Read string
	PeerEOF                 bool   `json:"peer_eof"`
	Error                   string `json:"error"`
	ObservedAt              int64  `json:"observed_at"`
}
type message struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
	Params struct {
		ThreadID, TurnID, Tool string
		Arguments              json.RawMessage             `json:"arguments"`
		Item                   struct{ Type, Text string } `json:"item"`
		Turn                   struct{ ID, Status string } `json:"turn"`
	} `json:"params"`
}
type derived struct {
	Incarnation, Thread, Turn, Final, Violation, Partial             string
	Initialized, SentInitialize, SentThread, SentTurn, SentInterrupt bool
	NeedInspect, SentInspect, Inspected, NeedStop, Stopped, Lost     bool
	InspectAt, ActivityAt                                            int64
	Terminal                                                         string
	Pending                                                          []message
	Answered                                                         map[string]bool
	Results                                                          []step // model-tool steps, in the order Codex's calls were invoked
}

// modelTools are the capabilities Codex is shown, with the description it
// sees. Everything else in reach (codex.*) is this adapter's own operation.
// A tool is offered only when it is also in the continuation's reach.
var modelTools = map[string]string{
	"echo": "Skyra echo",
	"read": "Read a UTF-8 text file, or list a directory, in the workspace. Paths are relative to the workspace root; \".\" lists the root. " +
		"Returns at most 2000 lines or 64 KiB: when truncated is true, call again with start_line set to end_line + 1. " +
		"sha256 and size describe the whole file. A failure is reported in error (not_found, denied, not_text, too_large).",
}

func offered(c continuation, name string) bool {
	_, ok := modelTools[name]
	return ok && c.Tools[name] != nil
}

func main() {
	var c continuation
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		fail(err)
	}
	out, err := decide(c)
	if err != nil {
		fail(err)
	}
	fmt.Println(out)
}
func fail(err error) { fmt.Fprintln(os.Stderr, "codexresident:", err); os.Exit(1) }
func emit(to string, payload any) string {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	return string(b)
}
func rpc(id, method string, params any) string {
	b, _ := json.Marshal(map[string]any{"id": id, "method": method, "params": params})
	return string(b)
}
func operation(to, inc, op string, deliver []string, prefix string) string {
	if deliver == nil {
		deliver = []string{}
	}
	return emit(to, map[string]any{"incarnation": inc, "op": op, "deliver": deliver, "wait_ms": 30000, "prefix": prefix})
}
func ioCall(inc string, deliver []string, prefix string) string {
	return operation("codex.exchange", inc, "exchange", deliver, prefix)
}
func id(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}
func decide(c continuation) (string, error) {
	if c.Frame == "" {
		return "", errors.New("no frame")
	}
	if len(c.Steps) > 2000 {
		return "", errors.New("turn exceeded operation limit")
	}
	d := derived{Incarnation: c.Input.Incarnation, Thread: c.Input.Thread, ActivityAt: c.Input.ActivityAt,
		NeedInspect: c.Input.Incarnation != "" && c.Input.Thread != "", Answered: map[string]bool{}}
	if d.Incarnation == "" {
		d.Thread = ""
	}
	for _, s := range c.Steps {
		if _, tool := modelTools[s.To]; tool {
			// A tool's return or fault is an observation for Codex, in the same turn.
			d.Results = append(d.Results, s)
			continue
		}
		if s.Fault != nil {
			switch s.To {
			case "codex.inspect":
				d.SentInspect, d.NeedStop, d.Lost = true, true, true
				continue
			case "codex.stop":
				if !d.Lost {
					return "", fmt.Errorf("stop fault (%s): %s", s.Fault.Kind, s.Fault.Cause)
				}
				d.Stopped, d.NeedInspect = true, false
				d.Incarnation, d.Thread, d.Partial = "", "", ""
				continue
			default:
				return "", fmt.Errorf("%s fault (%s): %s", s.To, s.Fault.Kind, s.Fault.Cause)
			}
		}
		if s.State != "return" {
			return "", fmt.Errorf("%s ended in %s", s.To, s.State)
		}
		switch s.To {
		case "codex.start":
			var started struct {
				Incarnation string `json:"incarnation"`
			}
			if err := json.Unmarshal(s.Value, &started); err != nil || started.Incarnation == "" {
				return "", errors.New("start did not return an incarnation")
			}
			d.Incarnation, d.Thread, d.Partial = started.Incarnation, "", ""
			d.NeedInspect, d.SentInspect, d.Inspected = false, false, false
			d.Initialized, d.SentInitialize, d.SentThread = false, false, false
		case "codex.stop":
			var o observation
			if err := json.Unmarshal(s.Value, &o); err != nil {
				return "", err
			}
			if !o.PeerEOF && !d.Lost {
				return "", errors.New("stop did not observe resident EOF")
			}
			d.Stopped, d.NeedInspect = true, false
			d.Incarnation, d.Thread, d.Partial = "", "", ""
		case "codex.exchange", "codex.inspect":
			var q struct {
				Deliver []string `json:"deliver"`
			}
			var o observation
			if err := json.Unmarshal(s.Payload, &q); err != nil {
				return "", err
			}
			if err := json.Unmarshal(s.Value, &o); err != nil {
				return "", err
			}
			if s.To == "codex.inspect" {
				d.SentInspect = true
				d.InspectAt = o.ObservedAt
			}
			for _, line := range q.Deliver {
				var m message
				if json.Unmarshal([]byte(line), &m) != nil {
					return "", errors.New("recorded delivery is malformed")
				}
				switch m.Method {
				case "initialize":
					d.SentInitialize = true
				case "thread/loaded/list":
					d.SentInspect = true
				case "thread/start":
					d.SentThread = true
				case "turn/start":
					d.SentTurn = true
				case "turn/interrupt":
					d.SentInterrupt = true
				}
				if m.Method == "" && len(m.ID) > 0 && len(m.Result) > 0 {
					d.Answered[id(m.ID)] = true
				}
			}
			if o.PeerEOF && s.To == "codex.inspect" {
				d.NeedStop, d.Lost = true, true
				continue
			}
			if o.PeerEOF {
				return "", errors.New("resident process closed during the turn")
			}
			if o.Delivery == "unknown" {
				return "", errors.New("delivery may have reached Codex; refusing to retry")
			}
			if o.Error != "" && !strings.Contains(o.Error, "timed out") && !strings.Contains(o.Error, "deadline exceeded") {
				return "", fmt.Errorf("exchange: %s", o.Error)
			}
			d.Partial = o.Partial
			for _, line := range o.Lines {
				var m message
				if json.Unmarshal([]byte(line), &m) != nil {
					return "", errors.New("Codex emitted malformed JSON-RPC")
				}
				if len(m.Error) > 0 && string(m.Error) != "null" {
					if id(m.ID) == c.Frame+":inspect" {
						d.NeedStop = true
						continue
					}
					return "", fmt.Errorf("Codex RPC error: %s", m.Error)
				}
				switch id(m.ID) {
				case c.Frame + ":inspect":
					if len(m.Result) > 0 {
						d.Inspected = true
					}
				case c.Frame + ":initialize":
					if len(m.Result) > 0 {
						d.Initialized = true
					}
				case c.Frame + ":thread":
					if len(m.Result) > 0 {
						var r struct {
							Thread         struct{ ID string } `json:"thread"`
							ApprovalPolicy string              `json:"approvalPolicy"`
							Sandbox        struct {
								Type    string
								Network bool `json:"networkAccess"`
							} `json:"sandbox"`
						}
						if json.Unmarshal(m.Result, &r) != nil || r.Thread.ID == "" || r.ApprovalPolicy != "never" || r.Sandbox.Type != "readOnly" || r.Sandbox.Network {
							return "", errors.New("thread/start returned unsafe policy")
						}
						d.Thread = r.Thread.ID
					}
				case c.Frame + ":turn":
					if len(m.Result) > 0 {
						var r struct {
							Turn struct{ ID string } `json:"turn"`
						}
						if json.Unmarshal(m.Result, &r) != nil || r.Turn.ID == "" {
							return "", errors.New("turn/start returned no turn")
						}
						d.Turn = r.Turn.ID
					}
				}
				if m.Method == "" || m.Params.ThreadID != "" && d.Thread != "" && m.Params.ThreadID != d.Thread || m.Params.TurnID != "" && d.Turn != "" && m.Params.TurnID != d.Turn {
					continue
				}
				if d.Turn == "" && (m.Method == "item/tool/call" || strings.HasPrefix(m.Method, "item/") || strings.HasPrefix(m.Method, "turn/")) {
					continue
				}
				if d.Turn != "" && (strings.HasPrefix(m.Method, "item/") || strings.HasPrefix(m.Method, "turn/")) && o.ObservedAt > d.ActivityAt {
					d.ActivityAt = o.ObservedAt
				}
				if len(m.ID) > 0 && m.Method != "" && m.Method != "item/tool/call" {
					d.Violation = m.Method
					continue
				}
				switch m.Method {
				case "item/tool/call":
					if !offered(c, m.Params.Tool) {
						d.Violation = "dynamicTool." + m.Params.Tool
					} else {
						d.Pending = append(d.Pending, m)
					}
				case "item/started", "item/completed":
					switch m.Params.Item.Type {
					case "agentMessage":
						if m.Method == "item/completed" {
							d.Final = m.Params.Item.Text
						}
					case "userMessage", "reasoning", "plan", "dynamicToolCall":
					default:
						d.Violation = m.Params.Item.Type
					}
				case "turn/completed":
					if d.Turn != "" && m.Params.Turn.ID == d.Turn {
						d.Terminal = m.Params.Turn.Status
					}
				case "turn/started":
				case "":
				default: // notifications without authority have no effect
				}
			}
		default:
			return "", fmt.Errorf("unexpected step %s", s.To)
		}
	}
	if d.NeedInspect && d.Inspected && c.Input.ActivityAt > 0 && d.InspectAt-c.Input.ActivityAt >= int64(30*time.Minute/time.Millisecond) {
		d.NeedStop = true
	}
	if d.NeedStop && !d.Stopped {
		return operation("codex.stop", d.Incarnation, "stop", nil, d.Partial), nil
	}
	if d.Incarnation == "" {
		return emit("codex.start", map[string]any{}), nil
	}
	if !d.SentInitialize && d.Thread == "" {
		return ioCall(d.Incarnation, []string{rpc(c.Frame+":initialize", "initialize", map[string]any{"clientInfo": map[string]string{"name": "skyra", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}})}, d.Partial), nil
	}
	if !d.Initialized && d.Thread == "" {
		return ioCall(d.Incarnation, nil, d.Partial), nil
	}
	if !d.SentThread && d.Thread == "" {
		tools := []any{}
		for _, name := range slices.Sorted(maps.Keys(modelTools)) {
			if offered(c, name) {
				tools = append(tools, map[string]any{"name": name, "description": modelTools[name], "inputSchema": jsonSchema(c.Tools[name])})
			}
		}
		initialized, _ := json.Marshal(map[string]any{"method": "initialized", "params": map[string]any{}})
		start := rpc(c.Frame+":thread", "thread/start", map[string]any{"approvalPolicy": "never", "sandbox": "read-only", "environments": []any{}, "dynamicTools": tools, "developerInstructions": "Only registered dynamic tools may perform effects."})
		return ioCall(d.Incarnation, []string{string(initialized), start}, d.Partial), nil
	}
	if d.Thread == "" {
		return ioCall(d.Incarnation, nil, d.Partial), nil
	}
	if d.NeedInspect {
		if !d.SentInspect {
			return operation("codex.inspect", d.Incarnation, "inspect", []string{rpc(c.Frame+":inspect", "thread/loaded/list", map[string]any{})}, d.Partial), nil
		}
		if !d.Inspected {
			return ioCall(d.Incarnation, nil, d.Partial), nil
		}
	}
	if !d.SentTurn {
		return ioCall(d.Incarnation, []string{rpc(c.Frame+":turn", "turn/start", map[string]any{"threadId": d.Thread, "input": []any{map[string]string{"type": "text", "text": c.Input.Text}}})}, d.Partial), nil
	}
	if d.Turn == "" {
		return ioCall(d.Incarnation, nil, d.Partial), nil
	}
	if d.Violation != "" {
		if !d.SentInterrupt {
			return ioCall(d.Incarnation, []string{rpc(c.Frame+":interrupt", "turn/interrupt", map[string]string{"threadId": d.Thread, "turnId": d.Turn})}, d.Partial), nil
		}
		if d.Terminal == "" {
			return ioCall(d.Incarnation, nil, d.Partial), nil
		}
		// The attempt is reported as what it is: an observation returned to
		// the parent, which decides what it means. Blocking it is the
		// sandbox's job; this is the report (next/24, improvements 2a).
		return emit(c.Frame, map[string]any{"text": "", "refused": "native." + d.Violation, "thread": d.Thread, "incarnation": d.Incarnation, "activity_at": d.ActivityAt}), nil
	}
	for i, call := range d.Pending {
		if d.Answered[id(call.ID)] {
			continue
		}
		if i >= len(d.Results) {
			args := call.Params.Arguments
			if len(args) == 0 {
				args = json.RawMessage(`{}`)
			}
			return emit(call.Params.Tool, args), nil
		}
		if d.Results[i].To != call.Params.Tool {
			return "", fmt.Errorf("tool call %s asked for %s, but step %d invoked %s", id(call.ID), call.Params.Tool, i, d.Results[i].To)
		}
		result := toolResult(call.ID, d.Results[i])
		line, _ := json.Marshal(result)
		return ioCall(d.Incarnation, []string{string(line)}, d.Partial), nil
	}
	if d.Terminal != "" {
		if d.Terminal != "completed" || d.Final == "" {
			return "", fmt.Errorf("turn ended %s without an answer", d.Terminal)
		}
		return emit(c.Frame, map[string]any{"text": d.Final, "thread": d.Thread, "incarnation": d.Incarnation, "activity_at": d.ActivityAt}), nil
	}
	return ioCall(d.Incarnation, nil, d.Partial), nil
}

func toolResult(id json.RawMessage, s step) map[string]any {
	text, success := string(s.Value), s.State == "return"
	var v string
	if json.Unmarshal(s.Value, &v) == nil {
		text = v
	}
	if s.Fault != nil {
		text = fmt.Sprintf("fault (%s): %s", s.Fault.Kind, s.Fault.Cause)
	}
	return map[string]any{"id": id, "result": map[string]any{"success": success, "contentItems": []any{map[string]string{"type": "inputText", "text": text}}}}
}
func jsonSchema(s *schema) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	switch s.Type {
	case "object":
		p := map[string]any{}
		for k, v := range s.Fields {
			p[k] = jsonSchema(v)
		}
		r := s.Required
		if r == nil {
			r = []string{}
		}
		return map[string]any{"type": "object", "properties": p, "required": r, "additionalProperties": false}
	case "array":
		return map[string]any{"type": "array", "items": jsonSchema(s.Items)}
	case "bool":
		return map[string]any{"type": "boolean"}
	default:
		return map[string]any{"type": s.Type}
	}
}

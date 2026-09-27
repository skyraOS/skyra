// Command codex is the Codex adapter step (next/15): a stateless
// continuation that owns every Codex semantic, so the runtime owns none.
//
// It runs once per step, under the sandbox the runtime renders, and reaches
// exactly one thing: the actor host in front of the Codex app-server
// (SKYRA_CONNECT). It keeps no state between runs. Each run rebuilds where the
// turn stands by reading the host's inbox from the start of this frame's turn,
// and every delivery has an id derived from the frame, so delivering again is
// a no-op: a resumed step re-derives what the previous step did.
//
// What it maps:
//
//	Codex asks for a dynamic tool       → addressed output {"to": tool, "payload": args}
//	                                      (the runtime derives an invoke, or a fault)
//	a step came back                    → the tool call's JSON-RPC result
//	the turn completes                  → the final agent message, as plain text (a reply)
//	Codex uses anything native          → addressed output naming it, e.g. "native.commandExecution".
//	(a shell command, a file change,       It is not in reach, so the runtime derives a horizon
//	an approval request)                   fault. Codex may already have run it: this detects, it
//	                                       does not prevent (next/15, open)
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type schema struct {
	Type     string             `json:"type"`
	Fields   map[string]*schema `json:"fields"`
	Required []string           `json:"required"`
	Items    *schema            `json:"items"`
}

type continuation struct {
	Frame string `json:"frame"`
	Input struct {
		Text         string `json:"text"`
		Conversation string `json:"conversation"` // optional: turns of one conversation share a Codex thread
	} `json:"input"`
	Tools map[string]*schema `json:"tools"`
	Steps []struct {
		To    string          `json:"to"`
		State string          `json:"state"`
		Value json.RawMessage `json:"value"`
		Fault *struct {
			Kind  string `json:"kind"`
			Cause string `json:"cause"`
		} `json:"fault"`
	} `json:"steps"`
}

func main() {
	out, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "codex adapter:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

func run() (string, error) {
	var c continuation
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		return "", err
	}
	h, err := dial(os.Getenv("SKYRA_CONNECT"))
	if err != nil {
		return "", err
	}
	defer h.conn.Close()

	// The connection: initialize once per host lifetime.
	seq, err := h.deliver("initialize", request("initialize", "initialize", map[string]any{
		"clientInfo": map[string]string{"name": "skyra", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}}))
	if err != nil {
		return "", err
	}
	if _, err := h.response(seq, "initialize"); err != nil {
		return "", fmt.Errorf("initialize: %w", err)
	}
	if _, err := h.deliver("initialized", map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		return "", err
	}

	// The thread: one per conversation, or per frame when there is none, so
	// Codex keeps its memory across a conversation's turns. Only the dynamic
	// tools in reach, and a read-only sandbox with no network and no approvals.
	threadID := c.Frame + ":thread"
	if c.Input.Conversation != "" {
		threadID = "conversation:" + c.Input.Conversation + ":thread"
	}
	seq, err = h.deliver(threadID, request(threadID, "thread/start", map[string]any{
		"approvalPolicy": "never", "sandbox": "read-only", "dynamicTools": dynamicTools(c.Tools),
		"developerInstructions": "Only the registered dynamic tools may perform effects."}))
	if err != nil {
		return "", err
	}
	res, err := h.response(seq, threadID)
	if err != nil {
		return "", fmt.Errorf("thread/start: %w", err)
	}
	var started struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
		ApprovalPolicy string `json:"approvalPolicy"`
		Sandbox        struct {
			Type    string `json:"type"`
			Network bool   `json:"networkAccess"`
		} `json:"sandbox"`
	}
	if json.Unmarshal(res, &started) != nil || started.Thread.ID == "" {
		return "", errors.New("thread/start: no thread")
	}
	if started.ApprovalPolicy != "never" || started.Sandbox.Type != "readOnly" || started.Sandbox.Network {
		return "", fmt.Errorf("thread/start: effective policy is %s/%s/network=%v, not never/readOnly/no network",
			started.ApprovalPolicy, started.Sandbox.Type, started.Sandbox.Network)
	}
	thread := started.Thread.ID

	// The turn: one per frame.
	turnID := c.Frame + ":turn"
	from, err := h.deliver(turnID, request(turnID, "turn/start", map[string]any{
		"threadId": thread, "input": []any{map[string]string{"type": "text", "text": c.Input.Text}}}))
	if err != nil {
		return "", err
	}

	// Walk the turn from its start. Tool calls already answered by a step are
	// answered again (a no-op after the first time); the first unanswered one,
	// or the turn's end, decides this step's output.
	calls, final := 0, ""
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		events, next, err := h.read(from, 30_000)
		if err != nil {
			return "", err
		}
		for _, e := range events {
			if e.Host != "" {
				return "", fmt.Errorf("host: %s", e.Host)
			}
			var m struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
				Params struct {
					ThreadID string          `json:"threadId"`
					Tool     string          `json:"tool"`
					Args     json.RawMessage `json:"arguments"`
					Item     struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"item"`
					Turn struct {
						Status string `json:"status"`
					} `json:"turn"`
				} `json:"params"`
			}
			if json.Unmarshal(e.Line, &m) != nil || m.Method == "" || m.Params.ThreadID != thread {
				continue // a response, another thread's event, or not JSON-RPC
			}
			request := len(m.ID) > 0
			switch {
			case request && m.Method == "item/tool/call":
				if calls < len(c.Steps) {
					if _, err := h.deliver(fmt.Sprintf("%s:tool:%d", c.Frame, calls), toolResult(m.ID, c.Steps[calls].State, c.Steps[calls].Value, c.Steps[calls].Fault)); err != nil {
						return "", err
					}
					calls++
					continue
				}
				args := m.Params.Args
				if len(args) == 0 {
					args = json.RawMessage(`{}`)
				}
				out, _ := json.Marshal(map[string]any{"to": m.Params.Tool, "payload": args})
				return string(out), nil
			case request:
				return native(m.Method), nil // an approval or any other request: never granted
			case m.Method == "item/started" || m.Method == "item/completed":
				switch m.Params.Item.Type {
				case "agentMessage":
					if m.Method == "item/completed" {
						final = m.Params.Item.Text
					}
				case "userMessage", "reasoning", "plan", "dynamicToolCall":
				default:
					return native(m.Params.Item.Type), nil
				}
			case m.Method == "turn/completed":
				if m.Params.Turn.Status != "completed" || final == "" || calls != len(c.Steps) {
					return "", fmt.Errorf("turn ended %q with %d of %d tool calls answered", m.Params.Turn.Status, calls, len(c.Steps))
				}
				// The runtime reads one line: address the reply, so newlines
				// survive JSON-escaped and text that looks like JSON stays a reply.
				out, _ := json.Marshal(map[string]any{"to": "reply", "payload": final})
				return string(out), nil
			}
		}
		from = next
	}
	return "", errors.New("the turn did not finish in time")
}

// native names a capability Codex tried to use on its own. It is not in any
// Horizon, so the runtime derives a horizon fault.
func native(what string) string {
	out, _ := json.Marshal(map[string]any{"to": "native." + what, "payload": map[string]any{}})
	return string(out)
}

func request(id, method string, params any) map[string]any {
	return map[string]any{"id": id, "method": method, "params": params}
}

func toolResult(id json.RawMessage, state string, value json.RawMessage, fault *struct {
	Kind  string `json:"kind"`
	Cause string `json:"cause"`
}) map[string]any {
	text, success := string(value), state == "return"
	var s string
	if json.Unmarshal(value, &s) == nil {
		text = s
	}
	if fault != nil {
		text = fmt.Sprintf("fault (%s): %s", fault.Kind, fault.Cause)
	}
	return map[string]any{"id": id, "result": map[string]any{"success": success,
		"contentItems": []any{map[string]string{"type": "inputText", "text": text}}}}
}

// dynamicTools turns the input contracts in reach into Codex tool definitions.
func dynamicTools(tools map[string]*schema) []any {
	var defs []any
	for name, s := range tools {
		defs = append(defs, map[string]any{"name": name, "description": "Skyra capability " + name, "inputSchema": jsonSchema(s)})
	}
	return defs
}

func jsonSchema(s *schema) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	switch s.Type {
	case "object":
		props := map[string]any{}
		for k, f := range s.Fields {
			props[k] = jsonSchema(f)
		}
		required := s.Required
		if required == nil {
			required = []string{}
		}
		return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
	case "array":
		return map[string]any{"type": "array", "items": jsonSchema(s.Items)}
	case "bool":
		return map[string]any{"type": "boolean"}
	}
	return map[string]any{"type": s.Type}
}

// host is a client of the universal actor host contracts.
type host struct {
	conn net.Conn
	r    *bufio.Reader
}

type event struct {
	Seq  int             `json:"seq"`
	Line json.RawMessage `json:"line"`
	Host string          `json:"host"`
}

func dial(sock string) (*host, error) {
	if sock == "" {
		return nil, errors.New("no actor host granted (SKYRA_CONNECT is empty)")
	}
	c, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}
	return &host{conn: c, r: bufio.NewReaderSize(c, 1<<20)}, nil
}

func (h *host) call(req map[string]any, resp any) error {
	b, _ := json.Marshal(req)
	if _, err := h.conn.Write(append(b, '\n')); err != nil {
		return err
	}
	line, err := h.r.ReadBytes('\n')
	if err != nil {
		return err
	}
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(line, &e) == nil && e.Error != "" {
		return errors.New(e.Error)
	}
	return json.Unmarshal(line, resp)
}

// deliver sends a line once per id, and returns the inbox position to read
// its effects from.
func (h *host) deliver(id string, line any) (int, error) {
	raw, _ := json.Marshal(line)
	var r struct {
		Seq int `json:"seq"`
	}
	err := h.call(map[string]any{"op": "deliver", "id": id, "line": json.RawMessage(raw)}, &r)
	return r.Seq, err
}

func (h *host) read(from, waitMS int) ([]event, int, error) {
	var r struct {
		Events []event `json:"events"`
		Next   int     `json:"next"`
	}
	err := h.call(map[string]any{"op": "read", "from": from, "wait_ms": waitMS}, &r)
	return r.Events, r.Next, err
}

// response waits for the JSON-RPC response to id, reading from seq.
func (h *host) response(from int, id string) (json.RawMessage, error) {
	want, _ := json.Marshal(id)
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		events, next, err := h.read(from, 30_000)
		if err != nil {
			return nil, err
		}
		for _, e := range events {
			if e.Host != "" {
				return nil, fmt.Errorf("host: %s", e.Host)
			}
			var m struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
				Result json.RawMessage `json:"result"`
				Error  json.RawMessage `json:"error"`
			}
			if json.Unmarshal(e.Line, &m) != nil || m.Method != "" || strings.TrimSpace(string(m.ID)) != string(want) {
				continue
			}
			if len(m.Error) > 0 {
				return nil, fmt.Errorf("provider error: %s", m.Error)
			}
			return m.Result, nil
		}
		from = next
	}
	return nil, fmt.Errorf("no response to %s", id)
}

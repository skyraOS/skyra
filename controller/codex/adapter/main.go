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
//	the turn completes                  → the final agent message, returned to this frame's slot
//	Codex uses anything native          → after interrupting the affected turn, a return naming
//	(a shell command, a file change,       the attempt, e.g. "native.commandExecution": an
//	an approval request)                   observation for the parent, not a fault it triggers.
//	                                       The provider launcher enforces OS file/exec restrictions.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"os"
	"slices"
	"strings"
	"time"
)

type schema struct {
	Title    string             `json:"title,omitempty"`
	Type     string             `json:"type"`
	Fields   map[string]*schema `json:"fields,omitempty"`
	Required []string           `json:"required,omitempty"`
	Items    *schema            `json:"items,omitempty"`
}

type continuation struct {
	Frame string `json:"frame"`
	Input struct {
		Text         string `json:"text"`
		Conversation string `json:"conversation"` // optional: turns of one conversation share a Codex thread
	} `json:"input"`
	Tools     map[string]*schema   `json:"tools"`
	Contracts map[string][]*schema `json:"contracts"`
	Steps     []struct {
		To          string          `json:"to"`
		Payload     json.RawMessage `json:"payload"`
		State       string          `json:"state"`
		Value       json.RawMessage `json:"value"`
		Observation string          `json:"observation"`
		Fault       *struct {
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
		"approvalPolicy": "never", "sandbox": "read-only", "environments": []any{}, "dynamicTools": dynamicTools(c),
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

	turnResponse, err := h.response(from, turnID)
	if err != nil {
		return "", fmt.Errorf("turn/start: %w", err)
	}
	var turnStarted struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if json.Unmarshal(turnResponse, &turnStarted) != nil || turnStarted.Turn.ID == "" {
		return "", errors.New("turn/start: no turn")
	}
	turn := turnStarted.Turn.ID
	reject := func(what string) (string, error) {
		if err := h.interrupt(c.Frame, thread, turn, from); err != nil {
			return "", fmt.Errorf("native %s: %w", what, err)
		}
		return native(c.Frame, what), nil
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
					TurnID   string          `json:"turnId"`
					Tool     string          `json:"tool"`
					Args     json.RawMessage `json:"arguments"`
					Item     struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"item"`
					Turn struct {
						ID     string `json:"id"`
						Status string `json:"status"`
					} `json:"turn"`
				} `json:"params"`
			}
			if json.Unmarshal(e.Line, &m) != nil || m.Method == "" || m.Params.ThreadID != thread {
				continue // a response, another thread's event, or not JSON-RPC
			}
			if m.Params.TurnID != "" && m.Params.TurnID != turn {
				continue
			}
			if m.Method == "turn/completed" && m.Params.Turn.ID != turn {
				continue
			}
			request := len(m.ID) > 0
			switch {
			case request && m.Method == "item/tool/call":
				if !offered(c, m.Params.Tool) {
					return reject("dynamicTool." + m.Params.Tool)
				}
				if calls < len(c.Steps) {
					step := c.Steps[calls]
					result := toolResult(m.ID, step.State, step.Value, step.Fault)
					if _, err := h.deliver(fmt.Sprintf("%s:tool:%d", c.Frame, calls), result); err != nil {
						return "", err
					}
					calls++
					continue
				}
				args := m.Params.Args
				if len(args) == 0 {
					args = json.RawMessage(`{}`)
				}
				out, err := toolEmission(c, m.Params.Tool, args)
				if err != nil {
					return "", err
				}
				return out, nil
			case request:
				return reject(m.Method) // an approval or any other request: never granted
			case m.Method == "item/started" || m.Method == "item/completed":
				switch m.Params.Item.Type {
				case "agentMessage":
					if m.Method == "item/completed" {
						final = m.Params.Item.Text
					}
				case "userMessage", "reasoning", "plan", "dynamicToolCall":
				default:
					return reject(m.Params.Item.Type)
				}
			case m.Method == "turn/completed":
				if m.Params.Turn.Status != "completed" || final == "" || calls != len(c.Steps) {
					return "", fmt.Errorf("turn ended %q with %d of %d tool calls answered", m.Params.Turn.Status, calls, len(c.Steps))
				}
				// The runtime reads one line: address the slot, so newlines
				// survive JSON-escaped and text that looks like JSON stays text.
				out, _ := json.Marshal(map[string]any{"to": c.Frame, "payload": final})
				return string(out), nil
			}
		}
		from = next
	}
	return "", errors.New("the turn did not finish in time")
}

// native names a capability Codex tried to use on its own. It is not in any
// Horizon, so the runtime derives a horizon fault.
// interrupt is adapter policy, not a runtime event. A terminal observation
// for this exact turn is required; an RPC acknowledgement alone is insufficient.
func (h *host) interrupt(frame, thread, turn string, from int) error {
	deadline := time.Now().Add(15 * time.Second)
	if err := h.conn.SetDeadline(deadline); err != nil {
		return err
	}
	defer h.conn.SetDeadline(time.Time{})
	id := frame + ":interrupt"
	if _, err := h.deliver(id, request(id, "turn/interrupt", map[string]string{"threadId": thread, "turnId": turn})); err != nil {
		return fmt.Errorf("interrupt delivery: %w", err)
	}
	for time.Now().Before(deadline) {
		events, next, err := h.read(from, 1000)
		if err != nil {
			return fmt.Errorf("interrupt confirmation: %w", err)
		}
		for _, e := range events {
			if e.Host != "" {
				return fmt.Errorf("interrupt confirmation: %s", e.Host)
			}
			var m struct {
				Method string `json:"method"`
				Params struct {
					Thread string `json:"threadId"`
					Turn   struct {
						ID     string `json:"id"`
						Status string `json:"status"`
					} `json:"turn"`
				} `json:"params"`
			}
			if json.Unmarshal(e.Line, &m) != nil {
				continue
			}
			if m.Method == "turn/completed" && m.Params.Thread == thread && m.Params.Turn.ID == turn {
				switch m.Params.Turn.Status {
				case "interrupted", "completed", "failed":
					return nil // completion may race interruption
				}
			}
		}
		from = next
	}
	return errors.New("interrupt was not confirmed; turn may still be active")
}

// native reports a native attempt as what it is: an observation returned to
// the parent, not a name chosen so the runtime would fault it.
func native(frame, what string) string {
	out, _ := json.Marshal(map[string]any{"to": frame, "payload": "refused: the agent attempted native." + what + "; the turn was interrupted"})
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

// modelTools are the capabilities Codex is shown, by the name it is shown
// (a dynamic tool name has no dots). Under a shared Horizon (next/24) reach
// holds every surface in the room, this adapter and the Horizon's syscalls
// included, so only these are offered, and only when they are in reach.
var modelTools = map[string]string{"echo": "echo", "agent": "agent"}

// tool is one tool Codex is offered: its name, what it says, its schema, and
// the name in reach a call to it invokes.
type tool struct {
	name, description, target string
	schema                    *schema
	kind                      string // "": a call with the model's arguments; remember; procedure
}

// tools is everything Codex is offered, in a fixed order (next/32): the model
// tools in reach; when the interpreter is in reach, remember (by the index of
// one of this turn's own calls: the adapter, a worker, holds its observation)
// and procedure (its input goes to the interpreter, which runs the procedure
// whose contract it fits); and one tool per procedure that exists when the
// thread starts, labelled by its contract's title, never a hash. Codex's
// tools are fixed at thread/start, so a procedure remembered in this turn is
// called through procedure.
func tools(c continuation) []tool {
	var ts []tool
	for _, name := range slices.Sorted(maps.Keys(modelTools)) {
		if c.Tools[modelTools[name]] != nil {
			ts = append(ts, tool{name: name, description: "Skyra capability " + modelTools[name], target: modelTools[name], schema: c.Tools[modelTools[name]]})
		}
	}
	interp := interpreter(c)
	if interp == "" {
		return ts
	}
	ts = append(ts,
		tool{name: "remember", target: interp, kind: "remember", description: "Remember how one of your calls in this turn was done, as a procedure: " +
			"call is its index among your calls, from 0. Calling the procedure later runs the same calls on new input, without you.",
			schema: &schema{Type: "object", Required: []string{"call"}, Fields: map[string]*schema{"call": {Type: "number"}}}},
		tool{name: "procedure", target: interp, kind: "procedure", description: "Run a remembered procedure on new input: input is what you would give the call it was remembered from.",
			schema: &schema{Type: "object", Required: []string{"input"}, Fields: map[string]*schema{"input": {Type: "any"}}}})
	for _, c := range contractsOf(c, interp)[1:] {
		if name := toolName(c.Title); name != "" && !slices.ContainsFunc(ts, func(t tool) bool { return t.name == name }) {
			ts = append(ts, tool{name: name, target: interp, schema: c, description: "A remembered procedure: " + c.Title})
		}
	}
	return ts
}

// interpreter is the name in reach that holds remember, or "".
func interpreter(c continuation) string {
	for _, name := range slices.Sorted(maps.Keys(c.Tools)) {
		if s := c.Tools[name]; s != nil && s.Title == "remember" {
			return name
		}
	}
	return ""
}

func contractsOf(c continuation, name string) []*schema {
	if all, ok := c.Contracts[name]; ok {
		return all
	}
	if s := c.Tools[name]; s != nil {
		return []*schema{s}
	}
	return nil
}

// toolName is a title as a dynamic tool name, or "" if it cannot be one.
func toolName(title string) string {
	if title == "" || len(title) > 64 {
		return ""
	}
	for _, r := range title {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return ""
		}
	}
	return title
}

func offered(c continuation, name string) bool {
	return slices.ContainsFunc(tools(c), func(t tool) bool { return t.name == name })
}

// toolEmission is the instruction a tool call becomes.
func toolEmission(c continuation, name string, args json.RawMessage) (string, error) {
	i := slices.IndexFunc(tools(c), func(t tool) bool { return t.name == name })
	t := tools(c)[i]
	var payload any = args
	switch t.kind {
	case "remember":
		var a struct {
			Call *int `json:"call"`
		}
		if json.Unmarshal(args, &a) != nil || a.Call == nil || *a.Call < 0 || *a.Call >= len(c.Steps) {
			payload = map[string]any{"remember": map[string]any{}} // no such call: intake refuses it
			break
		}
		s := c.Steps[*a.Call]
		payload = map[string]any{"remember": map[string]any{"observation": s.Observation,
			"contract": map[string]any{"target": s.To, "schema": selected(contractsOf(c, s.To), s.Payload)}}}
	case "procedure":
		var a struct {
			Input json.RawMessage `json:"input"`
		}
		_ = json.Unmarshal(args, &a)
		payload = a.Input
	}
	out, err := json.Marshal(map[string]any{"to": t.target, "payload": payload})
	return string(out), err
}

// selected is the one contract a payload fits, as the kernel selected it.
func selected(contracts []*schema, payload json.RawMessage) *schema {
	var v any
	d := json.NewDecoder(strings.NewReader(string(payload)))
	d.UseNumber()
	if d.Decode(&v) != nil {
		return nil
	}
	var fit []*schema
	for _, c := range contracts {
		if fits(c, v) {
			fit = append(fit, c)
		}
	}
	if len(fit) != 1 {
		return nil
	}
	return fit[0]
}

func fits(s *schema, v any) bool {
	switch s.Type {
	case "object":
		o, ok := v.(map[string]any)
		if !ok {
			return false
		}
		for k, e := range o {
			if f := s.Fields[k]; f == nil || !fits(f, e) {
				return false
			}
		}
		for _, r := range s.Required {
			if _, ok := o[r]; !ok {
				return false
			}
		}
		return true
	case "array":
		a, ok := v.([]any)
		if !ok {
			return false
		}
		for _, e := range a {
			if !fits(s.Items, e) {
				return false
			}
		}
		return true
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(json.Number)
		return ok
	case "bool":
		_, ok := v.(bool)
		return ok
	case "null":
		return v == nil
	case "any":
		return true
	}
	return false
}

// dynamicTools turns the tools Codex is offered into Codex tool definitions,
// in a fixed order. A resumed round delivers thread/start again under the
// same id, and the host refuses a line that differs from the first delivery,
// so the same continuation must produce the same bytes every time.
func dynamicTools(c continuation) []any {
	defs := []any{}
	for _, t := range tools(c) {
		defs = append(defs, map[string]any{"name": t.name, "description": t.description, "inputSchema": jsonSchema(t.schema)})
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
	case "any":
		return map[string]any{}
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

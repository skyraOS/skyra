// Command codexresident is a stateless app-server adapter. Every exchange and
// its observation arrives through the continuation; it owns no descriptor.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"skyra/controller/codex/proceduretools"
	"slices"
	"strings"

	"skyra/controller/codex/prompt"
	desktop "skyra/controller/cua/observation"
)

type schema = proceduretools.Schema
type step struct {
	To          string                        `json:"to"`
	Payload     json.RawMessage               `json:"payload"`
	State       string                        `json:"state"`
	Value       json.RawMessage               `json:"value"`
	Observation string                        `json:"observation"`
	Fault       *struct{ Kind, Cause string } `json:"fault"`
}
type continuation struct {
	Actor     string               `json:"actor"`
	Held      map[string][]*schema `json:"held"`
	Workspace string               `json:"workspace"`
	Scratch   string               `json:"scratch"`
	Frame     string               `json:"frame"`
	Input     struct {
		Text, Header, Conversation, Thread, Incarnation, Model string
		Fresh                                                  bool
		Context                                                json.RawMessage
		ActivityAt                                             int64 `json:"activity_at"`
	} `json:"input"`
	Tools     map[string]*schema   `json:"tools"`
	Contracts map[string][]*schema `json:"contracts"`
	Steps     []step               `json:"steps"`
}
type observation struct {
	Lines                   []string `json:"lines"`
	Partial, Delivery, Read string
	PeerEOF                 bool   `json:"peer_eof"`
	Error                   string `json:"error"`
	ObservedAt              int64  `json:"observed_at"`
	Incarnation             string `json:"incarnation"` // open: the session's handle
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
		Turn                   struct {
			ID, Status string
			Error      struct{ Message string } `json:"error"`
		} `json:"turn"`
	} `json:"params"`
}
type derived struct {
	SentResume                                                       bool
	Model                                                            string
	Incarnation, Thread, Turn, Final, Violation, Partial             string
	Initialized, SentInitialize, SentThread, SentTurn, SentInterrupt bool
	NeedInspect, SentInspect, Inspected, NeedStop, Stopped, Lost     bool
	InspectAt, ActivityAt                                            int64
	Terminal, TerminalError                                          string // how the turn ended, and Codex's reason when it failed
	Pending                                                          []message
	Answered                                                         map[string]bool
	Results                                                          []step // model-tool steps, in the order Codex's calls were invoked
}

// modelTools are the capabilities Codex is shown, with the description it
// sees. Everything else in reach (codex.*) is this adapter's own operation.
// A tool is offered only when it is also in the continuation's reach.
var modelTools = proceduretools.ModelTools

// tool is one tool Codex is offered: its name, what it says, its schema, and
// the name in reach a call to it invokes.
type tool struct {
	name, description, target string
	schema                    *schema
	kind                      string    // "": a call with the model's arguments; remember; procedure
	contracts                 []*schema // the contracts of target a call to it may select
}

func (t tool) shared() proceduretools.Tool {
	return proceduretools.Tool{Name: t.name, Description: t.description, Target: t.target, Kind: t.kind, Schema: t.schema, Contracts: t.contracts}
}

// tools is everything Codex is offered, in a fixed order (next/32): the model
// tools in reach; when the interpreter is in reach, remember (by the index of
// one of this turn's own calls: the adapter, a worker, holds its observation)
// and procedure (its input goes to the interpreter, which runs the procedure
// whose contract it fits); and one tool per procedure that exists when the
// thread starts, labelled by its contract's title, never a hash. Codex's
// tools are fixed at thread/start, so a procedure remembered in this thread
// is called through procedure.
func tools(c continuation) []tool {
	var ts []tool
	for _, name := range slices.Sorted(maps.Keys(modelTools)) {
		if c.Tools[name] != nil {
			description := modelTools[name]
			if name == "desktop_observe" && c.Tools[name].Title != "" {
				description += " " + c.Tools[name].Title
			}
			toolName := name
			if strings.Contains(name, ".") {
				toolName = strings.ReplaceAll(name, ".", "_")
			}
			ts = append(ts, tool{name: toolName, description: description, target: name, schema: c.Tools[name], contracts: proceduretools.ContractsOf(c.Tools, c.Contracts, name)})
		}
	}
	reserved := make([]string, 0, len(ts))
	for _, t := range ts {
		reserved = append(reserved, t.name)
	}
	for _, t := range proceduretools.Procedures(c.Tools, c.Contracts, reserved, "Remember how one of your tool calls in this turn was done, as a procedure: call is its index among your tool calls this turn, from 0. Calling the procedure later runs the same calls on new input, without you, and with no approval. Only a call that itself called tools can be remembered.") {
		ts = append(ts, tool{name: t.Name, target: t.Target, kind: t.Kind, description: t.Description, schema: t.Schema, contracts: t.Contracts})
	}
	ts = slices.DeleteFunc(ts, func(t tool) bool { return !proceduretools.Offered(c.Held, t.shared()) })
	for i := range ts {
		if actors := proceduretools.Delegable(c.Tools, c.Contracts); ts[i].name == "delegate" && len(actors) > 0 {
			ts[i].description += " Available actors: " + strings.Join(actors, ", ") + "."
		}
	}
	return ts
}

func shared(ts []tool) []proceduretools.Tool {
	out := make([]proceduretools.Tool, len(ts))
	for i, t := range ts {
		out[i] = t.shared()
	}
	return out
}

func lookup(c continuation, name string) (tool, bool) {
	ts := tools(c)
	i := slices.IndexFunc(ts, func(t tool) bool { return t.name == name })
	if i < 0 {
		return tool{}, false
	}
	return ts[i], true
}

// turnInput is what a turn gives Codex: the conversation so far, when the
// invoker passed it as context (proposal 38: a thread lasts one invocation,
// so memory is the lineage's), then the message itself, always last. Images
// pasted with the message are not passed on: Codex has no read of them until
// a recorded Skyra tool brings them back (proposal 49, decision 2).
func turnInput(context json.RawMessage, text string) []any {
	var input []any
	if earlier := recall(context); earlier != "" {
		input = append(input, map[string]string{"type": "text", "text": earlier})
	}
	return append(input, map[string]string{"type": "text", "text": text})
}

// recall renders a context observation for the model: a transcript of
// {role, text} entries as lines, anything else as its JSON. Nothing at all
// (absent, null, empty) is no context.
func recall(context json.RawMessage) string {
	var transcript []struct{ Role, Text string }
	if json.Unmarshal(context, &transcript) == nil {
		if len(transcript) == 0 {
			return ""
		}
		var b strings.Builder
		b.WriteString("Earlier in this conversation:\n")
		for _, e := range transcript {
			fmt.Fprintf(&b, "\n%s: %s\n", e.Role, e.Text)
		}
		return b.String()
	}
	if t := strings.TrimSpace(string(context)); t == "" || t == "null" || t == "{}" {
		return ""
	}
	return "Earlier in this conversation (as recorded):\n" + string(context)
}

// quiet reports whether an operation only waited out its deadline with
// nothing to read. service-io reports Go's net timeout ("i/o timeout").
func quiet(o observation) bool {
	return strings.Contains(o.Error, "i/o timeout") || strings.Contains(o.Error, "timed out") || strings.Contains(o.Error, "deadline exceeded")
}

func offered(c continuation, name string) bool {
	_, ok := lookup(c, name)
	return ok
}

// toolEmission is the instruction a tool call becomes. results are this
// turn's tool steps, in the order Codex's calls were invoked.
func toolEmission(c continuation, d derived, name string, args json.RawMessage) string {
	results := d.Results
	t, _ := lookup(c, name)
	calls := make([]proceduretools.Call, len(results))
	for i, s := range results {
		calls[i] = proceduretools.Call{Target: s.To, Observation: s.Observation, Payload: s.Payload}
	}
	return emit(t.target, proceduretools.Payload(t.kind, args, calls, c.Tools, c.Contracts, proceduretools.Caller{Actor: c.Actor, Model: d.Model, Tools: shared(tools(c))}))
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
		if !strings.HasPrefix(s.To, "codex.") {
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
		case "codex.open":
			var opened observation
			if err := json.Unmarshal(s.Value, &opened); err != nil || opened.Incarnation == "" {
				return "", fmt.Errorf("open did not return a session: %s", opened.Error)
			}
			d.Incarnation, d.Thread, d.Partial = opened.Incarnation, "", ""
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
				case "thread/resume":
					d.SentResume = true
				case "turn/start":
					d.SentTurn = true
				case "turn/interrupt":
					d.SentInterrupt = true
				}
				if m.Method == "" && len(m.ID) > 0 && len(m.Result) > 0 {
					d.Answered[id(m.ID)] = true
				}
			}
			// A health check that ends or goes unanswered is a lost server:
			// stop it and start another, rather than wait on it forever.
			if s.To == "codex.inspect" && (o.PeerEOF || quiet(o) && len(o.Lines) == 0) {
				d.NeedStop, d.Lost = true, true
				continue
			}
			if o.PeerEOF {
				return "", errors.New("resident process closed during the turn")
			}
			if o.Delivery == "unknown" {
				return "", errors.New("delivery may have reached Codex; refusing to retry")
			}
			// Silence during a turn is normal: the model may reason without
			// streaming. Before a turn, the server answers at once, so a
			// quiet wait there is still a failure.
			if o.Error != "" && !(quiet(o) && d.Turn != "") {
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
				case c.Frame + ":thread", c.Frame + ":resume":
					if len(m.Result) > 0 {
						var r struct {
							Thread         struct{ ID string } `json:"thread"`
							Model          string              `json:"model"`
							ApprovalPolicy string              `json:"approvalPolicy"`
							Sandbox        struct {
								Type    string
								Network bool `json:"networkAccess"`
							} `json:"sandbox"`
						}
						if json.Unmarshal(m.Result, &r) != nil || r.Thread.ID == "" || r.ApprovalPolicy != "never" || r.Sandbox.Type != "readOnly" || r.Sandbox.Network {
							return "", errors.New("thread/start returned unsafe policy")
						}
						if id(m.ID) == c.Frame+":resume" && r.Thread.ID != c.Input.Thread {
							return "", errors.New("resumed another thread")
						}
						d.Thread = r.Thread.ID
						d.Model = r.Model
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
					// Web search, a native read-only tool (proposal 38), runs
					// at the provider. Its reads are not recorded in the lineage.
					case "webSearch":
						if c.Input.Fresh {
							d.Violation = m.Params.Item.Type
						}
					case "userMessage", "reasoning", "plan", "dynamicToolCall":
					default:
						d.Violation = m.Params.Item.Type
					}
				case "turn/completed":
					if d.Turn != "" && m.Params.Turn.ID == d.Turn {
						d.Terminal, d.TerminalError = m.Params.Turn.Status, m.Params.Turn.Error.Message
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
	// Fresh servers have no next invocation to reuse them. Finish only
	// after their recorded stop, rather than retaining an idle child server.
	if c.Input.Fresh && d.Stopped && d.Terminal != "" {
		return finish(c, d)
	}
	// An idle server is the service host's to reclaim, by its backoff; a turn
	// finds out by its inspect, and opens another.
	if d.NeedStop && !d.Stopped {
		return operation("codex.stop", d.Incarnation, "stop", nil, d.Partial), nil
	}
	if d.Incarnation == "" {
		return emit("codex.open", map[string]any{"program": "codex"}), nil
	}
	if !d.SentInitialize && d.Thread == "" {
		return ioCall(d.Incarnation, []string{rpc(c.Frame+":initialize", "initialize", map[string]any{"clientInfo": map[string]string{"name": "skyra", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}})}, d.Partial), nil
	}
	if !d.Initialized && d.Thread == "" {
		return ioCall(d.Incarnation, nil, d.Partial), nil
	}
	if !d.SentThread && d.Thread == "" {
		defs := []any{}
		for _, t := range tools(c) {
			defs = append(defs, map[string]any{"name": t.name, "description": t.description, "inputSchema": jsonSchema(t.schema)})
		}
		params := map[string]any{"approvalPolicy": "never", "sandbox": "read-only", "dynamicTools": defs, "baseInstructions": prompt.Environment(c.Input.Header, c.Workspace, c.Scratch, c.Tools["repo.write"] != nil && proceduretools.Holds(c.Held, "repo.write"))}
		if c.Input.Model != "" {
			params["model"] = c.Input.Model
		}
		if c.Input.Fresh {
			params["config"] = map[string]any{"web_search": "disabled"}
		}
		// This invocation initialized the server: finish the handshake.
		initialized, _ := json.Marshal(map[string]any{"method": "initialized", "params": map[string]any{}})
		return ioCall(d.Incarnation, []string{string(initialized), rpc(c.Frame+":thread", "thread/start", params)}, d.Partial), nil
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
	// A delegate that names no actor inherits this thread's model, so learn
	// it before the turn: thread/start reports it, a resumed thread is asked.
	if c.Actor != "" && d.Thread != "" && d.Model == "" {
		if c.Input.Thread == "" {
			return "", errors.New("thread did not report its resolved model")
		}
		if !d.SentResume {
			return ioCall(d.Incarnation, []string{rpc(c.Frame+":resume", "thread/resume", map[string]any{"threadId": d.Thread, "approvalPolicy": "never", "sandbox": "read-only"})}, d.Partial), nil
		}
		return ioCall(d.Incarnation, nil, d.Partial), nil
	}
	if !d.SentTurn {
		return ioCall(d.Incarnation, []string{rpc(c.Frame+":turn", "turn/start", map[string]any{"threadId": d.Thread, "input": turnInput(c.Input.Context, c.Input.Text)})}, d.Partial), nil
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
		if c.Input.Fresh && !d.Stopped {
			return operation("codex.stop", d.Incarnation, "stop", nil, d.Partial), nil
		}
		return finish(c, d)
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
			return toolEmission(c, d, call.Params.Tool, args), nil
		}
		if t, _ := lookup(c, call.Params.Tool); d.Results[i].To != t.target {
			return "", fmt.Errorf("tool call %s asked for %s, but step %d invoked %s", id(call.ID), call.Params.Tool, i, d.Results[i].To)
		}
		result := toolResult(call.ID, d.Results[i])
		line, _ := json.Marshal(result)
		return ioCall(d.Incarnation, []string{string(line)}, d.Partial), nil
	}

	if d.Terminal != "" {
		if c.Input.Fresh && !d.Stopped {
			return operation("codex.stop", d.Incarnation, "stop", nil, d.Partial), nil
		}
		return finish(c, d)
	}

	return ioCall(d.Incarnation, nil, d.Partial), nil
}

func finish(c continuation, d derived) (string, error) {
	if d.Violation != "" {
		result := map[string]any{"text": "", "refused": "native." + d.Violation}
		if !c.Input.Fresh {
			result["thread"], result["incarnation"], result["activity_at"] = d.Thread, d.Incarnation, d.ActivityAt
		}
		return emit(c.Frame, result), nil
	}
	if d.Terminal != "completed" || d.Final == "" {
		if d.TerminalError != "" {
			return "", fmt.Errorf("turn ended %s: %s", d.Terminal, d.TerminalError)
		}
		return "", fmt.Errorf("turn ended %s without an answer", d.Terminal)
	}
	if c.Input.Fresh {
		return emit(c.Frame, map[string]string{"text": d.Final}), nil
	}
	return emit(c.Frame, map[string]any{"text": d.Final, "thread": d.Thread, "incarnation": d.Incarnation, "activity_at": d.ActivityAt}), nil
}

func toolResult(id json.RawMessage, s step) map[string]any {
	text, success := string(s.Value), s.State == "return" && proceduretools.Failure(s.Value) == ""
	if s.To == "repo.check" || s.To == "repo.write" || s.To == "execve" {
		var result struct {
			Error    string `json:"error"`
			ExitCode *int   `json:"exit_code"`
		}
		if json.Unmarshal(s.Value, &result) != nil || result.Error != "" || ((s.To == "repo.check" || s.To == "execve") && (result.ExitCode == nil || *result.ExitCode != 0)) {
			success = false
		}
	}
	items := []any{}
	if s.To == "desktop_observe" || s.To == "design.tools" || s.To == "blender.tools" {
		var result desktop.Result
		if json.Unmarshal(s.Value, &result) != nil || result.Version != 1 || result.Operation != s.To {
			success = false
		} else {
			if result.Error != nil {
				success = false
			}
			img := result.Screenshot
			result.Screenshot = nil
			metadata := map[string]json.RawMessage{}
			_ = json.Unmarshal(s.Value, &metadata)
			delete(metadata, "screenshot")
			b, _ := json.Marshal(metadata)
			text = string(b)
			if img != nil && success {
				if img.Validate() != nil {
					success = false
					text = "Invalid desktop screenshot"
				} else {
					items = append(items, map[string]string{"type": "inputImage", "imageUrl": "data:" + img.MIMEType + ";base64," + img.Data})
				}
			}
		}
	}

	var v string
	if json.Unmarshal(s.Value, &v) == nil {
		text = v
	}
	if s.Fault != nil {
		text = fmt.Sprintf("fault (%s): %s", s.Fault.Kind, s.Fault.Cause)
	}
	items = append([]any{map[string]string{"type": "inputText", "text": text}}, items...)
	if s.To == "read" && s.Fault == nil {
		items = nil
		for _, block := range proceduretools.ReadTexts(s.Value) {
			items = append(items, map[string]string{"type": "inputText", "text": block})
		}
	}
	return map[string]any{"id": id, "result": map[string]any{"success": success, "contentItems": items}}
}

var jsonSchema = proceduretools.JSONSchema

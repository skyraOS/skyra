// Command clauderesident is a stateless continuation adapter for the Claude
// Code provider. Each invocation receives the turn's accumulated steps on
// stdin and emits the next recorded action on stdout. It never owns a process,
// socket or descriptor; the runtime performs every action it emits.
//
// Turn lifecycle: open a one-shot session of claude-start on the service
// host, deliver the turn configuration, read stream events, dispatch bridge
// tool calls to recorded-edge targets, deliver bridge results back, and
// return the final text when the stream completes.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"skyra/controller/claude/bridge"
	"skyra/controller/claude/provider"
	"skyra/controller/codex/proceduretools"
	"skyra/controller/codex/prompt"
)

type schema = proceduretools.Schema

type continuation struct {
	Agents    map[string]json.RawMessage `json:"agents"`
	Actor     string                     `json:"actor"`
	Held      map[string][]*schema       `json:"held"`
	Contracts map[string][]*schema       `json:"contracts"`
	Workspace string                     `json:"workspace"`
	Scratch   string                     `json:"scratch"`
	Frame     string                     `json:"frame"`
	Input     struct {
		Text, Header, Conversation, Session, Model string
		Fresh                                      bool
		Context                                    json.RawMessage
	} `json:"input"`
	Tools map[string]*schema `json:"tools"`
	Steps []step             `json:"steps"`
}

type step struct {
	Observation string          `json:"observation"`
	To          string          `json:"to"`
	Payload     json.RawMessage `json:"payload"`
	State       string          `json:"state"`
	Value       json.RawMessage `json:"value"`
	Fault       *struct {
		Kind, Cause string
	} `json:"fault"`
}

type observation struct {
	Lines    []string `json:"lines"`
	Partial  string   `json:"partial"`
	Delivery string   `json:"delivery"`
	PeerEOF  bool     `json:"peer_eof"`
	Error    string   `json:"error"`
}

type bridgeCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type derived struct {
	Incarnation string
	ConfigSent  bool
	Session     string
	FinalText   string
	Done        bool
	Failed      string
	Partial     string
	Calls       []bridgeCall
	Results     []step
	Answered    map[string]bool
	ActivityAt  int64
}

// modelTools lists the tools Claude is offered. A tool appears only when it
// is also present in the continuation's reach (c.Tools).
var modelTools = proceduretools.ModelTools

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

func fail(err error) { fmt.Fprintln(os.Stderr, "clauderesident:", err); os.Exit(1) }

func loadSession() string {
	dir := os.Getenv("SKYRA_STATE")
	if dir == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(dir, "session"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func saveSession(session string) {
	dir := os.Getenv("SKYRA_STATE")
	if dir == "" || session == "" {
		return
	}
	os.WriteFile(filepath.Join(dir, "session"), []byte(session), 0600)
}

func emit(to string, payload any) string {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	return string(b)
}

func ioCall(inc string, deliver []string, prefix string) string {
	if deliver == nil {
		deliver = []string{}
	}
	return emit("claude.exchange", map[string]any{
		"incarnation": inc, "op": "exchange",
		"deliver": deliver, "wait_ms": 30000, "prefix": prefix,
	})
}

func quiet(o observation) bool {
	return strings.Contains(o.Error, "i/o timeout") || strings.Contains(o.Error, "timed out") || strings.Contains(o.Error, "deadline exceeded")
}

func decide(c continuation) (string, error) {
	if c.Frame == "" {
		return "", errors.New("no frame")
	}
	if c.Input.Session == "" && !c.Input.Fresh {
		c.Input.Session = loadSession()
	}
	if len(c.Steps) > 2000 {
		return "", errors.New("turn exceeded operation limit")
	}

	d := derived{Answered: map[string]bool{}}

	for _, s := range c.Steps {
		if !strings.HasPrefix(s.To, "claude.") {
			d.Results = append(d.Results, s)
			continue
		}
		if s.Fault != nil {
			return "", fmt.Errorf("%s fault (%s): %s", s.To, s.Fault.Kind, s.Fault.Cause)
		}
		if s.State != "return" {
			return "", fmt.Errorf("%s ended in %s", s.To, s.State)
		}
		switch s.To {
		case "claude.open":
			var opened struct {
				Incarnation, Error string
			}
			if json.Unmarshal(s.Value, &opened) != nil || opened.Incarnation == "" {
				return "", fmt.Errorf("open did not return a session: %s", opened.Error)
			}
			d.Incarnation = opened.Incarnation
			d.ConfigSent = false
			d.Done = false
			d.Failed = ""
		case "claude.stop":
			d.Incarnation = ""
		case "claude.exchange", "claude.inspect":
			var q struct {
				Deliver []string `json:"deliver"`
			}
			var o observation
			if json.Unmarshal(s.Payload, &q) != nil {
				return "", errors.New("malformed exchange payload")
			}
			if json.Unmarshal(s.Value, &o) != nil {
				return "", errors.New("malformed exchange observation")
			}
			if o.PeerEOF {
				if !d.Done && d.Failed == "" {
					d.Failed = "claude process exited before completing"
				}
				d.Incarnation = ""
				continue
			}
			if o.Error != "" && !quiet(o) {
				return "", fmt.Errorf("exchange: %s", o.Error)
			}
			d.Partial = o.Partial

			for _, line := range q.Deliver {
				var msg struct {
					Prompt string `json:"prompt"`
				}
				if json.Unmarshal([]byte(line), &msg) == nil && msg.Prompt != "" {
					d.ConfigSent = true
				}
				var br struct {
					Bridge string `json:"_bridge"`
					ID     string `json:"id"`
				}
				if json.Unmarshal([]byte(line), &br) == nil && (br.Bridge == "result" || br.Bridge == "error") {
					d.Answered[br.ID] = true
				}
			}

			for _, line := range o.Lines {
				processLine(line, &d)
			}
		default:
			return "", fmt.Errorf("unexpected step %s", s.To)
		}
	}

	if d.Failed != "" {
		return "", errors.New(d.Failed)
	}
	if d.Done {
		if !c.Input.Fresh {
			saveSession(d.Session)
		}
		return emit(c.Frame, map[string]any{
			"text":        d.FinalText,
			"session":     d.Session,
			"activity_at": d.ActivityAt,
		}), nil
	}

	if d.Incarnation == "" {
		return emit("claude.open", map[string]any{"program": "claude"}), nil
	}

	if !d.ConfigSent {
		config := turnConfig(c)
		b, _ := json.Marshal(config)
		return ioCall(d.Incarnation, []string{string(b)}, d.Partial), nil
	}

	// Dispatch pending bridge calls.
	for i, call := range d.Calls {
		if d.Answered[call.ID] {
			continue
		}
		if i >= len(d.Results) {
			args := call.Args
			if len(args) == 0 {
				args = json.RawMessage(`{}`)
			}
			t, ok := lookup(c, call.Name)
			if !ok {
				return "", fmt.Errorf("Claude requested unavailable tool %q", call.Name)
			}
			calls := make([]proceduretools.Call, len(d.Results))
			for i, s := range d.Results {
				calls[i] = proceduretools.Call{Target: s.To, Observation: s.Observation, Payload: s.Payload}
			}
			return emit(t.Target, proceduretools.Payload(t.Kind, args, calls, c.Tools, c.Contracts, proceduretools.Caller{Actor: c.Actor, Model: resolvedModel(c), Tools: tools(c)})), nil
		}
		result := d.Results[i]
		var br map[string]any
		if result.Fault != nil || result.State != "return" || proceduretools.Failure(result.Value) != "" {
			cause := "tool delivery failed"
			if failure := proceduretools.Failure(result.Value); failure != "" {
				cause = failure
			}
			if result.Fault != nil {
				cause = result.Fault.Cause
			}
			br = map[string]any{"_bridge": "error", "id": call.ID, "error": cause}
		} else {
			br = map[string]any{"_bridge": "result", "id": call.ID, "value": toolText(result)}
		}
		delivery, _ := json.Marshal(br)
		return ioCall(d.Incarnation, []string{string(delivery)}, d.Partial), nil
	}

	return ioCall(d.Incarnation, nil, d.Partial), nil
}

func processLine(line string, d *derived) {
	var bc struct {
		Bridge string          `json:"_bridge"`
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		Args   json.RawMessage `json:"args"`
	}
	if json.Unmarshal([]byte(line), &bc) == nil && bc.Bridge == "call" {
		d.Calls = append(d.Calls, bridgeCall{ID: bc.ID, Name: bc.Name, Args: bc.Args})
		return
	}
	var errMsg struct {
		Error string `json:"_error"`
	}
	if json.Unmarshal([]byte(line), &errMsg) == nil && errMsg.Error != "" {
		d.Failed = errMsg.Error
		return
	}
	// Stream events from claude. Parse for session, result, and final text.
	var e struct {
		Type      string `json:"type"`
		Subtype   string `json:"subtype"`
		SessionID string `json:"session_id"`
		IsError   bool   `json:"is_error"`
		Result    string `json:"result"`
		Message   struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	}
	if json.Unmarshal([]byte(line), &e) != nil || e.Type == "" {
		return
	}
	if e.SessionID != "" {
		d.Session = e.SessionID
	}
	if e.Type == "result" {
		if e.IsError || e.Subtype != "success" {
			d.Failed = "claude turn failed"
			if e.Subtype != "" {
				d.Failed = e.Subtype
			}
			if e.Result != "" {
				d.Failed = e.Result
			}
			return
		}
		d.Done = true
		if e.Result != "" {
			d.FinalText = e.Result
		}
		for _, block := range e.Message.Content {
			if block.Type == "text" {
				d.FinalText = block.Text
			}
		}
	}
}

// toolText extracts the text result from a tool step's Value. If the value
// is a bare string, it is used directly; otherwise it is the JSON encoding.
func toolText(s step) json.RawMessage {
	var v string
	if json.Unmarshal(s.Value, &v) == nil {
		b, _ := json.Marshal(v)
		return b
	}
	return s.Value
}

// turnConfig builds the configuration the wrapper reads from stdin.
// tools shares procedure discovery and contract dispatch with Codex.
func tools(c continuation) []proceduretools.Tool {
	var ts []proceduretools.Tool
	for _, name := range slices.Sorted(maps.Keys(modelTools)) {
		if c.Tools[name] != nil {
			ts = append(ts, proceduretools.Tool{Name: strings.ReplaceAll(name, ".", "_"), Target: name, Description: modelTools[name], Schema: c.Tools[name], Contracts: proceduretools.ContractsOf(c.Tools, c.Contracts, name)})
		}
	}
	reserved := make([]string, len(ts))
	for i, t := range ts {
		reserved[i] = t.Name
	}
	ts = append(ts, proceduretools.Procedures(c.Tools, c.Contracts, reserved, "Remember one of your tool calls by its index, starting at 0.")...)
	ts = slices.DeleteFunc(ts, func(t proceduretools.Tool) bool { return !proceduretools.Offered(c.Held, t) })
	for i := range ts {
		if ts[i].Name == "delegate" && len(c.Agents) > 0 {
			ts[i].Description += " Available actors: " + strings.Join(slices.Sorted(maps.Keys(c.Agents)), ", ") + "."
		}
	}
	return ts
}
func lookup(c continuation, name string) (proceduretools.Tool, bool) {
	for _, t := range tools(c) {
		if t.Name == name {
			return t, true
		}
	}
	return proceduretools.Tool{}, false
}
func resolvedModel(c continuation) string {
	if c.Input.Model != "" {
		return c.Input.Model
	}
	return provider.DefaultModel
}

func turnConfig(c continuation) map[string]any {
	ts := []bridge.Tool{}
	for _, t := range tools(c) {
		b, _ := json.Marshal(proceduretools.JSONSchema(t.Schema))
		ts = append(ts, bridge.Tool{Name: t.Name, Description: t.Description, InputSchema: b})
	}
	cfg := map[string]any{
		"prompt": c.Input.Text, "frame": c.Frame, "tools": ts, "model": resolvedModel(c),
		"system_prompt": prompt.Environment(c.Input.Header, c.Workspace, c.Scratch, c.Tools["repo.write"] != nil && proceduretools.Holds(c.Held, "repo.write")),
	}
	if !c.Input.Fresh && c.Input.Session != "" {
		cfg["session"] = c.Input.Session
	}
	return cfg
}

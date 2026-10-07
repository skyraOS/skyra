// Command clauderesident is a stateless continuation adapter for the Claude
// Code provider. Each invocation receives the turn's accumulated steps on
// stdin and emits the next recorded action on stdout. It never owns a process,
// socket or descriptor; the runtime performs every action it emits.
//
// Turn lifecycle: start a one-shot resident, deliver the turn configuration,
// read stream events, dispatch bridge tool calls to recorded-edge targets,
// deliver bridge results back, and return the final text when the stream
// completes.
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
	"skyra/controller/codex/proceduretools"
	"skyra/controller/codex/prompt"
)

type schema = proceduretools.Schema

type continuation struct {
	Frame string `json:"frame"`
	Input struct {
		Text, Header, Conversation, Session string
		Context                             json.RawMessage
	} `json:"input"`
	Tools map[string]*schema `json:"tools"`
	Steps []step             `json:"steps"`
}

type step struct {
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
	State   string          `json:"state"`
	Value   json.RawMessage `json:"value"`
	Fault   *struct {
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
var modelTools = map[string]string{
	"repo.check": "Build and verify repository changes using named recipes, without a shell command. action: go.test, go.build, go.format (reports formatting needed), swift.test, swift.build, python.test (Blender tests). Go actions accept optional package such as ./controller/blender or ./controller/... . No other arguments, commands or environment overrides. Runs current source in a disposable snapshot; source and installed toolchains are read-only, writes stay in the job and bounded Swift temporary replacement folders, credentials are absent, network is limited to loopback and job-local Unix sockets. Returns exit_code, output (64 KiB limit), truncated, error, duration_ms and report_path under ideas/checks. Nonzero exit, truncation, or a sandbox denial must be reported; never claim checks passed without their results. Build artifacts are temporary. Read report_path to review the result.",
	"repo.write": "Create, replace, or append UTF-8 text in this repository. Input: path (relative to repository root), text, optional append. Parent folders are created. Read it back using read. No delete or rename operation; .git, runtime state, symlinks and hardlinks cannot be written.",
	"execve":     "Run Git in this repository as a recorded actor: {target:\"git\",args:[\"status\",\"--short\"]}. args is an argv array, never a shell command. Commit, push, fetch, merge, add, branch/tag creation and inspection are available. No approval is needed for commit, push or merge. File/ref deletion, destructive reset/checkout/clean, config changes, hooks, custom filters, external drivers, arbitrary programs and repository switching are refused. Commits and merges cannot remove existing file paths; pushes cannot remove paths from an existing destination tree. Use a configured remote name. Merge takes one branch/commit and previews it; conflicting merges are refused before touching the worktree. Output includes exit_code, output, truncated and error.",
	"echo":       "Skyra echo",
	"read":       "Read a UTF-8 text file, or list a directory, in the workspace. Paths are relative to the workspace root; \".\" lists the root. Returns at most 2000 lines or 64 KiB: when truncated is true, call again with start_line set to end_line + 1. sha256 and size describe the whole file. A failure is reported in error (not_found, denied, not_text, too_large).",
	"find":       "Find files (or folders) in the workspace by name, path, depth, kind, age or size, without reading them. in is where to look, relative to the workspace root (default \".\"). include and exclude are globs: one without a slash matches a name (\"*.go\"), one with a slash the path (\"src/*.go\"); an excluded folder is not entered. kind is \"file\" (default), \"dir\" or \"any\". maxdepth limits depth below in; newer is an RFC 3339 time; larger is a size in bytes. .git and node_modules are skipped unless ignore is \"none\", and dot files unless hidden is true; the result counts what was skipped. Paths come back as read takes them. At most max_results (default 100, up to 500): truncated says when the list was cut and why.",
	"roots":      "Show the absolute paths you may see: workspace (where a relative read path resolves) and every path you may read.",
	"write":      "Write UTF-8 text inside the granted scratch directory shown by roots, for notes, drafts and intermediate results. path is relative to the scratch space and cannot leave it; folders are created as needed. append: true adds to the end instead of replacing.",
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
	if c.Input.Session == "" {
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
		case "claude.start":
			var started struct {
				Incarnation string `json:"incarnation"`
			}
			if json.Unmarshal(s.Value, &started) != nil || started.Incarnation == "" {
				return "", errors.New("start did not return an incarnation")
			}
			d.Incarnation = started.Incarnation
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
		saveSession(d.Session)
		return emit(c.Frame, map[string]any{
			"text":        d.FinalText,
			"session":     d.Session,
			"activity_at": d.ActivityAt,
		}), nil
	}

	if d.Incarnation == "" {
		return emit("claude.start", map[string]any{}), nil
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
			return emit(call.Name, json.RawMessage(args)), nil
		}
		result := d.Results[i]
		var br map[string]any
		if result.Fault != nil || result.State != "return" {
			cause := "tool delivery failed"
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
func turnConfig(c continuation) map[string]any {
	var tools []bridge.Tool
	for _, name := range slices.Sorted(maps.Keys(modelTools)) {
		if c.Tools[name] == nil {
			continue
		}
		schema := proceduretools.JSONSchema(c.Tools[name])
		b, _ := json.Marshal(schema)
		tools = append(tools, bridge.Tool{
			Name:        name,
			Description: modelTools[name],
			InputSchema: b,
		})
	}
	cfg := map[string]any{
		"prompt":        c.Input.Text,
		"frame":         c.Frame,
		"tools":         tools,
		"system_prompt": prompt.Base(c.Input.Header),
	}
	if c.Input.Session != "" {
		cfg["session"] = c.Input.Session
	}
	return cfg
}

package drive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"skyra/apprentice/tree"
)

// ClaudeModel is the model the machine startles: the local Claude Code CLI
// (`claude -p`) under the user's existing login, as controller/claude's
// provider runs it. Its own tools, commands, hooks and MCP servers are off,
// and inherited API credentials are refused rather than silently switching
// a subscription to metered billing.
//
// One ask is one CLI session. Each reply is one JSON object: an expand,
// answered by resuming the session (ExpandLimit times, never counted
// against the refinement bound), or the guard, which ends the ask.
type ClaudeModel struct {
	Program string // the CLI; "claude" on PATH by default
	Model   string
	Effort  string
	// Run runs the CLI with these arguments and environment and gives its
	// stdout; exec by default.
	Run func(args, env []string) ([]byte, error)
}

// NewClaudeModel is Claude Opus 5.5 through the claude CLI.
func NewClaudeModel() *ClaudeModel {
	return &ClaudeModel{Program: "claude", Model: "claude-opus-5-5", Effort: "high"}
}

const claudeSystem = `You write guards for a tree that learns. Each time it reaches a place it has not learned, it asks you what to do there. What you write is stored and reused: the next time anything arrives at the same base with values your guard selects, it runs with no model call.

Observations are named $1, $2, ... in the frame's scope. A tool observation names a binary the frame may run. Values are truncated where marked.

A guard is CEL that returns an instruction or null:
  size($2) > 0 ? "$1 $2" : null
  $3.exit != 0 ? "return $3" : $3.contains("b.log") ? "summarize $1 $3" : null
Conditions read observations by $n and their fields by key ($3.size, $3["content-type"]). Only $n listed in the base are learned; a guard reading another runs once and is not stored.

An instruction is read by its first field:
  return $n            returns $n to the frame's caller
  $t args...           $t is a tool: runs its binary with the args, each $n replaced by its value
  task $a $b ...       delegates to a child frame named task, passing only $n; the child must be passed fewer tools than this frame holds

Anywhere an instruction takes $n it may take a cut of it instead:
  $2[3:7]              lines 3 to 7 of $2's text, counted from 1
  $2[1,4:5]            several runs of lines, joined
  $2.dir               the field dir of $2, a JSON object ($2.dir.path goes deeper)
Never copy text out of an observation into a guard or an instruction: what you write is reused on new values, so a copied value is only right for this one. Say where it comes from with a cut, and condition on shapes and fields rather than exact values.

Your guard must select something for what has arrived.

Reply with exactly one JSON object and nothing else:
  {"expand": {"n": 3, "from": 2048}}   to read more of a truncated value from a byte offset
  {"guard": "<your guard>"}            to answer`

// arguments are the provider's: print mode, the user's login and settings,
// and nothing of Claude Code's own.
func (m *ClaudeModel) arguments(session, prompt string) []string {
	args := []string{"-p", "--output-format", "json", "--model", m.Model, "--tools", "", "--disable-slash-commands",
		"--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--setting-sources", "user",
		"--settings", `{"disableAllHooks":true,"switchModelsOnFlag":false}`, "--system-prompt", claudeSystem}
	if m.Effort != "" {
		args = append(args, "--effort", m.Effort)
	}
	if session != "" {
		args = append(args, "--resume", session)
	}
	return append(args, "--", prompt)
}

// environment refuses inherited API and provider credentials: the CLI uses
// the subscription login under HOME, never metered API billing.
func environment(env []string) ([]string, error) {
	out := make([]string, 0, len(env))
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		switch key {
		case "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_BASE_URL",
			"CLAUDE_CODE_USE_BEDROCK", "CLAUDE_CODE_USE_VERTEX", "CLAUDE_CODE_USE_FOUNDRY":
			if value != "" {
				return nil, fmt.Errorf("the subscription-only Claude model refuses %s", key)
			}
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}

func (m *ClaudeModel) run(args, env []string) ([]byte, error) {
	if m.Run != nil {
		return m.Run(args, env)
	}
	program := m.Program
	if program == "" {
		program = "claude"
	}
	cmd := exec.Command(program, args...)
	cmd.Env = env
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %v: %s", program, err, strings.TrimSpace(errOut.String()))
	}
	return out.Bytes(), nil
}

// A reply is one expand or the guard.
type reply struct {
	Expand *struct{ N, From int }
	Guard  *string
}

// parseReply reads the JSON object in the model's text, allowing a fence
// around it. Text that holds none is taken as the guard, and the machine
// refuses it if it is not one.
func parseReply(text string) reply {
	s := strings.TrimSpace(text)
	if i, j := strings.Index(s, "{"), strings.LastIndex(s, "}"); i >= 0 && j > i {
		var r reply
		if json.Unmarshal([]byte(s[i:j+1]), &r) == nil && (r.Expand != nil || r.Guard != nil) {
			return r
		}
	}
	return reply{Guard: &s}
}

// Guard asks Claude for a guard at the ask, answering its expansions in
// the same session.
func (m *ClaudeModel) Guard(a tree.AskModel) (string, error) {
	env, err := environment(os.Environ())
	if err != nil {
		return "", err
	}
	session, prompt := "", Render(a)
	for expands := 0; ; {
		out, err := m.run(m.arguments(session, prompt), env)
		if err != nil {
			return "", err
		}
		var res struct {
			Result    string `json:"result"`
			SessionID string `json:"session_id"`
			IsError   bool   `json:"is_error"`
		}
		if err := json.Unmarshal(out, &res); err != nil {
			return "", fmt.Errorf("the claude CLI's output is not its JSON result: %v", err)
		}
		if res.IsError {
			return "", errors.New("the claude CLI failed: " + res.Result)
		}
		session = res.SessionID
		r := parseReply(res.Result)
		if r.Guard != nil {
			return *r.Guard, nil
		}
		switch {
		case expands >= ExpandLimit:
			prompt = fmt.Sprintf("Expanded %d times in this ask, the most allowed: reply with your guard.", expands)
		default:
			if prompt, err = Expand(a, r.Expand.N, r.Expand.From); err != nil {
				prompt = "Not expanded: " + err.Error()
			} else {
				expands++
			}
		}
	}
}

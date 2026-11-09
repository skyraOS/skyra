package drive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"skyra/typesys-prototype/model"
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

// The system prompts, one per kind: each shows only what its kind may
// write. Neither says how many tools make a kind: the act bound is the
// owner's.
const systemHead = `Write one CEL expression. It decides what runs next.

You are given parameters $1, $2, ..., grouped by role, each with its type and value:
  Goal      $1: what you are to do
  Tools     programs
  Context   what you were passed, and what has arrived since

The expression returns an instruction, as a string, or null:
  size($4) > 0 ? "return 0 $4" : null
Each condition is followed by one instruction, and the chain ends in one null, after the last instruction:
  c ? "a" : null
  c1 ? "a" : c2 ? "b" : null
  c1 ? "a" : true ? "b" : null      b applies otherwise
There is no else: c1 ? "a" : "b" is refused, and so is c1 ? "a" : c2 ? "b" : "c" : null. For otherwise, end with true ? "c" : null.
It may use CEL's standard functions and the strings extension, called as methods: $1.trim(), $1.split("\n"), $1.lowerAscii(), $1.indexOf("x"), $1.substring(0, 3), $1.replace("a", "b"). A program's output usually ends in a newline: trim it before an exact match.

`

const actInstructions = `Instructions:
  $2 $1 ...                    run the tool $2, with the other parameters as its arguments
  return 0 $4                  finish: the goal is done, with $4 as the answer
  return 1 $4                  finish: the goal is not done, with $4 as what there is
  return 1 $4 $5               the same, with $5 as why (either return may name a why)
`

const delegateInstructions = `Instructions:
  delegate(goal($1), $2, $4)   delegate: goal(...) is the goal you delegate, taken from a parameter or part of one;
                               the rest is what you pass: at least one tool, and fewer tools than you have
  return 0 $4                  finish: the goal is done, with $4 as the answer
  return 1 $4                  finish: the goal is not done, with $4 as what there is
  return 1 $4 $5               the same, with $5 as why (either return may name a why)
You do not run tools yourself: you delegate.
`

const systemTail = `
An instruction is made of parameters, each a $n or part of one, in one of the forms above; a return also takes its exit status, 0 or 1. Where a tool needs text no parameter holds, such as a flag or a filter, write it as a literal in single quotes:
  $2 '-l' $1[2]
  $2 '.version' $1[2]
Where a tool needs text built from parameters, such as a path from a directory and a name, write the parts together with no space between them, parameters and literals:
  $2 $1[2]'/'$5.stdout[1]
A literal never writes what a cut of a parameter takes, a whole value, a line or lines, or a field: one that does is refused, so cut it instead. Write nothing else: any other text in an instruction is rejected.

Use part of a parameter instead of copying its text:
  $2[2]              line 2 of $2
  $2[3:7]            lines 3 to 7
  $2.dir             the field dir of a JSON value ($2.dir.path goes deeper)
  $4.stdout[1]       line 1 of $4's field stdout: each cut cuts what the last gave
Your expression is reused when new values arrive, so never write a value's text into it: a literal is reused too, so write only what holds for every value, a flag or a filter, never a name or a path.
It is reused for other goals too: any goal whose parameters have the same types reaches the same expression, whatever it asks. So guard on the goal whatever you take from what it asks, a filter, a flag, what to return, and leave a goal that asks something else to null:
  $1.contains("version") ? "$2 '--version'" : null
What you would do whatever the goal asked needs no such guard.

Reply with JSON and nothing else:
  {"cel": "<expression>"}
  {"expand": {"n": 3, "from": 2048}}   to see more of a value cut off at "…[truncated"
`

// System is the system prompt for a frame of this kind.
func System(k model.Kind) string {
	if k == model.KindDelegate {
		return systemHead + delegateInstructions + systemTail
	}
	return systemHead + actInstructions + systemTail
}

// arguments are the provider's: print mode, the user's login and settings,
// and nothing of Claude Code's own.
func (m *ClaudeModel) arguments(system, session, prompt string) []string {
	args := []string{"-p", "--output-format", "json", "--model", m.Model, "--tools", "", "--disable-slash-commands",
		"--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--setting-sources", "user",
		"--settings", `{"disableAllHooks":true,"switchModelsOnFlag":false}`, "--system-prompt", system}
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
		// The CLI reports most failures as its JSON result on stdout, with a
		// failing exit: keep both.
		return nil, fmt.Errorf("%s: %v: %s %s", program, err, strings.TrimSpace(errOut.String()), strings.TrimSpace(out.String()))
	}
	return out.Bytes(), nil
}

// A reply is one expand or the guard, written as CEL.
type reply struct {
	Expand *struct{ N, From int }
	Guard  *string `json:"cel"`
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
func (m *ClaudeModel) Guard(a model.Ask) (string, error) {
	env, err := environment(os.Environ())
	if err != nil {
		return "", err
	}
	session, prompt := "", Render(a)
	for expands := 0; ; {
		out, err := m.run(m.arguments(System(a.Kind), session, prompt), env)
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
			prompt = fmt.Sprintf("Expanded %d times, the most allowed: reply with your CEL.", expands)
		default:
			if prompt, err = Expand(a, r.Expand.N, r.Expand.From); err != nil {
				prompt = "Not expanded: " + err.Error()
			} else {
				expands++
			}
		}
	}
}

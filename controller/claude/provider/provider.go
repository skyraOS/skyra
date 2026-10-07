// Package provider describes the Claude Code transport. Process ownership and
// recording belong to the resident holder, not to a continuation worker.
package provider

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

// Arguments uses the user's existing Claude Code login. Native tools and
// ambient MCP servers are disabled; Skyra tools must use a recorded bridge.
func Arguments(session, prompt, systemPrompt string) []string {
	args := []string{"-p", "--output-format", "stream-json", "--verbose", "--include-partial-messages", "--model", "claude-opus-5-5", "--tools", "", "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--setting-sources", "", "--settings", `{"disableAllHooks":true}`}
	if systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}
	if session != "" { args = append(args, "--resume", session) }
	return append(args, "--", prompt)
}

// BridgeArguments adds a capability-protected loopback MCP server. These
// arguments contain an ephemeral capability and must not be logged/recorded.
// The holder must externally sandbox the CLI before launching these arguments.
func BridgeArguments(session, prompt, systemPrompt, endpoint, token string, tools []string) ([]string, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("Claude bridge must be a loopback HTTP endpoint")
	}
	if _, _, err := net.SplitHostPort(u.Host); err != nil { return nil, errors.New("invalid bridge address") }
	if len(token) < 32 || strings.ContainsAny(token, "\r\n") { return nil, errors.New("invalid bridge capability") }
	allowed := make([]string, 0, len(tools))
	seen := make(map[string]bool)
	for _, name := range tools {
		if name == "" || strings.ContainsAny(name, " ,/\\*()") || seen[name] { return nil, errors.New("invalid bridge tool name") }
		seen[name] = true
		allowed = append(allowed, "mcp__skyra__"+name)
	}
	config, _ := json.Marshal(map[string]any{"mcpServers":map[string]any{"skyra":map[string]any{"type":"http","url":endpoint,"headers":map[string]string{"Authorization":"Bearer "+token}}}})
	args := Arguments(session, prompt, systemPrompt)
	for i := range args { if args[i] == "--mcp-config" { args[i+1] = string(config); break } }
	if len(allowed) > 0 { args = append(args[:len(args)-2], append([]string{"--allowedTools", strings.Join(allowed, ","), "--", prompt}, []string{}...)...) }
	return args, nil
}

// Environment rejects inherited API/provider credentials rather than silently
// switching a subscription to metered API billing. HOME retains the CLI login.
func Environment(env []string) ([]string, error) {
	out := make([]string, 0, len(env))
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		switch key {
		case "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_BASE_URL", "CLAUDE_CODE_USE_BEDROCK", "CLAUDE_CODE_USE_VERTEX", "CLAUDE_CODE_USE_FOUNDRY":
			if value != "" { return nil, fmt.Errorf("subscription-only Claude provider refuses %s", key) }
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}

type Block struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Name string `json:"name"`
}

type Event struct {
	Type string `json:"type"`
	Subtype string `json:"subtype"`
	SessionID string `json:"session_id"`
	Result string `json:"result"`
	IsError bool `json:"is_error"`
	Errors []string `json:"errors"`
	Message struct { Content []Block `json:"content"` } `json:"message"`
	Event struct { ContentBlock Block `json:"content_block"` } `json:"event"`
	Raw json.RawMessage `json:"-"`
}

func Read(r io.Reader, observe func(Event) error) error { return ReadWithTools(r, nil, observe) }

// ReadWithTools admits only explicitly offered Skyra MCP tools, including
// partial content-block starts. Permission settings never expand this allowlist.
// Truncated records and streams without a terminal result are never successes.
func ReadWithTools(r io.Reader, names []string, observe func(Event) error) error {
	allowed := make(map[string]bool)
	for _, name := range names { allowed["mcp__skyra__"+name] = true }
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 8<<20)
	terminal := false
	session := ""
	for scanner.Scan() {
		if terminal { return errors.New("Claude emitted data after terminal result") }
		var e Event
		line := append([]byte(nil), scanner.Bytes()...)
		if err := json.Unmarshal(line, &e); err != nil { return fmt.Errorf("malformed Claude stream: %w", err) }
		if e.Type == "" { return errors.New("Claude event lacks type") }
		e.Raw = line
		if e.SessionID != "" {
			if session != "" && e.SessionID != session { return errors.New("Claude session changed during turn") }
			session = e.SessionID
		}
		blocks := append(append([]Block(nil), e.Message.Content...), e.Event.ContentBlock)
		for _, block := range blocks {
			if block.Type == "tool_use" && !allowed[block.Name] { return errors.New("refused unexpected Claude tool") }
		}
		if e.Type == "result" {
			terminal = true
			if e.IsError || e.Subtype != "success" { return errors.New("Claude turn failed") }
			if e.SessionID == "" { return errors.New("Claude result lacks session ID") }
		}
		if err := observe(e); err != nil { return err }
	}
	if err := scanner.Err(); err != nil { return err }
	if !terminal { return errors.New("Claude stream ended without terminal result") }
	return nil
}

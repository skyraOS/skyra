// Deterministic Claude transport fixture for recorded delegation tests.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	d := json.NewDecoder(os.Stdin)
	e := json.NewEncoder(os.Stdout)
	var cfg struct {
		Prompt, Frame, Session, Model string
		Tools                         []struct{ Name string }
	}
	if d.Decode(&cfg) != nil {
		return
	}
	if cfg.Session != "" {
		_ = e.Encode(map[string]string{"_error": "child reused a session"})
		return
	}
	if cfg.Prompt == "provider-exit" {
		return
	}

	call := func(name string, args any) string {
		_ = e.Encode(map[string]any{"_bridge": "call", "id": cfg.Frame + ":" + name, "name": name, "args": args})
		var result struct {
			Value json.RawMessage
			Error string
		}
		if d.Decode(&result) != nil {
			return "fixture stream closed"
		}
		if result.Error != "" {
			return result.Error
		}
		return string(result.Value)
	}
	text := cfg.Prompt
	if strings.HasPrefix(text, "delegate:") {
		var args map[string]any
		if json.Unmarshal([]byte(strings.TrimPrefix(text, "delegate:")), &args) != nil {
			return
		}
		text = call("delegate", map[string]any{"delegate": args}) + " | " + call("echo", map[string]string{"text": "parent continued"})
	}
	if strings.HasPrefix(text, "tool:") {
		offered := false
		for _, t := range cfg.Tools {
			if t.Name == "echo" {
				offered = true
			}
		}
		if !offered {
			_ = e.Encode(map[string]string{"_error": "echo unavailable"})
			return
		}
		_ = e.Encode(map[string]any{"_bridge": "call", "id": "child-call", "name": "echo", "args": map[string]string{"text": strings.TrimPrefix(text, "tool:")}})
		var result struct {
			Value json.RawMessage
			Error string
		}
		if d.Decode(&result) != nil {
			return
		}
		if result.Error != "" {
			_ = e.Encode(map[string]string{"_error": result.Error})
			return
		}
		text = string(result.Value)
	}
	_ = e.Encode(map[string]any{"type": "result", "subtype": "success", "session_id": "fresh-child-session", "result": fmt.Sprintf("%s; model=%s", text, cfg.Model)})
}

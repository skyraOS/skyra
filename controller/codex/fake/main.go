// The fixture speaks the selected app-server JSON-RPC subset, never Skyra Moves.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skyra/controller/codex/prompt"
)

type message struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
}
type turn struct{ thread, id, text string }

func main() {
	enc := json.NewEncoder(os.Stdout)
	d := json.NewDecoder(os.Stdin)
	send := func(v any) {
		if err := enc.Encode(v); err != nil {
			os.Exit(1)
		}
	}
	reply := func(id json.RawMessage, v any) { send(map[string]any{"id": id, "result": v}) }
	event := func(method string, t turn, item any) {
		params := map[string]any{"threadId": t.thread, "turnId": t.id}
		if item != nil {
			params["item"] = item
		} else {
			params["turn"] = map[string]string{"id": t.id, "status": "completed"}
		}
		send(map[string]any{"method": method, "params": params})
	}
	complete := func(t turn, text string) {
		event("item/completed", t, map[string]string{"type": "agentMessage", "text": text})
		event("turn/completed", t, nil)
	}
	threads, turns := 0, 0
	pending := map[string]turn{}
	// A scripted turn (remember:<x>, next/32): calls made one after another,
	// each after the last one's result, then an answer made of every result.
	type call struct {
		tool string
		args any
	}
	script := map[string][]call{}
	results := map[string][]string{}
	toolCall := func(t turn, n int, c call) {
		id := json.RawMessage(fmt.Sprintf("\"call-%s-%d\"", t.id, n))
		pending[string(id)] = t
		send(map[string]any{"id": id, "method": "item/tool/call", "params": map[string]any{"threadId": t.thread, "turnId": t.id, "callId": string(id), "tool": c.tool, "arguments": c.args}})
	}
	var held *turn
	history := map[string][]string{}
	prompts := map[string]string{}  // each thread's base prompt, as Skyra gave it
	recalled := map[string]string{} // each turn's context, as Skyra gave it
	seen := map[string][]string{}   // each turn's local images, by file name
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	for {
		var m message
		if d.Decode(&m) != nil {
			return
		}
		if m.Method == "" {
			t, ok := pending[string(m.ID)]
			if ok {
				delete(pending, string(m.ID))
				var result struct {
					ContentItems []struct {
						Text string `json:"text"`
					} `json:"contentItems"`
				}
				_ = json.Unmarshal(m.Result, &result)
				if len(result.ContentItems) >= 1 {
					if rest, scripted := script[t.id]; scripted {
						results[t.id] = append(results[t.id], result.ContentItems[0].Text)
						if len(rest) > 0 {
							script[t.id] = rest[1:]
							toolCall(t, len(results[t.id]), rest[0])
						} else {
							complete(t, strings.Join(results[t.id], " | "))
						}
						continue
					}
					texts := []string{}
					for _, item := range result.ContentItems {
						texts = append(texts, item.Text)
					}
					complete(t, "tool:"+strings.Join(texts, " | "))
				}
			}
			continue
		}
		var v struct {
			ThreadID string `json:"threadId"`
			TurnID   string `json:"turnId"`
			Input    []struct {
				Type string `json:"type"`
				Text string `json:"text"`
				Path string `json:"path"`
			} `json:"input"`
			Developer string `json:"developerInstructions"`
			Base      string `json:"baseInstructions"`
			Tools     []struct {
				Name string `json:"name"`
			} `json:"dynamicTools"`
		}
		_ = json.Unmarshal(m.Params, &v)
		switch m.Method {
		case "initialize":
			if mode == "hang-connect" {
				continue
			}
			if mode == "bad-connect" {
				reply(m.ID, map[string]any{})
				continue
			}
			reply(m.ID, map[string]string{"userAgent": "dna19-fake/1"})
		case "initialized":
		case "turn/interrupt":
			reply(m.ID, map[string]any{})
			// An interrupt ends the named turn, as Codex does, whether it is
			// held or waiting on a tool call the adapter will never answer.
			interrupted := held != nil && held.thread == v.ThreadID && held.id == v.TurnID
			if interrupted {
				held = nil
			}
			for id, t := range pending {
				if t.thread == v.ThreadID && t.id == v.TurnID {
					delete(pending, id)
					interrupted = true
				}
			}
			if interrupted {
				send(map[string]any{"method": "turn/completed", "params": map[string]any{
					"threadId": v.ThreadID, "turn": map[string]string{"id": v.TurnID, "status": "interrupted"}}})
			}
		case "thread/loaded/list":
			if mode == "hang-health" {
				continue
			}
			reply(m.ID, map[string]any{"data": []string{}})
		case "thread/start", "thread/fork":
			if m.Method == "thread/start" && (v.Developer != "" || !strings.HasPrefix(v.Base, prompt.Prefix)) {
				// Proposal 38: the whole prompt is Skyra's; the provider's
				// default and the old developer line must never reach a thread.
				send(map[string]any{"id": m.ID, "error": map[string]any{"code": -1, "message": "thread/start without Skyra's base prompt"}})
				continue
			}
			threads++
			id := fmt.Sprintf("thread-%d", threads)
			prompts[id] = v.Base
			if m.Method == "thread/fork" {
				history[id] = append([]string(nil), history[v.ThreadID]...)
				prompts[id] = prompts[v.ThreadID]
			}
			if mode == "reject-replace" && threads > 1 {
				send(map[string]any{"id": m.ID, "error": map[string]any{"code": -1, "message": "replacement unsupported"}})
				continue
			}
			reply(m.ID, map[string]any{"thread": map[string]string{"id": id}, "approvalPolicy": "never", "sandbox": map[string]any{"type": "readOnly", "networkAccess": false}})
		case "turn/start":
			turns++
			if len(v.Input) == 0 {
				send(map[string]any{"id": m.ID, "error": map[string]any{"code": -1, "message": "turn/start without input"}})
				continue
			}
			// The message is the turn's last input; anything before it is
			// the context its invoker passed (proposal 38).
			t := turn{v.ThreadID, fmt.Sprintf("turn-%d", turns), v.Input[len(v.Input)-1].Text}
			recalled[t.id], seen[t.id] = "", nil
			for _, earlier := range v.Input[:len(v.Input)-1] {
				if earlier.Type == "localImage" {
					seen[t.id] = append(seen[t.id], filepath.Base(earlier.Path))
					continue
				}
				recalled[t.id] += earlier.Text
			}
			history[t.thread] = append(history[t.thread], t.text)
			reply(m.ID, map[string]any{"turn": map[string]string{"id": t.id}})
			event("turn/started", t, nil)
			switch {
			case t.text == "provider-exit":
				return
			case t.text == "hold":
				held = &t
			case t.text == "native-running":
				held = &t
				event("item/started", t, map[string]string{"type": "commandExecution"})
			case t.text == "unregistered-tool":
				held = &t
				send(map[string]any{"id": "unregistered", "method": "item/tool/call", "params": map[string]any{
					"threadId": t.thread, "turnId": t.id, "tool": "reply", "arguments": map[string]string{"text": "must not bypass registration"}}})
			case t.text == "release":
				complete(t, "isolated:"+t.thread)
				if held != nil {
					complete(*held, "held:"+held.thread)
					held = nil
				}
			case strings.HasPrefix(t.text, "read:"):
				id := json.RawMessage(fmt.Sprintf("\"call-%d\"", turns))
				pending[string(id)] = t
				send(map[string]any{"id": id, "method": "item/tool/call", "params": map[string]any{"threadId": t.thread, "turnId": t.id, "callId": string(id), "tool": "read", "arguments": map[string]string{"path": strings.TrimPrefix(t.text, "read:")}}})
			case strings.HasPrefix(t.text, "find:"):
				// find:GLOB finds files in the workspace by name.
				id := json.RawMessage(fmt.Sprintf("\"call-%d\"", turns))
				pending[string(id)] = t
				send(map[string]any{"id": id, "method": "item/tool/call", "params": map[string]any{"threadId": t.thread, "turnId": t.id, "callId": string(id), "tool": "find", "arguments": map[string]any{"include": []string{strings.TrimPrefix(t.text, "find:")}}}})
			case t.text == "roots":
				id := json.RawMessage(fmt.Sprintf("\"call-%d\"", turns))
				pending[string(id)] = t
				send(map[string]any{"id": id, "method": "item/tool/call", "params": map[string]any{"threadId": t.thread, "turnId": t.id, "callId": string(id), "tool": "roots", "arguments": map[string]string{}}})
			case strings.HasPrefix(t.text, "write:"):
				// write:PATH=TEXT writes TEXT to PATH in the scratch space.
				path, text, _ := strings.Cut(strings.TrimPrefix(t.text, "write:"), "=")
				id := json.RawMessage(fmt.Sprintf("\"call-%d\"", turns))
				pending[string(id)] = t
				send(map[string]any{"id": id, "method": "item/tool/call", "params": map[string]any{"threadId": t.thread, "turnId": t.id, "callId": string(id), "tool": "write", "arguments": map[string]string{"path": path, "text": text}}})
			case strings.HasPrefix(t.text, "tool:") || t.text == "taint-child":
				id := json.RawMessage(fmt.Sprintf("\"call-%d\"", turns))
				pending[string(id)] = t
				send(map[string]any{"id": id, "method": "item/tool/call", "params": map[string]any{"threadId": t.thread, "turnId": t.id, "callId": string(id), "tool": "echo", "arguments": map[string]string{"text": strings.TrimPrefix(t.text, "tool:")}}})
				if t.text == "taint-child" {
					event("item/started", t, map[string]string{"type": "commandExecution"})
					complete(t, "must reject")
				}
			case t.text == "taint":
				event("item/started", t, map[string]string{"type": "fileChange"})
				complete(t, "must reject")
			case t.text == "unknown":
				send(map[string]any{"id": "unknown", "method": "approval/unknown", "params": map[string]string{"threadId": t.thread, "turnId": t.id}})
				complete(t, "must reject")
			case strings.HasPrefix(t.text, "remember:"):
				x := strings.TrimPrefix(t.text, "remember:")
				calls := []call{
					{"agent", map[string]string{"text": "relay:" + x}},
					{"remember", map[string]int{"call": 0}},
					{"procedure", map[string]any{"input": map[string]string{"text": "relay:" + x + ", again"}}},
				}
				script[t.id] = calls[1:]
				toolCall(t, 0, calls[0])
			case t.text == "describe":
				// What a real model would look at: the images it was given.
				complete(t, "saw:"+strings.Join(seen[t.id], ","))
			case t.text == "prompt":
				complete(t, strings.ReplaceAll(prompts[t.thread], "\n", " | "))
			case t.text == "history":
				// What this turn remembers: the context it was given, its
				// thread's own messages, and how many threads this app-server
				// has started (more than one: it was reused).
				complete(t, fmt.Sprintf("%s|%s|threads=%d", strings.ReplaceAll(recalled[t.id], "\n", " "), strings.Join(history[t.thread], "|"), threads))
			default:
				complete(t, t.thread+":"+t.text)
			}
		default:
			send(map[string]any{"id": m.ID, "error": map[string]any{"code": -32601, "message": "unsupported"}})
		}
	}
}

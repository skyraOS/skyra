// The fixture speaks the selected app-server JSON-RPC subset, never Skyra Moves.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	var held *turn
	history := map[string][]string{}
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
					complete(t, "tool:"+result.ContentItems[0].Text)
				}
			}
			continue
		}
		var v struct {
			ThreadID string `json:"threadId"`
			TurnID   string `json:"turnId"`
			Input    []struct {
				Text string `json:"text"`
			} `json:"input"`
			Developer string `json:"developerInstructions"`
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
			if held != nil && held.thread == v.ThreadID && held.id == v.TurnID {
				send(map[string]any{"method": "turn/completed", "params": map[string]any{
					"threadId": held.thread, "turn": map[string]string{"id": held.id, "status": "interrupted"}}})
				held = nil
			}
		case "thread/loaded/list":
			if mode == "hang-health" {
				continue
			}
			reply(m.ID, map[string]any{"data": []string{}})
		case "thread/start", "thread/fork":
			threads++
			id := fmt.Sprintf("thread-%d", threads)
			if m.Method == "thread/fork" {
				history[id] = append([]string(nil), history[v.ThreadID]...)
			}
			if mode == "reject-replace" && threads > 1 {
				send(map[string]any{"id": m.ID, "error": map[string]any{"code": -1, "message": "replacement unsupported"}})
				continue
			}
			reply(m.ID, map[string]any{"thread": map[string]string{"id": id}, "approvalPolicy": "never", "sandbox": map[string]any{"type": "readOnly", "networkAccess": false}})
		case "turn/start":
			turns++
			t := turn{v.ThreadID, fmt.Sprintf("turn-%d", turns), v.Input[0].Text}
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
			case t.text == "history":
				complete(t, strings.Join(history[t.thread], "|"))
			default:
				complete(t, t.thread+":"+t.text)
			}
		default:
			send(map[string]any{"id": m.ID, "error": map[string]any{"code": -32601, "message": "unsupported"}})
		}
	}
}

// Command user is the user's side of a conversation (next/16): a stateless
// continuation whose steps are the transcript. It is Codex's parent, so the
// user holds the higher authority: Codex can only reply to it, and only the
// user side can end the conversation.
//
// Its whole policy, one output per step:
//
//	the human just said something           → ask codex, in this conversation
//	the human said /close                   → reply "closed": the conversation ends
//	the channel to the human faulted        → reply with the fault: the conversation ends.
//	                                          Nobody is there to show it to, so it never retries
//	anything else (the start, Codex's
//	answer, or a fault in Codex's turn)     → show the transcript to the human, and wait for them
//
// A fault in a Codex turn is one step like any other: it is shown, and the
// conversation goes on. It shares no code with skyra.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type step struct {
	To    string          `json:"to"`
	State string          `json:"state"`
	Value json.RawMessage `json:"value"`
	Fault *struct {
		Kind  string `json:"kind"`
		Cause string `json:"cause"`
	} `json:"fault"`
}

type entry struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

func main() {
	var c struct {
		Input struct {
			Conversation string `json:"conversation"`
		} `json:"input"`
		Steps []step `json:"steps"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		fmt.Fprintln(os.Stderr, "user:", err)
		os.Exit(1)
	}

	transcript := []entry{}
	said := ""
	thread, incarnation := "", ""
	for _, s := range c.Steps {
		var text string
		_ = json.Unmarshal(s.Value, &text)
		if s.To == "codex" && s.Fault == nil {
			var answer struct{ Text, Thread, Incarnation string }
			if json.Unmarshal(s.Value, &answer) == nil && answer.Text != "" {
				text, thread, incarnation = answer.Text, answer.Thread, answer.Incarnation
			}
		}
		switch {
		case s.Fault != nil:
			if s.To == "codex" {
				thread, incarnation = "", ""
			}
			who := "the agent"
			if s.To == "browser" {
				who = "the browser"
			}
			transcript = append(transcript, entry{"agent", fmt.Sprintf("⚠ %s faulted (%s): %s", who, s.Fault.Kind, s.Fault.Cause)})
			said = ""
		case s.To == "browser":
			said = text
			if text != "/close" {
				transcript = append(transcript, entry{"user", text})
			}
		case s.To == "codex":
			transcript = append(transcript, entry{"agent", text})
			said = ""
		}
	}

	last := len(c.Steps) - 1
	switch {
	case last >= 0 && c.Steps[last].To == "browser" && c.Steps[last].Fault != nil:
		f := c.Steps[last].Fault
		fmt.Printf("ended: the browser faulted (%s): %s\n", f.Kind, f.Cause)
	case last >= 0 && c.Steps[last].To == "browser" && c.Steps[last].Fault == nil && said == "/close":
		fmt.Println("closed")
	case last >= 0 && c.Steps[last].To == "browser" && c.Steps[last].Fault == nil:
		emit("codex", map[string]string{"conversation": c.Input.Conversation, "text": said, "thread": thread, "incarnation": incarnation})
	default:
		emit("browser", map[string]any{"conversation": c.Input.Conversation, "transcript": transcript})
	}
}

func emit(to string, payload any) {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	fmt.Println(string(b))
}

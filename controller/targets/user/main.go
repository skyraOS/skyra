// Command user is the user's side of a conversation (next/16): a stateless
// continuation whose steps are the transcript. It is the parent of whatever
// it runs, so the user holds the higher authority: the agent can only reply to
// it, and only the user side can end the conversation.
//
// A conversation runs one procedure, fixed when it is created (proposal 38):
// the human selects it in the UI and the web host seeds it here. The user
// side never addresses an agent itself; what a procedure composes is the
// procedure's business.
//
// Its whole policy, one output per step:
//
//	the human just said something           → run the procedure on it, with the
//	                                          conversation so far as its context
//	the human said /close                   → return "closed": the conversation ends
//	the channel to the human faulted        → return the fault: the conversation ends.
//	                                          Nobody is there to show it to, so it never retries
//	anything else (the start, the
//	procedure's answer, or a fault in it)   → show the transcript to the human, and wait for them
//
// A fault in a turn is one step like any other: it is shown, and the
// conversation goes on. It shares no code with skyra.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	var raw json.RawMessage
	if err := json.NewDecoder(os.Stdin).Decode(&raw); err != nil {
		fmt.Fprintln(os.Stderr, "user:", err)
		os.Exit(1)
	}
	if handled, err := operation(raw); handled || err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, "user:", err)
			os.Exit(1)
		}
		return
	}
	var c struct {
		Frame string `json:"frame"`
		Input struct {
			Conversation string `json:"conversation"`
			Procedure    string `json:"procedure"`
			// A conversation continued in a new run, after the run it was in
			// ended (a host restart, a close, a fault): what was said before,
			// and the message that continues it. A conversation is the user's
			// and the agent's, not the kernel's: the new run is told the old
			// one as text.
			Transcript []entry                       `json:"transcript"`
			Text       string                        `json:"text"`
			Images     []struct{ Path, Name string } `json:"images"`
		} `json:"input"`
		Steps []step `json:"steps"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		fmt.Fprintln(os.Stderr, "user:", err)
		os.Exit(1)
	}
	if c.Input.Procedure == "" {
		c.Input.Procedure = "think" // the one procedure there is
	}
	p, ok := procedures[c.Input.Procedure]
	if !ok {
		fmt.Fprintf(os.Stderr, "user: no procedure %q\n", c.Input.Procedure)
		os.Exit(1)
	}

	transcript := append([]entry{}, c.Input.Transcript...)
	said := ""
	var images []string // the paths of the images pasted with what was said
	continuing := c.Input.Text != "" || len(c.Input.Images) > 0
	if continuing {
		said = c.Input.Text
		shown := said
		for _, im := range c.Input.Images {
			images = append(images, im.Path)
			shown += "\n🖼 " + im.Name
		}
		transcript = append(transcript, entry{"user", shown})
	}
	for _, s := range c.Steps {
		switch {
		case s.Fault != nil:
			who := "the agent"
			if s.To == "browser" {
				who = "the browser"
			}
			transcript = append(transcript, entry{"agent", fmt.Sprintf("⚠ %s faulted (%s): %s", who, s.Fault.Kind, s.Fault.Cause)})
			said = ""
		case s.To == "browser":
			var text string
			var pasted []struct{ Path, Name string }
			if json.Unmarshal(s.Value, &text) != nil {
				// A message with pasted images: its text, and each image's
				// path in scratch and hash, as the web host kept it.
				var m struct {
					Text   string                        `json:"text"`
					Images []struct{ Path, Name string } `json:"images"`
				}
				_ = json.Unmarshal(s.Value, &m)
				text, pasted = m.Text, m.Images
			}
			said, images = text, nil
			shown := text
			for _, im := range pasted {
				images = append(images, im.Path)
				shown += "\n🖼 " + im.Name
			}
			if text != "/close" {
				transcript = append(transcript, entry{"user", shown})
			}
		case s.To == p.Actor:
			transcript = append(transcript, entry{"agent", answer(s.Value)})
			said = ""
		}
	}

	last := len(c.Steps) - 1
	switch {
	case last < 0 && continuing:
		// The continuing message is run at once, with everything before it
		// as its context.
		if images == nil {
			images = []string{}
		}
		emit(p.Actor, map[string]any{c.Input.Procedure: map[string]any{"text": said, "context": transcript[:len(transcript)-1], "images": images}})
	case last >= 0 && c.Steps[last].To == "browser" && c.Steps[last].Fault != nil:
		f := c.Steps[last].Fault
		emit(c.Frame, fmt.Sprintf("ended: the browser faulted (%s): %s", f.Kind, f.Cause))
	case last >= 0 && c.Steps[last].To == "browser" && c.Steps[last].Fault == nil && said == "/close":
		emit(c.Frame, "closed")
	case last >= 0 && c.Steps[last].To == "browser" && c.Steps[last].Fault == nil:
		// The turn's context is the conversation before this message.
		context := transcript[:len(transcript)-1]
		if images == nil {
			images = []string{}
		}
		emit(p.Actor, map[string]any{c.Input.Procedure: map[string]any{"text": said, "context": context, "images": images}})
	default:
		emit("browser", map[string]any{"conversation": c.Input.Conversation, "transcript": transcript})
	}
}

// procedure is one entry of the table of what a human can run: the actor it
// runs on. A turn's payload is {<title>: {text, context, images}}: the human's
// message as text, the conversation so far as one context observation, and
// the paths of any images pasted with the message. Intake checks
// the fit, so a procedure that takes something else is refused, not misused.
type procedure struct{ Actor string }

// procedures is held here for now (proposal 38, step 5); later the UI's
// selection carries what a procedure runs on.
var procedures = map[string]procedure{
	"think": {Actor: "interpreter"},
}

// answer is what a turn returned, as the agent's line: its text when it has
// one, a refusal said as one, and its JSON otherwise.
func answer(v json.RawMessage) string {
	var a struct {
		Text    string `json:"text"`
		Refused string `json:"refused"`
	}
	if json.Unmarshal(v, &a) == nil {
		switch {
		case a.Refused != "" && strings.HasPrefix(a.Refused, "native."):
			return fmt.Sprintf("⚠ the agent attempted %s; the turn was interrupted", a.Refused)
		case a.Refused != "":
			return "⚠ " + a.Refused
		case a.Text != "":
			return a.Text
		}
	}
	var text string
	if json.Unmarshal(v, &text) == nil {
		return text
	}
	return string(v)
}

func emit(to string, payload any) {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	fmt.Println(string(b))
}

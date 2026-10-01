// Command agent is a fake continuation: deterministic, with no model. It
// stands where an LLM agent will: it reads a continuation on stdin (its input
// and every step it has taken, with what came back), and prints one output.
// It is stateless: everything it knows is on stdin. It shares no code with
// skyra.
//
// Its whole policy: with no steps yet, ask echo to repeat the text; once echo
// has answered, return what it said to the slot its frame discharges.
//
// With text "narrow:<text>", it first asks the room for echo under a narrowed
// Horizon (no reach, no grants), then asks echo through the reference it got
// back (next/26).
//
// With text "route:<text>", it writes a route at its novelty interrupt
// (next/30): echo's first return is novel, so it records a route that wakes
// it only when echo's answer changed, and sends an unchanged answer on
// instead. Then it asks echo the same thing again.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type continuation struct {
	Frame string `json:"frame"`
	Input struct {
		Text string `json:"text"`
	} `json:"input"`
	Steps []struct {
		To        string          `json:"to"`
		State     string          `json:"state"`
		Value     json.RawMessage `json:"value"`
		Interrupt string          `json:"interrupt"`
		Fault     *struct {
			Kind  string `json:"kind"`
			Cause string `json:"cause"`
		} `json:"fault"`
	} `json:"steps"`
}

// unchanged is the route the agent writes for echo: interrupt when the answer
// changed, otherwise send it on to echo, marked, so it is not delivered as is.
const unchanged = `def route():
    if skyra.changed("$"):
        return skyra.interrupt()
    return skyra.send("echo", {"text": "unchanged: " + skyra.get("$")})
`

func main() {
	var c continuation
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		fmt.Fprintln(os.Stderr, "agent:", err)
		os.Exit(1)
	}
	narrowed, isNarrow := strings.CutPrefix(c.Input.Text, "narrow:")
	routed, isRoute := strings.CutPrefix(c.Input.Text, "route:")
	if isRoute {
		switch {
		case len(c.Steps) == 0:
			emit("echo", map[string]string{"text": routed})
			return
		case len(c.Steps) == 1 && c.Steps[0].Interrupt == "novelty":
			emit("system.route", map[string]any{"step": 0, "source": unchanged})
			return
		case len(c.Steps) == 2:
			emit("echo", map[string]string{"text": routed})
			return
		}
	}
	if len(c.Steps) == 0 && isNarrow {
		emit("system.narrow", map[string]any{"actor": "echo", "reach": []string{}, "authority": map[string]any{}})
		return
	}
	if len(c.Steps) == 0 {
		emit("echo", map[string]string{"text": c.Input.Text})
		return
	}
	last := c.Steps[len(c.Steps)-1]
	if len(c.Steps) == 1 && isNarrow && last.Fault == nil {
		var got struct{ Reference string }
		_ = json.Unmarshal(last.Value, &got)
		emit(got.Reference, map[string]string{"text": narrowed})
		return
	}
	if last.Fault != nil {
		emit(c.Frame, fmt.Sprintf("echo failed (%s): %s", last.Fault.Kind, last.Fault.Cause))
		return
	}
	var said string
	_ = json.Unmarshal(last.Value, &said)
	emit(c.Frame, fmt.Sprintf("echo said %q", said))
}

func emit(to string, payload any) {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	fmt.Println(string(b))
}

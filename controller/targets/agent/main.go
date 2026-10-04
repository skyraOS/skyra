// Command agent is a fake continuation: deterministic, with no model. It
// stands where an LLM agent will: it reads a continuation on stdin (its input
// and every step it has taken, with what came back), and prints one output.
// It is stateless: everything it knows is on stdin. It shares no code with
// skyra.
//
// Its whole policy: with no steps yet, ask echo to repeat the text; once echo
// has answered, return what it said to the slot its frame discharges.
//
// With text "narrow:<text>", it first asks its Horizon (system, narrow) for echo under a narrowed
// Horizon (no reach, no grants), then asks echo through the reference it got
// back (next/26).
//
// With text "fan:<a>,<b>,...", it asks echo for every part in one emission
// (fan-out, next/32): a JSON array of instructions, one slot each. A part
// written "!<name>" asks <name> instead, which is out of reach: that one
// instruction faults, and its siblings still run. Once every slot is
// discharged, it returns what came back, in the order it asked.
//
// With text "extend:<text>", it adds a contract to itself (system,
// extend-reach: next/32), "shout", then invokes itself through it. Run
// against shout, it returns its input in capitals.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type continuation struct {
	Frame    string `json:"frame"`
	Contract *struct {
		Title string `json:"title"`
	} `json:"contract"` // the contract it is executed against, when it has several
	Input struct {
		Text  string `json:"text"`
		Shout string `json:"shout"`
	} `json:"input"`
	Steps []struct {
		To    string          `json:"to"`
		State string          `json:"state"`
		Value json.RawMessage `json:"value"`
		Fault *struct {
			Kind  string `json:"kind"`
			Cause string `json:"cause"`
		} `json:"fault"`
	} `json:"steps"`
}

func main() {
	var c continuation
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		fmt.Fprintln(os.Stderr, "agent:", err)
		os.Exit(1)
	}
	if c.Contract != nil && c.Contract.Title == "shout" {
		emit(c.Frame, strings.ToUpper(c.Input.Shout))
		return
	}
	if word, isExtend := strings.CutPrefix(c.Input.Text, "extend:"); isExtend {
		extend(c, word)
		return
	}
	narrowed, isNarrow := strings.CutPrefix(c.Input.Text, "narrow:")
	if parts, isFan := strings.CutPrefix(c.Input.Text, "fan:"); isFan {
		fan(c, strings.Split(parts, ","))
		return
	}
	if len(c.Steps) == 0 && isNarrow {
		emit("system", map[string]any{"narrow": map[string]any{"actor": "echo", "reach": []string{}, "authority": map[string]any{}}})
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

// extend adds the shout contract to itself, invokes itself through it, and
// returns what it said.
func extend(c continuation, word string) {
	switch len(c.Steps) {
	case 0:
		emit("system", map[string]any{"extend-reach": map[string]any{"schema": map[string]any{
			"title": "shout", "type": "object", "required": []string{"shout"},
			"fields": map[string]any{"shout": map[string]string{"type": "string"}}}}})
	case 1:
		if f := c.Steps[0].Fault; f != nil {
			emit(c.Frame, fmt.Sprintf("extend-reach failed (%s): %s", f.Kind, f.Cause))
			return
		}
		emit("agent", map[string]string{"shout": word})
	default:
		var said string
		_ = json.Unmarshal(c.Steps[1].Value, &said)
		emit(c.Frame, fmt.Sprintf("agent, through shout, said %q", said))
	}
}

// fan asks for every part at once, then answers with every step, in order.
func fan(c continuation, parts []string) {
	if len(c.Steps) == 0 {
		var all []any
		for _, p := range parts {
			if name, astray := strings.CutPrefix(p, "!"); astray {
				all = append(all, map[string]any{"to": name, "payload": map[string]string{"text": p}})
				continue
			}
			all = append(all, map[string]any{"to": "echo", "payload": map[string]string{"text": p}})
		}
		b, _ := json.Marshal(all)
		fmt.Println(string(b))
		return
	}
	var said []string
	for _, s := range c.Steps {
		if s.Fault != nil {
			var attempt struct{ To string } // a fault's value is the attempt it was derived from
			_ = json.Unmarshal(s.Value, &attempt)
			said = append(said, fmt.Sprintf("%s failed (%s)", attempt.To, s.Fault.Kind))
			continue
		}
		var v string
		_ = json.Unmarshal(s.Value, &v)
		said = append(said, fmt.Sprintf("%s said %q", s.To, v))
	}
	emit(c.Frame, strings.Join(said, "; "))
}

func emit(to string, payload any) {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	fmt.Println(string(b))
}

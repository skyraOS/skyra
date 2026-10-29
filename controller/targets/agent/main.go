// Command agent is a fake continuation: deterministic, with no model. It
// stands where an LLM agent will: it reads a continuation on stdin (its input
// and every step it has taken, with what came back), and prints one output.
// It is stateless: everything it knows is on stdin. It shares no code with
// skyra.
//
// Its whole policy: with no steps yet, ask echo to repeat the text; once echo
// has answered, return what it said to the slot its frame discharges.
//
// With text "narrow:<text>", it first asks its Boundary (system, narrow) for echo under a narrowed
// Boundary (no reach, no grants), then asks echo through the reference it got
// back (next/26).
//
// With text "fan:<a>,<b>,...", it asks echo for every part in one emission
// (fan-out, next/32): a JSON array of instructions, one slot each. A part
// written "!<name>" asks <name> instead, which is out of reach: that one
// instruction faults, and its siblings still run. Once every slot is
// discharged, it returns what came back, in the order it asked.
//
// With text "extend:<text>", it adds a contract to itself (system,
// export: next/32), "shout", then invokes itself through it. Run
// against shout, it returns its input in capitals.
//
// With text "relay:<text>", it does a task (next/32, H8): it asks echo for
// its whole input and for "relayed" in one emission, then echo again for
// what the first said, and returns everything that came back, authoring
// nothing.
//
// With text "remember:<text>", it is a worker that remembers (next/32, H6,
// H8): it asks agent to relay, remembers that trace through agent's
// contract, calls the procedure on new input (no agent runs), calls it again
// from a Boundary narrowed to keep it but drop echo, tries to remember the
// same contract a second time, calls the procedure on input outside the
// refinement the agent wrote (H9: the agent runs), and returns what each step
// gave back.
//
// With text "redundant:", it tries to extend itself with the contract it
// already holds: a contract fault at intake, which discharges its frame.
//
// Run against look (next/32, H7), it reads a path and returns what read
// observed. Run against resume, it is the original actor a procedure fell
// back to: it continues from the edge that refused, with what already ran,
// rather than starting over. With text "look:<dir>", it is a worker that
// remembers a look at <dir>/a.txt, looks at <dir>/b.txt through the
// procedure (the same base types: no agent runs), then at <dir>/sub, whose
// listing is not what the procedure was compiled for: the agent resumes.
//
// Run against refine (next/32, H9), it writes the refinements of a procedure
// remembered from it, in CEL: which inputs and returns its decisions still
// hold for. Its first answer reads a field that does not exist, so the
// interpreter refuses it and shows why; its second is right.
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
		Text   string `json:"text"`
		Shout  string `json:"shout"`
		Path   string `json:"path"`
		Refine *struct {
			Command struct {
				Text *string `json:"text"`
				Path *string `json:"path"`
			} `json:"command"`
			Rounds [][]struct {
				Node string `json:"node"`
			} `json:"rounds"`
			Refused string `json:"refused"`
		} `json:"refine"`
		Resume *struct {
			Ran []struct {
				Value json.RawMessage `json:"value"`
				Type  string          `json:"type"`
			} `json:"ran"`
			Refused struct {
				At  string `json:"at"`
				Why string `json:"why"`
			} `json:"refused"`
		} `json:"resume"`
	} `json:"input"`
	Reach map[string]struct {
		Contracts []json.RawMessage `json:"contracts"`
	} `json:"reach"`
	Steps []struct {
		To          string          `json:"to"`
		State       string          `json:"state"`
		Value       json.RawMessage `json:"value"`
		Observation string          `json:"observation"`
		Fault       *struct {
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
	if c.Contract != nil {
		switch c.Contract.Title {
		case "shout":
			emit(c.Frame, strings.ToUpper(c.Input.Shout))
			return
		case "look":
			if len(c.Steps) == 0 {
				emit("read", map[string]string{"path": c.Input.Path})
				return
			}
			emit(c.Frame, c.Steps[0].Value) // what read observed, passed up
			return
		case "refine":
			emit(c.Frame, refinements(c))
			return
		case "resume":
			r := c.Input.Resume
			var seen []json.RawMessage
			for _, s := range r.Ran {
				seen = append(seen, s.Value)
			}
			emit(c.Frame, map[string]any{"resumed at": r.Refused.At, "because": r.Refused.Why, "from what ran": seen})
			return
		}
	}
	if dir, isLook := strings.CutPrefix(c.Input.Text, "look:"); isLook {
		look(c, dir)
		return
	}
	if _, isRelay := strings.CutPrefix(c.Input.Text, "relay:"); isRelay {
		relay(c)
		return
	}
	if word, isRemember := strings.CutPrefix(c.Input.Text, "remember:"); isRemember {
		remember(c, word)
		return
	}
	if c.Input.Text == "redundant:" {
		emit("system", map[string]any{"export": map[string]any{"schema": registered(c)}})
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

// relay is a task whose result is made only of what came back: control, so
// remember can compile all of it.
func relay(c continuation) {
	echo := func(text any) map[string]any {
		return map[string]any{"to": "echo", "payload": map[string]any{"text": text}}
	}
	switch len(c.Steps) {
	case 0:
		b, _ := json.Marshal([]any{echo(c.Input.Text), echo("relayed")})
		fmt.Println(string(b))
	case 2:
		var first string
		_ = json.Unmarshal(c.Steps[0].Value, &first)
		b, _ := json.Marshal(echo(first))
		fmt.Println(string(b))
	default:
		emit(c.Frame, map[string]any{"echoes": []json.RawMessage{c.Steps[0].Value, c.Steps[1].Value}, "again": c.Steps[2].Value})
	}
}

// remember is the worker that remembers: it carries the observation's hash,
// which only a worker may (next/32).
func remember(c continuation, word string) {
	remembering := map[string]any{"remember": map[string]any{
		"observation": "", "contract": map[string]any{"target": "agent", "schema": registered(c)}}}
	switch len(c.Steps) {
	case 0:
		emit("agent", map[string]string{"text": "relay:" + word})
	case 1:
		remembering["remember"].(map[string]any)["observation"] = c.Steps[0].Observation
		emit("interpreter", remembering)
	case 2:
		emit("interpreter", map[string]string{"text": "relay:" + word + ", again"})
	case 3:
		emit("system", map[string]any{"narrow": map[string]any{"actor": "interpreter", "reach": []string{"interpreter"}}})
	case 4:
		var got struct{ Reference string }
		_ = json.Unmarshal(c.Steps[3].Value, &got)
		emit(got.Reference, map[string]string{"text": "relay:" + word + ", narrowed"})
	case 5:
		remembering["remember"].(map[string]any)["observation"] = c.Steps[0].Observation
		emit("interpreter", remembering)
	case 6:
		emit("interpreter", map[string]string{"text": "plain words"})
	default:
		said := map[string]any{}
		for i, name := range []string{"task", "remembered", "procedure", "narrow", "narrowed", "again", "outside its refinement"} {
			s := c.Steps[i]
			if s.Fault != nil {
				said[name] = fmt.Sprintf("fault (%s): %s", s.Fault.Kind, s.Fault.Cause)
				continue
			}
			said[name] = s.Value
		}
		emit(c.Frame, said)
	}
}

// refinements are the agent's refinements for a procedure remembered from
// it: a relay holds for any relay: input, and each call returning text; a
// look holds for any path, and each read returning a file.
func refinements(c continuation) map[string]any {
	r := c.Input.Refine
	header, edge := `input.text.startsWith("relay:")`, `type(returns.%s) == string`
	if r.Command.Path != nil {
		header, edge = `size(input.path) > 0`, `returns.%s.kind == "file"`
	}
	if r.Refused == "" {
		header = strings.Replace(strings.Replace(header, "input.text", "input.txt", 1), "input.path", "input.pth", 1) // a slip
	}
	edges := map[string]string{}
	for _, round := range r.Rounds {
		for _, call := range round {
			edges[call.Node] = fmt.Sprintf(edge, call.Node)
		}
	}
	return map[string]any{"header": header, "edges": edges}
}

// look is the worker that remembers a look and runs it on new paths.
func look(c continuation, dir string) {
	var lookContract json.RawMessage
	for _, s := range c.Reach["agent"].Contracts {
		var t struct{ Title string }
		if json.Unmarshal(s, &t) == nil && t.Title == "look" {
			lookContract = s
		}
	}
	switch len(c.Steps) {
	case 0:
		emit("agent", map[string]string{"path": dir + "/a.txt"})
	case 1:
		emit("interpreter", map[string]any{"remember": map[string]any{
			"observation": c.Steps[0].Observation, "contract": map[string]any{"target": "agent", "schema": lookContract}}})
	case 2:
		emit("interpreter", map[string]string{"path": dir + "/b.txt"})
	case 3:
		emit("interpreter", map[string]string{"path": dir + "/sub"})
	default:
		said := map[string]any{}
		for i, name := range []string{"a.txt, by the agent", "remembered", "b.txt, by the procedure", "sub, by the procedure"} {
			if f := c.Steps[i].Fault; f != nil {
				said[name] = fmt.Sprintf("fault (%s): %s", f.Kind, f.Cause)
				continue
			}
			said[name] = c.Steps[i].Value
		}
		emit(c.Frame, said)
	}
}

// extend adds the shout contract to itself, invokes itself through it, and
// returns what it said.
func extend(c continuation, word string) {
	switch len(c.Steps) {
	case 0:
		emit("system", map[string]any{"export": map[string]any{"schema": map[string]any{
			"title": "shout", "type": "object", "required": []string{"shout"},
			"fields": map[string]any{"shout": map[string]string{"type": "string"}}}}})
	case 1:
		if f := c.Steps[0].Fault; f != nil {
			emit(c.Frame, fmt.Sprintf("export failed (%s): %s", f.Kind, f.Cause))
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

// registered is the agent's own registered contract, as its reach shows it.
func registered(c continuation) json.RawMessage {
	if all := c.Reach["agent"].Contracts; len(all) > 0 {
		return all[0]
	}
	return nil
}

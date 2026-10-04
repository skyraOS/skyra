// Package remember compiles a strand into procedures (next/32, "Compiling a
// strand"). In the strand a model did two jobs: it decided what to call, and
// it was resumed when its calls came back. Compiling gives both jobs to the
// interpreter; the frames, rounds and discharges look the same to the kernel,
// only the program resumed changes.
//
//	strand(frame) = frame record
//	              → rounds (resolve records, linked through their round)
//	                → each round: the invokes it opened, in the order they were emitted
//	                  → each invoke: its discharge, and
//	                      no rounds of its own   → a leaf: a call
//	                      rounds of its own      → a continuation: compiled, if its return passes up what ran
//	                      another Horizon        → opaque: one call and its return
//
// Control, not content: a continuation's emissions become nodes; a return
// is control when every value in it was found by provenance, and a frame
// whose return the model authored stays a call to it.
package remember

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"skyra/interp/proc"
)

// Record is a kernel record as the strand hands it over.
type Record struct {
	Predecessor string          `json:"predecessor"`
	State       string          `json:"state"`
	To          string          `json:"to,omitempty"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	Fault       *proc.Fault     `json:"fault,omitempty"`
	Frame       string          `json:"frame,omitempty"`
	Round       string          `json:"round,omitempty"`
	Deps        []string        `json:"deps,omitempty"`
	Supervisor  bool            `json:"supervisor,omitempty"`
}

// Strand is what the kernel's strand syscall returns.
type Strand struct {
	Chain   []string          `json:"chain"`
	Records map[string]Record `json:"records"`
	Opaque  map[string]string `json:"opaque,omitempty"`
}

// Reach is every contract each name in the requester's Horizon holds, as its
// envelope shows them.
type Reach map[string][]*proc.Schema

// Refused is a strand remember will not compile: shown to whoever asked.
type Refused string

func (r Refused) Error() string { return string(r) }

func refuse(format string, args ...any) error { return Refused(fmt.Sprintf(format, args...)) }

// Compile turns the strand behind an observation into procedures, bottom
// up: every procedure before the procedure that calls it, the one for the
// requested contract last. self is the interpreter's own name in reach, the
// target a compiled child is called through.
func Compile(st Strand, observation, target string, schema *proc.Schema, reach Reach, self string) ([]proc.Source, error) {
	if len(st.Chain) == 0 {
		return nil, refuse("the strand is empty")
	}
	c := compiler{st: st, reach: reach, self: self, resolves: map[string]string{}}
	for stamp, r := range st.Records {
		if r.State == "resolve" {
			c.resolves[r.Round] = stamp
		}
	}
	cause := st.Chain[len(st.Chain)-1]
	inv, ok := st.Records[cause]
	if !ok || inv.State != "invoke" || inv.To != target {
		return nil, refuse("the strand does not end at an invoke of %s", target)
	}
	if !c.continuation(cause) {
		return nil, refuse("%s ran as a leaf: it called nothing, so there is nothing to compile", target)
	}
	o, ok := st.Records[observation]
	if !ok || o.State != "return" || o.Supervisor || !slices.Contains(c.rounds(cause), o.Predecessor) {
		return nil, refuse("remember what %s returned: this observation is not its return", target)
	}
	srcs, control, err := c.frame(cause, schema, target, o)
	if err != nil {
		return nil, err
	}
	if !control {
		return nil, refuse("%s authored its result rather than passing up what ran: nothing to compile", target)
	}
	return srcs, nil
}

type compiler struct {
	st       Strand
	reach    Reach
	self     string
	resolves map[string]string // round → the resolve that closed it
}

// rounds are a frame's rounds: the frame itself, then each resolve.
func (c *compiler) rounds(frame string) []string {
	rounds := []string{frame}
	for r := frame; ; {
		res, ok := c.resolves[r]
		if !ok || c.st.Records[res].Frame != frame {
			return rounds
		}
		rounds = append(rounds, res)
		r = res
	}
}

// continuation reports whether a frame opened slots: derived from the
// records, never claimed.
func (c *compiler) continuation(frame string) bool {
	if _, in := c.st.Records[frame]; !in {
		return false
	}
	res, ok := c.resolves[frame]
	return ok && c.st.Records[res].Frame == frame
}

// child is the record that opened the slot a discharge discharged: on the
// discharge's chain, the one emitted under the round.
func (c *compiler) child(discharge, round string) (string, error) {
	if call, ok := c.st.Opaque[discharge]; ok {
		return call, nil
	}
	for s := discharge; ; {
		r, ok := c.st.Records[s]
		if !ok {
			return "", fmt.Errorf("the strand lacks %s", short(s))
		}
		if r.Predecessor == round {
			return s, nil
		}
		s = r.Predecessor
	}
}

type known struct {
	name  string
	value any
}

// frame compiles one continuation's control into a procedure source for the
// contract it was invoked through, and every child procedure before it. It
// reports whether its return is control; when it is not, the frame stays a
// call to its original actor.
func (c *compiler) frame(frame string, schema *proc.Schema, original string, discharge Record) ([]proc.Source, bool, error) {
	inv := c.st.Records[frame]
	input, err := decode(inv.Payload)
	if err != nil {
		return nil, false, err
	}
	src := proc.Source{Contract: schema, Original: original, Nodes: map[string]proc.SourceNode{},
		Refinement: &proc.Refinement{Input: proc.InputType(schema, input), Edges: map[string]string{}}}
	var below []proc.Source
	var seen []known
	var prev []string
	for k, round := range c.rounds(frame) {
		res, ok := c.resolves[round]
		if !ok || c.st.Records[res].Frame != frame {
			break // the last round: the frame returned from it
		}
		var this []string
		for i, d := range c.st.Records[res].Deps {
			stamp, err := c.child(d, round)
			if err != nil {
				return nil, false, err
			}
			x, dis := c.st.Records[stamp], c.st.Records[d]
			switch {
			case x.State != "invoke":
				return nil, false, refuse("an instruction in the strand faulted at intake: a strand with a fault is not remembered")
			case dis.State != "return":
				return nil, false, refuse("a call to %s in the strand did not return: a strand with a fault is not remembered", x.To)
			case x.To == "system" && syscall(x.Payload) != "narrow":
				return nil, false, refuse("the strand calls system's %s: only a narrowing is remembered", syscall(x.Payload))
			}
			name := fmt.Sprintf("r%dc%d", k, i)
			call, _ := json.Marshal(x.To)
			if strings.HasPrefix(x.To, "system.ref.") { // a reference: bound to the narrowing that handed it back
				call, _ = c.template(x.To, input, seen, true)
			}
			in, _ := c.template(mustDecode(x.Payload), input, seen, false)
			if _, opaque := c.st.Opaque[d]; !opaque && c.continuation(stamp) {
				if cs := proc.Selected(c.reach[x.To], x.Payload); cs != nil {
					sub, control, err := c.frame(stamp, cs, x.To, dis)
					if err != nil {
						return nil, false, err
					}
					if control { // a continuation whose control is compiled: called through the interpreter
						below = append(below, sub...)
						call, _ = json.Marshal(c.self)
					}
				}
			}
			value, err := decode(dis.Payload)
			if err != nil {
				return nil, false, err
			}
			src.Nodes[name] = proc.SourceNode{Call: call, In: in, After: prev}
			src.Order = append(src.Order, name)
			src.Refinement.Edges[name] = proc.EdgeType(call, value)
			seen = append(seen, known{name, value})
			this = append(this, name)
		}
		prev = this
	}
	ret, control := c.template(mustDecode(discharge.Payload), input, seen, false)
	src.Return = ret
	return append(below, src), control, nil
}

// template is a value as a template: wherever a part of it equals the input,
// or something an earlier node returned, a binding to it (a parameter, or a
// binding); the rest constants. It reports whether every leaf was bound.
// Provenance is by equality: the input first, then the nodes in the order
// they ran. A null, a bool or an empty string is too common to say where it
// came from, so it stays a constant. A call's target may be bound only to
// what a node returned: a name in the input is the caller's to choose.
func (c *compiler) template(v any, input any, seen []known, isCall bool) (json.RawMessage, bool) {
	t, bound := c.bind(v, input, seen, isCall)
	b, _ := json.Marshal(t)
	return b, bound
}

func (c *compiler) bind(v any, input any, seen []known, isCall bool) (any, bool) {
	if path, ok := c.find(v, input, seen, isCall); ok {
		return path, true
	}
	switch x := v.(type) {
	case map[string]any:
		out, all := map[string]any{}, true
		for k, e := range x {
			t, b := c.bind(e, input, seen, false)
			out[k], all = t, all && b
		}
		return proc.Const(out), all
	case []any:
		out, all := make([]any, len(x)), true
		for i, e := range x {
			t, b := c.bind(e, input, seen, false)
			out[i], all = t, all && b
		}
		return out, all
	}
	return v, false
}

// find is where a value came from, as a binding: the input or a part of it,
// else an earlier node's return or a part of it.
func (c *compiler) find(v any, input any, seen []known, isCall bool) (map[string]any, bool) {
	switch x := v.(type) {
	case nil, bool:
		return nil, false
	case string:
		if x == "" {
			return nil, false
		}
	}
	want := canonical(v)
	if !isCall {
		if path, ok := search(input, want); ok {
			return proc.Binding("input", path...), true
		}
	}
	for _, k := range seen {
		if path, ok := search(k.value, want); ok {
			return proc.Binding(k.name, path...), true
		}
	}
	return nil, false
}

// search finds a part of v equal to want, shallowest first, keys in order.
func search(v any, want string) ([]any, bool) {
	type at struct {
		path []any
		v    any
	}
	queue := []at{{nil, v}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if canonical(cur.v) == want {
			return cur.path, true
		}
		switch x := cur.v.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			for _, k := range keys {
				queue = append(queue, at{append(slices.Clone(cur.path), k), x[k]})
			}
		case []any:
			for i, e := range x {
				queue = append(queue, at{append(slices.Clone(cur.path), i), e})
			}
		}
	}
	return nil, false
}

func syscall(payload json.RawMessage) string {
	var o map[string]json.RawMessage
	if json.Unmarshal(payload, &o) != nil || len(o) != 1 {
		return "?"
	}
	for k := range o {
		return k
	}
	return "?"
}

func decode(b []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var v any
	err := d.Decode(&v)
	return v, err
}

func mustDecode(b []byte) any {
	v, _ := decode(b)
	return v
}

func canonical(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

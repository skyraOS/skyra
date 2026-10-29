package proc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// Running a procedure (32, "Running a procedure"). On each resume:
//
//	read     the procedure and where this frame is; the frame's steps
//	advance  which nodes have discharged, and what they bound
//	emit     every call now ready, as one emission of N instructions,
//	         or the return value to its slot
//
// Plan is the advance, and the choice of what to emit. It is pure: the same
// procedure, input, position and steps give the same plan, so a resumed
// apprentice, or one that lost its process, decides the same thing again.

// Step is one slot as the kernel delivers it in the envelope.
type Step struct {
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
	State   string          `json:"state"`
	Value   json.RawMessage `json:"value,omitempty"`
	Fault   *Fault          `json:"fault,omitempty"`
}

// Fault is a step's fault, as the kernel derived it.
type Fault struct {
	Kind  string `json:"kind"`
	Cause string `json:"cause"`
}

// Next is what to do: emit a round, return, or stop at a call that did not
// return.
type Next struct {
	Emit   []json.RawMessage // the round's instructions, in order
	Round  []Call
	Return json.RawMessage // every node discharged: the return value
	Stuck  *Stuck          // a call faulted: the procedure cannot go on as compiled
	Values map[string]any  // what each discharged node returned
}

// Stuck is the first call that did not return: its node, and its step.
type Stuck struct {
	Node  string
	Index int
	Step  Step
}

// Plan advances a procedure over the steps its frame has taken.
func Plan(p Procedure, nodes map[string]Node, input json.RawMessage, pos Position, steps []Step) (Next, error) {
	in, err := decode(input)
	if err != nil {
		return Next{}, fmt.Errorf("input: %w", err)
	}
	var calls []Call
	for _, r := range pos.Rounds {
		calls = append(calls, r...)
	}
	if len(calls) != len(steps) {
		return Next{}, fmt.Errorf("the frame emitted %d calls but the envelope delivers %d steps", len(calls), len(steps))
	}
	order, err := topo(p, nodes)
	if err != nil {
		return Next{}, err
	}

	values := map[string]any{}
	emitted := map[string]int{} // node → how many calls it emitted
	parts := map[string][]any{}
	for i, c := range calls {
		emitted[c.Node]++
		st := steps[i]
		if st.State != "return" {
			return Next{Stuck: &Stuck{Node: c.Node, Index: c.Index, Step: st}, Values: values}, nil
		}
		v, err := decode(st.Value)
		if err != nil {
			return Next{}, fmt.Errorf("step %d: %w", i, err)
		}
		if c.Index < 0 {
			values[c.Node] = v
			continue
		}
		for len(parts[c.Node]) <= c.Index {
			parts[c.Node] = append(parts[c.Node], nil)
		}
		parts[c.Node][c.Index] = v
	}
	for id, ps := range parts {
		values[id] = ps
	}

	env := Env{Input: in, Values: values}
	var next Next
	for _, id := range order {
		if _, done := values[id]; done || emitted[id] > 0 {
			continue
		}
		n := nodes[id]
		if !ready(n, values) {
			continue
		}
		target, err := callee(n, env)
		if err != nil {
			return Next{}, fmt.Errorf("node %s: %w", short(id), err)
		}
		if len(n.Each) == 0 {
			payload, err := Eval(n.In, env)
			if err != nil {
				return Next{}, fmt.Errorf("node %s: %w", short(id), err)
			}
			next.Emit = append(next.Emit, instruction(target, payload))
			next.Round = append(next.Round, Call{Node: id, Index: -1})
			continue
		}
		list, err := Eval(n.Each, env)
		items, ok := list.([]any)
		if err != nil || !ok {
			return Next{}, fmt.Errorf("node %s: each is not a list: %v", short(id), err)
		}
		if len(items) == 0 {
			values[id] = []any{} // nothing to call: discharged at once
			continue
		}
		for i, item := range items {
			payload, err := Eval(n.In, Env{Input: in, Values: values, Item: item, HasItem: true})
			if err != nil {
				return Next{}, fmt.Errorf("node %s[%d]: %w", short(id), i, err)
			}
			next.Emit = append(next.Emit, instruction(target, payload))
			next.Round = append(next.Round, Call{Node: id, Index: i})
		}
	}
	next.Values = values
	if len(next.Emit) > 0 {
		alone(&next)
		distinct(next.Emit)
		return next, nil
	}
	for _, id := range order {
		if _, done := values[id]; !done {
			return Next{}, fmt.Errorf("node %s can never run", short(id))
		}
	}
	ret, err := Eval(p.Return, env)
	if err != nil {
		return Next{}, fmt.Errorf("return: %w", err)
	}
	next.Return, _ = json.Marshal(ret)
	return next, nil
}

// Emission is a round's instructions as one emission: one instruction alone,
// several as a JSON array (fan-out).
func (n Next) Emission() []byte {
	if len(n.Emit) == 1 {
		return n.Emit[0]
	}
	// Joined as written: json.Marshal would compact each instruction, and
	// with it the spaces that keep repeats distinct.
	parts := make([][]byte, len(n.Emit))
	for i, e := range n.Emit {
		parts[i] = e
	}
	return append(append([]byte("["), bytes.Join(parts, []byte(","))...), ']')
}

func ready(n Node, values map[string]any) bool {
	for _, a := range n.After {
		if _, done := values[a]; !done {
			return false
		}
	}
	return true
}

// callee is a node's target: a name, or a binding to one an earlier node
// returned (decision 26), such as a reference narrow handed back.
func callee(n Node, env Env) (string, error) {
	v, err := Eval(n.Call, env)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", errors.New("call is not a name")
	}
	return s, nil
}

func instruction(to string, payload any) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	return b
}

// distinct makes every instruction of a round its own bytes. Under one
// predecessor the kernel content-addresses what it receives, so two identical
// instructions would be one record, and it refuses them; every call runs
// (decision 13), so the k-th repeat carries k spaces inside its object,
// which no reading of it notices.
func distinct(emit []json.RawMessage) {
	for i := range emit {
		k := 0
		for _, earlier := range emit[:i] {
			if bytes.Equal(bytes.TrimRight(earlier[:len(earlier)-1], " "), emit[i][:len(emit[i])-1]) {
				k++
			}
		}
		if k > 0 {
			emit[i] = append(append(slices.Clone(emit[i][:len(emit[i])-1]), bytes.Repeat([]byte(" "), k)...), '}')
		}
	}
}

// alone keeps a syscall that changes the Boundary of whoever waits (narrow,
// export) out of a fan-out: the kernel refuses one emitted with
// siblings (H2), so it goes first, by itself, and the rest wait a round.
func alone(n *Next) {
	if len(n.Emit) < 2 {
		return
	}
	for i, e := range n.Emit {
		var ins struct {
			To      string                     `json:"to"`
			Payload map[string]json.RawMessage `json:"payload"`
		}
		if json.Unmarshal(e, &ins) != nil || ins.To != "system" {
			continue
		}
		_, narrow := ins.Payload["narrow"]
		_, extend := ins.Payload["export"]
		if !narrow && !extend {
			continue
		}
		n.Emit, n.Round = []json.RawMessage{e}, []Call{n.Round[i]}
		return
	}
}

// topo is the order nodes are considered in: every node after the nodes it
// depends on, from the body's sinks in order.
func topo(p Procedure, nodes map[string]Node) ([]string, error) {
	var order []string
	state := map[string]int{} // 1 visiting, 2 done
	var visit func(id string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("node %s depends on itself", short(id))
		case 2:
			return nil
		}
		n, ok := nodes[id]
		if !ok {
			return fmt.Errorf("no node %s", short(id))
		}
		state[id] = 1
		for _, a := range n.After {
			if err := visit(a); err != nil {
				return err
			}
		}
		state[id] = 2
		order = append(order, id)
		return nil
	}
	for _, id := range p.Body {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return order, nil
}

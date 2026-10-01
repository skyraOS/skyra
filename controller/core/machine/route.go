package machine

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Routes (next/30). The runtime computes a return's type; the parent's route
// for that type decides where the return goes:
//
//	child returns ─▶ its type ─▶ the route for (parent, binding, type): recorded, else declared
//	                             ├─ found: run it ─▶ interrupt | send | fault
//	                             └─ none:  a novelty interrupt
//
// An interrupt is raised by the runtime, never by the parent (next/29), and has
// a kind: novelty (no route for this type), route (the route asked) or depth
// (a chain of sends reached MaxChain). A send is an invoke under the parent's
// Horizon, whose return is routed again. A fault delivers the return to the
// parent as refused.
//
// Only a continuation's single-child rounds are routed in v1, and only
// returns: a fault, a fan-out round and a facet of the room wake the parent as
// before. Everything here is pure; the route runs in the route runner.

// Decisions and interrupt kinds.
const (
	Interrupt = "interrupt"
	Send      = "send"
	Refuse    = "fault"

	Novelty  = "novelty"
	ByRoute  = "route"
	Depth    = "depth"
	MaxChain = 8 // sends in a row, at one waiter, before its next return interrupts
)

// Typed is a return's type as the parent's routes know it (next/30): the
// binding it came back through, its type hash, and its presentation. It is
// derived on demand from what a round already holds, never stored in a record.
type Typed struct {
	Binding      string // the name the parent invoked: a name in reach, or the actor a reference stands for
	Type         string // contract.Type(actor identity, contract hash, shape): never shown to a model
	Presentation string // what a model is shown: binding → shape
}

// bound resolves what an invoke addressed under the parent's Horizon: the
// binding its routes are keyed on, and the surface that ran.
func bound(parent horizon.Horizon, to string) (string, horizon.Surface, bool) {
	if s, ok := parent.Reach[to]; ok {
		return to, s, true
	}
	if g, ok := parent.Refs[to]; ok {
		s, ok := parent.Reach[g.Actor]
		return g.Actor, s, ok
	}
	return "", horizon.Surface{}, false
}

// TypeOf derives the type of a child's discharge, if it returned. The actor is
// the identity of the surface that ran; the contract is the one it was
// invoked through.
func TypeOf(parent horizon.Horizon, c Child) (Typed, bool) {
	if c.Discharge == nil || c.Discharge.State != record.Return {
		return Typed{}, false
	}
	name, s, ok := bound(parent, c.Invoke.To)
	if !ok {
		return Typed{}, false
	}
	v, err := contract.Decode(c.Discharge.Payload)
	if err != nil {
		return Typed{}, false
	}
	return Typed{
		Binding:      name,
		Type:         contract.Type(horizon.Identity(s), horizon.ContractHash(s.Input), contract.Shape(v)),
		Presentation: name + " → " + contract.Present(v),
	}, true
}

// RouteKey is what a route is recorded under: the parent's actor identity, so
// a route outlives the frame, the binding, and the type.
type RouteKey struct {
	Parent  string `json:"parent"`
	Binding string `json:"binding"`
	Type    string `json:"type"`
}

// ID is the key's name in the tree.
func (k RouteKey) ID() string {
	b, _ := json.Marshal(k)
	return record.Sum(b)
}

func keyOf(parent horizon.Horizon, run RouteRun) RouteKey {
	return RouteKey{Parent: horizon.Identity(*parent.Surface), Binding: run.Binding, Type: run.Type}
}

// RouteRun is how one discharge was routed, as the resolve that follows it
// observed. Hashes are fine here: it is the runtime's, never a model's.
type RouteRun struct {
	Child     string            `json:"child"`               // the discharge routed
	Binding   string            `json:"binding"`             // the name it came back through
	Type      string            `json:"type"`                // its type
	Source    string            `json:"source,omitempty"`    // the route that ran, by hash; none for novelty and depth
	Input     string            `json:"input,omitempty"`     // the runner's stdin, by hash
	Output    string            `json:"output,omitempty"`    // the runner's emission, by hash
	Failed    string            `json:"failed,omitempty"`    // or why it emitted none, by hash
	Decision  string            `json:"decision"`            // interrupt, send or fault
	Interrupt string            `json:"interrupt,omitempty"` // an interrupt's kind: novelty, route or depth
	Send      json.RawMessage   `json:"send,omitempty"`      // a send's envelope, emitted under the parent's Horizon
	Fault     *record.FaultInfo `json:"fault,omitempty"`     // a fault's kind and cause
}

// RouteEntry is one route recorded under a key (F7, F8): its source by hash,
// the parent that wrote it, and the flips it accepted against the log.
type RouteEntry struct {
	Source   string `json:"source"`
	Author   string `json:"author"` // the parent's actor identity
	Frame    string `json:"frame"`  // the frame it was written in
	Accepted []Flip `json:"accepted,omitempty"`
}

// Flip is a recorded route run whose decision a rewritten route changes.
type Flip struct {
	Input string          `json:"input"` // the recorded run's runner input, by hash
	Was   json.RawMessage `json:"was"`   // the recorded decision
	Now   json.RawMessage `json:"now"`   // the rewritten route's
}

// RunnerInput is the route runner's stdin.
type RunnerInput struct {
	Source   string          `json:"source"`
	Return   json.RawMessage `json:"return"`
	Previous json.RawMessage `json:"previous"`
	Targets  []string        `json:"targets"`
}

// routable says whether the gathered round is routed: nil if not, a run with a
// decision if it is decided without a route (the chain is at its depth), and a
// run without one if its route is to be looked up.
func routable(waiter horizon.Horizon, e Settled) *RouteRun {
	round := e.Rounds[len(e.Rounds)-1]
	if waiter.Surface == nil || len(round.Children) != 1 {
		return nil
	}
	c := round.Children[0]
	t, ok := TypeOf(waiter, c)
	if !ok || strings.HasPrefix(t.Binding, SystemPrefix) {
		return nil
	}
	run := &RouteRun{Child: c.Result, Binding: t.Binding, Type: t.Type}
	if chain(e.Rounds) >= MaxChain {
		run.Decision, run.Interrupt = Interrupt, Depth
	}
	return run
}

// chain counts the sends in a row that opened the latest rounds.
func chain(rounds []Round) int {
	n := 0
	for i := len(rounds) - 1; i > 0 && sentBy(rounds[i]); i-- {
		n++
	}
	return n
}

// sentBy reports whether a round was opened by a route's send.
func sentBy(r Round) bool {
	return r.Opened != nil && r.Opened.Route != nil && r.Opened.Route.Decision == Send
}

// lookedUp: a recorded route wins over a declared one, and with neither the
// return is novel. A route runs in the route runner, under an empty Horizon.
func lookedUp(s Instance, e RouteLooked) (Instance, []Effect) {
	run, waiter := *s.Routing, *s.Self.Resume
	source := e.Recorded
	if source == nil {
		source = e.Declared
	}
	if source == nil {
		run.Decision, run.Interrupt = Interrupt, Novelty
		return resolve(s, &run)
	}
	run.Source = record.Sum(source)
	runner, ok := waiter.Reach[RouteRunner]
	if !ok {
		run.Decision, run.Fault = Refuse, &record.FaultInfo{Kind: ByRoute, Cause: "route: no route runner in the room"}
		return resolve(s, &run)
	}
	rounds := s.Gathered.Rounds
	stdin, _ := json.Marshal(RunnerInput{Source: string(source), Return: rounds[len(rounds)-1].Children[0].Discharge.Payload,
		Previous: previous(waiter, rounds, run), Targets: targets(waiter)})
	run.Input = record.Sum(stdin)
	s.Routing, s.Phase = &run, Route
	h := horizon.Horizon{Stamp: s.Stamp, Surface: &runner, Reach: map[string]horizon.Surface{}, Frame: run.Child}
	return s, []Effect{StoreInput{stdin}, runOf(h, stdin, s.Floor)}
}

// routed: the runner decided, or failed. Either way the round resolves, with
// the decision in its observation, and the run is indexed under its key.
func routed(s Instance, e Emitted) (Instance, []Effect) {
	run, waiter := *s.Routing, *s.Self.Resume
	var effs []Effect
	if e.Err == "" {
		run.Output = record.Sum(e.Out)
		effs = append(effs, StoreInput{e.Out}, RecordRun{Key: keyOf(waiter, run), Input: run.Input, Output: run.Output})
	} else {
		run.Failed = record.Sum([]byte(e.Err))
		effs = append(effs, StoreInput{[]byte(e.Err)})
	}
	run = Decided(waiter, run, e.Out, e.Err)
	return resolve(s, &run, effs...)
}

// Decided derives a route run's decision from the runner's emission, or from
// why it emitted none. It is pure, so replay derives it again. The emission
// must address the runner's slot (the discharge) with exactly one decision. A
// send must be an invoke under the parent's Horizon: anything else is
// delivered to the parent as the fault it would be.
func Decided(parent horizon.Horizon, run RouteRun, out []byte, failed string) RouteRun {
	run.Decision, run.Interrupt, run.Send, run.Fault = Refuse, "", nil, nil
	refuse := func(kind, cause string) RouteRun {
		run.Fault = &record.FaultInfo{Kind: kind, Cause: cause}
		return run
	}
	if failed != "" {
		return refuse(ByRoute, "route runner: "+failed)
	}
	var emitted struct {
		To      string `json:"to"`
		Payload struct {
			Interrupt *bool `json:"interrupt"`
			Send      *struct {
				To      string          `json:"to"`
				Payload json.RawMessage `json:"payload"`
			} `json:"send"`
			Fault *string `json:"fault"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(out, &emitted); err != nil || emitted.To != run.Child {
		return refuse(ByRoute, "route runner: its emission is not a decision for this return")
	}
	d := emitted.Payload
	switch {
	case d.Interrupt != nil && *d.Interrupt && d.Send == nil && d.Fault == nil:
		run.Decision, run.Interrupt = Interrupt, ByRoute
		return run
	case d.Fault != nil && *d.Fault != "" && d.Interrupt == nil && d.Send == nil:
		return refuse(ByRoute, *d.Fault)
	case d.Send == nil || d.Interrupt != nil || d.Fault != nil || len(d.Send.Payload) == 0:
		return refuse(ByRoute, "route runner: a decision is exactly one of interrupt, send or fault")
	}
	payload, err := contract.Decode(d.Send.Payload)
	if err != nil {
		return refuse(record.Malformed, "route: send: "+err.Error())
	}
	if _, s, ok := bound(parent, d.Send.To); ok && s.System != "" {
		return refuse(record.Horizon, fmt.Sprintf("route: send: %q is the room's, not an actor", d.Send.To))
	}
	envelope := contract.Canonical(map[string]any{"to": d.Send.To, "payload": payload})
	switch r := interpret(parent, envelope); r.State {
	case record.Invoke:
		run.Decision, run.Send = Send, envelope
		return run
	case record.Fault:
		return refuse(r.Fault.Kind, "route: send: "+r.Fault.Cause)
	}
	return refuse(record.Horizon, "route: send: a route cannot return for its parent")
}

// previous is the last earlier return in the frame with the same binding and
// type, or null.
func previous(waiter horizon.Horizon, rounds []Round, run RouteRun) json.RawMessage {
	var prev json.RawMessage
	for _, round := range rounds[:len(rounds)-1] {
		for _, c := range round.Children {
			if t, ok := TypeOf(waiter, c); ok && t.Binding == run.Binding && t.Type == run.Type {
				prev = c.Discharge.Payload
			}
		}
	}
	return prev
}

// targets are the names a route may send to: the actors in the parent's
// reach, and the references it holds. Never a facet of the room.
func targets(parent horizon.Horizon) []string {
	var names []string
	for name, s := range parent.Reach {
		if s.System == "" {
			names = append(names, name)
		}
	}
	for ref, g := range parent.Refs {
		if _, ok := parent.Reach[g.Actor]; ok {
			names = append(names, ref)
		}
	}
	slices.Sort(names)
	return names
}

// viewed shows a continuation how its route took a step: by names and
// presentations, never hashes.
func viewed(v *StepView, waiter horizon.Horizon, c Child, run RouteRun) {
	switch run.Decision {
	case Interrupt:
		v.Interrupt = run.Interrupt
		if t, ok := TypeOf(waiter, c); ok {
			v.Type = t.Presentation
		}
	case Send:
		var env struct{ To string }
		_ = json.Unmarshal(run.Send, &env)
		v.Sent = env.To
	case Refuse:
		v.State, v.Fault = record.Fault, run.Fault
	}
}

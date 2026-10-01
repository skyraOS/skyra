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

// The novelty interrupt and system.route (next/30, F7), and the log as the
// test suite (F8).
//
// A return with no route is novel: the parent wakes, the step says so, and
// the continuation carries the route contract: what a route may test, what it
// may decide, the parent's Horizon and what can happen to it. All of it as
// names and contracts: no hash reaches a model.
//
// The parent answers with system.route, naming the step it was shown. The
// runtime binds the source to that step's (binding, type), under the parent's
// identity, so the route outlives the frame. Before it is recorded, the new
// source runs over every recorded run of that key; a changed decision is a
// flip, and the route is refused unless the parent accepts every flip.

// RouteContract is what a model is shown at a novelty interrupt. The return
// itself is the novel step's type.
type RouteContract struct {
	Conditions map[string]string `json:"conditions"`
	Decisions  map[string]string `json:"decisions"`
	Horizon    ShownHorizon      `json:"horizon"`
	Faults     []string          `json:"faults"`
	Interrupts []string          `json:"interrupts"`
	Write      string            `json:"write"`
}

// ShownHorizon is a Horizon with every hash left out: names in reach with
// their input contracts, references by name, and the authority.
type ShownHorizon struct {
	Reach     map[string]*contract.Schema `json:"reach"`
	Refs      []string                    `json:"refs"`
	Authority horizon.Authority           `json:"authority"`
}

func routeContract(h horizon.Horizon) RouteContract {
	shown := ShownHorizon{Reach: map[string]*contract.Schema{}, Refs: []string{}, Authority: h.Authority}
	for name, s := range h.Reach {
		if s.System != "runner" {
			shown.Reach[name] = s.Input
		}
	}
	for ref := range h.Refs {
		shown.Refs = append(shown.Refs, ref)
	}
	slices.Sort(shown.Refs)
	return RouteContract{
		Conditions: map[string]string{
			"skyra.get(path)":     "the value at a path of the return ($, $.key, $[\"key\"], $[0]), of that path's kind",
			"skyra.previous()":    "the previous return of this type in this frame, or None",
			"skyra.changed(path)": "bool: whether the value at path differs from the previous return's",
			"skyra.diff()":        "float: the ratio of the return's leaves that changed since the previous return",
		},
		Decisions: map[string]string{
			"skyra.interrupt()":       "deliver the return to you",
			"skyra.send(to, payload)": "send an observation to a name in your Horizon; its return is routed again",
			"skyra.fault(cause)":      "refuse: you receive the return as a fault with this cause",
		},
		Horizon:    shown,
		Faults:     []string{record.Malformed, record.Horizon, record.Contract, record.Custody, record.Leaf},
		Interrupts: []string{Novelty, ByRoute, Depth},
		Write: `invoke system.route with {"step": <the novel step's index>, "source": "def route():\n    ..."}: ` +
			"Starlark, branching with if (no for, while or recursion), returning exactly one decision",
	}
}

// RecordedRun is one route run as the log holds it: the runner's input and
// its emission.
type RecordedRun struct{ Input, Output []byte }

// Writing is a system.route invoke while its route is tested against the log.
type Writing struct {
	Key          RouteKey
	Source       []byte
	Accept       []int
	Presentation string
	Author       string // the caller's actor identity
	Cases        []RecordedRun
	Next         int
	Flips        []FlipView
	Flipped      []Flip
}

// FlipView is a flipped case as the parent is shown it.
type FlipView struct {
	Case     int             `json:"case"`
	Type     string          `json:"type"`
	Return   json.RawMessage `json:"return"`
	Previous json.RawMessage `json:"previous"`
	Was      json.RawMessage `json:"was"`
	Now      json.RawMessage `json:"now"`
}

// writeRoute resolves the step the caller named to its type, then loads the
// runs recorded for that key.
func writeRoute(s Instance, e Settled) (Instance, []Effect) {
	caller := s.Self.Resume
	if record.Stamp(e.Frame) != caller.Frame || len(e.Rounds) == 0 || e.Rounds[len(e.Rounds)-1].Stamp != caller.Stamp {
		s.Phase = Done
		return s, []Effect{Exit{"gathered rounds do not belong to the caller"}}
	}
	var ask struct {
		Step   int    `json:"step"`
		Source string `json:"source"`
		Accept []int  `json:"accept"`
	}
	if err := json.Unmarshal(s.R.Payload, &ask); err != nil {
		return wrote(s, map[string]any{"ok": false, "error": "step and accept are indexes: " + err.Error()})
	}
	var steps []Child // what the caller was shown, in order
	for _, round := range e.Rounds {
		for _, c := range round.Children {
			if c.Discharge != nil {
				steps = append(steps, c)
			}
		}
	}
	if caller.Surface == nil {
		return wrote(s, map[string]any{"ok": false, "error": "only an actor writes routes"})
	}
	if ask.Step < 0 || ask.Step >= len(steps) {
		return wrote(s, map[string]any{"ok": false, "error": fmt.Sprintf("there is no step %d", ask.Step)})
	}
	t, ok := TypeOf(*caller, steps[ask.Step])
	if !ok || strings.HasPrefix(t.Binding, SystemPrefix) {
		return wrote(s, map[string]any{"ok": false, "error": fmt.Sprintf("step %d is not a return an actor made", ask.Step)})
	}
	slices.Sort(ask.Accept)
	s.Writing = &Writing{Key: RouteKey{Parent: horizon.Identity(*caller.Surface), Binding: t.Binding, Type: t.Type},
		Source: []byte(ask.Source), Accept: slices.Compact(ask.Accept), Presentation: t.Presentation,
		Author: horizon.Identity(*caller.Surface), Flips: []FlipView{}}
	s.Phase = Cases
	return s, []Effect{LoadRuns{Key: s.Writing.Key}}
}

func retest(s Instance, e RunsLoaded) (Instance, []Effect) {
	s.Writing.Cases = e.Runs
	return nextCase(s)
}

// nextCase runs the new source over the next recorded case: the same return
// and the same previous return.
func nextCase(s Instance) (Instance, []Effect) {
	w := s.Writing
	if w.Next == len(w.Cases) {
		return recorded(s)
	}
	runner, ok := s.Self.Resume.Reach[RouteRunner]
	var in RunnerInput
	if !ok || json.Unmarshal(w.Cases[w.Next].Input, &in) != nil {
		return wrote(s, map[string]any{"ok": false, "error": "the recorded runs cannot be rerun"})
	}
	in.Source = string(w.Source)
	stdin, _ := json.Marshal(in)
	h := horizon.Horizon{Stamp: s.Stamp, Surface: &runner, Reach: map[string]horizon.Surface{}, Frame: s.Stamp}
	s.Phase = Retest
	return s, []Effect{runOf(h, stdin, s.Floor)}
}

// retested compares the new decision with the recorded one.
func retested(s Instance, e Emitted) (Instance, []Effect) {
	w := s.Writing
	c := w.Cases[w.Next]
	was, now := decisionOf(c.Output, ""), decisionOf(e.Out, e.Err)
	if string(was) != string(now) {
		var in RunnerInput
		_ = json.Unmarshal(c.Input, &in)
		if in.Previous == nil {
			in.Previous = json.RawMessage("null")
		}
		w.Flips = append(w.Flips, FlipView{Case: w.Next, Type: w.Presentation, Return: in.Return, Previous: in.Previous, Was: was, Now: now})
		w.Flipped = append(w.Flipped, Flip{Input: record.Sum(c.Input), Was: was, Now: now})
	}
	w.Next++
	return nextCase(s)
}

// decisionOf is the decision a runner emitted, canonical, whatever slot it
// addressed: a recorded run addressed the discharge it routed.
func decisionOf(out []byte, failed string) json.RawMessage {
	if failed != "" {
		return contract.Canonical(map[string]any{"fault": "route runner: " + failed})
	}
	v, err := contract.Decode(out)
	if o, ok := v.(map[string]any); err == nil && ok && o["payload"] != nil {
		return contract.Canonical(o["payload"])
	}
	return contract.Canonical(map[string]any{"fault": "route runner: not a decision"})
}

// recorded records the route if every flip was accepted, with the parent as
// its author and the accepted flips beside it.
func recorded(s Instance) (Instance, []Effect) {
	w := s.Writing
	var flipped []int
	for _, f := range w.Flips {
		flipped = append(flipped, f.Case)
	}
	if !slices.Equal(flipped, w.Accept) {
		return wrote(s, map[string]any{"ok": false, "flips": w.Flips,
			"error": "the route changes recorded decisions: accept must name exactly the flipped cases to record it"})
	}
	entry := RouteEntry{Source: record.Sum(w.Source), Author: w.Author, Frame: s.Self.Resume.Frame, Accepted: w.Flipped}
	return wrote(s, map[string]any{"ok": true, "flips": w.Flips}, PutRoute{Key: w.Key, Source: w.Source, Entry: entry})
}

// wrote is system.route's result, returned to its slot like any facet's.
func wrote(s Instance, result map[string]any, effs ...Effect) (Instance, []Effect) {
	out, _ := json.Marshal(map[string]any{"to": s.Self.Frame, "payload": result})
	s.Phase = Execute
	s, more := execute(s, Emitted{Out: out})
	return s, append(effs, more...)
}

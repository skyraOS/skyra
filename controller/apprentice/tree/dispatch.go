// Proposal 58: dispatch, as Dispatch.lean defines one transition's control,
// over the walk. A Machine is one frame and the refinements it reads. Each
// input is a return's bytes or the guard the model wrote; each step gives
// one effect: run an invoke, ask the model, or end. Nothing here does I/O:
// the shell runs what RunInvoke names, asks the model what AskModel shows, and feeds
// the answers back.
//
// An arrival is distinguished on its base's refinement. One instruction is
// trusted; a conflict is reconciled, the older branch keeping the values;
// null startles, and the model's guard is committed and its own
// instruction run. Every run ends a transition, counted against the step
// bound. The model is asked at most RefinementBound times in one transition, its
// refinement bound: past that the frame ends with an error observation for its caller, as when
// the transition bound is exhausted. The refinements are in memory until the store (58, slice 4).
package tree

import (
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
)

// A Machine is one frame, the refinements it reads, and the transition in
// progress.
type Machine struct {
	Hashes          WalkHashes
	Outputs         map[string]Output  // each binary's declared output; text without one
	Refs            map[string]Version // each base's current refinement
	Bound           TransitionBound
	RefinementBound int // the most asks in one transition: its refinement bound
	Frame           Frame

	at      *Arrival   // the arrival whose transition is in progress
	states  Transition // its states so far
	asks    int        // the asks it has made
	learned bool       // whether the refinement selects what trust last ran
}

// GoalContract is the contract a frame's goal answers: the frame starts
// waiting on it, and the goal arrives as its return.
const GoalContract = "goal"

// Bounds are a frame's two bounds. Their values are not settled (58, open
// question 6): DefaultBounds is a variable, to be tuned.
type Bounds struct {
	Transitions int // the transition bound: the most transitions a frame takes
	Refinements int // the refinement bound: the most asks in one transition
}

// DefaultBounds are the bounds a frame gets when its caller sets none.
var DefaultBounds = Bounds{Transitions: 64, Refinements: 3}

// NewMachine is a frame waiting on its goal, which arrives as text.
func NewMachine(h WalkHashes, outputs map[string]Output, refs map[string]Version, b Bounds) Machine {
	return Machine{Hashes: h, Outputs: outputs, Refs: refs, Bound: TransitionBound{Transitions: b.Transitions},
		RefinementBound: b.Refinements,
		Frame:           Frame{Pending: &Awaiting{Input: h.Input(GoalContract, nil), Output: OutputText}}}
}

// Inputs.
type (
	// Answered is what the pending invoke's process gave: its stdout, its
	// exit status and its stderr. A goal is answered with exit 0.
	Answered struct {
		Bytes  []byte
		Exit   int
		Stderr []byte
	}
	// Wrote is the guard the model wrote when asked, reading the frame's $n.
	Wrote struct{ Guard string }
)

// Effects.
type (
	// RunInvoke is an invoke for the shell to run: Argv is the template's fields,
	// each $n replaced by its value's text. Its bytes come back as Answered.
	RunInvoke struct {
		Contract string
		Argv     []string
	}
	// AskModel startles the model: what arrived, where, and the frame's scope.
	// Reason is why the last guard was refused, if one was. Its guard comes
	// back as Wrote.
	AskModel struct {
		Base    string
		Arrival Arrival
		Scope   []Obs
		Reason  string
	}
	// Finished is the frame's return: `return $n` ran.
	Finished struct{ Value any }
	// Exhausted is the frame's outcome when a bound is exhausted, its transition bound
	// or one transition's asks: an error observation for its caller.
	Exhausted struct{ Value map[string]any }
)

// ErrAsking is an input the machine is not waiting for.
var ErrAsking = errors.New("the machine is not waiting for this input")

// Transitions is how many transitions the frame has taken.
func (m Machine) Transitions() int { return m.Bound.Taken() }

// Asking is whether the machine waits on the model.
func (m Machine) Asking() bool { return m.at != nil }

// Learned is whether the refinement selects what trust last ran for the
// values it ran on, or it ran once, unlearned.
func (m Machine) Learned() bool { return m.learned }

// States is the transition in progress, so far.
func (m Machine) States() Transition { return slices.Clone(m.states) }

// Step takes one input and gives the next effect.
func (m Machine) Step(in any) (Machine, any, error) {
	m.Refs = maps.Clone(m.Refs)
	if m.Refs == nil {
		m.Refs = map[string]Version{}
	}
	switch in := in.(type) {
	case Answered:
		if m.at != nil {
			return m, nil, ErrAsking
		}
		a, err := SenseReturn(m.Hashes, m.Frame, in.Bytes, in.Exit, in.Stderr)
		if err != nil {
			return m, nil, err
		}
		m.Frame, m.at, m.states, m.asks = a.Next, &a, nil, 0
		return m.decide()
	case Wrote:
		if m.at == nil {
			return m, nil, ErrAsking
		}
		return m.wrote(in.Guard)
	}
	return m, nil, ErrAsking
}

func (m Machine) version(base string) Version {
	if v, ok := m.Refs[base]; ok {
		return v
	}
	return EmptyVersion()
}

// decide distinguishes the arrival on its base's refinement.
func (m Machine) decide() (Machine, any, error) {
	a := m.at
	v := m.version(a.Base)
	sel, err := Distinguish(v.Branches, a.Vals)
	if err != nil {
		return m, nil, err
	}
	if sel.Kind == SelectConflict {
		m.states = append(m.states, StateReconcile)
		if sel, v, _, err = Resolve(v, a.Vals); err != nil {
			return m, nil, err
		}
		m.Refs[a.Base] = v
	}
	if sel.Kind == SelectOne {
		m.learned = true
		return m.trust(sel.Instr, a.Refs)
	}
	return m.ask("")
}

// ask startles the model, or, its asks spent, ends the frame.
func (m Machine) ask(reason string) (Machine, any, error) {
	if m.asks >= m.RefinementBound {
		base := m.at.Base
		m.at = nil
		return m, Exhausted{Value: map[string]any{"error": []byte("refinement bound exhausted: the model was asked " + itoa(m.asks) +
			" times at " + base + " without a guard that runs")}}, nil
	}
	m.states = append(m.states, StateStartle)
	m.asks++
	return m, AskModel{Base: m.at.Base, Arrival: *m.at, Scope: slices.Clone(m.Frame.Scope), Reason: reason}, nil
}

// reject refuses the guard and asks again.
func (m Machine) reject(err error) (Machine, any, error) {
	m.states = append(m.states, StateReject)
	return m.ask(err.Error())
}

// wrote stores, checks and commits the model's guard, and runs its own
// instruction for what arrived. A guard reading a $n outside the base is
// not stored: it runs once, unlearned.
func (m Machine) wrote(guard string) (Machine, any, error) {
	a := m.at
	chain, err := StoreGuard(guard, a.Table)
	var u Unstorable
	if errors.As(err, &u) {
		return m.unlearned(guard)
	}
	if err != nil {
		return m.reject(err)
	}
	if err := CheckGuard(chain, shapesOf(*a)); err != nil {
		return m.reject(err)
	}
	v, _, _, err := CommitChain(m.version(a.Base), chain)
	if err != nil {
		return m.reject(err)
	}
	m.Refs[a.Base] = v
	m.states = append(m.states, StateReconcile)
	after, err := AfterStartle(chain, v.Branches, a.Vals)
	if err != nil {
		return m, nil, err
	}
	if after.Startle {
		return m.ask("the guard does not select what arrived")
	}
	m.learned = after.Learned
	return m.trust(after.Instr, a.Refs)
}

// unlearned evaluates the guard over every $n in the frame, giving those
// outside the base positions of their own, and runs what it selects
// without storing anything.
func (m Machine) unlearned(guard string) (Machine, any, error) {
	a := m.at
	table, refs, vals := maps.Clone(a.Table), maps.Clone(a.Refs), maps.Clone(a.Vals)
	next := len(refs)
	for i, o := range m.Frame.Scope {
		n := i + 1
		if _, ok := table[n]; ok {
			continue
		}
		p := Passed(next)
		next++
		table[n], refs[p], vals[p.cel()] = p, n, o.Value
	}
	chain, err := StoreGuard(guard, table)
	if err != nil {
		return m.reject(err)
	}
	after, err := AfterStartle(chain, nil, vals)
	if err != nil {
		return m.reject(err)
	}
	if after.Startle {
		return m.reject(errors.New("the guard does not select what arrived"))
	}
	m.learned = false
	return m.trust(after.Instr, refs)
}

func shapesOf(a Arrival) map[Position]Shape {
	out := map[Position]Shape{}
	for p := range a.Refs {
		out[p] = ShapeOf(a.Vals[p.cel()])
	}
	return out
}

// trust ends the transition, counting it, and runs the instruction: bound
// to the frame's $n, then invoked, or the frame's return.
func (m Machine) trust(stored string, refs map[Position]int) (Machine, any, error) {
	tmpl, err := Bind(stored, refs)
	if err != nil {
		return m, nil, err
	}
	m.states = append(m.states, StateTrust)
	if err := m.Bound.Take(m.states); errors.Is(err, ErrTransitionBound) {
		m.at = nil
		return m, Exhausted{Value: m.Bound.Exhausted()}, nil
	} else if err != nil {
		return m, nil, err
	}
	m.at = nil
	fields := strings.Fields(tmpl)
	if len(fields) == 2 && fields[0] == "return" {
		if n, w := number(fields[1]); w == len(fields[1]) {
			if o, ok := m.Frame.get(n); ok {
				return m, Finished{Value: o.Value}, nil
			}
			return m, nil, Unbound{n}
		}
	}
	contract, args := invokeOf(tmpl)
	var binary string
	if len(fields) > 0 {
		binary = fields[0]
	}
	a, err := SenseInvoke(m.Hashes, m.Frame, contract, m.Outputs[binary], args)
	if err != nil {
		return m, nil, err
	}
	m.Frame = a.Next
	argv := make([]string, len(fields))
	for i, f := range fields {
		argv[i], _ = references(f, func(n int) (string, error) {
			o, _ := m.Frame.get(n)
			return argText(o.Value), nil
		})
	}
	return m, RunInvoke{Contract: contract, Argv: argv}, nil
}

// invokeOf is an instruction's invoke contract, the template with its k-th
// reference as the slot @passed[k], and the $n it passes, in order.
func invokeOf(tmpl string) (string, []int) {
	var args []int
	contract, _ := references(tmpl, func(n int) (string, error) {
		args = append(args, n)
		return Passed(len(args) - 1).ref(), nil
	})
	return contract, args
}

// argText is a value as an argument: a string or bytes as themselves, anything
// else as JSON.
func argText(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	b, _ := json.Marshal(v)
	return string(b)
}

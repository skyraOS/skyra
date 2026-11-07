// Proposal 58: dispatch, as Dispatch.lean defines one transition's control,
// over the walk and frames. A Machine is a stack of frames, the root first,
// and the refinements they read. Each input is a process's result or the
// guard the model wrote; each step gives one effect: run an invoke, ask the
// model, or end. Nothing here does I/O: the shell runs what RunInvoke names,
// asks the model what AskModel shows, and feeds the answers back.
//
// An arrival is distinguished on its base's refinement. One instruction is
// trusted; a conflict is reconciled, the older branch keeping the values;
// null startles, and the model's guard is committed and its own
// instruction run. Every run ends a transition, counted against the frame's
// transition bound; the model is asked at most its refinement bound's times
// in one transition. An instruction runs a tool, returns, or delegates to a
// child frame (frames.go), which is pushed and runs to its return; what it
// returns, or its error observation when a bound is exhausted, arrives at
// its parent as the return of the parent's pending invoke. The refinements
// are in memory until the store (58, slice 4).
package tree

import (
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
)

// A level is one frame on the stack and the transition in progress in it.
type level struct {
	frame   Frame
	bound   TransitionBound
	at      *Arrival   // the arrival whose transition is in progress
	states  Transition // its states so far
	asks    int        // the asks it has made
	learned bool       // whether the refinement selects what trust last ran
	goal    bool       // whether it waits on its goal: only the root, at its start
}

// A Machine is a stack of frames, the root first, and the refinements they
// read.
type Machine struct {
	Hashes  WalkHashes
	Outputs map[string]Output  // each binary's declared output; text without one
	Refs    map[string]Version // each base's current refinement
	Bounds  Bounds             // every frame's

	stack []level
}

// GoalContract is the contract a root's goal answers: the root starts
// waiting on it, passed its tools, and the goal arrives as its return.
const GoalContract = "goal"

// Bounds are a frame's two bounds. Their values are not settled (58, open
// question 6): DefaultBounds is a variable, to be tuned.
type Bounds struct {
	Transitions int // the transition bound: the most transitions a frame takes
	Refinements int // the refinement bound: the most asks in one transition
	// Act is the act bound, k: a frame holding at most Act tool types acts,
	// one holding more delegates. It is the owner's, set where the tree is
	// started; no model is shown it.
	Act int
}

// DefaultBounds are the bounds a frame gets when its caller sets none.
var DefaultBounds = Bounds{Transitions: 64, Refinements: 3, Act: 1}

// NewMachine is a root frame granted its tools, waiting on its goal, which
// arrives as text and becomes its $1, the tools $2 on.
func NewMachine(h WalkHashes, outputs map[string]Output, refs map[string]Version, b Bounds, tools []string) Machine {
	var scope []Obs
	var args []int
	for i, t := range tools {
		scope = append(scope, Obs{Value: t, Base: h.Tool(t), Tool: t})
		args = append(args, i+1)
	}
	bases := make([]string, len(scope))
	for i, o := range scope {
		bases[i] = o.Base
	}
	root := Frame{Scope: scope, Pending: &Awaiting{Input: h.Input(GoalContract, bases), Output: OutputText,
		Args: args, Passed: slices.Clone(scope)}}
	return Machine{Hashes: h, Outputs: outputs, Refs: refs, Bounds: b,
		stack: []level{{frame: root, bound: TransitionBound{Transitions: b.Transitions}, goal: true}}}
}

// goalFirst is the root's goal arrival with its scope in roles' order: the
// goal $1, the tools after it. The goal's base is the goal contract's
// return, keyed on the tools' bases and never on the goal's text.
func goalFirst(a Arrival) Arrival {
	scope := a.Next.Scope
	goal := scope[len(scope)-1]
	goal.Goal = true
	a.Next.Scope = append([]Obs{goal}, scope[:len(scope)-1]...)
	a.Table, a.Refs = map[int]Position{1: Arrived}, map[Position]int{Arrived: 1}
	for i := range scope[:len(scope)-1] {
		a.Table[i+2], a.Refs[Passed(i)] = Passed(i), i+2
	}
	return a
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
	// RunInvoke is an invoke for the shell to run: Argv is the template's
	// fields, each $n replaced by its value's text, the first the binary.
	// Its result comes back as Answered.
	RunInvoke struct {
		Contract string
		Argv     []string
	}
	// AskModel startles the model: what arrived, where, the frame's scope,
	// its kind (what it may write), and how deep it is (the root is 0).
	// Reason is why the last guard was refused, if one was. Its guard comes
	// back as Wrote. The act bound is not in it: a model sees only its kind.
	AskModel struct {
		Base    string
		Arrival Arrival
		Scope   []Obs
		Kind    FrameKind
		Depth   int
		Reason  string
	}
	// Finished is the root's return: its `return $n` ran.
	Finished struct{ Value any }
	// Exhausted is the root's outcome when one of its bounds is exhausted:
	// an error observation. A child's goes to its parent instead.
	Exhausted struct{ Value map[string]any }
)

// ErrAsking is an input the machine is not waiting for.
var ErrAsking = errors.New("the machine is not waiting for this input")

func (m *Machine) top() *level { return &m.stack[len(m.stack)-1] }

// Frame is the frame running now: the deepest.
func (m Machine) Frame() Frame { return m.stack[len(m.stack)-1].frame }

// Depth is how deep the running frame is: the root is 0.
func (m Machine) Depth() int { return len(m.stack) - 1 }

// Transitions is how many transitions the running frame has taken.
func (m Machine) Transitions() int { return m.stack[len(m.stack)-1].bound.Taken() }

// Asking is whether the machine waits on the model.
func (m Machine) Asking() bool { return m.stack[len(m.stack)-1].at != nil }

// Learned is whether the refinement selects what trust last ran for the
// values it ran on, or it ran once, unlearned.
func (m Machine) Learned() bool { return m.stack[len(m.stack)-1].learned }

// States is the running frame's transition in progress, so far.
func (m Machine) States() Transition { return slices.Clone(m.stack[len(m.stack)-1].states) }

// Step takes one input and gives the next effect.
func (m Machine) Step(in any) (Machine, any, error) {
	m.Refs = maps.Clone(m.Refs)
	if m.Refs == nil {
		m.Refs = map[string]Version{}
	}
	m.stack = slices.Clone(m.stack)
	if len(m.stack) == 0 {
		return m, nil, ErrAsking
	}
	t := m.top()
	switch in := in.(type) {
	case Answered:
		if t.at != nil {
			return m, nil, ErrAsking
		}
		a, err := SenseReturn(m.Hashes, t.frame, in.Bytes, in.Exit, in.Stderr)
		if err != nil {
			return m, nil, err
		}
		if t.goal {
			a, t.goal = goalFirst(a), false
		}
		m.arrive(a)
		return m.decide()
	case Wrote:
		if t.at == nil {
			return m, nil, ErrAsking
		}
		return m.wrote(in.Guard)
	}
	return m, nil, ErrAsking
}

// arrive starts the running frame's next transition at a.
func (m *Machine) arrive(a Arrival) {
	t := m.top()
	t.frame, t.at, t.states, t.asks = a.Next, &a, nil, 0
}

func (m Machine) version(base string) Version {
	if v, ok := m.Refs[base]; ok {
		return v
	}
	return EmptyVersion()
}

// decide distinguishes the running frame's arrival on its base's
// refinement.
func (m Machine) decide() (Machine, any, error) {
	t := m.top()
	a := t.at
	v := m.version(a.Base)
	sel, err := Distinguish(v.Branches, a.Vals)
	if err != nil {
		return m, nil, err
	}
	if sel.Kind == SelectConflict {
		t.states = append(t.states, StateReconcile)
		if sel, v, _, err = Resolve(v, a.Vals); err != nil {
			return m, nil, err
		}
		m.Refs[a.Base] = v
	}
	if sel.Kind == SelectOne {
		t.learned = true
		return m.trust(sel.Instr, a.Refs)
	}
	return m.ask("")
}

// ask startles the model, or, the frame's asks spent, ends the frame.
func (m Machine) ask(reason string) (Machine, any, error) {
	t := m.top()
	if t.asks >= m.Bounds.Refinements {
		return m.end(map[string]any{"error": []byte("refinement bound exhausted: the model was asked " +
			itoa(t.asks) + " times at " + t.at.Base + " without a guard that runs")}, true)
	}
	t.states = append(t.states, StateStartle)
	t.asks++
	return m, AskModel{Base: t.at.Base, Arrival: *t.at, Scope: slices.Clone(t.frame.Scope),
		Kind: KindOf(m.Bounds.Act, t.frame.Scope), Depth: m.Depth(), Reason: reason}, nil
}

// reject refuses the guard or instruction and asks again.
func (m Machine) reject(err error) (Machine, any, error) {
	t := m.top()
	t.states = append(t.states, StateReject)
	return m.ask(err.Error())
}

// wrote stores, checks and commits the model's guard, and runs its own
// instruction for what arrived. A guard reading a $n outside the base is
// not stored: it runs once, unlearned.
func (m Machine) wrote(guard string) (Machine, any, error) {
	t := m.top()
	a := t.at
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
	if err := m.runnable(chain, a.Refs); err != nil {
		return m.reject(err)
	}
	v, _, _, err := CommitChain(m.version(a.Base), chain)
	if err != nil {
		return m.reject(err)
	}
	m.Refs[a.Base] = v
	t.states = append(t.states, StateReconcile)
	after, err := AfterStartle(chain, v.Branches, a.Vals)
	if err != nil {
		return m, nil, err
	}
	if after.Startle {
		return m.ask("the guard does not select what arrived")
	}
	t.learned = after.Learned
	return m.trust(after.Instr, a.Refs)
}

// unlearned evaluates the guard over every $n in the frame, giving those
// outside the base positions of their own, and runs what it selects
// without storing anything.
func (m Machine) unlearned(guard string) (Machine, any, error) {
	t := m.top()
	a := t.at
	table, refs, vals := maps.Clone(a.Table), maps.Clone(a.Refs), maps.Clone(a.Vals)
	next := len(refs)
	for i, o := range t.frame.Scope {
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
	if err := m.runnable(chain, refs); err != nil {
		return m.reject(err)
	}
	after, err := AfterStartle(chain, nil, vals)
	if err != nil {
		return m.reject(err)
	}
	if after.Startle {
		return m.reject(errors.New("the guard does not select what arrived"))
	}
	t.learned = false
	return m.trust(after.Instr, refs)
}

// runnable judges each instruction a guard returns by its first position,
// before anything is committed: a guard whose instruction cannot run is
// refused. A position's base fixes whether it holds a tool, so what is
// runnable here is runnable wherever the guard is selected.
func (m Machine) runnable(chain []Link, refs map[Position]int) error {
	for _, l := range chain {
		if l.Instr == "" {
			continue
		}
		tmpl, err := Bind(l.Instr, refs)
		if err != nil {
			return err
		}
		in, err := ParseInstr(tmpl)
		if err != nil {
			return err
		}
		if _, err := ClassifyInstr(m.Bounds.Act, m.Hashes, m.top().frame.Scope, in); err != nil {
			return err
		}
	}
	return nil
}

func shapesOf(a Arrival) map[Position]Shape {
	out := map[Position]Shape{}
	for p := range a.Refs {
		out[p] = ShapeOf(a.Vals[p.cel()])
	}
	return out
}

// trust binds the instruction to the frame's $n and judges it by its first
// position; an invalid one is rejected and the model asked again. Otherwise
// it ends the transition, counting it, and runs: a tool as a process, a
// return to the caller, or a child frame.
func (m Machine) trust(stored string, refs map[Position]int) (Machine, any, error) {
	t := m.top()
	tmpl, err := Bind(stored, refs)
	if err != nil {
		return m, nil, err
	}
	in, err := ParseInstr(tmpl)
	if err != nil {
		return m.reject(err)
	}
	c, err := ClassifyInstr(m.Bounds.Act, m.Hashes, t.frame.Scope, in)
	if err != nil {
		return m.reject(err)
	}
	t.states = append(t.states, StateTrust)
	if err := t.bound.Take(t.states); errors.Is(err, ErrTransitionBound) {
		return m.end(t.bound.Exhausted(), true)
	} else if err != nil {
		return m, nil, err
	}
	t.at = nil
	if c.Kind == InstrReturn {
		return m.end(c.Return.Value, false)
	}
	contract, args := invokeOf(tmpl)
	if c.Kind == InstrDelegate {
		contract, args = DelegateContract(c.Order)
	}
	out := m.Outputs[c.Binary]
	a, err := SenseInvoke(m.Hashes, t.frame, contract, out, args)
	if err != nil {
		return m, nil, err
	}
	t.frame = a.Next
	if c.Kind == InstrDelegate {
		m.stack = append(m.stack, level{bound: TransitionBound{Transitions: m.Bounds.Transitions}})
		m.arrive(ChildArrival(a, c.Passed))
		return m.decide()
	}
	// argv: each field with its arguments as their values' text, cut, in
	// the order classified.
	fields := strings.Fields(tmpl)
	argv := make([]string, len(fields))
	next := 0
	for i, f := range fields {
		argv[i], _ = cutRefs(f, func(int, Sel) (string, error) {
			next++
			return argText(c.Passed[next-1].Value), nil
		})
	}
	return m, RunInvoke{Contract: contract, Argv: argv}, nil
}

// end ends the running frame with a value: the root's is the run's outcome;
// a child's walks back to its parent's pending invoke, which decides it.
// exhausted marks an exhausted bound's error observation.
func (m Machine) end(v any, exhausted bool) (Machine, any, error) {
	m.top().at = nil
	if len(m.stack) == 1 {
		if exhausted {
			return m, Exhausted{Value: v.(map[string]any)}, nil
		}
		return m, Finished{Value: v}, nil
	}
	m.stack = m.stack[:len(m.stack)-1]
	a, err := SenseReturned(m.Hashes, m.top().frame, v)
	if err != nil {
		return m, nil, err
	}
	m.arrive(a)
	return m.decide()
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

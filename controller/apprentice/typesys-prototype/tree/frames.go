// Proposal 58: frames and children, as Frames.lean defines them. A frame's
// tools are observations in its scope, each keyed on its binary. Its kind
// is counted from them under the act bound k, the owner's: an act frame
// holds at most k and runs them; a delegate holds more and hands work on.
// An instruction's first position decides what it is, as exec's first
// argument does: `return $n` returns to the caller; in an act frame, a $n
// holding a tool runs that binary; in a delegate, `delegate(goal(a), …)`
// starts a child, its goal taken from an observation, never written, and
// passed observations, tools among them like any other. A child holds
// strictly fewer tool types than its emitter, and at least one. An argument
// is a $n or a cut of one (cut.go).
//
// An observation's role is goal, tool or context. A scope is its goal, $1,
// then its tools, then its context, in order: tools are fixed when a frame
// starts and context only grows at the end, so a $n a learned guard names
// keeps its meaning.
package tree

import (
	"errors"
	"strings"
)

// Tools is a scope's tool types: the binaries among it, each once, in the
// order they first appear.
func Tools(scope []Obs) []string {
	var out []string
	seen := map[string]bool{}
	for _, o := range scope {
		if o.Tool != "" && !seen[o.Tool] {
			seen[o.Tool] = true
			out = append(out, o.Tool)
		}
	}
	return out
}

// A FrameKind is what a frame may write: an act frame runs its tools, a
// delegate hands work on; both return.
type FrameKind uint8

const (
	KindActs FrameKind = iota
	KindDelegates
)

// KindOf is a frame's kind under the act bound k.
func KindOf(k int, scope []Obs) FrameKind {
	if len(Tools(scope)) <= k {
		return KindActs
	}
	return KindDelegates
}

// An InstrKind is what an instruction's first position makes it.
type InstrKind uint8

const (
	InstrReturn InstrKind = iota
	InstrRun
	InstrDelegate
)

// An Arg is a $n and what it takes of it.
type Arg struct {
	N   int
	Sel Sel
}

// An Instr is an instruction as the core reads it. Return is the returned
// argument, its stdout, and Exit its exit status, 0 when the goal is done and
// 1 when it is not; N is the tool run, a whole $n; Goal is a delegation's
// child's goal; Args are the other arguments, in order: a return's is its
// stderr, if it names one.
type Instr struct {
	Kind   InstrKind
	Return Arg
	Exit   int
	N      int
	Goal   Arg
	Args   []Arg
}

// A Role is what an observation is to its frame.
type Role uint8

const (
	RoleGoal Role = iota
	RoleTool
	RoleContext
)

func (o Obs) Role() Role {
	switch {
	case o.Goal:
		return RoleGoal
	case o.Tool != "":
		return RoleTool
	}
	return RoleContext
}

// arg reads a field that is one argument: $n and its selector, nothing
// else.
func arg(f string) (Arg, bool) {
	n, w := number(f)
	if w == 0 {
		return Arg{}, false
	}
	s, sw, err := selector(f[w:])
	if err != nil || w+sw != len(f) {
		return Arg{}, false
	}
	return Arg{n, s}, true
}

// ParseInstr reads a bound instruction, its $n the frame's, by its first
// field.
func ParseInstr(tmpl string) (Instr, error) {
	fields := strings.Fields(tmpl)
	if len(fields) == 0 {
		return Instr{}, errors.New("an empty instruction")
	}
	head := fields[0]
	if head == "return" {
		if (len(fields) == 3 || len(fields) == 4) && (fields[1] == "0" || fields[1] == "1") {
			in := Instr{Kind: InstrReturn, Exit: int(fields[1][0] - '0')}
			a, ok := arg(fields[2])
			in.Return = a
			if ok && len(fields) == 4 {
				var e Arg
				e, ok = arg(fields[3])
				in.Args = []Arg{e}
			}
			if ok {
				return in, nil
			}
		}
		return Instr{}, errors.New("return takes an exit status, 0 when the goal is done or 1 when it is not, then its stdout, a $n or a cut of one, and optionally its stderr, another")
	}
	if n, w := number(head); w > 0 {
		if w != len(head) {
			return Instr{}, errors.New("a tool is run whole: the first position is a $n, not a cut")
		}
		in := Instr{Kind: InstrRun, N: n}
		for _, f := range fields[1:] {
			if _, err := cutRefs(f, func(n int, s Sel) (string, error) {
				in.Args = append(in.Args, Arg{n, s})
				return "", nil
			}); err != nil {
				return Instr{}, err
			}
		}
		return in, nil
	}
	inner, ok := strings.CutPrefix(strings.TrimSpace(tmpl), "delegate(")
	if inner, ok = strings.CutSuffix(inner, ")"); !ok {
		return Instr{}, errInstr
	}
	parts := splitArgs(inner)
	g, ok := strings.CutPrefix(parts[0], "goal(")
	if g, ok = strings.CutSuffix(g, ")"); !ok {
		return Instr{}, errors.New("a delegation's first argument is the goal: goal($n), or goal of a cut")
	}
	in := Instr{Kind: InstrDelegate}
	if in.Goal, ok = arg(g); !ok {
		return Instr{}, errors.New("a goal is taken from a $n, or a cut of one: " + g + " is neither")
	}
	for _, f := range parts[1:] {
		a, ok := arg(f)
		if !ok {
			return Instr{}, errors.New("a delegation passes only observations: " + f + " is not a $n or a cut of one")
		}
		in.Args = append(in.Args, a)
	}
	return in, nil
}

var errInstr = errors.New("this is not an instruction: write one of the forms under Instructions")

// Canonical is an instruction as it is compared: its form and its $n, in
// order, nothing else. Text the model wrote between them is dropped, so two
// instructions agree exactly when they pass the same parameters in the same
// order. One that is not an instruction is left as it is, to be refused
// where it is bound.
func Canonical(tmpl string) string {
	in, err := ParseInstr(tmpl)
	if err != nil {
		return tmpl
	}
	return in.String()
}

func (a Arg) String() string { return "$" + itoa(a.N) + a.Sel.String() }

// String is the instruction in its canonical form.
func (in Instr) String() string {
	args := make([]string, len(in.Args))
	for i, a := range in.Args {
		args[i] = a.String()
	}
	switch in.Kind {
	case InstrReturn:
		return strings.Join(append([]string{"return", itoa(in.Exit), in.Return.String()}, args...), " ")
	case InstrDelegate:
		return "delegate(" + strings.Join(append([]string{"goal(" + in.Goal.String() + ")"}, args...), ", ") + ")"
	}
	return strings.Join(append([]string{"$" + itoa(in.N)}, args...), " ")
}

// splitArgs splits at the commas between arguments, not those inside a
// cut's brackets or a goal's parentheses, each trimmed.
func splitArgs(s string) []string {
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(s[start:]))
}

// NotTool is a first position that is a $n but not a tool.
type NotTool struct{ N int }

func (e NotTool) Error() string {
	return "$" + itoa(e.N) + " is not a tool: the first position is a tool to run"
}

// CutFails is a cut that does not apply to $n's value: not text, lines it
// does not have, or a field it does not hold.
type CutFails struct {
	N   int
	Sel Sel
}

func (e CutFails) Error() string {
	return "$" + itoa(e.N) + e.Sel.String() + " does not apply: lines cut text, and a field cut an object holding it"
}

// ActDelegates is a delegation from an act frame, DelegateRuns a run from a
// delegate, and NoTools a delegation passing no tool.
var (
	ActDelegates = errors.New("you run your tools and return: you do not delegate")
	DelegateRuns = errors.New("you delegate and return: you do not run tools yourself")
	NoTools      = errors.New("a delegation passes at least one tool")
)

// GoalTool is a goal taken from a tool, whole.
type GoalTool struct{ N int }

func (e GoalTool) Error() string {
	return "$" + itoa(e.N) + " is a tool, and a goal is what to do: take it from the goal or the context"
}

// Narrows is a delegation that does not narrow: a child holds fewer tool
// types than its emitter.
type Narrows struct{ Passed, Held int }

func (e Narrows) Error() string {
	return "a delegation passes fewer tools than you have: this passes " + itoa(e.Passed) +
		" of your " + itoa(e.Held)
}

// A Classified instruction: what it returns, the binary it runs, or the
// child it starts, and what it passes: a child's whole scope, its goal,
// then its tools, then its context, and the arguments they were taken by,
// in that order.
type Classified struct {
	Kind   InstrKind
	Return Obs
	Exit   int
	Binary string
	Passed []Obs
	Order  []Arg
}

// ClassifyInstr judges an instruction against the frame's scope under the
// act bound k, cutting its arguments; h keys a cut's base. A rejection is
// Unbound, CutFails, NotTool, Narrows, GoalTool, NoTools, ActDelegates or
// DelegateRuns.
func ClassifyInstr(k int, h WalkHashes, scope []Obs, in Instr) (Classified, error) {
	cutBase, kind := h.Cut, KindOf(k, scope)
	f := Frame{Scope: scope}
	argObs := func(a Arg) (Obs, error) {
		o, ok := f.get(a.N)
		if !ok {
			return Obs{}, Unbound{a.N}
		}
		c, ok := CutObs(cutBase, a.Sel, o)
		if !ok {
			return Obs{}, CutFails{a.N, a.Sel}
		}
		return c, nil
	}
	resolve := func(args []Arg) ([]Obs, error) {
		var ps []Obs
		for _, a := range args {
			o, err := argObs(a)
			if err != nil {
				return nil, err
			}
			ps = append(ps, o)
		}
		return ps, nil
	}
	switch in.Kind {
	case InstrReturn:
		o, err := argObs(in.Return)
		if err != nil {
			return Classified{}, err
		}
		ps, err := resolve(in.Args)
		if err != nil {
			return Classified{}, err
		}
		return Classified{Kind: InstrReturn, Return: o, Exit: in.Exit, Passed: ps}, nil
	case InstrRun:
		if kind == KindDelegates {
			return Classified{}, DelegateRuns
		}
		o, ok := f.get(in.N)
		if !ok {
			return Classified{}, Unbound{in.N}
		}
		if o.Tool == "" {
			return Classified{}, NotTool{in.N}
		}
		ps, err := resolve(in.Args)
		if err != nil {
			return Classified{}, err
		}
		return Classified{Kind: InstrRun, Binary: o.Tool, Passed: append([]Obs{o}, ps...)}, nil
	}
	if kind == KindActs {
		return Classified{}, ActDelegates
	}
	g, err := argObs(in.Goal)
	if err != nil {
		return Classified{}, err
	}
	if g.Tool != "" {
		return Classified{}, GoalTool{in.Goal.N}
	}
	ps, err := resolve(in.Args)
	if err != nil {
		return Classified{}, err
	}
	switch p, held := len(Tools(ps)), len(Tools(scope)); {
	case p == 0:
		return Classified{}, NoTools
	case p >= held:
		return Classified{}, Narrows{p, held}
	}
	// The child's scope in roles' order: its goal, then the tools, then the
	// rest, each in the order passed. Nothing passed is a goal.
	g.Goal = true
	c := Classified{Kind: InstrDelegate, Passed: []Obs{g}, Order: []Arg{in.Goal}}
	for _, tools := range []bool{true, false} {
		for i, o := range ps {
			if (o.Tool != "") == tools {
				o.Goal = false
				c.Passed, c.Order = append(c.Passed, o), append(c.Order, in.Args[i])
			}
		}
	}
	return c, nil
}

// DelegateContract is a delegation's invoke contract, its arguments in the
// child's order, the goal first: the k-th as the slot @passed[k], cut. Its
// child's $k+1 is what @passed[k] holds.
func DelegateContract(order []Arg) (string, []int) {
	slots, ns := make([]string, len(order)), make([]int, len(order))
	for i, a := range order {
		slots[i], ns[i] = Passed(i).ref()+a.Sel.String(), a.N
	}
	slots[0] = "goal(" + slots[0] + ")"
	return "delegate(" + strings.Join(slots, ", ") + ")", ns
}

// ChildArrival is a child's first arrival: at the input base its parent's
// invoke stands at, with the passed positions, the child's $k+1 the k-th
// observation passed.
func ChildArrival(invoke Arrival, passed []Obs) Arrival {
	a := Arrival{Dir: Forward, Base: invoke.Base, Vals: map[string]any{},
		Table: map[int]Position{}, Refs: map[Position]int{}}
	for k, o := range passed {
		a.Table[k+1], a.Refs[Passed(k)], a.Vals[Passed(k).cel()] = Passed(k), k+1, o.Value
	}
	a.Next = Frame{Scope: passed}
	return a
}

// Proposal 58: frames and children, as Frames.lean defines them. A frame's
// tools are observations in its scope, each keyed on its binary. An
// instruction's first position decides what it is, as exec's first argument
// does: `return $n` returns to the caller; a $n holding a tool runs that
// binary; a word is a task, delegating to a child frame passed observations,
// tools among them like any other. A child holds strictly fewer tool types
// than its emitter, so frames nest no deeper than the root's tool types.
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

// An InstrKind is what an instruction's first position makes it.
type InstrKind uint8

const (
	InstrReturn InstrKind = iota
	InstrRun
	InstrDelegate
)

// An Instr is an instruction as the core reads it. N is the returned $n or
// the tool run; Task names a delegation; Args are the other $n, in order.
type Instr struct {
	Kind InstrKind
	N    int
	Task string
	Args []int
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
		if len(fields) == 2 {
			if n, w := number(fields[1]); w == len(fields[1]) {
				return Instr{Kind: InstrReturn, N: n}, nil
			}
		}
		return Instr{}, errors.New("return takes one $n")
	}
	if n, w := number(head); w > 0 {
		if w != len(head) {
			return Instr{}, errors.New("the first position is a $n holding a tool, or a task")
		}
		in := Instr{Kind: InstrRun, N: n}
		for _, f := range fields[1:] {
			references(f, func(n int) (string, error) {
				in.Args = append(in.Args, n)
				return "", nil
			})
		}
		return in, nil
	}
	if strings.Contains(head, "$") {
		return Instr{}, errors.New("the first position is a $n holding a tool, or a task")
	}
	in := Instr{Kind: InstrDelegate, Task: head}
	for _, f := range fields[1:] {
		n, w := number(f)
		if w == 0 || w != len(f) {
			return Instr{}, errors.New("a delegation passes only observations: " + f + " is not a $n")
		}
		in.Args = append(in.Args, n)
	}
	return in, nil
}

// NotTool is a first position that is a $n but not a tool.
type NotTool struct{ N int }

func (e NotTool) Error() string {
	return "$" + itoa(e.N) + " is not a tool: the first position runs a tool or names a task"
}

// Narrows is a delegation that does not narrow: a child holds fewer tool
// types than its emitter.
type Narrows struct{ Passed, Held int }

func (e Narrows) Error() string {
	return "a child holds fewer tools than its emitter: this passes " + itoa(e.Passed) +
		" of the frame's " + itoa(e.Held)
}

// A Classified instruction: what it returns, the binary it runs, or the
// child it starts, and what it passes.
type Classified struct {
	Kind   InstrKind
	Return Obs
	Binary string
	Task   string
	Passed []Obs
}

// ClassifyInstr judges an instruction against the frame's scope. A rejection is
// Unbound, NotTool or Narrows.
func ClassifyInstr(scope []Obs, in Instr) (Classified, error) {
	f := Frame{Scope: scope}
	resolve := func(args []int) ([]Obs, error) {
		var ps []Obs
		for _, n := range args {
			o, ok := f.get(n)
			if !ok {
				return nil, Unbound{n}
			}
			ps = append(ps, o)
		}
		return ps, nil
	}
	switch in.Kind {
	case InstrReturn:
		o, ok := f.get(in.N)
		if !ok {
			return Classified{}, Unbound{in.N}
		}
		return Classified{Kind: InstrReturn, Return: o}, nil
	case InstrRun:
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
	ps, err := resolve(in.Args)
	if err != nil {
		return Classified{}, err
	}
	if p, h := len(Tools(ps)), len(Tools(scope)); p >= h {
		return Classified{}, Narrows{p, h}
	}
	return Classified{Kind: InstrDelegate, Task: in.Task, Passed: ps}, nil
}

// ChildArrival is a child's first arrival: at the input base its parent's
// invoke stands at, with the passed positions, the child's $k+1 the k-th
// observation passed.
func ChildArrival(invoke Arrival, passed []Obs) Arrival {
	a := Arrival{Dir: Forward, Base: invoke.Base, Vals: invoke.Vals,
		Table: map[int]Position{}, Refs: map[Position]int{}}
	for k := range passed {
		a.Table[k+1], a.Refs[Passed(k)] = Passed(k), k+1
	}
	a.Next = Frame{Scope: passed}
	return a
}

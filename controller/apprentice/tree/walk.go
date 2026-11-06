// Proposal 58: the walk, as Walk.lean defines it. Sense takes a frame and
// what its emitter sent, an invoke or a return, and gives where the step
// stands: the base, the values at the base's positions, and the scope table
// mapping the frame's $n to those positions. An invoke walks forward to its
// input base; a return walks back to the invoke it answers. The direction
// decides which base, and nothing else: the same transition runs at either.
//
// A frame holds at most one pending invoke, so a return answers it and
// names nothing. Parallel work is several frames.
package tree

import (
	"errors"
	"slices"
	"strings"
)

// An Obs is an observation in a frame's scope: its value and the hash of
// the base it arrived at.
type Obs struct {
	Value any
	Base  string
}

// An Awaiting is the invoke a frame waits on: its input base, how its
// contract's output reads, the $n it was passed and what they held.
type Awaiting struct {
	Input  string
	Output Output
	Args   []int
	Passed []Obs
}

// A Frame is its scope, $1 first, appended to and never rewritten, and its
// one pending invoke, if any.
type Frame struct {
	Scope   []Obs
	Pending *Awaiting
}

// WalkHashes hash an input base (the contract and the bases of what is
// passed, in order) and a return base (the input base it answers and the
// encoding of the shape that came back).
type WalkHashes struct {
	Input  func(contract string, passed []string) string
	Return func(input, shape string) string
}

type Direction uint8

const (
	Forward Direction = iota
	Back
)

// An Arrival is where the step stands. Vals are the values at the base's
// positions, by CEL name, what Distinguish evaluates. Table maps each $n to
// its position, the first it was passed at, what StoreGuard stores under;
// Refs maps each position back to the $n holding it, what Bind binds with.
type Arrival struct {
	Dir   Direction
	Base  string
	Vals  map[string]any
	Table map[int]Position
	Refs  map[Position]int
	Next  Frame
}

var (
	// ErrBusy is an invoke while one is pending: one at a time.
	ErrBusy = errors.New("an invoke is pending: a frame invokes one at a time")
	// ErrIdle is a return with nothing pending, a custody fault.
	ErrIdle = errors.New("a return with no invoke pending answers nothing")
)

// Unbound is a $n that names nothing in the frame's scope.
type Unbound struct{ N int }

func (u Unbound) Error() string { return "$" + itoa(u.N) + " names nothing in this frame" }

// get is $n: $1 is the first observation; there is no $0.
func (f Frame) get(n int) (Obs, bool) {
	if n < 1 || n > len(f.Scope) {
		return Obs{}, false
	}
	return f.Scope[n-1], true
}

// SenseInvoke walks forward: the invoke's base is its contract and the bases of
// what it passes, and it becomes the frame's pending invoke.
func SenseInvoke(h WalkHashes, f Frame, contract string, out Output, args []int) (Arrival, error) {
	if f.Pending != nil {
		return Arrival{}, ErrBusy
	}
	passed := make([]Obs, len(args))
	bases := make([]string, len(args))
	for k, n := range args {
		o, ok := f.get(n)
		if !ok {
			return Arrival{}, Unbound{n}
		}
		passed[k], bases[k] = o, o.Base
	}
	base := h.Input(contract, bases)
	a := Arrival{Dir: Forward, Base: base, Vals: map[string]any{},
		Table: map[int]Position{}, Refs: map[Position]int{}}
	a.pass(args, passed)
	a.Next = Frame{Scope: slices.Clip(f.Scope),
		Pending: &Awaiting{Input: base, Output: out, Args: slices.Clone(args), Passed: passed}}
	return a, nil
}

// SenseReturn walks back to the pending invoke: what its process gave
// arrives (its stdout as its contract's output reads if it exited 0, the
// failure otherwise), the base is its input base and that shape, the value
// is appended to the scope, and nothing is pending.
func SenseReturn(h WalkHashes, f Frame, stdout []byte, exit int, stderr []byte) (Arrival, error) {
	p := f.Pending
	if p == nil {
		return Arrival{}, ErrIdle
	}
	v, s := ArriveExit(p.Output, stdout, exit, stderr)
	base := h.Return(p.Input, s.Enc())
	n := len(f.Scope) + 1
	a := Arrival{Dir: Back, Base: base, Vals: map[string]any{Arrived.cel(): v},
		Table: map[int]Position{n: Arrived}, Refs: map[Position]int{Arrived: n}}
	a.pass(p.Args, p.Passed)
	scope := append(slices.Clip(f.Scope), Obs{Value: v, Base: base})
	a.Next = Frame{Scope: scope}
	return a, nil
}

func (a *Arrival) pass(args []int, passed []Obs) {
	for k, n := range args {
		a.Vals[Passed(k).cel()] = passed[k].Value
		a.Refs[Passed(k)] = n
		if _, ok := a.Table[n]; !ok {
			a.Table[n] = Passed(k)
		}
	}
}

// StoreInstr stores an instruction template under a scope table: each $n
// becomes its position's reference. A $n the table lacks is Unstorable: it
// runs once, unlearned.
func StoreInstr(tmpl string, table map[int]Position) (string, error) {
	return references(tmpl, func(n int) (string, error) {
		p, ok := table[n]
		if !ok {
			return "", Unstorable{n}
		}
		return p.ref(), nil
	})
}

// Bind binds a stored instruction in a frame: each @arrived and @passed[k]
// becomes the $n holding that position there. What a step stores binds back
// to what it was written as.
func Bind(stored string, refs map[Position]int) (string, error) {
	var out strings.Builder
	for i := 0; i < len(stored); {
		p, w := position(stored[i:])
		if w == 0 {
			out.WriteByte(stored[i])
			i++
			continue
		}
		n, ok := refs[p]
		if !ok {
			return "", errors.New(p.ref() + " is not a position of this base")
		}
		out.WriteString("$" + itoa(n))
		i += w
	}
	return out.String(), nil
}

// position reads @arrived or @passed[k] at the start of s, giving the
// position and its width, or width 0.
func position(s string) (Position, int) {
	if strings.HasPrefix(s, "@arrived") {
		return Arrived, len("@arrived")
	}
	rest, ok := strings.CutPrefix(s, "@passed[")
	if !ok {
		return 0, 0
	}
	k, w := 0, 0
	for w < len(rest) && rest[w] >= '0' && rest[w] <= '9' && w < 9 {
		k = k*10 + int(rest[w]-'0')
		w++
	}
	if w == 0 || w >= len(rest) || rest[w] != ']' {
		return 0, 0
	}
	return Passed(k), len("@passed[") + w + 1
}

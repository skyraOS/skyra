// Package machine is the whole machine. An instance is a pure state machine:
//
//	Step(state, event) → (state, effects)
//
// The top-level decision is the Move's state (D9). Intake decides it, and
// each state is its own sequence of phases:
//
//	intake ─┬─ invoke → release → execute → handoff   in custody: run the target, hand off its output
//	        ├─ return → release → settle parent        terminal
//	        ├─ fault  → release → settle parent        terminal
//	        └─ resolve                                 holes; not in the minimal shape
//
// The four phases:
//
//  1. Intake   receive bytes, parse them as a Move, check them against the
//     Horizon they were emitted under, commit a content-addressed
//     record. The record's sha256 is the stamp.
//  2. Release  the parent checks the stamp and sends the Horizon: which program
//     this instance runs, and which targets it may call (Reach).
//  3. Execute  run the target inside a Seatbelt sandbox rendered from the
//     Horizon. .skyra/ is always denied.
//  4. Handoff  pass the target's output, unparsed, to a new successor; check its
//     stamp, release it, and exit without waiting for it.
//
// Handoff ends by issuing the successor's Release, so one instance's Handoff
// overlaps the next instance's Intake and Release.
//
// Step makes every decision. Package instance performs the effects and reports
// what happened as the next event; it holds all the I/O and makes no decisions.
// This package is pure: no I/O, no clock, no randomness. It imports no package
// that touches the OS, which machine_test.go enforces. That purity is what lets
// replay re-derive a lineage from its recorded inputs.
package machine

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"skyra-v1/specs/2026-09-23/minimal-shape/horizon"
	"skyra-v1/specs/2026-09-23/minimal-shape/move"
	"skyra-v1/specs/2026-09-23/minimal-shape/record"
)

// Instance is the state of one instance as it moves through its phases.
type Instance struct {
	Floor string          // the .skyra path every rendered sandbox denies
	State move.State      // the Move's state, decided at intake: the top-level branch
	Phase Phase           // the phase whose event this instance is waiting for
	Under horizon.Horizon // the Horizon the input was emitted under
	R     record.Record   // what intake committed
	Stamp string          // hash of R
	Self  horizon.Horizon // this instance's own Horizon, once released
	Out   []byte          // the target's emission, while it is handed off
	Child string          // the successor's stamp, once verified and released
}

type Phase string

const (
	Intake  Phase = "intake"  // waiting for bytes
	Release Phase = "release" // committed and stamped; waiting for the Horizon
	Execute Phase = "execute" // the target is running; waiting for its output
	Handoff Phase = "handoff" // the successor is intaking the output; waiting for its stamp
	Done    Phase = "done"
)

// Events are what the world reports to Step. Each phase accepts exactly one.
type Event interface{ event() }

type Received struct { // → Intake: the bytes, and the Horizon they were emitted under
	Under horizon.Horizon
	In    []byte
}
type Released struct { // → Release: this instance's Horizon, and whether a node fd came with it
	Self horizon.Horizon
	Node bool
}
type Emitted struct { // → Execute: the target's one emission (empty: it exited without one)
	Out []byte
	Err string
}
type Stamped struct { // → Handoff: the successor's stamp, and the record the tree holds under it (nil: none)
	Stamp string
	R     *record.Record
	Err   string
}

func (Received) event() {}
func (Released) event() {}
func (Emitted) event()  {}
func (Stamped) event()  {}

// Effects are what Step asks the shell to do, in order. SendStamp, Run and
// Spawn each end with the shell waiting for the next event.
type Effect interface{ effect() }

type StoreInput struct{ In []byte }   // record the exact input bytes before using them
type Commit struct{ R record.Record } // durable write: temp, flush, rename, flush directory
type SendStamp struct{ Stamp string } // then wait for Released
type Run struct {                     // then wait for Emitted
	Surface horizon.Surface
	Stdin   []byte
	Profile string
}
type Spawn struct { // start the successor, forward the bytes, then wait for Stamped
	Under horizon.Horizon
	Out   []byte
}
type KillSuccessor struct{}
type SendHorizon struct { // issue the successor's Release: its Horizon plus a node fd
	Child    horizon.Horizon
	NewFrame bool // true: a new frame for the child; false: the node this instance holds
}
type Settle struct{ Result string } // write the result into the held node, once
type Exit struct{ Err string }

func (StoreInput) effect()    {}
func (Commit) effect()        {}
func (SendStamp) effect()     {}
func (Run) effect()           {}
func (Spawn) effect()         {}
func (KillSuccessor) effect() {}
func (SendHorizon) effect()   {}
func (Settle) effect()        {}
func (Exit) effect()          {}

// Step is the decision tree. Intake comes first, because intake is where the
// state is decided. After that, the state picks the program and the phase picks
// the step within it.
func Step(s Instance, e Event) (Instance, []Effect) {
	if s.Phase == Intake {
		if ev, ok := e.(Received); ok {
			return intake(s, ev)
		}
		return stop(s, e)
	}
	switch s.State {
	case move.StateInvoke:
		switch ev := e.(type) {
		case Released:
			if s.Phase == Release {
				return release(s, ev)
			}
		case Emitted:
			if s.Phase == Execute {
				return execute(s, ev)
			}
		case Stamped:
			if s.Phase == Handoff {
				return handoff(s, ev)
			}
		}
	case move.StateReturn, move.StateFault:
		if ev, ok := e.(Released); ok && s.Phase == Release {
			return settleParent(s, ev)
		}
	case move.StateResolve:
		return s, []Effect{Exit{"resolve (holes) is not in the minimal shape"}}
	}
	return stop(s, e)
}

func stop(s Instance, e Event) (Instance, []Effect) {
	msg := fmt.Sprintf("%T is not valid in phase %s", e, s.Phase)
	if s.State != "" {
		msg += fmt.Sprintf(" of a Move in a state of %s", s.State)
	}
	return s, []Effect{Exit{msg}}
}

// 1. Intake: parse the bytes as a Move, check them against the Horizon they
// were emitted under, decide the Move's state, and commit a record. A rejected
// Move is still committed, in a state of fault: a stamp means custody, not
// approval.
func intake(s Instance, e Received) (Instance, []Effect) {
	r := record.Record{Predecessor: e.Under.Stamp, Input: record.Sum(e.In)}
	if m, err := interpret(e.Under, e.In); err != nil {
		r.State, r.Reason = move.StateFault, err.Error()
	} else {
		r.State, r.Move = move.StateOf(m), &m
	}
	s.Under, s.R, s.Stamp, s.State, s.Phase = e.Under, r, record.Stamp(r), r.State, Release
	return s, []Effect{StoreInput{e.In}, Commit{r}, SendStamp{s.Stamp}}
}

// interpret is intake's authorization: the bytes must parse as exactly one
// Move, and an Invoke must name a surface the Horizon reaches.
func interpret(h horizon.Horizon, in []byte) (move.Move, error) {
	m, err := move.Parse(in)
	if err != nil {
		return move.Move{}, err
	}
	if m.Invoke != nil {
		if _, ok := h.Reach[m.Invoke.Target]; !ok {
			return move.Move{}, fmt.Errorf("invoke of %q is not in reach (a leaf may only return)", m.Invoke.Target)
		}
	}
	return m, nil
}

// accept is the child's half of Release, for every state: the Horizon must be
// for this commit and come with a node fd, or the instance stops.
func accept(s Instance, e Released) error {
	if e.Self.Stamp != s.Stamp || !e.Node {
		return errors.New("release is not for this commit")
	}
	return nil
}

// 2. Release, invoke: accept the Horizon and run exactly the committed
// surface, under a sandbox rendered from the Horizon.
func release(s Instance, e Released) (Instance, []Effect) {
	if err := accept(s, e); err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	args, _ := json.Marshal(s.R.Move.Invoke.Args)
	s.Self, s.Phase = e.Self, Execute
	return s, []Effect{Run{Surface: *e.Self.Surface, Stdin: args, Profile: horizon.Render(e.Self, s.Floor)}}
}

// 2. Release, return or fault: accept the Horizon, whose node is the parent's
// frame, and settle that frame with this record. The frame's state is now this
// record's state.
func settleParent(s Instance, e Released) (Instance, []Effect) {
	if err := accept(s, e); err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	s.Self, s.Phase = e.Self, Done
	return s, []Effect{Settle{s.Stamp}, Exit{}}
}

// 3. Execute (invoke): the target ran exactly the committed surface, under a
// sandbox rendered from the Horizon. Its output is handed off unparsed. A
// target that failed or exited without output is faulted by this instance,
// which still holds custody.
func execute(s Instance, e Emitted) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, e.Err)
	}
	if len(e.Out) == 0 {
		return fault(s, "target exited without emitting")
	}
	s.Out, s.Phase = e.Out, Handoff
	return s, []Effect{Spawn{Under: s.Self, Out: e.Out}}
}

// 4. Handoff (invoke): the successor has committed. Issue its Release (check
// the stamp, narrow the Horizon), send it, and exit. The release is the last
// write.
func handoff(s Instance, e Stamped) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, "successor died before stamping: "+e.Err)
	}
	child, err := issueRelease(s.Self, s.Out, s.Stamp, e)
	if err != nil {
		return killAndFault(s, err.Error())
	}
	s.Child, s.Phase = e.Stamp, Done
	return s, []Effect{SendHorizon{Child: child, NewFrame: e.R.State == move.StateInvoke}, Exit{}}
}

// issueRelease is the parent's half of Release: check the successor's stamp,
// then compute its Horizon.
func issueRelease(parent horizon.Horizon, emitted []byte, parentStamp string, e Stamped) (horizon.Horizon, error) {
	if err := checkSuccessor(e.R, emitted, parentStamp); err != nil {
		return horizon.Horizon{}, fmt.Errorf("stamp failed verification: %w", err)
	}
	return Narrow(parent, *e.R, e.Stamp)
}

// checkSuccessor is the parent's check of a stamp (D3). The shell already
// refused any record that does not hash to its stamp. This checks that the
// successor read exactly the bytes that were emitted and linked to this frame.
func checkSuccessor(r *record.Record, emitted []byte, parentStamp string) error {
	switch {
	case r == nil:
		return errors.New("stamp names no valid record")
	case r.Input != record.Sum(emitted):
		return errors.New("record does not cover the bytes that were emitted")
	case r.Predecessor != parentStamp:
		return errors.New("record does not link to this parent")
	}
	return nil
}

// Narrow computes the child's Horizon from the parent's and the child's
// record. The child's reach is always a subset of the parent's.
func Narrow(parent horizon.Horizon, r record.Record, stamp string) (horizon.Horizon, error) {
	switch r.State {
	case move.StateReturn, move.StateFault:
		// Terminal: the child settles the parent's frame, and reaches nothing.
		return horizon.Horizon{Stamp: stamp, Reach: map[string]horizon.Surface{}, Settles: parent.Stamp}, nil
	case move.StateInvoke:
	default:
		return horizon.Horizon{}, fmt.Errorf("no Horizon for a Move in a state of %s", r.State)
	}
	s, ok := parent.Reach[r.Move.Invoke.Target]
	if !ok {
		return horizon.Horizon{}, fmt.Errorf("%q is not in the parent's reach", r.Move.Invoke.Target)
	}
	reach := map[string]horizon.Surface{}
	if !s.Leaf {
		reach = maps.Clone(parent.Reach) // at most the parent's
	}
	return horizon.Horizon{Stamp: stamp, Surface: &s, Reach: reach}, nil
}

// fault is the "maybe commit": this instance still holds custody, so it
// commits a fault and settles its own frame with it.
func fault(s Instance, reason string) (Instance, []Effect) {
	r := record.Record{Predecessor: s.Stamp, Input: record.Sum(nil), State: move.StateFault, Reason: reason, Supervisor: true}
	s.Phase = Done
	return s, []Effect{Commit{r}, Settle{record.Stamp(r)}, Exit{}}
}

func killAndFault(s Instance, reason string) (Instance, []Effect) {
	s, effs := fault(s, reason)
	return s, append([]Effect{KillSuccessor{}}, effs...)
}

// Committed returns the record an intake transition asks to commit.
func Committed(effs []Effect) record.Record {
	for _, e := range effs {
		if c, ok := e.(Commit); ok {
			return c.R
		}
	}
	return record.Record{}
}

// Package machine is the whole machine. An instance is a pure state machine:
//
//	Step(state, event) → (state, effects)
//
// The top-level decision is the state (D9). Intake derives it from the
// emitter's operands (thought 14), and each state is its own sequence of
// phases:
//
//	intake ─┬─ invoke  → release → execute → handoff              in custody: run the target, hand off its output
//	        ├─ return  → release → settle → gather → claim ─┐     discharge the slot, then maybe resume the waiter
//	        └─ fault   → release → settle → gather → claim ─┤
//	                                                        └─ resolve → execute → handoff   the waiter's continuation, resumed
//
// The phases:
//
//  1. Intake   receive bytes, interpret them as operands against the Horizon
//     they were emitted under, derive the state, commit a
//     content-addressed record. The record's sha256 is the stamp.
//  2. Release  the parent checks the stamp and sends the Horizon: which program
//     this instance runs, which targets it may call (Reach), and the
//     frame stack.
//  3. Execute  run the target inside a Seatbelt sandbox rendered from the
//     Horizon. .skyra/ is always denied.
//  4. Handoff  pass the target's output, unparsed, to a new successor; check its
//     stamp, open a slot for it if it is an invoke, release it, and exit
//     without waiting for it. An emission that is a JSON array is that many
//     instructions (fan-out, next/32): one successor each, every slot opened
//     before any successor is released.
//  5. Settle   a return or fault discharges the slot its reply names, once.
//  6. Gather   read what the waiting continuation is still waiting on. If
//     every slot of its round is discharged, commit a resolve record
//     naming every discharge (thought 13's join rule, by hashing).
//  7. Claim    the fan-in claim: exactly one sibling resumes the continuation,
//     in this same instance, as a new execute.
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
	"strings"

	"skyra/controller/core/record"
)

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
	case record.Invoke, record.Resolve:
		switch ev := e.(type) {
		case Released:
			if s.Phase == Release && s.State == record.Invoke {
				return release(s, ev)
			}
		case Emitted:
			if s.Phase == Execute {
				return execute(s, ev)
			}
		case Started:
			if s.Phase == Execute && s.Self.Surface != nil && s.Self.Surface.Resident {
				if ev.Err != "" {
					return fault(s, ev.Err)
				}
				return execute(s, Emitted{Out: ev.Out})
			}
		case Parked:
			if s.Phase == Execute && s.Self.Surface != nil && s.Self.Surface.Host != "" {
				return parked(s, ev)
			}
		case StrandRead:
			if s.Phase == Execute && s.State == record.Invoke && s.Self.Surface != nil && s.Self.Surface.System == "horizon" {
				return stranded(s, ev)
			}
		case Stamped:
			if s.Phase == Handoff {
				return handoff(s, ev)
			}
		}
	case record.Return, record.Fault:
		switch ev := e.(type) {
		case Released:
			if s.Phase == Release {
				return settle(s, ev)
			}
		case Settled:
			if s.Phase == Gather {
				return gather(s, ev)
			}
		case Claimed:
			if s.Phase == Claim {
				return claim(s, ev)
			}
		}
	}
	return stop(s, e)
}

func stop(s Instance, e Event) (Instance, []Effect) {
	msg := fmt.Sprintf("%T is not valid in phase %s", e, s.Phase)
	if s.State != "" {
		msg += fmt.Sprintf(" of an instance in a state of %s", s.State)
	}
	return s, []Effect{Exit{msg}}
}

// 1. Intake: interpret the bytes as operands against the Horizon they were
// emitted under, derive the state, and commit a record. A refused output is
// still committed, in a state of fault: a stamp means custody, not approval.
//
// The Horizon is read by id and verified here. If it does not verify there is
// no predecessor to commit against, so this instance refuses custody.
func intake(s Instance, e Received) (Instance, []Effect) {
	under, err := HorizonOf(e.Under)
	if err != nil {
		s.Phase = Done
		return s, []Effect{Exit{"the Horizon this input was emitted under: " + err.Error()}}
	}
	r := interpret(under, e.In)
	r.Predecessor, r.Input = under.Stamp, record.Sum(e.In)
	s.Under, s.R, s.Stamp, s.State, s.Phase = under, r, record.Stamp(r), r.State, Release
	return s, []Effect{StoreInput{e.In}, Commit{r}, SendStamp{s.Stamp}}
}

// 2. Release, invoke: accept the Horizon and run exactly the committed
// surface, under a sandbox rendered from the Horizon. A leaf gets its payload;
// a continuation gets its payload inside a continuation (no steps yet).
func release(s Instance, e Released) (Instance, []Effect) {
	self, err := accept(s, e)
	if err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	s.Self, s.Phase = self, Execute
	if self.Surface.System != "" {
		// A syscall: nothing runs; its result is derived. A strand is read
		// from the tree first.
		if name, args := syscall(s.R.Payload); name == Stranding {
			read, err := stranding(*self.Resume, args)
			if err != nil {
				return refuse(s, record.Horizon, err.Error())
			}
			return s, []Effect{read}
		}
		return execute(s, Emitted{Out: runFacet(s)})
	}
	if self.Surface.Host != "" {
		// An async surface: nothing runs here. The host takes custody and
		// resumes this instance, in Execute, with the event as the output.
		return s, []Effect{Park{Host: self.Surface.Host, Payload: s.R.Payload}}
	}
	stdin := []byte(s.R.Payload)
	if self.Surface.Continuation {
		stdin, _ = json.Marshal(Continuation{Actor: agentName(self), Agent: self.Agent, Agents: agents(self), WorkTools: workToolNames(self), Delegated: self.Delegated, Frame: s.Stamp, Caller: caller(self), Input: s.R.Payload, Contract: against(*self.Surface, s.R.Payload),
			Workspace: workspaceOf(self), Scratch: self.Authority.Scratch, Tools: tools(self), Contracts: contracts(self), Steps: []StepView{}})
	}
	socket := ""
	if self.Surface.Resident || self.Surface.Connection {
		var err error
		if socket, err = edgeSocket(self, s.Floor, s.Sockets); err != nil {
			return fault(s, err.Error())
		}
	}
	if self.Surface.Resident {
		return s, []Effect{StartResident{Run: runOf(self, stdin, s.Floor, ""), Socket: socket, Horizon: self, Floor: s.Floor}}
	}
	return s, []Effect{runOf(self, stdin, s.Floor, socket)}
}

// parked: the host holds custody now, so this instance exits. If it refused,
// this instance still holds custody and faults its own frame.
func parked(s Instance, e Parked) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, "host refused custody: "+e.Err)
	}
	s.Phase = Done
	return s, []Effect{Exit{}}
}

// 3. Execute: the target ran exactly the committed surface, under a sandbox
// rendered from the Horizon. Its output is handed off unparsed, one
// successor per instruction it carries. A target that failed or exited
// without output is faulted by this instance, which still holds custody.
func execute(s Instance, e Emitted) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, e.Err)
	}
	if len(e.Out) == 0 {
		return fault(s, "target exited without emitting")
	}
	outs, err := Instructions(e.Out)
	if err != nil {
		return fault(s, err.Error())
	}
	s.Out, s.Outs, s.Kids, s.Phase = e.Out, outs, nil, Handoff
	return s, []Effect{Spawn{Under: s.Self, Out: outs[0]}}
}

// Instructions splits an emission into the instructions it carries (next/32,
// decision 1). An emission that is a JSON array is that many instructions,
// each taken by its own successor as its exact bytes; anything else is one
// instruction, the emission itself, interpreted at intake as whatever it is.
// It is pure, so replay could split the same emission the same way.
func Instructions(out []byte) ([][]byte, error) {
	if t := strings.TrimSpace(string(out)); !strings.HasPrefix(t, "[") {
		return [][]byte{out}, nil
	}
	var parts []json.RawMessage
	if err := json.Unmarshal(out, &parts); err != nil {
		return [][]byte{out}, nil // not an array after all: one instruction, which intake faults
	}
	switch {
	case len(parts) == 0:
		return nil, errors.New("target emitted no instruction: an empty array")
	case len(parts) > MaxFanOut:
		return nil, fmt.Errorf("target emitted %d instructions; at most %d fan out at once", len(parts), MaxFanOut)
	}
	outs := make([][]byte, len(parts))
	for i, p := range parts {
		for _, earlier := range outs[:i] {
			if string(earlier) == string(p) {
				// Under one predecessor, the same bytes are the same record.
				return nil, errors.New("an emission carries the same instruction twice: they would be one record")
			}
		}
		outs[i] = p
	}
	return outs, nil
}

// 4. Handoff: a successor has committed. Check its stamp and compute its
// Horizon; spawn the next instruction's successor, until every one is
// stamped. Then open every new slot, and only then release every successor,
// so no child runs before the round's size is on disk and gather never sees
// a partial round. The releases are the last writes.
func handoff(s Instance, e Stamped) (Instance, []Effect) {
	if e.Err != "" {
		if len(s.Kids) == 0 {
			return fault(s, "successor died before stamping: "+e.Err)
		}
		return killAndFault(s, "successor died before stamping: "+e.Err)
	}
	fanned := len(s.Outs) > 1
	r, child, err := issueRelease(s.Self, s.Outs[len(s.Kids)], s.Stamp, e.Record, fanned)
	if err != nil {
		return killAndFault(s, err.Error())
	}
	s.Kids = append(s.Kids, Kid{Stamp: e.Record.Key, R: r, Horizon: child})
	if len(s.Kids) < len(s.Outs) {
		return s, []Effect{Spawn{Under: s.Self, Out: s.Outs[len(s.Kids)]}}
	}
	if fanned {
		if err := alone(s.Kids); err != nil {
			return killAndFault(s, err.Error())
		}
	}
	s.Child, s.Phase = s.Kids[0].Stamp, Done
	var effs []Effect
	for _, k := range s.Kids {
		if opensSlot(k) {
			effs = append(effs, StoreInput{s.Out}, PublishEmission{Round: s.Self.Stamp, Hash: record.Sum(s.Out)})
			break
		}
	}
	for i, k := range s.Kids {
		if opensSlot(k) {
			effs = append(effs, OpenDep{Round: s.Self.Stamp, Child: k.Stamp, Pos: i})
		}
	}
	for i, k := range s.Kids {
		effs = append(effs, SendHorizon{Child: k.Horizon, NewFrame: opensSlot(k), Index: i})
	}
	return s, append(effs, Exit{})
}

// opensSlot: the successor runs in a frame of its own, which waits in the
// emitter's round.
func opensSlot(k Kid) bool { return k.Horizon.Frame == k.Stamp }

// alone refuses what cannot share an emission with another instruction: a
// return discharges the emitter's own slot, which its siblings' round waits
// on, and an invoke that changes the Horizon of whoever waits (system.narrow)
// would change it once per sibling, each on its own copy.
func alone(kids []Kid) error {
	for _, k := range kids {
		switch {
		case k.R.State == record.Return:
			return errors.New("an emission of several instructions cannot return")
		case k.R.State == record.Invoke && k.Horizon.Surface != nil && changesWaiter(k.R, *k.Horizon.Surface):
			return fmt.Errorf("%s changes the Horizon of whoever waits on it, so it is emitted alone", k.R.To)
		}
	}
	return nil
}

// 5. Settle: accept the Horizon, whose first frame is the slot this record
// discharges, and discharge it. If a continuation waits on that slot, read
// what it still waits on.
func settle(s Instance, e Released) (Instance, []Effect) {
	self, err := accept(s, e)
	if err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	s.Self = self
	return discharged(s, Settle{s.Stamp})
}

func discharged(s Instance, effs ...Effect) (Instance, []Effect) {
	if s.Self.Resume == nil {
		s.Phase = Done // the operator waits on this slot; it reads the result itself
		return s, append(effs, Exit{})
	}
	s.Phase = Gather
	return s, append(effs, GatherWaiter{})
}

// 6. Gather: the waiting continuation resumes only when every slot its round
// opened is discharged (the join rule; wait for all). The resolve record names
// every discharge, so it cannot be formed early. Then the resolve is committed
// and the fan-in claim made: the winner resumes the waiter.
func gather(s Instance, e Settled) (Instance, []Effect) {
	waiter := s.Self.Resume
	frame, rounds, err := gathered(e)
	if err == nil && (e.Frame.Key != waiter.Frame || len(rounds) == 0 || rounds[len(rounds)-1].Stamp != waiter.Stamp) {
		err = errors.New("gathered rounds do not belong to the waiting continuation")
	}
	if err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	for _, c := range rounds[len(rounds)-1].Children {
		if !c.Discharged {
			s.Phase = Done // a sibling's discharge will be the last one
			return s, []Effect{Exit{}}
		}
	}
	return resolve(s, frame, rounds)
}

// gathered verifies every record the shell read for gather.
func gathered(e Settled) (record.Record, []round, error) {
	frame, err := RecordOf(e.Frame)
	if err != nil {
		return record.Record{}, nil, err
	}
	rounds := make([]round, len(e.Rounds))
	for i, r := range e.Rounds {
		rounds[i] = round{Emission: r.Emission, Stamp: r.Stamp}
		for _, c := range r.Children {
			inv, err := RecordOf(c.Invoke)
			if err != nil {
				return record.Record{}, nil, err
			}
			sl := slot{Stamp: c.Invoke.Key, Invoke: inv}
			if c.Discharge != nil {
				d, err := RecordOf(*c.Discharge)
				if err != nil {
					return record.Record{}, nil, err
				}
				sl.Result, sl.Discharge, sl.Discharged = c.Discharge.Key, d, true
			}
			rounds[i].Children = append(rounds[i].Children, sl)
		}
	}
	return frame, rounds, nil
}

// resolve commits the resolve record for the gathered round and makes the
// fan-in claim. The winner resumes the waiter.
func resolve(s Instance, frame record.Record, rounds []round) (Instance, []Effect) {
	waiter := s.Self.Resume
	round := rounds[len(rounds)-1]
	var results []string
	for _, c := range round.Children {
		results = append(results, c.Result)
	}
	obs, _ := json.Marshal(Observation{Frame: waiter.Frame, Round: round.Stamp, Discharges: results, Emission: round.Emission})
	r, _ := Resolved(s.Stamp, obs)
	s.Held, s.Phase = r, Claim
	s.Next = continuation(frame, rounds, *waiter)
	return s, []Effect{StoreInput{obs}, Commit{r}, TryClaim{Round: round.Stamp, Stamp: record.Stamp(r)}}
}

// 7. Claim: the winner resumes the waiting continuation in this instance: it
// holds custody of the resolve record, so it runs the continuation next.
func claim(s Instance, e Claimed) (Instance, []Effect) {
	if !e.Won {
		s.Phase = Done
		return s, []Effect{Exit{}}
	}
	resumed := *s.Self.Resume
	resumed.Stamp = record.Stamp(s.Held)
	if resumed.Surface == nil {
		s.Phase = Done // the operator's continuation: boot reads the claim and prints
		return s, []Effect{Exit{}}
	}
	s.State, s.R, s.Stamp, s.Self, s.Phase = record.Resolve, s.Held, resumed.Stamp, resumed, Execute
	return s, []Effect{Pop{}, runOf(resumed, s.Next, s.Floor, "")} // a continuation, never a connection surface
}

// stranded: the walk found the strand, which is the syscall's result, or
// refused it, which is a horizon fault of the syscall's frame.
func stranded(s Instance, e StrandRead) (Instance, []Effect) {
	if e.Refused != "" {
		return refuse(s, record.Horizon, "strand: "+e.Refused)
	}
	out, _ := json.Marshal(map[string]any{"to": s.Self.Frame, "payload": e.Strand})
	return execute(s, Emitted{Out: out})
}

// refuse is a syscall's refusal: a supervisor fault of the syscall's own
// frame, of the kind the syscall derived. Its observation says so, so replay
// derives the same kind (Observed).
func refuse(s Instance, kind, cause string) (Instance, []Effect) {
	return fault(s, string(Refusal(kind, cause)))
}

// Refusal is the observation of a syscall's refusal.
func Refusal(kind, cause string) []byte {
	b, _ := json.Marshal(map[string]any{"refused": record.FaultInfo{Kind: kind, Cause: cause}})
	return b
}

// fault is the "maybe commit": this instance still holds custody, so it
// records what it observed, discharges its own frame's slot with the fault that
// observation derives, and wakes whoever waits on that slot.
func fault(s Instance, observed string) (Instance, []Effect) {
	obs := []byte(observed)
	r := Observed(s.Stamp, obs)
	s.State, s.R, s.Stamp = record.Fault, r, record.Stamp(r)
	return discharged(s, StoreInput{obs}, Commit{r}, Settle{s.Stamp})
}

// Observed derives a supervisor's fault record from what it observed. The
// observation is stored as the record's input, so replay derives it again.
// What a supervisor observes is custody breaking, unless it is a syscall's
// refusal, which carries its kind. Every other observation begins with the
// supervisor's own words, never with JSON, so no target can forge one.
func Observed(stamp string, observation []byte) record.Record {
	info := record.FaultInfo{Kind: record.Custody, Cause: string(observation)}
	var r struct {
		Refused *record.FaultInfo `json:"refused"`
	}
	if json.Unmarshal(observation, &r) == nil && r.Refused != nil && r.Refused.Kind == record.Horizon &&
		string(Refusal(r.Refused.Kind, r.Refused.Cause)) == string(observation) {
		info = *r.Refused
	}
	return record.Record{Predecessor: stamp, Input: record.Sum(observation), State: record.Fault, Fault: &info, Supervisor: true}
}

func killAndFault(s Instance, reason string) (Instance, []Effect) {
	s, effs := fault(s, reason)
	return s, append([]Effect{KillSuccessor{}}, effs...)
}

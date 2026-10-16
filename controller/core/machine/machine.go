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
// Each phase is one or more cycles of Read* Commit? Execute* (41, V2): reads
// of the tree, which the core verifies itself, then one commit of write-once
// writes, then the acts that leave this process. Step makes every decision.
// Package instance performs the effects and reports what happened as the next
// event; it holds all the I/O and makes no decisions.
// This package is pure: no I/O, no clock, no randomness. It imports no package
// that touches the OS, which machine_test.go enforces. That purity is what lets
// replay re-derive a lineage from its recorded inputs.
package machine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Step is the decision tree. The phase says which event this instance is
// waiting for; anything else stops it.
func Step(s Instance, e Event) (Instance, []Effect) {
	switch ev := e.(type) {
	case Received:
		if s.Phase == Intake {
			return intake(s, ev)
		}
	case Committed:
		if err := integrity(ev); err != nil {
			return exit(s, err.Error())
		}
		switch s.Phase {
		case Stamping:
			s.Phase = Release
			return s, []Effect{Execute{Act: SendStamp{s.Stamp()}}}
		case Acknowledging:
			s.Phase = Acknowledged
			return s, []Effect{Execute{Act: Acknowledge{}}}
		case Spawning:
			s.Phase = Handoff
			return s, []Effect{Execute{Act: Spawn{Under: HeldHorizon(s.Self).Key, Out: s.Outs()[len(s.Kids)]}}}
		case Opening:
			return opened(s, ev)
		case Settling:
			return settled(s, ev)
		case Claim:
			return claim(s, ev)
		}
	case Executed:
		switch s.Phase {
		case Release:
			return arrived(s, ev)
		case Acknowledged:
			if s.State() == record.Invoke {
				return release(s)
			}
			return settle(s)
		case Running:
			if s.State() != record.Invoke && s.State() != record.Resolve {
				break // a terminal state never runs a target
			}
			return execute(s, ev)
		case Parking:
			return parked(s, ev)
		case Handoff:
			return stamped(s, ev)
		case Killing:
			return fault(s, s.Cause)
		}
	case ReadResult:
		if len(ev.Found) != len(s.Reads) {
			return exit(s, fmt.Sprintf("a read of %d keys was answered with %d", len(s.Reads), len(ev.Found)))
		}
		switch s.Phase {
		case Accepting:
			return accepting(s, ev)
		case Running:
			if s.State() == record.Invoke && s.Self.Surface != nil && s.Self.Surface.System == "horizon" {
				return stranded(s, ev)
			}
		case Verifying:
			return handoff(s, ev)
		case Gather:
			return walk(s, ev)
		}
	}
	return stop(s, e)
}

func stop(s Instance, e Event) (Instance, []Effect) {
	msg := fmt.Sprintf("%T is not valid in phase %s", e, s.Phase)
	if s.State() != "" {
		msg += fmt.Sprintf(" of an instance in a state of %s", s.State())
	}
	return exit(s, msg)
}

// exit ends this instance: with an error, or cleanly when err is "".
func exit(s Instance, err string) (Instance, []Effect) {
	s.Phase = Done
	return s, []Effect{Execute{Act: Exit{err}}}
}

// read asks the shell for keys; the answer comes back in the same order.
func read(s Instance, keys ...Key) (Instance, []Effect) {
	s.Reads = keys
	return s, []Effect{Read{Keys: keys}}
}

// integrity refuses a commit that found different bytes under a content
// address: corruption or a hash collision, never a race to resolve (V2).
func integrity(e Committed) error {
	for _, o := range e.Outcomes {
		if o == Conflict {
			return errors.New("integrity: a write found different bytes already under its name")
		}
	}
	return nil
}

func object(kind string, b []byte) Write { return ObjectWrite{Kind: kind, Bytes: b} }

func recordWrite(r record.Record) Write {
	b, _ := record.Encode(r)
	return object("records", b)
}

func horizonWrite(h horizon.Horizon) Write { return object("horizons", HeldHorizon(h).Raw) }

// 1. Intake: interpret the bytes as operands against the Horizon they were
// emitted under, derive the state, and commit a record. A refused output is
// still committed, in a state of fault: a stamp means custody, not approval.
//
// The Horizon is read by id and verified here. If it does not verify there is
// no predecessor to commit against, so this instance refuses custody.
func intake(s Instance, e Received) (Instance, []Effect) {
	under, err := HorizonOf(e.Under)
	if err != nil {
		return exit(s, "the Horizon this input was emitted under: "+err.Error())
	}
	r := Intaken(under, e.In)
	s.Under, s.R, s.Phase = under, r, Stamping
	return s, []Effect{Commit{Writes: []Write{object("inputs", e.In), recordWrite(r)}}}
}

// Intaken is the record intake derives from bytes emitted under a Horizon:
// the one derivation behind intake and replay. Replay holds the Horizon
// already, re-derived from genesis, so it calls this directly instead of
// handing Step bytes to verify again.
func Intaken(under horizon.Horizon, in []byte) record.Record {
	r := interpret(under, in)
	r.Predecessor, r.Input = under.Stamp, record.Sum(in)
	return r
}

// 2. Release: the parent answered the stamp with the id of this instance's
// Horizon and its frame stack. Read the Horizon, verify it, check it is what
// this commit derives and that the stack is whole, then commit the
// acceptance receipt and acknowledge custody.
func arrived(s Instance, e Executed) (Instance, []Effect) {
	if e.Err != "" {
		return exit(s, "no Horizon, cannot execute: "+e.Err)
	}
	s.Accepting, s.Arrived, s.Phase = string(e.Obs), e.Handles, Accepting
	return read(s, ObjectKey{Kind: "horizons", Hash: s.Accepting})
}

func accepting(s Instance, e ReadResult) (Instance, []Effect) {
	self, err := accept(s, Object{Key: s.Accepting, Raw: e.Found[0]})
	if err != nil {
		s.Phase = Done
		return s, []Effect{Execute{Act: Acknowledge{Err: err.Error()}}, Execute{Act: Exit{err.Error()}}}
	}
	s.Self, s.Phase = self, Acknowledging
	return s, []Effect{Commit{Writes: []Write{AcceptanceWrite{Stamp: s.Stamp(), Horizon: s.Accepting}}}}
}

// release, invoke: run exactly the committed surface, under a sandbox
// rendered from the Horizon. A leaf gets its payload; a continuation gets its
// payload inside a continuation (no steps yet).
func release(s Instance) (Instance, []Effect) {
	self := s.Self
	s.Phase = Running
	if self.Surface.System != "" {
		// A syscall: nothing runs; its result is derived. A strand is read
		// from the tree first.
		if name, args := syscall(s.R.Payload); name == Stranding {
			key, err := stranding(*self.Resume, args)
			if err != nil {
				return refuse(s, record.Horizon, err.Error())
			}
			return read(s, key)
		}
		return execute(s, Executed{Obs: runFacet(s)})
	}
	if self.Surface.Host != "" {
		// An async surface: nothing runs here. The host takes custody and
		// resumes this instance, running, with the event as the output.
		s.Phase = Parking
		return s, []Effect{Execute{Act: Park{Host: self.Surface.Host, Payload: s.R.Payload, Stamp: s.Stamp(), Frames: self.Depth()}, Force: true}}
	}
	stdin := []byte(s.R.Payload)
	if self.Surface.Continuation {
		stdin, _ = json.Marshal(Continuation{Actor: agentName(self), Agent: self.Agent, Agents: agents(self), WorkTools: workToolNames(self), Delegated: self.Delegated, Frame: s.Stamp(), Caller: caller(self), Input: s.R.Payload, Contract: against(*self.Surface, s.R.Payload),
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
		return s, []Effect{Execute{Act: StartResident{Run: runOf(self, stdin, s.Floor, ""), Socket: socket, Horizon: self, Floor: s.Floor}, Force: true}}
	}
	return s, []Effect{Execute{Act: runOf(self, stdin, s.Floor, socket), Force: true}}
}

// parked: the host holds custody now, so this instance exits. If it refused,
// this instance still holds custody and faults its own frame.
func parked(s Instance, e Executed) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, "host refused custody: "+e.Err)
	}
	return exit(s, "")
}

// 3. Execute: the target ran exactly the committed surface, under a sandbox
// rendered from the Horizon. Its output is handed off unparsed, one
// successor per instruction it carries. A target that failed or exited
// without output is faulted by this instance, which still holds custody.
func execute(s Instance, e Executed) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, e.Err)
	}
	if len(e.Obs) == 0 {
		return fault(s, "target exited without emitting")
	}
	if _, err := Instructions(e.Obs); err != nil {
		return fault(s, err.Error())
	}
	s.Out, s.Kids = e.Obs, nil
	return spawnNext(s)
}

// spawnNext commits the Horizon the next successor intakes under, then
// spawns it.
func spawnNext(s Instance) (Instance, []Effect) {
	s.Phase = Spawning
	return s, []Effect{Commit{Writes: []Write{horizonWrite(s.Self)}}}
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

// 4. Handoff: a successor has committed. Read the record under its stamp,
// check it, and compute its Horizon; spawn the next instruction's successor,
// until every one is stamped. Then commit the round (its emission and every
// new slot) and every successor's frame and Horizon, and only then release
// every successor, so no child runs before the round's size is on disk and
// gather never sees a partial round. The releases are the last acts.
func stamped(s Instance, e Executed) (Instance, []Effect) {
	if e.Err != "" {
		if len(s.Kids) == 0 {
			return fault(s, "successor died before stamping: "+e.Err)
		}
		return killAndFault(s, "successor died before stamping: "+e.Err)
	}
	s.Phase = Verifying
	return read(s, ObjectKey{Kind: "records", Hash: string(e.Obs)})
}

func handoff(s Instance, e ReadResult) (Instance, []Effect) {
	outs := s.Outs()
	fanned := len(outs) > 1
	stamp := s.Reads[0].(ObjectKey).Hash
	r, child, err := issueRelease(s.Self, outs[len(s.Kids)], s.Stamp(), Object{Key: stamp, Raw: e.Found[0]}, fanned)
	if err != nil {
		return killAndFault(s, err.Error())
	}
	s.Kids = append(s.Kids, Kid{Stamp: stamp, R: r, Horizon: child})
	if len(s.Kids) < len(outs) {
		return spawnNext(s)
	}
	if fanned {
		if err := alone(s.Kids); err != nil {
			return killAndFault(s, err.Error())
		}
	}
	s.Phase = Opening
	var writes []Write
	for _, k := range s.Kids {
		if opensSlot(k) {
			writes = append(writes, object("inputs", s.Out), EmissionWrite{Round: s.Self.Stamp, Hash: record.Sum(s.Out)})
			break
		}
	}
	for i, k := range s.Kids {
		if opensSlot(k) {
			writes = append(writes, SlotWrite{Round: s.Self.Stamp, Child: k.Stamp, Pos: i})
		}
	}
	for _, k := range s.Kids {
		if opensSlot(k) {
			writes = append(writes, FrameWrite{Stamp: k.Stamp})
		}
		writes = append(writes, horizonWrite(k.Horizon))
	}
	return s, []Effect{Commit{Writes: writes}}
}

// opened: the round is on disk. Every write in a frame is exclusive, so one
// that already existed means this round was opened before: stop.
func opened(s Instance, e Committed) (Instance, []Effect) {
	for _, o := range e.Outcomes {
		if o == Lost {
			return exit(s, "the round was already opened")
		}
	}
	s.Phase = Done
	var effs []Effect
	for i, k := range s.Kids {
		next := ""
		if opensSlot(k) {
			next = k.Stamp
		}
		effs = append(effs, Execute{Act: ReleaseChild{Index: i, Horizon: HeldHorizon(k.Horizon).Key, NewFrame: next, Stamp: k.Stamp}})
	}
	return s, append(effs, Execute{Act: Exit{}})
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

// 5. Settle: the Horizon's first frame is the slot this record discharges:
// discharge it, once. If a continuation waits on that slot, read what it
// still waits on.
func settle(s Instance) (Instance, []Effect) {
	s.Phase = Settling
	return s, []Effect{Commit{Writes: []Write{ResultWrite{Stamp: s.Stamp(), Frame: s.Self.Frame}}}}
}

// settled: the discharge is the last of the commit; it is exclusive.
func settled(s Instance, e Committed) (Instance, []Effect) {
	if len(e.Outcomes) == 0 || e.Outcomes[len(e.Outcomes)-1] == Lost {
		return exit(s, "frame already settled")
	}
	if s.Self.Resume == nil {
		return exit(s, "") // the operator waits on this slot; it reads the result itself
	}
	return gather(s)
}

// 6. Gather: the waiting continuation resumes only when every slot its round
// opened is discharged (the join rule; wait for all). The core walks the
// tree itself: the waiter's frame record, then each round from the latest
// back to the frame, following resolve records, and for each round its
// slots, their invokes and their discharges. The resolve record names every
// discharge, so it cannot be formed early. Then the resolve is committed and
// the fan-in claim made: the winner resumes the waiter.
type walked struct {
	Stage  int // 0: the frame and a round's slots; 1: its invokes and results; 2: its discharges; 3: the round before it
	Frame  Object
	Rounds []Round // oldest first; the round being read is not yet in it
	Cur    Round
	Slots  []string
}

func gather(s Instance) (Instance, []Effect) {
	waiter := s.Self.Resume
	s.Phase, s.Walk = Gather, &walked{Cur: Round{Stamp: waiter.Stamp}}
	return read(s, ObjectKey{Kind: "records", Hash: waiter.Frame}, SlotsKey{Round: waiter.Stamp}, EmissionKey{Frame: waiter.Frame, Round: waiter.Stamp})
}

func walk(s Instance, e ReadResult) (Instance, []Effect) {
	waiter, w := s.Self.Resume, *s.Walk
	s.Walk = &w
	switch w.Stage {
	case 0, 3:
		if w.Stage == 0 {
			w.Frame = Object{Key: waiter.Frame, Raw: e.Found[0]}
			e.Found = e.Found[1:]
		}
		if e.Found[0] == nil {
			return exit(s, fmt.Sprintf("round %s opened no slots", record.Short(w.Cur.Stamp)))
		}
		if err := json.Unmarshal(e.Found[0], &w.Slots); err != nil {
			return exit(s, "slots: "+err.Error())
		}
		w.Cur.Emission = string(e.Found[1])
		w.Stage = 1
		var keys []Key
		for _, c := range w.Slots {
			keys = append(keys, ObjectKey{Kind: "records", Hash: c}, ResultKey{Frame: c})
		}
		return read(s, keys...)
	case 1:
		var keys []Key
		w.Cur.Children = nil
		for i, c := range w.Slots {
			child := Child{Invoke: Object{Key: c, Raw: e.Found[2*i]}}
			if result := e.Found[2*i+1]; result != nil {
				child.Discharge = &Object{Key: string(result)}
				keys = append(keys, ObjectKey{Kind: "records", Hash: string(result)})
			}
			w.Cur.Children = append(w.Cur.Children, child)
		}
		w.Stage = 2
		if len(keys) > 0 {
			return read(s, keys...)
		}
		return walk(s, ReadResult{})
	case 2:
		i := 0
		for _, c := range w.Cur.Children {
			if c.Discharge != nil {
				c.Discharge.Raw = e.Found[i]
				i++
			}
		}
		w.Rounds = append([]Round{w.Cur}, w.Rounds...)
		if w.Cur.Stamp == waiter.Frame {
			s.Walk = nil
			return gathered(s, Settled{Frame: w.Frame, Rounds: w.Rounds})
		}
		w.Stage = -1 // reading the resolve that opened this round
		return read(s, ObjectKey{Kind: "records", Hash: w.Cur.Stamp})
	case -1:
		prev, err := RecordOf(Object{Key: w.Cur.Stamp, Raw: e.Found[0]})
		if err != nil {
			return exit(s, err.Error())
		}
		if prev.State != record.Resolve || prev.Frame != waiter.Frame {
			return exit(s, fmt.Sprintf("round %s is not a resolve of frame %s", record.Short(w.Cur.Stamp), record.Short(waiter.Frame)))
		}
		w.Cur, w.Stage = Round{Stamp: prev.Round}, 3
		return read(s, SlotsKey{Round: prev.Round}, EmissionKey{Frame: waiter.Frame, Round: prev.Round})
	}
	return exit(s, "gather lost its place")
}

// gathered: every record the walk read is verified, and if every slot of the
// waiter's round is discharged, the round resolves.
func gathered(s Instance, e Settled) (Instance, []Effect) {
	waiter := s.Self.Resume
	frame, rounds, err := verifyGathered(e)
	if err == nil && (e.Frame.Key != waiter.Frame || len(rounds) == 0 || rounds[len(rounds)-1].Stamp != waiter.Stamp) {
		err = errors.New("gathered rounds do not belong to the waiting continuation")
	}
	if err != nil {
		return exit(s, err.Error())
	}
	for _, c := range rounds[len(rounds)-1].Children {
		if !c.Discharged {
			return exit(s, "") // a sibling's discharge will be the last one
		}
	}
	return resolve(s, frame, rounds)
}

// verifyGathered verifies every record the shell read for gather.
func verifyGathered(e Settled) (record.Record, []round, error) {
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
// fan-in claim in the same commit. The winner resumes the waiter.
func resolve(s Instance, frame record.Record, rounds []round) (Instance, []Effect) {
	waiter := s.Self.Resume
	round := rounds[len(rounds)-1]
	var results []string
	for _, c := range round.Children {
		results = append(results, c.Result)
	}
	obs, _ := json.Marshal(Observation{Frame: waiter.Frame, Round: round.Stamp, Discharges: results, Emission: round.Emission})
	r, _ := Resolved(s.Stamp(), obs)
	s.Held, s.Phase = r, Claim
	s.Next = continuation(frame, rounds, *waiter)
	return s, []Effect{Commit{Writes: []Write{object("inputs", obs), recordWrite(r), ClaimWrite{Round: round.Stamp, Stamp: record.Stamp(r), Frame: waiter.Frame}}}}
}

// 7. Claim: the claim is the last write of the commit. The winner resumes the
// waiting continuation in this instance: it holds custody of the resolve
// record, so it runs the continuation next.
func claim(s Instance, e Committed) (Instance, []Effect) {
	if len(e.Outcomes) == 0 || e.Outcomes[len(e.Outcomes)-1] == Lost {
		return exit(s, "")
	}
	resumed := *s.Self.Resume
	resumed.Stamp = record.Stamp(s.Held)
	if resumed.Surface == nil {
		return exit(s, "") // the operator's continuation: boot reads the claim and prints
	}
	s.R, s.Self, s.Phase = s.Held, resumed, Running
	return s, []Effect{Execute{Act: Pop{}}, Execute{Act: runOf(resumed, s.Next, s.Floor, ""), Force: true}} // a continuation, never a connection surface
}

// stranded: the walk found the strand, which is the syscall's result, or
// refused it, which is a horizon fault of the syscall's frame.
func stranded(s Instance, e ReadResult) (Instance, []Effect) {
	var got struct {
		Strand  json.RawMessage `json:"strand"`
		Refused string          `json:"refused"`
	}
	if err := json.Unmarshal(e.Found[0], &got); err != nil {
		return exit(s, "strand: "+err.Error())
	}
	if got.Refused != "" {
		return refuse(s, record.Horizon, "strand: "+got.Refused)
	}
	out, _ := json.Marshal(map[string]any{"to": s.Self.Frame, "payload": got.Strand})
	return execute(s, Executed{Obs: out})
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
// records what it observed and discharges its own frame's slot with the
// fault that observation derives, in one commit; then it wakes whoever waits
// on that slot.
func fault(s Instance, observed string) (Instance, []Effect) {
	obs := []byte(observed)
	r := Observed(s.Stamp(), obs)
	s.R, s.Phase, s.Cause = r, Settling, ""
	return s, []Effect{Commit{Writes: []Write{object("inputs", obs), recordWrite(r), ResultWrite{Stamp: s.Stamp(), Frame: s.Self.Frame}}}}
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

// killAndFault kills every successor spawned so far, then faults.
func killAndFault(s Instance, reason string) (Instance, []Effect) {
	s.Phase, s.Cause = Killing, reason
	return s, []Effect{Execute{Act: KillSuccessors{}}}
}

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
//     without waiting for it.
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
	"unicode/utf8"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Instance is the state of one instance as it moves through its phases.
type Instance struct {
	Floor     string             // the .skyra path every rendered sandbox denies
	State     record.State       // the state intake derived: the top-level branch
	Phase     Phase              // the phase whose event this instance is waiting for
	Under     horizon.Horizon    // the Horizon the input was emitted under
	R         record.Record      // what this instance committed last
	Stamp     string             // hash of R
	Self      horizon.Horizon    // this instance's own Horizon, once released
	Out       []byte             // the target's emission, while it is handed off
	Child     string             // the successor's stamp, once verified and released
	Next      []byte             // a resumed continuation's input, while the claim is decided
	Held      record.Record      // the resolve record, while the claim is decided
	Resources []horizon.Resource // non-frame descriptors carried beside the stack
}

type Phase string

const (
	Intake  Phase = "intake"  // waiting for bytes
	Release Phase = "release" // committed and stamped; waiting for the Horizon
	Execute Phase = "execute" // the target is running; waiting for its output
	Handoff Phase = "handoff" // the successor is intaking the output; waiting for its stamp
	Gather  Phase = "gather"  // a slot was discharged; waiting to read what the waiter still waits on
	Claim   Phase = "claim"   // a resolve was committed; waiting for the fan-in claim
	Done    Phase = "done"
)

// Events are what the world reports to Step. Each phase accepts exactly one.
type Event interface{ event() }

type Received struct { // → Intake: the bytes, and the Horizon they were emitted under
	Under horizon.Horizon
	In    []byte
}
type Released struct { // → Release: this instance's Horizon, and how many frame fds came with it
	Self      horizon.Horizon
	Frames    int
	Resources []horizon.Resource
}
type Emitted struct { // → Execute: the target's one emission (empty: it exited without one)
	Out              []byte
	Err              string
	ReleasedResource string // an explicit close-connection surface released this incarnation
}
type Started struct { // → Execute: a resident and its connection entered custody
	Out      []byte
	Resource horizon.Resource
	Err      string
}
type Stamped struct { // → Handoff: the successor's stamp, and the record the tree holds under it (nil: none)
	Stamp string
	R     *record.Record
	Err   string
}
type Settled struct { // → Gather: the waiting frame's record and every round it opened, oldest first
	Frame  record.Record
	Rounds []Round
}
type Claimed struct{ Won bool }  // → Claim: whether this instance won the fan-in claim
type Parked struct{ Err string } // → Execute (async): whether the host took custody of the frame

// Round is one emission of a continuation and the slots it opened.
type Round struct {
	Stamp    string
	Children []Child
}

// Child is one slot: the invoke that opened it, and what discharged it (nil: still open).
type Child struct {
	Stamp     string
	Invoke    record.Record
	Result    string
	Discharge *record.Record
}

func (Received) event() {}
func (Released) event() {}
func (Emitted) event()  {}
func (Started) event()  {}
func (Stamped) event()  {}
func (Settled) event()  {}
func (Claimed) event()  {}
func (Parked) event()   {}

// Effects are what Step asks the shell to do, in order. SendStamp, Run,
// Spawn, GatherWaiter and TryClaim each end with the shell waiting for the
// next event.
type Effect interface{ effect() }

type StoreInput struct{ In []byte }   // record the exact input bytes before using them
type Commit struct{ R record.Record } // durable write: temp, flush, rename, flush directory
type SendStamp struct{ Stamp string } // then wait for Released
type Run struct {                     // then wait for Emitted
	Surface         horizon.Surface
	Stdin           []byte
	Profile         string
	Env             []string // the target's whole environment: nothing is inherited
	ConnectionOwner string   // optional: pass this owner's connection as fd 3
	CloseConnection bool
}
type StartResident struct {
	Run     Run
	Owner   string
	Horizon horizon.Horizon
	Floor   string
}
type Spawn struct { // start the successor, forward the bytes, then wait for Stamped
	Under horizon.Horizon
	Out   []byte
}
type KillSuccessor struct{}
type OpenDep struct{ Round, Child string } // open a slot in this instance's frame, before the child is released
type SendHorizon struct {                  // issue the successor's Release: its Horizon plus its frame stack
	Child    horizon.Horizon
	NewFrame bool // true: push a new frame for the child; false: hand on this instance's stack
}
type Settle struct{ Result string }         // discharge this instance's frame's slot, once
type GatherWaiter struct{}                  // read the waiting frame's rounds, then wait for Settled
type TryClaim struct{ Round, Stamp string } // the fan-in claim on the waiting frame, then wait for Claimed
type Pop struct{}                           // drop the discharged frame: the resumed continuation holds the rest
type Park struct {                          // hand this instance, its payload and its frame stack to an actor host, then wait for Parked
	Host    string
	Payload []byte
}
type Exit struct{ Err string }

func (StoreInput) effect()    {}
func (Commit) effect()        {}
func (SendStamp) effect()     {}
func (Run) effect()           {}
func (StartResident) effect() {}
func (Spawn) effect()         {}
func (KillSuccessor) effect() {}
func (OpenDep) effect()       {}
func (SendHorizon) effect()   {}
func (Settle) effect()        {}
func (GatherWaiter) effect()  {}
func (TryClaim) effect()      {}
func (Pop) effect()           {}
func (Park) effect()          {}
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
				s.Resources = append(s.Resources, ev.Resource)
				return execute(s, Emitted{Out: ev.Out})
			}
		case Parked:
			if s.Phase == Execute && s.Self.Surface != nil && s.Self.Surface.Host != "" {
				return parked(s, ev)
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
func intake(s Instance, e Received) (Instance, []Effect) {
	r := interpret(e.Under, e.In)
	r.Predecessor, r.Input = e.Under.Stamp, record.Sum(e.In)
	s.Under, s.R, s.Stamp, s.State, s.Phase = e.Under, r, record.Stamp(r), r.State, Release
	return s, []Effect{StoreInput{e.In}, Commit{r}, SendStamp{s.Stamp}}
}

// interpret derives the state from the emitter's operands (thought 14). An
// emitter supplies only an address and a payload:
//
//	{"to": "<address>", "payload": <JSON>}
//
// It never says whether that is a call or a return. The runtime derives it
// from where the emitter stands (improvements, items 1 and 2):
//
//	addresses the slot its frame discharges, payload fits its output contract   → return
//	names a capability in reach, payload fits its input contract                → invoke
//	names anything else                                                         → fault (horizon)
//	addresses nothing                                                           → fault (leaf)
//	not readable as operands                                                    → fault (malformed)
//	payload does not fit the contract of what it addresses                      → fault (contract)
//
// No name is reserved: the slot is the frame's id, a stamp no capability can
// be named. Every fault carries, as its payload, the observation of why it
// faulted. It is total and pure, so replay derives the same state again.
func interpret(h horizon.Horizon, out []byte) record.Record {
	if len(out) == 0 {
		return faulted(record.Custody, "no output", nil)
	}
	if !utf8.Valid(out) {
		return faulted(record.Malformed, "output is not UTF-8 text", nil)
	}
	printed := contract.Canonical(string(out))
	v, err := contract.Decode(out)
	o, isObject := v.(map[string]any)
	if _, addressed := o["to"]; err != nil || !isObject || !addressed {
		return faulted(record.Leaf, "output addresses nothing", printed)
	}
	to, isName := o["to"].(string)
	payload, hasPayload := o["payload"]
	if !isName || !hasPayload || len(o) != 2 {
		return faulted(record.Malformed, `addressed output must be exactly {"to": <address>, "payload": <JSON>}`, printed)
	}
	attempt := contract.Canonical(o)
	if h.Surface != nil && to == h.Frame {
		if err := contract.Fits(h.Surface.Output, payload); err != nil {
			return faulted(record.Contract, "return: "+err.Error(), attempt)
		}
		return record.Record{State: record.Return, Payload: contract.Canonical(payload)}
	}
	target, ok := h.Reach[to]
	if !ok {
		return faulted(record.Horizon, fmt.Sprintf("%q is not in reach", to), attempt)
	}
	if err := contract.Fits(target.Input, payload); err != nil {
		return faulted(record.Contract, to+": "+err.Error(), attempt)
	}
	return record.Record{State: record.Invoke, To: to, Payload: contract.Canonical(payload)}
}

// faulted is an intake fault. Its payload is the observation it was derived
// from, so the parent can pass it on (next/24, "Faults carry their observation").
func faulted(kind, cause string, observation json.RawMessage) record.Record {
	return record.Record{State: record.Fault, Payload: observation, Fault: &record.FaultInfo{Kind: kind, Cause: cause}}
}

// accept is the child's half of Release, for every state: the Horizon must be
// for this commit and come with its whole frame stack, or the instance stops.
func accept(s Instance, e Released) error {
	if e.Self.Stamp != s.Stamp {
		return errors.New("release is not for this commit")
	}
	if e.Frames != e.Self.Depth() {
		return fmt.Errorf("release carries %d frames, its Horizon needs %d", e.Frames, e.Self.Depth())
	}
	if len(e.Resources) > 64-e.Frames {
		return errors.New("too many custody resources")
	}
	return nil
}

// 2. Release, invoke: accept the Horizon and run exactly the committed
// surface, under a sandbox rendered from the Horizon. A leaf gets its payload;
// a continuation gets its payload inside a continuation (no steps yet).
func release(s Instance, e Released) (Instance, []Effect) {
	if err := accept(s, e); err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	s.Self, s.Phase = e.Self, Execute
	s.Resources = append([]horizon.Resource(nil), e.Resources...)
	if e.Self.Surface.Host != "" {
		// An async surface: nothing runs here. The host takes custody and
		// resumes this instance, in Execute, with the event as the output.
		return s, []Effect{Park{Host: e.Self.Surface.Host, Payload: s.R.Payload}}
	}
	stdin := []byte(s.R.Payload)
	if !e.Self.Surface.Leaf() {
		stdin, _ = json.Marshal(Continuation{Frame: s.Stamp, Input: s.R.Payload, Tools: tools(e.Self), Steps: []StepView{}})
	}
	if e.Self.Surface.Resident {
		owner := resourceOwner(e.Self)
		if owner == "" {
			return fault(s, "resident has no owning frame")
		}
		return s, []Effect{StartResident{Run: runOf(e.Self, stdin, s.Floor), Owner: owner, Horizon: e.Self, Floor: s.Floor}}
	}
	return s, []Effect{runOf(e.Self, stdin, s.Floor)}
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
// rendered from the Horizon. Its output is handed off unparsed. A target that
// failed or exited without output is faulted by this instance, which still
// holds custody.
func execute(s Instance, e Emitted) (Instance, []Effect) {
	if e.ReleasedResource != "" {
		found := false
		for i, r := range s.Resources {
			if r.Incarnation == e.ReleasedResource {
				s.Resources = append(s.Resources[:i:i], s.Resources[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fault(s, "released resource was not in custody")
		}
	}
	if e.Err != "" {
		return fault(s, e.Err)
	}
	if len(e.Out) == 0 {
		return fault(s, "target exited without emitting")
	}
	s.Out, s.Phase = e.Out, Handoff
	return s, []Effect{Spawn{Under: s.Self, Out: e.Out}}
}

// 4. Handoff: the successor has committed. Issue its Release (check the stamp,
// narrow the Horizon); if it is an invoke, open its slot first; send it, and
// exit. The release is the last write.
func handoff(s Instance, e Stamped) (Instance, []Effect) {
	if e.Err != "" {
		return fault(s, "successor died before stamping: "+e.Err)
	}
	child, err := issueRelease(s.Self, s.Out, s.Stamp, e)
	if err != nil {
		return killAndFault(s, err.Error())
	}
	s.Child, s.Phase = e.Stamp, Done
	if e.R.State == record.Invoke {
		return s, []Effect{OpenDep{Round: s.Self.Stamp, Child: e.Stamp}, SendHorizon{Child: child, NewFrame: true}, Exit{}}
	}
	return s, []Effect{SendHorizon{Child: child, NewFrame: false}, Exit{}}
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
// record. An invoke opens a new frame, waited on by the parent as it is now;
// the child is granted what its surface asks for, and only if the parent holds
// it. A return or fault reaches nothing: it discharges the parent's slot.
func Narrow(parent horizon.Horizon, r record.Record, stamp string) (horizon.Horizon, error) {
	switch r.State {
	case record.Return, record.Fault:
		return horizon.Horizon{Stamp: stamp, Reach: map[string]horizon.Surface{}, Frame: parent.Frame, Resume: parent.Resume}, nil
	case record.Invoke:
	default:
		return horizon.Horizon{}, fmt.Errorf("no Horizon for a record in a state of %s", r.State)
	}
	s, ok := parent.Reach[r.To]
	if !ok {
		return horizon.Horizon{}, fmt.Errorf("%q is not in the parent's reach", r.To)
	}
	reach := map[string]horizon.Surface{}
	for _, name := range s.Reach {
		if granted, ok := parent.Reach[name]; ok {
			if limit, narrowed := s.Delegate[name]; narrowed {
				if !granted.Authority.Allows(limit) {
					return horizon.Horizon{}, fmt.Errorf("%q delegates authority outside %q's grant", r.To, name)
				}
				granted.Authority = limit
			}
			reach[name] = granted
		}
	}
	waiter := parent
	return horizon.Horizon{Stamp: stamp, Surface: &s, Reach: reach, Frame: stamp, Resume: &waiter}, nil
}

// 5. Settle: accept the Horizon, whose first frame is the slot this record
// discharges, and discharge it. If a continuation waits on that slot, read
// what it still waits on.
func settle(s Instance, e Released) (Instance, []Effect) {
	if err := accept(s, e); err != nil {
		s.Phase = Done
		return s, []Effect{Exit{err.Error()}}
	}
	s.Self = e.Self
	s.Resources = append([]horizon.Resource(nil), e.Resources...)
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
// every discharge, so it cannot be formed early.
func gather(s Instance, e Settled) (Instance, []Effect) {
	waiter := s.Self.Resume
	if record.Stamp(e.Frame) != waiter.Frame || len(e.Rounds) == 0 || e.Rounds[len(e.Rounds)-1].Stamp != waiter.Stamp {
		s.Phase = Done
		return s, []Effect{Exit{"gathered rounds do not belong to the waiting continuation"}}
	}
	round := e.Rounds[len(e.Rounds)-1]
	var results []string
	for _, c := range round.Children {
		if c.Discharge == nil {
			s.Phase = Done // a sibling's discharge will be the last one
			return s, []Effect{Exit{}}
		}
		results = append(results, c.Result)
	}
	obs, _ := json.Marshal(Observation{Frame: waiter.Frame, Round: round.Stamp, Discharges: results})
	r, _ := Resolved(s.Stamp, obs)
	s.Held, s.Next, s.Phase = r, continuation(e, tools(*waiter)), Claim
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
	return s, []Effect{Pop{}, runOf(resumed, s.Next, s.Floor)}
}

// Observation is what a resolve observed: the round whose slots are all
// discharged, and the discharges, in slot order.
type Observation struct {
	Frame      string   `json:"frame"`
	Round      string   `json:"round"`
	Discharges []string `json:"discharges"`
}

// Resolved derives a resolve record from its observation. It is stored as the
// record's input, so replay derives it again.
func Resolved(predecessor string, obs []byte) (record.Record, error) {
	var o Observation
	if err := json.Unmarshal(obs, &o); err != nil {
		return record.Record{}, err
	}
	return record.Record{Predecessor: predecessor, Input: record.Sum(obs), State: record.Resolve,
		Frame: o.Frame, Round: o.Round, Deps: o.Discharges}, nil
}

// runOf is the Run effect for a Horizon: its surface, under the sandbox the
// Horizon renders to, with an environment that names only the sockets its
// surface was granted, and the slot it returns to.
func runOf(h horizon.Horizon, stdin []byte, floor string) Run {
	// A target returns by addressing the slot its frame discharges. Being told
	// it is information, not authority: intake derives the return from the
	// emitter's own Horizon, whatever the target believes.
	env := []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=" + h.Frame}
	for _, sock := range h.Surface.Authority.Connect {
		env = append(env, "SKYRA_CONNECT="+sock)
	}
	// A target is told where its read grants are, as PATH lists directories.
	// This is information, not authority: the rendered profile enforces them.
	if reads := h.Surface.Authority.Read; len(reads) > 0 {
		paths := reads[0].Path
		for _, grant := range reads[1:] {
			paths += ":" + grant.Path
		}
		env = append(env, "SKYRA_READ="+paths)
	}
	r := Run{Surface: *h.Surface, Stdin: stdin, Profile: horizon.Render(h, floor), Env: env}
	if h.Surface.Connection {
		r.ConnectionOwner = resourceOwner(h)
		r.CloseConnection = h.Surface.CloseConnection
	}
	return r
}

func resourceOwner(h horizon.Horizon) string {
	depth := h.Surface.ResourceOwnerDepth
	if depth == 0 {
		depth = 1
	}
	for i := 0; i < depth; i++ {
		if h.Resume == nil {
			return ""
		}
		h = *h.Resume
	}
	return h.Frame
}

// tools is what a continuation is shown: the input contract of everything in
// its reach. What it sees is exactly what it may invoke (frame ⊆ authority).
func tools(h horizon.Horizon) map[string]*contract.Schema {
	t := map[string]*contract.Schema{}
	for name, s := range h.Reach {
		t[name] = s.Input
	}
	return t
}

// Continuation is what a continuation reads on stdin: the frame it runs in,
// the payload it was invoked with, the tools it may invoke, and every step it
// has taken, with what came back. It is stateless: everything it knows is here.
type Continuation struct {
	Frame string                      `json:"frame"`
	Input json.RawMessage             `json:"input"`
	Tools map[string]*contract.Schema `json:"tools"`
	Steps []StepView                  `json:"steps"`
}

// StepView is one slot as the continuation sees it: what it invoked, and what
// discharged the slot.
type StepView struct {
	To      string            `json:"to"`
	Payload json.RawMessage   `json:"payload"`
	State   record.State      `json:"state"`
	Value   json.RawMessage   `json:"value,omitempty"`
	Fault   *record.FaultInfo `json:"fault,omitempty"`
}

func continuation(e Settled, tools map[string]*contract.Schema) []byte {
	c := Continuation{Frame: record.Stamp(e.Frame), Input: e.Frame.Payload, Tools: tools, Steps: []StepView{}}
	for _, round := range e.Rounds {
		for _, child := range round.Children {
			c.Steps = append(c.Steps, StepView{To: child.Invoke.To, Payload: child.Invoke.Payload,
				State: child.Discharge.State, Value: child.Discharge.Payload, Fault: child.Discharge.Fault})
		}
	}
	b, _ := json.Marshal(c)
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
func Observed(stamp string, observation []byte) record.Record {
	return record.Record{Predecessor: stamp, Input: record.Sum(observation), State: record.Fault,
		Fault: &record.FaultInfo{Kind: record.Custody, Cause: string(observation)}, Supervisor: true}
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

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
	Outs      [][]byte           // the instructions it carries, one per successor (fan-out: several)
	Kids      []Kid              // the successors stamped and verified so far, in instruction order
	Child     string             // the first successor's stamp, once every one is verified and released
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

// Kid is a successor this instance supervises during handoff: its stamp, its
// record, and the Horizon it will be released with.
type Kid struct {
	Stamp   string
	R       record.Record
	Horizon horizon.Horizon
}

// MaxFanOut bounds the instructions one emission may carry: every successor
// holds a slot, a process and its frame descriptors until it is released.
const MaxFanOut = 32

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
type StrandRead struct {         // → Execute (strand): the strand, or why the walk refused it
	Strand  json.RawMessage
	Refused string
}

// Round is one emission of a continuation and the slots it opened.
type Round struct {
	Emission string // hash of the complete emission that opened these slots
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

func (Received) event()   {}
func (Released) event()   {}
func (Emitted) event()    {}
func (Started) event()    {}
func (Stamped) event()    {}
func (Settled) event()    {}
func (Claimed) event()    {}
func (Parked) event()     {}
func (StrandRead) event() {}

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
type KillSuccessor struct{} // kill every successor spawned so far
type PublishEmission struct{ Round, Hash string }

func (PublishEmission) effect() {}

type OpenDep struct { // open a slot in this instance's frame, before any child is released
	Round, Child string
	Pos          int // the instruction's position in the emission: a round's steps come back in this order
}
type SendHorizon struct { // issue a successor's Release: its Horizon plus its frame stack
	Child    horizon.Horizon
	NewFrame bool // true: push a new frame for the child; false: hand on this instance's stack
	Index    int  // which successor, in instruction order
}
type Settle struct{ Result string }         // discharge this instance's frame's slot, once
type GatherWaiter struct{}                  // read the waiting frame's rounds, then wait for Settled
type TryClaim struct{ Round, Stamp string } // the fan-in claim on the waiting frame, then wait for Claimed
type ReadStrand struct {                    // walk the tree from an observation up to the invoke made through a contract, then wait for StrandRead
	Observation  string
	Invocation   string // tree mode: retained root, with a nonempty accepted frontier
	Observations []string
	Target       string
	Schema       *contract.Schema
	Origin       string // the requester's Horizon: the walk stays inside it
	Room         string // the enclosing genesis; different boots are not one Horizon
}
type Pop struct{}  // drop the discharged frame: the resumed continuation holds the rest
type Park struct { // hand this instance, its payload and its frame stack to an actor host, then wait for Parked
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
func (ReadStrand) effect()    {}
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
//	addresses the slot its frame discharges                                     → return
//	names a capability in reach, payload fits exactly one of its contracts      → invoke
//	names anything else                                                         → fault (horizon)
//	addresses nothing                                                           → fault (leaf)
//	not readable as operands                                                    → fault (malformed)
//	payload fits none of the contracts of what it addresses, or more than one   → fault (contract)
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
		// Any JSON value: a return declares no contract.
		return record.Record{State: record.Return, Payload: contract.Canonical(payload)}
	}
	target, ok := h.Reach[to]
	if g, isRef := h.Refs[to]; isRef && !ok {
		target, ok = h.Reach[g.Actor]
	}
	if !ok {
		return faulted(record.Horizon, fmt.Sprintf("%q is not in reach", to), attempt)
	}
	if _, err := Selected(target, payload); err != nil {
		return faulted(record.Contract, to+": "+err.Error(), attempt)
	}
	if target.System == "horizon" {
		if kind, err := checkSyscall(h, contract.Canonical(payload)); err != nil {
			return faulted(kind, to+": "+err.Error(), attempt)
		}
	}
	return record.Record{State: record.Invoke, To: to, Payload: contract.Canonical(payload)}
}

// Selected is the one contract of a target that a payload fits (next/32,
// decisions 20, 27). A contract is (target, schema): the invoke names the
// target, and fit picks the schema. None fits, or more than one does, is a
// contract fault; the kernel never reasons about whether two schemas overlap,
// it refuses when an ambiguity actually occurs.
func Selected(target horizon.Surface, payload any) (*contract.Schema, error) {
	var fit []*contract.Schema
	var first error
	closest := false // the error kept is from a contract whose required fields the payload has
	for _, c := range target.All() {
		if err := contract.Fits(c, payload); err != nil {
			if near := hasRequired(c, payload); first == nil || near && !closest {
				first, closest = err, near
			}
			continue
		}
		fit = append(fit, c)
	}
	switch {
	case len(fit) == 1:
		return fit[0], nil
	case len(fit) > 1:
		return nil, fmt.Errorf("ambiguous: the payload fits %d of its contracts", len(fit))
	case first == nil:
		return nil, errors.New("no contract")
	}
	return nil, first
}

// faulted is an intake fault. Its payload is the observation it was derived
// from, so the parent can pass it on (next/24, "Faults carry their observation").
func faulted(kind, cause string, observation json.RawMessage) record.Record {
	return record.Record{State: record.Fault, Payload: observation, Fault: &record.FaultInfo{Kind: kind, Cause: cause}}
}

// accept is the child's half of Release, for every state: the Horizon it
// looked up must be for this commit, must be exactly what this commit derives
// from the Horizon it was emitted under, and must come with its whole frame
// stack, or the instance stops. Authority is derived, never taken on trust
// (next/24): the id is only a handle.
func accept(s Instance, e Released) error {
	if e.Self.Stamp != s.Stamp {
		return errors.New("release is not for this commit")
	}
	want, err := Narrow(s.Under, s.R, s.Stamp)
	if err != nil {
		return fmt.Errorf("release cannot be derived: %w", err)
	}
	if a, b := encode(want), encode(e.Self); a != b {
		// One instruction of several: an invoke is forked, a fault holds a
		// slot of its own.
		if alt, _ := Fanned(s.Under, s.R, s.Stamp); encode(alt) != b {
			return errors.New("release is not what this commit derives")
		}
	}
	if e.Frames != e.Self.Depth() {
		return fmt.Errorf("release carries %d frames, its Horizon needs %d", e.Frames, e.Self.Depth())
	}
	if len(e.Resources) > 64-e.Frames {
		return errors.New("too many custody resources")
	}
	return nil
}

func encode(h horizon.Horizon) string {
	b, _ := json.Marshal(h)
	return string(b)
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
	if e.Self.Surface.System != "" {
		// A syscall: nothing runs; its result is derived. A strand is read
		// from the tree first.
		if name, args := syscall(s.R.Payload); name == Stranding {
			read, err := stranding(*e.Self.Resume, args)
			if err != nil {
				return refuse(s, record.Horizon, err.Error())
			}
			return s, []Effect{read}
		}
		return execute(s, Emitted{Out: runFacet(s)})
	}
	if e.Self.Surface.Host != "" {
		// An async surface: nothing runs here. The host takes custody and
		// resumes this instance, in Execute, with the event as the output.
		return s, []Effect{Park{Host: e.Self.Surface.Host, Payload: s.R.Payload}}
	}
	stdin := []byte(s.R.Payload)
	if e.Self.Surface.Continuation {
		stdin, _ = json.Marshal(Continuation{Frame: s.Stamp, Caller: caller(e.Self), Input: s.R.Payload, Contract: against(*e.Self.Surface, s.R.Payload),
			Tools: tools(e.Self), Contracts: contracts(e.Self), Steps: []StepView{}})
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
// rendered from the Horizon. Its output is handed off unparsed, one
// successor per instruction it carries. A target that failed or exited
// without output is faulted by this instance, which still holds custody.
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
	outs, err := Instructions(e.Out)
	if err != nil {
		return fault(s, err.Error())
	}
	if len(outs) > 1 && len(s.Resources) > 0 {
		// Custody of a connection passes to exactly one successor.
		return fault(s, "an emission of several instructions cannot carry custody of a resource")
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
	child, err := issueRelease(s.Self, s.Outs[len(s.Kids)], s.Stamp, e, fanned)
	if err != nil {
		return killAndFault(s, err.Error())
	}
	s.Kids = append(s.Kids, Kid{Stamp: e.Stamp, R: *e.R, Horizon: child})
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

// issueRelease is the parent's half of Release: check the successor's stamp,
// then compute its Horizon.
func issueRelease(parent horizon.Horizon, emitted []byte, parentStamp string, e Stamped, fanned bool) (horizon.Horizon, error) {
	if err := checkSuccessor(e.R, emitted, parentStamp); err != nil {
		return horizon.Horizon{}, fmt.Errorf("stamp failed verification: %w", err)
	}
	if fanned {
		return Fanned(parent, *e.R, e.Stamp)
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
// record. Every call shares its caller's Horizon (next/24): the child runs
// with the same reach, references and grants. Nothing about the callee
// narrows them; a narrowing is requested, through the room (next/26):
//
//	invoke system, narrow        the caller's Horizon, and the waiting caller is
//	                             recorded as holding a new reference: the grant
//	                             its payload derives, named system.ref.<n>
//	invoke system, extend-reach  the caller's Horizon, and the waiting caller's
//	                             reach entry for itself holds one more contract,
//	                             in every frame below that shares its Horizon
//	invoke a reference           the actor, under the narrowed Horizon it names
//
// An invoke opens a new frame for the child, waited on by the parent as it is
// now. A return or fault runs nothing: it discharges the parent's slot.
func Narrow(parent horizon.Horizon, r record.Record, stamp string) (horizon.Horizon, error) {
	switch r.State {
	case record.Return, record.Fault:
		return horizon.Horizon{Stamp: stamp, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority, Frame: parent.Frame, Resume: parent.Resume}, nil
	case record.Invoke:
	default:
		return horizon.Horizon{}, fmt.Errorf("no Horizon for a record in a state of %s", r.State)
	}
	waiter := parent
	if g, isRef := parent.Refs[r.To]; isRef {
		if _, named := parent.Reach[r.To]; !named {
			s, ok := parent.Reach[g.Actor]
			if !ok {
				return horizon.Horizon{}, fmt.Errorf("reference %q names %q, which is not in the parent's reach", r.To, g.Actor)
			}
			// A Horizon of its own: created by this invoke, so named by it.
			return horizon.Horizon{Stamp: stamp, Surface: &s, Reach: kept(parent.Reach, g.Reach), Authority: g.Authority,
				Name: g.Actor, Origin: stamp, Frame: stamp, Resume: &waiter}, nil
		}
	}
	s, ok := parent.Reach[r.To]
	if !ok {
		return horizon.Horizon{}, fmt.Errorf("%q is not in the parent's reach", r.To)
	}
	if s.System == "horizon" {
		switch name, args := syscall(r.Payload); name {
		case Narrowing:
			g, err := narrowing(parent, args)
			if err != nil {
				return horizon.Horizon{}, err
			}
			refs := map[string]horizon.Grant{refName(parent.Refs): g}
			for name, held := range parent.Refs {
				refs[name] = held
			}
			waiter.Refs = refs
		case Extending:
			c, err := extension(parent, args)
			if err != nil {
				return horizon.Horizon{}, err
			}
			waiter = extended(parent, parent.Origin, parent.Name, c)
		}
	}
	return horizon.Horizon{Stamp: stamp, Surface: &s, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority,
		Name: r.To, Origin: parent.Origin, Forked: parent.Forked, Frame: stamp, Resume: &waiter}, nil
}

// Fanned is Narrow for a successor that took one instruction of several
// (next/32, H2). An invoke opens its frame as always. A fault holds a slot of
// its own in the emitter's round and discharges it, so its siblings still run
// and the emitter resumes with the fault among its steps; alone, the same
// fault discharges the emitter's own slot.
func Fanned(parent horizon.Horizon, r record.Record, stamp string) (horizon.Horizon, error) {
	if r.State != record.Fault {
		h, err := Narrow(parent, r, stamp)
		if r.State == record.Invoke {
			h.Forked = true // its siblings share this Horizon, each on its own copy
		}
		return h, err
	}
	waiter := parent
	return horizon.Horizon{Stamp: stamp, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority, Frame: stamp, Resume: &waiter}, nil
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
// every discharge, so it cannot be formed early. Then the resolve is committed
// and the fan-in claim made: the winner resumes the waiter.
func gather(s Instance, e Settled) (Instance, []Effect) {
	waiter := s.Self.Resume
	if record.Stamp(e.Frame) != waiter.Frame || len(e.Rounds) == 0 || e.Rounds[len(e.Rounds)-1].Stamp != waiter.Stamp {
		s.Phase = Done
		return s, []Effect{Exit{"gathered rounds do not belong to the waiting continuation"}}
	}
	for _, c := range e.Rounds[len(e.Rounds)-1].Children {
		if c.Discharge == nil {
			s.Phase = Done // a sibling's discharge will be the last one
			return s, []Effect{Exit{}}
		}
	}
	return resolve(s, e)
}

// resolve commits the resolve record for the gathered round and makes the
// fan-in claim. The winner resumes the waiter.
func resolve(s Instance, e Settled) (Instance, []Effect) {
	waiter := s.Self.Resume
	round := e.Rounds[len(e.Rounds)-1]
	var results []string
	for _, c := range round.Children {
		results = append(results, c.Result)
	}
	obs, _ := json.Marshal(Observation{Frame: waiter.Frame, Round: round.Stamp, Discharges: results, Emission: round.Emission})
	r, _ := Resolved(s.Stamp, obs)
	s.Held, s.Phase = r, Claim
	s.Next = continuation(e, *waiter)
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
	Emission   string   `json:"emission,omitempty"`
	Frame      string   `json:"frame"`
	Round      string   `json:"round"`
	Discharges []string `json:"discharges"`
}

// hasRequired reports whether a payload is an object holding every field a
// contract requires: the contract it was most likely meant for.
func hasRequired(c *contract.Schema, payload any) bool {
	o, ok := payload.(map[string]any)
	if !ok || c.Type != "object" {
		return false
	}
	for _, r := range c.Required {
		if _, ok := o[r]; !ok {
			return false
		}
	}
	return true
}

// Resolved derives a resolve record from its observation. It is stored as the
// record's input, so replay derives it again.
func Resolved(predecessor string, obs []byte) (record.Record, error) {
	var o Observation
	if err := json.Unmarshal(obs, &o); err != nil {
		return record.Record{}, err
	}
	return record.Record{Predecessor: predecessor, Input: record.Sum(obs), State: record.Resolve,
		Frame: o.Frame, Round: o.Round, Deps: o.Discharges, Emission: o.Emission}, nil
}

// runOf is the Run effect for a Horizon: its surface, under the sandbox the
// Horizon renders to, with an environment that names only the sockets its
// surface was granted, and the slot it returns to.
func runOf(h horizon.Horizon, stdin []byte, floor string) Run {
	// A target returns by addressing the slot its frame discharges. Being told
	// it is information, not authority: intake derives the return from the
	// emitter's own Horizon, whatever the target believes.
	env := []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=" + h.Frame}
	for _, sock := range h.Authority.Connect {
		env = append(env, "SKYRA_CONNECT="+sock)
	}
	// A target is told where its read grants are, as PATH lists directories.
	// This is information, not authority: the rendered profile enforces them.
	if reads := h.Authority.Read; len(reads) > 0 {
		paths := reads[0].Path
		for _, grant := range reads[1:] {
			paths += ":" + grant.Path
		}
		env = append(env, "SKYRA_READ="+paths)
	}
	if s := h.Authority.Scratch; s != "" {
		env = append(env, "SKYRA_SCRATCH="+s)
	}
	if dir := horizon.StateDir(h, floor); dir != "" {
		env = append(env, "SKYRA_STATE="+dir) // the actor's own state (next/32): the profile grants it
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

// tools is what a continuation is shown: the first contract of everything in
// its reach. What it sees is exactly what it may invoke (frame ⊆ authority).
// Which of them a model is offered is the adapter's choice, not the runtime's.
func tools(h horizon.Horizon) map[string]*contract.Schema {
	t := map[string]*contract.Schema{}
	for name, s := range shownReach(h) {
		if all := s.All(); len(all) > 0 {
			t[name] = all[0]
		}
	}
	return t
}

// contracts is every contract of each name in reach that holds more than
// one, so a continuation sees all of them (next/32, H3).
func contracts(h horizon.Horizon) map[string][]*contract.Schema {
	var c map[string][]*contract.Schema
	for name, s := range shownReach(h) {
		if all := s.All(); len(all) > 1 {
			if c == nil {
				c = map[string][]*contract.Schema{}
			}
			c[name] = all
		}
	}
	return c
}

// shownReach is a Horizon's reach as a continuation sees it: every name,
// and every reference it holds, as the actor it stands for.
func shownReach(h horizon.Horizon) map[string]horizon.Surface {
	r := map[string]horizon.Surface{}
	for name, s := range h.Reach {
		r[name] = s
	}
	for ref, g := range h.Refs {
		if s, ok := h.Reach[g.Actor]; ok {
			r[ref] = s
		}
	}
	return r
}

// against is the contract a program is executed against, named in its
// envelope when its surface has more than one: the one its input fits.
func against(s horizon.Surface, input json.RawMessage) *contract.Schema {
	if len(s.All()) < 2 {
		return nil
	}
	v, err := contract.Decode(input)
	if err != nil {
		return nil
	}
	c, _ := Selected(s, v)
	return c
}

// Continuation is what a continuation reads on stdin: the frame it runs in,
// who invoked it, the payload it was invoked with, the contract it is executed
// against (when it has more than one), the tools it may invoke, and every step
// it has taken, with what came back. It is stateless: everything it knows is here.
type Continuation struct {
	Frame     string                        `json:"frame"`
	Caller    string                        `json:"caller,omitempty"` // the name its invoker was invoked through; "" for the operator
	Input     json.RawMessage               `json:"input"`
	Contract  *contract.Schema              `json:"contract,omitempty"`
	Tools     map[string]*contract.Schema   `json:"tools"`
	Contracts map[string][]*contract.Schema `json:"contracts,omitempty"`
	Steps     []StepView                    `json:"steps"`
}

// StepView is one slot as the continuation sees it: what it invoked, and what
// discharged the slot. Observation is the discharge's stamp: a worker passes
// an observation on by it (next/21), and remembers by it (next/32). It is a
// hash, so an adapter never shows it to a model.
type StepView struct {
	To          string            `json:"to"`
	Payload     json.RawMessage   `json:"payload"`
	State       record.State      `json:"state"`
	Value       json.RawMessage   `json:"value,omitempty"`
	Fault       *record.FaultInfo `json:"fault,omitempty"`
	Observation string            `json:"observation,omitempty"`
}

// caller is who invoked the frame h runs in: the name its invoker was itself
// invoked through, derived from the Horizon, so a program cannot claim another.
// Two actors name the edge between them by it (proposal 38). The operator has
// no name.
func caller(h horizon.Horizon) string {
	if h.Resume == nil {
		return ""
	}
	return h.Resume.Name
}

// continuation is what the waiter reads when it resumes: every step of every
// round, in the order they were emitted.
func continuation(e Settled, waiter horizon.Horizon) []byte {
	c := Continuation{Frame: record.Stamp(e.Frame), Caller: caller(waiter), Input: e.Frame.Payload, Tools: tools(waiter), Contracts: contracts(waiter), Steps: []StepView{}}
	if waiter.Surface != nil {
		c.Contract = against(*waiter.Surface, e.Frame.Payload)
	}
	for _, round := range e.Rounds {
		for _, child := range round.Children {
			c.Steps = append(c.Steps, StepView{To: child.Invoke.To, Payload: child.Invoke.Payload,
				State: child.Discharge.State, Value: child.Discharge.Payload, Fault: child.Discharge.Fault, Observation: child.Result})
		}
	}
	b, _ := json.Marshal(c)
	return b
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

// Committed returns the record an intake transition asks to commit.
func Committed(effs []Effect) record.Record {
	for _, e := range effs {
		if c, ok := e.(Commit); ok {
			return c.R
		}
	}
	return record.Record{}
}

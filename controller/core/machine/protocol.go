package machine

import (
	"encoding/json"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Instance is the state of one instance as it moves through its phases.
type Instance struct {
	Floor   string          // the .skyra path every rendered sandbox denies
	State   record.State    // the state intake derived: the top-level branch
	Phase   Phase           // the phase whose event this instance is waiting for
	Under   horizon.Horizon // the Horizon the input was emitted under
	R       record.Record   // what this instance committed last
	Stamp   string          // hash of R
	Self    horizon.Horizon // this instance's own Horizon, once released
	Out     []byte          // the target's emission, while it is handed off
	Outs    [][]byte        // the instructions it carries, one per successor (fan-out: several)
	Kids    []Kid           // the successors stamped and verified so far, in instruction order
	Child   string          // the first successor's stamp, once every one is verified and released
	Next    []byte          // a resumed continuation's input, while the claim is decided
	Held    record.Record   // the resolve record, while the claim is decided
	Sockets string          // the directory edge sockets live in (runtime/hold.Dir); set by the shell, like Floor
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

type Received struct { // → Intake: the bytes, and the Horizon they were emitted under, as read by id
	Under Object
	In    []byte
}
type Released struct { // → Release: this instance's Horizon as read by id, and how many frame fds arrived with it
	Self   Object
	Frames int
}
type Emitted struct { // → Execute: the target's one emission (empty: it exited without one)
	Out []byte
	Err string
}
type Started struct { // → Execute: a resident was started, or its edge already had a live one
	Out []byte
	Err string
}
type Stamped struct { // → Handoff: the record read under the successor's stamp (Key: the stamp; Raw nil: none)
	Record Object
	Err    string
}
type Settled struct { // → Gather: the waiting frame's record and every round it opened, oldest first
	Frame  Object
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
	Invoke    Object
	Discharge *Object
}

// A gathered round, verified: what continuation and resolve read.
type round struct {
	Emission, Stamp string
	Children        []slot
}
type slot struct {
	Stamp, Result     string
	Invoke, Discharge record.Record
	Discharged        bool
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
	Surface  horizon.Surface
	Stdin    []byte
	Profile  string
	Env      []string // the target's whole environment: nothing is inherited
	Resident string   // a connection surface's edge socket: the one path it may reach the resident by
}

// StartResident launches a resident behind its edge's holder (runtime/hold),
// or finds the live one the edge already has.
type StartResident struct {
	Run     Run
	Socket  string // the edge's socket: derived from the floor and the two actors
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

// Committed returns the record an intake transition asks to commit.
func Committed(effs []Effect) record.Record {
	for _, e := range effs {
		if c, ok := e.(Commit); ok {
			return c.R
		}
	}
	return record.Record{}
}

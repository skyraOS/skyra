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
	Phase   Phase           // the phase whose event this instance is waiting for
	Under   horizon.Horizon // the Horizon the input was emitted under
	R       record.Record   // what this instance committed last: its state and stamp are R's
	Self    horizon.Horizon // this instance's own Horizon, once released
	Out     []byte          // the target's emission, while it is handed off
	Kids    []Kid           // the successors stamped and verified so far, in instruction order
	Next    []byte          // a resumed continuation's input, while the claim is decided
	Held    record.Record   // the resolve record, while the claim is decided
	Sockets string          // the directory edge sockets live in (runtime/hold.Dir); set by the shell, like Floor

	Accepting string  // the id of the Horizon a release named, while it is read and accepted
	Arrived   int     // how many descriptors arrived with that release (41, V3)
	Reads     []Key   // the keys of the read in flight: a ReadResult answers them in order
	Cause     string  // a supervisor's observation, held while its successors are killed
	Walk      *walked // gather's reads, while they are in flight
}

// State is the state intake derived, the top-level branch: R's (41, V4).
func (s Instance) State() record.State { return s.R.State }

// Stamp is the hash of R, or "" before this instance has committed anything.
func (s Instance) Stamp() string {
	if s.R.State == "" {
		return ""
	}
	return record.Stamp(s.R)
}

// Outs is the instructions Out carries, one per successor (fan-out: several).
// Out was split when it was emitted, so it splits the same way again.
func (s Instance) Outs() [][]byte {
	outs, _ := Instructions(s.Out)
	return outs
}

// Child is the first successor's stamp, once one is verified.
func (s Instance) Child() string {
	if len(s.Kids) == 0 {
		return ""
	}
	return s.Kids[0].Stamp
}

type Phase string

const (
	Intake        Phase = "intake"        // waiting for bytes
	Stamping      Phase = "stamping"      // the input and record are being committed; then the stamp goes up
	Release       Phase = "release"       // stamped; waiting for the parent's release
	Accepting     Phase = "accepting"     // reading the released Horizon
	Acknowledging Phase = "acknowledging" // the acceptance receipt is being committed
	Acknowledged  Phase = "acknowledged"  // custody is being acknowledged to the parent
	Running       Phase = "running"       // the target is running, or the strand is being read; waiting for its output
	Parking       Phase = "parking"       // custody is being handed to an actor host; waiting to hear whether it took it
	Spawning      Phase = "spawning"      // the Horizon a successor intakes under is being committed
	Handoff       Phase = "handoff"       // the successor is intaking the output; waiting for its stamp
	Verifying     Phase = "verifying"     // reading the record under the successor's stamp
	Opening       Phase = "opening"       // the round's emission, slots, frames and Horizons are being committed
	Killing       Phase = "killing"       // successors are being killed before this instance faults
	Settling      Phase = "settling"      // a discharge is being committed
	Gather        Phase = "gather"        // reading what the waiter still waits on
	Claim         Phase = "claim"         // the resolve and the fan-in claim are being committed
	Done          Phase = "done"
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

// Events are what the world reports to Step (41, V2). Received starts an
// instance; after it, every event answers the effect before it: a read, a
// commit, or the last execute of a batch.
type Event interface{ event() }

type Received struct { // → Intake: the bytes, and the Horizon they were emitted under, as read by id
	Under Object
	In    []byte
}

// ReadResult answers a Read: the bytes found under each key, in order, raw
// and unverified (nil: nothing there).
type ReadResult struct{ Found [][]byte }

// Committed answers a Commit: what happened to each write, in order.
type Committed struct{ Outcomes []Outcome }

// Executed answers an execute: what came back (an emission, a stamp, a
// release's Horizon id), how many descriptors arrived with it, or why it
// failed.
type Executed struct {
	Obs     []byte
	Handles int
	Err     string
}

func (Received) event()   {}
func (ReadResult) event() {}
func (Committed) event()  {}
func (Executed) event()   {}

// Outcome is what a write found.
type Outcome int

const (
	Created   Outcome = iota // the write is new
	Identical                // it existed with the same bytes: a content-addressed retry
	Conflict                 // it existed with different bytes: corruption, never a race
	Lost                     // an exclusive create someone else won
)

// Effects are what Step asks the shell to do (41, V2). There are three kinds,
// by what a crash leaves behind:
//
//	Read     a query of the tree         idempotent: retry freely
//	Commit   write-once writes           idempotent: a retry creates the object or finds it
//	Execute  anything that leaves this   the only kind with an unknown outcome
//	         process: run, spawn, send, park, exit
//
// A cycle is Read* Commit? Execute*: any number of reads, each able to depend
// on the last, at most one commit, then the executes. Each read and commit is
// answered before the next effect; of a batch of executes, the last one's
// result is the next event.
type Effect interface{ effect() }

type Read struct{ Keys []Key }
type Commit struct{ Writes []Write }
type Execute struct {
	Act   Action
	Force bool // everything committed so far must be durable first: the act leaves the failure domain (40)
}

func (Read) effect()    {}
func (Commit) effect()  {}
func (Execute) effect() {}

// Key is one thing a read asks for. An object is read by hash and is the
// same forever. Every other key is a mutable query: its answer depends on
// timing, so these are the concurrency points, and they are this fixed set
// (41, V6). A decision made on a mutable read ends up in a committed record,
// as a resolve names every discharge, so replay can re-derive it.
type Key interface{ key() }

type ObjectKey struct { // an object by hash: records, inputs or horizons
	Kind string
	Hash string
}
type SlotsKey struct{ Round string } // the slots a round opened in the waiting frame (frame handle 1): a JSON array of stamps, in emission order
type EmissionKey struct {            // the hash of the emission a round of a frame published
	Frame, Round string
}
type ResultKey struct{ Frame string } // the stamp of the record that discharged a frame's slot
type StrandKey struct {               // the strand walk: {"strand":…} or {"refused":"…"}
	Observation  string
	Invocation   string // tree mode: retained root, with a nonempty accepted frontier
	Observations []string
	Target       string
	Schema       *contract.Schema
	Origin       string // the requester's Horizon: the walk stays inside it
	Room         string // the enclosing genesis; different boots are not one Horizon
}

func (ObjectKey) key()   {}
func (SlotsKey) key()    {}
func (EmissionKey) key() {}
func (ResultKey) key()   {}
func (StrandKey) key()   {}

// Write is one write of a commit. Objects are content-addressed; the rest
// are write-once names in a frame, reached through a frame handle.
type Write interface{ write() }

type ObjectWrite struct { // an object under its hash: records, inputs or horizons
	Kind  string
	Bytes []byte
}
type AcceptanceWrite struct{ Stamp, Horizon string } // evidence that this instance accepted the release of Stamp under Horizon
type FrameWrite struct{ Stamp string }               // a new frame directory, for a successor that runs in a frame of its own
type EmissionWrite struct{ Round, Hash string }      // this frame's round seals its full emission, before any child runs
type SlotWrite struct {                              // a slot in this frame's round, before its child runs
	Round, Child string
	Pos          int // the instruction's position in the emission: a round's steps come back in this order
}
type ResultWrite struct { // discharge this frame's slot (frame handle 0), once
	Stamp string
	Frame string // the frame it discharges: for the trace only
}
type ClaimWrite struct { // the fan-in claim on the waiting frame (frame handle 1): Lost if a sibling won
	Round, Stamp string
	Frame        string // the frame that resumes: for the trace only
}

func (ObjectWrite) write()     {}
func (AcceptanceWrite) write() {}
func (FrameWrite) write()      {}
func (EmissionWrite) write()   {}
func (SlotWrite) write()       {}
func (ResultWrite) write()     {}
func (ClaimWrite) write()      {}

// Action is what an execute does.
type Action interface{ action() }

type SendStamp struct{ Stamp string } // tell the parent this stamp; its release is the result
type Acknowledge struct{ Err string } // tell the parent whether this instance accepted custody
type Run struct {                     // run the target; its emission is the result
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
type Spawn struct { // start the successor, hand it the bytes and the id of the Horizon they were emitted under; its stamp is the result
	Under string
	Out   []byte
}
type ReleaseChild struct { // issue a successor's release: its Horizon id and its frame stack
	Index    int    // which successor, in instruction order
	Horizon  string // the id of the Horizon it is released with
	NewFrame string // the frame pushed for it, or "": it is handed this instance's stack
	Stamp    string // its stamp: for the trace only
}
type KillSuccessors struct{} // kill every successor spawned so far
type Pop struct{}            // drop the discharged frame (handle 0): the resumed continuation holds the rest
type Park struct {           // hand this instance's frame stack to an actor host; whether it took custody is the result
	Host    string
	Payload []byte
	Stamp   string
	Frames  int
}
type Exit struct{ Err string }

func (SendStamp) action()      {}
func (Acknowledge) action()    {}
func (Run) action()            {}
func (StartResident) action()  {}
func (Spawn) action()          {}
func (ReleaseChild) action()   {}
func (KillSuccessors) action() {}
func (Pop) action()            {}
func (Park) action()           {}
func (Exit) action()           {}

// CommittedRecord returns the record an intake transition asks to commit.
func CommittedRecord(effs []Effect) record.Record {
	for _, e := range effs {
		if c, ok := e.(Commit); ok {
			for _, w := range c.Writes {
				if o, ok := w.(ObjectWrite); ok && o.Kind == "records" {
					var r record.Record
					_ = json.Unmarshal(o.Bytes, &r)
					return r
				}
			}
		}
	}
	return record.Record{}
}

// Settled is what gather read: the waiting frame's record and every round it
// opened, oldest first, as read.
type Settled struct {
	Frame  Object
	Rounds []Round
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

// Package record owns the record: what a step commits, its canonical
// encoding, and the content address (stamp) derived from it. It is pure:
// storing and loading records belongs to package tree.
package record

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"skyra/controller/core/horizon"
)

// State is the state the system is in at a record, named by the test "the
// system is in a state of ___" (D9). The runtime derives it; no emitter can
// state it (thought 14).
type State string

const (
	Resolve State = "resolve" // every dependency of a frame is discharged; its continuation resumes
	Invoke  State = "invoke"  // a capability in reach, with a payload that fits its input contract
	Return  State = "return"  // the slot the emitter's frame discharges, with a payload that fits its output contract, if it declares one
	Fault   State = "fault"   // the runtime proved something broke
)

// Fault kinds (design/faults.md). A target cannot emit a fault; only the
// runtime derives one.
const (
	Malformed = "malformed" // the output cannot be read as operands
	Horizon   = "horizon"   // it names a capability outside its Horizon
	Contract  = "contract"  // the payload does not fit the capability's input contract
	Custody   = "custody"   // the protocol broke: no output, a dead successor, a failed stamp
	Leaf      = "leaf"      // the output addresses nothing: it did not follow the protocol, and is kept as the fault's observation
)

type FaultInfo struct {
	Kind  string `json:"kind"`
	Cause string `json:"cause"`
}

// Record is what a step commits. It is stored under its own hash, and that
// hash is the stamp. Everything but Predecessor and Input is derived from
// them; replay derives it again and must get the same bytes. Predecessor is
// the custody chain: the record whose holder committed this one.
type Record struct {
	Predecessor string                     `json:"predecessor"`          // the stamp of the Horizon it was emitted under; "" only for genesis
	Input       string                     `json:"input"`                // sha256 of the exact bytes observed
	State       State                      `json:"state"`                // derived
	To          string                     `json:"to,omitempty"`         // invoke: the capability
	Payload     json.RawMessage            `json:"payload,omitempty"`    // invoke: its input; return: the value; fault: the observation it was derived from. Canonical JSON
	Fault       *FaultInfo                 `json:"fault,omitempty"`      // fault: kind and cause
	Frame       string                     `json:"frame,omitempty"`      // resolve: the frame that resumes
	Round       string                     `json:"round,omitempty"`      // resolve: the round whose slots are all discharged
	Deps        []string                   `json:"deps,omitempty"`       // resolve: the discharges, one per slot: the bottom-up link (thought 13)
	Reach       map[string]horizon.Surface `json:"reach,omitempty"`      // genesis only: the root Horizon's reach
	Authority   *horizon.Authority         `json:"authority,omitempty"`  // genesis only: the grants everything under the root shares
	Author      string                     `json:"author,omitempty"`     // genesis only: who created the room (next/24: the one authored origin)
	Nonce       string                     `json:"nonce,omitempty"`      // genesis only: makes this act, and every stamp after it, unique
	Supervisor  bool                       `json:"supervisor,omitempty"` // derived from a custody observation, not from a target's output
}

// Sum is the content address of any bytes.
func Sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Encode is the canonical encoding a record is hashed and stored as.
func Encode(r Record) ([]byte, error) {
	return json.Marshal(r)
}

// Stamp is the hash a record is committed under.
func Stamp(r Record) string {
	b, _ := Encode(r)
	return Sum(b)
}

// Short abbreviates a stamp for logs.
func Short(stamp string) string {
	if len(stamp) < 8 {
		return stamp
	}
	return stamp[:8]
}

// Describe is a one-line reading of a record, led by its state.
func Describe(r Record) string {
	if r.Predecessor == "" {
		return "genesis"
	}
	switch r.State {
	case Invoke:
		return fmt.Sprintf("invoke %s %s", r.To, r.Payload)
	case Return:
		return fmt.Sprintf("return %s", r.Payload)
	case Resolve:
		return fmt.Sprintf("resolve %s: %d of %d slots discharged", Short(r.Frame), len(r.Deps), len(r.Deps))
	case Fault:
		by := "derived at intake"
		if r.Supervisor {
			by = "observed by the supervisor"
		}
		return fmt.Sprintf("fault (%s, %s): %s", r.Fault.Kind, by, r.Fault.Cause)
	}
	return string(r.State)
}

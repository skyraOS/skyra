// Package record owns the record: what intake commits, its canonical encoding,
// and the content address (stamp) derived from it. It is pure: storing and
// loading records belongs to package tree.
package record

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"skyra-v1/specs/2026-09-23/minimal-shape/horizon"
	"skyra-v1/specs/2026-09-23/minimal-shape/move"
)

// Record is what intake commits. It is stored under its own hash, and that
// hash is the stamp.
type Record struct {
	Predecessor string                     `json:"predecessor"`          // the parent's stamp; "" only for genesis
	Input       string                     `json:"input"`                // sha256 of the exact bytes received
	State       move.State                 `json:"state"`                // resolve | invoke | return | fault (D9)
	Move        *move.Move                 `json:"move,omitempty"`       // the interpreted Move, when accepted
	Reason      string                     `json:"reason,omitempty"`     // why, when intake or a supervisor faulted it
	Reach       map[string]horizon.Surface `json:"reach,omitempty"`      // genesis only: the root Horizon's reach
	Supervisor  bool                       `json:"supervisor,omitempty"` // committed by a supervisor, not produced by intake
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
	case move.StateInvoke:
		return fmt.Sprintf("invoke %s %v", r.Move.Invoke.Target, r.Move.Invoke.Args)
	case move.StateReturn:
		return fmt.Sprintf("return %q", r.Move.Return.Value)
	case move.StateFault:
		switch {
		case r.Supervisor:
			return "fault (supervisor): " + r.Reason
		case r.Move == nil:
			return "fault (rejected at intake): " + r.Reason
		default:
			return "fault (emitted): " + r.Move.Fault.Reason
		}
	}
	return string(r.State)
}

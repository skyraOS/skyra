// Package move owns the wire vocabulary: the three Move variants, and the
// states a Move can be in. It imports nothing from this module.
package move

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Move is the only message kind. Exactly one variant is set.
type Move struct {
	Invoke *Invoke `json:"invoke,omitempty"`
	Return *Return `json:"return,omitempty"`
	Fault  *Fault  `json:"fault,omitempty"`
}

type Invoke struct {
	Target string            `json:"target"`
	Args   map[string]string `json:"args"`
}

type Return struct {
	Value string `json:"value"`
}

type Fault struct {
	Reason string `json:"reason"`
}

// State is the state of a Move, named by the test "the system is in a state
// of ___" (D9).
type State string

const (
	StateResolve State = "resolve" // holes unfilled or no custody yet (not in the minimal shape)
	StateInvoke  State = "invoke"  // in an instance's custody: its target runs
	StateReturn  State = "return"  // terminal: a value
	StateFault   State = "fault"   // terminal: stuck (emitted, rejected at intake, or committed by a supervisor)
)

// Parse interprets unchecked bytes. Unknown fields and anything other than
// exactly one variant are rejected.
func Parse(b []byte) (Move, error) {
	var m Move
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return Move{}, err
	}
	n := 0
	for _, set := range []bool{m.Invoke != nil, m.Return != nil, m.Fault != nil} {
		if set {
			n++
		}
	}
	if n != 1 {
		return Move{}, errors.New("expected exactly one Move variant")
	}
	return m, nil
}

// StateOf names the state an accepted Move puts its frame in.
func StateOf(m Move) State {
	switch {
	case m.Invoke != nil:
		return StateInvoke
	case m.Return != nil:
		return StateReturn
	default:
		return StateFault
	}
}

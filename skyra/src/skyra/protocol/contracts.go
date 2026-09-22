// Package protocol declares the bounded experiment's JSON Move vocabulary.
// No actual port, OS handle, process identity, or runtime pointer crosses this wire.
package protocol

type ContextID string
type ExecutableID uint64
type FrameID uint64
type Observation struct {
	Key   string `json:"key"`
	Bytes []byte `json:"bytes"`
}
type Invoke struct {
	Frame    FrameID       `json:"frame,omitempty"`
	ReturnTo FrameID       `json:"return_to,omitempty"`
	Args     []Observation `json:"args"`
}
type Return struct {
	Answering   FrameID     `json:"answering"`
	Observation Observation `json:"observation"`
}

// Fault is a fixture diagnostic, not a definition of the proposal's FaultType.
type Fault struct {
	Frame   FrameID `json:"frame"`
	Message string  `json:"message"`
}
type Move struct {
	Identity uint64       `json:"identity"`
	Context  ContextID    `json:"context,omitempty"`
	Target   ExecutableID `json:"target,omitempty"`
	Source   FrameID      `json:"source,omitempty"`
	Invoke   *Invoke      `json:"invoke,omitempty"`
	Return   *Return      `json:"return,omitempty"`
	Fault    *Fault       `json:"fault,omitempty"`
	// State is an externalized parent continuation, carried back on resume.
	State []byte `json:"state,omitempty"`
}

const Echo ExecutableID = 1
const Parent ExecutableID = 2
const User ExecutableID = 3

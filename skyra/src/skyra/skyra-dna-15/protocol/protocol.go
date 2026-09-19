package protocol

import (
"crypto/rand"
"encoding/hex"
)

func (Invoke) opcode() {}

func (Return) opcode() {}

func (MoveBatch) opcodeEvent() {}

func (DispatchResult) opcodeEvent() {}

func (AdmissionFailure) opcodeEvent() {}

func (ExecutionFailure) opcodeEvent() {}

func NewInvocationID() (InvocationID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return InvocationID{}, err
	}
	return InvocationID{Value: hex.EncodeToString(b[:])}, nil
}

func (Envelope) opcodeEvent() {}


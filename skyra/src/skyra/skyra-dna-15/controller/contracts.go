package controller

import (
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"context"
"skyra-v1/src/skyra/skyra-dna-15/driver"
)

// Validate returns a validation verdict outside the scheduler lock. Submit and
// execute translate it into events. This lets execute check
// all outputs before ONE atomic batch write, without a second scheduler writer.
// Dispatch/execute retain write callbacks and no persistent invocation state.
type Controller interface {
	Validate(ctx context.Context, move protocol.Move) error
	Dispatch(ctx context.Context, move protocol.Move, sequence uint64, write func(protocol.OpcodeEvent) error) error
}

type ContinuationController struct {
	registry driver.Resolver
}


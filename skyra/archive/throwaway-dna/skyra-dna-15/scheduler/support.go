package scheduler

import (
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
)

func CloneContinuation(c Continuation) Continuation {
	c.Move = support.CloneMove(c.Move)
	c.Handoff.PendingOutput = support.CloneMoves(c.Handoff.PendingOutput)
	if c.Result != nil {
		o := support.CloneObservation(*c.Result)
		c.Result = &o
	}
	if c.CompiledRegister != nil {
		c.CompiledRegister = &CompiledRegister{}
	}
	return c
}


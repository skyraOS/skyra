// Package ctu is the execution endpoint — the architecture (the-quadrant.md:
// system space, unprivileged). It owns the live possibility tree, the two
// dispatch tables, the interpreter, and the scheduler. One CTU per endpoint.
//
package ctu

import c "skyra-v1/src/skyra/skyra-dna-v7/causal"

// CTU is the endpoint machine — the primitives it performs. Issuing any of these
// is unprivileged; the CTU reaches the kernel only by NAMING it (dispatch.md),
// and reaches durable storage only at ACTIVATE / REMEMBER (boundaries.md).
type CTU interface {
	// Move issues custody transfer — the only transport. Returns the
	// three-outcome (Success | Failure | NotYet). move.md.
	Move(m c.Move) c.Outcome

	// Greet is the REQUEST port: admits work, checks authority. Refuses by
	// returning Failure. primitive.md, traffic-routing.md.
	Greet(from c.Target, m c.Move) c.Outcome

	// Recognize is the RESPONSE port: checks the returned Observation against the
	// frame-resident Contract, in the parent's frame. primitive.md.
	Recognize(m c.Move, returned c.Observation) c.Outcome

	// Think proposes potentials at the frontier — the propose side only,
	// unprivileged and effectless. At most five. think.md.
	Think(g c.Goal, f c.Frame) []c.Potential

	// Bound ATTENUATEs a held horizon: append a caveat, narrowing only. Local,
	// unprivileged (boundaries.md SEAM 2 resolution). The privileged DEFINE half
	// is not here — it NAMES the kernel. horizon-and-root.md §2a.
	Bound(h c.Horizon, add c.Caveat) c.Horizon

	// Orient reads the endpoint's own outstanding set. Purely local; possibly not
	// a protocol act at all. primitive.md (UNPLACED).
	Orient() c.Outstanding
}

// Interpreter reads the MOVE INTERFACE (target + sys-instruction), never the
// payload, into the eligible set. decode-and-run.md.
type Interpreter interface {
	Interpret(m c.Move) (c.EligibleSet, c.Outcome)
}

// Scheduler picks where + when within the eligible set (operand wait, resource
// gate, placement). Local-vs-distributed is OPEN. decode-and-run.md.
type Scheduler interface {
	Schedule(e c.EligibleSet, operands []c.Observation) (c.Resource, c.Outcome)
}

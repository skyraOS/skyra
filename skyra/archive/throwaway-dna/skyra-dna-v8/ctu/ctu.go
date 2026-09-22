// Package ctu is the execution endpoint — the architecture (the-quadrant.md:
// system space, unprivileged). In the distributed model it EXECUTES A POSSIBILITY:
// it steps the possibility, running what it can locally and FORWARDING the slice
// that needs another CTU. One CTU per endpoint.
//
// AUTHORITY: skyra/specs/2026-08-22/ctu/ (possibility.md, orient.md).
package ctu

import c "skyra-v1/src/skyra/skyra-dna-v8/causal"

// CTU is the endpoint machine. Issuing any primitive is unprivileged; the CTU
// reaches the kernel only by NAMING it (dispatch.md).
type CTU interface {
	// Execute runs a possibility to the extent this CTU can. It is the CORE of the
	// distributed model — "the CTU executes a possibility" — not a ninth primitive
	// but the loop that steps a possibility by issuing the primitives below. When a
	// dependency cannot be resolved locally, Execute forwards a slice (PUSH) rather
	// than parking a waiting parent. ctu/possibility.md. (Execute is Claude's name
	// for the loop — claude-ideas.md.)
	Execute(p *c.Possibility) c.Outcome

	// Move issues a cut — custody transfer. A REQUEST leg either forwards a Slice
	// (push) or carries a pull instruction; a RESPONSE leg returns an observation at
	// a join. move.md, fabric/signed-provenance.md.
	Move(m c.Move) c.Outcome

	// Greet is the REQUEST port: admits an arriving cut, checks authority, and
	// RESUMES its Slice as a child possibility to execute (or makes a leaf for a
	// pull). Returns that possibility. The forwarded slice becoming a possibility is
	// the child-activation the v7 build had no home for. ctu/possibility.md,
	// traffic-routing.md.
	Greet(from c.Target, m c.Move) (*c.Possibility, c.Outcome)

	// Recognize is the RESPONSE port: it locates the owning possibility from the cut
	// edge (m.ID), VERIFIES the returned observation's signed provenance across the
	// trust boundary, checks it against the frame-resident Contract at the correlated
	// join, and records it. fabric/signed-provenance.md, continuation-language.md.
	Recognize(m c.Move, returned c.Observation) c.Outcome

	// Think proposes potentials at the frontier — propose side only, unprivileged and
	// effectless. At most five. think.md.
	Think(g c.Goal, f c.Frame) []c.Potential

	// Bound ATTENUATEs a held horizon: append a caveat, narrowing only. Central to
	// the distributed model — the horizon that travels with a slice is attenuated to
	// that slice (least-knowledge). boundaries.md SEAM 2, ctu/possibility.md.
	Bound(h c.Horizon, add c.Caveat) (c.Horizon, c.Outcome)

	// Orient reads this fragment's own outstanding FRONTIER — local by construction.
	// The whole-possibility view composes by propagation over the returned edges;
	// there is no global observer without a cut. ctu/orient.md.
	Orient(p *c.Possibility) c.Frontier
}

// Interpreter reads the MOVE INTERFACE (target + sys-instruction), never the
// payload, into the eligible set. decode-and-run.md.
type Interpreter interface {
	Interpret(m c.Move) (c.EligibleSet, c.Outcome)
}

// Scheduler picks WHERE + when within the eligible set. In the distributed model
// WHERE is a PUSH-vs-PULL choice: a local Resource (RunHere) or a remote endpoint
// to forward the slice to (SendTo). decode-and-run.md, ctu/possibility.md.
type Scheduler interface {
	Schedule(e c.EligibleSet, operands []c.Observation) (c.Selection, c.Outcome)
}

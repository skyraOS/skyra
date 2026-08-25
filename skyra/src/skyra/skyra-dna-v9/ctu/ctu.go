// Package ctu is the execution endpoint — the CORE (live-possibility.md §1). It is
// nearly stateless pipeline hardware that operates on whatever Possibility (the
// CONTEXT) is resident. It EXECUTES a possibility by DRIVE — the reservation-station
// cycle — issuing the eight primitives, running what it can locally and FORWARDING a
// narrow child when a dependency needs another CTU. One CTU per endpoint.
//
// AUTHORITY: skyra/specs/2026-08-23/ (execution-cycle.md, live-possibility.md), on
// skyra/specs/2026-08-22/ctu/ and 2026-08-18/the-quadrant.md, dispatch.md.
package ctu

import c "skyra-v1/src/skyra/skyra-dna-v9/causal"

// CTU is the endpoint machine — the core. Issuing any primitive is unprivileged; the
// CTU reaches the kernel only by a TRAP (naming it). dispatch.md, the-quadrant.md.
type CTU interface {
	// Drive is the reservation-station CYCLE over a RESIDENT possibility — the engine
	// (supersedes v8's Execute). It turns landings into ready moves, hands the ready set
	// to SCHEDULE, dispatches locally or PARKS a placed cut on the frontier, and returns
	// when the context quiesces. It never blocks and never recurses through a port. It is
	// NOT a ninth primitive — it is the loop that issues the eight. execution-cycle.md §2.
	Drive(p *c.Possibility) c.Outcome

	// Move issues a cut — a Transfer on the wire. A REQUEST leg pushes a narrow child
	// possibility; a RESPONSE leg returns an observation at a join. move.md,
	// possibility-graph.md §8.
	Move(t c.Transfer) c.Outcome

	// Greet is the REQUEST port and an EVENT SOURCE. It admits an arriving Transfer: if it
	// carries a runnable, authorized child (a CONTINUATION), it stamps a LOCAL arena, makes
	// the child resident, and DRIVEs it — no kernel. If a reference does not resolve, it
	// FAULTS to the kernel (LOAD via Activate). It returns the resumed possibility. The
	// forwarded child becoming resident is the migration the v7/v8 builds had no home for.
	// possibility-residency.md §2,4, execution-cycle.md §2.
	Greet(from c.Target, t c.Transfer) (*c.Possibility, c.Outcome)

	// Recognize is the RESPONSE port and an EVENT SOURCE. It VERIFIES the returned
	// observation's signed segment (the return-path trust checkpoint), resolves the owning
	// possibility LOCALLY from the cut edge (t.ID), LANDS the observation into
	// Registers[InFlight[t.ID]], checks THAT register's contract, re-readies dependents, and
	// DRIVEs. It does NOT treat every return as a final Result. fabric/signed-provenance.md,
	// execution-cycle.md §2,5.
	Recognize(t c.Transfer) c.Outcome

	// Think proposes potentials at the frontier — propose side only, unprivileged and
	// effectless. At most five, read from the possibility's window. think.md.
	Think(p *c.Possibility, g c.Goal) []c.Potential

	// Bound ATTENUATEs a held horizon: append a caveat, narrowing only. The horizon that
	// travels with a child is attenuated to it (least-knowledge). The CTU can ONLY narrow —
	// it cannot MINT authority, which is why origination is the kernel's. boundaries.md SEAM 2,
	// possibility-residency.md §3.
	Bound(h c.Horizon, add c.Caveat) (c.Horizon, c.Outcome)

	// Orient reads this fragment's own outstanding FRONTIER — DERIVED from its InFlight,
	// local by construction. The whole-possibility view composes by propagation over the
	// edges; there is no global observer without a cut. ctu/orient.md.
	Orient(p *c.Possibility) c.Frontier
}

// Interpreter is BLOCK-SELECTION (the trust axis, made concrete). It reads the move's
// INTERFACE (Target + Sys), never its Payload, against the two tables (host = system,
// entity = user) and quadrant-filters, so the eligible set NEVER spans the forbidden
// diagonal. A name that MISSES is the FAULT signal. dispatch.md, execution-cycle.md §3.
type Interpreter interface {
	Interpret(p *c.Possibility, mv *c.Move) (c.EligibleSet, c.Outcome)
}

// Scheduler picks WITHIN the eligible set: a free resource + placement. RunHere is a
// local dispatch (host BELIEVED / entity STAMPED+CHECKED / kernel TRAP); SendTo forwards
// the child. Placement never touches the graph. decode-and-run.md, execution-cycle.md §3.
type Scheduler interface {
	Schedule(e c.EligibleSet, operands []c.Observation) (c.Selection, c.Outcome)
}

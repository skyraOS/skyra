// Package entity is the ENTITY BLOCK — untrusted hardware, user space (the-quadrant.md,
// right column). It is decode + ALU: it reads a move's payload, computes, and yields RAW
// content. Its output is NEVER trusted — the CTU stamps provenance and, before it crosses
// a cut, SIGNS it as a segment. The running shape above it changed (a reorder buffer of
// moves); this interior did not — one decode → one run → one content.
//
// AUTHORITY: skyra/specs/2026-08-18/entity-ctu-boundary/decode-and-run.md;
// skyra/specs/2026-08-23/execution-cycle.md §3 (the ENTITY-block fate: run, STAMP + CHECK).
package entity

import c "skyra-v1/src/skyra/skyra-dna-v9/causal"

// Entity is an entity-block entry. It returns Content, NOT an Observation — the difference
// is the trust boundary: the entity produces content, the CTU produces (and signs)
// causality. decode-and-run.md, fabric/signed-provenance.md.
type Entity interface {
	// Run: DECODE the payload (the entity reads it; the CTU never does), compute over the
	// operands, and yield EXACTLY ONE raw content. decode-and-run.md.
	Run(operands []c.Observation, p c.Payload) (c.Content, c.Outcome)
}

// Function adapts a Go function into foreign execution hardware, so the boundary can run
// without the CTU knowing a payload language. Mechanism, not a machine concept.
type Function func([]c.Observation, c.Payload) (c.Content, c.Outcome)

func (f Function) Run(operands []c.Observation, p c.Payload) (c.Content, c.Outcome) {
	return f(operands, p)
}

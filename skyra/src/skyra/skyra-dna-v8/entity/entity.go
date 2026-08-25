// Package entity is the ENTITY TABLE — untrusted hardware, user space
// (the-quadrant.md, right column). It is decode + ALU: it reads the payload,
// computes, and yields RAW content. Its output is NEVER trusted — the CTU stamps
// provenance and, in the distributed model, SIGNS it as a segment before it crosses
// a cut.
//
// The distributed model leaves this interior untouched: one decode → one run → one
// content, unchanged. AUTHORITY: skyra/specs/2026-08-18/entity-ctu-boundary/
// decode-and-run.md; skyra/specs/2026-08-22/entity/README.md (no new interface).
package entity

import c "skyra-v1/src/skyra/skyra-dna-v8/causal"

// Entity is an entity-table entry. It returns Content, NOT an Observation — the
// difference is the trust boundary: the entity produces content, the CTU produces
// (and signs) causality. decode-and-run.md, fabric/signed-provenance.md.
type Entity interface {
	// Run: DECODE the payload (the entity reads it; the CTU never does), compute, and
	// yield EXACTLY ONE raw content. decode-and-run.md.
	Run(operands []c.Observation, p c.Payload) (c.Content, c.Outcome)
}

// Function adapts a Go function into foreign execution hardware, so the boundary can
// run without the CTU knowing a payload language. Mechanism, not a machine concept.
type Function func([]c.Observation, c.Payload) (c.Content, c.Outcome)

func (f Function) Run(operands []c.Observation, p c.Payload) (c.Content, c.Outcome) {
	return f(operands, p)
}

// Package entity is the ENTITY TABLE — untrusted hardware, user space
// (the-quadrant.md, right column; dispatch.md). Foreign code (e.g. a python
// interpreter). It is decode + ALU: it reads the payload, computes, and yields
// RAW content. Its output is NEVER trusted — the CTU stamps provenance and checks
// the contract afterward. decode-and-run.md.
//
package entity

import c "skyra-v1/src/skyra/skyra-dna-v7/causal"

// Entity is an entity-table entry. It returns Content, NOT an Observation — the
// difference is the trust boundary: the entity produces content, the CTU produces
// causality (provenance). decode-and-run.md.
type Entity interface {
	// Run: DECODE the payload (the entity reads it; the CTU never does), compute,
	// and yield EXACTLY ONE raw content. One decode → one run → one content: no
	// streaming, no side-outputs, no zero. decode-and-run.md.
	//
	// operands are the resolved input observations the CTU dispatched with.
	Run(operands []c.Observation, p c.Payload) (c.Content, c.Outcome)
}

// Function adapts a Go function into foreign execution hardware. It exists so
// the boundary can run without making the CTU aware of a payload language.
type Function func([]c.Observation, c.Payload) (c.Content, c.Outcome)

func (f Function) Run(operands []c.Observation, p c.Payload) (c.Content, c.Outcome) {
	return f(operands, p)
}

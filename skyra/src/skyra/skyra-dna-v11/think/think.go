// Package think is the COLD-path compiler.
//
// THINK turns an unstructured request into a compiled Body — it PROPOSES, it does not execute, mint,
// or provision. Read-only horizon; no ask_permission. It runs via the kernel (the cold path). Its
// output (a Body) is what the kernel provisions, mints a capability for, and REMEMBERs.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/research/think-act-paths.md, think-act-signature.md.
package think

import c "skyra-v1/src/skyra/skyra-dna-v11/causal"

// Think is the compiler. Compile is the only surface; whether THINK is one call or a possibility-tree
// (acquire/compose as children) is a seam — see claude-ideas.md.
type Think interface {
	// Compile turns a request (an Observation) into a compiled Body — the runnable move graph. It
	// authors the RUNNABLE half of each move's signature (target + reachable + shape); the input/return
	// TYPE is left for ACT to negotiate. It NEVER executes, mints, or provisions.
	Compile(request c.Observation) (c.Body, c.Outcome)
}

// Acquire and Compose are THINK's children (from v10 + the v11 direction). Named here so the reader
// sees the pipeline; the implementer decides whether they are real interfaces or moves in a think-tree.
type Acquire interface {
	// Acquire onboards a NOVEL capability (discover/test/admit) so a target becomes reachable. It
	// returns a reference THINK can name in a Body; the actual admit is the kernel's (a fault).
	Acquire(intent c.Observation) (c.Target, c.Outcome)
}

type Compose interface {
	// Compose assembles registered capabilities into the Body's move graph — the authoring surface.
	Compose(intent c.Observation, available []c.Target) (c.Body, c.Outcome)
}

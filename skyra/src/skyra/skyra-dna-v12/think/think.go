// Package think is the COLD-path compiler.
//
// THINK turns an unstructured request into a compiled Body. It PROPOSES — it never executes, mints, or
// provisions. It plans against a READ-ONLY view of what is REGISTERED (discoverable) and can only compose
// moves whose targets the kernel will authorize. Its output (a Body) is what the kernel BOUNDs a horizon
// for, MINTs a ticket for, runs, and REMEMBERs.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/research/think-act-paths.md, think-act-signature.md;
// skyra/specs/2026-08-27/horizon-and-attention.md (THINK plans ATTENTION within the kernel's horizon).
package think

import c "skyra-v1/src/skyra/skyra-dna-v12/causal"

// Think is the compiler. Compile is the only outward surface; acquire/compose are its children (a seam:
// real interfaces, or moves in a think-tree). THINK plans a workflow's ATTENTION — the per-node children
// in the frame — NOT its horizon (the kernel creates that from the compiled Body).
type Think interface {
	// Compile turns a request into a compiled Body — the runnable move graph. It authors the RUNNABLE half
	// of each move (target + reach + shape); the input/return TYPE is left for ACT. `registered` is the
	// READ-ONLY discovery view (what EXISTS) — THINK composes from it but grants nothing. NEVER executes,
	// mints, or provisions.
	Compile(request c.Observation, registered []c.Target) (c.Body, c.Outcome)
}

// Acquire onboards a NOVEL capability (discover/test/admit) so a target becomes registered/reachable. It
// returns a reference THINK can name; the actual admit is the KERNEL's (through the privilege boundary —
// discover/test are user-side, per v10). THINK proposes the need; it does not admit.
type Acquire interface {
	Acquire(intent c.Observation) (c.Target, c.Outcome)
}

// Compose assembles registered targets into the Body's move graph — the authoring surface, the ATTENTION
// layout. Pure structure: no I/O, no execution.
type Compose interface {
	Compose(intent c.Observation, available []c.Target) (c.Body, c.Outcome)
}

// Package dnav2 is the slice-based-dispatch rebuild of the dispatcher layer.
//
// The v1 dispatcher cells (sensory/midlevel/other/self/motor) collapse into ONE
// canonical base cell — the DNA (the genome) — holding all durable indexes and all
// competences (namespaced genes). Per-wave work runs on disposable transcripts
// (ephemeral roots / mRNA). Routing is centralized in the orchestrator (membrane),
// which reads a static workflow definition + the dynamic job board and dispatches.
//
// See specs/2026-07-03/slice-based-dispatch.md.
package dnav2

import "skyra-v1/src/skyra"

// A competence is a GENE: a stateless transform run by an ephemeral root (the
// transcript). It reaches DURABLE state via root.DNA() and PER-WAVE state via the
// root; it MUST NOT stash durable state on itself — that invariant is what protects
// the accumulated experience from behavior churn (swap competences freely, the DNA
// is untouched). Two phases match the kernel's bookends: Observe (descent — the
// work) and Express (ascent — report / route). Most pipeline genes do their work in
// Observe and leave Express to the orchestrator.
type ObserveCompetence func(root *EphemeralRoot, origin skyra.Space, impulse any)

type ExpressCompetence func(root *EphemeralRoot, origin skyra.Space, impulse any) any

// Package engram is the DISK tier — the persistent store under the CTU's RAM caches.
//
// FIRST SHAPE (v11 left this un-combed). The engram is what REMEMBER writes and ACTIVATE reads: the warm
// tier between hot RAM (Cache/BodyCache) and the cold kernel/THINK compile. Content-addressed: an Engram's
// key IS its BodyID (a digest), so the store is a CAS — dedup and integrity-check come for free.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/research/memory-hierarchy-and-faults.md ("the engram
// disk is UN-COMBED — next area").
package engram

import c "skyra-v1/src/skyra/skyra-dna-v12/causal"

// Engram is the persisted COMPILED artifact. FIRST CALL: it persists the Body only (the static graph),
// not a live Context — a run is re-materialized fresh, not resumed. (COMB POINT: is a checkpointed Context
// ever persisted? for now, no — Remember writes bodies; Contexts are RAM-only until Remembered as bodies.)
type Engram struct {
	ID   c.BodyID // == digest(Body); the CAS key
	Body c.Body
	// meta (COMB): provenance, mint/compile time, refcount for GC — deferred until the shape settles.
}

// Store is the disk seam. The kernel owns it (one per host); the CTU never touches it — it faults to the
// kernel, which ACTIVATEs (reads) or, on a miss, compiles fresh. Symmetric with Remember (writes).
type Store interface {
	// Activate is the PAGE-IN (the warm hit): resolve a Body by content address. A miss (false) is the
	// deepest fault → the kernel compiles fresh via THINK (the COLD path). Verify by hash on read.
	Activate(id c.BodyID) (c.Body, bool)
	// Persist is REMEMBER's write (the page-out): store a compiled Body under its content address. Idempotent
	// — an identical Body re-persists to the same key (CAS dedup). This is what makes a future run WARM.
	Persist(body c.Body) (c.BodyID, c.Outcome)
	// Evict reclaims disk (GC). A Body is safe to evict when no resident Context references its BodyID and
	// no held capability depends on it. (COMB: refcount vs sweep; deferred.)
	Evict(id c.BodyID) c.Outcome
}

// Digest computes the content address of a Body — the single source of a BodyID. Same bytes → same ID, so
// two hosts that compile the same Body agree on its key (the basis for the deferred distributed tier).
type Digest interface {
	Digest(body c.Body) c.BodyID
}

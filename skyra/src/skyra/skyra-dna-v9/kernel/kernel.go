// Package kernel is system space + privileged (the-quadrant.md, top-left). It owns the
// durable graph, the root/lifetime structure, and the namespace. In the residency model
// it is the MATERIALIZE handler: it turns an unresolvable reference into a live
// possibility, either LOADing an existing engram (a page fault) or MINTing a brand-new
// one (origination). Its commit job, REMEMBER, GATHERs a cone that spans CTUs.
//
// AUTHORITY: skyra/specs/2026-08-23/possibility-residency.md, on
// skyra/specs/2026-08-22/kernel/remember-gather.md and 2026-08-18/kernel-ctu-boundary/.
package kernel

import c "skyra-v1/src/skyra/skyra-dna-v9/causal"

// Kernel is the privileged side, reached only by a TRAP (naming it). MATERIALIZE splits
// into Activate (LOAD) and Mint (MINT) on the EVENT/REQUEST axis; Remember is the STORE
// crossing; Define is the authority-tier write. possibility-residency.md §3, dispatch.md.
type Kernel interface {
	// Activate is LOAD — the PAGE-FAULT handler (durable → live, EVENT/involuntary entry).
	// Resolve a name, load the engram under the admitted horizon, and return a possibility.
	// A forwarded child from MIGRATION does NOT come through here — it resumes at ctu.Greet;
	// Activate is only the durable load. possibility-residency.md §2–3, boundaries.md.
	//
	// The returned possibility's Move graph MATERIALIZES from Engram.Body's static
	// templates. See causal.Materialize and boundary-crossings.md.
	Activate(name c.Name, h c.Horizon) (c.Possibility, c.Outcome)

	// Mint is MINT — origination of a possibility with NO backing engram (REQUEST/voluntary
	// entry, a syscall). It creates a fresh root identity under an authenticated horizon.
	// It is the kernel's BY CONSTRUCTION: the CTU can only Bound (attenuate), never mint
	// fresh authority. possibility-residency.md §3.
	Mint(name c.Name, h c.Horizon) (c.Possibility, c.Outcome)

	// Remember is STORE (live → durable), ROOT-ONLY, and DISTRIBUTED: the cone a kept
	// observation depended on may span CTUs, so the body GATHERS it — walking each retired
	// move's Provenance.Causes ([]SegmentRef) across CTU boundaries (via fabric.Fetch),
	// dedup'd by content-address, converging at the root's kernel. remember-gather.md.
	//
	// SEAM: root-only stays STRUCTURAL (root has Parent == nil); the arena back-reference on
	// a non-root possibility must not become a commit handle.
	Remember(root *c.Possibility, keep c.Observation) (c.Cone, c.Outcome)

	// Define is BOUND's privileged DEFINE half — the authority-tier write. It authenticates
	// a Pin and returns the account and its root horizon (which Mint then originates under).
	// horizon-and-root §2a, boundaries.md SEAM 2 (Model A).
	Define(p c.Pin) (c.Account, c.Horizon, c.Outcome)
}

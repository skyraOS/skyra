// Package kernel is system space + privileged (the-quadrant.md, top-left). It owns
// the durable graph, the root/lifetime structure, and the namespace. In the
// distributed model its one changed job is REMEMBER: the cone now spans CTUs, so
// commit is a GATHER to the root.
//
// AUTHORITY: skyra/specs/2026-08-22/kernel/remember-gather.md, on
// skyra/specs/2026-08-18/kernel-ctu-boundary/.
package kernel

import c "skyra-v1/src/skyra/skyra-dna-v8/causal"

// Kernel is the privileged side. Activate and Remember are the two memory-tier
// crossings; Define is the authority-tier write (Model A). All three are reached by
// NAMING the kernel.
type Kernel interface {
	// Activate is LOAD (durable → live): resolve a name, load the engram under the
	// admitted horizon, and return a possibility. NOTE: child possibilities from
	// MIGRATION do NOT come through here — a forwarded slice is resumed at ctu.Greet.
	// Activate is only the durable load. boundaries.md, ctu/possibility.md.
	Activate(name c.Name, h c.Horizon) (c.Possibility, c.Outcome)

	// Remember is STORE (live → durable), ROOT-ONLY, and now DISTRIBUTED: the cone a
	// kept observation depended on may span CTUs, so the body GATHERS it — walking
	// Provenance.Causes across CTU boundaries (via fabric.Fetch), dedup'd by
	// content-address, converging here at the root's kernel. Signature is unchanged
	// from v7; the WHERE of the walk is new. remember-gather.md.
	//
	// SEAM: root-only stays STRUCTURAL (root has Parent == nil); the arena
	// back-reference on a non-root possibility must not become a commit handle.
	Remember(root *c.Possibility, keep c.Observation) (c.Cone, c.Outcome)

	// Define is BOUND's privileged DEFINE half — the authority-tier write. It NAMES
	// the kernel, authenticates a Pin, and returns the account and its root horizon.
	// Ties Account + Pin. horizon-and-root §2a, boundaries.md SEAM 2 (Model A).
	Define(p c.Pin) (c.Account, c.Horizon, c.Outcome)
}

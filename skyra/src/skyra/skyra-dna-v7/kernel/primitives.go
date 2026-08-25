// Package kernel is system space + privileged (the-quadrant.md, top-left). It
// owns the durable graph, the root/lifetime structure, and name-meaning (the
// namespace). It is the ONE privileged executor — reached by being NAMED
// (dispatch.md). It meets the CTU at exactly two crossings.
//
package kernel

import c "skyra-v1/src/skyra/skyra-dna-v7/causal"

// Kernel is the privileged side. Activate and Remember are the two MEMORY-tier
// crossings (boundaries.md); Define is the AUTHORITY-tier write (Model A) — a
// separate axis, not a third memory crossing. All three are reached by NAMING
// the kernel (dispatch.md); that is why they are kernel methods.
type Kernel interface {
	// Activate is LOAD (durable → live): resolve a name against the namespace,
	// load the engram under the admitted horizon, and return a possibility.
	// Read-resolution happens here. primitive.md, boundaries.md.
	Activate(name c.Name, h c.Horizon) (c.Possibility, c.Outcome)

	// Remember is STORE (live → durable), ROOT-ONLY: commit the cone the kept
	// observation depended on, and tear down the arena. Write-resolution — new
	// names entering the namespace — happens here. The kernel computes WHAT
	// persists (the cone); a driver persists HOW. primitive.md, durable-storage.md.
	Remember(root c.Possibility, keep c.Observation) (c.Cone, c.Outcome)

	// Define is BOUND's privileged DEFINE half — the authority-tier write. It
	// NAMES the kernel, so it lives here, not on the CTU (which only ATTENUATEs).
	// The seed case authenticates a Pin and returns the account and its root
	// horizon; naming further durable horizons generalizes this. Ties Account +
	// Pin to an interface. horizon-and-root §2a, boundaries.md SEAM 2 (Model A).
	Define(p c.Pin) (c.Account, c.Horizon, c.Outcome)
}

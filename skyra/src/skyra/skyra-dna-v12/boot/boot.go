// Package boot is the thin OUT-OF-BAND bootloader — main IS this.
//
// Order (bootstrap/sequence): SEED the kernel with the human's authority; START the CTU (provisioned;
// OS-abstraction via the driver seam, baremetal-capable); run the HANDSHAKE (the kernel runs acquire on
// the CTU, registering its advertised drivers into the sysblock); then EXIT. Boot is NOT a resident
// service — it hands off and returns. One kernel per host (distributed deferred).
//
// AUTHORITY: specs/2026-08-24/skyra-dna-v11/bootstrap/ (sequence, kernel, ctu).
package boot

import c "skyra-v1/src/skyra/skyra-dna-v12/causal"

// Boot brings the host up from a Seed and returns — the bootloader exits once the kernel and CTU are up
// and the handshake is done. It mints nothing itself; it only installs the human's seed authority into
// the kernel, which owns everything from there.
type Boot interface {
	Boot(seed Seed) c.Outcome
}

// Seed is the out-of-band authority the human hands the kernel at boot: the root principal and the initial
// grants it holds. Everything else is derived THROUGH the kernel from here — this is the only authority
// that does not come from a mint.
type Seed struct {
	Root   c.Entity  // the seeded human principal (self-authorizing)
	Grants c.Horizon // the initial authority the root holds
}

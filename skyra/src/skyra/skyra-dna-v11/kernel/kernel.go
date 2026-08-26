// Package kernel is the OS-independent AUTHORITY.
//
// FRESH BUILD. The kernel owns IDENTITY and the FULL capability table; it MINTS, PROVISIONS, and guards
// the BARRIER. It does NOT coordinate (that is entity behavior) and it does NOT execute (that is the
// CTU's). It is loaded out-of-band and seeded by the human. seL4-shaped: a capability authority.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/ownership.md, research/capability-table-and-horizon.md,
// research/authorization-and-registration.md, bootstrap/kernel.md.
package kernel

import (
	c "skyra-v1/src/skyra/skyra-dna-v11/causal"
	"skyra-v1/src/skyra/skyra-dna-v11/think"
)

// Kernel is the cold-path entry and the authority root.
type Kernel interface {
	// Handle is the COLD path: a NEW request move arrives (entity=kernel). The kernel THINK-compiles it
	// into a Body, PROVISIONS a horizon, MINTS a capability, runs it once, and REMEMBERs — returning the
	// CAPABILITY the user then HOLDS for the hot path, plus the result. (think-act-paths)
	Handle(from c.Entity, request c.Observation) (held c.Capability, result c.Observation, out c.Outcome)
}

// Table is the FULL capability table — kernel-owned. The CTU only ever sees a provisioned VIEW (a
// horizon). Mint and Provision are kernel-only; Lookup is how the kernel resolves a target.
type Table interface {
	Lookup(t c.Target) (Resource, bool)
	// Mint originates authority — registers a capability (a macaroon) into the table. Kernel-only.
	Mint(cap c.Capability) c.Outcome
	// Provision hands out a capability over a registered target for a principal — WIDENING. `Bound`
	// only narrows; only the kernel widens. This is the page-table "map". (capability-table-and-horizon)
	Provision(t c.Target, principal c.EntityID) (c.Capability, c.Outcome)
	// Remember caches a compiled Body (content-addressed) so a future request skips THINK — the bridge
	// from cold to hot. (think-act-paths; authorization-and-registration = registration.)
	Remember(b c.Body) c.Outcome
}

// Resource is a table entry. Block decides the trust rule (host=BELIEVED, entity=CHECKED). Privilege is
// NOT a block — it is the trap axis (D5).
type Resource interface {
	Target() c.Target
	Block() Block
}

type Block int

const (
	Host   Block = iota // sysblock — BELIEVED. drivers live here (registered at handshake).
	Entity             // userspace — CHECKED.
)

// Barrier is ingress/egress + the in-flight SOCKET TABLE (kernel-owned; the CTU holds only fds).
type Barrier interface {
	// Egress: attenuate the horizon, SIGN (custody), PARK in the socket table, send. cuts-with-a-return
	// park; fire-and-forget does not. (move-and-the-kernel-barrier)
	Egress(from c.Target, t c.Transfer) c.Outcome
	// Ingress: VERIFY custody, then — Response: POP the socket table and hand to CTU.Move; Request:
	// identity-gate (principal self-auth / foreign policy — stubbed) and hand to CTU.Move.
	Ingress(from c.Target, t c.Transfer) c.Outcome
}

// Identity — kernel-owned. Authentication is stubbed for now (bootstrap); the principal is self-auth.
type Identity interface {
	Authenticate(from c.Entity) (Principal, c.Outcome)
}

type Principal struct {
	Entity   c.EntityID
	IsRoot   bool // the seeded human principal (self-authorizing)
}

// Handshake is the bootstrap acquire: the kernel discovers a CTU, reads its drivers, and admits them
// into the sysblock. The handshake IS acquire, run with the kernel's seed authority. (bootstrap C)
type Handshake interface {
	Handshake(ctu CTURef) c.Outcome
}

type CTURef struct{ Endpoint string }

// The kernel composes a Think to run the cold path.
type Cold interface {
	Thinker() think.Think
}

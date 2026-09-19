// Package kernel is the OS-independent AUTHORITY.
//
// FRESH BUILD. The kernel owns IDENTITY and the FULL capability table; it guards the BARRIER and is the
// only minter. It does NOT coordinate (that is entity behavior) and does NOT execute (that is the CTU's).
// Loaded out-of-band, seeded by the human. seL4-shaped: a capability authority.
//
// The CTU-facing surface is just TWO FAULTS — Materialize (read) and Remember (write). Everything
// privileged (ACTIVATE / MINT / PROVISION / THINK-compile / admission-consent) is INTERNAL to Materialize;
// the CTU is oblivious to the memory tiers.
//
// AUTHORITY: skyra/specs/2026-08-24/skyra-dna-v11/ownership.md, research/capability-table-and-horizon.md,
// research/authorization-and-registration.md, research/memory-hierarchy-and-faults.md, bootstrap/kernel.md.
package kernel

import (
	c "skyra-v1/src/skyra/skyra-dna-v11/causal"
	"skyra-v1/src/skyra/skyra-dna-v11/think"
)

// Faults is the CTU-facing kernel surface — the ONLY two verbs a CTU names. A fault is a move that
// targets the kernel: move(target=kernel, action=materialize|remember). MATERIALIZE brings a Context
// INTO the cache (the read fault); REMEMBER writes an engram OUT to disk (the write fault). Symmetric.
type Faults interface {
	// Materialize resolves a FAULTING MOVE into a runnable Context. Pass the whole move — it IS the
	// request: Target/Action = what to resolve, Reads = the seed inputs, Writes (=move.ID) = the fd,
	// Contract = the return type. `under` is the horizon the child runs within (the CTU attenuates by
	// REMOVAL first). The kernel does the internal step — ACTIVATE (page-in from disk) / MINT (compile
	// fresh via THINK) / PROVISION (map the horizon) — and the CTU is oblivious to which. May PARK (wake
	// the human under policy) and re-drive on grant.
	//
	// It RETURNS the child Context; the CTU (the OWNER) installs it in its own map and re-drives the move
	// — trap → handle → retry, idempotent on move.ID (a repeated fault overwrites the same slot, no
	// duplicate). The kernel never writes CTU memory. (memory-hierarchy-and-faults)
	Materialize(move c.Move, under c.Horizon) (*c.Context, c.Outcome)
	// Remember persists an engram (open: and/or a checkpointed Context) to disk — the write / page-out.
	Remember(ctx *c.Context) c.Outcome
}

// ColdEntry is the USER-facing cold path — the principal issues move(entity=kernel, request). Distinct
// from the CTU faults: it materializes + runs + returns the CAPABILITY the user then HOLDS for the hot
// path, plus the result. (think-act-paths)
type ColdEntry interface {
	Handle(from c.Entity, request c.Observation) (held c.Capability, result c.Observation, out c.Outcome)
}

// ---- INTERNAL kernel operations (behind Materialize; a CTU NEVER issues these) ------------------

// Table is the FULL capability table — kernel-internal. The CTU sees only its provisioned horizon.
// ACTIVATE (load an EXISTING engram from disk) vs MINT (compile a FRESH one via THINK) is the disk
// hit/miss choice the kernel makes inside Materialize.
type Table interface {
	Lookup(t c.Target) (Resource, bool)
	// Mint ORIGINATES authority — registers a capability (a macaroon) into the table. Internal.
	Mint(cap c.Capability) c.Outcome
	// Provision MAPS a capability over a target into a horizon — WIDENING (Bound only narrows). Internal.
	Provision(t c.Target, principal c.EntityID) (c.Capability, c.Outcome)
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
	Egress(from c.Target, m c.Move) c.Outcome
	// Ingress: VERIFY custody, then — Response: POP the socket table and hand to CTU.Transfer; Request:
	// identity-gate (principal self-auth / foreign policy — stubbed) and hand to CTU.Transfer.
	Ingress(from c.Target, m c.Move) c.Outcome
}

// Identity — kernel-owned. Authentication is stubbed for now (bootstrap); the principal is self-auth.
type Identity interface {
	Authenticate(from c.Entity) (Principal, c.Outcome)
}

type Principal struct {
	Entity c.EntityID
	IsRoot bool // the seeded human principal (self-authorizing)
}

// Handshake is the bootstrap acquire: the kernel discovers a CTU, reads its drivers, and admits them
// into the sysblock. The handshake IS acquire, run with the kernel's seed authority. (bootstrap C)
type Handshake interface {
	Handshake(ctu CTURef) c.Outcome
}

type CTURef struct{ Endpoint string }

// Cold composes a Think — the kernel invokes it inside Materialize's MINT step to compile a fresh body.
type Cold interface {
	Thinker() think.Think
}

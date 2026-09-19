// Package kernel is the OS-independent AUTHORITY — and in v12, the COLD path.
//
// The kernel owns IDENTITY and the FULL capability table; it is the ONLY minter and the ONLY creator of a
// horizon (BOUND). It guards the BARRIER (custody + socket table) and holds the engram DISK. It does NOT
// coordinate (entity behavior) and does NOT execute (the CTU's). Seeded by the human; seL4-shaped.
//
// The CTU-facing surface is the CTU-OWNED trap seam (ctu.Handler + the Kind vector table); the kernel is the
// HANDLER BLOCK behind it (08-27 kernel-handler-block.md). The kernel does not DEFINE the CTU's interface —
// it IMPLEMENTS the handlers and INSTALLs them at boot (ctu.Install, alongside driver registration). Its
// first two vectors are Materialize (read) / Remember (write). Behind Materialize the kernel runs the COLD
// PATH: COMPILE (think) → BOUND (create the workflow horizon) → MINT (the held ticket) → run → REMEMBER
// (engram). This file exposes those steps for the comb.
//
// AUTHORITY: specs/2026-08-24/skyra-dna-v11/{ownership,capability-table-and-horizon,authorization-and-
// registration,memory-hierarchy-and-faults,bootstrap} + specs/2026-08-27/horizon-and-attention.
package kernel

import (
	c "skyra-v1/src/skyra/skyra-dna-v12/causal"
	"skyra-v1/src/skyra/skyra-dna-v12/engram"
	"skyra-v1/src/skyra/skyra-dna-v12/think"
)

// ---- the handler block: what the kernel installs behind the CTU's vectors ----
//
// These are the kernel's HANDLER IMPLEMENTATIONS behind the CTU-owned trap seam (ctu.Handler). Each satisfies
// one Kind vector and is Installed into the CTU's trap table at boot. The signatures below are the per-handler
// SHAPES (what each needs); reconciling them to ctu.Fault's uniform frame is deferred to when the handler
// BODIES are written — the boundary (the seam + table) is set first, on purpose (kernel-handler-block.md).

// Faults names the kernel's first two vector handlers — Materialize (read) / Remember (write). A fault is
// move(target=kernel, action=materialize|remember); the CTU raises the vector, this block implements it.
type Faults interface {
	// Materialize resolves a FAULTING MOVE into a runnable Context. Pass the whole move — it IS the request.
	// The kernel runs the RESOLVE CASCADE, oblivious to the CTU:
	//     cache HIT      → (already resident; the CTU would not have faulted)
	//     BodyCache/disk → engram.Activate(BodyID)          the WARM page-in
	//     disk MISS      → COLD PATH: think.Compile → Bound → Mint                  the deepest fault
	// then it PROVISIONs the horizon onto a fresh Context and returns it. The CTU (owner) installs it and
	// re-drives — idempotent on move.ID. The kernel never writes CTU memory. May PARK (wake the human).
	// The wire unit IS the Move (identity + authority + TargetFd/Action + ReturnFd — inlined, no Possibility
	// struct); Materialize is handed that faulting Move (+ resolved Body) to BUILD the Context. When wired via
	// the trap seam, it returns the Context through ctu.Fault.Result; this direct signature is the same fact.
	Materialize(move c.Move, under c.Horizon) (*c.Context, c.Outcome)
	// Remember persists the compiled Body to the engram disk (engram.Persist) — the WRITE fault / page-out.
	// It is what turns a future run from COLD into WARM. (COMB: does it ever checkpoint a live Context? no,
	// this phase — see engram.Engram.)
	Remember(ctx *c.Context) c.Outcome
}

// ---- the cold path (the v12 focus) -------------------------------------------

// Cold is the USER-facing cold entry AND its decomposition. Handle is what a principal calls with an
// unstructured request; the three inner steps are the kernel work behind it, surfaced for the comb.
type Cold interface {
	// Handle runs the full cold path for a NEW request and returns the HELD ticket (for the hot path next
	// time) + the result. = the deepest fault, driven from the top instead of trapped from a running move.
	Handle(from c.Entity, request c.Observation) (held c.Capability, result c.Observation, out c.Outcome)
}

// Provisioning is the kernel's authority-origination surface — the two things ONLY the kernel may do, and
// the CTU never touches. This is where v11's CTU-side bound() moves to (08-27).
type Provisioning interface {
	// Bound CREATES the workflow HORIZON (08-27). For each (Target, Action) the compiled Body needs, it
	// checks the principal's authority (Table.Lookup) and includes the capability. The TOTAL set IS the
	// horizon — workflow-scoped, created up front, not narrowed per node. Kernel authority; the CTU never
	// calls this. (COMB: is the horizon ever WIDENED mid-run by a fault→provision, or always whole up front?)
	Bound(body c.Body, principal c.EntityID) (c.Horizon, c.Outcome)
	// Mint ORIGINATES the HELD ticket — the SINGLE capability the user presents to re-invoke this workflow
	// on the HOT path (offline-verifiable). DISTINCT from the horizon: the horizon is the workflow's INTERNAL
	// authority; the ticket is the ENTRY. Signs it into the table. Only the kernel mints.
	Mint(root c.Target, principal c.EntityID) (c.Capability, c.Outcome)
}

// Thinker hands the kernel its compiler — invoked inside the cold path's COMPILE step. THINK proposes a
// Body; the kernel Bounds/Mints around it. Kept a seam so THINK stays a swappable proposer.
type Thinker interface {
	Thinker() think.Think
}

// ---- internal: the registry (behind the cold path) ---------------------------

// Table is the FULL capability table — kernel-internal. The CTU sees only its provisioned horizon. Lookup
// answers BOTH registration (does it EXIST) and authorization (is the principal PERMITTED) — the two axes.
// Bound consults it per needed (Target, Action).
type Table interface {
	// Lookup returns the registry entry for a target and whether the principal is authorized for an action.
	Lookup(t c.Target, action c.Name, principal c.EntityID) (Resource, bool)
	// Register admits a target into the registry (existence) — e.g. a driver at handshake, or a novel
	// capability the acquire path admitted. Registration is NOT authorization (that is Bound/Lookup).
	Register(r Resource) c.Outcome
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
	Entity              // userspace — CHECKED.
)

// ---- the barrier: custody + the socket table ---------------------------------

// Barrier is ingress/egress + the in-flight SOCKET TABLE (kernel-owned; the CTU holds only fds = MoveIDs).
// The forward wire unit is a Move (the activation); the return leg is a Landing. The socket table keys on the
// activation's fd (ReturnFd, a MoveID).
type Barrier interface {
	// Egress: SIGN the activation (custody), PARK it in the socket table (an invoke-with-a-return;
	// fire-and-forget does not park), and send. The signature makes provenance unforgeable across the fabric.
	Egress(from c.Target, mv c.Move) c.Outcome
	// Ingress: VERIFY custody, then — return leg (a Landing): POP the socket table, hand to the CTU; forward
	// leg (a Move): identity-gate (Identity.Authenticate: principal self-auth / foreign policy) and hand to
	// CTU.Transfer.
	Ingress(from c.Target, mv c.Move) c.Outcome
}

// ---- identity & bootstrap ----------------------------------------------------

// Identity — kernel-owned. Authentication is stubbed at bootstrap; the seeded human principal is self-auth.
type Identity interface {
	Authenticate(from c.Entity) (Principal, c.Outcome)
}

type Principal struct {
	Entity c.EntityID
	IsRoot bool // the seeded human principal (self-authorizing)
}

// Handshake is the bootstrap acquire: the kernel discovers a CTU, reads its advertised drivers, and
// REGISTERs them into the sysblock. The handshake IS acquire, run with the kernel's seed authority.
type Handshake interface {
	Handshake(ctu CTURef) c.Outcome
}

type CTURef struct{ Endpoint string }

// Disk hands the kernel its engram store — the one seam to the persistent tier. Held by the kernel, never
// by the CTU. (engram.Store is Activate/Persist/Evict.)
type Disk interface {
	Store() engram.Store
}

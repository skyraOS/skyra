// Package causal holds the types that cross boundaries in v12.
//
// CARRIED from v11 (proven hot path) with the 2026-08-27 folds applied: the HORIZON is workflow-scoped
// and kernel-created; a Context names its workflow ROOT. The v12 focus is the COLD path (kernel/, think/,
// engram/), so this vocabulary is mostly settled — it is the shared ground the cold path builds on.
//
// AUTHORITY (current): skyra/specs/2026-08-27/move-and-possibility.md + register-lifecycle-and-transition.md
// + skyra/specs/2026-09-01/ (the wire Move, three-table state, the reactor, the fault path). Older docs
// (v11 execution model, 08-27 pre-collapse) are HISTORICAL — see the 08-27 README banner. Docs win.
package causal

// Outcome is the total result of a step. FAULT = a name resolved to nothing PROVISIONED (a page fault →
// cross to the kernel; the deepest fault is the COLD path). NOTYET = operands / a parked cut not landed.
type Outcome int

const (
	NotYet Outcome = iota
	Success
	Failure
	Fault
)

// ---- identity & addressing: entity → host → address --------------------------

type EntityID struct{ Value string }
type HostID struct{ Value string }
type Address struct{ Value string } // VOLATILE, host-resolved; never stored on the entity.

// Entity is WHO — a userspace identity. MULTI-HOMED: one or more home hosts resolve its address via
// host(entity) locally. The human is an Entity too — uniform in structure, special only in authority.
type Entity struct {
	ID    EntityID
	Homes []HostID
}

type Name struct{ Value string }

// Target names a move's destination: an entity, resolved via its host, by name.
type Target struct {
	Entity EntityID
	Host   HostID
	Name   Name
}

// ---- capability (macaroon-shaped) = the unit of a horizon --------------------

// Capability is FINE-GRAINED and TIMESTAMPED: it authorizes ONE action on ONE target, until ExpiresAt.
// (Target, Action) is DEFINED by the target's driver. Minted ONLY by the kernel. Verified OFFLINE (sig +
// not-expired + target/action match) at ingress. NO CAVEATS — attenuation is by REMOVAL.
type Capability struct {
	Target    Target
	Action    Name
	ExpiresAt int64
	Sig       Signature
}

type Signature struct{ Bytes []byte }

// Horizon is the PROVISIONED VIEW — the total permission set of a ROOT / WORKFLOW (08-27). CREATED BY THE
// KERNEL (kernel.Bound), carried ON THE WIRE (inlined in a Move), verified offline by the CTU. Authority travels
// WITH the work because a remote CTU CANNOT reconstruct it — there is no shared kernel to call; that is the
// whole point of offline-verifiable capabilities. NOT per-node and NOT CTU-narrowed: the whole possibility
// tree of a workflow runs under one horizon. = what the workflow COULD do; what it IS doing is the in-flight set.
// The CTU NEVER constructs one.
type Horizon struct{ Capabilities []Capability }

// ---- Body (the recursive move node) → Context (per-node frame) ---------------
//
// move-and-possibility.md: a MOVE and a POSSIBILITY are two faces of ONE edge, joined at the fd
// (parent Move.ID == child ReturnFd). So there is NO separate `Possibility` struct and NO separate
// `MoveTemplate` — a Body IS the move. "moves hold moves, recursive all the way down": a Body holds child
// Bodies (by digest). The ROOT move (no parent, carrying the Horizon) IS the possibility; a leaf (empty
// Children) is a Driver.Perform. Identity + authority (ID/Root/Horizon) is the root move's — inlined onto
// the Context (per-node frame) and the Move (wire), never a struct of its own.

// Body is the ENGRAM node: a content-addressed MOVE (Target·Action·Contract) + its CHILD edges. STATIC,
// shareable, read-only. THINK's output; what REMEMBER persists (engram/) and MINT provisions a horizon for.
// The register topology IS the tree (move-and-possibility.md): a move's OUTPUT register is its KEY in its
// parent (its fd); its INPUTS are the keys of its Children. So there is no separate Reads/Writes and no
// Return field — the return is the DAG sink (the root's output). FAN-OUT is git's object model: a shared
// sub-move is the SAME child digest under two parents, stored once (cursor.com/blog/git-at-any-scale).
type Body struct {
	ID       BodyID
	Target   Target            // WHERE/WHOM this move runs — resolved via host(entity)
	Action   Name              // the method name — checked against the horizon as (Target, Action)
	Contract Contract          // the RUNNABLE half THINK authors; the input/return TYPE is negotiated at ACT
	Children map[RegKey]BodyID // sub-moves this move INVOKES, keyed by the child's fd (its output register),
	//                            referenced BY DIGEST — the content-addressed DAG. empty ⇒ a leaf (Perform).
}

type BodyID struct{ Value string } // content address (digest) — the disk key (a CAS).

// Context is the LOCAL FRAME of a run on ONE machine: identity + authority (INLINED — there is no Possibility
// struct; ID/Root/Horizon ARE the root move's) + the CODE it runs (Body, the root node) + its register file +
// result. MUTABLE, and LOCAL — BORN at its CTU (Materialize builds it FROM a Move) and NEVER crosses the
// barrier. Contexts do not travel; the MOVE travels. Registers are produced LOCALLY (a PARTIAL view) —
// values flow UP the fd chain via ReturnFd, single-assignment keeping each register single-homed. Holds NO
// move state (moves live in the CTU's WorkList). LIVE until REMEMBERed. One run on N machines = N Contexts,
// each resolving its Body into its own i-cache.
type Context struct {
	ID        PossibilityID       // the run identity = the root move's identity (flies on the wire)
	Root      PossibilityID       // the workflow this run belongs to (the horizon's scope); self==Root for a root
	Horizon   Horizon             // the authority this run executes under — INVARIANT across ALL its Contexts
	Body      BodyID              // the ROOT move node this frame runs — resolved from THIS CTU's i-cache
	Registers map[RegKey]Observation // the VALUE-SLOT table: fd → landed value, resting until consumed
	Result    *Observation
	ReturnFd  *MoveID // the parent's fd this run's result fills; destination is reply-to-sender. nil ⇒ root run
}

// PossibilityID is the run's identity = the identity of its ROOT MOVE — the MEMBERSHIP key, the "these moves
// go together" tag. It rides the wire inside a Move's ReturnFd (the parent's MoveID.Possibility); the CHILD
// run's own PossibilityID is minted at the TARGET (content-address of body+inputs), not sent. Deliberately
// the MIDDLE grain: coarser than a move (MoveID adds Writes), finer than a workflow (that is Root — the
// horizon's scope). Membership = Possibility · ordering = the child edges · scope = Root — three tags, three
// jobs; do NOT use one for another. Context creation is idempotent on THIS id (materialize-once; the
// thundering-herd dedup — see the Cache residency latch). ADOPTED (move-and-possibility.md): a CONTENT
// ADDRESS of (root Body + seed inputs) → idempotent create + memoization of a whole RUN. CAVEAT: a
// side-effecting/NOW/RANDOM move is NOT memoizable — purity gates the reuse (carry-over-from-08-24.md).
type PossibilityID struct{ Value string }

// MoveID identifies a runtime move by the register it WRITES — its output register, which is its KEY IN ITS
// PARENT (its fd). Writes IS its identity AND its fd. Scoped by its Possibility for the wire. The WorkList
// keys on it; a child's response lands in the parent's register named by Writes. It is ALSO the whole
// backward pointer a run carries home (ReturnFd): the DESTINATION is not stored — it is reply-to-sender
// (barrier-derived from who sent the activation), so the fd alone suffices. A child echoes it as its Landing.Fd
// (parent Move.ID == the child's ReturnFd — the two faces of one edge; move-and-possibility.md).
type MoveID struct {
	Possibility PossibilityID
	Writes      RegKey
}

// ---- moves: the wire units ---------------------------------------------------

// Move is the WIRE FORMAT — the portable ACTIVATION that crosses the barrier to invoke a child run
// (what-travels.md: "a Move is the portable activation that travels"; move-and-possibility.md: it is the
// CALLER's face of the call edge). It carries NO runtime state — state lives in the Context's tables (Work /
// in-flight / Registers). It names only WHAT to run, under WHAT authority, and WHERE the result lands. A REMOTE
// invoke serializes this and crosses; a LOCAL invoke uses the same descriptor to drive a Driver. Authority
// RIDES here (the Horizon) because a remote CTU cannot reconstruct it — Verify checks it OFFLINE at ingress.
//
// The invoke names (Target, Action); the TARGET resolves its OWN Body from that (the caller knows only the
// NAME — contracts-and-remember.md), so the Body does NOT ride for a plain invoke. (Shipping code to a
// machine that lacks it — a mobile-plan body-blob fetch — is a SEPARATE engram concern: peer-fetch by
// digest, one whole blob expanded locally, never pointer-chased hop by hop. OPEN whether it reuses this type.)
//
// A Move carries no ID of its own: its identity IS its ReturnFd (a MoveID) — parent Move.ID == the child's
// ReturnFd, the two faces of one edge. The WorkList keys on that same MoveID. The CHILD run's PossibilityID
// is NOT on the wire either: it is content-addressed = hash(body + inputs), and the caller knows only the
// Name — so the TARGET mints it locally after resolving its Body and receiving its inputs. Baked in, not sent.
type Move struct {
	Root     PossibilityID // the workflow scope (the horizon's Root) — inherited from the caller
	Horizon  Horizon       // the authority — Verify OFFLINE against this; NOT reconstructable at the target
	TargetFd Target        // WHOM to invoke (the forward fd); the target resolves its own Body from (Target, Action)
	Action   Name          // the method
	ReturnFd *MoveID       // the parent's fd the value lands in (destination is reply-to-sender); nil ⇒ root run
}

// Landing is the RESPONSE wire format — the value flowing UP the fd chain. NOT a Move: the forward leg is an
// activation, the return leg is just {fd, observation}. It writes the value into the parent Context's value
// slot named by Fd (Cache[Fd.Possibility].Registers[Fd.Writes] = LANDED), pops the in-flight entry, and wakes
// dependents. (Retires Move.Header/Move.Returned: request and response are now DIFFERENT types, not one Move.)
type Landing struct {
	Fd    MoveID      // the parent register this value fills (the child echoes the parent's Move.ID)
	Value Observation // the observation flowing up
}

// WorkList is the CTU's READY table — the reservation station: the activations whose inputs are all present
// in their Context's Registers, awaiting issue. It is ONE of the three tables that stand in for a register
// state machine (there is no state flag): READY here, ISSUED in the in-flight/socket table, LANDED in
// Registers. A Landing that completes a move's inputs enqueues it here; issuing it POPS it (the pop IS the
// single-writer claim) and dispatches (local Driver or remote Egress). (CONCURRENCY: the pop/claim + land
// must be atomic once N>1 — the CAS discipline; OPEN.)
type WorkList map[MoveID]Move

// Cache is THIS CTU's local Contexts — the RAM tier, keyed by PossibilityID (this machine's frame for each
// distributed Possibility; a possibility on N machines has a Context in N Caches). It is also the
// RESIDENCY LATCH that dedups concurrent faults on the SAME context (the demand-paging race): a slot is
// tri-state — Absent (no key) → Materializing (fault in flight) → Resident (*Context present). The first
// faulter flips to Materializing and faults; concurrent faulters SEE Materializing and PARK (they do NOT
// re-fault); on return the slot goes Resident and parked moves wake. = a page-table entry with a
// fault-in-progress bit. LAND is CAS-LINEARIZED against the content-addressed store (move-and-possibility.md,
// after Continuity): any CTU can be the writer, no consensus — reads verify against the source of truth, so a
// lost gossip/replication message is harmless. (COMB/deferred: the Materializing marker's concrete shape and
// the CAS compare-key.)
type Cache map[PossibilityID]*Context

// BodyCache is the i-cache: content-addressed compiled bodies. A MISS is the page fault to the kernel:
// ACTIVATE from the engram disk, else MINT fresh via THINK (the COLD path). DISTRIBUTION NOTE: a Move carries
// NO BodyID — the target resolves its Body from (TargetFd, Action), then the BYTES here. For a body a machine
// lacks, ACTIVATE must be able to FETCH it by BodyID (content-addressed — shared
// engram or peer-fetch-by-digest), NOT recompile — else distribution would MINT the same body per machine.
type BodyCache map[BodyID]Body

// ---- wire & value ------------------------------------------------------------

// The VALUE-SLOT table is Context.Registers (map[RegKey]Observation): fd → the LANDED value, RESTING there
// until the parent CONSUMES it (fires). There is NO register-state flag and NO promise-cell struct
// (register-lifecycle-and-transition.md): a register's "state" is WHICH TABLE holds its fd, not a field —
//   in WorkList              ⇒ READY (inputs all present)
//   in the in-flight table   ⇒ ISSUED (dispatched, pending; the socket table for remote, claim-out-of-Work
//                              for local — the single-writer latch)
//   present in Registers      ⇒ LANDED (value resting until the parent consumes; freed on consume, not on land)
//   in NONE                  ⇒ not started, or already consumed-and-freed
// A CONSUMER is Ready when every register it reads is present in Registers. The move's promised return NAME
// lives on its Body node (Body.Contract) — the STRUCTURAL type is discovered by running and written at
// REMEMBER; RETURN guards the landed value against the REMEMBERED contract (contracts-and-remember.md).

type Observation struct {
	Content Content
	Segment Segment
}

type Content struct{ Bytes []byte }
type Segment struct{ By Name }

// Contract is a NAME — what the caller knows at CALL (the invoked child / expected observation name), NOT a
// structural type. The caller never sees a child's return shape up front; the structural return type is
// NEGOTIATED at ACT and DISCOVERED by running, then AUTHORED at REMEMBER (bottom-up from what the run
// composed). RETURN judges the landed Value against the REMEMBERED contract. See contracts-and-remember.md.
type Contract struct{ Name Name }

type RegKey struct {
	Name Name
	Tag  Tag
}
type Tag struct{ Iter int }

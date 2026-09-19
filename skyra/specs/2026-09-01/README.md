> **Canonical identity update:** [register-frame.md](register-frame.md) defines MoveID as action + ordered
> observation params, with one register table and explicit ParentFd return publication. Engine proposal
> types reflect this; algorithms and identity encoding/scope remain open. v12 is unchanged.

# Spec — 2026-09-01

> **REGISTER UPDATE (2026-09-06):** [One register per fd](register-frame.md) records the latest agreed
> shape: `Register{Contract, Children []ChildObservation}`, keyed by MoveID in Memory. Each child entry
> has its fd and optional returned Observation; declared order survives out-of-order completion.
> Contract is schema-only; Action/Params remain on invoke. This supersedes the contract sketch below.

> **CONTRACT UPDATE (2026-09-06):** [Parent and child contracts](body-registers-and-return.md#parent-and-child-contracts-on-the-engram)
> live explicitly on the engram: parentContract is what it owes its parent; childContracts are what its
> children owe it. Both use [Contract = Action + Schema](body-registers-and-return.md#contract--action--schema):
> Action is the sole model-visible method name/header (160-token cap), and Schema is opaque JSON
> (`json.RawMessage`). No separate name/title/description. The return opcode delivers an Observation;
> it is not the contract type. Binding, validation, and lifecycle details remain open.

**Register review:** [register-decision-review.md](register-decision-review.md) collects the earlier
register decisions and current code shapes for the next Body/register design pass. It separates
superseded claims from open proposals; no new decisions are retired by the review.

> **BODY UPDATE (2026-09-06):** [body-registers-and-return.md](body-registers-and-return.md) records Body
> as the register requirements and return specification: which observations are needed to return this
> move's observation. Declaration names/partial contracts remain compatible with structural discovery
> during execution and recording at REMEMBER. Concrete types and satisfaction rules are still open.

> **DECLARATION UPDATE (2026-09-06):** [focus-and-move.md](focus-and-move.md) makes available operations
> explicit as Focus on a live move: Horizon is authority, Focus is visible choices, Dependencies are
> committed relationships. Focus replaces the older Attention terminology. Includes a proposed Move
> shape with `Focus []c.Capability`, selecting existing Horizon capabilities without an OperationDecl
> wrapper; runtime code remains unchanged.

> **UPDATE (2026-09-06):** [engram-loading-and-fault-return.md](engram-loading-and-fault-return.md)
> records the latest fault-path decision: scheduler belongs to the kernel; the kernel can install
> engrams into CTU memory; involuntary faults return to scheduler through the execution-event transport.
> EngramID identifies durable executable backing; MoveID now follows the canonical invocation rule above. Engine/v12 code has
> not yet been reconciled with this decision. It supersedes the conflicting ownership claims below.
> Durable engrams live in configurable user application storage. Kernel.Load(EngramID) performs exact-ID
> loading directly through the filesystem driver, with distinct authority and boot availability for
> recovery; no separate EngramStore interface is required.
> Kernel coordinates/provisions from CTU returns; shared Memory owns resident structures and consistent
> storage operations; CTU executes and writes results through Memory. New invocations and recovery
> notifications also enter kernel coordination.

> **UPDATE (2026-09-05):** [ingress-and-fire.md](ingress-and-fire.md) records the latest execution-ownership
> decision: Scheduler owns one map `MoveID → State`, merging Work/InFlight and retaining completed
> records until retirement. The CTU owns contexts/registers/results and receives ready batches through
> `Fire(fds)`. It supersedes the older ownership/table split below
> and in `the-reactor.md`. The engine proposal reflects this split as of 2026-09-06; algorithms and
> authority-check placement remain open. The v12 surface is unchanged.

**Next focus:** [closed-loop-outstanding.md](closed-loop-outstanding.md) — close the fault/recovery path
through Scheduler, Memory, MMU, and CTU, then verify the local execution loop.

**Date:** 2026-09-01
**Status: FOCUS / OPEN.** Carries from `specs/2026-08-27/` (the move/possibility collapse, the wire-format
Move, the three-table state model, the two-part reactor). This spec names **the next thing to work: THE FAULT
PATH** — demand paging (Materialize), its clean separation from the dataflow wait, and the per-access MMU
placement that falls out of the same CPU analogy.

```text
Docs in this spec:
  README.md          (this) the fault path — demand paging vs dataflow wait; the per-access MMU.
  the-reactor.md     DECISION. the engine is a two-part reactor (ISSUE + LAND), the 4→2 collapse. authoritative.
  operand-model.md   OPEN. seam A — where a leaf's inputs come from. THE issue to resolve next.
```

---

## What to focus on next — the fault path

The engine is a reactor (ISSUE + LAND). The one part NOT yet worked is **what happens when a move can't
proceed** — and the CPU has a precise, two-headed answer we should mirror exactly. There are **two different
"don't have it"s**, and conflating them is the trap:

```text
                    WHAT IS MISSING              WHO RESOLVES        SKYRA
PAGE FAULT          the memory BACKING is not    the KERNEL —        Body/Context not resident → Materialize
(a trap)            resident (RAM↔disk miss)     fetch a lower tier, trap → page-in from the ENGRAM (WARM) or
                                                 then RETRY          COLD-compile (THINK) → re-drive (idempotent)
DATAFLOW WAIT       the operand is not PRODUCED  in-pipeline —       input register not yet LANDED → the move is
(not a trap)        yet (a sibling hasn't run)   NO kernel; wait     simply not Ready; a sibling's Landing wakes it
```

- **PAGE FAULT = a memory-hierarchy miss → go to the kernel → retry.** This is Skyra's `Fault` /
  `Materialize` vector: the CPU (CTU) looks for a Body/Context and doesn't have it resident, traps to the
  kernel (involuntary — the hardware yanks it there, not a syscall), the kernel pages it in (engram, or the
  COLD path), and the move re-drives. Demand paging, verbatim.
- **DATAFLOW WAIT = an unproduced operand → no kernel, just wait.** This is `NotYet`: the value hasn't been
  computed by another move; the move sits out of `WorkList` until its inputs Land. The reactor handles it;
  the kernel never sees it.

**Keep these two apart.** A page fault crosses to the kernel; a dataflow wait never does.

## The MMU, per-access — the same analogy, the other half

Falls out of the same CPU model (captured 09-01 as `register-lifecycle-and-transition.md` seam I +
`kernel-handler-block.md` refinement, to be worked here): the MMU is **inline per access**, not a one-time
ingress gate. Three ports, only one of which the current framing names:

```text
DTLB      Interpret + Verify at ISSUE (PER MOVE)      "does this run's Horizon authorize THIS (TargetFd,
                                                       Action)?" — recurs every issue; a MISS faults precisely.
                                                       ← the real per-access check, MISSING from today's framing.
ITLB      Verify at the barrier CROSSING              a remote invoke's own (Target, Action), at ingress.
CR3-load  install/validate the Horizon at MATERIALIZE coarse, once per context (the address-space switch).
```

Today `causal.go`/`ctu.go` label only ingress `Verify` "the MMU." The work: **move the per-access
authorization into the ISSUE step** so it recurs per move (a protection-fault-on-miss), leaving ingress as the
crossing check and materialize as the CR3-load.

## The work items (what this spec will resolve)

```text
1. THE FAULT FRAME + RE-DRIVE. nail the Materialize path end to end: ISSUE hits an unresolved Body → raise
   the fault vector → kernel Faults.Materialize (engram.Activate WARM, else COLD think.Compile→Bound→Mint) →
   provision Context → CTU installs + re-drives (idempotent on move.ID). the ctu.Fault frame shape (deferred
   in kernel-handler-block.md) gets fixed HERE.
2. FAULT vs NOTYET vs FAILURE — the three non-success outcomes, kept distinct. Fault = trap to kernel (page
   in); NotYet = dataflow wait (no kernel); Failure = a landed errno Observation (a value, not a trap). does
   the three-table model need a FAULTING table, or is a fault purely a trap-and-retry that leaves the fd in no
   table? (register-lifecycle seam E.)
3. THE MMU SPLIT — move the per-access authorization to ISSUE (DTLB); keep ingress Verify (ITLB) and the
   materialize CR3-load. relabel in causal.go/ctu.go. (register-lifecycle seam I.)
4. DEMAND-PAGING RACE — N moves faulting on the same not-resident Body/Context: the residency latch
   (Absent→Materializing→Resident) dedups; single faulter pages in, the rest park. tie to the CAS/concurrency
   discipline. (context-identity-and-races.md already seeded this.)
```

## Still-blocking, carried from 08-27 (not this spec's focus, but gating)

```text
· SEAM A — the operand model: what a move READS beyond its children's returns (leaf inputs, constants,
  a parent's datum threaded down). gates the ISSUE "ready" predicate; unresolved.
· SEAM B — the flip index: how LAND finds "the moves that read register R" to wake them.
  the sole coupling between the reactor's two halves.
```

These two still block writing `Transition` at all; the fault path is the layer that sits ON the reactor, so
expect to touch A/B while working item 1.

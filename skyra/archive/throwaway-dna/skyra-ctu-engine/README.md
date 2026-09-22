# skyra-ctu-engine — PROPOSAL

> **Reviewed current docs:** [Surviving design notes](<../../../specs/2026-09-07 /README.md>) selectively reconcile the
> August 24, August 27, and September 1 specs. [The audit](<../../../specs/2026-09-07 /source-review.md>) accounts for every
> source file. Use these notes over conflicting historical prose below or in dated specs.

> **Current design thread (2026-09-07):** [Egress, entity, and host](<../../../specs/2026-09-07 /host-entity-address-horizon/README.md>).

> **Resume here:** [2026-09-06 checkpoint](checkpoint.md) records the latest loop decisions and the
> egress/host/entity routing discussion, including decisions not yet reconciled into code below.

**Date:** 2026-09-06  
**Status: partial implementation probe, not an executable runtime.** The proposal now reflects
[one register per fd](../../../specs/2026-09-01/register-frame.md),
[Focus and Move](../../../specs/2026-09-01/focus-and-move.md), and
[kernel loading and shared Memory](../../../specs/2026-09-01/engram-loading-and-fault-return.md).
Boot, ordered child publication, and scheduler claim/completion transitions run. Admission,
resolution, dispatch, and return remain explicitly unimplemented. See [DECISIONS.md](DECISIONS.md)
and [boundary-crossings.md](boundary-crossings.md) for results and blockers.
The follow-up [surface proposal](surface-proposal.md) explicitly lists new structs/interfaces and
changes needed for a serial preloaded loop; it has not been applied to the implementation.
v12 remains unchanged; this package reuses its possibility,
capability, and observation vocabulary, but defines its own MoveID, Fd, Contract, and Register.

> **Latest loop revision:** [Pending and Move returns](surface-proposal.md#latest-decided-loop-pending-move-in--move-out)
> supersede the older execution-event flow below. CTU determines readiness and executes; scheduler
> publishes returned observations through Memory. Issued is removed from code. Claim/Claimed/Landed
> and Event remain legacy probe surfaces awaiting reconciliation before the next handoff.

## Overnight implementation probe

The [work order](overnight.md) targets a small local execution loop using the current surface in
`engine.go`. Follow the [source implementation rules](../README.md#implementation-probes): no new
structs or fields, no reshaped signatures, and orchestration through specified methods. Method
bodies and local implementation choices are the implementer's responsibility.

The surface is a proposal being tested. Discover and document where it fails to support the loop;
do not silently extend it or bypass ownership to manufacture a successful demo. Keep proposed
interface changes in the report for review. The open questions below are context, not a required
implementation sequence.

The [current proposal](surface-proposal.md#accepted-distinction-ingress-dispatch-and-completion)
now defines Move as the universal wire envelope and one ingress that receives it and splits on opcode: Call provisions an fd entry; Ret reaches the
existing return behavior. Scheduler claims eligible entries and `Fire(batch)` is their handoff.
Registers stay resident in Memory. Issued, Landed, and Faulted are internal scheduler events.
Landed reports publication already performed, so it never triggers a second register write.

Before another handoff, review the [remaining interface questions](surface-proposal.md#remaining-issues-before-the-next-handoff).
These are design decisions to work through, not instructions prescribing implementation algorithms.

**Latest identity/resolution decision:** [MoveID and EngramID are the same identity](surface-proposal.md#decided-one-identity-resident-lookup-then-durable-resolution).
Scheduler looks up the arriving MoveID in Memory; on absence, kernel resolves the same ID through the
filesystem driver or obtains executable behavior through declaration. The filesystem driver owns `Resolve(id MoveID) (*Fd, error)`; kernel installs the returned fd.
Creating a register alone does not supply executable behavior. The separate ID types/Move.Engram field in engine.go await reconciliation
with the resolution interface; the descriptions below reflect that older code where they differ.

## Register and contract

```go
type Contract struct {
    Schema json.RawMessage
}

type Register struct {
    Contract Contract
    Children []ChildObservation
}

type ChildObservation struct {
    Fd    MoveID
    Value *c.Observation
}
```

MoveID canonically represents action + ordered observation params. The fd is that MoveID;
there is no separate fd lookup table. Canonical encoding and live-state sharing/scope remain open.
Target is outside the identity, requiring unambiguous actions within the lookup scope.

Memory holds one fd table:

```go
type Fd struct {
    Move     Move
    Register *Register
}

type Memory struct {
    Fds map[MoveID]*Fd
}

func (ct *CTU) Fire(fds []*Fd)
```

Kernel provisions the move and register together through Memory. Scheduler claims their MoveIDs;
the selected entries are handed to CTU.Fire. The batch contains references to the one Memory-owned
table, not a copy of scheduling state. CTU checks access and executes each entry's Move.Op using its
Register. Installation and batch-selection methods remain to be specified.

One register is a return frame collecting ordered child
observations, not one scalar result slot. Each child entry has an execution fd and a value that is nil
until arrival. Declaration order survives completion order. Arrival alone does not establish contract
satisfaction. The CTU uses the observations to produce one return to the parent.

Action and `Params []c.Observation` remain on invoke; Contract contains only the opaque schema. Action is the single
model-visible method name, capped at 160 tokens in the design; tokenization/enforcement remain open.
Move carries `Focus []c.Capability` selected from Horizon, optional `Engram *EngramID` backing, and
`ParentFd *MoveID` (nil for root). ParentFd selects the parent register; the move's own Fd selects its
entry in that register's ordered Children. These execution addresses survive fault recovery.
ExecCtx uses the Register from the handed-off Fd. CTU does not resolve an executable during dispatch.
Ret.Value carries the proposed output; its contract comes from the execution frame. Ret.Fire checks
requirements and output, calls `Memory.Land(parentFd, childFd, value)`, then emits Landed. Land fills
the matching ordered child entry. Return construction/validation and root output remain open.

## Ownership and flow

| Surface | Responsibility |
|---|---|
| Kernel | Contains Scheduler; owns Load and authority; provisions shared Memory through its operations |
| Kernel.Sched | Single writer of MoveID → State; claims before dispatch, parks faults, retains completion until retirement |
| Memory | Shared resident storage, register consistency, lookup, publication, and lifetime rules |
| CTU | Executes ready, claimed fd batches using Memory and MMU; reports execution events |
| MMU / Touching | Check declared target/action access before execution; never mint grants |
| Driver | Synchronous mechanism; CTU's eventual async adapter must preserve Observation and Outcome |
| Kernel.FS | The filesystem Driver used directly for exact-ID loading; no EngramStore/Pager layer |
| Runtime.Boot | Wires one Memory instance into Kernel and CTU and installs kernel recovery handlers |

Runtime.Memory is the authoritative boot-time reference, allocated if omitted. Supply Kernel and CTU
before Boot; provide any preloaded Memory through Runtime.Memory. Boot does not start a scheduler loop,
select a storage directory, or install an actual filesystem driver.

```text
Kernel provisions fd entries → Scheduler claims MoveIDs → hand off selected entries
CTU.Fire([]*Fd) → MMU.Check → Op.Fire
                         ← Issued / Landed / Faulted execution events

Faulted{MoveID, missing EngramID} → scheduler parks and routes kernel recovery
Kernel.Load(EngramID) → exact filesystem path → FS read → decode/verify → Memory installation
                     → readiness established → scheduler reclaims the same MoveID
```

Kernel.Load is an explicit stub. StorageRoot is configurable; the intended macOS default is
`~/Library/Application Support/Skyra/`, but platform path resolution is not implemented. Missing files,
I/O errors, and invalid contents cannot count as a successful load. Loading infrastructure must be
available at boot independently of pageable engrams. Sharing the filesystem mechanism does not grant
workflows the kernel's authority over application data.

Involuntary faults return through the execution-event path; CTU no longer owns a handler table or
independently starts the same recovery. Kernel owns the recovery table. Explicit Trap remains a stub
whose remaining role is open. Pager and the old Context result channel are removed.

## Open before implementation

- Durable Body/Engram struct, canonical encoding, and Memory.Install argument type. EngramID is a
  provisional string carrier; no fake complete engram codec or placeholder executable Body is introduced.
- Memory installation and pre-handoff selection of Fd entries; unknown fd versus missing backing.
  Move and Register now have an explicit home together in the fd table.
- Canonical Action/Params encoding and table scope; register sharing for identical calls, repeated
  effects, and multiple parents. Root output and input injection. Engine no longer uses MoveID.Writes
  or RegKey parameters; v12 retains its earlier types.
- Schema discovery, validation, unknown schemas, return construction, and satisfaction/readiness.
- Event delivery, fault resolution states/results, duplicate handling, and failed handoff recovery.
  No separate Attempt ID is adopted; Materialize preserves MoveID and usable inputs.
- Memory publication before consumer admission, register lifetime, shared-load dedup, and safe eviction.
- Completed-record retention, quiescence, trace retention, and REMEMBER compilation/persistence.

Tests cover shared boot memory, ordered publication and duplicate/conflict handling, scheduler
claim/completion transitions, and explicit failures at blocked loop boundaries. Build, vet, and tests
pass; there is no running local graph or root observation. Scheduler tests seed lifecycle preconditions
directly for unit testing; that is not a runtime admission path.

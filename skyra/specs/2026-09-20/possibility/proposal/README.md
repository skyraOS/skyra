# Possibility and Horizon — code on paper

Date: 2026-09-20

Status: Discussion draft. These declarations are not an implementation. Like the
earlier dated proposals, Go files use `//go:build ignore`; cross-directory type
references are conceptual and package wiring remains provisional.

## Starting point

Every Move enters interpretation. The runtime establishes the caller from its
trusted execution binding and resolves the Possibility. Horizon is the first
semantic gate: it admits or traps the Move before that Move can settle
dependencies or cause execution. The gate may read invocation relationships to
make that decision. Commit records the prepared transition; external execution
happens after the state lock is released.

Horizon needs a trap/fault gate and a reference to a TrapController. The
TrapController is an addressable Surface. It is not an arbitrary callback run
inside the authority check. Controller work is prepared and committed before
being dispatched, and its emitted Moves enter interpretation too.

## Files

The declarations are now composed into three proposed trust roles. These are
responsibility boundaries to test, not a decision to create three OS processes
or to label any one of them a kernel.

```text
proposal/
    authority-owner/
        authority-checker/
        possibility/
        horizon/
        state-owner/
        surface-registry/
        trap-controller/
        reference/
    execution-broker/
    agent-surface-worker/
        surface/
```

- [Authority owner](authority-owner/README.md): provisioning, authoritative
  records, registry mutation, and runtime trap handling.
- [Execution broker](execution-broker/README.md): committed bindings, restricted
  resource preparation, child lifecycle, and response attribution.
- [Agent/Surface worker](agent-surface-worker/README.md): capability behavior
  executed under the authority it was given.

The broker has no existing standalone declaration to move: its behavior was
spread across the earlier execution-pool, process, and receive-pump discussions.
Its role note makes that missing composition explicit. The reference snapshot
is parked under authority-owner as historical source material; its contents span
all roles and are not being assigned wholesale to authority.

| File | Starting point and proposed change |
|---|---|
| [authority-checker/](authority-owner/authority-checker/README.md) | `Check(Move)` returns `AuthorizedMove{Move, Horizon}` or an error; namespace ownership and validity rules remain open. |
| [possibility/proposal.go](authority-owner/possibility/proposal.go) | September 18 Possibility shape, extended with a Possibility-owned invocation map. Frame fields adapted from DNA-17. |
| [horizon/proposal.go](authority-owner/horizon/proposal.go) | Existing generation/grant shape, plus a registered TrapController reference. Gate signatures remain open. |
| [trap-controller/proposal.go](authority-owner/trap-controller/proposal.go) | TrapController represented as a Surface using the September 19 structure. |
| [surface/proposal.go](agent-surface-worker/surface/proposal.go) | September 19 Surface/contract structs and the September 18 Driver interface that proposal retained. |
| [surface-registry/](authority-owner/surface-registry/README.md) | Registered Surface identities and handler bindings, resolved through Horizon during interpretation. |
| [state-owner/](authority-owner/state-owner/README.md) | Canonical state and a long-lived state-turn queue/pool, separate from execution. |
| [reference/dna17.go](authority-owner/reference/dna17.go) | Verbatim current DNA-17 contract declarations, with build exclusion and package header added. Reference only. |

The proposal retains the dated specs' UUID identities. DNA-17's reference uses
string identities and a uint64 Identifier. This draft does not settle their
encoding or equate MoveID, invocation identity, execution handle, and OS fd.

## Ownership under discussion

- Possibility owns its Horizon and invocation frames logically. StateOwner remains
  the canonical storage and mutation authority; callers do not receive mutable
  map aliases.
- The invocation map indexes a branching call stack: each frame has a recorded
  return destination. Concurrent children require more than DNA-17's single
  `waitingOn` slot. The precise dependency representation remains open.
- Horizon governs executable resolution and permission together. The new
  SurfaceRegistry draft proposes a separate registered-binding table contained
  by StateOwner; per-Possibility handler binding and replacement remain open.
- The executor consumes committed work. It does not independently grant authority.
- TrapController handles admitted trap/fault work. Naming it does not itself
  define its privileges or allow it to widen Horizon.

## Trap/fault questions to resolve next

The reference is represented provisionally as a SurfaceID, rather than a pointer
to a live handler, following the canonical-state identity convention. Its exact
placement and resolution contract are open.

1. What are the gate's input and result types, including its trusted caller context?
2. What distinguishes a trapped Move from a reported execution fault?
3. Does the controller return a denial, prepare corrective work, or choose between
   them? How does its result answer the original invocation?
4. How is access to the TrapController provisioned so a denied Move can be handled
   without bypassing authority or recursively trapping controller dispatch?
5. What happens if the controller itself fails or emits a denied Move?
6. The SurfaceRegistry draft proposes resolving grants against a separate table
   during interpretation. What pins its binding through commit and execution?

No new gate, controller, or state-access method signatures are invented yet.
The purpose of these declarations is to make the next discussion concrete.

## Sources

- [DNA-17 contracts](../../../../src/skyra/skyra-dna-17/contracts.go)
- [DNA-17 invocation records](../../../../src/skyra/skyra-dna-17/state.go)
- [September 18 Possibility](../../../2026-09-18/possibility/proposal.go)
- [September 18 Surface and Driver](../../../2026-09-18/surface/proposal.go)
- [September 19 revised Surface](../../../2026-09-19/surface.md)
- [September 19 phase rules](../../../2026-09-19/improvements.md)

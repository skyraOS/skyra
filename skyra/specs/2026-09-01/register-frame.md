# One register per fd

**Date:** 2026-09-06  
**Status: SHAPE AGREED / runtime unimplemented.** Records the accepted register proposal and the
contract simplification leading to it. Supersedes the single-value register shape and the
Action-plus-Schema contract in [body-registers-and-return.md](body-registers-and-return.md).

## Agreed shape

```go
// Canonical representation of Action + ordered observation Params.
type MoveID struct { Value string }

type Contract struct {
    Schema json.RawMessage
}

type Register struct {
    Contract Contract
    Children []ChildObservation
}

type ChildObservation struct {
    Fd    MoveID
    Value *c.Observation // nil until this child returns
}

type Fd struct {
    Move     Move
    Register *Register
}

type Memory struct {
    Fds map[MoveID]*Fd
    // Durable engram storage remains to be specified.
}

func (ct *CTU) Fire(fds []*Fd)
```

This shape is reflected in the engine proposal's types and boot wiring. Execution algorithms remain
unimplemented. MoveID is engine-local; v12 retains its earlier Possibility/Writes identity.

**One table:** the fd is the MoveID itself, directly indexing Memory.Fds. Each entry holds its Move and Register. MoveID canonically
represents action + ordered observation params. There is no separate InvocationKey-to-fd table.
Canonical encoding and the scope/sharing of live registers remain open. Target is outside the identity;
actions must be unambiguous within the chosen lookup scope. Root retains the possibility identity.

One fd identifies **one register that collects the child observations needed to fulfill its return
contract**. The register is the execution's return frame, rather than the earlier single-result slot.
Each child entry carries its canonical MoveID; each child can have its own register.

## Executable handoff

Kernel provisions the Move and Register together through Memory. Scheduler claims their MoveIDs,
then the selected Fd entries are handed to CTU.Fire as a batch of references into Memory.Fds.
CTU checks access and executes each entry's Move.Op, constructing ExecCtx from the supplied Move
and Register. It does not independently resolve executables during dispatch.

There is one fd table; no separate register table is introduced. Installation and selection must
have Memory methods before the loop is implemented. Event delivery remains a separate open seam.
This accepts the Fd shape, not the broader additions in the overnight surface proposal.

## Ingress distinction

MoveID and EngramID are the same identity under the latest
[resolution decision](../../src/skyra/skyra-ctu-engine/surface-proposal.md#decided-one-identity-resident-lookup-then-durable-resolution).
Call ingress looks up that ID in Memory, then kernel resolves persisted behavior or provisions from
declaration when no remembered executable exists. Creating the register alone cannot supply behavior.

Move is the decided universal wire envelope. One ingress receives `Move` and routes incoming
work/results by opcode. Registers stay resident in Memory; Issued, Landed, and Faulted remain internal
execution events for the scheduler. Call provisions a Move/Register entry in
Memory, then readiness and scheduler claim precede Fire(batch). Ret routes to the existing return
behavior using the returning move's context; it does not create a new invocation or replace its
register. Ret.Fire defines publication through Memory. Landed subsequently informs the scheduler
and does not publish again. Scheduler interprets lifecycle events rather than opcodes.

Ingress receiver/result signature, opcode wire encoding, context resolution, event delivery, and
readiness APIs remain open. The ingress input itself is settled as Move.
The [current review checklist](../../src/skyra/skyra-ctu-engine/surface-proposal.md#remaining-issues-before-the-next-handoff)
collects these and the other decisions for the next handoff.

## Many child observations, one return

```text
fd A → Register
         Contract: Comparison schema
         Children, in declared order:
           [0] fd B → Document observation
           [1] fd C → pending

fd B → its own Register
fd C → its own Register
```

When C returns, its observation fills entry 1 of A's register. Children retain declared order,
regardless of completion order; results do not append in arrival order. Once A's requirements are
satisfied, the scheduler can claim/dispatch A. The CTU uses A's child observations to produce one
Comparison observation, which fills A's entry in its parent's register.

Every entry identifies a child execution and holds at most its one returned observation; the register
holds many entries. Non-nil means an observation arrived, not automatically that it satisfied the
required contract. Empty/zero-child returns, root results, and validation rules still need definition.

## Contract and invoke

The proposed invoke carrier contains the call information:

```go
type invoke struct {
    Target c.Target
    Action c.Name
    Params []c.Observation
}
```

Action and parameters belong to the invocation, not Contract. **Contract now contains only Schema.**
The explored ObservationParams contract field is not adopted; Params holds the actual input observations
in order. If a future opcode carries references, resolve them before deriving MoveID. Action remains the single model-visible method name, capped at 160 tokens, with no separate
name/title/description; removing it from Contract does not remove it from invoke or Capability/Focus.

Return uses the contract and the frame's observations to produce an Observation. The schema supplies
the shape, not the values. Ret.Value carries the proposed output. Ret.Fire owns return behavior; its algorithm remains unimplemented.
This uses the existing execution-event transport.

The fd associates a live invocation with its register/return obligation and eventual result. This avoids
copying Action and Params into the contract merely to preserve that association. The rejected
parent-managed "current version"/activation scheme is not revived: observed input/output relationships
and learned schemas are recorded at REMEMBER. Canonical key encoding, contract discovery, and durable encoding of
those relationships remain open; no parameter-keyed schema database is selected here.

## Durable and live ownership

The engram preserves the contract/specification and reusable child structure. Memory holds live fds and
observations; arrivals do not mutate immutable engrams. Kernel/scheduler owns scheduling lifecycle;
CTU executes and accesses registers through Memory's operations.

The earlier parentContract/childContracts terminology can still describe the two ends of a relationship,
but the accepted live shape has one Register.Contract and ordered child-result entries. Do not add a
second live contract-management structure from the older sketches. The exact durable Engram/Body fields
and how child contract references are reached still need reconciliation.

## Move and return delivery

```go
type Move struct {
    Op       Op
    Root     c.PossibilityID
    Fd       MoveID
    ParentFd *MoveID // nil for root
    Horizon  c.Horizon
    Focus    []c.Capability
    Engram   *EngramID
}

type ExecCtx struct {
    Host  *CTU
    Move  *Move
    Frame *Register // entry.Register from the Fd handed to CTU.Fire
}

type Ret struct { Value c.Observation }

// Proposed Memory publication interface; algorithm remains a stub.
func (m *Memory) Land(parentFd MoveID, childFd MoveID, value c.Observation) c.Outcome
```

Ret.Fire reads its frame, checks required child observations and the proposed output against its
contract, then reads Move.ParentFd to select the parent register. Move.Fd identifies the child entry
within that register. Memory.Land fills that entry, preserving declaration order. Only after publication
does Ret.Fire emit Landed; the scheduler records completion and coordinates readiness without writing
registers. A nil ParentFd requires a separate workflow-result destination, still to be designed.

Preserve Fd and ParentFd through materialization/reissue. ParentFd is a live return address and does
not become a durable engram dependency ID. EngramID still identifies remembered executable backing.

## Remaining seams

1. **Identity and scope:** define canonical Action/Params encoding and whether register tables are scoped
   to a possibility. Identical action/params yield the same MoveID: decide register sharing, repeated calls,
   and multiple-parent behavior before implementation. A single ParentFd does not settle fan-out.
   No separate fd translation table or Attempt ID is introduced.
2. **Roles and requirements:** how same-type roles and required input conditions relate to ordered child
   entries and the contract. An array preserves order, but does not define semantic roles by itself.
3. **Return construction and roots:** how child observations produce Ret.Value, how root output is delivered,
   and how asynchronous/remote returns reach the same publication boundary.
4. **Satisfaction/publication:** validation, duplicate/conflicting returns, conditional requirements,
   and synchronization of publication before readiness is reported.
5. **Lifetime:** preserve inputs across faults and asynchronous execution; define consume/free and
   shared-value retention without restoring the old table-membership lifecycle model.
6. **Discovery and persistence:** represent unknown schemas, compile discovered contracts at REMEMBER,
   and encode the reusable Body/Engram without persisting live execution fds as compiled identities.

The accepted shape is **one MoveID/fd → one register → contract plus ordered child observations →
one observation returned to the parent.**

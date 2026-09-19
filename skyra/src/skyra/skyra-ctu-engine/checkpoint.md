# Checkpoint — 2026-09-06

Resume here tomorrow. This checkpoint records the latest conversation decisions, including changes
not yet reconciled into engine.go or the older proposal sections. Do not restart implementation from
the old overnight order before finishing the interface review.

## Where we stopped

We were connecting CTU output back to the loop. A callback/channel was suggested, but the user's
latest direction is to examine the **existing egress boundary and host/entity routing** first.
No callback or channel was selected.

The user asked to see the two host/entity maps the CTU maintains. Inspection found no separate maps
in the current engine or v12. The actual engine surface is:

```go
type CTU struct {
    Self   c.Target
    Memory *Memory
    MMU    MMU
    Local  map[c.Target]Driver
    Out    Egress
}

type Egress interface {
    Egress(from c.Target, mv Move) c.Outcome
}
```

The existing v12 addressing vocabulary is:

```go
type Entity struct {
    ID    EntityID
    Homes []HostID
}

type Target struct {
    Entity EntityID
    Host   HostID
    Name   Name
}

type Address struct { Value string }
```

v12 describes entity → home hosts → address resolution. Local is currently a map of full Target to
Driver. Separate host/entity routing maps have not been declared. **Next discussion: how egress uses
that addressing model to route emitted Moves, including the local return to scheduler ingress.**
Do not assume new routing fields or their ownership before discussing them.

## Latest agreed loop

- Move is the universal wire envelope. One ingress receives Move and splits on opcode.
- Registers stay in Memory. One table is `Memory.Fds map[MoveID]*Fd`, where
  `Fd{Move Move, Register *Register}` holds executable and live register together.
- `CTU.Fire(fds []*Fd)` is the handoff itself. Its argument is a batch of references to selected
  entries, not the entire Memory map. No additional handoff abstraction is needed.
- MoveID and EngramID are the same identity. The arriving Move carries the lookup key.
- Scheduler checks Memory first. On absence, kernel calls the filesystem driver's
  `Resolve(id MoveID) (*Fd, error)` and installs the returned fd through Memory.
- If nothing is remembered, declaration must supply executable behavior. Creating a register alone
  does not make an executable. Missing backing and failed/corrupt reads remain distinct.
- Pending replaces Claimed. It describes handed-off work awaiting return, not proof of dependency
  satisfaction. Issued is removed. Landed is superseded by a Move carrying Ret.
- CTU executes work and emits Moves. Scheduler-side ingress processes those Moves and writes results
  through Memory. The earlier CTU-publication-then-Landed model is superseded.

## Readiness refinement — most recent discussion

Do not restore a generic scheduler → CTU readiness query → scheduler → CTU execution round trip.
Readiness follows opcode semantics:

- Call already carries its invocation dependencies and can fire once resolved and permitted.
- The return path records child observations and holds parent work until its dependency/contract
  requirements permit release to CTU.
- Other opcodes have their own semantics; do not generalize Call or Ret into a universal policy.

The user does not want to design full dependency/contract validation now and asked to stub that
boundary. Keep the parent check separate from recording the observation and releasing work.
A minimal “all declared child slots have observations” rule was suggested for the local probe;
it is a provisional stub choice, not full contract validation or a finalized universal requirement.

```text
Call ingress → resolve/provision → Pending → Fire
Ret ingress  → write child observation through Memory
             → parent dependency/contract check (stub boundary)
             → hold or release parent work to Fire
CTU emits Moves → egress/routing connection under discussion → ingress
```

## What actually runs, and what remains stale

The overnight probe implemented Memory.Land storage behavior and limited scheduler claim/completion
transitions. It did not close a local loop. Scoped tests pass, but test success does not demonstrate
execution, driver resolution, or a root return. No execution/loading loop has been implemented.

Already reflected in code: Fd, Memory.Fds, Fire([]*Fd), and removal of Issued.

Still to reconcile before handoff:

- Separate MoveID/EngramID types and Move.Engram versus the agreed shared identity.
- Generic Kernel.FS Driver surface versus filesystem-specific Resolve returning *Fd.
- Claim/Claimed, Landed, Event, Scheduler.Apply, and Op.Fire's old Event result versus Pending and
  emitted Moves. Legacy tests still exercise portions of the old scheduler surface.
- Comments assigning publication to Ret.Fire before Landed versus scheduler-side return publication.
- Installation and entry-selection methods, concrete ingress ownership/result, and the egress return
  connection. Do not invent a transport to make the old signatures look complete.

## Working style and next handoff

The goal remains a small local loop. Review the boundaries together, then freeze the surface for
another implementation probe. Do not over-specify algorithms or micromanage the implementer.
The implementation order forbids new structs/fields and reshaped interfaces; architectural changes
are made during this design review, not silently invented during the next run.

The remaining checklist in surface-proposal.md also covers root result/completion, non-success
outcomes, canonical identity/revisions/reuse, authority, and entry lifetime. The readiness discussion
above supersedes that checklist where it suggests repeatedly dispatching unready work.
Durable storage implementation, full declaration/REMEMBER, remote transport, and multicore execution
remain separate scope decisions. Egress discussion does not itself expand the local build scope.

## Reading pointers

- [Current proposal and open issues](surface-proposal.md): accepted revisions first; the broad earlier
  proposal below them is unapproved material, not a list of mandatory new types.
- [Initial probe findings](boundary-crossings.md) and [implementation choices](DECISIONS.md): historical
  evidence from the frozen surface at the time.
- [Work order](overnight.md): initial probe completed; next handoff still under design review.
- [Register specification](../../../specs/2026-09-01/register-frame.md): useful background, but its older
  publication/readiness language must be reconciled with this checkpoint.

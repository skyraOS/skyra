# StateOwner

Date: 2026-09-20

Status: Code-on-paper proposal. Restores the StateOwner name for canonical state
ownership while keeping SurfaceRegistry a separate component. No runtime changes.

## StateOwner is a Surface

**StateOwner conforms to the Skyra protocol. It receives Moves and emits Moves
like any other Surface.** It must use the common Surface shape: registered
Identity, Contracts, Dependencies, Driver, IngressProfile, and EgressProfile.
The [declaration](proposal.go) embeds that
[Surface structure](../../agent-surface-worker/surface/proposal.go).

System-level state reads and writes are invoked as Moves through the provisioned
StateOwner binding. Their outcomes are emitted as Moves and return through
runtime interpretation and authority. StateOwner does not expose a separate
public Request protocol or use a private return path to advance another invocation.
An internal Request may still represent the prepared writes within a state turn.

```text
Admitted state-operation Move
    -> StateOwner Surface
    -> state operation
    -> emitted Move
    -> runtime interpretation and authority
```

This Surface role does not give StateOwner permission to widen grants or bypass
admission. The coordinator retains preparation/dependency bookkeeping. StateOwner
implements its primitive reads/writes without recursively emitting another Move
for each primitive access.

The older storage and loop sketch below is retained for discussion. Its placement
of registry, runtime records, and coordination within StateOwner is provisional;
the Surface protocol above is the current requirement. See the
[execution-context refinement](../../../authority-namespace-and-execution-context.md).

## What the prior docs actually say

The [separate-pools sketch](../../../../../2026-09-19/persistent-parent-and-separate-pools.md)
proposes one long-lived state queue/pool with **one worker initially**, separate
from a concurrent execution queue/pool. Reads and writes are phases of the same
state turn. There is no separate read pool and write pool.

DNA-17 implements a single interpreter goroutine plus a mutex-serialized
StateExecutor; it does not implement that proposed StateOwner worker pool.
Its execution-side emission commits also take the mutex. The newer direction
routes execution outcomes back through interpretation instead of preserving an
independent execution-side state-transition path.

The older MessageBroker proposal had a serialized request channel and restricted
profiles, but also Interpretation/Supervisor mailboxes and custody transitions.
The September 18 reconciliation marked that two-mailbox model as superseded.
This proposal carries forward restricted access and the single owner, not those
old mailboxes or separately queued stage commits.

## Proposed ownership

StateOwner stores canonical Possibilities, their Horizons and invocation frames,
Move records, and runtime execution records. It contains SurfaceRegistry as a
distinct component, with registry access serialized by the same state turn.

Logical ownership of the invocation frames stays with each Possibility. There is
no second global invocation map introduced here. Execution records hold runtime
correlation and lifecycle facts; their exact fields and fd relationship are open.

The long-lived state pool processes inputs to interpretation. It is a host for
the state-turn behavior, not a new source of permission: Horizon still supplies
the Possibility's authority. StateOwner owns storage and serialization.

```text
State-turn queue -> one state worker initially
    lock
    interpret
        establish trusted caller / Possibility
        check Horizon and resolve registered binding
        resolve invocation/dependency relationships
        prepare Request
    commit(Request)
    unlock
    hand authorized work to the separate execution pool
```

Prepared Requests do not wait in another commit queue. Keeping interpretation's
reads valid through commit avoids introducing stale decisions between the phases.
The executor, receive pump, and registry Driver never execute under the state lock.

The declaration groups the queue/pool with StateOwner for discussion. Whether the
loop is physically a StateOwner method or a separate runtime coordinator is open;
this grouping does not merge authority, dependency logic, and storage into one
implementation method.

## State access and lifecycle

- Runtime callers receive restricted profiles, not direct maps or a general
  mutable StateOwner reference.
- State operations expressed as Moves still enter interpretation and authority.
  Internal primitive reads/writes implement the turn without recursively queuing
  another Move for every record access.
- Registry registration and Horizon admission must use one consistent view.
- Process objects, pipes, buffers, and live receive loops are execution resources;
  canonical records describe their ownership and lifecycle without storing them
  as durable values.
- Long-lived state is not crash durability. Persistence, replay, and recovery
  remain open.

## Next questions

1. Concrete state-turn input envelope and trusted-origin binding.
2. Restricted read profiles and prepared Request operations.
3. Where the state loop is hosted and how shutdown drains or settles work.
4. Queue capacity and re-entry without the DNA-17 self-enqueue deadlock.
5. Atomic recording and scheduling of committed execution work.
6. Registry binding lifetime and runtime execution-handle records.

## Sources

- [Declaration](proposal.go)
- [SurfaceRegistry proposal](../surface-registry/README.md)
- [DNA-17 StateExecutor](../../../../../../src/skyra/skyra-dna-17/state.go)
- [Older MessageBroker/StateOwner](../../../../../2026-09-18/MessageBroker/README.md)
- [September 18 superseded mailboxes](../../../../../2026-09-18/README.md)
- [Explicit interpret/commit phases](../../../../../2026-09-19/improvements.md)

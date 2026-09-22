# MessageBroker and component state profiles

Updated: 2026-09-18.

Status: proposal for the initial demo. MessageBroker supplies restricted profiles
over the single State Owner. The stateless TransitionUnit profile replaces the
former CustodyState contract. Exact operations and underlying records remain open.

## Ownership

StateOwner owns authoritative runtime state behind the MessageBroker profiles.
Components do not receive general Read/Write authority or direct access to maps.
They receive typed profiles exposing only their permitted state operations.
All profile adapters submit requests to the same serialized state-access loop.

TU forwards the whole Request (Move and current-stage Operations) with bound
routing context. MessageBroker queues it and carries the acknowledgment. StateOwner
validates authority, ownership, freshness, and operation validity, then commits all
of that stage's changes atomically. Enqueueing alone is not a commit acknowledgment.

Every stage commits its required runtime operations before work advances.
Interpreter commits its changes together with Supervisor mailbox placement and
custody. Supervisor commits execution preparation before releasing execution.
StateOwner validates that execution registration follows the committed Interpreter
handoff for the current processing attempt. No cross-stage staging area is needed,
and previously committed operations are not resubmitted as pending operations.

State Owner owns exactly two Move mailboxes: Interpretation and Supervisor.
There is no standalone Admission service. Its former queue is the Interpretation
mailbox. The broker's request channel serializes state operations and is not a
third Move mailbox. Revision reconciliation,
detached read results, operation validation, and acknowledgment after commit remain
part of the state-access contract.

## System and user state

StateOwner contains its request channel, systemState, and userState. systemState
holds the revision, SystemProfiles, continuation map, possibility map, custody,
worker records, and both mailboxes. The continuation map is the map of Moves;
there is no separate Move map.

userState holds user-registered Surfaces, such as Codex, keyed by Surface identity,
and UserSurfaces. Surface-native state remains on the Surface. SystemProfiles
contains Interpretation, Supervisor, Transition, and TrapSurface profiles.
UserSurfaces contains the SysCallSurface profile; it is not nested in SystemProfiles.
Boot wires profile adapters; ordinary state operations do not change their authority.

Kernel holds the TrapSurface and SysCallSurface references. Their profiles define
state access separately from those Surface references. All profiles share the same
StateOwner request loop.

## Boot injection and compile-time contracts

Boot constructs each component with its assigned profile:

- Interpretation receives InterpretationState.
- Supervisor receives SupervisorState.
- Stage transitions use TransitionUnit, bound to permitted component identities.
- TrapSurface receives TrapSurfaceState from SystemProfiles.
- SysCallSurface receives SysCallSurfaceState from UserSurfaces.

These are named interfaces, not profiles selected from a generic list at runtime.
A component's constructor parameter and stored field use its specific interface.
Go checks that the supplied adapter implements that method set. Concrete adapter
conformance assertions can be added when the methods and adapters are declared.

Each profile has a distinct private adapter implementation exposing only its
permitted methods. Do not give every component the full broker merely narrowed
to an interface: type assertions could recover other interfaces it implements.
The profile should be bound to the component's identity during construction.
Components do not call an AssumeProfile operation to gain authority dynamically.

Compile-time conformance checks method compatibility. It does not prove that boot
assigned the correct runtime identity or that a requested transition is valid;
boot wiring and broker validation still own those checks.

## Custody remains a state fact

The [TU profile](../tu/proposal.go) provides the atomic stage transition protocol:

```text
Move identity -> Interpretation | Supervisor
```

State Owner commits the custody change and destination mailbox insertion in one
operation. There is no post-commit receiver call that must accept the Move to
complete that handoff. Until commit, the source retains responsibility; after
commit, the destination owns the Move, even while it is queued. Initial submission
records Interpretation ownership and mailbox placement together.

Interpretation transfers work to Supervisor, and Supervisor transfers emitted Moves
back to Interpretation through the same outer handoff mechanics: destination
mailbox placement, custody recording, then commit acknowledgment. Transferring an
existing Move and introducing a new emitted Move have different record semantics
inside the state operation, but do not require different processing loops. An
emitted Move keeps its own identity; it does not overwrite the original Move's
execution record. The exact profile signatures for both cases remain to be defined.

Taking a Move for processing does not change its component owner. Its canonical
record and unfinished-work tracking remain in State Owner. Exact take/wakeup and
capacity contracts remain open; waiting for work cannot block the State Owner loop.

TransitionUnit authorizes stage operations. It does not create another Move
owner or require a separate custody service. State Owner owns storage, not the
computation represented by every record. Surface identity, pending acceptance,
retry policy, and Trap semantics are outside CustodyRecord. Trap remains a Surface.

## Stateless processing services

Interpreter and Supervisor own behavior, while State Owner owns canonical state.
Interpreter processes one Move at a time. Supervisor owns dispatch and the
supervisor/worker behavior, with concurrent workers. Worker progress, execution
feedback, and retry bookkeeping live in State Owner and are accessed through the
SupervisorState profile. Temporary locals, active calls, and live worker resources
still exist within Supervisor; Surface-native state remains on the Surface.

State Owner also owns Move/continuation records and revision reconciliation.
Interpretation must be able to request graph changes and the corresponding
supervisor mailbox handoff in one commit. This makes the state operation atomic;
it does not make external Surface effects part of that transaction.

## Next declarations

Define the operations needed by each component, then declare its profile interface.
Do not use empty interfaces or arbitrary mutation callbacks as placeholders for
restricted authority. Specify related changes that must commit together and the
revision assumptions each operation validates.

[proposal.go](proposal.go) records the agreed struct shape. It intentionally leaves
request types, map and worker record contracts, and profile interfaces unresolved
rather than inventing their fields or method sets. Proposal files are excluded
from normal builds.

# Loop-owned authority and Move serialization

Date: 2026-09-20

Status: Discussion checkpoint. Records the loop-owner refinement; exact service
interfaces and serialization representation remain open. No runtime changes.

## Governing invariant

**Not everything entering the loop is an AuthorizedMove. Everything leaving the
loop for execution is an AuthorizedMove.**

The component that owns the loop holds the authority to prepare and authorize
the work it emits. An incoming Move does not need to grant that authority to the
loop, and a separate authority service is not required merely to authorize the
loop's own preparation work.

For example, the loop may authorize a read of the installed Horizon in order to
check an incoming Move. That does not yet authorize execution of the incoming
Move's target. The inspection can be permitted while the requested target
execution is later denied.

```text
Incoming Move — may be unauthorized
    -> loop services declare their required operations
    -> loop assembles ordered preparation/execution work
    -> all work emitted for execution is authorized
    -> executor runs the serialized work
```

This uses the existing AuthorizedMove concept: a Move paired with its resolved
Horizon. The relationship between the loop owner's authority and the incoming
Possibility's Horizon still needs concrete provisioning and binding contracts.
Loop privilege must not be mistaken for permission to grant an incoming target
unrestricted access.

## Services declare; the coordinator serializes

Each service encountered in the loop declares the system-level operations it
needs. This includes the services that inspect a Move and the target Surface's
declared preparation. The operations are serialized in the required order.

The dependency coordinator owns assembly and tracking of this work. Each service
owns the declaration and meaning of its operations; the coordinator need not
hard-code every service's adapter-state semantics.

Illustrative operations discussed:

```text
TrapController   -> write the trapped Move record
AuthorityChecker -> write the authority stamp
Target Surface   -> its declared adapter-state reads/writes
```

These examples name responsibilities, not an unconditional execution order.
A write that depends on a decision must follow that decision and occur only on
the appropriate branch. In particular, an authority success stamp and target
execution must not run after authority denial.

## Responsibilities

| Component | Responsibility |
|---|---|
| Loop owner | Hold the authority to admit and emit work. |
| Each service | Declare the required operations and their ordering. |
| Dependency coordinator | Assemble and track the ordered work and its results. |
| TrapController | Transfer control into preparation and handle trap bookkeeping. |
| Executor | Run the supplied authorized serialized work. |

The dependency coordinator and TrapController are closely related. They can be
modeled as responsibilities of one coordinating service while their exact
separation is tested. That is not yet a process-layout decision.

## Where preparation terminates

The earlier recursive model assumed that each prerequisite had to re-enter as a
fresh untrusted request and establish its preparation authority again. That is
not the intended model.

**Preparation terminates in the serialized executable operations emitted by the
loop.** Primitive state reads/writes become executable entries; they do not
require another copy of the same preparation cycle merely to run the primitive.
The loop owner authorizes this work using its provisioned authority.

Authorization and termination are distinct properties: the output invariant
establishes authorization, and lowering requirements to executable entries
supplies the preparation base case. StateOwner remains a Surface executing state
operations and emitting Moves through the Skyra protocol.

## Ordering, readiness, and execution context

The executor must not search for undeclared prerequisites. Serialized work
preserves declared order; an operation that consumes an earlier result runs only
after that result is available. Conditional execution must preserve the outcome
of the preceding decision.

An authorized operation and a ready operation are distinct. How serialization
represents dependent entries, conditional branches, failures, and release of
ready work remains to be specified. The invariant against dispatching work with
unresolved execution prerequisites remains in force.

The execution context retains the agreed minimal shape: serialized Moves and
the handlers required to execute them, ideally a single executable handler list.
Invocation frames and preparation bookkeeping do not need to travel with it.

## Open contracts

1. How the loop owner receives its authority and how emitted Moves carry the
   correct authority scope for runtime inspection versus target execution.
2. Service declarations and their dependency/ordering representation.
3. Conditional writes, such as an authority stamp, and failure short-circuiting.
4. Whether the executor receives a whole sequence or successive ready portions,
   and how results feed subsequent entries without undeclared dependency lookup.
5. Related state operations' transaction boundaries and protection against
   conflicting updates; ordered execution alone does not define atomicity.
6. TrapController/dependency-coordinator composition and correlation of results
   to the original invocation.

## Related

- [Authority namespace and execution context](authority-namespace-and-execution-context.md)
- [Earlier takeaways](takeaways.md)
- [AuthorityChecker interface](proposal/authority-owner/authority-checker/README.md)
- [StateOwner as a Surface](proposal/authority-owner/state-owner/README.md)

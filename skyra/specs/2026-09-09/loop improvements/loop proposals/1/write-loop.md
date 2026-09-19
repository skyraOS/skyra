# Proposal — scheduler write loop

Status: working proposal, not an implementation order. This records the current
discussion so it can be revised. An earlier assumption still needs clarification;
the unresolved points below must not be treated as settled protocol decisions.

## One writer, one lock

`Write` is the single entry point for state transitions. It owns the lock around
protocol interpretation, state updates, and selection/reservation of eligible work.
Dispatch callbacks return to `Write`. Helpers do not introduce additional entry
points or acquire their own locks.

The proposed protocol implementation is supplied to the scheduler through
dependency injection at boot. `Write` uses that instance to interpret each input.
The protocol interface and transition representation remain open.

## Two kinds of input

| Input | Meaning | Handler |
| --- | --- | --- |
| Dispatch outcome | Work was successfully or unsuccessfully dispatched; update its continuation state. | `applyDispatch` |
| Work | An `Invoke` or `Return` to apply. | `applyWork` |

Both kinds are state transitions accepted by `Write` and composed under the
protocol. Dispatch feedback reports the disposition of a handoff. Successful
dispatch does not establish successful execution or a returned observation;
execution may subsequently produce new work or a fault.

A `Return` belongs to the work path. It is not a dispatch acknowledgement.
The distinction is dispatch outcome versus work, not external input versus
execution feedback: execution can itself produce new work.

## Minimal core

The intended structure is:

```text
Write(input)
  lock
    interpret input through the injected protocol
      dispatch outcome → applyDispatch
      work             → applyWork
                           Invoke → applyInvoke
                           Return → applyReturn

    select and reserve eligible work affected by the transition
  unlock

  propagate processing error, notify host, or dispatch selected snapshots
    dispatch callbacks → Write
```

Illustrative pseudocode only; names such as `Input`, transition kinds, and helper
return shapes do not establish concrete Go types:

```text
Write(ctx, input):
    lock

    transition, error = protocol.Interpret(input)
    if no error:
        switch transition.kind:
            Dispatch:
                affected, notification, error = applyDispatch(transition)
            Work:
                affected, notification, error = applyWork(ctx, transition)

    if no error:
        selected = selectMoves(affected)

    unlock

    if error: return error
    if notification: return complete(notification)
    return run(selected)
```

## Helper responsibilities

| Helper | Responsibility |
| --- | --- |
| `applyDispatch` | Resolve the referenced continuation and apply its dispatch outcome. Preserve terminal facts against late acknowledgements. |
| `applyWork` | Branch on the work opcode and delegate to `applyInvoke` or `applyReturn`. |
| `applyInvoke` | Validate admission and linkage, then install continuations and parent register entries. Validate an entire batch before recording any of it. |
| `applyReturn` | Validate return binding, record the observation, and update the receiving parent or prepare a root notification. Handle duplicate and conflicting observations. |
| `selectMoves` | Assess affected continuations, reserve eligible execution, and produce detached dispatch snapshots. |
| `run` | Dispatch selected snapshots after unlocking. Route dispatch callbacks back through `Write`. |

`applyDispatch` and `applyWork` return affected work, an optional notification,
and an error. Their exact signatures remain open. They do not select, dispatch,
notify, or acquire locks. Selection stays visible in the core.

Helper names describe their work, without a `Locked` suffix. The lock is owned by
`Write`; the helpers execute within that scope. Existing root lookup and fault
bookkeeping can remain supporting helpers under the same naming convention.

The protocol has no independent authority to mutate the continuation table.
`Write` applies its interpretation through the private helpers. No additional
persistent scheduler state or public entry points are proposed here.

## Current implementation and unresolved assumptions

The current implementation is
[`scheduler.go`](../../../../../src/skyra/skyra-dna-v14/scheduler/scheduler.go).
Its main event method is `Write`.

- The current code reserves `Dispatched` before handoff and records `InFlight`
  after successful acknowledgement. The discussion described successful handoff
  as `Dispatched`. The protocol must settle the meaning and mapping of these states.
- The current runtime applies `Return` directly and dispatches only `Invoke`.
  The discussion describes dispatch feedback for work, potentially Invoke or
  Return. Whether dispatch semantics change is unresolved; the decomposition
  alone must not silently decide it.
- The injected protocol's responsibilities, inputs, and output representation
  need definition. The pseudocode is a structural proposal, not a final interface.
- Placement of execution faults and admission failures within the proposed
  protocol needs an explicit mapping. They must not be conflated with an
  unsuccessful dispatch merely to fit the two branches.
- The pseudocode retains the current choice between a notification and selected
  dispatches. Whether future protocol transitions may require both remains open.
- The user has identified an earlier assumption to revisit. Its correction has
  not yet been specified; this document preserves the proposal for that discussion.

## Implementation scope, once settled

Keep a single state-transition path. Preserve atomic batch admission, return
binding, duplicate handling, parent resumption reservation, saved execution
contexts, detached snapshots, callback error propagation, and calls outside the
scheduler lock unless a subsequent protocol decision explicitly changes them.

The current contract interpreter remains a stub; this decomposition does not
define compiled matching, retries, or repeated execution waves.

Verify any later implementation with the existing immediate, reversed,
concurrent, duplicate-return, fault-suppression, and callback-error tests, including
the race detector. This document makes no runtime changes.

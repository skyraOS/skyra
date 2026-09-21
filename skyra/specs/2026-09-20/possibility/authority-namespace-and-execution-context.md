# Possibility as authority namespace; execution as a handler list

Date: 2026-09-20

Status: Discussion refinement. Supersedes the earlier description of execution
context as a Move plus concrete input/resource fields. No runtime changes.

## Possibility and invocation frames

The Possibility can be the namespace that holds the run's authority through its
installed Horizon. It does not need to be the object handed to the executor.

Invocation frames track callers, return destinations, preparation progress, and
dependencies. A separate coordinating service can maintain this stack-like
structure on behalf of the Possibility. The executor does not need the stack.
Concurrent invocation branches still fit the earlier cactus analogy.

This separates logical ownership from placement: the frames belong to the run,
but they need not live inside the object used for execution or travel with it.

## Execution context — precise refinement

**All the execution context is, is a list of serialized Moves and the handlers
required to execute them. Ideally it is just one list of handlers that the
executor executes.**

```text
ExecutionContext
    serialized Moves + handlers required to execute them

Preferred shape
    executable handler list
```

There is no additional top-level collection of concrete inputs/resources in this
definition. There is no invocation stack, Horizon copy, or dependency scheduler
inside the execution context. How handlers embody or reference their associated
Moves remains to be specified; do not invent a second payload structure here.

The list representation, handler signature, and meaning of serialization remain
open. Encoding Moves for transport and establishing execution order are distinct
questions. A local callable handler and a handler reference carried across a
process boundary may require different representations without changing this
logical model.

## Reads and writes use the same execution model

System-level reads and writes are Moves executed under the same Possibility's
authority as the Surface invocation they prepare. They do not require a second
authority namespace or a separate public state-request protocol. StateOwner is
the Surface addressed for those operations.

A Surface declares which system-level reads/writes must occur before invocation.
The coordinator arranges that preparation; no unresolved prerequisite reaches
the executor. The same execution-context form is used for state-operation work
and subsequent prepared Surface work.

Related reads and writes still need an explicit consistency/atomicity rule.
Membership in one Possibility alone does not prevent another execution from
interleaving state updates. Their grouping does not require sending the whole
invocation stack to the executor.

## Trap and fault

- **Trap:** transfer control to preparation.
- **Fault:** an error.

Preparation is not itself a failure. Required state that cannot be obtained must
be handled through the trap/fault protocol; the exact transition and recovery
contract remain open. The earlier rule remains: an invocation with unavailable
required state is trapped rather than dispatched unprepared.

```text
Move enters coordinator
    -> resolve through the Possibility's Horizon
    -> preparation required: trap
        -> emit declared state-operation Moves
        -> execute their handlers through the ordinary execution model
        -> results return through authority
        -> coordinator updates preparation progress
    -> dispatch the prepared invocation's execution context
```

The coordinator retains continuation bookkeeping throughout. Its exact service
composition is open: authority and dependency preparation may be combined in one
privileged service. Runtime-generated Moves must remain within the provisioned
authority of the work they serve; coordination privilege is not an unrestricted
grant to use another Surface's state.

StateOwner's primitive implementation needs a base case: performing an actual
state read/write must not recursively require another preparation Move for that
same primitive forever.

## Minimize authority-to-execution transformations

The direction is to provision executable bindings together with their authority
in Horizon. Preparation selects and assembles the admitted work into the
execution context. Execution should not receive a bare permission decision and
then independently rediscover a handler through unrestricted registry access.

The desired relationship is:

| Part | Responsibility |
|---|---|
| Possibility / Horizon | Authority namespace and provisioned executable bindings |
| Coordinator / frame service | Admission, traps, preparation, dependencies, and return bookkeeping |
| Execution context | Serialized Moves and their required handlers; ideally one handler list |
| Executor | Execute the supplied handlers and report outcomes |

Handler binding, ordering, result correlation, and retirement still need exact
contracts. Those contracts should preserve the small execution context rather
than move the coordinator's state into it.

## Related

- [Possibility investigation](README.md)
- [Earlier takeaways](takeaways.md)
- [Three-role proposal under examination](proposal/README.md)

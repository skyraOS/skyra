# Disposable workers, state turns, and execution policy

Date: 2026-09-20

Status: Design idea captured from discussion; no implementation changes. The
persistent-parent/disposable-worker direction is the premise. Queue, scheduler,
receive-pump, and policy details below are proposals, with open decisions called
out explicitly.

## Persistent parent, disposable request/response workers

Keep one persistent parent process that owns canonical state and spawns ephemeral
workers to perform invocation work. A worker is a request/response shell. Its
process does not need to survive just to preserve a continuation: the parent
stores the state needed for subsequent work and can spawn a fresh worker.

The parent owns Possibilities, Horizons, invocation records, dependencies, return
destinations, committed continuation state, and worker lifecycle records. Workers
receive the input and restricted authority needed for their execution and report
outcomes to the parent.

For example, a worker can report that work requires a child result and exit. The
parent records that dependency, executes the child, and launches fresh work with
the committed continuation and child result. The original process need not wait.
The exact yield/continuation response contract remains to be defined.

Distinguish these lifetimes:

- **Logical invocation:** recorded by the parent until completion or fault.
- **Worker execution:** one disposable request/response shell. A continuation or
  retry may require another worker for the same logical invocation.
- **Provider resource:** a connection, session, or external process, whose
  lifetime depends on the Surface's semantics.

An invocation identity survives worker replacement. Each worker execution needs
its own temporary handle, so retiring an old worker rejects its late emissions
without necessarily settling the logical invocation. This differs from DNA-17's
single fd retained across the invocation session and its resumes.

This design requires continuation state to be externalized before the worker
exits. DNA-17's opaque DriverState alone does not establish that guarantee: some
provider continuity currently resides in private driver session bookkeeping.
If a Surface requires a persistent provider connection, ownership of that
connection must be outside the disposable execution shell, or the Surface must
define how it reconstructs the connection/session.

## Separate state work from Move execution

The proposed separation is two scheduling domains:

1. State turns: interpret incoming Moves and execution events, then commit their
   prepared state changes.
2. Execution: admit and run the external work authorized by those turns.

Separate queues do not require two concurrent worker pools. A first implementation
can keep one serialized state-turn processor and a concurrent execution pool.
Whether state interpretation later gains parallelism is a separate decision.

```text
Incoming Move / worker event
            |
            v
      State-turn queue
            |
      step():
        lock
        interpret -> Request
        commit(Request)
        unlock
            |
            v
      Execution scheduler
      admission / concurrency / delays
            |
            v
      Disposable execution workers
            |
            v
      Receive pump
      reads and frames worker output
            |
            +---- event -> State-turn queue
```

The diagram shows the success path through scheduling; interpretation can also
record a deferred execution, completion, or fault without dispatching a worker.

## Request and step discipline

Interpret reads canonical state, resolves authority and relationships, evaluates
applicable policy, and prepares a **Request**. Commit applies its specified
writes. The concrete Request fields remain open, but all information required
for those writes must be prepared during interpretation.

Hold the required locks across interpret and commit, following
[improvements.md](improvements.md). Execute external work after releasing them.

Queue the **inputs to interpretation**, rather than interpreting a Request and
leaving it in an independently drained commit queue. Separating read and write
with a queue would allow the prepared Request's assumptions to become stale and
would require a new validation/concurrency mechanism.

Thus a "queue for requests to state" means a queue of requested state turns at
the boundary. The prepared Request is the interpret-to-commit artifact within a
turn. These two roles should not be confused even if final type names differ.

## Receive pump

The receive pump is the loop that reads worker output and submits execution
events to the runtime. Its proposed responsibilities are:

- Read bytes and assemble complete frames according to the transport contract.
- Bind each event to the worker execution assigned by the runtime; identity is
  not trusted merely because the worker claims it in a payload.
- Submit frames, EOF, transport failures, and process-exit observations through
  the state-turn boundary. Process waiting may be a separate lifecycle observer
  feeding that same boundary.

Surface-specific interpretation gives those events meaning in a new state turn.
The pump does not independently settle invocations, choose retries, or mutate
canonical state. EOF or process exit alone does not imply successful completion.

For a request/response worker, reading can end when its response and lifecycle
have been handled. A persistent provider stream may need its own longer-lived
pump. "Pump" names the reading responsibility, not a requirement that every
worker connection persist across invocations.

Ordering between final frames and exit observations, framing limits,
backpressure, and pump/process cleanup ownership still need concrete contracts.
Queue design must also address the existing interpreter self-enqueue deadlock;
separating execution from state does not automatically fix that issue.

## Surface policy and runtime enforcement

Surfaces have different retry, rate-limit, timeout, and concurrency semantics.
Introduce a small policy abstraction that lets a Surface describe those
semantics while the runtime owns scheduling and enforcement.

| Surface-specific semantics | Runtime responsibility |
|---|---|
| Which outcomes permit retry | Record attempts and schedule eligible retries |
| Whether an operation can safely be repeated | Admit or reject another attempt under that rule |
| Meaning of provider retry-delay signals | Record the earliest permitted execution time |
| Which operations share a rate-limit scope | Account for and reserve the shared budget |
| Concurrency and timeout defaults | Enforce admission, deadlines, and cancellation |

Rate-limit scope need not equal SurfaceID. Multiple Surfaces may use one provider
account or resource and consume the same budget. Independent per-Surface limits
would not enforce that shared constraint.

Policy evaluation belongs in interpret. A candidate decision vocabulary is:

```text
Run
Defer(until)
Retry(until)
Complete(result)
Fail(reason)
```

These are conceptual decisions, not finalized Go types or protocol opcodes.
Interpret prepares the decision and any budget reservation; commit records them.
The scheduler executes admitted work and submits wake-up events for deferred work.
Eligibility checks that depend on canonical state happen in a state turn, rather
than through independent scheduler mutations. Waiting for a rate limit or retry
delay must not occupy a disposable worker process.

Configuration, provider outcome classification, and mutable policy accounting are
distinct: Surface code supplies meaning and rules; the parent owns canonical
counters, deadlines, and lifecycle state. How runtime-wide constraints combine
with Surface defaults remains open.

## Attempts, continuations, and ambiguous outcomes

A retry is another execution attempt for the same logical invocation. It receives
a fresh execution handle. Responses must correlate to the appropriate execution
so an old worker cannot complete replacement work with a late response.

A continuation after a child result is not necessarily a retry. Both may launch
fresh workers, but continuation advances recorded work while retry repeats an
operation. The identity model must preserve that distinction; concrete attempt
and continuation identifiers remain open.

Retry policy must distinguish **the operation failed** from **its answer was not
received**. A worker can perform an external effect and die before reporting it.
Centralized state does not make that effect atomic with the parent commit.
Surfaces need to classify ambiguous outcomes and, where supported, supply
idempotency or reconciliation semantics instead of treating every lost response
as permission to repeat the operation.

## Initial scope and open decisions

Suggested starting shape: one serialized state-turn queue, one execution
scheduler, disposable workers, and receive/lifecycle events returning through
state turns. Start the policy interface around admission and outcome decisions.

Still open:

1. Concrete Request and worker response/yield contracts, including how complete
   continuation state is committed before a worker exits.
2. Logical invocation, continuation, attempt, and execution-handle relationships.
3. Ownership of persistent provider sessions and resources, including whether
   any are shared across Possibilities.
4. Receive-pump framing, event ordering, backpressure, and shutdown behavior.
5. Policy signatures, limit scopes, budget reservation/release rules, and how
   Surface defaults combine with runtime constraints.
6. Cancellation and deadlines across dependencies, delayed work, and live workers.
7. Whether a separate concurrent state pool is needed beyond the initial single
   serialized state-turn processor.

## Related

- [Explicit interpret/commit phases](improvements.md)
- [Surface ingress and egress profiles](surface.md)
- [Invocation cactus and Possibility isolation](invocation-cactus-and-possibility-isolation.md)
- [Identity relationships](surfaceID-moveID-fd.md)
- [Stdio Surface and persistent streams](stdio/stdio-surface.md)
- [Execution handoff primitives](claude-syscall/the-handoff-set.md)
- [Open correctness issues](issues.md)

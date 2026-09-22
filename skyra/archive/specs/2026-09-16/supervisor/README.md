# Supervisor

Date: 2026-09-17.

Status: conceptual proposal. Construction and ownership follow the agreed loop;
profile decomposition, execution records, and worker helper contracts remain open.

## Construction and ownership

Boot constructs Supervisor with restricted SupervisorState and CustodyState
adapters and a configured WorkerPool. The constructor fixes Supervisor's identity
and Interpretation as its permitted destination. Boot binds and StateOwner
validates adapter authority. Construction stores dependencies; boot starts Run
after wiring components and supervises its return.

WorkerPool owns live goroutines, capacity reservations, cancellation handles,
and independent monitoring. Canonical worker, progress, and outcome records belong
to StateOwner. Boot configures capacity; execution policy comes from the authorized
binding.

## Profile decomposition and naming

Profile methods should express the bare atomic state operations decomposed from
the loop. Define them around which related changes need to commit together.
The final loop and profile should use matching names for the same operation.
If execution preparation is named Reserve, use Reserve consistently at both layers.
Avoid keeping separate names and wrappers that merely describe the same operation.

The current sketch retains WorkerPool.Reserve before Take and
SupervisorState.PrepareExecution during startWorker. Their similar naming is
provisional: capacity reservation currently waits locally, while preparation
atomically authorizes and registers a specific Move. Reconcile their names and
decomposition when defining the profile operations, preserving capacity before
Take and registration before launch. Waiting for capacity or external execution
must not block StateOwner's request loop. These notes also appear in proposal.go.

## Processing

The outer loop reserves capacity, takes a Move, and hands the reservation and
Move to startWorker. startWorker owns the reservation from entry. If no worker
starts it releases capacity; otherwise the pool retains capacity until the actual
worker lifecycle ends. A timeout report alone does not release a slot.

Preparation loads the canonical Move, derives its Possibility and target, checks
the installed Horizon, resolves and validates the Surface/contract binding, and
finds or registers the correlated execution in one StateOwner operation. Finding
an existing execution must not launch another worker. PreparedExecution's exact
identity and opaque binding fields remain to be defined.

```text
prepare authorized execution through StateOwner
    -> launch registered worker
    -> call bound ingress
        -> record acceptance separately
        -> receive emissions
            -> egress normalizes each emission to one Move
            -> retain correlated outcome through StateOwner
            -> transfer emitted Move to Interpretation
    -> record worker exit when the worker actually ends
```

Workers support multiple emissions through the Driver callback. An emitted child
Invoke need not complete its source invocation. Each emission needs stable
identity and its own reconciled handoff. Exact callback and worker lifetime
contracts remain open.

Supervisor binds provenance to the actual source execution, retains each outcome
through StateOwner before its handoff, and transfers through CustodyState. A stale
revision rebuilds only the state proposal from the retained outcome; it never
reruns ingress or assigns a new emission identity. Ambiguous acknowledgment is
reconciled before resubmission. Custody transfer commits related changes, mailbox
placement, and ownership together.

## Monitoring and shutdown

Monitoring runs independently of ingress. It records lost progress or deadlines
through StateOwner and arranges a correlated failure-report Move to Trap through
Interpretation. Trap decides corrective work. Late output reconciles against
canonical records to prevent duplicate settlement. Timeout does not establish
worker exit or absence of external effects.

Shutdown must stop taking work, request cancellation according to policy, and
preserve unfinished execution and retained outcome records. Run and WorkerPool
shutdown coordination remains to be specified. Ingress return, worker exit,
emission handoff, and invocation completion are distinct events.

See [proposal.go](proposal.go), the [minimal loop](../ctu/minimal-loop.md), and
the [Possibility authority proposal](../possibility/README.md). The Go proposal
is excluded from normal builds.

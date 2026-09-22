# Driver composition and opaque state

**Status: DESIGN DIRECTION.** This note records the agreed driver composition and
state-ownership boundary. It extends the Codex-specific identity model in
[`codex-ingress-boundary.md`](codex-ingress-boundary.md) into a provider-neutral
driver rule. The stable target and binding relationships are recorded separately in
[`entity-address-driver-transport.md`](entity-address-driver-transport.md).

## Composition

A driver is composed from transport mechanics, protocol behavior, and translation
at the Skyra boundary:

```text
Driver
    = Transport
    + {
        Ingress: Skyra -> protocol messages
        Egress:  protocol messages -> Skyra
        State:   opaque correlation shared by both directions
      }
```

End to end:

```text
Skyra Invoke
    |
    v
Ingress translation
    |
    v
protocol request
    |
    v
Transport <-> external entity
    |
    v
protocol response / event
    |
    v
Egress translation
    |
    v
Skyra Invoke / Return / fault
```

### Transport

Transport moves bytes or frames. It does not understand Skyra or application
semantics.

Examples:

```text
managed process stdin/stdout
Unix socket
WebSocket
Streamable HTTP
one-shot process pipes
```

For a managed process, transport mechanics include process startup and shutdown,
stdin serialization, stdout framing, stderr collection, backpressure, cancellation,
and crash detection. A subprocess should normally be executed directly rather than
through a shell.

### Ingress

Ingress translates a Skyra invocation and its current state into operations the
external entity understands.

For Codex this includes:

```text
InvocationID -> create or resume ThreadID
Focus        -> dynamicTools
Invoke input -> turn/start input
child Return -> answer the pending Codex tool callback
```

### Egress

Egress translates external responses and events into Skyra moves.

For Codex this includes:

```text
tool callback   -> child Invoke
agent completion -> Return
provider failure -> fault
```

Ingress and egress share protocol correlation, but that correlation remains private
to the driver. The runtime never learns what a `ThreadID`, `TurnID`, `ToolCallID`,
JSON-RPC request id, or MCP request id means.

## State ownership

The scheduler is the authoritative in-memory owner of invocation state. It stores
driver-specific protocol state with the continuation, but treats that state as an
opaque value. The driver owns the schema and meaning of the value.

```text
Scheduler:
    "Invocation A has opaque driver state X."

Codex driver:
    "X says Thread 123 is waiting for tool call 456."

MCP driver:
    "X says request 789 is awaiting a remote response."
```

The scheduler stores state, not protocol semantics.

## Do not use an unrestricted `any` as the durable contract

An `any` field is useful for an early in-memory prototype, but is the wrong durable
boundary. It permits mutable pointers, arbitrary Go values, values the persistence
layer cannot encode, and driver mutations that escape the scheduler's commit.

Use a versioned opaque envelope instead:

```go
// DriverState is owned as data by the scheduler and interpreted only by the driver.
type DriverState struct {
	Kind    string
	Version uint32
	Payload []byte
}
```

`Kind` prevents the wrong driver from interpreting the payload. `Version` gives the
driver a migration seam. `Payload` makes the persistence requirement explicit.

The scheduler must treat the envelope as immutable. A driver receives the previous
value and returns a replacement; it must not retain a mutable pointer into scheduler
state.

## Driver as a transition

Invocation behavior should be reconstructible from the stored state and the next
event:

```go
type Driver interface {
	Step(
		invocation Invocation,
		state DriverState,
		event DriverEvent,
	) (DriverTransition, error)
}

type DriverTransition struct {
	State DriverState
	Moves []Move
}
```

These types are conceptual, not an approved contract. The important property is:

```text
(invocation, previous opaque state, event)
                    |
                    v
          (replacement state, moves)
```

The scheduler commits the replacement state together with the continuation and
register changes caused by the same transition:

```text
continuation state
+ register / handoff changes
+ opaque driver state replacement
= one scheduler commit
```

This is what makes a driver stateless in the useful sense: its instance does not
contain authoritative invocation state. Recreating the driver and supplying the
stored state is sufficient to continue interpreting that invocation.

## What remains live outside the continuation

Not all driver machinery is invocation state. Process handles and open connections
are live resources; they cannot be serialized into a continuation.

The split is:

| Possibility-scoped live host | Invocation-scoped opaque state |
| --- | --- |
| OS process handle | provider execution handle |
| stdin/stdout pipes | thread or request id |
| connection initialization | current protocol phase |
| frame reader and writer | active turn id |
| process health | pending tool-call id |
| backpressure queues | child-return correlation |
| connection-level request routing | resumable provider metadata |

Live host state is ephemeral and reconstructible. Invocation state is authoritative
and participates in the durable transition boundary.

## Process scope follows possibility scope

A possibility is the kernel-minted authority context. A process-backed driver host
belongs to that possibility:

```text
Possibility A
+-- installed Horizon A
+-- process-backed driver host A
    +-- Invocation A1 <-> opaque driver state A1
    +-- Invocation A2 <-> opaque driver state A2

Possibility B
+-- installed Horizon B
+-- process-backed driver host B
    +-- Invocation B1 <-> opaque driver state B1
```

For Codex the intended shape is one `codex app-server` process per possibility and
one Codex thread per Codex-backed invocation. Multiple invocation threads may use
the same possibility-scoped process without sharing invocation identity.

Process separation is not authority by itself. The kernel must derive the process
launch policy from the possibility's installed authority, including filesystem,
network, environment, credentials, executable, and resource policy. Provider-level
thread sandboxing may narrow that authority further but is not its source.

## Restore sequence

If live invocations survive process death:

```text
1. Restore the possibility and its installed authority.
2. Recreate the possibility-scoped driver host.
3. Restart and initialize the external process or connection.
4. Restore each continuation and its opaque DriverState.
5. Let the owning driver decode and reconcile that state.
6. Resume provider work and normal Skyra transitions.
```

The kernel persistence mechanism need not understand a provider payload. The driver
must be able to decode its own stored versions and decide whether a provider-side
execution can be resumed, retried, reconciled, or faulted.

## Example: Codex

The scheduler may durably hold a payload which the Codex driver interprets as:

```go
// Private to the Codex driver. Serialized into DriverState.Payload.
type codexInvocationState struct {
	ThreadID           string
	ActiveTurnID       string
	Phase              string
	PendingToolCallID  string
	ChildInvocationID  string
}
```

The scheduler sees bytes. The Codex driver sees an invocation waiting for a child.
When the child returns, the scheduler supplies the stored state and return event to
the driver. The driver answers the appropriate Codex callback and returns replacement
state for the scheduler to commit.

## Invariants

```text
scheduler owns invocation state as data
driver alone interprets its protocol payload
driver returns replacement state; it does not mutate scheduler state by alias
opaque state is serializable and versioned
provider identifiers do not enter kernel or scheduler behavior
live transports are possibility-scoped and reconstructible
process authority is derived from the kernel-minted possibility
driver reconstruction does not depend on hidden authoritative instance state
```

## Open

- Encoding choice for `Payload` and limits on its size.
- Driver state version registration and migration failure behavior.
- The exact event and transition contracts used around asynchronous transport input.
- How connection-level request routing is rebuilt after a process restart.
- Atomicity between committing driver state and performing external I/O. A durable
  outbox, provider idempotency key, or reconciliation step may be required to avoid
  losing or duplicating an operation across a crash.
- Whether terminal driver state remains with the completed continuation for audit or
  is compacted into a smaller terminal record.

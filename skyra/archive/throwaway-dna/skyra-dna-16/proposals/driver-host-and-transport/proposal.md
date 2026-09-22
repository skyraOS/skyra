# Proposal — driver host, transport, and opaque state

**Status: IMPLEMENTED FOR THE LOCAL DNA16 PROOF.** The transition, transport,
opaque state, process-host, and Codex boundaries below are implemented. Live
account provisioning and OS sandbox policy installation remain outside this
local proof.

## Pressure

DNA15's `Ingress(move, output)` and `Egress(raw)` are sufficient for the local
synchronous probe. The Codex App Server probe adds three requirements:

- a long-lived process can emit protocol events after `Ingress` returns;
- protocol identifiers must survive separately from live process resources; and
- process loss must not make the scheduler forget authoritative invocation state.

Codex is evidence for these boundaries, not vocabulary for the runtime.

## Ownership

```text
Kernel
    constructs a possibility-scoped driver host
    derives its process, credential, filesystem, network, and resource authority

Driver host
    owns live process/connection resources and a Transport
    routes opaque protocol frames to the owning driver

Transport
    moves one framed message in each direction
    does not interpret the protocol or Skyra Moves

Driver
    interprets capability, operands, private protocol frames, and DriverState
    requests Skyra effects through DriverEmission

Scheduler
    stores and atomically replaces opaque DriverState with continuation changes
    never interprets provider sessions, threads, turns, or request identifiers
```

Live resources are deliberately not durable state:

```text
live:     process, pipe/socket, reader loop, writer lock, request router
durable:  versioned opaque payload needed to reconcile an invocation
```

## Transport contract

DNA16 adds the following optional host-side seam:

```go
type Transport interface {
    Send(context.Context, []byte) error
    Receive(context.Context) ([]byte, error)
    Close() error
}
```

Each byte slice is exactly one complete protocol frame. Framing belongs to the
implementation—for example JSONL over stdio or messages over a socket—but the
frame's meaning belongs to the driver. `Close` must unblock a blocked `Receive`.

`Transport` is not embedded in `Driver`, `DriverRegistration`, the scheduler, or
the continuation language. A driver may be in-process and have no transport. A
driver host may also use a protocol-specific transport implementation without
changing the public entity address.

Transport errors are observations about delivery, not continuation-language
instructions. The driver translates them into `DriverTransition.Failure`; the
transport does not emit Skyra `Move`s itself.

## Opaque driver state contract

DNA16 adds scheduler-owned storage with driver-owned interpretation:

```go
type DriverState struct {
    Version uint32
    Payload []byte
}
```

`Continuation.DriverState` is separate from `Invoke.Register`:

- `Register` is semantic continuation state visible to the continuation laws;
- `DriverState` is private adapter state required to continue or reconcile the
  external protocol; and
- live host resources never enter either value.

The scheduler must copy the payload on admission, replacement, and snapshot. A
driver receives a detached value and returns a complete replacement. It never
receives a mutable pointer into resident scheduler state. The driver owns payload
version decoding and migration; an unknown or invalid version faults the affected
invocation rather than silently discarding state.

For Codex, the payload may contain thread ID, turn ID, phase, pending tool-call
correlation, tool-name mapping, candidate output, and contamination state. Those
fields are examples owned by the Codex driver and must not become DNA16 fields.

## Selected driver transition

DNA16 changes driver ingress to receive the last committed state and emit a full
replacement:

```text
(detached Move, previous DriverState)
    -> DriverEmission(replacement DriverState, raw output, HostCommands, failure)
    -> DriverTransition(replacement DriverState, validated Moves, HostCommands, failure)
```

The controller validates and translates raw output. The scheduler atomically
commits replacement state with all accepted Moves, or rolls the entire transition
back. Only after that write returns successfully does the controller call the
optional `CommandExecutor`. Placement is committed before `Ingress`, so an
emission is never merely buffered while its command escapes.

Command IDs are stable protocol-owned values. A crash after commit and before send
is reconciled from the committed opaque payload. The first Codex recovery policy
is deliberately conservative: when a new host cannot prove an outstanding live
request safe to resume, it faults the invocation instead of replaying or resetting
it.

`DriverTransition.Failure` atomically commits final driver state—such as Codex
contamination—and faults the workflow. Provider errors remain runtime signals,
not continuation-language opcodes.

## Invariants

1. Entity addresses do not encode transport choice or provider session IDs.
2. Transport interprets no protocol and creates no Skyra effects.
3. No driver is required to have a process, connection, or Transport.
4. The kernel owns construction and ambient authority of external driver hosts.
5. The scheduler is the sole owner of resident `DriverState` storage.
6. Only the owning driver interprets its state payload and version.
7. Provider vocabulary does not enter scheduler or continuation-language laws.
8. Host loss can destroy live machinery, but not committed driver state.

## Implemented proof

The local implementation:

1. move the Codex JSONL transport behind DNA16's `Transport` contract;
2. prove framing, serialized writes, cancellation, and close-unblocks-receive;
3. replace the Codex prototype's private session payload with detached
   `DriverState` round trips;
4. classify native Codex items and persist contamination before accepting a final
   return; and
5. kill and recreate the host, then either reconcile safely or fault explicitly
   from stored state.

The generic process host is intentionally narrow. Promotion of request routing or
additional host policy still requires a second materially different driver or an
explicit DNA16 decision.

## Related decisions

- [`specs/2026-09-10/driver-composition-and-state.md`](../../../../../specs/2026-09-10/driver-composition-and-state.md)
- [`specs/2026-09-11/codex-driver/04-shared-driver-seams.md`](../../../../../specs/2026-09-11/codex-driver/04-shared-driver-seams.md)
- [`specs/2026-09-11/codex-driver/05-prototype-boundary.md`](../../../../../specs/2026-09-11/codex-driver/05-prototype-boundary.md)
- [`specs/2026-09-11/codex-driver/06-native-tools-and-containment.md`](../../../../../specs/2026-09-11/codex-driver/06-native-tools-and-containment.md)

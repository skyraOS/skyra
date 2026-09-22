# Possible shared driver seams

## Rule

Codex is evidence for a driver contract, not proof of a universal abstraction.
Extract a shared mechanism only after another driver needs the same semantic
boundary.

## Likely shared

These concepts already follow from Skyra ownership rather than Codex itself:

```text
DriverBinding       implements one stable entity/address
CapabilityDescriptor advertises functions but grants no authority
DriverState         versioned opaque invocation state stored by scheduler
DriverEvent         something the driver may interpret
DriverTransition    replacement state plus requested runtime effects
DriverSignal        completion, failure, availability, or pressure indication
```

A candidate driver core remains:

```go
type Driver interface {
	Step(
		invocation InvocationView,
		state DriverState,
		event DriverEvent,
	) (DriverTransition, error)
}
```

The useful invariant is more important than this signature:

```text
(invocation, previous opaque state, event)
              → (replacement opaque state, effects)
```

The scheduler can commit continuation changes and replacement driver state as one
transition. The driver never retains a mutable pointer into scheduler state.

## Possibly shared, but not yet guaranteed

- lifecycle management for process- or connection-backed hosts;
- request-id allocation and response correlation;
- framed reader/writer queues and backpressure;
- restart and reconciliation hooks;
- telemetry and normalized driver signals; and
- an outbox for external writes that must agree with durable state commits.

These belong in a driver kit only after their ownership is demonstrated by more
than one protocol.

## Must not be assumed universal

A generic driver contract must not require:

- an OS process;
- JSON or request/response messaging;
- a persistent connection;
- threads, turns, sessions, or tool calls;
- exactly one output;
- asynchronous events; or
- recoverable remote state.

An in-process deterministic driver and a streaming remote driver must both remain
possible.

## Host versus driver core

Codex suggests—but does not yet prove—a reusable split:

```text
Driver core
    interprets capability, operands, opaque state, and protocol events

Driver host
    owns live resources and delivers protocol commands/events
```

The host may be absent for an in-process driver. Protocol commands and events
should therefore remain driver-owned opaque envelopes rather than becoming Codex-
shaped runtime types.

## Pressure on DNA15

DNA15 currently exposes:

```go
Ingress(move, output)
Egress(raw)
```

That works for the synchronous probe but does not naturally express unsolicited
events, durable correlation, replacement opaque state, or atomic transition
effects. The Codex implementation should prototype the transition contract beside
the existing interface before Skyra promotes it to a shared driver API.

## Extraction test

Promote something into the shared driver kit only when all are true:

1. at least two materially different drivers require it;
2. its semantics can be named without provider vocabulary;
3. the scheduler can store or route it without interpreting protocol meaning; and
4. omitting it would duplicate correctness-sensitive machinery.


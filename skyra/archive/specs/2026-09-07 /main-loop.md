# The main loop

**Date:** 2026-09-08  
**Status:** DESIGN — decided. Scheduler owns registers and continuation state; Continuation Controller
(CC) is stateless across invocations. This supersedes the earlier CC-private-store decision, CTU-as-executor,
kernel-contains-scheduler, and Issued/Landed lifecycle framing. Not yet reflected in `engine.go`.

## Ownership

```text
CTU
  ├─ Scheduler
  │    └─ private continuation store / registers
  ├─ Continuation Controller (CC)
  └─ addressed drivers / registry
       └─ kernel is the root driver
```

Scheduler owns resident continuations, registers, child observations, outstanding-work correlation,
and lifecycle state. It is the single writer and determines which work may proceed.
An `Fd{Move, Register}` is the current starting shape for a resident continuation.
This live store is distinct from durable engrams and their artifact/history records.

CC receives the context needed for authorization and dispatch and reports what happened. It retains
no per-invocation continuation state. Temporary call context does not constitute another state store.
Drivers perform the actual work; CTU coordinates it.

## The loop

```text
Driver emits Move
       ↓
CC.authorize(move)
       ↓
Scheduler receives authorized Move
  ├─ writes registers / continuation state
  └─ selects eligible work
       ↓
CC.dispatch(work)
       ↓
Dispatch result / callback: what happened
       ↓
Scheduler writes continuation state

Driver result arrives as a new Move → CC.authorize(move) → Scheduler → repeat
```

There is one scheduler-owned write path with two kinds of input:

- An authorized Move carries an invocation and its payload. Opcode semantics determine its effect;
  a Ret can contribute an observation to a parent register.
- A dispatch result or callback describes what happened while handing work to the driver. Scheduler
  records the corresponding state change. It is not the work's observation payload.

Successful dispatch does not necessarily mean completed work. An asynchronous driver may accept work
now and emit result Moves later. Those Moves enter through authorization again. CC reports dispatch
outcomes; Scheduler records them and decides what proceeds. The exact result type, callback signature,
and outcome vocabulary remain open; this does not restore Issued/Landed as driver messages.

The diagram specifies responsibilities, not a synchronous call stack. Driver output can arrive out of
order or contain several Moves. State needed to correlate these arrivals belongs to Scheduler.

## Authorization and dispatch

`authorize()` and `dispatch()` remain provisional method names. Authorization checks established Horizon
grants; it does not mint authority. An authorized incoming Move is handed to Scheduler for state changes.

The binding authority check must hold at access. A capability can expire while a Move waits, so an
admission-time check alone is insufficient. Dispatch must enforce valid authority when touching the
target driver, using supplied context. Exact method signatures and check placement remain open.

Driver ingress/egress describe boundary crossings. Register reads/writes describe Scheduler's state
operations. They are related but are not interchangeable terms.

## Every Move is an invocation

Every Move is a new invocation because it invokes something. Call invokes child behavior; Ret invokes
parent behavior. Its connection to an earlier invocation is correlation, not reuse of that invocation's
identity. New invocation does not mean new authority.

All driver output toward the runtime is Move(s), including operation errors represented as Ret
observations. Internal dispatch outcomes are a separate report to Scheduler about the handoff; they do
not replace result Moves or carry a second driver-result protocol.

A Move's ParentFd supplies parent linkage in the current shape. Its own InvocationID and correlation
with prior work must remain distinct. The concrete live key and parent linkage still need reconciliation
for simultaneous invocations sharing the same MoveID.

Call carries its invocation inputs and may proceed once resolved and permitted. The return path
accumulates child observations and holds or releases parent continuation work according to its requirements.
Full dependency/contract checking remains deliberately stubbed for the minimal loop. Pending describes
outstanding handed-off work; completion and retirement semantics remain open.

## The kernel is the root driver

Kernel is registered first at boot with seed capabilities for minting authority, loading, and persistence.
Scheduler is a CTU component, not part of Kernel. Only the root driver holds authority-minting capabilities;
grants flow downward from human policy. CC enforces those grants.

Recovery is reached through a Move addressed to the kernel driver, using the same loop. A missing
executable discovered before driver execution is a resolution concern, not output from a driver that
never ran. Turning that condition into kernel-directed work, correlating recovery, and installing the
resolved continuation through Scheduler still need concrete interfaces.

## Open and implementation reconciliation

- CC and Scheduler methods, including dispatch outcomes/callbacks and the single scheduler write path.
- Live invocation correlation and parent linkage for repeated action/params.
- Registration, continuation installation, and failed recovery handling.
- Root input, root-result delivery, and completion detection; no ready work does not imply completion.
- Pure computation/composition behind a driver so CC remains a coordinator.
- Reconcile `engine.go`: Kernel.Sched, shared external Memory, Event-returning opcodes, and legacy
  Claim/Landed surfaces do not implement this model. Keep full parent-contract checking stubbed.

## Related

- [Current architecture summary](../../summaries/2026-09-08-the-main-loop-and-execution-boundaries.md).
- [Drivers and the boundary](drivers-and-the-boundary.md).
- [The inference substrate](inference-substrate.md).
- [Move, register, and return](move-register-return.md) — earlier live shapes; current ownership is above.

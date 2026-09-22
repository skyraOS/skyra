# Focus points — 2026-09-09

**Status: WORK LIST.** These are implementation and design threads extracted
from the current v14 runtime and the driver-device discussion. They are ordered
by dependency, not by estimated effort.

## 1. Clean up driver registration — MINIMAL LOOP FIRST

Promote registration from a raw `{Driver, InputValidator}` map on the CTU into a passive
top-level `DriverRegistry` store with two structurally-separated doors:

```text
Resolver{ Resolve(address) -> driver }   read door,  handed to the CTU/controller (hot path)
Registrar{ Register(address, driver) }   write door, handed only to the kernel (traps)
```

Keep live registration: a driver must be loadable after CTU construction, with no
restart. `Register` claims a name (duplicate = error). `Resolve` is direct and ungated.
That is the whole pass.

Lifecycle — replacement, draining, unloading, generation, in-flight leases — is
**deferred** (items 4/5/7 below) until this loop exists and the kernel-as-driver trap
path is in place.

Primary specs: `minimal-registration.md` (the decision), then `driver-registration.md`.

## 2. Clean up the loop

Make the execution loop legible as one owned path from admission through
dispatch, driver output, scheduler write, child completion, parent resumption,
and root notification. Remove duplicated or competing transition paths, keep
state mutation under scheduler ownership, and keep driver/host callbacks outside
the scheduler lock.

As part of that cleanup, separate the kernel protocol from native driver
protocols. Preserve the v14 compatibility surface while allowing JSON, RPC,
subprocess, filesystem, GPU, or hardware drivers to use different native
protocols behind the same kernel-facing shape.

Primary specs: `ingress-egress-adapters.md` and the v14 scheduler/controller
implementation.

## 3. Add the kernel as the first driver

Represent the kernel itself through the driver boundary first. It should be the
initial trusted/system driver registered by the runtime, exercising the same
ingress, egress, descriptor, authority, and lifecycle path that later drivers
use.

The kernel driver is special in privilege, not in bypassing the protocol. Its
implementation may own privileged operations and durable state, but the runtime
should still make the crossing explicit, source-bound, and observable. This
gives the system a concrete first device against which driver registration,
adapter behavior, and dispatch semantics can be tested.

The kernel driver should establish:

```text
kernel address
privileged descriptor
durable-operation boundary
authority/capability check
completion and fault path
```

Do not make the kernel a hidden alternate execution path. If the kernel needs a
special privilege mechanism, specify that as part of its driver contract.

## 4. Define driver lifecycle ownership

Decide who opens and closes a driver, where health and fault state live, and what
the runtime promises when an asynchronous driver remains silent or cannot drain.
Do not imply that unloading cancels work already accepted by a remote device.

## 5. Bind execution to driver generations

Record the selected driver address and generation on a continuation or its
dispatch record. Expose this in snapshots and failure reports so execution
provenance is inspectable.

## 6. Preserve the kernel boundary

Keep source binding, identity allocation, parent linkage, horizon inheritance,
and scheduler mutation in the controller/scheduler. A driver or adapter may
produce candidate output but may not forge architectural state.

## 7. Test hot loading and replacement

Add focused tests for:

```text
load after CTU construction
failed load leaves no registration
replacement is atomic
old generation drains
new work uses the new generation
in-flight work finishes on the old generation
unload rejects new work
late output remains source-bound
```

## 8. Revisit the v14 compatibility facade

Keep [`api.go`](../../src/skyra/skyra-dna-v14/api.go) stable while the internal
driver model evolves. Decide whether `RegisterDriver` remains the simple add
operation or becomes a compatibility alias for `Load`.

## 9. Do not expand scope prematurely

Compiled contract interpretation, retries, durable storage, network transport,
and multi-wave scheduling remain separate threads. Driver lifecycle work should
not quietly become a redesign of those systems.

## Open

The immediate next decision is the lease boundary for asynchronous execution:
when exactly a driver generation's in-flight count is decremented. That choice
controls safe draining and must be settled before replacement is implemented.

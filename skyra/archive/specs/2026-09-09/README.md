# 2026-09-09 — drivers as loadable execution devices

**Status: WORKING PROPOSAL.** This set follows the local Invoke/Return runtime in
`src/skyra/skyra-dna-v14/`. The current pass is deliberately narrowed to the
**minimal registration loop**: promote the driver registry into a passive top-level
store with two structurally-separated doors — a `Resolver` (read, on the hot path,
handed to the CTU) and a `Registrar` (write, kernel-gated). Making a driver *reachable*
at runtime is the whole goal for now. Device management (replacement, draining,
unloading, generations, leases) and the byte-level adapter split are **deferred** until
that loop exists. The decision is recorded in `minimal-registration.md`; the larger
device model is preserved as deferred material in the other docs.

## Read in this order

```text
1  minimal-registration.md   the DECISION: the passive store, the two doors, the trap-to-kernel. read first.
2  driver-registration.md    the registry contract for this pass; larger device model kept as deferred.
3  ingress-egress-adapters.md the driver boundary; v14 Ingress/Egress retained, adapter split deferred.
4  focus points.md           the implementation focus list; this is the work queue, not new canon.
```

## Relationship to the existing runtime

`skyra-dna-v14/driver/driver.go` currently defines `Driver` as `Ingress` plus
`Egress`, and `DriverRegistry` as a mutex-protected address-to-registration
map. `ctu.RegisterDriver` already permits adding an endpoint after construction.
This set keeps that direction and specifies the missing device semantics around
it; it does not replace the v14 Invoke/Return protocol.

The current `DriverRegistration` is therefore the seed of the new model, not a
mistake to discard:

```text
current:      address → { Driver, InputValidator }   (a map the CTU holds)
this pass:    a top-level DriverRegistry store; two interfaces over it —
              Resolver{Resolve} to the CTU, Registrar{Register} to the kernel
deferred:     address → { Descriptor, Adapter, Driver, Lifecycle, Generation }
```

The scheduler continues to own continuation state (the matching store); the registry is
its own passive instruction store, owning driver presence. This pass adds no lifecycle:
`Register` claims a name (duplicate = error) and `Resolve` reaches it — nothing more.
The deferred device model below (generation, lease, drain) returns only after the
minimal loop exists.

## The CPU correspondence

```text
instruction       Invoke
opcode/target     Action + Target
register frame    Invoke.Register
device/controller registered Driver
bus protocol      Adapter
completion        bound Return / callback
capability        Horizon / Authority decision
hotplug           Load / Replace / Drain / Unload
```

The useful abstraction is the stable bus contract, not a claim that every
driver is hardware. The kernel owns the envelope and custody rules; the driver
owns its native execution protocol; the adapter translates between them.

## Design rules

1. A driver is published only after its descriptor and adapter validate and its
   `Open` lifecycle step succeeds.
2. Publication is atomic. New dispatches see either the old ready generation or
   the new ready generation, never a partially initialized device.
3. Existing work is not silently moved to a replacement. It keeps the driver
   generation selected for its dispatch.
4. Replacement is two-phase: publish the new generation, then drain and close
   the old generation after its in-flight work reaches zero.
5. Removal rejects new dispatches before closing the device. It does not claim
   that already accepted remote work was cancelled.
6. Driver output is untrusted data at the boundary. The controller binds its
   source, root, parent route, and invocation identity; the adapter cannot mint
   those facts.
7. Driver registration is a runtime operation, not a process-startup-only
   configuration step.

## Open

The set does not yet choose whether drivers are Go plugins, in-process objects,
subprocesses, or remote services. It also does not settle capability signing,
health checks, persistence of registrations, or the exact public API for
replacement. Those choices must preserve the stable kernel-facing contract.

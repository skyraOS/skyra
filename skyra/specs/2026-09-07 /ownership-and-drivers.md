# Ownership and drivers

**Reviewed:** 2026-09-07  
**Status:** Surviving boundaries, reconciled with the current loop. These are design responsibilities,
not a claim that all methods are implemented.

## Responsibility

| Owner | Responsibility |
|---|---|
| Human / principal | Goals, policy, and delegated authority. |
| Kernel | Coordinate and provision execution, contain Scheduler, arrange loading and recovery, and provision authority. |
| Scheduler within Kernel | Maintain lifecycle state, coordinate resident lookup and provisioning, mark handed-off work Pending, and process returned Moves. |
| Memory | Own the fd table and register consistency, publication, and lifetime operations. |
| CTU | Execute handed-off fd entries according to opcode semantics, check access, and emit Moves. |
| Driver | Supply a concrete mechanism behind its interface. The filesystem driver also resolves remembered executables. |

Kernel and CTU share one Memory instance. Kernel provisioning and scheduler-side return processing
use Memory operations. Memory maintains its own structures; orchestration does not reach into its
maps to bypass those methods.

The CTU produces results. Scheduler-side return ingress publishes the returned observations through
Memory. The return path holds parent work until its requirements permit release. Pending describes
outstanding handed-off work, not a general readiness certificate.

## Mechanism is not authority

A driver being installed does not authorize a workflow to use it. Knowing a resource exists,
presenting it as an available choice, and granting access to it are different responsibilities.
The human supplies policy; kernel provisioning establishes authority; CTU checks the authority it
receives. An operation or model choice cannot authorize itself.

Driver interfaces separate behavior from platform-specific mechanisms. An implementation can change
with the hosting environment without requiring the remembered behavior to use a platform-specific
wire representation. This is the continuity boundary, not a claim that all backends already exist.

Host identity and entity routing are being designed separately in
[egress-entity-host](host-entity-address-horizon/README.md). Do not equate Host with a trusted privilege ring
or infer routing-map ownership from the older capability-table discussions.

## Recovery boundary

The CTU reports an execution problem; kernel coordination arranges recovery. The CTU must not also
independently start the same recovery through a second handler path. Recovery infrastructure must be
available without first executing the missing engram.

Missing backing, an unfinished child, denied access, and an operation failure are distinct conditions.
Their concrete return representations and transitions remain open in the Move-in/Move-out loop.
This document does not restore the removed Issued/Landed events or a CTU-owned handler table.

## Reviewed sources

Selected principles only, not whole-file imports:

- [08-24 ownership](../2026-08-24/skyra-dna-v11/ownership.md): human policy, kernel provisioning, CTU execution.
- [08-24 CTU bootstrap](../2026-08-24/skyra-dna-v11/bootstrap/ctu.md): driver mechanism versus authority; platform-independent interface.
- [08-24 registration and authorization](../2026-08-24/skyra-dna-v11/research/authorization-and-registration.md): availability does not confer permission.
- [09-01 loading](../2026-09-01/engram-loading-and-fault-return.md): shared Memory and kernel recovery coordination.

Excluded from these sources: fixed bootstrap order, main-exits recipe, kernel-does-not-coordinate,
Host=trusted-system-block, live horizon-table view, specific grant protocols, and CTU direct recovery.

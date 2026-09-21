# Agent / Surface worker

Date: 2026-09-20

Status: Proposed composition to test. This role describes execution of untrusted
agent or capability behavior, not every use of the Surface type.

Power: perform the particular work using the resources and authority supplied to it.

## Composition

The existing [Surface and Driver declarations](surface/proposal.go) are placed
here as the behavior contract. The worker executes the assigned behavior and
emits outcomes or further requested Moves. The parent retains continuation state;
a disposable worker process need not own the logical invocation's whole lifetime.

Emitted Moves are requests. They re-enter interpretation and Horizon; they do
not directly invoke another registered Driver or mutate canonical state.

## The Surface abstraction crosses roles

A Surface is an addressable capability, not a trust classification. The same
structure is used for the authority-side TrapController. Keeping the shared
declaration here is an organizational choice, not a claim that TrapController is
an untrusted agent or that the worker owns the registry's Surface record.

The present Surface struct combines identity, contracts, dependencies, profiles,
and Driver behavior. Those fields do not all confer worker powers:

- Registration and profile installation belong to trusted runtime machinery.
- The worker executes supplied behavior with limited inputs/resources.
- A worker cannot change its authority by changing a local copy of a profile,
  contract, dependency list, or SurfaceID.
- The ingress/egress state profiles are invoked by the runtime under the state
  turn discipline; they are not permission for worker code to access state maps.

This is a real pressure on the three-role proposal: we may need to separate the
registered Surface definition, trusted protocol adapter, and worker entry point.
Do not settle that split by moving a file alone.

## Test of the boundary

Assume the worker is fully compromised. What can it do without another admitted
Move? The resources and privileges it actually receives must fit its authorized
work. A separate pool or child process alone does not demonstrate that constraint.

The worker must not gain registry mutation, Horizon provisioning, general
StateOwner access, or the ability to attribute output to another invocation.
Trusted machinery still needs to enforce those limits and define any permitted
direct external effects.

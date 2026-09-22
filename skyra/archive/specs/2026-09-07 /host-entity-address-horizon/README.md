# Host, entity, address, and Horizon

Current engram thread: [Engram, invocation, and work unit](work-unit-and-invocation.md).
EngramID → InvocationID → timestamped WorkUnits, each with workArtifact, operatingEntity, and change.
Declarations are in [the engram proposal](proposal-engram.md).

Host stub: [Host address and forwarding](host-address-and-forwarding.md), with declarations in
[the structs/methods proposal](proposal-entity-address-host.md).

Latest decision: [Address owns representation](address-and-representation.md). The address uniquely
identifies the entity–host endpoint binding and owns its parser/representation contract.

Continue here: [Working context, compilation, and Remember](execution-history-and-live-work.md).
History informs changes; live work is compiled. The latest successful compile is current for new
invocations. User-initiated Remember is the separate commit boundary.

Latest decision: [Move and Engram](move-and-engram.md) — entity, address, and host exist on the Move;
the executable thing exists on the Engram.

**Date:** 2026-09-07  
**Status:** Working design thread. Records the agreed conceptual distinctions before choosing routing
structures or interfaces.

## Why we are here

The CTU executes work and emits Moves. We are designing how egress routes those Moves to the intended
entity, including local delivery back to scheduler ingress. The current CTU has `Local map[c.Target]Driver`
and `Out Egress`; separate host/entity routing maps have not been declared. No callback, channel, or
new routing structure has been selected.

## Host is identity, not address

A host identifies the underlying hosting environment. Its identity is intended to remain stable while
software, endpoints, or execution resources within that environment change.

The motivating examples are a device identified independently of its installed OS, and an AWS
environment identified by its account rather than by individual VMs or functions. These examples
express the desired abstraction; the concrete identifier sources and guarantees still need choosing.
The cloud account does not need to be subdivided into host identities for each underlying VM merely
because those machines exist. A separately identified environment can provide its own stable boundary.

Replacing the identified machine or moving to a different identified account is a host-identity change.
Do not invent a persistent logical HostID whose purpose is to conceal that replacement.

An existing suitable unique identifier can be used directly; a generated UUID is not required merely
to replace it. Preserve the issuing namespace so identifiers from different systems remain distinct.
The exact HostID encoding is open.

## Entity and address

Entity identifies what we want to contact. For example, a device's underlying OS can be an entity;
the device is its host.

**An address belongs to the binding between an entity and a host.** It answers:

> How do I contact this entity on this host?

An entity can have bindings on multiple hosts, and a host can support multiple entities. The contact
information on a binding can change without changing either identity. An address is neither the host's
identity nor a permanent identity for the entity.

The existing vocabulary offers `Entity{ID, Homes []HostID}` and `Target{Entity, Host, Name}`. These are
starting points, not a completed representation of the address-bearing binding.

## Horizon: authority alongside routing

Host, entity, and address identify the hosting environment, the intended participant, and how to
reach that participant on that host. Horizon answers a different question: **what is this workflow
permitted to do?** Knowing an address or finding an engram does not authorize an operation.

Horizon is kernel-provisioned workflow authority carried by Move. Its capabilities use the existing
Target/Action vocabulary, including expiry and signature. CTU checks access when executing an operation;
choosing a route cannot mint or widen authority. Focus selects visible choices from Horizon without
becoming another permission boundary.

The relationship between routing bindings, engram availability, and capability targets belongs in this
design thread. We have not decided to store Horizon on a host, entity, or address binding. Keep its
workflow ownership explicit while deciding how the structures connect. Moving an entity to a new host
does not by itself authorize the replacement target; authority for that access must still be established.

See [authority and Focus](../authority-and-focus.md) for the surviving authority decisions.

## Continuity across host replacement

**We do want continuity across host replacement.** That does not require the replacement to keep the
old HostID. Entity continuity, remembered behavior in engrams, and the driver abstraction separate the
system's behavior and data from the underlying host.

The replacement host has its own identity. Entity bindings can move to it, while engrams preserve
remembered behavior and drivers provide the mechanisms available in the new environment. The migration
and provisioning mechanics are not yet specified; this is the architectural separation supporting them.

The earlier suggestion to preserve HostID across machine/account replacement confused host identity
with continuity of the entity and its behavior. That suggestion is not the accepted direction.

## Engrams and the next decisions

The entity/host mapping also needs to support engram resolution. Engram identity must remain distinct
from its current location; moving or replicating remembered behavior should not make its address part
of its identity. The current execution decision equates MoveID and EngramID; this routing discussion
does not introduce another executable identity.

Still to discuss:

- The entity–host binding data structure and how its addresses are represented.
- Where engram availability belongs in relation to entities, hosts, and their bindings.
- How Horizon capabilities relate to those bindings and engram availability, including host replacement.
- Ownership and maintenance of these mappings, including local and remote egress routing.
- How existing Target, Entity, Local, and Egress fit the resulting structure.

No concrete maps, fields, routing algorithm, or transport have been accepted yet. Continue from these
distinctions rather than treating a host as an address or equating host replacement with loss of continuity.

See the [previous checkpoint](../../../src/skyra/skyra-ctu-engine/checkpoint.md) for the execution-loop decisions leading to this thread.

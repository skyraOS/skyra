# Entity address, driver, protocol, and transport

**Status: DECIDED DIRECTION.** This note reconciles the stable entity-address model
from the September 7 specs with the driver composition and transport boundary
developed on September 10.

Supersedes neither the entity identity nor address-owns-representation decisions.
It assigns driver, protocol, transport, and route their places beneath that stable
address.

## Decision

Callers address a stable semantic face of an entity. They do not address a driver
implementation or transport.

```text
public target:

    entity / entity-address / capability-address

resolved implementation:

    (entity, address)
            |
            v
      DriverBinding
        +-- driver
        +-- protocol
        +-- transport
        +-- route
```

The rejected public shape is:

```text
entity / transport / driver / action...
```

Transport and driver are omitted because they may change while the addressed
entity contract remains the same. Putting either in the public target would turn
an implementation replacement into a false change of semantic destination.

## Vocabulary

```text
Entity       who is being reached
Address      which stable semantic face of that entity is being used
CapabilityAddress
             which stable callable operation is requested through that face
Driver       the local implementation of the address contract
Protocol     the message rules implemented by ingress and egress
Transport    how protocol messages physically move
Route        the current host, process, socket, or network locator
```

An address is stable because it names a representation and behavior contract, not
because its implementation can never be replaced.

## Capability address

An entity address contains stable callable addresses:

```text
Entity
    1 ---- * EntityAddress
                  1 ---- * CapabilityAddress
```

The target hierarchy is:

```text
entity / entity-address / capability-address
```

For example:

```text
openai/gpt-6/codex/run
openai/gpt-6/codex/review
openai/gpt-6/responses/create
```

Semantically:

```text
openai/gpt-6   Entity
codex          EntityAddress
run            CapabilityAddress
```

A `CapabilityAddress` names a callable function and its contract. A `Capability`
is not the function: it is kernel-minted authority to invoke that function, possibly
under narrower constraints.

```text
CapabilityAddress   names a callable function
Capability          grants authority to call it
Horizon             contains the possibility's granted Capabilities
Focus               contains the currently exposed subset
```

Conceptually:

```go
type CapabilityRef struct {
	Entity     EntityID
	Address    AddressID
	Capability CapabilityAddress
}

type Capability struct {
	Ref         CapabilityRef
	Constraints Constraints
}
```

These types are explanatory, not an approved Go surface.

### Description is not authority

A driver binding exposes descriptors for the functions it implements:

```go
type CapabilityDescriptor struct {
	Address      CapabilityAddress
	Description  string
	InputSchema  Schema
	OutputSchema Schema
}

type DriverDescriptor struct {
	Entity       EntityID
	Address      AddressID
	Capabilities []CapabilityDescriptor
}
```

This produces four distinct views:

```text
DriverDescriptor.Capabilities
    everything the address implementation claims to provide

Horizon
    everything the possibility is authorized to invoke

Focus
    the authorized subset currently exposed to the executing entity

Invoke target
    the exact CapabilityRef selected for this child execution
```

Advertising a function does not authorize it. A driver may describe availability,
but only the kernel may mint a corresponding `Capability` into Horizon.

The full flow is:

```text
driver registration
    |
    +-- publishes CapabilityDescriptors
            |
            +-- kernel policy admits and mints
                    |
                    +-- Horizon []Capability
                            |
                            +-- narrowed into Focus []Capability
                                    |
                                    +-- ingress exposes entity-native tools
                                            |
                                            +-- entity selects CapabilityRef
                                                    |
                                                    +-- scheduler checks Horizon
                                                            |
                                                            +-- child Invoke
```

For Codex, ingress projects Focus into `dynamicTools`. The dynamic registry is
discoverability only; the child invocation must still pass the ordinary Horizon
check.

### Existing target shape

The current DNA15 target surface does not yet fully encode the decided hierarchy:

```text
Target.Entity   -> Entity
Target.Address  -> EntityAddress
Target.Capability -> CapabilityAddress (intended)
```

`Invoke.Action` is not the capability-address slot. It is superseded by the
method-header observation at `Invoke.Operands[0]`; the capability address belongs
in `Target`. See
[`continuation-language-operands.md`](../2026-09-11/continuation-language-operands.md).

## Cardinality

```text
Entity         1 ---- * EntityAddress
EntityAddress  1 ---- * CapabilityAddress
Driver instance  ---- 1 EntityAddress
EntityAddress  1 ---- 1 active DriverBinding per possibility
DriverBinding  1 ---- 1 configured transport initially
```

Over time, one `EntityAddress` may be served by multiple compatible driver
generations. Only one generation is active for a particular possibility/binding.

Driver code or a driver factory may be reused to construct many bindings. The
one-address rule applies to an instantiated driver binding: it implements one
address and cannot ambiguously speak for several contracts.

The initial design configures one transport per binding. A future driver may hide
failover across several routes or transports, but that does not change the public
address or expose route selection to an engram.

## Address owns the contract; driver implements it

The earlier relationship remains:

```text
Entity
    |
    +-- EntityAddress
            +-- representation / behavior contract
            +-- active DriverBinding
                    +-- ingress translation
                    +-- egress translation
                    +-- protocol behavior
                    +-- transport
                            +-- current route
```

The address is the contract: where the entity can be addressed and how interaction
through that face is represented. The driver is the local implementation: how this
host actually crosses that address.

Ingress and egress are therefore interpretations of the address contract in each
direction:

```text
Ingress: Skyra Invoke / Return -> address representation -> protocol
Egress:  protocol event         -> address representation -> Skyra Move(s)
```

## Proposed conceptual shapes

These are explanatory shapes, not an approved Go surface:

```go
type EntityAddress struct {
	Entity   EntityID
	Address  AddressID
	Contract ContractID
}

type DriverBinding struct {
	Address    EntityAddress
	DriverKind DriverKind
	Transport  TransportConfig
	Route      RouteConfig
	Generation uint64
}
```

The registry resolves the stable key to an implementation binding:

```go
type Resolver interface {
	Resolve(entity EntityID, address AddressID) (DriverBinding, error)
}
```

Whether the current implementation keeps `Entity` and `Address` as separate lookup
arguments or encodes them into one canonical address key remains an encoding choice.
The semantic distinction must survive either encoding.

## Example: one entity, two addresses

An exact model entity can be reachable through Codex and through the OpenAI API:

```text
Entity: openai/gpt-6

+-- Address: codex
|   +-- contract: Codex agent invocation
|   +-- driver: Codex App Server driver
|   +-- protocol: Codex App Server JSON-RPC
|   +-- transport: managed process over stdin/stdout
|
+-- Address: responses
    +-- contract: Responses API request
    +-- driver: OpenAI Responses driver
    +-- protocol: Responses API
    +-- transport: HTTPS
```

These are distinct addresses even though they ultimately reach the same model
entity. They expose different representation and execution contracts:

```text
codex address
    threads, turns, agent harness, dynamic tools, Codex auth and approvals

responses address
    API requests, response ids, API auth, Responses tool-call semantics
```

Example public targets:

```text
openai/gpt-6/codex/run
openai/gpt-6/codex/review
openai/gpt-6/responses/create
```

The exact string encoding is still open. What is decided is that the target selects
entity, stable address, and capability address—not its transport or driver
implementation.

## Stable replacement

A transport replacement does not change the address:

```text
before:
    openai/gpt-6/codex -> CodexDriver(ProcessStdio)

after:
    openai/gpt-6/codex -> CodexDriver(UnixSocket)
```

Engrams, Focus, and Horizon retain the same semantic target. Only the binding and
possibly its generation change.

A driver replacement may retain the address only when it preserves the address's
contract. An incompatible change to representation or behavior requires a new
address. Stable addressing must not be used to disguise a semantic discontinuity.

## Authority consequence

Capabilities should name the stable semantic target:

```text
(Entity, EntityAddress, CapabilityAddress)
```

They should not grant a particular transport, route, process id, or driver
generation. The kernel authorizes what may be reached; the active binding determines
how it is reached inside the possibility's authority envelope.

This permits compatible transport and driver replacement without reminting Horizon.
If replacement changes the authority implications of the address, then it is not a
compatible replacement and must be registered as a distinct address/contract.

## Placement with possibility

The `EntityAddress` is stable across possibilities. The active `DriverBinding` is
installed per possibility because process, credentials, filesystem access, network
policy, and routes belong to the possibility's kernel-minted authority context.

```text
stable EntityAddress
        |
        +-- Possibility A DriverBinding
        |       +-- process / transport / route under Horizon A
        |
        +-- Possibility B DriverBinding
                +-- process / transport / route under Horizon B
```

Two possibilities can therefore address the same entity face while using isolated
processes, credentials, and transport resources.

## Invariants

```text
an entity may expose multiple stable addresses
an entity address may expose multiple stable capability addresses
an instantiated driver binding implements exactly one address
the public target names entity + entity address + capability address
the public target does not name driver, protocol, transport, or route
the address owns the representation and behavior contract
the capability address names a callable function beneath that contract
describing a callable function does not authorize it
only the kernel mints Capability grants into Horizon
Focus is an exposed subset of authorized capabilities
the driver implements that contract through ingress and egress
transport and route remain replaceable implementation details
compatible replacement preserves the address
incompatible semantic replacement receives a new address
bindings are installed inside a possibility's authority envelope
```

## Open encoding and lifecycle questions

- Canonical encoding of `EntityID`, `AddressID`, and `CapabilityAddress`.
- Whether the registry key is a tuple or one encoded stable address value.
- Where contract identity and compatibility checks live during registration.
- Descriptor/schema dialect and capability-address collision rules.
- Whether capability constraints are data, predicates, or references to
  kernel-owned policy.
- How generation, drain, and in-flight binding retention are represented.
- Whether a binding may select among several equivalent routes without exposing
  that choice above the driver.
- How entity identity is stated when a coarse entity, such as Codex plus its model,
  is later split after failure-domain evidence.

## Prior and related decisions

- [September 7 entity/address/host proposal](../2026-09-07%20/host-entity-address-horizon/proposal-entity-address-host.md)
- [September 7 address owns representation](../2026-09-07%20/host-entity-address-horizon/address-and-representation.md)
- [September 7 drivers and the boundary](../2026-09-07%20/drivers-and-the-boundary.md)
- [Driver composition and opaque state](driver-composition-and-state.md)
- [Codex ingress boundary](codex-ingress-boundary.md)
- [Continuation-language ordered operands](../2026-09-11/continuation-language-operands.md)

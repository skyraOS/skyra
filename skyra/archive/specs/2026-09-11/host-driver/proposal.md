# Surface: the recursive execution boundary

Status: proposal for review. This document does not authorize implementation.

## Decision

`Surface` is the one recursive interface used to expose anything Skyra can
address and execute:

```go
type Surface interface {
	Address() Address
	Capabilities() []Surface
	Driver() Driver
}
```

An entity address, a transport, a host address, and a leaf capability are not
separate kinds of runtime machinery. They are concrete surfaces with different
semantics.

A surface provides:

- one stable address at its level;
- zero or more child surfaces;
- one driver that translates execution at that level.

This removes the proposed `HostDriver`, `ProcessDriver`, `TransportResolver`,
`CapabilityDescriptor`, `DriverInput`, `HostInput`, and `HostRequest` concepts.
Their useful distinctions remain in concrete surface implementations, not in
parallel execution architectures.

## The common driver contract

Every surface is entered with the canonical execution object, `Move`. Every
driver emission is normalized back into canonical moves:

```go
type Driver interface {
	Ingress(
		ctx context.Context,
		move Move,
		output func(DriverEmission, error) error,
	) error

	Egress(emission DriverEmission) []Move
}
```

`Ingress` interprets the addressed surface's input semantics and begins or
continues work. `Egress` interprets that driver's output semantics and returns
zero or more moves to the runtime.

There is no first-class input validator. A driver may use private validation
helpers. If its operand is malformed, ingress emits the failure and egress
normalizes the failure into the runtime's language.

## Recursive composition

```text
Entity
└── entity-address Surface
    ├── local-transport Surface
    │   └── capability Surface
    └── remote-transport Surface
        └── capability Surface
```

The exact graph is supplied at boot. The interface does not prescribe a fixed
depth or require every entity to expose both local and remote transport.

At any node:

```text
address identifies this surface
capabilities expose the next reachable surfaces
driver interprets execution at this surface
```

`Capabilities() []Surface` is recursive. A leaf returns no children. Alternative
implementations are sibling surfaces, rather than a `Drivers() []Driver` field.
Each surface has exactly one driver.

## Entity and host lineage

The September 7 proposal supplied these concrete topology objects:

```go
type Entity struct {
	ID        EntityID
	Addresses []*EntityAddress
}

type Host struct {
	ID HostID
}

type HostAddress struct {
	Address string
	Host    *Host
}

type EntityAddress struct {
	Address       string
	HostAddresses []*HostAddress
}

type History struct{}
type LiveWork struct{}

func (a *EntityAddress) Representation(history *History, liveWork *LiveWork)
```

This proposal retains the identities and relationships while reconciling their
executable boundary with `Surface`:

- `Entity` remains a concrete identity and the root of its address surfaces.
- `EntityAddress` is a concrete `Surface` implementation and continues to own
  representation semantics.
- `Host` remains the identity of a hosting environment.
- `HostAddress` is a concrete `Surface` through which host behavior is reached.
- transport and capability surfaces compose below an address as required.

The interface does not erase the semantic distinction between entity, host, and
transport. It only says that all three expose execution in the same shape.

The stable part of a host address is its semantic address. Volatile connection
details belong behind the relevant transport driver's boundary. This refines the
older interpretation in which `HostAddress` itself could be treated as mutable
contact information.

## Concrete surface families

These names describe implementations, not new runtime interfaces:

```go
type EntityAddressSurface struct { /* identity, children, driver */ }
type HostAddressSurface struct { /* host identity, children, driver */ }
type LocalTransportSurface struct { /* local transport semantics */ }
type RemoteTransportSurface struct { /* remote transport semantics */ }
type CapabilitySurface struct { /* leaf or recursively composed API */ }
```

Local and remote transport stay distinct because their failure, lifecycle, and
framing semantics differ. A `Local bool` flag would hide that distinction.

Transport framing belongs to the transport surface's driver. For example, JSONL
framing is interpreted there; it is not a scheduler concept. Entity protocol
meaning belongs to the relevant entity or capability driver. Host provisioning
meaning belongs to a privileged host surface's driver.

Transport normally is not selected in advance by the scheduler or dispatcher.
The addressed driver can select among its child transport surfaces using local
knowledge of the entity address and available graph. If a caller eventually
needs a transport constraint, it may be expressed as an optional boundary hint;
absence means the driver selects. It is not a required field on `Move`.

## Resolution

Execution requires three semantic facts to resolve:

```text
entity + transport + host -> executable leaf surface
```

They need not be represented by three special interfaces. Resolution traverses
the surface graph until it reaches the driver capable of executing the move.

```text
requested entity
└── matching entity-address surface
    └── viable transport surface
        └── host/capability dependency
            └── executable leaf driver
```

Where the resolver is housed is secondary to the surface contract. A map may
index stable addresses for efficient lookup, but the map does not become a new
execution abstraction.

Resolution may discover work that must complete first. The existing continuation
language already carries that dependency relation:

- `InvocationID` identifies the dependency execution;
- `ReturnFd` identifies its caller;
- `Register` records pending children and their observations;
- a `Return` fills the dependency and makes the parent eligible to continue.

A driver can therefore emit child moves while walking toward a leaf. A second
dependency graph or planner is not required initially. Static resolution should
reject cycles; dynamic recursion remains bounded by the possibility's Horizon.

## Authority is not surface description

`Surface.Capabilities()` describes what can be reached below a surface. It is not
the kernel-minted `Capability` carried by a `Horizon`.

```text
surface graph  = what implementations exist and how they compose
horizon        = what this possibility is authorized to traverse or invoke
```

The shared word is intentional at the conceptual level, but the types have
different ownership. A driver cannot grant authority by returning another child
surface. The kernel checks the possibility's Horizon before allowing traversal
or execution.

## Possibility is visibility and authority

A possibility is not a Surface and is not executable. It supplies the visibility,
authority, and state frame for executable Moves:

```go
type Possibility struct {
	ID      PossibilityID
	Horizon Horizon

	Durable *DurableState
	Live    *LiveState
	First   *MoveNode
}
```

Execution starts at `Possibility.First`, follows the MoveNode's Move, and reaches
the addressed Surface driver. Durable and live state remain distributed; the
possibility points to their roots rather than aggregating all surface state.

Possibility lifecycle operations such as terminate are leaf surfaces under
`kernel/ctu`. The possibility ID is an operand to those operations. An OS process
signal remains a separate host-surface mechanism.

## Kernel and boot ownership

Skyra depends on the underlying OS, so the kernel assumes exactly one local host
during boot. That host cannot be replaced after boot. Boot also installs the
initial surface graph and address indexes.

The kernel alone holds authority to invoke privileged host surfaces, including
process provisioning. It creates a possibility at that authority boundary,
installs its Horizon and state references, and points it to the first MoveNode.

The scheduler performs work beginning with that MoveNode. It does not receive raw
commands, process handles, stdin/stdout objects, or transport sessions. Those
remain behind their surface drivers.

## Current DNA-16 move baseline

The currently compiled contract guarantees this shape:

```go
type Move struct {
	Root   PossibilityID
	Target Target
	Op     Opcode
	Stamp  AuthorizationStamp
}

type Target struct {
	Invocation InvocationID
	Entity     EntityID
	Address    string
	Host       HostID
}

type Opcode interface{ opcode() }

type Invoke struct {
	InvocationID InvocationID
	Action       Action
	Params       []Observation
	ReturnFd     InvocationID
	Register     Register
	Horizon      Horizon
	Focus        []Capability
}

type Return struct {
	Observation Observation
}
```

This is the implementation baseline, not the final surface contract. Decisions
already made after that baseline still need one deliberate protocol migration:

- target routing becomes entity/address/capability-aware;
- `Action` moves into the ordered operand list;
- operand position determines continuation-language interpretation;
- the possibility owns the live Horizon instead of accepting new authority from
  each `Invoke`;
- the shared `Move` changes once for all drivers; no host- or transport-specific
  execution envelope is added.

## Ownership map

| Concern | Owner |
|---|---|
| Stable identity | `Entity` / `Host` |
| Recursive address and API graph | `Surface` implementations |
| Address indexing and traversal | boot-installed resolution layer |
| Meaning of a surface's input/output | that surface's `Driver` |
| Local or remote framing | transport surface driver |
| Host provisioning and OS interaction | privileged host surface driver |
| Continuation structure and dependencies | scheduler |
| Permission to traverse or execute | kernel-minted Horizon |
| Live execution handle plus permission | `Possibility` |

## Invariants

1. Every executable boundary is exposed as a `Surface`.
2. Every surface has one stable address, zero or more child surfaces, and one
   driver.
3. Drivers accept `Move`; their emissions normalize to `[]Move`.
4. Runtime-native values enter the boundary before driver-specific values.
5. Driver-specific process, protocol, and transport details do not leak into the
   scheduler.
6. Surface discovery never grants authority.
7. The local host and initial surface graph are fixed at boot.
8. Only the kernel may cross a privileged host boundary.
9. A possibility supplies Horizon, durable/live state references, and its first
   MoveNode; it is not executable.
10. Malformed operands become driver emissions and are normalized through
    egress; validation is not a peer abstraction.

## Open implementation questions

- Whether `Address` is one value type for every surface or a tagged value whose
  variants remain comparable and stable.
- Whether `Capabilities()` returns immutable snapshots or boot-frozen references.
- Whether address indexes live in the kernel, dispatcher, or a boot-owned graph
  object.
- The exact `DriverEmission` fields needed once transport and host prototypes are
  exercised through the common driver contract.
- How the current `Target` fields migrate to an explicit recursive surface path
  without duplicating address information.

These questions can be tested in DNA-16 without adding another top-level
architecture.

## Related decisions

- `universal-surface.md`
- `ctu-kernel-proposal/kernel-assumption.md`
- `ctu-kernel-proposal/possibility-boundary.md`
- `../dispatch-operand-ownership.md`
- `../continuation-language-operands.md`
- `../horizon.md`
- `../codex-driver/03-ingress-egress-mapping.md`
- `../../2026-09-07 /host-entity-address-horizon/proposal-entity-address-host.md`

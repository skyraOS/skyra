# Surface proposal

Status: proposal for review. This is the first contract in the September 12
set; later proposals may depend on it but may not redefine it implicitly.

## Purpose

`Surface` is the universal recursive unit for addressable executable behavior.
Entity addresses, hosts, transports, processes, and capabilities remain
different concrete types while satisfying one boundary.

## Surface is not Move

`Move` and `Surface` remain separate even where their shapes appear similar.

```text
Move       the recorded representation of a procedure
           a decision record
           recorded and interpreted
           not executable

Surface    stable durable execution
           an invocation boundary the runtime can rely on
           addressable
           reachable only through a Move dispatched by a Horizon
```

Stable means the surface identity and invocation contract endure. It does not
mean every value returned by `State()` is immutable.

A Move may name a target, carry ordered procedure state, and relate to other
Moves. A Surface may have an address, state, and child capabilities. That
structural resemblance does not make them the same runtime object.

There is no separate runtime `Procedure` type. The direction of authorship and
interpretation is:

```text
human or LLM
    → authors procedure
        → represented as recorded Move
            → CTU reads live procedure state
                → dispatches Move against Possibility.Horizon
                    → authorized handler reaches stable Surface
```

A Surface cannot be invoked without a Move dispatched through the invocation's
provisioned Horizon. Passing a Move to driver ingress supplies the recorded
procedure; the Move is not independently executable. Directly invoking a
surface driver without horizon dispatch is an effect outside the runtime.

Only a human or an LLM authors a procedure. Drivers may normalize an authored
expression into a Move, and compiled surfaces may materialize previously authored
procedure steps, but neither the CTU nor an adapter invents semantic procedure
intent.

In a later compilation design, recorded Moves may contribute to a compiled
artifact that is installed as a Surface for future invocation:

```text
Move history
    → compile                         future contract
        → artifact
            → install as Surface
                → invoke
```

This proposal defines no compiler, artifact, or installation contract. It keeps
the distinction explicit so those concepts can be added later without making a
live Move executable.

## Contracts

```go
type Surface interface {
	Address() Address
	State() StateMap
	Capabilities() []Surface
	Driver() Driver
}

type Driver interface {
	Ingress(
		ctx context.Context,
		move Move,
		output func(DriverEmission, error) error,
	) error

	Egress(emission DriverEmission) []Move
}
```

The provisional normalized state shape is:

```go
type StateKey struct {
	Value string
}

type StateMap map[StateKey]Observation
```

The exact state value contract remains open for this review. The ownership law
does not: each map describes only the state owned by that surface. It does not
recursively aggregate child state.

## Ownership

```text
Address       stable identity at this boundary
State         normalized local state owned at this boundary
Capabilities reachable child surfaces
Driver        local ingress and egress interpretation
```

Native implementation state stays private. `State()` returns a detached,
normalized view; it must not return a mutable alias into the surface.

For example, a Codex surface may own a first map keyed by ThreadID. A child
capability surface may own further state for that same thread when its protocol
requires it. The parent does not have to copy or combine those maps.

A stateless surface returns `nil` from `State()`. A leaf returns `nil` from
`Capabilities()`.

## Recursive law

```text
Surface
├── Address()
├── State()
├── Capabilities()
│   └── Surface
│       ├── Address()
│       ├── State()
│       ├── Capabilities()
│       │   └── ...
│       └── Driver()
└── Driver()
```

Recursion ends only when a surface has no child capabilities:

```go
func (s *LeafSurface) Capabilities() []Surface {
	return nil
}
```

Traversal uses:

```go
len(surface.Capabilities()) == 0
```

as the leaf condition. `nil` is the canonical empty representation.

The child does not hold a parent reference:

```text
surface graph    parent → child    discovery and dispatch
continuation     child  → ReturnFd live return route
```

The Move proposal owns the exact continuation fields. Surface only requires that
parent knowledge does not leak into child implementations.

## Driver law

The target recorded on a Move selects a surface. The selected surface's driver
interprets that live procedure record and its operands:

```text
Move.Target
    → Possibility.Horizon.Dispatch(Move)
        → authorized Handler
            → Surface.Driver().Ingress(Move)
                → invoke stable Surface behavior
                    → local/native execution
                        → DriverEmission
                            → Driver.Egress(...)
                                → []Move
```

Ingress is inbound normalization. Egress is outbound normalization. The kernel
proposal will own `Realize`, the authorized execution wave that binds them.

`Surface.Driver()` currently exposes the driver handle in the public method set.
That shape alone cannot enforce the horizon-only rule: any Go caller holding a
Surface could call `Driver().Ingress(...)`. Before implementation, the Surface
and Move proposals must choose one enforcement shape:

```text
hide driver access behind a runtime-private surface refinement
or
require an unforgeable horizon-issued handler at driver ingress
```

Convention is insufficient. The eventual type boundary must make invocation
that bypasses horizon dispatch unavailable or invalid.

A surface with child capabilities may still have a driver for behavior addressed
at that surface. A leaf driver reaches the actual implementation. The interface
does not create separate host, process, transport, or capability driver kinds.

## Process pressure test

```text
process Surface
├── local process state
├── stdin capability Surface
├── signal capability Surface
├── file/socket/pipe capability Surfaces
└── process Driver
```

At a leaf:

```text
stdin/signal Move
    → leaf ingress
        → actual process operation

stdout/stderr/exit/fault
    → leaf egress
        → []Move
```

Raw `exec.Cmd`, stdin, stdout, stderr, sockets, and protocol sessions remain
private implementation state. They do not enter Move, scheduler state, or the
normalized state map.

## What is not a Surface

The CTU is not a Surface. It interprets recorded Moves and coordinates surface
resolution while owning a
privileged surface implementation that the kernel mounts.

A Possibility is not a Surface. It provides visibility, authority, durable/live
state references, and the first MoveNode. The CTU interprets the recorded Move
at that node and invokes its target Surface.

## Compile-time checks

Every concrete surface carries a local conformance assertion:

```go
var _ Surface = (*ConcreteSurface)(nil)
```

This applies to every entity-address, host, transport, process, kernel, CTU
privilege, and capability surface.

Typed owner structs declare required child shapes, then project them through the
universal slice:

```go
type ExampleCapabilities struct {
	Read  *ReadSurface
	Write *WriteSurface
}

func (s *ExampleSurface) Capabilities() []Surface {
	return []Surface{
		s.capabilities.Read,
		s.capabilities.Write,
	}
}
```

The typed struct makes required fields compile-time-visible. The assertions prove
that each field's concrete type satisfies `Surface`. Boot remains responsible
for non-nil values, unique addresses, and a coherent graph.

## Invariants

1. Every addressable executable boundary is a Surface.
2. Every capability returned by a Surface is another Surface.
3. A leaf has no capabilities; `nil` is the canonical representation.
4. Each surface has one driver.
5. Drivers accept canonical Move values and normalize emissions to `[]Move`.
6. State is local and distributed, never recursively aggregated.
7. `State()` never exposes a mutable native-state alias.
8. A child surface does not need to know its parent.
9. Surface discovery does not grant authority.
10. Every concrete implementation has a compile-time Surface assertion.
11. A Move is the recorded representation of a procedure and is never itself
    independently executable.
12. A Surface is reached only through a Move dispatched by the invocation's
    provisioned Horizon.
13. Only a human or LLM authors procedure intent.
14. CTU and driver transformations record, interpret, or normalize that intent;
    they do not author it.
15. The final type boundary must prevent or reject driver ingress that bypasses
    Horizon dispatch.

## DNA-16 migration pressure

The following DNA-16 concepts are expected to change when implementation begins:

- string address → driver registration becomes recursive surfaces plus indexes;
- first-class `InputValidator` becomes driver-local interpretation;
- peer `Transport` becomes concrete transport surfaces and drivers;
- special `ProcessHost` becomes implementation behind host/process surfaces;
- `DriverInput` stops wrapping Move with a universal state payload;
- provider state becomes distributed surface-local state;
- kernel becomes the single root Surface.

The following DNA-16 findings survive:

- driver calls occur outside the continuation-state transaction;
- provider-native protocol and live resources terminate at the driver boundary;
- driver output returns through a callback and normalizes to Moves;
- stale or malformed output fails closed;
- children do not require parent implementation references.

## Open for review

1. Is `State()` required on every Surface, or should state remain an optional
   capability surface?
2. If `State()` remains, is `map[StateKey]Observation` the canonical normalized
   value or only the first concrete candidate?
3. Does `Address()` return a globally complete stable address or one segment
   scoped by its parent?
4. What behavior is valid when a non-leaf surface itself is targeted?
5. Must `Capabilities()` return a detached slice, or is the graph immutable
   after boot?
6. Does the runtime hide `Driver()` behind a private refinement, or does ingress
   require an unforgeable Horizon-issued handler?

These questions must close before drafting the Move target path.

# Universal Surface

Status: design finding for review.

## Finding

Anything Skyra can address and execute has the same recursive shape:

```go
type Surface interface {
	Address() Address
	Capabilities() []Surface
	Driver() Driver
}
```

`Surface` is the universal interface. Entity addresses, hosts, transports,
processes, and capabilities remain semantically different concrete types, but
they do not require different execution architectures.

Each surface guarantees:

- a stable address at its level;
- zero or more child surfaces;
- one driver that interprets ingress and egress at that level.

```text
Surface
├── Address()       identity at this boundary
├── Capabilities() reachable child boundaries
└── Driver()        local interpretation
```

## Recursive law

The same interface composes at every depth:

```text
entity Surface
└── address Surface
    └── transport Surface
        └── process Surface
            ├── stdin Surface
            ├── signal Surface
            └── other capability Surface
```

This diagram is illustrative, not a mandatory hierarchy. A surface may expose a
different graph, and a leaf returns an empty capability list.

The child does not need to know its parent. Static surface relationships are
used to discover and dispatch downward. The canonical `Move` carries the live
continuation route back upward.

```text
Surface graph: parent → child       discovery and dispatch
ReturnFd:      child  → caller      continuation and return
```

## Driver law

Every surface uses the same driver boundary:

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

The target selects the surface. Operands describe what the selected surface must
interpret.

```text
Target   = where execution goes
Operands = what the target interprets
ReturnFd = where resulting execution returns
```

For a process surface, ingress can discriminate operand types locally:

```text
Move operands
├── stdin input
├── signal
├── file/socket/pipe operation
└── lifecycle operation
```

The same surface normalizes its emissions:

```text
stdout ─┐
stderr ─┼─> DriverEmission ─> Egress ─> []Move
exit   ─┤
fault  ─┘
```

Raw stdin, stdout, stderr, and OS process handles are not runtime objects. The
host surface owns authority to provision them; the transport/process driver owns
their interpretation and live handles. A possibility may reference the live
state but does not own an executable handler.

## Surface versus possibility

```text
Surface       addressable executable behavior
Possibility   visibility, authority, and state references
MoveNode      first executable continuation node
```

A possibility is not a Surface. It points to its Horizon, durable state, live
state, and first MoveNode. Execution follows that node's Move to a Surface.
Process handles and protocol sessions stay behind their owning surfaces and
never enter the scheduler.

## The one special case

The kernel assumes the local host surface during boot. The host is special by
authority and boot position, not by interface shape.

```text
kernel boot
└── assumed local host Surface
    └── ordinary recursive capability surfaces
```

The assumed host supplies the bottom authority needed to interact with the OS.
It still satisfies `Surface`, uses a `Driver`, and exposes further behavior as
child surfaces.

## What this collapses

The universal interface removes the need for peer execution abstractions such as
`HostDriver`, `ProcessDriver`, `TransportResolver`, and capability descriptors.
Their semantics survive as concrete surface and driver implementations.

It does not collapse:

- concrete entity, host, transport, and process identities;
- kernel authority versus driver implementation;
- static discovery versus live continuation;
- descriptive child surfaces versus Horizon-granted permission.

## Invariants

1. Every addressable execution boundary is a `Surface`.
2. Every child returned by `Capabilities()` is also a `Surface`.
3. Each surface has exactly one driver.
4. Drivers accept canonical `Move` values and normalize emissions to `[]Move`.
5. Children do not retain parent references for return routing.
6. Discovering a surface does not authorize its use.
7. Driver-private resources never become scheduler state.
8. The kernel's assumed host is a refined use of `Surface`, not a second
   execution abstraction.

## Related

- `proposal.md`
- `ctu-kernel-proposal/kernel-assumption.md`
- `ctu-kernel-proposal/possibility-boundary.md`
- `../dispatch-operand-ownership.md`
- `../continuation-language-operands.md`

# Observe, Express, Realize

Status: design finding for review.

## Finding

The execution triad has rederived. The same three-part shape has now appeared in
four generations under different vocabulary:

| Generation | Inbound | Outbound | Binding (execution) |
|---|---|---|---|
| v2 / spacetime (`Space` + `Time`) | `Observe` | `Express` | `Time.Realize` |
| v.05 (`Reality`) | inbound half of `Realize` | `Express` + `Attach` | `Reality.Realize` |
| DNA-16 (`Driver`) | `Ingress` | `Egress` | unnamed; spread across scheduler + `controller.execute` |
| host-driver (`Surface`) | `Ingress` | `Egress` | unnamed |

v.05 collapsed the three into one method that returned a value. spacetime kept
them separate and split their ownership. DNA-16 has ingress/egress but leaves the
binding step unnamed. This note names it and fixes the ownership using the split
spacetime already used.

References:

- `../../../archive/spacetime/space.go` — `Space.Observe`, `Space.Express`, and
  `Time.Realize`.
- `../../../archive/_skyra-dna-v2/skyra-dna-v2/membrane.go` — a `Space` that
  implements `Observe`/`Express` while `o.kernel.Realize(m, m, impulse)` drives
  the wave.

## Ownership

```text
Surface   owns Ingress and Egress   (observe in, express out)
Kernel    owns Realize              (the wave that binds them per execution)
```

Realize is not a method on the surface; it takes the surface as an argument. Only
the kernel opens a Realize wave, so the binding step is also the authority
chokepoint. This is consistent with "only the kernel may cross a privileged host
boundary" and with the possibility supplying the Horizon under which its first
MoveNode reaches an executable surface.

## Perception Invariant

Realize returns nothing. A wave's result is written to the traversed slice and
returns to the origin as an involuntary return perception, routed by the
continuation, not collected by the caller as a chosen return value. The invariant
is stated on `Time.Realize` in `space.go`. It is already implemented in DNA-16 as
the `output func(DriverEmission, error) error` callback plus `ReturnFd`.

Consequences for signatures:

- Realize is void.
- Ingress emits through the output callback; it does not return the result.
- Express performs the synchronous upward normalization (`-> []Move`). It is the
  one member of the triad that returns a value.

## State

Opaque continuation state is not carried on the Move or collapsed into one
possibility-wide map. State remains distributed across its owning surfaces. The
possibility points to durable state, live state, and its first MoveNode. This is
why Move-only `Ingress` remains sufficient.

## Resolution mechanism

Given an invoked surface, trace its dependencies back to the runtime and execute
from there.

```text
invoked surface
└── trace the dependency chain downward
    entity -> address -> transport -> host -> ... -> leaf driver
        the kernel Realizes from the resolved point
            results return upward as perception via ReturnFd
```

Because every surface is the same interface, each stays dumb and local: it knows
only its own Ingress/Egress and its own children; it does not know its parent or
the whole graph. The runtime composes the chain by walking `Capabilities()`
downward for dispatch and following `ReturnFd` upward for return. A driver may
emit child moves while walking toward a leaf; the existing continuation language
(`InvocationID`, `ReturnFd`, `Register`, `Return`) already carries that
dependency relation.

## Distribution

Local-only, uniform surfaces make local and remote execution the same shape. A
caller does not distinguish a local surface from a remote one; resolution
traverses the graph until it reaches a driver that can execute, wherever that
driver lives. The universal interface plus surface-local knowledge is the basis
for distribution: no surface needs global knowledge for the system to compose
one.

## Invariants

1. A surface owns Ingress and Egress. The kernel owns Realize.
2. Realize takes the surface as an argument and returns nothing.
3. Ingress emits through the output callback; results return as perception via
   `ReturnFd`.
4. Express is the only member of the triad that returns a value.
5. Opaque state is slice/possibility-resident, not carried on the Move.
6. A surface knows only its own boundary and children; it holds no parent
   reference and no global graph.

## Related

- `universal-surface.md`
- `proposal.md`
- `ctu-kernel-proposal/kernel-assumption.md`
- `ctu-kernel-proposal/possibility-boundary.md`

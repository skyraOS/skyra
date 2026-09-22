# Horizon

**Date:** 2026-09-11
**Status: FOUNDATIONAL / DECIDED DIRECTION.** Consolidates the authority model
settled across the September 7–11 specs and the containment thread. `DECIDED`
marks a settled rule; `OPEN` marks a choice still owed.

## What Horizon is

> **Horizon is the installed authority context of a running possibility: the set
> of kernel-minted Capabilities that bound everything the run may reach.**

It is the cognitive light cone of a possibility — the edge of what that run can
call and what may cross back into it. Below the kernel a possibility dissolves
into "an invocation tree running under an installed Horizon"; the Horizon *is*
that installed authority context (the CR3 / current address space of the run).

Nothing executes, and nothing returns, outside its Horizon.

## Kernel mints, scheduler checks

Authority has two motions, and they live in different places `DECIDED`:

```text
MINT     kernel only     creates Capabilities and installs a possibility's Horizon
CHECK    scheduler       fast-path check of every crossing against the installed Horizon
```

The analogy is the MMU: the kernel writes the page tables; the MMU checks every
access against them and is consulted on none of the lookups. Horizon-checking
stays on the hot path in the scheduler; only *minting* is a kernel privilege.

DNA15 stamps in the scheduler and trusts `Invoke.Horizon` as a host-supplied
field — that was jumping the gun. The decided direction: Horizon stops being a
trusted field on the Invoke and becomes a property of the possibility the kernel
minted.

## Horizon is not Focus, Capability, or CapabilityAddress

These are routinely conflated and must stay distinct `DECIDED`:

| Term | Meaning |
|---|---|
| `CapabilityAddress` | Names a callable function and its contract. Naming, not authority. |
| `Capability` | Kernel-minted **authority** to invoke a CapabilityAddress, possibly under narrower constraints. |
| `Horizon` | The set of Capabilities granted to a possibility — its full authority envelope. |
| `Focus` | The subset of Horizon currently **exposed** to the executing entity. Discoverability, not authority. |

Advertising a function does not authorize it. A driver descriptor may claim a
function exists; only the kernel may mint a `Capability` for it into Horizon; only
some of Horizon is projected into Focus. For a Codex-backed invocation, Focus is
projected into `dynamicTools` — a model-facing discoverability surface. The child
invocation the model then requests is still checked against Horizon.

```text
descriptor  what an implementation claims to provide
Horizon     what the possibility is authorized to invoke
Focus       the authorized subset currently exposed
target      the exact capability selected for one child execution
```

## Two enforcement boundaries

Horizon is checked at **both** ends of an invocation, and the two checks are not
redundant `DECIDED`:

```text
invoke-side Horizon check    gates what an invocation may CALL
return-side Horizon check    gates what content may cross BACK into it
```

### Return-side check = information-flow control

Each invocation is a type. A returned observation whose **type, capability, or
provenance** is outside the invocation's Horizon is rejected before it enters the
semantic Register — it never becomes the invocation's result, never propagates to
the parent, never becomes a future operand. Within Skyra's causal world the
forbidden result never happened.

This is the enforcement point for effects that cannot be stopped at the OS layer.
A generated image returning into an invocation whose Horizon admits no
image-typed result is rejected at the return boundary — regardless of the fact
that the OS could not distinguish the hosted call from inference traffic. See
[`codex-driver/06-native-tools-and-containment.md`](codex-driver/06-native-tools-and-containment.md).

**Contract requirement `OPEN`:** the return-side check must gate on the returned
observation's type/capability/provenance, not merely on the producing endpoint's
identity. "Each invocation is a type" only enforces if a returned observation
carries a type tag Horizon can reject on. If the current contract checks
endpoint + action but not returned-content kind, that tag and its return check
are the surface to add.

## Horizon and the OS containment boundary

Horizon is the **semantic** authority ceiling. It has a physical shadow: the
process the kernel launches for a possibility must be contained to no more than
its Horizon allows `DECIDED`.

```text
Horizon (semantic ceiling)
    |
    +--derives--> process launch policy (ambient/effect ceiling)
                  filesystem, network, environment, credentials, executable,
                  process-spawn, devices, resource limits
```

The kernel derives the sandbox from the possibility's Horizon; the guest's own
configuration is advisory and can never be the boundary. Provider-level sandboxing
may narrow further but is never the source of authority.

The two ceilings cover different failure modes:

- **Local effects** (shell, file write, subprocess, browser, device) — bounded by
  the OS sandbox; even if the guest re-enables a tool, the effect is denied.
- **Hosted effects that share the inference connection** (web search, image
  generation, apps/connectors) — the OS cannot separate them from inference.
  These are bounded at the **return boundary** (typed-result rejection) and, where
  that is insufficient (state-mutating connectors), by registry exclusion.

Measured baseline for the OS ceiling (macOS 15.6 / M1 / Seatbelt): file-read,
file-write, process-exec (allow one binary, deny all other exec), and
network-outbound denial are all enforceable during the process at operation
granularity. macOS has no per-syscall filter; enforcement is operation-level via
MACF/Seatbelt. Escape-resistance beyond that is a VM concern, deferred.

## The two boundaries

Horizon is enforced in two independent planes. They are not a stack of the same
check; they cover disjoint failure modes, fail independently, and are both
derived from the same Horizon `DECIDED`.

**1. The process boundary — physical / effect.** *What can this process do to the
world?* Enforced below the guest by the OS sandbox (Seatbelt/MACF): filesystem,
network, subprocess exec, devices. It is **preventive of the action** — the
effect never occurs. It bounds local effects (shell, patch, browser, spawned MCP
servers). It is blind to anything that rides the inference connection.

**2. The runtime boundary — semantic / admission.** *What can become part of
Skyra?* Enforced above the guest by the scheduler checking Horizon at invoke and
return. It is **preventive of the consequence** — out-of-Horizon content never
enters the causal world (never a result, never propagates to a parent, never a
future operand). It bounds exactly the hosted effects the OS cannot see: a
returned image, a typed result.

```text
                   prevents         catches                  blind to
process boundary   the action       local effects            hosted-connection effects
runtime boundary   the consequence  out-of-Horizon ingress   local actions
```

Why two is the win, not merely two:

- Neither is reconfigurable by the guest — one lives *below* it (kernel/OS), one
  *above* it (runtime semantics).
- The process boundary does not trust the guest's configuration; the runtime
  boundary does not trust the OS's ability to classify traffic. A broken
  assumption in one does not cascade into the other.
- Their union covers strictly more than either alone; where they overlap, the
  coverage is redundant on purpose.

Both are projections of one Horizon — authority is specified once and enforced in
two planes, with no second source of truth to drift:

```text
Horizon
   ├── projected DOWN  -> OS launch policy    (physical / effect ceiling)
   └── projected IN    -> invoke/return check (semantic / admission ceiling)
```

Outside both planes: **egress of data.** A forbidden computation's input (for
example, context embedded in a search query) can still leave over the inference
connection. The control there is registry exclusion — never expose the tool —
plus the inference-channel exposure already accepted for the model call itself.

## Horizon and identity

Authority is separate from continuation-language composition `DECIDED`:

```text
Target capability    authority and dispatch
Invoke / Return      runtime transition form
ordered Operands     continuation-language application
```

`EngramID = hash(ordered Operands)` — identity is composed from the operand list,
not from Horizon. Horizon authorizes an execution of an engram; it is not part of
the engram's identity. An observation is inert data; it gains a role only by its
position in an admitted Invoke, and that admission is where Horizon is enforced.

## Scope and placement

- Horizon is **per-possibility**. Two possibilities addressing the same entity
  face run under different Horizons and therefore different installed authority.
- The stable thing across possibilities is the `EntityAddress`, not the Horizon.
  The active driver binding — process, credentials, transport, route — is
  installed per possibility inside that possibility's Horizon.
- Capabilities name the stable semantic target `(Entity, EntityAddress,
  CapabilityAddress)` and their constraints. They do **not** grant a transport,
  route, process id, or driver generation. Compatible implementation replacement
  does not remint Horizon; an incompatible authority change requires a new
  address.

## Invariants

```text
the kernel is the sole minter of Horizon
the scheduler checks against the installed Horizon on the fast path
Horizon is a property of the minted possibility, not a trusted field on the Invoke
Focus is an exposed subset of Horizon; advertising is not authorizing
both Invoke and Return are checked against Horizon before effects
a returned observation outside Horizon is rejected before entering the Register
the OS process boundary is derived from Horizon and is the enforcement of last resort
guest configuration is advisory and is never the authority boundary
Horizon authorizes an execution; it is not part of engram identity
capabilities name a stable semantic target, not a transport or driver
```

## Open

- `OPEN` The returned-observation type/capability/provenance tag and the
  return-side check that reads it (contract surface in DNA15+).
- `OPEN` Whether Horizon constraints are data, predicates, or references to
  kernel-owned policy.
- `OPEN` Exact representation of a minted Capability and the possibility-table
  entry that installs a Horizon (`{ id, horizon, role, returnTo, state }`).
- `OPEN` When the scheduler must become a possibility-aware *reader* (the
  three-possibility star) versus staying possibility-free for one possibility.
- `OPEN` The precise Horizon → process-launch-policy derivation (which
  Capabilities imply which filesystem/network/exec/credential grants).
- `OPEN` Outbound-data exposure: the return check bounds what comes back, not what
  a forbidden computation's input carries out over the inference connection.

## Related

- [`codex-driver/06-native-tools-and-containment.md`](codex-driver/06-native-tools-and-containment.md) — return-boundary enforcement and the OS containment plan.
- [`continuation-language-operands.md`](continuation-language-operands.md) — operands, identity, and the authority/composition split.
- [`../2026-09-10/entity-address-driver-transport.md`](../2026-09-10/entity-address-driver-transport.md) — CapabilityAddress vs Capability vs Horizon vs Focus.
- [`../2026-09-10/focus-ideas.md`](../2026-09-10/focus-ideas.md) — kernel mints / scheduler checks; the possibility table.
- [`../2026-09-10/codex-ingress-boundary.md`](../2026-09-10/codex-ingress-boundary.md) — Focus → dynamicTools; provider identifiers terminate at ingress.

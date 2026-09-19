# Drivers and the boundary

**Date:** 2026-09-07
**Status:** DESIGN — decided this session. The driver mechanism behind [the main loop](main-loop.md); connects
the [entity/host thread](host-entity-address-horizon/README.md) (addresses, representation) to the runtime.

## A driver is an addressed, bidirectional transducer

A **driver** translates between the runtime's **Move plane** and an **entity's representation plane**, at one
address:

```text
   Move  ⟷  [ driver: normalize ]  ⟷  entity representation
   (runtime side — Moves both ways)     (entity side — per the address's representation contract)
```

The governing rule is **the receiver owns the format**: whoever receives a message dictates the representation
it must arrive in, and the driver translates the sender's form into the receiver's form.

```text
runtime → entity  (inbound to the entity)   render into the ENTITY's representation   ← entity owns this contract
entity  → runtime (inbound to the runtime)  render into a canonical Move              ← runtime owns this contract
```

This is the same principle from [address owns representation](host-entity-address-horizon/address-and-representation.md):
the entity address owns the parser/representation contract for its endpoint; the driver implements it.

## The invariant: all driver egress produces Move(s)

**Everything a driver emits toward the runtime is a Move.** A filesystem read normalizes to a `Ret` move; an
LLM tool call normalizes to a `Call` move; an error is a `Ret` carrying an errno-observation (an error is a
value that returns, not a special channel).

This is not a new axiom — it **follows** from two things already decided: (1) the receiver owns the format, and
(2) the runtime's canonical inbound form is a Move. The runtime receives everything a driver emits, so every
driver's outbound normalization *must* produce Move(s). It is a corollary, which is why it holds so cleanly.

Three scoping notes keep it tight:

```text
· 1-to-N. one driver egress can be MULTIPLE moves (a model doing parallel tool calls emits several Calls). "Move(s)."
· THE OUTBOUND NORMALIZER ASSIGNS THE OPCODE, from what the entity produced: FS return → Ret; LLM tool call → Call;
  LLM final answer → Ret; driver error → Ret with an errno-observation.
· SCOPE IT TO DRIVER OUTPUT. a Materialize (engram not resident) happens BEFORE the driver runs — nothing egressed,
  because nothing executed. that is a resolve-time concern upstream, not driver egress; it does not violate the invariant.
```

Symmetric on the way in: a driver *receives* the `Exec`/`Call` move directed at it and *emits* result Move(s).
Moves both directions on the runtime face; the entity's representation both directions on the entity face.

CC's dispatch result/callback reports the handoff outcome to Scheduler, which owns registers and
continuation state. This internal report is distinct from driver output. Successful dispatch can mean
accepted work; completion arrives later as a driver result Move through authorization.

## The boundary is ingress and egress

The driver boundary **is** the ingress/egress membrane — they are not separate concepts sitting near each other.
Every move going out crosses egress through an addressed driver toward its entity; every move coming in arrived
from one. Because every invocation is **registered at ingress** (identity, address, opcode) before it flows,
egress is pure routing and every return is correlatable back to a known work item — nothing crosses unregistered.
That registration is the `callId ↔ MoveID` correlation generalized.

Two axes at the boundary stay clean (and match [the main loop](main-loop.md)):

```text
ADDRESS  →  WHICH driver   (routing — where this move goes)
OPCODE   →  HOW it renders (the driver transduces a Call vs a Ret vs an Exec differently for its entity)
```

The opcode-switch *here* is rendering (invoke the entity vs. deliver a return to it), not lifecycle bookkeeping —
so it does not conflict with "the scheduler must not switch on opcode."

## Cardinality and placement

```text
Entity        1 ──── *  EntityAddress        an entity has many addresses
EntityAddress 1 ──── 1  Driver               one address, exactly one driver
```

Driver is 1:1 with EntityAddress, which already owns the representation contract — so the driver (the
implementation) and the contract it satisfies are co-located: **the address is the contract ("where + how to
represent"), the driver is the implementation ("how to actually talk").** The driver's *handle* lives in the
CTU (the driver registry); the *entity* it fronts may be local or remote — the driver is the local proxy.

**This collapses `Local map[Target]Driver` and `Out Egress` into one address→driver table.** Local vs. remote
stops being two mechanisms — it is just which kind of driver the address resolves to (a local mechanism driver,
or a remote-entity driver whose transduction includes network transport). Routing is a field on the move plus a
lookup, not a subsystem. This answers the entity/host thread's open "how do Target, Local, and Egress fit."

## The kernel is the root driver

The kernel is a driver — the **root** one, registered first at boot, holding the seed capabilities (mint, load,
persist) it grants downward. Everything the kernel used to be a separate category for — authority, fault
handling, persistence — is reached the same way as any driver: a move to its address. See
[the main loop](main-loop.md#the-kernel-is-the-root-driver) for the fault-seam and authority-invariant details.

## Three mechanisms become one

`Driver` (world-touch), `Egress` (remote), and the MCP/LLM adapter were three things. Under **entity + address +
driver** they are one uniform mechanism — an addressed transducer:

```text
Filesystem   entity + driver + address   (POSIX in, bytes/errno-Ret out)
LLM / Codex  entity + driver + address   (a turn in, Call/Ret moves out — see the inference-substrate doc)
Remote CTU   entity + driver + address   (Egress is just the remote-entity driver)
Kernel       the ROOT driver             (authority/fault/persist reached by address)
```

## Open

```text
· COMPOSE AS A DRIVER. pure computation needs a compute driver, or the CC (not a driver) secretly executes. lean: yes.
· INGRESS/EGRESS FRAME. named as driver-boundary gates here; register reads/writes belong to Scheduler; CC authorizes and dispatches (main-loop).
  reconcile the earlier runtime-frame "ingress" wording in move-register-return.md to this two-axis split.
· DRIVER REGISTRY SHAPE. the address→driver table, registration, and how the root driver admits others (bootstrap).
· REMOTE TRANSPORT. the remote-entity driver's actual wire/transport is unspecified; the abstraction names the seam only.
```

## Related

- [The main loop](main-loop.md) — Scheduler-owned state, stateless CC, and dispatch outcomes.
- [The inference substrate](inference-substrate.md) — the LLM as a driver in detail.
- [Address owns representation](host-entity-address-horizon/address-and-representation.md) — the per-address representation contract.
- [Host, entity, address, and Horizon](host-entity-address-horizon/README.md) — the routing thread this resolves into drivers.

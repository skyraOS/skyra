# Overnight v9 build — boundary crossings

**Date:** 2026-08-23
**Status: THROWAWAY IMPLEMENTATION NOTES.** The trail from making a possibility run as an
event-driven reorder buffer, gather its distributed cone, and unfold a two-iteration loop.

## What runs

```text
ACTIVATE materializes Engram.Body → Moves + Registers
  → DRIVE runs a local seed
  → two ready remote moves issue queued Transfers and PARK in InFlight
  → the issuing stack returns before either worker runs
  → fabric events GREET two narrow children on the worker CTU
  → each child runs, signs, and queues a response
  → each RECOGNIZE verifies and uses the same LAND path as local completion
  → the second landing wakes the join; the root reaches quiescence
  → root-only REMEMBER walks signed Causes through Fetch and persists four deduped segments
```

The separate dispatch trail runs all three local fates: host output is BELIEVED, entity output is
STAMPED+CHECKED, and a kernel-named move enters `Trap` even when a peer entity is registered under
the same spelling. The loop trail runs `increment → SWITCH` twice. SWITCH is an ordinary Move; its
runner appends the next tagged pair only on continue. On exit it appends nothing, the frontier is
empty, the watermark reaches the end, and DRIVE terminates by quiescence.

## Interfaces that did not hold

### `Possibility.Return Contract` cannot identify completion

A contract says what may land, not which register is the fragment's result. With a fan-out/join,
DRIVE could not distinguish an intermediate matching value from completion. `Return` is now a
`RegKey`; its `Register.Contract` retains the type expectation. This is also what lets a tagged loop
map its exit back to the outer tag without special-casing a counter.

### Arena on the crossing type cannot make least-knowledge inexpressible

`*Arena == nil` was only a convention: any sender could populate it. Arena now lives in the CTU's
residency entry, stamped at GREET, and is absent from `Possibility`/`Transfer` entirely. The child is
still the same executable object; only residency metadata stays with the core.

### Synchronous local transport reintroduced the waiting parent

Calling `Ingress.Arrive` from `Fabric.Approach` would recreate v8's recursive
`Drive → Move → Greet → Drive` stack. `Local.Approach` now validates and enqueues. Its throwaway
`Drain` harness advances events later, proving that the parent is represented only by `InFlight`
and the local MoveID ownership index.

### A return contract must be checked before shared LAND, not inside it

LAND is shared by trusted host completion, checked entity completion, and verified cut return. A
contract check inside LAND incorrectly gates host output, contradicting BELIEVED. Entity checks in
`runLocal`; RECOGNIZE checks the destination register after signature verification; LAND only
updates the register, node, watermark, and result.

## Where two documents disagree

`2026-08-22/ctu/possibility.md` says the arena binding must live on every possibility because the CTU
is nearly stateless. `2026-08-23/possibility-residency.md` says an in-flight/non-root possibility must
be structurally unable to carry its parent's arena and that the receiver stamps a local arena. One
exported Go struct cannot satisfy both. The newer residency rule wins: the CTU keeps only residency
metadata—not computation state—and the crossing object has no arena field.

There is a smaller trust-document tension too. `execution-cycle.md` says host results are BELIEVED
and notes that whether their provenance is signed is unresolved, while GATHER requires every cause
to be fetchable and verified. This build signs every retired move, including host moves. The
signature records custody for the cone; it does not turn the host result into an entity-style
contract-checked value.

## A proposal in `claude-ideas.md` that was wrong

Proposal 7, `Arena` as `*Arena`, did not provide the structural guarantee it claimed to signal. The
build rejected it and moved arena into residency. Proposal 2's `Return Contract` also proved too
weak once the graph had more than one result-shaped observation.

The Move/Transfer split survived this run. A resident `Move` is mutated through
Waiting→Ready→Issued→Retired while its request and response legs may sit in the event queue; forcing
both lifetimes into one struct would make the resident node alias transient cargo. `Transfer` remains
the declared name for the custody leg, and the two cargo arms are runtime-validated XORs.

## The concept that had no home: the ENGRAM BODY

`Engram{Name, Contracts}` could not materialize a graph. `Engram.Body []MoveTemplate` now holds the
compiled continuation and `Engram.Return` identifies its exit register. `Materialize(body, tag)`
instantiates runtime Moves and their register file at ACTIVATE. `MoveTemplate` uses the spec's own
static TEMPLATE/POSITION terms; it is a representation, not a new machine mechanism.

## The loop seam this run exposed

The existing `Entity.Run → one Content` interface has no general surface for SWITCH to request graph
growth. The proof keeps SWITCH an ordinary entity-dispatched Move, but its throwaway runner is bound
to the resident possibility and appends the next tagged templates itself. DRIVE contains no switch
branch and termination is genuine quiescence, but this binding is not a durable interface. A later
contract must let an ordinary Move express its one routed emission plus template instantiation
without letting arbitrary entity code mutate causal state. The specs deliberately leave
`SWITCH SURFACE` open; this build found the exact boundary where that open lands.

## New implementation names

```text
residency     CTU-local mapping of a resident possibility to its local arena and return route.
Trapper       the execution-cycle's named TRAP crossing, kept separate from both dispatch blocks.
Drain         throwaway deterministic advancement of Local's queued ingress events.
Pending       throwaway observation used to prove a cut parked before downstream work ran.
MoveTemplate  the spec's static TEMPLATE represented in the engram body.
```

No process type, process table, lifecycle, or control unit was introduced.

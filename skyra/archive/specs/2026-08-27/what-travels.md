# What travels — the Move is portable, the Context is local

**Date:** 2026-08-27
**Status: DECISION, CODE APPLIED.** Working through execution-with-contexts surfaced a reframe: **contexts
do not travel.** The Move is the work; the Context is the local register frame it becomes when it lands.
Consequence: the HORIZON must ride on the wire, not be reconstructed at the destination.

> **UPDATE (2026-09-01):** this doc's PRINCIPLE is the current model and was REINFORCED — contexts don't
> travel; the **Move IS the portable activation / wire format**, and authority (the Horizon) rides on it. The
> mechanics drifted, though: the `Seed` struct here dissolved into `Message` (message-the-wire-unit.md), and
> `Message` itself is now **RETIRED** — the wire unit is the `Move` = {Root, Horizon, TargetFd, Action,
> ReturnFd} (move-and-possibility.md). So read "Seed"/"Message" below as "the Move"; `Route` is gone
> (`ReturnFd` is a `MoveID`); the child's PossibilityID is minted at the target, off the wire.

---

## The execution sketch (validated)

```text
1  a MOVE arrives (crosses the barrier — self-contained: identity + authority + body-ref + inputs + return).
2  the CTU has NO local Context for it → it FAULTS (Materialize; the i-cache/body miss is the read fault).
3  the fault BUILDS the Context LOCALLY — from the move's Seed: load the Body, install the Horizon, seed the
   input registers, set ReturnTo. the Context is BORN here, at this CTU.
4  run. the Context accumulates registers as the body's moves land; at quiescence, Remember writes it out.
```

Step 3 is the tell: **the Context does not arrive — it is CREATED at the destination.** So everything the
destination needs to build it must be ON THE MOVE.

## The three roles — code, activation, frame

The reframe separates three things the old struct blurred:

```text
BODY     the CODE.        content-addressed, immutable, shareable, cacheable. FETCHED (Materialize) on a
                          miss — it does not ride the move; the move REFERENCES it (a BodyID). the "image."
MOVE     the ACTIVATION.  the PORTABLE unit that crosses the wire: identity (MoveID) + authority (Horizon) +
                          body-ref + inputs + return-route (Seed). a reified call / continuation. it TRAVELS.
CONTEXT  the FRAME.       the live register file + result. MUTABLE, per-run, and LOCAL — born at its CTU,
                          never crosses the barrier. like a CPU's registers: they never leave the core.
```

"The move is the executable" (the reframe that clicked) → sharpened: the **Body** is the code; the **Move**
is the traveling activation that points at it; the **Context** is the local frame it runs in. The move is
what's PORTABLE — the Context is what a move BECOMES when it lands.

## Why the horizon MUST travel on the move

Distribution forces it. A cut goes to a **remote** CTU. That CTU **cannot reconstruct the authority**:

```text
· there is no shared kernel to call back to — the hot path is deliberately kernel-free (Verify is offline).
· reconstructing a horizon at the destination = a kernel round-trip per hop = the exact latency the
  offline-verifiable capability model exists to AVOID.
· so authority must be SELF-DESCRIBING and travel WITH the request, signed/unforgeable (the barrier signs).
```

This is capability security 101 (the capability IS the authority, carried and checked, not looked up) and
the actor model (messages travel; actor state is local). **We re-derived both at this seam.** The horizon
is authority; authority rides the move.

## What changed in code (causal.go)

```text
· Move.Seed WAS `*Context` — a category error: shipping a whole Context across the wire. RETYPED to a new
  `Seed` struct: the portable INGREDIENTS (ID, Root, Body, Horizon, seed Registers, ReturnTo) — explicitly
  NOT a Context. Now the type system ENFORCES "contexts don't travel": a Move cannot carry one.
· Horizon rides on the Seed (on the Move). doc updated: "carried ON THE MOVE … a remote CTU cannot
  reconstruct it."
· Context doc: now LOCAL — "born at its CTU (Materialize builds it from a Move's Seed) and NEVER crosses the
  barrier … what a Seed BECOMES when it lands."
· build + vet green; no field accesses broke (nothing dereferenced Move.Seed yet).
```

Seed and Context look similar on purpose — a Context is what a Seed becomes. The differences are the point:
Seed is immutable/portable/inputs-only; Context is mutable/local/accumulates-registers + Result.

## The invariant

**Contexts are local. Moves are portable. Authority rides the move.** The Materialize fault is the
transition point — it turns a traveling Seed into a resident Context. (This is also why there is no
CTU-side loader that pulls a Context from elsewhere: there is nowhere to pull it FROM — it is built here.)

## Open

```text
· WHICH horizon travels — the full workflow horizon vs an attenuated per-hop subset. THE DAY'S OPEN CALL
  (cut/child-horizon attenuation, README). This reframe does not resolve it; it only fixes WHERE it rides
  (the Seed). whatever the subset, it travels there.
· the two authorities on a cut: (a) the capability to MAKE this call (Verify at ingress, drawn from the
  PARENT's horizon) vs (b) the horizon the CHILD RUNS under (in the Seed). same in an in-workflow cut,
  possibly different on a cross-workflow RPC. needs settling with the attenuation call.
· could Context EMBED Seed (Context = Seed + Result + accumulated registers)? tighter, but couples them;
  left as a comb, not done.
· does a Seed's input Registers carry full Register{Contract,Value} or just Values (Contracts come from the
  Body's templates)? lean: values in, contracts from the Body — but deferred to handler/materialize work.
```

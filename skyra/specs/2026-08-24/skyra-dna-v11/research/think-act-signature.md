# THINK / ACT — the co-authored signature

**Date:** 2026-08-24
**Status: DIRECTION — NOT CONFIRMED.** Still being worked; recorded so the thread isn't lost. Nothing
here is settled. "ACT" is a *proposed* name for the execution phase (to separate it explicitly from
"execute" and other terms). Builds on the body/context split in `cpu-execution-model.md`.

---

## The direction

A move's **signature is co-authored across two phases**, by two different owners, at two different
times:

```text
THINK  (propose · BODY)    authors the RUNNABLE half:  TARGET + that it is REACHABLE + the move params.
                           the part that makes the move dispatchable AT ALL — like a linkable
                           declaration. does NOT fix the exact types or the values.
ACT    (execute · CONTEXT) settles the PAYLOAD half:   the actual OBSERVATION passed, PLUS a parent↔child
                           NEGOTIATION of the input-observation type and the return-contract type.
                           late-bound at the crossing.
```

THINK writes the *declaration* ("target T is reachable, here's the arity"); ACT writes the *typed call*
("parent and child agree what goes in and what comes back, and here is the value"). This aligns with
body/context: **THINK produces the BODY, ACT fills the CONTEXT**, and the signature spans both.

---

## The negotiation is a 2×2 — but LOCAL to ACT, not a system quadrant

The type negotiation at a crossing is a genuine 2×2:

```text
                    INPUT (observation type)        RETURN (contract type)
PARENT              what the parent SENDS           what the parent EXPECTS back
CHILD               what the child ACCEPTS          what the child RETURNS
```

Negotiation = reconcile the rows: *parent-sends* must satisfy *child-accepts*; *child-returns* must
satisfy *parent-expects*. This is the quadrant the shape keeps suggesting — but current reading is that
it is **local to each ACT crossing**, not a new top-level system quadrant. Open: is the quadrant the
seen shape bigger than this negotiation, or is it exactly this?

---

## Why types are negotiated at ACT, not fixed at THINK

THINK cannot know the exact types. It proposes *reach target T* — but the precise observation-in and
return-out shapes depend on the actual child, and for a NOVEL target those shapes are **discovered by
`test` at runtime** (the ConnectionContract's Fitness = request/response types). So types are
**late-bound at the call site**, like structural typing or content negotiation (HTTP Accept /
Content-Type, or an ALPN handshake). THINK guarantees *reachable*; ACT settles *typed*.

### How acquire ties in

```text
acquire/test   discovers the CHILD's negotiation-side (its input/return shapes) → registers it in the
               ConnectionContract.
THINK          proposes reaching the target — reachable BECAUSE it was acquired.
ACT            parent↔child negotiate using {registered ConnectionContract = child's side} vs
               {parent's payload + expectation}.
```

So the ConnectionContract is the child's half of the negotiation quadrant, pre-discovered by test; the
parent brings the other half at ACT.

---

## What this would reshape

`Move.Contract` today is a single, THINK-ish field (a return expectation). Under this direction, **the
contract is not solely THINK's — it is negotiated at ACT.**

```text
THINK sets on the Move    Target, reachable, structural params (Reads/Writes positions).   → BODY half
ACT settles at crossing   input-observation type, return-contract type (negotiated), payload value. → CONTEXT half
```

---

## Candidate statement (NOT confirmed)

> A move's signature has two halves. THINK authors the *reachability* half into the BODY — target,
> runnable-at-all, params. ACT settles the *payload* half into the CONTEXT — the observation value plus
> a parent↔child negotiation of the input and return types. The negotiation is a 2×2 (parent/child ×
> in/return) local to each ACT crossing; it is not a system-level quadrant.

---
---

## Open / leaning (CLAUDE'S — not decided)

```text
THINK: TYPE HINT or PURE REACHABILITY?   lean: THINK proposes a target + PROVISIONAL types (enough to
                                         pick the child); ACT ratifies or narrows them. not "no types
                                         until ACT," but "provisional at THINK, settled at ACT."
WHERE THE NEGOTIATION RESULT LIVES        lean: in the CONTEXT (the resolved contract for that crossing),
                                         not back in the BODY — the body stays the shareable static
                                         skeleton; negotiated types are per-run.
IS THE QUADRANT BIGGER THAN NEGOTIATION?  open — the user is still weighing whether the shape they see
                                         is exactly the ACT negotiation 2×2 or something larger.
```

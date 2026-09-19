# SYS params vs payload — the reframe that IS the seam

**Date:** 2026-09-02
**Status: BRAINSTORM / a clarifying framing (lean: adopt as the vocabulary for seam A).**

---

## The idea (yours)

> The sys parameters are the params the kernel needs to find the executor. The payload + observation params
> are the data the LLM or human reasons over and emits computation over.

Two categorically different kinds of thing ride (or should ride) with a move:

```text
SYS PARAMS   what the KERNEL/BARRIER need to FIND + AUTHORIZE the executor. routing + authority.
PAYLOAD      the DATA the computation reasons over. what an LLM/human emits code to operate on.
```

## Printed against today's wire Move (causal.go)

```text
SYS PARAMS  (present on the Move today)
  Root      PossibilityID   the workflow scope (horizon's Root)
  Horizon   Horizon         the authority — Verify OFFLINE against this
  TargetFd  Target ────┬─ Entity  EntityID   WHO   — a userspace identity
                       ├─ Host    HostID     WHERE — resolves the volatile Address via host(entity)
                       └─ Name    Name       WHICH — the endpoint name
  Action    Name            the method to invoke; checked as (Target, Action) against the horizon
  ReturnFd  *MoveID         where the value lands (destination is reply-to-sender, barrier-derived)

PAYLOAD / DATA
  ⟶  Observation { Content{Bytes}, Segment{By Name} }
  ⟶  ...is NOWHERE on the Move, Body, or Context as an operand.
```

## Why this is the whole seam

**The Move carries a complete SYS half and a totally empty DATA half.** That asymmetry IS seam A:

- `Cold.Handle(from, request Observation)` receives the payload.
- THINK reads it to SELECT a Body — and then it vanishes. No field carries that Observation forward to the
  driver that needs it (gap #9: "request data disappears before execution").
- `Driver.Perform(action, in []Observation)`'s `in` was defined as "children's returns" only — so for a leaf
  (no children) `in` is empty. A leaf that needs data has no source for it.

So the operand model is precisely: **give the payload half a home the SYS half already has.** Naming it this
way is cleaner than operand-model.md's "literals vs children" framing — it says the operand is categorically
the Observation, distinct from Target/Action/Horizon, full stop.

## Consequence carried into the next docs

Note what the SYS/payload split does NOT by itself decide: WHERE the payload half lives (on the node? bound at
materialize? produced by a child?) and WHETHER a payload can travel FORWARD or only UP. Those are
`leaf-and-directions.md` and `args-decision.md`.

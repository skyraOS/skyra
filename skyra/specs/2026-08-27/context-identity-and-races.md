# Context identity and races — PossibilityID is the membership key

**Date:** 2026-08-27
**Status: DECISION (leaning), CODE-DOC APPLIED.** Follows what-travels.md. Concern: moves for the SAME
context arrive separately at a remote machine and race on who creates the context. Resolution: the grouping
identifier already exists and flies on the wire — `PossibilityID` — and the race is OS demand paging
(dedup concurrent faults on one page). Concurrency CODE deferred; the model + the key are settled here.

---

## The concern

```text
a bunch of moves cross to another machine SEPARATELY but all belong to the SAME context. if "the kernel
makes the context" (Materialize builds it), then N arriving moves → N faults → N attempts to create the
same context. RACE / double-create. and we CANNOT group them by DEPENDENCY (a precondition-not-arrived
gate) — transitive deps would pull the WHOLE TREE into one context, and the point of a context is to be a
SMALL, REUSABLE component. so we need an identifier that says "these go together" and flies on the move.
```

## Resolution 1 — the identifier exists: PossibilityID

It already flies on every move (`MoveID.Possibility`) and on a spawning move (`Seed.ID`). No new field —
name its role. It is the MIDDLE of a three-way split, and the bug is using one tag for another's job:

```text
Root         (workflow)   the whole tree; the HORIZON's scope.          too COARSE for membership.
Possibility  (context)    ONE run of one Body; "these go together".     ← the grain. SMALL + REUSABLE.
Dependency   (Reads/Writes) firing ORDER (Waiting→Ready).               too TRANSITIVE for membership.

MEMBERSHIP = Possibility   ·   ORDERING = dependency   ·   SCOPE = Root
```

Both rejections in the concern are correct: **dependency can't be membership** (collapses the tree);
**Root can't be membership** (too coarse). `PossibilityID` is the per-run grain — which is exactly what
makes a context small and reusable. Membership is groupable across machines by comparing PossibilityIDs,
with NO dependency reading.

## Resolution 2 — the race IS demand paging

N moves faulting on one not-yet-resident context = N threads faulting on one not-present page. Page it in
ONCE; the rest find it present/in-flight and wait. Three parts:

```text
1  CREATE IS IDEMPOTENT ON PossibilityID.  materialize-once, keyed on the CONTEXT identity (not move.ID —
   that key is for per-move retry). two moves for X → one context X.

2  RESIDENCY LATCH (tri-state) in the Cache:  Absent → Materializing → Resident.  the FIRST faulter flips X
   to Materializing and faults; concurrent faulters SEE Materializing and PARK (do NOT re-fault); on return
   the slot flips Resident and parked moves wake. = a page-table entry with a fault-in-progress bit.

3  SINGLE-WRITER LAND per PossibilityID.  once resident, concurrent responses into X's registers serialize
   (write → pop → flip): one writer per context at a time. (the concurrency area in dispatch-and-decode F.)
```

### Two DIFFERENT waits — keep them apart

```text
CONTEXT-NOT-RESIDENT   park on the residency latch, keyed on PossibilityID   (creation not done)
INPUTS-NOT-LANDED      park in Waiting, dataflow, keyed on Reads             (dependency not satisfied)
```

A move can be blocked on either; both are park-and-wake, different keys. Conflating them was the trap the
concern spotted — dependency is the SECOND wait, never the grouping key.

## The seed worth holding — content-addressed PossibilityID

Make `PossibilityID` a CONTENT ADDRESS of `(BodyID + seed inputs)` — as `BodyID` already is for code. Then:

```text
· "these go together"      = same PossibilityID → AUTOMATIC, no coordination.
· idempotent create        = FREE — same inputs = same id = the dedup IS content-addressing (race-safe).
· reusable in later flows   = MEMOIZATION — same (body, inputs) resolves to the same id, so a later workflow
                             FINDS the cached context/result instead of recomputing. exactly what a context
                             is FOR.
```

Collapses membership + idempotency + reuse into ONE mechanism. Cost: seed inputs known at id-time +
determinism assumptions. STRONG SEED, not a decision — settle when the reuse/memoization story is drawn.

## What changed in code (causal.go, build+vet green)

```text
· PossibilityID — doc added naming it the CONTEXT-MEMBERSHIP key (the three-tag split; idempotent-create;
  the content-address comb). no new field.
· Cache — doc added: it is ALSO the residency latch (Absent→Materializing→Resident) that dedups concurrent
  faults. concrete Materializing marker + single-writer discipline DEFERRED with the concurrency model.
```

## Open

```text
· the concurrency CODE: the Materializing marker's shape (a sentinel entry? a separate in-flight set?), and
  the single-writer-per-PossibilityID LAND discipline. deferred with the Core concurrency model (Core has
  the transient-state TODO; dispatch-and-decode F).
· content-addressed PossibilityID (the seed above) — decide with the memoization/reuse story.
· eviction/reactivation: a context Remembered then re-Activated should keep its PossibilityID (stable
  identity, paged out and back) — confirm Remember/Materialize preserve it (they should; it's the disk key).
```

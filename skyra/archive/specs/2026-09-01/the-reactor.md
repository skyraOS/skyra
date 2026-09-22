# The reactor — the engine is two halves, not four stages

**Date:** 2026-09-01 (extracted to its own doc 2026-09-02)
**Status: DECISION, CODE APPLIED.** The execution engine is a **two-part reactor** — an ISSUE loop and a LAND
handler — not the four-step `pull → dispatch → land → flip` loop it started as. Two collapses got us here.
This is the authoritative statement; `register-lifecycle-and-transition.md` (the three-table STATE model)
points here for the ENGINE that steps it.

---

## The starting point (what collapsed)

The engine began as one core's four-step pass:

```text
1. PULL READY   take a ready move from Work.
2. DISPATCH     run it (local Driver.Perform, or remote Egress); mark it in-flight.
3. LAND         the result returns; write it into the register.
4. FLIP         wake the moves whose inputs are now all present.
```

Two observations collapse this to two halves.

## Collapse 1 — LOCAL and REMOTE do not differ in the engine

Dispatch is **uniformly asynchronous**: it issues and FIRE-AND-FORGETS, and the result **always** returns as
a `Landing` — whether from a local `Driver` next door or a remote CTU across the fabric. Locality is a
**transport** concern, BELOW the engine (does the Landing cross the barrier?). The reactor never knows.

```text
· correctness, not just tidiness: a SYNCHRONOUS local dispatch STALLS the core for the driver's whole
  duration. uniform-async keeps cores issuing other ready moves while one is outstanding.
· this is the same principle as "the cut does not exist" (move-and-possibility.md) carried into the engine:
  local vs remote is a disposition, not a kind of run.
```

Consequence: steps 2 and 3 stop being one core's inline sequence. DISPATCH never waits for its own result;
LAND is driven by INCOMING landings, from either side, handled identically.

## Collapse 2 — PULL and DISPATCH are one atomic step

A move in `Work` is ready; **taking it OUT of Work IS issuing it.** The atomic remove is simultaneously the
single-writer CLAIM (no second core fires the same move) and the dispatch. There is no "pulled but not yet
dispatched" state to occupy — nothing happens between selection and issue.

```text
(choosing WHICH ready move to pull first is scheduling POLICY — layer it later; it is not a separate state.)
```

## The reactor — two decoupled halves

```text
ISSUE  (the loop)      while Work non-empty:
                         claim-and-issue a ready move (atomic remove from Work → in-flight). FIRE-AND-FORGET —
                         never blocks on a result. local ⇒ Driver.Perform (async); remote ⇒ Egress.

LAND   (event-driven)  on ANY incoming Landing (local completion OR remote Ingress), for its fd:
                         1. WRITE Registers[Fd.Writes]   (the value; = LANDED)
                         2. POP the in-flight entry
                         3. FLIP dependents — each move reading the new value whose reads are now all present
                                              → enqueue to Work (which the ISSUE loop drains)
```

The two halves are connected ONLY through the tables: ISSUE drains `Work`; landings refill it. **Dispatch
never blocks on a result.** This is the classic reactor / reservation-station shape: a work queue drained by
fire-and-forget dispatch, refilled by completions.

```text
QUIESCENCE = Work empty AND in-flight empty → the root's value is in Registers → Remember → the Context retires.
```

**Two POPs, different tables — do not conflate:** LAND pops the *in-flight* entry (on arrival); CONSUME frees
the *value slot* in `Registers` (later, when the parent fires and reads its inputs).

## What this supersedes

```text
· dispatch-and-decode.md's four-step, `Disposition = Local | Cut | Trap` framing PREDATES this. the
  decode-dispatch CONCEPT survives (something still classifies a move's transport), but "park in Issued" and
  the local-vs-cut branch as ENGINE stages are gone: local and remote are the same to the reactor; there is
  no `cut` edge; state is table membership, not a bucketed WorkList.
· the "four-step loop" language anywhere else — it is two halves now.
```

## In code (causal.go / ctu.go)

```text
· Core.Transition   the ISSUE pass — claim-and-issue ready moves, fire-and-forget. (currently a stub.)
· CTU.Land(Landing) the LAND half's door — both local completion and remote Ingress arrive here. (pre-added.)
· CTU.Work          the READY table (map[MoveID]Move).
· CTU.InFlight      the ISSUED table (pre-added) — dispatched-and-pending; a Landing pops it.
· Context.Registers the LANDED table (map[RegKey]Observation) — the value slots.
```

## Open

```text
· CONCURRENCY. with N cores the claim-and-issue (remove from Work + add to in-flight) and the LAND
  (write + pop + flip) must be atomic against each other — the CAS/linearization discipline (Continuity).
  quiescence must be linearized with LAND so a context is not retired while a Landing is in flight.
· THE ISSUE↔LAND COUPLING IS THE FLIP INDEX (seam B). LAND step 3 needs "the moves that read register R";
  no reverse index exists yet. this is the ONLY coupling between the two halves, and it must be synchronized
  with both. resolving B resolves how the reactor's halves actually talk.
```

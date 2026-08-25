# Overnight work order — v9

**Date:** 2026-08-23
**To:** the implementer.
**Status:** THROWAWAY build. Move fast, don't polish. The value of this run is learning **where the
boundaries cross** in the RUNNING model, not shipping something clean.

---

## The job

Fill the v9 bodies and wire the packages so a possibility RUNS as a reorder buffer: an event-triggered
`Drive` loop over a graph of moves, ingress that lands-and-drives, a cut that ships a narrow child, and
a REMEMBER that gathers the signed cone home. Then push one step past the DAG: **a two-iteration loop
that unfolds by tag and terminates.**

```
Interfaces to implement:  skyra/src/skyra/skyra-dna-v9/   (start at its README.md)
```

## The authority

```
skyra/specs/2026-08-23/   (start at its README.md — reading order: residency ▸ live ▸ graph ▸ execution-cycle)
skyra/specs/2026-08-22/   and  2026-08-18/   (the foundation it stands on)
```

Read `2026-08-23/execution-cycle.md` closely — it is the engine, and it carries inline Go sketches of
`drive`, `land`, the thin ports, `runLocal` (the two-blocks-plus-trap dispatch), and the loop's SWITCH.
**If your code and the docs disagree, the docs win.**

## What is settled, and what is yours to change

- The **specs** are settled design. Honor them.
- The **interfaces** are a contract to implement — but their *shape* is a set of proposals.
  `claude-ideas.md` lists every one and which are least certain. **Reject or reshape any of them.** If
  you do, say why.
- `interface-rationale.md` explains every type and lists the seams already known.

## The target trail — what to prove, in order

```
1  DAG RUNS       one possibility, several moves, fan-out/join across two CTUs. Ports land-and-drive;
                  a placed cut PARKS in InFlight; Recognize lands it and re-drives. No synchronous
                  recursion — if a port blocks on a downstream leg, you regressed (rationale seam #1).
2  DISPATCH       runLocal branches on the resource block: host BELIEVED, entity STAMPED+CHECKED,
                  kernel TRAP. Prove a kernel-named move traps rather than dispatching as a peer.
3  GATHER         root-only REMEMBER walks each retired move's Provenance.Causes across the cut via
                  fabric.Fetch, dedup'd by content-address, and persists the cone.
4  THE LOOP       a two-iteration loop: a SWITCH move (an ordinary Move) routes Iter+1 twice, then
                  exit. Registers key by (Name, Tag); watch the tags advance and the frontier drain.
                  Termination must be QUIESCENCE, not a counter you special-case.
```

## Report back — this is the actual deliverable

Leave a trail of **where boundaries crossed** (a `boundary-crossings.md`, as v8 did):

- an interface that didn't hold and had to be reshaped (Move-vs-Transfer is the likely one),
- a place two docs disagreed and you had to pick,
- a proposal in `claude-ideas.md` that was wrong,
- a concept that had no home until you gave it one (the ENGRAM BODY is the flagged gap — how the Move
  graph materializes at Activate is undecided; if you invent a form, say so).

## Two hard lines — everything else is yours

- **Terms come from the docs.** If you must name something new, say so and why.
- **No "process" concept.** The docs defer it on purpose; do not smuggle OS process semantics in. And
  **no control unit beside DRIVE** — the SWITCH is a Move, not a scheduler branch (execution-cycle.md §6).

Build it.

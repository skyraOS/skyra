# skyra-dna-v12 — the COLD path

**Date:** 2026-08-27
**Status: PROPOSAL (to be combed).** v11 proved the HOT path (the CTU executes a Body as a Context,
kernel-free, verified offline). v12 turns to the other half: **how a request BECOMES a Body + a
kernel-created horizon + a held capability, and how the engram disk persists it.**

---

## What carries forward (proven, do not re-litigate)

```text
causal/   the boundary vocabulary — Body/Context split, Move (wire+work unit), WorkList, Route,
          Horizon=[]Capability. carried from v11 with the 2026-08-27 folds applied.
ctu/      the hot path — Transfer (door) · Run (park/wake loop) · Core.Transition. carried as the SEAM
          surface; its internals were proven in v11 (ctu_test green). NOT the comb focus.
```

## What is the comb focus (new / undesigned surface)

```text
kernel/   the COLD path made concrete: Cold.Handle decomposed into COMPILE (think) → BOUND (create the
          workflow horizon — kernel authority, 08-27) → MINT (the held ticket) → materialize+run →
          REMEMBER. plus the barrier custody + socket table, and Materialize's INTERNAL cascade.
think/    COMPILE (request → Body), and the acquire/compose children.
engram/   the DISK tier — the un-combed area from v11's memory-hierarchy doc. content-addressed store,
          what REMEMBER writes, ACTIVATE (page-in), eviction. FIRST SHAPE for combing.
boot/     the thin bootloader — seed kernel, start CTU, handshake, EXIT. main IS this.
```

## The 2026-08-27 folds baked in

```text
· HORIZON is the total permission set of a ROOT / WORKFLOW, CREATED BY THE KERNEL (Bound). not per-node,
  not CTU-narrowed. the CTU never constructs a horizon — it verifies against the one it is handed.
· ATTENTION is per-node, THINK-planned; its runtime-active subset is Work[Issued] (no stored field).
· NO per-cut narrowing this phase (a cut carries the ticket + request; the workflow horizon is the root's).
```

## The through-line

```text
COLD   a NEW request → the deepest fault → THINK compiles a Body, the kernel BOUNDs a horizon + MINTs a
       ticket, it runs once, REMEMBER writes the engram. returns (held capability, result).
WARM   a REMEMBERED body → ACTIVATE from disk (page-in) → materialize → run.
HOT    the held capability presented → verified OFFLINE → run. (v11, proven.)
```

The comb question: **do these three tiers share ONE resolve path (cache → disk → compile) with the cold
path as its deepest fault — and are Bound/Mint/Remember the right three kernel verbs around it?**

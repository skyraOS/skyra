# CTU surface for the Starlark frontend — what THINK exposes to propose a graph

**Date:** 2026-08-24
**Status: SPECULATIVE RESEARCH — not a decision.** Companion to
`starlark-computation-and-dataflow.md` (Role 2). Question: *which CTU vocabulary does the
skyra-starlark package expose so THINK can author a continuation graph — and which must it
never expose?* Claude proposes; human confirms.

---

## The governing principle

The frontend exposes **inert constructors** — functions that return a *description* of a graph
node, never the effect it names. It exposes **no execution method and no authority method.**

```text
EXPOSED       INERT CONSTRUCTORS — name every NODE KIND of the continuation graph. read(cap,x)
              returns a Move DESCRIPTION, never bytes. Evaluating the module yields an unpaved graph.
NEVER EXPOSED EXECUTION (Drive, dispatch, Land, Park, Recognize, Interpret, Schedule) and
              AUTHORITY (ask_permission, mint, admit, Bound/attenuate). THINK proposes; the kernel paves.
```

This is the same wall as everywhere else: THINK is read-only and holds no `ask_permission`, so
the *language it authors in* must be unable to cross or to originate authority. If one builtin
ever executes a crossing "for convenience," the boundary is gone. That invariant — **a
constructor cannot execute the move it names** — is the thing to prove first.

---

## The exposed surface

### 1 · References (name, never bind)

```text
cap(name)          → CapabilityReference. names an admitted-OR-proposed capability. authority-free:
                     it is a name (+ at most a novelty verdict), resolved to a live binding by the
                     KERNEL at paving. a novel name triggers acquire/ask_permission DOWNSTREAM, not here.
reg(name)          → RegisterRef. usually implicit (a move's .result IS a register); explicit only for
                     loop-carried state and joins.
```

### 2 · Crossings — value-producing Moves (inert descriptions, one per crossing kind)

```text
read(cap, selector)   → Move   READ    capability crossing  → observation register
request(cap, args)    → Move   REQUEST entity/endpoint contact (a cut carrying a narrow child) → obs
now()                 → Move   NOW     time crossing        → observation register
random(spec)          → Move   RANDOM  entropy crossing     → observation register
execute(cap, args)    → Move   EXECUTE process crossing     → observation register
```

Each returns a Move whose `.result` is the register it fills. THINK *names* the capability; the
crossing is performed later by the executor. These are the only world-touching node kinds, and
here they are pure descriptions.

### 3 · Computation — the pure leaf (Role 1)

```text
compute(pure_fn, *inputs) → Move   COMPUTE. pure_fn names a PAVED pure Starlark module/function;
                                   inputs are register refs. reads only registers, crosses nothing.
                                   result register is DERIVED (digest + input cone), not observed.
```

### 4 · Control / structure

```text
branch(cond, then, else) → Move    a gate on a position. cond is a register ref (usually a COMPUTE bool).
loop(state, condition, body) → Move  the CONTINUATION loop (CTU-owned, tagged). crosses the world each
                                   iteration; parks; may wake THINK. NOT a Starlark for-loop.
group(*moves)   → Group            independent moves — fan-out, parallel-eligible.
sequence(*moves) → Group           ordered composition where EFFECT order matters beyond dataflow deps.
```

### 5 · Distributed / causal

```text
gather(*refs)  → Move     REMEMBER — assemble the signed cone by content-address → a register.
return_(ref)   → Return   the exit register / Result of the possibility.
```

`cut` is deliberately **not** a separate constructor — `request` is the high-level contact; whether
it runs local or ships a cut is the paver/executor's placement decision, not THINK's.

### 6 · Wrapper

```text
program(root)  → Program   wraps the authored graph into the top-level unpaved Possibility. This is the
                           value the module evaluates to and hands to kernel validation/paving.
```

---

## What is deliberately NOT exposed (the wall)

```text
EXECUTION      Drive / Land / Park / Recognize / Interpret / Schedule — the engine is the executor's.
AUTHORITY      ask_permission / mint / admit / Bound — THINK cannot bake authority into a proposal.
SUSPENSION     yield / await / CTU callback — the CTU owns waiting, park, NotYet, retry.
EFFECTS        real read-bytes / socket / clock / random-source / process — crossings return DESCRIPTIONS.
RUNTIME LOOPS  Starlark while / recursion — disabled; the continuation loop is loop(), owned by the CTU.
```

A Starlark `for` may still expand *repeated graph structure* at authoring time when its collection is
known — but it must never impersonate a runtime CTU loop over values discovered later.

---

## Mapping to the CTU model

```text
CONSTRUCTOR        CTU NODE                         OUR MODEL
read/request/…     a value-producing Move           the crossing Moves; result lands in a register
request            a cut → narrow child possibility  possibility-graph §8; paver decides local vs cut
compute            a COMPUTE Move (pure)             new trust tier: REPRODUCIBLE (digest + input cone)
branch             a gated position                 conditions/gates seam
loop               a tagged continuation loop        the SWITCH/loop-by-tag machinery (v9)
gather             REMEMBER-as-gather                the signed cone, root-only
return_            the Return register               live-possibility Return/Result
cap                a capability reference (unbound)  named; kernel binds at paving; novel → acquire
program            the unpaved Possibility           the ENGRAM BODY the kernel materializes
```

Note the clean split for the v11 question (*where think and execute live in the CTU*): **the
frontend surface is THINK-side** (inert authoring), **the crossings + the continuation loop are
execute-side**, and **COMPUTE is executor-side but pure**. THINK authors → kernel paves → executor
runs. The surface here never touches the execute side.

---

## Why this is safe to build on (addresses the "someone else's work" nerve)

```text
SHALLOW DEPENDENCY   Starlark is only a PURE EVALUATOR that builds OUR data structures. the canonical
                     artifact is the continuation graph / IR — ours. Starlark is a replaceable frontend.
SMALL BLAST RADIUS   the package exposes INERT constructors + PURE helpers only. no causal semantics
                     depend on Starlark; if it broke or we swapped it, the graph IR is unaffected.
HERMETIC BY IDENTITY  the reason to use it at all — "the external world is absent." we ADD nothing that
                     crosses; we only expose inert constructors. keep it unmodified; do not fork.
```

So "a skyra-starlark package" is fine **because** it is inert-only. The separation is: crossings and
authority live OUTSIDE Starlark, forever.

---

## Open questions (carry forward)

```text
CROSSING GRANULARITY   five constructors (read/request/now/random/execute) or one cross(kind, cap, args)?
                       five names the taxonomy; one keeps the surface tiny. test which reads better.
CAP AUTHORITY-FREEDOM   confirm cap(name) can carry a novelty verdict without carrying a binding —
                       the proposal must stay authority-free even when it names a novel capability.
REGISTER TYPING        how loop-carried state, child returns, and joins are typed WITHOUT moving
                       validation into Starlark. the kernel validates; the frontend only describes.
DERIVED vs OBSERVED     the frontend must mark a COMPUTE result DERIVED and a crossing result OBSERVED,
                       automatically by constructor kind — preserve the world/derivation distinction.
BRANCH / LOOP SHAPE     does the CTU represent the loop as a cycle, a tagged unfolding, or a
                       possibility-producing node? Starlark must not decide the runtime mechanism.
EXECUTE-PROOF          the minimal test: attempt a crossing from inside a constructor; it must be
                       INEXPRESSIBLE. this is the load-bearing invariant.
COMPOSE FIT            confirm this surface is exactly what THINK's `compose` child emits — the frontend
                       IS compose's authoring surface, not a separate thing.
```

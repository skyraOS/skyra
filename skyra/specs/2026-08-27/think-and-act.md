# THINK and ACT — the roles, corrected (ACT is the compiler)

**Date:** 2026-08-27
**Status: DECISION / CORRECTION.** Prior framing (v11 think-act docs, and the v12 `think.go`/`kernel.go`
proposal) put COMPILATION on THINK. That is backwards. This fixes the THINK/ACT division and follows it
through to who a contract-return failure wakes.

---

## The correction

```text
THINK   TREE STRUCTURE + METHOD NAMES.  the ahead-of-time PLAN — the possibility tree (the frame shape,
        the children) and which actions/targets to call. produces a PLAN (tree + names), an IR. NEVER
        touches bodies.
ACT     COMPILES THE BODIES + owns OBSERVATION TYPES / CONTRACTS.  takes THINK's plan and compiles it into
        a runnable Body, settling the payload types. runtime-invokable — it can be re-woken to recompile.
```

The pipeline is two stages, not one:

```text
request → THINK (plan: tree + method names) → ACT (compile: Body + observation types) → run
          ^ the architect                     ^ the compiler
```

The Body is **ACT's** compiled artifact, seeded by THINK's plan. (This reconciles think-act-signature's
"THINK writes the runnable half" — that half is the *skeleton*, tree + names — with "ACT compiles the
body": ACT assembles the runnable Body and settles the types.)

## ACT is a JIT

Because ACT compiles bodies **and can be re-woken at runtime**, ACT is a **JIT compiler**. A contract-return
failure is a **type-guard failure in JIT'd code**: the compiled body assumed an observation type, the
return violated it, so the machine **traps and wakes the JIT to deopt + recompile** — V8 deoptimizing on a
failed guard. It wakes **ACT**, never **THINK**. The plan doesn't change; the compilation does.

## Who owns a contract-return failure

```text
CTU owns OUTCOMES, not SEMANTICS.  the CTU sees a BINARY outcome (contract satisfied? y/n) and, on failure,
   VECTORS to the handler by the failure KIND — a mechanical classification, like an exception-vector
   number (#PF vs #DE). "return-contract failed" is a vector that points at ACT. the CTU never interprets
   the meaning; it routes the bit.
```

The three contracts, each with its owner:

```text
SYSTEM contract   entity · host · horizon      MECHANICAL   the CTU/kernel decide it   → NACK/Fault (control plane)
ENGRAM contract   observation types            ACT's        CTU traps → wakes ACT       → JIT deopt   (data plane)
THINK contract    tree · method names          THINK's      a DIFFERENT vector — re-PLAN, not re-compile
```

The contract-return failure is the **ENGRAM contract** (observation types) — ACT's, because ACT compiled
the engram and settled its types.

## The type consequence

```text
Outcome       the CTU's bit — did it fault/fail?              (mechanical, CTU-owned, stays binary)
Observation   the error SIGNATURE — what the failure MEANS     (semantic, ACT-produced, context-emitted)
```

Do NOT grow `Outcome` with semantic error codes. The meaning is DATA — an `Observation` carrying an error
signature, emitted by the context, landing in the parent's frame where its base case reacts. The CTU
routes the bit and vectors to ACT; **ACT owns the meaning and the recompile.**

## What this corrects in the build

```text
· v12 think.go — "Compile(request) → Body" framed as THINK compiling. WRONG: THINK produces the PLAN
  (tree + method names); ACT compiles the Body. split the two stages.
· v12 kernel.go — the cold path's "COMPILE (think)" step is ACT, not THINK. the cold path is
  THINK (plan) → ACT (compile) → Bound → Mint → run → Remember.
· v11 research/think-act-paths.md + think-act-signature.md — reconcile: THINK = tree + names (the
  skeleton/IR); ACT = the (JIT) compiler that settles observation types.
```

## Open

```text
· how THINK's plan (tree + names) is represented as it hands off to ACT (the IR between the two stages).
· is ACT a base-case LLM like THINK, and where does the JIT run — hosted by the CTU (woken on a guard
  fault), or a kernel-side service? (leans: woken like a fault handler; the CTU vectors to it.)
· the THINK-contract vector (structure/naming failure) — its own failure path, distinct from the ACT one.
```

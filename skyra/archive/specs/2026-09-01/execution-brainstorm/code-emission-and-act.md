# Code emission and the return — "two things, not one"

**Date:** 2026-09-02
**Status: BRAINSTORM. A SEPARATE seam from the operand model — captured here so it is not fused into it.
Lean: world-touching emitted code is ACT (new forward moves), not a return value.**

---

## The idea (yours)

> We need to find where the LLM can emit code and be the steward of that code. I think that lives in the
> RETURN — an LLM emitting code within its return statement should be the contractual observation it returns.
> But I'm worried I'm shoving too much complexity into one corner. Code that touches the world DOES need to be
> validated by its parent. But what's also important is what the code DOES — the actual return value. Two
> things, not one. Or instead of forward/return, we derive a "Leaf" opcode.

## The key distinction: the "two things" split cleanly onto two seams

The two-ness is the tell. Split the returned code by whether it INVOKES:

```text
return is PURE DATA (a summary, a value, inert code-AS-TEXT)   →  a clean LEAF return. one thing: a VALUE.
return is code that INVOKES (Target, Action) — touches world   →  NOT a leaf at all.
```

The moment returned code *invokes a capability*, it has authored **forward edges** — new moves, new children.
By the leaf definition (leaf-and-directions.md: "orchestrates no dependencies") it has stopped being a leaf.
It is the compiler emitting moves — that is **ACT / the JIT** (think-and-act.md: "ACT compiles the bodies").

So the "two things" you named are two DIFFERENT seams:

```text
THE VALUE the code yields          →  the operand/return world (this brainstorm's main thread).
THE CODE-THAT-INVOKES              →  ACT: authorized as NEW MOVES against the horizon (a different seam).
```

The parent-validation you want ("code that touches the world must be validated") is **already** the MMU check
at ISSUE (`register-lifecycle-and-transition.md` seam I / `specs/2026-09-01` DTLB): every emitted move is
Interpret+Verify'd against the run's Horizon, per move. Emitting world-touching code = a **deopt/recompile
re-entry** into ACT, whose moves face that check — NOT a value smuggled through a register.

## Why this must stay OUT of the operand corner (your worry, confirmed)

Putting world-touching code in the return slot fuses the JIT into the register file — exactly the "too much in
one corner" you flagged. Keep them apart:

```text
· the RETURN carries OBSERVATIONS (values). it MAY carry inert code-as-data (text the parent stores/passes on).
· code that RUNS and touches the world is NEW FORWARD MOVES — checked where all moves are checked (the horizon).
```

## The "Leaf opcode" question

You probably do NOT need a third DIRECTION — two directions (forward/back) is right and leaf-as-producer
reconfirmed it. What you are reaching for is a leaf **KIND**, and it already has a home: the crossing taxonomy
carried over from v10 (`carry-over-from-08-24.md`):

```text
READ · REQUEST · NOW · RANDOM · EXECUTE · COMPUTE       (Starlark = the COMPUTE leaf; the continuation
                                                          language IS dataflow.)
```

That taxonomy is where "this leaf touches the world → validate + do-not-memoize" vs "this leaf just returns a
value" lives — on the DRIVER/crossing classification, not as a new opcode and not in the operand slot. Same
corner-hygiene, correct corner. It also feeds seam E: NOW/RANDOM/world-touching leaves are non-memoizable, so
they salt or disable the PossibilityID reuse (args-decision.md open task).

## Net for the operand model

```text
IMPACTS operand model:      nothing new here — this thread is deliberately kept OUT of it.
BELONGS elsewhere:          ACT (emitted invoking code = new moves) + the crossing taxonomy (leaf KIND).
the one carry-back:         a leaf's return is a VALUE (possibly inert code-as-data); it is never itself the
                            act of touching the world. that keeps the operand/return slot single-purpose.
```

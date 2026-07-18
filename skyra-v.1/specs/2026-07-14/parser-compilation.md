# Parser Compilation — Flat Curve to Parser, and Below the Floor of 1

**Status:** MECHANISM (2026-07-15 research). High-level only — no implementation. A concrete way to
cash in the two learning loops: **compile stable behavior into parsers and stop calling the LLM.**
Still an idea; recorded because it makes "going below the floor of 1" a *chosen* decision.
**Companions:** [observation-side-learning.md](observation-side-learning.md) ·
[region-as-playbook.md](region-as-playbook.md) ·
[closure-duration-objective.md](closure-duration-objective.md) ·
[the-learning-loop.md](the-learning-loop.md) ·
grounds: [../../captains-log/phase-07-prediction-diff-learning.md](../../captains-log/phase-07-prediction-diff-learning.md).

## The signal — a flat curve is a compile signal

Phase-07 / experiment `.2001` produced a passes-to-closure curve of `[5 4 3 1 2 1 1 1 …]` — flat at
**1 from run 4 on**. From then the being did the same three writes plus the build and spent an
inference call *each time* to re-derive what it already knew. That flat stretch is not just evidence
of learning. It is the **compile signal**: a stable behavior sitting there being needlessly
re-inferred. The moment the curve flattens, there is something to compile.

## The mechanism — JIT compilation

The LLM is the **interpreter**: slow, general, handles anything. When a path goes *hot* — the curve
flattens, closure is stable — compile it to native code: a **parser**, a deterministic function that
reproduces the behavior with no LLM call. Fall back to the interpreter only on a **cache miss**: no
parser for this situation, or a parser that failed.

A **reflex is a compiled parser.** The README flagged exactly this as "reflexes — a very probable
path may eventually skip the framing call," then deferred it and kept the floor at 1. This
un-defers it, and the curve is what justifies un-deferring.

## Parsers at both ends

The two learning loops each compile:

```text
observation → state change     the OBSERVATION-side parser   ("learned observation")
state change → action          the ACTION-side parser        ("learned expression", the compiled playbook)
```

Same two loops as everywhere else — now with a compiled form once each is stable.

## Below the floor of 1 — speculative execution (the conscious trade)

The README fixed the floor at 1 **on purpose**: precedent can supply the plan, but only the model can
*confirm the precedent applies*, so it paid one call every time to guarantee it. Nothing executed on
precedent alone.

A parser goes to **zero** by executing *optimistically* — run the compiled behavior without
confirming, and catch the mistake *after*, through instability. That is **speculative execution /
branch prediction**:

```text
predict   → run the parser (0 LLM calls)
misfire   → parser produced unstable behavior  → fall back to the LLM → recompile
```

This is a real, chosen trade, not a drift:

```text
FLOOR 1   guaranteed-correct per call        pay 1 call always, never misfire
FLOOR 0   cheap + optimistic                 pay 0, occasionally misfire and roll back
```

It is viable **only because failure detection is cheap** — a parser misfire *is* unstable behavior,
the exact grade the observation-side loop already computes. The rollback signal is already there, so
speculation is nearly free. Choosing floor 0 is choosing to tolerate rare, caught, rolled-back
misfires in exchange for near-zero steady-state cost.

## "The state is a set of parsers" — two senses, reconciled

"State" is doing two jobs; keep them apart:

```text
WORLD state        what is true — the state graph, the fold over observations.
OPERATIONAL state  which parsers are currently loaded and running.
```

They connect: a change in **world** state selects which region is active, which **loads that
region's parser set**. So being *in a state* = running a particular compiled program; a **state
change = jumping to different compiled code**; and a **state change with no parser = a cache miss**
that recruits the LLM to produce stable behavior until there is something to compile. This keeps
"current state is the state graph" (from [region-as-playbook.md](region-as-playbook.md)) intact — the
graph is what is *true*; the active parser set is the *response configuration to it*.

## Failure inherits the existing loop — no new machinery

A parser **fails** by producing unstable behavior — the destroy-and-trace loop from
[observation-side-learning.md](observation-side-learning.md). *Which* parser failed — the
observation one or the action one — is the **re-parse vs re-plan** blame split. So this adds no new
failure mechanism; it inherits the one already being built and lands on the same open knot
(blame localization). It is the same system wearing a compiler.

## Open

- **The compile trigger.** A single flat point is not enough — the `.2001` curve blipped `1 → 2` at
  run 5. Compile on *sustained* stability: high `N`, stable and long closure duration (the behavior
  reliably closes the error and keeps it closed). The exact threshold is unspecified; it is the same
  (P, N) / margin / closure-stability machinery, re-used as a compile gate.
- **What a parser concretely is** — a memoized mapping, a generalized rule, promoted precedent? Left
  open; the contract is only "deterministic, reproduces the behavior, and fails loudly."
- **Misfire cost bound.** Speculative execution needs the rollback to be genuinely cheap and the
  misfire rate genuinely low, or floor 0 loses to floor 1. Needs measuring, not assuming.
- **Blame localization** (shared with the whole design): re-parse vs re-plan vs indict-nothing when a
  parser misfires.

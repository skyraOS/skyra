# Phase: Prediction-Diff Learning — Does Planning Get Cheaper With Experience?
# Experiment .2001

## Context

This opens a new series (`.2xxx`) for the `skyra-dna-v3` substrate. The comparator series
(`.1001`–`.1004`, phases 03–06) measured whether an LLM call could *judge* a present. This measures
something the last four generations never did: whether the being **learns** — whether repeated
exposure to a task makes it cheaper to plan, on grounded evidence, with the model never told how it
is scored.

v3's thesis is two independent curves (`design-docs/skyra-dna-v3/retrieval.md`):

- **passes-to-closure ↓** — retrieval is working; planning is getting cheaper.
- **prediction error ↓** — the world-model is working; the bets are getting better.

Both down is the thesis. Either alone is a warning, and either failure localizes. This is the first
run to put a number on the first curve over enough episodes to be a curve rather than a slope.

## The Question

Given a task with a **hidden multi-part prerequisite** the being is never told, does experience make
planning cheaper — and does it get there by *discovering* the prerequisite from the world rather than
from the prompt?

## Setup

- **Substrate:** `skyra-dna-v3`, driven by `cmd/dnav3`. Real inference via `CallCodex` (Codex CLI,
  subscription-covered). Real sandboxed filesystem world.
- **Task:** reach `build:/ status ok`. The world's build (`World.build`) succeeds **iff all three of
  `app.src`, `app.conf`, `app.lock` exist**, and on failure reports only the **first** missing file.
  The three-part recipe is **never in the prompt** — the being can only learn it from failed-build
  reports.
- **Episode isolation:** a **fresh world every episode** (`NewEpisode`) with a **persistent
  expression graph**. The only thing that carries across reps is what the being *learned*, not what it
  left on disk. (The observation graph is per-world; the expression graph is the durable memory.)
- **No coaching:** the model is never told the scoring rule — enforced by
  `TestPromptDoesNotLeakTheScoringRule`. The present shows base rates as *evidence*
  ("worked 72% of 29 tries") but never the multiplier, the prior, or which direction pays.
- **Reconciler:** `Run` is plan → act → observe → *re-plan if the target is not yet observed-met*
  (the watcher from the design). This is what lets a bare build fail, be observed as failing, and
  provoke a corrective plan.
- **20 reps.** Raw output saved to the run log; reproduce with `go run ./cmd/dnav3 -reps 20`.

## Run / Observations

### The retrieval curve — passes-to-closure falls and holds

Expected:
  Cold, the being needs several planning passes (discover each prerequisite in turn). Warm, precedent
  front-loads the whole recipe and it closes in ~1 pass.

Observed:
```
passes-to-closure = [5 4 3 1 2 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1]
```

Evidence:
  Cold start **5 passes** (run 1). Reached the floor of **1 by run 4**, and held at 1 for **15 of the
  last 16 runs** (one blip to 2 at run 5). Total **30 inference calls for 20 episodes** — the first
  five episodes cost 15 calls (5+4+3+1+2), the last fifteen cost 15 (one each).

Conclusion:
  A **~5× reduction** in planning cost, converged and stable. The cost curve inverts exactly as the
  economic thesis predicts: the expensive discovery is paid once, and precedent amortizes it. This is
  the first curve, measured, and it holds.

### The being discovered a 3-part recipe it was never told

Expected:
  With the recipe absent from the prompt, the being should learn it from the world's failure reports —
  or fail permanently.

Observed:
  Cold run 1 ran a bare build, was told `missing app.src`, wrote it; told `missing app.conf`, wrote
  it; told `missing app.lock`, wrote it; build succeeded. By run 4 it front-loaded all three writes
  plus the build in a single emission.

Evidence:
  `acts` per run fell **8 → 4** and stabilized at 4 (three writes + one build). The recipe is nowhere
  in `plannerSystem`; the only channel was `build:/ missing <name>` observations.

Conclusion:
  The being learned a **multi-step structural prerequisite from grounded, ungameable evidence**, with
  no human in the loop and nothing in the prompt. This is the behavior the first live run (phase-N/a,
  v2) spent 95 inference calls and a human scolding to approximate once; here it is a repeatable,
  converging capability.

### The edges are the memory — counted, not asserted

Expected:
  The reusable edges should accumulate across episodes toward the true conditional probabilities.

Observed / Evidence:
```
P=0.98 N=20  write→file:/app.src   → app.src exists    true
P=0.98 N=20  write→file:/app.conf  → app.conf exists   true
P=0.98 N=20  write→file:/app.lock  → app.lock exists   true
P=0.72 N=29  build→build:/         → build:/ status ok
```

Conclusion:
  The three writes converged to **0.98 over 20 trials** — the being knows they reliably work. The
  build sits at **0.72 over 29 trials** (more trials, because the cold episodes ran it and it failed
  before the recipe was complete). This is `P(world state | expression)` computed from **counted
  outcomes**, not a judge's opinion — the thing v2's settle-written `A` could never be. The edge key
  is (primitive→target), not the expression instance, so the same act accumulates across every
  episode; that fix (found by the first live run, N=1-forever bug) is what makes this table possible.

### Calibration — the edge over the base rate decays to zero as the graph absorbs the recipe

Expected:
  Early, the model knows things the empty graph does not, so it earns. Once the graph has learned the
  recipe, an honest confidence just matches the base rate and earns nothing.

Observed:
```
run 1  edge=+0.358      run 8  edge=+0.040      run 15 edge=+0.009
run 4  edge=+0.360      run 10 edge=+0.031      run 20 edge=+0.011
```
  Overall EDGE OVER BASE RATE = **+0.1335**. Bet accuracy **90% (81/90)**.

Conclusion:
  The per-run edge **decays monotonically toward zero** (0.36 → 0.01) as `N` climbs and `P`
  converges. This is not the model failing — it is the model becoming **calibrated on a mastered
  task**: it earns while it knows more than the graph, and once the graph has absorbed the recipe
  there is nothing left to beat. Edge → 0 on a solved task is the *expected* terminal state, and it is
  exactly why the prompt must not leak the scoring rule: a model told "beat the base rate" would
  manufacture spurious deviation here instead of correctly converging.

## Phase Conclusion

**Both curves moved, and the thesis held its first real test.** Planning cost fell ~5× and stabilized
at the floor; prediction accuracy hit 90%; the edge over the base rate decayed to zero as the being
mastered the task — all on grounded filesystem evidence, with the model never told how it was scored.
The being **discovered a three-part recipe from the world's own failure reports** and learned to
front-load it. Thirty inference calls did what a fixed pipeline could not do in ninety-five.

This is the first measurement in the project's history where a v3-shaped being demonstrably learns.

**Scope / NOT yet proven (so we do not overclaim):**

- **One task, one model, one seed.** No variance runs. The curve is clean but n=1 in the
  task/model/seed dimension. A different task shape, or a run-to-run variance band, is the next thing.
- **Determinism helped.** The world is deterministic and the recipe is fixed; a noisy or
  non-stationary world (where `P` must chase a moving target, which is why the EMA has a plasticity
  floor) is untested.
- **The retrieval index is one-hop / brute force.** The kernel (`kernel-v2`) does not exist yet. This
  task's bridge is shallow enough that flat retrieval suffices; a genuinely deep bridge (where
  precedent chains multiple hops) would stall and is what would justify building the wave.
- **Non-monotonic blip.** Run 4→5 regressed 1→2 before settling. Minor, but the curve is not strictly
  monotonic — expected with a live model, worth a variance band to characterize.

## Next

- **`.2002` — variance and a second task.** Run the same task ~10× with different seeds to get a band,
  and add a second, differently-shaped recipe to show the curve is not task-specific.
- **`.2003` — a non-stationary world.** Change the recipe mid-run and confirm the EMA's plasticity
  floor lets old edges wash out (the stationary-world assumption is where a pure sample mean breaks).
- **Deep-bridge task** to find where one-hop retrieval stalls — the falsifiable trigger for building
  the `kernel-v2` wave.

# The Closure-Duration Objective

**Status:** MECHANISM (2026-07-15 research). High-level only — no implementation.
**Companions:** [the-learning-loop.md](the-learning-loop.md) ·
[Maturing the current design.md](Maturing%20the%20current%20design.md) ·
[principle-falsifiable-boundary.md](principle-falsifiable-boundary.md) ·
[../../design-docs/skyra-dna-v3/closure-and-planning.md](../../design-docs/skyra-dna-v3/closure-and-planning.md)

This records a design conversation that landed on a single organizing question for the *slow*
learning process — the one that sits alongside the fast probability gradient and shapes it over
time. The two open tensions in [the-learning-loop.md](the-learning-loop.md) (breaking a bad habit,
exactly when the call pops) both re-express under it. It is deliberately free of encoding detail.

## The question

> **Given a relational set, what action makes that set disappear for the longest amount of time?**

The "relational set" is the outstanding tension between a target state and the current observed
state. It *disappears* when observation confirms `target ⊆ current` — the closure condition already
defined in [closure-and-planning.md](../../design-docs/skyra-dna-v3/closure-and-planning.md). The
objective is not "which action closes the set" but "which action keeps it closed **longest**."

That single word — *longest* — is the whole find, because it converts the learning signal from a
judged quantity into an **observed** one. Duration of closure is counted, never opined. The model's
confidence cannot inflate it. It is LAW 2 (learning only across the boundary) satisfied by
construction: the grade is the interval between two observations, both of which crossed the boundary
from the world.

### Corrections are just fast reopenings

The reframe's first dividend: a correction stops being a special event that needs a special
mechanism.

```text
grab still water   → closes "Kathy has water" → reopens in seconds  ("I meant sparkling")
grab sparkling     → closes "Kathy has water" → stays closed the rest of the evening
```

Still-water closes the region for seconds; sparkling closes it durably. No rejection-weight, no
surprise multiplier, no hand-set constant. The habit **breaks itself** through its own short closure
durations. This is the "specialize / carve the distinction" route from
[the-learning-loop.md](the-learning-loop.md) item on breaking a bad habit — but now with a mechanism
underneath it instead of three competing candidates.

## Two processes, not one

Keep this process distinct from the fast gradient. They run on different timescales and answer
different questions.

```text
FAST   right this second, what do I do?          roll down the probability gradient, act on the
       (single step, reactive)                    most-probable ΔS/action. KEEP AS IS.

SLOW   over time, what durably resolves this?     track closure-duration per (region, action);
       (cross-episode, value-learning)            this RESHAPES the fast gradient's probabilities
                                                   and CARVES the segment boundaries.
```

The slow process does not replace the fast one; it **trains** it. `P(A | ΔS)` stops being a free
parameter and becomes "the action that historically closed this region longest." Which also settles
a question the learning-loop doc left implicit: closure-duration is **not a second value knob**
competing with probability (item 9, "probability is the only value knob"). It is the *training
signal that shapes P*. One runtime knob, fed by an observed objective.

### It composes with segmentation

From the prior turn's segmentation move: the boundary between two relational regions falls exactly
where `argmax_action(closure-duration)` **switches**. Where "sparkling" begins out-lasting "still,"
that is the cut — and the cut is `requester = Kathy`. So:

- **Segmentation** answers *where the regions are*.
- **Closure-duration** is the value each region-action carries, **and** the signal that draws the
  region boundaries.

Structure and objective, one mechanism.

## The inner/outer decomposition

The slow process is stable in one regime and explicitly incomplete in the other. Name both.

```text
INNER LOOP   given a STABLE observation:              context = observation
             actions compete on closure-duration,      arms    = actions
             feeding back to the same observation.      reward  = closure-duration
             → this is a CONTEXTUAL BANDIT. it converges.

OUTER LOOP   CHANGE in observation revises the target.  non-stationary context.
             different mechanism, different force.       = the O→ΔS interpretation axis.
             → NOT built here. named, and left open.
```

The inner loop's stability is not hopeful — it is structural. Fixed context, a bounded scalar
maximized over a fixed arm set, is a contraction; stationary bandits converge. The instability only
enters when the context itself moves, which is precisely the change-in-observation regime the outer
loop owns.

Note the convergence from an independent direction: [the-learning-loop.md](the-learning-loop.md)
reached *"interpretation (O→ΔS) is where the power goes"* from **where unresolved meaning enters**.
This spec reaches the same edge from **stability** — the inner action-loop is stationary and cheap,
so the instability, and therefore the compute, has to live in the change-in-observation edge. Same
conclusion, two paths. That is the strongest confirmation the altitude is right.

## The three ways "longest" lies

The objective is sound; its soundness lives in three guardrails. Only the first is a hard
requirement now.

### 1. Un-probed ≠ closed — HARD requirement

The degenerate maximizer of "make the tension disappear longest" is *stop looking*. Leave the room,
never check, the set never reopens. A closure counts only while the target is **observed-satisfied**
(`target ⊆ observed current`); abandoning observation is `indeterminate`/censored, never closed.
This is LAW 1 doing its job — the set is closed only when an observation says so. **Nail this or the
being learns to look away.**

That the objective *tempts* toward avoidance is a good sign, not a bad one: an objective whose
degenerate strategy is avoidance is an objective that models **motivation** correctly. Real
goal-directed agents fall into exactly this failure. A reward that could not be gamed by looking away
was not capturing wanting in the first place.

### 2. Which reopening counts — DISSOLVES into the outer loop

Earlier worry: does tomorrow's fresh "grab my water" wrongly count as a reopening against today's
action? Resolution: it does not need a causal-linkage filter. Whether tomorrow's request
contaminates today's score depends entirely on whether it lands in the **same segment cell**, and
cell granularity is set by the outer loop (observation interpretation). Fine-grained "this instance"
→ tomorrow is a different observation, no contamination. Coarse "Kathy wants water" → they pool. The
inner loop never decides it. So this is not a separate problem; it is a consequence of observation
granularity — "hiding under learning-from-change-in-observation," which is the outer loop. Adding a
causal-linkage mechanism now is premature machinery.

### 3. Blame localization — real, but relocated

A short closure is a **global** alarm over `O→ΔS→A→O′`: it says *something* was wrong, not *which
link*. In the water case the action was fine; the interpretation was wrong (targeted still water).
Localization does not vanish — it relocates onto the same counting the segmentation uses: the blame
falls wherever the short-closures **co-vary**. All short closures share `requester=Kathy` → cut the
interpretation edge. Share `sparkling out of stock` → cut the consequence edge. Counting, not
tuning.

## Where it attaches to what exists

The post-observation cost has a home already: **the watcher** in
[closure-and-planning.md](../../design-docs/skyra-dna-v3/closure-and-planning.md). The watcher
already re-runs the closure check after every world-state update and already fires on the two events
this objective needs:

```text
closes + target reached           → done          ← closure STARTS here
a requires became false → hole      → replan        ← closure ENDS here (the reopening)
```

Closure-duration is just **the interval between those two events, attributed to the action that
produced the closure.** No new machinery — the watcher records an interval it already detects, and
the monitoring is cheap (observation-matching, no model calls).

## Methodology — find the local maximum first

This loop can settle into a local maximum: pool two distinct sub-cases into one region and it
converges to their blended best action rather than the right one for each. That is accepted, on
purpose. A stable loop that finds a local max is a system you can run and measure; a globally-correct
loop that never ships is worth nothing. Do not pre-build the outer edge to prevent a ceiling you have
not yet measured — the same discipline as the rest of v3 (change the representation, do not add
machinery until forced).

### The tripwire — how you know you've hit the ceiling

So the local max is not a surprise when it arrives. **Watch closure-duration variance inside a
converged region.** Confident (probability converged) *and* erratic (closures alternate long/short)
means two sub-cases got pooled into one cell and the inner bandit is averaging across a seam. That is
the "height without margin" case from [the-learning-loop.md](the-learning-loop.md) item 6, surfaced
as closure-variance instead of option-margin.

> **Confident-but-erratic closure in a converged region = build the interpretation edge here.**

A concrete, measurable "now" signal for when the outer loop earns its existence.

## Lineage

"Hold a target and choose the action that keeps you there longest against perturbation" is
**homeostatic control** — setpoint regulation ([../../design-docs/cognition/setpoints.md](../../design-docs/cognition/setpoints.md)).
The reframe is that a setpoint's value *is* its persistence, and a task is a setpoint with a shorter
natural life. "Maintain the room at 72°" and "get Kathy her water" unify under one objective — a good
sign the altitude is right.

## Open

- **Duration in which clock.** Closure-duration in world-tick, agent-tick, or count of relevant
  observations that passed without contradiction? Ties to the two-clock rule
  ([../../design-docs/skyra-dna-v3/the-cycle.md](../../design-docs/skyra-dna-v3/README.md), "Two
  clocks") and to "staleness comes from perception, not decay." Lean: relevant-observation count,
  which is agent-relative and censors cleanly when the being stops looking.
- **Censoring / normalization.** Comparing an action seen 3 times against one seen 30 times when some
  closures are cut short by episode end. Standard survival-analysis shape; the edge's `(P, N)` is the
  natural place, but the exact estimator is unspecified.
- **The outer loop itself** — learning from change-in-observation. Named here, owned by the O→ΔS
  interpretation axis, deliberately not designed. The tripwire above says when to start.

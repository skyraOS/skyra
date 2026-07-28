# v4 Design Set — 2026-07-24

The current, self-consistent design set. Brought forward from `2026-07-22` after a
reconciliation pass that removed the last known contradictions between docs.

> **Start with [COHERENCE.md](COHERENCE.md)** — the sweep result. It says which docs are the
> LIVE CORE (build from these), which are v3 reference, and which are stale/excluded. If you
> only read one meta-doc, read that one.

**Nothing in this set contradicts anything else in this set.** Where a doc corrects an
earlier position, the correction is marked inline with its date and reasoning — the old
position is kept, not deleted.

---

## Reading order

**1 · What things are**

| doc | status |
|---|---|
| [definitions](definitions.md) | **THE MEASUREMENT LAW + FORMAL DEFINITIONS.** §0: the runtime admits only the measurable; measurement is a position on the falsification–truth scale. Then Attention `A` (ranks), Truth `p` (decides), Grade `g` (pays), Dynamics `ΔB` (moves). **§3a the two graded arms · §3b the tense split — *sight runs backward, forward is only ever committed to*.** Start here. |
| [v4-data-model](v4-data-model.md) | **CORE VOCABULARY.** §1–3 settled. The spine: Observation → Belief → Action, and what each record is. |
| [tokenizer](tokenizer.md) | The stable sensory alphabet and the **concrete/segmented cycle** — the representation everything else assumes. |

**2 · How the wave moves**

| doc | status |
|---|---|
| [perturb-observe-traversal](perturb-observe-traversal.md) | **CORE DECISION.** Segmented descent, concrete ascent, `CreationPhase` visibility. |
| [kernel-alteration](kernel-alteration.md) | **DECIDED.** The kernel executes the returned Space and does zero cognitive math. |
| [interaction-primitives](interaction-primitives.md) | **SUSPENDED overall; §9 DECIDED.** The excitation field is the first traversal layer. §1–§8 describe v3 and are not being built. |

**3 · What cognition does**

| doc | status |
|---|---|
| [belief-expression-inference](belief-expression-inference.md) | **CORE DECISION.** Two inference calls, the desire-blind boundary, resolve / hold / disregard. |
| [settlement-and-surprise](settlement-and-surprise.md) | **SPLIT DECIDED, measure open.** Per-expectation settlement is credit + falsification; a window distribution diff is surprise + coverage. The diff gates; settlement grades. §7's metric question closed by definitions §3b. |
| [staged-emission-and-credit](staged-emission-and-credit.md) | **RULE SETTLED.** Citation routes, activation weights — the model names its triggers, the system weighs its credit. `trigger→product` emission; prediction-as-a-stage proposed — now with its strongest argument (ΔB becomes gradeable) and its hidden cost (the action worker loses implicit future-awareness). |
| [time](time.md) | Proper time. §3b–d and §4 decided; §3a and §3e still open. |

**4 · Getting there**

| doc | status |
|---|---|
| [v4-port](v4-port.md) | **WORKING DOC.** What the port is, measured against the v3 code. |
| [workers](workers.md) | **OPEN LIST.** The worker registry; predates some current vocabulary. |
| [episode-names-two-different-things](episode-names-two-different-things.md) | **DECIDED, not yet applied.** `EpisodeID → RunID`, `Class → Episode`. |

**5 · Reference — what v3 actually does**

| doc | status |
|---|---|
| [v3-mechanisms](v3-mechanisms.md) | **DESCRIPTIVE.** The 16 mechanisms, what each owns, and where the boundaries are not clean. |
| [dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md) | **DESCRIPTIVE.** The triple-nested loop and what does/doesn't map to v2's pattern. |
| [simplifications-found-in-review](simplifications-found-in-review.md) | **OPEN work order**, v3 scope. Quality only; nothing applied. |

---

## Left behind in `2026-07-22`, deliberately

| doc | why |
|---|---|
| `traversal-spine.md` | **RETIRED.** Describes a field model that no longer exists (entities/triples/spans) and collides on layer names. Everything worth keeping was folded into this set; its banner says where each piece went. |
| `defects-found-in-review.md` | v3 work order, **all in-scope items closed 2026-07-22**. Kept as the record. Still linked from this set where cited. |

---

## The model in one page

```text
REPRESENTATION      CONCRETE    bounded, ordered, identifiable whole
                    SEGMENTED   distributed, individually addressable token material
                    PAYLOAD     temporary bounded selection assembled to cross a field

LAYERS              L0  tokenized observations
                    L1  tokenized beliefs
                    L2  tokenized actions   ← expectations are a PHASE of this layer

TRAVERSAL           Perturb   traverses segmented material, returns ONE executable Space
                    Observe   post-order concrete fold; nil from Perturb is the base case
                    kernel    keeps lineage; performs no cognitive math

AMPLITUDE           space-owned. amplitude = evidence phase + magnitude.
                    the kernel carries an opaque impulse payload only.

THE CUT             budget = model context window · Cost = tokens · magnitude = share
                    affordability IS price-per-token, computed in a Space's own behavior

VISIBILITY          same-kind hidden; other-kind visible when done
                    no inference call ever reads a concrete Expectation

TIME                a tick = one Observation. count is the being's own; wall sets the metric.
                    Within is wall-denominated — a claim about the world.
```

---

## What is actually open

Consolidated so it is visible rather than buried in ten docs.

```text
belief-expression §10   the leverage function for observation detail
belief-expression §10   whether leverage is local per Belief or root-allocated
belief-expression §10   entity resolution as a mechanism — V0 defers it deliberately
settlement-surprise §7a MEASURE β FIRST — one logged field decides whether the whole
                        calibration-curve line of work has any content (β = 0 ⇒ drop it)
settlement-surprise §7b the p ≈ 0 suppress branch, and tail accuracy of the curve
settlement-surprise §8  the divergence measure: coverage vs calibration (or both)
settlement-surprise §8  what bounds the window · does the diff ever grade, or only gate
settlement-surprise §7  the FLOOR's magnitude, and reps-to-reach-it (the metric question
                        itself is CLOSED — see definitions §3b)
definitions §3a         the baseline fork is ONE PER ARM, and both are open
definitions §3a         where a reading's discriminating test lives (v4-data-model §5)
kernel-alteration       the Observe signature migration across every Space implementation
interaction-prim §9.7   distinct-origin dedup identity for raw injections
time §3a                does the cognitive agent take a Slice
time §3e                RECORD THE BASELINE — a distribution, not a run. blocks measurement.
v4-data-model §4b       compaction: the log grows without bound once runs stop wiping it
```

**One open bet, stated so it is visible** — [definitions §3b](definitions.md): act selection
is steered by belief and retrieval alone, never by lookahead. It loses if `GradeSum` rises
while `PassesToClosure` stays flat, and the fallback is already half-built
(`Expectation.By` is act-keyed and unread).

**One of these blocks measurement rather than design.** `time` §3e — no baseline
distribution, so no claim about the port preserving behavior is falsifiable, and no way to
say how many reps either curve needs to reach its floor.

> **Closed 2026-07-25.** `settlement-and-surprise` §7's *which metric carries "prediction
> error"* is answered in favour of **calibration**. `GradeSum` is the **headroom** curve, and
> its decay toward the floor is expected rather than a failure — so a converged run and a
> broken run no longer look alike. See [definitions §3a–§3b](definitions.md).

**One prerequisite now has two dependents.** `interaction-primitives` §7 (projected/observed
phase) is required by both the settlement-eligibility surface
([v4-data-model §4a](v4-data-model.md)) and the window distribution diff
([settlement-and-surprise §4](settlement-and-surprise.md)). It is marked SUSPENDED.

# v4 Design Set — 2026-07-24

The current, self-consistent design set. Brought forward from `2026-07-22` after a
reconciliation pass that removed the last known contradictions between docs.

**Nothing in this set contradicts anything else in this set.** Where a doc corrects an
earlier position, the correction is marked inline with its date and reasoning — the old
position is kept, not deleted.

---

## Reading order

**1 · What things are**

| doc | status |
|---|---|
| [definitions](definitions.md) | **THE MEASUREMENT LAW + FORMAL DEFINITIONS.** §0: the runtime admits only the measurable; measurement is a position on the falsification–truth scale. Then Attention `A` (ranks), Truth `p` (decides), Grade `g` (pays), Dynamics `ΔB` (moves). Start here. |
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
| [settlement-and-surprise](settlement-and-surprise.md) | **SPLIT DECIDED, measure open.** Per-expectation settlement is credit + falsification; a window distribution diff is surprise + coverage. The diff gates; settlement grades. |
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
settlement-surprise §7  WHICH METRIC CARRIES "prediction error" — GradeSum decays to zero
                        as the graph converges, so its healthy shape is an inverted U
kernel-alteration       the Observe signature migration across every Space implementation
interaction-prim §9.7   distinct-origin dedup identity for raw injections
time §3a                does the cognitive agent take a Slice
time §3e                RECORD THE BASELINE — a distribution, not a run. blocks measurement.
v4-data-model §4b       compaction: the log grows without bound once runs stop wiping it
```

**Two of these block measurement rather than design.** `time` §3e — no baseline distribution,
so no claim about the port preserving behavior is falsifiable. And
`settlement-and-surprise` §7 — if `GradeSum` is the thesis's "prediction error," a converged
run and a broken run look alike, because a converging system drives it toward zero.

**One prerequisite now has two dependents.** `interaction-primitives` §7 (projected/observed
phase) is required by both the settlement-eligibility surface
([v4-data-model §4a](v4-data-model.md)) and the window distribution diff
([settlement-and-surprise §4](settlement-and-surprise.md)). It is marked SUSPENDED.

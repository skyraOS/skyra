# Coherence Map — the 2026-07-24 Set

**Date:** 2026-07-24
**Status: SWEEP RESULT.** The outcome of a doc-by-doc coherence pass. It records which docs
form the consistent live core, which are reference, which are stale, and every incoherence
found (with its resolution or its open flag).

**How to read it:** if you are building, read only the **LIVE CORE** (§1). Everything else is
reference, stale, or descriptive of the v3 code being replaced.

---

## 1. The live core — these cohere

Checked pairwise against each other and against [definitions](definitions.md). No
contradictions remain after the 2026-07-24 sweep fixes (§4).

| doc | owns | anchored to |
|---|---|---|
| [definitions](definitions.md) | **THE MEASUREMENT LAW** · A · p · g · ΔB · the crank | the root |
| [tokenizer](tokenizer.md) | concrete/segmented cycle · the token alphabet | representation |
| [v4-data-model](v4-data-model.md) | the concrete structs (Observation/Belief/Action/Expectation) | §1–3 settled |
| [perturb-observe-traversal](perturb-observe-traversal.md) | Perturb descends · Observe ascends · CreationPhase | traversal |
| [kernel-alteration](kernel-alteration.md) | kernel does zero math · executes the returned Space | kernel |
| [belief-expression-inference](belief-expression-inference.md) | the two inference calls · resolve/hold/disregard | cognition |
| [settlement-and-surprise](settlement-and-surprise.md) | settlement vs window-diff · calibration curve · MI identity | learning |
| [staged-emission-and-credit](staged-emission-and-credit.md) | **citation routes, activation weights** · the trigger→product format · prediction-as-a-stage (proposed) | credit |
| [time](time.md) | proper time · the sensory organ · wall sets the metric | time |
| [interaction-primitives](interaction-primitives.md) **§9 only** | the excitation field as first traversal layer | field |

### The spine that ties them, one line each

```text
definitions       only the measurable is admitted; measurement is a scale, not a verdict.
tokenizer         a claim is CONCRETE (a whole) and SEGMENTED (addressable tokens).
v4-data-model     the concrete records: Observation → Belief → Action(+Expectation).
perturb-observe   descend the segmented field (Perturb); realize concrete on ascent (Observe).
kernel            the kernel executes returned Spaces; all math lives in the Space's Observe.
belief-expression the model interprets (belief call) and acts (expression call); desire-blind.
settlement        the world settles addressed claims; a second phase-free surprise signal (ΔB).
time              a tick is one Observation; the being ages by flow through it.
field (§9)         heat is on the segmented tokens; the field is where the walk begins.
```

**The four keystones that make the join hold** (each was a contradiction until closed):

```text
heat lives on TOKENS, not beliefs        v4-data-model §1/§4a · interaction-primitives §9 · definitions §0
a Belief is a CONCRETE WHOLE             v4-data-model — reached because its tokens are hot
Expectations are a PHASE OF L2           perturb-observe §6a · v4-data-model §3 · belief-expression §3
the kernel does ZERO MATH                kernel-alteration · perturb-observe · traversal-spine (retired)
```

---

## 2. Reference — describes the v3 code being ported, not the v4 design

These are **correct as descriptions** of the existing implementation. They use v3 vocabulary
(`triple`, `Express`, the three nested loops) because that is what v3 has. Do not read them as
the v4 design; read them to know what is being replaced.

| doc | is |
|---|---|
| [v3-mechanisms](v3-mechanisms.md) | the 16 v3 mechanisms, what each owns, where boundaries are clean/not |
| [dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md) | the v3 triple-nested loop and what maps to v2's pattern |
| [simplifications-found-in-review](simplifications-found-in-review.md) | v3 quality work order — none applied |
| [v4-port](v4-port.md) | **§2 valid** (measures v3 code). **§5 Q1–Q2 resolved** (marked inline); Q3–Q6 open. |

---

## 3. Stale — excluded from the live core until rewritten

| doc | why it is out | what survives |
|---|---|---|
| [workers](workers.md) | pre-token-model. Asserts *"Belief is the unit of heat"* (wrong), uses `sensory:*`/`plan:*` names that predate Perturb/Observe and the two inference calls, and `[]Triple` ingest. Marked SUPERSEDED-IN-PART in its own header. | the *responsibilities* (grade write-path, surprise gate, desire-blind boundary) and the naming convention |
| [episode-names-two-different-things](episode-names-two-different-things.md) | **DECIDED, not stale — just not applied.** `EpisodeID→RunID`, `Class→Episode`. Coheres; it is a pending rename, not a contradiction. | all of it |

---

## 4. Incoherences found this sweep — and their disposition

```text
FIXED — internal contradiction in v4-data-model
  §1 vocabulary table (line 135) and §2 struct comment (line 199) still read "unit of
  heat"/"heated unit" while §1 box, §4a table, and §8 struct-budget were corrected. Now
  all five agree: Belief is a concrete whole; heat is on tokens.

FLAGGED — workers.md contradicts the heat model and the traversal model
  Marked SUPERSEDED-IN-PART in its header (§3 above). Not deleted — its responsibility
  inventory is still the reference for the eventual worker rewrite.

FLAGGED — v4-port §5 Q1/Q2 were resolved elsewhere
  Marked RESOLVED inline with links. The rest of v4-port stays as v3 measurement.

GAP (not a contradiction) — the measurement law is stated but not propagated
  definitions §0 establishes it; the other live docs predate it and do not yet reference
  it. They do not CONTRADICT it — settlement, LAW 5, and the P∈[0,1] scale are all already
  consistent with "only the measurable, on a scale." Threading the vocabulary through is a
  pending pass, tracked in §5.

GAP (not a contradiction) — the struct↔token join is stated in v4-data-model but thin
  v4-data-model now references tokenizer (Belief is concrete, tokens carry heat), and
  tokenizer's concrete/segmented cycle covers it. Coherent, but the JOIN deserves one
  explicit paragraph rather than living across two docs. Pending.
```

---

## 5. Pending passes — coherence debt, not contradictions

None of these block building from the live core; they are threading work.

```text
1. workers.md rewrite      re-derive worker boundaries against Perturb/Observe + the two calls
2. measurement-law thread  reference definitions §0 from settlement, v4-data-model, belief-expr
3. struct↔token paragraph  one explicit statement that v4-data-model structs are the CONCRETE
                           forms whose segmented projection is the field
4. interaction-primitives  split the SUSPENDED v3 description (§1–8) from the DECIDED §9, or
                           retire §1–8 the way traversal-spine was retired
```

---

## 6. The one dependency the whole set still routes through

Unchanged by this sweep, restated because it is the critical path:

> **`interaction-primitives` §7 — projected/observed evidence phase — is SUSPENDED, and four
> live things depend on it:** the §5 emission-leak cleanup, the settlement-eligibility surface
> ([v4-data-model §4a](v4-data-model.md)), the grading calibration curve
> ([settlement-and-surprise §7a](settlement-and-surprise.md)), and per-layer truth
> ([settlement-and-surprise §7b](settlement-and-surprise.md)).

And the cheapest decisive experiment, also unchanged: **measure `β`** — one logged field
(activation-at-arm on the Expectation, paired with the settled outcome) collapses or confirms
the entire calibration limb.

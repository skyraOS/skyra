# Toy #8 — Density Opens the Valley Before the Predicate Names It

**Date:** 2026-07-21  
**Status:** PARTIAL PASS · density-before-naming geometry passes; trusted predicate
promotion and pocket-driven commits are blocked by the
[Phase 3 corrective audit](phase-3-audit-geometry-is-not-trusted-interpretation.md)  
**Parent:** [Interpretation Layer — Implementation Phases](interpretation-implementation-phases.md)  
**Reads on:** [Episodes Over the Field](episodes-over-the-field.md),
[Predicate Diagnosticity](predicate-diagnosticity.md), and
[Scale, Structure, Content](scale-structure-content.md)

---

## The forcing constraint

Phase 3 cannot search for a marker such as `school` and then announce that the marker
proved two pockets exist. That would impose the criterion first and recover the requested
split by construction.

The order must be:

```text
positive settlement spans
    → span-to-span geometry
    → a valley earns the split
    → only then fit a predicate that names each side
```

The smallest geometry already available in v3 is a span's set of raw observable triples.
Candidate completion triples are removed before comparison so the outcome cannot trivially
separate itself. Interpreted observations and marks are also excluded: Phase 3 may discover
structure only in world-owned material.

For each Phase-1 branch, use the **currently active prediction-time source spans** whose
expressions later produced positive settlements, shared by all completions in the branch.
Resolving spans retain the outcome evidence but do not supply contextual features. The
intersection matters for a conjunctive branch: an occurrence of one conjunct alone is not
an occurrence of the conjunction. The current span is held out from fitting so an instance
cannot train the predicate that matches itself.

```text
within(branch) = mean Jaccard(raw-feature sets for pairs of its occurrence spans)
between(a,b)   = mean Jaccard(raw-feature sets across the two branches)

raw valley     = min(within(branches)) - max(between(branch pairs))
```

The minimum and maximum are conservative. Every claimed pocket must be internally dense,
and the closest two pockets determine whether a real valley separates the region. Empty
feature sets carry no similarity; two empty spans are not evidence of a pocket.

A single repeated example must not mint a code. Let:

```text
N_pocket       = minimum number of positive occurrence spans in any branch
maturity       = N_pocket / (N_pocket + 5)
effectiveDepth = max(0, raw valley) × maturity
```

The existing `confHalf=5` supplies the saturation curve. The split opens only when
`effectiveDepth >= cfg.SplitValleyDepth`. This keeps the Phase-3 threshold in its own unit
while making recurrence part of the geometry rather than a separate promotion rule.

---

## Resolvable fixture

One source has two alternative completions, A and B. Every occurrence contains a shared
regional feature. A occurrences also contain `context=school`; B occurrences contain
`context=skyra`.

```text
A1 A2 A3:  {shared, school}
B1 B2 B3:  {shared, skyra}
```

Within either branch, every pair is identical:

```text
within(A) = within(B) = 1.000
```

Across branches, only `shared` overlaps:

```text
between(A,B) = 1 / 3 = 0.333
raw valley   = 1.000 - 0.333 = 0.667
```

The graduation is density-driven:

```text
two occurrences per branch:
    maturity       = 2 / 7 = 0.286
    effectiveDepth = 0.667 × 0.286 = 0.190  < 0.200 → still one blur

three occurrences per branch:
    maturity       = 3 / 8 = 0.375
    effectiveDepth = 0.667 × 0.375 = 0.250  ≥ 0.200 → valley opens
```

Only after the `0.250` valley exists are naming predicates fitted. For this first bounded
vocabulary, a predicate is one exact raw triple. Selection uses region-relative
diagnosticity:

```text
score(d, branch) = coverage(d | branch) × log(coverage(d | branch) / baseRate(d | region))

school for A: coverage=1.0, baseRate=3/6=0.5 → score=log(2)=0.693
skyra  for B: coverage=1.0, baseRate=3/6=0.5 → score=log(2)=0.693
shared for either: coverage=1.0, baseRate=6/6=1.0 → score=0
```

The fitted names therefore describe the already-discovered pockets; they did not create
them. The prototype lets a unique predicate match in the current raw span select its branch
as the counted interpretation. The audit found that this is not yet earned: the predicate
is selected and scored on the same historical examples, and the held-out current span only
shows that it fires—not that it predicts the branch. Until disjoint evaluation and sparse
promotion pass, a match is a provisional pocket description and must not force or override
an operational commit. Independent interpretation settlement remains Phase 4.

---

## Irreducible control

Give A and B the same distribution of contextual features:

```text
A: two {shared, red},  two {shared, blue}
B: two {shared, red},  two {shared, blue}
```

Within-branch and between-branch similarity converge to the same geometry. In this exact
fixture, the closest cross-pocket similarity is at least the internal similarity, so:

```text
raw valley ≤ 0
effectiveDepth = 0 at every N
```

More observations increase certainty that the fork is irreducible; they never manufacture
a differentiator. No pocket predicate is fitted and the source remains a well-counted
`AMBIGUOUS` region.

---

## Adversarial checks

- **Outcome leakage:** including A and B themselves makes every branch perfectly separable.
  All candidate completion keys must be removed from the feature vectors.
- **Predicate-first leakage:** diagnosticity is not evaluated until the geometry passes the
  valley gate.
- **Sparse-code minting:** one occurrence has no within-pair density; two perfect pairs are
  still below the default effective-depth threshold.
- **Conjunction inflation:** branch occurrences are the intersection, not union, of member
  completion spans.
- **Ordering:** span, branch, and candidate insertion order cannot change the valley or
  selected predicate.
- **Multiple fires:** a current span matching multiple pocket predicates vetoes even a
  generic heat commit; zero matches add no pocket-specific selection.
- **Active bound:** settlement membership is queried over heated spans; pocket discovery
  never scans an edge's unbounded historical settlement list.
- **Unresolved topology:** a Phase-1 non-clique remains unresolved. Phase 3 does not invent
  branch ownership for material Phase 1 could not partition.
- **Selection/evaluation separation — NOT YET PASSED:** candidate predicate search and
  branch-predictive evaluation must use disjoint labeled evidence. The unlabeled current
  application is not an evaluation set.
- **Sparse multiple-search support — NOT YET PASSED:** promotion must use a support bound
  that accounts for selecting the maximum among many features; `PromotionEvidence` must be
  consumed by the live path rather than merely configured.
- **Decision-time feature ownership — PASSED 2026-07-21:** settlement separately indexes
  the prediction-time source span, and density fits on that raw source evidence rather than
  arbitrary facts in the later resolving span. A correlated post-outcome-error fixture
  cannot change the pocket name.
- **Action-source identity — PASSED CONSERVATIVELY 2026-07-21:** primitive, target, payload
  digest, and sorted requirements form the initial key. Different payloads/guards separate;
  requirement order does not fragment identity. Learned equivalence remains open.
- **Cross-episode interpretation recall — NOT YET PASSED:** learned pocket structure must
  remain retrievable after the per-world observation/span indexes are replaced, without
  laundering old world facts into the new world.

---

## What is settled and what is not

**Settled for the Phase-3 geometry:** density is measured before naming and only over the active region;
the initial geometry is within-versus-between Jaccard over raw span features; recurrence
discounts the valley with the existing confidence curve; predicates are fitted by
region-relative diagnosticity only after the valley passes; and the irreducible control
does not split.

**Implemented but not trusted:** exact predicates are fitted after the valley and a unique
current match currently selects a branch. That path demonstrates plumbing, not validated
predictive authority. It must not be treated as a settled operational rule.

**Still open:** exact triples are only the first criterion vocabulary. Regex, token,
structural, and embedding sensors still need the composable calibration described in
Scale/Structure/Content. Mean Jaccard treats every raw triple as one dimension and may be
too crude for large multimodal spans. Split hysteresis/merge behavior is also unspecified.
These are replaceable behind the density-first contract; they are not reasons to reverse
the order and let a predicate impose the split.

## Corrected result

Toy #8 partially passes. The resolvable region remains blurred at two occurrences, opens a measured
valley at three, receives `school`/`skyra` names only afterward, and can deterministically
compute a uniquely matching current branch. The equal-distribution control accumulates
forever without splitting. What has not passed is the move from deterministic match to
trusted commit. Density discovers a possible region; naming describes it provisionally;
out-of-sample promotion, causal availability, source identity, and durable recall must earn
operational authority.

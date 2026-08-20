# Phase 3 Audit: Geometry Is Not Yet Trusted Interpretation

**Date:** 2026-07-21  
**Status:** CORRECTIVE AUDIT · operational pocket commits are now mechanically blocked;
prediction-time feature ownership and conservative action identity have landed. Disjoint
predicate promotion and durable cross-episode interpretation recall remain open  
**Audits:** [Toy #8 — Density Opens the Valley Before the Predicate Names It](toy-8-density-pocket-differentiation.md)
and [Interpretation Layer — Implementation Phases](interpretation-implementation-phases.md)  
**Reads on:** [Biggest Risks and Required Gates](biggest-risks.md),
[Predicate Diagnosticity](predicate-diagnosticity.md), and
[Toy #9F — Span-Owned Prediction Fields](toy-9f-span-owned-prediction-fields.md)

---

## Why this correction exists

Toy #8 established a useful ordering: measure recurrence geometry before fitting a name.
The implementation follows that order and preserves several important controls—positive
settlement polarity, candidate-outcome exclusion, permutation invariance, unresolved
vetoes, and the active-region bound.

Those facts do not establish that a fitted name has earned predictive authority. A review
of the landed path found four ways the current prototype can turn a real geometric split
into an unjustified operational interpretation.

The corrected dependency is:

```text
density discovers a possible region
    ≠ predicate has earned predictive authority
    ≠ a current match may safely force a commit
```

The geometry prototype is not being discarded. Its epistemic status is being narrowed to
what the tests actually demonstrate.

## Finding 1 — selection and evaluation reuse evidence

`fitPocketPredicate` searches the features in the branch occurrence spans and calculates
coverage, base rate, and diagnosticity on those same spans. The current span is excluded
from that fit, but it does not evaluate whether the selected predicate predicts the branch;
it merely checks whether the selected predicate fires and then may override heat with a
confident target.

With many candidate features, the maximum diagnosticity on a small sample is selection
bias. Holding out the current unlabeled application does not repair it. This conflicts with
the existing gates requiring disjoint selection/evaluation evidence and sparse-data support
bounds. `PromotionEvidence` is configured but is not consumed by the current pocket path,
so there is no separate promotion barrier behind the selected maximum.

**Status:** the unsafe authority path is blocked. A unique match is retained as provisional
`PocketMatch` metadata but cannot create, override, or veto a commit. The estimator and
promotion gate remain open.

**Required gate:** candidate search and evaluation use disjoint labeled historical
evidence—through a fixed holdout, cross-fitting, or an equivalent precommitted scheme—and
promotion uses a lower support/confidence bound that accounts for candidate search. The
current span is application evidence only. Until this passes, a pocket name may be shown as
a provisional description but must not force `SourceConfident` or override another route.

## Finding 2 — durable statistics do not imply durable interpretation recall

`NewEpisode` correctly replaces the per-world observation graph and span index while
preserving `ExpressionGraph`. Consequently, counted causal-edge statistics survive world
replacement. But normal impulse retrieval, density feature reconstruction, and pocket
selection depend on the new episode-local span postings and observations. Old expression
birth spans may still exist in the expression graph, yet ordinary impulses cannot generally
reheat or reconstruct their old contextual feature vectors after the span index is reset.

The precise failure is therefore not "all learning resets." It is:

> Edge statistics survive episodes, but learned interpretation recall, pocket features,
> and ordinary impulse-driven access to prior expressions generally do not.

**Status:** architectural gap. It prevents demonstrated interpretation learning from
behaving as durable memory across worlds.

**Required gate:** after learning a discriminating pocket in episode A, start episode B
with a fresh observation graph and present an appropriate cue. The learned code must be
retrievable without treating episode-A world observations as current truth. This likely
requires a durable compressed code/pocket index containing sufficient statistics and
provenance, distinct from the per-world observation graph.

## Finding 3 — resolving-span features may contain effects

The current implementation builds a pocket feature vector from every raw observation in a
span that settled a branch, excluding only the candidate completion triples. A resolving
span may also contain error messages, secondary consequences, or downstream effects. Those
facts can become purported context predicates even though they were unavailable when the
action or prediction was chosen.

Removing the exact target prevents trivial target leakage; it does not enforce temporal or
causal availability.

**Status:** fixed in the landed sampling path. Settlement now records both the resolving
span and the expression's prediction-time source span. Phase 3 reads raw features through
the source-span index; resolving effects remain settlement evidence and cannot name the
pocket. An adversarial regression injects perfectly branch-correlated post-outcome errors
and confirms they do not enter the fitted name.

**Required gate:** training examples pair features available at prediction time with the
later branch settlement:

```text
source / decision-span evidence  →  later settled branch outcome
```

The feature side should be owned by the prediction's birth span or explicit `BornFrom`
evidence, with a frozen eligibility boundary. Filtering the resolving span by exclusion
alone is insufficient. Phase 4F's span-owned prediction and comparison fields should carry
this ownership explicitly.

## Finding 4 — source identity aliases materially different actions

The present source key contains only each responsible act's `Primitive → Target`. `Payload`
and `Requires` are omitted. Two writes to the same file with different contents, or the
same primitive under different necessary guards, therefore share one source and one edge
history. Their consequences may be classified as alternatives, conjunctions, or pockets
despite coming from materially different interventions.

This is not harmless abstraction: the equivalence was never learned.

**Status:** fixed conservatively. The reusable action signature now includes primitive,
target, a SHA-256 payload digest, and sorted canonical requirements. Tests prove different
payloads and guards separate while requirement insertion order does not fragment identity.
Learned action-equivalence compression remains future work.

**Required gate:** begin with a conservative canonical action identity containing
primitive, target, canonical payload or payload digest, and sorted canonical requirements.
Prove that materially different actions do not contaminate each other. Later compression
may earn equivalence between action identities; accidental omission must not supply it.

## Corrected Phase-3 status

The settled result is narrower than the previous documentation claimed:

**Settled:** active-region geometry can be computed before predicate fitting; candidate
completion outcomes are excluded; recurrence modulates the valley; the equal-distribution
control remains unresolved; insertion order does not choose a winner.

**Implemented but untrusted:** exact-feature predicate fitting and unique-current-match
selection. These demonstrate a provisional mechanism, not an earned interpretation.

**Blocked:** allowing `SelectedByPocket` to create or override a confident operational
target until disjoint evaluation, multiple-search-aware support, and durable code recall
pass. Prediction-time feature ownership and conservative action identity are now rails.

The safest implementation sequence is:

1. ~~prevent pocket selection from forcing an operational commit~~ — landed;
2. ~~correct source identity and prediction-time feature ownership~~ — landed;
3. add disjoint evaluation, multiple-search-aware support, and a real promotion gate;
4. add durable learned-code retrieval across episode replacement;
5. rerun Toy #8 with the remaining promotion and episode-transition fixtures before
   restoring commits.

## Bottom line

Phase 3 found a possible geometry. It has not yet shown that the name selected from that
geometry predicts a branch out of sample or survives world replacement as retrievable
learned structure. Prediction-time availability and conservative intervention identity are
now enforced. Geometry remains evidence for differentiation; trusted interpretation still
requires disjoint promotion and durable recall.

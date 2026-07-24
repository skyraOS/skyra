# The Observe Wave — Spec

**Date:** 2026-07-20
**Status:** SPEC (draft). Mechanism paper-validated (tests #1–#4); segmentation
substrate landed (`segment.go`); this document specifies the wave itself for
implementation in v3.

> **Revised 2026-07-20:** the class-mediated form in this document is superseded
> as the active direction by
> [span-owned-predictions-and-activation](span-owned-predictions-and-activation.md).
> Spans now own expressions/predictions directly; entity/address/triple posting
> lists heat spans without a `Class` layer; short-lived EMA activation over
> entities, spans, and causal edges scopes grading. The implementation here is
> retained as prototype history until the direct-span replacement lands.

**Position:** the mechanism inside `observation → state change` — the slot
[retrieval.md](../../design-docs/skyra-dna-v3/retrieval.md) reserved ("the one arrow
with the entire cognitive act in it"). It is NOT a third kernel wave:
[the-wave.md](../../design-docs/skyra-dna-v3/the-wave.md)'s planning wave keeps waves
1 and 2; the observe wave is the upgrade to their FREE index steps (plan-loop steps
1, 2, 5) plus the resolution commit at ingest.
**Derived from:** [observe-mechanism-over-spine](../2026-07-19/observe-mechanism-over-spine.md)
and the four paper tests; ground decisions
[episode-as-region](episode-as-region.md), [settlement-set](settlement-set.md).

## One-line version

> An impulse's references resolve by set operations over class-grouped precedent —
> roots heat classes, diversity ranks them, statistics select bindings, coverage and
> joint precedent gate the commit — and the model appears only at the two bookends.

## Contract

```text
COSTS       zero inference calls. index lookups + arithmetic, like every other
            set operation in v3. budget unit: index operations, bounded by lens size.
READS       observation log, Spans/Classes/Lens (derived, rebuildable), edge EMAs.
WRITES      one state-change node per collapse (the resolution), interpreted
            provenance. NOTHING else durable. per-wave activation is transient
            and never stored.
MODEL       decomposer at ingest (already exists); the planning wave's framing
            call downstream (already exists). the wave itself: LAW 4 clean.
```

## The algorithm

### 1. Roots

A **root** is one independently resolvable reference in the decomposed impulse.
Three kinds, no special machinery for any:

```text
entity roots    subjects/objects of the impulse's triples
verb roots      predicates (paper test #3: dropping "back" flips the outcome —
                relational tokens are load-bearing)
context roots   relevant CURRENT world state (the fold), e.g. the being's location.
                witnessed context participates identically to linguistic roots.
```

Each root proposes candidates: world-given identities cost no coreference and are
singular; model-named identities may be plural (the decomposer's interpreted triples —
superposition lives within a root). Candidate priors are the decomposer's stated
confidences; equal split when it declines to say.

### 2. Lens-hop, conservation

Each root carries amplitude 1.0, split across its candidates, then across each
candidate's classes via `Lens` posting lists. Ubiquitous terms are absent from the
lens by construction (weight 0), so they cannot hand every impulse every class.

### 3. Rank classes: diversity first, heat second

```text
rank by  (distinct roots that heated the class,  summed amplitude)
```

A root's candidates are alternatives — they can pile heat but never corroborate.
Corroboration is only ever across roots (test #2's tie: 0.75 from two roots beats
0.75 from one root's own two candidates). Heat is the tie-break, not the signal.

### 4. Select bindings: class statistics

Within the ranked class, each candidate binding is selected by its counted record —
**edge EMAs grouped by class** (P, N per binding; recency F per `RetrievalWeight`).
This is the Born-rule row with v3's existing learned object doing the counting;
nothing new is learned here, only grouped.

### 5. Coverage gate and joint precedent

```text
COVERAGE    the winning class must instantiate EVERY root. if not, bindings
            stitch from the next classes down — legal, but:
JOINT       the committed binding SET must be checked for co-occurrence in some
PRECEDENT   member span. count > 0: constructive — attach the count.
            count = 0: a NOVEL STITCH — commit is allowed (first times must be
            committable) but FLAGGED. the flag is the express-side signal to
            prefer a clarifying act over confident execution.
```

This is the phase slot, settled (test #3): compatibility = witnessed joint
co-occurrence at commit time. No signed amplitude anywhere — retired after three
tests found nothing to cancel.

### 6. Descend to the exemplar

Only the winning class's exemplar span (most recent settled member) descends into
the present, inside wave 2's token budget. Classes are read where the statistics
live; spans are read only for particulars. Descent is not a conservation sink.

### 7. Collapse: the commit

The resolution commits as a **state-change node**: the chosen bindings, interpreted
provenance, confidence from step 4's statistics, `DerivedFrom` pointing at the raw
impulse triples, the joint-precedent count and any novelty flag attached.

This closes an open integration point deliberately: **the resolution IS the state
change** — v3's loop already has the slot (`observation → state change`), and the
commit is a claim like any other. The tabled decomposer feedback loop
(observation.go) gets its grading path here for free: a later contradiction (bring
the water; "no, I meant the report") grades the resolution's edge —
P(reading correct | impulse class) — by ordinary settlement. Confusion-then-
correction is retained as precedent because it is simply more spine.

## What the planning wave sees (the second integration point)

The observe wave does not hand workers a heat map. Its entire output is:

```text
to the plan loop's INDEX/PRESENT steps:
  - ranked classes with their statistics        (precedent, better-ranked than
                                                 flat posting lists: diversity +
                                                 heat + RetrievalWeight)
  - the exemplar span excerpt                   (within token budget)
to the framing call's present, verbatim:
  - the committed bindings and their confidence
  - the novelty flag / joint-precedent count    (the model SEES "this combination
                                                 is unprecedented" as a fact in
                                                 its present — what it does with
                                                 that is the framing judgment)
to nobody:
  - per-wave amplitudes. transient, disposable, recomputed next impulse.
```

The kill condition, budgets, and the goroutine lifecycle of the planning wave are
untouched.

## Code deltas (in order)

**Deltas 1–5 landed 2026-07-20** in `src/skyra/cognitive-agent/skyra-dna-v3/wave.go` (+`wave_test.go`,
all against the paper-toy fixture):

1. ~~ClassStats~~ → `ExpressionGraph.ClassEdges(log, spans, classes)`: edges grouped
   by class, joined **through evidence** — `ResolvedBy` → observation → span → class.
2. ~~Roots~~ → `Roots(impulse, context, rx)` with `ResolutionIndex` /
   `BuildResolutionIndex` as the tier-1 proposer (counted priors from `resolved-to`
   commits; first contact stands as its own reading).
3. ~~Heat~~ → `HeatClasses`: conservation lens-hop, ranked (distinct roots, heat).
4. ~~Select/JointPrecedent~~ → `Resolve` with the `Stats` Born-rule hook (measured
   readings outrank unmeasured; the no-stats tie is deliberately documented in
   `TestBornRuleSelectsGarage` as the wrong answer the hook exists to fix),
   coverage-gated stitching, joint-precedent count, novelty flag.
5. ~~Collapse~~ → `Resolution.Commit`: interpreted `resolved-to` observations,
   counted confidence, `DerivedFrom` the raw evidence — and the test closes the
   loop: commits feed the next `BuildResolutionIndex`.
6. ~~The seam~~ — **landed 2026-07-20** as `ingest.go`: `Being.ObserveImpulse`
   (impulse written raw → derived substrate rebuilt → wave → readings committed via
   `ObserveInterpreted` → target read off the primary class's measured edges, P ≥
   0.5, N > 0). Novel/unmeasured → `Target == nil` and authoring stays with the
   caller's framing call, per the ground below. `Run(target)` untouched. The tests
   caught and fixed two ground violations in `Resolve`: joint precedent must skip
   OPEN spans (the impulse's own span is trivially self-precedent) and identity
   first-contact bindings (a term standing as itself is not a reading).

## Ground: targets are read, not written (delta-6 shape, decided 2026-07-20)

*Do we need a target?* Architecturally yes — irreducibly. Two standing arguments
close it: region-as-playbook's *error is created by the act of targeting* (no target
→ no error → no action, ever), and the-wave.md's *attention is a goal* (perception
only happens because a standing target requires perceiving). Kill targets and both
action and perception die.

But **target-as-input is a harness artifact.** `Run(target)` takes the target from
`cmd/dnav3` — the experimenter playing the role of desire. v3's own vocabulary
already names the fix: a state change is "a revision to current **or target**
state." The target is interpretation's *output*, never a parameter.

And the wave makes most target-derivation free: the winning class's member spans
contain the settled trajectory — the acts that followed, the expects that resolved
true, graded by edges. **The class's high-P settled end-state is what this kind of
situation completes to.** A target read off precedent, counted, not authored:

```text
familiar impulse   wave resolves → class completion = target        FREE
novel impulse      wave resolves, novelty flag / low P →
                   the FRAMING CALL authors the target; its commit
                   founds the precedent the next one reads          MODEL
```

Cost scales with surprise a third time — for *wanting*, after reading and planning.

The one underivable target is the seed: the standing purpose that bootstraps
perception before any precedent exists (the-wave.md's genome slot, `~purpose` as
the bootstrap target state). One seeded standing target; everything else read.

So the ingest seam is:

```text
impulse observed → wave resolves → target derived
                   (class completion; framing call on novelty)
                 → Run(target)          ← unchanged. the skeleton's lower half
                                          survives; only the SUPPLIER changes —
                                          from the harness to the being's history.
```

This is also traversal-as-excitation's "current and target as derived roles inside
one field" given mechanics: the directional side of the pattern is literally the
settled tail of the hot class.

## Non-goals

- No signed amplitude, no cancellation, no stored heat, no LLM in the middle.
- No hierarchy of classes yet (session-tier grouping is real but parked with
  closure-duration-objective).
- No reflex path (the-wave.md defers it; the novelty flag is this spec's only
  express-side opinion).

## Open ledger

1. **Candidate proposal quality** — RESOLVED IN DIRECTION (2026-07-20): compression,
   not the model. This was the parent compression thread's target all along, and it
   is [parser-compilation](../2026-07-14/parser-compilation.md)'s observation-side
   parser applied at ingest. Two tiers:

   ```text
   tier 1 (FREE)   resolution history as posting lists: candidates = past committed
                   readings of this surface form; priors = resolution-edge EMAs
                   (counted, never equal-split). Generalization to novel-but-
                   analogous forms by slot-filling from class signatures.
   tier 2 (MODEL)  cache miss only — first contact with a reference. The
                   decomposer's reading founds the new resolution edge; the
                   reference never costs inference again.
   ```

   Misfires are handled by the loop that already exists: wrong proposal → committed →
   contradicted → correction retained → edge P drops → low P is the consult-the-model
   signal. Cost scales with surprise at ingest, same as planning. Remaining open:
   candidate *extraction* (payload → triples for familiar sentence shapes) stays with
   the decomposer until the observation-side parser compiles — the long road
   parser-compilation scopes.
2. **Context-root selection** — which of the world fold's triples count as roots for
   a given impulse (all of them is too many; the being's location clearly is one).
   Lean: the standing attention target's requires define the context set. Undecided.
3. **Exemplar choice** — most-recent-settled vs highest-graded member. Working rule:
   recency (matches F-decay's bias); revisit with data.
4. **Class-stat cold start** — an edge with N=0 has no opinion; a class whose every
   edge is unmeasured should rank below a measured competitor at equal
   diversity+heat. Confirm `prior()`'s existing handling suffices once grouped.

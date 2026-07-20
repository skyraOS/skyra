# Span-Owned Predictions and Activation

**Date:** 2026-07-20
**Status:** IMPLEMENTING — continuous replacement is live in v3; real
conversation/decomposer endurance remains
**Revises:** [the-observe-wave](the-observe-wave.md), specifically its class layer,
class-grouped statistics, and class-derived targets
**Keeps:** observation-first world state, prediction-diff grading, exact reverse
indexes, causal edges, settlement, and the expression grammar

## One-line version

> An expression and every prediction it emits belong to the span that birthed
> them; an incoming impulse creates a short-lived activation field over entities,
> triples, spans, and causal edges; only armed predictions owned by active spans
> are eligible to be graded.

## Why this revision exists

The walking skeleton's grader currently asks the expression graph for every open
prediction and matches them globally by `(subject, predicate)`. That loses the
episode in which the prediction was made. Two unrelated spans can carry claims
about the same address, and one incoming observation can settle both even though
only one was active.

The observe-wave prototype introduced `Class` to group similar spans and made
classes the unit heated by an impulse. That is more ontology than the loop needs.
Entity and triple posting lists can retrieve spans directly. Repeated causal
structure already accumulates on the learned edges. A clustering object between
those two mechanisms duplicates responsibility and makes a transitive similarity
decision before the causal record has earned one.

The correction is:

```text
span owns expression
expression emits predictions
actions arm predictions
impulse heats entities/triples
reverse indexes heat spans
causal propagation extends the active region
only active-span, causally armed predictions may grade
```

There is no previous-focus log. Continuity is a mutable, transient activation
field with short decay.

## Vocabulary

```text
SPAN          one witnessed occurrence. An external impulse starts it; actions,
              targeted perceptions, and observations causally descended from that
              impulse retain its SpanID even when they arrive late.

EXPRESSION    one planner emission: framing + acts + predictions. It is born into
              exactly one span.

PREDICTION    an `Expect` at emission time: a falsifiable claim attributed to one
              or more acts, with confidence and a maturity.

BET           the same prediction after every responsible act has completed. It
              is armed, has an ArmedAt, and now owes the world an observation.

ACTIVE        the prediction's owning span currently has activation above the
              eligibility floor.

ELIGIBLE      active + armed + unresolved + observation occurred after arming.

RESOLVED      one eligible observation matched the prediction's address, produced
              TruthTrue or TruthFalse, and was recorded in ResolvedBy.
```

Do not collapse `armed` and `active`:

```text
armed     is causal readiness: did the responsible action set happen?
active    is attention: did this impulse recruit the prediction's birth region?
eligible  requires both.
```

An impulse may recall a prediction before its cause has happened. Activation alone
must never grant causal credit.

## Existing expression/prediction wiring

The lower v3 loop already emits predictions with expressions:

```text
Planner
  → <act ...>
  → <expect of="..."> confidence within subject predicate object </expect>
  → Parse
  → Validate: every act must project at least one expect
  → Affordances stamp measurement mode
  → ExpressionGraph.Place
  → action fires
  → expect arms
  → observation settles it
```

This remains. `Expect` is the code type for a prediction/bet; no parallel
`Prediction` type is introduced.

The missing wiring is ownership and activation-scoped eligibility.

## Durable identity and ownership

### Observation identity

Every observation needs a stable identity independent of its triple and tick:

```go
type Observation struct {
    ID        ObservationID
    EpisodeID EpisodeID
    SpanID    SpanID

    Triple
    Provenance Provenance
    At         int
    Wall       int64
}
```

Suggested identity:

```text
episode:<n>/observation:<sequence>
```

`Triple.Key() + tick` is not an identity. Several observations may legitimately
carry the same triple at the same tick.

### Span identity

A span has a stable ID. It is not an index into a recomputed `[]Span`.

```text
SpanID = EpisodeID + first external observation ID
```

An independent external impulse starts a new span. Runtime-generated descendants
carry the originating `SpanID`:

```text
external impulse I
  → planner expression E
  → act A
  → world outcome O
  → matured perceive P
  → settlement observation S

I, E, A, O, P, and S retain one SpanID.
```

Wall-clock renewal gaps may propose a boundary for external input. They do not
detach a late causal descendant from its originating span.

### Expression ownership

```go
type Expression struct {
    ID        string
    EpisodeID EpisodeID
    SpanID    SpanID
    BornFrom  []ObservationID

    Framing string
    Acts    []Act
    Expects []Expect
}
```

Every planning pass caused by one impulse receives the same `SpanID`. Every
prediction in that expression inherits the expression's ownership.

The ownership query is exact:

```text
bet → expression → SpanID
```

It is not reconstructed from the observation that eventually settled the bet.
`ResolvedBy` answers where the evidence came from; `SpanID` answers where the bet
was born. Both are required.

## Classes are removed from the active mechanism

`Class` attempted to represent a reusable type of episode by grouping
non-contiguous spans with baseline-weighted shared terms.

The active design does not carry this object:

```text
OLD   entity → class → member spans → class-grouped edges
NEW   entity/triple → spans → expressions/causal edges
```

The jobs previously assigned to classes have existing owners:

| Former class job | New owner |
|---|---|
| Find similar prior occurrences | entity/address/triple posting lists over spans |
| Accumulate repeated outcome statistics | causal edge `P`, `N` |
| Prefer convergent precedent | distinct-root heat accumulated on spans |
| Hold short-term relevance | transient activation field |
| Derive likely completion | activated high-confidence causal edges |

This avoids similarity thresholds, transitive union, provisional class
membership, and a second statistics container.

### The real reason: lifecycle, not redundancy

Redundancy is the tidy argument. The load-bearing one is that a class has no
eviction story, and the problem is unbounded — and it gets catastrophically worse
the moment we move off triples.

The chain:

```text
1  the settlement guarantee is what keeps unresolved predictions from piling up:
   Within matures a bet, fires a targeted perceive, exact-match resolves it.
2  that guarantee is TRIPLE-SHAPED — it works only because the expect names its
   own determinate address, so the perceive is fully determined.
3  off clean subject/predicate/object, the target is no longer a determinate
   address, the perceive is no longer determined, and the bet cannot be cleanly
   settled. It sits at TruthUnknown. This is not a bug to fix; it is the direct
   consequence of a richer representation.
4  so a growing pool of never-answered predictions is inevitable. The only
   question is WHERE they live.
```

A class is the worst possible home for them:

```text
CLASS   durable + defined by similarity. Similarity never expires, so a class
        never evicts. Every impulse that heats the class re-exposes the entire
        accumulated pool of unanswerable bets to the grader. Monotonic accretion,
        no forgetting. A dumpster zone.

SPAN    a bounded occurrence, gated by activation. It closes; it decays out of
        relevance (invariant 7 — an inactive span's predictions are invisible to
        the grader); it is a finite unit that span-level GC can archive. Dead bets
        in a cold span cost nothing until collected.
```

The deepest form, and the one that generalizes past triples:

> **Ownership by provenance survives representation change; ownership by
> similarity does not.** A span is "this impulse plus everything causally
> descended from it" — a definition that never mentions what the content looks
> like, so it holds whether a prediction is a triple, a payload, or an embedding.
> Class membership is a function of the content representation, and we are
> deliberately making that representation fuzzier. Hanging lifecycle-critical
> ownership on a content-similarity structure means the ownership degrades exactly
> as fast as the representation does — backwards from what is wanted. A prediction
> must be owned by the EVENT THAT MADE IT, not by what it happens to resemble.

Two consequences worth stating:

- This is why invariant 7 is load-bearing, not merely tidy: activation gating is
  the eviction mechanism the class never had, and it is what makes span-level GC
  tractable (a fully-decayed span whose bets are all past-horizon and unresolved
  can be archived) — because the unit is bounded, which a similarity-type is not.
- The one capability a class genuinely offered — fuzzy "this is like that"
  generalization — is exactly the thing that does not scale and must not be
  durable in the grading path. It belongs to the framing call on novelty now, and
  to learned similarity later; never to a persistent clustering object the grader
  has to walk.

The existing `Class`, `Classes`, `Lens`, class-based wave, and tests remain useful
as prototype history until the replacement is built and compared. They are not
load-bearing forward architecture.

## Direct reverse indexes

All indexes are derived, rebuildable, and exact.

### Observation/span postings

```go
type SpanIndex struct {
    ByEntity  map[string][]SpanID
    ByAddress map[Address][]SpanID
    ByTriple  map[TripleKey][]SpanID
}
```

```text
ByEntity   broad contextual recall: thing:bottle
ByAddress  same world variable:      (build:/, status)
ByTriple   exact precedent:           build:/ status ok
```

Subjects, predicates, and objects may all be roots. A relational token is
load-bearing and is indexed like an entity; it is not reduced to metadata.

### Prediction postings

```go
type BetIndex struct {
    ByAddress map[Address][]BetRef
    ByTriple  map[TripleKey][]BetRef
}
```

The grader begins from `ByAddress[observation.Address()]`, never
`ExpressionGraph.Open()` over all memory.

### Planning precedent

The existing expression reverse index remains:

```text
expected Triple.Key() → expression IDs that projected it
```

It answers: *How have I made this state true before?*

These are three different questions and must remain separate:

```text
SpanIndex   Where has this entity/state been witnessed?
BetIndex    Which predictions could this observation settle?
Expr index  Which prior expressions produced a desired state?
```

## The activation field

Activation is transient attention, not durable belief:

```go
type ActivationField struct {
    Tick int

    Entities map[string]float64
    Triples  map[TripleKey]float64
    Spans    map[SpanID]float64
    Edges    map[EdgeKey]float64
}
```

Durable memory:

```text
observation log
expressions
edge P and N
```

Transient state:

```text
entity/triple/span/edge activation A
```

Thinking about an edge may raise `A`. It must never raise `P`.

### Decay and bump

At each external impulse:

```text
A_t(x) = min(1, λ A_{t-1}(x) + H_t(x))
```

Values below `ε` are deleted.

Working initial parameters, not laws:

```text
λ = 0.5
ε = 0.05
```

A single full activation then decays approximately:

```text
1.0 → 0.5 → 0.25 → 0.125 → 0.0625 → gone
```

Reactivation sustains a region. Several regions may remain warm concurrently.
There is no append-only focus record and no one-impulse wraparound object.

The exact decay curve is an experimental parameter. Replay records the parameter
set and the activation bumps, not every derived decay value.

## Heating spans

The incoming impulse creates independent roots. Each root bumps its entity/triple
activation and reads the direct posting lists.

```text
thing:bottle    ─┐
fridge:kitchen  ├─→ span S12
put-back        ─┘
```

Span activation accumulates by distinct root. Alternative candidates within one
ambiguous root divide that root's heat; they do not corroborate one another.

Conceptually:

```text
for each root:
    divide root amplitude across its candidates
    divide candidate amplitude across its posting-list spans
    bump each span
    record that this distinct root reached the span

rank spans by:
    distinct root count
    then total heat
```

This retains the useful convergence rule from the class prototype without
creating classes.

## Causal propagation

Heated spans expose the expressions born in them. Activation propagates through
their learned causal edges:

```text
span
  → expression
  → responsible act set
  → learned edge
  → expected/observed world-state triple
  → posting-list spans containing that state
```

Crossing strength is gated by measured evidence:

```text
crossing =
    source activation
    × edge.P
    × confidence(edge.N)
```

The crossing bumps:

```text
edge activation
target triple activation
spans reached through the target triple's postings
```

Cycles do not run indefinitely. One impulse has an activation budget and a
visited `(source, edge, target)` set. The first implementation should use a small
deterministic hop cap; the future kernel replaces the flat walk only when data
shows the need.

## Prediction eligibility and grading

### Arming

An individual prediction arms after its responsible act completes.

A joint prediction:

```xml
<expect of="a1 a2">...</expect>
```

arms only after **all** responsible acts complete. `ArmedAt` is the completion
tick of the last source act:

```go
type Expect struct {
    Of []string
    Triple

    Within   int
    ArmedAt  int
    MaturesAt int

    Resolution Truth
    ResolvedBy []ObservationID
}
```

### Candidate selection

For each incoming observation:

```text
candidates = BetIndex.ByAddress[observation.Address()]
```

Then:

```text
eligible(bet, observation) =
    activation.Spans[bet.Expression.SpanID] ≥ activationFloor
    AND every responsible act completed
    AND bet.ArmedAt > 0
    AND observation.At ≥ bet.ArmedAt
    AND bet.Resolution == TruthUnknown
    AND observation provenance may settle this claim
```

Only eligible bets reach the existing exact object comparison and proper score.

### Fidelity

Evidence fidelity caps the grade:

```text
raw           1.0
interpreted   0.3
mark          0.0
```

The effective grade weight cannot exceed either the claim channel or evidence
fidelity. An interpreted observation never settles an exact prediction at full
authority.

Marks may activate their own mark-addressed region. They may not satisfy or grade
an ordinary world-state claim.

## Settlement

A matured bet emits a targeted perceive. The request carries its ownership:

```go
type PerceiveRequest struct {
    Triple
    BetID     BetID
    OwnerSpan SpanID
}
```

When its observation returns, the owner span is explicitly reactivated before
eligibility is checked:

```text
bet obliges perceive
  → perceive carries OwnerSpan
  → return bumps OwnerSpan
  → eligible bet resolves
```

This is not an exception to activation-scoped grading. Settlement is itself a
causal continuation of the span and therefore legitimately reactivates it.

Settlement outlives the plan. It does not outlive the bet's ownership.

## Required write/grade separation

Current `Being.Observe` writes and immediately calls the global grader. That
ordering cannot support activation-scoped eligibility.

The phases become explicit:

```go
written := b.AppendObservations(decomposed)
span := b.AssignSpan(written, cause)

b.Activation.Decay()
b.ActivateImpulse(span, written)
b.PropagateActivation()

eligible := b.EligibleBets(written)
b.Grade(written, eligible)

target := b.DeriveTarget(...)
b.Plan(span, target)
```

New observations are appended before grading. New predictions are emitted after
grading. Therefore an impulse cannot settle a prediction that did not exist when
the impulse arrived.

World outcomes and settlement returns use the same phases but carry their causal
`SpanID`; they do not start independent spans.

## Target derivation without classes

The class prototype read a target from the winning class's measured completions.
Without classes, target derivation reads the activated causal field:

```text
heated spans
  → reliable outgoing causal edges
  → high-activation expected world states
  → candidate target
```

Candidate targets require:

```text
edge.N > 0
edge.P above reliability floor
target activation above target floor
joint support when several roots contributed
```

If no measured causal continuation survives, the situation is novel and the
framing call authors a target. Its expression and predictions belong to the
current span; their eventual results found future precedent.

## The loop

```text
raw impulse
    ↓
decompose into observations
    ↓
append observations without grading
    ↓
assign observations to a stable span
    ↓
decay entity/triple/span/causal activation
    ↓
bump entities and triples from incoming observations
    ↓
reverse index activated entities/triples → spans
    ↓
heat spans by distinct-root convergence
    ↓
propagate through reliable causal edges
    ↓
prediction-address reverse lookup
    ↓
filter to active-span, causally armed predictions
    ↓
grade against incoming observations
    ↓
derive or update target from activated causal continuation
    ↓
planner emits Expression + predictions
    ↓
stamp Expression with EpisodeID, SpanID, BornFrom
    ↓
execute closed plan
    ↓
world outcomes retain SpanID and re-enter the same loop
```

## Relationship to edge recency `F`

The current edge retrieval weight is:

```text
P × confidence(N) × F
```

Activation now owns *relevance right now*. `P` and `N` own historical
reliability. `F` may duplicate activation.

Do not remove `F` in the first landing. Instrument:

```text
rank with P × confidence(N) × F
rank with activation gating + P × confidence(N)
```

If `F` contributes no independent predictive value, delete it. There should not
be two unexplained short-decay systems.

## Invariants

1. Every action is covered by at least one prediction in the expression that
   emits it.
2. Every expression and prediction has exactly one birth `SpanID`.
3. Every resolved prediction names stable observation evidence.
4. An unarmed prediction cannot resolve.
5. A joint prediction cannot arm until every responsible act completes.
6. An observation cannot settle a prediction born or armed after it.
7. An inactive span's predictions are invisible to the grader.
8. A targeted settlement return reactivates its owning span.
9. Activation may change retrieval and eligibility; it may never change edge
   truth statistics.
10. `TruthUnknown` never updates an edge.
11. Marks never satisfy or grade ordinary world-state claims.
12. Interpreted evidence never carries raw fidelity.
13. No class object is required for retrieval, aggregation, activation, or target
    derivation.

## Landing order

### Landed 2026-07-20 — bounded milestone 1

The first implementation slice is in `src/skyra/skyra-dna-v3/`:

- Stable episode, span, and observation identities are stamped on runtime writes.
- Expressions carry `EpisodeID`, birth `SpanID`, and `BornFrom`.
- Expression-by-span and prediction-by-address indexes are exact and deterministic.
- Per-act completion is durable; joint predictions arm only after the final source
  act completes.
- World outcomes are appended after completion/arming and retain the expression's
  birth span.
- Impulses append their raw observation batch before grading.
- The grader reads the address index and admits only armed, unresolved predictions
  whose birth span matches the active observation span.
- Settlement returns retain and reactivate the bet owner's span.
- Stable observation IDs replace triple-plus-tick evidence IDs, with legacy lookup
  retained for replayed fixtures.
- Evidence provenance caps grade weight.

The current activation set is intentionally minimal: the causal span carried by the
incoming observation. There is no EMA persistence or causal recruitment yet. The
class-mediated wave remains as prototype code until phase 4 establishes direct-span
parity.

### Stress validation 2026-07-20

Milestone 1 was checked against an independent deterministic eligibility model:

- Five fixed random seeds, each spanning three episodes, 360 predictions, and
  5,400 adversarial events.
- Random one-to-three-act joint completion order, repeated completions, raw,
  interpreted, and mark evidence, repeated same-address observations, old-span
  labels after episode replacement, and same-tick identity pressure.
- Twenty-five repeated passes of the deterministic stress group.
- Two native Go fuzz runs totaling 546,293 generated executions, including
  320,158 after the final fix, with no remaining mismatch.
- A delayed-settlement flood: 1,000 predictions across 200 owner spans.
- Race detector, full repository tests, and `go vet`.

Two real defects were found and fixed:

1. **Cross-episode contamination.** The grader compared `SpanID` but not
   `EpisodeID`. A new-world observation carrying an old span label could settle an
   old-world prediction. Eligibility now requires both identities, and settlement
   never perceives an old bet against a replacement `Realm`.
2. **Mark laundering through world state.** Marks could not grade, but
   `ObservationGraph.Write` still folded them into current state, allowing a mark
   to satisfy guards/targets and overwrite raw evidence. Marks now remain in the
   durable log but are excluded from the evidence-bearing current-state fold.

The prediction-address benchmark held grading at approximately 312–370 ns/op
with 0, 1,000, and 100,000 unrelated open predictions, with the same four
allocations per operation. This supports the intended complexity boundary:
unrelated bets do not enter the grading candidate set.

This validates mechanical ownership and eligibility, not semantic span policy.
Still untested by construction: whether conversational continuation should reuse a
span, direct-span retrieval parity after class removal, EMA decay/floor values,
causal propagation, concurrent mutation (the runtime is intentionally
single-threaded), persistent identity across process restart, and lifecycle/GC for
old unresolved predictions.

### Landed 2026-07-20 — continuous replacement

There is no accumulated model conversation and no compaction operation in the
v3 inference path:

```text
durable observation/expression history
  → decay transient activation once per external impulse
  → bump incoming observation nodes/entities/triples/current span
  → direct entity/address/triple postings retrieve spans
  → measured causal edges propagate activation
  → every renderable observation/expression has a token cost
  → fixed-budget selection ranks memory nodes by activation/token
  → present is materialized anew
```

Implementation:

- `activation_field.go`: entity, triple, span, edge, and memory-node activation;
  `A_t=min(1, λA_{t-1}+H_t)`, epsilon pruning, bounded causal propagation,
  active-span grading, and activated target derivation.
- `memory.go`: incremental direct `SpanIndex`, token-cost stamps, renderable
  `MemoryNode`, deterministic fixed-budget selection, and `A/token` density.
- `present.go`: no full-log scan. The planner sees a newly materialized active
  projection under `ActivationConfig.MemoryTokens`.
- `ingest.go`: the class layer is removed from the active ingest path. The class
  implementation remains only as prototype/paper-test history.
- `cmd/dnav3` and `cmd/globe`: live harness impulses now enter through
  `ObserveImpulse` before planning.

Token count is a provider seam. The default is a conservative Unicode-aware
approximation (`ceil(runes/3)`); inference providers may inject their exact
tokenizer through `Being.CountTokens`. Durable observations and expressions stamp
their base count, and each rendered `MemoryNode` counts its concrete presentation
again so labels/statistics cannot silently escape the budget.

The current impulse is the traversal root and its nodes are marked mandatory;
the remaining field competes by `activation/token`. Broad posting lists whose
per-span heat cannot clear epsilon are not scanned—the direct-index form of the
fan/promiscuity discount.

An automated endurance test stores more than 100,000 estimated tokens across
4,000 spans while holding the active projection to 256 tokens. The first memory
decays out, remains absent, then returns through entity reactivation without a
summary or compaction step. The continuous-replacement changes also passed 84,057
new fuzz executions, full/race/vet checks, ten repeated focused replacement runs,
and a 25-pass combined stress/endurance run.

This proves the substrate mechanism, not yet graceful live conversation. The
remaining v1 gate is a real decomposer/conversation source with interleaved
threads, corrections, interruption/resumption, and provider-token accounting.

### Phase 1 — identity and ownership

- Stable `EpisodeID`, `ObservationID`, and `SpanID`.
- Stamp observations, expressions, and predictions.
- Add expression-by-span and prediction-by-address indexes.
- Tests for duplicate same-triple/same-tick observations.

### Phase 2 — causal readiness

- Add per-act completion records.
- Add `ArmedAt`.
- Arm joint predictions after the last responsible act.
- Reject observations older than the arm.

### Phase 3 — split append from grade

- Observation append becomes side-effect-free with respect to grading.
- Batch activation and eligibility precede `Grade`.
- Preserve the existing arithmetic grader once candidates are selected.

### Phase 4 — direct span indexes

- Build entity/address/triple → span postings.
- Heat spans by distinct-root convergence.
- Remove classes from `ObserveImpulse`.
- Compare direct-span results against the four existing paper fixtures.

### Phase 5 — activation

- Entity, triple, span, and edge activation maps.
- Deterministic decay/bump/prune.
- Bounded causal propagation.
- Activation-scoped bet eligibility.

### Phase 6 — settlement and target

- Settlement carries and reactivates `OwnerSpan`.
- Derive familiar targets from activated reliable causal edges.
- Novel continuation hands target authorship to framing.

### Phase 7 — trace and live source

- Typed append-only trace for activation, eligibility, grade, and edge deltas.
- Wire a real impulse source through the complete ingest loop.
- Run ambiguity, interruption, delayed-return, and nonstationarity trials.

## Required tests before a live grade

- Two spans predict the same address; only the active span resolves.
- An inactive old bet remains open when an unrelated matching observation lands.
- An unarmed bet never resolves.
- A joint bet remains unarmed after only one source action.
- A late settlement return resolves its owner after reactivation.
- Two identical observations at one tick retain different IDs.
- Repeated entity activation sustains a span; unrelated spans decay out.
- Two independent roots outrank two alternatives from one root.
- Causal propagation recruits a span absent from the direct entity postings.
- Activation does not alter `P` or `N`.
- Interpreted evidence caps the grade.
- A mark cannot satisfy a guard or target.
- Removing classes does not break the paper-toy reference resolutions.

## Open ledger

1. **External span boundary:** an independent external impulse starts a span;
   continuation/coreference may keep one open. The exact independence predicate
   needs live traces.
2. **Decay curve:** start with configurable `λ=0.5`, `ε=0.05`; fit from
   interruption/return behavior rather than intuition.
3. **Activation floor:** whether eligibility is any nonzero heat or a threshold.
   Lean: explicit floor equal to prune epsilon.
4. **Propagation budget:** initial deterministic hop and crossing caps.
5. **Concurrent spans:** several spans may stay active; action arbitration remains
   separate from grading eligibility.
6. **Edge `F`:** retain through instrumentation, then keep or delete based on
   independent value.
7. ~~**Prediction coverage:** add validation that every `Expect.Of` references an
   act in the same expression.~~ **Landed 2026-07-20**, including duplicate and
   empty act-ID rejection.

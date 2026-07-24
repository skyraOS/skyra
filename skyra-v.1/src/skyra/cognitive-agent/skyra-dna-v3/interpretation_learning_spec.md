# Interpretation Learning Through Blind Waves

Status: proposal  
Date: 2026-07-20  
Scope: high-level v3 design; no wire format or storage layout is final

> **NEXT TOY TEST:**
> [Genuinely Balanced Ambiguity With No Dominant Signal](../../../specs/2026-07-21/genuinely-balanced-ambiguity.md).
> The fixture removes every earned tie-breaker and requires the being to preserve
> plural uncertainty, reach a query fixed point, and ask for discriminating evidence
> rather than collapse by ordering or self-confirmation.

> **ARCHITECTURAL BOUNDARY:**
> [Interpretation, Prediction, and Induction Are Separate Operations](../../../specs/2026-07-21/interpretation-prediction-induction-boundary.md).
> Abduction proposes a category; deduction computes the consequences licensed by
> that candidate; the world settles those consequences; only then may induction
> revise confidence in the category. Deterministic prediction does not turn its
> uncertain interpretive premise into a fact.

> **CURRENT RESEARCH NOTE:**
> [Two-Axis Inference: Interpretation Grades Through the World, Not Itself](../../../specs/2026-07-21/two-axis-inference.md).
> The note proposes distinct resemblance and consequence edge types, traces backward
> credit through their licensing dependency, and records abductive search and failure
> attribution as unresolved rather than hiding both inside one confidence scalar.

> **LUCKY-HIT CORRECTION:**
> [The Two-Bit Outcome: Interpretation and Prediction Are Separately Falsifiable](../../../specs/2026-07-21/the-two-bit-outcome.md).
> A successful action prediction cannot validate the interpretation that selected it.
> The interpretation must license its own category-discriminating prediction, settled
> independently. [Toy #9](../../../specs/2026-07-21/toy-9-two-address-claim-paths.md)
> now implements the requirement as two typed claim paths and histories sharing one
> `Edge` probability primitive—not two duplicated edge implementations.

> **PREDICATE MECHANISM:**
> [Predicate Diagnosticity: Learning the Consequence Check, Not Hand-Writing It](../../../specs/2026-07-21/predicate-diagnosticity.md).
> The AI may propose tool-rich boolean predicates, but held-out firing behavior and
> base-rate diagnosticity select them; runtime settlement remains deterministic and
> LLM-free. Sparse-data behavior remains open.

> **COMPOSABLE OBSERVATION GRADE:**
> [Scale, Structure, Content](../../../specs/2026-07-21/scale-structure-content.md).
> Token range licenses an observation budget; structure spends it into slots; content
> fills those slots with same-claim-fused predicates. Only materialized, settled
> predicates grade. Predictions owned by absent slots are void rather than false.

## Claim

Interpretation can acquire a learning signal without an LLM judging another LLM.

The being holds a prediction of where the next observation wave will activate,
allows the next impulse to create its own blind activation field, and compares the
two only after the observation field is frozen.

When the initial observation wave is insufficient to interpret an impulse, the LLM
may issue queries that traverse and heat the existing graph. Those queries remain
inside the impulse's interpretation session. They retrieve memory but do not write
world state and do not count as evidence.

The design produces two learning curves:

```text
queries-to-interpretation-closure  ↓
future interpretation error       ↓
```

Both curves must fall. Fewer queries without lower error is premature confidence.
Lower error without fewer queries is search without learning.

This is the interpretation-side counterpart of v3's existing measurement:

```text
passes-to-plan-closure  ↓
world prediction error ↓
```

## The cycle

```text
raw impulse
    ↓
blind observation wave
    ↓
initial candidates and interpretation residual
    ↓
zero or more graph queries
    ↓
mechanical interpretation closure
    ↓
committed interpretation
    ↓
held expectation field
    ↓
next blind impulse
    ↓
alignment, surprise, and grade over base rate
```

This cycle crosses three different epistemic operations:

```text
"this is a P"  ──abduction──▶  "if P, expect Q"  ──deduction──▶  Q or not-Q
       ▲                                                               │
       └──────────────────── induction after settlement ────────────────┘
```

Interpretation may be uncertain while prediction remains deterministic conditional
on each candidate. The system therefore preserves candidate-specific expectations
instead of forcing an interpretation winner merely to make a prediction.

The LLM participates in proposing queries and interpretations. It does not decide
whether its interpretation was correct, whether closure was reached, or how much a
prediction is worth.

## The three fields

The architecture must keep three kinds of activation distinct.

### Observation field

The activation produced by an incoming impulse while blind to open predictions,
targets, desires, and graph queries that have not yet been measured.

It is the only field that crossed the world boundary. It may contribute evidence
according to the provenance of the observations that produced it.

### Retrieval field

Activation produced by a self-authored query over durable memory.

It answers "what does my history make available?" It does not answer "what is true
now?" Retrieval heat may change attention and candidate availability, but it may not:

- write current world state;
- settle a prediction;
- satisfy a target or action guard;
- overwrite an observation;
- count as independent evidence.

### Expectation field

A held prediction about the activation expected from a future eligible impulse.

It is mine, not the world's. It remains separate from the next observation wave until
that wave has been frozen and measured.

The ordering is:

```text
predict → hold
observe → freeze
compare → grade
then, optionally, combine for attention
```

It must never become:

```text
predict → preheat observation → observe → grade the manufactured agreement
```

## Blindness invariant

The pass that constructs the observation field must not see:

- the expectation field being tested;
- the target or desired answer;
- confidence assigned to open predictions;
- graph-query results produced after the impulse arrived.

It may see stable vocabulary and world-given identity required to speak the graph's
language.

After the blind field is frozen, expectation and retrieval fields may influence
attention, planning, or the selection of a clarifying action. They may not rewrite
the frozen field.

## Interpretation sessions and spans

An external impulse begins one durable span. Interpretation work occurs in a session
owned by that span:

```text
external span
├── raw observations
├── blind field
├── query 1
├── retrieval field 1
├── query 2
├── retrieval field 2
├── committed interpretation
└── expectation field
```

Internal queries do not create new external spans and must not extend or merge span
identity based on model latency. A later external impulse starts its own span even if
the prior interpretation session exhausted its budget without closing.

Every query, returned activation, binding, interpretation, and expectation retains
the owning episode and span IDs.

## Initial exact pass

Every impulse begins with the existing zero-inference machinery:

1. write raw observations;
2. derive spans, classes, lens, and resolution history;
3. extract independently resolvable roots;
4. heat classes with conserved amplitude;
5. select bindings from counted precedent;
6. compute coverage and joint precedent;
7. expose only the unresolved residual.

If this pass reaches interpretation closure, the LLM makes no graph query.

## Graph queries

A graph query is an act of attention over durable memory. The LLM chooses what it
wants to retrieve; the substrate deterministically decides what the query activates.

A query should be able to name:

- unresolved surface roots;
- known graph identities;
- relevant relations;
- temporal or span scope;
- the kind of evidence sought.

The first implementation should prefer a constrained representation, even if a
natural-language surface is later accepted. An unrestricted natural-language query
creates a recursion problem: the system must interpret the query being used to
interpret the impulse.

Example conceptual query:

```text
roots: "it", "over there"
relations: recent-referent, held-object, destination
scope: current episode
```

Embedding or lexical retrieval may propose candidates. Retrieval never asserts that
a candidate is the correct reading.

Queries may be emitted in a batch when independent. Query execution does not require
an inference call.

## Interpretation residual and closure

The LLM does not declare that it understands. The substrate computes closure.

The residual may contain:

- roots with no grounded candidate;
- roots with unresolved competing candidates;
- bindings not covered by the primary class;
- a novel stitch with no joint precedent;
- a committed reading with no falsifiable consequence;
- missing provenance between a reading and its raw evidence.

An interpretation is structurally closed when:

1. every root is bound or explicitly retained as unresolved;
2. every committed binding is traceable to raw evidence and any retrieval used;
3. alternative candidates have not been silently converted into certainty;
4. the joint-precedent or novelty status is recorded;
5. the interpretation produces a falsifiable expectation or a clarifying action.

Novelty does not prevent closure. It reduces confidence and should favor clarification
over irreversible execution.

A fixed-point rail stops the session when another query fails to reduce the residual.
A budget rail stops broad or endless search.

## Cognitive cost

Raw query count is not a sufficient cost measure. One query that heats the entire
graph is not cheaper than two precise queries.

Interpretation cost should eventually account for:

```text
inference passes
+ inference tokens
+ query count
+ nodes traversed
+ activation breadth or entropy
+ latency on the critical path
```

The exact scalar is deferred. The invariant is that broad retrieval consumes more
budget than narrow retrieval.

The cost record must remain attached to the interpretation path so later success can
credit the retrieval edges that made interpretation cheaper.

## Expectation fields

An interpretation may emit a sparse normalized distribution over the graph locations
expected to activate on the next eligible impulse.

The first representation should operate over classes because the observe wave already
produces class heat. Later representations may add:

- candidate bindings per root;
- expected triples;
- explicit novelty or "other" mass;
- timing and eligibility;
- joint expectations from multiple acts.

Conceptually:

```text
ExpectationField {
    source interpretation
    episode and span
    armed-at
    within
    class amplitudes
    optional binding amplitudes
    provenance
}
```

Each expectation remains independently attributable. Multiple active expectations may
form a mixture for attention, but they are graded separately so a good prediction
cannot hide a bad one.

## Interference, alignment, and surprise

Let:

- `Q` be the normalized expectation field;
- `O` be the next normalized blind observation field;
- `B` be the historical base-rate field for the applicable context.

A bounded similarity may be displayed on a `0..100` scale:

```text
0   = disjoint activation
100 = identical distributions
```

Jensen-Shannon similarity or Bhattacharyya overlap are candidate display measures.
The choice is deferred.

Similarity alone must not drive learning. A diffuse prediction can overlap almost
anything. The durable grade should reward information over the base rate:

```text
grade = Σ Oᵢ log(Qᵢ / Bᵢ)
```

with probabilities clamped away from zero. This is a proposal, not yet a committed
formula. It preserves the v3 principle:

> The model earns credit only for information that counted history did not already
> provide.

Expected behavior:

- a precise correct prediction earns positive credit;
- reciting the base-rate field earns approximately zero;
- a broad uninformative prediction earns little or nothing;
- a confident wrong prediction loses credit;
- correctly predicting unusual activation earns more than predicting the ordinary.

The system should retain separate diagnostics for:

- class alignment;
- binding alignment;
- triple alignment;
- novelty surprise.

This localizes whether the being predicted the kind of situation, the meaning of an
ambiguous term, or the resulting world claim.

## What the grade means

Alignment between an expectation and a later interpreted impulse demonstrates
predictive continuity. It does not by itself establish semantic truth.

A resolver can make the same internally consistent mistake twice. Therefore:

- alignment through interpreted observations remains fidelity-capped;
- raw observations retain greater authority;
- self-authored marks and query heat carry zero evidentiary fidelity;
- later contradiction may revise an interpretation;
- discriminating action is preferred when memory cannot resolve ambiguity.

Example:

```text
candidate A: "bank" means financial institution
candidate B: "bank" means river edge

internal retrieval cannot distinguish them
    ↓
ask a clarifying question
    ↓
the answer creates a new blind impulse
    ↓
the competing predictions make different commitments
```

Interpretation becomes strongly learnable when it predicts consequences that the
world can independently distinguish.

## Just-in-time cognition

The query loop makes cognition proportional to unresolved novelty:

```text
first exact pass closes
    → zero interpretation queries

residual remains
    → query only the residual

query does not reduce residual
    → stop or ask the world
```

The intended progression for a repeated ambiguous situation is:

```text
early:   several graph queries and high prediction error
later:   fewer queries and lower prediction error
mature:  exact or counted resolution with zero queries
```

More cognition is justified for a larger or more surprising human impulse. Familiar
structure should return to the fast path.

## Competence crystallization

Interpretation learning and just-in-time tool development are compatible instances
of the same process:

```text
deliberate cognition
→ repeated falsifiable attempts
→ reliable predictions
→ recurring successful structure
→ guarded procedural competence
```

On the interpretation side, repeated query paths may crystallize into direct
resolution edges.

On the motor side, repeated successful action subgraphs may crystallize into a
parameterized tool or competence.

In both cases:

- reliability and prediction error supply selection pressure;
- applicability conditions become guards;
- execution becomes cheaper as the structure consolidates;
- failed guards or surprising outcomes deoptimize back into cognition;
- yesterday's plan may become today's motor primitive.

This spec does not define tool promotion. It establishes the shared learning shape
that future tool development should preserve.

## Required measurements

An experiment must report at least:

```text
queries-to-interpretation-closure
interpretation inference calls
activated nodes or retrieval breadth
future field prediction error
grade over base rate
clarification actions
unresolved or novel interpretations
```

For end-to-end evaluation it should also retain:

```text
time to first useful response or action
time to observed target
total inference tokens
passes-to-plan-closure
world prediction error
```

The architecture succeeds only if interpretation becomes both cheaper and more
accurate across repeated exposures.

## Minimum experiment

The first implementation should be deliberately small:

1. use the existing class-heat output as `O`;
2. add a held class-level expectation field `Q`;
3. allow one constrained graph query inside the current span;
4. keep query activation out of observation state and grading;
5. define one mechanical closure residual;
6. compare `Q` with the next eligible blind field;
7. record query cost and prediction error;
8. replay repeated ambiguous fixtures and measure both curves.

The first fixture is the linked **genuinely balanced ambiguity** test. It must be run
before adding a fixture with a dominant signal: the substrate first has to prove it can
represent honest underdetermination without turning deterministic ordering into belief.

Required tests:

- a correct narrow expectation beats the base rate;
- a uniform expectation earns approximately zero;
- a confident wrong expectation is punished;
- novel support produces surprise;
- expectation heat cannot change the blind observation field;
- query heat cannot satisfy a target, guard, or prediction;
- an expectation resolves only against an eligible future span;
- interpreted evidence remains fidelity-capped;
- multiple expectations retain independent attribution;
- repeated exposure reduces queries without increasing error.

## Deferred questions

- What exact field level should v0 predict: classes, bindings, or both?
- What bounded similarity is most useful for inspection?
- What base-rate context is neither too broad nor too sparse?
- When does a novel stitch close versus require clarification?
- How are unfinished interpretation sessions represented when the next impulse arrives?
- How should query breadth be priced?
- When may a prior expectation influence posterior attention without contaminating the
  committed observation?
- What evidence is sufficient to crystallize a recurring query path into a direct edge?

These are experimental questions. None changes the invariants: observation is blind,
queries are not evidence, expectations are held separately, closure is mechanical, and
learning requires both falling cognitive cost and falling prediction error.

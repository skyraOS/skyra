# Skyra — Span-Owned Activation and Interpretation Learning

**Date:** 2026-07-21  
**Scope:** Changes after the 2026-07-18 project snapshot: commits from July 19–20,
the observe-wave paper walks and live replay, the current dnav3 implementation,
the compression research, and the uncommitted interpretation-learning proposal.  
**Status:** Project snapshot. Span ownership, activation-scoped grading, continuous
memory replacement, and the 3D globe are implemented. The class-mediated observe
wave remains tested prototype history. Learned compression codes and interpretation
expectation fields are research proposals.

## Executive summary

The last two days converted the July 18 field direction into a concrete v3 working-memory
mechanism, but not by keeping the first mechanism that was built.

The work proceeded in four moves:

```text
1. Paper-walk observation-side interpretation over an immutable experience spine.
2. Implement a zero-inference observe wave using spans, classes, and counted precedent.
3. Find that classes are the wrong owner for predictions and remove them from the active path.
4. Replace them with exact span ownership plus a continuously replaced activation field.
```

The current shape is:

```text
external impulse starts a span
    ↓
raw observations are appended
    ↓
the impulse heats entities, triples, memory nodes, and prior spans
    ↓
direct posting lists retrieve witnessed occurrences
    ↓
measured causal edges propagate activation
    ↓
only active-span, causally armed predictions may grade
    ↓
memory is rendered under a fixed token budget by activation/token
    ↓
reliable activated continuations may provide a target
    ↓
novel residual goes to framing/planning
```

This is the first implemented form of the July 18 claim that the active context should be
reconstructed rather than appended. Durable history continues to grow, but the model-facing
present does not. Each impulse decays the old transient field, recruits what matters now, and
materializes a fresh bounded projection. There is no accumulated model conversation and no
compaction pass.

The most important correction is ownership:

> A prediction belongs to the event span that caused it, not to a similarity class and not to
> every open prediction sharing its address.

That rule made causal eligibility structural. An observation can grade a prediction only when
the prediction belongs to the same episode and span, its responsible actions have completed,
and the observation arrived after arming. Targeted settlement explicitly reactivates the owner
span when the delayed observation returns.

The next research edge is interpretation. A new proposal separates blind observation heat,
self-authored retrieval heat, and held expectation heat. The being may query its graph inside an
open interpretation session, but query results are attention rather than evidence. Interpretation
learning would be measured by two curves:

```text
queries-to-interpretation-closure  ↓
future interpretation error       ↓
```

This extends the same just-in-time principle already present in planning: exact machinery first,
model inference only on the unresolved residual, and repeated successful cognition eventually
crystallizing into direct resolution edges or guarded motor competences.

## Review window

The previous summary is
[2026-07-18-traversal-excitation-and-field-representation.md](2026-07-18-traversal-excitation-and-field-representation.md).
The committed review window begins immediately after it.

| Commit | Date | Result |
|---|---|---|
| `42b3f31` | 2026-07-19 | Interactive cube-sphere globe realm and browser harness |
| `dd13202` | 2026-07-19 | Observe mechanism, four paper-test chain, segmentation and settlement grounds |
| `5a49ff6` | 2026-07-19 | Class-mediated observe wave implemented in dnav3 with zero inference |
| `b2c68f9` | 2026-07-20 | Span ownership, causal arming, direct indexes, activation field, bounded memory; classes removed from active ingest |
| `74c33ef` | 2026-07-20 | Research direction: triples as a bootstrap compression codebook |
| `7f4d3b9` | 2026-07-20 | Architecture timeline backfilled through phases 26–34 |

The working tree also contains two relevant uncommitted documents:

- an expanded compression note connecting v3's log score to compression gain and recording
  cross-entropy/entropy research;
- `interpretation_learning_spec.md`, which combines blind waves, graph queries, expectation
  fields, just-in-time cognition, and competence crystallization.

## What the July 18 snapshot left open

The July 18 snapshot established the direction but deliberately did not choose the representation:

```text
interaction + carried structure
    → multi-root traversal
    → overlap/interference pattern
    → bounded field representation
    → collapse/carry and expression
```

It retained several non-negotiable constraints:

- observation is the only writer of durable world state;
- unknown is not false;
- raw evidence is durable while interpretation remains revisable;
- exact resolution precedes inference;
- learning crosses the observation/action boundary;
- the kernel remains ignorant of cognitive meaning.

The immediate open questions were how to represent overlap, how recurrent context stays bounded,
what owns an unresolved prediction, and how stable patterns compress without losing provenance or
falsifiability.

The July 19–20 work answers the ownership and bounded-context questions in code. It advances but
does not close learned representation and interpretation.

## The interactive globe realm

The July 19 globe commit adds an embodied spherical world rather than another cognitive layer.

The surface is a cube-sphere: six square grids projected onto a sphere. Cells retain exact,
world-given addresses while rendering at real 3D positions:

```text
coord:<face>,<x>,<y>
```

Cross-face adjacency is calculated geometrically by projecting the step and selecting the new
dominant face. It does not rely on a manually maintained table of 24 edge transitions. This keeps
movement identity exact while avoiding a class of silent wrap errors.

The realm includes:

- exact placement and identity;
- spherical movement and shortest-path behavior;
- perception and engagement surfaces compatible with the existing world contract;
- a browser-based interactive renderer and game harness;
- dnav3 integration through the generic `Realm` interface.

Architecturally, the globe is useful because it demonstrates that v3's being is not defined by the
filesystem laboratory. A realm supplies acts, outcomes, and perceivable predicates; the cognition
loop remains separate.

## The observe-wave paper walks

The paper-test sequence is worth preserving because each test was constructed to force one open
decision. The important output is not that the toy passed; it is which proposed machinery failed
to earn a place.

### Paper test 1 — the hottest point is not the precedent

The first toy used a 15-moment experience spine containing:

- a successful water-bottle routine using the kitchen fridge;
- a failed kitchen-fridge interpretation corrected to the garage fridge;
- an unrelated plant-watering routine.

The impulse was:

```text
"go get my water from the fridge"
```

Plural candidate readings were allowed for `water` and `fridge`. Episode-level activation tied
the water routine and fridge-fetch history exactly. The single hottest moment was the plant-water
distractor because its entity occurred in only one place and therefore received concentrated heat.

The useful conclusions were:

- isolated heat spikes are fan-in artifacts, not corroboration;
- the useful unit was a span of jointly supported history, not one hot node;
- ubiquitous entities such as the user self-attenuate and carry little information;
- negative settlement did not require signed activation in the middle of the wave;
- root diversity appeared more meaningful than raw summed heat.

The toy resolved `water → bottle` and `fridge → kitchen` because all independent roots converged
on the same successful span.

### Paper test 2 — superposition is within a root; interference is across roots

The spine was extended with another successful garage-fridge fetch and spans were grouped into
hand-assigned episode classes.

For:

```text
"grab something from the fridge"
```

the fridge-fetch class won, and its counted outcomes selected garage over kitchen. The result did
not require negative amplitude: heat selected the relevant precedent region; counted settlement
statistics selected the binding.

For the original water-and-fridge impulse, the water-routine and fridge-fetch classes tied in raw
heat. The distinction was the source of that heat:

```text
water routine:  equal heat from two independent roots
fridge fetch:   equal heat from two alternatives of one root
```

The durable rule became:

> A root's candidates are alternatives and cannot corroborate each other. Rank by distinct-root
> support first and heat second.

### Paper test 3 — context and verbs are roots; phase is joint precedent

The contradiction case placed Michael in the garage holding the blue bottle and used:

```text
"put my water back in the fridge"
```

Without the relational root `back`, garage context heated the fridge-fetch history and produced a
novel cross-class stitch: bottle plus garage fridge, a combination with no witnessed joint
precedent.

With `back` included as a root, the water-return routine gained support from three independent
roots and correctly selected the kitchen fridge.

The test established:

- context participates as an ordinary root;
- relational and verb tokens can be more decisive than entity nouns;
- diversity, coverage, and joint precedent are separate checks;
- a winning region that cannot cover every root must stitch bindings from elsewhere;
- a zero-precedent stitch is legal but must be marked novel;
- signed cancellation in the middle remained unnecessary and was retired.

The field analogy's open `phase` slot became a collapse-time fact:

```text
compatible bindings = witnessed joint co-occurrence in a prior span
zero joint precedent = novel stitch, prefer clarification over confidence
```

### Paper test 4 — segmentation is recoverable mechanically

The final toy removed all episode labels and added witnessed timestamps. It recovered:

- 4/4 intended span boundaries from renewal gaps;
- 3/3 intended classes from rarity-weighted co-occurrence signatures;
- a provisional membership for the still-open tail span.

Important details:

- signatures must use resolved identities, not raw words; otherwise bottle-water and
  plant-water blur together;
- ubiquitous identities receive zero information weight;
- open-span membership is provisional and commits only when the occurrence closes;
- strict observation/action alternation was exposed as an idealization—the time gap, not a
  synthetic rest action, carries the boundary signal.

The toy therefore supplied a rule for every slot in the class-mediated mechanism before code was
written.

### The live segmentation replay

The segmentation rules were replayed against the July 13 live trace. They recovered six sensible
conversation spans: greeting, harness/location, move-to-me, failed movement/look-first,
movement postmortem, and maze work.

The real trace changed the segmentation composition:

```text
rule 1: eagerly over-segment at the finest witnessed rhythm
rule 2: merge spans whose resolved co-occurrence signatures identify one class
```

This was more stable than tuning a single gap threshold. The live replay also exposed the true
bottleneck: canonical identity. V2 emitted several strings for the same movement event. V3's exact
world-addressed keys solve that for the exact layer; interpreted identity remains the harder case.

## The first observe-wave implementation

Commit `5a49ff6` implemented the complete class-mediated prototype with no inference in the wave:

```text
roots
→ conservation-based heat
→ class rank by root diversity and heat
→ binding selection from counted edge statistics
→ coverage-gated stitching
→ joint-precedent novelty check
→ interpreted resolution commit
```

It added:

- witnessed wall-clock timestamps;
- mechanical span derivation;
- co-occurrence classes and a lens index;
- resolution-history priors;
- a Born-rule statistics hook;
- committed `resolved-to` observations with provenance;
- target derivation from measured class completions;
- a zero-inference `ObserveImpulse` seam.

The target policy was:

```text
familiar impulse → read the class's reliable completion as the target
novel impulse    → framing authors the target and founds precedent
```

This implementation is still present as prototype and test history, but it is no longer the active
ingest architecture.

## Why the class layer came out

The class mechanism passed its paper tests and was still removed almost immediately. The reason
was not that its arithmetic failed. Its ownership and lifecycle semantics were wrong.

The existing grader matched an incoming observation against every open prediction at the same
address. Two unrelated episodes could predict the same address, and one observation could settle
both. Similarity could retrieve a prediction, but it could not establish that the current event
caused or owned that prediction.

Classes also duplicated responsibilities already held elsewhere:

| Responsibility | Existing owner |
|---|---|
| Exact recall of witnessed occurrences | entity/address/triple posting lists |
| Repeated outcome reliability | causal edge `P` and `N` |
| Current relevance | transient activation |
| Prediction lifecycle | the birth span and action completion |

The deeper problem appears when representation becomes less exact. Similarity classes have no
natural eviction boundary. Unresolved predictions can collect monotonically inside a durable
class and be re-exposed whenever anything similar arrives. A span is bounded by provenance: one
external event and its causal descendants. It can decay, close, archive, and later reactivate.

The resulting principle is:

> Ownership by provenance survives representation change; ownership by similarity does not.

Classes may still be useful as analyses, offline compression proposals, or historical fixtures.
They are not on the active grading path.

## Span-owned predictions

V3 now stamps stable identities onto observations, spans, episodes, and expressions.

```text
external impulse I
  → span S
  → planner expression E owned by S
  → acts A owned by E
  → predictions X owned by E/S
  → world outcomes retain S
  → targeted settlement returns retain S
```

A prediction has two independent readiness conditions:

```text
armed   its responsible action or action set completed
active  its birth span is recruited by the current field
```

Eligibility requires both, plus:

- the prediction is unresolved;
- the evidence belongs to the same episode and span;
- the evidence arrived at or after `ArmedAt`;
- evidence provenance is allowed to settle the claim.

Joint predictions arm only after the final responsible action completes. Immediate world outcomes
are appended after completion/arming so real effects cannot arrive "before" their causal source in
the grader. Old-world predictions cannot settle against a replacement Realm.

Targeted settlement is a causal continuation rather than an exception: the perceive request
carries the owner's span, and its return reactivates that span before grading.

## The continuously replaced activation field

Working memory is now implemented as transient activation over durable keys:

```text
entities
triples
spans
causal edges
renderable memory nodes
```

At each external impulse:

```text
A_t(x) = min(1, λ A_(t-1)(x) + H_t(x))
```

with configurable decay, prune floor, span eligibility floor, propagation hop limit, and memory
token budget. The current defaults are experimental, not architectural constants.

The impulse heats its own nodes and exact roots. Direct posting lists distribute heat to prior
spans. Broad postings whose per-span share falls below epsilon are skipped, implementing the fan
effect without scanning an unbounded ubiquitous list.

Reliable learned edges propagate activation:

```text
active span
→ expression born there
→ measured act-to-outcome edge
→ expected world-state key
→ spans witnessing that state
```

The crossing reads historical reliability; it does not change it. Thinking about an edge may raise
activation but never `P` or `N`.

## Bounded memory without compaction

Each observation and expression has a provider-dependent token cost. The current default is a
conservative Unicode-aware approximation, with an injectable exact tokenizer seam.

The model-facing present is rebuilt from active memory under a fixed budget:

```text
mandatory current-root nodes first
remaining nodes ranked by activation / token cost
```

This produces continuous replacement:

```text
durable history grows
transient relevance decays and is renewed
the presented context stays bounded
cold memory remains retrievable through exact indexes
```

The endurance fixture stores more than 100,000 estimated tokens across 4,000 spans while holding
the active projection to 256 tokens. The earliest memory decays out of the present and later
returns through entity reactivation, without a summary or compaction step.

## Validation and defects found

The span/activation implementation records substantial validation:

- five deterministic stress seeds over three episodes, 360 predictions, and 5,400 adversarial
  events per seed;
- 546,293 native fuzz executions for the eligibility milestone;
- another 84,057 fuzz executions for continuous replacement;
- 1,000 delayed predictions across 200 owner spans;
- a benchmark with 0, 1,000, and 100,000 unrelated open bets;
- a 100k-token / 4,000-span endurance run;
- race, vet, focused repetition, and repository-wide checks recorded in the spec.

The stress work found two important defects:

1. **Cross-episode grading.** Span equality alone was insufficient; an old span label could settle
   an old-world prediction after the Realm changed. Eligibility now requires episode identity too.
2. **Mark laundering.** Marks were excluded from grading but still entered current state, allowing
   self-authored content to satisfy guards or targets and overwrite raw evidence. Marks remain in
   durable history but no longer enter the evidence-bearing current-state fold.

The address index keeps grading cost independent of unrelated open predictions. The recorded
benchmark remained approximately 312–370 ns/op across 0 to 100,000 unrelated bets.

For this review, the currently affected packages also pass:

```text
go test ./src/skyra/skyra-dna-v3 ./src/skyra/world3d ./cmd/globe
```

## Compression research: triples are a bootstrap codebook

The representation research reframes a triple as a hand-authored compression code:

```text
(subject, predicate, object)
    = stable discrete address for a recurring relational pattern
```

The decomposer is therefore a compressor using a fixed human-designed dictionary. Replacing
triples should not mean replacing exact keys with continuous embeddings. The proposed move is to
learn the codebook while preserving the substrate contract:

- stable discrete identity;
- exact posting-list retrieval;
- three-valued truth;
- compositional joints;
- world/model provenance;
- a determinate way to confirm or refute the code.

The substrate is already largely representation-agnostic. Span ownership, activation maps,
posting lists, causal edges, and settlement need discrete addressable keys; they do not inherently
need subject/predicate/object syntax.

The proposed objective is constrained Minimum Description Length:

```text
mint code C when:

L(codebook + C) + L(stream | codebook + C)
    <
L(codebook) + L(stream | codebook)

subject to:
    C retains a determinate falsification address
```

The current research connects v3's proper score directly to compression gain:

```text
grade = log(p_claim) - log(p_edge)
```

is the number of log-probability units the claim saved over the learned base rate. Edge-EMA
learning is therefore an online prediction/compression loop rather than a separate reward system.

The research also records two necessary departures from unconstrained Shannon-optimal coding:

1. A statistically predictive code is not automatically falsifiable.
2. Maximum compression destroys the redundancy needed for partial matching, composition, and
   checkable addresses.

Skyra therefore needs a constrained sub-Shannon representation: compress only while
falsifiability and action-relevant distinctions survive.

The proposed migration is gradual:

```text
triples remain the bootstrap and exact floor
→ an offline dictionary/grammar compressor proposes recurring codes
→ MDL and falsifiability gate admission
→ learned codes enter the same indexes and edges
→ learned codes take over interpreted structure first
→ exact world-given addresses remain symbolic
```

The hardest open problem is still the falsification address. Compression can discover predictive
regularity more easily than it can establish what independent observation would confirm or refute
the resulting abstraction.

## Interpretation learning through blind fields

The new uncommitted interpretation spec proposes a way to obtain an interpretation learning signal
without placing an LLM judge at the bottom.

It separates three activation fields:

```text
observation field   produced by an incoming impulse, blind to predictions and desire
retrieval field     produced by self-authored graph queries; attention, not evidence
expectation field   a held prediction about a future eligible observation field
```

The order is load-bearing:

```text
predict → hold
observe → freeze
compare → grade
then optionally combine for attention
```

The expectation field must not preheat the observation pass it will later be compared against.
Otherwise the being manufactures agreement with its own prediction.

### LLM graph traversal inside an interpretation session

One external span may own an internal interpretation session:

```text
raw observations
→ blind exact wave
→ unresolved interpretation residual
→ LLM emits a constrained graph query
→ graph deterministically returns heated memory
→ closure is recomputed mechanically
→ repeat until closed, fixed, or budgeted out
```

Queries do not become observations, settle predictions, satisfy guards, or update world state.
The LLM chooses what to look for; the graph determines what the query retrieves; the substrate
determines whether the residual closed.

Query count alone is not a sufficient cost because one broad query can heat everything. Eventual
cognitive cost should include inference passes, tokens, graph operations, nodes traversed,
activation breadth, and critical-path latency.

### Interpretation prediction

After closure, the interpretation may hold a normalized expectation field `Q`. The next eligible
blind impulse produces observation field `O`; historical activation supplies base rate `B`.

A bounded overlap can be rendered on a human-readable `0..100` scale, but the learning grade must
penalize diffuse predictions. The proposed form is information over base rate:

```text
grade = Σ O_i log(Q_i / B_i)
```

The exact formula is not yet decided. The invariant is that predicting everything earns little,
reciting the base rate earns approximately zero, and a precise surprising success earns credit.

Alignment between two interpreted fields demonstrates predictive continuity, not semantic truth.
The same resolver can repeat the same mistake. Interpretation evidence therefore remains
fidelity-capped, and unresolved ambiguity should eventually recruit a discriminating world action
such as a clarifying question.

### Just-in-time cognition and competence crystallization

The proposed interpretation loop generalizes the existing planning economy:

```text
familiar impulse  → exact/counting path, few or zero interpretation queries
novel impulse     → LLM queries only the unresolved residual
repeated success  → query path crystallizes into a direct resolution edge
```

The same learning shape can later govern motor/tool development:

```text
deliberate plan
→ repeated falsifiable execution
→ stable action subgraph and predictable outcomes
→ guarded procedural competence
→ surprising failure deoptimizes back into cognition
```

In that framing, just-in-time cognition and just-in-time tool development are not separate product
features. They are two consolidation levels. Yesterday's expensive plan becomes today's motor
primitive when observed reliability justifies the compression.

### Reconciliation needed with the current implementation

The interpretation proposal currently names class-level heat in its minimum experiment because it
was written against the observe-wave vocabulary. The active architecture removed classes in favor
of direct span and node activation.

Before implementation, the experiment should be restated over the current field:

```text
Q and O over stable discrete keys and/or span/node activation
not over a durable Class object
```

The core separation—blind observation, non-evidentiary retrieval, held expectation—survives this
change. The experimental carrier needs updating.

## Current architecture by status

### Implemented

- v3 observation/expression loop and exact prediction-diff grading;
- stable episode, span, observation, and expression ownership;
- prediction arming and joint-action causal readiness;
- episode/span/time eligibility for grading;
- owner-span settlement reactivation;
- marks excluded from evidence-bearing current state;
- direct entity/address/triple-to-span indexes;
- continuously decayed entity/triple/span/edge/node activation;
- bounded causal propagation;
- fixed-token active-memory projection by activation density;
- familiar target derivation from activated measured edges;
- interactive cube-sphere globe realm.

### Implemented prototype/history, not active direction

- renewal-gap `Spans` derivation;
- co-occurrence `Classes` and `Lens`;
- class-mediated observe wave;
- class-level binding resolution and target derivation;
- paper-test fixtures for root diversity, coverage, and joint precedent.

These remain valuable as tests and as evidence for the principles later carried into direct-span
activation. They are not the live ingest path.

### Research/specification

- learned discrete compression codes replacing hand-authored interpreted triples;
- constrained MDL and falsifiability-address admission;
- interpretation sessions with LLM-authored graph queries;
- held expectation fields and future-wave interference grading;
- query-cost and interpretation-error learning curves;
- crystallization of repeated interpretation or action paths into competences;
- just-in-time tool development.

## What changed conceptually

The two-day arc compresses to several durable decisions:

1. **The active field is now implemented.** It is transient activation over durable exact keys,
   continuously replaced per impulse.
2. **Ownership is event-based.** Spans own predictions; similarity retrieves but does not own.
3. **Causal readiness and attention are separate.** A prediction must be both armed and active.
4. **Context is bounded by selection, not by rewriting history.** Memory is projected, not
   compacted.
5. **Ubiquity carries no information.** Broad postings self-discount by conservation/flooring.
6. **Signed middle-state was not earned.** Compatibility is checked through joint precedent at
   collapse rather than cancellation during traversal.
7. **Classes are analyses, not runtime ontology.** Their useful rules survive without their
   durable object.
8. **Triples are recognized as bootstrap codes.** The substrate needs discrete checkable keys,
   not necessarily a permanently hand-authored grammar.
9. **Interpretation learning should have its own two curves.** Cognitive effort and future error
   must fall together.

## Open questions and risks

### Span policy

- What makes a later human turn a continuation of the current span versus a new independent span?
- How should interleaved threads, interruption, resumption, and delayed unsolicited effects retain
  causal ownership?
- When and how are permanently unresolved predictions archived without being counted false?

### Activation

- Fit decay, eligibility floor, and hop limits from live interruption/return behavior.
- Determine whether edge recency `F` adds predictive value beyond current activation.
- Validate causal recruitment and bounded projection under real conversational ambiguity, not only
  synthetic endurance.

### Interpretation

- Keep exact guards and targets from accepting lower-authority interpreted evidence as if it were
  raw world state.
- Define a constrained graph-query representation without creating an infinite interpretation
  recursion.
- Choose the field carrier and base-rate context for expectation scoring.
- Distinguish repeated semantic truth from a resolver repeating the same coherent mistake.
- Prefer clarifying actions when memory cannot discriminate competing readings.

### Representation and compression

- Define a falsification address for learned abstract codes.
- Keep code identity stable through mint, merge, split, and deprecation.
- Preserve composition and partial matching without allowing fuzzy identity into settlement.
- Avoid compressing away distinctions that later change action quality.

### Capability and embodiment

- The filesystem `exec` realm still needs a real capability boundary before autonomous tool
  development can be treated as safe.
- Generated motor competences need versioned identity, explicit guards, observable effects, and a
  deoptimization path.
- The globe proves realm portability, but not yet learning across rich, noisy, multi-agent
  interaction.

## Recommended next experiments

1. **Real conversation endurance.** Feed interleaved requests, corrections, interruption, return,
   delayed outcomes, and ambiguous references through the active span field. Measure false grading,
   recall, context size, latency, and target completion.
2. **Restate the interpretation-field toy over direct activation.** Preserve blind `O`, held `Q`,
   and base rate `B`, but use stable span/node/code keys rather than the retired class layer.
3. **Add one constrained interpretation query.** Verify that it reduces a mechanical residual while
   remaining unable to alter observation state or settle a prediction.
4. **Run the compression paper test.** Apply Sequitur/BPE or another discrete grammar compressor to
   a real resolved spine. Count useful recurring codes, one-off rejections, and codes lacking a
   falsification anchor.
5. **Measure the two interpretation curves.** Repeated ambiguous fixtures should require fewer
   graph queries while future-wave prediction error also falls.
6. **Delay tool crystallization until the capability boundary is real.** Once safe, detect repeated
   successful action subgraphs and test guarded, versioned consolidation with explicit rollback.

## Bottom line

The July 18 direction said that context should be a recurrent field reconstructed from interaction
and carried structure. The July 19–20 work now provides an implemented v3 version of that claim:

```text
durable experience
    + new impulse
    → transient activation
    → exact span retrieval
    → bounded active projection
    → causally eligible learning
```

The key simplification was not a new clustering algorithm. It was recognizing that provenance and
similarity have different jobs. Spans own predictions because they record what caused them.
Similarity and activation may retrieve those spans, but they do not get to rewrite ownership.

The next frontier is to make interpretation obey the same discipline. Let the LLM query memory and
propose readings, but keep observation blind, queries non-evidentiary, closure mechanical, and the
learning signal downstream in independently arriving consequences. If both queries-to-closure and
future error fall, interpretation will be learning rather than merely becoming internally
consistent.

# Scale, Structure, Content: A Composable Grading Structure for Predicted Observations

**Date:** 2026-07-21  
**Status:** RESEARCH NOTE · base-rate-relative calibration, same-claim fusion, the
scale→structure→content dependency, and void-on-unmaterialized-scale are settled
design structure. The retained dimensional comparison field is now also settled as
the representation target; its mature parser vocabulary, budget calibration, and
autonomous same-claim detector remain open.  
**Extends:**
[Predicate Diagnosticity: Learning the Consequence Check, Not Hand-Writing It](predicate-diagnosticity.md)  
**Two-axis line:**
[The Two-Bit Outcome: Interpretation and Prediction Are Separately Falsifiable](the-two-bit-outcome.md) ·
[Two-Axis Inference: Interpretation Grades Through the World, Not Itself](two-axis-inference.md)

**Reflective companion:**
[Phase as Relational Alignment](phase-as-relational-alignment.md).
It treats same-claim fusion as the current phase analogue while keeping redundancy
discounting distinct from contradiction cancellation until a true node forces the
latter.

**Implementation checkpoint:** v3 now implements paragraph-count as a deterministic
structural parser. Content dimensions carry an explicit paragraph-carrier identity and
one-based slot. When fewer paragraphs arrive than predicted, materialized slots settle and
absent slots become `VOID`; only settled rows enter the generic learning projection. List
and code-block carriers have not yet been added.

## The unifying rule

The foundation is settled:

> **Every tool converts its raw output into base-rate-relative surprise. Tools
> measuring the same claim fuse into one predicate; tools measuring different claims
> sum. Range or threshold width is not a separate penalty—it is base-rate coverage,
> priced automatically.**

Each clause closes a different ambiguity in the prior note.

### Every tool is calibrated against its own base rate

A tool's raw value has no universal evidentiary meaning:

```text
boolean match       true / false
regex               matched / did not match
embedding           cosine = 0.78
token count         143
range membership    inside / outside [120, 180]
marker count        4
```

`true`, `0.78`, and `143` are different shapes of output, but the same question makes
them informative:

```text
How unusual is this result among eligible observations in the relevant base-rate
window?
```

Each tool therefore measures its own held-out distribution:

```text
boolean tool       base rate = P(predicate fires)
embedding tool     base rate = distribution of similarity scores
length tool        base rate = distribution of token counts
range predicate    base rate = P(value lands inside range)
count tool          base rate = distribution of observed counts
```

The distributions have different shapes; the epistemic conversion is the same. A raw
tool output becomes evidence only after it is located relative to the outcomes that
tool ordinarily produces.

### Same claim fuses; different claims sum

Several tools may observe one proposition. Their outputs are sensors, not separate
facts. They must fuse into one predicate settlement before credit is assigned.

Distinct propositions may each be right or wrong. Once independently settled, their
base-rate-relative contributions may sum.

### Width is coverage

A wide range or permissive threshold fires on more of the held-out world. That raises
its base-rate coverage and makes a match less surprising. No second `width penalty` is
needed:

```text
wider predicate
    → higher P(fire | base rate)
    → smaller likelihood ratio when it fires
    → less credit
```

The same principle closes three loopholes:

```text
match-anything
    loose predicate → high base rate → near-zero credit

redundant stacking
    same-claim sensors fuse → one settlement, not several rewards

uncalibrated embedding
    raw cosine has no grade until placed in its held-out similarity distribution
```

**Settled:** tools may differ; evidentiary currency does not. They all enter grading as
base-rate-relative surprise.

## Same-axis fusion

Suppose Skyra predicts that the next eligible observation will contain traversal talk.
It constructs three checks:

```text
string sensor
    contains token "traversal"

regex sensor
    /\b(depth[- ]first|breadth[- ]first|graph walk)\b/i

embedding sensor
    cosine(observation, traversal-centroid) ≥ 0.72
```

All three measure the same claim:

```text
C₁ = "this observation contains traversal talk"
```

They are not three independent content predictions. They are multiple sensors on one
predicate:

```text
string ───┐
regex ────┼──▶ fused predicate for C₁ ──▶ one settlement ──▶ one grade
embedding ┘
```

If all three fire on the same phrase, Skyra did not observe traversal three times.
Summing their scores would pay repeatedly for correlated restatements of one fact. A
proposer could then inflate credit by adding synonymous regexes, nearby embedding
centroids, and duplicated token sets.

Fusion removes that incentive by construction. The predicate definition records how
its sensors combine—such as a calibrated joint rule, threshold, or learned fusion—but
the outward claim has one identity and one settlement.

By contrast, these are different claims:

```text
C₁ = traversal talk appears
C₂ = a code block appears
C₃ = a question is asked
```

They can fail independently and may sum after each is calibrated to its own base
rate. The difference is semantic and structural, not merely low empirical
correlation: distinct claims name distinct falsification addresses or distinct
properties at an address.

Threshold and range breadth still require no standalone complexity charge. A loose
embedding threshold or expansive regex fires more frequently in the held-out window;
its high base rate automatically drains its surprise. This claim assumes honest
held-out measurement. Searching many fusions against the same small sample remains a
finite-data overfitting risk carried from Predicate Diagnosticity.

**Settled rule:** fuse sensor outputs before grading when they measure one claim; sum
only after independently settling genuinely different claims.  
**Open mechanism:** autonomous detection of that boundary appears at the end of this
note.

## Token count is scale, not a peer axis

Length is not another content claim competing for weight beside `traversal talk` or
`contains code`. Token count predicts the scale of the future observation. Scale sets
the budget for how many structural and content sub-observations the prediction may
license.

```text
predicted token range
        ↓ sets observation budget
number of structural slots
        ↓ licenses
number of distinct content predicates
```

More expected tokens create room for more independently checkable observations. A
three-token reply cannot honestly carry the same predictive decomposition as a
three-page design review.

This resolves two symmetric failures.

### Short scale, too many content predictions

```text
scale prediction:  5–12 tokens
content claims:    traversal + recursion + complexity + code + benchmark + caveat
```

The predictor over-commits. The expected observation is too small to support all
those distinct checks. Most will whiff even if the short answer is appropriate. A
scale-derived budget prevents licensing that many claims.

### Long scale, almost no content prediction

```text
scale prediction:  800–1,200 tokens
content claims:    "mentions traversal"
```

The predictor under-commits. It predicts a large blob while saying almost nothing
about what fills it. Length alone becomes a cheap way to appear predictive without
risking content.

Scale therefore calibrates appropriate specificity. It is a budget on the number of
terms that may enter the lower-tier sum, not a multiplier or modifier applied to the
grading equation.

The dependency shape is settled. The precise conversion from token range to budget is
not. It may depend on historical proposition density, observation modality, language,
or task type; a fixed `tokens per predicate` constant would be premature.

## Three-tier grading structure

The settled structure has three dependent tiers.

```text
TIER 1 — SCALE
token-count range
    ↓ sets observation budget

TIER 2 — STRUCTURE
spends that budget into expected shape
    ↓ licenses structural observation slots

TIER 3 — CONTENT
fills licensed slots with distinct claims
    ↓ same-claim sensors fuse; distinct claims sum
```

Tier 1 gates tier 2. Tier 2 gates tier 3. Length licenses structure; structure licenses
content.

## Representation correction: preserve the comparison field

The tiered structure above describes what is predicted and what becomes eligible to
settle. It must not be collapsed into one Boolean—or even into the four labels
`TRUE / FALSE / VOID / UNKNOWN`—at the point of observation. Those labels are useful
local settlement metadata. They are not the learned object.

The LLM authors a **span-owned predicted-observation field**. It chooses a parser from a
fixed vocabulary and emits that parser's arguments together with the dimensions it is
staking. After the next eligible impulse has been fully written and frozen, the supplied
parser receives only those arguments plus the impulse and emits an observed field in the
same coordinate system. The substrate then constructs and retains the comparison:

```text
LLM, before observation
    PredictionField {
        owner:          span / claim path
        parser:         fixed parser identity
        parser_args:    frozen arguments
        scale:          token range
        structure:      paragraph / list / code-block expectations
        content:        marker, regex, and similarity expectations
        attribution:    signed observation ranges → stable dimension IDs
        eligibility:    address + horizon
    }

world, after the eligible impulse is fully written
    fixed parser(parser_args, frozen impulse)
        → ObservedField {
              token_count
              paragraph_count
              list_count
              code_block_count
              marker_hits
              regex_results
              similarity_values
              raw_evidence_ids
          }

substrate
    compare(PredictionField, ObservedField)
        → ComparisonField {
              predicted values / ranges
              actual values
              residuals or relations per dimension
              materialized / void / unknown metadata
              per-tool base-rate calibration
              same-claim fusion membership
              ownership + raw evidence provenance
          }
```

The distinction matters. A token prediction `[120, 180]` followed by an actual count of
`143` contains more information than `TRUE`. A semantic threshold of `0.72` followed by
`0.84` contains more information than `TRUE`. A prediction of three paragraphs followed
by one contains a signed structural residual, while content claims owned by the two absent
slots are void. Throwing those measurements away recreates the low-dimensional learning
surface this line of work was intended to escape.

The comparison field is therefore the durable learning record. A Boolean settlement,
proper score, or base-rate-relative scalar grade is a **projection** of that record for a
particular update—not a replacement for it. Changing the scoring projection later must not
require reconstructing raw measurements that were discarded earlier.

Interpretation and consequence use the same field representation and fixed parser
vocabulary, but each claim owns a separate `PredictionField`, eligibility address,
comparison history, and grade projection. Shared representation does not merge the two
axes. The Michael case can therefore retain a rich traversal-content comparison for the
consequence claim and a separate school-versus-Skyra context comparison for the
interpretation claim.

The first proof deliberately excludes LLM-authored parsers. The fixed vocabulary target is:

```text
exact_address
token_count
paragraph_count
list_count
code_block_count
string_markers
regex_match
semantic_similarity
```

**Landed minimum 2026-07-21:** `exact_address`, `token_count`, and `string_markers` now
run through the fixed registry and retain comparisons. Paragraph/list/code-block counts,
regex, and semantic similarity remain implementation work; listing them here does not mark
them landed.

The LLM may select a supplied parser and emit its typed arguments. It may not write a new
parser during this phase. Just-in-time parser development can be reconsidered only after
the prediction → parse → compare → learn path preserves dimensionality correctly with
known evaluators.

### Signed attribution into prediction dimensions

The prediction field also preserves which passages in already-available observations the
LLM says influenced each dimension:

```text
[O41][120:168]:+0.82->[P17.content.topic]
[O41][220:251]:-0.55->[P17.content.topic]
```

The range uses start-inclusive/end-exclusive Unicode character offsets into immutable raw
observation `Object` text. The pointer names a stable prediction-dimension ID, not a positional
array element. `+1` means strong claimed support, `-1` strong claimed counterevidence, and
zero no directional relevance. The number is authored alignment, not calibrated
probability and not earned credit.

This compact signed link replaces an unnecessary emitted relation vocabulary. The sign
carries support versus contradiction; the dimension pointer carries what is distinguished;
parser identity carries measurement; and licensing and settlement remain substrate-owned
events. Birth-evidence attributions are frozen with the prediction and kept separate from
the later observation IDs that settle it.

Value remains relational:

```text
claimed influence(observation passage, prediction, dimension, birth time)
```

No flat value is written onto an observation. Repetition may later calibrate whether the
LLM's signed attributions were useful, but a downstream hit cannot directly credit every
passage it named.

**Settled representation:** retain predicted value, observed value, relation/residual,
calibration, settlement metadata, ownership, and raw evidence per dimension. Fuse
same-claim sensors and sum distinct claims only when producing an optional grade
projection. Retain signed birth-evidence passage links separately from settlement evidence.
Do not dispose of the underlying dimensions into four outcomes.

**Open mechanism:** the concrete typed schema, parser registry API, and update rule over
the retained residual history still require
[Toy #9F](toy-9f-span-owned-prediction-fields.md) as an implementation gate. This is a
data-model step before differential abduction, not a new inference subsystem.

### Tier 1 — scale

```text
prediction:  token count ∈ [lo, hi]
effect:      sets the observation budget
graded by:   did eligible length land in the range?
calibration: base-rate probability mass inside [lo, hi]
```

A wide range contains more of the historical length distribution, so a hit earns
less surprise. A narrow range earns more when right and misses more often. Width is
already priced as coverage.

### Tier 2 — structure

Given the scale budget, Skyra predicts shape:

```text
three paragraphs
one code block
one concluding list
```

Each is a structural sub-predicate with its own boolean settlement and base-rate
distribution. Structure allocates the abstract token budget into observable regions
that can license content.

```text
predicted 3-paragraph response
    ├── paragraph slot 1
    ├── paragraph slot 2
    └── paragraph slot 3
```

Structure is not inferred after the response merely to make the prediction fit. The
licensed slots and checks exist before the eligible observation.

### Tier 3 — content

Within each licensed observation slot, Skyra predicts claims:

```text
paragraph 1: traversal problem statement
paragraph 2: algorithm or code
paragraph 3: complexity and caveat
```

String, regex, marker, and embedding tools measuring the same claim fuse into one
predicate. Different content claims settle independently and their calibrated
contributions sum.

The positive-credit view is:

```text
grade = Σ base-rate-relative surprise of settled predicates that fired
```

The learning record must also retain settled false outcomes so failed predicates lose
reliability under the chosen proper settlement rule. A non-fire at an eligible,
materialized address is a miss; absence of the address itself may instead be void, as
specified below.

No raw cosine, token count, or duplicate regex enters the **scalar sum** directly. The raw
measurements and their predicted counterparts remain in the comparison field. Calibrated
contributions are fused within a claim and added across distinct claims only when a scalar
projection is required.

## The matched-budget safeguard

The hierarchy closes a second route to gaming. Skyra cannot predict `three paragraphs`
and then stack twenty content predicates merely to create twenty chances for credit.
The scale-derived observation budget caps the structure and content claims that may be
licensed.

Predicting a longer observation permits more checks, but the additional claims are
not free lottery tickets. If the scale materializes, each licensed predicate becomes
independently falsifiable:

```text
predict longer
    → receive larger observation budget
    → license more distinct claims
    → owe more settled predictions if that scale arrives
```

Conversely, predicting a long response and spending almost none of its budget earns
little content information. The system has forecast size without forecasting the
observation's internal structure.

There is therefore a matched sweet spot between predicted scale and predicate count:

```text
too few predicates for the scale
    → under-specific; little information earned

matched predicates for the scale
    → useful coverage and diagnosticity

too many predicates for the scale
    → over-committed; settled misses accumulate
```

This is the Predicate Diagnosticity inverted U one level up. Tightness and coverage
govern one predicate; scale and licensed specificity govern a composed observation.

The budget is a cap and accounting surface, not itself an automatic reward. Unspent
budget does not need a fabricated penalty if the system simply earns no content
credit for claims it did not make. Whether mature evaluation needs an explicit
under-commitment diagnostic remains open; it must not become a coverage penalty that
rewards guessing.

## Void, not failed

The observation budget is a license, not a requirement. This distinction is settled
and load-bearing.

Suppose Skyra predicts:

```text
scale:      3 paragraphs

paragraph 1 content: problem statement
paragraph 2 content: code example
paragraph 3 content: complexity discussion
```

Only one paragraph arrives.

The correct settlement is:

```text
TIER 1
3-paragraph scale prediction       → MISS

TIER 2 / TIER 3
paragraph 1 predicates             → SETTLE against paragraph 1
paragraph 2 predicates             → VOID
paragraph 3 predicates             → VOID
```

The content predicates for paragraphs 2 and 3 predicted properties of observations
that never materialized. There is no eligible object against which to evaluate them.
They are unknown, not false.

This is the existing invariant one level down:

> A content prediction about an observation that never materialized is unsettled and
> void, not refuted. Unknown is not false.

The two failure directions are different:

```text
predicted scale did not materialize
    → grade the scale miss
    → void predicates owned by absent slots

predicted scale and slot did materialize, but content was absent or contrary
    → settle the content predicate false
```

Getting this wrong double-counts the scale error and corrupts content learning. It
punishes Skyra for content claims about paragraphs the world never supplied, causing
the content layer to learn from missing carriers rather than observed content. Over
time that teaches the being to predict fewer details, which is backward.

The scalar grade domain is therefore:

```text
grade = base-rate-discounted sum across SETTLED predicates only

VOID predicates contribute neither success nor failure
```

Every lower-tier predicate retains the identity of the scale and structural slot that
licensed it, so materialization can be decided mechanically rather than inferred from
semantic similarity.

`SETTLED`, `VOID`, and `UNKNOWN` are dimension-local metadata in the comparison field.
They do not flatten the field: a settled dimension still retains its prediction, actual
measurement, residual, calibration, and evidence. An absent carrier has no fabricated
actual value; it is marked void with the licensing relationship intact.

## Token-count standard deviation is rejected for now

A token-count range already expresses location and tightness:

```text
[lo, hi]
```

Its width is observable and its historical coverage is measurable. Wider ranges fire
more often and receive less base-rate-relative surprise. Narrower ranges fire less
often and make a sharper claim.

Adding a standard deviation or full predictive distribution encodes similar forecast
information without a demonstrated need. The observed value and residual are retained;
only the predictive envelope stays simple:

```text
[lo, hi] prediction + actual token count
    ↓
range settlement + signed residual + base-rate calibration
```

That additional surface is not currently earned. It adds calibration and serialization
choices without resolving a demonstrated failure.

**Settled decision:** keep `[lo, hi]` as the prediction, while retaining actual count and
residual in the comparison field. Upgrade the predictive side to a distribution only when
a concrete case shows the range is inadequate. Do not confuse preserving the observed
measurement with requiring a Gaussian forecast.

## Open problem: deciding when sensors share a claim

This note treats string match, traversal regex, and traversal embedding similarity as
sensors for one claim because a human can see that they all measure `traversal talk`.
Skyra needs to make that distinction autonomously.

A likely route is empirical:

```text
measure each sensor's held-out firing distribution
        ↓
measure their joint firing distribution
        ↓
detect persistent correlation or redundant information
        ↓
propose fusion into one claim
```

This bottoms out in the same base-rate machinery. Same-claim detection needs the joint
base-rate distribution and a correlation or conditional-information test. The
recursion is productive rather than circular: every layer converts raw coincidence
into information relative to counted history.

But correlation is not identity. Two genuinely distinct claims may co-occur because
the world couples them, while two sensors for one claim may disagree because one is
poor. A detector must preserve provenance and remain revisable rather than permanently
merging every correlated pair.

The concrete detector is unspecified. Open questions include:

- which dependence statistic distinguishes redundancy from useful complementarity;
- how much history is required before fusion;
- whether tools fuse at the predicate, code, or transient observation level;
- how a fused predicate splits after drift;
- how to prevent the fusion search itself from overfitting;
- whether a shared falsification address is necessary or sufficient for same-claim
  identity.

**Settled boundary:** redundant sensors cannot sum as independent evidence.  
**Open mechanism:** how Skyra discovers that redundancy without human semantic
judgment.

## Bottom line

The principle is one composable dependency:

> Token count predicts scale and sets the observation budget. Structure spends the
> budget into shape predicates. Content fills each licensed observation with
> same-claim-fused, base-rate-discounted predicates. The durable result is the
> dimensional comparison field; a grade is a base-rate-discounted projection across
> settled predicates. Predicates licensed by a scale that did not materialize are void,
> not failed.

One mechanism—base-rate-relative surprise—operates at every tier. Tool outputs differ
in shape, but all are calibrated against what that tool ordinarily returns. Same-claim
sensors fuse, distinct claims sum, and breadth pays for itself through base-rate
coverage rather than a separate complexity penalty. None of those projections erase the
predicted values, actual measurements, residuals, or provenance that made them possible.

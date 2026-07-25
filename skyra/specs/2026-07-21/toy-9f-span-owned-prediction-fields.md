# Toy #9F — Span-Owned Prediction Fields Preserve the Measurement

**Date:** 2026-07-21  
**Status:** PARTIAL PASS · minimum live Phase-4F slice landed 2026-07-21; exact-address,
token-count, paragraph-count, marker, signed attribution, retained-comparison, and
paragraph-carrier void rails pass. Conservative explicit same-claim fusion also passes.
List/code-block structure, regex, semantic similarity, full non-fire scoring, learned joint
fusion, and residual-history learning remain open

**Parent:** [Interpretation Layer — Implementation Phases](interpretation-implementation-phases.md)  
**Reads on:** [Scale, Structure, Content](scale-structure-content.md),
[Predicate Diagnosticity](predicate-diagnosticity.md), and
[Toy #9 — One Impulse, Two Addresses, Two Independent Updates](toy-9-two-address-claim-paths.md)

---

## What this gate corrects

Toy #9 proved that interpretation and consequence require independent ownership,
predictions, addresses, and histories. Its exact `Triple` checks and two Boolean outcomes
were the minimum representation needed to prove routing. They are not the mature shape of
a predicted observation.

Collapsing the next impulse to `TRUE / FALSE / VOID / UNKNOWN` would preserve routing while
discarding the dimensionality needed for learning. This gate requires the system to retain
what was predicted, what was measured, and how they differed on every licensed dimension.
Settlement state remains attached to a dimension; it does not replace the dimension.

## The fixed-parser contract

For the first proof, parsers are supplied by the substrate. The LLM may select a parser and
emit typed arguments, but may not write a parser.

```text
PredictionField belongs to the predicting span / claim
    parser_id
    parser_args
    eligibility address and horizon
    predicted dimensions
    signed evidence attributions into those dimensions

fully written eligible impulse
    ↓ fixed parser(parser_args, impulse)
ObservedField in the same dimensions
    ↓ substrate comparison
ComparisonField retained with the span / claim history
```

The initial parser registry is intentionally plain:

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

Every parser is deterministic at evaluation time. `semantic_similarity` means a fixed
embedding model, fixed reference or centroid, and fixed threshold/range named before the
impulse arrives. The embedding is a sensor; it is not the judge.

The parser receives the frozen impulse and precommitted arguments only. It cannot read the
expected answer in order to reshape its output, and settlement cannot begin until the raw
impulse is completely written.

## Minimal evidence-attribution language

Some observations are long. Recording only an observation ID says that the observation
participated in authorship, but not which passage mattered or which part of the prediction
it influenced. A single value stored on the observation is also wrong: evidentiary value is
relative to a prediction dimension, not an intrinsic property of the observation.

The LLM therefore emits zero or more signed span references alongside the prediction:

```text
[O41][120:168]:+0.82->[P17.content.topic]
[O41][220:251]:-0.55->[P17.content.topic]
[O44][0:36]:+0.61->[P17.structure.code_blocks]
```

The grammar is deliberately small:

```text
[observation-id][start-char:end-char]:claimed-influence->[prediction-id.dimension-id]
```

Its semantics are:

```text
-1.0    this passage strongly argues against the predicted dimension
 0.0    this passage is irrelevant or has no directional contribution
+1.0    this passage strongly argues for the predicted dimension
```

`-1` means active counterevidence, not merely absence of support. Absence belongs near
zero. Intermediate values preserve the LLM's claimed direction and strength without
pretending that `+0.82` is an 82-percent probability.

Ranges use Unicode character offsets into the immutable `Object` text of the identified
raw observation: start inclusive, end exclusive. The target is a stable prediction and
dimension identity, never an array position. Reordering dimensions must not redirect an
attribution.

The attribution is frozen with the prediction before the eligible world observation
arrives. It records:

```text
which prior observation passage
    → the LLM deemed relevant
    → to which predicted dimension
    → in which direction and claimed strength
```

It does **not** record earned credit. The later comparison field records what the world
actually supplied. Over repeated settlements, the substrate may calibrate whether passages
assigned strong positive or negative influence were reliably useful for that dimension.
A single successful downstream prediction cannot by itself prove that every positive
attribution named the right reason.

No emitted relation taxonomy is required. Support and contradiction are the sign;
distinction is expressed by the dimension pointer; parser identity already says what
measures the dimension; and licensing and settlement remain structural events owned by the
substrate. Existing action `Requires` semantics are not duplicated here.

The retained form is conceptually:

```text
EvidenceAttribution {
    observation_id
    start_char
    end_char
    prediction_id
    dimension_id
    claimed_influence  // [-1, 1], authored rather than earned
}
```

These are **birth-evidence links**. They point only to observations available before the
prediction was authored. Later parser-produced evidence IDs are **settlement links**. The
two sets remain separate so post-outcome material cannot be rewritten as the reason the
prediction was made.

## Worked field

Before Michael's next message, the consequence claim predicts:

```text
owner:       consequence claim C17 on span S41
prediction:  P17
eligibility: next message:/reply

EVIDENCE ATTRIBUTION
    [O41][120:168]:+0.82->[P17.content.traversal-talk]
    [O41][220:251]:-0.55->[P17.content.traversal-talk]

SCALE
    token_count predicted [80, 180]

STRUCTURE
    paragraphs predicted [2, 4]
    code_blocks predicted [0, 1]

CONTENT CLAIM: traversal-talk
    string_markers predicted one-of {traversal, graph walk, depth-first, breadth-first}
    semantic_similarity predicted [0.72, 1.00] against frozen traversal centroid
```

The fully written eligible impulse is parsed as:

```text
token_count         actual 143
paragraphs          actual 3
code_blocks         actual 1
string_markers      actual {graph walk}
semantic_similarity actual 0.84
raw_evidence        observation IDs O201…O205
```

The retained comparison is not `TRUE`:

```text
dimension             predicted       actual        relation / residual       state

token_count           [80,180]        143           in range; +63 from lo     SETTLED
paragraphs            [2,4]           3             in range; +1 from lo      SETTLED
code_blocks           [0,1]           1             in range; +1 from lo      SETTLED
traversal markers     marker set      {graph walk}  member hit                SETTLED
traversal similarity  [0.72,1.00]     0.84          +0.12 over threshold      SETTLED
```

Each row additionally retains its base-rate calibration window and raw evidence IDs.
The two traversal sensors belong to one claim. Their calibrated readings fuse for an
optional traversal grade; their raw measurements remain separately inspectable. Scale,
structure, and traversal are different claims and may contribute separately to a scalar
projection without becoming one stored number.

## Missing carriers

Now change only the structural arrival: predict three paragraphs, but receive one.

```text
scale / paragraph-count prediction     → SETTLED MISS; actual count retained
paragraph 1 content predictions        → SETTLED against paragraph 1
paragraph 2 content predictions        → VOID; no fabricated actual value
paragraph 3 content predictions        → VOID; no fabricated actual value
```

The void rows remain in the comparison field with their owners and licensing links. They
have no observed value and do not enter the grade. `UNKNOWN` is reserved for a prediction
whose eligibility has not yet been settled. Unknown is not false; void is not false; and
neither label deletes the predicted dimension.

## Two axes, one representation, independent fields

Interpretation and consequence may parse the same incoming span but do not share a
`PredictionField` or `ComparisonField`:

```text
consequence field
    predicts traversal-content dimensions
    compares against topic observations

interpretation field
    predicts Skyra-context dimensions
    compares against category-context observations
```

Michael's message can therefore produce rich versions of the Toy #9 result:

```text
consequence comparison
    traversal marker hit; traversal cosine 0.84; ...

interpretation comparison
    Skyra markers absent; school marker present; context cosine residual; ...
```

Any convenient projection may say consequence right and interpretation wrong. The
retained records explain how right and how wrong without permitting one axis to update the
other.

## Required invariants

Toy #9F passes only if all of these hold:

- changing the scalar scoring projection does not mutate or erase a stored comparison;
- prediction-dimension order and observed-token order do not change the comparison;
- predicted and actual values, residual/relation, calibration, ownership, and evidence
  survive for every settled dimension;
- same-claim sensors fuse for grading but keep their individual raw measurements;
- distinct claims may sum only in an explicit projection;
- child predicates owned by an unmaterialized structural slot are void, not failed;
- an as-yet-ineligible prediction is unknown, not failed;
- consequence and interpretation own separate fields even when one impulse supplies both;
- every evidence attribution points to an existing immutable raw observation range and a
  stable dimension owned by that prediction;
- attribution offsets are start-inclusive/end-exclusive Unicode character indexes;
- invalid ranges, future observations, unknown dimensions, non-finite numbers, or values
  outside `[-1,1]` reject the attribution before arming;
- dimension or attribution insertion order cannot change the links;
- claimed influence remains distinct from earned credit and cannot directly update an
  edge, category, or observation;
- birth-evidence links and later settlement-evidence links remain separately queryable;
- the parser runs only after the complete impulse is written and sees only the impulse plus
  precommitted typed arguments;
- an unknown parser identity or invalid argument shape rejects the prediction before
  arming;
- the existing exact-address `Triple` evaluator remains usable as the smallest parser,
  proving compatibility without defining the ceiling;
- no LLM call occurs at grade time.

## What is settled and what remains open

**Settled by the design:** the comparison field—not a Boolean or one scalar—is the durable
learning object; parsers are deterministic evaluators selected before observation; the
prediction belongs to its span/claim; the parser output uses the prediction's dimensions;
interpretation and consequence retain independent fields; and signed passage attributions
record which pre-observation evidence the LLM says influenced each prediction dimension
without treating that statement as earned credit.

**Not yet settled:** the Go schema, parser registry interface, storage lifetime, exact
base-rate record attached to each dimension, scalar projection used by the current `Edge`
learner, and how histories learn from residual shape rather than only Bernoulli outcomes.
How claimed-influence magnitudes are calibrated over repetition is also open. LLM-proposed
parsers are explicitly deferred.

## Gate result

The minimum gate partially passes in code. `PredictionField`, `ObservedDimension`,
`ComparisonField`, and `EvidenceAttribution` are durable typed records. The fixed registry
implements `exact_address`, `token_count`, and `string_markers`. Existing consequence and
interpretation `Expect`s now receive separate stable prediction IDs; exact settlement
records the complete comparison before projecting its truth into the existing `Edge`
update. The model-facing grammar accepts
`[observation][start:end]:influence->[prediction.dimension]`, namespaces local prediction
IDs, validates raw pre-birth Unicode ranges and signed values, and keeps those birth links
separate from later settling observations. No grading-time LLM call was added.

Executable Phase-4F tests cover multidimensional retention, range/marker/exact parsers,
Unicode attribution bounds, future/invalid reference rejection, insertion-order
invariance, storage-copy isolation, no-credit-on-comparison, live parser authorship, and
separate consequence/interpretation comparison histories. The missing-carrier forcing
case is also live: a predicted three-paragraph response receiving one paragraph settles
the paragraph-count miss and paragraph-one content, retains paragraph-two and
paragraph-three predictions as evidence-free `VOID` rows, and excludes those rows from
the generic settled-only learning projection.

The full gate remains open for list/code-block parsers, regex, fixed embedding similarity
and version identity, full non-fire scoring, and learned residual-history updates.
Paragraph slots prove the carrier/void structure; additional structural parsers should
reuse it rather than introduce another settlement rule.

The positive-fire calibration rail is now live. A named window points only to durable raw
pre-birth observations and category labels owned by separate admitted predictions whose
comparisons already settled from raw evidence. Declared predicate-selection evidence may
not overlap the calibration observations, and the prediction may not label itself.
Per dimension the frozen record retains category/non-category support, fire counts,
coverage, base rate, `PMI = log(coverage/base-rate)`, and
`coverage × PMI`. The worked `Q_A/Q_B/Q_C/Q_evil` arithmetic is executable: the
diagnostic middle wins and the match-anything predicate approaches zero. Missing support
is recorded as `INSUFFICIENT_SUPPORT`; a never-firing measured predicate is
`MEASURED_INERT`; zero category coverage with positive base rate, or any finite
non-positive PMI, is `MEASURED_ANTI_DIAGNOSTIC`; only finite positive PMI is
`CALIBRATED`. Every state preserves measured counts, and `Ready()` is derived only from
`CALIBRATED`. This is arithmetic readiness, not promotion; true blindness still depends on
the caller withholding the named window from candidate generation.

The conservative same-claim projection is also live as of 2026-07-22. Every authored
dimension must name the claim it measures; missing claim identity is rejected rather than
silently treating redundant sensors as independent. Within a claim, the representative is
selected from frozen calibration by maximum expected credit before the current outcome is
inspected. Only that representative may contribute its per-fire PMI if it settles
true. There is no fallback to a weaker sensor that happened to fire. Distinct claim winners
sum; duplicates, broad siblings, unready sensors, non-positive sensors, unknown rows, and
void rows cannot inflate the result. Calibration itself is substrate-owned and cannot be
supplied on an authored prediction.

This is a conservative selection projection, not the mature learned joint predicate. It
proves the identity and anti-stacking structure while leaving autonomous claim discovery
and jointly calibrated multi-sensor fusion open.
Phase 5 must not assume those unfinished dimensions exist.

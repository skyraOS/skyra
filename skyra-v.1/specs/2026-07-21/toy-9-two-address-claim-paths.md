# Toy #9 — One Impulse, Two Addresses, Two Independent Updates

**Date:** 2026-07-21  
**Status:** PASS · paper gate for Interpretation Phase 4  
**Parent:** [Interpretation Layer — Implementation Phases](interpretation-implementation-phases.md)  
**Reads on:** [The Two-Bit Outcome](the-two-bit-outcome.md),
[Two-Axis Inference](two-axis-inference.md), and
[Interpretation, Prediction, and Induction Are Separate Operations](interpretation-prediction-induction-boundary.md)

---

## The representation question

The requirement is settled: interpretation and action consequence need separate claim
identities, licensed predictions, falsification addresses, and confidence histories. It
does not follow that they need two implementations of probability learning.

Toy #9 chooses the minimum representation that makes fusion impossible:

```text
CONSEQUENCE CLAIM INSTANCE
    existing Expression.Expect
    source identity: responsible action set
    prediction:      expected action/world observation
    learned history: consequence namespace → Edge{P,N}

INTERPRETATION CLAIM INSTANCE
    new InterpretationClaim
    source identity: proposed category/code for this instance
    prediction:      its own expected category-context observation
    licenses:        audited reference to the consequence Expect it selected
    learned history: interpretation namespace → Edge{P,N}
```

Both histories may reuse the existing `Edge` math—EMA, base-rate prior, proper score,
confidence saturation. They do not share a map, key, target, settlement index, or metrics.
This is **one learning primitive with two typed claim paths**, not two confidence variables
on one edge and not two duplicated Go edge implementations.

The interpretation claim is not written into `ObservationGraph` as truth. It lives beside
expressions as an agent-authored falsifiable claim. Its `DerivedFrom` list points to the raw
observations that caused the abduction. Its context prediction is armed when the claim is
placed and can settle only on a later observation sequence.

Admission requires:

- a non-empty category/code identity;
- at least one real raw observation in `DerivedFrom`;
- one existing consequence Expect that the interpretation selected/licensed;
- one mechanically checkable context prediction with a deadline;
- a context falsification address distinct from the consequence address.

Without its own prediction, the interpretation remains a proposal and is not admitted to
the learned claim graph.

---

## The Michael walk, preserved

```text
Michael:    "Can you help me with something?"

Abduction:  "this is Michael-working-on-Skyra"
Deduction:  "so expect traversal talk"
Expression: "Sure — is it about that traversal algorithm from yesterday?"

Michael:    "Hey Skyra, can you help me with a traversal algorithm for school"
```

Before the final impulse, two claims exist:

```text
CONSEQUENCE
    prediction: message:/reply topic traversal
    address:    (message:/reply, topic)
    confidence: 0.90

INTERPRETATION: Michael-working-on-Skyra
    prediction: message:/reply context skyra
    address:    (message:/reply, context)
    confidence: 0.90
    licenses:   the consequence claim above
```

The final message is decomposed into two raw observations in one external span:

```text
message:/reply topic   traversal
message:/reply context school
```

`ObserveImpulse` writes the complete raw impulse before grading either claim. Settlement is
then address-indexed:

```text
(message:/reply, topic) = traversal
    → consequence prediction TRUE
    → consequence history: N 0→1, P 0.50→0.75

(message:/reply, context) = school
    → interpretation prediction "skyra" FALSE
    → interpretation history: N 0→1, P 0.50→0.25
```

With exact channels and stated confidence `0.90`, the pre-update proper scores are:

```text
consequence true:     log(0.90 / 0.50) =  0.588
interpretation false: log(0.10 / 0.50) = -1.609
```

The two-bit record is therefore mechanically:

```text
(consequence = TRUE, interpretation = FALSE)
```

The traversal hit cannot touch interpretation history because the interpretation bet is not
indexed at the topic address. The school contradiction cannot touch consequence history
because the action bet is not indexed at the context address. No judge reads the sentence
and decides which reason was good.

---

## Adversarial matrix

Run all four logical outcomes by varying the two independently observed values:

```text
                         consequence RIGHT       consequence WRONG

interpretation RIGHT     both histories +        interpretation +, consequence -
interpretation WRONG     LUCKY HIT:               both histories -
                         consequence +,
                         interpretation -
```

Also require:

- reversing observation order changes nothing;
- the consequence observation alone leaves interpretation unresolved;
- the context observation alone leaves consequence unresolved;
- unrelated addresses settle neither;
- a mark settles neither;
- interpreted evidence may settle only with its existing fidelity cap;
- an old episode cannot adjudicate against a replacement world;
- a claim cannot be admitted after seeing its own settling observation and then scan
  backward for credit;
- identical context and consequence addresses are rejected at admission;
- missing/raw-invalid `DerivedFrom`, missing consequence references, or missing context
  predictions are rejected;
- two competing interpretations at one context address may settle in opposite directions
  from the same world observation; two candidates predicting the same value receive the
  same bit and are not discriminated.

---

## Why the shared `Edge` implementation is safe

Reusing `Edge` does not fuse the claims. `Edge` is only the counted Bernoulli learner. The
epistemic type lives in the path that owns it:

```text
consequenceEdges[ action-source + target ]
interpretationEdges[ category-code + context-prediction ]
```

No function that grades a consequence receives an interpretation edge, and no function
that grades an interpretation receives a consequence edge. Metrics remain separate as
well. The same arithmetic can therefore be reused without allowing one outcome to update
both histories.

The opposite design—one map entry with `consequenceP` and `interpretationP` fields—would
still fuse identity and settlement routing. Two concrete structs would duplicate reliable
math without adding epistemic separation. Typed paths are the load-bearing middle.

---

## Dimensional correction after the independence proof

Toy #9's exact `Triple` predicates and `TRUE / FALSE` results prove independent ownership
and routing. They must not be mistaken for the final prediction representation. A real
prediction belongs to its span/claim and may stake several scale, structure, and content
dimensions. A fixed parser evaluates the fully written impulse in those dimensions, and the
substrate retains predicted values, actual measurements, residuals/relations, calibration,
provenance, and dimension-local settlement state.

The two-bit projection remains useful:

```text
consequence projection    RIGHT
interpretation projection WRONG
```

But the projection does not replace the two underlying comparison fields. Traversal marker
hits and semantic similarity may explain the consequence result; absent Skyra markers,
present school markers, and their similarity residuals may explain the interpretation
result. Same-claim sensors fuse only for a grade projection, while raw dimensional
measurements remain available to learning.

[Toy #9F — Span-Owned Prediction Fields Preserve the Measurement](toy-9f-span-owned-prediction-fields.md)
is the required follow-on gate. Phase 5 cannot treat this toy's minimum exact-address
evaluator as the ceiling.

---

## What this settles and what remains open

**Settled:** two claim instances, two addresses, two typed history namespaces, independent
settlement, pre-observation arming, raw provenance back to the abductive evidence, and an
explicit link from the interpretation to the consequence it licensed. One impulse may
settle both because it contains two observations, never because one Boolean is copied.

**Open:** how the LLM and Phase-3 pocket selector author the context prediction in normal
runtime; how multiple context predicates aggregate into one category confidence; how long
conversation predicates remain eligible when targeted `Realm.Perceive` is unavailable;
and whether learned interpretation histories should later participate directly in
retrieval weight. Phase 4 builds the honest learning boundary, not the full abductive
authoring policy.

## Result

Toy #9 passes on paper. The lucky hit produces consequence `TRUE` and interpretation
`FALSE` from two addresses in one blind impulse. The representation needs two typed claim
paths and histories, but not two duplicated edge implementations. Interpretation can now be
wrong for the right answer without either fact corrupting the other.

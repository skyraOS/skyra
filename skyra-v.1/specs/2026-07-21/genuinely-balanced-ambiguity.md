# Toy Test #5 — Genuinely Balanced Ambiguity With No Dominant Signal

**Date:** 2026-07-21  
**Status:** NEXT TO RUN  
**Tests:** interpretation closure, held expectation fields, graph-query fixed points,
and behavior when the substrate has no earned basis for collapse.  
**Parent:**
[interpretation-learning](../../src/skyra/skyra-dna-v3/interpretation_learning_spec.md)

## The question

The first four observe-wave paper tests deliberately constructed ties and then made a
new signal earn the right to break each tie:

```text
test 1  span convergence beats isolated heat
test 2  distinct-root diversity beats same-root alternatives
test 3  coverage and joint precedent expose a novel stitch
test 4  witnessed timing and resolved co-occurrence recover segmentation
```

This test removes the next signal too.

> **What does the being do under genuinely balanced ambiguity with no dominant
> signal?**

The correct answer must not be an arbitrary winner produced by iteration order,
lexical order, recency, target leakage, expectation preheating, or an LLM declaring
its own interpretation correct.

## Required symmetry

Construct one impulse with two mutually exclusive readings, `A` and `B`, such that
every currently legitimate discriminator is balanced:

```text
candidate prior              A = B
distinct-root support        A = B
total activation heat        A = B
root coverage                A = B
joint-precedent count        A = B
edge P and N                 A = B
edge recency                 A = B
memory token cost            A = B
world-given context          none that distinguishes them
```

The two candidates must imply different future observations or different useful
actions. Otherwise the ambiguity is behaviorally irrelevant and may be compressed
away.

The fixture should use neutral opaque IDs and then run again with the IDs swapped.
No semantically suggestive name may accidentally serve as a prior.

## Minimal fixture

Two prior spans are mirror images:

```text
S1
  raw impulse pattern: R
  interpretation:      ambiguous-root → candidate:A
  clarifying answer:   signal:a
  eventual outcome:    world-state:a

S2
  raw impulse pattern: R
  interpretation:      ambiguous-root → candidate:B
  clarifying answer:   signal:b
  eventual outcome:    world-state:b
```

Both histories have identical:

- number of observations;
- provenance and fidelity;
- action structure;
- settlement count;
- success rate;
- age/recency;
- rendered size;
- graph distance from the new impulse.

The new impulse `I3` reproduces `R` while omitting `signal:a` and `signal:b` and every
other distinguishing context root.

The blind observation field should therefore contain a balanced distribution:

```text
O(candidate:A) = 0.5
O(candidate:B) = 0.5
```

The exact numeric representation may differ, but the substrate must preserve the
symmetry.

## Interpretation-session walk

### Pass 0 — blind exact/retrieval pass

The incoming impulse runs without seeing an open target, desired reading, or held
expectation field.

Expected result:

```text
bindings:  {A, B}
residual:  one unresolved mutually-exclusive choice
closure:   not singularly closed
novelty:   false — both readings have precedent
```

This is ambiguity, not novelty. The system has seen both readings before; it lacks the
evidence required to choose between them now.

### Pass 1 — graph query

The LLM may issue a constrained query for information that could distinguish `A` and
`B`.

The graph returns equal support because the fixture is symmetric.

Expected result:

```text
retrieval field: balanced
residual:        unchanged
query heat:      attention only, not evidence
closure:         still unresolved
```

The query must not break the tie merely because one posting was visited first.

### Pass 2 — fixed point

Another query with no new discriminating predicate does not reduce the residual.

Expected result:

```text
fixed point reached
stop internal querying
do not spend the remaining budget performing ritual search
```

Queries-to-closure is not defined as success here: the honest result is
queries-to-recognized-underdetermination.

## Correct behavioral outcomes

### If a discriminating act is available

Emit the cheapest reversible act whose predicted observations differ between the
candidate readings.

For a conversation this is normally a clarification:

```text
ask for signal:a versus signal:b
```

The action must carry separate falsifiable expectations:

```text
if A, expect answer/effect a
if B, expect answer/effect b
```

The answer arrives as a new blind external impulse. Only then may one candidate gain
evidentiary authority.

### If no discriminating act is available

Retain the plural reading and either wait or choose an action safe under both
interpretations.

The being must not convert missing evidence into `TruthFalse`, and it must not choose
one branch solely to satisfy a singular output type.

### If an irreversible action requires singular interpretation

Do not act until evidence breaks the symmetry or policy explicitly authorizes the
risk. Ambiguity is itself a state the planner must be able to represent.

## Held expectation-field check

If the interpretation session holds an expectation field, it must preserve the
balanced belief:

```text
Q(A) = 0.5
Q(B) = 0.5
```

The next blind observation produces `O`. Comparison happens only after `O` is frozen.

Required outcomes:

- If the answer supports `A`, the `A` branch receives the corresponding outcome and
  the `B` branch does not.
- If the answer supports `B`, the mirror result occurs.
- If the answer remains ambiguous, neither branch is falsely resolved.
- The original balanced forecast earns approximately the base-rate score; it did not
  know which branch would occur.
- A fabricated confident `0.99/0.01` forecast is punished half the time across the
  mirrored fixtures and must not outperform the calibrated forecast in expectation.

## Metamorphic symmetry test

Run the entire fixture twice:

```text
run 1: IDs and insertion order are A then B
run 2: IDs and insertion order are B then A
```

Before discriminating evidence, the committed state, residual, query count, and chosen
behavioral policy must be equivalent under the rename.

After evidence, the result must mirror with the evidence.

This catches hidden tie-breakers such as:

- map iteration;
- lexical ID ordering;
- insertion order;
- most-recent exemplar selection;
- deterministic parser preference;
- first candidate emitted by the LLM.

Determinism may order the presentation of candidates. It may not turn presentation
order into epistemic preference.

## Broad-query adversary

Allow one query that heats every memory related to either candidate.

Expected result:

- the query costs more than a narrow query;
- it returns equal support;
- it supplies no evidence;
- it does not create closure;
- it cannot improve the interpretation grade merely by increasing total heat.

This pins the difference between cognitive effort and evidentiary authority.

## Pass conditions

The test passes only if:

1. The initial blind field preserves both readings.
2. Equal retrieval cannot force singular collapse.
3. A fixed query residual stops the internal loop.
4. Query heat never becomes observation evidence.
5. The system chooses clarification, a branch-safe action, or waiting—not arbitrary
   commitment.
6. The held expectation remains calibrated rather than inventing confidence.
7. Mirroring candidate IDs mirrors the post-evidence result and changes nothing
   pre-evidence.
8. Later raw evidence can break the symmetry and names the observation that did so.
9. Interpreted evidence remains fidelity-capped.
10. No target or desired answer is visible to the blind observation pass.

## Failure signatures

```text
always chooses candidate A
    → hidden ordering or naming prior

queries until budget exhaustion
    → no fixed-point rail / ritual cognition

one broad query creates certainty
    → activation mistaken for evidence

held expectation changes blind field
    → self-confirming interpretation

returns TruthFalse for the unchosen candidate
    → ambiguity collapsed into falsehood

clarification answer settles an old unrelated span
    → ownership/eligibility failure

confident arbitrary choice beats calibrated 0.5 over mirrored runs
    → scoring or attribution failure
```

## Why this test is next

The preceding toys always contained a signal waiting to be recognized. This one asks
whether Skyra can recognize when no such signal exists.

A learning system that cannot preserve genuine underdetermination will eventually turn
implementation details into beliefs. Before interpretation prediction or graph-query
learning is implemented, the substrate must demonstrate:

> **When reality has not supplied a reason to choose, Skyra knows that it does not
> know—and acts to acquire the missing distinction rather than appointing a winner.**


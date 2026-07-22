# Interpretation, Prediction, and Induction Are Separate Operations

Status: architectural decision  
Date: 2026-07-21  
Scope: v3 interpretation learning and deterministic prediction

## Decision

Skyra must not treat interpretation and prediction as the same operation.

```text
"this is a P"          interpretation — abduction: propose the category
    ↓ licenses
"so expect Q"          prediction — deduction: P implies a consequence
    ↓ world
"Q or not-Q"           settlement — the world supplies the outcome
    ↓
update belief in P     learning — induction: revise category confidence
```

The arrows are provenance boundaries, not merely stages in one model response.

Interpretation is an abductive hypothesis over what the present impulse means.
Prediction is a deterministic consequence licensed by that hypothesis plus known
causal structure. Settlement is an observation from the world. Induction updates
the interpretation only after comparing the licensed consequence with that
observation.

## Why the split matters

If interpretation and prediction collapse into one score, a system can mistake a
fluent guess for a consequence, then mistake its own consequence for confirming
evidence. Keeping them separate exposes exactly where uncertainty lives.

The category may be uncertain while its conditional consequences are exact:

```text
P(A | impulse) = 0.50      if A, deterministically expect Q
P(B | impulse) = 0.50      if B, deterministically expect R
```

There is no need to make deduction probabilistic merely because the abductive choice
is uncertain. Hold both branches. Let the world discriminate between them.

This is especially important for genuinely balanced ambiguity. With no dominant
signal, Skyra must preserve competing interpretations and their separately licensed
predictions. Candidate ordering is not belief and cannot select a branch.

## Required separation

The system must retain four independently attributable records:

1. **Interpretation hypothesis** — candidate category, confidence, and the evidence
   or retrieval path that proposed it.
2. **Licensed prediction** — the consequence derived conditional on that candidate,
   including the rule or causal edge used in the deduction.
3. **World settlement** — the eligible blind observation that yielded `Q`, `not-Q`,
   or an unresolved outcome.
4. **Inductive update** — the change to candidate confidence caused by that
   settlement, with its provenance and fidelity cap.

No record may impersonate another:

- an interpretation is not evidence that its consequence occurred;
- a prediction is not a world observation;
- retrieval heat is neither settlement nor confirmation;
- deterministic derivation does not make its abductive premise certain;
- an unresolved or ineligible observation produces no belief update;
- a failed prediction updates only the premises and causal rules that licensed it,
  with credit assignment preserved rather than spread across the whole graph.

## Multiple candidates

Every live interpretation candidate may license its own predictions before a winner
exists:

```text
candidate A ──deduction──▶ expects Q
candidate B ──deduction──▶ expects not-Q
                              │
                        eligible world act
                              │
                 settles one discriminating fact
                              │
                  inductively reweights A and B
```

This turns ambiguity into an experiment design problem. When passive observation is
unlikely to distinguish the branches, the correct next act is a clarification or
another safe action whose possible outcomes differ across the candidates.

An action is discriminating because the candidates predict different outcomes, not
because an LLM declares that the action will be informative.

## Learning boundary

Induction should consume prediction settlements, not raw similarity between a
candidate and later text. A successful update requires:

- a prediction born before the settling observation;
- a trace from the prediction back to the interpretation that licensed it;
- an eligible observation that was blind to the held expectation;
- an exact or otherwise mechanically defined settlement result;
- comparison against the relevant base rate where confidence is graded;
- fidelity and provenance caps on the resulting update.

Repeated success may crystallize an abductive path into cheap recognition, but it
does not erase the boundary. The resulting competence must still be able to emit
licensed expectations, encounter contrary world evidence, and deoptimize back into
explicit interpretation.

## Consequences for implementation

- Interpretation confidence belongs to candidates, not to their display order.
- Prediction records identify the interpretation candidate and causal rule that
  licensed them.
- The deterministic predictor accepts a candidate as a conditional premise; it does
  not decide whether that candidate is the correct interpretation.
- The grader settles predictions against world observations; it does not reinterpret
  the original impulse.
- The learner updates candidate confidence only from settled, attributable
  predictions.
- Query count and interpretation cost describe abduction; prediction error describes
  the consequences it licensed. Both are needed to measure learning.

## Tests implied by the decision

1. Two equally supported candidates remain equally supported before world evidence,
   even when their deterministic predictions differ.
2. Swapping candidate IDs or insertion order changes neither confidence nor the
   licensed prediction set.
3. A query that repeats the candidate's own prediction cannot confirm the candidate.
4. An eligible observation of `Q` updates the candidate that licensed `Q` and the
   competing candidate according to their prior predictions.
5. An unrelated, late, or self-authored `Q` does not settle the prediction.
6. When both candidates predict the same `Q`, observing `Q` does not discriminate
   between them; it may test the shared causal rule but cannot select a category.
7. When no eligible outcome arrives, confidence remains unchanged rather than
   decaying into an arbitrary winner.

## Compact invariant

> Abduction proposes the premise. Deduction computes what follows if the premise is
> true. The world settles the consequence. Only then may induction revise confidence
> in the premise.


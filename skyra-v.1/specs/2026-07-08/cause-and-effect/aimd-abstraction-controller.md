# AIMD Abstraction Controller

**Status:** RESEARCH DRAFT

## Purpose

Use additive increase and multiplicative decrease (AIMD) to probe how much
specificity the model can currently support. The controlled variable is the
query's token budget, not traversal amplitude and not belief strength.

```text
B = current QueryTokenBudget

supported:    B' = min(B_max, B + alpha)
contradicted: B' = max(B_min, beta * B)   where 0 < beta < 1
unknown:      B' = B
```

Initial tuning candidates are `alpha = 10 tokens` and `beta = 0.5`. They are
starting values, not architecture constants.

Additive increase earns specificity slowly. Multiplicative decrease abandons
unsupported detail quickly and requests a more abstract requery. Bounds prevent
degenerate prompts and unlimited growth.

## Controller Episode

```text
1. Select the current evidence snapshot.
2. Query with budget B; receive a sentence and corresponding subgraph.
3. Preserve that immutable result and its provenance.
4. Grade the resolvable claim as supported, contradicted, or unknown.
5. Update B using AIMD.
6. On contradiction, requery the same evidence at the reduced budget.
7. Leave Efficient/Final belief edges intact unless their own evidence grades
   move their EMA strength toward zero.
```

The controller belongs to substrate state. The LLM sees the immediate query and
budget, not AIMD state, cascade thresholds, or global graph consequences.

Controller state should initially be local to a query lineage. A single global
budget would let failure in one domain erase justified specificity in unrelated
domains.

> **Decided 2026-07-12:** in the minimal loop, "lineage" means **one AIMD
> lineage per compound worker** — present, efficient, and final each carry their
> own budget. The three workers emit separately (present first, then
> efficient ∥ final in parallel), so a contradicted claim taxes exactly the
> lineage that made it.

## Interaction With Existing Mechanics

```text
TraversalAmplitude  -> which nodes can be reached within context
QueryTokenBudget    -> how detailed the emitted query may be
signed edge EMA     -> resonance/tension of a particular relationship
residual error      -> whether a higher layer receives unresolved mismatch
```

These values may influence the same episode but are not aliases. In particular,
a smaller query budget must not reduce the stored strength of Efficient or Final
belief edges.

## Why AIMD Is A Candidate, Not A Borrowed Proof

AIMD is established in congestion control as a way to probe unknown shared
capacity through slow increase and rapid decrease. See
[RFC 5681](https://www.rfc-editor.org/rfc/rfc5681) and the convergence analysis
by [Chiu and Jain](https://www.cs.cmu.edu/~srini/15-744/F02/readings/CJ89.pdf).

The analogy is structurally useful, but TCP's convergence and fairness results
do not automatically prove epistemic calibration here. Query verification is
not packet loss, model behavior may be nonstationary, and reducing output tokens
may cause truncation rather than abstraction. Those properties require local
experiments.

## Alternatives Deferred

- A decayed Beta posterior could map a smoothed support estimate continuously
  to the budget. It adds probabilistic state and requires care because observed
  outcomes may not be independent Bernoulli trials.
- Conformal methods can calibrate prediction sets under their stated exchangeability
  conditions; they do not automatically calibrate arbitrary causal sentences.
  See [Vovk et al.](https://proceedings.mlr.press/v60/vovk17a.html).

Begin with AIMD because its controller is small and inspectable. Keep complete
episode history so a later controller can be replayed against the same data.

## Required Measurements

- budget before and after each grade;
- sentence and serialized subgraph token costs;
- support, contradiction, or unknown outcome with provenance;
- whether a reduced budget yielded abstraction, truncation, or semantic drift;
- number of requeries before local fit recovered;
- residual error propagated beyond the local query.


# Skyra v3 — Biggest Risks and Required Gates

**Date:** 2026-07-21  
**Status:** RISK REGISTER · blocking constraints for the interpretation-layer build  
**Scope:** v3 activation, interpretation, prediction, settlement, memory, and long-running behavior  
**Reads on:** [episodes-over-the-field](episodes-over-the-field.md),
[interpretation-implementation-phases](interpretation-implementation-phases.md),
[interpretation-prediction-induction-boundary](interpretation-prediction-induction-boundary.md),
[the-two-bit-outcome](the-two-bit-outcome.md),
[predicate-diagnosticity](predicate-diagnosticity.md),
[scale-structure-content](scale-structure-content.md), and
[spacetime-kernel-event](spacetime-kernel-event.md).

---

## Purpose

This document names the ways Skyra can appear to learn while becoming less truthful.
Those failures are more dangerous than an explicit inability to solve a task: a failed
agent stops; a falsely-learning agent accumulates confidence in the wrong structure and
uses that structure to select its future evidence.

The governing distinction is:

```text
implemented rail   mechanically impossible to bypass on the live path
specified rail     designed and paper-tested, but still bypassable in code
open problem       named, but no mechanism has earned implementation
```

A written invariant is not an implemented rail. A passing task score is not evidence that
the internal attribution was correct. A dense memory region is not evidence that its
interpretation is true.

The build must preserve one end-to-end dependency:

```text
raw observation
    ↓ proposes
interpretation candidate
    ↓ licenses
precommitted, world-checkable prediction
    ↓ settled by
eligible blind world observation
    ↓ updates only
the claim that licensed that prediction
```

Any shortcut across this chain is a potential self-confirmation path.

---

## Severity and release rule

```text
P0  can silently teach the being a false interpretation or corrupt grading
P1  can force arbitrary behavior, prevent deoptimization, or poison retrieval
P2  can prevent scaling, calibration, or reliable longitudinal operation
```

**Release rule:** no phase may consume a structure operationally before the gate for that
structure passes. In particular:

- Phase 2 may not commit counted targets until Phase 1's structured target is live.
- Phase 3 may not mint durable codes until recurrence is mechanically separated from truth.
- Phase 4 may not claim interpretation learning until the two-bit gate passes.
- Phase 5 may not inherit a neighbor's consequences until direct differential falsification
  and broad-failure attribution are represented separately.

---

## Risk 1 — Lucky-hit corruption

**Severity:** P0  
**Current state:** specified, not implemented  
**Owned by:** Phase 4

### Failure

An action consequence comes true even though the interpretation that selected it was wrong.
If one confidence represents both facts, the successful consequence reinforces the false
interpretation.

```text
interpretation:  this is Michael working on Skyra                 WRONG
prediction:      traversal will be mentioned                      RIGHT
world:           "a traversal algorithm for school"

fused update:    interpretation strengthened                      CORRUPTION
required update: consequence +1, interpretation -1 or unknown     TWO BITS
```

This is the most direct way for good task performance to conceal worsening cognition.

### Existing protection

- Consequence predictions have explicit targets, eligibility, and settlement evidence.
- The two-bit architecture requires independent interpretation and consequence claim paths.

### Missing protection

V3 currently has the consequence edge. It does not yet have an independently settled
interpretation claim that licenses a category-discriminating prediction at a distinct
address.

### Required gate

The Toy #9 fixture must demonstrate all four outcomes without an LLM judge:

```text
action consequence   interpretation consequence
       right                  right
       right                  wrong
       wrong                  right
       wrong                  wrong
```

Removing the interpretation-specific observation must leave interpretation **unknown**,
not infer it from action success. No shared scalar may update both histories.

### Stop condition

Do not describe Skyra as learning interpretations while this gate is open.

---

## Risk 2 — Interpreted observations laundering model output into world state

**Severity:** P0  
**Current state:** partially protected  
**Owned by:** Phase 4 and the decomposition boundary

### Failure

The decomposer reads a raw impulse and writes an inferred triple into the observation graph:

```text
raw:          "Can you grab me a glass?"
interpreted:  michael wants water
```

The second record is model output on the world's side of the boundary. Even at reduced
grading fidelity, it can heat memory, influence retrieval, select an action, and determine
which evidence is collected next.

### Existing protection

- `raw`, `interpreted`, and `mark` provenance are distinct.
- Interpreted observations carry confidence and `DerivedFrom` pointers.
- Interpreted observations have capped fidelity.
- Marks have zero grading fidelity.

### Missing protection

- The interpretation claim itself is not yet graded by later category-discriminating
  observations.
- The decomposer's blindness is an architectural rule but is not yet a hard capability
  boundary at every caller.
- Low grading fidelity does not limit routing influence.

### Required gates

1. A decomposer invocation cannot access open bets, desired targets, confidences, or the
   planner present by type/API construction, not prompt convention.
2. Every interpreted observation retains a complete path to raw evidence.
3. A repeated wrong interpretation loses future routing influence after contrary world
   evidence.
4. Replaying the same raw impulses with permuted candidate IDs yields the same candidate
   field before settlement.

### Stop condition

No interpreted triple may be promoted into an exact fact or durable category label merely
because it was repeatedly retrieved.

---

## Risk 3 — Flat targets turn ambiguity into conjunction or ordering into belief

**Severity:** P0  
**Current state:** present on the live path; Phase 1 is code-ready  
**Owned by:** Phase 1

### Failure

`deriveActivatedTarget` currently returns a flat `[]Triple`. The representation cannot
distinguish:

```text
[A, B] = A AND B
[A, B] = A OR B, unresolved
```

Deterministic lexical ordering is necessary for replay, but an LLM consumer may treat the
first item as preferred. An implementation detail then appoints a winner with no evidence.

### Required mechanism

Return source-owned `TargetGroup`s containing explicit branches, unresolved completions,
and source state. Meaning must be invariant to insertion order and opaque ID renaming.

Legacy projection is allowed only when **every** group is confident. Any ambiguous or
unresolved group makes the legacy `Target` nil.

### Required gate

Phase 1a and 1b must pass the full Toy #5/#6/#6B suite, including:

- balanced alternatives remain plural;
- conjunction stays one branch;
- cross-source evidence cannot merge branches;
- false settlements cannot count as positive co-occurrence;
- a non-transitive compatibility component is unresolved, not greedily partitioned;
- ID swapping and insertion reversal do not change meaning;
- an unresolved component vetoes a dominant clean branch;
- no rendered order is described to the model as preference.

### Stop condition

No caller may execute or frame a flat derived target while any source group is ambiguous or
unresolved.

---

## Risk 4 — Retrieval changes the evidence distribution and becomes self-confirming

**Severity:** P0  
**Current state:** partially protected; selection bias remains open  
**Owned by:** activation, action selection, and experimental action policy

### Failure

Skyra correctly separates heat from evidence, but heat still changes behavior:

```text
precedent heats
    ↓
precedent is retrieved
    ↓
agent repeats its associated action
    ↓
agent observes the region selected by that action
    ↓
more evidence accumulates around the precedent
```

The grader may remain honest while the sampling policy hides alternatives. This is the
agentic form of confirmation bias.

### Existing protection

- Query heat cannot settle a prediction.
- Activation is transient and decays.
- Only non-mark observations with fidelity can grade.
- Every settled prediction names its evidence.

### Missing protection

There is no settled policy for balancing exploitation of reliable precedent against actions
that discriminate live alternatives or inspect cold regions.

### Required gates

1. A balanced fork with an available discriminating act chooses or proposes that act before
   an irreversible branch-specific act.
2. An irreducible fork remains balanced across repeated exposure rather than collapsing due
   only to sampling asymmetry.
3. With one branch artificially over-exposed, correcting for exposure restores the expected
   branch relation.
4. Retrieval heat alone never increases interpretation confidence.
5. Report observation opportunity counts beside positive/negative settlements so "rarely
   inspected" cannot masquerade as "rarely true."

### Open decision

Whether discriminating actions are chosen by explicit expected information gain, a cheaper
branch-disagreement rule, or another measured policy remains open. The LLM may propose acts;
it may not certify that its own act is discriminating.

---

## Risk 5 — Density is mistaken for truth, causality, or utility

**Severity:** P0  
**Current state:** principle stated; Phase 3 mechanism open  
**Owned by:** Phase 3

### Failure

A density pocket proves recurrence in the field. It does not prove that the pocket is:

- a true interpretation;
- causally coherent;
- useful;
- externally grounded;
- free from repeated model error;
- independent of the agent's sampling policy.

A joke, misconception, ritualized failure, or retrieval loop can all become dense.

### Required distinction

```text
density claim:       these materials repeatedly co-activate
interpretation claim: this instance belongs to the named code
consequence claim:   if it belongs, a discriminating world state should follow
utility claim:       using the code reduces cost or error
```

No one claim may update another without a witnessed licensing edge and eligible settlement.

### Required gates

1. A recurring false interpretation forms a detectable dense pocket but does not gain
   interpretation confidence without its own consequence settlement.
2. A dense useless ritual does not gain utility merely from recurrence.
3. A dense association with no diagnostic predicate remains an unnamed region, not a
   semantic fact.
4. Naming or renaming a pocket does not change its historical grades.
5. Removing LLM-generated labels leaves the density split unchanged.

### Stop condition

Phase 3 may mint a code for compression before it mints a trusted interpretation. Code
existence and code truth must remain separate records.

---

## Risk 6 — Broad or adaptive predicates make interpretations unfalsifiable

**Severity:** P0  
**Current state:** diagnosticity objective specified; landed fitter still reuses selection
evidence, but its ability to create, override, or veto a commit is now blocked  
**Owned by:** Phase 3 predicate fitting and Phase 4 settlement

### Failure

A category prediction such as "Skyra-context will appear" can be widened until nearly any
continuation satisfies it. It then looks reliable while carrying almost no information.

An embedding threshold or LLM classifier in the grading path creates a second failure: the
meaning of the check can drift after the prediction was made.

### Existing design response

- Predicate proposals may use any tool.
- Runtime evaluation must be frozen, deterministic, and LLM-free.
- Candidate predicates are scored against blind held-out data by diagnosticity:

  ```text
  coverage × log(coverage / base_rate)
  ```

- Predicate/version identity changes when its model or threshold changes.

### Confirmed Phase-3 failure

The current exact-feature fitter searches candidates and computes their diagnosticity on
the same branch occurrence spans. The held-out current span only tests whether the chosen
feature fires; it does not independently test whether that feature predicts branch
membership. `PromotionEvidence` is configured but unused by this path. A maximum selected
from many sparse candidates previously received immediate operational authority. It now
survives only as provisional match metadata; promotion remains unavailable.

### Required gates

1. A match-all predicate earns approximately zero.
2. An exact but nearly never-firing predicate loses through insufficient coverage.
3. Candidate selection and evaluation use disjoint evidence.
4. `unknown` and "predicate did not get an observation opportunity" remain distinct from
   false.
5. Threshold, embedding model, normalization, and tool versions are part of predicate
   identity.
6. Sparse-data promotion uses support bounds, not a maximum point estimate selected from
   many candidates.

### Open problems

- cold-start priors and confidence bounds;
- multiple-comparison cost when searching many predicates;
- deoptimization after diagnosticity drift;
- discovering when multiple sensors measure the same claim.

---

## Risk 7 — Source identity aliases different situations or fragments equivalent ones

**Severity:** P1  
**Current state:** conservative non-aliasing key landed 2026-07-21; learned equivalence open  
**Owned by:** Phase 1 edge/source identity

### Failure

Ambiguity and co-occurrence are conditioned on an act-signature source. If that identity is
too coarse, different situations collide and create false alternatives or conjunctions. If
it is too fine, equivalent experiences never accumulate enough evidence to differentiate.

The audited `edgeSourceKey` contained only `Primitive → Target`. It now also contains a
SHA-256 payload digest and sorted canonical requirements, so materially different writes or
differently guarded actions do not share counts by omission.

### Required gates

1. Construct two contexts with the same primitive/target acts but different necessary
   conditions; prove they do not contaminate each other's completion groups.
2. Construct syntactically different but behaviorally equivalent act sequences; measure
   whether useful evidence is fragmented.
3. Any source-identity revision is versioned and preserves provenance to the old identity.
4. Report collision and fragmentation diagnostics before tuning thresholds around their
   artifacts.
5. The conservative initial key includes primitive, target, canonical payload or payload
   digest, and sorted canonical requirements; any later fusion is learned and versioned.

### Open decision

The minimum sufficient conditioning context for an edge source is not settled. Do not hide
this representational question inside `ConjunctionOverlap` or `DominanceMargin` tuning.

---

## Risk 8 — Thresholds encode the ontology

**Severity:** P1  
**Current state:** separate configuration specified; estimators open  
**Owned by:** Phases 1, 3, and 5

### Failure

Small threshold changes can cause category explosion, forced merging, false confidence, or
permanent unresolved regions. A single shared threshold would silently make statistically
different decisions equivalent.

### Existing design response

`ConjunctionOverlap`, `DominanceMargin`, `SplitValleyDepth`, and promotion support are
separate named quantities. Permutation and boundary fixtures are required.

### Required gates

1. Sweep each threshold independently over a declared range and publish the phase diagram
   of resulting states.
2. Small perturbations away from fixture boundaries must not change classification.
3. Boundary cases must remain explicitly unstable/uncertain rather than flip silently.
4. Support count or confidence intervals accompany overlap and dominance point estimates.
5. Threshold changes mint a configuration/version identity so replay remains auditable.

### Stop condition

Do not tune thresholds on the same fixtures used to claim discrimination.

---

## Risk 9 — Nonstationarity preserves obsolete competence

**Severity:** P1  
**Current state:** partially protected  
**Owned by:** edge learning, activation, and episode transition

### Failure

The expression graph survives world episodes by design. A formerly reliable procedure may
therefore remain active after the world changes. Retrieval can keep it warm, slowing or
preventing deoptimization.

### Existing protection

- Edge probability uses a plastic EMA with an alpha floor.
- Retrieval use/recency decays.
- Observation graphs are per-world.
- Predictions cannot be settled by a replacement Realm.

### Missing protection

- explicit change-point/drift evidence;
- context-conditional validity for old procedures;
- proof that repeated retrieval cannot indefinitely preserve a failing precedent;
- predicate and environment-version migration.

### Required gate

Run a reversal fixture:

```text
world v1: procedure A succeeds repeatedly
world v2: rule changes; A fails and B succeeds
world v3: original context returns; A is useful again
```

Required behavior:

- A is initially retrieved in v2, fails, and loses influence.
- B takes over without deleting A's witnessed v1 history.
- A can recover in v3 under the restored context.
- Global success must not overwrite context-specific failure.

---

## Risk 10 — Bounded present hides unbounded durable growth

**Severity:** P1  
**Current state:** retrieval bounded; storage lifecycle open  
**Owned by:** memory and indexes

### Failure

The activation field and token budget bound what reaches inference, but the durable
observation graph, expression graph, indexes, edge histories, and settlement-span postings
continue growing.

Long-lived operation eventually faces:

- index and replay cost;
- duplicated expressions;
- obsolete predicate versions;
- graph repair after bad writes;
- archival and deletion semantics;
- schema migration;
- backup/recovery;
- evidence retained after its derived summary is removed.

### Required gates

1. A million-event synthetic soak maintains bounded per-impulse activation and present
   construction cost.
2. Settlement-span windows remain bounded without changing recent classification.
3. Archiving a cold span preserves every grade's evidence trace.
4. Rebuilding all materialized indexes from the durable log yields the same active present.
5. A failed migration is recoverable without deleting source observations.
6. Duplicate or equivalent expressions do not gain dominance by multiplicity.

### Open decision

The durable log may be append-only while indexes and summaries are disposable. The exact
archival boundary is not settled, but evidence required to audit a grade may never be
silently discarded.

---

## Risk 11 — Correlation is mislabeled as causation

**Severity:** P1  
**Current state:** measured conditional edges exist; causal identification remains limited  
**Owned by:** expression edge semantics and experimental action selection

### Failure

An edge estimates:

```text
P(world state W | expression E was run)
```

That is a measured conditional. It is not automatically the causal effect of E. Context,
selection policy, hidden prerequisites, and concurrent acts may explain W.

Joint expectations improve attribution but do not eliminate confounding.

### Required gates

1. A confounded fixture where E and W share a hidden prerequisite must not be described as
   proof that E causes W.
2. Individual and joint expectations retain independent histories.
3. An act receives causal credit only for predictions attributed to it before settlement.
4. Where safe interventions are available, interventional evidence is labeled separately
   from passive conditional evidence.
5. Reports use "measured conditional" unless an identification criterion has actually been
   satisfied.

### Stop condition

Do not let causal vocabulary outrun the evidence stored on the edge.

---

## Risk 12 — Aggregate improvement hides local epistemic damage

**Severity:** P1  
**Current state:** metrics are promising but insufficient  
**Owned by:** evaluation

### Failure

Passes-to-closure and prediction error can improve while a subset of interpretations becomes
systematically wrong. Expression-level mean score can also hide a mixture of excellent and
catastrophic claims.

### Existing protection

V3 reports passes-to-closure, resolved prediction accuracy, and edge over base rate.

### Required additions

Report at least:

- consequence and interpretation grades separately;
- calibration by confidence bucket;
- unknown/void rate;
- grade by provenance/fidelity;
- worst-decile edge performance, not only means;
- ambiguity, unresolved, and forced-commit counts;
- reversal/drift recovery time;
- retrieval interference: performance with and without retrieved precedent;
- memory growth and present-construction cost;
- evidence opportunity counts;
- rate of grades lacking complete provenance (must be zero).

### Required gate

An adversarial run must be able to improve task success while corrupting interpretations;
the metric suite must expose that divergence rather than report one successful curve.

---

## Risk 13 — Cold-to-counted handoff becomes premature automation

**Severity:** P1  
**Current state:** Phase 2 specified; coldness estimator open  
**Owned by:** Phase 2

### Failure

If "not ambiguous" is mistaken for "mature," a sparse region may be treated as counted
confidence. If support is allowed to dominate too early, one or two accidental successes
can remove the LLM before the substrate has learned anything reliable.

Conversely, if the LLM remains in control after counted structure is dense, the being never
graduates and continues paying for—and being exposed to—model judgment.

### Required four-way routing

```text
COLD        insufficient counted support            → LLM judges provisionally
CONFIDENT   dense, dominant, no unresolved material  → counted commit
AMBIGUOUS   dense, multiple non-dominant branches    → clarify/discriminate/wait
UNRESOLVED  dense, unsafe partition                  → clarify/wait; never commit
```

### Required gate

Toy #7 must show a monotone graduation across support while preserving all four routes.
Dense confidence cannot be mistaken for cold merely because it raises no ambiguity signal.
Unresolved material must veto a neighboring confident branch.

---

## Risk 14 — The world instrument defines what can ever be learned

**Severity:** P2  
**Current state:** explicit affordance registry; natural-language instrumentation open  
**Owned by:** Realm, affordances, and predicate learning

### Failure

The filesystem world has exact predicates such as `exists`, `contains`, and `status`.
Open-world interpretation rarely does. If Skyra cannot construct a stable, blind,
world-checkable address for a claim, then the claim cannot be settled honestly.

The danger is responding by quietly reintroducing an LLM judge.

### Existing protection

- The world stamps claim mode; the model cannot choose its grade multiplier.
- Unmeasurable targets are standing compasses rather than exact bets.
- Predicate diagnosticity keeps embeddings and models on the proposal side.

### Required gates

1. Every runtime predicate exposes a deterministic evaluator and version identity.
2. The evaluator cannot access the expected answer.
3. An unavailable instrument produces unknown, not false.
4. Standing goals generate concrete sub-targets but never grade themselves.
5. No new domain is called supported until its observation vocabulary can settle both
   positive and negative claims.

---

## Risk 15 — Durable edge counts are mistaken for durable interpretation recall

**Severity:** P1  
**Current state:** confirmed across episode replacement  
**Owned by:** episode transition, durable learned codes, activation retrieval, and Phase 3

### Failure

`NewEpisode` preserves `ExpressionGraph` but replaces the observation graph, span index,
and activation field. This correctly prevents a prior world's observations from becoming
the current world's truth. It also means ordinary impulse retrieval and density feature
reconstruction can no longer reach the old episode's contextual observations. Counted edge
statistics survive, but the interpretation structure that made those counts usable is
generally cold or unreconstructable.

Calling this "durable interpretation learning" would conflate stored probability with
retrievable learned structure.

### Required gates

1. Learn a discriminating pocket and its consequence histories in episode A.
2. Replace the world and all per-world observation indexes.
3. Present the corresponding context in episode B through normal `ObserveImpulse`, without
   an exact hand-authored query.
4. Retrieve the learned code and its provenance without treating episode-A facts as
   episode-B observations.
5. Preserve enough compressed feature and calibration statistics to apply and revise the
   code; do not retain the entire old world merely to make recall pass.

### Open mechanism

A durable learned-code/pocket index appears necessary. It must sit between ephemeral
per-world observation truth and global raw-log retention: compact enough to survive world
replacement, but explicit about which historical observations earned it.

---

## Risk 16 — Resolving-span effects masquerade as decision context

**Severity:** P1  
**Current state:** fixed in Phase-3 source-span sampling; adversarial regression landed  
**Owned by:** prediction ownership, density examples, and Phase 4F comparison fields

### Failure

The audited Phase-3 path took every raw observation in a branch-settling span except the exact
candidate completion triples. A resolving span can also contain errors, downstream effects,
or other consequences that were unavailable when the action or interpretation was chosen.
Those effects can become high-diagnosticity pocket names and later be presented as causal
context.

Excluding the target prevents direct answer leakage. It does not establish that remaining
features predate the prediction.

The landed correction records prediction-time source-span membership separately from the
resolving span and fits density features only from the source side. Resolving observations
remain evidence for the outcome, never retroactive context.

### Required gates

1. Every density training example explicitly pairs a source/decision feature span with a
   later settlement label.
2. Feature eligibility is frozen at prediction birth through `BornFrom` evidence or an
   equivalent sequence boundary.
3. Adding a perfectly branch-correlated fact only after settlement cannot change the fitted
   predicate or pocket geometry.
4. Adding the same fact before prediction may change the fit, with provenance showing why.
5. The comparison record preserves source evidence and settling evidence separately.

### Stop condition

Do not promote a resolving-span feature as an interpretation predicate merely because the
candidate outcome itself was excluded. The system must prove the feature was available on
the propose side of the boundary.

---

## Cross-cutting invariants

The following must remain mechanically queryable:

1. **No grade without evidence.** Every resolved claim names its settling observation.
2. **No future evidence at birth.** The claim predates every observation eligible to settle it.
3. **No self-authored world proof.** Marks never settle claims about the world.
4. **No retrieval-as-belief.** Heat, rank, density, and query count never directly update confidence.
5. **No unknown-as-false.** Lack of an eligible observation is neither success nor failure.
6. **No shared lucky-hit update.** One outcome cannot update two claims unless it independently
   addresses both.
7. **No hidden winner.** Ordering and opaque IDs carry no belief semantics.
8. **No mutable historical predicate.** A changed evaluator receives a new identity.
9. **No world replacement settlement.** Evidence from one Realm cannot adjudicate another.
10. **No aggregate-only diagnosis.** Every global metric decomposes to claims and evidence.
11. **No post-outcome context.** A feature used to select an interpretation existed before
    the prediction it conditions.
12. **No accidental action abstraction.** Different payloads or required guards do not
    share a source identity until evidence earns their equivalence.
13. **No stored-but-unreachable learning claim.** Durable learned structure remains
    retrievable after per-world indexes are replaced.

An integrity scan should run these as graph queries, not code-review conventions.

---

## Required longitudinal test program

Paper toys establish local discrimination. They do not establish lifespan behavior. Before
calling the system continually learning, run four longer classes of tests.

### 1. Poison test

Repeatedly supply a plausible but wrong interpretation. Confirm that it may become familiar
and dense while its interpretation confidence falls through independent world evidence.

### 2. Drift test

Change a previously reliable world rule. Confirm deoptimization, replacement learning, and
context-specific recovery without global forgetting.

### 3. Ambiguity test

Maintain a genuinely irreducible balanced fork for many episodes. Confirm that increasing
support raises confidence in the **fork**, not in an arbitrary branch.

### 4. Growth test

Run enough events to force archival, index rebuild, predicate versioning, and duplicate
pressure. Confirm bounded working cost and complete grade provenance.

For every test, publish both task performance and epistemic integrity metrics. A task pass
with a failed integrity query is a failed run.

---

## Phase blockers

```text
PHASE 1 blocked by   none; implement the paper-passed structured target and span index
PHASE 2 blocked by   Phase 1 live shape + Toy #7 coldness/graduation gate
PHASE 3 blocked by   Toy #8 audit: disjoint predicate evaluation, sparse promotion,
                     cross-episode recall (commit block, decision-time features,
                     and conservative source identity have landed)
PHASE 4 blocked by   Toy #9 two-address/two-bit representation
PHASE 4F blocked by  Toy #9F remaining structure/regex/semantic, calibration,
                     void-carrier, and residual-history rails; minimum field landed
PHASE 5 blocked by   Toy #10 direct differential falsification and attribution
```

The kernel `Spacetime` migration is orthogonal unless it changes observation ownership,
return perception, or causal eligibility. If it does, all evidence-boundary gates must be
rerun.

---

## Bottom line

Skyra's largest risk is not that the LLM makes a bad prediction. Bad predictions are visible
and learnable.

The largest risk is that the substrate records the wrong reason for a successful outcome,
then uses that false reason to choose what it sees next. That failure can improve task scores,
increase memory density, reduce planning passes, and still make the being less truthful.

The defense is not a better judge. It is a complete witnessed chain:

```text
claim before outcome
    + independent falsification address
    + blind eligible observation
    + claim-specific update
    + preserved unknown
    + auditable provenance
```

Until every interpretation crosses that chain, Skyra has consequence learning and an
interpretation hypothesis—not interpretation learning.

# Settlement and Surprise — Two Signals, Two Jobs

> ## ⚠ SUPERSEDED IN PART — 2026-07-26
>
> **The split stands** — settlement grades, the diff gates, and §5's argument for why the
> diff must not replace settlement is untouched. What changes is settlement's *outcome type*.
>
> ```text
> R2  §2/§3 describe settlement as a deterministic compare producing TRUE|FALSE.
>     WRONG. settlement(B) = ΣW over B's window — a signed displacement (definitions §3e).
>     §3's chain survives with one substitution: "deterministic compare → Settlement"
>     becomes "interpret the return → accumulate agreement."
>     §6's "absence IS falsity within a targeted perceive" SURVIVES, and improves: each
>     empty sweep contributes a negative term instead of one verdict at expiry.
>
> R6  §7's open metric question is CLOSED IN THIS DOC — calibration carries "prediction
>     error"; GradeSum is the headroom curve. no repair needed.
> ```

**Date:** 2026-07-24
**Status: THE SPLIT IS DECIDED.** Per-expectation settlement stays as the credit and
falsification unit; a window distribution diff is added as a separate surprise and coverage
signal. **The divergence measure, the window definition, and whether the diff ever produces a
grade are OPEN** (§8) — this doc exists to be fleshed out.

**Reads on:** [definitions](definitions.md) §3a–§3b (the two graded arms, and which curve
carries "prediction error" — §7's metric question is answered there),
[belief-expression-inference](belief-expression-inference.md) (the two calls),
[interaction-primitives](interaction-primitives.md) §7 (projected/observed phase — the
prerequisite), [v4-data-model](v4-data-model.md) §3 (individual settlement, the credit walk),
[time](time.md) §4 (the sensory organ that sweeps open addresses), [workers](workers.md) (the
existing surprise gate this replaces).

---

## 1. The hole this closes

> **Nothing currently measures what the being failed to anticipate.**

Every learning signal is **bet-gated**. An expectation names an address, the settler perceives
that address, and a grade follows. No bet at an address ⇒ no perceive ⇒ no settlement ⇒ no
signal. So a being that bets narrowly and safely is never penalized for the world being full
of things it did not model.

The docs already brushed against this and patched it the weaker way —
[workers](workers.md):

> *"Surprise alone is degenerate at cold start: a being that has predicted nothing can never
> be surprised, so it would go quiet exactly when it knows least."*

That fix bolted on a second trigger (incompleteness). The better fix is to **make surprise
itself non-degenerate** by measuring divergence over a *region* instead of agreement at a
*point*.

Note the shape of the bias, because it is self-reinforcing:

```text
bets steer perception → perception makes observations → observations make beliefs
     → beliefs ground the next bets → bets steer perception
```

---

## 2. The two signals

```text
PER-EXPECTATION SETTLEMENT      CREDIT + FALSIFICATION
  individual · addressed         which claim failed · who gets attributed
  targeted perceive              absence IS falsity, within the address
  drives edge P and credit walk

WINDOW DISTRIBUTION DIFF        SURPRISE + COVERAGE
  regional · heated              how much of what happened did I not see coming
  projected-phase vs observed    catches material NO bet covered
  gates attention and planning
```

> **The diff decides when to think. Settlement decides what to learn.**

They are not competing implementations of one idea. Trying to make one mechanism do both jobs
is what makes the design feel unresolved.

---

## 3. Per-expectation settlement — and why it stays individual

Unchanged from [v4-data-model](v4-data-model.md) §3. The chain:

```text
Expectation{address, claimed p, Within}
    → organ perceives the address across the window
    → deterministic compare → Settlement{Resolution: TRUE|FALSE, Grade, ResolvedBy}
    → g = log(p_claim) − log(p_prior)        graded against the prior it HAD
    → edge.Observe(outcome)                   the frequency estimate moves
    → credit walk: By[]Act → Expression (running mean) → Beliefs (activation-at-assembly)
                    → triples (partial) → Observation
```

**Individual, not collective, and the reason is diagnostic power:**

> **SETTLEMENT INDIVIDUAL.** … Forced by the two-bit rule — independent falsifiability. …
> If the set were the settlement unit, **cells 2 and 3 of the two-bit table become
> unreachable — and cell 3 is the lucky hit, the failure most worth catching.**

and

> *"`Predicts` stays singular … three claims tell you **which** failed; one conjunctive claim
> only tells you **something** did."*

No model call anywhere on this path. That is LAW 4 as a call graph.

---

## 4. The window distribution diff — the heated diff

The signal is a **field operation**, not a new store, because
[interaction-primitives](interaction-primitives.md) §7 already wants phase on the excitation
field:

```text
PROJECTED heat over the window     what the being expected to be live
OBSERVED  heat over the window     what actually landed
        divergence                 SURPRISE, over a region, not a point
```

**What it catches that settlement structurally cannot:** regions that lit up which *nothing
projected*. Under settlement, unbet material produces no signal at all. Under the diff it
produces the **largest** divergence — which is the correct sign, because material you did not
model at all is exactly what you most need to notice.

**It upgrades the existing gate.** Planning currently fires on `|grade| > θ`, computable only
over bets already placed — structurally blind to the case this doc is about. The diff makes
the surprise trigger non-degenerate, and `incompleteness` stops being a patch for surprise's
blindness and goes back to meaning only what it says.

**Phase is a hard prerequisite for THIS form.** Without projected/observed on the field there
is nothing to diff. This is the second independent thing that requires §7 — the first was the
settlement-eligibility surface in [v4-data-model §4a](v4-data-model.md).

**A phase-free alternative form: `ΔB`.** The belief-field delta
([definitions §4](definitions.md#4-dynamics--δb)) measures surprise as movement of the belief
set — `BORN` beliefs (novelty) and `‖MOVED‖` (revision) — over what beliefs *did*, not over
projected-vs-observed heat. Because it never touches the field's evidence phase, it **does not
require §7**, which makes it the more buildable candidate for this signal. Open which form
wins; `ΔB` has the fewer dependencies.

---

## 5. Why the diff must not replace settlement

A window-level distribution diff **is a conjunctive claim**. It reports that a region
diverged; it cannot report that the act succeeded while the interpretation was wrong.

That is the lucky hit — cell 3 of the two-bit table, the failure the entire two-axis design
exists to catch. Replacing settlement with the diff would make it unreachable, undoing a
deliberate and well-argued decision for a signal that answers a different question.

```text
diff says     "this region surprised me"
settlement    "THIS claim, at THIS address, was wrong — and here is who is accountable"
```

Attribution needs the second. Attention needs the first.

---

## 6. What the diff cannot inherit

Settlement's deepest property does not generalize:

> *"a targeted perceive comes back authoritative for exactly the claim bet on, so within it,
> absence IS falsity. The bet manufactures its own evidence. The general closed-world problem
> never has to be solved."*

**Absence in a region is not falsity.** It is ambiguous between *"did not happen"* and *"did
not look."* Only the addressed perceive collapses that ambiguity.

Consequently:

```text
the diff widens the training sample     from "addresses I bet on" → "addresses I observed"
it does NOT debias it                    observation is still partial
```

That widening is a real gain — observations arrive unbidden through impulses and act effects,
so the observed region is strictly larger than the bet region. But the diff is a measure over
what was *seen*, and must be read as such.

**This is why the diff gates rather than grades.** A signal that cannot distinguish "absent"
from "unobserved" must not move an edge EMA, because `P` is a frequency estimate and feeding
it unobserved-as-false would corrupt every future grade scored against it.

---

## 7. What each signal's curve should look like — WHICH METRIC IS RESOLVED; magnitude open

The two signals have different expected shapes, and one of them is counter-intuitive.

**`GradeSum / Resolved` is the model's marginal information over the graph** —
`Σ (log p_claim − log p̂_edge)`, per settled bet. Watch it as the system succeeds:

```text
COLD   edge N=0 → prior 0.5 · claim 0.80, TRUE   →  +0.470
WARM   edge P=0.85 · claim 0.85, TRUE            →   0.000   parroting a CORRECT graph
WARM+  edge P=0.85 · claim 0.95, TRUE            →  +0.112   real but shrinking headroom
```

As the edge converges, the headroom to beat it shrinks. The expected shape is an **inverted
U**: rise, peak, decay.

### The metric has an exact identity — and a floor that is not zero

> **Corrected 2026-07-24.** An earlier draft said this metric "drives toward zero." That
> holds only when the model has **no** instance-specific information. The true floor is the
> mutual information, derived below.

With the edge converged to base rate `q`, and the model reporting honestly on an instance
whose true probability is `qᵢ`:

```text
E[grade | instance i] = qᵢ·ln(qᵢ/q) + (1−qᵢ)·ln((1−qᵢ)/(1−q))
                      = KL( Bern(qᵢ) ‖ Bern(q) )
```

Averaging over instances, and since the edge converges to the marginal `q = E[qᵢ]`:

```text
E[ GradeSum / Resolved ]  →  Eᵢ[ KL( P(Y|Xᵢ) ‖ P(Y) ) ]  =  I( outcome ; model's side-information )
```

**The metric converges to the mutual information between what the model sees and what
happens**, in nats. So the decay runs from *(base-rate information + instance information)*
down to *(instance information alone)* — the graph absorbs the countable part, and what
remains is exactly the model's irreducible contribution.

That makes the asymptote the headline number, and falsifiable in both directions:

```text
converges to a POSITIVE constant   the model contributes instance-level information that
                                   no amount of counting would produce
converges to ZERO                  the counting graph could have done it alone
```

The second is a real, sharp negative result — arguably the most important thing this
experiment can report. The metric is well-founded; what is open is its expected magnitude and
how many reps it takes to reach the floor.

That collides with how the thesis is stated in `doc.go`:

> *passes-to-closure decreases AND **prediction error decreases**. Both curves down is the
> thesis.*

`GradeSum` is **credit** — higher is better — so "down" reads as failure, and "down" is
exactly what a converging system produces. **The two readings give opposite success
criteria.**

Probable resolution — they are different quantities and both are wanted:

```text
GradeSum / Resolved   marginal information over the graph   inverted-U. near-zero = converged.
calibration error     |p_claim − outcome| style              monotone ↓, independent of
                      (v3 has prediction_calibration.go)      whether the graph has converged
coverage (the diff)   unanticipated share of what landed     should ↓ as the being models
                                                              more of its world
```

*"Prediction error decreases"* is honestly served by **calibration**, not by `GradeSum`.

> **RESOLVED 2026-07-25 — in favour of calibration.** `GradeSum` is not the prediction-error
> curve and never was; it is the **headroom** curve. Prediction improving looks like
> calibration error falling *while* `GradeSum` decays, and reading that decay as failure is
> the trap. `GradeSum`'s approach to its floor is therefore **not** a failure condition and
> must not be stated as one — what is falsifiable is the *floor's* value (§7's identity:
> positive ⇒ instance-level information, zero ⇒ counting could have done it alone).
> Full statement, including why the two graded arms reach their floors at different times:
> [definitions §3a–§3b](definitions.md).

**Still open here:** the expected *magnitude* of the floor, and how many reps it takes to
reach it — which is what makes [time §3e](time.md)'s baseline distribution a prerequisite for
reading either curve.

This bears directly on the baseline ([time](time.md) §3e): a non-monotone metric can only be
compared **at equal rep counts**, and its expected shape has to be known before anyone can
say a port preserved behavior. Otherwise a converged run and a broken run look alike.

---

## 7a. Conditioning the baseline on activation — a calibration curve

**Status: PROPOSED, not decided.** The estimator is worked out; the premise is untested and
the grading use is gated. Recorded 2026-07-24.

### The idea

The edge stores one scalar, `P(Y = 1 | edge invoked)`, which pools every activation level into
one homogeneous population. If activation strength systematically predicts truth rate, that
scalar discards information. The conditional version:

```text
P(Y = 1 | A = a)        a calibration curve, not a single base rate
grade = log P(Y|X) − log P(Y|A)
```

### The estimator — one global slope, per-edge intercept

Do **not** estimate activation bins independently; that quarters the evidence per edge and,
given `confidence = N/(N+5)`, would leave most bins unmeasured. Fit a curve instead, so
sparse points share strength:

```text
logit P̂(Y | edge, A)  =  logit(P_edge)  +  β · (A − Ā)
```

- `P_edge` — **the scalar the edge already stores**, reinterpreted as the rate at mean
  activation. No new per-edge state.
- `β` — a **single global parameter** fitted across every settled bet in the system.
- `Ā` — mean activation-at-arm.

Three properties make this the cheap version:

```text
β = 0 recovers TODAY'S SYSTEM EXACTLY     ⟹ ship it dark; measuring β IS the experiment
logistic is linear in log-odds            ⟹ slots into the log-score grade unchanged, still proper
interpolation falls out of the FORM       ⟹ no coverage needed at every activation level
```

That last point is the answer to sparsity: with truth rates observed at `A = .06` and
`A = .18`, the fit yields a value at `A = .12` without ever having seen data there.
Isotonic regression is the more flexible successor if the linear-in-logit form proves too
stiff; it costs more data.

### Two uses, different value and different risk

```text
GRADING BASELINE   P(Y|A) as the prior the model must beat
                   sharpens the identity in §7 to  I(Y ; X | A)  — conditional MI.
                   the model is paid only for what NEITHER counting NOR field
                   geometry provides. a strictly harder, more honest test.

COLLAPSE PRIOR     P(Y|A) as a truth estimate for material with no settled outcome.
                   converts activation from an UNCALIBRATED RANKING SCORE into a
                   CALIBRATED PROBABILITY — usable for the leverage cut, for ranking,
                   and for giving inference calibrated confidence on held ambiguity
                   instead of raw heat.
```

Their failure modes are not symmetric, which sets the order:

```text
grading    contamination moves EDGES — permanent, and corrupts every future grade scored against P
collapse   contamination biases WHAT IS SHOWN — a retrieval problem, and
           "a bad retrieval decays where a bad merge does not"
```

**So the collapse use ships first.** It pays immediately and its worst case decays.

### Blocker: activation currently contains the being's own claim

The caution is not hypothetical here. [interaction-primitives](interaction-primitives.md) §5:

> `Place` bumps the field when the being **emits** … an observation landing and the being's
> own emission are the **same primitive at the same amplitude** … the concrete mechanism
> behind the self-fulfilment leak: **predicting something excites it.**

Conditioning any baseline on raw activation would let the baseline absorb exactly the
information being credited. **Condition on OBSERVED-phase heat only, never projected** — which
makes this the **third** dependent on §7 phase, alongside the settlement-eligibility surface
and the window diff (§4).

Two further constraints:

1. **Activation must be stamped at arm/emission time**, not read at settlement — settlement-time
   heat is influenced by the settling observation, i.e. by the outcome. Precedent exists:
   `Expression.Beliefs map[BeliefID]float64 // activation AT ASSEMBLY TIME`.
2. **`P`'s double duty breaks.** `edge.go` states that one number serving as both grading prior
   and `RetrievalWeight` is deliberate. A conditional-on-activation `P` cannot be the retrieval
   weight — retrieval would rank by a quantity conditioned on the activation retrieval itself
   produces. Circular. Marginal-for-retrieval and conditional-for-grading would have to split.

### The population caveat — and how the diff resolves it

The curve is fitted on **settled bets**: activation-at-arm paired with an outcome. The collapse
use applies it to arbitrary heated material, most of which was never bet on.

```text
fitted on    things the model CHOSE to claim, at their arm-time activation
applied to   any heated material, claimed or not
```

That extrapolates outside the fitted distribution — and it is the **same sampling bias** §1
identifies, since bets steer perception.

It is checkable rather than fatal: the window diff (§4) produces outcomes for material nobody
bet on, which is precisely the held-out sample needed to test whether the curve transfers.
**The two proposals support each other** — the diff supplies the unbiased sample the curve must
be validated against.

### Recommended order

```text
1. stamp activation-at-arm on the Expectation; log it with the outcome     no behavior change
2. fit β, inert. if β ≈ 0 the premise is false and this section is dropped
3. if β ≠ 0: use the curve at COLLAPSE first (decaying failure mode)
4. gate the GRADING use on §7 phase (observed-only heat)
5. validate transfer against the diff's unbet outcomes
```

Step 2 is the whole test, and it costs one logged field.

---

## 7b. Two axes — attention is ordinal, truth is cardinal

**Status: VOCABULARY DECIDED 2026-07-24.** The display quantity in §7b.3 is proposed and
answers an open question elsewhere; flagged where it is not yet settled.

### The distinction, and why it is not pedantic

```text
RANKING / ATTENTION    descent · search · spreading activation
                       ORDINAL — only relative order matters
                       an UNCALIBRATED score is fine here

TRUTH                  ascent · display decision
                       CARDINAL — the number must mean something absolute
                       REQUIRES calibration (§7a)
```

> **Ranking is a search term. Importing it onto the truth axis is a category error.**

This is what makes the calibration curve non-redundant. A monotone curve leaves *ranking*
unchanged — order-preserving transforms preserve order — so if truth were used for ranking,
`P(Y|A)` would buy nothing over raw `A`. It is not used for ranking. It is used for absolute
decisions, and **an uncalibrated score cannot support an absolute decision.** "Activation
0.12" licenses nothing; `P(truth) = 0.15` licenses a threshold.

### 7b.1 Where each is computed — SETTLED

```text
PERTURB   produces A. NOTHING ELSE. no probabilities, no grades, no truth values — just heat.
          attention spreads L0 → L1 → L2, ranking all the way down. terminates at nil.
OBSERVE   applies P(T|A) to the arrived heat and PRODUCES THE TRUTH VALUE.
          post-order, so each layer decides for itself what to realize.
```

**This is the definition of what Observe does at each layer, not a side computation.**
Calibration is an Observe operation — Perturb could apply the curve but would then carry a
truth value down through layers that do not use it; Observe applies it exactly where the
show / drop / uncertain decision is made. A probability never crosses the descent. See
[definitions](definitions.md#where-each-is-evaluated--the-phase-boundary).

Nothing new is traversed — the ascent already visits every node with its activation known.

**Truth ESTIMATE, not truth.** `P(T|A)` is derived from a contaminated field (§5 leak), so
Observe produces a prior for what to *show*, never a settlement of what is *true*. Only the
world settles.

The dependency chain per layer, in descent terms:

```text
belief truth   ← belief activation   ← reverse-index injection from the observation layer
act truth      ← act activation      ← reverse-index injection from the belief layer
```

Each `Observe` stage computes the truth value for its own layer from the attention that
reached it.

### 7b.2 The recursion has a floor — Observations are not derived

`Observation.Fidelity` is the channel's quality — mechanical `1.0`, inferential `~0.3` —
**assigned at observation time, not inferred from the field.** And `PerceivedAct` is
unfalsifiable by construction. So the recursion starts at **L1 beliefs**, with fidelity as the
ground reference.

*(Parked: fidelity as an error signal — the one truth-like quantity assigned by the channel
rather than derived from the field, which makes it the natural reference to measure derived
values against. Not worked through.)*

### 7b.3 The display quantity is `p(1−p)`, not `p` — PROPOSED

`belief-expression-inference` §5 already states what belongs under a grounding header:

> **HIGH-VARIANCE OBSERVATIONS** — exact token spans that remain **surprising, disputed,
> sparse, or poorly explained**

So display wants what is **uncertain**, not what is **probably true**. A "high `p` survives"
rule would surface exactly the settled material the model already knows, spending context on
confirmation.

The reconciling quantity is Bernoulli variance over the calibrated truth:

```text
p = P(truth | A)
display = p(1 − p)        maximal at p = 0.5 · minimal as p → 0 or 1

p ≈ 1.0   settled. the model already knows it. showing it wastes tokens.
p ≈ 0.5   maximally uncertain. where an action can still gain information.
p ≈ 0.0   probably false — a SUPPRESS case, not a show case. distinct from the above.
```

**This gives a closed form for [belief-expression-inference](belief-expression-inference.md)
§10 #1** — *"the exact variance measure for Observations in the action present"* — which was
open. Proposed, not settled: the `p ≈ 0` suppress branch is a separate decision from the
`p(1−p)` show branch, and nothing has measured whether the curve is accurate enough near the
tails to act on either.

### 7b.4 How truth and the budget cut coexist

The context window forces a ranked fill, so truth cannot be the only mechanism. Two stages,
each on its own axis:

```text
TRUTH      a FILTER    cardinal · absolute · drop the settled and the likely-false
LEVERAGE   the RANKER  ordinal · relative · fill the remaining budget
```

Filter cardinally, then rank ordinally. This keeps the separation intact all the way to the
payload, and it stops `leverage` from having to secretly encode a probability in order to work
— which is what an ordinal-only formulation would have forced.

### 7b.5 The guard

> **An activation-derived truth value is a PRIOR, never a SETTLEMENT.** It may inform display,
> filtering, and leverage. It may never close a claim, move `edge.P`, or defeat a belief.

It is derived from the field's own geometry, not from the world. Only a world-derived
observation settles. This is the third instance of the same shape — *grouping may confer
attention, never settlement rights* — and the wording is deliberate, because *"compute the
truth value of belief A"* is language that drifts toward settlement if nobody pins it.

### 7b.6 Costs carried over from §7a

```text
one curve PER LAYER      Y differs — a belief's Y is its own falsification address resolving;
                         an expectation's Y is the predicted triple holding. different base
                         rates, different data volumes. beliefs settle far more rarely.
compounding error        chaining calibrations down the layers multiplies fitting error
§5 contamination         every layer's activation is partly self-authored until phase lands
```

---

## 8. Open — the things to flesh out

1. **What is the divergence measured over?** The full observed region, or only the projected
   support? This decides what the signal *is*:
   ```text
   over the full observed region   COVERAGE   "how much that landed did I not project"
   over the projected support only CALIBRATION "where I did project, how wrong was the shape"
   ```
   These are different measures and the system probably wants both.

2. **What bounds the window?** Candidates: wall duration, tick count, the union of open
   `Within` windows, or the current perturbation. Each gives a different meaning to "during."
   Note [time](time.md) §3d — `Within` is wall-denominated; the field decays per impulse.

3. **Does the diff ever produce a grade, or only a gate?** §6 argues gate-only, because it
   cannot separate absent from unobserved. If it ever grades, what may it move — expression
   credit, belief credit — and never `edge.P`?

4. **What replaces or joins θ?** The existing gate is `|grade| > θ`. Does the diff supply a
   second threshold, or does it subsume θ entirely?

5. **Is the diff computed per-wave or maintained continuously?** A field operation could be
   evaluated on demand at gate time, or accumulated as the field evolves.

6. **Divergence function.** KL, cross-entropy, or something cheaper. Whatever it is, if it
   ever touches credit it must stay a proper scoring rule — the property that currently makes
   hedging unprofitable.

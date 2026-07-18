# Grading

**Status:** DESIGN · part of [skyra-dna-v3](README.md) · implemented shape in `src/skyra/skyra-dna-v3/edge.go`

## What a grade is

```text
an EXPECT is a claim:        (triple, confidence)  — "this act will make this true, at 0.85"
an OBSERVATION resolves it:  EXACT MATCH — the decomposer already interpreted it on the way in
the GRADE is the residue:    it folds into the EDGE, and ResolvedBy names the evidence
```

A grade is **the accuracy of one specific claim, measured against one specific observation.** Not a
verdict on an episode. Not a score for a turn. A number attached to a bet.

## The expect is the measurement. The edge is what learns.

**An expect resolves once.** `file:/src/auth.go exists true` at 0.95 → observed → done. It is never
re-graded. So "accuracy over time" is not a property of an expect, and an expect carries no running
economy.

What repeats is the **edge**:

```text
E ──────→ W        P(W | E)   "given this expression, does this world state come back?"
```

That is the learned object — **an estimate of a conditional probability**, updated by counted
outcomes. Not a strength, not an affinity, not a judge's opinion of coherence.

### Joints are not a special case

An edge's **source is a set**. Cardinality does all the work:

```text
{open-door}              → door:front  state     open              P(W₁ | E₁)
{tell-kathy}             → being:kathy heard     "walk through"    P(W₂ | E₂)
{open-door, tell-kathy}  → being:kathy location  beyond:door       P(W₃ | E₁ ∧ E₂)   ← ONE edge
```

`of="a1"` vs `of="a1 a2"` **is** the edge's source cardinality. Nothing new is needed.

Which is exactly why *"open door + told her ⟹ she walks through"* can be **wrong** while both
individual edges stay **right**. Three edges, three EMAs, three independent things to be wrong about.

## The EMA

```text
p̂ ← p̂ + α(outcome − p̂)          outcome ∈ {0, 1}
α = max(α_floor, 1/(N+1))
```

`1/(N+1)` converges to the sample mean. **The floor keeps it plastic** — a pure sample mean assumes a
*stationary world*, and the world is not stationary. Old evidence has to be able to wash out.

> **`TruthUnknown` MUST NOT move the EMA.** An unresolved expect is not a failed one — the being
> simply did not look. Treating *"I didn't check"* as *"it didn't happen"* would drag every edge in
> the graph toward zero. This is the single most destructive bug available in this design, and it is
> why every expect carries a **`within`** — a bet with no deadline is never settled, and an
> unsettled bet is either a lie waiting to happen or a fact nobody went to get.

## The score — one formula, both properties

> **`grade = log(p_claim) − log(p̂_edge)`**, evaluated at the outcome that occurred.

The edge's EMA **is the prior**. Which means the being is graded on **beating its own base rate**:

```text
edge 0.70   claim 0.70   W true     →   0.000    parroted the graph. added NOTHING.
edge 0.70   claim 0.95   W true     →  +0.305    saw something specific about THIS instance.
edge 0.70   claim 0.20   W FALSE    →  +0.981    called an upset. huge credit.
edge 0.70   claim 0.95   W FALSE    →  −1.792    overconfident against history. brutal.
silence (no claim)                  →   0.000    says nothing, learns nothing.
```

Strictly **proper** (hedging is punished honestly) *and* **information-weighted** (trivia and silence
earn nothing) — in one formula. `information gain` and `no coverage penalty` stop being two
mechanisms and become the same one.

Clamp `p ∈ [ε, 1−ε]` so the log cannot blow up.

### What this makes the model's job

The graph knows *"builds pass 70% of the time after this kind of edit."* Only the model can look at
**this** diff and say *"…but this one touches the auth interface, so 0.2."*

> **The LLM only earns credit for information the graph did not already have.**

That is precisely what a language model is good at, and precisely what a graph cannot do.

### But the model is NEVER TOLD ANY OF THIS

**The prompt does not describe the scoring rule.** Not the multiplier, not the prior, not the fact
that beating the base rate is what pays. It asks for an honest probability and says nothing about
consequences.

```text
the KERNEL   does not know what a weight MEANS
the MODEL    does not know what its confidence is USED FOR
```

Three reasons, and each alone is sufficient:

**It replaces calibration with strategy.** The incentive-compatibility theorem assumes an
expected-utility maximizer. An LLM told *"you are scored on beating the base rate"* does not become
calibrated — it **roleplays a score-maximizer**, reasoning narratively about the rule instead of
about the world, and deviating from history *for its own sake*. That is contrarianism, not
information.

> **A calibrated model beats the base rate automatically whenever it actually knows something. Asking
> for it is how you stop getting it.**

**It contaminates the only metric that matters.** `Score` is the check on whether the model is
earning its keep. A model that *knows* it is being tested on its edge over the base rate is no longer
being tested on anything. Goodhart, arriving through the prompt.

**And the rule will change.** α, τ, the fidelity weights, the retrieval formula — all of it gets
tuned. A prompt encoding the scoring rule invalidates every prior run each time it moves.

Enforced by `TestPromptDoesNotLeakTheScoringRule`, which fails on any of *scored / punish / reward /
earn / paid / base rate / beat / credit / grade* appearing in the prompt — because this "improvement"
will be attempted again by whoever next reads the prompt and finds it under-specified.

### And it gives you a falsifiable check on the architecture

A framing accumulates `Score` = the running mean of `(log p_claim − log p̂_edge)` over every expect it
has ever made — **the model's edge over the base rate.**

> **If Score hovers near zero across all framings, the model is parroting base rates and is not
> earning its keep.**

That is a test, not a reassurance. Watch it from the first run.

## Channel fidelity

```text
weighted = grade × Mode.Weight()

exact       ~1.0    file / exit code / window state — free, instant, unambiguous
interpret   ~0.3    the decomposer INTERPRETED an observation. noisy, but real.
(no source)  0      FORBIDDEN — LAW 2. no observation → no grade.
```

Provenance as a **coefficient**, not a field. A grade read off the filesystem and a grade inferred
from someone's mood are both legitimate, and **neither gets to pretend it is the other.**

## The fields — smaller than v2's

```text
ON THE EDGE  (expression → world state, or {expressions} → world state)
  P     EMA of P(W | E).  THE PRIOR for scoring, AND the retrieval weight.
  N     trials. sets α (learning rate) and confidence(N).
  F     recency, lazily decayed. stale edges fall out of retrieval.

ON THE FRAMING
  Score  the model's edge over the base rate. ≈ 0 means the model is adding nothing.
```

**`A` is gone** — it was *"coherence of node→framing, settle-written"*: a judge's opinion, which is
the thing v3 exists to delete. **`P` takes the same slot in the arithmetic but is computed from
counted outcomes.**

**`S` is gone** — it belonged to the settle's economy. Nothing in v3 asks for it back.

Retrieval ranks by:

```text
retrieval_weight = P × confidence(N) × F(now)
```

Reliability, weighted by how much evidence backs it, decayed by staleness. `confidence(N)` is what
keeps **untested** apart from **wrong**: an edge with `N=0` is not a bad edge, it is an *unmeasured*
one, and its `P` must not be ranked as though it means anything yet.

---

## The diff

```text
1. EXACT PASS (free, no model)
   For each expect, look up Key() = subject∥predicate∥object in world state.

2. THERE IS NO STEP 2.
   Interpretation already happened — at OBSERVATION time, in the decomposer (LAW 4).
   By the time a bet is settled, the comparison is a plain exact match. THE MODEL CANNOT
   TOUCH THE GRADING PATH.

3. FIVE OUTCOMES
   matched       predicted, observed, same value           → score it (true)
   wrong-value   predicted X, observed Y for same (s,p)    → score it (false)
   missing       predicted, PERCEIVED, not there           → false  (the targeted perceive
                                                             is authoritative for its own triple)
   pending       predicted, not yet matured                → UNKNOWN. wait. DO NOT SCORE.
                                                             the `within` timer will settle it.
   spurious      observed, not predicted                   → additive: NEWS, no penalty
                                                             contradicting: SHOCK (below)
```

Note there is no "never resolved" outcome. **Every bet matures**, and at maturity the substrate goes
and looks. `UNKNOWN` is a waiting room, not a destination.

### Absence is not falsity — and `within` is what solves it

If the being predicted `file:/src/auth.go exists true` and world state says *nothing* about
that file, **that is not a failed prediction. It did not look.**

Treating *"I didn't check"* as *"it didn't happen"* would drag every edge in the graph toward zero.
This was the last blocker on the grader.

**The maturity timer settles it.** Every expect carries a `within` (see
[expression-grammar.md](expression-grammar.md)). At maturity, the substrate emits a **perceive at
exactly `(subject, predicate)`** — mechanically, no model call, because the expect already names its
own target.

```text
timer fires  →  perceive(subject, predicate)  →  observation  →  resolve normally
```

A **targeted** perceive comes back **authoritative for the triple that was bet on**. So within it,
absence *is* falsity — and the general closed-world/open-world problem never has to be solved,
because the being goes and looks at precisely the thing it claimed.

> **The bet manufactures its own evidence.**
> **Making a bet obliges you to go look. `within` is when.**

No scope annotations on every observation. No special grader semantics. `TruthUnknown` still never
moves the EMA — it just stops being a permanent state, because every bet has a deadline after which
it gets settled.

And this gives LAW 5 teeth: not merely *"every act must carry a falsifiable expect"* but ***"every
bet gets actively settled."*** The being **cannot leave claims hanging**, and it becomes curious
about its own open bets for free.

**V0 is strict:** whatever the perceive returns resolves the bet. A still-running build scores false
and the being learns to bet longer. Crude, honest, collects data. (See the `τ` note in
[expression-grammar.md](expression-grammar.md) — the log-score-against-base-rate rule already
punishes lazily long deadlines, once the edge learns typical timing.)

## Incompleteness is not contradiction

```text
PURELY ADDITIVE     the graph GREW. nothing believed is now false.
                    → INCOMPLETE. the permanent condition of being finite.

REORGANIZING        triples you ASSERTED are now invalidated.
                    → WRONG. that is the learning signal.
```

Punishing incompleteness punishes the being for the world being *detailed*. A `.bak` file it did not
foresee does not make its model **wrong** — it makes it **partial**, and it will always be partial.

**There is no coverage penalty.** The log-score-against-prior already makes silence worthless; it does
not additionally need to be punished.

## Surprise shock — the slow path

```text
purely additive          world grew           → update. no shock.        FAST
small contradiction      a few beliefs flip   → regrade, adjust EMAs.    FAST
severe reorganization    many beliefs flip    → SHOCK: revise, replan.   SLOW
```

```text
shock = Σ over invalidated triples ( prior_confidence × grade_weight )
```

A 0.95 belief collapsing is an earthquake. A 0.30 belief collapsing is a Tuesday. That is Bayesian
surprise, and it is the right gain term: the slow path fires when **strong commitments** break, not
when the world merely surprises at the margins. It makes the watcher's response **proportional** — it
fires *as hard as it was wrong.*

This is the trigger two earlier documents were waiting for (the two-speed runtime; *"only an
exceptional positive surprise delta may change the material/formal summary"*). Both described the
regime change. Neither had a way to detect it.

## Open

- ~~SCOPE~~ — SOLVED by the `within` timer. A targeted perceive is authoritative for its own triple.
- **Shock threshold: fixed or adaptive?** A being in a novel world is surprised constantly; a fixed
  threshold would pin it in the slow path forever. Probably adaptive to a recent-surprise baseline.
- **Retrieval is not graded.** If precedent P keeps being retrieved and keeps leading to failure,
  nothing punishes P — the failures land on the *new* expression's claims. Directly-placed precedent
  is self-correcting (its own edges are re-run and re-scored); **adapted** precedent needs a
  `DerivedFrom` link and a discounted back-propagation, or retrieval becomes a ratchet: precedent
  gets used forever on the strength of grades it earned once and never re-earns.
- **α_floor, confHalf, λ** are guesses. First things to tune against real runs.

# Predicate Diagnosticity: Learning the Consequence Check, Not Hand-Writing It

**Date:** 2026-07-21  
**Status:** RESEARCH NOTE · the separation between predicate proposal and boolean
runtime settlement is settled. Selecting predicates by held-out diagnosticity is the
settled mechanism direction. Sparse-data estimation, non-fire accounting, and search
selection bias remain open implementation details.  
**Extends:**
[The Two-Bit Outcome: Interpretation and Prediction Are Separately Falsifiable](the-two-bit-outcome.md) ·
[Two-Axis Inference: Interpretation Grades Through the World, Not Itself](two-axis-inference.md)

**Extended by:**
[Scale, Structure, Content: A Composable Grading Structure for Predicted Observations](scale-structure-content.md).
It applies base-rate-relative surprise across tool types, fuses redundant sensors by
claim, and makes predicted length the budget that licenses structural and content
predicates.

## Where this sits

The two-axis notes separated interpretation from deterministic prediction. Abduction
proposes a category. Deduction produces a consequence `Q`. An eligible world
observation settles `Q`, and only the claim whose prediction was settled receives the
update.

Those notes established two constraints:

```text
grading is deterministic
grading contains no LLM judgment
```

They did not yet say what `Q` concretely is when the observable world is human
language. Exact string equality is deterministic, but ordinary utterances rarely
repeat byte-for-byte. A consequence that is technically falsifiable but almost never
fires cannot teach the system.

This note answers the remaining question:

> How does Skyra obtain a world-checkable consequence predicate that survives messy
> expression without moving an LLM into the grading path?

The answer is to learn the predicate by diagnosticity, then execute it as a pure
boolean check.

**Settled:** the arrived observation grades a previously licensed predicate without
an inference call.  
**Settled direction:** the predicate is selected by measured diagnosticity rather
than permanently hand-written.  
**Open:** robust estimation when history is sparse or the candidate search is large.

## Determinism is in the check; interpretation is in predicate selection

`Q` is a deterministic boolean predicate over observable tokens or features:

```text
Q(observation) → {true, false}
```

The predicate may be assembled from any tool whose runtime result is reproducible:

- exact string equality;
- regular expressions;
- keyword or marker sets;
- partial token matches;
- normalized structural fields;
- an embedding-similarity threshold;
- conjunctions, disjunctions, and bounded combinations of these.

At grade time, the predicate fires or it does not. There is no model call and no
semantic judge:

```text
licensed Q + eligible observation
        ↓
deterministic evaluation
        ↓
true, false, or unknown if the address was not observed
```

The fallible part is the choice of predicate. Deciding that a marker set or threshold
captures `traversal talk` is an abductive proposal. Skyra does not certify that choice
at proposal time. It measures how the predicate behaves across settled history and
keeps or revises it over time.

This preserves the epistemic boundary:

```text
creative/tool-rich side        deterministic/world-graded side

propose candidate Q      ───▶  evaluate Q against observations
                                 count where it fires
                                 compare with base-rate firing
                                 retain or reject from measured history
```

Determinism therefore lives in the check. Interpretation lives in selecting the
check. Confusing those locations either makes language too brittle to settle or puts
an LLM back in the judgment seat.

## The tightness spectrum

Predicates occupy a spectrum between exactness and breadth.

### Q1 — exact string

```text
msg == "traversal"
```

This is falsifiable by construction. It is also almost useless. Real instances such
as `a traversal algorithm for school`, `walk the graph`, or `the traversal pass` do
not equal the string. The predicate is highly precise but rarely covers the category.

### Q2 — keyword or marker set

```text
contains any of {
    "traversal",
    "graph walk",
    "depth-first",
    "breadth-first"
}
```

This remains deterministic and is loose enough to fire across ordinary phrasing. It
is the intuitive sweet spot: a checkable approximation that retains discriminating
power.

### Q3 — interpretive category

```text
"is about traversal"
```

This states what Skyra means, but it is not itself a deterministic predicate. Asking
an LLM whether the observation is about traversal merely moves interpretation into
the grader.

The goal is the deterministic analogue of Q3: discover a Q2-style predicate that
fires reliably on real category instances without requiring an LLM at match time.

Tightness alone cannot select it. Q1 wins on precision while losing nearly all useful
coverage. Coverage alone cannot select it either; a predicate that matches everything
has perfect coverage and no meaning.

## The toolbox loophole

Allowing the proposer to use regexes, token sets, embeddings, thresholds, and boolean
composition creates a direct attack surface. A sufficiently elaborate predicate can
contain enough alternatives and loose thresholds to match almost anything:

```text
contains "traversal"
OR resembles any of 200 examples above threshold 0.31
OR contains any graph term
OR contains any school term
OR ...
```

It remains mechanically deterministic. It may even be written as an auditable
program. But if it fires on nearly every observation, it is unfalsifiable in practice.
Every outcome appears to confirm the category, so lucky-hit corruption returns dressed
as determinism.

An explicit complexity ban is the wrong primary safeguard. It couples safety to a
particular predicate language and prevents genuinely useful compositions. The
safeguard must make looseness self-defeating regardless of which tools produced it.

## Diagnosticity discounts broad predicates

A fire carries information only to the degree that the predicate was more likely to
fire inside the category than on observations in general.

For a predicate `Q`:

```text
credit_per_fire = log(
    P(fire | category-right)
    /
    P(fire | base rate)
)
```

This is the same likelihood-ratio shape as the existing compression/proper-score
grade, not a separate penalty system. A broad predicate has base-rate firing
probability near `1`, so even a successful match says little and its log-ratio
approaches zero. A tight predicate with a low base rate makes a surprising claim, so
a fire can carry substantial information.

The conditioning set `category-right` must contain independently settled historical
instances. LLM-labeled examples would let interpretation manufacture the data used to
validate its own predicate.

Per-fire credit alone favors predicates too tight to be useful. Therefore candidate
selection also measures coverage:

```text
coverage(Q) = P(fire | category-right)

positive-fire expected credit per real category instance
    = coverage(Q) × log(coverage(Q) / base_rate(Q))
```

This is the positive-fire contribution of the log-likelihood ratio. It expresses the
specific tradeoff under study: a useful predicate must fire on the category and must
be unusual outside it.

The full treatment of non-fires is not silently assumed. A complete binary proper
score may also include:

```text
(1 - coverage) × log((1 - coverage) / (1 - base_rate))
```

Whether predicate selection uses only positive-fire utility or the full Bernoulli
information remains to be paper-tested against Skyra's `unknown ≠ false` settlement
semantics. The worked result below concerns the stated positive-fire objective.

## Worked example

Take a toy stream of `100` settled past observations:

```text
20 genuinely traversal-talk
80 not traversal-talk
```

Three candidate predicates behave as follows:

```text
              fires on        fires on          base-rate
              traversal(20)   non-traversal(80) P(fire)

Q_A exact        6                0              0.06
Q_B keyword     18                7              0.25
Q_C embedding   20               45              0.65
```

Their category coverage is:

```text
Q_A: P(fire | traversal) =  6 / 20 = 0.30
Q_B: P(fire | traversal) = 18 / 20 = 0.90
Q_C: P(fire | traversal) = 20 / 20 = 1.00
```

Per-fire credit is:

```text
Q_A: log(0.30/0.06) = log(5.00)  = 1.609
Q_B: log(0.90/0.25) = log(3.60)  = 1.281
Q_C: log(1.00/0.65) = log(1.538) = 0.431
```

Per-fire reward favors tightness. `Q_A` makes the most informative individual match,
but it misses fourteen of the twenty real instances. Weight by coverage to obtain
expected positive-fire credit per real category instance:

```text
Q_A: 0.30 × 1.609 = 0.483
Q_B: 0.90 × 1.281 = 1.153   ← maximum
Q_C: 1.00 × 0.431 = 0.431
```

The sweet spot is `Q_B`, and it falls out of the arithmetic as an inverted U:

```text
too tight
    surprising when it fires
    but coverage kills it

diagnostic middle
    fires reliably on real instances
    remains unusual outside the category

too loose
    always fires
    but diagnosticity kills it
```

The maximum is the tightest predicate that still reliably covers the real category.
No hand-written rule declares keywords intrinsically superior to equality or
embeddings. The history of where each candidate fires decides.

**Settled result for this objective:** neither precision nor coverage alone selects
the useful predicate. Their information-weighted product selects the diagnostic
middle in the toy.

## The match-anything predicate earns almost nothing

Now allow a gamed predicate assembled from enough tools to cover all traversal cases
plus `70` of the `80` non-traversal observations:

```text
Q_evil fires on traversal:      20 / 20 = 1.00
Q_evil fires on non-traversal:  70 / 80 = 0.875
Q_evil base-rate fire:          90 / 100 = 0.90
```

Its per-fire and expected credit are:

```text
Q_evil: log(1.0/0.90) = 0.105
expected credit = 1.0 × 0.105 = 0.105
```

It covers every real instance and still loses badly to `Q_B`'s `1.153`. Elaborateness
self-defeats because every added clause that catches unrelated observations raises
the base rate and drains information from a match.

There is no explicit `do not be elaborate` rule. The proposer may use any tool. The
only durable path to credit is to produce a predicate that is diagnostic.

This safeguard holds at the population level or on genuinely independent evaluation
data. It does not make finite-sample overfitting impossible. If the proposer searches
thousands of predicates and selects the maximum on the same small window, some monster
may win by chance. Candidate generation and grading therefore require a blind held-out
window, bounded search accounting, or later prospective validation. That is an
estimation rail, not a restriction on which predicate tools are legal.

## Embeddings belong on the proposal side

Embeddings are useful here without becoming epistemic judges.

```text
proposal-time tool:
    similarity(embed(observation), centroid) ≥ threshold

runtime result:
    true or false
```

Once the embedding model, reference vector, preprocessing, distance metric, and
threshold are frozen into the predicate version, evaluation is deterministic for
that version. No LLM decides whether the match feels correct.

Threshold breadth is priced like any other looseness. A permissive threshold fires
on unrelated observations, raises the measured base rate, and loses diagnostic
credit. A threshold too strict loses coverage. The same search selects the useful
middle if one exists.

Versioning is load-bearing: changing the embedding model or threshold changes the
predicate and must mint a new identity or witnessed revision. Otherwise historical
fire counts refer to a check that no longer exists.

Embeddings therefore remain on the propose side, behind the falsifiability gate. They
help compose a boolean consequence predicate; they never grade the consequence.

## Settled mechanism

The mechanism direction is:

1. The AI proposes several candidate predicates for a category, using any tools.
2. Skyra fires each predicate against a blind held-out window of past observations.
3. For each predicate it measures:

   ```text
   coverage  = P(fire | category-right)
   base rate = P(fire | any observation)
   ```

4. It scores the positive-fire objective:

   ```text
   score = coverage × log(coverage / base_rate)
   ```

5. It keeps the maximum as the learned consequence predicate for this category.
6. Runtime grading becomes pure boolean predicate evaluation with no LLM call.
7. It periodically re-measures the predicate; when diagnosticity decays, Skyra
   reopens abduction and proposes replacements.

In compact form:

```text
abduction proposes                    creative, tool-rich
        ↓
blind base-rate measurement grades    deterministic, cheap
        ↓
diagnostic search selects              learned, not hand-written
        ↓
runtime predicate fires or not         deterministic, LLM-free
```

The sweet-spot predicate is selected by search. It is not authored into the ontology
as permanent semantic law.

What is settled here is the separation of roles and the diagnosticity objective. The
final estimator must still respect independent evaluation, `unknown ≠ false`, version
identity, and the open sparse-data problem below.

## Open problem: cold start

The formula requires enough settled history to estimate both:

```text
P(fire | category-right)
P(fire | any observation)
```

Early estimates are noisy. A single accidental match can make a predicate appear
perfectly diagnostic, and selecting the maximum across several candidates magnifies
that noise. The apparent sweet spot can jitter as observations arrive.

The provisional stance is conservative:

> When data is thin, prefer tight predicates with high potential per-fire
> information. Require evidence before loosening them for coverage.

This prefers missed learning opportunities over indiscriminate confirmation. It is a
safety posture, not yet a formal solution.

Still open:

- smoothing or Bayesian priors for both probabilities;
- minimum positive and negative support;
- confidence bounds rather than point estimates;
- how to price the number of predicates searched;
- whether selection uses positive-fire utility or full binary information;
- how much prospective evidence is required before promotion;
- how quickly a once-diagnostic predicate may deoptimize after drift.

Until those are settled, a new predicate should remain provisional and its apparent
diagnosticity should not be mistaken for mature confidence.

## Implementation checkpoint

The positive-fire objective now has a Phase-4F implementation. A calibration window names
raw observations that predate the candidate prediction and points to category truth through
a separate admitted prediction's already-settled comparison. The candidate cannot label
itself, declared selection attributions cannot overlap the evaluation observations, and a
settling impulse cannot recalibrate its own information value. Every dimension retains the
support counts behind coverage and base rate plus the resulting per-fire PMI and
coverage-weighted expected credit.

The worked 20/80 arithmetic is an executable regression, including `Q_evil`. Calibration
retains counts in four explicit states: insufficient support, measured inert, measured
anti-diagnostic, and calibrated. Zero-probability cases remain finite; only calibrated
positive PMI derives `Ready()`. This is not yet autonomous blind-window allocation: the caller must genuinely
withhold the named evaluation window from candidate generation, and undeclared exposure
cannot be detected from the graph. Full non-fire scoring, confidence intervals, search-size
pricing, drift, and promotion remain open exactly as above.

## Bottom line

The consequence check is deterministic and LLM-free, exactly as required. The
predicate itself is learned by diagnosticity rather than hand-written.

Determinism lives in the check. Interpretation lives in predicate selection. The AI
may use any tool to propose candidate checks, including embeddings, because the world
still grades their frozen boolean behavior.

The base-rate discount closes the toolbox loophole at the objective level: a
match-anything predicate makes its own matches unsurprising and therefore nearly
worthless. Tightness loses through coverage; looseness loses through diagnosticity;
the useful middle is selected from witnessed behavior rather than asserted meaning.

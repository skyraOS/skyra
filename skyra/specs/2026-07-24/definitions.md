# Definitions — Attention, Truth, Grade

**Date:** 2026-07-24
**Status: FORMAL DEFINITIONS.** The three quantities are settled. Where a term inside a
definition is itself open (the weight `w(t)`, the baseline choice, the slope `β`), it is
marked OPEN in place — the *definition* is fixed even where a *parameter* is not.

**Reads on:** [settlement-and-surprise](settlement-and-surprise.md) (the two-axis split and
the calibration curve), [interaction-primitives](interaction-primitives.md) §9 (the field and
the §5 emission leak), [belief-expression-inference](belief-expression-inference.md) (the
inference calls truth and attention serve), [edge.go](../../src/skyra/cognitive-agent/skyra-dna-v3/edge.go)
(the grade and EMA as they exist).

---

## 0. The foundation — the measurement law

> ## THE MEASUREMENT LAW
> **The runtime admits only the measurable. A claim that cannot be placed on the
> falsification–truth scale is rejected from measurement. Measurement is a POSITION on that
> scale — never a binary verdict.**

The center, stated in measurement language:

> **The being proposes · only the measurable is admitted · the world measures · nothing
> measures itself.**

### The falsification–truth scale — for beliefs

A belief's measured truth is a position on one scale, not a TRUE/FALSE flag:

```text
0.0 ── falsified ─────────── uncertain ─────────── true ── 1.0
                               (0.5)
```

A single settlement is still a **binary event** — the world said so, at this address, at this
time (this preserves individual settlement; do not lose it). Binary events **fold** into a
continuous position. *The event is binary; the measurement is the scale.*

### Measurable ≡ falsifiable ≡ has a world-address

To be measurable is to have a place the world can report on. No address → no possible reading
→ unmeasurable. But "rejected" is scoped:

```text
ATTENTION path     open to ALL — measurable or not, everything can be heated and ranked
MEASUREMENT path   only the measurable — only an addressed claim can move the accumulator
```

An unmeasurable interpretation is rejected **from measurement**, not from thought — it may
still be attended (HELD). This is the same line as *grouping confers attention, never
settlement* and *nothing measures itself*, now named by measurability.

### Three states — `UNKNOWN` was doing two jobs

```text
UNMEASURABLE           no address exists    rejected from measurement (attend only)
MEASURABLE, UNMEASURED  addressed, N=0       admitted · NO OPINION (position 0.5)
MEASURED               addressed, N>0       a POSITION on the scale (P, backed by N)
```

`prior()=0.5 at N=0` is *"measurable, no reading yet,"* not *"unknown."*

### Two axes, kept apart

```text
POSITION    WHERE on the scale    P               a belief at 0.5 because it is 50/50 …
CONFIDENCE  HOW SURE               N/(N+5)          … is not one at 0.5 because it is unmeasured
```

### The crank turns the scale

One enduring belief accumulates measurements into its scale position. **That accumulation is
the crank — one turn per tick** (a tick is one turn). Everything else is a **READ** of the one
accumulated history — the nouns proliferate because reads get named as if they were things:

```text
THE THING     a belief, enduring, with its accumulated position-history      ONE object
THE CRANK     fold a measurement in                                          ONE write · per settlement
THE READS ────────────────────────────────────────────────────────────────────────────────
   LEVEL        confidence · P · baseline           "where is it now"
   CHANGE       ΔB · stability · convergence         "how is it moving"
   CONDITIONED  calibration P(T|B,A)                 "how does it depend on context"
   BOUNDARY     grade                                the ONLY read that moves the accumulator
```

The internal reads (level, change, conditioned) inform attention, display, and timing and
**never touch learning**. Only the boundary read — grade — crosses into the world and moves
the position. *Nothing measures itself* is exactly this.

---

## The one-line map

```text
A       ranks       ORDINAL    — the descent · spreading attention          (static, per wave)
p       decides     CARDINAL   — the ascent · display, filter, weight       (static, per wave)
ΔB      moves       DYNAMICS   — across waves · how belief is CHANGING       (§4)
grade   pays        MARGINAL   — settlement · information the model added
```

Four jobs, four mathematical requirements. `A`, `p`, and `g` are **snapshots**; `ΔB` is the
**time derivative** they lack — which is why it, and only it, can carry surprise and
convergence, both of which are inherently about change. The single quantity that threatens
`A`, `p`, and `g` is the same one — the §5 emission leak — which is why **phase gates
everything**. `ΔB` is notable for **not** needing phase (§4).

## Where each is evaluated — the phase boundary

```text
PERTURB (descent)   produces A. NOTHING ELSE.
                    no probabilities · no grades · no truth values · just heat.

OBSERVE  (ascent)   applies P(T | A) to the arrived heat and produces the TRUTH VALUE.
                    branches: likely true · likely false · uncertain.
                    calibration lives HERE, because this is where its output is consumed.

WORLD    (settle)   produces the grade. the only place truth (not an estimate) enters.
```

**Calibration is an Observe operation, not a Perturb one — and this is forced.** Perturb
could technically apply the curve, but it would then carry a truth value down through layers
that do not use it. Observe applies it exactly where the show / drop / uncertain decision is
made. A probability never crosses the descent; heat never survives as heat past the ascent.

> **Observe produces a truth ESTIMATE that governs display. The world produces truth.** The
> word "estimate" is load-bearing: `P(T|A)` is derived from a contaminated field (§5), so it
> is a prior for what to SHOW, never a settlement of what is TRUE. "Observe produces the truth
> value" drifts toward "Observe settles claims" if that line is not held.

---

## 1. Attention — `A`

> **Attention is the scalar a record accumulates from the reverse-index resonance of the
> wave's perturbation, plus propagation over learned edges.**

The tokens are the **carriers**; `A` is what **lands**. Attention is not the tokens
themselves — it is the value they deposit on a record `R`.

### The resonance term — the core

```text
A_resonance(R) = Σ over DISTINCT shared tokens t of  w(t)
```

Three parts, each load-bearing:

```text
distinct    a token cannot corroborate itself. resonance is convergence of DIFFERENT
            tokens on the same record. three rare tokens agreeing > one rare token thrice.
            (the observe-wave rule: "distinct-root count first")

w(t)        the per-token weight.  OPEN — one of:
              1/df(t)        CONSERVATION.  bounds total field mass per impulse. steep.
              log(N/df(t))   IDF / EVIDENCE. a rare term is stronger evidence. tempered.
            leaning IDF (strength-of-connection), at the cost of the mass bound.

df(t)       posting size — read directly from the reverse index.
```

### The full sum

```text
A(R) = A_resonance(R)                              reverse-index overlap        (the largest term)
     + Σ propagation crossings (a · P · conf)       spread over learned edges    (graph structure)
     + emission bump                                the being's own claim — THE §5 LEAK
```

> **The emission bump is contamination, not signal.** `Place` currently excites the field on
> the being's own emission at the same amplitude as a world observation
> ([interaction-primitives](interaction-primitives.md) §5). Until §7 phase separates
> projected from observed heat, `A` is partly self-authored, and every quantity derived from
> it inherits that confound.

### What `A` is and is not

```text
A IS      an UNCALIBRATED relevance/resonance score. ordinal. unbounded (v4 dropped the clamp).
A IS NOT  a probability. not "the model's confidence" — that is `claim` (see §3).
          not conserved, not normalized: total field mass rises with propagation.
```

`A` ranks. That is its whole job. Importing it onto the truth axis (§2) is a category error.

---

## 2. Truth — `p`

> **Truth is a derived value: when this belief reaches attention `A`, how often is it
> actually true?**

```text
p = P(T = 1 | B, A)
```

`p` is an **estimate of where the belief will land on the falsification–truth scale** (§0),
read from the field before the world measures. The measured position is the accumulated `P`;
`p` is the field's guess at it. Both live on `[0,1]`; only the measured one is admitted to
learning.

Both conditions are load-bearing:

```text
| B    PER-BELIEF address. not global. this is what stops "hot ⇒ true":
       two beliefs at the same A have genuinely different truth rates.

| A    the CALIBRATION SLOPE. the entire content of the proposal.
       β = 0 ⇒ A tells you nothing, and p collapses to the per-belief base rate.
```

**`p` is evaluated on the ASCENT, by Observe.** Perturb produces `A` and stops; Observe applies
`P(T|A)` to the arrived heat and branches on it. This is the definition of what Observe does at
each layer — not a side computation. See the phase boundary above.

### The estimator — OPEN parameter, fixed form

```text
logit p = logit(P_B) + β · (A − Ā)
```

- `P_B` — the belief's own base rate (the edge scalar today), = `p` at mean attention.
- `β` — a single **global** slope, fitted across all settled bets. **OPEN: measure it.**
  `β = 0` reproduces the current system exactly, so the fit is inert until read.
- sparse beliefs shrink to the global rate: `p = conf·P_B + (1−conf)·P_global`, `conf = N/(N+5)`.

### The two guards — part of the definition

```text
DERIVED, NEVER SETTLED   p is read off the FIELD's geometry, not the world. it ranks,
                         filters, and weights display. it NEVER closes a claim, moves an
                         edge, or defeats a belief. only a world observation settles.
                         (third instance of: grouping confers attention, never settlement.)

A is ordinal, p is       the calibration curve exists precisely to turn an uncalibrated
CARDINAL                 score into a probability. if A were already a probability, the
                         curve would be the identity and there would be nothing to fit.
```

### What `p` is for

```text
display   show the UNCERTAIN, not the probably-true:  display = p(1 − p)
                                                       max at p=0.5, min at the tails
filter    drop the settled (p→1) and the likely-false (p→0) before ranking
diagnostic the slope/sharpness of p over A measures whether the FIELD's heat tracks truth
           at all — a health metric for the attention mechanism, independent of grading.
```

---

## 3. Grade — `g`

> **A grade answers: how much information did the model contribute over the baseline?**

```text
g = log P(T | claim) − log P(T | baseline)
```

The subtraction **is** the point: the model is paid only for information the baseline lacks.
Reciting the baseline earns exactly zero.

### The baseline is the open choice — and it sets what the metric measures

```text
baseline = P(T | B)       TODAY.      the edge scalar.       Σg/Resolved → I(T ; X)
baseline = P(T | B, A)    PROPOSED.   conditional (§2).      Σg/Resolved → I(T ; X | A)
```

Raising the baseline (conditioning on `A`) pays the model only for what **neither counting
nor the field's geometry** already knew — a strictly harder, more honest test.

### Three properties baked into the arithmetic

```text
STRICTLY PROPER    argmax at honest report. hedging and lying both lose.
                   E[score] = q·log p + (1−q)·log(1−p) is maximized at p = q.

GRADED BEFORE      g is computed against the prior the edge HAD, before Observe folds the
THE EDGE MOVES     outcome in. the baseline is never contaminated by the result it scores.

UNITS = NATS       the grade reads as information contributed. this is why the running sum
                   converges to a mutual information (below).
```

### The convergence identity

```text
E[g | instance i] = KL( Bern(qᵢ) ‖ Bern(q) )
E[ Σg / Resolved ] → I( T ; model's side-information )
```

Floor is the mutual information, **not zero**:

```text
converges POSITIVE   the model contributes instance-level information counting cannot produce
converges to ZERO    the counting graph could have done it alone   ← the sharp negative result
```

Inverted-U over reps; comparable only at equal rep counts.

### What a grade is NOT

```text
NOT the model's confidence          that is `claim`, the number being graded.
NOT settleable by attention          only a world observation produces the T in the formula.
NOT moved by activation heat          A informs the baseline (proposed), never the outcome.
```

---

## 4. Dynamics — `ΔB`

> **`ΔB` is how the belief field MOVES between waves.** It is the only quantity measured
> across time rather than within one wave, and it measures movement of belief — never
> correctness of belief.

```text
ΔB_t = B_t − B_{t−1}
```

**Status: CANDIDATE, promising.** The five inferences below are real; the decomposition in
§4.1 is the correction that makes them separable.

### 4.1 `B` is an OPEN basis — the decomposition is the point

The naive `‖ΔB‖` assumes `B_t` and `B_{t−1}` share a fixed coordinate space. They do **not**:
belief inference *creates and destroys* beliefs each wave, so the basis itself moves. A flat
subtraction hides three different events, and separating them is more informative than the
norm:

```text
BORN    beliefs that appeared this wave       no prior coordinate — PURE NOVELTY
DIED    beliefs that fell out                 no current coordinate — ABANDONMENT
MOVED   beliefs present in both waves          Δconfidence — the true UPDATE
```

> **`‖ΔB‖` lumps novelty, abandonment, and revision into one scalar. Always decompose first.**
> This is the coverage-vs-calibration distinction (settlement-and-surprise §8) reappearing:
> BORN is coverage (what the world showed that nothing anticipated); MOVED is calibration
> (how existing readings shifted).

### 4.2 The five inferences, scored against the decomposition

```text
1  SURPRISE      ‖ΔB‖, and specifically HIGH BORN     a candidate for the §8 window diff —
                 = the world showed something new      and PHASE-FREE (measures beliefs, not heat)
2  COMPETITION   signs within MOVED (+0.4 / −0.3)      substitution, ONLY if the two share an
                 = one reading gains as another loses   address/reading. correlation ≠ a link.
3  CO-MOVEMENT   beliefs that rise/fall together        learns graph STRUCTURE from temporal
                 across many waves                       correlation. Hebbian. may discover edges
                                                         the counting graph never posted.
4  STABILITY     consistently small ΔB                  AMBIGUOUS ALONE — converged vs disengaged
                                                         look identical. needs settlement to read.
5  CONVERGENCE   ‖ΔB₁‖ > ‖ΔB₂‖ > …                       candidate for the missing thesis curve
                 = the field is settling                 (prediction-error ↓). same #4 caveat.
```

### 4.3 What it is and is not

```text
ΔB IS      the DYNAMICS axis — the time derivative A, p, and g lack.
           inherently about change, which is why it carries surprise (#1) and convergence (#5).
           phase-free: it measures what beliefs DID, not projected-vs-observed heat, so it may
           sidestep the §7 blocker the heat-diff form of the window diff needs.

ΔB IS NOT  a correctness signal. every one of the five is an ATTENTION/DYNAMICS diagnostic —
           it says what cognition is DOING, never whether it is RIGHT.
```

> **`ΔB` measures belief movement, never belief correctness.** #4 and #5 are the trap: a
> shrinking `‖ΔB‖` is *converged* or *disengaged*, and belief movement cannot tell them apart.
> Only settlement (`g`) says whether stable beliefs are also right. Blur this and "the being
> feels settled, therefore it is correct" — the vibe the whole system exists to refuse.

### 4.4 Where it is open

```text
what is B's coordinate    the confidence vector over WHICH beliefs — the addressable set? all?
                          decides what BORN/DIED/MOVED range over.
the norm on MOVED         L1, L2, weighted by activation-at-assembly — undecided.
co-movement window        #3 needs a history length over which correlation is measured.
does it grade             NO — same rule as p: a dynamics signal gates and diagnoses, never
                          settles. it may inform when-to-think; it may not move an edge.
```

---

## 4a. The loop — the crank unrolled, and where each quantity governs

The four quantities are not four ideas. They are **the four arcs of one cycle** — the crank
(§0) unrolled across the record types. This is the dynamic spine; the measurement law (§0) is
the static one.

```text
   observation ──▶ belief ──▶ act ─ ─▶ prediction ──▶ [ THE WORLD ] ──▶ observation ──▶ …
        │            │          │           ┊               │                 │
   interpret     select      guard      (proposed         SETTLES          new reading,
   (belief       an act      while      stage — §3 of     the addressed    graded on the
    worker)      from the    the        staged-emission)  claim            way in
                 belief                                                     │
        └──────────────────────── ONE TURN OF THE CRANK ───────────────────┘
```

**The dashed arc is proposed, not decided.** `act ─ ─▶ prediction` as a *separate* stage is
[staged-emission-and-credit §3](staged-emission-and-credit.md). Today the prediction is
bundled with the act (expectations are a phase of L2). Drawn dashed so the diagram does not
encode a proposal as fact.

### The loop is NOT closed inside the being

> **The only arc that pays is the one through the world.** `belief → act → prediction` is a
> chain in the being's head; it becomes a *loop* only when the world settles the prediction
> and returns an observation. Remove the world step and you have the being spinning in its own
> head — the self-reference *nothing measures itself* (§0) exists to forbid. **The world
> closure is not one stage among four; it is the quarter that makes the other three
> legitimate.**

### Which arc each quantity governs

```text
A   ATTENTION   the DESCENT into each stage — what is live to interpret/select/guard    ranks
p   TRUTH       the ASCENT out of each stage — what is shown, filtered, weighted         decides
g   GRADE       the WORLD-CLOSURE arc — the prediction meeting its outcome               pays
ΔB  DYNAMICS    the NET MOTION of one full turn — how the belief set moved               moves
```

So `A` and `p` operate *within* every stage (descent in, ascent out); `g` operates *only on
the world arc* (the boundary read, §5); `ΔB` measures the *whole turn*. That is why only `g`
touches learning — it is the only quantity that lives on the arc through the world.

### The crank rhythm

One turn of this loop is one settlement folded into the accumulator — **one turn of the
crank, and (roughly) one tick** (a tick is one Observation; the world arc produces it). The
loop and the crank (§0) are the same object: the crank is the loop seen as accumulation, the
loop is the crank seen as motion.

*(Each arc is owned by a doc; this section routes, it does not re-derive.
`observation→belief` and `belief→act`: [belief-expression-inference](belief-expression-inference.md).
`act→prediction` and the citation format: [staged-emission-and-credit](staged-emission-and-credit.md).
world settlement: [settlement-and-surprise §3](settlement-and-surprise.md). the descent/ascent
mechanics: [perturb-observe-traversal](perturb-observe-traversal.md).)*

---

## 5. How they compose

```text
                 wave descends
                      │
     A  ─────────────┤  reverse-index resonance + propagation      ORDINAL. ranks the descent.
                      │  (contaminated by the §5 leak until phase)
                      ▼
                 wave ascends
                      │
     p = P(T|B,A) ────┤  calibrate A → probability                 CARDINAL. filters + weights display.
                      │  DERIVED, never settles.
                      ▼
                 world settles a bet
                      │
     g ──────────────┘  log P(T|claim) − log P(T|baseline)         MARGINAL. pays the model.
                         baseline may be P(T|B) or P(T|B,A).
                         only THIS touches learning.

     ΔB  = B_t − B_{t−1}   spans WAVES, not one wave. DYNAMICS. gates + diagnoses,
                           never settles. phase-free (§4).
```

- **`A`, `p`, and `ΔB` never touch learning.** They shape attention, display, and
  when-to-think. Getting them wrong costs a bad retrieval or a mistimed call, which decays.
- **`g` is the only one that moves an edge.** Getting it wrong is permanent, which is why the
  baseline is graded before the edge moves and why settlement is world-derived.

---

## 6. The single shared dependency

`A`, `p`, and `g` degrade under the **§5 emission leak**: the being's own claim excites the
field, so `A` is partly self-authored, `p` calibrates against a self-authored score, and a
leak-based baseline would let the model beat a prior it itself inflated.

> **`interaction-primitives` §7 (projected/observed phase) is the prerequisite for those three
> being clean.** It is marked SUSPENDED. It is the critical path.

**`ΔB` is the exception** — it measures belief movement, not heat, so it does not depend on
phase. That is its main structural advantage as a surprise/convergence signal (§4.3).

---

## 7. What is fixed vs open

```text
FIXED   the DEFINITIONS. A ranks, p decides, g pays. their forms and their guards.
        A = Σ distinct-token weights + propagation · p = P(T|B,A) · g = log-ratio over a baseline.

OPEN    w(t)       1/df vs log(N/df)                      §1
        β          measure it — one logged field           §2  ← the cheapest decisive move
        baseline   P(T|B) vs P(T|B,A)                      §3
        phase      §7 — the shared prerequisite            §5
```

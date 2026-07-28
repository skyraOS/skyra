# Definitions — Attention, Truth, Grade

**Date:** 2026-07-24 · **updated 2026-07-25** (§3a the two graded arms, §3b the tense split,
§4.3a what `ΔB` settles against) · **2026-07-26** (§0a the kernel contract; §3c Support —
the evidence→belief edge)
**Status: FORMAL DEFINITIONS.** The quantities are settled. Where a term inside a
definition is itself open (the weight `w(t)`, the baseline choice, the slope `β`), it is
marked OPEN in place — the *definition* is fixed even where a *parameter* is not.

**Reads on:** [settlement-and-surprise](settlement-and-surprise.md) (the two-axis split and
the calibration curve), [interaction-primitives](interaction-primitives.md) §9 (the field and
the §5 emission leak), [belief-expression-inference](belief-expression-inference.md) (the
inference calls truth and attention serve),
[staged-emission-and-credit](staged-emission-and-credit.md) (prediction as its own stage —
assumed by §3b), [edge.go](../../src/skyra/cognitive-agent/skyra-dna-v3/edge.go)
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

> **Corrected 2026-07-26 — see §3e.** This paragraph read: *"A single settlement is still a
> **binary event** … the event is binary; the measurement is the scale."* The binariness was an
> artifact of v3's categorical predicates (`exists true|false`, `status ok|failed`), not a
> property of measurement — and it does not survive a world whose returns arrive as payloads
> that must be interpreted. **What that clause was protecting is INDIVIDUAL and ADDRESSED, and
> both survive.**

```text
KEEP    one settlement = one reading, at one address, over one window
DROP    that reading is a bit
```

A settlement is a **signed displacement** that folds into the position. Absence still works:
nothing at an addressed claim is negative displacement, not a missing value.

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

## 0a. The kernel contract — signature + subject

**Status: DECIDED 2026-07-26.** Top-level abstraction only. The internal structure of either
half belongs with the mechanisms and is deliberately not fixed here.

Every agent receives the same two-part contract:

```text
ExecutionSignature   the contract by which the RUNTIME must HANDLE the subject.
                     necessarily deterministic — a handling contract that itself
                     required interpretation could not be handled.

ExecutionSubject     the thing to be handled within that runtime.
                     THE interpretive surface, and the only one.
```

> **The signature says HOW to handle. The subject is WHAT to handle.**

Handling is **execution** outbound and **interpretation** inbound — the same contract in both
directions, which is why the return leg needs no second shape even though it carries all the
interpretive load:

```text
Skyra → macOS   subject: "write file:/app.conf …"           handled by EXECUTING it
macOS → Skyra   subject: "ls: no such file or directory"    handled by INTERPRETING it
```

Determinism is therefore a *consequence* of the signature's job rather than a separate claim
about it.

The output has the same shape, so transformations compose:

```text
(Signature, Subject) ──▶ resolve transformer ──▶ (Signature′, Subject′)
```

The signature selects which transformation is valid; it never carries the subject's meaning.
This is the instruction-decoder shape already in `world/space.go` (`switch imp.Primitive`),
and it makes one guarantee structural rather than remembered:

> **The model cannot select its own transformer.**

Precedent — the closed Phase-4F parser registry: *"the LLM may choose a parser and its
arguments; it cannot add code to the grading path."*

### The split is receiver-relative

The line is not a property of the message. It is a property of the **(message, receiver)**
pair — what is subject to one agent is signature to the next.

```text
Skyra emits    place=world   target=macOS            primitive=engage
               subject: "write file:/app.conf …"      ← opaque to Skyra's kernel

macOS emits    place=macOS   target=file:/app.conf   primitive=write
               subject: "…file contents…"             ← `write` WAS subject, IS now signature
```

> **Interpretation is the promotion of subject into signature.**

The line moves not because agents disagree but because the runtime changed:

> **The subject's meaning is scoped to the runtime the signature names.**

`write file:/app.conf` means something inside macOS's runtime and is bytes outside it.

### Termination is a fixed point, not a classification

This is what makes the contract recursive rather than merely uniform: on each hop inward the
receiver promotes part of the subject into signature. When nothing is promoted, there is
nowhere left to go.

```text
TERMINATION   the promotion reaches a FIXED POINT — nothing was promoted, so
              Sig′ = Sig and Subj′ = Subj, and the next hop would be identical.

              DECLARED    by nil from Perturb
              EXPERIENCED as a nil realized Space at Observe
```

**The kernel must not ask "is this a resource?"** — that is a classification, and classifying
is cognitive work the kernel does not do. It only notices that nothing new came back.

This also gives `nil` a meaning it currently lacks (`kernel-alteration` §13 Q5 asks whether nil
suffices or needs an explicit result type): **nil from `Perturb` means *I promoted nothing*.**

Two consequences worth having:

```text
PER-MESSAGE, NOT PER-AGENT   an agent handed something it cannot parse promotes nothing and
                             terminates — without being a resource. "did not interpret THIS"
                             is a FINDING; "cannot interpret anything" is not the kernel's
                             to decide. `Outcome.Status: unavailable` already names it.

AN EMPTY SUBJECT IS LEGAL    a targeted perceive is nearly pure signature. nothing needs
                             interpreting, so it terminates at the runtime immediately —
                             an instance of the rule, not an exception to it.
```

The Slice keeps the branch either way: *"a branch that later returns nothing was still
genuinely perturbed."*

### Authorship is not the test — handling is

The model writes `Requires` on an act; a worker carries its own `CreationPhase`. Both are
signature regardless, because **what puts something in the signature is that the runtime
handles it — not that the substrate wrote it.** The model may *propose* a guard and can never
*satisfy* one: `actTruth` resolves it against observation.

### Addressing is produced, not received

There is no third slot. Two addresses are easily confused:

```text
DELIVERY address   where the message goes              signature
CLAIM address      where a claim lands in world state   sp(triple) — DOES NOT EXIST YET
```

`sp()` needs a Triple, and a Triple exists only after the subject has been interpreted. The
claim address is an **output** of interpretation — which is why an uninterpreted payload
cannot be settled, since §0 admits only what has an address.

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

## 3a. The two graded arms — two baselines, two headrooms

**Status: DEFINITIONS SETTLED 2026-07-25.** The baseline fork (§3) turns out to be one
decision *per arm*, and both stay open.

`g` has two independent instances. Two calls stake a confidence, and they do **the same job
in different material**: each takes something not yet on the falsification–truth scale and
puts it there. That is §0's admission step, performed twice.

```text
BELIEF CALL       observation → belief         "is this what happened?"
                  stakes a READING             graded through the discriminating
                                               address that reading authored

PREDICTION CALL   acts → expectation           "is this what follows?"
                  stakes a CONSEQUENCE         graded at the address it names
```

Each is then paid only for what its own counting graph could not already have said.

### The baselines are disjoint — and that is the lucky-hit correction

```text
BELIEF       baseline = P(T | this reading)   interpretationEdgeKey(category, contextKey)
PREDICTION   baseline = P(T | these acts)     edgeKey(actMap, Of, target)
```

They never share keys, settlement indexes, or updates. Collapse them and cells 2 and 3 of
the two-bit table become unreachable — including cell 3, the lucky hit
([v4-data-model §5a](v4-data-model.md)).

A Belief has no address of its own, so its confidence is not graded directly: it is graded
**through** a separate discriminating claim the reading authored, at a distinct `sp()`.
Where that test lives is still open — [v4-data-model §5](v4-data-model.md).

**Neither call is told it is being scored.** A calibrated model beats its base rate
automatically whenever it actually knows something; asking for it is how you stop getting it
(`present.go`).

### The asymmetry — headroom against data

The arms are not symmetric, and the asymmetry is structural rather than incidental:

```text
              GENERATIVITY                        DENSITY
BELIEF        HIGH. no counting process turns     SPARSE. needs an authored
              a payload into an addressed         discriminating test; most
              claim. the hard arrow.              readings never settle at all.

PREDICTION    LOW. given belief + acts, this      DENSE. LAW 5 forces an expect
              is interpolation over history.      on every act.
```

```text
BELIEF        more headroom, less data
PREDICTION    less headroom, more data
```

### What that predicts, and how to read it

Both arms carry §3's convergence identity, so both are inverted-U over reps — but they
should **not** peak together.

```text
prediction arm decays toward its floor FIRST     more samples. expected. NOT a failure.
belief arm converging to ZERO                    THE REAL NEGATIVE RESULT — the counting
                                                 graph could have read the world alone.
```

Two arms, two curves. `Metrics.GradeSum` and `Metrics.InterpretationGradeSum` are already
separate, and `being.go` says why: folding them would recreate the lucky-hit corruption in
reporting. The instrument for this question is already built.

### Where the baseline fork lands, per arm

```text
PREDICTION   P(T | acts)            TODAY. the model earns for conditioning on belief.
             P(T | acts, beliefs)   PROPOSED. earns only for what neither counting NOR
                                    the belief state provides. may be near-empty.

BELIEF       P(T | reading)         TODAY.
             P(T | reading, A)      the §2 calibration curve, same form.
```

> **Raising a baseline can remove an arm's job entirely.** Hold prediction at `P(T | acts)`
> until `β` is measured — there is no way to know whether headroom survives conditioning
> before knowing whether attention predicts truth at all.

---

## 3b. Sight runs backward — the temporal division of labour

**Status: DECIDED 2026-07-25.** Assumes prediction takes its own inference call
([staged-emission §3](staged-emission-and-credit.md)); the sight rule holds either way.

Each worker owns one tense, and the tense decides what it may see.

```text
BELIEF       articulates the PAST into a present snapshot   sees heated Observations
                                                            + completed Actions
ACTION       decides in the PRESENT                          guards resolve against
                                                            OBSERVED state
PREDICTION   commits to a FUTURE                             emits addressed claims,
                                                            each with its own window
```

This is why desire-blindness is natural rather than imposed: a worker responsible for the
past has no business seeing what answer is wanted.

> **SIGHT RUNS BACKWARD. FORWARD IS ONLY EVER COMMITTED TO.**

One rule covering all three workers instead of three separate blindness rules. **No worker
observes the future — including the prediction worker.** Prediction does not *see* a future;
it **authors** one and is graded on it. If it could see one it would not need to bet, and
`g` would have no gap to measure — LAW 2 makes a grade a cross-boundary diff, and an oracle
removes the boundary.

The §5 emission leak is a **partial oracle already**: `Place` excites the field when the
being emits, so predicting something makes it likelier to surface. Same failure at low
amplitude. Third reason phase is the critical path.

### What "possible future observations" resolves to

The prediction worker is reasoning *forward* over *backward* data, and that is the job. The
phrase decomposes into two halves, neither of them speculative:

```text
ADDRESS SPACE       which addresses can this act produce a reading at?
                    → the affordance registry. WORLD-GIVEN, present-tense.
                    not a projection at all.

VALUE DISTRIBUTION  what values came back at those addresses before?
                    → the act-keyed edge. COUNTED, backward.
```

For `exec sh build.sh`: the world exposes `dir:/ status` and `dir:/ output` — a fact about
instruments. What those addresses returned across prior runs — the distribution. The model
combines them with current beliefs and states one number per address.

> **This makes [v4-data-model §5a](v4-data-model.md)'s affordance defect operational rather
> than a curiosity.** `world.go` emits `build:/ missing app.conf`; `DefaultAffordances`
> registers only `build: {status, output}`. If the address space comes from the affordance
> list, the worker cannot predict at the most diagnostic address the build exposes — or
> predicts there and settles at 0.3 despite a mechanical report. *The instruments a world
> exposes should BE the affordance list* becomes a prerequisite of staged prediction, not a
> cleanup.

### The line: who computes the projection

Not past-vs-future. **Who computes it, and whether it enters as material.**

```text
LEGAL     the frame carries COUNTED HISTORY at the addresses this act can touch.
          the MODEL reasons forward from it and states a number.
          "7 of 10 produced status ok" → 0.15, because app.lock is missing.

ILLEGAL   the frame carries a PRE-COMPUTED PROJECTED OUTCOME as an item.
          "projected: build:/ status ok"
```

The second fails twice. It is the base rate wearing a costume — the model anchors on a
**point estimate** where it should be reasoning over a distribution, destroying exactly the
uncertainty it is paid to resolve. And a committed projection is field material, so it
excites the addresses it names: the §5 leak, self-inflicted.

**A boundary that will come up in implementation:** seeing the act set — including acts that
have not fired — is *not* seeing the future. Those acts are commitments that exist **now**,
and the worker needs them for `By` and for joint expectations. What it may not see is
anything asserting what those acts will produce. *The act is present; its consequence is the
bet.*

### Every worker's frame has the same three slots

```text
commitments   what THIS cycle has committed so far     within-cycle, phase-filtered
options       what this worker may commit — PLURALLY    resolve · hold · disregard
precedent     counted outcomes of similar commitments   BACKWARD · aggregated · world-derived
```

The third slot is deliberately **not** "projected futures." A rejected framing made futures
the third slot and had each worker select its commitment by evaluating them — which is
selection-by-downstream-consequence, i.e. LAW 6 broken, the lucky hit unreachable, and the
being scoring its own branches. Recorded because the shape is attractive and will be
proposed again:

> **This architecture is a commit-and-be-graded loop, not a search.** No branch evaluation
> appears anywhere in it, and that absence is load-bearing. *Nothing measures itself* forbids
> the agent scoring its own options.

`options` is plural on purpose: singular commitment kills HOLD, and a predicted belief-*set*
must decompose into individually addressed expectations
([staged-emission §4](staged-emission-and-credit.md)).

### The prediction worker does not learn — the frame does

No weights update. "Getting better at prediction" is therefore never the worker improving;
it is the **frame improving**, which is [v4-data-model](v4-data-model.md)'s standing claim
that the only thing varying across runs is what the model was shown. Three channels, all
counting:

```text
BASE RATES SHARPEN   [worked 63% of 8 tries] is noise at N=1 and a real prior at N=30
BELIEFS SHARPEN      the instance-discriminator improves — prediction quality is
                     DOWNSTREAM of belief quality, which is what the tense split predicts
`Within` SHARPENS    the settlement log has observed latencies; today it is guessed.
                     probably the largest near-term win: v3 is strict at maturity, so a
                     share of current "prediction error" is bets maturing at the wrong
                     moment rather than bets being wrong.
```

### Which curve carries "prediction error" — RESOLVED

```text
CALIBRATION ERROR    |p_claim − outcome|, aggregated     monotone ↓
                     "my stated probabilities match reality"
                     THIS is prediction improving. this is "reduce errors."

GRADESUM/RESOLVED    log p_claim − log p_prior            inverted U (§3a)
                     "I know things counting does not"
                     NOT error. decays as the graph converges, BY DESIGN.
```

> **`GradeSum` is not the prediction-error curve and never was. It is the headroom curve.**
> Prediction improving looks like calibration error falling *while* `GradeSum` decays. Those
> happen together, and reading the decay as failure is the trap.

This closes [settlement-and-surprise §7](settlement-and-surprise.md)'s open item — *which
metric carries "prediction error"* — in favour of calibration. `prediction_calibration.go`
already exists; the curve that answers the question is the one not currently plotted.

### The design bet — the action worker does not see futures

**Stated as a bet, because it is one.**

> **Steering runs through belief and retrieval, never through lookahead.** The action worker
> is given no projected consequences of its options. The wager is that the learning loop
> alters belief state — and therefore what surfaces — enough to steer act selection.

The bet is already made, in [v4-data-model](v4-data-model.md): *"Nothing here changes an act
… the only thing that varies across runs is what the model was shown."* The channel is the
cascade in [time §7](time.md): grades move `Edge.P` → `P` is both grading prior **and**
`RetrievalWeight` → different precedent wins the knapsack → different present → different
emission.

It already has a curve. `PassesToClosure ↓` **is** this claim. The failure signature is
distinguishable without a new metric:

```text
GradeSum rises, passes-to-closure FLAT     predicting better, acting no better — BET LOST
```

**The cost of staging, stated plainly.** Today express and expect are one call, so the action
worker has *implicit* future-awareness: it cannot propose an act whose outcome it cannot
state, because it must state it in the same breath. [workers](workers.md) names this — *"in
one call the model picks acts knowing it must bet on them … splitting removes that
constraint, and nothing currently replaces it."* **Staged prediction is what makes this bet
real.** Adopting the split and the bet together is a larger step than either alone.

**The fallback is already half-built.** `Expectation.By` makes edges act-keyed —
`P(state | these acts)` — and nothing reads it. If passes-to-closure flattens, that seam is
consumed rather than the design being reopened.

---

## 3c. Support — the evidence→belief edge, and why it is not yet a third arm

**Status: VOCABULARY SETTLED 2026-07-26. NOT a graded arm** — the units do not permit it
(below). Recorded here so the field is shaped for the upgrade rather than against it.

> **Partly superseded by §3d, same day.** Two things below were wrong and are corrected there,
> not deleted: **(1)** `AVAILABILITY` is not a third column on this edge — activation runs the
> *other direction* (observation → belief) and is keyed on tokens, not on the support edge at
> all. **(2)** `DIAGNOSTICITY` is keyed on the **belief**, not on the address — a token form or
> an address is instance-shaped, and the belief is the only reusable identity. The
> authored/counted split and the units argument stand unchanged.

A Belief is licensed by evidence. Today that link is an **unweighted list** —
`Belief.ObservationID`, `Observation.DerivedFrom`, `InterpretationClaim.DerivedFrom` — so
nothing says how much any one piece of evidence did. Three distinct numbers can sit on that
edge, and they are routinely conflated because two of them are per-instance floats:

```text
SUPPORT         authored     [-1,1]   per instance    what the MODEL says this evidence did
DIAGNOSTICITY   counted      signed   per ADDRESS     what material at this address is WORTH
AVAILABILITY    structural   ≥0       per instance    what was HOT when the reading formed
```

### The first division: earning vs distributing

```text
EARNING       does the model know what its reading rested on?   SUPPORT vs DIAGNOSTICITY
DISTRIBUTING  a settlement arrived — who gets what share?        AVAILABILITY
```

`AVAILABILITY` is **not a measurement of contribution and must never be read as one.** It is
heat at formation: *this material was live*, not *this material bore weight*. It is used
because it is the only non-authored number present at formation time, and its job is to deny
the model the pen — [v4-data-model §3](v4-data-model.md) is explicit that attribution is
collective *"forced by the runtime not knowing which belief produced which bet."* A fallback
under ignorance, not a finding. `Expression.Beliefs map[BeliefID]float64 // activation AT
ASSEMBLY TIME` is the same fallback one layer up, and must be stamped at formation for the
same reason: the field decays per impulse, so by settlement the heat that caused the reading
is gone.

### `Support` is signed, and the sign is the point

```text
+1   this evidence strongly argues FOR the reading
 0   neutral
−1   this evidence strongly argues AGAINST it — active counterevidence
```

A magnitude-only scale cannot distinguish **opposition** from **absence**, which destroys the
ability to record that a reading was held *despite* something. This is the same quantity the
prediction arm already emits as `EvidenceAttribution.ClaimedInfluence` — *"authored alignment,
not probability and not earned credit"* — and the same shape as weight-of-evidence, which is
signed for the same reason.

> **Naming.** `Support` names the quantity; the sign names the direction. Negative support is
> opposition. The v3 prediction-side field is `ClaimedInfluence`; no v3 code is renamed by this
> doc.

### `Diagnosticity` is already implemented, twice

What a counted evidence baseline represents is **how informative material at this address is
as a signal** — and it is not a new mechanism:

```text
density_pockets.go          coverage · baseRate · coverage·log(coverage/baseRate)
prediction_calibration.go   BaseRateCalibration{Coverage, BaseRate, PMI, ExpectedCredit}
                            + MeasuredInert · MeasuredAntiDiagnostic
```

Evidence grain instead of predicate grain. **But it is weaker than the act edge and should not
be described as the same kind of thing.** `P(W | E)` is causal — the acts produce the outcome.
Evidence does not make a reading true; the world does. `Diagnosticity` is a property of an
*address as a signal*, not a relation between two events.

### Why this is NOT a graded arm — the units

```text
SUPPORT         [-1,1] signed influence.   NOT A PROBABILITY — the prompt says so outright.
DIAGNOSTICITY   a frequency / a PMI.
```

`g = log p_claim − log p_prior` requires both operands to be probabilities. They are not, so
the log-score does not apply and **§3a's two arms stay two.**

What *is* comparable is that PMI is signed too, so the two can be set against each other
directly:

```text
claimed sign      vs  measured sign        does the model know WHICH WAY evidence cut?
claimed magnitude vs  measured magnitude   does it know HOW MUCH?
```

Sign agreement is the sharper reading and the one worth plotting first — a model that
systematically cites counterevidence as support is making a legible, specific error, and it is
a **self-knowledge** signal distinct from correctness. Nothing in the set currently produces
one.

> **The limit that follows.** Agreement is not a proper scoring rule, and
> [settlement-and-surprise §8](settlement-and-surprise.md) requires anything touching credit to
> stay strictly proper — that property is what makes hedging unprofitable. **So `Support`
> reports and diagnoses. It does not move an edge.** Same status the prediction side already
> assigns it.

### What the weighted edge fixes regardless

[v4-data-model §2](v4-data-model.md) has `Belief.ObservationID` **singular**, "never empty."
[tokenizer §4](tokenizer.md) and [belief-expression §4](belief-expression-inference.md) both
have `Evidence []TokenSpan`, and spans carry their own `ObservationID` — plural by
construction. Belief inference sees *"the heated region's Observations"*, plural, so a reading
grounded in both a current utterance and a prior build failure is expressible in one doc and
not the other. **An evidence set with per-edge fields is the plural link**, and it also
completes the credit walk, whose last hop (`Belief → its Observation`) is currently the only
unweighted one.

### Open

```text
GRAIN        ~~per TokenSpan or per Observation?~~  CLOSED by §3d — NEITHER. the key is
             the BELIEF. token forms and addresses are instance-shaped.

THE UPGRADE  making this a genuine third arm needs a PROBABILITY-shaped evidence claim,
             which costs the sign — and a third keyspace, disjoint from the belief and
             prediction arms or the lucky-hit collapse reappears at evidence grain (§3a).
```

---

## 3d. The belief graph — activation up, support down

**Status: DECIDED 2026-07-26.**

### One node type, one edge type

Interpretation produces beliefs. The belief a payload was read into and the belief being
tested are **the same type in different positions** — nothing distinguishes them but what is
resting on what.

```text
u   the raw subject PERCEIVED — foreign-authored, uninterpreted    (PerceivedAct carries it)
r   the belief u was transformed into
B   the belief whose position is moving
```

`u → r → bears on → B`. `r` is not a lesser kind of thing than `B`.

### Role is positional, not intrinsic

The same move as §0a's receiver-relativity: what something *is* depends on the relation being
looked at.

```text
ROOT       a belief with no ACTIVE parent            acting as HYPOTHESIS
INTERNAL   cited by something, citing something      both at once
LEAF       a raw observation                          evidence only
```

*Active* makes rootness per-wave rather than a property stamped on a record.

### Every belief cites its support, or it is rejected

LAW 5's shape on the belief side:

```text
LAW 5   every act appears in at least one expect, or the emission is REJECTED
        every belief cites at least one evidence edge, or the emission is REJECTED
```

Validation at emission, not discipline. [v4-data-model §2](v4-data-model.md) already has the
singular unweighted form: `Belief.ObservationID // what produced this. never empty.`

### The regress terminates at raw observation

Evidence may itself be a belief, so the chain must bottom out. A raw observation needs no
support because it is **unfalsifiable by construction — it happened.** Same shape as §0a's
termination: the recursion stops at something that owes nothing further.

### Two directions, and they never cross

```text
ACTIVATION   observation → belief    heat · what SURFACES     keyed on TOKENS    transient
SUPPORT      belief → evidence       structure · what COUNTS  keyed on BELIEFS   durable
```

> **Tokens group. Beliefs settle.**

That is *grouping may confer attention, never settlement rights*, assigned to the two edge
directions. The tokenized field is **solely** the activation substrate; nothing in the learning
layer is keyed on it.

**Why the belief is the only legal key.** A token form is instance-shaped —
`ls: cannot access '/app.conf'` and `…'/app.lock'` are different sequences and the same *kind*
of thing, so counts keyed on them stay at `N=1` forever. That is `store.go`'s recorded failure:
*"the reusable thing was keyed by the instance that made it, not by what it is."* **The belief
IS the compression** — unbounded payload variation collapsed into a bounded addressed claim —
which makes it the only thing carrying a reusable identity.

### Ownership follows from that

```text
SUPPORT      belief-owned · written once at creation · backward-pointing   DURABLE — §7 rule 1
ACTIVATION   field-owned · decays per impulse · rebuildable                OUTSIDE §7
```

The belief is written once, so its citations are too. Activation can be lost and recomputed
at no cost, which is why [interaction-primitives](interaction-primitives.md) puts the field
outside the storage law.

### Ambiguity is plural beliefs, never a scalar

An ambiguous payload does not yield one belief at reduced confidence. It yields **several
beliefs**, each owning a support edge to the same observation, each carrying its own position.

```text
u = "ls: /app.conf: Permission denied"
      ├── file:/app.conf exists true
      └── dir:/ readable false
```

LAW 4 stated properly: *"ambiguous references are preserved as low-weight candidates, never
hard-linked."* A channel-reliability scalar discounting a single reading was considered and
**rejected** — it smuggles a scalar in where the design wants plurality.

### This does not violate perturb-observe §6a

§6a forbids *"Beliefs justifying new Beliefs."* Belief-to-belief support is not that, and the
distinction is the one running through the whole design:

```text
FORBIDDEN   the MODEL using prior beliefs to AUTHOR a new one    generation · authored
ADMITTED    the SYSTEM counting how beliefs co-occur             measurement · counted
```

The belief worker still cannot see prior concrete Beliefs. **What stacks is the fold, not the
prompt.** Worth stating loudly wherever this is built, because the surface reading is a
contradiction and will be raised.

### Acyclicity — one rule, no DAG to maintain

If `r` moves `B`, nothing structurally stops `B` moving `r`.

> **Every agreement in a wave is computed against PRE-WAVE positions; all folds apply at the
> end of the wave.**

No within-wave propagation, so no cycle. This is §3's *graded before the edge moves* applied to
the belief graph, and it makes the wave the unit of update — which matches a tick being one
Observation ([time §3b](time.md)).

### Units stay apart

```text
Support   [-1,1]   AUTHORED   what the model says this evidence did
W         nats     COUNTED    weight of evidence, r → B
```

One relation, two readings, different ranges. They compare by **sign and rank, never by
value** (§3c).

---

## 3e. Settlement is accumulated agreement

**Status: DECIDED 2026-07-26.** Supersedes the binary-event clause in §0 (corrected there, not
deleted).

Settlement needs no mechanism of its own. A predicted belief **is** a belief; evidence arriving
after it moves its position like any other evidence. The causal chain is already durable —
`Expression.Beliefs → Expression.Acts → Expectation.By` — so nothing new has to be linked.

> **settlement(B) = Σ W(rᵢ, B), over evidence arriving inside B's window.**

### Two jobs v3 runs together, and they split

```text
THE SWEEP    goes and looks at addresses with open stakes   an ACT — produces observations
SETTLEMENT   the accumulated agreement those produce        ARITHMETIC — produces nothing
```

v3's `Settle()` perceives, compares, and resolves in one pass. Split, the sweep's returns are
ordinary observations flowing through ordinary interpretation, and there is **no special
settlement path** to keep the model out of. LAW 4 gets easier rather than harder.

### The window bounds which evidence counts

Passive accumulation never finishes, and grading needs a fixed outcome. `Within` supplies the
bound — which is what [v4-data-model §3](v4-data-model.md) already redefined it to be: a
stability claim over a window, not a deadline.

```text
INSIDE the window   settles the claim
AFTER it            about the world, not about the bet
```

Individual settlement survives, so the two-bit table does.

### Grading survives as a strict generalization

The outcome is no longer a bit but a final position `P_f = σ(L₀ + ΣW)`. The log score takes a
soft target unchanged:

```text
g = P_f · log(p_claim / p_prior)  +  (1 − P_f) · log((1 − p_claim) / (1 − p_prior))

P_f = 1  →  log p_claim − log p_prior             the current TRUE branch
P_f = 0  →  log(1 − p_claim) − log(1 − p_prior)   the current FALSE branch
```

Still strictly proper, so [settlement-and-surprise §8](settlement-and-surprise.md)'s
requirement holds and hedging stays unprofitable.

### Three readings that were one value

```text
ΣW < 0       falsifying displacement
ΣW = 0       OBSERVED — and the world said nothing about this claim
NO RECORD    nobody looked. a window that expired UNOBSERVED (time §3d)
```

The middle row does not exist today; `TruthUnknown` collapses it into the third.

### Absence gets stronger, not weaker

Each sweep returning nothing contributes a negative `W`, and they accumulate. That is
v4-data-model §3's *"absence becomes a sustained finding rather than one sample"* implemented
rather than asserted — and the old property (*a targeted perceive comes back authoritative for
exactly the claim bet on*) is what still licenses each individual negative term.

### What it costs

```text
Settlement.Resolution Truth         →  ΣW, plus the window bounds
Edge.Observe(outcome Truth, tick)   →  takes a DISPLACEMENT, not a bit
compareExactExpectation             →  GONE. it existed to manufacture the bit.
Settle()                            →  splits: the organ sweeps; accumulation is a fold
```

**The load-bearing break is `Edge.P`.** It is a Bernoulli frequency *and* it is
`RetrievalWeight` — one number doing two jobs, deliberately (`edge.go`). Continuous outcomes
break the frequency reading. The staged path: keep `P` as a coarse within-tolerance rate **for
ranking**, let `ΣW` carry **grading**, and do not collapse them until the residual has been
measured to carry information.

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

### 4.3a Staged prediction gives `ΔB` something to settle against — 2026-07-25

Under the tense split (§3b), the prediction worker and the belief worker emit **the same type
at the same addresses, one step apart**:

```text
PREDICTION WORKER   commits the expected post-action belief set
BELIEF WORKER       commits the actual next belief set
```

They subtract. So `ΔB` stops being a quantity merely *observed* about the field and becomes
the **settlement of one worker's output against another's** — `BORN`/`MOVED` measured against
what was projected, at named addresses, where absence within an address is falsity.

That is the property [settlement-and-surprise §6](settlement-and-surprise.md) says the window
diff cannot inherit. It does not close the gap — beliefs that were `BORN` with no prediction
covering them are coverage failures and still need the diff — but it converts the *predicted*
portion from a gate into something gradeable. **This is the strongest argument for staging
prediction, and it is not currently in [staged-emission §3](staged-emission-and-credit.md).**

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

FIXED   THE KERNEL CONTRACT is signature + subject. the model cannot select     §0a
        its own transformer. the split is RECEIVER-RELATIVE — interpretation
        is the promotion of subject into signature, and addressing is one of
        its products. the subject's meaning is SCOPED to the runtime the
        signature names. terminates at a FIXED POINT of the promotion —
        declared by nil from Perturb; never a classification.
        (the structure of either half is deferred to the mechanisms.)

FIXED   TWO GRADED ARMS on disjoint keyspaces, asymmetric in headroom and data.   §3a
FIXED   SIGHT RUNS BACKWARD. forward is only ever committed to.                    §3b
FIXED   calibration carries "prediction error." GradeSum is the HEADROOM curve.    §3b
FIXED   SUPPORT is authored and DIAGNOSTICITY is counted — one relation, two          §3c
        readings. support is signed and is NOT graded; the units forbid a log
        score, so they compare by sign and rank only.

FIXED   ONE NODE TYPE (belief), ONE EDGE TYPE (support). role is POSITIONAL —         §3d
        root / internal / leaf. every belief cites support or is rejected.
        the regress terminates at a raw observation.
        ACTIVATION runs observation→belief on TOKENS · SUPPORT runs
        belief→evidence on BELIEFS. tokens group; beliefs settle. never cross.
        ambiguity is PLURAL BELIEFS, never a scalar.
        acyclicity: all agreements computed against PRE-WAVE positions.

FIXED   SETTLEMENT IS ACCUMULATED AGREEMENT — ΣW over the window, a signed          §3e
        DISPLACEMENT, never a verdict. no mechanism of its own; the causal
        chain is already durable. the SWEEP (an act) splits from SETTLEMENT
        (arithmetic). grading generalizes to a soft target and stays proper.
        ΣW=0 (observed, orthogonal) ≠ no record (nobody looked).

OPEN    w(t)       1/df vs log(N/df)                      §1
        β          measure it — one logged field           §2  ← the cheapest decisive move
        baseline   ONE FORK PER ARM, both open             §3a
                     prediction  P(T|acts) vs P(T|acts,beliefs)
                     belief      P(T|reading) vs P(T|reading,A)
        phase      interaction-primitives §7 — shared      §5
        the test   where a reading's discriminating claim lives — v4-data-model §5
        grain      Support per TokenSpan or per Observation §3c
        the upgrade a probability-shaped evidence claim would make §3c a third
                    arm — and would cost the sign                  §3c

BET     act selection is steered by belief + retrieval alone, never lookahead.     §3b
        loses if GradeSum rises while PassesToClosure stays flat.
        fallback already half-built: `Expectation.By` is act-keyed and unread.
```

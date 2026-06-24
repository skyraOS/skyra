# Abductive Reasoning — Decomposition

> **Reconciliation (2026-06-23).** Read **"Coherence" → Global**, **"binding
> realities" → the self/other agents**. The missing-vs-wrong fork is realized in
> [error-recovery-loop.md](error-recovery-loop.md) (the prediction × assumption 2×2).

## Status

**Research / planning — to map onto the architecture.** A decomposition of abductive reasoning (inference to the best explanation) into ten steps. Captured to discuss where each step already lives in Skyra's machinery and where it doesn't yet. The running example is a first-contact scenario; the steps are general. Pairs with [comparator.md](comparator.md) (graded alignment, uncertainty), [predictive-hierarchy.md](predictive-hierarchy.md), and [setpoints.md](setpoints.md) (goal inference).

## The Ten Steps

### 1. Observation Extraction

Convert raw input into discrete observations. No explanation yet — just facts.

```
O1: Countdown appears on retinas.
O2: Stars appear to flicker.
O3: Aliens communicate.
O4: Aliens learn humans can lie.
O5: Aliens stop communicating.
O6: Aliens stop retinal countdowns.
```

### 2. State Change Detection

Find events where the world changed. Abduction is mostly about explaining **transitions**.

```
Before:                 After:
  Communication active    Communication inactive
  Retinal effects active  Retinal effects inactive
```

State changes become the things needing explanation.

### 3. Causal Dependency Discovery

Identify observations that might be related — generate candidate links, not asserting truth yet.

```
Aliens learn humans can lie  →  Communication stops
Communication stops          →  Retinal effects stop
```

### 4. Goal Inference

Infer what hidden objectives an actor may have. **Without goals, causality is weak** — many explanations become obvious once goals are hypothesized.

```
G1: Establish trust
G2: Warn humanity
G3: Gather information
G4: Manipulate humanity
G5: Avoid revealing capabilities
```

### 5. Latent Variable Generation

Invent hidden causes not explicitly observed. **The heart of abduction — you create entities that aren't in the data.**

```
Observed:   Aliens learn humans lie · Aliens stop talking
Generated:  H1: Aliens now distrust humans
```

### 6. Explanation Synthesis

Build causal graphs into a coherent story.

```
Humans can lie
   → Aliens update trust model
   → Aliens distrust humanity
   → Communication ceases
   → Visible interventions cease
```

### 7. Counterfactual Testing

Try to break the explanation. Weak explanations collapse here.

```
If humans could not lie, would communication continue?
Could the same observations occur if distrust never happened?
```

### 8. Competing Hypothesis Generation

Generate multiple explanations. **Humans often stop too early here. Strong reasoners don't.**

```
A: Distrust explanation
B: Interventions were expensive
C: Mission phase completed
D: Political faction change
```

### 9. Bayesian Ranking

Score explanations:

```
Score = Coverage × Simplicity × Goal Consistency × Predictive Power
```

- **Coverage** — how many observations explained?
- **Simplicity** — how many hidden assumptions?
- **Goal Consistency** — does it fit the inferred goals?
- **Predictive Power** — what new observations would it predict?

### 10. Uncertainty Preservation

**The most important piece. Most AI systems fail because they collapse too early.** Maintain the distribution rather than picking a winner:

```
Distrust hypothesis:   42%
Mission complete:      31%
Resource limitation:   17%
Internal politics:     10%
```

Humans call this "keeping an open mind."

## Mapping to Skyra — To Discuss

The decomposition lines up with machinery that already exists, which is the reason to capture it. First-pass correspondences (not yet decided — the point is to find where abduction *is* the architecture and where it needs new structure):

| Step | Skyra machinery |
|---|---|
| 1 Observation extraction | sensory intake — observations land on the slice as cells |
| 2 State-change detection | the comparator's **trend** (Δalignment across frames) — transitions, not levels |
| 3 Causal dependency | co-activation linkage; edges between cells |
| 4 Goal inference | Coherence's **Observe/Predict** (other-modeling on the descent — "what the other agent will do") + setpoint inference ([setpoints.md](setpoints.md)) |
| 5 Latent variable generation | **topology growth** — Coherence splits a cell when a mismatch can't be absorbed; `doesNotUnderstand` seeds a new cell. Generating `H1: distrust` *is* depositing a latent cell |
| 6 Explanation synthesis | the topology **is** the causal graph; understandings deposit on ascent (trace → understanding) |
| 7 Counterfactual testing | the frontopolar **Resolve** audit; adversarial verification |
| 8 Competing hypotheses | multiple goals/explanations coexist ([comparator.md](comparator.md): "multiple goals coexist") |
| 9 Bayesian ranking | the **multiplicative** activation product — `Goal Consistency` = alignment to setpoint, `Predictive Power` = Coherence's Observe/Predict |
| 10 Uncertainty preservation | **the comparator's graded `<alignment>` level, not a binary** — the exact .1001 finding (Run-1 polarized to ±1 and lost the middle; Run-2 recovered the distribution). "Don't collapse too early" is already the design's core discipline |

The sharpest connection is **step 10 ↔ the comparator**: "most AI systems fail because they collapse too early" is the same failure the `−1..1` level was designed to avoid, and "maintain the distribution" is why alignment is a graded level rather than a done-flag. The sharpest *gap* is **step 5 (latent variable generation)** — inventing a hidden cause that isn't in the data is the most generative move here, and it maps to topology growth under coherence pressure, which is specced but not built.

Open question for the mapping: is abduction a *separate* process, or is it what the existing descent/ascent loop already does when run over observations with goal inference on? Leaning the latter — abduction may be the comparator + Prediction + topology-growth already composed, not a new module.

## What LLMs Are Bad At

The failure modes are the negative space the architecture is built around — each one maps to a mechanism Skyra adds *on top of* the LLM.

### 1. Distinguishing plausibility from truth

The biggest problem. Abduction asks *"what explanation best explains the observations?"* LLMs often answer *"what explanation sounds most reasonable?"* Those aren't the same — **a compelling story can dominate a more likely explanation.**

### 2. Preserving uncertainty

LLMs are notorious for collapsing onto a single narrative. Instead of `A: 40% · B: 35% · C: 25%`, they produce *"the answer is probably A"* and move on. (Humans do this poorly too — but LLMs worse.)

### 3. Inventing truly novel latent causes

Where it gets interesting. In a synthetic world — new physics, new social rules, new incentives, no resemblance to training data — **performance drops**. The model can no longer retrieve analogous structures from memory; it has to actually **build a world model**. Current LLMs do some of this, but not nearly as well as their familiar-domain performance suggests.

### 4. Active information gathering

Human scientists don't just hypothesize — they ask *"what observation would distinguish these explanations?"* Seeing that both A and B explain the data, they design the experiment that separates them. LLMs can do this when prompted, but **they don't naturally drive the process.**

## Why LLMs Look Better at Abduction Than They Are

Many benchmarks accidentally leak structure from training data. `fever + cough + low oxygen → "pneumonia"` looks like reasoning, but it may largely be **compressed pattern recall**. The real test is when the domain is novel, the causal structure unfamiliar, there are multiple competing explanations, and no analogous examples exist in training. That's where the gap becomes visible.

## How the Architecture Targets These

| LLM weakness | What Skyra adds to counter it | Status |
|---|---|---|
| **1. Plausibility vs. truth** | **observed-not-asserted** — alignment scores against the *observable* `satisfied-when`, not what sounds right; the LLM can only cite cells that exist in the frame ([comparator.md](comparator.md): "cannot fabricate a thought"), so a compelling story can't win without grounding | validated (.1001 "no self-satisfaction") |
| **2. Preserving uncertainty** | the graded `<alignment>` **level, not a binary** — maintain the spread until collapse is *earned*; the .1001 Run-1 → Run-2 fix was precisely recovering the middle the model had polarized away | validated (.1001 granularity) |
| **3. Novel latent causes** | the world model lives in the **topology** (accumulated history), not the LLM's weights — built by traversal + **topology growth** (latent-cell deposit under coherence pressure; `doesNotUnderstand`). Novelty doesn't depend on training analogues because the model is *grown*, not retrieved | specced, not built |
| **4. Active information gathering** | frontopolar **Resolve** gather-before-commit — raises *"here's specifically what I'd need"* as a `perceive` primitive (the `<need>`). Active info-gathering is a *designed* loop, not a prompt the user has to supply | specced, not built |
| **Benchmark leakage / pattern recall** | the **topology-is-the-compression** thesis — cognition is the present the topology assembles; the LLM is a judge *over* it, not the seat of the model. The .1003 scale result (judgment scale-invariant, recall ceiling ~100k) is the first evidence the topology, not the window, carries the load | partially validated (.1003) |

The pattern: weaknesses **1 and 2 are the judgment layer** the comparator experiments already tested (and passed). **3 and 4 are the structural bets** — the topology as world-model, and Resolve as the active-gathering driver — specced but not yet built, and exactly where abduction stops being "the LLM does it" and becomes "the architecture does it."


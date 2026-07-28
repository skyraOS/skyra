# 2026-07-27

**A strip-down, not a fork.** `2026-07-26` stands as the record of what the design accumulated.
Nothing has been carried across. Things enter this set deliberately, one at a time.

The rule for this set: **strip complexity, do not add it.** A mechanism earns its place by being
required by the loop below, not by being coherent with the rest of the doc set.

---

## The loop

```text
an OBSERVATION comes in
    │
    ▼
it is assigned a SUPPORT VALUE against an expectation
    │
    ▼
given that expectation, what is the next ACTION
    │
    ▼
given that action, what is the next EXPECTATION
    │
    └──────────────────────────▶ (an observation comes in)
```

It is the four-cause loop:

```text
observation  ──▶  state change  ──▶  action  ──▶  observation
 (material)        (final)            (efficient)
```

The expectation *is* the final cause; the action *is* the efficient cause. Not annotations on
the loop — its two node types.

## The expectation field

There is no staked / unstaked distinction. **Everything is prediction.** The field is the
expectation field, and an expectation is the only kind of thing in it.

That is what makes the observation worker blind by construction rather than by rule: there is
no status for it to see, because status does not exist.

**An expectation is a state CHANGE, not a state**, and its action is the direct cause of that
change. Position measures **closure** — how much of the change has happened.

## Edge ownership

> **An expectation owns every edge it points at. It never owns an edge that points at it.**

Two kinds, and they are opposite relations:

```text
EVIDENCE   ──▶ observations / other expectations   EPISTEMIC — what grounds this change
ACTION     ──▶ its one action                      CAUSAL — what makes this change happen
```

The expectation sits between what it rests on and what closes it, and holds both.

**Incoming is a query, never a field.** *"Which expectations does `E` support?"* is a walk over
the edge set; *"candidate actions for `E`"* is a fold over `Q`. Neither is stored, which is what
stops an old expectation from being rewritten every time something new comes to rest on it.

So an expectation's record closes as soon as both halves exist — its evidence at the time, and
its action. **Nothing appends to it afterward.**

## What each worker cannot see

```text
OBSERVATION WORKER   never sees ACTIONS
                     a completed act leaks intent. it needs none: it assigns support,
                     and causation lives in Q, where the action is known anyway.

ACTION WORKER        never sees OBSERVATIONS
                     it runs on interpreted state only, so a raw payload it might
                     misread cannot reach it.
```

Each blindness is what licenses trusting the other's output.

## The observation worker

It sees the field from its own perspective — addresses, objects, everything in frame. It
assigns support values. It does not know what is wanted, what was claimed, or what any of it
is for.

**It sets the edge. The value is support, signed on `[-1,1]`.** That range is load-bearing: it
is what gives trust free depth attenuation and what stops chains from amplifying without bound.

**It scores what exists and never creates.** If an observation bears on nothing in the field,
the observation worker correctly does nothing. Novelty belongs to the expectation worker, which
sees observations too. That is what keeps this worker narrow enough to stay blind.

### What is in its frame

It cannot score every expectation against every observation, so the frame is selected — using
both quantities for exactly what each is for:

```text
ACTIVATION   SELECTS   which expectations this observation is RELEVANT to     tokens
TRUST        RANKS     which of those matter most                             mass
budget       CUTS
```

This is **trust's first concrete consumer.** It stops being a quantity we defined and becomes
the thing that decides what the observation worker looks at.

> **The observation worker is the system's operational definition of the world.**

There is no separate deterministic oracle to compare it against, and at a boundary there never
was — the world produces raw returns and something has to read them. As models improve, the
quality of that reading improves with no redesign. That is a property, not a compromise.

What it makes the most important measurable in the system is the interpreter's own stability,
and that is checkable without an oracle:

```text
revision rate    how often a reading is displaced when more evidence arrives
                 rising  →  the interpreter is unstable, and so is everything downstream
                 falling →  it is converging on a consistent reading of this world
```

## Three layers

```text
observation → expectation      LEARNS      interpretation      trust · truth
expectation → action           LEARNS      policy              actions
the transition itself          CONTROLS    runtime dynamics    dynamics
```

Two learners and one controller. Dynamics claims nothing, so it needs no grading arm.

---

## In this set

**One thing has entered.** `kernel-alteration` was moved from `2026-07-26` on 07-28 — the first
admission under this set's rule. It qualifies on the rule's own terms: the loop needs a substrate,
and this doc is where the strip-down had already happened. The node interface went from eight
methods to two, `Perturb` and `Observe`, both returning `Space`. Descent and ascent are closed
over the same type, so nothing leaves the algebra, and the kernel cannot classify what it walks
because it has no method that would let it.

It arrives carrying vocabulary this set does not define — `CreationPhase`, amplitude, `Cost`,
affordability, concrete/segmented. That material is marked in the doc rather than deleted, since
stripping it is a design decision and this was a move.

| doc | owns | status |
|---|---|---|
| [kernel alteration](kernel-alteration.md) | the substrate the loop runs on. `Perturb` / `Observe`, both closed over `Space`. the kernel executes what it is returned and does no cognitive math. | **DECIDED** — admitted 07-28 |
| [trust](trust.md) | accumulated observational magnitude, scaled by expectation edges. **unsigned mass** — unbounded, ranks, never settles. | **FINAL** |
| [truth](truth.md) | the normalized **signed** grounding. bounded. **the signal is its change over time, not its level.** | **TENTATIVE** — normalizer unsettled |
| [actions](actions.md) | expectations are state, actions are operators. **an action predicts SUPPORT, never an observation.** the attribution rule, `Q(E,A)`, selection. | — |
| [dynamics](dynamics.md) | the controller. global values derived from the transition, modulating the next pass. | — |
| [toy runthroughs](toy-runthroughs.md) | five adversarial traces for interpreter stability, correlated evidence, causal cardinality, action composition, and attention capture. | **TESTS FOR THE DESIGN** |

```text
T = Σ|c|      MASS        ranks — a well-refuted claim is as worth attending to as a
                          well-supported one, and a contested one more than either
G = Σ c       DIRECTION   which way it leans
|O|           COUNT       independent grounding observations
P = G / |O|   POSITION
```

All derived at evaluation, never persisted. The graph stores only edge values.

## Not yet in this set

```text
THE EXPECTATION WORKER   what it is handed, and what licenses it to create a new
                         expectation rather than support an existing one.
                         the only worker without a frame yet.

INDEPENDENCE             trust.md — distinct observation RECORDS are not necessarily
                         independent EVIDENCE. correlated sources inflate T, and
                         nothing here defines independence.

TRUTH NORMALIZER         truth.md — the mean is not monotone in evidence
Q's STATE CONDITIONING   actions.md — Q marginalizes over field state
THE ACTION WINDOW        actions.md — what bounds a support prediction

SLICE CONCURRENCY        kernel-alteration.md §13 Q4 — the kernel holds one lock for a
                         wave and cell methods never re-lock. a grouping Space that runs
                         its members concurrently either serializes, defeating the
                         fan-out, or races the transcript the Slice is authoritative for.
                         the one place a Space-owned decision reaches back through the
                         boundary.

THE SET'S OWN VOCABULARY kernel-alteration.md §§2, 6, 8, 9 — admitted with terms this set
                         does not define. either they earn entry by the loop requiring
                         them, or those sections strip.
```

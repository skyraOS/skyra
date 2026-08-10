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
CHILDREN   ──▶ 0:* expectations    EPISTEMIC — what grounds this change
ACTIONS    ──▶ 0:* actions         CAUSAL — what makes this change happen
                                   at least one of the two must be present
```

The expectation sits between what it rests on and what closes it, and holds both. Actions are
leaves and own nothing. See [expectations](expectations.md).

**Incoming is a query, never a field.** *"Which expectations does `E` support?"* is a walk over
the edge set; *"candidate actions for `E`"* is a fold over `Q`. Neither is stored, which is what
stops an old expectation from being rewritten every time something new comes to rest on it.

### What is immutable, and what is not

An earlier draft here said *nothing appends to a record afterward*. Read literally that freezes
`O(E)` at creation, which freezes `P` for all time — and the whole set runs on `ΔP`. The
substrate is append-only, not immutable, and the rule is narrower than it was stated:

```text
ACTION EDGE      write-once. it closes the record.
SUPPORT VALUE    immutable once written. never revised in place.
EVIDENCE SET     APPEND-ONLY. new grounding may arrive at any time.
```

`P` therefore moves by accretion and never by rewrite. That is what the original sentence was
protecting — an old expectation is not *rewritten* when something new comes to rest on it — and
it holds without freezing the evidence.

## Two workers

There were three. The action worker is folded into the expectation worker — see
[workers](workers.md) for the full statement.

```text
OBSERVATION WORKER    grades state      how far are we from target     CURRENT
EXPECTATION WORKER    proposes state    what should be true, and how   TARGET + BRIDGE
```

The observation worker emits **two numbers** — `relationship ∈ [-1,1]`, signed, which is the
only thing that enters the arithmetic; and `confidence ∈ [0,1]`, which routes and must never
weight. Together they separate *unrelated* from *not understood*, two states one number
conflates at the bottom of its range.

The rule that licenses this is narrower than the two blindness rules it replaces:

> **Blindness is required of ASSESSORS, not of claimants.**

The observation worker is the assessor and is still blind to actions. Both other roles are
claimants, and claimants may see whatever helps them claim well, because something blind grades
them afterward. Merging two claimants changes nothing about the licensing.

The cost is real and stated in [workers](workers.md): the worker that sets targets now also sees
what has worked, so targets will tend to bend toward existing competence. That cannot be
prevented while the merge holds. It can be measured.

## Three layers

```text
observation → expectation      LEARNS      interpretation      trust · truth
expectation → action           LEARNS      policy              actions
the transition itself          CONTROLS    runtime dynamics    dynamics
```

Still two learning problems and one controller — but two learners is now a statement about
**signals**, not about workers. Both learning signals land on the expectation worker; they must
not be run together. Dynamics claims nothing, so it needs no grading arm.

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
| [workers](workers.md) | the two workers. what each owns, what each sees, and the rule that licenses the arrangement — **blindness is required of assessors, not claimants.** | **TWO WORKERS** — 07-30 |
| [expectations](expectations.md) | the shape. `0:*` actions and `0:*` expectations, grading at leaf actions, the falsifiability invariant, order. | **MODEL SETTLED** — 07-30 |
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
CREATION LICENSE         the expectation worker now HAS a frame — activation high,
                         support ≈ 0 (workers.md). what remains open is what licenses
                         CREATING a new expectation rather than refining an existing one.

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

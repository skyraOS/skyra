# Workers

**Date:** 2026-07-22
**Status:** OPEN LIST. The registry of v4 workers and their responsibilities. Grows as the
port is designed; nothing here is built.

**Reads on:** [v4-port](v4-port.md) (the cycle inventory these are derived from),
[v4-data-model](v4-data-model.md) (**the vocabulary — read first**),
[time](time.md) (proper time — what `Within` is denominated in),
[dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md).

**Vocabulary.** This doc uses v4 words. A *PerceivedAct* is what happened; an *Observation*
is the mapping many PerceivedActs → *Beliefs*; a *Belief* is a set of triples and the unit
of heat; an *Expression* is the mapping beliefs → *EngageActs*. Full table in
[v4-data-model](v4-data-model.md) §1.

**Worker names below predate the final vocabulary** (`sensory:interpret` performs an
Observation; `plan:articulate` assembles an Expression's beliefs). Renaming them is a
pending pass, not a decision — the responsibilities are unchanged.

---

## Conventions

**Naming** — `namespace:gene`, carried over from v2.

**Cost** — `pure` arithmetic only · `model` one inference call · `world` crosses the
membrane.

**Dispatch** — `routed` the transition table decides whether it runs · `write-path` invoked
directly by the observation write and never scheduled. A `write-path` worker must be
unreachable from the table, report no `Condition`, and be transcribed onto a type with no
`Infer` method. A missing table entry must never be able to skip one.

**Exists** — is this behavior in the v3 code today, or is it new work?

---

## The list

| # | Worker | Responsibility | Cost | Dispatch | Exists |
|---|---|---|---|---|---|
| 1 | `sensory:interpret` | **PerceivedActs → Beliefs.** Reads the raw payloads and emits structured claims, recording one **Observation** carrying the channel's fidelity. Preserves ambiguity as multiple competing Beliefs from the same Observation; never hard-links. | **model** | routed | **NO — new** |
| 2 | `sensory:intake` | Opens the external span. Writes the PerceivedActs, the Observation, and its Beliefs into their stores. | pure | routed | yes |
| 3 | `sensory:activate` | Decays the excitation field once, injects at the impulse's terms, excites the reverse-index postings that mention them. | pure | routed | yes |
| 4 | `sensory:propagate` | Propagates excitation across learned causal edges to `HopLimit`, weighted by `edge.P × edge.confidence()`, cut at `ε`. | pure | routed | yes |
| 5 | `sensory:target` | Reads activated measured edges into `TargetGroup`s; projects a target only if every group is confident and committed. | pure | routed | yes |
| 6 | `grade:consequence` | Settles matching open bets against a landed observation; folds the outcome into the edge EMA. | pure | **write-path** | yes |
| 7 | `grade:interpretation` | Same, for Beliefs, through their own falsification address. Action success never enters as backward credit. | pure | **write-path** | yes |
| 8 | `plan:close` | Computes holes — requires that nothing in the plan produces and nothing observed satisfies. Set membership, no threshold. | pure | routed | yes |
| 9 | `plan:articulate` | Assembles the present and **records `Expression.Beliefs`** — which Beliefs were in play and how hot each was. Propagate, gather candidates, select under a token budget. **Replaces three hand-rolled traversals with one budgeted descent.** | pure | routed | partly (assembles, never records) |
| 10a | `plan:express` | Given a belief set, the best set of acts. Owns `EngageAct`. **Fires on surprise or incompleteness, not per pass.** | **model** | routed | yes (fused, ungated) |
| 10b | `plan:expectation` | Given a set of acts, their expectations. Owns `Expectation`. May own `Within`; may NOT own confidence. | **model?** | routed | yes (fused) |
| 11 | `plan:place` | Parse, validate, stamp modes from the world's affordances, index into the expression graph. | pure | routed | yes |
| 12 | `act:select` | Returns acts whose every `Requires` is satisfied by *observed* state. Derives parallelism and conditionals; declares neither. | pure | routed | yes |
| 13 | `act:do` | Executes one act against the world and returns a terminal outcome. **The leaf that crosses into consequence, and where a device-place attaches.** Needs a deadline and must be able to terminate `indeterminate`. | **world** | routed | yes (no deadline) |
| 14 | `act:complete` | Arms expects whose `Of` acts have all completed. Must be reachable from the deadline path, not only the success path. | pure | routed | yes |
| 15 | `settle:sweep` | Runs perceives **across** each open bet's `Within` window, not once at maturity. No model call — the expect names its own address. An expired window with no confirmation turns the bet false. | pure + **world** | routed | partly (one perceive at maturity) |
| 16 | `reconcile:check` | Is every target Belief *observed* true. Not projected — observed. | pure | routed | yes |

**17 workers: 2–3 call a model, 2 cross the world boundary, 2 are write-path, the rest are
pure arithmetic.** The range is because `plan:express` and `plan:expectation` are a settled
boundary whose call count is open — see below.

---

## Notes on individual workers

### 1 · `sensory:interpret` — the only genuinely new one

Today `ObserveImpulse` takes `[]Triple` and every caller hands it pre-made triples
(`cmd/dnav3/main.go:79` is a literal; `cmd/globe` passes world-produced `seed.Effects`).
Nothing turns a raw payload into structured claims. This worker is that seam.

Three constraints the existing code imposes on it:

**It cannot see the bets.** LAW 6 — the pass that interprets an Observation into Beliefs
must not see open bets, their confidences, or the target. Only the entity and predicate
vocabulary. This is an interface constraint, not a discipline: if it can see what answer
was wanted, it resolves ambiguity toward the being's own bets and manufactures the evidence
that grades them.

**It is the second model call in the system, and it lands at intake.** `being.go:39`
currently claims `Infer` is the one model call site. Intake is presently free — no model
call, no tick advance. This worker changes that for every impulse, not just for planning.

**Its output is two-provenance, and today's front door cannot carry that.**
`ObserveImpulse` hardcodes `ProvenanceRaw`, confidence `0`, `DerivedFrom` nil
(`ingest.go:40`). Interpreted output pushed through it would be laundered to fact authority
(1.0) — the exact failure `observation.go` exists to prevent. Either the translator routes
its readings through `ObserveInterpreted`, or `ObserveImpulse` stops assuming raw.

### 6, 7 · Why grading is not routed

`grade` is invoked unconditionally from inside `Observe()` (`being.go:184`, `:194`). Every
observation write grades; no table is consulted. v2's router treats a missing entry as an
**inert edge — nothing dispatched**, which is right for cognition and silently wrong for
grading. Routed grading would make LAW 5 depend on table completeness.

### 9 · `plan:articulate` is where the port actually happens

It absorbs `activateQuery`, `activeMemoryNodes`, and `SelectMemoryDetailed` — three
separate traversals run back-to-back on every planning pass, one of which re-does
propagation the intake path already performed. Whether it becomes attenuation
(`traverse.Assemble`) or stays density-greedy is [v4-port](v4-port.md) §5 question 2, and
it is the real technical fork.

### 13 · `act:do` is the device seam

The only worker whose implementation differs between a sandbox filesystem, a Mac, and a
TV. Everything above it is unchanged by which world is mounted.

---

## When planning fires — the surprise gate

Planning does not run per pass. **It runs on divergence.** (This gates `plan:express`; if
planning is two calls, `plan:expectation` follows it rather than being gated separately.)

```text
SURPRISE          the world diverged from what I claimed        |grade| > θ
INCOMPLETENESS    I have no claim here at all                   holes, or N = 0
```

Both triggers, not one. Surprise alone is degenerate at cold start: a being that has
predicted nothing can never be surprised, so it would go quiet exactly when it knows least.
Incompleteness is what `holes` already is, and what `prior() = 0.5, N = 0` already flags.

### What this changes

v3 fires the model on **structural incompleteness** (holes) and **unmet target** (the
reconcile loop). Neither is surprise. Gating on divergence makes the first thesis curve
structural rather than emergent:

```text
today        passes-to-closure drops BECAUSE precedent surfaces a good plan   (emergent)
this rule    passes-to-closure drops BECAUSE a calibrated being is rarely surprised
```

A warm being running a known task makes the irreducible framing call and then just *runs* —
no call per pass, because nothing diverged. That matches doc.go's stated floor exactly:
*"the irreducible call is the FRAMING, not the plan."* **One call, plus one per surprise,
is a complete cost model.**

### θ is the knob

`|grade|` is already the magnitude — `log p_claim − log p_prior`. So the gate is
*fire when |grade| exceeds θ*, and θ is literally **how wrong the being has to be before it
thinks about it.**

```text
θ low     ruminates on noise
θ high    ignores real divergence
```

It is the one tunable that trades cost against adaptivity in a single number, and it is
measurable: θ against `Expression.Spent` against grade. **OPEN:** its value, and whether it
is global or per-namespace.

### The absence case runs through the same gate

An expired `Within` window with no confirmation turns the bet FALSE. That is a belief
change, and a belief change that diverges from the prediction is a surprise. So *"look, see
nothing, then ask"* needs no special path — it is the ordinary gate firing on an ordinary
divergence. See [v4-data-model](v4-data-model.md) §3.

---

## The model calls — what each may see

**LAW 6 makes a single combined call structurally impossible.** At minimum there are two
paid workers; whether planning is one call or two is open (below).

> *"The pass that turns an impulse into observation triples must not see the open bets,
> their confidences, or the target... Otherwise the being resolves ambiguity toward its own
> bets and manufactures the evidence that grades it."*

The planning half needs the target. The interpreting half must not see it. One call would
have to simultaneously see it and not see it.

### Why the interpret/plan split is forced

The planning half needs the target. The interpreting half must not see it. One call would
have to simultaneously see it and not see it. That boundary is settled and is not a style
choice.

---

## `express` / `expectation` — boundary SETTLED, call count OPEN

Two competences, and they are not the same skill:

```text
express       given a belief set  →  what is the best set of acts
expectation   given a set of acts →  what are the expectations
```

**A worker boundary is an ownership boundary, not a call boundary.** Two workers can consume
one structured emission; one worker can make two calls. Deciding the boundary does not
commit the call count.

### SETTLED — the boundary

| | |
|---|---|
| **struct ownership** | `express` owns `EngageAct`; `expectation` owns `Expectation`. No shared write. |
| **separate validation** | act well-formedness and bet well-formedness are different checks failing for different reasons |
| **reversibility** | with the boundary in the code, splitting the call later is a local change; without it, splitting later is a rewrite |

The third is why the boundary lands now and the call count waits — that ordering keeps the
decision reversible rather than deferring it into a corner.

### OPEN — one call or two

The invariant holds either way; only the mechanism differs:

```text
ONE call    LAW 5 is a VALIDATION   Validate() rejects an act with no expect   (v3 today)
TWO calls   LAW 5 is a GATE         no act reaches the world without its expectation
```

The two-call form is threadable because models stream and `parse.go` was already built for
it — *"XML envelope (self-delimiting — a truncated emission still yields every complete
element before the cut) with line-oriented triples inside (per-line failure isolation)."*
Acts release as delimiters land, each held until its expectation arrives. A stream dying
mid-act leaves nothing to clean up, because nothing was released.

**What actually decides it: whether `expectation` needs a different present.** If it wants
past predictions at these addresses and their outcomes, and the planner does not, it wants
its own call. If it can work from the planner's frame, one call is cheaper. That is a
question about *presents*, answerable by writing out what each would contain — not by
argument. The presents table below currently gives them one column; splitting that column
is the same decision.

### Attribution does NOT require the split

An earlier draft of this doc claimed a fused call makes `Expression.Score` un-attributable.
That was wrong. Attribution needs two **measurements**, not two calls, and both already
exist — `targetMet` and `GradeSum`:

| | predictions accurate | predictions wrong |
|---|---|---|
| **target met** | both good | planning ok, **prediction bad** |
| **target not met** | **planning bad**, prediction good — it correctly called its own failure | both bad |

Same shape as the two-bit outcome, one level up. The bottom-left cell is the useful one and
it is legible today, with no split at all.

### The debt the split would create

In one call the model picks acts **knowing it must bet on them** — a real constraint on act
selection: *do not choose an act whose outcome you cannot call.* Splitting removes that
constraint, and nothing currently replaces it, because there is no act-side learning.

The eventual replacement is already in the model: `Expectation.By` makes edges **act-keyed**,
so acts that reliably produce uncallable outcomes accumulate weak edges, and weak edges sink
in retrieval. That path is unbuilt. Between now and then, a split `express` is less
constrained than today's fused one.

### What `expectation` may and may not own

**May own `Within`.** Model-guessed today, and it has no business being. *"Claims at this
address historically settle within N ticks"* is counted, not judged — the settlement log has
the latencies. This is the slice that measurably improves with counts.

**May NOT supply confidence.** If a calibrator sets `p_claim` from the edge's own history,
then `p_claim == p_prior` and `Grade = log(p_claim) − log(p_prior)` is **zero by
construction** — the model parroting the base rate by design, which is exactly the failure
`Expression.Score` exists to detect. Confidence stays the model's own stake or the
measurement measures nothing.

---

## The presents

| | `sensory:interpret` | `plan:express` / `plan:expectation` |
|---|---|---|
| the impulse | ✓ | ✓ |
| current Beliefs (world state) | ✓ — for coreference: *"he"*, *"it"*, *"that file"* | ✓ |
| entity + predicate vocabulary | ✓ | ✓ |
| past **Interpretations** for similar impulses, with own-address outcome | ✓ | — |
| closed-run credit on those readings | ✓ | — |
| the target | **✗ LAW 6** | ✓ |
| open Expectations and their confidences | **✗ LAW 6** | ✓ |
| precedent Expressions with base rates `[worked 63% of 8 tries]` | **✗** | ✓ |

### Why interpretation history is safe to show — and what makes it safe

LAW 6 names *open* bets. Settled history is not desire; it is accuracy. But the reason that
holds is narrower than it looks:

```text
"you read X as `michael wants water`, and that reading's OWN test failed"
    → a fact about your accuracy. carries no target.

"you read X as `michael wants water`, and the PLAN built on it failed"
    → carries the target backwards. exactly what LAW 6 forbids.
```

**LAW 6 and the no-backward-credit rule are the same protection from two ends.** v3 is
explicit that *"action success never enters this function as backward credit."* That is not
only epistemic hygiene — it is the precondition for interpretation history being showable
at all. If interpretations were graded by downstream success, their history would encode
desire and could never enter the interpreter's present.

### The residual leak, and the boundary — OPEN

Own-address settlement is LAW-6-clean but **sparse**: it requires every model reading to
author a discriminating prediction, which is why `ProposeInterpretation`'s bar is high
enough that almost nothing clears it. Credit flow-back is the denser signal — and it
reintroduces the leak, because `ArticulationMember.Credit = Activation × Articulation.Grade`
and that grade is downstream of the target.

Proposed boundary, by aggregation:

```text
credit from the CURRENT run       → target-shaped. withhold from the interpreter.
credit accumulated over CLOSED runs → a statistic about this reading's general
                                      usefulness. no live target in it.
```

Same logic as held-out evaluation. *"This reading has earned negative credit across nine
finished runs"* is learning; *"the plan you are in the middle of is failing"* is a hint.

### Selection: heat vs variance — OPEN

Both presents are assembled under a token budget, and today selection is purely **heat**
(`Activation / TokenCount`, greedy). What a learning loop wants is partly
**informativeness** — where the graph is uncertain and a model call could actually move it.

```text
heat      what is relevant right now
variance  where my model is weakest, so the call is worth paying for
```

They compete for one budget. All-heat re-reads what the being already knows; all-variance
is unmoored from the situation. So it is probably a **split**, not a choice.

This is measurable rather than a matter of taste: `Articulation.Spent` against
`Articulation.Grade` says whether the split is paying (see
[v4-data-model](v4-data-model.md) — the spend curve). The raw material already exists —
precedent nodes render `ed.P` and `ed.N`, mean and sample size, enough for a posterior
spread.

Variance also justifies its own slot on first principles: **a reading the graph is already
confident about does not need a model call.** The uncertainty is the case for spending the
money — which is the same argument `Expression.Score` makes when it warns that a score near
zero means the model is parroting base rates and not earning its keep.

---

## Ordering constraints between workers

Not everything is free to be reordered by the router.

**Intake is batched, not per-claim.** `ObserveImpulse` runs three separate passes — write
all, then activate once, then grade all — with the comment *"The complete impulse is
durable before activation and grading begin"* (`ingest.go:49`). A worker chain that ran
`intake → activate → grade` per claim would grade claim #1 against a world state missing
#2 and #3, and would decay the field three times instead of once.

> `sensory:intake` runs to completion over the whole impulse before `sensory:activate`
> runs at all. Decay is once per external impulse, never once per claim.

**Arming precedes observation of the act's own effects.** `act:complete` must run before
the effects of `act:do` are appended, or a bet settles against evidence that predates it.

# Interaction Primitives — the Contract of the Excitation Field

**Date:** 2026-07-22 · **updated 2026-07-23**
**Status: SUSPENDED, WITH ONE DECIDED SECTION.** §1–§8 are recorded, not being acted on.
**§9 is a deliberate exception, solidified in-session on 2026-07-23:** it establishes the
excitation field as the **first layer of traversal**, which is a decision about the *shape*
of the field — not a resolution of the §5/§7 v4 keyspace questions, which remain suspended.
§9 was allowed to land ahead of the worker layer because it is a boundary (field vs.
selection) rather than an interface, and because it changes what the later questions are
about. Where §9 refines an earlier section it says so.

The current work is otherwise at the **ephemeral-worker layer** — [v4-port](v4-port.md) and
[workers](workers.md). Everything in §1–§8 below is measured from `activation_field.go` and
is accurate as a description of v3; it is simply not the thing being built yet, and none of
its OPEN questions should be resolved ahead of the worker layer.

Picked up again when the worker boundaries are settled and the field needs a contract to
sit behind them.

**Name.** v3 calls this `ActivationField`. v4 calls it the **Excitation Field**, and the
quantity it holds is **excitation**. The word already carries this meaning in the project —
[traversal-as-excitation](../2026-07-14/traversal-as-excitation.md): *"excitation is not a
separate cognitive operation performed before traversal. The traversal itself produces the
excitation and the active field."* That is the same claim as §6: transport is the field's
response, not a call.

**Reads on:** [v3-mechanisms](v3-mechanisms.md) (the field is flagged there as the least
encapsulated thing in the package), [v4-data-model](v4-data-model.md) (the field is
explicitly *outside* §7's storage law — it is transient and is not expected to be durable).

---

## Why this doc

v3's `ActivationField` has **five writers and no owner**. `bumpObservation`, `heatTriple`,
`heatPosting`, `propagateActivation`, and `BeginImpulse` each touch a different subset of
five maps. To answer *"why did that surface?"* you have to hold all five in your head at
once.

Naming the primitives turns that into a **closure claim**:

> **These are the only ways excitation moves. Everything else calls them.**

That is the same discipline as §7's storage law, applied to the one structure the law
deliberately does not cover.

---

## 1. The algebra — separate from the primitives

```go
bump(a, h) = min(1, a + h)      // saturating add
```

Additive and **capped at 1**. Repeated mention saturates; nothing dominates the field by
being said ten times.

This is the *combination rule*, not a primitive. It could be replaced (max, log-sum,
normalised) without touching a single primitive — which is exactly why it is stated apart.

---

## 2. The primitives — THREE, not four

An earlier draft of this doc listed four (DECAY / STRIKE / SPREAD / CONDUCT) and put SPREAD
under diffusion. **That was wrong**, and the error is worth keeping visible because it is
the kind that hides in a name.

The real partition is by **what each one reads**:

```text
DECAY       SINK        a ← a·λ,  drop below ε         reads its OWN state
SOURCE      INJECT      h at a key, or A/n over an     reads NOTHING
                        index; gated at ε
TRANSPORT   COPY        a · P · N/(N+k),  gated at ε   reads CURRENT state · ITERATES
```

> **The line is not point-vs-distributed. It is: is the amount a function of the field's
> current state, or not?**

**TRANSPORT does not move anything.** An earlier draft called it "the only thing that moves
excitation." It copies — see below. The source is never depleted.

### DECAY — the only sink

```go
a *= λ
if a < ε { delete }        // deletion, not clamping — the field stays sparse
```

Runs **once per external impulse**, in `BeginImpulse`. Causal descendants and planning
queries bump without advancing it, so **excitation is measured in impulses** — not ticks, not wall
time. Four impulses of silence and an entry is gone: `1 → .5 → .25 → .125 → .0625 → ∅`.

### SOURCE — injection, at one key or many

Two fan-outs of one primitive. Same amplitude budget; different spread of it.

```go
// point:       amplitude 1.0 at ONE key
BumpEntity(term, 1)

// distributed: amplitude 1.0 across N keys → 1.0/N each
share := amplitude / len(spans)
if share < ε { return false }      // ← the discrimination gate
```

Neither reads the target's current excitation. `heatPosting` touches `Activation.Spans[span]`
exactly once — to detect *whether* anything changed — and never to compute *how much* flows.

The distributed form divides amplitude by how many places mention the term, so **a rare term
excites its few targets strongly and a promiscuous one excites each weakly.** That is IDF,
arrived at by conservation rather than by a log.

Its gate is load-bearing, and the code says why:

> *"A posting so broad that no member can clear the activation floor carries no
> discriminating information. Do not scan an unbounded `said`/`state`/user list."*

**A broad term costs nothing rather than costing everything** — the scan is skipped, not
performed and discarded. Then it recurses one level, `share / len(matched)`, gated again.

Injection amplitudes in v3:

```text
direct    span · node · entity (×3 terms) · triple key    1.0
posting   entity    term → spans                          1.0    broad recall
          address   (subj,pred) → spans                   0.5    same address, any value
          triple    full key → spans                      0.5    exact claim
```

### TRANSPORT — copies along learned edges; the source is never depleted

```go
for hop := 0; hop < HopLimit; hop++ {
    for _, source := range RankedSpans() {                    // ← reads CURRENT excitation
        crossing := source.Activation × ed.P × ed.confidence()
        if crossing < ε { continue }
        ...
    }
    if !changed { return }                                    // ← iterates to fixpoint
}
```

with `confidence = N/(N+5)`, `HopLimit = 2`.

**This is spreading activation** in the classic form — `a × w`, summed, iterated, cut at a
threshold. It is **the only place the field is not a bag of words**: the amount is weighted
by how reliably an edge proved (`P`) and how much evidence backs it (`confidence`). An
untested edge (`N=0`) contributes nothing, which keeps *untested* apart from *wrong*.

Precedent therefore surfaces itself, with no query issued.

#### It creates, it does not conserve

**Nothing decrements the source.** The source keeps its full activation and every target
gains `a × P × conf`. Total field mass goes *up* at every hop.

And one line above the crossing calculation:

```go
b.Activation.BumpNode("expression:"+e.ID, source.Activation)
```

**Full source activation, no attenuation whatsoever** — no `P`, no `confidence`. Every
expression in a hot region receives that region's entire excitation, copied. So expression
activation is purely **structural** (co-location), never **learned**, which is a different
rule from every other crossing in this primitive and is nowhere stated.

A region with 5 expressions × 3 expects emits 15 attenuated crossings *plus* 5 full-strength
copies, from a source that loses nothing. `min(1, ·)` per node and `HopLimit = 2` are the
only things bounding it.

#### The fan-out asymmetry — real, and undeclared

The two primitives take **opposite policies on fan-out**:

```text
SOURCE      share = amplitude / |targets|      DIVIDES by fan-out
TRANSPORT   crossing = a × P × conf            does NOT divide by fan-out
```

Both are arguably correct, for different reasons:

```text
SOURCE      about the INFORMATIVENESS OF A TERM — a term in 40 places says less per place
TRANSPORT   about the RELIABILITY OF A LINK    — "A reliably causes B" does not weaken
                                                  because A also causes C
```

So the asymmetry is probably right. But it currently sits in the code as an unexplained
difference, and it should be a stated decision. **OPEN.**

#### The risk it creates: hubs

A claim appearing in many expressions, each with reliable edges, broadcasts at full strength
once per expression per hop, from a source that is never depleted. Nothing structural
prevents a hub from dominating the field — **inflation is bounded by clamps, not by
structure.**

Tolerable at current scale. Less so now that the run wipe is gone and memory is unbounded
(see [v4-data-model](v4-data-model.md) §4b). **OPEN.**

### Why the collapse matters

Naming SPREAD as a transport primitive implied excitation flowed hot-to-cold between related
things. It does not. Injection is **stateless** — the same impulse produces the same
distribution regardless of what the field already holds. Only TRANSPORT is state-dependent,
and only TRANSPORT is learned.

---

## 3. The keyspace — and two dead maps

v3 has five maps. **Two are write-only.**

| map | written by | read by |
|---|---|---|
| `Spans` | decay, strike, spread | `SpanActive` (the grading gate), `RankedSpans` (retrieval), `heatPosting` |
| `Nodes` | decay, strike, spread, conduct | `memory.go` — `MemoryNode.Activation`, which drives density selection |
| `Edges` | decay, conduct | once, in target derivation (`activation_field.go:394`) |
| `Entities` | decay, strike | **nothing** |
| `Triples` | decay, strike, conduct | **nothing** |

`Entities` and `Triples` are decayed and bumped on every impulse and **never read — not once,
not in a test.** Same class as `BornFrom` and `Observation.Confidence`: a field that exists,
costs work, and informs nothing.

The live field is three maps, and one of those is read once.

---

## 4. The constants

```text
λ         = 0.50    decay rate, per impulse
ε         = 0.05    THREE different jobs — see below
SpanFloor = 0.05    the grading gate (defaults to ε)
HopLimit  = 2       conduct depth
confHalf  = 5.0     evidence half-point in confidence
cap       = 1.0     bump saturation
```

### ε is doing three unrelated jobs

```text
decay floor      "this is too cold to keep"
spread gate      "this posting is too broad to discriminate"
conduct gate     "this edge is too weak to conduct"
```

Three different questions sharing one number, because nobody separated them. The spread gate
is the load-bearing one — it is what prevents an unbounded scan — and it is the one most
likely to want a different value from the other two.

**Naming the primitives makes this a choice rather than an accident. OPEN.**

---

## 5. Questions this raises

### SOURCE has two origins and does not distinguish them

`Place` bumps the field when the being **emits**:

```go
b.Activation.BumpNode("expression:"+e.ID, 1)
b.Activation.BumpSpan(e.SpanID, 1)
```

So an observation landing and the being's own emission are the **same primitive at the same
amplitude**. Under LAW 2 those are opposite sides of the boundary.

It is also the concrete mechanism behind the self-fulfilment leak: **predicting something
excites it**, so a prediction about what will be hot is partly self-fulfilling through the act
of recording it.

That may be correct — a being probably *should* attend to what it just committed to. But it
should be a decision, and today it is an unmarked coincidence of amplitude.

**Resolved in §7: this is a phase property of the node, not a property of the injection.**

### What does the field key on in v4?

Spans are gone (§4a of the data model). `Belief` is the excited unit, explicitly. And `Nodes`
already holds heterogeneous IDs — both observation IDs and `"expression:"+id`.

So the v4 keyspace is plausibly **one map**: `ID → excitation`, over Beliefs and Expressions —
with terms, addresses, and claim keys demoted to *indexes used to route excitation*, rather than
things that hold excitation themselves.

That would be five maps → one, and it is exactly what the two dead maps are already telling
us: `Entities` and `Triples` were never excitation, they were routing. **OPEN, and the single
biggest simplification available here.**

### Does `Edges` survive?

Read once. Edge excitation is a function of the source belief's excitation and the edge's own `P` — both
available at read time. It may be derivable rather than stored. **OPEN.**

---

## 6. The interface — what gives the field an owner

**Only one of the three primitives comes from outside.** Read the recurrence again:

```text
aᵢ  ←  min(1,  λ·aᵢ  +  Σⱼ wᵢⱼ·aⱼ  +  sᵢ )
               └────────┬────────┘     └┬┘
              the field's own step    the only input
```

DECAY and TRANSPORT are not invoked. No worker decides to propagate — propagation is the
field's **response** to having been injected into. No worker decides to decay — decay is
what an impulse boundary *means*.

So the interface is small:

```go
// WRITE — the only two doors
field.Impulse(claims)    // decay, then inject, then propagate to fixpoint
field.Derive(claims)     // inject, then propagate. NO DECAY.

// READ
field.Excitation(id) float64
field.Ranked() []Activation
field.Active(id) bool     // the grading eligibility gate
```

**The `Impulse` / `Derive` split IS the invariant.** *"Decay happens exactly once per
external impulse"* is currently a comment plus a convention spread across five call sites.
As two methods it is structural: you cannot decay twice because `Derive` has no decay to
call, and you cannot skip it because `Impulse` does not let you.

Workers never touch a map, never call decay, never call propagate, and never control the
order.

### Two workers collapse

| worker | fate |
|---|---|
| `sensory:activate` | **gone** — this is `Impulse`'s decay + inject half |
| `sensory:propagate` | **gone** — this is the field's own step |
| `sensory:intake` | calls `field.Impulse(...)` |
| act effects · settle returns | call `field.Derive(...)` |
| `plan:articulate` | reads `Ranked()` |
| `grade:*` | reads `Active(id)` |

Those two were workers **only because the field had no owner.** They are not units of work;
they are stages of one operation that got named separately because five different functions
were each doing a piece of it. See [workers](workers.md) — 17 → 15 when this lands.

---

## 7. Phase — a property of the node, not of the injection

The being's own emissions currently inject at the same amplitude as world observations (§5).
The fix is not a third write door. **Origin is a property of the thing being excited, not of
the act of exciting it.**

### The concept already exists at both ends of the pipe, and nowhere in the middle

```text
Provenance      raw 1.0 · interpreted 0.3 · mark 0.0     fixed at write
MemoryNode.Kind memory · current · mark · precedent      computed at render
THE FIELD       excitation is excitation                             PHASE-BLIND
```

And `observation.go` already states the rule this would enforce:

> *"A MARK MAY BE PERCEIVED AS A FACT ABOUT ITSELF. IT MAY NEVER BE EVIDENCE FOR A CLAIM
> ABOUT ANYTHING ELSE. It is a SEARCH AID, not EVIDENCE."*

**A mark may excitation. It may not ground.** That is exactly the distinction the three-door
version was trying to encode at the call site.

### Why "phase" rather than "provenance"

Provenance is fixed at write. Phase **transitions**, and a real transition is happening:

```text
PROJECTED   I claimed it. no external support.
     │      ↓ the world reports it
OBSERVED    it came back. grounded.
     │      ↓ a bet on it resolves
SETTLED     it has a history now.
```

The same claim genuinely occupies different epistemic states over its life, and settlement
is the transition. Provenance cannot express that, because it never changes.

### It costs no field state

```text
the field stays    ID → float64      one map, unchanged
phase is read      from the Belief   at the three places that branch on it
```

No new dimension, no per-phase maps, no growth in what the field owns. A lookup, not a
field.

Which is also why the write interface stays at two doors: `Impulse` vs `Derive` is about
**decay timing**; phase is about **what is being excited**. Two axes that a three-way method
choice had tangled into one.

### What it makes askable

All three are currently unaskable:

**Should projected excitation conduct?** If the being predicts X, and X has edges to Y, should Y
excitation? That is reasoning forward from an unconfirmed assumption — useful in planning, and
**exactly what rumination is** when it loops. Phase makes it a policy rather than an
accident.

**Should projected excitation decay faster?** It has no external support holding it up. Today it
decays at λ like everything else.

**Should retrieval budget by phase?** *"At most X% of the frame may be the being's own
projections."* That is a direct anti-fiction rail on the present, and today nothing prevents
a frame composed entirely of the being's own prior claims.

The third is the self-fulfilment leak with a dial on it instead of a warning in a doc.

### The guard on this

Three phases is a taxonomy, and taxonomies grow. The test that keeps it honest is the one
used throughout this doc set: **does anything need to branch on this value?**

Projected-vs-observed clearly does — conduct, decay, and budget all want it. Whether
`settled` earns a third phase is less clear, since *"has a settlement record"* is already a
lookup. **OPEN.**

---

## 8. What this is not

The field is **transient**. §7's storage law does not cover it and should not: it is
rebuilt continuously, decays by design, and losing it costs nothing but recency.

But *outside the law* and *unowned* are different things. This doc is the contract that
makes it owned.

---

## 9. The field is the first layer of traversal — DECIDED 2026-07-23

Everything above §9 describes the field as a thing that holds and gates excitation. **That
was the category error.** The field does not gate anything. It is a **heatmap** — a potential
landscape — and it is the **first layer of traversal**, not a store that traversal later
consults.

> **The field heats and decays. It never selects. Every bound in the system lives in the
> walk, never in the map.**

### 9.1 Three layers, previously tangled into one map

```text
HEATMAP      SOURCE + DECAY          the potential landscape. unbounded. no cutoff.
                                     heat pools where the impulse and precedent put it;
                                     it fades where they do not. that is the whole job.

CONDUCTOR    the traversal EDGE      what the walk crosses. carries from.heat × P × conf.
                                     NOT a field cell — see 9.3.

WALK         TRANSPORT + budget      the traversal itself. THE ONLY THING THAT BOUNDS.
                                     enters at the peaks, descends until the budget is spent.
```

TRANSPORT is not a field primitive that runs *before* traversal. **It is the traversal's
first step.** This is [traversal-as-excitation](../2026-07-14/traversal-as-excitation.md)
made structural: *"the traversal itself produces the excitation and the active field."* The
walk produces the heat; the heat does not precede the walk.

### 9.2 No selection lives in the field — which splits ε by layer

The doc's own §4 finding (*"ε does three unrelated jobs"*) resolves once selection is removed
from the field. The three jobs were never one constant because they were never one layer:

| job | layer | fate |
|---|---|---|
| decay deletion floor | HEATMAP | **housekeeping.** a sub-ε entry is ≈0; drop it to stay sparse. NOT a threshold — "don't store dust." |
| SOURCE discrimination gate | HEATMAP | **stays.** the IDF gate: skip an uninformative broad scan. it shapes *injection*, not selection. |
| retrieval / grading gate | WALK | **leaves.** this is the only "threshold" in the selection sense, and it was never the field's to hold. |

**`SpanActive` gets truer, not broken.** The grading-eligibility gate — *"was the birth
region attended when the observation landed"* — needs no floor, because the field is
sparse-by-decay: **presence is the gate.** Still in the map ⇒ warm; decayed out ⇒ cold. Cold
is not *"below 0.05,"* cold is *gone*. One constant (`SpanFloor`) dissolves and the semantics
improve.

### 9.3 The stored value is a Conductor, not a scalar

The field's defect is that it stores the **product** and discards the **factors**:

```text
v3         Triples[K] = 0.35              injection? propagation? how old? unknowable.
v4         Conductor{From, P, Conf, K}    every factor still present; Heat() derives 0.35.
```

A **Conductor** is the excitation of a single triple, realized from the entity path. Named to
avoid collision with the learned `Edge` (`edge.go`) — this package already carries two
name-collisions (`Resolve`, `episode`) and will not add a third.

```go
// carries heat; does not store it.
type Conductor struct {
    From    BeliefID   // what excites this. empty for a raw injection.
    P, Conf float64    // the learned edge's factors. 1.0 / 1.0 for a raw injection.
    Triple  TripleKey  // what it excites.
}
```

The win in one line: **excitation is summed at READ, not at write.** v3's `bump` summed at
write and the factors were gone forever after. A conductor keeps them, so heat is derived,
provenance survives, and phase (§7) falls out — a raw injection has no `From`; a propagation
does. Nothing else needs to encode phase.

### 9.4 Heat is the SUM across distinct origins, unclamped

```text
Heat(triple) = Σ over DISTINCT origins of  ( From.Heat() × P × Conf )
```

**Sum, because the terms of a triple are independent evidence that it is relevant, and
independent evidence accumulates.** Max would discard convergence; a claim lit from its
subject *and* its predicate *and* a reliable precedent edge is more relevant than one lit
once, and sum is the only rule that says so.

**Sum is forced, not merely preferred.** Phase-budgeting (§7's anti-fiction rail — *"at most
X% of the frame may be the being's own projection"*) needs a denominator to take a fraction
of. Only sum provides one; max cannot express "fraction of total." The moment §7 is wanted,
sum is entailed.

**The caveat, imported from the observe wave** (`wave.go`: *rank is distinct-root count
first*): sum **across distinct origins, not across paths from one origin.** Two impulse terms
converging on a triple → sum. One upstream belief reaching it by two propagation paths → one
origin, counted once. The conductor set must dedup by `From`, not blindly accumulate.

**Unclamped.** v3's `min(1, ·)` existed to stop a term dominating by being said ten times —
but SOURCE's IDF gate already self-limits promiscuous terms *before* the sum, and ranking
needs order, not a ceiling. Clamping at 1 flattens the top of the distribution exactly where
the signal is strongest (most-corroborated triples become indistinguishable). So heat is
unbounded in storage; if anything is clamped, it is clamped at *display* into the present,
which is a WALK concern, not a field one.

### 9.5 Why "top-k" is the wrong shape, not a different cut

Selection is not "materialize the field, sort, slice." A heatmap that is the start of a
traversal is **never materialized**: the walk enters at the peaks and descends until the
budget is spent, and never enumerates the map to rank it. `top-k` and a budgeted walk are not
two sizes of the same operation — they are enumerate-then-cut vs. walk-from-hot-until-broke.

This picks a side in a fork the specs already named — [v4-port](v4-port.md) §5 Q2:

> Does `plan:articulate` become `traverse.Assemble`, or does `traverse` get replaced by the
> knapsack?

v3's `memory.go` knapsack is the enumerate-sort-cut model (`activeMemoryNodes` builds every
candidate, then sorts — [v4-port](v4-port.md) §2 flags it as the expensive thing in the
innermost paid loop). **§9 votes `traverse.Assemble` and against the knapsack:** let the map
heat unbounded, bound the *walk*.

### 9.6 What this section ANSWERS

| question | source | answer |
|---|---|---|
| Does the field select / threshold / top-k? | this session | **No. Never. The walk bounds; the map does not.** |
| Why does ε do three unrelated jobs? | §4 | Because they were three layers. Two are heatmap housekeeping/injection; one is selection and leaves the field (9.2). |
| Does `SpanActive` need a floor once heat is unbounded? | raised 9.2 | No — presence in the sparse decayed field *is* activeness. `SpanFloor` dissolves. |
| Should the field value be a scalar or carry its factors? | §5, §7 | A **Conductor** — factors preserved, heat derived at read (9.3). |
| sum, max, or clamped-sum when heat converges? | §1 (`bump`) | **Sum, unclamped, across distinct origins** (9.4). |
| Where does phase live? | §7 | In the conductor's shape — injection has no `From`, propagation does. No enum needed (9.3). |
| knapsack or `traverse.Assemble`? | [v4-port](v4-port.md) §5 Q2 | `traverse.Assemble` — the walk, not enumerate-sort-cut (9.5). |

### 9.7 What this section OPENS or leaves open

1. **Triple-grained or belief-grained heat?** A Conductor excites *a triple*
   (entity-path routing is triple-grained, and TRANSPORT crosses per-expect). But
   [v4-data-model](v4-data-model.md) §1 is emphatic that **the Belief is the heated unit** —
   *"a Belief is what heats; a Triple is what settles."* These disagree about the grain of
   excitation. Either the conductor rolls up to a belief before ranking, or the "heated unit"
   claim softens. **Unresolved, and it is the load-bearing one for the next part.**

2. **The combination rule now runs at read.** v3's saturating `bump` made the sum-vs-max
   choice silently at write. Deferring it to read is correct (9.3) but means the rule must be
   *made* somewhere explicit, and re-evaluated on every read. Cost is bounded only while
   `HopLimit` is small; a deeper walk turns this into a memoization problem (§9.5's walk is
   what would own that).

3. **Distinct-origin dedup needs an identity for "origin."** 9.4 requires it; `From BeliefID`
   is the obvious candidate, but a raw injection has no `From`, so injections need their own
   origin identity (the impulse term? the perceived act?) or they cannot be deduped against
   each other. Open.

4. **Unbounded heat removes every absolute cutoff.** Downstream code that compared to ε now
   has nothing fixed to compare to. The walk's budget replaces it — but every current
   absolute-threshold read (`RankedSpans() >= ε`, any `SpanFloor` consumer) has to be
   re-expressed as a walk-relative decision. Blast radius unmeasured.

5. **Decay is still eager; the conductor invites lazy.** If a conductor's injection carries
   its birth tick, `Heat()` could apply `λ^(now−born)` at read — reconciling the field's
   eager sweep with `Edge.readF`'s lazy timestamped decay ([time](time.md) contrast). Not
   decided here; flagged because the conductor is what makes it possible.

### 9.8 The one-paragraph statement

> The excitation field is a heatmap and the first layer of traversal. It heats where the
> impulse and precedent put it and decays everywhere else, with no threshold, no top-k, and
> no selection of any kind — sparsity is the only thing deletion does, and presence is the
> only gate. Its stored value is a **Conductor**, which carries the factors and derives heat
> as their **sum across distinct origins, unclamped**. Selection, budget, and every bound
> live in the **walk** that reads this map, never in the map itself.

---

## Appendix — analogies tried, and what each one caught

Two framings were tested against these primitives. **Both were wrong**, and both were
useful — each failed at a different point and the failure was the finding.

| framing | claim | why it fails | what it caught |
|---|---|---|---|
| reaction–diffusion | SPREAD is `D∇²a` | injection reads nothing; diffusion is gradient-driven | the **cap is a reaction term** hiding in the algebra |
| network flow | TRANSPORT conserves and moves mass | the source is never depleted — it **copies** | the **fan-out asymmetry** and the **hub risk** |

The pattern is worth noting: an analogy is useful here **precisely when it breaks**, because
the break is always at a place where the code made an undeclared choice.

### The reaction–diffusion reading

The four primitives partition almost exactly onto a reaction–diffusion system, which is
worth recording because it names two things we had not, and it says where our intuitions
should *not* transfer.

```text
∂a/∂t  =  R(a)  +  D∇²a  +  S
          reaction  diffusion  source
```

| primitive | R–D term |
|---|---|
| **DECAY** | negative reaction — local loss |
| **SOURCE** | source `S` — external forcing, point or index-distributed |
| **TRANSPORT** | *(see below — not in this equation)* |

**Note the earlier draft had this wrong.** It listed SPREAD under `D∇²a` (diffusion). But
diffusion is gradient-driven and reads the neighbour's current state; injection reads
nothing. **Injection is `S` however widely it fans out.** The field has no diffusion term at
all.

### Correction: it is not a PDE at all

Two drafts got this wrong in the same direction — first calling TRANSPORT *diffusion*, then
calling it *advection*. **Both are conserving transport terms, and nothing here conserves.**
Advection moves a quantity along a velocity field; our crossing copies without depleting.

The honest form is a **thresholded linear recurrence** — which is what spreading activation
always was:

```text
aᵢ  ←  min(1,  λ·aᵢ  +  Σⱼ wᵢⱼ·aⱼ  +  sᵢ )
               decay     coupling      injection

wᵢⱼ = P × confidence          the learned edge weight
```

No `D∇²a` — nothing spreads down a gradient. No `v·∇a` — nothing is transported. It is a
recurrent weighted sum with decay, injection, and a clamp, run for two steps. Discrete,
non-conservative, and structural.

**The coupling weight is `wᵢⱼ = P × confidence` — the learned reliability of an edge.** The
field induces activation preferentially along paths that have proven to carry consequence.
That is what TRANSPORT does and what injection cannot, and it is the one place the field is
not a bag of words.

### What the analogy names that we had not

**1 · The cap is a reaction term, and it is not one of the four primitives.**

`bump(a,h) = min(1, a+h)` looked like plumbing. In R–D terms it is a **carrying capacity** —
a crude logistic, `R(a) = r·a(1−a)` clipped to a hard ceiling. So the field has *local
dynamics of its own*, independent of anything striking, spreading, or conducting into it.

That is a fifth behaviour hiding in the algebra. Whether it should stay a hard clip or become
a smooth saturation is now a question that can be asked. **OPEN.**

**2 · The transport/source ratio is a real quantity, and it is the being's temperament.**

Advection–diffusion has the **Péclet number**: the ratio of directed transport to spreading.
Ours is TRANSPORT over SOURCE, and it reads more sharply than the earlier framing:

```text
how much attention is MINTED by what just arrived
versus
how much is INDUCED by what the being has learned
```

A being with no learned edges is **pure source**: every impulse floods the index and nothing
propagates. As edges accumulate, transport takes over.

**That ratio should rise as the being learns**, and it is directly measurable — transport
excitation over source excitation, per impulse.

It may be a better learning-curve instrument than anything currently in the thesis, because
it measures *how* the being is retrieving rather than only how well.

### Where the analogy misleads

**There is no diffusion at all.** This is the error the earlier draft made. Nothing here
flows down a gradient or relaxes toward a smooth state — the field is injected into, moved
along learned edges, and decayed. Equilibration intuitions do not transfer.

**Not conservative.** Diffusion conserves mass. Distributed injection conserves its own
amplitude (`÷n`), but injection adds and the cap destroys. So *total excitation is not a meaningful quantity* — worth knowing
before anyone tries to normalise over it.

**The ε gates are outside the physics.** They are computational truncation, not dynamics —
present so a broad posting costs nothing to skip. Nothing in R–D corresponds to them, and
they should not be reasoned about as though they were a physical threshold.

### The honest summary

Neither analogy describes this system. It is a **discrete, non-conservative, thresholded
recurrence over a learned graph** — spreading activation, run twice, with an index-driven
injection stage and uniform decay.

What the analogies gave, and it was worth the trip:

```text
from R–D           the cap is a reaction term, not plumbing
from flow          TRANSPORT copies rather than moves; the fan-out asymmetry; hubs
from both          the MINTED / INDUCED ratio as a learning instrument
```

Three findings from two wrong models. The lesson to carry: **reach for the analogy, take the
question it raises, and drop the frame.**

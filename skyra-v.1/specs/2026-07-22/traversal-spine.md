# The Traversal Spine — the Descent, Layer by Layer

**Date:** 2026-07-23
**Status: WORKING — in-session design.** Layers 0–2 and the flow rule are solidified. Two
questions are deliberately HELD (grain, cardinality — §6) to be answered after the spine is
fully walked. The descent stops at the span boundary, which is where segmentation,
prediction, and spans intersect — the next conversation (§7).

**Interface:** [Perturb and Observe](perturb-observe-traversal.md) owns the phase
contract. Frontier selection, allocation, affordability, and descent belong to
`Perturb`. `Observe` is the post-order concrete ascent.

**Reads on:** [interaction-primitives](interaction-primitives.md) §9 (the field is the
first layer of traversal — the Conductor, the heatmap, unbounded, no selection),
[v4-data-model](v4-data-model.md) (§1 the triple/belief grain, §4a span dissolves into
Observation, §7 the storage law), [episode-names-two-different-things](episode-names-two-different-things.md)
(the two-word collision and the grouping law), [time](time.md) (proper time — the walk's tick).

---

## 0. What the spine is

The excitation field is not a store the traversal consults; it **is** the traversal's early
layers (§9). This doc writes down the descent as it stands — how heat enters, how it flows
downward, and what is known at each layer boundary — up to the point where it stops.

> **The spine descends. The walk bounds. Nothing between the two selects.**

Every layer here is unbounded. Selection, budget, and cutoff belong to the *walk* that reads
this spine, never to the spine itself (§9.5). Read the whole doc with that invariant held.

---

## 1. The layers

```text
LAYER 0   ENTITIES     routing index. terms point inward.        NOT heat — the path in.
              │        (subject / predicate / object, per §9)
              ▼
LAYER 1   TRIPLES      HEATMAP. the Conductor carries heat.      we know HOW HOT each triple is.
              │        heat = Σ distinct origins (from·P·conf)   unbounded (§9.4).
              ▼
LAYER 2   SPANS        reverse index off each triple.            we know WHICH SPANS are active,
              │        map[Conductor] → span. same pattern.      and each span's heat.
              ▼        a span is active BECAUSE its triples are   unbounded.
              ·        hot — activation is derived, not bumped.
              ·
EPISODES  (implicit)   emergent — spans that co-light because     a reverse index over co-heat.
                       their triples co-heat. NEVER stored.       attention only, never settlement.
```

**Vocabulary note.** "Span" is the v3 word; in settled v4 it is the **Observation** (§4a).
Same concept — occurrence identity plus settlement scope. This doc says "span" to match the
working conversation; read it as Observation wherever the v4 model is authoritative.

---

## 2. The recurring pattern — Conductor + reverse index

Each layer relates to the one above by the **same two-part pattern**, which is why the spine
looks self-similar as it descends:

**The Conductor** (§9.3) — the stored value is never a bare scalar. It carries its factors
and derives heat on read:

```go
type Conductor struct {
    From    ID       // what excites this. empty for a raw injection.
    P, Conf float64  // the learned edge's factors. 1.0 / 1.0 for a raw injection.
    Target  ID       // what it excites (a triple at L1; a span at L2).
}
func (c Conductor) Heat() float64 { return c.From.Heat() * c.P * c.Conf }
```

**The reverse index** — each layer is keyed *by the layer above*. A triple holds the
conductors that lit it; a span is reached from the triples that name it. The descent is read
backwards from where heat landed:

```text
triple  ──its conductors──▶  the spans that instantiate it
span    ──co-activation──▶   the episode it implicitly belongs to
```

> **Episodes are the terminal case of the pattern: a reverse index with no stored layer at
> all.** They are computed as *"what is hot as a result of what is hot one layer up,"* read
> backward over co-heat. That is what keeps them attention-only, and it is the grouping law
> satisfied by construction — *grouping may confer attention, never settlement rights.*

---

## 3. Heat flow — two rules, and they differ by direction

The spine has two kinds of heat movement, and the distinction is load-bearing:

```text
WITHIN a layer   (TRANSPORT / propagation)   crossing = a · P · conf     CREATES. source not depleted.
                                             per-link, reliability-weighted (§9.4).

BETWEEN layers   (descent)                   share = heat(parent) / |children|   CONSERVES. divide equally,
                                             sum at the child.                    then sum.
```

So descent is the **distributed-SOURCE / IDF pattern**, applied downward:

```text
heat(span S) = Σ over triples T that name S  of   heat(T) / |spans(T)|
```

Two consequences, both intended:

- **A triple's heat divides equally among its spans** — it does not copy in full to each.
  Total heat is conserved on the way down, unlike within-layer propagation, which inflates.
- **A span accumulates** across every hot triple that names it. A span under many hot triples
  is hot; a span under one is faint. Convergence is rewarded, exactly as at L1 (§9.4, sum
  across distinct origins).

> **Within-layer propagation creates; between-layer descent conserves. Different arithmetic,
> and neither is the other's default.**

**Decided this session:** the descent divides **equally**. A weighted descent — dividing by
span recency or reliability instead of uniformly — is a knob we are choosing *not* to turn
yet. Recorded as a minor open (§6.3).

---

## 4. What is known at the span boundary

When the descent reaches L2, the spine has produced, with no selection applied:

```text
1. every heated triple, and how hot          (L1)
2. every active span, and how hot            (L2, derived by §3's descent)
3. for each span, which triples fed it        (the reverse index)
```

This is the state the walk will later bound. It is also the state segmentation reads — which
is why the descent stops here.

---

## 5. The one invariant that governs the whole spine

> **UNBOUNDED UNTIL THE WALK.** No layer thresholds, tops-k, or selects. Heat pools and
> decays; sparsity (deletion of ≈0 entries) is the only thing that ever removes an entry, and
> that is housekeeping, not selection (§9.2). Every real bound lives in the walk that reads
> this spine.

If a proposed layer mechanism cuts, ranks-and-slices, or gates on a fixed threshold, it does
not belong in the spine. It belongs in the walk.

---

## 6. Questions HELD — to answer after the full spine is walked

1. **Grain: are triples and beliefs one thing or two?** The conversation collapsed them
   (*"triples or beliefs, they are the same thing"*); [v4-data-model](v4-data-model.md) §1
   holds them deliberately apart (*"a Belief is what heats; a Triple is what settles"*). This
   is [interaction-primitives](interaction-primitives.md) §9.7 #1, still the load-bearing
   unresolved decision. **Held by agreement — resolve after the spine.**

2. **Cardinality: does a Conductor point to one span or many?** A claim recurs across
   occurrences, so a triple's reverse index may fan out. Whether each conductor ties to one
   span (each heat-origin to one occurrence) or many is deferred with #1. **Held.**

3. **(Minor) Equal vs weighted descent.** §3 divides a parent's heat uniformly. Weighting by
   span recency or reliability is available and not chosen. Low stakes; noted so it is a
   decision rather than an accident.

---

## 7. Where the spine stops — segmentation ∩ spans ∩ prediction

The descent halts at the span boundary because that is where three concerns meet, and none
of them is settled:

- **Spans** are the temporal/causal occurrence and the owner of settlement (§4a, the
  grouping law).
- **Segmentation** is what *produces* spans — the wall-gap cut (`DefaultGapK = 5.0`) that
  bundles PerceivedActs into one occurrence ([v4-data-model](v4-data-model.md) §4a). In the
  live v3 path the over-segment half runs and the merge half is dead
  ([episode-names-two-different-things](episode-names-two-different-things.md)).
- **Prediction** is owned per span: a span carries its open bets, and settlement eligibility
  is span-scoped.

The spine has just told us *which spans are active and how hot*. Segmentation decides *what a
span is in the first place*, and prediction rides on that boundary. So the intersection is:
**how does a descent that activates spans relate to the process that creates them, and to the
bets they own?**

§8 records where that conversation landed.

---

## 8. Segmentation — WORKED, 2026-07-23

Segmentation is not an independent mechanism bolted beside the field. **It falls out of the
descent**, with one external signal (the bet window) supplying the only thing the field
cannot feel. The starting instinct — *"an impulse joins the span of whatever it activates"* —
is the exact merge the docs forbid ([episode-names-two-different-things](episode-names-two-different-things.md):
*"it manufactures evidence for a bet the impulse never placed … merging on activation is
circular"*). The three rules below are what make a heatmap-driven segmentation legal anyway.

### 8.1 Membership is the heatmap; the temporal boundary is the bet window

The two segmentation axes split across two owners:

```text
MEMBERSHIP   which occurrences are the same KIND      → the heatmap (co-heat). decay + propagation
                                                         already carry recency AND relatedness.
TEMPORAL     where does one occurrence END            → the bet's `Within`. an occurrence ends
                                                         when its bets mature and settle.
```

**Causation is temporal by definition**, so the temporal boundary need not be detected
separately — the causal commitment *is* the extent. `Within` is wall-denominated
([time](time.md) §3d), so wall time is present, but carried **by the bet**, not by an
independent wall-gap cut. This is why the field can drive segmentation despite being blind to
idle: the wall signal rides in on the bet, not on a gap rule.

The build-then-idle-then-build case, which pure co-heat would wrongly merge (the field does
not decay while nothing happens), separates correctly:

```text
build#1 bets "status ok, within N"      extent = N (wall). matures, settles at its OWN address.
build#1 CLOSED before build#2 arrives    occurrence#1 is done; nothing of build#2 can grade it.
build#2 arrives                          fresh observation, fresh bet, fresh window.
```

The tricky sub-case — build#2 arrives *inside* build#1's open window at the same address — is
**not** contamination. It is [v4-data-model](v4-data-model.md) §3's *early change = surprise*:
the stability claim *"I don't expect to revise this"* was violated, which is a real finding.

### 8.2 Two regimes, by whether a triple carries an open bet

```text
triples WITH an open bet     temporal care REQUIRED    supplied by `Within` (§8.1)
triples WITHOUT a bet         attention only            group aggressively; worst case is a bad
                                                        retrieval, which decays
```

This is the grouping law holding — *grouping may confer attention, never settlement rights.*
Care is spent only where settlement can happen; everywhere else grouping is free to be broad.

### 8.3 Settlement touches only `OBSERVED-heat ∩ open-bet` — the phase prerequisite

The overlap of *active triples* and *triples with open bets* is the surface where settlement
is allowed — **but "active" is not sufficient**, because active heat has two phases and only
one may settle:

```text
OBSERVED heat     the world reported this triple          → may settle a bet at this address
PROPAGATED heat   hot by co-heat / spreading only          → attention ONLY. never settles.
```

Settling on propagated heat manufactures evidence a second way — mistaking excitation for
observation. The conductor already distinguishes them (§9.3 of
[interaction-primitives](interaction-primitives.md)): a raw observation has no `From`
(`P=Conf=1`); a propagation carries `From · P · Conf`. So the legal settlement surface is:

> **(triples with an open bet) ∩ (triples with an OBSERVED-phase conductor).**

**This makes §7 phase a prerequisite of segmentation, not an enhancement.** Recorded as the
first hard dependency.

### 8.4 The log is append-only — which is what keeps segmentation reversible

Observations are appended, never merged or mutated: an **append-only causation log**. Because
segmentation is then a *derived read* over that log, it is never irreversibly stamped at write
time — the precise escape from the doc's *"rebuilding it at write time makes it
irreversible."* A span is a view over the log, not a store (§1 of this doc, and §4a of the
data model).

### 8.5 The cut — at the observation boundary, priced against the context window

Segmentation ends where assembly begins. At the span boundary the spine holds, per
observation, a **token count** and a **magnitude**. The frame is cut here, and the threshold
is not arbitrary:

> **The bound is the model's context window** — the one physical constraint in the system.
> Everything upstream was kept unbounded (§5) precisely so this single real limit is the only
> cut.

**The cut is not a fork — magnitude and density are one operation.** An earlier draft framed
*magnitude-descent* and *density knapsack* as competing implementations. They are not.
`traverse.go`'s own contract already separates the two concerns this conflated:

> *"the cell owns the policy (which edges, weights, how broad / conserved); the kernel only
> scales."* — `traverse.go`

**Affordability IS local density — and it lives in `Perturb`, not the kernel.** With budget =
the context window, `Cost()` = an observation's token count, and `weight` = magnitude, the
allocation-with-affordability rule the assembly space runs —

```text
allocation   = (incoming share) attenuated by magnitude      magnitude attenuates the window
keep iff     allocation ≥ Cost()                             tokens must be covered
```

— reaches a leaf with `W × (weight product down its path)`, i.e. that observation's
**magnitude-share of the context window**, and admits it iff that share covers its tokens.
That is price-per-token, evaluated **locally during the walk, never enumerating**. Density
falls out of affordability; there is no separate knapsack pass and nothing to choose against.

**The kernel does zero math** — it executes the one ephemeral Space `Perturb`
returns ([kernel-alteration](kernel-alteration.md)). That Space privately carries
the surviving material, its grouping, and its behavior. Allocation and
affordability are one computation owned by the assembly Space. Density is not a
rival to the walk; it is part of how that Space builds its executable
continuation (v4-port §5 Q2 and §9.5 resolved this way, not by picking a side).

**What stays real (properties, not forks):**

1. **The walk is greedy-local, not globally optimal** — the ephemeral grouping
   Space owns member order and never backtracks. It must execute its members in
   magnitude order for the window to be spent hottest-first. The kernel never
   receives that ordered collection.
2. **Assembly-layer weights must conserve (sum to 1)** or the context window does not bound
   total tokens. Broadcast weights (1 each) are legal in the kernel and right for
   propagation, wrong at the cut. "Conserve vs broadcast" is the per-layer knob; assembly
   picks conserve.

### 8.6 What §8 settled and what it left open

| | |
|---|---|
| **settled** — membership axis | the heatmap (co-heat), 8.1 |
| **settled** — temporal axis | the bet's `Within`, wall-denominated, 8.1 |
| **settled** — grouping law honored | two regimes by open-bet presence, 8.2 |
| **settled** — reversibility | append-only causation log, 8.4 |
| **settled** — cut location + threshold | observation boundary, the context window, 8.5 |
| **settled** — cut shape | not a fork; affordability = local density; kernel pure, space owns weight policy, 8.5 |
| **open, hard** — phase prerequisite | settlement needs OBSERVED vs PROPAGATED (§7), 8.3 |
| **held** — grain, cardinality | still §6 |

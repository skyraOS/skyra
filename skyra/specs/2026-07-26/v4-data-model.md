# v4 Data Model — Vocabulary and Structs

> ## ⚠ SUPERSEDED IN PART — 2026-07-26
>
> Three reforms hit this doc and **the struct pass has not been made.** The vocabulary and the
> storage law stand; three specific shapes below are wrong. See
> [definitions](definitions.md) and [COHERENCE §1](COHERENCE.md).
>
> ```text
> R2  §8 `Settlement.Resolution Truth  // TRUE | FALSE`
>     WRONG. settlement is a signed DISPLACEMENT — ΣW over the window (definitions §3e).
>     the ARGUMENT survives: absence of a record still means "I did not look." what
>     changes is the field's type, plus a new third reading (ΣW = 0, observed and
>     orthogonal) that `Truth` cannot express.
>
> R3  §3 `Observation.Fidelity` fixed at write · §5a's affordance table
>     WRONG at a boundary. everything a foreign agent returns is interpretive, including
>     syscall-shaped values. Fidelity, Mode, and the affordance registry become LEARNED.
>     §5a's complaint that "the instruments a world exposes should BE the affordance list"
>     is no longer a defect report — it is the requirement.
>
> R4  §2 `Belief.ObservationID` (singular) · `Belief.Confidence` (scalar)
>     WRONG. a belief owns PLURAL weighted support edges, to observations OR to other
>     beliefs. §3's credit walk terminates one hop early for the same reason.
> ```
>
> **This is repair item #1** — every other pending rewrite reads off it.

**Date:** 2026-07-22
**Status:** §1–§3 **SETTLED**. §4–§5 are descriptive analysis of v3. §6–§7 are the storage
law. §8–§9 are the struct budget and what remains open.

**Scope:** spec vocabulary. **No existing v3 code is renamed by this doc.** Where a v3
identifier appears (`Triple`, `InterpretationClaim`, `ModeInterpret`) it means the v3
identifier, not the v4 word.

**Reads on:** [v3-mechanisms](v3-mechanisms.md) (what the code actually does),
[episode-names-two-different-things](episode-names-two-different-things.md),
[v4-port](v4-port.md), [workers](workers.md).

**Superseded within this doc:** earlier drafts named the mappings `Construal`,
`Interpretation`, and `Articulation`, and made `Belief` a single triple. All three
mapping-structs are gone — the mappings are `Observation` and `Expression`, and a Belief
is a set of triples. Sections below that still use the old words are marked.

---

## What this model is, and what it deliberately is not

> **This is a belief-learning system. It is one side of the equation, and the other side
> could not have been built first.**

Everything here learns about **what is the case and what follows from what**:

```text
Belief credit       was this reading worth holding
Edge P              does this state actually come back
Expression grade    did the model beat its own base rate
```

**Nothing here changes an act.** Acts appear from one model call, and the only thing that
varies across runs is *what the model was shown* — the beliefs, their heat, and precedent
with its base rates. A winning plan resurfaces because its Expression was retrieved, not
because anything learned that the act was good.

So there is no mechanism yet for:

| absent | what it would be |
|---|---|
| **act repair** | trying a variant when one fails |
| **act composition** | *"write config, then build"* becoming one reusable thing |
| **capability discovery** | the affordance list is static — and must stop being static the moment a device world exists |
| **promotion** | `act-service.md`'s *"when a generated adapter proves itself through use, it gets promoted to a proven block"* has no home here |

### Why the order was forced

**Act learning optimises against belief quality.** If what came back is noisy or laundered,
act selection reinforces whatever happened to be adjacent to a lucky outcome — the lucky-hit
failure at act scale, where it compounds instead of settling.

The belief side had to be trustworthy first. That is a dependency, not a preference.

### The consequence for the thesis

**Both current numbers measure the belief side.**

```text
passes-to-closure ↓    improves via RETRIEVAL     — belief-side
prediction error  ↓    improves via CALIBRATION   — belief-side
```

`PassesToClosure` counts planning iterations, and those drop when precedent surfaces a good
plan — retrieval getting better, not acts getting better. `GradeSum` is edge-over-base-rate,
which is prediction quality.

So when the act side arrives it needs its own number, and *"both curves down"* becomes an
incomplete claim rather than a wrong one.

**The seam is already here, unconsumed.** `Expectation.By` makes edges **act-keyed** rather
than expression-keyed, which v3 never had. `P(state | these acts)` is act-level knowledge —
it simply has nothing reading it yet.

---

## The two rules

> **Nothing in the code is wrong. The word is.**

1. **One representation.** One type for every structured claim — facts, predictions,
   readings. A claim that is not that type cannot be indexed, resolved, matched, or graded.
2. **One home per concern.** A thing stored twice, or split across two structs with no
   join, is a model lying about what it is.

---

## 1. The spine

```text
                          ┌────────────────────────────────┐
                          │            WORLD               │
                          └────────────────────────────────┘
                       acts happen │              ▲ acts done
                                   ▼              │
                        ┌───────────────┐  ┌───────────────┐
                        │  OBSERVATION  │  │  EXPRESSION   │
                        │  many acts    │  │  beliefs      │
                        │    → beliefs  │  │    → many acts│
                        └───────────────┘  └───────────────┘
                                   │              ▲
                                   ▼              │
                        ┌────────────────────────────────┐
                        │             BELIEF             │
                        │  many triples · a concrete whole│
                        └────────────────────────────────┘
```

> **Corrected 2026-07-24.** This box read *"the heated unit."* It is not. Heat lives on the
> **segmented** field — `SegmentedToken{ParentID, Position, Token}` — and every concrete type
> (Observation, Belief, Action) has a segmented projection that conducts it. A Belief is a
> **concrete bounded whole**; it is reached *because* its tokens are hot, not because it is
> itself the unit of heat. See [tokenizer §3](tokenizer.md#the-concretesegmented-cycle).

**Endpoints are things. The two mappings are named for the acts that make them.** They are
mirror images: Observation collapses many acts into belief; Expression expands belief into
many acts.

### The atom is the same on both ends

An **act is an event in the world** — whether the world caused it or the being did. Whose it
was is a field, not a type. That keeps LAW 2 intact (learning is still across the MINE /
WORLD boundary) while collapsing two vocabularies into one.

### Vocabulary mapping

| v4 word | v3 called it | Means |
|---|---|---|
| **Act** | `Act` (partly) | An event at a place, at a time. The atom on both ends. |
| **PerceivedAct** | *(nothing)* | An act that happened — the world's, another being's, or mine returning. Unfalsifiable. |
| **EngageAct** | `Act` | An act the being intends. Carries guards. |
| **Observation** | *(nothing — v3's `Observation` was a triple)* | **The mapping** many PerceivedActs → Beliefs, at a fidelity. |
| **Belief** | *(nothing)* | **Many triples · a concrete bounded whole.** Reached *because* its segmented tokens are hot — NOT itself the unit of heat (corrected 2026-07-24; see the box above and tokenizer §3). |
| **Triple** | `Triple`, minus `Mode`/`Confidence`/`Payload` | One subject-predicate-object claim. **The unit of settlement.** |
| **`Expression.Beliefs`** | *(nothing — the frame is discarded)* | The activation-weighted Beliefs one Expression was assembled from. A field, not a struct — a Belief is a Belief whether or not it is currently in one. |
| **Expression** | `Expression` | **The mapping** beliefs → EngageActs. One emission. |
| **Expectation** | `Expect` | A falsifiable claim, attributed to the acts responsible. |
| **Settlement** | *(nothing)* | The event that closes a claim. |
| **Boot** | *(nothing)* | Which process wrote a record. An audit stamp, not a scope — see §4c. |
| **Episode** | `Class` | The derived grouping of Observations of the same kind. |

**Two granularities, deliberately.** A Belief is what *heats*; a Triple is what *settles*.
One is the unit of retrieval, the other the unit of falsification, and they are not the same
size.

---

## 2. The structs — SETTLED

### Acts

```go
// The atom: an event at a place, at a time.
type Act struct {
    ID        ActID
    Primitive string // perceive | write | exec | engage
    Target    string
    Payload   string
}

// WORLD contract — it happened.
type PerceivedAct struct {
    Act
    Source string // whose act
    At     int
    Wall   int64
}

// DNA contract — it is intended.
type EngageAct struct {
    Act
    Requires []Triple // guards. only meaningful before it runs.
}
```

> **An EngageAct completing does NOT become a PerceivedAct claiming success.** It causes the
> world to *report*, and that report is a separate PerceivedAct. `write` returns
> `written: true` — never `exists: true`.

The gap between the two records **is LAW 1 made structural** rather than remembered. *"The
act completed, so mark it perceived"* is the natural shortcut, and it is exactly the fiction
the law forbids.

Embedding `Act` is safe only under the rule that no field may collide with an embedded
type's fields — that rule is what stops this becoming v3's `Observation.Confidence`.

### The spine

```go
// MAPPING: many PerceivedActs → Beliefs.
type Observation struct {
    ID       ObservationID
    From     []ActID // the perceived acts, bundled by the wall-time gap rule. backward ref.
    Fidelity float64 // the CHANNEL's quality. mechanical 1.0, inferential ~0.3.
}

// Many triples. A concrete bounded whole; its SEGMENTED tokens carry heat (corrected 2026-07-24).
type Belief struct {
    ID            BeliefID
    ObservationID ObservationID // what produced this. never empty.
    Triples       []Triple
    Confidence    float64 // this reading's own probability
}

// MAPPING: beliefs → EngageActs. One emission.
type Expression struct {
    ID          ExpressionID
    TriggeredBy ObservationID // what provoked this emission. NOT derivable.

    Beliefs map[BeliefID]float64 // activation AT ASSEMBLY TIME
    Acts    []EngageAct
    Framing string // the model's own stated stance
    Budget  int
    Spent   int
}

type Expectation struct {
    ID           ExpectationID
    ExpressionID ExpressionID // the ONLY link. carries the whole package.

    By       []ActID // the engage acts responsible. attribution AND cardinality.
    Predicts Triple
    Within   int // the STABILITY WINDOW. two-sided — see §3.
}
```

### Why `Observation` earns its existence

It is thin, so the test matters: **it is what makes two competing readings siblings.**
Without it you cannot distinguish *"two readings of the same evidence"* — preserved
ambiguity, which LAW 4 requires — from *"two readings that happen to cite identical acts."*

### Why `Beliefs` activation must be stamped at assembly

The excitation field decays every impulse (`λ=0.5`). By the time a bet settles, the
excitation that caused the emission is gone. If it is not written down when the Expression is built,
credit assignment is not approximate — **it is impossible.**

That is exactly what `traversal-implementation-plan.md` Phase 2 asks for and cannot have
today: *"distribute `credit = activation · g` to **all** edges in that cycle's frame."*

### Why `By` is not called `observations`

`By` is `[]ActID`, and those are **engage acts** — the being's own doing. Naming them
observations would collapse the one boundary the design rests on (LAW 2: MINE vs THE
WORLD'S), and would destroy the cardinality semantics:

```text
By = [a]        a failure grades THE ACT           — the door didn't open
By = [a, b]     a failure grades THE CAUSAL MODEL  — the door DID open, she DID hear,
                                                     and she still didn't walk through
```

An expectation is the **join of the two sides of the boundary**, which is what makes it
gradeable at all:

```text
ExpressionID ──▶ beliefs ──▶ observations ──▶ perceived acts   WHAT I SAW   (the world's)
By           ──▶ engage acts                                    WHAT I DID   (mine)
```

### Why `Expectation` names no Observation

Everything is already reachable, in three hops:

```text
Expectation
  ├─ By []ActID ──────────────▶ engage acts                      MINE
  │
  └─ ExpressionID ──▶ Expression
                        ├─ Acts                                  MINE
                        └─ Beliefs map[BeliefID]activation
                              └─▶ Belief
                                    ├─ Triples
                                    └─ ObservationID ──▶ Observation
                                                           └─ From []ActID
                                                                └─▶ perceived acts  WORLD'S
```

An Expression's beliefs come from **many** observations, across many runs, because the frame
includes retrieved precedent. Naming one would mean picking an arbitrary member of a set
that is already fully reachable — and creating a second path to the same fact, which is the
`DerivedFrom`-in-two-places defect.

**What is not derivable** is *what triggered this cycle* — you cannot recover it by filtering
the frame, because the triggering observation looks like any other belief once it is in
there. So it is stored once per emission, on the Expression. That is `BornFrom` finally
placed correctly and actually reachable.

---

## 3. Settlement, credit, and the two units — SETTLED

### `Within` is a claim, not a timer

v3 treats `within` as a deadline: at maturity, fire one targeted perceive and settle. v4
makes it a **stability claim**, which is falsifiable in both directions:

```text
DURING the window   "I do not expect to revise this."
                    perceives run across the window and return nothing → NO surprise.
                    an EARLY change is a surprise — the stability claim was wrong.

AT expiry           "by now it should have landed."
                    still nothing → the bet turns FALSE → the belief changes → SURPRISE.
```

So the perceives inside the window are not wasted polling. They are the evidence that the
stability claim held right up to the moment it did not.

**Two consequences.**

*Absence becomes a sustained finding rather than one sample.* A single perceive returning
nothing is weak evidence; a series of them across a declared window is what licenses
calling the bet false.

*The clock is no longer frozen when the being idles.* Perceives are world-touching, so they
tick (proper time — see [time](time.md)). doc.go's *"if the being idles for an hour and thinks nothing,
nothing has happened to it"* holds only with **no open bets** — with open bets, looking is
something happening.

**Cost, which is now real:** a window of N ticks polling every tick is N world calls per
open bet. This is the first place `ephemeral-world-workers.md`'s spend budget stops being
theoretical.

### No `Thought` struct — an absence is an ordinary observation

A failed perceive is not nothing. *"I looked at address X and found nothing"* is a
PerceivedAct → an Observation → a Belief (`X exists false`). A series of them across an
expired window is a genuine **belief change**, and that change diffs against the prediction
like any other.

So a being prompted by its own silence is already expressible: **an Expression whose
`TriggeredBy` names an absence-Observation rather than an external one.** The trigger
differs; the shape does not. No new struct.

**But it needs a rail.** A being that can prompt itself from its own silence can loop —
expression → nothing observed → expression → nothing observed — spending a model call per
turn and producing no external evidence. That is the fiction risk at cycle scale. v2 had
exactly this guard shape in `TryReframe`: one reframe per impulse, subsequent attempts
suppressed.

Recommended, and not yet decided:

```text
Expression.TriggeredBy    distinguishes external Observation from absence-Observation
a rail                    bounds consecutive self-triggered cycles
a metric                  self-triggered / total — a runaway is visible before it is costly
```

The metric matters to the thesis: self-triggered cycles spend model calls without external
prompting, so `PassesToClosure` would quietly inflate if the being started thinking to
itself — and that would read as *retrieval getting worse*.

### Individual settlement, collective attribution

```text
SETTLEMENT   INDIVIDUAL.  different addresses, different `within`, arriving at different
                          times. Forced by the two-bit rule — independent falsifiability.

ATTRIBUTION  COLLECTIVE.  one frame, one basis, credit to all of it. Forced by the runtime
                          not knowing which belief produced which bet.
```

A set cannot settle: one bet matures at tick 3, another at tick 20, and an observation lands
at one address answering exactly one claim. If the set were the settlement unit, cells 2 and
3 of the two-bit table become unreachable — and cell 3 is the lucky hit, the failure most
worth catching.

Conversely, attribution **must** be collective. The model may claim which evidence it used,
but authored attribution is never treated as earned credit — so there is no principled way
to assign a bet to a subset of the frame.

> **Every expectation from one emission shares the same evidential basis: the entire frame.**

### The credit walk

One settlement, one traversal, everything reached:

```text
settlement on ONE expectation  (grade g)
   │
   ├─▶ By []ActID              credit the engage acts responsible
   │
   └─▶ ExpressionID
          │
          ├─▶ the Expression's running grade  (mean over its settled expectations)
          │
          └─▶ BELIEFS — every Belief, weighted by activation-at-assembly
                 │
                 ├─▶ its component triples   partial: only those that settled
                 │
                 └─▶ its Observation ──▶ perceived acts
```

The frame is credited **once per settlement**, not once per emission — correct if the
Expression's grade is a running mean, wrong if it is a sum. An expression whose bets never
settle credits nothing, which is the honest outcome.

**Partial settlement of a Belief:** its credit is a running mean over settled triples only,
so a Belief accrues as evidence arrives rather than waiting on its slowest bet.

### Predictions and beliefs are the same KIND OF CLAIM, but not the same layer

> **Refined 2026-07-24.** This section originally read *"the same type at different stages."*
> The **content** claim stands; the **type-identity** claim does not. An Expectation lives on
> the Action layer, not as a phase of Belief — see
> [perturb-observe §6a](perturb-observe-traversal.md#expectations-are-a-phase-of-the-action-layer--decided-2026-07-24).

```text
belief      Belief + activation           HELD.   heat, no stake.
prediction  Triple + confidence + within   STAKED. a bet with a deadline.
```

**What is still true — the content.** Both are state claims. `doc.go`'s four causes place the
*state change* (final cause) between observation and action, and a prediction is that final
cause stated as a falsifiable address. *A prediction is a belief to come true* is accurate
about what it **is**.

**What changed — the layer.** Layer assignment encodes creation and traversal structure, not
ontology, and by creation a prediction is inseparable from its act:

> LAW 5 — *every act appears in at least one expect, and every expect carries a `within`, or
> the emission is rejected.*

An act without an expect is rejected; an expect without acts has no `By`. One emission, two
faces. So the Expectation is a **phase of the Action layer (L2)**, which also makes
`CreationPhase` map 1:1 onto layers (`BeliefPhase → L1`, `ActionPhase → L2`) and keeps the
same-kind visibility audit free of sub-type discrimination.

They convert: a prediction settles → the observation that settled it becomes a belief →
that belief enters the heated region → it is a belief. The loop is
beliefs → expression → predictions → observations → beliefs.

**Note the conversion runs through the world, not through a phase flip.** A prediction does
not *become* a belief by changing state; it is settled by an Observation, and that Observation
grounds a Belief. LAW 1 stays load-bearing in the cycle.

`Predicts` stays **singular**. A set of predictions is just N expectations, and splitting
them is strictly more diagnostic — three claims tell you *which* failed; one conjunctive
claim only tells you *something* did. A genuine conjunction (*"these three co-occur"*) is a
different claim and should be explicit, not fall out of a slice.

### Fidelity is one mechanism, not two

World-given facts are produced by an Observation too:

```text
os.Stat("/x") returns a stat struct     ← PerceivedAct. raw.
    │  Observation. mechanical. Fidelity 1.0
    ▼
build:/ missing app.conf                ← Belief

michael said "get me water"             ← PerceivedAct. raw.
    │  Observation. inferential. Fidelity 0.3
    ▼
being:michael thirsty                   ← Belief
```

v3 had two mechanisms — `Provenance` on the observation and `Mode` on the expectation,
multiplied at grade time. v4 asks one question: *what observed this, and how much is that
worth.* Answered by `Observation.Fidelity`.

**Cost, stated honestly:** an Observation record becomes mandatory for every structured
claim, including the thousands the world emits mechanically. Either those share an
observation per source-kind, or observations become the highest-volume record in the system.
Open storage question, not an objection.

---

## 4. What this fixes in the v3 model

Measured, not asserted.

**a) The Observation→Belief mapping did not exist.** `ProposeInterpretation` never wrote
an observation. Nothing linked an interpreted `Observation` to an `InterpretationClaim` —
no field, no lookup, no join key.

| | the claim | the evidence | the test | the score |
|---|---|---|---|---|
| `Observation{interpreted}` | ✓ | ✓ `[]string` | ✗ | ✗ |
| `InterpretationClaim` | ✗ (a bare label) | ✓ `[]ObservationID` | ✓ | ✓ |

**b) `Category string` broke the one-representation rule** — and that was the root cause of
(a). A bare string has no key to join on. `interpretationEdgeKey` just concatenated it.
Making the claim the one representation makes the missing link stop being missing.

**c) `InterpretationClaim` held four unrelated concerns** — the reading, the evidence, its
test, and an audit link it explicitly never learned from, plus inline score. Its own
comment admitted it: *"Context is an Expect for settlement mechanics, but it is indexed and
learned outside the action-consequence path."* The tell was `Context.Of`, never set — a
field belief "the acts responsible," structurally meaningless for a reading.

**d) `DerivedFrom` existed twice with two types** — `[]ObservationID` on the claim,
`[]string` on the observation. Collapses into `Interpretation.From`.

**e) The Belief→Expression mapping did not exist at all.** `Expression.BornFrom` recorded
the *impulse* (1–3 claims), not the *frame* (up to 2048 tokens of retrieved memory and
precedent), and was write-only — its only reads are in a test. The assembled node list was
rendered to a string and discarded; only `{Budget, Spent, MandatoryTokens, OverBudget}`
survived, overwritten per call.

---

## 4a. Span dissolves into Observation — SETTLED

Span is **gone**. Not deferred — every one of its five v3 jobs has acquired a home, so
there is nothing left for it to do.

| span's v3 job | now belongs to |
|---|---|
| **occurrence identity** — "this one impulse arriving" | **Observation** — it *is* one act of observing |
| **grading eligibility scope** — a bet settles only if its birth region is warm | **the Expression's Beliefs** — is any of them still hot |
| **activation unit** — heat bumped and decayed per span | **the segmented field** — heat is carried by `SegmentedToken`s, which name their concrete parent. *(corrected 2026-07-24; this cell read "Belief — the heated unit")* |
| **retrieval key** — postings map terms → spans | **the segmented field** — token/address indexes reach segmented material, whose `ParentID` names the concrete record |
| **frame membership proxy** | **`Expression.Beliefs`** — taken when that field was recorded |

### Segmentation survives the dissolution

The temporal rule does not disappear with the struct. The wall-time renewal-gap cut
(`DefaultGapK = 5.0`) stops producing *spans* and instead decides **which PerceivedActs get
bundled into one Observation**.

```text
v3   log ──cut by wall-time gaps──▶ spans ──merge by shared identity──▶ classes
v4   PerceivedActs ──cut by wall-time gaps──▶ Observations ──merge──▶ Episodes
```

Same signal, same rule, one fewer layer. Episodes group Observations rather than spans, and
the derivation still runs one way.

### Segmentation happens on the OBSERVE axis

*(Folded in 2026-07-24 from the retired `traversal-spine.md` §8, which filed this on the
descent. No decisions changed — only the axis, and the reasoning is unchanged.)*

Boundary is a property of the **concrete** form: *"a concrete form owns identity, boundary,
order, provenance"* ([tokenizer §3](tokenizer.md#the-concretesegmented-cycle)). `Perturb`
traverses segmented material that already carries a `ParentID` — it forms no boundaries.
Bundling PerceivedActs into one Observation **is** concrete realization, which happens during
`Observe`.

> **Perturb routes. Observe bounds.** Segmentation is a realization decision, not a routing
> one.

**The two axes split across two owners:**

```text
MEMBERSHIP   which occurrences are the same KIND    the heatmap — co-heat already carries
                                                     recency (decay) and relatedness (propagation)
TEMPORAL     where does one occurrence END           the bet's `Within` — an occurrence ends
                                                     when its expectations mature and settle
```

Causation is temporal by definition, so the temporal boundary is not detected separately —
the causal commitment *is* the extent. `Within` is wall-denominated ([time](time.md) §3d), so
wall time enters through the bet rather than through an independent gap rule. That matters
because the field cannot feel idle time pass: it decays per impulse, so a
build-then-idle-then-build sequence would wrongly merge under pure co-heat, and separates
correctly under `Within`.

A second occurrence arriving *inside* the first's open window is **not** contamination — it is
§3's *early change = surprise*: the stability claim *"I do not expect to revise this"* was
violated, which is a real finding.

**Two regimes, by whether a claim carries an open expectation:**

```text
WITH an open expectation     temporal care REQUIRED   supplied by `Within`
WITHOUT one                   attention only           group aggressively; the worst case is a
                                                       bad retrieval, which decays
```

This is the grouping law holding — *grouping may confer attention, never settlement rights.*
Care is spent only where settlement can happen.

**Settlement touches only `OBSERVED-heat ∩ open expectation`.** Active heat has two evidence
phases and only one may settle:

```text
OBSERVED heat     the world reported this            → may settle at this address
PROPAGATED heat   hot by co-heat / spreading only     → attention ONLY, never settles
```

Settling on propagated heat manufactures evidence by mistaking excitation for observation.
Note this is the **amplitude evidence phase**, not `CreationPhase` — `perturb-observe` §6a is
explicit that the two are different.

**The log is append-only, which is what keeps segmentation reversible.** Observations are
appended, never merged or mutated, so segmentation is a *derived read* over the log and is
never irreversibly stamped at write time. The Slice preserves this on the traversal side too:
*"a branch that later returns nothing was still genuinely perturbed."*

---

## 4b. World identity lives in the ADDRESS — SETTLED

v3 quarantined worlds with a run boundary: *"an old world's prediction cannot be adjudicated
by its replacement Realm."* That was the right intent through the wrong mechanism.

Put the world in the address instead. `triple.go` already reserved the namespace:

```go
var worldNamespaces = []string{"file:", "dir:", "proc:", "url:", "coord:",
                               "build:", "test:", "device:"}
```

```text
device:mac/file:/x     exists  true
device:phone/file:/x   exists  true      ← a DIFFERENT address. nothing to quarantine.
```

**Cross-world contamination stops being something a boundary prevents and becomes something
that cannot be expressed.** The addresses simply do not match.

And it is strictly better than the run check, because it permits the thing the design
actually wants: **a bet about the Mac, settled while standing on the phone.** The run
boundary forbade that. The namespace allows it, and cross-world work is the point.

### The cost this exposes

The observation log was bounded by the run wipe. Under §7 nothing is wiped, and a device
world is long-lived by nature — so the log grows without limit. `cur` is fine (bounded by
distinct addresses); `log` is not.

**This design has no compaction story.** It did not need one while runs were short and
disposable. It needs one the moment the Mac is a place the being keeps returning to.
**OPEN.**

---

## 4c. Run is deleted — SETTLED

`RunID` was v3's `EpisodeID`: the lifecycle boundary around one world. Every job it held has
moved.

| job | where it went |
|---|---|
| world lifecycle | the address namespace (§4b) |
| settlement quarantine | the address namespace — automatic, nothing to check |
| trial grouping | **`Cycle`, and always was** |
| ID prefix `run:1/observation:7` | a monotonic counter |

**The trial point is worth stating, because the experiment disguised it.**
`cmd/dnav3/main.go` calls `NewEpisode(freshWorld())` and then `Run(target, 6)` once per rep,
so a run and a trial look identical. They are not: `PassesToClosure` is appended per
`Run(target)` call — **per Cycle** — and `Metrics` survives `NewEpisode` entirely. The trial
unit was never the run.

Applying the same test used for structs — *does anything need to filter by this for
correctness?* — the answer for Run is now no.

### What replaces it: `Boot`, and it earns more than forensics

```go
BootID   // which process wrote this record
```

Nothing filters by it, nothing is bounded by it, no query starts with it. The stores outlive
the process; a restart changes nothing the being knows, believes, or is betting on.

**But it turns out to be the orphan detector**, which matters more than audit:

Under the windowed-`Within` model (§3), an armed bet cannot dangle — the settler sweeps it
across its window, and expiry with no confirmation turns it FALSE. So a bet that is armed,
past its window, and *still unsettled* did not go stale. **Nothing was running.**

```text
abandoned  = armed · window expired · no settlement · BootID != current
           = the bets a crash or restart orphaned
```

That is a defect signal, not a calibration signal — and it is invisible without a boot
stamp.

### What this costs

The abandonment metric in §8 was defined *"at run end,"* which was a free horizon needing no
threshold. That horizon is gone. Under the windowed model the metric changes meaning rather
than losing it — from *"bets the being drifted away from"* to *"bets a restart stranded"* —
but the calibration signal it was originally meant to give (`Within` values set wrong) now
has no home. **OPEN.**

---

## 5. What still needs a home — OPEN

**The interpretation test — NARROWED by §5a.** A Belief whose triples land on *measurable*
addresses is its own test; nothing extra is needed. Only an unmeasurable reading
(*"michael thirsty"* names no world state) needs a second discriminating claim at a
different address, and `Affordances.Measurable()` already computes that fork.
→ still open: what the field is called, and whether it is a distinct type or an
`Expectation` pointing at a `BeliefID` the way an ordinary one points at acts.

**The licence.** What downstream claim a reading justified. v3 had `Consequence
ConsequenceRef` inline, explicitly never learned from — audit metadata inside a learning
object. Probably an index, not a field.

**Invariant that survives whatever we pick:** a reading is graded through its **own**
address, never by whether the action it licensed succeeded. Otherwise a lucky action
launders a bad reading.

---

## 5a. Worked example — "get me water", and who takes the blame

The question this answers: *if the reading was "Michael is thirsty and wants water" and the
being brings the wrong kind of water, is that the Interpretation's fault or the
Expression's?*

### There is no assignment step

Nothing decides whose fault it was. **Each claim is tested at its own address, and an
arriving observation settles whatever addresses it matches.** One observation can settle
both, either, or neither. The independence is the mechanism — see
[the-two-bit-outcome](../2026-07-21/the-two-bit-outcome.md): *"deterministic prediction
success does not, by itself, validate the interpretation that produced it."*

Two bits, four cells, and the diagnosis is automatic:

| interpretation | prediction | diagnosis |
|---|---|---|
| ✓ | ✓ | understood it and executed it |
| ✓ | ✗ | understood it, failed to do it — **the act is the problem** |
| ✗ | ✓ | misread it and executed the misreading perfectly — **the reading is the problem** |
| ✗ | ✗ | misread it and botched it too |

The water case is cell 3, the same cell as the spec's Michael/traversal lucky hit.

### First: what an "address" is

Not a field. It is derived from the claim itself:

```go
func sp(t Triple) string { return t.Subject + "\x00" + t.Predicate }
```

A claim is three parts. The **address** is the first two; the object is the value predicted
at it.

```text
build:/ | status | ok
└─ address ─┘   └─ predicted value
```

This is why world state is `cur[sp(claim)] = observation` — *the latest value at each
address* — and why `Resolve` is a lookup plus a comparison:

```text
address present, object matches   → TRUE
address present, object differs   → FALSE
address absent                    → UNKNOWN     absence is not falsity
```

### When an interpretation needs a SEPARATE test

This narrows §5's homeless "test" considerably.

| the reading lands on | needs a separate test? |
|---|---|
| a **measurable** address — the world has an instrument | **No.** The Belief is its own test. Go perceive it. |
| an **unmeasurable** address — no instrument exists | **Yes.** It needs a second claim, at a different address, whose truth discriminates the reading. |

```text
build:/ missing app.conf     → the world reports it → its own test
being:michael thirsty true   → no instrument        → needs a discriminating claim
```

`Affordances.Measurable()` already computes exactly this fork. So the test is **not a
universal field** — it is populated only for readings the world cannot check directly.

### A concrete pair, and all four cells

```text
Observation      michael said "the build is broken — it's the config"

INTERPRETATION   reads "the config" as the missing build input
                 Belief:   build:/ | missing | app.conf
                            └ address ┘   └ predicted value      measurable → its own test

PREDICTION       act: write file:/app.conf
                 Belief:   build:/ | status  | ok
                            └ address ┘   └ predicted value
```

Same subject, different predicate, so `sp()` differs and the distinct-address rule holds.

| | interp `(build:/, missing)` | pred `(build:/, status)` | what happened |
|---|---|---|---|
| **✓✓** | `app.conf` | `ok` | read it right, fixed it |
| **✓✗** | `app.conf` | `failed` — now missing `app.lock` | read it right; *"writing app.conf fixes the build"* is a bad causal model |
| **✗✓** | `app.lock` — a different input was missing | `ok` | misread it — **a teammate fixed the build meanwhile.** The bet came true and validates nothing |
| **✗✗** | `app.lock` | `failed` | misread it, and the write did not help |

Each cell is two independent perceives at two addresses. Nobody adjudicates.

### Defect this example surfaced in v3

`world.go` emits `build:/ missing app.conf` as a real committed effect. But
`DefaultAffordances` registers `build: {status, output}` — **`missing` is not in the
list.** So `Measurable()` returns false for `(build, missing)`, the claim is stamped
`ModeInterpret`, and a bet on it settles at **0.3 weight instead of 1.0** despite the world
reporting it exactly and mechanically.

A measured fact graded as an inference. Not for fixing in v3, but it names a requirement
for v4: **the affordance table and the world's actual emissions are two lists that can
silently disagree.** The instruments a world exposes should *be* the affordance list, not
be mirrored by a second hand-maintained one.

### Why distinct addresses are what make the two bits TWO

v3 hard-rejects a shared address:

```go
if sp(proposal.Context) == sp(consequence.Triple) {
    return nil, fmt.Errorf("interpretation and consequence require distinct falsification addresses")
}
```
— `interpretation_claim.go:81`

If both claims lived at `(build:/, status)`, one observation would settle both identically
and **the table would collapse to its diagonal.** Cells 2 and 3 become structurally
unreachable — and cell 3 is the lucky hit, the failure mode you most need to see. The rule
is not bookkeeping; it is the precondition for the diagnosis existing at all.

**And this is the sharper statement of the confounded-test problem below.** The water
example's addresses were distinct as *strings* but not in their *causes*:

```text
reading   being:michael thirsty
test      being:michael drinks within 5   ← downstream of the act.
                                             a delivery error lands on the reading.
```

Distinctness must hold in the causal graph, not just in the key. An address that the act
can influence is not independent of the act, whatever it is called.

### Trace it with ambiguity COLLAPSED — the failure

```text
Observation      michael said "get me water"

Interpretation   being:michael requested water          ← TRUE. he did.
                 being:michael thirsty true             ← TRUE. he is.

Expectation      being:michael receives water:sparkling ← TRUE. he received it.

Correction       michael said "no, I wanted still"
```

**Every claim resolved TRUE and the outcome was still wrong.**

That is not a bug in the grading. It is the honest report: *the being made no claim at the
granularity where it erred.* Water type was never bet on. It was **assumed**, and

> **an unaddressed assumption is an ungradeable one.**

So the answer to "which of the two was wrong" is, mechanically, *neither* — and that
answer is itself the diagnosis. The fix is not better blame attribution. It is that the
ambiguity should have become a claim.

### Trace it with ambiguity PRESERVED — the design working

LAW 4: *"Ambiguous references are preserved as low-weight candidates, never hard-linked."*
Interpretation is many→one, and nothing says one observation yields one belief.

```text
Interpretation   being:michael requested water:still      conf 0.5
                 being:michael requested water:sparkling  conf 0.5

                 → the planner now SEES an unresolved fork.
                   `ask` is an act that produces an observation; the answer
                   becomes a Requires guarding the fetch.

Correction       michael said "no, I wanted still"
                 → settles requested:sparkling  FALSE
                 → settles requested:still      TRUE
```

Blame assigned. No arbiter, no judgment — the correction simply hit two addresses that
existed. **Collapsing the ambiguity is what denied the planner the information to ask**; by
the time it saw the belief there was nothing ambiguous left in it.

### The part that decides assignment: test design

An interpretation settles through its own discriminating claim, so **that claim's choice is
where blame is actually assigned.** A confounded test mis-assigns it, silently.

```text
BAD    being:michael drinks within 5
       fails if he is not thirsty — OR if I brought the wrong thing.
       conflates the reading with the act. punishes "thirsty" for a delivery error.

GOOD   a claim that is true iff he is thirsty, whatever I bring.
```

> A test that can fail for reasons other than its own claim being false will assign blame
> to the wrong mapping, and nothing downstream can detect that it did.

This is the same requirement [predicate-diagnosticity](../2026-07-21/predicate-diagnosticity.md)
formalises: select the predicate by held-out diagnosticity, finding the middle between
brittle exactness and unfalsifiable breadth. Blame assignment is a *consequence* of
diagnostic address choice, not a separate step.

### What still catches the coarse bet

`being:michael receives water` after a fetch has a base rate near 1.0, so
`log p_claim − log p̂_edge ≈ 0`. The scoring punishes an uninformative bet not by calling it
wrong but by granting **no edge** — the correct treatment for a bet that could not have been
wrong. LAW 5's *"a prediction that cannot be wrong is a vibe"* has teeth through the score,
not through a verdict.

### What this leaves genuinely open

If the interpretation authored **no** discriminating test, the correction settles nothing
and the being repeats the error. Own-address settlement is the clean channel and it is
sparse — this is a concrete case where the sparseness costs the lesson outright. It is the
strongest argument yet that credit flow-back (§1, open) is needed rather than optional.

---

## 6. Other open questions

1. **`Mode`.** v3's `ModeExact` / `ModeInterpret` answers *"does the world have an
   instrument for this claim?"* — a property of the expectation, stamped by the world's
   affordances. That is genuinely different from interpretation fidelity (*how good was
   the evidence?*), so both survive. But under §3 **every** Belief comes from an
   interpretation, so "interpret" can no longer be the contrast term. → `ModeMeasured` /
   `ModeInferred`?
2. **ID prefixes.** The episode spec deferred this to "the rewrite" — this is that
   rewrite. `run:1/span:2/observation:7` is consistent.
3. **Does `Observation.Payload` stay a `string`?** A syscall return is not text. `any`, a
   `[]byte`, or a typed union?
4. **Do mark-provenance records survive** as a third thing, or become an interpretation at
   fidelity 0?
5. **Does an Expression record the beliefs it EXCLUDED?** Near-misses under budget are the
   signal for calibrating the budget itself. Cheap to add, easy to regret leaving out.
6. ~~**Is credit stored or derived?**~~ **CLOSED by §8: derived.** The mapping holds the
   basis, the settlement log holds the grade, and storing the product would be a third copy
   that can go stale.
7. **Does `Attribution.Weight` destroy the cardinality signal?** v3 used `len(Of)` as
   semantics — one act grades the ACT, several grade the CAUSAL MODEL, and the confidence
   gradient between them measures how much of the outcome is out of the being's hands.
   Weights can encode that (single act = 1.0, joint = split) and `len(From)` still reads,
   but a crisp binary has become a continuum. Keep both, or keep the binary?
8. **What settles an Interpretation whose Belief is never bet on?** Credit needs a
   `Grade`, and a grade needs a settlement. A belief that is construed and then never
   used has no outcome to distribute — it stays `Settled: false` forever. That is probably
   correct (no evidence, no learning) but it means most interpretations never earn
   anything, and the store grows regardless.

---

## 7. Storage — LAW

**SETTLED.** Not a proposal. Everything below follows from it.

### The six stores

Durable, append-only, **one writer each**.

```text
perceivedActs  ActID          → PerceivedAct
observations   ObservationID  → Observation    (holds From []ActID)
beliefs       BeliefID      → Belief        (holds ObservationID)
expressions    ExpressionID   → Expression     (holds TriggeredBy, Beliefs, Acts)
expectations   ExpectationID  → Expectation    (holds ExpressionID)
settlements    SettlementID   → Settlement     (the event log — §8)
```

EngageActs are owned by their Expression (`Acts []EngageAct`) rather than living in a
store of their own, so `Expectation.By` resolves through an index — see below.

### The indexes

Derived, rebuildable, many-valued. **They never mutate a store.**

```text
perceivedAct →  observations that used it
observation  →  beliefs it produced
belief      →  expressions that believed it
expression   →  expectations it made
engageActID  →  the expression that authored it
claimID      →  its settlement, if any
```

### The three rules

**1 · References point backward, never forward.**
A record names what already existed when it was created. Write-once, at creation.
A forward pointer would require going back and mutating a durable record — a second
writer on a store that has exactly one door.

**2 · Forward traversal is an index, never a field.**
The reverse of a many→one mapping is one→many *and it grows without bound*: a single
Belief enters the beliefs of many Expressions across many runs, indefinitely. A
stored `next` would be an ever-growing list requiring writes to a record laid down months
earlier.

**3 · The entire index layer is a fold over the stores.**
Losable, rebuildable, never authoritative — the same discipline as world state being a fold
over the observation log, extended to the whole model. A corrupted index is a rebuild, not
data loss.

### What is NOT a store

| | why |
|---|---|
| **world state** | the fold `(subject,predicate) → latest Triple`. Derived. |
| **span** | a view over the log, cut by wall-time renewal gaps. |
| **episode** | a grouping of spans. Derived from spans, which are derived from the log. |
| **excitation field** | transient excitation, decayed per impulse. Outside this law — not durable, not expected to be. Its own contract is [interaction-primitives](interaction-primitives.md). |
| **the temporal chain** | `Sequence` + `Wall` are already on every record. Sort the log. Never stored as links. |

### What this makes impossible

Each of these was found in v3 (see [v3-mechanisms](v3-mechanisms.md)):

- **A mapping with no join.** `ProposeInterpretation` never wrote an observation and
  nothing linked the two. Under rule 1 the backward reference is mandatory at creation, so
  an unjoined record cannot be constructed.
- **A store with five writers.** v3's `ActivationField` has five, and no owner. One writer
  per store is the rule; the excitation field is outside it because it is transient — but
  it gets its own contract in [interaction-primitives](interaction-primitives.md) §6, which
  reduces it to two write doors.
- **A write-only field.** `BornFrom` and `Observation.Confidence` are both written and
  never read. A field that exists only as a link is replaced by the index, which is
  derived — so there is nothing to write and forget.

---

## 8. The settlement log — SETTLED

The seventh store. Keyed by the durable struct being settled; the settlement is the value.

```go
type Settlement struct {
    ID     SettlementID
    ClaimID string  // the Interpretation or Expectation being settled
    At     int      // tick
    Wall   int64

    Resolution Truth           // TRUE | FALSE. never UNKNOWN — unsettled means NO RECORD.
    Grade      float64         // (log p_claim − log p_prior) × fidelity
    ResolvedBy []ObservationID // the evidence that closed it
}
```

**`Resolution` is never UNKNOWN.** Unknown is not a value — it is the *absence of a
settlement record*. That is the same move as world state: absent means unknown, and
"I did not look" cannot be confused with "it did not happen."

### What this removes from the spine structs

`Grade`, `Settled`, `Resolution`, and every credit field are **absent from §2**. They would
be mutations of durable records, which §7 rule 1 forbids. They are answers to a lookup.

The `Settled bool` redundancy dissolves on its own: settled *is* "a settlement exists."

### Three consequences

**1 · An Expression never settles, because it has no address.**

Only *claims* settle, and a claim settles at an address. An Interpretation has one; an
Expectation has one. An Expression is not a claim — it is a record of what was in front of
the being.

So an Expression's grade is **derived**: a fold over the settlements of its own
Expectations. It accrues as those bets come due one at a time, rather than
arriving at a single moment. Storing it would have been wrong for that reason alone.

**2 · Credit is derived, not stored.** *(this closes open question 6)*

```text
Evidence.Credit            = Weight     × the Interpretation's settlement Grade
belief credit               = Activation × the Expression's derived Grade
Attribution.Credit         = Weight     × the Expectation's settlement Grade
```

Every operand is already durable — the basis on the mapping, the grade in the settlement
log. Storing the product would be a third copy that can go stale.

**3 · The Edge becomes derived too, and this is the big one.**

`Edge.P/N/F` is a fold of `Observe(outcome, tick)` calls. If every settlement is logged with
its outcome and tick, replaying the settlement log **rebuilds every edge exactly.** The
learned object is not a store; it is an index.

What that buys:

| | |
|---|---|
| **replay** | *what did this edge believe at tick N* — not just what it believes now |
| **audit** | every P is traceable to the settlements that produced it |
| **re-fold** | **changing the learning rule becomes a recomputation, not a data migration** |

The third is the one worth the trouble. The EMA is `α = max(floor, 1/(N+2))`, `confHalf`,
`λ` for staleness — hand-tuned constants. Under this law, retuning them is a re-fold over
the settlement log. Today it is unrecoverable: the old P is overwritten and the evidence
that produced it is gone.

### A metric the log makes free: abandonment

Today *"unsettled because I am still waiting"* and *"unsettled because I abandoned it"* are
indistinguishable — both read `Resolution == UNKNOWN`. The second is diagnostic and the
first is not.

```text
orphan rate = armed · window expired · no settlement · BootID != current
              ────────────────────────────────────────────────────────────
                                 expectations armed
```

**The boot stamp is the horizon**, so no threshold is needed. Under the windowed-`Within`
model (§3) an armed bet cannot dangle — the settler sweeps it across its window and expiry
turns it FALSE. So a bet still unsettled past its window did not go stale; **nothing was
running.** See §4c.

What it detects: bets a crash or restart stranded. That is a **defect** signal.

The *calibration* signal this metric originally aimed at — `Within` values set wrong, a
being pulled off-task before its predictions come due — no longer has a home now that runs
are gone (§4c). **OPEN.**

Note this is not a correctness cost — `Edge.Observe` ignores UNKNOWN, so an abandoned bet
contributes exactly zero. It is a storage cost and a blind spot, not a corruption.

### Open

1. **One settlement store, or one per claim type?** v3 keeps interpretation edges on a
   *"fully separate keyspace — they never share keys, settlement indexes, or updates with
   action consequence edges."* Separate stores continue that discipline; one store with a
   `Kind` field is simpler and relies on convention.
2. **Record the prior that was graded against?** `Grade` is computed against the edge's P
   *before* the fold. A complete ordered settlement log makes that recomputable, so storing
   it is redundant — but it makes each settlement self-contained and auditable without a
   replay. Cheap insurance, or a fourth copy.

---

## 9. Struct budget — what is actually NEW

**Struct contracts are immutable without the builder's explicit permission.** This section
exists so the cost is visible per-struct rather than buried in §2.

### Renames and reshapes — no new contract

| v4 | was |
|---|---|
| `Triple` | `Triple`, minus `Mode` / `Confidence` / `Payload` (all three dead on it — see §4) |
| `Expression` | `Expression`, reshaped: `Beliefs` + `TriggeredBy` in, `BornFrom` + `Score` out |
| `Expectation` | `Expect`, minus the settlement fields |
| `EngageAct` | v3's `Act`, unchanged in substance |

### Net-new — five

| # | Struct | Why | Earns it? |
|---|---|---|---|
| 1 | `Act` | the shared base — an event at a place, at a time | **Yes.** It is what makes perceived and engaged the same kind of thing, which is the symmetry the spine rests on. |
| 2 | `PerceivedAct` | the world's contract: it happened | **Yes.** `Source`, `At`, `Wall` are all non-derivable. |
| 3 | `Observation` | the mapping many acts → beliefs | **Yes** — and on a specific test: it is what makes two competing readings *siblings*. Without it, preserved ambiguity is indistinguishable from coincidence. |
| 4 | `Belief` | many triples; a concrete bounded whole | **Yes** — but not for the reason first given. *(corrected 2026-07-24: this read "the heated unit," and justified itself by separating retrieval from settlement.)* Heat is on the segmented field, so that is not what it earns. It earns its place as the unit of a **reading** — what makes two competing interpretations of the same evidence independently identifiable, which HOLD requires. |
| 5 | `Settlement` | the event that closes a claim | **Yes.** §7 rule 1 requires it; it is what makes the other five immutable. |

### Deleted from earlier drafts — four

`Construal`, `Interpretation` (as a struct), `Articulation`, `Evidence`,
`ArticulationMember`, `Attribution`. All were either 1:1 with what they mapped to — in
which case the mapping is a field, not a record — or `ID → weight` pairs wearing a struct.

### The test a struct must pass

**More than one field that cannot be derived.**

```text
PerceivedAct   source, at, wall, primitive, target, payload   → earns it
Observation    from, fidelity  (+ sibling identity)           → earns it
Belief        triples, confidence                            → earns it
Expression     triggeredBy, beliefs, acts, framing, budget    → earns it
Expectation    by, predicts, within                           → earns it
Settlement     resolution, grade, resolvedBy                  → earns it
```

Six structs plus `Act` as the embedded base, and `Triple` which already existed.

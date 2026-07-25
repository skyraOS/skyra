# v3 Mechanisms — What the Code Actually Does

**Date:** 2026-07-22
**Status:** DESCRIPTIVE. Read off `src/skyra/cognitive-agent/skyra-dna-v3` on branch `slice-based-dispatch`.
**No proposals in this doc.** Where something looks tangled it is reported, not redesigned.

**Purpose:** inventory the distinct pieces of machinery, and for each one record **what it
owns (writes)**, **what it reads**, and **the invariant it exists to enforce**. Ownership
boundaries are then read off the evidence rather than argued from a model.

---

## The mutable state, in full

Everything the system can write. Nine stores across four owners.

```text
ObservationGraph   log        []Observation           append-only
                   cur        sp(claim) → Observation the fold = world state
                   byID       ObservationID → Observation

SpanIndex          bySpan     SpanID → []Observation
                   spanTokens SpanID → int
                   spanOrder  []SpanID
                   entity     term → []SpanID         broad posting
                   address    sp(claim) → []SpanID     medium posting
                   triple     Key() → []SpanID         exact posting
                   seen       dedup guard

ActivationField    Entities / Triples / Spans / Edges / Nodes    key → float64

ExpressionGraph    exprs, edges, index, bets, spans, order,
                   settlementSpans, settlementSourceSpans,
                   interpretations, interpretationBets, interpretationEdges,
                   interpretationOrder, predictionFields, comparisons

Being (scalars)    tick, observationSeq, expressionSeq, interpretationSeq, spanSeq,
                   episodeID, currentSpan, currentBornFrom, Interpretation,
                   Metrics, LastMemorySelection
```

---

## The mechanisms

### 1 · The fold — world state as a derived index
**File:** `store.go` · **Owns:** `log`, `cur`, `byID`

`Write` appends to the log and folds the claim into `cur[sp(claim)] = observation`. World
state is not stored; it is *the latest value at each address*. `Resolve` is a lookup plus a
comparison.

```text
address present, object matches   → TRUE
address present, object differs   → FALSE
address absent                    → UNKNOWN
```

**Invariant:** absence is not falsity. Treating "I didn't check" as "it didn't happen"
would drag every edge toward zero.

**Ownership is clean.** `Obs.Write` and `Spans.Add` are each called from exactly one place —
`appendObservation`, `being.go:172-174`. One door.

**Exception it enforces:** a `mark` (self-authored) enters the log but **not** `cur`, so the
being's own notes can never satisfy a guard or overwrite a fact at the same address.

---

### 2 · The Edge — the learned object
**File:** `edge.go` · **Owns:** `P`, `N`, `F`, `Flast` on each edge

```text
E ──→ W        P = P(W | E)     counted, not asserted
```

`Source []string` — length > 1 **is** the conjunction. A joint expectation needs no special
machinery; it is an edge whose source is a set, carrying its own EMA.

Four operations:

| | what it does |
|---|---|
| `Observe(outcome, tick)` | folds one resolved outcome. `α = max(floor, 1/(N+2))` — one observation must not produce certainty (first update lands P at 0.75, not 1.0) |
| `Grade(claim, outcome, mode)` | `(log p_claim − log p_prior) × mode.Weight()` |
| `prior()` | `N == 0 → 0.5`. **No evidence, no opinion** |
| `RetrievalWeight(tick)` | `P × confidence × F` |

**Invariants:**
- `TruthUnknown` must not move the EMA — unresolved is not failed.
- `P = 0, N = 0` does not mean "never happens." Grading against it would pay +6.86 for
  beating a base rate that was never measured.
- `confidence = N/(N+half)` keeps **untested** apart from **wrong**.

**The same number does two jobs:** `P` is the prior that grades a claim *and* the weight
that ranks precedent. That is deliberate and stated in the header.

---

### 3 · Grading — arithmetic on the write path
**File:** `being.go:200` · **Owns:** `Expect.Resolution`, `ResolvedBy`, `Expression.Score`,
`Runs`, `Metrics`; drives `Edge.Observe`

Runs unconditionally inside `Observe()` (`being.go:184`, `:194`). No routing, no condition.
Six gates before a bet may settle:

```text
fidelity ≠ 0          a mark may never grade anything about the world
episode matches       an old world's bet cannot be adjudicated by its replacement
birth span active     cold birth regions are invisible to the grader
o.At ≥ ArmedAt        the evidence must post-date the bet
not already resolved  a doubled observation is a lie; a lost one is only a gap
compareExact ≠ UNKNOWN
```

Then: `g := edge.Grade(...)` **before** `edge.Observe(...)` — graded against the prior it
had, not the one it is about to have.

**Invariant:** the model cannot touch this path. There is no grading-time aligner to
misbehave.

---

### 4 · Arming and maturity — the bet manufactures its own evidence
**File:** `expression.go` · **Owns:** `Armed`, `ArmedAt`, `MaturesAt`

```text
World.Do returns → Complete(actID, tick) → for each expect whose every `Of` act is done:
                                            Arm(tick) → MaturesAt = tick + Within
```

**Invariant (LAW 5):** every act appears in at least one expect, and every expect carries a
`Within`. A prediction that cannot be wrong is a vibe; a bet that is never settled is the
same thing.

---

### 5 · The settler — making absence decidable
**File:** `being.go:311` · **Owns:** nothing; issues `World.Perceive` and writes through
mechanism 1

Every matured unresolved bet gets a targeted perceive. **No model call** — the expect
already names its own address.

**Invariant:** a targeted perceive comes back authoritative for exactly the claim bet on, so
*within it*, absence IS falsity. The general closed-world problem never has to be solved.

---

### 6 · The activation field — transient attention
**File:** `activation_field.go` · **Owns:** five `key → float64` maps

```text
BeginImpulse()   decay λ=0.5 across all five, drop below ε=0.05    ONCE per external impulse
bump             span, node, subject/predicate/object as entities, the claim key
retrieve         heat every past span whose posting contains those terms
propagate        ≤ HopLimit(2) hops: crossing = activation × edge.P × edge.confidence()
```

**Invariant:** decay happens exactly once per external impulse. Causal descendants and
planning queries bump the same field without advancing it.

**Note:** activation crosses edges weighted by *how reliable that edge proved*. Precedent
does not have to be searched for; it heats itself.

---

### 7 · Segmentation — spans and classes as derived indexes
**File:** `segment.go` · **Owns:** nothing. Pure functions over `[]Observation`

Two rules, two signals, deliberately not one judgment:

```text
BOUNDARY    temporal    where does one occurrence end?   renewal gaps in WALL time (K=5)
MEMBERSHIP  spatial     which are the same KIND?         shared identity above baseline (1.5)
```

Composition that survived live data: **over-segment on the finest rhythm, then merge.**

**Invariant:** rebuildable, never authoritative, no model anywhere. Conflating the two rules
was v2's edit/new trap — one LLM judgment doing both jobs at write time.

---

### 8 · Retrieval — budget-bounded selection
**File:** `memory.go` · **Owns:** `LastMemorySelection`

Candidates are gathered from every warm span (observations + expressions), plus exact
precedent labelled with its measured base rate `[worked 63% of 8 tries]`. Then:

```text
sort:   Mandatory first, then Density() = Activation / TokenCount, then Activation, then ID
admit:  every Mandatory node, even if that alone busts the budget
then:   optional nodes while Spent + cost ≤ 2048
```

**Invariant:** deterministic, and mandatory overrun is *recorded* (`OverBudget`) rather than
silently dropped.

---

### 9 · The observe wave — entity resolution by counted statistics
**File:** `wave.go` · **Owns:** commits `resolved-to` observations through mechanism 1

```text
ROOTS → HEAT → RANK → SELECT → COVERAGE → JOINT → COMMIT
```

Rank is **distinct-root count first, heat second** — a root's own candidates are
alternatives and can never corroborate each other; corroboration is *across* roots.

**Invariant:** zero inference calls in the entire file. A term with no history stands as its
own reading until a collapse founds its first edge.

**Status:** built and tested against fixtures, **not wired** — v3's skeleton is
target-driven and has no decomposer path yet.

---

### 10 · Affordances — the world stamps the stakes
**File:** `affordance.go` · **Owns:** `Expect.Mode` (via `Stamp`, at Place time)

```text
mode = exact      iff the world exposes a perceive for (namespace, predicate)
mode = interpret  otherwise
```

`Mode.Weight()` is the grade multiplier — exact ×1.0, interpret ×0.3.

**Invariant, and the reason it is not the model's to choose:** a model picking its own mode
would be picking its own stakes. Marking something measurable as `interpret` cuts losses to
0.3× — *a badly calibrated model would dampen the very signal this system exists to
measure.*

Also derives `Standing`: a target with no exact channel anywhere is a **compass, not a
bet** — it never closes and contributes zero learning signal.

---

### 11 · Structural closure — set membership, no threshold
**File:** `being.go:339` · **Owns:** nothing

A plan is closed when every `Requires` is either produced by some expect in the plan, or
already observed TRUE. No graded overlap, no threshold, no judge.

---

### 12 · Guards — parallelism and conditionals, derived
**File:** `expression.go` · **Owns:** nothing

An act fires when every `Requires` resolves TRUE against **observation**.

- no unmet requires → runs now; the rest topo-sorts itself
- if/else is two acts with opposite guards

**Invariant:** the model never writes control flow. It writes conditions under which things
are true, and the world decides which fires — so a conditional cannot be resolved by
fiction.

---

### 13 · The parser — failure isolation
**File:** `parse.go` · **Owns:** nothing

XML envelope (self-delimiting: a truncated emission still yields every complete element
before the cut) with line-oriented claims inside (per-line failure isolation).

**Invariant:** *an unparseable line is dropped and logged, never fatal.* Fewer predictions is
acceptable; no action is not. Deliberately not `encoding/xml` — an act payload is arbitrary
content and a strict parser would choke on exactly the emissions most worth saving.

---

### 14 · Interpretation claims — the second axis
**File:** `interpretation_claim.go` · **Owns:** `interpretations`, `interpretationBets`,
`interpretationEdges`

Admitted only when it already owns a falsification address distinct from the consequence it
licensed, with raw same-span evidence.

**Invariant:** settles through its **own** address. Action success never enters as backward
credit — the same counted `Edge` math on a fully separate keyspace.

---

### 15 · Prediction fields, calibration, fusion — Phase 4F
**Files:** `prediction_field.go`, `prediction_calibration.go`, `prediction_fusion.go`
**Status:** newest, partly in flight

- **Closed parser registry.** The model may choose a parser and its arguments; it **cannot
  add code to the grading path.**
- **Calibration windows** are blind: cases chosen outside predicate grading, support floors
  making sparse estimates explicitly provisional. Meeting them makes the arithmetic *ready*,
  not the predicate *trusted*.
- **Fusion** is a derived, reversible scalar. The dimensional comparison stays
  authoritative.

---

### 16 · The world — reports, never grades
**File:** `world.go` · **Owns:** the sandbox filesystem

```go
type Realm interface {
    Do(Act) Outcome
    Perceive(Triple) Outcome
}
```

**Invariant:** it has never seen the being's prediction and must not. Its independence is
the entire reason its evidence is worth anything.

Two details that are load-bearing rather than incidental:
- `write` reports `written: true`, **not** `exists: true`. Whether the file exists must be
  perceived. LAW 1 enforced by the world's own return shape.
- a parameterised predicate folds its argument into the predicate —
  `contains:"func main"` — so each parameterisation is its own addressable state.

---

## Ownership map — who writes what

| Store | Written by | Doors |
|---|---|---|
| `log`, `cur`, `byID` | `appendObservation` → `Obs.Write` | **1** |
| `SpanIndex` (7 maps) | `appendObservation` → `Spans.Add` | **1** |
| `ActivationField` (5 maps) | `BeginImpulse`, `bumpObservation`, `heatTriple`, `heatPosting`, `propagateActivation` | **5** |
| `Edge.P/N/F` | `grade`, `gradeInterpretations` | 2 |
| `exprs` / `index` / `bets` / `spans` | `Place` | 1 |
| `settlementSpans` | `RecordSettlement` | 1 |
| `Expect.Armed/MaturesAt` | `Complete` → `Arm` | 1 |
| `Expect.Resolution` | `grade` | 1 |
| `tick` | `express` only | 1 (2 statements) — **not proper time; see [time](time.md)** |
| `currentSpan`, `currentBornFrom` | `startExternalSpan`, `ObserveImpulse`, `startEpisode` | 3 |

---

## Where the boundaries are clean

**Observation writing.** One door, enforced. `Obs.Write` is called from exactly one line,
and it is the only thing that can move world state. LAW 1 is not a convention here — it is
a call graph with one edge.

**Learning.** Only `Edge.Observe` moves a P, and only a resolved outcome reaches it.

**The world.** Two methods, no back-channel, cannot see a prediction.

**Segmentation.** Pure functions, no state, rebuildable.

---

## Where they are not

**The activation field has five writers and no owner.** (Now given a contract in
[interaction-primitives](interaction-primitives.md), which also finds two of its five maps
are write-only.) `bumpObservation`, `heatTriple`,
`heatPosting`, `propagateActivation`, and `BeginImpulse` each touch a different subset of
five maps. To answer *"why did that memory surface?"* you have to hold all five in your head
at once. This is the least encapsulated thing in the package.

**The package sits on no substrate.** `cognitive-agent/skyra-dna-v3` imports nothing from
`spacetime` — no `Space`, no `Slice`, no `Time`, no `Impulse` — while every other agent in
the tree does. Most of what follows is downstream of that. See
[v4-port](v4-port.md) §4a.

**`Being` is a god object.** Eleven scalars of run state, five graph pointers, the model
seam, the token counter, and the three-loop engine, all on one struct. Every mechanism above
reaches through it.

**Interpretation is split across three files with no join.**
`Observation{Provenance: interpreted}` carries the claim and its evidence;
`InterpretationClaim` carries the category, its test, and its score; `wave.go` commits
`resolved-to` claims. **Nothing links them.** `ProposeInterpretation` never writes an
observation, and no field on either side points at the other.

**Retrieval reads state that grading also writes.** `Precedent` ranks by
`RetrievalWeight = P × confidence × F`, and `grade` mutates all three. Retrieval quality is
therefore a function of settlement order, and nothing marks that dependency.

**Frame membership is never recorded.** `activeMemoryNodes` builds the candidate set,
`SelectMemoryDetailed` chooses under budget, `present` renders it to a string — and the node
list is dropped. Only `{Budget, Spent, MandatoryTokens, OverBudget}` survives, overwritten
every call. `Expression.BornFrom` records the impulse instead, and nothing outside a test
reads it.

**Two lists that can silently disagree.** `world.go` emits `build:/ missing app.conf`;
`DefaultAffordances` registers `build: {status, output}`. `missing` is absent, so a claim at
that address is stamped `interpret` and settles at 0.3 despite being mechanically measured.

**`wave.go` is unwired.** Fully built, fully tested, no caller. Verified: none of
`BuildResolutionIndex`, `Roots`, `HeatClasses`, or `Resolve` has a non-test caller outside
the file. The largest single mechanism in the package that does not run.

**`Resolve` names two different things in one package** — the same collision as `episode`:

```text
ObservationGraph.Resolve(Triple) (Truth, Observation)   THE DIFF. world-state lookup.
Resolve(log, spans, classes, lens, roots, stats)        ENTITY RESOLUTION. the observe wave.
```

A grep for `Resolve(` returns both and they have nothing to do with each other. This is
what made the wiring status of `wave.go` hard to establish — the dead function looks alive
because its name is shared with a hot one.

# Flags

## Two learning signals

These may be related, but they are not the same measurement:

```text
assigned support − predicted support    prediction calibration
Pₜ₊₁(E) − Pₜ(E)                        action efficacy · Q(E,A)
```

The update rule must state which learner receives each signal.

## Strongest route

For multiple paths from one observation, “strongest” must mean the contribution with the
greatest absolute magnitude while preserving its sign. Signed `max` would incorrectly prefer
`−0.2` over `−0.8`.

Selecting on `|∏ s(e)|` while carrying the sign is also **discontinuous in sign**. With negative
edges present, a small change to one scalar can change which path wins and flip `c(o,E)` from
positive to negative. Ties are likewise unspecified. Both feed `G`, hence `P`, hence `Q`.

## Distinct is not necessarily independent

Two distinct observation records may describe the same underlying event. Repeated observations
could therefore inflate trust unless observational independence or shared provenance is defined.

## Truth normalizer

`P = σ(T)` cannot preserve direction because `T` is unsigned. Strong refutation produces high
`T` just as strong support does. A directional truth value must retain `G` or another signed
quantity.

## Confidence is self-reported

**New 07-30.** `confidence ∈ [0,1]` gives the system a way to say *I read this and could not
place it*, which is what Toy 5 needed. But it is authored by the same worker whose reading it
describes.

```text
CATCHES    known unknowns — unfamiliar vocabulary, unplaceable observations
BLIND TO   confident misreadings — Toy 1, where the interpreter never hesitates
```

Nothing grades confidence. It is the one number in the system with no blind assessor, which is
tolerable only because it routes rather than settles. If it ever starts weighting anything, that
exemption stops being safe.

## Goal degeneration under the merged worker

**New 07-30.** The expectation worker sets targets *and* selects actions, so it must see `Q`.
The earlier separation — repertoire yes, `Q` no — is no longer available.

Targets will tend to bend toward existing competence. Closure rates rise, confidence rises,
ambition contracts. Every local measurement improves while the system gets smaller, which is why
this is hard to catch by inspection.

```text
average P at creation time     rising  → targets are getting easier
novelty rate                   falling → fewer genuinely new expectations
```

Log both from the first run. Neither is derivable after the fact.

## Motivation requires re-scoring

**New 07-30.** An observation that triggers a creation scored ≈0 against a field that lacked its
target. If observations are scored only on arrival, it stays at ≈0 forever and never acquires
the edge that would make motivation derivable.

The expectation worker must re-present triggering observations for scoring after it creates or
refines. The substrate permits it — the evidence set is append-only — but it is an explicit step,
and without it the mechanism silently does not work.

## The basis moves, and now it moves routinely

**Promoted 07-30.** `truth.md` flags that `ΔP` is undefined over a changing node set until
`BORN / DIED / MOVED` are separated. That was theoretical. Goal refinement makes it routine:
adding children beneath an expectation drops its `P` sharply **with nothing having failed**.

`dynamics.md` reads `‖ΔP‖ large → surprise → widen attention`. So the controller will read the
being getting smarter as the world breaking. `ΔP` must distinguish *the basis changed* from
*the world moved* before dynamics can consume it at all.

## Slice concurrency is live again

**Updated 07-30.** Sequential actions were expected to defuse kernel §13 Q4. Sibling
expectations run concurrently, so two action sequences can be in flight, both writing to a Slice
whose lock is held per wave.

Narrower than before — credit attribution survives, since grading happens per leaf action and
each expectation inherits along its own path. It is a transcript-integrity problem, not a credit
problem.

## “Nothing appends” contradicts ΔP

`README.md` — *“an expectation’s record closes as soon as both halves exist — its evidence at the
time, and its action. Nothing appends to it afterward.”*

Read literally, `O(E)` is frozen at creation and so is every scalar on every path beneath it.
`T(E)` and `P(E)` are then **constant for all time**, and the set has no second hand:

```text
truth      the signal is the CHANGE in truth over time
Q(E,A)     E[ Pₜ₊₁(E) − Pₜ(E) | Eₜ , A ]
dynamics   every controller input is a delta
Toy 1      later cross-form evidence must revise operational weight
Toy 5      o* must eventually bear on an already-trusted E
```

The observation worker’s own description in the same document — *“it sets the edge… it scores what
exists and never creates”* — is appending: scoring what exists means attaching support to
expectations conceived in earlier waves.

**Resolved 07-30** — the README now states the narrowed rule directly. Retained here because the
reasoning is the argument for it.

The immutability instinct is aimed at the wrong edge. What the passage protects against is stated
two lines earlier — *stops an old expectation from being rewritten every time something new comes
to rest on it* — and that concerns **incoming** edges, which are already a query rather than a
field. Evidence *for* `E` is the opposite relation and must stay open.

Candidate split, recorded not adopted:

```text
ACTION EDGE       write-once. closes the record.
SUPPORT VALUE     immutable once written.
EVIDENCE SET      append-only.
```

`P` then moves by accretion and never by rewrite.

## Two arrow conventions, and a self-contradicting citation

`README.md` points the expectation **at** its evidence; `trust.md` points observations **at** the
expectation:

```text
README    EVIDENCE   ──▶ observations / other expectations
trust     Observation ── support ──▶ Expectation
```

`trust.md` then cites the ownership rule against its own diagram: *“These edges are owned by the
expectation they ground — an expectation owns every edge it points at, and never one that points
at it.”* In that diagram the edge points at `A` and the sentence says `A` owns it, which is what
the quoted rule forbids.

There appear to be two directions in play — a **citation** direction (`E ──▶ what it rests on`,
which ownership is about) and a **flow** direction (`O ──▶ E`, which the arithmetic is about) —
merged into one arrow. One convention must be chosen for the diagrams and the other named
explicitly, because ownership is what decides the append question above.

## Closure is capped by depth

Depth attenuation is correct for a ranking signal (`trust.md` — *“every hop shrinks the
contribution… needs no decay constant”*). It is wrong for closure.

`truth.md` says `P` measures how much of the change has happened, `P → +1` being closed. But in
the worked example `B` sits behind a `0.7` edge, so `P(B)` cannot exceed `0.7` however completely
the change closes. A state change that has fully occurred reads as partially closed because of
where it sits in the graph.

Either `P` is not closure — it is normalized grounding, which is what the arithmetic computes —
or closure needs a derivation that does not inherit path decay. This is prior to the normalizer
question: it concerns what `P` *means*, and `Q` is defined on `ΔP`.

## Coherence is undefined at T = 0

`coherence = |G| / T` is `0/0` for an expectation with no grounding. Dynamics keys on coherence
and Toy 5 turns on it. An unexamined expectation needs either a defined value or an explicit
undefined state that the controller is required to handle.

## The unspecified worker is the keystone

The expectation worker sits under *Not yet in this set*, but three toys resolve through it:

```text
Toy 1   cross-form revision of an established reading
Toy 2   what licenses counting a new ground
Toy 5   novelty as the ingress path for contradictory o*
```

It is not one gap among several. It is the missing piece the toys depend on.

## Toy 3 — a candidate type boundary

Recorded not adopted. The open question asks whether one-to-one governs an authored operator, an
execution occurrence, or physical causation. A three-way split answers it:

```text
OPERATOR     authored, reusable, carries Q
OCCURRENCE   one execution, immutable, referenceable by many
E → OPERATOR the edge “exactly one” governs
```

`Elight` and `Efan` each own one edge to the same operator; one occurrence is recorded; each
measures its own `ΔP`. Credit splitting does not return, because `Q` was always per `(E,A)` and
each expectation grades only its own movement.

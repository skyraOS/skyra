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

# Toy Test #6 — Ambiguity-Signal Discrimination

**Date:** 2026-07-21
**Status:** RUN (paper walk) · the regression spec for the first build step named in
[episodes-over-the-field](episodes-over-the-field.md) — *revive joint-precedent on the
live path.*
**Tests:** that the three-condition ambiguity signal **fires on ambiguity and stays
silent on belonging** — specifically that each belonging case is silenced by a *different*
one of the three conditions, proving all three are necessary.

---

## The signal under test

```text
v1 (boolean, superseded):  same source ∧ empty co-occurrence          ∧ no dominant heat
v2 (rate, current):        same source ∧ co-occurrence rate < θ        ∧ no dominant heat
```

Grounded in the real substrate (`store.go`, `activation_field.go`):

- **same source** — group activated completions by the **act-signature prefix** of their
  `edgeKey` (`sorted(primitive→target) \x00 target`). Two completions are competing iff
  they share that prefix, i.e. the same acts producing different world states. Already
  computed at derive-target time; today it is discarded (grouping is by target suffix).
- **co-occurrence rate** — `|Spans.Triple(A) ∩ Spans.Triple(B)| / |Spans.Triple(A) ∪
  Spans.Triple(B)|` — the Jaccard overlap of the two completions' span postings,
  conditioned on same-source (so the denominator is effectively the source's firing
  history). `rate → 1` = witnessed together = belonging; `rate → 0` = witnessed apart =
  alternatives. The original **boolean** was the degenerate case `rate > 0`, which a
  *single* accidental co-occurrence trips — see the re-run below. `θ` is the valley-depth
  threshold (illustrative `θ = 0.2` here).
- **no dominant heat** — the top completion's activation heat is not clearly above the
  second. Heat is `Edges[ek]` from `propagateActivation`, i.e.
  `source.Activation · ed.P · ed.confidence()` — note it does **not** include recency `F`,
  so recency cannot leak into the balance.

The **v1 boolean walk is kept below** because it is what motivated the change; the **v2
re-run** follows it.

## Fixture — one impulse heats four sources at once

Constants: `λ=0.5, ε=0.05, spanFloor=0.05`. Every source region heats to
`source.Activation ≈ 0.67` (as in Toy #5). `confidence(N=10)=0.667`, `confidence(N=20)=0.8`.
Every edge clears the derive gate (`N>0 ∧ P≥0.5`).

```text
GROUP  source sig  completions   edge P,N     heat  ≈           Spans.Triple postings
────────────────────────────────────────────────────────────────────────────────────
AMB    S_amb       A             0.9, 10      0.67·0.9·0.667 ≈ 0.40   {p1, p2}
                   B             0.9, 10      0.40                    {p3, p4}
CONJ   S_conj      C             0.9, 10      0.40                    {q1, q2, q3}
                   D             0.9, 10      0.40                    {q1, q2, q3}
CONF   S_conf      E             0.95, 20     0.67·0.95·0.8 ≈ 0.51    {r1 … r19}
                   F             0.55, 4      0.67·0.55·0.44 ≈ 0.16   {r5}
CTX    S_x         G             0.9, 10      0.40                    {x1, x2}
       S_y         H             0.9, 10      0.40                    {y1, y2}
```

The foils are built to be *individually* deceptive:

- **CONJ** has perfectly balanced heat (0.40 = 0.40) — it would fool a heat-only detector.
- **CTX** has balanced heat **and** empty co-occurrence — it would fool a
  "balanced + never-together" detector that forgot the same-source clause.
- **CONF** is same-source with (assume) empty co-occurrence — it would fool a
  "same-source + apart" detector that forgot to check dominance.

## The walk

```text
GROUP  same source?   empty co-occ?              no dominant heat?     → SIGNAL
──────────────────────────────────────────────────────────────────────────────
AMB    yes (S_amb)    {p1,p2}∩{p3,p4} = ∅  yes   0.40 vs 0.40  yes     ►► FIRES
CONJ   yes (S_conj)   {q1,q2,q3}∩{…} ≠ ∅   NO    0.40 vs 0.40  yes     silent
CONF   yes (S_conf)   (irrelevant)              0.51 vs 0.16  NO       silent
CTX    NO (S_x ≠ S_y) ∅                    yes   0.40 vs 0.40  yes     silent
```

**Each foil is defeated by exactly one condition, and a different one each time:**

| foil | balanced heat | apart | verdict | the single clause that saves it |
|------|:---:|:---:|---|---|
| CONJ | ✓ | ✗ | silent | **co-occurrence** |
| CONF | — | ✓ | silent | **dominance** |
| CTX  | ✓ | ✓ | silent | **same-source** |
| AMB  | ✓ | ✓ | **fire** | — (all three hold) |

Drop any one condition and its foil leaks through: without co-occurrence, CONJ fires
(a conjunction misread as a choice); without dominance, CONF fires (a settled call
reopened); without same-source, CTX fires (two independently-wanted targets misread as
rivals). So all three are load-bearing. **PASS.**

## Metamorphic check

Swap the opaque IDs `A ↔ B` in AMB. The source prefix `S_amb` is unchanged, the two
targets swap, heat stays `0.40 = 0.40`, the intersection stays `∅`. The signal fires
identically — because it is raised **about the source's slot** ("this situation has two
competing completions and no reason to pick"), not about a named winner. Nothing
downstream is handed an ordered head. Pre-evidence result is invariant under the rename.
**PASS.**

## Re-run with co-occurrence as a rate (v2)

The v1 walk passed on the four original groups because their overlaps were extreme
(`∅` or `1.0`) — the boolean handled those. The rate earns its keep in the **middle band**,
and specifically on the case the boolean got **wrong**: a genuine fork that co-occurred
once by accident. So the re-run keeps the originals and adds the cases that exercise the
rate. `θ = 0.2`.

```text
GROUP       A postings          B postings              |∩|  |∪|   rate    → SIGNAL (v2)
──────────────────────────────────────────────────────────────────────────────────────
AMB         {p1,p2}             {p3,p4}                   0    4    0.00    ►► FIRES
AMB-NOISE   {p1,p2,p3,x}        {p4,p5,p6,x}              1    7    0.14    ►► FIRES   ← v1 WRONGLY silenced
CONJ        {q1,q2,q3}          {q1,q2,q3}                3    3    1.00       silent
CONJ-WEAK   {q1..q8}            {q1..q8,q9,q10}           8   10    0.80       silent
AMB-EDGE    {s1,s2,s3}          {s3,s4}                   1    4    0.25       silent (at θ=0.2)
CONF        {r1..r19}           {r5}          — dominance decides (0.51 vs 0.16) — silent
CTX         different source    —             — same-source fails —            silent
```

What the rate changes, precisely:

- **AMB-NOISE is rescued.** Two branches that are alternatives 6 times and happened to
  share span `x` once: boolean `∩ = {x} ≠ ∅` → *silenced a real fork forever*. Rate
  `1/7 = 0.14 < θ` → **fires**, correctly. This is the entire point of the change.
- **CONJ-WEAK confirms the other side holds.** A conjunction where `D` sometimes appears
  without `C` (co-occur 8, D-alone 2): rate `0.80 ≥ θ` → still **silent**. The rate does
  not start leaking conjunctions just because they are imperfect.
- **AMB-EDGE shows the threshold now *decides the middle*.** Overlap `0.25` sits just above
  `θ = 0.2` → silent; raise `θ` to `0.33` and it fires. This is not a defect — it is the
  **valley-depth parameter made concrete**. The boolean pretended this band did not exist;
  the rate exposes it as a tunable dial, which is the honest state.

And a robustness note the AND-structure buys for free: the rate's one blind spot — a rare
add-on completion that always appears alongside a dominant one (low Jaccard, but not a real
fork) — is caught by the **dominance** clause anyway (the dominant branch's heat swamps the
add-on's). A misread on co-occurrence is backstopped by dominance, because the three
clauses are conjoined, not voted.

**v2 verdict: PASS, and strictly better than v1** — same discrimination on the clean cases,
plus it now fires on `AMB-NOISE` (which v1 failed) without silencing any belonging case.

## Holes found (the honest part)

1. **The threshold placement is the real unknown (was: the binary is brittle — now
   fixed).** v2 replaced the brittle boolean with a Jaccard rate, so a single accidental
   co-occurrence no longer silences a fork. What remains is *where* the co-occurrence
   cutover (`ConjunctionOverlap`) sits — the middle band (`AMB-EDGE` at `0.25`) is decided
   by it. It shares a *principle* with differentiation's valley depth, the dominance
   margin, and the promotion threshold — "how much separation before it's a real
   distinction" — but **not a value**: these are statistically different quantities (and
   promotion is a count, not a `[0,1]` fraction), so they are **independent named fields**
   in `DifferentiationConfig`, each tuned by its own test, not one fused `θ`. Expose as
   config, do not hard-code.
   **RESOLVED (was flagged here as a secondary refinement):** the co-occurrence rate must
   be both **source-conditioned** and **polarity-separated**. The index is
   `settlementSpans[SIG\x00target][Truth]`, populated at settlement:
   - *source-conditioned* — keying by the full `edgeKey` means an unrelated source `T`
     emitting `A`/`B` together lands under a different key and cannot contaminate `S`'s
     intersection.
   - *polarity-separated (blocking)* — `grade()` settles every open prediction at the
     observed address, so when `B` occurs an open `A` prediction settles **false** there;
     Jaccard must use **`TruthTrue` spans only** (`settlementSpans[S→a][TruthTrue]` vs
     `[S→b][TruthTrue]`), else competing alternatives always share spans and every fork is
     misread as a conjunction.

   See [interpretation-implementation-phases](interpretation-implementation-phases.md)
   Phase 1 and [toy-6b-branch-partition.md](toy-6b-branch-partition.md); the cross-source
   contamination and polarity cases are required 1a fixtures.
2. **Source granularity is inherited from `edgeKey`.** "Same source" is exactly as
   fine/coarse as the act-signature `primitive→target`. If two runs that are *really* the
   same situation carry different signatures, their completions land in different groups
   and a real ambiguity is missed; if two genuinely different situations collide on a
   signature, a false ambiguity is raised. This is not a new risk — it is the existing
   edge-identity granularity — but the ambiguity grouping now depends on it.
3. **The dominance line is undefined.** The walk only shows that *clear* balance fires and
   *clear* dominance stays silent (0.51 vs 0.16). Where the cutover sits is the
   valley-depth parameter, unresolved here by construction.

## Toy #6B — multi-way groups (moved to its own gate doc)

Toy #6 proved only the **two-completion** case. The multi-way gate — where a flat
`confident | conjunction | ambiguous` enum breaks down (rule overlap for 3+ completions;
mixed conjunctive-within/ambiguous-between structures) — is walked in its own document:
**[toy-6b-branch-partition.md](toy-6b-branch-partition.md)**.

The resolution there supersedes an earlier draft of this section on two points that turned
out to be blocking:

- **Partition:** *not* complete-linkage greedy (which is order-dependent and breaks the
  ID-swap invariant) but **connected components + clique test** — a clique component is a
  branch, a non-clique (bridge) component is left **whole-component unresolved** for Phase 3.
- **Co-occurrence uses positive settlement spans only** (see Hole #1) — otherwise an
  alternative that settles *false* where its rival occurs would count as co-occurrence and
  invert the classification.

Branch heat is settled as **`max`** (a weak/duplicate conjunct cannot inflate dominance),
and `conjunction`/`ambiguous` are orthogonal derived readings with dominance evaluated at
the **branch** level. #6B has passed its paper gate.

## Pass conditions

1. Fires on AMB. ✓
2. Silent on CONJ, CONF, CTX. ✓
3. Each foil silenced by a distinct necessary condition. ✓
4. Invariant under ID swap pre-evidence; the signal names the slot, not a winner. ✓
5. Recency cannot tip the balance (heat excludes `F`). ✓ (by construction of the crossing)

## Verdict

The three-condition signal **discriminates ambiguity from all three belonging shapes**,
and the fixture proves the discrimination is not accidental — remove any clause and a
specific foil leaks. With **co-occurrence as a Jaccard rate (v2)** it is strictly better
than the boolean: it additionally fires on `AMB-NOISE` (a real fork with one accidental
co-occurrence) without silencing any belonging case, and it exposes the middle band as a
tunable **valley-depth** threshold instead of pretending the band does not exist. It is
safe to build. The one thing to carry into the code: `θ` is a **config parameter shared**
with differentiation and the dominance margin — not a hard-coded constant, and not a
boolean intersection.

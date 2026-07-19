# Paper Test #4 — Segmentation Recovered from the Unsegmented Spine

**Date:** 2026-07-20
**Tests:** the last hand-wave — [episode-as-region](episode-as-region.md)'s claim that
span boundaries (renewal gaps) and class membership (co-occurrence signature) are
computable off-wave from spine history, with no LLM judge in the durable path.
**Verdict:** both rules recover the hand segmentation — 4/4 boundaries, 3/3 classes —
plus a principled provisional slot for the open tail span, and the identity rule
independently corroborates test #3's collapse. Two new substrate requirements
surfaced (witnessed timestamps; signatures over *resolved entities*, not tokens) and
one erratum in the toy was found and repaired. **Every slot of the observe mechanism
now has a paper-validated rule.**

## Erratum in the toy (found preparing the input)

Test #2's extension created an O→O adjacency (m15 `O8` → m16 `O9`), violating strict
alternation; tests #1's boundary fillers (A3, A6) were present at the other seams but
this one was missed. Repaired here by inserting a rest action `m15b`. Two lessons:

1. With timestamps, rest fillers are revealed as artifacts — **the time gap is the
   boundary signal**, not the filler. The fillers existed only to satisfy alternation.
2. Whether strict O→A alternation is a spine *invariant* or an idealization (real
   streams have observation runs with no intervening act; a null-action is a
   modeling patch) is now a **parked question** for the spine substrate.

## Input — the unsegmented spine, with witnessed timestamps

Segmentation by renewal requires wall-clock time. The runtime witnesses every moment
anyway; the requirement is only that **timestamps are spine facts** (witnessed,
immutable), not annotations. The toy, timestamped plausibly and stripped of all
episode labels:

```text
day 1   08:00 m1    08:01 m2    08:03 m3    08:04 m4    08:06 m5    08:07 m6(rest)
day 1   19:30 m7    19:31 m8    19:33 m9    19:34 m10   19:36 m11   19:37 m12(rest)
day 2   09:00 m13   09:02 m14   09:04 m15   09:05 m15b(rest)
day 2   15:20 m16   15:22 m17   15:24 m18
day 2   15:40 m19   15:41 m20
```

Also given (all witnessed): kind-2 resolved-into events per moment, kind-6
settlements. Nothing interpreted is consumed.

## Rule 1 — boundaries from renewal gaps

Successive gaps, in minutes:

```text
intra-activity   1, 2, 1, 2, 1 | 1, 2, 1, 2, 1 | 2, 2, 1 | 2, 2 | 1
candidates       m6→m7: 683    m12→m13: 803    m15b→m16: 375    m18→m19: 16
```

A single global threshold works on this toy (anything in 5 min – 6 h), but the honest
structure is **multi-scale**: the gap spectrum has three bands (minutes / tens of
minutes / hours). The adaptive local rule that needs no global constant:

> **boundary ⇔ gap > k × median of recent intra-gaps** (k ≈ 5 here: 16 > 5×2; 1–2 not)

This cuts all four boundaries, including the instructive m18→m19 case (16 min — same
afternoon, different occurrence). A naive "largest multiplicative jump" cut would have
merged m16–m20 into one span. Recovered spans:

```text
S1 = m1–m6     S2 = m7–m12     S3 = m13–m15b     S4 = m16–m18     S5 = m19–m20 (OPEN)
```

S5 has no settlement yet (test #3's uncommitted impulse) — the **open tail span** is
simply the span not yet closed, which is what dnav2 calls the active episode.
Boundary-strength being graded (16 min ≪ 11 h) points at hierarchical segmentation
(activity < session < day) — closure-duration-objective's territory; parked.

## Rule 2 — membership from co-occurrence signatures

Signature = resolved entities + verb classes committed within the span (kind-2 facts).
Weight by rarity across spans (independence baseline): df 5/5 → 0.0 (michael), df 2/5
→ 0.92, df 1/5 → 1.61.

```text
S1  {michael₀, blue-bottle, kitchen-fridge, drink-verb, put-back-verb}
S2  {michael₀, kitchen-fridge, garage-fridge, soda, fetch-verb}
S3  {michael₀, plants, water-substance, watering-can, water-verb}
S4  {michael₀, soda, garage-fridge, fetch-verb}
S5  {michael₀, garage-place, blue-bottle, put-back-verb}   (fridge unresolved — open)
```

Pairwise weighted overlap:

```text
S2·S4 = garage-fridge + soda + fetch-verb          = 2.76   MERGE
S1·S5 = blue-bottle + put-back-verb                = 1.84   MERGE (provisional: S5 open)
S1·S2 = kitchen-fridge                             = 0.92   below threshold
all others                                         ≈ 0
```

Cluster at threshold ≈ 1.5 (≥ two informative shared terms):

```text
{S2, S4}      = EP-fridge-fetch      ✓
{S1, S5*}     = EP-water-routine     ✓  (* provisional until S5 settles)
{S3}          = EP-plant-care        ✓
```

**3/3 classes recovered.** The near-miss is the important row: S1·S2 share
kitchen-fridge (0.92) and correctly do *not* merge — a single shared place must never
found a class. The independence weighting plus the two-term threshold is the
operational form of relational-region-identity's "co-occurrence above the independence
baseline."

## Findings

1. **The hand segmentation is recoverable mechanically.** 4/4 boundaries from
   arithmetic on witnessed timestamps; 3/3 classes from weighted co-occurrence over
   kind-2 commits. No judge in the durable path — confirming episode-as-region's
   placement of the LLM edit/new as, at most, a real-time boundary *proposer* for
   ambiguous gaps (the 16-minute band is where it could earn that keep).
2. **Timestamps must be witnessed spine facts.** New substrate requirement; the
   runtime already has them, the spine schema must keep them.
3. **Signatures must be over resolved entities, not surface tokens** — token-level
   "water" would have blurred S1 (bottle) into S3 (plant-water). No circularity
   results: today's segmentation consumes *yesterday's settled commits* (kind-2 spine
   facts) — a temporal bootstrap the mutability table already licenses ("materialized
   caches over settled resolution history").
4. **The fan effect has an identity-side twin.** michael (df 5/5) weighs 0.0 in every
   signature — the same promiscuity normalization as test #1 finding 5, now doing the
   same work on the membership side. One principle, both sides: ubiquity carries no
   information.
5. **Segmentation has bookends too.** A span's membership is provisional while the
   span is open and commits at closure (settlement) — "strict at commit, plural at
   propose," extended: *strict at close, provisional while open.* If michael corrects
   S5 toward the garage fridge, its final signature changes and its class assignment
   rebuilds — cheap to be wrong about, exactly as designed.
6. **Cross-validation with test #3:** the identity rule assigns S5 to
   EP-water-routine (bottle + put-back, 1.84) — the same answer test #3's wave reached
   through root diversity (variant B). Two independent mechanisms, one verdict. The
   segmentation rule and the convergence rule agree about what the garage moment *is*.

## Consequences

1. **The observe mechanism is fully paper-validated.** Wave shape, conservation at
   class, Born-rule statistics, root-diversity tie-break, coverage gate, joint-
   precedent novelty flag, and now derived segmentation — every slot has a rule that
   survived its toy. The compression thread also closes cleaner: nothing above the
   spine needs interpretation to *exist*; the judge is optional everywhere.
2. **Substrate deltas before code:** witnessed timestamps on moments; signatures
   keyed by resolved entities + verb classes; open-span provisional membership; the
   strict-alternation question parked with a null-action patch as fallback.
3. **Next grade comes from reality, not paper.** The toys are constructed (ties and
   all, by design). The next step is implementation against the live pipeline —
   replay a real session's trace through segmentation first (rules 1+2 are pure
   functions over spine history, testable standalone), then wire the wave.

# Toy Test #6B — Branch Partition (Multi-Way, Corrected Mechanism)

**Date:** 2026-07-21
**Status:** RUN (paper walk) · **the gate Phase 1's branch partition must pass before
code.** Toy #6 proved only the two-completion case; #6B is the multi-way gate.
**Parent:** [ambiguity-signal-discrimination](ambiguity-signal-discrimination.md) (Toy #6),
[interpretation-implementation-phases](interpretation-implementation-phases.md) (Phase 1).
**Walks the mechanism AFTER two blocking corrections** (both raised by Codex): polarity-
separated settlement spans, and a permutation-invariant partition.

---

## The corrected mechanism under test

1. **Group** activated completions by source (the `edgeKey` prefix).
2. **Co-occurrence over POSITIVE settlement spans only.** Settlement spans are indexed by
   outcome — `settlementSpans[edgeKey][TruthTrue]` and `[TruthFalse]` — because `grade()`
   settles *every* open prediction at the observed `(subject, predicate)` address: when the
   world shows `B`, an open `A` prediction settles **false** at that same address. Recording
   both indiscriminately makes competing alternatives always share spans (`Jaccard → 1`,
   every fork misread as a conjunction). Co-occurrence therefore uses **TruthTrue spans
   only** — "the completion *occurred* here," not "was graded here."
   ```text
   cooc(A,B) = jaccard( positiveSpans[S→A], positiveSpans[S→B] )
   ```
3. **Partition — permutation-invariant, not greedy:**
   ```text
   a. compatibility graph:  node per completion; edge iff cooc ≥ cfg.ConjunctionOverlap
   b. connected components
   c. component is a CLIQUE   → emit as one branch (all pairwise co-occur = real conjunction)
      component is NOT clique → whole component UNRESOLVED (hand to Phase 3; never guess a split)
   ```
   (Complete-linkage greedy is rejected: starting at `A` vs `C` on a bridge yields different
   partitions — an ID-order dependence that violates the metamorphic invariant.)
4. **Branch heat = `max(member heat)`** — never `sum`/`avg`. A branch is as hot as its
   hottest member, so adding a weak conjunct cannot inflate dominance.
5. **Derived readings (precedence order; `Unresolved` VETOES confidence):**
   `UNRESOLVED ⟺ len(Unresolved) > 0` (a non-clique component exists — never commit,
   regardless of any clean branch); else `ambiguous ⟺ ≥2 branches ∧ no dominant branch`
   (`margin < cfg.DominanceMargin`); else `confident ⟺ exactly one branch OR one branch
   dominates`. A single branch needs no fictional runner-up: `{A}` is confident and one
   conjunctive branch `{A,B}` is confident + conjunction.
   `conjunction ⟺ ANY emitted branch has >1 completion` — **branch-local** and orthogonal
   (not "the committed branch": MIXED has a conjunctive `{A,B}` while nothing is committed).
   Confidence and ambiguity are source-level; conjunction is branch-local.

## The walk (`cfg.ConjunctionOverlap = 0.2`, positive spans)

```text
CASE            positive spans                       compat edges   components→partition   reading
──────────────────────────────────────────────────────────────────────────────────────────────────────
3-WAY FORK      A{s1,s2} B{s3,s4} C{s5,s6}           none           {A}{B}{C} (all cliques) ambiguous ×3
MIXED           A{s1,s2,s3} B{s1,s2,s3} C{s4,s5}     A–B (J=1)      {A,B}✓clique | {C}      conj WITHIN {A,B},
                                                                                            ambiguous BETWEEN
DOMINANT BRANCH A,B co-occur hot .4; C cold .05      A–B            {A,B}* | {C}            confident → {A,B}
NON-TRANSITIVE  A{s1,s2,s3} B{s2,s3,s4} C{s4,s5,s6}  A–B(.5) B–C(.2) {A,B,C} not clique     UNRESOLVED (→Phase 3)
VETO            D{s7,s8} hot .5; A~B~C bridge cold .4 D isolated; A–B,B–C  {D} | {A,B,C}✗   UNRESOLVED, NOT confident-D
```

- **3-WAY FORK** — three singleton components; a singleton is trivially a clique, so three
  branches, balanced heat, no dominant → **ambiguous** three ways.
- **MIXED** (Codex's case the flat enum could not hold) — component `{A,B}` is `K₂`, a
  clique → one conjunction branch; `{C}` a singleton branch. Two branches, `max` heat
  `0.4 = 0.4` → **ambiguous between**, conjunctive within `{A,B}`.
- **DOMINANT BRANCH** — same partition; branch heat `{A,B}=max=0.4` vs `{C}=0.05`, margin
  large → **confident**, commit the conjunction `{A,B}`.
- **NON-TRANSITIVE** — `B` bridges `A` and `C` but `A≁C`; the component `{A,B,C}` is
  connected yet not a clique → the **whole component is unresolved**, identically whatever
  order the nodes are visited. No arbitrary `{A,B}|{C}` vs `{A}|{B,C}` choice is made.
- **VETO** (unresolved components veto confidence) — a clean dominant branch `{D}` (heat
  `0.5`) sits beside an unresolved bridge component `{A,B,C}` (heat `~0.4`). If confidence
  counted **only emitted branches**, `{D}` would look sole/dominant and the source would
  **commit while ignoring un-compared evidence**. The rule: `len(Unresolved) > 0` →
  state = **UNRESOLVED**, never confident. An unresolved lump is evidence that has not been
  compared, so it must **block** commitment until Phase 3 differentiates it.

## Regression 1 — polarity separation (blocking bug #1)

Source `S`, address `(S, pred)`, with both `A` and `B` predictions open in each trial.
Span `s1`: world shows `A` → `A` true, `B` false. Span `s3`: world shows `B` → `B`
true, `A` false.

```text
INDISCRIMINATE (all graded spans):
    trialSpans[A] = {s1, s3}   trialSpans[B] = {s1, s3}   →  Jaccard 1.0  →  CONJUNCTION   ✗ WRONG
POSITIVE-ONLY (TruthTrue):
    positiveSpans[A] = {s1}    positiveSpans[B] = {s3}    →  Jaccard 0.0  →  AMBIGUOUS     ✓ CORRECT
```

Polarity separation flips the classification from a silent, catastrophic false-conjunction
to the correct fork. This is a **required** #6B fixture.

## Regression 2 — permutation / metamorphic invariance (blocking bug #2)

Run every case with completion IDs relabelled and insertion order reversed. Connected
components, the clique test, and `max` heat are all set operations → **partition, branches,
and unresolved set are identical under the rename**. The bridge in NON-TRANSITIVE resolves
to the same whole-component-unresolved result regardless of start node. Committed state and
residual are invariant pre-evidence.

Metamorphic heat check: duplicating a member of a branch, or adding a **weak** conjunct
(heat below the others), does **not** change that branch's `max` heat and therefore does not
change dominance — a conjunction cannot win a fork by member count.

## Pass conditions

1. 3-way independent fork → three branches, ambiguous. ✓
2. Conjunction-within / ambiguity-between represented (not collapsible to one enum). ✓
3. Dominant branch → confident commit to the conjunction; minor branch not lost. ✓
4. Non-transitive bridge → whole component unresolved, **order-independent**. ✓
5. **Unresolved vetoes confidence** — a dominant clean branch beside an unresolved component
   reads UNRESOLVED, never confident (VETO case). ✓
6. Polarity: a false-settling alternative never counts as co-occurrence. ✓
7. ID/insertion-order permutation changes nothing pre-evidence. ✓
8. Weak/duplicate conjunct does not raise branch dominance (`max` heat). ✓
9. `conjunction` is branch-local (any emitted branch >1 completion), not tied to a commit. ✓
10. One singleton branch → confident; one multi-completion branch → confident + conjunction. ✓
11. In a multi-candidate source, missing positive-span provenance → UNRESOLVED, not
    evidence of separation (`unknown ≠ false`). ✓

## Verdict

The **corrected** mechanism — positive-settlement co-occurrence + connected-components/clique
partition + `max` branch heat + `Unresolved`-vetoes-confidence — passes all eleven, including
the two regressions that sank the earlier draft. The partition is now deterministic and conservative: it emits only real
cliques as branches and refuses to guess a split for a bridge, deferring that to Phase 3
differentiation. **Phase 1's branch partition has passed its paper gate.** The greedy
algorithm and outcome-blind index in the prior draft are superseded by this walk.

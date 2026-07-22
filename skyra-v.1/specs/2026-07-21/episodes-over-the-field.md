# Episodes Over the Field

**Date:** 2026-07-21
**Status:** DESIGN SYNTHESIS · the organizing spine for the interpretation layer.
The seven 07-21 research docs turn out to be facets of one machine; this records it.
**Grew out of:** the [balanced-ambiguity run](balanced-ambiguity-run.md) (Toy #5 walked
against the live path) and the design conversation that followed it.
**Re-hangs:** [two-axis-inference](two-axis-inference.md),
[predicate-diagnosticity](predicate-diagnosticity.md),
[the-two-bit-outcome](the-two-bit-outcome.md),
[scale-structure-content](scale-structure-content.md),
[novel-categories-as-differentials](novel-categories-as-differentials.md),
[genuinely-balanced-ambiguity](genuinely-balanced-ambiguity.md).
**Supersedes the framing of:** [episode-as-region](../2026-07-20/episode-as-region.md)
(episode = class of similar spans) — an episode is not a class computed off to the side;
it is an **owned heated region of the activation field**.

---

## Thesis

> An episode owns a particular heated region. The criterion that defines it **is the
> field** — membership is co-heating, not a predicate applied from outside. Combination
> and differentiation are one machine driven by density, it scales because the field
> bounds the working set, and abstraction falls out for free.

Everything below is that sentence unpacked.

## 1. Episode = owned heated region; the criterion is the field

Membership is not tested, it is **observed as co-activation**. Two things belong to the
same episode because they heat together, and they heat together because they share
structure — the posting lists already spread activation through shared
entities/triples/edges (`heatPosting` in `activation_field.go`). So there is nothing to
cluster and no predicate to define: the episode *is* the region that lights up.

This is why the [Toy #5 run](balanced-ambiguity-run.md) returning `[A, B]` together was
**the field telling the truth**, not a flattening bug. With no density to separate them,
A and B genuinely belong to one undifferentiated region. The earlier instinct to bolt on
an exclusivity marker was imposing a split the field had not earned.

## 2. Combination and differentiation are one machine

- **Combination** — co-heating spans (and criteria) collapse into one region. Criteria
  conjoin (`R ∧ context-X`) into compound episodes; this is `Edge.Source` with `len > 1`.
- **Differentiation** — as a region accumulates, density concentrates unevenly. A
  **pocket** forms, a valley opens between pockets, and the region splits *at the valley*.

The direction matters: differentiation is **bottom-up, density-first**. A criterion is
not imposed top-down as a diagnostic predicate `D` and then used to split; the pocket
forms first, and a predicate is fitted **after the fact only to name it** — for
retrieval, for communication, for a differential's "differs on `D`." Density discovers
the structure; the predicate describes it.

```text
sparse region     one blur — A and B co-heat, no valley           → undifferentiated (truth)
      │ accumulation
      ▼
dense region      ·A·A·A·   ·B·B·B·   two pockets, a valley opens  → differentiate at the valley
```

**A pocket of density is a code.** A tight sub-region is exactly what MDL would mint a
symbol for (see [replacing-triples-with-compression](../2026-07-20/replacing-triples-with-compression.md)):
recurrence pays for the code, the code is small. "Differentiate where density
concentrates" and "mint a code where recurrence pays" are the same act from two sides.

**Abstraction is free.** The region is the abstract grouping; its pockets are the
specializations; the ladder between them is **resolution on a continuous activation
value**. Lower the threshold → coarser, more abstract episode. Raise it → finer, more
specific. You do not build a hierarchy — you read it off the density at whatever
resolution you ask for, because the field holds continuous heat rather than discrete
classes.

## 3. Belonging must not blur into ambiguity

The danger in raising an "ambiguity signal": every region contains many co-heated
members, so *balanced heat alone is not ambiguity* — it is also what "two things I both
want" looks like. Raising on balanced heat would flag belonging as ambiguity constantly.

The discriminator is **co-occurrence** — the wave's `JointPrecedent`, in new clothes.

```text
                     balanced heat   same source   co-occur in history   verdict
confident                no              —               —               commit
conjunction (belong)     yes             yes             YES              both wanted
context (belong)         yes             no              —               members
AMBIGUOUS                yes             yes             NO               pending → clarify
```

**Belonging things have been witnessed together; ambiguous things are alternatives
witnessed apart.** The ambiguity signal is the *conjunction* of three conditions, none
sufficient alone:

```text
same source  ∧  empty co-occurrence  ∧  no dominant heat
```

Co-occurrence is a cheap set intersection over the existing index:

```text
Spans.Triple(A) ∩ Spans.Triple(B)  ==  ∅   → never together → alternatives → ambiguous
                                   ≠  ∅   → together        → belonging
```

**State of the substrate:** the signal is *not wired* on the live path.
`deriveActivatedTarget` groups candidates by target and drops both the source and the
co-occurrence, so its `[A@h, B@h]` output cannot tell the three worlds apart. But every
ingredient is present — `Edge.Source` and `SpanIndex` postings — so raising it is
**reviving joint-precedent on the live path**, not inventing a subsystem. The one free
parameter is what counts as "no dominant heat" (the margin between the top two
completions) — the same **valley depth** threshold differentiation needs.

## 4. Cold rides the LLM; dense earns the signal

- **Cold / newly formed region** — no counted structure exists, so there is *no
  deterministic signal to compute*. The decision to raise a clarifying question bottoms
  out in **expression + prediction**: the LLM raises it or it does not, per its
  lab-trained competence. This is not a curve we fight; it is one we ride. Do not rebuild
  a worse clarification-detector than the labs already shipped.
- **The non-negotiable** — the span label must preserve the **counted** material (source,
  span membership, co-occurrence) even while the LLM carries the decision, so the region
  can **graduate off the LLM** as it densifies. Cold: the LLM guesses. Dense: a counted
  signal takes over, *better calibrated because it is counted*. LLM behavior is the prior;
  the counted structure is the cheaper, sharper, distilled replacement. "Bottom out in
  the LLM" must not mean "stay there forever" — the label is the seed of the hand-off.
- **The signal fades in** — it is a function of density/span/episode, so its very
  *existence* is the maturity gate. No signal → the LLM decides. Signal present → the LLM
  acts on a counted input. Maturity chooses automatically; there is no manual switch
  between "LLM handles it" and "we compute it."

**Calcification, carefully.** As runs accumulate:

- **Resolvable ambiguity** (a differentiator exists) — a valley opens, the region splits
  into resolved branches, and the ambiguity leaves the calcified core. Clustering narrows.
- **Irreducible ambiguity** (no differentiator — the world is genuinely 50/50) — more
  runs do *not* narrow it; they **confirm** it. The region calcifies *with* the fork, as a
  well-counted coin flip. This is required, not a failure: a system that "cleans out"
  genuine underdetermination turns implementation noise into false belief (Toy #5's
  warning). Calcification here means "I now know with high `N` that this is a coin flip,"
  not "I resolved it." The same co-occurrence test sorts the two: a valley that
  eventually opens = resolvable; density that grows without ever opening one = irreducible
  and correctly permanent.

## 5. Why this scales — episodes live on the field

Episodes are not a global clustering pass. The activation field already bounds the
working set: only spans above `spanFloor` sit in `RankedSpans`. Combination and
differentiation are computed **only over currently-active spans** — `O(active)`, never
`O(|log|)`. An episode is a **transient projection over the heated region**, refreshed
when the region reheats: the same "materialized, rebuildable, cheap-to-be-wrong-about"
pattern as `ResolutionIndex`. This is what reconciles "episodes are back" with the
flat-scaling continuous-replacement path — episodes do not live in a persistent
clustering index, they live on the field, and the field is the bound.

## 6. The two load-bearing unknowns

This synthesis unifies the mechanism; it does **not** dissolve the two open questions it
is made of. It just puts them in one place instead of seven.

1. **The criterion vocabulary** — what can a pocket-naming predicate even *be*? This is
   where [scale-structure-content](scale-structure-content.md) enters: observation budget
   → structure slots → falsifiable predicates. Criteria are learned predicates fitted to
   density, never hand-written.
2. **The split/merge thresholds** — how deep a valley before a region differentiates, how
   shallow before branches recombine? This shares a **principle** with "no dominant heat"
   in the ambiguity test and the promotion threshold in
   [novel-categories-as-differentials](novel-categories-as-differentials.md) — *how much
   separation before it's a real distinction* — but **not a single value**: they are
   statistically different quantities (promotion is an evidence count, not a `[0,1]`
   overlap), so they are independent named fields in `DifferentiationConfig`, unified only
   if evidence later shows one latent quantity. Share the principle, not the scalar.

## 7. The A–F phases, re-hung under the spine

The earlier phase list was seven parallel islands. Under this spine they are facets of one
build:

```text
episode = owned heated region  ── criterion IS the field ── scales on the field
   │
   ├─ combination        co-heating merges                     (Edge.Source ∧)
   ├─ differentiation    density pockets → split → name        (diagnosticity, differentials)
   ├─ belonging≠ambiguity  co-occurrence intersection          (revive JointPrecedent live)
   ├─ cold→dense          LLM prior → counted signal fades in  (two-axis, boundary doc)
   └─ abstraction         resolution on continuous density     (free)
```

Concrete first build step, unchanged in spirit but re-rooted: **revive joint-precedent on
the live path** — group `deriveActivatedTarget` candidates by `Edge.Source`, compute the
co-occurrence intersection, and raise the three-condition ambiguity signal (`same source ∧
empty intersection ∧ no dominant heat`) for the express side to act on. That single move
gives the region a legible differentiation-state without imposing any split the field has
not earned, and Toy #5 becomes its regression test.

## Bottom line

Episodes are not classes computed beside the field — they are regions **of** the field.
One machine combines by co-heating and differentiates at density valleys; abstraction is
just the resolution you read it at; belonging and ambiguity are held apart by
co-occurrence, not by heat; and the whole thing rides the LLM while cold and distills a
counted signal as it calcifies — without ever deleting a genuine coin flip. It scales
because it never leaves the active region. The two hard unknowns — the criterion
vocabulary and the valley-depth threshold — are named and localized, which is the most an
honest synthesis should claim before code.

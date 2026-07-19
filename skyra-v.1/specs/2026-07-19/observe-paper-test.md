# Paper Test — Observe Wave over a Toy Spine

**Date:** 2026-07-19
**Tests:** [observe-mechanism-over-spine.md](observe-mechanism-over-spine.md)

> **2026-07-20 notice:** Findings 3 and 4 (episode layer and kind 4 as deletion
> candidates) held only for episode-as-single-span-wrapper. Superseded by
> [episode-as-region](../2026-07-20/episode-as-region.md): episode = class of similar
> spans, which [paper test #2](../2026-07-20/observe-paper-test-2.md) validates. Open
> question 1 (convergence metric) is settled there: root diversity first, heat second.
**Verdict:** the wave resolves both ambiguous references correctly, but only at span
granularity — episode-level totals produce an exact tie. Three edge kinds fold into
derivations of the others. No signed representation was needed; the negative settlement
did its work at the collapse bookend, not in the middle.

## Setup

### The spine (15 moments, strictly alternating O/A)

Three derived episodes over one continuous spine. E2 is the retained
confusion-then-correction precedent. E3 carries the "water" homonym distractor.

```text
 m   node  content                                              episode  settlement
 1   O1    michael: "I'm thirsty"                               E1
 2   A1    resolve my-water -> blue-bottle; fetch from          E1       + (by O2)
           kitchen-fridge
 3   O2    michael has bottle, drinks, satisfied                E1
 4   A2    return bottle to kitchen-fridge                      E1       + (by O3)
 5   O3    bottle stored in kitchen-fridge, refilled            E1
 6   A3    session close / rest                                 —        ungraded
 7   O4    michael: "grab a drink from the fridge"              E2
 8   A4    resolve fridge -> kitchen-fridge; fetch              E2       − (by O5)
 9   O5    michael: "no — the garage fridge"                    E2
10   A5    re-resolve fridge -> garage-fridge; fetch soda       E2       + (by O6)
11   O6    michael has soda                                     E2
12   A6    log / rest                                           —        ungraded
13   O7    michael: "the plants look dry"                       E3
14   A7    water plants from watering-can                       E3       + (by O8)
15   O8    plants perked up                                     E3
```

### The five lenses (entity -> episodes)

```text
water-bottle     {E1}
kitchen-fridge   {E1, E2}
garage-fridge    {E2}
michael          {E1, E2, E3}
water-substance  {E3}
```

### The impulse

`"go get my water from the fridge"` — appended as new observation O9, pre-commit.

Proposals (interpreted, plural, equal split under conservation):

```text
"my water"    -> water-bottle 0.5, water-substance 0.5
"the fridge"  -> kitchen-fridge 0.5, garage-fridge 0.5
speaker/"my"  -> michael 1.0
```

## Trace

### Lens-hop (edge kind 4), conservation split across each lens's episodes

```text
water-bottle    0.5  -> E1 0.500
water-substance 0.5  -> E3 0.500
kitchen-fridge  0.5  -> E1 0.250, E2 0.250
garage-fridge   0.5  -> E2 0.500
michael         1.0  -> E1 0.333, E2 0.333, E3 0.333
```

Episode arrival totals:

```text
E1 = 0.500 + 0.250 + 0.333 = 1.083
E2 = 0.250 + 0.500 + 0.333 = 1.083     <- exact tie with E1
E3 = 0.500 + 0.333         = 0.833
```

**Finding 1 — the tie.** Episode-level accumulation cannot decide this impulse. E1 and
E2 tie by symmetry (the bottle's 0.5 mirrors the garage fridge's 0.5; kitchen-fridge
splits evenly; michael is uniform). Discrimination only appears below, at span/moment
granularity. The episode layer carried no information the span didn't.

### Descend (kinds 3 + 5): deposit each entity's episode amplitude across its resolved-from moments

```text
E1:  water-bottle   0.500 over {m2,m3,m4,m5} -> 0.125 each
     kitchen-fridge 0.250 over {m2,m4,m5}    -> 0.083 each
     michael        0.333 over {m1,m3}       -> 0.167 each

E2:  kitchen-fridge 0.250 over {m8}          -> 0.250
     garage-fridge  0.500 over {m9,m10,m11}  -> 0.167 each
     michael        0.333 over {m7,m9,m11}   -> 0.111 each

E3:  water-substance 0.500 over {m14}        -> 0.500
     michael         0.333 over {m13}        -> 0.333
```

Moment heat after additive Slice accumulation (before walk-spreading):

```text
E1:  m1 0.167   m2 0.208   m3 0.292   m4 0.208   m5 0.208     span sum 1.083
E2:  m7 0.111   m8 0.250   m9 0.278   m10 0.167  m11 0.278    span sum 1.083
E3:  m13 0.333  m14 0.500                                     span sum 0.833
```

**Finding 2 — hottest moment is the wrong read.** The single hottest moment in the
whole field is m14 (0.500) — the plant-watering distractor, hot only because
water-substance has exactly one resolved-from moment. m8, the *negatively settled*
kitchen resolution, is E2's most concentrated deposit for the same reason. Single-root
spikes are an artifact of fan-in, not evidence. The doc already says "hottest **span** =
precedent" — the paper test confirms span is the correct unit, and suggests convergence
should additionally weigh **root diversity**: E1's span is heated by all three impulse
roots; E2's by two; E3's by two but with no fridge binding at all.

### Grade-read (kind 6) at each action in the warm spans

```text
E1:  A1 +   A2 +            both support {my-water=bottle, fridge=kitchen}
E2:  A4 −   A5 +            kitchen reading defeated here; garage reading affirmed
E3:  A7 +                   supports plants-watering, no fridge binding present
A3, A6: ungraded — grade-read skips them (no heat contribution)
```

### Converge and collapse

Joint candidate readings, judged by shared-span convergence plus settlement counts
(the Born-rule row):

```text
R1  {my-water -> water-bottle, fridge -> kitchen-fridge}
    shared span m1–m5: all three roots converge; settlements + +      SELECTED
R2  {fridge -> garage-fridge}
    span m7–m11: settlement + on A5, but no "my water" candidate
    lands anywhere in the span — no joint convergence
R3  {my-water -> water-substance}
    span m13–m15: hot but fridge-free; settlement is about plants
```

R1 wins by constructive interference in exactly the doc's sense: compatible paths
(bottle ∧ kitchen-fridge ∧ michael) converging on the same spine span. Collapse commits
O9 -> {water-bottle, kitchen-fridge} as a spine event, precedent span m1–m5 attached
with its two positive settlements. Strict, singular, falsifiable: if michael answers
"no — I moved it to the garage fridge," that correction is witnessed, becomes a new
E2-shaped precedent, and shifts the counted statistics for the next impulse.

**The correct precedent span heats and both ambiguous references resolve by
convergence.** The mechanism passes its own stated check.

## Edge-kind audit (the deletion candidates the doc asked for)

```text
kind  used in the wave?              status after test
1     spine-walk                     load-bearing, witnessed
2     written at collapse only       load-bearing, but only as the commit record
3     descend (targeted deposit)     = inverse index of kind 2
4     lens-hop                       = kind 3 composed with kind 5 membership
5     descend (span boundary)        load-bearing as the walk region
6     grade-read                     load-bearing, witnessed
```

**Finding 3 — the inventory compresses.** Witnessed facts are {1, 2, 6}. Kind 3 is
mechanically the inverse of kind 2; kind 4 is 3 ∘ 5. The lens layer is confirmed as a
pure materialized cache — it was never consulted for anything the composition couldn't
answer. A compression-native mechanism could keep only {1, 2, 5, 6} durable and treat
3 and 4 as rebuildable indexes (which the doc's mutability table already licenses:
"deletable, rebuildable, cheap to be wrong about").

**Finding 4 — the episode layer moved with the span layer at every step.** Amplitude
passed *through* episodes without episodes ever discriminating (Finding 1 is the proof:
they tie). Candidate simplification: lenses point directly at hot spans
(entity -> span), deleting episode as an activation carrier and keeping it only as a
narrative label.

**Finding 5 — the fan effect behaved as designed.** michael participates in every
episode and contributed a flat 0.333 floor everywhere — zero discrimination,
self-attenuated by conservation. Compression candidate: skip lens-hop entirely for
entities whose episode coverage exceeds a threshold; their contribution is a constant
that normalizes away.

**Finding 6 — no signed representation was needed.** A4's negative settlement never had
to *subtract* amplitude mid-wave; it acted only at the collapse bookend, where R2/the
kitchen-in-E2 reading was already losing on convergence. This is a data point for
keeping the middle non-negative and interpretation-free, and letting settlement
statistics do their work at selection. Destructive interference would only earn its
keep on *contradicting bindings converging on the same span* — which this toy cannot
produce. That is the case paper test #2 should construct.

## Proposed paper test #2

Impulse: `"grab something from the fridge"` — no water binding at all. Expected: E2's
precedent should now win on counted settlement statistics alone (garage +, kitchen −),
i.e. the Born-rule row selecting between candidates that *both* converge. If the
mechanism cannot prefer garage there without a signed middle, the open
destructive-interference decision becomes forced rather than optional.

## Open questions surfaced

1. Convergence metric: raw span heat, or span heat weighted by distinct-root count?
   The toy says root diversity is the real signal (Finding 2).
2. Walk-spreading radius: the trace stopped at deposit + grade-read; how far the
   spine-walk bleeds heat across episode boundaries (m5 -> m6 -> m7) was never
   exercised. A boundary-bleed case belongs in test #2.
3. Where does the resolution act live on the spine — is the collapse commit itself the
   next A node (making O9 -> A8 succession witnessed), per praxis-ground's "a completed
   action is an ordinary CausalNode"? The trace assumed yes.

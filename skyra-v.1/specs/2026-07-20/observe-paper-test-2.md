# Paper Test #2 — Class Statistics and the Root-Diversity Tie-Break

**Date:** 2026-07-20
**Tests:** [episode-as-region.md](episode-as-region.md) against the case
[paper test #1](../2026-07-19/observe-paper-test.md) said would force the
signed-middle decision.
**Verdict:** class statistics resolve the no-binding impulse without a signed middle —
the destructive-interference decision stays optional. Episode-as-class demonstrably
aggregates evidence the flat model discards. The class-level tie recurs and is broken
by root diversity, yielding a clean principle: **superposition lives within a root,
interference only happens across roots.**

## Setup

Test #1's spine extended by one occurrence (a second garage fetch), so the fridge-fetch
class has real statistics:

```text
16   O9    michael: "get me a soda"                     +EP-fridge-fetch
17   A8    resolve fridge -> garage-fridge; fetch       settled + (by O10)
18   O10   michael has soda from garage-fridge
```

> **Erratum (found 2026-07-20 in test #4):** m15 `O8` → m16 `O9` is an O→O adjacency,
> violating strict alternation — the boundary filler action was missed here. Repaired
> as `m15b(rest)` in [observe-paper-test-4.md](observe-paper-test-4.md), which also
> shows the fillers are artifacts once timestamps carry the boundary signal.

Episodes are now classes of similar spans (hand-assigned membership; the derived
segmentation rule is episode-as-region's open gap):

```text
EP-water-routine   members {m1–m5}            fridge->kitchen 2/2 (A1+, A2+)
EP-fridge-fetch    members {m7–m11, m16–m18}  fridge->kitchen 0/1 (A4−)
                                              fridge->garage  2/2 (A5+, A8+)
EP-plant-care      members {m13–m15}          (no fridge binding)
```

Conservation per the addendum: amplitude heats the episode, statistics are read there,
descent only to the exemplar. michael's flat floor omitted (test #1 finding 5).

## Test 2a — `"grab something from the fridge"`

No water binding. One contentful root:

```text
"the fridge" -> kitchen-fridge 0.5, garage-fridge 0.5
```

Lens-hop to classes:

```text
kitchen-fridge 0.5 -> EP-water-routine 0.25, EP-fridge-fetch 0.25
garage-fridge  0.5 -> EP-fridge-fetch 0.50

EP-fridge-fetch 0.75    EP-water-routine 0.25    EP-plant-care 0
```

EP-fridge-fetch is the hot class. Born-rule read of its counted statistics:
kitchen 0/1, garage 2/2 → **collapse: fridge → garage-fridge**, exemplar span
m16–m18 attached as precedent.

**Finding 1 — no signed middle needed, again.** The kitchen candidate's amplitude
*helped heat the winning class* (its 0.25 into EP-fridge-fetch); its defeat happened
at selection, off counted settlements. Clean division of labor: **heat answers "which
precedent class"; counts answer "which candidate."** Test #1's prediction — that this
impulse would force the destructive-interference decision — is falsified. The decision
stays optional.

**Finding 2 — the flat model fragments the evidence.** Under test #1's
single-span-episode model the same impulse reads only the hottest span (E2: kitchen −,
garage +) and the m16–m18 confirmation contributes nothing to selection. It happens to
give the same answer here, off a 1-observation margin; the class model selects off
2/2 vs 0/1. Aggregation-across-members is the mechanism's actual memory of "this is
how fridge fetches go" — the playbook, literally (region-as-playbook).

## Test 2b — `"go get my water from the fridge"` re-run under classes

Two contentful roots:

```text
"my water"   -> water-bottle 0.5, water-substance 0.5
"the fridge" -> kitchen-fridge 0.5, garage-fridge 0.5

water-bottle    0.5 -> EP-water-routine 0.50
water-substance 0.5 -> EP-plant-care    0.50
kitchen-fridge  0.5 -> EP-water-routine 0.25, EP-fridge-fetch 0.25
garage-fridge   0.5 -> EP-fridge-fetch  0.50

EP-water-routine 0.75    EP-fridge-fetch 0.75    EP-plant-care 0.50
```

**The tie is back** — at class level this time, and raw heat cannot break it. But look
at where the heat comes from:

```text
EP-water-routine   0.75 from TWO roots  ("my water" via bottle + "the fridge" via kitchen)
EP-fridge-fetch    0.75 from ONE root   ("the fridge" twice — its own two candidates)
```

**Finding 3 — root diversity is the convergence metric.** A root's candidates are
*alternatives* — mutually exclusive readings of the same token. They can pile heat on
a class but can never corroborate each other. Corroboration — constructive
interference — is only ever *across* roots. Stated in the mechanism's own QM
vocabulary:

> **Superposition lives within a root; interference only happens across roots.**

Rank classes by distinct-root count first, heat second: EP-water-routine (2 roots)
beats EP-fridge-fetch (1 root). Collapse: {my water → water-bottle,
fridge → kitchen-fridge} — same answer as test #1, now derived at class level. This
settles test #1's open question 1 and hardens finding 2 (single-root spikes are
artifacts) into a rule rather than a heuristic.

## Consequences

1. **Signed representation:** still not needed at either bookend-to-bookend distance
   tested. It would now take a case of *cross-root contradiction converging on one
   class* to motivate it. Park it.
2. **Wave shape amendment:** `converge` operates at class level — rank by
   (distinct roots, heat), read class statistics, descend to exemplar. Span-walk
   demotes to the descent step.
3. **The remaining hand-wave is segmentation** — both toys hand-assigned members. The
   derived boundary/membership rules (renewal gaps; co-occurrence signature) are now
   the critical path to implementation, ahead of any code for the wave itself.
4. **Sub-span counting held up:** statistics per resolution-act within members (A4
   counted against kitchen inside a span that also contains A5 for garage) worked
   without member = attempt. Keep member = span until it breaks.

## Proposed paper test #3

> **Run same day** — [observe-paper-test-3.md](observe-paper-test-3.md). Outcome:
> compatibility filtering handles it; the signed middle is retired and phase is
> defined as joint-precedent co-occurrence at collapse.

Force the parked decision or retire it: construct *cross-root contradiction* — two
roots whose candidates converge on the same class but bind incompatibly (e.g. "put my
water back in the fridge" said in the garage). If compatibility filtering at collapse
handles it, the middle stays non-negative for good; if the wrong class wins on
(roots, heat), destructive interference earns a mechanism.

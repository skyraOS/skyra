# Paper Test #3 — Cross-Root Contradiction and the Phase Slot

**Date:** 2026-07-20
**Tests:** the parked signed-representation decision, per
[test #2](observe-paper-test-2.md) §Proposed paper test #3.
**Verdict:** destructive interference as *signed amplitude* is retired — three tests,
never needed, and even the contradiction case has nothing to cancel. What the
contradiction actually forces is the **phase slot**: compatibility = witnessed joint
co-occurrence of the committed binding set, checked at the collapse bookend. Novel
stitches stay legal but flagged. Two new root findings: context is a root, and
relational/verb tokens ("back") are load-bearing roots whose omission flips the
outcome.

## Setup

Spine extended past m18:

```text
19   A9    skyra follows michael to the garage           ungraded
20   O11   [in garage, michael holding the blue bottle]
           "put my water back in the fridge"             the impulse
```

One observation carries utterance *and* situation — the location context is witnessed
inside the impulse observation itself, not annotated onto it.

New lens entries:

```text
garage-place         -> {EP-fridge-fetch}    (both garage fetches happened there)
put-back verb-class  -> {EP-water-routine}   (A2: return bottle to kitchen-fridge)
```

Class statistics carried forward: EP-water-routine fridge->kitchen 2/2;
EP-fridge-fetch fridge->kitchen 0/1, fridge->garage 2/2.

## The roots

```text
"my water"     -> water-bottle 0.5, water-substance 0.5
"the fridge"   -> kitchen-fridge 0.5, garage-fridge 0.5
location       -> garage-place 1.0            (witnessed, unambiguous)
"back"         -> put-back verb-class 1.0     (variant B only)
```

Equal-split discipline kept even though michael is visibly holding the bottle —
priors belong to the ingest bookend and are not what this test measures.

## Variant A — without the "back" root

```text
water-bottle    0.5 -> EP-water-routine 0.50
water-substance 0.5 -> EP-plant-care    0.50
kitchen-fridge  0.5 -> EP-water-routine 0.25, EP-fridge-fetch 0.25
garage-fridge   0.5 -> EP-fridge-fetch  0.50
garage-place    1.0 -> EP-fridge-fetch  1.00

EP-fridge-fetch    1.75   roots {fridge, location}   = 2
EP-water-routine   0.75   roots {water, fridge}      = 2
EP-plant-care      0.50   roots {water}              = 1
```

Root diversity ties (2 = 2); heat breaks it → **EP-fridge-fetch wins**. Class
statistics say fridge → garage (2/2). But now the collapse must instantiate *all*
roots, and EP-fridge-fetch cannot: **no member span contains any "my water"
candidate.** The water binding has to be fetched from EP-water-routine, and the
committed joint reading

```text
{my water -> water-bottle, fridge -> garage-fridge}
```

is a **stitch across classes with zero witnessed joint precedent** — no span anywhere
contains bottle and garage-fridge together.

**Finding 1 — diversity, coverage, and joint precedent are three different things.**
Test #2's metric (distinct roots, heat) counts roots that *heated* a class. This test
separates two more properties it cannot see: whether the winning class can *cover*
(instantiate) every root, and whether the final binding set has *joint precedent*.
EP-fridge-fetch won diversity+heat at 2-of-3 coverage, forcing a zero-precedent
stitch.

**Finding 2 — nothing needed to cancel.** The kitchen path and the garage path arrive
at *different candidates*; there is no shared outcome at which signed contributions
would destructively interfere. Signed amplitude is mechanically inert even in the case
constructed to demand it. After three tests: **retire the signed middle.**

**Finding 3 — the contradiction is a collapse-time fact, and it defines phase.** The
QM table's open row — *phase = relational compatibility of bindings (Combine rule;
open)* — gets its definition:

> **Phase compatibility of paths = their bindings co-occur in some witnessed span.
> Constructive interference = joint co-occurrence > 0. "Destructive" = a committed
> binding set with zero joint precedent — not cancellation mid-wave, but a novelty
> flag at the collapse bookend.**

Novel stitches must stay legal (every first time is unprecedented — the mechanism has
to be able to commit one), but they commit *knowing* they are novel. Zero joint
precedent is exactly the signature that should make the express bookend prefer a
clarifying act ("the kitchen fridge where it lives, or this one?") — which is itself
an ordinary act, graded, whose answer becomes the founding span of the new joint
precedent. The bookend discipline holds: observe delivers readings with their
precedent counts; express decides act-vs-ask.

## Variant B — with the "back" root

```text
put-back 1.0 -> EP-water-routine 1.00

EP-water-routine   1.75   roots {water, fridge, verb}  = 3
EP-fridge-fetch    1.75   roots {fridge, location}     = 2
```

Heat ties exactly; root diversity breaks it → **EP-water-routine**, collapse
{my water → water-bottle, fridge → kitchen-fridge}, full coverage, joint precedent
2/2 (A1, A2). The human answer — "back" means back to where it lives.

**Finding 4 — relational/verb tokens are load-bearing roots.** "Back" is not an
entity, but it roots into an action-class lens and is precisely the token carrying
return-to-origin semantics. Dropping it (variant A) flips the outcome to a
zero-precedent stitch. Roots are *independently resolvable references* of any kind —
entities, places, verb classes — not entity mentions only.

**Finding 5 — context is an ordinary root.** The witnessed situation (in the garage)
participated in the wave identically to linguistic roots: same lens-hop, same
conservation, same seat at the diversity count. No special context machinery.

## Consequences

1. **Signed representation: retired.** Three tests, zero need. The honest-divergence
   note in the 07-19 doc resolves: non-negative middle is not a compromise, it is the
   design. Remove it from the open-decisions list.
2. **Phase/Combine rule: settled in shape.** Compatibility = witnessed joint
   co-occurrence of committed bindings; computed at collapse from spine facts (kind 2
   events per span); zero = novelty flag on the resolution event.
3. **Collapse gains one obligation:** instantiate every root; record per-binding
   source class and the joint-precedent count on the commit. The count is what the
   express bookend and later grading read.
4. **Convergence metric amendment:** (distinct roots, heat) ranks classes, but
   coverage gates commitment — a class that cannot instantiate all roots cannot be
   sole precedent; stitching across classes is what triggers the joint-precedent
   check.
5. **The toy's ties are constructed**, deliberately — each test builds the tie that
   makes the next metric earn its place. Live traces won't tie exactly; the *ordering*
   (diversity before heat, coverage as gate) is the durable result.

## What remains before implementation

The segmentation rules are now the only hand-wave left in the whole mechanism
(boundaries by renewal gaps, membership by co-occurrence signature —
[episode-as-region](episode-as-region.md)). Every other slot has a paper-validated
rule: wave shape, conservation, class statistics, tie-breaking, coverage gate,
novelty flag. Paper test #4, if wanted, should hand-derive segmentation itself: take
the 20-moment toy spine *unsegmented* and check that renewal gaps + co-occurrence
recover the three classes.

> **Run same day** — [observe-paper-test-4.md](observe-paper-test-4.md). 4/4
> boundaries, 3/3 classes recovered mechanically; the mechanism is fully
> paper-validated. Next grade comes from a live trace.

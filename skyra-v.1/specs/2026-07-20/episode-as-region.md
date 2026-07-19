# Episode as Region — the Holder of Similar Spans

**Date:** 2026-07-20
**Status:** Ground decision for the observe mechanism. Revises the 07-19 edge-kind
audit; validated on paper by [observe-paper-test-2.md](observe-paper-test-2.md).
**Parents:** [../2026-07-19/observe-mechanism-over-spine.md](../2026-07-19/observe-mechanism-over-spine.md) ·
[../2026-07-14/relational-region.md](../2026-07-14/relational-region.md) ·
[../2026-07-14/region-as-playbook.md](../2026-07-14/region-as-playbook.md)

## The decision

Paper test #1 proved that episode-as-single-span-wrapper is redundant: it moved with
the span layer at every step and its totals tied where spans discriminated. The layer
survives by changing job:

> **A span is one witnessed occurrence. An episode is the class of similar spans — the
> precedent class.**

This is the relational-region result restated on the spine. `relational-region.md`:
*"the state change → action axis is graded over relational regions, not single
instances."* Episode-as-class is the spine-native region; span membership is the
region-identity question (weighted co-occurrence signature above the independence
baseline), already partitioned there into its own process. Region-as-playbook's
corrective framing carries over too: the episode is the playbook, its member spans are
the recorded runs.

## What this changes in the observe mechanism

### 1. The Born rule gets its denominator

Counted settlement statistics need a class to count over. Statistics attach at the
episode and aggregate across member spans:

```text
EP-fridge-fetch:  fridge -> kitchen   0/1 (defeated)
                  fridge -> garage    2/2
```

Without the class, this evidence fragments across spans and the wave only ever reads
the single hottest one (paper test #2 shows the flat model discarding a confirming
member).

### 2. Conservation counts at the episode; descent is optional

If episode amplitude split equally across member spans, a well-worn class with 40
members would dilute to nothing — punishing rich precedent, which is backwards. So:

- amplitude heats the **episode**, and class statistics are read there;
- **descent** into a member span happens only when specifics are needed (the exemplar:
  most recent settled member), and is not a conservation sink.

Converge moves up a level: decide at the class, descend to instantiate.

### 3. The convergence metric is root diversity first, heat second

Class-level ties recur (test #2 reproduces one exactly), and raw heat cannot break
them: a single impulse root splitting across its own candidates can match the combined
heat of two roots agreeing. Distinct-root count is the discriminator — constructive
interference is *several roots* converging, not one root arriving twice. This settles
open question 1 of paper test #1.

### 4. The edge-kind audit partially reverses

07-19 finding 3 ("kind 4 = 3 ∘ 5, deletable") held only for single-span episodes.
Membership adds information beyond co-location, so the revised inventory:

```text
1. O -> A -> O           spine succession            witnessed
2. O -> entity           resolved-into (event)       spine fact
3. entity -> O           resolved-from               derived (= inverse of 2)
4. entity -> episode     participates-in-class       derived, carries membership signal
5. episode -> spans      members (plural now)        derived (co-occurrence signature)
6. A -> O                settled-by                  witnessed
+  episode stats cache   counted settlements         counted (updatable)
```

Kinds 3 remains rebuildable from 2. Kinds 4/5 remain derived and rebuildable — the
mutability table still holds — but they are no longer redundant.

## Epistemic placement

Membership is interpretation. It is built off-wave as a revisable derived index
(the mutability table's "interpreted/counted" rows), so the bookend discipline —
mechanics only between the two collapses — survives intact. Being wrong about
membership stays cheap: reassigning a span to another class rebuilds indexes and
recounts statistics; the spine never moves.

## How segmentation stands today

Asked directly during this session: *how are we currently segmenting spans/episodes?*

**In the built system (dnav2):** a per-turn LLM judge at ingest/midlevel time — the
**edit/new decision** of active-grouping.md. *Edit* = the latest message extends the
active episode (gist updated, episode continues); *new* = topic shift (create a
grouping, switch active). A catch-all default episode catches deposits before midlevel
has framed anything (`genes_ingest.go`, `ensureDefaultEpisode`). So segmentation is
currently **interpreted at write time, by inference, with boundary = topic shift** —
the doc already notes a cheap similarity gate could front-run the judge.

**In the observe-over-spine design:** episodes must be *derived, rebuildable indexes*
— but the 07-19 doc gives no segmentation rule, and both paper tests hand-segmented
their toys. This is the gap.

**What episode-as-region says the rule should be:** dnav2's single edit/new act
conflates two questions that the region partition separates:

```text
span boundary   where one occurrence starts/ends on the spine
                -> BEHAVIOR signal: the renewal profile. a span closes when its
                   region goes dormant / its targets close (closure-duration-
                   objective); a gap in renewal is a boundary.
membership      which spans belong to the same episode class
                -> IDENTITY signal: co-occurrence signature above the
                   independence baseline (relational-region-identity).
```

Both are computable off-wave from spine history — no LLM judge in the durable path.

> **Corrected 2026-07-20 (v3 keying):** the edit/new judge is v2 history, not a
> component to preserve — v3 has no judge in the loop at all (grades are computed;
> the decomposer at ingest is the one open interpretation site). Paper test #4
> subsequently showed segmentation fully mechanical, so nothing needs a proposer.
> See [settlement-set.md](settlement-set.md) §Keying note.

## Open refinements

1. **Sub-span attempts.** E2's single span contains both a defeated kitchen resolution
   and a successful garage one. Test #2 counted statistics per resolution-act *within*
   members, with member = span. Whether attempts should be first-class members is open.
2. **Signature drift.** When does a new span found a new class vs. join and blur an
   existing one? This is relational-region-identity's boundary question; not resettled
   here.
3. **Exemplar choice.** Most-recent-settled is the working rule; recency vs. highest
   grade is undecided.

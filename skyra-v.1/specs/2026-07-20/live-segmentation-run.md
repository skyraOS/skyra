# Live Segmentation Run — the Rules Meet a Real Trace

**Date:** 2026-07-20
**Tests:** paper test #4's segmentation rules against the 2026-07-13 live session
(`logs/traces/*.jsonl`, nine impulses, 11:03–11:32; wall clock reconstructed as
file-mtime − max tMs; signatures from the decomposer's resolved referents).
**Verdict:** the rules survive contact with reality, but compose differently than the
toy suggested: **rule 1 over-segments at fine scale and rule 2 merges same-class
spans** — that composition, not threshold tuning, is the stable design. The bottleneck
is not segmentation; it is canonical entity resolution, and v3's exact-triple layer
already solves it. Method note: replay script in the session scratchpad; rules are
~60 lines of arithmetic, no inference calls.

## What the rules recovered

Six spans, matching conversational reality (utterance labels abridged):

```text
S1  11:03–11:04   greeting                     «skyra hello how are you?»
S2  11:08–11:11   harness / where-am-I         «testing your harness … tell me where I am»
S3  11:17–11:18   move-to-me                   «nice can you move to me?»
S4  11:21–11:24   didn't-move / look-first     «you didnt move … look where you're going»
S5  11:24–11:26   movement postmortem          «why didnt you look again before moving?»
S6  11:30–11:32   the maze                     «complete the maze below?»
```

The zero-gap continuation turns (3, 6, 9 — skyra's follow-through impulses with no
utterance) folded into their parent spans automatically. Rule 2's pairwise weights
then found exactly one class merge:

```text
S4·S5 = 1.79  {look, move}     ← the movement-postmortem class. correct.
S3·S4 = S3·S5 = 0.69  {move}   ← below the two-term threshold (see bottleneck)
all others ≤ 0.41 ('ask')
```

**The promiscuity principle fired twice on real data, unprompted:** michael and skyra
(present in every span) weigh 0.00, and 'ask' (5 of 6 spans) weighs 0.41 — the
ubiquitous verb self-discounted exactly as the toy predicted.

## Finding 1 — the gap spectrum inside one sitting is two-band, not three

The toy had minutes/hours bands; a single session has `{0s}` (continuation) vs
`{57–329s}` (exchange gaps) and nothing above. Test #4's `k × median` rule fails
naively (median lands in the upper band → one span). The fix that worked: baseline =
the **finest witnessed rhythm** (within-turn moment deltas), which makes rule 1
eagerly cut every exchange boundary — over-segmentation — and rule 2 merge what
belongs together. **Over-segment then merge is the composition.** Spans stay small
witnessed occurrences; classes do the grouping. This also re-confirms the hierarchy
question (activity < session < day) as real: this whole trace is one span at day
scale.

## Finding 2 — the bottleneck is canonicalization, and it is already solved in v3

S3 («move to me») failing to join the movement class is not a segmentation failure:
the v2 decomposer emitted 'move' / 'that move' / 'skyra's earlier move' as three
strings for one referent, and crude stemming only partially repairs it. This is
precisely v2's deferred term→entity merge — and precisely what v3's **exact layer
does natively**: world-addressable `Triple.Key()` identifiers are canonical by
construction, and retrieval.md's posting lists index them exactly. Replaying a *v3*
trace would feed rule 2 canonical keys with no repair step. The interpreted layer
rides on top, capped-fidelity, per observation.go.

## Finding 3 — timestamps

Wall-clock was recoverable only by archaeology (file mtimes minus tMs). v3's
`Observation.At` is a tick; the substrate delta from test #4 stands: **wall-clock on
the observation node, witnessed.**

## The v3-or-v4 question

**v3. Not v4.** The grounds:

1. **The wave already has a home in v3.** retrieval.md declares itself "the mechanism
   inside `observation → state change`" and is already posting-lists + set operations,
   no judge. The observe mechanism is that spec *matured* — lens = posting lists
   grouped by class; the wave = the set operation extended with root diversity,
   coverage, and joint precedent. Same representation, more mechanism.
2. **Version numbers here mark representation breaks.** v2→v3 happened because
   judge-graded free text hit a ceiling. Nothing in paper tests #1–#4, the ground
   decisions, or this live run breaks v3's representation — every addition is a
   derived, rebuildable index (spans, classes, lenses) plus one field (wall-clock),
   and v3's vocabulary explicitly reserves the derived-index slot.
3. **The live run's one real blocker is fixed by v3**, not by anything a v4 would add
   (Finding 2).
4. **v3 is half-landed and the plan says "archive v2 once v3 stands."** A v4 label now
   would abandon an aligned substrate mid-build and re-derive it under a new name —
   cost, no representational payoff.

Reserve v4 for an actual representation break, if one ever arrives.

## Landing steps (in order)

1. ~~Wall-clock on `Observation`~~ — **landed 2026-07-20**: `Observation.Wall` (unix
   ms), stamped by the one writer (`ObservationGraph.Write`) when zero, preserved when
   preset (replay carries historical time).
2. ~~A `segment` package~~ — **landed 2026-07-20** as `segment.go` in dnav3:
   `Spans` / `Classes` / `Lens`, pure functions, no inference. Paper test #4's toy is
   the unit-test fixture (`segment_test.go`) and the live run's composition finding is
   `TestOverSegmentThenMerge`. One estimator correction found by the tests: the rhythm
   baseline is the **lower quartile** of positive deltas, not the median — a median
   lands in the wrong band when boundary gaps rival intra-deltas in count.
3. Lenses as class-grouped posting lists — `Lens` gives the structure; wiring edge
   EMAs to classes as the class statistics remains.
4. The wave on top of retrieval's set operation — **specced 2026-07-20**:
   [the-observe-wave.md](the-observe-wave.md). Both integration points decided
   there: the resolution commit IS the state-change node (giving observation.go's
   tabled feedback loop its grading path), and workers see ranked classes + exemplar
   + flags in the present, never a heat map.
5. Replay the next live v3 session through 1–3 before 4 exists.

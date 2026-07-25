# Episode Names Two Different Things

**Date:** 2026-07-22
**Status:** DECIDED, pre-implementation. Rename not yet applied; blast radius measured
against `src/skyra/cognitive-agent/skyra-dna-v3` at `9693df3`.

**Reads on:** [Episode as Region](../2026-07-20/episode-as-region.md),
[Live Segmentation Run](../2026-07-20/live-segmentation-run.md), and
`src/skyra/cognitive-agent/skyra-dna-v3/segment.go`

---

## The collision

`episode` denotes two unrelated things in v3, and they are not reconcilable.

```text
DESIGN (segment.go, episode-as-region.md)
  episode = the CLASS of similar spans
  derived, rebuildable, DISCONTIGUOUS in time, never authoritative

IMPLEMENTED (being.go)
  EpisodeID = a runtime lifecycle boundary around one Realm
  caller-driven, contiguous, authoritative, wipes Obs/Spans/Activation
```

The implemented one took the name and none of the meaning. `NewEpisode(w Realm)` is a
reset: fresh world, fresh observation graph, fresh span index, fresh activation field,
sequence counters back to zero. The only thing that survives is `Expr`. That is a RUN of
an experiment, not a region of similar experience.

Nothing in the code is wrong. The word is.

## The axes, stated once

The two segmentation rules answer different questions off different signals, and
conflating them was v2's edit/new trap — one LLM judgment doing both jobs at write time.

```text
SPAN       TEMPORAL     one witnessed occurrence      renewal gaps in WALL time
CLASS      MEMBERSHIP   which occurrences are the     co-occurrence of resolved
                        same KIND of thing?           identity above baseline
```

Derivation runs one way, and the direction is load-bearing:

```text
observation log
    │ cut by wall-time gaps        spans exist FIRST, on their own axis
    ▼
spans
    │ merge by shared-term weight  classes are computed FROM spans
    ▼
classes
```

A span is not a subdivision of a class. A class is a grouping of spans. A span is a VIEW
over the log, not a store.

## The rename

```text
EpisodeID  ->  RunID       the runtime lifecycle boundary
Class      ->  Episode     the derived similarity grouping
```

`Episode` is only free after `EpisodeID` moves out, so the order is not optional.

| current | becomes | note |
|---|---|---|
| `EpisodeID` (type + field) | `RunID` | 116 occurrences |
| `episodeID` | `runID` | 14 |
| `NewEpisode` | `NewRun` | 11 |
| `startEpisode` | `startRun` | 3 |
| `episodeNumber` | `runNumber` | 3 |
| `Class` (type) | `Episode` | 28 |
| `Classes()` | `Episodes()` | 8 |
| `ClassHeat` / `ClassEdges` | `EpisodeHeat` / `EpisodeEdges` | 6 / 4 |
| `HeatClasses` | `HeatEpisodes` | 3 |
| `classOfSpan` / `byClass` | `episodeOfSpan` / `byEpisode` | 3 / 7 |

Plus four test names in the episode family and two in the class family.

### Do not touch

`classifySource` and the five `TestClassifySource*` cases are a DIFFERENT WORD FAMILY —
classify, not class. They live in `activation_field.go` and concern ambiguity resolution
over target sources. A careless `class -> episode` sweep corrupts them.

The class family is otherwise confined to four files: `segment.go`, `wave.go`,
`segment_test.go`, `wave_test.go`. The episode family is spread across the package.

### Open: the ID string prefix

`startEpisode` mints `episode:%d`, and span and observation IDs are built on it:

```text
episode:1/span:2
episode:1/observation:7
```

Renaming the concept argues for `run:%d`. Against: at least seven test sites assert those
literal strings, and any durable artifact or trace already written uses the old prefix.

Decide with the rewrite, not before it. If the identifier rename lands alone, leave the
prefix and record the inconsistency here rather than half-migrating the format.

## What the rename exposes

`Classes()` has NO live non-test callers. `Lens()` has none at all. The only remaining
consumer is `wave.go`, which is itself prototype history and no longer the ingest path.

So the composition the live-segmentation run validated is half-dead:

```text
OVER-SEGMENT (finest rhythm)   ALIVE, and then some — ObserveImpulse mints a span
                               unconditionally, one per impulse, finer than the
                               temporal rule would cut
THEN MERGE                     DEAD. nothing in the live path ever groups two spans
```

Every external impulse starts a new span with no merge behind it. Against a real
conversational interface, three messages typed in succession become three spans, each
owning its own predictions, with no mechanism to recognize them as one exchange. That is
the fragmentation half of Risk 7 — currently unguarded, not solved.

Renaming `Class` to `Episode` does not fix that. It makes the gap legible: the layer that
would fix it exists, is tested, and is unreachable.

## The rule that must survive the rewrite

The temptation, on finding the merge missing, is to merge at write time — let an incoming
impulse join an existing span when it activates one. Do not.

A span is the ownership root for settlement. Joining span S makes an impulse's
observations eligible to settle S's open bets, so a wrong merge does not merely mislabel
an occurrence — IT MANUFACTURES EVIDENCE FOR A BET THE IMPULSE NEVER PLACED. That is the
cross-episode grading defect the stress run found, walking back in through the front door.

Worse, merging on activation is circular: activation is bumped BY the impulse, so "does
this impulse activate an existing span?" reduces to "does it share terms with recent
memory?" — which is what `Class` did, and what the class layer was removed from the
active path for. Rebuilding it at write time makes it irreversible.

The grouping the instinct wants is already correct as a DERIVED READ. State it as a law:

> **Grouping may confer attention, never settlement rights.**

Spans stay fine-grained and own their bets. Episodes (the renamed class layer) group spans
for retrieval, framing, and rendering. Two layers, two jobs. Under that rule grouping can
be as aggressive as it likes, because the worst case is a bad retrieval, and a bad
retrieval decays where a bad merge does not.

## Consequences for the thesis

`RunID` is the honest name for what the measurement iterates. `cmd/dnav3 -reps N` runs the
same task N times with a fresh world and persistent `Expr` — those are runs. The proposal's
"repeated exposure: N episodes of one task class" should read N RUNS of one task class,
and the ablation disables durable edge accumulation across runs, which is a `b.Expr` seam
and touches `Obs` not at all.

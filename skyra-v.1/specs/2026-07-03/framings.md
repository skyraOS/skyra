# Framings — Reify, Save, Compete (Stage C contract)

**Date:** 2026-07-03 · **Status:** in spec · **Parent:** [accumulation.md](accumulation.md) Stage C · **Design source:** [midlevel.md](../2026-06-16/skyra-dna/midlevel.md) "Grouping vs. framing" (2026-06-26)

> **This spec is instance one of a larger pattern:** match/edit/new are the
> **inner world's primitives** — affordances of MemoryCells, mirroring
> perceive/engage as the outer world's. See
> [memory-as-world.md](../../design-docs/cognition/memory-as-world.md). Entity
> resolution is instance two; no generalization until both run.

## The problem

Framings are not saved. Today the gist is a **field** (`Gist` on the grouping's
Content + `g.gist` string on the cell) — not a node. Nothing persists it, nothing
competes with it, and cold-start mints a **new grouping every cycle**: the maze
session produced **42 gists for one situation** (19× "navigate maze together" and
cousins), every one a fresh anchor. That's the empirical case for the warm path in
one number.

## The anchor decision: same groupID, new framing node per re-frame

**One stable groupID per situation; framings are separate child nodes that turn
over on it.** Not a new groupID per framing. The reasons are load-bearing, not
stylistic:

- **`grouping → prediction` edges must stay put** while framings churn
  (working-memory.md loose coupling). A per-framing groupID moves the anchor every
  time the label shifts — every prediction re-hangs, every bet orphans.
- **`setpoint → groupID` tags would dangle.** A setpoint tags the situation it
  belongs to; the situation doesn't become a different situation because MidLevel
  found better words for it.
- **The framing sequence IS the version history** — recent hot, old decaying, no
  separate versioning system, death by decay (midlevel.md). That only works if the
  versions hang off one anchor.

## Node structure

Framing = **an understanding that hasn't crystallized yet** (groupings-and-
understandings.md) — so it is a `KindUnderstanding` node with **no `BirthAlignment`
stamp**. No new Kind: crystallized-vs-not is the stamp, not the type
(working-memory.md "two orthogonal axes"). The same node matures in place.

```
grouping g:N (stable anchor; Content = nil/meta — the gist moves OFF Content)
 ├── g:N→framing:t12   KindUnderstanding, unstamped   Content = Gist{Text, Timestamp}
 │     └── gist tokens (attention roots — move here from the grouping)
 ├── g:N→framing:t40   (a re-frame — the old framing decays, both remain)
 ├── g:N→trigger       (the interaction that started the grouping — situational, stays)
 ├── g:N→other-pred    (per working-memory.md, unchanged)
 └── g:N→self-pred     (unchanged)
```

- **Current framing** = the framing child with the highest activation (its edge
  record: re-selection bumps, abandonment decays; pre-Stage-D, highest timestamp).
- **Gist tokens move onto the framing node** (they are the gist's words); the
  trigger interaction stays on the grouping (it belongs to the situation).
- **Persistence is free:** framings are ordinary `MemoryCell`s in the graph —
  Stage F serializes them with everything else. "Saving framings" = this
  reification; no separate store.

## The wire format — match / rewrite / create

Render the active framings into MidLevel's present as a numbered list (hot-first,
top-K capped; `f#` indices are **cycle-local** — a per-cycle map resolves
`f# → (groupID, framingID)` at parse time, same posture as motor's forms):

```
Active framings:
[f1] navigate maze together
[f2] playful greeting
```

One wrapper, three explicit verbs (parse.go: `ParseFrame`) — same grammar family as
the motor (`<engage><michael>…`) and other's `<prediction p="…">`: verb tags, `id`
as attribute, payload as body. **No colon-sigils, no implicit forms** (match vs.
edit must never hinge on body emptiness — models echo the old gist back, and every
echo would mint a junk framing node):

| Response | Action | Effect |
|---|---|---|
| `<frame><match id="f1"/></frame>` | **match** | re-elect f1 — bump its edge (no new node) |
| `<frame><edit id="f1">leading michael through the maze</edit></frame>` | **edit** | new framing node on **f1's grouping**; old framing left to decay |
| `<frame><new>solve the shared maze</new></frame>` | **create** | new grouping + first framing (the current cold-start path, kept) |

The `<frame>` wrapper namespaces the verbs (`new`/`edit` are too generic to live
top-level once a response ever carries more than one tag) — the same role
`<engage>` plays for its target.

Malformed → retry-with-correction (the motor's proven pattern). A failed call
leaves last cycle's framing standing (current behavior, kept).

## Attribution — framings earn their weight (no self-grading)

Per the Stage A decision: MidLevel never grades its own framing (unfalsifiable —
a framing has no observable of its own). Instead:

- **Bets stamp their framing.** `Prediction` gains a `FramingID`, set at bet time —
  credit goes to the framing that was context when the bet was made, not whichever
  framing is current when the grade lands (re-frames between bet and grade must not
  steal or absorb credit).
- **Grades flow up.** Stage A writes each grade onto the `grouping → prediction`
  edge AND onto the stamped framing's edge (`g→framing:t`) as `Alignment`. Framings
  then compete on **earned predictive success + recency** through the ordinary bump —
  no new mechanism.
- **Re-frame pressure is derived:** bad grades under a framing → its alignment
  drops, stress rises → the next match/rewrite call sees a cooling framing. The
  frame-lock-in seam (ancestor map #3) gets its first mechanism for free.
- **The fast path — grade-triggered re-framing** (two-stroke-cycle.md): a
  DRAMATIC failure (total miss / hard execution failure) doesn't wait for the
  slow competition — the grading agent spawns MidLevel's framing pass directly,
  failure in the present, one reframe per impulse. Slow path = attribution and
  decay; fast path = ascent. Same pressure, two timescales.

## Crystallization — the framing's real grade, at closure

At Stage B closure the **current framing node is the thing that crystallizes**:
`StampBirth(closureGrade, stressAtBirth)` on the framing → it becomes the durable
Understanding (same node, matured — content already in place, tokens already
attached). Framings that never see a closure decay out as unstamped history.
Lifecycle: **compete while live (selection + inherited alignment) → stamped at
closure → or decay.** Graded exactly when there's something honest to grade with.

## Changes

- [ ] reify: `startGrouping` → grouping anchor (Content off) + first framing child node; gist tokens onto the framing
- [ ] `currentFraming(grouping)` read (activation-first, timestamp fallback) — replaces `g.gist` and the agents' `Content().(Gist)` reads
- [ ] render `Active framings: [f#] …` into MidLevel's present (hot-first, capped); cycle-local `f#` map
- [ ] `ParseFrame` in parse.go: `<frame><match id/>` / `<frame><edit id>body</edit>` / `<frame><new>body</new>` + retry-with-correction
- [ ] `Prediction.FramingID` stamped at bet time (other + self)
- [ ] Stage A grades also write `Alignment` onto the stamped framing's edge
- [ ] closure stamps the framing node (`StampBirth`) — the gist→Understanding move becomes literal

## Open

- **rewrite-vs-create ambiguity** (the clustering problem): the model will sometimes
  create where it should rewrite. Decay is the cleanup (a duplicate grouping that
  stops being selected fades) — accept the slop, watch the rate in the logs.
- Cap for the rendered framing list (start ~5; the maze session suggests one live
  situation dominates anyway).
- Whether match should also allow a **probability/confidence** later (a soft
  re-election) — not now.
- Cross-grouping rewrite (moving a framing between anchors) — explicitly not
  supported; that's what create + decay handles.

## Done when

- A second cycle **matches** instead of creating (the 42-gists-one-situation
  pathology gone from the logs)
- A rewrite births a second framing node on the same anchor; both visible in the
  graph; the old one decays over subsequent cycles
- A prediction's grade lands on the framing that was active at bet time
- A closure stamps the framing node and it survives restart (with Stage F) as a
  durable Understanding

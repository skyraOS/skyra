# Perceive Build — Uncertainties to Resolve Before/While Building

**Date:** 2026-07-07 · **Status:** ✅ RESOLVED (same day) — all 8 items decided; two
threads deliberately tabled for the motor-independence work (1b's continue-vs-local-motor
revisit, 5's correlation id). Items are ordered by how much they blocked.
**Parents:** [perceive-internalization.md](../../design-docs/cognition/perceive-internalization.md)
(the build this de-risks), [affordance-record.md](../../design-docs/being/affordance-record.md),
[prediction-motor-and-closure.md](../../archive/design-docs/prediction-motor-and-closure.md),
[slice-based-dispatch.md](../2026-07-03/slice-based-dispatch.md) (the workflow DAG these
decisions land in).

Everything here came from checking the 2026-07-07 design docs against the live dnav2 code.
The deposit mechanics themselves (bridge translation, ingest gene, default grouping,
normalization pass, self-entity) are **not** in question — they map cleanly. These are the
places the specs are silent or the code contradicts them.

---

## 1. Routing after a perceive ingest — where does the wave go?

**Blocks:** the whole perceive build (it's a workflow-DAG row).

**The gap.** perceive-internalization §5 says a perceive-return "enters the pipeline as an
internal event" and routes to a **sensory ingest**, not `self:grade`. The transition table
(`workflow.go`) has no vocabulary for this: `classify` (`genes_sensory.go`) returns
`reafference` for *any* `Return` impulse, and reafference routes to `self:grade`. Two
sub-decisions:

**1a. The condition split.** `classify` must fork the reafference lane by primitive:
`move`/`engage`-return → `self:grade` (grade my action's execution), `perceive`-return → the
ingest. New condition name (`perceive-return`?), new DAG row.

**1b. After the deposit, continue or stop?**
- **(A) Continue → `midlevel:frame`.** Every perceive spawns a full cognitive turn — the
  self-driven navigation loop (move→perceive→move…), matching the maze behavior and "enters
  the pipeline." Cost: ~4 inference calls per glance.
- **(B) Stop → Done.** Deposit only; the perception waits lit on the graph for the next
  impulse to traverse it. Cheaper; the being only *thinks* when engaged or when its own
  action-grading loop drives it.

These are very different beings. Lean: **(A)** — the workflow already predicts the
reafference loop bounded by `midlevel:empty`, and (B) quietly kills self-driven navigation.
But cost per glance is real; maybe (A) with the ingest allowed to report a cheap "nothing
changed" condition that short-circuits to Done.

**Decision (2026-07-07):**

- **Classification corrected: a perceive-return is a REAFFERENCE, not an internal event.**
  `internal` is reserved for **interrupts** — e.g. motor engaging the self
  (`<engage><skyra/self>`). The perceive-return is the world answering my own act; it stays
  on the reafference side of the fork. What distinguishes it from an action-reafference is
  only **where the record lands** — the perceive-return's job is depositing the
  `PerceiveRecord` in the proper spot on the spine (the gist-owned chain), which is what
  perceive-internalization outlines. (perceive-internalization §5's "enters the pipeline as
  an internal event" is a wording error — correct it to *reafference* when building.)
- **1a: yes** — `perceive-return` is a new condition + DAG row. The reafference lane forks
  by primitive: `move`/`engage`-return → `self:grade`, `perceive`-return → the ingest.
- **1b: (A) continue for now — tabled for revisit.** Continue into `midlevel:frame` today.
  This will change: with location now *recorded* on the spine, motor can act somewhat
  independently (the prediction-motor bound direction — motor riding its local world slice
  without re-running the full brain per glance). Revisit when that lands; don't build the
  short-circuit optimization yet.

---

## 2. Where do predictions live in the new spine?

**Blocks:** the gist restructure (item 3), the bound persistence in
prediction-motor-and-closure, four load-bearing readers.

**The gap.** The §3 spine diagram — `grouping → gist → record → entity` — omits predictions.
Today predictions hang **directly off the grouping**, and four readers scan
`grouping.Children()` for them: `LatestUngradedOther`, `LatestUngradedRealizedSelf`,
`readSelfProposal`, `Snapshot` (all `nucleus.go`). Meanwhile:

- affordance-record.md lists `"predict"` as a Kind of AffordanceRecord, and
- the gist-ownership logic ("interactions belong to the framing they arrived under") applies
  at least as strongly to a prediction — a bet *made under* gist-1 was made under that framing.

**Options:**
- **(A) Predictions move to the gist** (uniform spine: the gist owns everything that arrived
  under it). All four readers change to walk `grouping → gists → predictions`. Consequence:
  a gist edit strands the open bet under a decaying prior gist — arguably *correct* (the bet
  was conditioned on the old framing; a re-frame is grounds to re-predict), arguably a bug
  (the bound should survive a sharpen — prediction-motor wants the bound "written onto the
  active grouping" and graded at closure, and an edit is the *same* episode continuing).
- **(B) Predictions stay on the grouping** (mixed spine: records are gist-owned, predictions
  are grouping-owned). Readers unchanged. The grouping = the prediction unit (as
  slice-based-dispatch already says: "the grouping is only the prediction unit"); the gist =
  the framing/record unit. The mix is principled if we can say *why* in one line: records are
  evidence (framing-scoped), predictions are the episode's open state (episode-scoped).
- **(C) Split by edit-vs-new:** predictions survive an `edit` (same episode), die with the
  episode. This is (B) — an edit keeps the grouping — plus the observation that `new`
  already strands nothing (new grouping, fresh bets).

Lean: **(B)/(C)** — smaller blast radius, matches "grouping = prediction unit," and the
edit path already preserves exactly the continuity the bound needs. But it must be written
down as a rule, not left as an accident of which reader was cheapest to keep.

**Decision (2026-07-07): (A) — gist-owned. Predictions are a record Kind
(`AffordanceRecord{Kind:"predict"}`), and the spine is uniform: the gist owns everything
that arrived under it.** Re-predict-on-edit is a **feature**, not a bug — a bet was
conditioned on the framing it was made under; a re-frame (edit → new gist) is grounds to
re-predict. Consequences to build:

- The spine is `grouping → gist → {record, prediction} → …` — predictions sit beside
  perceive/engage records as children of the gist they were made under.
- The four readers (`LatestUngradedOther`, `LatestUngradedRealizedSelf`,
  `readSelfProposal`, `Snapshot` callers) walk `grouping → gists → predictions` — one level
  deeper. Grading still closes bets under **prior** gists (evidence can grade a bet made
  under the old framing; the readers scan all gists, not just the active one). New
  predictions land under the **active** gist.
- A stranded bet (under a decayed prior gist) fades with its framing — no cleanup logic.
- Migration per affordance-record.md: `dna.Prediction`/`KindPrediction` collapse into
  `AffordanceRecord{Kind:"predict"}` is a follow-on, not same-day; the *placement* (gist
  ownership) is the same-day change.

---

## 3. The gist-as-child restructure — active-gist tracking + who sees prior gists

**Blocks:** perceive-internalization §3 (the three listed deltas are necessary but not
sufficient).

**The gap.** `grouping.Content().(dna.Gist)` is read in more places than the edit path:
`OpenEpisodes`, `Latest()` (`nucleus.go`), the settle's `gistOf` (`settle.go`), and
`renderCellText` (`frame.go` — renders *any* Gist as `"Framing: …"`). Two concrete holes:

**3a. How is the active gist tracked?** Options: a per-grouping active-gist pointer on the
`GroupingIndex` (explicit, one more piece of mutable index state) vs. max-timestamp
`KindGist` child (derived, no state, one scan). Lean: **pointer** — it's read on every
deposit and every settle; and "active" is already an explicit index concept
(`GroupingIndex.active`), so this matches.

**3b. Prior gists leak into the frame and the settle.** With gists as children:
- a traversal renders **multiple `"Framing: …"` lines** in one present (the sequence of
  prior gists all render), and
- `collectSettleItems` would hand the grader **the framing as an item to grade against
  itself** (today it escapes only because the gist is the root's `Content`, and the root is
  skipped).

Needed rules (proposal): the **active** gist renders as `Framing:` and is the settle's
truth; **prior** gists render as nothing (or a one-line `earlier framing:` history entry —
pick one); **`KindGist` nodes are excluded from settle items** unconditionally. Also pin:
does the settle traverse from the grouping (skipping gist nodes as items but descending
through them to their records) — required, since the records now hang *under* gists.

**Decision (2026-07-07):**

- **3a: on the SLICE.** `activeGroupingID` + `activeGistID` live on the being's slice —
  they are *active state*, and the slice is the canonical home for active state (same as
  the surroundings map and stress; §5b already says "when the slice has no active
  grouping"). Pins: **(1)** this forces the being's one nucleus slice to become real
  (dnav2 currently passes `nil` — `NewEntityIndex(nil)`); **(2)** store **IDs, not
  pointers** — the slice is kernel-side, readers resolve ID → node through the index. No
  kernel change: it's a cognition map via the existing `Slice.Maps()`, kernel oblivious.
  `GroupingIndex.active` migrates to reading/writing the slice map.
- **3b: prior gists ARE allowed in the frame — that's the point.** A prior gist is a
  previous representation of what the agent thought the situation was about; the agent
  reorganizes around the new gist, bringing the perceive/engage records over under the new
  framing — "before I thought x, now I think y because of z." So prior gists render as
  frame lines, and the **settle grades them like any other item** against the active
  framing: a superseded framing in tension gets pruned/decays; one still coherent survives
  as history. The trajectory falls out of the economy. **The only exclusion rule: the
  ACTIVE gist is never a settle item (never graded against itself)** — it renders as
  `Framing:` and is the settle's truth. Priors render distinctly (e.g. `earlier framing:`).

---

## 4. Perceive-only turns frame nothing — should perceive records enter the framing present?

**Blocks:** the maze / any non-conversational world; the default grouping's §5b lifecycle.

**The gap.** `renderExchange` (`genes_midlevel.go`) only reads `InteractionRecord`s. If a
perceive deposit lands and (per item 1) the pipeline continues to `midlevel:frame`, the
exchange is empty → `"empty"` → Done. In a world with no conversation, midlevel never
frames and the default grouping is never sharpened — but §5b's lifecycle *depends* on
midlevel eventually sharpening it.

**Options:**
- **(A) Perceive records render into the framing present** ("[t12] perceived: michael at
  2,1 · self at 5,3") — midlevel can frame a purely spatial episode ("exploring the maze,
  michael is north of me").
- **(B) Perceive turns skip framing by design** — framing is conversational; spatial state
  is only ever *context* for frames created by exchanges.

Lean: **(A)** — the maze run already demonstrated framing-worthy spatial episodes, and (B)
makes §5b's default-grouping lifecycle unreachable in embodied worlds.

**Decision (2026-07-07): (A) — yes. Perceive records are the locations of entities; they
must be displayed.** Placement + mechanism pinned:

- **Displayed at the TOP of the frame** — the normalization pass gives the frame this
  structure before the present is derived.
- **The mechanism is display-normalization over the TRAVERSED set — not a graph query.**
  The traversal delivers all the nodes; the normalization pass then decides what to
  *display* out of those nodes (per entity: the most recent location). It runs after
  traversal, before present-derivation.
- **Normalization is NOT grading.** It writes nothing — no A, no F, no economy effect. It
  is a deliberate cheat to gain control over the frame and give it structure (the §2
  shortcut, now generalized to frame layout: locations pinned at top, deduped
  freshest-per-entity).

---

## 5. Reafference linkage — which action provoked this return?

**Blocks:** the causal `prev` chain on the board; correct self-grading under two in-flight
actions.

**The gap.** slice-based-dispatch wants the world's reply linked to its causal predecessor,
but `dnav2Being.Observe` (`cmd/tui/being.go`) threads `prev` as "the last inbound engage,"
and a perceive-return arriving async carries nothing identifying the action that fired it.
Today's grading already uses the heuristic `LatestUngradedRealizedSelf` (newest ungraded
realized proposal) — fine at one-action-in-flight, silently mis-links at two.

**Options:**
- **(A) Accept the heuristic** for now (single-action worlds), note the limit.
- **(B) Motor stamps an outstanding-action id** (the action node's ID) somewhere the bridge
  can match on the return — but the world's return impulse carries no such field today, so
  this touches the world's return path (`world/space.go` PerceiveCell.Express / the engage
  return) or adds a correlation field to `skyra.Impulse`.

Lean: **(A) now, (B) when fan-out is real** — but write the limit down where the mis-link
would bite (`genes_self.go` / the ingest).

**Decision (2026-07-07): (A) — the heuristic.** `LatestUngradedRealizedSelf` stays; correct
at one-action-in-flight, which the join currently guarantees (motor fires once per impulse).
Write the limit down at the two sites where a mis-link would bite (`genes_self.go`, the new
ingest). The correlation id (B) lands **with** the motor-independence work (tabled 1b) —
that's what makes overlapping actions real — not before. Note: perceive-returns are
low-stakes here regardless (the ingest deposits locations, it closes no bet); the heuristic
risk lives only on the `move`/`engage` grading lane.

---

## 6. "Amplitude economics" — confirm the token budget is the intended realization

**Blocks:** nothing (a confirmation, not a build decision) — but if wrong, it changes item 3/4.

**The gap.** perceive-internalization §4 speaks kernel-amplitude language (attenuation,
falloff), but the dnav2 frame doesn't use kernel amplitude on this path at all:
`traverse.Assemble` with a **normalized-weight token budget** (`frame.go` — weights
normalized to sum 1 = conservation). Note the mesh hop is **already live**: `LinkGrouping`
makes groupings children of entities, so `FrameOf` can already walk
`grouping → … → entity → other-grouping` and bleed cross-context, budget-bounded.

**To confirm:** budget-share *is* the amplitude stand-in (falloff = shrinking budget share
down the tree), and no second scoping mechanism should be built. If confirmed, one line in
perceive-internalization §4 saying so ("amplitude" on this path = frame budget share).

**Decision (2026-07-07): confirmed.** Budget-share is the intended realization of
"amplitude" on the frame path — falloff = shrinking normalized budget share down the tree.
Build **no second scoping mechanism** (no radius rule, no wall, no separate amplitude
accounting on frame assembly). The already-live mesh hop (`entity → grouping-B` via
`LinkGrouping`) is the §4 bleed, correctly bounded by budget share. Add the one-line
clarification to perceive-internalization §4 at build time.

---

## 7. prediction-motor-and-closure — what is actually buildable now

**Blocks:** nothing in the perceive build; guards against overreach.

**The read.** From that doc, the only near-term buildable piece is **persist the bound**
(predictions stop evaporating — which is item 2) plus **reafference reaching it** (the
perceive build itself). Everything else needs design before code:

- **Concurrent settle-during-motor:** no mechanism exists — the settle runs synchronously
  *inside* `midlevelFrame` (`genes_midlevel.go`). A persistent settle worker or re-trigger
  loop is a structural change to the workflow engine.
- **The nervousness channel:** quantification explicitly open in the doc.
- **Closure's write target:** explicitly open (fourth quantity vs. a channel on prediction
  edges); `activation.go`'s `DistributeCredit` already carries the parked/mis-assigned note.
- **Extent of a bound:** open.

**Proposed scope line:** the perceive build treats prediction-motor-and-closure as
*direction only*; nothing from it lands except what items 1–5 already cover.

**Decision (2026-07-07): confirmed.** prediction-motor-and-closure is direction only for
this build. What lands is already covered: predictions persist on the spine (item 2 —
gist-owned, the bound stops evaporating) and reafference reaches them (the build itself).
Concurrent settle-during-motor, the nervousness channel, closure's write target, and the
extent of a bound all stay design-open, no code — the settle-during-motor work is likely
the same work as the tabled motor-independence direction (item 1b), so they return
together.

---

## 8. AffordanceRecord pins (small)

**Blocks:** the ingest gene's first line of code (a), nothing hard (b).

**8a. Kind vocabulary — typed constants vs raw strings.** `MemoryCell.Kind` is a
typed-constant vocabulary (`memory_cell.go`); affordance-record.md implies raw dynamic
strings. And today's interactions are `KindResponse`, while the docs call the same thing
`Kind:"engage"` — the migration has to reconcile the two names. Proposal: new Kinds enter as
raw strings stamped from the affordance (`Kind("perceive")`), existing constants stay until
their records migrate; `KindResponse → "engage"` is renamed *at* the EngageRecord collapse,
not before.

**8b. Timestamp source under async returns.** Records stamp `DNA.Tick()`; perceive-returns
arrive on world goroutines. Ticks are mutex-safe but cross-wave ordering is unspecified —
the §2 max-Timestamp normalization leans on it. Probably fine (proper time doesn't promise
cross-wave ordering; per-entity freshest-wins only needs *within-entity* monotonicity, which
one mutexed clock gives). Note it; no action expected.

**Decision (2026-07-07):**

- **8a: new Kinds are born as raw strings stamped from the affordance** (`Kind("perceive")`
  — no constant added; the `Kind` type stays, the closed vocabulary opens). Existing
  constants stay until their record migrates; `KindResponse → "engage"` renames **at** the
  EngageRecord collapse, not before.
- **8b: accepted as the note** — the mutexed clock's within-being monotonicity is all the
  freshest-per-entity normalization needs; cross-wave ordering was never promised.

---

## Housekeeping (not design)

- `phase1-ab` / `phase2-accumulate` at the repo root are compiled binaries (~5.6MB each),
  untracked — gitignore them.

## The decided shape (summary)

- **Routing:** perceive-return is a **reafference** (internal = interrupts only, e.g. motor
  `<engage><skyra/self>`). New condition + DAG row; reafference forks by primitive
  (`move`/`engage` → `self:grade`, `perceive` → ingest); ingest continues into
  `midlevel:frame` for now (revisit with motor independence).
- **Spine:** uniform, gist-owned — `grouping → gist → {record, prediction} → entity → …`.
  Predictions are `AffordanceRecord{Kind:"predict"}`; re-predict-on-edit is a feature.
- **Active state:** `activeGroupingID` + `activeGistID` on the being's slice (IDs, not
  pointers, via `Slice.Maps()`); the nucleus slice becomes real.
- **Prior gists:** legitimate frame items, graded by the settle like anything else; only
  the active gist is exempt (never graded against itself; renders as `Framing:`).
- **Frame layout:** normalization = display-only pass over the traversed set (not a query,
  not grading); freshest-location-per-entity pinned at the top of the frame.
- **Linkage:** `LatestUngradedRealizedSelf` heuristic stands; correlation id lands with
  motor independence.
- **Scoping:** budget-share IS amplitude on the frame path; no second mechanism.
- **prediction-motor-and-closure:** direction only for this build.
- **Kinds:** new Kinds born as raw affordance strings; renames at collapse time.

## Tabled (returns with motor independence)

- 1b revisit: motor riding its local world slice instead of every perceive spawning a full
  turn — likely the same work as concurrent settle-during-motor + the nervousness channel.
- 5(B): the action-id correlation field on the impulse/return path.

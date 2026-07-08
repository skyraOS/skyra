# Perceive Internalization — Location as a Hot Leaf, Scoped by Grouping

**Date:** 2026-07-07 · **Status:** Spec (concrete enough to build; open questions marked).
**Revises:** [entities-and-surroundings.md](entities-and-surroundings.md) (the single global
surroundings map dissolves), [working-memory.md](working-memory.md) (the traversal-chain
direction flips — see §6). **Composes with:**
[prediction-motor-and-closure.md](../architecture/prediction-motor-and-closure.md) (the bound),
[emit-as-traversal.md](../kernel/emit-as-traversal.md) (the world-cell neighborhood),
[operator-runtime-and-interrupts.md](../architecture/operator-runtime-and-interrupts.md) (the
local world slice).

## The problem

Skyra can *decide* to perceive and the world faithfully runs it, but nothing she perceives
re-enters cognition. The emit half works; the whole return half is stubbed at three walls
(§5). This spec fixes internalization **and** settles the topology the perceived location
lands in — because the two are the same question: *where does a location go, and who can
reach it.*

## 1. There is no global surroundings map

"Surroundings" was a stand-in for **hot location-edges on entities**. It is not a store the
being reads. Perception writes onto the shared graph; a being's "what's around me" is
whatever hot locations its **currently inhabited grouping** can reach. Motor never reads a
union — it reads its grouping's slice. The map dissolves the same way the working-memory
"buffer" did: into *the lit subgraph, reached by traversal*.

## 2. A perception is a hot, stale-valued PerceiveRecord (coord in its payload)

**Records are named after the primitive that mints them.** A perceive mints a
**`PerceiveRecord` (`kindPerceive`)** — a sense event (entity + coord, no utterance); an engage
mints an **`EngageRecord` (`kindEngage`)** — a communication event (from/to/utterance). So the
record type reads straight off the action: `skyra perceive <entity>` → `PerceiveRecord`,
`skyra engage <entity>` → `EngageRecord`. Both hang off the entity and route through the
grouping; the kind keeps the sense stream separable from the conversation stream.

> **These are Kinds, not types.** `PerceiveRecord`/`EngageRecord` are **`Kind` values of one
> `AffordanceRecord`** (Kind filled from the affordance that came back), not separate Go structs
> — see [affordance-record.md](../being/affordance-record.md). So `PerceiveRecord` ≙
> `AffordanceRecord{Kind:"perceive"}`, `EngageRecord` ≙ `AffordanceRecord{Kind:"engage"}`. Since
> `PerceiveRecord` is unbuilt, it can be **born directly as `AffordanceRecord{Kind:"perceive"}`**
> — the natural first site for the standardization. `EngageRecord` is today `dna.InteractionRecord`
> (`skyra-dna/memory_cell.go`); its collapse into `AffordanceRecord` is a follow-on.

**The location is the `PerceiveRecord`'s payload — not a separate leaf on the entity.** The
`PerceiveRecord` is an `AffordanceRecord{Kind:"perceive", payload:{coord}}`; it sits **between
gist and entity** (`gist → PerceiveRecord → entity`) and *carries the coord*. The entity does
not own a `location` child; the coord rides on the perception event that references the entity.

```
gist ──▶ PerceiveRecord{coord} ──▶ entity        (the record's edge carries the heat)
```

- Perceive **appends** a new `PerceiveRecord` (hot edge) under the active gist, pointing at the
  perceived entity, coord in its payload.
- The coord **is old the instant it is written** — "current" is carried by **which
  `PerceiveRecord` is freshest**, never by any single value.
- Prior `PerceiveRecord` edges **decay**; the current location is the freshest record's coord.
- **Staleness = decay, for free.** An entity Skyra stops perceiving **cools** → "I'm no longer
  sure where B is" is those records going cold. Confidence-in-location *is* their heat.

No mutable "current location" field, no overwrite, no eviction logic — append + decay.

> **How "current" is read — the shortcut we ship first.** The principled read is "freshest
> record by activation." For now we **cheat with a normalization pass — *after* traversal, not
> in lieu of it.** The wave traverses `gist → PerceiveRecord → entity` normally (activation
> surfaces the reachable records). Then, **right before the present is derived**, a
> **normalization abstraction** collapses the traversed `PerceiveRecord`s **per entity** to the
> **max-`Timestamp`** one (they're typed `kindPerceive`, so the dedup is cheap). Traversal stays
> the mechanism; normalization only dedups its output before present-assembly. This sidesteps
> needing perceive records to decay faster than everything else to stay correct. See §7 (decay)
> for the real mechanism that lets this normalization drop away, deferred.

## 3. The spine — the gist owns records; the entity routes back to groupings

The grouping does **not** point to entities directly, and it does **not** own the records
directly either. **The grouping holds gists; the gist (the framing) owns the records; the
record points at the entity; the entity routes back out to every grouping it has been part of,
weighted by activation.** The full spine (and the mesh it closes):

```
grouping ──▶ gist ──▶ record{payload} ──▶ entity ──┬─▶ grouping#7   ┐ all groupings this
 (id/       (the       (event)            (who)     ├─▶ grouping#3   │ entity is in,
  cluster)   framing)   PerceiveRecord{coord}       └─▶ grouping#12  ┘ weighted by activation
                        | EngageRecord{utterance}         │
                        (Kinds of AffordanceRecord)       └─▶ (each is another grouping → gist → record → entity → …)
```

Every `──▶` is a `children` edge. The records (`PerceiveRecord`/`EngageRecord` — Kinds of
`AffordanceRecord`) are children of the **gist**; each **carries its affordance payload** (a
perceive's coord, an engage's utterance) and points at the entity it concerns. The **entity is
an associative router**: `entity.Children()` = the groupings it has participated in, each edge
activation-weighted. So the graph is a **mesh** — `gist → record → entity → grouping → gist →
…` — and the wave hops between groupings *through the entities they share*. This is
entities-and-surroundings.md's hub insight lifted from setpoints to **groupings**: retrieval is
spreading activation, no separate index.

**Write-time discipline:** when a record under grouping-G's gist references entity E, append
`E → G` to the entity's children (E has now been part of G). That edge is what later lets a wave
arriving at E surface G — the reverse-read, maintained at write time (same contract as
entities-and-surroundings.md's "append the setpoint to the entity on completion").

**Why the gist owns interactions (not the grouping, not `grouping → entity`):**

- **Entity is shared, framing is not.** One `michael` cross-cuts many contexts; a record belongs
  to **one gist**. Routing through the gist gives framing-scoped reachability **without copying
  the entity**. `grouping → entity` would (falsely) assert the entity *belongs to* the grouping.
- **The gist is finer than the grouping.** A grouping persists while its **framing changes**
  (the edit/sharpen path). Interactions belong to the **framing they arrived under**, so a
  re-frame mints/sharpens a gist that owns subsequent records while the prior gist keeps its
  own. Ownership tracks the framing, not just the episode id.
- **Crystallization needs no re-parenting.** A gist *is* an uncrystallized understanding
  ([groupings-and-understandings.md](groupings-and-understandings.md)); on closure it becomes an
  **Understanding owning its constituents**. Since the gist **already** owns its interactions,
  `gist → interaction` *is* the pre-crystallization form of `Understanding → children` — closure
  flips status, not structure.

> **Decided: a sequence of gists.** A grouping holds a **sequence** of gist nodes — each edit
> mints a **new** gist (priors retained, decaying), each owning the records that arrived under
> it. Midlevel's **edit-vs-new decision already exists** (`genes_midlevel.go` — `<edit id=…>` /
> `<new>`), but the **structure does not yet support the sequence**; three code deltas:
> 1. **Gist becomes a child node** of the grouping, not the grouping's `Content`. Today
>    `Create` puts the gist in `KindGrouping.Content`; make it a `KindGist` child so several
>    coexist.
> 2. **Edit mints a new gist node**, not `SetContent`. Today `edit` does
>    `g.SetContent(Gist{…})`, overwriting (the prior text is lost). Instead append a new gist
>    child and make it active; the prior gist stays and decays.
> 3. **Interactions attach to the active gist**, not the grouping. Today
>    `consolidateInteraction` does `AddChild(root.Grouping, …)`; retarget to the active gist.
>
> This is the structure crystallization-without-re-parenting (§3) and framing-scoped ownership
> depend on. Non-trivial — it touches the midlevel edit path and the grouping node shape.

## 4. Motor is handed the location by inhabiting the grouping

Motor does **not** subscribe (pull). The entity — reached through the active gist's
`PerceiveRecord` — is already in the frame midlevel built; motor inhabiting that grouping
**reaches the location by plain traversal**, re-freshed every cycle perception fires. Motor
never reaches for it; **it is delivered.** This is the Perception Invariant at motor's own
frame: *what motor perceives is not chosen — it is whatever its inhabited grouping hands it.*

**Scoping is amplitude economics, not a wall.** Nothing is prohibited. The inhabited grouping is
the **traversal root** — full amplitude — so it (and its entities, records, locations) is
**~guaranteed in the frame**. From there the wave attenuates as it branches:

> **"Amplitude" on the frame path = normalized budget share** (confirmed 2026-07-07,
> perceive-build-uncertainties.md §6). The dnav2 frame realizes this economics as
> `traverse.Assemble`'s token budget: each node's children get weight-normalized shares summing
> to 1 (conservation), so falloff = a shrinking budget share down the tree. No second scoping
> mechanism (no radius rule, no wall) is to be built.

- **Entity A is in grouping A** (the root) → high amplitude → in frame.
- **Entities A and B share grouping B.** The wave can hop `grouping A → gist → record → entity A
  → grouping B → entity B` — so **entity B may enter the frame** *iff* there is **leftover
  amplitude** to carry the wave through grouping B. Related context bleeds in when the wave still
  has energy; it is **not forbidden**, it is **paid for in amplitude**.
- **Unrelated agent C** (no shared grouping with A) either has no path, or one so long the
  amplitude has died before reaching it. So it stays out — not by a rule, by falloff.

So motor's **tight bound = the high-amplitude core** (the root grouping); **associative reach =
the low-amplitude bleed** into shared-entity neighbors. The `entity → grouping` edges are the
routes; amplitude decides how far the wave actually travels them. (Same agent seen in two
contexts reaches the **same** entity — one michael, one freshest coord.)

This is operator-runtime's "local world slice" made concrete (the slice = the amplitude-bounded
neighborhood of the root grouping) and the spatial extent of prediction-motor-and-closure's
**bound**: the entities-with-fresh-coords the wave reaches **are** the world-cell neighborhood
motor traverses (emit-as-traversal). Perception is a **constituent of the bound**, not a side
channel motor polls.

## 5. The internalization mechanism (the three walls → the fix)

Today the perceive **return** dies at three walls. The build cuts here:

1. **Bridge drops the return wave.** `cmd/tui/being.go:46` — `dnav2Being.Observe` discards
   `imp.Return`. The world's `PerceptionResult` never reaches the pipeline.
2. **Sensory no-ops the reafference lane.** `genes_sensory.go:22` — `classify` correctly labels
   it `"reafference"` but the gene records nothing ("phase 2" stub).
3. **Nothing ingests a `PerceptionResult`.** dnav2 has zero production references to
   `world.PerceptionResult`; the reafference route goes to `self:grade`, which reads a **string**
   result (a struct falls through as "no result").

**The fix — bridge translates, sensory ingests, kept world-agnostic:**

- **Translation lives in the bridge (Option A), not the DNA.** `dnav2Being` (the harness — it
  already imports both sides) unwraps `world.PerceptionResult` into a **DNA-native surroundings
  payload** (`{entityID → coord}` + self-coord) and delivers *that* in. The DNA package stays
  world-agnostic (only tests import `world` today — keep it that way). This is the sensory
  transduction layer: the world speaks its own result type; the membrane bridge translates it
  into cognition's vocabulary.
- **Stop dropping perceive returns at the bridge.** A `Return` perceive wave enters the pipeline
  as a **reafference** (threaded by prev-id — the reafference chain the board already carries).
  It is NOT an internal event — `internal` is reserved for interrupts (e.g. motor engaging the
  self); a perceive-return is the world answering my own act, and what distinguishes it from an
  action-reafference is only *where the record lands* (decided 2026-07-07,
  perceive-build-uncertainties.md §1).
- **Split the reafference lane.** `perceive`-return (bring in surroundings) is distinct from
  `move`/`engage`-return (grade my action's execution). Perceive routes to a **sensory
  ingest**, not to `self:grade`.
- **The ingest gene appends a record under the active gist.** For each `{entityID, coord}`:
  `Ensure(entity)`, append a **`PerceiveRecord` (`AffordanceRecord{Kind:"perceive", payload:{coord}}`)**
  as a child of the **active gist**, pointing at the entity, with a **hot** edge; append the
  `entity → grouping` route (the entity has now been in this grouping); and `CrossBirth` the new
  edges (born active → participate in activation immediately). No `entity → location` leaf — the
  coord rides on the record (§2).

**Self is the wave's origin, not a discovery (§5a).** The perceive wave is rooted at Skyra's
own point and sweeps outward; the world returns `SelfCoord = origin's coord`
(`world/space.go:564`) alongside the swept `Agents`. So:

- **Self-location is given** — `SelfCoord` is *where the wave started*, the anchor, not
  something found in the periphery.
- **Other-locations are swept** — discovered out in the radius (the `Agents` map).

Skyra **has an entity of herself**; the ingest writes `SelfCoord` as a `PerceiveRecord` (coord
in payload) under the active gist, pointing at that **self-entity**, by the identical mechanism
as any agent — the only difference is the coord's *source* (origin vs. sweep). This matches
self-relies-on-other and the Perception Invariant: the origin is known, the periphery is
perceived.

**Who groups it:** sensory only *deposits* the `PerceiveRecord` + entity + hot location into the
**active grouping** (the default one if nothing is active — §5b). **Midlevel** then frames
(sharpens or splits). Clean separation — sensory transduces, midlevel frames, motor inhabits.

**Reading it back is post-traversal.** Motor/present-assembly does **not** scan records instead
of traversing. The wave traverses `grouping → PerceiveRecord → entity → location` as normal; the
per-entity most-recent-location dedup is a **normalization pass on the traversed set, right
before the present is derived** (§2 shortcut). Present-derivation is the single seam where that
normalization lives.

## 5b. The default grouping — the catch-all (impl carve-out)

**Problem (confirmed in code).** The only grouping-creation path is `midlevelFrame →
gi.Create(groupID, gist, ts)` (`genes_midlevel.go:54`), gated on a **non-empty exchange** + a
parsed gist. When the slice has **no active grouping** and an impulse lands (Skyra only ever
acts *on* an inbound impulse — there is no free-running tick, so this is the real cold-start),
sensory has nowhere to deposit its `PerceiveRecord`/interaction. There is **no default grouping**.

**Fix.** Create a **default, gist-less grouping** whenever no active grouping is on the slice —
the catch-all that incoming records land in before midlevel has framed anything.

- **Where:** on `Fire` / at sensory-intake, before deposit: if the board row has no active
  grouping and `gi` has no active episode, `gi.Create(defaultID, "" /* no gist */, ts)` and set
  it active. (`gi.Create` already "sets it active" — reuse it; the new part is the **empty
  gist** and the **no-exchange-required** trigger.)
- **Shape:** a normal grouping node holding **one empty-text gist node** (records attach to the
  gist, §3, so the default needs a gist to receive them — an empty-`Text` `KindGist` child). It
  is a legitimate grouping — records hang off its gist, edges are born, it can be inhabited — it
  simply has no framing *yet*. Midlevel's first `edit` fills the gist text (or `new` splits off).
- **Lifecycle into midlevel:** when `midlevel:frame` next runs, the default grouping shows up in
  `OpenEpisodes()` like any other; the edit/new operator **sharpens it** (`edit` → it gains a
  real gist and keeps its edges — the compounding path) or **splits a new episode** off it if
  the topic genuinely differs. So the default is the seed midlevel grows a framed episode from,
  not a throwaway.
- **Open (impl detail):** whether the default grouping is **per-being singleton** (one
  ever-present "unframed" bucket that keeps getting re-sharpened) or **minted fresh** each time
  the slice goes empty. Leaning singleton (a stable id → its edges accumulate); pin at build.

## 6. What this revises in the existing docs

- **entities-and-surroundings.md:** the single per-slice global surroundings map is gone (§1).
  Entities-as-cells + wave-from-context stays; the "reverse read" now enters through the gist's
  records, not a global map query.
- **working-memory.md traversal chain:** flips. It currently reads `entity →(overlap)→ grouping`
  and `entity → {location}` as an inline value. The spine is now
  `grouping → gist → record → entity → grouping …` (records owned by the **gist**, carrying their
  affordance payload; entity reached *through* the record; the entity routes back out to its
  groupings). No `entity → location`: the coord rides on the `PerceiveRecord`. Records are an
  added child-kind under the gist.
- **token-web-and-entity-resolution.md:** its `entity → interaction` edge **flips** to
  gist-owned `gist → record → entity` (interactions are no longer entity-owned). Tokens stay as
  record children (`record → token`). Aliases-as-grouping-children is *unchanged* and now even
  more aligned. Everything else in that doc (the resolution pipeline, confidence/TMS, the three
  phases) is untouched. Marked unbuilt in dnav2 → **port fresh in this direction.** (Membership =
  `record → entity` (into the entity, who's in frame); recall = `entity → grouping` (out of the
  entity, to its other groupings, activation-weighted) — two edges, opposite directions, no
  conflict.)

## 7. Decay — the real "current location" mechanism (deferred)

The normalization pass (§2) is a stand-in. The principled read is edge heat, and it becomes
correct once **`PerceiveRecord` edges decay faster than everything else** — a position goes stale
faster than a fact, so the freshest perceive record naturally outranks older ones by activation
and the timestamp-dedup drops away.

Current mechanics (`activation.go`, the live four-factor edge record — **not** the stale
two-EMA in activation-and-traversal.md): `Strength = rectSoft(A)·S·F·conf(N)`; decay lives on
**F** only — `readF = F·λ^(tick−Flast)` with a **single global `λ = 0.98`/tick**. A perceive-
specific faster decay = a **per-kind `λ`** (below 0.98) for `kindPerceive` edges.

**Not now.** We ship the §2 normalization first. When we do this: add the per-kind `λ` in
`activation.go` **and** drop the normalization pass — leave a code note at both sites tying
them together (the normalization exists *because* λ is not yet per-kind).

## Open questions

- **Entity resolution** — *come back to this.* `perceived SpaceID → stable entity id` is
  exact-match (`resolveEntity`), which is **correct and sufficient on the world/perceive path**
  (world SpaceIDs are stable + unique, so a `PerceiveRecord` and an `EngageRecord` for the
  same agent land on one shared entity for free). The hard problem — fuzzy/alias resolution of
  *linguistic* mentions, cross-session re-id — lives on the conversational path and does **not**
  gate this build. Parked.
- **Default grouping: singleton vs. fresh-per-empty** (§5b) — leaning per-being singleton with a
  stable id (edges accumulate); pin at build.
- **Perceive decay `λ`** (§7) — the exact per-kind value (below 0.98) is empirical; deferred
  behind the §2 normalization.

### Resolved (folded in)

- **Record kind** → a perceive mints a **`PerceiveRecord` (`AffordanceRecord{Kind:"perceive"}`)**,
  distinct from `EngageRecord`; **coord in the record's payload**, not an entity leaf (§2).
- **The spine** → `grouping → gist → record → entity → grouping…`: gists are a **sequence**
  (append on edit), the **gist owns records**, the entity **routes back to its groupings**
  (activation-weighted associative recall) (§3).
- **Scoping** → **amplitude economics, not a wall**: the inhabited grouping is the traversal
  root (~guaranteed in frame); shared-grouping neighbors enter *iff* leftover amplitude carries
  the wave there (§4).
- **Cold-start routing** → a **default grouping with one empty-text gist** catches deposits when
  nothing is active; midlevel fills the gist (§5b). Skyra only acts on an inbound impulse, so
  there is no spontaneous-new-agent path to route.
- **Self-location** → a `PerceiveRecord` under the active gist pointing at Skyra's **self-entity**,
  written from `SelfCoord` (the wave's origin — given, not swept) (§5a).

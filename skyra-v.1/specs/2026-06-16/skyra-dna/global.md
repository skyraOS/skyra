# Global — Cell Type Spec (vote tally / tie-break)

**Date:** 2026-06-19 · **Status:** in spec (updated)

> Working-memory indexes (`groupID → setpoints` (observe), `action → {self,other prediction}` (express), the grouping spine, loose coupling) are specced in [working-memory.md](../../../design-docs/cognition/working-memory.md).

## File
- `../../../src/skyra/skyra-dna/global.go`

## What it is

The orchestrator. Three responsibilities:

1. **Vote tally** — collects scores, runs the Hopfield weighted sum
2. **Tie-break** — fires when the vote doesn't resolve cleanly; the slow path
3. **Grouping** — maintains the **index over active setpoints**: clusters them into higher-order (longer-horizon) setpoints, recursively. This *is* the future layer (see below).

## Grouping — the future as a top-down index

The self/other agents handle the **short-term**: predict the near move and close it. Global handles the **abstraction / standing-intent layer**: group active setpoints into a higher-order setpoint.

> short-term `other`: "the user is going to talk about the skyra project"
> short-term `self`: "mention the skyra project and prior details"
> Global grouping: **"help the user with skyra"**

Recursive: when groupings get dense, group the groupings — an index over an index. Don't name the tiers (no "medium"/"long"); horizon is emergent from grouping depth. **Loosely coupled** — a grouping is a soft association over what's currently active (the pre-crystallization phase of an edge; it grooves if it keeps proving useful, decays if not). Grouping is the Hebbian crystallization pointed *upward*.

**Gist = an understanding that hasn't crystallized yet.** Each grouping carries a gist — a compressed framing (LLM pass over its constituents) — a *potential* Understanding. On closure it crystallizes into a durable Understanding; if the grouping never resolves it decays away. The grouping/gist pipeline and the episode-closure pipeline are the **same pipeline** at two altitudes. Full model: [groupings-and-understandings.md](../../../design-docs/cognition/groupings-and-understandings.md).

**This replaces the parent/future edge.** Setpoints are past-only; the future is not stored per-setpoint. It lives here, as Global's grouping index, and is **passed down as context** to the active setpoints — so each setpoint is **reinterpreted under whatever long-term goal is currently active** ("mention skyra project" means one thing under "help with skyra," another under "wrap up the call").

### Grouping runs in Global's Observe (on impulse)
- On impulse, score the impulse against the **active groupings**, in the context of the lit subgraph (working memory) — really *lit-set-vs-grouping*, **one inference call**.
- **Skips entirely if there is no completed setpoint** (cold start has nothing to abstract over).
- The chosen/active grouping becomes the context passed down for the rest of the cycle.

> **The read-half mechanics** of this pass — how Global gathers the candidate setpoints
> (entity-outward from sensory's warm surroundings, scored by overlap × recency, cell-written
> IndexScore, dynamic `Children()`) — are specced in
> [global-observe-gather.md](global-observe-gather.md). The inference call here is the write
> trigger on top of that gather.
>
> **MVP (above) vs. the improvement.** The single inference call here is the **MVP** — build
> it first. The planned upgrade — Global as an **active LLM-driven probe loop** that explores
> nodes *before* grouping (passed-competence probe waves, a mini-BSP join on the slice) with a
> **drift-diffusion stop controller** (alignment-plateau + model-tier cascade + time→stress
> urgency bound) — is specced separately in
> [global-active-grouping.md](global-active-grouping.md). Same metric (alignment), better
> groupings; layered on once the loop runs.

### The reverse read is a wave, not a hash
Turning the **surroundings map** (current active entities, updated on every `perceive`) into the relevant setpoints is **not** an index lookup — it's a **wave fired from the surroundings entities**. Entities are cells whose `Children()` are the setpoints stamped on them at completion; the wave spreads `understanding → entity → understanding → entity …`, bounded by amplitude. Associative recall = spreading activation, free from the kernel, no separate index to maintain or scan. Decay is automatic (edge-owned activation + threshold); salience falls out of branching (hub entities spread thin). See [entities-and-surroundings.md](../../../design-docs/cognition/entities-and-surroundings.md).

## Agents — self and other

`Agents()` → **two** agents, each its own descent/ascent structure:

| Key | Role |
|---|---|
| `other` | models the other — the estimation/prediction pass (descent) and its scoring (ascent) |
| `self` | models the self — composes alignment toward closure; feeds the vote |

This supersedes both the WorkingMemory+Category two-hemisphere model and the "one binding reality" collapse. The structure is now **self vs. other as two agents**, each a full descent/ascent (per [self-and-other-modeling.md](../../../design-docs/cognition/self-and-other-modeling.md)) — not one global descent=other / ascent=self.

## The cognitive cycle

### Step 0 — Observe / grouping (Global, one LLM call)
On impulse: score it against the active groupings in the lit set's context, set the active grouping as the cycle's top-down context. Skips if no completed setpoint exists. (See "Grouping" above.)

### Step 1 — Filter (Global, kernel, no LLM, pure math)
The kernel walks the setpoint topology, scoring by `activation × stress × alignment` (cell-owned — kernel calls `Activate()`, gates by threshold). The impulse + warm activation cut down which setpoints are live. Survivors (> 0.1) **are the lit subset** — working memory is not a buffer assembled alongside the graph, it is the activation-selected *view* of the one graph the traversal produces (see [working-memory.md](../../../design-docs/cognition/working-memory.md)).

**Working memory is the lit subgraph, not a separate store.** Everything is appended to the one graph; what makes a node "in working memory" is its recency activation (the fast Familiarity EMA), not membership in a second structure. So the lit set just *is* the high-activation nodes the wave reaches this cycle:
- Incoming evidence (responses, actions, observations) — appended, lit because fresh
- Recent failed predictions — lit because recent (the gap between expected and actual is the most informative signal for the next prediction); they fade by decay, not by being drained from a buffer

On episode closure nothing is moved between stores — the lit arc's nodes (already in the graph) are **relinked in place** as children of the closure node and the new node is stamped:
- **Success** → the resolved arc becomes children of the resulting Understanding
- **Failure cascade** → the retracted delta becomes children of the resulting Stress

### Step 2 — The self/other pass — **per grouping** (not per setpoint)
One predict-score-close pass **per active grouping**, with the grouping's constituent setpoints as *material*, not as the unit of iteration. (Supersedes the earlier per-setpoint fan-out.) This is cheaper and ties grouping width to explore/exploit: a narrow grouping (few setpoints) leaves more spillover per setpoint to explore; a wide grouping spreads attention thinner and exploits.

**Frame for the pass:** high-activation history of the grouping's setpoints + the currently selected setpoints + Global's **gist** (the framing) + predictions carried from the previous pass on the slice.

**Self and other fire concurrently (closure phase).** A grouping holds both self-content and other-content under one gist (one situation, two faces). At closure they ask independent retrospective questions — self: "did my action close my commitment?"; other: "did this impulse close the other's open setpoint?" — so the pass spawns **N ≥ 2 concurrent waves** (≥1 self, ≥1 other). Each is a separate-slice agent reading a **snapshot** of the shared frame; Global owns the lit set and integrates only at the join. Safe by construction: **snapshot-in, join-out.** (Concurrent at closure; still ordered other→self at prediction — "self relies on other.")

- **predict / grade.** Produce the grouping's prediction (appended to the graph under the grouping, lit because fresh); grade against incoming evidence (broadcast the lit set, each constituent self-selects relevance, alignment accumulates).
- **closure.** On proven-true: the prediction **graduates into a new setpoint at level n**; constituents become its **history (children, n-1)**; the **gist crystallizes into a durable Understanding**; setpoint(n) is eligible at n+1 (stacking).

Same predict-score-close loop as Frame 1 / Frame 2 below, run at the grouping altitude. See [groupings-and-understandings.md](../../../design-docs/cognition/groupings-and-understandings.md).

### Step 3 — Orchestration: route the 2-D outcome (Global, code)
The concurrent waves return independently → a 2-D outcome (self × other). A **pure-code selector** routes on `(stress-delta, closure-grade)` — **cascade dominates**:

1. **Any cascade → reconcile (failure).** Roll the setpoint state back to solid ground (~0.2 threshold), pop the delta to WM, reorganize → **Stress**. See [error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md#cascade--a-failure-wave-de-crystallization).
2. **Else any closure → record + continue (success).** Recurse to Global, regroup, kick the next wave (predict → self); gist → Understanding.
3. **Else open → act + await.** Re-predict → self → emit a setpoint action → await the next impulse.

Routing never calls an LLM; inference lives only in the branch's work.

### Step 4 — Vote (Global, no LLM, pure math)
Hopfield weighted sum over the scores. Threshold crossing → action. No clear crossing → Global fires (tie-break / gather signal). Scores feed back into the activation records for the next filter pass (alignment is lagged).

## Observe — two internal functions

`Observe` routes internally based on something on the payload (discriminator TBD — do not generalize to all cells yet):

| Internal fn | Triggers when | Does |
|---|---|---|
| `observeExternal` | incoming signal from outside | **the grouping pass** — score impulse vs active groupings (lit-set-vs-grouping, one inference call); skip if no completed setpoint; set the active grouping as context for the cycle |
| `observeSelf` | returning result from an agent Global dispatched | integrate the returned scores (toward the vote) |

## Visibility (scoped altitude)

Global reads exactly two things: **crystallized understandings** of completed setpoints, and **potential gists** of currently active groupings. It never reads raw, ungrouped, unresolved setpoints directly — that altitude is the kernel traversal + the raw predict/score layer (Frame 1 / Frame 2), below Global. A grouping has **no explicit graded alignment**; its relevance is read implicitly off the attention Global keeps giving it (selection constitutes significance).

## Interface (`skyra.Space`)
- `Children()` → **groupings** (Global's index — higher-order setpoints over the active field), keyed/indexed by person/place/thing + stress. The raw setpoints are below Global's altitude.
- `Agents()` → `{"self": ..., "other": ...}`
- `Observe(origin, impulse)` → routes to `observeSelf` or `observeExternal`
- `Express(origin, impulse)` → dispatches the agents, runs vote, fires on conflict
- `Content()` → the **grouping index** (higher-order setpoints over the active field — the durable future layer) + the transient vote synthesis

## Open
- Exact discriminator for `observeSelf` vs `observeExternal`
- Whether the two-observeFn pattern generalizes beyond Global
- Hopfield weight derivation (from alignment history? stress? both?)
- Tie-break mechanism: inner-voice re-entry vs. gather signal vs. explicit conflict resolution
- **Grouping formation:** how a *new* grouping is born vs. matched to an existing one (does the Observe inference call propose new groupings, or only score against existing?)
- **Grouping density trigger:** what threshold of grouping density fires the group-the-groupings recursion / stacking to n+1
- **Setpoint(n) vs. Understanding(n):** on closure, one node (the Understanding is the setpoint's framing) or two tightly-bound nodes? (see [groupings-and-understandings.md](../../../design-docs/cognition/groupings-and-understandings.md))
- **self/other vs. Frame 1 / Frame 2:** captured as the same predict-score-close mechanism at two altitudes — split if they're meant to be distinct

## Done when
- Grouping runs in Observe: impulse scored vs active groupings (lit-set context), active grouping passed down; skips if no completed setpoint
- The self/other pass runs **per grouping** (constituents as material); prediction appended to the graph under the grouping (lit, uncrystallized); on closure → new setpoint(n), constituents → children(n-1), gist → durable Understanding
- Self/other fire concurrently at closure (N≥2 waves, snapshot-in/join-out)
- The 3-branch orchestration routes the 2-D outcome in code (cascade > closure > open)
- Runs the Hopfield vote over returned scores; fires on conflict (tie-break)
- The cycle completes end-to-end: observe/grouping → filter → per-grouping concurrent self/other → orchestrate (cascade/closure/open) → vote → action

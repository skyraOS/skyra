# The Active Grouping — Episodic Consolidation

**Status:** design, not built. Resolves how incoming interactions are recorded, and turns
per-turn grouping *islands* into growing *episodes*. Builds on `memory-as-world.md`
(match/edit/new), `episodic-encoding.md`, `groupings-and-understandings.md`, and the
economy in `../kernel/traversal-mechanics.md`.

## The problem

Today `sensory:intake` records an incoming message as a child of the **entity**
(`michael → interaction`, `KindResponse`). Three costs:

1. **Hot entities.** The entity accumulates every message as a flat, unbounded fan-out —
   traverse to `michael` and you face thousands of unchunked interactions.
2. **Inputs are outside the economy.** The frame only walks the *grouping* spine; the
   entity index is never crossed or credited, so the evidence that grounds cognition is
   invisible to the traversal (`renderExchange` hand-renders it).
3. **Islands.** `midlevel:frame` mints a **new** grouping every turn, so each turn is a
   disconnected graph credited once (`N=1`) — the accumulation loop can't compound.

## The move

Interactions hang off **groupings** (episodes), not entities. Entities *point to*
groupings. Memory tiers: `entity → episodes → interactions` — bounded fan-out at each
level, the way episodic recall actually works (you recall a context, then its specifics).

The "no home before the grouping exists" tension dissolves with a **reference to the
active grouping**, not a buffer: there is always a current episode; `sensory` attaches to
it; cold-start lazily creates one. The grouping *is* the buffer — always present, extended
in place.

And `midlevel:frame` becomes **episodic consolidation** — the match/edit/new already
specced in `memory-as-world.md`:
- **edit** — update the active grouping's gist; the episode *continues*.
- **new** — create a grouping and switch active; a *topic shift*.

## Four for one

This single change settles four open problems together:

| problem | how this settles it |
|---|---|
| home for a just-arrived interaction | the **active grouping** (created on cold start) |
| hot entities | interactions live under groupings; `entity → grouping` edges (bounded) |
| inputs outside the economy | interactions now sit on the **grouping spine** the frame walks + credits |
| islands / `N=1` (no compounding) | **edit** grows one episode → its edges persist and re-grade → `N` climbs |

The last row is the important one: *the edit branch is the fix for compounding.* Extend
the episode instead of minting a new one, and the same edges are crossed and re-graded
turn after turn — which is exactly what the economy needs to accumulate.

## The associative web

The edges complete a cycle and turn memory from a tree into an associative graph:

```
entity ──▶ grouping ──▶ interaction ──▶ entity
   ▲                                        │
   └────────────────────────────────────────┘
```

- `entity → grouping` — the episodes a participant was in (cool, bounded fan-out).
- `grouping → interaction` — the exchanges within an episode.
- `interaction → {origin, other-participant}` — who was in this exchange, as **children of
  the interaction** (out-edges from the interaction, so entities stay cool — the
  participants aren't piled onto the entity).

The payoff is **associative recall**: from an interaction you reach its participants, and
from a participant you spread to *their other episodes* — "what else has happened with this
person." That is the graph gather-mode traverses (`../kernel/traversal-mechanics.md`,
Drill and gather).

Three deliberate consequences:

1. **Off the compounding critical path.** These edges enable associative recall and, later,
   entity-valence ("interactions with X tend to close well / badly"). They do **not** fix
   `N=1` — the edit branch does. Build them *after* the active-grouping consolidation.
2. **The self edge is near-constant.** In the dyadic skyra↔michael case skyra is a
   participant in *every* interaction, so `interaction → self` discriminates nothing; the
   *other* (varying) participant is the informative edge. Link both for generality
   (multi-party futures — michael *and* the philosopher), but the self-edge earns its keep
   later, not now.
3. **It densifies the graph** — everything becomes reachable from everything via
   participants. The visited-once rail handles the cycles and weights + budget prune the
   density, but a *cold* graph (no learned weights) will over-spread under gather; the
   economy has to earn its selectivity. Deliberate, not a bug.

## Mechanics

**`GroupingIndex`** holds the active episode:
- `active *MemoryCell`, `Active() *MemoryCell`, `SetActive(*MemoryCell)`.

**`sensory:intake`** (structure — runs first):
- `g := gi.Active()`; if `nil` → `g = gi.Create(proto, "" , tick); gi.SetActive(g)` (cold start).
- attach the interaction (`KindResponse`) as a child of `g`.
- hang the interaction's **participants** as its children — `interaction → origin`,
  `interaction → other-participant` (see "The associative web").
- ensure an `entity → g` participant edge (dedup per `(entity, g)` pair).

**`midlevel:frame`** (semantics — decides continuity):
- read `g := gi.Active()` and its recent interactions.
- decide **edit vs new** (LLM judge: does the latest message *extend* this episode or
  *start* a new one? — a cheap similarity gate can front-run it later):
  - **edit** → infer an updated gist, `g.SetContent(Gist{...})`; active unchanged.
  - **new** → `g2 := gi.Create(...); gi.SetActive(g2)`.

So cold start splits cleanly: `sensory` mints the structural node (empty gist), `midlevel`
fills the semantics (edit). Structure at intake, meaning at framing.

## The realization edge (cross-ref)

The incoming interaction is *also* the ground-truth outcome that grades the prior bet.
Per the node-ownership thread, `other:grade` should link the interaction to what it
realized — the **matched branch** (hold) or the **grouping** (miss) — instead of stashing
an evidence string. So the interaction is born wired two ways: child of the active
grouping (episode home) **and** linked-from the prediction it realized (outcome). That may
be its own build step; noted here so the two don't fight.

## Alignment vs closure — the settle writes alignment; closure is parked

Two **separate axes**, not to be conflated (cross-ref `../kernel/traversal-mechanics.md`,
"Activation & the bump" correction):

- **Alignment / edge strength (`A`)** — the *coherence of a node to its framing*. This is
  what the midlevel **coherence-settle** writes: propose a framing (one inference), then
  **batch-grade** the window's nodes against that *fixed* framing, prune the ones below 0,
  re-traverse the survivors (freed budget swaps misfit-near for coherent-far), regrade —
  until the average alignment **plateaus**. Midlevel withholds its frame until then. `A` is
  read by the traversal as weight and climbs `N` as episodes persist (the compounding fix).
  The batch grader is validated to ~100k tokens.
- **Closure** — did the prediction hold / the action work (the two-stroke outcome). A
  *separate, not-yet-designed axis*. It probably factors into edge strength somehow, but it
  is **not** `A`. The realization edge above (what a prediction realized) belongs to
  *closure*, not alignment. **Parked** until designed.

So the settle is the alignment writer; the current `distributeClosure → A` wiring
(traversal-mechanics Phase 2c/2e) is a **placeholder** the settle will replace. *edit/new*
falls out of the settle: if the active episode's nodes survive coherence with the new
interaction → **edit**; if the settle prunes them to cohere around the new material →
**new** (still to be confirmed/built).

## Open decisions / wrinkles

1. **Boundary interaction.** On a **new** decision, the shift-triggering message was
   already attached to the *old* active grouping at `sensory` time. *Lean: leave it* (the
   closing beat of the old episode) rather than re-parent. Revisit if it muddies recall.
2. **Where `active` lives.** On `GroupingIndex` (nucleus) — one active episode per being.
   (A per-slice active ref is possible but one current episode is the natural model.)
3. **Edit-vs-new mechanism.** LLM judge first (midlevel already infers); a cheap embedding
   / gist-similarity gate can pre-filter later to avoid an inference per turn.
4. **Episode end.** Explicit (a **new** decision) to start; a decay/timeout close is
   deferred.
5. **Gist under edit.** Does the gist *replace* or *accrete*? Start with replace (latest
   understanding); accretion (a running gist) is a later refinement.

## Build notes

- Touches `sensory:intake` (attach to active) and `midlevel:frame` (edit vs new).
- `renderExchange` / the frame now sources interactions from the active grouping, not the
  entity index — which also fattens the thin self-frame (`lit=1`) finding, since the frame
  now includes the episode's interactions.
- Existing pipeline tests assume new-grouping-per-turn; they'll need updates.
- Scope this as its own step **before** more Phase-2/economy work: it's what makes
  accumulation actually accumulate, so it's higher-leverage than anything left downstream.

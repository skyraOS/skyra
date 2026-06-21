# Entities, the Surroundings Map, and Reverse Read as Traversal

**Date:** 2026-06-20

## The problem

Given the current context (who/what/where is around right now), find the setpoints
relevant to it — every time Global fires, cheaply, no inference. This is the
"reverse index" function. The resolution: **there is no separate index structure.**
The entities are cells in the graph; the reverse read is a wave.

## Entities are cells

A **person / place / thing** is a cell (`skyra.Space`) with a stable **ID** and a
**type**. Entities are nodes in the same graph as setpoints — not a side table.

The hard part is **entity resolution**: knowing that the "michael" perceived now is
the same `michael` stamped on a setpoint last month. Everything else here is
bookkeeping; stable IDs are the work. Start dumb (exact-match `descriptor → id`),
upgrade to fuzzy/embedding resolution later.

## The surroundings map

A map on the slice of the **currently active entities**, updated **every time an
agent emits `perceive`**. The being's live "what's around me right now." It is the
query for the reverse read, and the snapshot a setpoint stamps at completion.

**Written by the membrane's `Observe`** (sensory intake): on a `perceive` result it
**resolves** the signal to stable entity IDs and **deposits** them into the map. The
resolver lives in this path — "putting signals in the right place in active memory"
*is* the entity resolution. (Resolves the open question of who owns the write.)

## The structure — all `children`, no backlinks

Everything is a forward `children` edge. There is **no bidirectional edge** (no pair
stored twice). The graph alternates two layers, and the alternation closes a cycle:

```
understanding --children--> entity --children--> setpoint(=understanding) --children--> entity --> …
```

- **A setpoint is one node.** `Content()` = its **understanding** (the crystallized
  gist — what it means). `Children()` = the **entities** involved + responses +
  actions + stresses (the surroundings snapshot + evidence at closure).
- **An entity's `Children()`** = the **setpoints stamped on it** (appended at
  completion — see below). This is the "reverse index," but as graph edges, not a
  hash.

Two *different* forward edges between *different* node-pairs (`understanding→entity`,
`entity→setpoint`) — together they form a directed cycle through alternating layers.

## Reverse read = a wave (retrieval as traversal)

Global's reverse read is **not** a hash lookup. It is a **wave fired from the
surroundings entities**: it spreads `understanding → entity → understanding → entity
…`, bouncing between the two layers, **bounded by amplitude**. Entity nodes are
routing hubs that let the wave hop between setpoints that share context. That *is*
associative recall — spreading activation, free from the kernel, no separate index
to maintain or scan.

This is why it scales: the wave doesn't visit the whole graph (O(N)); it propagates
the active neighborhood and dies at threshold. Dormant setpoints surface when their
entity is in the surroundings, because the wave enters through that entity.

## Overlap → priority

A setpoint touched by more of the active surroundings entities outranks one touched
by a single entity. That overlap *is* the `IndexScore` term in the bump
(`IndexScore × Alignment × stress`) — either counted directly (gather
`entity.Children()` across the surroundings) or realized as entity-waves converging
on the same setpoint (more waves → more amplitude). **Compute it once per cycle into
a single bump** — don't let a setpoint take N separate bumps for N entity-waves, or
a single high-overlap moment over-grooves the durable EMAs (conflating "relevant to
a lot right now" with "consistently relevant over time"). Overlap is the this-cycle
multiplier; the EMAs groove once per cycle.

## Two consequences that fall out for free

- **Decay is automatic.** Each `entity → setpoint` edge is edge-owned activation
  (its own EMAs). A decayed edge drops below threshold and the wave doesn't follow
  it — no tombstone cleanup, no lazy-skip logic. "Nothing deleted, just suppressed,"
  handled by the threshold gate.
- **Salience falls out of branching.** A hub entity with many setpoint-children
  divides amplitude across all of them (branching-divides-amplitude). A super-common
  entity spreads thin and doesn't dominate recall; a rare/specific entity spreads
  focused and pulls hard. Entity specificity → recall strength, no IDF weighting.

Termination is the usual guarantee: visited-once rail + amplitude attenuation.

## The write-time invariant

**On setpoint completion, append the setpoint to each involved entity's
`Children()`.** This is the contract the reverse read depends on — miss the append
and the setpoint is invisible to entity recall even though it exists. Same family as
the justification index (`action → justifying-assumptions`, written at action-time)
and the epistemic-status stamp: *the index is only as good as the write-time
discipline that maintains it.*

A setpoint is **created when the predicted self AND other both come true** (both
faces of the situation close). That completion is when the entities get stamped and
the understanding crystallizes.

## What this resolves

- **The "reverse index" question** — no separate structure; entities-as-cells +
  wave-from-surroundings.
- **setpoint(n) vs Understanding(n)** (was open) — **one node**: the understanding is
  the setpoint's `Content`; the entities + evidence are its `Children`.

## Open

- **Entity resolution** — `perceived descriptor → stable id` (exact-match first;
  fuzzy/embedding later). The real work.
- Physical entities (place/device/tool) arrive via the world/`perceive`; cognitive
  ones (person/goal) from the runtime — both land in the same map.

## Source

- `design-docs/cognition/episodic-encoding.md` — Layer 1 typed indexes, retrieval as traversal (this is its realization)
- `specs/2026-06-16/skyra-dna/entity.md`, `setpoint.md`, `global.md`

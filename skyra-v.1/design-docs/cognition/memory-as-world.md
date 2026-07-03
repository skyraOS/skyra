# Memory as a World — Inner Affordances

**Date:** 2026-07-03 · **Status:** recognition captured; framings are the first
instance (see [framings.md](../../specs/2026-07-03/framings.md)); do **not**
generalize until the second instance (entities) proves the shape.

## The recognition

**MemoryCells are a world of affordances, the way the outer world is.** The outer
world offers primitives — `perceive` / `engage` / `move` / `wait` — and disciplines
them with physics (distance, walls, radius). The memory graph is the being's
*inner* world, and it has the same anatomy:

- **primitives:** `match` / `edit` / `new` — the verbs an agent can fire at memory
- **physics:** activation, decay, threshold — **decay is the inner world's gravity**

The primitives are **affordances of the cells themselves** — the cell declares what
can be done to it — not methods of the agent doing the acting. Any agent facing
memory gets the same verb set, the same way any agent standing in the world gets
`perceive`.

This was seeded in notes.md months before the memory system existed to receive it
("custom primitives at any level — a brain's internal message-passing as a protocol
*defined as primitives on the agent's inner environment* — same kernel, same
delivery"). The framings wire format is that seed germinating.

## The symmetry, laid flat

| | Outer world | Inner world (memory) |
|---|---|---|
| substrate | the grid (Points) | the graph (MemoryCells) |
| primitives | perceive / engage / move / wait | match / edit / new |
| physics | distance, walls, radius | activation, decay, threshold |
| perception | `perceive` → surroundings | **the wave** — recall is perceive pointed inward |
| consequence | blocked, bumped, out of range | decayed, evicted, resisted |
| grammar | `<engage><michael>…</michael></engage>` | `<frame><edit id="f1">…</edit></frame>` |

The being lives between the two worlds: membrane faces out, cognition faces in,
and **both faces act through affordances and get disciplined by physics they don't
control.** Neither face gets to decree outcomes — the outer world can block a move;
the inner world can decay an edit into irrelevance.

## What unifies immediately

**Entity resolution is the same primitive.** `perceived descriptor → stable id` —
"the real work" flagged in every spec — *is* match-vs-new fired at entity cells.
When sensory resolves "michael," it asks the identical question MidLevel asks of
framings: match an existing cell or mint one. The alias-entity design is the match
primitive wearing entity clothes. Without this recognition the same verb set was
about to be built twice under different names; with it, resolution is the **second
instance** that earns (or refutes) the general layer.

## The new design lever: state-gated affordances

Affordances living on the cells means the **cell's state gates its verbs**:

- an **unstamped** framing affords `edit` freely — it's provisional, that's the point
- a **crystallized** node resists — `BirthAlignment` as **edit-resistance**, so
  proven ground can't be casually rewritten, and the **cascade becomes the only
  editor strong enough for crystallized structure**

This re-expresses the TMS / solid-ground design as inner-world physics: bedrock
isn't protected by rule, it's protected by cost. (Exact resistance function: open —
likely the same solid-ground threshold the cascade already stops at.)

## The load-bearing guard: affordances vs. physics transitions

Two kinds of change happen to memory, and they must never blur:

- **Agent affordances** — what an agent *may do*: `match`, `edit`, `new` a
  **prediction / framing / entity**. Proposals, revisions, attention.
- **Physics transitions** — what *happens to* cells regardless of anyone's say-so:
  crystallization (closure stamps a setpoint into existence), decay, eviction,
  cascade rollback.

An agent may `new` a prediction. It may **never** `new` a setpoint. Setpoints are
*born* from closure because they are predictions that came true — and truth cannot
be asserted into memory. This is **observed-not-asserted applied to the write
path**: agents propose, physics crystallizes. Lose this line and the LLM can
hallucinate proven experience — the one corruption the architecture exists to
prevent (immune-system/safety.md is downstream of exactly this).

So the affordance set is **Kind-scoped**: which verbs a cell offers depends on what
it is and what state it's in. That table (Kind × state → verbs) is the future
general layer's contract — to be written only after the second instance.

## Implementation posture (when the collapse comes — not yet)

The verbs ride the machinery that already exists: a memory primitive is an
`Impulse.Primitive` (`"match"` / `"edit"` / `"new"`) fired into the graph, exactly
as `"perceive"` / `"engage"` are fired into the world — same kernel, same delivery,
the cells' `Observe` dispatching on the primitive the way `world.Observe` does.
No new mechanism; the recursion was built for this. **But**: framings first
(specced), entities second (resolution), collapse third — the pattern being right
does not exempt it from spread-before-compress.

## Open

- The Kind × state → affordances table (the general contract; write after instance two)
- Edit-resistance function for crystallized nodes (solid-ground threshold? BirthAlignment-proportional cost?)
- Whether `match` is one verb or two (re-elect a framing vs. resolve an entity — same verb, different evidence; watch whether the implementations actually converge)
- Does `wait` have an inner analog (deliberate non-action on memory — "leave it lit, decide next cycle")? Unforced; note and ignore.

## Sources / related

- `notes.md` — "Custom primitives at any level" (the seed, v.2+ section)
- `specs/2026-07-03/framings.md` — instance one (match/edit/new on framings)
- `design-docs/cognition/token-web-and-entity-resolution.md` — instance two's home (resolution as match/new)
- `design-docs/cognition/error-recovery-loop.md` — revision as `edit` under grading physics; the cascade as the only editor of bedrock
- `design-docs/laws/invariants/` — observed-not-asserted; the write-path guard above is its memory-side application
- `design-docs/kernel/syscalls.md`, `world-kernel-coupling.md` — the outer primitives this mirrors

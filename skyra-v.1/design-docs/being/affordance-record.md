# AffordanceRecord — One Record, Kind Filled from the Affordance

**Date:** 2026-07-07 · **Status:** Direction (a call, not yet built). Standardizes the
*content* layer of `MemoryCell` the way `MemoryCell` already standardized the *node* layer.
**Related:** [cell-types.md](cell-types.md) (`MemoryCell` = one node, Kind is a role),
[perceive-internalization.md](../cognition/perceive-internalization.md) (PerceiveRecord/
EngageRecord — now Kinds of this), [token-web-and-entity-resolution.md](../cognition/token-web-and-entity-resolution.md)
(KindToken/KindAlias — also just Kinds), [emit-as-traversal.md](../kernel/emit-as-traversal.md)
(affordances are exercised by traversal). Echoes notes.md "Four Signals" (Kind is *selected*, not
predeclared) and "Unified graph retrieval" (everything is a node).

## The call

Stop typing records ahead of time. Every record a being lays down — an engage, a perceive, a
midlevel grade, a prediction, a gist — is **one struct, `AffordanceRecord`**, whose **`Kind`
field is filled from the affordance payload that came back**. We do not predeclare a Go type
per affordance; the Kind vocabulary grows from what the being actually does.

```
AffordanceRecord {
    Kind      string   // the affordance that produced it: "engage" | "perceive" |
                       // "midlevel-grade" | "predict" | ...  (filled, not predeclared)
    Payload   any      // the affordance's own result shape
    Timestamp int
    // (from/to, coord, score, etc. live in Payload, read by Kind at the use site)
}
```

## Why this is not a premature collapse

`MemoryCell` is **already** one node type: *"differentiated by Kind, Content, and topological
position,"* with `Kind` a **role tag, NOT a behavior switch**, and `Content any`. Given that,
introducing typed `PerceiveRecord` / `EngageRecord` / `MidlevelGradeRecord` structs is itself
the **premature specialization** — adding static structure the node was designed not to need.

So per overbuild→compress / spread-before-collapse: `AffordanceRecord` is the **spread-friendly
default** (decline to over-type), not the collapse. The typed-records path is the one that
fights the existing design. This note is "match the node layer's uniformity at the content
layer," nothing more.

## What it buys

1. **Dynamic accumulation of graph types.** A new affordance → a new `Kind` string at runtime,
   no new type, no enum edit. `midlevel:grade` starts producing records and they're just
   `Kind:"midlevel-grade"`. The graph's type system is **open** — the same spirit as
   "doesNotUnderstand as growth" and "custom primitives at any level" (notes.md). Kinds are
   *selected by the signal* (the four-signals note), not written ahead.

2. **Static lookups/queries without the kernel — as an escape hatch.** Uniform records with a
   `Kind` field make a plain index (`by Kind`, `by entity×Kind`) **cheap and possible** if a
   query pattern ever arrives that traversal doesn't serve well. We are **not building those
   indexes now** (see the caution) — the point is the *option* costs nothing once the record is
   uniform.

3. **One read shape.** Use sites go from N type assertions (`content.(Prediction)`,
   `content.(InteractionRecord)`, …) to **one assertion + a Kind switch**
   (`content.(AffordanceRecord)`, then dispatch on `.Kind`). Fewer types crossing the seam.

## The caution — don't rebuild the index we killed

[entities-and-surroundings.md](../cognition/entities-and-surroundings.md) deliberately removed
the separate index structure: *"there is no separate index structure; the reverse read is a
wave."* `AffordanceRecord` makes side-indexes **possible again** — that's optionality, not an
invitation. **Traversal stays the primary query path.** Build a `by-Kind` index only when a
concrete query demands one that the wave genuinely can't serve; otherwise the uniform Kind is
just latent capability, and we don't pay the write-time-maintenance cost of a side-index we
swore off.

## Trade accepted

`Payload` is `any` (as `Content` already is), so field access at read sites is Kind-gated
rather than statically typed — the same discipline the codebase already uses
(`mc.Content().(dna.Prediction)`). No new safety loss; the assertion just moves to one type +
a Kind switch. Keep Payload **inspectable** (a small typed struct per Kind *inside* Payload, or
a `map[string]any`) so use sites stay readable.

## `MemoryCell.Kind` **is** `AffordanceRecord.Kind` (decided)

They are the **same value** — one Kind, not two. The affordance stamps it once; it types the
node and tags the record identically. There is no separate record-Kind field to reconcile: the
`MemoryCell`'s `Kind` *is* the affordance's Kind, and `Content` holds the affordance's payload.
So "record" and "node" aren't two things wearing two Kinds — a `MemoryCell` of `Kind:"perceive"`
whose `Content` is the perceive payload **is** the `PerceiveRecord`.

## To pin when we build

- **Where the Kind is stamped** — the affordance/primitive result must be self-describing
  (carry its own name). World primitives already do (`Impulse.Primitive`); cognitive steps use
  the gene/step name (`sensory:intake`, `midlevel:frame`). So Kind = the producing step's name.
- **Payload shape** — one `map[string]any`, or a per-Kind typed struct held in Payload. Lean
  typed-struct-in-Payload for readability; decide at build.
- **Migration** — `InteractionRecord`/`ActionRecord`/`Prediction`/`Gist` are live content types
  today. Collapsing them is a follow-on, not a same-day change; do it when the record layer is
  next touched (the perceive build is the natural first site — `PerceiveRecord` is unbuilt, so
  it can be born as `AffordanceRecord{Kind:"perceive"}` directly).

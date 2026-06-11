# Kernel Change Plan — v.1

## Status

**Planning — discuss before implementing.** This maps the design locked this session (the [Perception Invariant](../invariants/perception-invariant.md), the `perceive` attention model, the dispatch decision, the [two-recursion split](activation-and-traversal.md#two-recursions)) against the kernel as committed at `5e7306c`. Items marked **[decided]** are settled; **[open]** items are collected at the end for discussion.

## Baseline — the kernel as it stands

`relation.go`: `Realize(root, payload)` binds a frame via `ActiveState[root.ID()]`, runs an iterative DFS over `Children()`, crosses into Content-Realities recursively (`r.realize(inner, …)`), and returns the root's Express output. `Relation` holds only `ActiveState` + `Energy`. `score()` reads the slice's records/threshold/count.

What predates the session's design: no origin/`Self`, no result delivery, no dispatch entry point, and none of the activation-physics correctness (decay, visited-once, tree buffering, energy) is wired.

## 1. Perception Invariant — signature + delivery  [decided]

Every wave must deliver its result to its origin, with `Self` carried in the signature.

- **Signature** (`reality.go`, `Kernel.Realize`): `Realize(root, payload any) any` → `Realize(root, self, payload any) any`. **The only change is adding `Self`** — the origin the result is delivered to, a required argument so the invariant is mechanical. The act/medium **stays `any`** (it already is): the kernel never reads it, its shape is endpoint knowledge (`world.observe` and the executing cells treat it as the string act, but the kernel carries it opaquely). `Observe`/`Express` keep their `any`. Breaking interface change (pre-1.0, fine). Parameter names being disambiguated — see Naming.
- **Delivery**: at the end of `realize`, after the loop, call `self.Observe(r, result)` — the reafference. Today the result only returns up the Go stack; nothing perceives it.
- **Recursive origin**: the cross-plane call sets the inner wave's origin to the inner reality itself — `r.realize(inner, inner, content, …)`. The inner being perceives its own interior wave's output. The invariant holding recursively, in code.
- **Result representation**: minimally the root's Express output; properly a per-traversal **result field on the slice** (the "record of what it did" decision). Origin reads it.

Landing this fully needs a `Self` to deliver to, which at the top level comes from `world.observe(self, content)` (group 2) — so it pulls in the World.

## 2. Dispatch — world.observe entry point  [partly open]

- The kernel does **not** auto-dispatch Express returns today — it captures only the root return. Already matches the superseded "no dispatch-via-return" decision; nothing to remove. **[decided]**
- `Express` keeps returning output (`any`), never a wave-start trigger. **[decided]**
- New, world-side: `world.Observe(self, content)` spins a kernel wave — a goroutine (see Resolved) — when `content` is present, threading `self` as origin. **[open — build]** — spawn-vs-cross is resolved (always spawn); what remains is building the World Reality, which doesn't exist in code yet.

## 3. Activation physics — the .1001 correctness layer  [decided, all known fixes]

None of this is in the code. Without it, .1001's activation numbers are meaningless.

- **Double-bump**: `score()` activates each cell twice — once as a child (survivor selection), once as self (node score). A descended cell is bumped twice per traversal. Fix: a non-root cell's self-score is the value its parent already computed for it as a child — pass it into the Node; activate self only for the root.
- **Decay**: `Activate(false)` is never called; bump-only EMAs flatten toward 1.0. Needs `LastChildren` on the Slice (what each cell reported last pass), diffed against the current report; absent children get `Activate(false)`.
- **Tree double-buffer**: the ascent builds `parent.Children` and drops it; `slice.Tree()` Prior/Current never populate. At traversal start `Prior=Current, Current=nil`; on root pop, write the tree to `Current`.
- **Visited-once / cycle brake**: the grid is bidirectionally wired; energy is the only thing stopping A→B→A, which the locked decision forbids (energy is the rail, not the mechanism). Add stack-membership (`onPath`) to block re-descending a node on the current path. Directionality-via-attention is the richer version — future; visited-once suffices for .1001's single being.
- **Energy minting**: `NewRelation` leaves `Energy=0`, so nothing descends. Seed it at wave creation.

## 4. Supporting interface additions

- **Slice** (`reality.go`): `LastChildren()` + setter (decay); `Result()` / `SetResult()` (record-what-it-did). Future, not .1001: `Orientation()` — the attention aperture the kernel reads for the directionality activation term.
- **Reality**: `Observe` eventually dual-mode (descent-intake vs result-reafference, per [episodic-encoding](../cognition/episodic-encoding.md)). For now reuse `Observe` with the result as payload; the cell branches. Refinement.

## Sequencing

- **Path A — invariant first**: add `Self` + `origin.observe`. Architecturally correct keystone, but pulls in the World (`world.observe`), gated on the spawn-vs-cross lock.
- **Path B — activation physics first**: double-bump, decay, tree buffer, visited-once, energy, plus the two Slice additions. Pure kernel, no new design decisions, no World needed, and it's what unblocks a measurable .1001.

**Recommendation: Path B.** Self-contained, every item a known fix, turns the skeleton into something whose numbers mean something. The invariant signature is small and lands cleanly once `world.observe` is locked; it does not block the experiment.

## Naming — to disambiguate

The act/medium type is settled (`any`). What's left is naming: **"content" is overloaded.** The four roles in and around the kernel signature, with their jobs, to be named cleanly (see discussion):

| Role / job | Current / sketched name | Issue |
|---|---|---|
| The reality this wave traverses (its root substrate) | `root` (code) / `ContentReality` (sketch) | "Content" in the sketched name |
| The origin — who perceives the result, delivery target | `Self` | clear, no collision |
| The act/message the wave carries and executes (`any`, opaque to kernel) | `payload` (code) / `content` (sketch) | "content" collides with `Content()` |
| The cell's nested slot — may hold a Reality → boundary crossing | `Content() any` (Reality method) | source of the collision |

## Resolved this session

- **Spawn vs cross → always spawn at the actor boundary.** Every wave is a goroutine. In-plane traversal (a being's own `Children`) is the explicit stack, sequential. Crossing into a `Content` Reality is crossing into another actor (it has its own slice), so it **spawns a goroutine** — async — and that actor delivers its result to its own origin; nothing bubbles back. A response is a separate wave. Sync exists only within an actor; sync-vs-async is a placement decision (`Child` vs `Content`). This changes the content-check from `r.realize(inner, …)` to `go realize(inner, inner, …)`, and makes concurrency-safety (slice mutex, shared-substrate reads) real work, not later. See [Two Recursions](activation-and-traversal.md#two-recursions).

## Open questions

1. **Result representation** — root Express output, or a dedicated per-traversal result field on the slice? The "record of what it did" decision leans toward the slice field.
3. **Result delivery and Observe** — does `self.Observe(result)` overload the descent `Observe`, or do we add the dual-mode reafference entry (episodic-encoding) now rather than later?
4. **Energy budget** — what seeds it, and what value/curve? Per-wave constant for now, or attenuation-based (Phase 23)?
5. **Path choice** — A or B first.

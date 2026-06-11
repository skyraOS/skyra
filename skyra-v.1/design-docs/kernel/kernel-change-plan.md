# Kernel Change Plan — v.1

## Status

**Planning — discuss before implementing.** This maps the design locked this session (the [Perception Invariant](../invariants/perception-invariant.md), the `perceive` attention model, the dispatch decision, the [two-recursion split](activation-and-traversal.md#two-recursions)) against the kernel as committed at `5e7306c`. Items marked **[decided]** are settled; **[open]** items are collected at the end for discussion.

## Baseline — the kernel as it stands

`relation.go`: `Realize(root, payload)` binds a frame via `ActiveState[root.ID()]`, runs an iterative DFS over `Children()`, crosses into Content-Realities recursively (`r.realize(inner, …)`), and returns the root's Express output. `Relation` holds only `ActiveState` + `Energy`. `score()` reads the slice's records/threshold/count.

What predates the session's design: no origin/`Self`, no result delivery, no dispatch entry point, and none of the activation-physics correctness (decay, visited-once, tree buffering, energy) is wired.

## 1. Perception Invariant — signature + delivery  [decided]

Every wave must deliver its result to its origin, with `Self` carried in the signature.

- **Signature** (`reality.go`, `Kernel.Realize`): `Realize(root, payload any) any` → `Realize(upperReality, lowerReality, impulse any) any`. `UpperReality` is the wave's **owner** (the being whose wave it is) — threaded unchanged through crossings, the discriminator passed to `Observe`, and who the result returns to. `LowerReality` is the **realm** traversed at this level (the world, then a crossed-into being's interior); it descends at each crossing while Upper stays put. The act/medium **stays `any`** (it already is): the kernel never reads it, its shape is endpoint knowledge (`world.observe` and the executing cells treat it as the string act, but the kernel carries it opaquely). `Observe`/`Express` keep their `any` (renamed `payload` → `impulse`). Breaking interface change (pre-1.0, fine). Names locked — see Naming.
- **Delivery**: at the end of `realize`, after the loop, call `self.Observe(r, result)` — the reafference. Today the result only returns up the Go stack; nothing perceives it.
- **Recursive origin**: the cross-plane spawn threads `UpperReality` unchanged and sets `LowerReality` to the inner reality — `go realize(upper, inner, impulse)` — so the inner being reads the wave as external (`upper != my being`) and the result returns to the owner. A being perceives waves *it* owns (upper = itself → reafference); waves that cross in from outside are exafference. The invariant holding recursively, in code. (See Resolved: crossing semantics.)
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
- **Visited-once = the rail.** The grid is bidirectionally wired; visited-once is what guarantees termination (a traversal can't visit more nodes than exist). Add stack-membership (`onPath`) to block re-descending a node on the current path. This *is* the runaway safety — it replaces the step-counter (below). Directionality-via-attention is the richer version — future; visited-once suffices for .1001's single being.
- **Energy → amplitude (not a step-counter).** Drop `r.Energy--` and the `Energy > 0` descent gate. Energy is the wave's **amplitude / signal strength**: set by the agent at spawn (its investment — for v.1 a generous constant, later Motivation/Emotion-driven), it **folds into activation** (`activation = base × amplitude`) and **attenuates** as the wave propagates. The wave stops *naturally* when amplitude drops `activation` below threshold — no separate budget gate. Locked at the start (initial value fixed; only attenuates after). Amplitude also attenuates **across crossings**, so a sub-wave inherits the attenuated value — this is what bounds the goroutine tree (too quiet → don't spawn the crossing). Open tuning sub-question: attenuate per *step* or per *distance / activation-cost* (structural model is settled either way).

## 4. Supporting interface additions

- **Slice** (`reality.go`): `LastChildren()` + setter (decay); `Result()` / `SetResult()` (record-what-it-did). Future, not .1001: `Orientation()` — the attention aperture the kernel reads for the directionality activation term.
- **Reality**: `Observe` eventually dual-mode (descent-intake vs result-reafference, per [episodic-encoding](../cognition/episodic-encoding.md)). For now reuse `Observe` with the result as payload; the cell branches. Refinement.

## Sequencing

- **Path A — invariant first**: add `Self` + `origin.observe`. Architecturally correct keystone, but pulls in the World (`world.observe`), gated on the spawn-vs-cross lock.
- **Path B — activation physics first**: double-bump, decay, tree buffer, visited-once, energy, plus the two Slice additions. Pure kernel, no new design decisions, no World needed, and it's what unblocks a measurable .1001.

**Recommendation: Path B.** Self-contained, every item a known fix, turns the skeleton into something whose numbers mean something. The invariant signature is small and lands cleanly once `world.observe` is locked; it does not block the experiment.

## Naming — resolved

The act/medium type is settled (`any`); the names are locked:

| Role / job | Name |
|---|---|
| The wave's **owner** — the being whose wave it is; threaded unchanged, the `Observe` discriminator, the result recipient | `UpperReality` |
| The **realm** traversed at this level (world → crossed-into being's interior); descends at each crossing | `LowerReality` |
| The act/message the wave carries and executes (`any`, opaque to kernel) | `impulse` |
| The cell's nested slot — may hold a Reality → boundary crossing | `Content()` (unchanged) |

`UpperReality` / `LowerReality` capture the directional intuition — the owner reaches *down* into the realm — and Upper is threaded unchanged while Lower descends. Upper is the **owner, not the structural container** (skyra sits inside the world but owns the wave). Order is owner-first: `realize(UpperReality, LowerReality, impulse)`. `impulse` matches the spec's existing language and v.05's `Relation.Impulse`, pairing with `result` — impulse in, result out. `Content()` keeps the word "content" (it genuinely means *what the cell contains*).

Signature: **`Realize(UpperReality, LowerReality, impulse any) any`**.

## Resolved this session

- **Spawn vs cross → always spawn at the actor boundary.** Every wave is a goroutine. In-plane traversal (a being's own `Children`) is the explicit stack, sequential. Crossing into a `Content` Reality is crossing into another actor (it has its own slice), so it **spawns a goroutine** — async. This changes the content-check from `r.realize(inner, …)` to `go realize(upper, inner, …)` (`UpperReality` threaded, `LowerReality` = the inner reality), and makes concurrency-safety (slice mutex, shared-substrate reads) real work, not later. See [Two Recursions](activation-and-traversal.md#two-recursions).
- **Crossing semantics / origin threading → thread `UpperReality`.** The spawned wave keeps `UpperReality` (the owner) unchanged and sets `LowerReality` to the inner reality (`go realize(upper, inner, …)`). So the inner actor's cells read the wave as external (`upper != my being` → exafference), and the result returns to the owner, not the inner actor. The inner actor's interior runs on its own goroutine — concurrent and non-blocking — merely *labeled* with the external owner (no boundary violation). A being processes another's wave (labeled external, result to the owner); a being responds with its *own* new wave (it becomes the `UpperReality`).
- **Result representation → a dedicated per-traversal result field on the slice.** The wave's result is written to the origin's slice (`Result` / `SetResult`) — for an agent, the action it just fired and its outcome; for the world, the result of primitive execution. The being's record of what it did and the perception-invariant delivery target in one.
- **Result delivery → `origin.observe(result)` is the ending wave condition.** When a wave completes, the kernel calls the origin's `Observe` with the result — the actor that started the wave perceives what it produced. **Reuses `Observe`** (no separate method): the origin observes at the *start* (the impulse it sends — efference) and at the *end* (the result that returns — reafference). That is exactly episodic-encoding's dual-mode Observe, and the prediction-error delta lives in the bracket between them. Result content is primitive-specific and opaque to the kernel (`any`): `perceive` → the collected perception; `engage` → engaged / missed. Delivered via `Observe`, recorded on the slice `Result` field.
  - **Sub-detail (a) → resolved by identity.** `Observe` carries the wave's owner: `Observe(upper, impulse any)`, where `upper` is the `UpperReality` threaded through `realize`. The cell compares it to its own **being** (`upper.ID() == myBeingID`, the slice's `BeingID` — not the individual cell): upper is my being → the wave is mine (efference/reafference); upper is someone else → external (exafference). No flag — identity is the discriminator (von Holst). This makes the old `k Kernel` param on `Observe` vestigial (cells don't start waves during intake).
  - Sub-detail (b): latest-wins vs queue when multiple concurrent waves deliver to one origin (mutex-guarded either way).

## Open questions

1. **Path choice** — A or B first.

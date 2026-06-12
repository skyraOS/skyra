# Kernel Change Plan — v.1

## Status

**Planning — discuss before implementing.** This maps the design locked this session (the [Perception Invariant](../invariants/perception-invariant.md), the `perceive` attention model, the dispatch decision, the [two-recursion split](activation-and-traversal.md#two-recursions)) against the kernel as committed at `5e7306c`. Items marked **[decided]** are settled; **[open]** items are collected at the end for discussion.

## Baseline — the kernel as it stands

`relation.go`: `Realize(root, payload)` binds a frame via `ActiveState[root.ID()]`, runs an iterative DFS over `Children()`, crosses into Content-Realities recursively (`r.realize(inner, …)`), and returns the root's Express output. `Relation` holds only `ActiveState` + `Energy`. `score()` reads the slice's records/threshold/count.

What predates the session's design: no origin/`Self`, no result delivery, no dispatch entry point, and none of the activation-physics correctness (decay, visited-once, tree buffering, energy) is wired.

## 1. Perception Invariant — signature + delivery  [decided]

Every wave must deliver its result to its origin, with `Self` carried in the signature.

- **Signature** (`reality.go`, `Kernel.Realize`): `Realize(root, payload any) any` → `Realize(originReality, traversedReality, impulse any) any`. `OriginReality` is the wave's **owner** (the being whose wave it is) — threaded unchanged through crossings, the discriminator passed to `Observe`, and who the result returns to. `TraversedReality` is the **realm** this wave traverses; if traversal finds another Reality in `Content`, that discovery starts a *new* wave rather than making this one wait. The act/medium **stays `any`** (it already is): the kernel never reads it, its shape is endpoint knowledge (`world.observe` and the executing cells treat it as the string act, but the kernel carries it opaquely). `Observe`/`Express` keep their `any` (renamed `payload` → `impulse`). Breaking interface change (pre-1.0, fine). Names locked — see Naming.
- **Delivery**: at the end of `realize`, after the loop, call `origin.Observe(r, result)` — the reafference (`origin` = `OriginReality`). Today the result only returns up the Go stack; nothing perceives it.
- **Recursive origin**: the cross-plane spawn threads `OriginReality` unchanged and sets `TraversedReality` to the inner reality — `go realize(origin, inner, impulse)` — so the inner being reads the wave as external (`origin != my being`) and the result returns to the owner. A being perceives waves *it* owns (origin = itself → reafference); waves that cross in from outside are exafference. The invariant holding recursively, in code. (See Resolved: crossing semantics.)
- **Result representation**: minimally the root's Express output; properly a per-traversal **result field on the slice** (the "record of what it did" decision). Origin reads it.

Landing this fully needs a `Self` to deliver to, which at the top level comes from `world.observe(self, content)` (group 2) — so it pulls in the World.

## 2. Dispatch — world.observe entry point  [partly open]

- The kernel does **not** auto-dispatch Express returns today — it captures only the root return. Already matches the superseded "no dispatch-via-return" decision; nothing to remove. **[decided]**
- `Express` keeps returning output (`any`), never a wave-start trigger. **[decided]**
- New, world-side: `world.Observe(self, content)` spins a kernel wave — a goroutine (see Resolved) — when `content` is present, threading `self` as origin. **[open — build]** — spawn-vs-cross is resolved (always spawn); what remains is building the World Reality, which doesn't exist in code yet.

## 3. Activation physics — the .1001 correctness layer  [Path B — mostly DONE at `5067903`]

The kernel correctness layer, implemented and tested (mock-based) in `relation_test.go`. Without it, .1001's activation numbers are meaningless.

- **Double-bump → DONE.** `score()` was activating each cell twice (once as a child, once as self). Fixed: survivors are pre-scored `*Node`s (`Node.Survivors []*Node`); a survivor's `Score` carries into its own node when descended. Root scored once. Each cell activates exactly once.
- **Visited-once = the rail → DONE.** An `onPath` set blocks re-descending a node already on the current path — cycles terminate; multi-path reachability preserved. This is the runaway safety; it replaced the step-counter.
- **Energy → amplitude → DONE.** Dropped `r.Energy--` and the `Energy > 0` gate. Each `Node` carries an amplitude that attenuates per depth (0.9) and folds into the score (`activation = base × amplitude`); a quiet wave fails threshold and stops. Locked at spawn (`Relation.Amplitude`, a generous constant for now; later Motivation/Emotion-driven). It attenuates across crossings too, bounding the goroutine tree. Open tuning sub-question: attenuate per *step* or per *distance/activation-cost* (structural model settled either way).
- **Tree double-buffer → DONE.** `Prior=Current, Current=nil` at start; the finished tree written to `Current` on root pop.
- **Decay → DEFERRED (not on .1001's critical path).** `Activate(false)` is still never called. The archived `BaseActivation` shows why this can wait: `activation = Identity × Familiarity × recency`, where every *scored* cell bumps its EMAs toward 1.0 and **recency** does the dormant-decay (it self-corrects from traversal count when a cell is next scored). `.1001` is a *single traversal over pre-set weights*, so cross-traversal decay never fires — the first measurement doesn't need it. It matters for multi-traversal learning, and carries a real semantic fork (does `Activate(false)` fire on *absent children* per the doc, or on *non-survivors* to keep EMAs from flattening?) to settle then.

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
| The wave's **owner** — the being whose wave it is; threaded unchanged, the `Observe` discriminator, the result recipient | `OriginReality` |
| The **realm** this wave traverses (the world, then a crossed-into being's interior) | `TraversedReality` |
| The act/message the wave carries and executes (`any`, opaque to kernel) | `impulse` |
| The cell's nested slot — may hold a Reality → boundary crossing | `Content()` (unchanged) |

`OriginReality` is the **owner, not the structural container** (skyra sits inside the world but owns the wave); it's threaded unchanged across crossings. `TraversedReality` is the realm; finding a Reality in its `Content` starts a *new* wave, not a deeper leg of this one. Order is owner-first: `realize(OriginReality, TraversedReality, impulse)`. `impulse` matches the spec's existing language and v.05's `Relation.Impulse`, pairing with `result` — impulse in, result out. `Content()` keeps the word "content" (it genuinely means *what the cell contains*).

Signature: **`Realize(OriginReality, TraversedReality, impulse any) any`**.

## Resolved this session

- **Spawn vs cross → always spawn at the actor boundary.** Every wave is a goroutine. In-plane traversal (a being's own `Children`) is the explicit stack, sequential. Crossing into a `Content` Reality is crossing into another actor (it has its own slice), so it **spawns a goroutine** — async. This changes the content-check from `r.realize(inner, …)` to `go realize(origin, inner, …)` (`OriginReality` threaded, `TraversedReality` = the inner reality), and makes concurrency-safety (slice mutex, shared-substrate reads) real work, not later. See [Two Recursions](activation-and-traversal.md#two-recursions).
- **Crossing semantics / origin threading → thread `OriginReality`.** The spawned wave keeps `OriginReality` (the owner) unchanged and sets `TraversedReality` to the inner reality (`go realize(origin, inner, …)`). So the inner actor's cells read the wave as external (`origin != my being` → exafference), and the result returns to the owner, not the inner actor. The inner actor's interior runs on its own goroutine — concurrent and non-blocking — merely *labeled* with the external owner (no boundary violation). A being processes another's wave (labeled external, result to the owner); a being responds with its *own* new wave (it becomes the `OriginReality`).
- **Result representation → a dedicated per-traversal result field on the slice.** The wave's result is written to the origin's slice (`Result` / `SetResult`) — for an agent, the action it just fired and its outcome; for the world, the result of primitive execution. The being's record of what it did and the perception-invariant delivery target in one.
- **Result delivery → `origin.observe(result)` is the ending wave condition.** When a wave completes, the kernel calls the origin's `Observe` with the result — the actor that started the wave perceives what it produced. **Reuses `Observe`** (no separate method): the origin observes at the *start* (the impulse it sends — efference) and at the *end* (the result that returns — reafference). That is exactly episodic-encoding's dual-mode Observe, and the prediction-error delta lives in the bracket between them. Result content is primitive-specific and opaque to the kernel (`any`): `perceive` → the collected perception; `engage` → engaged / missed. Delivered via `Observe`, recorded on the slice `Result` field.
  - **Sub-detail (a) → resolved by identity.** `Observe` carries the wave's owner: `Observe(origin, impulse any)`, where `origin` is the `OriginReality` threaded through `realize`. The cell compares it to its own **being** (`origin.ID() == myBeingID`, the slice's `BeingID` — not the individual cell): origin is my being → the wave is mine (efference/reafference); origin is someone else → external (exafference). No flag — identity is the discriminator (von Holst). This makes the old `k Kernel` param on `Observe` vestigial (cells don't start waves during intake).
  - Sub-detail (b): latest-wins vs queue when multiple concurrent waves deliver to one origin (mutex-guarded either way).

## Open questions

*(none structural — the spec is settled)*

Remaining work, not open questions:
- **Path B finish** — decay (deferred above), and the two delivery sub-details (latest-vs-queue; the reafference Observe). None block .1001's first measurement.
- **Path A** — the World Reality + `world.observe`, the goroutine spawn at the content-check, and the `Realize(OriginReality, …)` signature + `origin.observe` delivery. Needed for multi-being / external action, not for single-being .1001.
- **To run .1001** — promote `BaseActivation` to the live `skyra` package; build the fake-memory harness; read the activation product off the tree.

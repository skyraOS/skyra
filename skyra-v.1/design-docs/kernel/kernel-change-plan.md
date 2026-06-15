# Kernel Change Plan — v.1

## Status

**Planning — discuss before implementing.** This maps the design locked this session (the [Perception Invariant](../laws/invariants/perception-invariant.md), the `perceive` attention model, the dispatch decision, the [two-recursion split](activation-and-traversal.md#two-recursions)) against the kernel as committed at `5e7306c`. Items marked **[decided]** are settled; **[open]** items are collected at the end for discussion.

## Baseline — the kernel as it stands

`relation.go` at `5e7306c`: `Realize(root, payload)` bound a frame via `ActiveState[root.ID()]`, ran an iterative DFS over `Children()`, crossed into Content-Realities recursively (`r.realize(inner, …)`), and returned the root's Express output. `Relation` held only `ActiveState` + `Energy`. `score()` read the slice's records/threshold/count.

What predates the session's design: no origin/`Self`, no result delivery, no dispatch entry point, and none of the activation-physics correctness (decay, visited-once, tree buffering, energy) is wired.

## 1. Perception Invariant — signature + delivery  [Path A1 — DONE at `9c86cbe`]

The invariant is mechanical in the kernel, tested (mock-based, in `relation_test.go`).
Every wave delivers its result to its origin, with the owner carried in the signature.

- **Signature** (`reality.go`, `Kernel.Realize`): `Realize(root, payload any) any` → `Realize(originReality, traversedReality, impulse any)`. `OriginReality` is the wave's **owner** (the being whose wave it is) — the discriminator passed to `Observe`, and who receives this wave's result. `TraversedReality` is the **realm** this wave traverses; if traversal finds another Reality in `Content`, that discovery starts an independent wave rather than making this one wait. The act/medium **stays `any`** (it already is): the kernel never reads it, its shape is endpoint knowledge (`world.observe` and the executing cells treat it as the string act, but the kernel carries it opaquely). `Observe`/`Express` keep their `any` (renamed `payload` → `impulse`). Breaking interface change (pre-1.0, fine). Names locked — see Naming.
- **Delivery**: at the end of `realize`, after the loop, call `origin.Observe(r, result)` — the reafference (`origin` = `OriginReality`). Today the result only returns up the Go stack; nothing perceives it.
- **Boundary spawn**: the cross-plane spawn starts a new wave for the inner actor — `go realize(inner, inner, impulse)`. The outer wave continues toward its own result and does not wait on the inner actor. The spawned wave delivers to its own origin, not to the parent wave's origin. A response from the inner actor is a later chosen send, not the return value of the parent wave. (See Resolved: crossing semantics.)
**Done in A1/A2:** signature shape (`Realize(origin, traversed, impulse any)`); `Observe`/`Express` take `(origin Reality, impulse any)` (kernel param dropped); `origin.Observe(origin, result)` delivery at every wave's end; cross-actor spawn is independent (`go realize(inner, inner, impulse)`), non-waiting, and does not deliver the inner result to the parent origin.

**Remaining:** the per-traversal **`Result` field on the slice** (record-what-it-did) — delivery currently goes via `origin.Observe`; the slice `Result` field for introspection is not yet wired. Small.

## 2. Dispatch — world.observe entry point (A3)  [goroutine DONE; world.observe to build]

- The kernel does **not** auto-dispatch Express returns; `Express` returns output, never a wave-start. **[decided, holds in code]**
- **Goroutine spawn at the crossing → DONE (A2).** `go realize(inner, inner, impulse)`, slice-locked, independent, non-waiting. The parent wave does not receive the child wave's result.
- **`world.observe(self, content)` entry point → A3.** A being acts outward by calling into the world it stands on, which spawns `realize(self, pointOf(self), content)` — the world threads `self` as the **owner** but the being's **location** as the traversed root (a doorway, not a destination: the world relays, it is not itself the realm). The result returns to `self` via the kernel's delivery. Resolved design:
  - **No new verb.** `observe` is the universal "receive a signal," with different competence per Reality: the world's `observe` *facilitates* (relays to the kernel); a memory's absorbs; the origin's perceives the result. `world.observe` (world receives the act) and `self.observe` (being receives the result) are the same uniform method on different objects — not a collision. The world is a facilitator.
  - **Recursion break — origin-presence split.** `world.observe` would loop if relaying were unconditional (the kernel walks the world as space, which re-calls `observe`). The world splits on the presence of the origin label: `origin != nil` → a being is acting → relay to the kernel; `origin == nil` → being walked as a realm → intake/propagate, never relay. The kernel strips the self label on the traversal path, disarming facilitation. No kernel special-case; the break lives in the world's competence. See [slice-and-world.md](../world/slice-and-world.md#why-it-cant-recurse).
  - **Reach: the environment reference.** Each cell holds a reference to the layer below it (its environment), set at registration — "a reference to what it's standing on." A being acts via `self.Environment.observe(self, content)`. The world holds the kernel; the chain is being → world → kernel → slices.
  - **Plane slices.** Registration wires the being's environment ref (down, to act) and assigns Realities to their slice plane. The kernel does not hold `ActiveState`; it binds active state through `traversed.Slice()`.
  - **The kernel is stateless** — it does not persist slice references. Slices belong to Reality planes; the kernel traverses inactive Reality and writes active state into the traversed Reality's plane Slice.
  - **Build remaining:** the World Reality (grid `Children`, the `Reality` methods, holding the kernel). Overlaps with migrating a minimal being layer (the current DNA imports the deleted `src/reality` and doesn't compile) — the same work the .1001 harness needs. See [slice-and-world.md](../world/slice-and-world.md#how-a-being-acts).

## 3. Activation physics — the .1001 correctness layer  [Path B — mostly DONE at `5067903`]

The kernel correctness layer, implemented and tested (mock-based) in `relation_test.go`. Without it, .1001's activation numbers are meaningless.

- **Double-bump → DONE.** `score()` was activating each cell twice (once as a child, once as self). Fixed: survivors are pre-scored `*Node`s (`Node.Survivors []*Node`); a survivor's `Score` carries into its own node when descended. Root scored once. Each cell activates exactly once.
- **Visited-once = the rail → DONE.** An `onPath` set blocks re-descending a node already on the current path — cycles terminate; multi-path reachability preserved. This is the runaway safety; it replaced the step-counter.
- **Energy → amplitude → DONE.** Dropped `r.Energy--` and the `Energy > 0` gate. Each `Node` carries an amplitude that attenuates per depth (0.9) and folds into the score (`activation = base × amplitude`); a quiet wave fails threshold and stops. Locked at spawn (`Relation.Amplitude`, a generous constant for now; later Motivation/Emotion-driven). It attenuates across crossings too, bounding the goroutine tree. Open tuning sub-question: attenuate per *step* or per *distance/activation-cost* (structural model settled either way).
- **Tree double-buffer → DONE.** `Prior=Current, Current=nil` at start; the finished tree written to `Current` on root pop.
- **Decay → DEFERRED (not on .1001's critical path).** `Activate(false)` is still never called. The archived `BaseActivation` shows why this can wait: `activation = Identity × Familiarity × recency`, where every *scored* cell bumps its EMAs toward 1.0 and **recency** does the dormant-decay (it self-corrects from traversal count when a cell is next scored). `.1001` is a *single traversal over pre-set weights*, so cross-traversal decay never fires — the first measurement doesn't need it. It matters for multi-traversal learning, and carries a real semantic fork (does `Activate(false)` fire on *absent children* per the doc, or on *non-survivors* to keep EMAs from flattening?) to settle then.

## 4. Supporting interface additions

- **Slice** (`reality.go`): `Lock()` / `Unlock()` → **DONE (A2)** — the kernel owns wave locking. Remaining: `Result()` / `SetResult()` (record-what-it-did, small); `LastChildren()` + setter (decay, deferred). Future, not .1001: `Orientation()` — the attention aperture for the directionality activation term.
- **Reality**: `Observe(origin, impulse)` → **DONE (A1)**. Remaining `[A3]`: an `Environment` reference — set at registration ("a reference to what it's standing on"), used for outward action (`self.Environment.observe(self, content)`). Uniform: every Reality carries it (nil for leaf cells).
- **Registration / boot `[A3]`**: assign Realities to their slice plane and wire the environment reference down. The kernel does not hold `ActiveState`; it binds active state through `traversed.Slice()`.

## Sequencing

- **Path A — invariant first**: the signature + `origin.observe` delivery. Architecturally the keystone, but pulls in the World (`world.observe`).
- **Path B — activation physics first**: double-bump, decay, tree buffer, visited-once, energy, plus the two Slice additions. Pure kernel, no new design decisions, no World needed, and it's what unblocks a measurable .1001.

**Chosen: B, then A1/A2.** Path B landed first (`5067903`), then the invariant signature + delivery (A1), stateless slice binding, and independent goroutine spawn (A2) followed. A3 (`world.observe` + the World) is the remaining piece.

## Naming — resolved

The act/medium type is settled (`any`); the names are locked:

| Role / job | Name |
|---|---|
| The wave's **owner** — the being whose wave it is; the `Observe` discriminator, the result recipient for this wave | `OriginReality` |
| The **realm** this wave traverses (the world, then a crossed-into being's interior) | `TraversedReality` |
| The act/message the wave carries and executes (`any`, opaque to kernel) | `impulse` |
| The cell's nested slot — may hold a Reality → boundary crossing | `Content()` (unchanged) |

`OriginReality` is the **owner, not the structural container** (skyra sits inside the world but owns its own wave). `TraversedReality` is the realm. Finding a Reality in `Content` starts a *new* wave, not a deeper leg of this one, and the new wave has its own origin. Order is owner-first: `realize(OriginReality, TraversedReality, impulse)`. `impulse` matches the spec's existing language and v.05's `Relation.Impulse`, pairing with `result` — impulse in, result out. `Content()` keeps the word "content" (it genuinely means *what the cell contains*).

Signature: **`Realize(OriginReality, TraversedReality, impulse any)`**.

## Resolved this session

- **Spawn vs cross → always spawn at the actor boundary.** Every wave is a goroutine. In-plane traversal (a being's own `Children`) is the explicit stack, sequential. Crossing into a `Content` Reality is crossing into another actor (it has its own slice), so it **spawns a goroutine** — async. This changes the content-check from `r.realize(inner, …)` to `go realize(inner, inner, …)` — an independent wave whose origin is the inner actor. Concurrency-safety (slice mutex, shared-substrate reads) is real work, not later. See [Two Recursions](activation-and-traversal.md#two-recursions).
- **Crossing semantics / origin reset → new wave, new owner.** The spawned wave does not keep the parent wave's `OriginReality`. It starts with the inner actor as both `OriginReality` and `TraversedReality` (`go realize(inner, inner, …)`). The parent wave continues and delivers only its own result to its own origin. The inner actor's interior runs on its own goroutine — concurrent and non-blocking. A being responds with its *own* new wave; it does not return a response into the parent wave.
- **Result representation → a dedicated per-traversal result field on the slice.** The wave's result is written to the origin's slice (`Result` / `SetResult`) — for an agent, the action it just fired and its outcome; for the world, the result of primitive execution. The being's record of what it did and the perception-invariant delivery target in one.
- **Result delivery → `origin.observe(result)` is the ending wave condition.** When a wave completes, the kernel calls the origin's `Observe` with the result — the actor that started the wave perceives what it produced. **Reuses `Observe`** (no separate method): the origin observes at the *start* (the impulse it sends — efference) and at the *end* (the result that returns — reafference). That is exactly episodic-encoding's dual-mode Observe, and the prediction-error delta lives in the bracket between them. Result content is primitive-specific and opaque to the kernel (`any`): `perceive` → the collected perception; `engage` → engaged / missed. Delivered via `Observe`, recorded on the slice `Result` field.
  - **Sub-detail (a) → resolved by nil-vs-named, not identity comparison.** Intake `Observe` carries a **nil** origin: during traversal the kernel walks the realm anonymously (`traversed.Observe(nil, impulse)`, `child.Observe(nil, impulse)`), so a cell never compares an origin to its own being. The owner is named only at the terminal delivery (`origin.Observe(origin, result)`), where `origin ==` the recipient — that is how a being recognizes its own wave's result (reafference). The rule is universal: **nil origin = being traversed; named origin = invoked-or-perceived.** The earlier von-Holst framing (cell compares `origin.ID()` to its being to tell efference from exafference *during traversal*) is retired — interior traversal has no foreign origin to compare against. Cross-being exafference, when it exists, arrives at the environment boundary (`world.Observe(self, …)`), not inside a being's interior.
  - Sub-detail (b): latest-wins vs queue when multiple concurrent waves deliver to one origin (mutex-guarded either way).

## Open questions

- **`Express` origin asymmetry → OPEN.** Intake `Observe` carries `nil` (the kernel walks anonymously), but `Express` is still called with the wave owner (`top.Cell.Express(origin, impulse)`, relation.go). `Express` runs *during* the traversal (the ascent), so the "anonymous while being walked" principle would argue it should be `nil` too — yet a cell producing output may legitimately need to know the owner. Left open deliberately: either Express goes `nil` (full anonymity during traversal), or the divergence is intentional and the docs must say *why* intake is anonymous while expression is owner-aware.

Remaining work, not open questions:
- **Join → DEFERRED (let it earn its place).** The stateless kernel fire-and-forgets spawned sub-waves — no `WaitGroup`, no join handle. Nothing waits on a crossing's outcome, so `main` can exit before sub-waves finish and there's no backpressure bounding the goroutine tree. Deferred deliberately: it resolves one of two ways and we don't pick now — either a concrete need to wait surfaces and *that* defines the join's shape, or sub-waves just deposit into their own slices and "doneness" becomes readable accumulated state (join never needs to exist). `.1001` is single-being internal traversal with no crossings on its path, so this costs nothing now; building it now would guess a shape the architecture might dissolve. The complexity has to live somewhere — we let it show up rather than pre-place it.
- **Path B finish** — decay (deferred above), and the two delivery sub-details (latest-vs-queue; the reafference Observe). None block .1001's first measurement.
- **Path A** — A1 (signature + delivery) and A2 (independent goroutine spawn + stateless slice binding) **DONE**. Remaining: **A3** — the World Reality + `world.observe`, the `Environment` ref, and plane-slice registration (see §2). Needed for multi-being / external action, not for single-being .1001.
- **To run .1001** — promote `BaseActivation` to the live `skyra` package; build the fake-memory harness; read the activation product off the tree.

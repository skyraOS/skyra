# Sensorimotor Boundary — Spec

**Date:** 2026-06-25 · **Status:** in spec — supersedes the **boundary model** in
[membrane.md](membrane.md) (membrane-does-the-gate-and-surroundings-write). The
**cognitive layer** (MidLevel / grouping / self·other) is unchanged; only **how the wave
reaches it** changes.

> Scope: the membrane boundary, the sensory/motor sub-actors, the per-being slice
> topology, and cell-initiated crossing. Captures the 2026-06-25 design conversation.

## 1. The shape — membrane as a thin boundary; sensory/motor as peer sub-actors

The membrane is **not** where perception work happens. It is a **boundary / dispatcher**
— a trampoline:

- `Observe` = **nothing.**
- `Children()` = **empty.** (The kernel goes straight from Observe to Express; the
  membrane is a leaf.)
- `Express` = **fire a wave inward to sensory** (the cell-initiated spawn, §3).
- `Content` = the **root drive** (durable).

Sensory and motor are the membrane's two faces, but they are **separate sub-actors**, not
inline helpers, and they fire on **opposite waves**:

| Face | Wave | Role |
|---|---|---|
| **Sensory** | the **in**-wave (membrane spawns it) | perceive: resolve the sender, write the surroundings, forward to cognition |
| **Motor** | the **out**-wave (cognition spawns it) | realize the chosen action, emit it to the world |

They never share a traversal — sensory on the inbound wave, motor on the outbound — so
"two faces of one boundary" creates **no contention** by construction.

> This **moves the surroundings write off the membrane** (old `membrane.md` had `Observe`
> do it). The membrane no longer resolves entities or writes the surroundings — **sensory
> does** (it already does, in `sensory_cell.go`). The membrane just routes.

## 2. Slice topology — three slices per being

Each being is a small constellation of actors, each with its **own slice**: **membrane**
(public boundary), **sensory** (private), **motor** (private) — plus the cognitive core
(MidLevel / self·other; see §8 open). The interface already supports this for free —
`Space.Slice()` is per-cell, so each organelle returning its own slice is all the
(cell-initiated) crossing needs.

### What lives where — the durable graph is shared by reference

The graph is **not owned by any one organelle.** Separate three things and ownership falls
out:

| Thing | What it is | Home |
|---|---|---|
| **Node identity** | the entity/setpoint/understanding `MemoryCell` (a heap object) | **shared by reference** — one canonical node, reachable from many indexes |
| **Index** | a cell's `{key → *node}` lookup (sensory's surroundings, motor's actions) | **private** to that cell's slice — holds *references* into the shared graph |
| **Edge weights** | the activation records (EMAs) | **on the slice of whoever traverses that edge** — already disjoint |

- **No canonical node registry.** Identity stays singular because of **single-authorship**
  (entities minted only in sensory at perception; setpoints at closure; understandings at
  crystallization) — not a registry enforcing uniqueness. Nodes are reached by **reference
  (snapshot handoff) + traversal (recall wave)**, never by id-lookup. The one id-keyed
  lookup — entity **resolution** at perception — is **local to sensory** (its `ByEntity`,
  stable `SpaceID` today; an associative resolution step later if entities start arriving
  without ids).
- **Weights partition by edge-owner.** `sensory→entity` (presence) → sensory's slice;
  `entity→setpoint`, `setpoint→understanding` (recall) → the cognitive slice. Same node,
  different edges, different slices, **zero contention** — this is the edge-owned
  activation model extended across slices (the kernel keys records off
  `traversed.Slice()`).
- **Structural writes to a shared node** (new entity, interaction record, setpoint stamp)
  are the only hazard (mutating a shared `children` map). Sensory (in-wave) and cognition
  are **sequential**, so sensory appends during its wave and cognition reads it after — no
  join needed between them. The formal **snapshot-in / join-out** is reserved for the
  genuinely concurrent pair (**self/other at closure**). See
  [working-memory.md](../../../design-docs/cognition/working-memory.md#concurrency).

## 3. Crossing is cell-initiated — the kernel no longer owns it

The kernel is **pure in-plane transport** (`relation.go`: "no crossing"). The old "kernel
detects `Content`-is-a-Reality → spawn" model is **gone and stays gone** — crossings are
now **explicit**: a dispatcher cell *knows* who it wants to message.

A dispatcher cell holds a `Time` (kernel) reference and fires:

```go
go kernel.Realize(origin=<perceiver>, traversed=<target>, impulse)
```

- **`go` (async).** Fire-and-forget. A synchronous call would block the launcher inside the
  target's traversal (and rip out its response) — wrong.
- **Deadlock-free *because* of §2's own-slices.** The launcher holds its slice lock through
  Express; the spawned wave locks the **target's** slice — a different lock. (A shared slice
  would re-enter and deadlock. Own-slices and cell-spawn are the same decision wearing two
  hats.)
- **Only dispatchers hold the ref** — membrane (→ sensory/motor), sensory (→ cognition),
  MidLevel (→ self/other). Passive `MemoryCell`s never spawn.
- **`Realize` is the runtime's "send" primitive**; cells are actors that call it. The kernel
  never knows a crossing happened — it just gets handed a new root.

This **supersedes** the earlier "`Express` returns a `Spawn{}` directive the kernel
executes" idea — cell-initiated is cleaner (no return overloading, full control over
multiple/conditional spawns) and matches `space.go`'s already-stated "a being acts outward
by calling its environment's entry with itself as origin." The trade: spawn-timing is a
**convention** (spawn on Express, not mid-Observe), not kernel-enforced — the same trust the
actor model runs on.

## 4. origin = the perceiver, not the launcher

When the membrane spawns sensory, it sets **`origin = sensory`**, not `origin = membrane`:

```go
go kernel.Realize(origin=sensory, traversed=sensory, impulse)   // "sensory's own intake"
```

- `origin` = the wave's **owner** (who it returns to). The membrane only **launches**; the
  perception is **sensory's**. So sensory is the rightful owner — `origin == traversed` is
  the "a being's own Observe is pure intake" pattern from `space.go`.
- **The membrane drops out of the loop** — the return path re-runs *sensory* (cheap, local),
  never the membrane. That is the point: the kernel does not run the membrane again.
- **Nothing is lost.** The external sender was never on `origin` — it rides on
  **`impulse.Source`** (`space.go` separates *sender* from *owner*). So origin=perceiver
  keeps the sender intact for the eventual motor reply.
- **Return-guard** (don't re-spawn on a `Return` impulse) lives on the **perceiver's**
  Express — where the forward/processing logic already is.

## 5. The kernel returns void — Perception Invariant temporarily suspended

**For the build, the kernel does not fire the return perception wave** (`returnToOrigin`
suspended). This **knowingly suspends** the
[Perception Invariant](../../../design-docs/laws/invariants/perception-invariant.md) — a
deliberate build-time scaffold so the **forward pipeline** (membrane → sensory → cognition
→ motor) can be gotten right first. The invariant remains foundational and is the
destination; the return perception wave is restored — and the suspension lifted — once the
forward path works and the reafference loop is wired. The invariant doc carries the
matching suspension banner.

## 6. The flow

```
external send (sender on impulse.Source)
  → membrane: Observe(nothing), no children → Express:
        go Realize(origin=sensory, traversed=sensory, impulse)   // EVERY cycle (unconditional)
  → sensory: resolve sender, write surroundings (own slice), append to shared graph
        → Express: forward → go Realize(origin=cognition, traversed=cognition, frame)
  → cognition (MidLevel / grouping / self·other): the cognitive cycle (unchanged) → choose action
        → Express: forward → go Realize(origin=motor, traversed=motor, intent)
  → motor: realize intent → emit action to the world (sender carried for the reply)
```

**Unconditional fire — but traffic-relative (load-bearing).** The membrane fires sensory on
**every wave that hits it**, including a wave that resolves to **no known entity** — so the
per-wave decay sweep always runs: the addressed entity **bumps**, everyone else **decays a
step** (someone addresses you → you focus on them → everyone else fades a little). Decay is
therefore **relative to the stream of waves, not wall-clock time**: proper time *is* the
per-slice traversal count (`slice.Count()`), so presence is measured in **interactions, not
seconds.** An idle being (no waves arriving) holds its surroundings **frozen** until the next
wave — there is **no heartbeat** decaying them during silence (a clock-driven heartbeat is a
*later* option if wall-clock fade is ever wanted, gated on the not-yet-built scheduler;
`relation.go`). The thing to avoid: gating the spawn on "perceived a *known* entity," which
would skip the decay sweep on ambient / unresolved waves.

## 7. Build punch-list

- [ ] **Kernel returns void** — suspend `returnToOrigin`; the only result path for now is
  forward spawns. *[scaffold]*
- [ ] **Membrane goes thin** — `Observe` no-op, `Children()` empty, `Express` fires
  `go Realize(origin=sensory, sensory, impulse)` unconditionally; drop the
  sensory-delegation. *[DNA]*
- [ ] **Sensory/motor own slices** — each mints its own at construction instead of receiving
  the membrane's. *[DNA / construction]*
- [ ] **Dispatcher cells hold the kernel (`Time`) ref** — passed at construction; only
  membrane / sensory / MidLevel use it. *[DNA / construction]*
- [ ] **Return-guard** on sensory (and each forwarder) `Express` — don't re-spawn on a
  `Return` impulse (moot while the kernel returns void; needed when the return wave is
  restored). *[DNA]*
- [ ] Sensory already does the surroundings write — keep; ensure decay fires every cycle.
  *[verify]*

## 8. Open

- **Where the durable cognitive graph (the "core" slice) sits** — the membrane's public
  slice, or a distinct core/MidLevel slice. Leaning core/MidLevel (the membrane is a boundary;
  the memory is the cognitive core), so "public slice" = the core memory slice the membrane
  routes to.
- **Origin-threading for the eventual reply** — how the external sender (on
  `impulse.Source`) is carried forward so the motor out-wave routes the considered response
  back.
- **Restoring the return wave** — when, and exactly where the reafference re-enters once the
  forward pipeline works (ties to un-suspending the invariant, §5).
- **Stress as a sharded counter** (model-b) — parked; see
  [activation-and-traversal.md](../../../design-docs/kernel/activation-and-traversal.md)
  ("Stress is global, not per-edge").

## 9. Done when

- Membrane is a thin trampoline (Observe / Children empty; Express fires sensory
  unconditionally).
- Sensory and motor are sub-actors on their own slices, reached by `go kernel.Realize`.
- The durable graph is shared by reference; each cell's slice holds only an index + its own
  edge weights; no registry.
- A real external impulse drives membrane → sensory → cognition end-to-end via forward
  spawns (kernel return void).
- *(later)* the return perception wave is restored and the Perception Invariant
  un-suspended.

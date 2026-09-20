# The invocation cactus and per-possibility isolation

Date: 2026-09-20
Status: **Design idea, captured from discussion.** Two decisions are open and
marked as such below. Rejectable — fold in, reshape, or discard per the
propose→confirm lane.

OS means the underlying operating system. Runtime means Skyra's runtime.

## Where this came from

Extending the `claude-syscall/` spawn work with two framings that surfaced in
discussion: (a) that `spawn` is a mechanism rather than an address, and (b) that
the map from an invocation back out to its result — and to the process that
serves it — is stack-shaped. Following those two through against the DNA-17 code
and its open issues produced one model with four faces. The two hard questions
it raises are not yet answered; they are recorded here rather than guessed.

## Summary — the four faces

1. **`spawn` is not addressable** — a mechanism the runtime invokes, not a
   destination with a `SurfaceID`. It *mints* an addressable capability; it is
   not one.
2. **`spawn` is a first-class system operation** — same class family as the OS
   `read`/`write`/`close`/`wait`/`signal`, all governed by
   `interpret → commit → execute → egress`. Its effect lands in **execute**
   (outside the locks); it is *not* an interpret-phase canonical read.
3. **The invocation structure is a cactus** rooted at `rootOnce` / the
   primordial space. `invoke` pushes a frame (`ReturnTo` = caller), `return`
   pops; concurrent live fds make each run a tree, each root-to-leaf path a
   stack. Process *resolution* is a walk toward the root.
4. **One cactus per Possibility** — a sealed forest. No shared frames, no
   cross-edges: trees cannot talk. `rootOnce` is both a cactus's floor and (if
   used) the one sanctioned edge back out.

## 1. `spawn` is a mechanism, not an address

In this system "being a Surface" *means* "having a `SurfaceID`" — the durable
registered **address** a Move routes to (`surfaceID-moveID-fd.md`). Nothing ever
routes a Move to "create a process," so `spawn` has no `SurfaceID` and is not a
Surface. It is a syscall the runtime invokes as a mechanism
(`claude-syscall/the-handoff-set.md`).

The asymmetry to keep straight: **spawning is not addressable; the thing it
produces is.** A successful `spawn` *mints* the running child's addressable
capability — the registered stdio Surface (`stdio-surface.md`) that `invoke` /
`return` actually target, i.e. the MoveID↔OS-process binding promoted to
canonical state (`claude-syscall/two-axes-one-seam.md`). "spawn is not
addressable" must **not** collapse into "the child is not addressable."

Consequence: the lifecycle wrappers `create-child-process`,
`ensure-child-process`, `attach-child-process` (`syscall/start-process/`) stay
**deferred** — no `SurfaceID`, not built — until some workflow must *address*
creation as a routed Move (a second, dynamically requested child; reconnection
policy). Until then the runtime just calls `spawn` when it needs the capability
to exist. `ensure` in particular reframes as *an invariant the runtime maintains
for an addressed capability*, not a Surface you call.

## 2. `spawn` is a first-class system operation, not a special case

Every operation obeys one phase discipline — `interpret → commit → execute →`
`(egress: interpret → commit)` (the genuine shape from
`main-loop-as-surfaces.md`). `spawn` is a first-class member of that family, a
peer of the OS `read`/`write`/`close`/`wait`/`signal` — not a bolted-on
process-creation category.

Class membership does **not** mean same phase. Within the family, an operation is
classified by **which phase carries its effect**:

| Family | Effect lands in | Under the state lock? | Examples |
|---|---|---|---|
| Canonical-state ops | `interpret` (reads), `commit` (writes) | **yes** | `resolve-runtime-process`, committed writes |
| OS system operations (syscalls) | **`execute`** | **no** | `read`, `write`, `close`, `wait`, `signal`, **`spawn`** |

So `spawn` is the peer of the **OS** `read`/`write` (effect in execute, outside
locks, outcome interpreted in egress) — *not* the interpret-phase canonical
reads. Putting an OS crossing into interpret would drop a `fork`/`execve` inside
the state lock and violate "no canonical reads past interpret" and "execution
outside the locks" (the fence, `claude-syscall/handle-liveness-fence.md`).

## 3. The invocation structure is a cactus rooted at `rootOnce`

The `ReturnTo` chain is a call stack. An invocation record carries `ReturnTo`;
a child Invoke's `ReturnTo` should point at its caller's fd; issue #1's bug is
exactly a frame whose `ReturnTo` "points to itself" — a corrupted frame. Healthy
shape: **`invoke` pushes a frame (`ReturnTo` = caller); `return` pops to
`ReturnTo`; the bottom frame is `rootOnce` / the primordial space.** That names
the "root return destination" hole `main-loop-as-surfaces.md` parked.

With concurrent live fds inside one run, the shape is a **cactus** (a tree where
every root-to-leaf path is a stack), not a single linear stack. **Decided:**
cactus, not linear stack — intra-run concurrency is allowed, and fan-out does
**not** force a new Possibility.

**Process resolution is a walk *up* the same structure.**
`resolve-runtime-process(parentMoveID)` is "the nearest enclosing frame that
carries a process binding," with the root frame's binding = the runtime itself.
So:

- **result → caller** = the cactus unwinding (pop along `ReturnTo`)
- **invocation → process** = walking toward the root until a frame owns a process

are the same structure read in two directions — `two-axes-one-seam` made
literal: the cactus *is* the seam. Frame depth/generation supplies the
settlement fence; `rootOnce` is the bottom sentinel (issue #11's
belt-and-suspenders).

Kept separate: a bound process's **lifetime**. A spawned process outlives the
frame that spawned it and its async emissions branch rather than unwind
(issue #9), so lifetime is **ownership/liveness** (the handle fence — refcount +
generation), never push/pop. The frame holds a *reference*; the resource lives
on its own fence.

## 4. One cactus per Possibility — a sealed forest

A Possibility is already the authority-and-identity unit in the code:

- `Submit` "opens a Possibility with the supplied Horizon and posts its first
  Move" (`cu.go`).
- Every `Move` carries a `PossibilityID` (`contracts.go`), and
  `authorize(view, poss, target, focus)` decides authority **keyed on the
  possibility** — "unknown possibility" is a rejection (`interpret.go`). A
  Possibility is "pure identity-and-authority DATA" (`contracts.go`).

So scope one cactus to one Possibility: its floor (`rootOnce`) is the
Possibility's First move, and the cactus never leaves the Possibility. The global
picture is a **forest — one cactus per Possibility, zero shared frames, no
cross-edges.** That is "trees cannot talk to each other."

```text
        (runtime root)
        /      |       \
   Poss A    Poss B    Poss C          <- each rootOnce = a First move
     |         |          |
   cactus    cactus     cactus         <- sealed; no edges between them
```

### What this makes unrepresentable (not merely checked)

- **Issue #1** (IDs overwrite records *across possibilities*): scope
  `InvocationID` **under** `PossibilityID`. `inv-1` in cactus A and `inv-1` in
  cactus B are different frames in different trees; an old fd cannot address a
  replacement in another tree because it cannot *name* another tree.
- **Issue #2** (a Fault targets another *possibility*): a `Fault` / `ReturnTo` /
  fd resolves only within its own cactus. A foreign-possibility target is not
  "rejected" — it is unrepresentable.

Two of the three "fix first" highs thus become structural, not runtime checks.

## Open decisions — not yet answered

### D-A. The return-out-of-a-Possibility edge — **OPEN**

If trees cannot talk, how does a Possibility's result reach whoever opened it?
Sketch (unconfirmed): exactly **one** sanctioned edge — the Possibility's
`rootOnce` / primordial space, wired **at `OpenPossibility` time** to a single
**write-once** sink in the opener, runtime-mediated; everything else stays
sealed. That would be the concrete fill for the "root still needs an agreed
result destination" hole.

Sub-question that decides the whole shape: **can a Move open a child Possibility,
or is a Possibility only ever opened externally by `Submit`?** If only `Submit`,
there are zero internal cross-edges — a fully sealed forest — and the result edge
is external. If a Move can open one, the single `rootOnce` join edge above is
required. Not resolved.

### D-B. Shared process vs. isolation — **OPEN**

A long-lived process referenced from two Possibilities is a *covert cross-tree
edge* and re-breaks the isolation face 4 buys. Options:

- **(a) Per-possibility spawn** — each Possibility spawns its own child. Clean
  isolation; costs a process per Possibility.
- **(b) Shared process below the forest** — one child at the runtime root, with
  async emissions **demuxed to the owning Possibility** by the runtime (issues
  #8/#9, "route async emissions to the right session"). A mediated seam, never
  free talk.

Leaning (a) as the isolation default, but not resolved — the Codex case may want
(b) for cost reasons, which then makes emission-routing load-bearing.

## To fold into the other specs (when confirmed)

- Keep the lifecycle Surfaces deferred (face 1) in `syscall/start-process/`.
- State `spawn`'s class + phase (face 2) in `skyra-syscall-surface.md`.
- Scope `InvocationID` under `PossibilityID` (face 4) — the concrete fix for
  issue #1 — and note the isolation is structural.
- Name `rootOnce` as both the cactus floor and the sole candidate join edge
  (face 3, D-A).

## Related

- [spawn is not fork](claude-syscall/spawn-not-fork.md),
  [the handoff set](claude-syscall/the-handoff-set.md),
  [syscall boundary](claude-syscall/syscall-boundary.md),
  [handle-liveness fence](claude-syscall/handle-liveness-fence.md),
  [two axes, one seam](claude-syscall/two-axes-one-seam.md)
- [SurfaceID / MoveID / fd](surfaceID-moveID-fd.md),
  [stdio surface](stdio/stdio-surface.md),
  [main loop as surfaces](main-loop-as-surfaces.md),
  [create-child-process](syscall/start-process/proposal.md)
- [DNA-17 review issues](issues.md) (#1, #2, #8, #9, #11)

# The Perception Invariant

## Status

**Foundational.** This is the first concrete invariant on world physics, and it must hold unchanged through the entirety of v.1. Every mechanism — internal traversal, external action, dispatch, result delivery — is built to preserve it. If a design choice would break this invariant, the design choice is wrong, not the invariant.

## The Invariant

> **What a being sends is chosen; what it perceives is not.**

In full:

> Every wave deposits its result in a field its origin reads. A being perceives the result of every wave that runs through it — internal or external, whether it started the wave or not. Perception of results is unconditional; the only choice is whether to send a wave outward.

## What It Means

- **Perception is involuntary.** A being does not choose what it perceives. Whatever a wave brings back, and whatever wave passes through the being, the being perceives its result. There is no filter that lets a being decline to see the outcome of a wave that touched it.
- **Sending is the only choice.** The single volitional act is initiating a wave outward — `world.observe(self, content)`. A being decides whether and what to send. It does not decide what returns.
- **Self-started or not, it perceives.** A wave the being initiated and a wave another being or the world sent through it are perceived the same way. The being always sees the output of any wave that ran through it.
- **Internal or external, it perceives.** A wave through the being's own interior (thought) and a wave out into the world (action) both deliver their result back to the being. The being sees the output of its internal pass with no more choice than it sees the world's response to its act.

## The Mechanism

One rule enforces the invariant, and it is identical for internal and external waves:

> Every wave writes its result to a field its origin reads, and on completion delivers that result to the origin.

A wave is a goroutine — its own Relation — carrying its origin, its primitive, and a result field it populates during the traversal. Every call to `kernel.realize` is a new wave. When that traversal completes (the tree is built, the ascent is done), the kernel delivers that wave's result to that wave's origin: `origin.observe(result)`. The origin reads what its own wave produced.

Internal and external waves are not two mechanisms. They end the same way — write a result, deliver it to the origin. They differ only in **what the result contains**:

- **Internal wave** — the result is the being's own output: what that traversal concluded, expressed, or changed inside it.
- **External wave** — the result is the world's response: what the primitive did when it landed.

So "the kernel calls `origin.observe` at the end" is correct and the same in both cases. There is one physics — a wave runs, writes a result, the origin perceives it — and "internal vs external" is only where the wave was started and what its own traversal brought home.

## Holds Recursively

The invariant never names the level it applies at. "Every wave delivers its result to its origin" speaks only of *wave*, *origin*, and *result* — things that exist at every level of the nesting. An invariant stated only in terms that exist at every level is automatically recursive; it has no level to be specific to. That is the test of a real invariant rather than a top-level rule dressed up as one, and this one passes.

It falls out of two uniformities already in the runtime: every participant implements the same `Reality` interface, and the crossing rule is uniform (Content that is a Reality → the kernel spawns a goroutine). Crossing a boundary does not extend the current wave into another actor as a blocking subroutine; it starts a new wave. The outer wave continues toward its own result, and the spawned wave delivers to its own origin. The inner universe — specialists as beings inside an agent — inherits the invariant for free: an inner being perceives the result of its waves exactly the way a top-level being does. The rule is written once.

The signature carries this. The kernel is called as:

```
realize(OriginReality, TraversedReality, impulse)
```

- **OriginReality** — the wave's owner: the being whose wave it is, the `Observe` discriminator, and whoever the result is delivered back to.
- **TraversedReality** — the realm traversed by this wave. If traversal discovers another Reality in `Content`, that discovery starts another wave rather than making this one wait on the inner actor.
- **impulse** — the act.

No world-specific parameter (no grid location) appears — location is the environment's knowledge about its own occupants, at whatever level, never threaded through the kernel. Because the signature names only what exists everywhere, the same call works at every depth. And because **`OriginReality` is a required argument, the invariant is mechanical**: you cannot start a kernel wave without naming who perceives the result, and the kernel will deliver it there. The owner is not a rule to remember — it is an argument you cannot omit.

A being acts by calling into its environment — `world.observe(self, content)` at the top level, `innerWorld.observe(self, content)` for a being inside an agent. This looks inverted: the environment *contains* the being, yet the being *calls into* it to act. There is no contradiction — **containment and call-direction are different axes.** You are *in* an environment, and you act by *invoking* it. The container is also the executor; the recursion is what lets one Reality be both.

## Enforcement Boundary

The invariant holds **by construction for any wave that goes through the kernel.** `OriginReality` is a required argument and the kernel delivers the result to it; the standard traversal path cannot run a wave that fails to deliver.

It does **not** hold for code that bypasses the kernel. If DNA calls another cell's `Observe`/`Express` directly, or realizes itself without going through the kernel, it has stepped outside the mechanism that threads `OriginReality` and delivers the result — and the developer owns the invariant there by hand. This is kernel-is-mechanism / DNA-is-policy applied honestly: the kernel is the sanctioned path that enforces the invariant; policy that calls its own methods is outside what the mechanism can vouch for.

State this plainly so "invariant" is not misread as "unbreakable": the Perception Invariant is **guaranteed on the kernel path, honored-by-design off it** — kernel-enforced, developer-honored. That is the right trade for this runtime. The kernel makes the standard way correct for free; the escape hatch exists, and using it makes the invariant your responsibility.

## What It Forbids

The invariant rules out a class of designs, permanently for v.1:

- **No silent waves.** A wave cannot run through a being and leave no perceivable result. Every traversal delivers.
- **No opt-out perception.** A being cannot choose not to perceive the result of a wave that touched it. Perception is not gated by the being's will.
- **Dispatch is not a return.** A cell's `Express` returns its *output*, never a signal that starts a wave. Starting a wave is an explicit `world.observe(self, content)` call, made by the motor when it decides to act. Burning wave-initiation into the `Express` return would steal the developer's control over *when* the kernel fires, and would make every expression an involuntary send — violating "sending is the only choice."

## Why It Must Hold

The invariant is the reafference loop, and the being's learning depends on it. A being acts, and the result of the act returns unconditionally — so the being can compare what it predicted against what came back. That delta is the signal that drives prediction error, episodic segmentation, emotion, and weight updates. If a being could decline to perceive the result of its own wave, it could act without consequence and never learn. The internal echo — the being hearing itself think — is the same mechanism: a being always perceives the output of its own thought, which is what lets it monitor and correct itself.

Sending is chosen so the being has agency. Perception is unconditional so the being is always accountable to what its waves did. **Agency out, accountability back.** That pairing is the invariant.

## Relationships

- [world-kernel-coupling.md](../kernel/world-kernel-coupling.md) — dispatch and the world boundary; how a being's outward send reaches the world.
- [descent-ascent.md](../kernel/descent-ascent.md) — the wave's two phases; the result is formed on the ascent.
- [world-physics.md](../world/world-physics.md) — the world's always-present physics; this invariant governs how their results reach beings.

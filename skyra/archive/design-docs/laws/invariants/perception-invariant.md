# The Perception Invariant

## Status

**Foundational.** This is the first concrete invariant on world physics, and it must hold unchanged through the entirety of v.1. Every mechanism — internal traversal, external action, dispatch, and return perception — is built to preserve it. If a design choice would break this invariant, the design choice is wrong, not the invariant.

### ⚠️ Temporarily suspended during implementation (2026-06-25)

The invariant is **knowingly suspended for the current build.** While the sensorimotor
**forward pipeline** (membrane → sensory → cognition → motor) is being stood up,
**`kernel.Realize` returns void and does not fire the return perception wave** — there is no
reafference loop yet. This is a deliberate **build-time scaffold to get the forward path
right first**, not a revision of the design: the invariant remains foundational and is the
destination. The return perception wave (and this suspension's removal) is restored once the
forward pipeline works and the reafference loop is wired back in. Tracking spec:
[../../../specs/2026-06-16/skyra-dna/sensorimotor.md](../../../specs/2026-06-16/skyra-dna/sensorimotor.md)
(§5). **Until then, read "every wave returns to its origin" as the target state, not current
runtime behavior.**

## The Invariant

> **What a being sends is chosen; what it perceives is not.**

In full:

> Every kernel wave returns its result to its origin by starting a perception wave rooted at that origin. A being cannot choose whether to perceive the result of a wave it owns; the only choice is whether to send a wave outward.

## What It Means

- **Perception is involuntary.** A being does not choose what it perceives. Whatever one of its waves brings back becomes an inward perception wave through the being's own topology. There is no filter that lets a being decline to process the outcome of its own wave.
- **Sending is the only choice.** The single volitional act is initiating a wave outward — `world.observe(self, content)`. A being decides whether and what to send. It does not decide what returns.
- **Ownership defines return.** A wave returns to its origin. Crossing into another Space starts another wave whose origin is that inner Space, so the outer actor does not receive the inner actor's result.
- **Internal or external, it perceives.** A wave through the being's own interior (thought) and a wave out into the world (action) both return their result through that wave's origin. The being processes the output of its internal pass with no more choice than it processes the world's response to its act.

## The Mechanism

One rule enforces the invariant, and it is identical for internal and external waves:

> Every wave forms one result during traversal, and on completion starts a return perception wave rooted at the origin.

A wave is a call to `kernel.realize`: it carries its origin, traversed Space, impulse, and traversal result. `Relation` is the shared stateless kernel object, not the wave object. When traversal completes (the tree is built, the ascent is done), the result does not merely callback into the actor. It becomes a new inward wave:

```go
Realize(origin, origin, resultImpulse)
```

That return wave roots in the actor's own topology, so the actor truly perceives the consequence by traversing itself. `origin.Observe(origin, result)` is only a shallow callback; it is not sufficient for true perception.

The return perception wave is the **terminus**, not a new send: it does not itself start a further return wave — that would recurse without end. The being perceives the result and may then *choose* to act again, but that next send is a fresh volitional act, never an automatic return-of-return. This is **"sending is the only choice"** enforced at the close of the loop — action returns to be perceived involuntarily, and there it stops; the being is not mechanically driven to respond. (The runtime currently marks the return wave terminal so it cannot re-return; the durable rule is the terminus, not the marker that enforces it today.)

The wave origin is not the same thing as impulse source metadata. `Impulse.Source` may describe where an impulse appears to arise in a traversed plane — for example a world point and coordinate — but that source is context only. It does not own the result path, and it does not let the external wave continue into the perceiver. The kernel `origin` still names the owner of the wave and the root of the return perception wave.

During traversal the kernel walks the realm **anonymously**: every intake `Observe` carries a **nil** origin — a cell is never told whose wave it is. The owner is named only when a wave is invoked from outside a traversal. So the discriminator is **nil vs named**, never a comparison of identities: a nil origin means *I am being traversed*; a named origin means *I am being invoked*. Self-versus-external is not decided inside a being's interior (there is no foreign origin there to compare against) — it lives at the environment boundary (`world.Observe(self, …)` → a being invoking the world) and in the return wave entering the actor's own topology.

Internal and external waves are not two mechanisms. They end the same way — write a result, then return that result through the origin's topology. They differ only in **what the result contains**:

- **Internal wave** — the result is the being's own output: what that traversal concluded, expressed, or changed inside it.
- **External wave** — the result is the world's response: what the primitive did when it landed.

So "the result returns to the origin" means "the result starts an inward perception wave rooted at the origin." There is one physics — a wave runs, forms a result, and that result is processed by the origin's own topology — and "internal vs external" is only where the first wave was started and what its own traversal brought home.

## Holds Recursively

The invariant never names the level it applies at. "Every wave returns its result through its origin" speaks only of *wave*, *origin*, *result*, and *return perception wave* — things that exist at every level of the nesting. An invariant stated only in terms that exist at every level is automatically recursive; it has no level to be specific to. That is the test of a real invariant rather than a top-level rule dressed up as one, and this one passes.

It falls out of two uniformities already in the runtime: every participant implements the same `Space` interface, and the crossing rule is uniform (Content that is a Space → the kernel spawns a goroutine). Crossing a boundary does not extend the current wave into another actor as a blocking subroutine; it starts a new wave. The outer wave continues toward its own result, and the spawned wave returns through its own origin. The inner universe — specialists as beings inside an agent — inherits the invariant for free: an inner being perceives the result of its waves exactly the way a top-level being does. The rule is written once.

The signature carries this. The kernel is called as:

```
realize(OriginSpace, TraversedSpace, impulse)
```

- **OriginSpace** — the wave's owner: the being whose wave it is and the root of the return perception wave.
- **TraversedSpace** — the realm traversed by this wave. If traversal discovers another Space in `Content`, that discovery starts another wave rather than making this one wait on the inner actor.
- **impulse** — the act.

No world-specific parameter (no grid location) appears — location is the environment's knowledge about its own occupants, at whatever level, never threaded through the kernel. Because the signature names only what exists everywhere, the same call works at every depth. And because **`OriginSpace` is a required argument, the invariant is mechanical**: you cannot start a kernel wave without naming where the return perception wave roots. The owner is not a rule to remember — it is an argument you cannot omit.

A being acts by calling into its environment — `world.observe(self, content)` at the top level, `innerWorld.observe(self, content)` for a being inside an agent. This looks inverted: the environment *contains* the being, yet the being *calls into* it to act. There is no contradiction — **containment and call-direction are different axes.** You are *in* an environment, and you act by *invoking* it. The container is also the executor; the recursion is what lets one Space be both.

## Enforcement Boundary

The invariant holds **by construction for any wave that goes through the kernel.** `OriginSpace` is a required argument and the kernel has enough information to start the return perception wave there; the standard traversal path cannot run a wave that has no return root.

It does **not** hold for code that bypasses the kernel. If DNA calls another cell's `Observe`/`Express` directly, or realizes itself without going through the kernel, it has stepped outside the mechanism that threads `OriginSpace` and starts the return wave — and the developer owns the invariant there by hand. This is kernel-is-mechanism / DNA-is-policy applied honestly: the kernel is the sanctioned path that enforces the invariant; policy that calls its own methods is outside what the mechanism can vouch for.

State this plainly so "invariant" is not misread as "unbreakable": the Perception Invariant is **guaranteed on the kernel path, honored-by-design off it** — kernel-enforced, developer-honored. That is the right trade for this runtime. The kernel makes the standard way correct for free; the escape hatch exists, and using it makes the invariant your responsibility.

## What It Forbids

The invariant rules out a class of designs, permanently for v.1:

- **No silent waves.** A *sent* wave cannot complete without starting one return perception wave rooted at its origin — action always returns to be perceived. The return perception wave is the terminus of that loop and does not itself re-return.
- **No opt-out perception.** A being cannot choose not to process the result of a wave it owns. Perception is not gated by the being's will.
- **No return-valued dispatch.** `kernel.Realize` returns nothing. A return value would create a second, chosen result channel for the caller, bypassing involuntary return perception.
- **Dispatch is not a return.** A cell's `Express` returns its *output*, never a signal that starts a wave. Starting a wave is an explicit `world.observe(self, content)` call, made by the motor when it decides to act. Burning wave-initiation into the `Express` return would steal the developer's control over *when* the kernel fires, and would make every expression an involuntary send — violating "sending is the only choice."

## Why It Must Hold

The invariant is the reafference loop, and the being's learning depends on it. A being acts, and the result of the act returns unconditionally as a perception wave through its own topology — so the being can compare what it predicted against what came back. That delta is the signal that drives prediction error, episodic segmentation, emotion, and weight updates. If a being could decline to perceive the result of its own wave, it could act without consequence and never learn. The internal echo — the being hearing itself think — is the same mechanism: a being always processes the output of its own thought, which is what lets it monitor and correct itself.

Sending is chosen so the being has agency. Perception is unconditional so the being is always accountable to what its waves did. **Agency out, accountability back.** That pairing is the invariant.

## Relationships

- [actor-frame-invariant.md](actor-frame-invariant.md) — the companion: where each wave *begins* (thought roots in the actor, action in its location); this one is where it *returns*.
- [world-kernel-coupling.md](../kernel/world-kernel-coupling.md) — dispatch and the world boundary; how a being's outward send reaches the world.
- [descent-ascent.md](../kernel/descent-ascent.md) — the wave's two phases; the result is formed on the ascent.
- [world-physics.md](../world/world-physics.md) — the world's always-present physics; this invariant governs how their results reach beings.

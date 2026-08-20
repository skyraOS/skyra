# The Local Relativity Invariant

## Status

**Foundational.** This invariant defines the frame in which every runtime quantity is measured. It is the umbrella under which proper time, distance, direction, and active state are corollaries — none of them is a global fact; each is relative to the observer whose wave measured it. Must hold unchanged through v.1.

## The Invariant

> **Every quantity is relative to the observer. There is no global frame.**

In full:

> A being knows only what its own waves have measured, in its own frame. Time, distance, direction, and the state of everything else are not properties of the world — they are relations between the world and the observer, established by observation and valid only as of that observation. There is no privileged frame holding the current truth; there is no global now.

## The Corollaries

Four quantities, one principle. Each is relative to the observer, and each already lives in the runtime this way:

- **Time is proper time.** Recency is traversal count, not wall clock. Each Slice carries its own `count`, bumped once per wave (`sl.SetCount(sl.Count() + 1)`). A being ages by the relation-flow passing through it; two beings in the same runtime age at different rates. No universal clock.
- **Distance is activation.** There is no global map. Activation strength *is* distance — `1.0` is inside the observer, `0.0` is unreachable. Two beings on the same point experience different proximities to the rest of the world because their weights differ. Geometry emerges from traversal, not from coordinates (see [Space Is Local](../../world/slice-and-world.md#space-is-local)).
- **Direction is attention.** Heading is relative to the being's own aperture — there is no global north and no per-act bearing computation. "Toward michael" is just where the being is already oriented because it looked at michael and held focus.
- **Active state is the last observation.** The world Slice's `Records` keyset *is* the observed subset (its complement is potential, unobserved). The double-buffer `Prior` holds the last completed traversal — the being operates on its last snapshot while building the next. State is held, not live.

## The Staleness Rule

The temporal corollary has teeth, so state it on its own:

> An observation is stale until observed again.

- **No state is pushed.** The world never notifies an observer when something changes elsewhere. State is pulled by re-observing, never pushed.
- **The only refresh is re-observation.** Between observations, a being holds its last-measured model unchanged. It does not drift toward truth; it sits frozen until a new wave corrects it.
- **The correction arrives with the next signal.** When an entity moves, no observer learns of it until that entity's next signal reaches them — and that signal, originating from the new condition, carries the correction implicitly. Skyra at `1,1` engages michael at `1,5`; michael responds and moves to `1,6`. Skyra's model still reads `michael@1,5` until michael's next message arrives — and that message propagates *from* `1,6`, so hearing from him again *is* the update. The lightcone delivers the news.
- **Staleness is correctness, not desync.** Skyra's model is stale, the world's record is stale, and that is not a bug to repair. Forcing all observers to agree would require a privileged global frame — the absolute-space mistake. Coherence is reached the only legitimate way: by observing again.

## Reached vs Rooted

The staleness rule forces a state distinction worth naming. A point can be **reached** — observed as a child, a leaf in another wave's tree — without ever having **rooted** a wave of its own. When michael moves into `1,6`, occupancy is there and `1,6` was reached by the move wave as its destination, but `1,6` has never observed *from itself*: it has authored no tree, holds no view. **Reached is not the same as has-observed.** A point can be looked-at without having looked.

## World Effects vs. Lived State

The world's frame is authoritative for **world effects** — occupancy and execution. When michael moves to `1,6`, the world's placement record *is* the fact of the matter: that is where the move actually landed, and it is what a future observation will read. Effects are real in the world's frame, and that is what "stale" is measured against — Skyra is stale *relative to* an occupancy that genuinely changed.

But the world is **not** globally authoritative as another observer's **lived state**. It cannot reach into Skyra's frame and correct `michael@1,5`; her lived model is her own, valid in her frame until she observes again. And the world's *own* observed record — its activation `Records` — is itself just another frame's last look, no truer than a being's.

So the world wears two frames, and they must not be conflated:

- **Executor / registrar of effects** — authoritative. Occupancy and primitive execution are real here. This is the durable [occupancy](../../world/world-observe.md#durable-vs-ephemeral) store.
- **An observer with its own activation record** — not privileged. The world's `Records` are stale snapshots like anyone's.

This does not reintroduce a global now. Authoritative occupancy is the record of *executed effects*, not a synchronized god's-eye view of every observer's model — and it is still only knowable by observation, never pushed. There is a fact of where michael is; no observer has privileged instantaneous access to it. The fact grounds "stale"; observation remains the only way to learn it.

## Not the Collapse Principle

Local relativity governs *observer-relative quantities and retarded information*. It is **not** the rule that observation creates state from potential — that is the measurement / collapse principle (`Realize` as measurement; disk = potential, active = the collapsed subset; see the [Active State Invariant](active-state-invariant.md)). Two physics threads, kept separate:

- **Local relativity** — relativistic. No global frame; quantities are observer-relative; information is retarded. (this document)
- **Measurement / collapse** — quantum. Observation realizes potential into active state.

A single wave does both at once: it collapses potential into active state (measurement) *and* it does so in the observer's frame, stale everywhere else (relativity). But the two answer different questions, and neither subsumes the other. Don't let one name swallow the other.

## Why It Must Hold

Having no global state to consult is what makes the runtime honest and what makes it scale. There is nothing to synchronize, no authoritative store to bottleneck, and a being's model costs exactly what it has bothered to observe — no more.

It is also what makes prediction load-bearing rather than decorative. Between observations a being runs on its last frame (`Prior`); the next observation's delta against that frame is the surprise that drives prediction error, episodic segmentation, emotion, and weight updates. A relativistic, retarded-information world is a *predicting* world by necessity — the being must run on a model because the truth is never directly available, only its last measurement of it. This is the same loop the [Perception Invariant](perception-invariant.md) guarantees the return path for.

## What It Forbids

- **No global now.** No structure anywhere holds the current state of the world. Any "current" is some observer's last observation.
- **No pushed updates.** Nothing notifies an observer of a change it did not observe. State refreshes only by re-observation.
- **No synchronization layer.** There is no reconciliation pass that makes observers agree. Agreement, where it happens, is two observers having recently observed the same thing — coincidence of frames, not enforced consistency.
- **No privileged frame for lived state.** No observer's *observed* model is authoritative over another's — the world's activation record included; it is just another frame's last look. (The world's authority over *effects* — occupancy, execution — is a separate thing and does survive; see [World Effects vs. Lived State](#world-effects-vs-lived-state).)

## Naming

"Local relativity," deliberately — and deliberately **not** "local realism." Local realism (locality + observer-*independent* definite values) is the conjunction Bell's theorem refuted; this runtime is on the side that survived, where values are observer-*dependent* and established by measurement. What it keeps from physics is the **relativity**: observer-relative quantities and finite information propagation, both confirmed. "Local" because physics here is always defined in a local frame — each being, each point, its own — and never a global one. The same claim [Space Is Local](../../world/slice-and-world.md#space-is-local) already makes, raised to a law.

## Relationships

- [active-state-invariant.md](active-state-invariant.md) — the disk → active realization; the measurement/collapse thread this one is distinct from.
- [plane-slice-invariant.md](plane-slice-invariant.md) — which Slice a Space's frame belongs to; the per-plane keying active state rides on.
- [perception-invariant.md](perception-invariant.md) — the return path; why a predicting world needs results processed unconditionally.
- [world-kernel-coupling.md](../../kernel/world-kernel-coupling.md#direction-is-the-third-relative-quantity) — the three relative quantities stated alongside the wave-shape model.
- [slice-and-world.md](../../world/slice-and-world.md#space-is-local) — "Space Is Local," the canon line this law generalizes.

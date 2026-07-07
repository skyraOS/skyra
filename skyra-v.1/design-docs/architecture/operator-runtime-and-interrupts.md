# Operator Runtime and Interrupts

## Status

High-level direction note. Not an implementation plan.

## Thesis

Skyra is moving toward a runtime where cognition is not one monolithic full-pass loop on
every event. The system should support several active **operators** maintaining different
motor functions over the world at the same time, each working inside its own framing and
local slice of context.

The expensive global pass remains, but it becomes the exception path. Most of the time the
runtime should continue locally. A broader cognitive pass happens when continuity breaks.

## The shape

The runtime is split into two modes:

- **local operator mode** — a framed operator continues work inside an already-coherent
  slice of the world
- **interrupt mode** — a surprise, contradiction, or out-of-context impulse escalates the
  system into broader cognition, regrouping, or reframing

This means the system does not recompute the whole being on every small update. It spends
most cycles continuing local work and only pays for global cognition when the current frame
is no longer trustworthy.

## Operators

An operator is a bounded motor faculty, not the whole mind.

Each operator has:

- a **framing** — what this operator believes it is doing
- a **local world slice** — the part of the substrate it is allowed to work over
- a **motor function** — the kind of action or maintenance it performs
- a **continuity rule** — what counts as normal progress vs. a break in context

Examples later might include cleanup, navigation, observation, communication, repair, or
 device-specific control. The important point is that these are all operating over the same
 substrate, not separate products.

## Shared substrate

The load-bearing idea is unification:

- memory can be traversed
- devices can be traversed
- workspaces can be traversed
- places can be traversed
- episodes can be traversed

Operators differ in framing and affordance, but not in the underlying law. The same kernel
and world mechanics should hold across all of them.

## Why this helps

This direction solves two practical problems at once:

1. **Latency**
   The user should not pay for a full cognitive pass on every small change. Operators can
   continue locally until something meaningful breaks.

2. **Scalability of context**
   Each operator carries only the context for its own framed work, not the full world-state
   of the being on every cycle.

## Interrupts

An interrupt is an event that current local framing cannot comfortably absorb.

Interrupts are the trigger for escalation. They are the point where the runtime stops
continuing locally and performs a wider cognitive operation.

At a high level, interrupts may include:

- **novelty interrupt** — the impulse does not fit an active grouping or current frame
- **coherence interrupt** — what is being observed conflicts with the operator's framing
- **frame degradation interrupt** — incoming and outgoing impulses begin grading poorly
  against the current frame across rounds
- **closure interrupt** — action outcome sharply contradicts expectation
- **priority interrupt** — an external event preempts current local work
- **world-state interrupt** — another place or device changes in a way that invalidates the
  current slice
- **self-query interrupt** — a motor operator explicitly engages `self/skyra` to ask an
  internal question and wake the full organism

Different interrupts may require different handlers. Some can be repaired locally. Others
should force regrouping, reframing, or a larger global pass.

## Frame degradation interrupt

Grading does not only matter at episode formation time. If local operation continues across
rounds, the runtime should keep grading both incoming and outgoing impulses against the
current frame.

If those grades begin to fall, the system should treat that as loss of frame integrity.
This means the operator's local continuation is no longer trustworthy, even if no single
event was dramatic enough to count as a novelty shock.

High-level rule:

- good grades across rounds -> the operator may keep running locally
- degrading grades across rounds -> wake the organism and run a broader pass

This gives the runtime a way to detect slow drift, not only sharp contradiction.

## Self-query interrupt

One important interrupt class is an introspective handoff:

```xml
<engage><self/skyra>ask self question</self/skyra></engage>
```

This is not just another tool call. It is a motor operator deliberately pausing local
continuation and waking the full organism to request more context, broader cognition, or a
fresh pass through the active loop.

Use cases:

- the operator is uncertain how to interpret the current frame
- several local continuations are possible and need judgment
- the current work lacks enough context and needs a fuller cognitive pass
- the operator suspects it is drifting and wants the organism to re-orient at full depth

This is a true global interrupt. It is the operator's way of saying: I cannot responsibly
continue from my local frame alone; wake the organism and run the larger loop.

## Relation to active grouping

Active grouping gives the runtime continuity. Operators ride that continuity.

The local/default path should be:

- continue the active grouping
- keep operators working inside that grouping's frame
- only escalate when a message, perception, or result no longer coheres with that frame

So the system's normal state is not full recomputation. Its normal state is framed
continuation.

## Relation to workspace projection

The "big brain" or base world can be treated as the being's **workspace**. Remote worlds,
devices, and subworlds remain locally owned, but relevant pieces can be projected into the
workspace for coordination.

This allows several operators to work over a shared workspace while still respecting local
world ownership:

- a filesystem slice from one machine
- a photo library slice from a phone
- a message thread from another place

These can appear together in one workspace without losing the fact that they come from
different worlds.

## Design intent

The end-state is not "many chat agents." The end-state is a world runtime with:

- one shared substrate
- multiple framed operators
- local continuation as the default
- interrupts as the escalation path
- global cognition reserved for surprise, contradiction, and reorientation

If this works, the runtime should feel less like a tool router and more like a system that
can remain active, distributed, and coherent across time.

## Open questions

- how operators are created, retired, and merged
- what exact signals raise each interrupt class
- which interrupts can be handled locally vs. globally
- how operators share or compete for workspace projection
- how closure, coherence, and priority interact when they disagree

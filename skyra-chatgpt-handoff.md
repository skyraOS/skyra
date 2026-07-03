# Skyra Handoff

## What Skyra Is

Skyra is an experimental agent runtime built around the idea that cognition should be a traversable substrate, not just a prompt loop wrapped in tools.

The core architectural bet is:

- intelligence needs somewhere to live
- the runtime should provide topology, traversal, memory, action, and world contact
- the model should operate inside that structure, not be the structure

Skyra is not organized like most agent systems. It is not mainly:

- planner -> tool router -> retrieval -> action executor
- a centralized orchestrator around an LLM
- a bag of modules stitched together by workflow rules

Instead, it is organized around:

- a kernel that is intentionally ignorant of what it traverses
- a `Space` interface that everything implements
- waves / traversal dynamics
- `Observe` on descent and `Express` on ascent
- local activation, thresholds, amplitude, and child topology
- worlds, beings, cells, and places all participating through the same substrate law

The point is to make memory, thinking, action, and world interaction manifestations of the same underlying mechanism rather than separate subsystems bolted together.

## The Kernel Bet

One of the deepest bets in Skyra is that the kernel should stay ignorant.

The kernel should know:

- how to traverse
- how to carry a wave
- how to score and threshold
- how to walk children
- how to preserve origin and return structure

The kernel should not know:

- what a maze is
- what a world primitive means
- what a device is
- what inference is
- what a memory is in semantic terms

So far, this is holding. The kernel supports the runtime while staying largely domain-agnostic. That separation took serious work before inference could even be wired in.

## Why This Feels Different From Most Agent Architectures

Skyra is different from most current agent architectures in a meaningful way.

Most current systems are built around workflows and orchestration. Skyra is built more like a physics layer for cognition.

Differences:

- the kernel is not a planner, it is a transport / traversal law
- cognition is spatialized and topological, not just procedural
- world interaction is not just "tool use"; it is boundary crossing into affordance-bearing spaces
- the LLM is an operator inside the system, not the seat of the runtime itself

The result should feel different because it is different:

- there is a somewhere
- there are boundaries
- there is locality
- there are reachable and unreachable things
- actions happen from positions
- movement, perception, and contact matter

## Why This Direction Was Chosen

The project arrived here after multiple prior runtimes.

The big bet is contrarian:

- as models get better, they will outgrow thin harnesses
- better models will make weak agent scaffolding more obviously inadequate
- stronger models will need better substrate, not less substrate

That is why Skyra tries to merge:

- memory
- action
- thinking
- traversal
- world contact

into one deeper mechanism.

This is costly work, which is probably one reason few people are pursuing it casually.

## Current Milestone

A major recent milestone: the new runtime completed a maze by itself.

That matters because it suggests:

- the wave dynamics are working
- the kernel / world boundary is working
- inference can now inhabit the architecture instead of faking behavior on top of it

This was the first version of the new architecture that did something concrete and complete in the world on its own. That makes it an important proof point.

## Long-Term Direction: GPS and Device Places

One of the long-term goals is to map GPS onto the world and let devices become "places."

The concept:

- the physical world is one world
- devices are places layered onto that world
- a phone, laptop, or other device becomes a traversable environment
- an agent can enter that place and interact with its local affordances

This means a device is not just a tool registry.

It is a place the agent can inhabit.

Examples:

- entering a phone place
- perceiving what is inside
- engaging local affordances
- moving through device-specific surfaces
- building continuity and memory inside that place

Under the hood, this can be implemented through a local API layer. For example:

- a Mac mini brokers calls
- a phone companion app exposes approved capabilities
- the runtime presents that as a coherent place-world

Spatially, the agent experiences itself as being "in the phone," even though the actual transport layer is an API bridge.

That is not fake in a bad way. It is a boundary abstraction that preserves the ontology of the system.

## Device APIs and Practical Constraint

The plan is to surface only what is actually available through public APIs or supported system surfaces.

For iPhone specifically:

- some domains are writable through public frameworks
- some are only partially controllable
- some are not broadly exposed

So the architecture should not assume total device control.

Instead:

- each place exposes only real affordances
- unsupported capabilities simply do not exist in that place
- the world stays honest

## Product Direction We Discussed

A strong near-term product direction is:

"an AI that can inhabit your devices and organize them"

This could include:

- photos
- duplicate photos
- screenshots
- large videos
- duplicate contacts
- stale reminders
- repeated files
- other cleanup and organization tasks where public APIs permit it

This is interesting because it is not just one-shot tool use.

It is a place-based job:

- inspect
- perceive patterns
- surface duplicates
- cluster likely junk
- ask for confirmation where needed
- act where permitted
- remember what has already been reviewed

That could potentially compete with or replace many paid iPhone cleanup apps, while feeling much more intelligent and situated.

## Why the Device-Place Idea Matters

The device-place idea could be a real differentiator.

It is contrarian, but in a good way:

- it disagrees with the current agent default for structural reasons
- it is testable in concrete behavior
- it could produce a noticeably different user experience

If most agents feel like:

"request goes into a capability cloud and response comes back"

Skyra could feel more like:

- the agent is somewhere
- it entered your phone
- it perceives what is around it
- it can only do what that place affords
- it can organize that place as an inhabitant, not just as an external script

That difference in feel could become a real product advantage because it comes from the architecture itself, not just presentation.

## Short Version

Skyra is an ambitious attempt to build a mind-shaped runtime where:

- topology matters
- traversal matters
- locality matters
- the world matters
- devices become places
- the kernel stays ignorant
- cognition is not just a workflow around an LLM

The latest proof point is that the new runtime successfully completed a maze by itself.

The next major horizon is turning devices into traversable places so an agent can inhabit your phone, perceive its contents through allowed APIs, and help organize and clean it up.

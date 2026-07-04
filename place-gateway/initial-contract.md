# Place Gateway Initial Contract

## Core Contract

The first gateway contract should stay very small.

Arguments:

- `place`
- `primitive`
- `impulse`

That means the gateway receives:

- what place is being addressed
- what affordance is being invoked
- the payload / envelope needed to perform it

## Why This Shape

This keeps the gateway aligned with Skyra's ontology.

The runtime does not call random device APIs directly.

Instead it says:

- route this primitive
- at this place
- with this impulse

The gateway then decides how that maps to the backing API surface.

## First Primitive Set

The initial primitive surface is:

- `perceive`
- `engage`
- `move`
- `wait`

For the current device-place discussion, the important first three are:

- `perceive`
- `engage`
- `move`

## First Routing Rule

Routing should be based on:

- place kind / identity
- primitive

The gateway should then dispatch to the correct adapter or helper.

Conceptually:

- `macbook + perceive` -> macbook perception helper
- `macbook + engage` -> membrane crossing helper, or later space-specific engage helper
- `macbook + move` -> world/place-local movement helper

## Early Principle

The gateway should not be modeled as a generic RPC bag.

It should be modeled as:

- place-aware
- primitive-aware
- impulse-driven

## First Interface Shape

Conceptual form:

```go
Handle(place PlaceRef, primitive string, impulse skyra.Impulse) (any, error)
```

Or if we keep `primitive` inside the impulse envelope:

```go
Handle(place PlaceRef, impulse skyra.Impulse) (any, error)
```

Both are viable, but the core idea is the same:

- place
- primitive
- impulse

## Open Question

Should `primitive` be passed separately, or should the gateway trust the
primitive on the impulse and just route on:

- `place`
- `impulse`

My current leaning:

- keep `primitive` on the impulse
- pass `place + impulse`

That keeps the contract closer to the runtime's existing shape.

## Distributed Runtime Principle

An important clarification:

the companion / remote device should likely be understood as another runtime
realm, not merely as an adapter host.

Why:

- the device owns the world it represents
- an agent can move into that world
- once inside, it is traversing that world's local topology
- the same kernel law applies there as everywhere else

So the architecture is better described as:

- multiple local worlds
- same kernel semantics everywhere
- devices own the worlds they expose
- agents traverse between worlds through membranes

This means the gateway is not merely API middleware.

It is better understood as an inter-world boundary / transport layer that helps:

- discover remote worlds
- cross membranes
- route perception
- route engagement
- preserve the same kernel semantics across device-owned worlds

## Consequence

The distributed system is not:

- one runtime calling remote helpers

It is closer to:

- many worlds
- one kernel law
- local ownership of world state
- crossing between realms

## Short Version

The same kernel should exist everywhere.

Devices own worlds.

Agents move between them.

## Workspace Projection Principle

Another important clarification:

the "big brain" is not best understood as a separate magical control center.

It is better understood as the place where the base world lives.

That base world can also be understood as the agent's workspace.

## What The Workspace Is

The workspace is:

- the base world where the agent stands
- the active coordination surface for task work
- the place where remote world state can be projected for local manipulation

So instead of flattening all device worlds into one global soup, the system can:

- keep each device world locally owned by its device
- project relevant subsets of those worlds into the workspace
- let the agent arrange and manipulate those subsets to complete a task

## Projection Idea

The workspace can host projected or mounted subsets of remote worlds.

Examples:

- a subset of a Mac filesystem
- a subset of phone photos
- a subset of browser tabs
- a subset of reminders
- a subset of app state from multiple devices

This is similar to mounting multiple filesystems into one working area, except
the mounted things are more general than filesystems. They are slices of
spaces/worlds.

## Why This Matters

This gives a clean answer to cross-device task work.

The system does not need to:

- globally merge all worlds permanently

Instead it can:

- selectively project task-relevant subsets into workspace
- preserve local ownership of truth at the source world
- let the workspace become the active assembly surface

## Consequences

- locality still matters
- source worlds still own truth
- workspace is task-shaped
- cross-device composition becomes natural
- the agent can shape its working environment to fit the task

## Examples

For a cleanup / organization task, the agent could project into workspace:

- duplicate photos from phone
- duplicate files from laptop
- screenshots from tablet
- recent downloads from MacBook

Then reason over them together in one active task surface without destroying
the fact that each originated in a different world.

## Short Version

The base world is the workspace.

Remote worlds stay locally owned.

Task-relevant subsets can be projected into workspace for manipulation.

## Constructed Workspace Components

Another natural consequence of the workspace model:

agents may be able to construct their own temporary components inside the
workspace.

These would not need to be predeclared device worlds. They could be temporary,
task-shaped structures created for coordination, comparison, synthesis, or
multi-agent work.

## Example

If there are multiple agents active in the same workspace, they could create a
shared coordination space such as a "meeting room."

That room could:

- gather participants
- hold projected subsets from multiple worlds
- provide a bounded local surface for discussion or task work
- exist only for the duration of the task

The specific example may change, but the deeper idea is:

the workspace is not just a passive projection surface.

It may also be a generative surface where agents create intermediate
task-specific components of their own.

## Possible Constructed Components

- meeting room
- cleanup desk
- comparison table
- merge space
- investigation board
- handoff room
- planning cell

## Three Classes of Workspace Things

The workspace may eventually contain at least three classes of things:

1. native base-world spaces
2. projected subsets of remote worlds
3. constructed task components created by agents

## Why This Matters

This allows the workspace to adapt not only informationally, but also
structurally and socially.

Instead of only bringing information into the workspace, agents can also shape
the workspace itself to better fit the task they are trying to accomplish.

## Open Question

Are constructed workspace components:

- full spaces
- temporary worlds
- lighter-weight task artifacts
- some staged progression between those states

This remains open, but the important principle is already visible:

workspace can be shaped by the agents that inhabit it.

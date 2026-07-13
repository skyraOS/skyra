# Distributed Systems Overview Spec

## Working Idea

Devices on a local network are modeled as places in the world.

Examples:

- laptop
- phone
- PC
- TV

An agent can enter a device-place and interact with that place through an API abstraction layer that exposes the device's available affordances.

The goal is not to treat devices as generic remote tools. The goal is to treat them as inhabitable places with local capabilities, boundaries, and state.

## Core Shape

- the world contains places
- some places are physical / spatial
- some places are device-backed
- each place is a collection of spaces
- each device is also a collection of spaces
- the local network connects device-places into one traversable environment
- when an agent enters a place, it perceives and acts through that place's local interface

World affordance execution follows the durable-substrate / ephemeral-worker
pattern described in [Ephemeral World Workers](./ephemeral-world-workers.md).
Places retain authoritative state; a situated `perceive`, `engage`, `move`, or
device affordance is instantiated as bounded temporary computation with an
explicit outcome.

This means a device-place is not one flat API surface. It is an environment made
of multiple internal spaces that may expose different affordances and
boundaries.

Examples:

- a phone may expose spaces like photos, contacts, messages, reminders, files
- a laptop may expose spaces like terminal, filesystem, browser, downloads,
  clipboard
- a TV may expose spaces like media playback, queue, inputs, now-playing state

## Why This Matters

This creates a distributed system that still preserves Skyra's core ontology:

- locality matters
- boundaries matter
- traversal matters
- capabilities are place-specific
- the agent acts from somewhere, not from nowhere

Instead of "call a device API from a central orchestrator," the system becomes:

- enter a place
- perceive what is available there
- engage local affordances
- move between places
- preserve continuity across a distributed environment

## Settled Spine — The Membrane Is the Mount

This section is decided, and it reconciles the exploratory framings below (the
Membrane / Portal model, the Locality model, and — in `place-gateway/initial-contract.md`
— the Workspace Projection model). Read the rest as elaboration on this spine.

### One membrane

There is not a "being membrane" and a separate "device-place membrane." There is
one construct: the **agent ↔ world mount** — the boundary where an agent meets
whatever world it is standing in.

- the base world is one thing the agent can seat onto
- a device (MacBook, phone, TV) is another
- moving worlds = re-seating the same membrane onto a different world
- the membrane was never being-specific

A device world implements the same `world.Space` interface the grid implements
now. The agent's Observe/Express against it is unchanged.

### The decomposition

- **agent = cognition** — the nucleus (DNA, entity index, groupings, accumulated
  experience). This travels with the membrane.
- **world = topology + live state** — owned by the device, kept local, hydrated by
  perception (see Locality and State Sync below). This stays home.
- **membrane = the mount** — one re-seatable construct where the two meet.

### This is already the code

The dnav2 Orchestrator *is* this membrane: it holds the nucleus and runs waves
against a world. To "load into the MacBook," point that same orchestrator's world
reference at the MacBook's `world.Space`. Nothing in cognition or the kernel
changes. All new work lives strictly *below* the membrane — the transport that
lets a `world.Space` live on another machine and marshal a crossing.

### Relocation and projection are one operation at two granularities

Both are `mount(Space)`, differing only in granularity and attachment:

- **Relocation** — mount a *whole* device world; it *replaces* what you stand on
  (re-seat). The mounted thing is the entire MacBook `world.Space`.
- **Projection** — mount a *partition* of a device world; it *attaches* to the
  base world / workspace as a child (add-child). The mounted thing is a sub-`Space`
  (a slice of the MacBook).

Because they are the same primitive, building relocation first is not throwaway —
it builds the exact machinery projection needs. Projection is relocation where the
mounted thing is a partition and it hangs off the workspace instead of swapping
under the agent's feet. You do not pay twice.

### Sequencing

1. **Now — relocation.** The agent goes to the MacBook world and traverses it.
2. **Next — projection.** The agent carves partitions of device state /
   capabilities and projects them onto the base-world workspace, composing across
   devices without moving.

### Above / below the membrane

- **Ontology (above the membrane)** — crossing, re-seating, mounting. Provable
  *entirely locally*: the first MacBook world can be a stub `world.Space` with one
  or two spaces and no transport at all. Agent perceives the `macbook` membrane in
  the base world, engages it, membrane re-seats onto the stub.
- **Transport (below the membrane)** — discovery, marshalling, the external API
  layer. Swapped in later by replacing the stub `world.Space` with an API-backed
  one. The test that the boundary held: **the agent cannot tell the stub from the
  API-backed world.**

### Minimal first slice

- a device-membrane node in the base world (perceivable, engageable)
- a stub MacBook `world.Space` behind it (one or two spaces, no network)
- the re-seat on engage

No network, no adapters. Prove the agent can cross into a second world and perceive
it is somewhere new. Everything real slots in underneath afterward.

## First-Pass Components

### 1. World

The containing environment that holds all places.

### 2. Device Places

Each networked device is represented as a place.

Examples:

- laptop place
- phone place
- PC place
- TV place

Each device-place contains its own internal spaces.

The device model is therefore:

- world
  - place
    - spaces

or, when the place is device-backed:

- world
  - device-place
    - device spaces

### 3. Place API Layer

Each place exposes a controlled API abstraction that maps real device capabilities into traversable affordances.

This layer should:

- expose only real available capabilities
- preserve place-specific boundaries
- hide transport / implementation details from the agent
- map device APIs into spaces, not just commands

The adapter question is not just:

- what API calls does the device support?

It is:

- what spaces does this device expose?
- what affordances belong to each space?
- what can be perceived, engaged, traversed, or modified there?

### 4. Network Layer

The substrate that allows device-places to communicate over a local network.

This is the actual distributed systems layer under the ontology.

### 5. Agent Entry / Presence

An agent can enter a device-place and act from inside it.

Open question:

- is entry persistent presence, temporary traversal, session binding, or some combination?

## Early Design Principles

- devices are places, not just tools
- places are made of spaces
- local network topology should matter
- the system should preserve spatial / ontological coherence
- transport details should stay below the place boundary
- each place should expose only its true affordances
- the architecture should feel different because it is different

## Device Adapter Direction

The next practical step is to identify what each type of device exposes as
spaces and then write adapters around those spaces.

This suggests a workflow like:

1. choose a device type
2. enumerate the spaces it exposes
3. define the affordances of each space
4. identify the real API surfaces available
5. write an adapter that maps those APIs into the place/space model

Examples:

- phone adapter
  - photos space
  - contacts space
  - reminders space
  - files space
- laptop adapter
  - terminal space
  - filesystem space
  - browser space
  - app/window space
- TV adapter
  - playback space
  - library space
  - input/control space

This keeps the design anchored in ontology first and implementation second.

## First-Pass API Layer Design

The first implementation should likely use an external API layer that the
runtime calls across a boundary.

Initial instinct:

- keep the runtime as cognition / traversal
- keep device integration outside the runtime
- let the runtime call a local external layer when a place needs to expose or
  act on real device state

This matches the broader Skyra pattern:

- the runtime should not become a pile of device-specific code
- transport and platform details should live below the place boundary
- the core system should stay ignorant of device internals

### Proposed Shape

The external layer sits between:

- Skyra runtime
- local networked devices and device-specific APIs

So the stack becomes:

- Skyra runtime
  - place / space boundary
  - external API layer
  - device adapters
  - actual devices

### Responsibilities of the External API Layer

- discover devices on the local network
- identify device type
- expose the spaces available on each device
- normalize capability access behind a stable contract
- handle auth / sessions / pairing
- perform reads and writes on behalf of the runtime
- return results in a shape the runtime can map back into place-space ontology

### Responsibilities of the Runtime

- decide where to go
- enter places
- traverse spaces
- perceive available affordances
- choose actions
- preserve continuity, state, and cognition

The runtime should not need to know:

- whether a phone call used Swift, Shortcuts, HTTP, gRPC, WebSocket, or a local
  bridge
- whether the TV is LG, Apple TV, or something custom
- whether a filesystem request was local disk, SMB, or an app sandbox bridge

It should only know:

- this place exists
- these spaces exist within it
- these affordances are available here
- this action succeeded / failed / returned state

### Why External First Makes Sense

This seems like the right first design because:

- it preserves kernel/runtime cleanliness
- it makes adapters independently evolvable
- it gives you one place to handle heterogeneous devices
- it lets the distributed system grow without bloating the core runtime
- it matches the "places expose affordances" idea better than linking platform
  code directly into the runtime

### First-Pass Contract Idea

At a high level, the runtime probably needs only a few categories of calls into
the external layer:

- discover places
- describe a place
- describe a space
- perceive current state
- engage an affordance
- mutate state where permitted

Possible conceptual contract:

- `ListPlaces()`
- `DescribePlace(placeID)`
- `ListSpaces(placeID)`
- `DescribeSpace(placeID, spaceID)`
- `Perceive(placeID, spaceID, query?)`
- `Engage(placeID, spaceID, action, payload)`

This is not a final API, just a first design direction.

### Important Design Constraint

The API layer should not collapse the ontology back into generic RPC.

Bad shape:

- runtime sends arbitrary commands to arbitrary devices

Better shape:

- runtime enters place
- sees spaces
- perceives local state
- engages local affordances through the external layer

The external layer is an implementation substrate, not the conceptual model.

## Membrane / Portal Model

For the first version, the cleanest design is to treat a device place as a
place with a membrane that acts like a portal boundary.

Example:

- base world contains a `macbook` membrane
- the `macbook` membrane is the boundary of the macbook place
- engaging that membrane moves the agent into the macbook world
- once inside, the agent traverses the macbook world as its own place

This is better than treating the first device interaction as direct API usage.

The API layer remains underneath the ontology. The agent engages a membrane, not
an endpoint.

### First Bare Shape

For now, keep the membrane very simple.

It can accept:

- `perceive`
- `move`
- `engage`

But:

- `perceive` reveals that the membrane / place exists
- `move` allows spatial approach / positioning
- `engage` acts as a portal crossing

### Base World Example

In the base world:

- the agent can perceive the `macbook` membrane
- the agent can move toward or onto it
- the agent can engage with it

When the agent engages with the membrane:

- the system crosses the boundary
- the agent is now inside the macbook world

### Macbook World Example

Inside the macbook place:

- the agent is now in the macbook world
- the macbook world has its own internal spaces
- the membrane still exists as a boundary / return portal

Possible internal spaces later:

- clipboard
- downloads
- running apps
- browser
- desktop

But those do not need to be fully designed yet.

### Why This Is Good

- it preserves the place ontology
- it keeps the boundary spatial and legible
- it avoids collapsing the first interaction into raw RPC
- it gives a clear distinction between crossing into a place and perceiving
  things within that place

### Membrane Responsibilities

- advertise that the place exists
- expose minimal metadata on `perceive`
- allow crossing on `engage`
- serve as the return boundary from inside the place

### Place Responsibilities

Once crossed into the place:

- expose internal topology
- expose internal spaces
- allow internal perception and engagement
- use the external API layer only as backing implementation

### API Layer Responsibilities Under This Model

The API layer does not define the conceptual crossing.

Instead it:

- backs the internal spaces of the place
- provides real device state once the agent is inside
- remains invisible at the ontology layer

### First Principle

The agent should engage a membrane, not an API endpoint.

## Locality and State Sync

An important distributed-systems insight:

the runtime does not need to keep a fully accurate global mirror of every
device-space at all times.

Instead, locality can be used as an advantage.

### Core Idea

Perception updates state on the Mac mini runtime.

This means:

- a space becomes current when it is being observed
- detailed state only needs to stay fresh while that space is under active observation
- the runtime does not need to maintain high-fidelity live state for every
  device-space all the time

### Consequence

The distributed world model becomes attention-shaped rather than globally
synchronized.

That means:

- if no one is looking at a space, it does not need to remain fully hydrated
- if an agent enters a place and perceives a space, that perception updates the
  local runtime representation
- while a space remains under observation, the runtime can keep it current
- when observation stops, the space can cool, compact, decay, or become stale

### Proposed State Model

#### 1. Structural State

State that can remain available even when the place is not actively observed.

Examples:

- this device exists
- this place exists
- these spaces exist within it
- these affordances are available here

This changes infrequently and can stay cached cold.

#### 2. Dynamic State

State that should only be kept highly current when it is actively perceived.

Examples:

- current photos listing
- now playing
- current reminders
- active window state
- temporary device conditions

This state is refreshed and maintained through perception.

### Better Sync Philosophy

Bad model:

- globally sync all device state all the time

Better model:

- keep structural knowledge persistently
- hydrate dynamic knowledge through perception
- keep detailed state current only while the space is being observed

### Why This Is Good

- reduces sync pressure
- keeps the representation honest
- uses locality as an operational advantage
- makes the distributed world feel like a lived perceptual surface, not a giant
  stale database
- preserves the idea that what is real in the runtime depends on what has
  actually been observed recently

### Short Version

The Mac mini runtime should not be a universal mirror of every device.

It should be the active perceptual surface of the distributed world.

## Open Questions

- What does it mean, concretely, for an agent to "enter" a place?
- What does it mean to enter a space within a place?
- What capabilities are universal across places, and what capabilities are local?
- How does movement between places work?
- How should presence be represented?
- What state belongs to the place versus the world versus the agent?
- How visible should the network topology be to the agent?
- How should discovery work on the local network?
- How should permissions and trust work between places?
- What is the minimum viable slice of this system?
- What are the canonical spaces for each device class?
- What is the minimal stable contract between runtime and external API layer?
- Should the external layer be one broker or many per-device services?

## Word Vomit Section

Dump raw thoughts here. We can organize them afterward.

# The sky dreams: the desktop sky as a machine hallucination

Proposal: **Not in the proposal.** A long-term direction for the desktop sky;
nothing here is planned work yet.

Date: 2026-10-02  
Status: proposed; cut 1 built (2026-10-02): the sky is drawn with Metal at the
display's rate, with the nebula and the flow's filaments as styling ahead of
the mapping. Builds on [23](23-desktop-app.md) (the app and its sky),
[24](24-the-room.md) (Boundaries, the room), 30
(types, routes, the novelty interrupt) and [20](20-boundry-crossing.md) (the
rule for a new crossing). The look it reaches for is Refik Anadol's *Machine
Hallucinations*: an archive pushed through a machine and shown as what the
machine dreams of it, dense, fluid, pigment-like.

## The rule

> **Every pixel is derived from the tree. Nothing on the sky is decoration.**

Borrowed as a style, the Anadol look is a screensaver and works against the
thesis. Borrowed as a method, it fits the runtime exactly. Skyra already has
both halves of a machine hallucination:

- **The archive** is the tree: hash-linked and content-addressed.
- **The dreaming** is replay: a machine re-deriving its own past, byte for
  byte, from genesis.

So the sky is a projection of the lineage, nothing more. Authority is derived,
never held; so is the image. And because replay is deterministic, so is the
picture: the same tree hash at frame N draws the same frame, and a test can pin
it the way `SkyTests` pins the star hash today.

## Three states

| State | When | What the sky shows |
|---|---|---|
| **Rest** | no conversation, or nothing new | the sky as built (23): the TUI's stars, the satellites, the quiet. Density is earned |
| **Work** | a conversation is running | Boundaries become fields instead of faint ellipses: particle clouds as dense as what they reach, currents along routes. A call is a plume drawn from the core to its surface's star, not a line of light |
| **Dream** | the app is idle or in the background, or on request | the lineage replayed from genesis as a fluid time-lapse, scrubbable. Each committed record lands as a burst of pigment; a conversation's history becomes a sculpture to move through |

The current dark and quiet stay the default. The density grows with what has
actually happened: a fresh workspace is nearly empty, a month-old one is a
nebula.

## The mapping

Each visual element answers to one thing in the runtime:

| In the runtime | On the sky | Why |
|---|---|---|
| a **record** in the tree | a particle burst, seeded by its hash | content-addressed, so it lands in the same place every time, as surfaces do in `BoundaryMap` today |
| a **lineage** | a current in the flow field: curl noise steered along parent → child | conversations read as rivers, branches as forks |
| a **type**, `origin{shape}` (30) | a pigment: hue and grain from the shape hash, placed in a 2D latent laid out by type | returns of one type pool into one colour |
| a **novelty interrupt** | a bloom of a pigment the field has never shown | the event that matters and is rare: the parent meeting something new |
| a **route** | a standing channel; later returns of that type flow down it without blooming | the model is called once per kind of return, and you can see it |
| a **Boundary** | a basin whose wall the fluid cannot cross | boundedness made visible |
| a **refusal** (amber) | fluid pressing on the wall and pooling | the Boundary doing its job, not a failure (as `Theme.amber` already says) |
| a **request to cross** (23, cut 3) | the one place the field tears: the wall thins and light comes through | the app's only interruption, so the only visual shock |

**Structure only.** The sky draws hashes, types, edges and counts, never file
contents or model text. The grant says contents go to the model provider; the
sky adds no other place they go, and it does not become a second, worse
transcript.

## Where it sits

The app stays what 23 says it is: a client of the webhost's conversation
protocol, with no descriptors, no `.skyra/`, no authority. The one new thing it
needs is the lineage as data.

### The rule for a new crossing, answered

| Question (20) | For the lineage stream |
|---|---|
| Who emits it, who checks it? | the webhost emits it from the lineage it already reads for Boundaries (`hosts/web/boundaries.go`); the app only renders it |
| Is it authority? | no. Records' hashes, types, origins and parents describe what was decided; nothing in the app acts on them |
| Is it state the runtime acts on? | no. It is a read of the tree, sent one way |
| Does it name an actor? | no new one |
| Is it a descriptor? | no. `{hash, parent, type, origin, kind}`, no content and nothing from `.skyra/` |
| Who holds it while nothing runs? | the tree. The app keeps at most a cache of the layout |

## Cuts

| Cut | Scope | Depends on |
|---|---|---|
| **1. Metal** (built) | the sky moves from `Canvas` to an `MTKView`: a draw list of sprites, segments and lights, drawn in instanced passes at the display's rate. Parity: same stars (the tests still pin the hash), satellites, lens. The particles are still stepped on the CPU (about 1,800); a compute pass is what 10⁵–10⁶ will need | nothing new |
| **2. The lineage stream** | a protocol message carrying records as structure, extending what the webhost sends for Boundaries | the webhost's lineage read |
| **3. Work** | Boundaries as basins, calls as plumes, routes as channels; today's `BoundaryMap` layout becomes the field's attractors | 1, 2 |
| **4. The type latent** | a deterministic, stable 2D layout of type space from shape hashes, so a type keeps its colour forever; novelty blooms | 3; types by origin (30, built) |
| **5. Dream** | replay as time-lapse, a scrubber, speed from commit times. Optionally a lineage exported as a rendered video: a run you can show | 3, 4 |
| **6. The tear** | the request to cross, drawn | 23's cut 3: faults that climb to the human and the room's admin facet (24, 26) |
| **7. Room scale** | full screen, an external display, a screensaver target: a workspace's whole history as a wall-sized data sculpture | 5 |

Cut 1 is engineering with nothing to show for it, and everything after needs it.
Cut 6 lands with the mechanism, never before.

## Invariants

- **Derived, not decorative.** Every element traces to a record. Golden-frame
  tests pin `(tree hash, frame) → image hash`; a change that breaks one is a
  change to the mapping, and says so.
- **Text wins.** In Work the conversation stays fully legible; the field dims
  and calms behind it. Only Dream is maximal.
- **Motion is quiet and costs little** (23, unchanged). Reduce Motion makes
  Dream a still frame. On battery, Rest and Work at reduced particle counts.
  Hidden or occluded, the sky stops.
- **The TUI is the sketch.** A terminal cannot draw this and need not. The
  clients share the palette and the meaning (the same hashes put things in the
  same places), not the fidelity. This answers 23's "two copies of the
  palette": one `design/theme.json` for the tokens, each client rendering as
  richly as it can.

## Limits, and what is open

- **The gimmick risk.** If the mapping ever loosens to "make it look good",
  the idea is gone. The golden frames are the guard.
- **Density needs history.** A new user sees almost nothing, which is right,
  but the first launch may want a demo lineage (`skyra serve --fake`) to show
  what is coming.
- **Is Dream ambient or asked for?** Starting on idle makes the app a presence;
  asking keeps it a tool.
- **The browser frontend.** A WebGPU port, or the reason 23's open question
  ("does the browser frontend retire?") is answered yes.
- **Room scale** as a real goal (a demo, an exhibition piece) or a north star
  only.

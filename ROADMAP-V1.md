# Roadmap to v.1

June 2025 → January 2026. Six phases. From first multi-agent alpha to production-ready ship.

## What's done

- Reality interface, three decay curves, thread binding, four primitives (v.1 branch)
- Exchange routing, think/act layers, memory graph, operators (v.05 on main)
- Frontend scaffold with live snapshot rendering, mock server, Zustand + Zod pattern
- Specs: places, binding realities, activation equation, nervous signals, primitives

## Demo target

A session where 3-6 beings (skyra, builder, philosopher, louise, + user) are freely exchanging — multi-party threads, beings addressing each other, the user participating as a peer. Activation weights govern who responds, how deep they think, what they remember. A being can navigate a changing maze using memory of where it's been — places as realities, engage to enter, look to survey, memory to recall the path.

## Milestones

### Weeks 1-2 — June 1 → June 15: Multi-agent exchange

- All beings bootstrapped on v.1 from the genome
- Multi-party thread routing working — being A talks to being B, being B routes to being C
- Activation experiments running — captain's log entries with real numbers
- Competence thresholds tuned: pass-through / pick-up / propagate boundaries
- Frontend connected to real v.1 runtime (not just mock)

### Weeks 3-4 — June 15 → June 29: Places, act service, and persistence

- Places implemented as Realities — engage crosses into them, tools surface on the Relation
- Act service wired — pipe boundary between runtime and consequence, Composio free tier as phase 1 backend (20k calls/month, $0)
- Place tools are act expressors — JSON over pipe, block name + params out, result back
- Tutorial room: a place with tools, a being enters and uses them through the act service
- Memory persistence — save/load graph across restarts (BoltDB or flat file, whatever the pain says)
- Recall by entity working — being queries memory, graph returns relevant nodes
- Place-as-binding-reality: being in a place has higher activation on memories formed in that place

### Weeks 5-6 — June 29 → July 13: Maze and demo

- Maze built as a place with changing topology — walls move, paths close
- Being navigates from memory — remembers where it's been, avoids dead ends, finds the exit
- Weight tuning under real multi-agent load — activation thresholds, decay rates, signal attenuation
- Frontend renders the live session — exchanges flowing, reality graph updating, beings thinking
- Demo polish: clean session flow, being personalities visible, maze navigation legible

### July 14: Buffer

Something will slip. This day absorbs it.

### July 15: Alpha

## Risks

| Area | Risk | Mitigation |
|---|---|---|
| Maze navigation | Memory + place binding + spatial decisions composing for the first time | Build the tutorial room first — validate place + memory before adding the maze |
| Multi-party routing | Thread routing loop in v.1 is new — v.05 had it but the traversal mechanics changed | Get 2-being exchange solid before scaling to 3-6 |
| Activation tuning | Thresholds are empirical — could take more iterations than expected | Start experiments week 1, don't wait for everything to be built |
| Memory persistence | Storage format not decided — could eat time on serialization | Start with flat JSON files, upgrade when it hurts |

---

## Phase 2 — July 16 → August 5: Ad-hoc agent bridges

Beings build their own adapters to external agent architectures at runtime. No config files, no downloads, no restarts. The user says "connect to my CrewAI team" and the being generates the adapter, validates it through the act service pipe, and wires it into its own topology. If it works, it gets reinforced. If it breaks, stress catches it and the being rewrites it.

### Weeks 7-8 — July 16 → July 29: Bridge generation

- Adapter shape standardized — every bridge is JSON in, pipe out, JSON back. Same act service contract.
- Being can generate an adapter from a description of the external system's API (REST, CLI, MCP, pipe)
- Generated adapters land as act expressors in the being's topology — reachable through activation, not config
- Validation loop: being generates adapter → tests it through the pipe → result comes back → reinforce or rewrite
- First bridges: Claude Code (CLI pipe), AutoGen (REST), CrewAI (REST), MCP servers (stdio/SSE)

### Week 9 — July 30 → August 5: Promotion and discovery

- Adapters that prove themselves under use get promoted from generated glue to proven blocks
- Beings share proven adapters — a block that builder validated becomes available to skyra through the topology
- doesNotUnderstand seeds new adapters — being tries to reach something that doesn't exist, generates the bridge
- No user-facing config. The user asks for a capability, the being figures out the bridge, the topology grows.

### What this phase proves

- External agents are native citizens — two lines in the genome, a Reality whose invariant is a pipe
- The being owns its own integration surface — it builds, tests, and promotes its own bridges
- The promotion gradient applies to adapters the same way it applies to memories and skills
- Zero config for the user. The being handles the boundary.

---

## Phase 3 — August 6 → September 2: The Projection — cold start (4 weeks)

The user's own digital brain. Same DNA, same `Realize()`, but the leaf is the human, not a provider. The topology thinks and waits for you. The hard problem is cold start — a new user has an empty topology. No memories, no weights, no grooves. The system has to bootstrap a useful digital brain from nothing without asking the user to fill out a profile or import data.

### Weeks 10-11 — August 6 → August 19: Seeding

- User topology bootstrapped from two sources: conversation (real-time) and device (historical)
- Every exchange with any being creates memory nodes in the user's topology, not just the being's
- Device scrape as binding realities — the laptop is years of accumulated signal already sitting there:
  - **Calendar** → temporal binding reality. Meetings, deadlines, recurring patterns. The projection knows what's coming.
  - **Filesystem** → project topology. Directory structure, recent files, modification patterns. The projection knows what the user is working on.
  - **Browser history** → attention pattern. What the user has been reading, searching, returning to. The projection knows what the user is thinking about.
  - **Contacts / messages** → relationship graph. Who the user talks to, how often, about what. The projection knows who matters.
- These aren't data imports — they're binding realities. They imprint fields on the Relation the same way thread and emotion do. The user's calendar doesn't get "loaded into memory." It becomes a binding reality that raises activation on everything time-relevant.
- The user's competence reads which device has attention — terminal vs websocket vs phone shapes what surfaces
- First experiment: fresh user, device scrape + 15 minutes of conversation → projection has deep enough weight structure to surface something useful without being asked

### Weeks 12-13 — August 20 → September 2: Surfacing and validation

- The projection runs in parallel — observes the user's exchanges, develops opinion, promotes understandings
- Surfacing mechanism: the projection pushes to the frontend when activation crosses threshold — not a notification, a weight event
- The frontend renders what the projection thinks matters through the user's own activation landscape
- Iterate on the seeding → surfacing loop — how fast does the topology become useful, what's missing, what's noise
- Cold start validation: new user, empty topology, 30 minutes of natural use → the projection surfaces something the user didn't ask for but needed
- The line between "the user's memory" and "the user's assistant" dissolves — same topology, same traversal

### What this phase proves

- The user is a Reality with their own topology, not a consumer of Skyra's
- Cold start solves through natural use + device scrape — the conversation and the laptop together seed a rich topology
- The projection surfaces relevance without being asked — weighted by the user's own attention history
- Same DNA, different leaf. The architecture holds.

---

## Phase 4 — September 3 → September 30: Personal network — distributed devices (4 weeks)

The user is one being across many devices. The laptop, the phone, the tablet — each is a world with its own components. The user's topology spans all of them. Competence reads which device has attention and routes there. An HTTP call is a Relation — the payload serializes across the boundary, the remote universe's `Realize()` fires, a string comes back. The network is just another port on the device.

### Weeks 14-15 — September 3 → September 16: Multi-device world

- Each device is a MacOS/iOS/Android world with its own components — same genome pattern, different devices
- The user's being exists in all worlds simultaneously — same identity, same topology, same weights
- Competence determines which device has attention right now — the Relation routes to the active device
- Serialization across device boundary — Relation serializes to JSON, crosses HTTP/WebSocket, deserializes on the other side
- Same `Realize()` on both sides. The remote device doesn't know the Relation came from the network.

### Weeks 16-17 — September 17 → September 30: Topology sync and convergence

- Weight updates propagate across devices — the user's phone conversation updates the laptop's topology
- Conflict resolution through the activation equation — concurrent updates on two devices merge through weight, not timestamps
- The projection runs on whichever device has capacity — compute follows attention
- Offline resilience: each device holds a full topology copy, syncs on reconnect, weights converge through EMA
- Edge cases: simultaneous input on two devices, network partition during traversal, stale weights after long offline
- Validation: user starts a conversation on the laptop, walks away, picks up the phone — the projection continues with full context, no handoff

### What this phase proves

- The personal network is the uncopyable move — the user's topology spanning devices is not a feature someone adds, it's a substrate property
- The network boundary is invisible — same interface, same traversal, the pipe is just longer
- The activation equation handles distributed convergence — no CRDT, no vector clocks, weights merge naturally

---

## Phase 5 — October 1 → October 31: Dynamic attention UI (4 weeks)

The frontend renders through the user's own activation landscape. Components don't just check if data exists — they check if the data scores high enough on the user's topology right now. The UI is the user's nervous system. What's bright is what the topology thinks matters. What's dim is what decayed. What's gone is what fell below threshold. This needs to feel right — not just work, but feel like the UI is alive and paying attention to you.

### Weeks 18-19 — October 1 → October 14: Activation-driven rendering

- Each frontend component receives its activation score from the user's binding fields
- Components scale visibility by activation — size, brightness, position, presence all derived from weight
- High-activation content surfaces without the user asking — the projection pushes, the UI renders
- Low-activation content fades — not hidden, just quieter. Still reachable if the user looks.
- The three decay curves (recency, familiarity, identity) drive the visual hierarchy — what you just touched is loud, what you always touch is steady, what you haven't touched fades
- Animation and transition design — the fading, surfacing, reshaping needs to feel organic, not mechanical

### Weeks 20-21 — October 15 → October 31: Attention feedback loop and polish

- User interaction feeds back into the topology — clicking, scrolling, dwelling on a component bumps its activation
- The UI reshapes itself from use — a user who always checks thought history sees it prominent. A user who ignores the desk watches it shrink.
- Per-user rendering without configuration — the activation equation IS the personalization. No settings, no preferences, no profiles.
- The higher-order activation function: the user's topology modulates what the activation equation surfaces for them. Each user sees a different UI from the same snapshot because their weights are different.
- Polish pass: transitions, timing, visual weight, the feel of a UI that knows you
- Validation: two users, same runtime, same beings — each frontend renders a different view shaped by their own attention history

### What this phase proves

- The UI is not a dashboard. It's a sense organ.
- Personalization is a property of the substrate, not a feature of the frontend
- The attention feedback loop is self-reinforcing — use shapes weights, weights shape the UI, the UI shapes use
- No one can replicate this without the user being a Reality inside the same physics as everything else

---

## Timeline summary

| Phase | Window | Duration | What lands |
|---|---|---|---|
| 1 — Alpha | June 1 → July 15 | 6.5 weeks | Multi-agent exchange, activation tuning, places, maze from memory |
| 2 — Bridges | July 16 → August 5 | 3 weeks | Beings build their own adapters at runtime |
| 3 — Projection | August 6 → September 2 | 4 weeks | User's digital brain, cold start from device + conversation |
| 4 — Personal network | September 3 → September 30 | 4 weeks | One being across many devices, topology sync |
| 5 — Attention UI | October 1 → October 31 | 4 weeks | Frontend renders through user's activation landscape |

**Total: June 1 → October 31 — 5 months, 5 phases.**

---

---

## Phase 6 — November 1 → January 31: Production hardening (3 months)

The architecture works. Now make it real. Permissions, testing, failure modes, security boundaries, operational maturity. Everything before this was "does it work." This phase is "does it hold under pressure from people we don't know."

### November — Permissions and trust boundaries

- Permission system designed from the topology — the being's reachable graph IS its access control. No ACLs bolted on.
- User consent model: what can the projection see? What can a being act on? What requires the human to approve?
- Device scrape permissions — calendar read vs calendar write, filesystem scope, browser history opt-in
- Act service security audit — the pipe boundary is the trust line. Validate that nothing crosses it that shouldn't.
- Multi-user isolation — two users on the same runtime can't see each other's topology, projection, or device data
- Governance spec from world-physics.md moves from theory to enforcement — proposals, thresholds, vote mechanics

### December — Testing and failure modes

- Integration test suite: multi-agent exchange, place navigation, memory persistence, act service round-trip, device sync
- Chaos testing: kill a device mid-traversal, corrupt a weight file, drop the network during sync, exhaust the token budget
- Performance under load: 6 beings, 10 open exchanges, 50 memories per being, continuous traversal — does the activation equation stay responsive?
- Cold start regression: new user, device scrape, 30 minutes — does the projection still surface useful signal at scale?
- Memory pressure: what happens when the graph has 10,000 nodes? 100,000? Where does it break and what does the degradation curve look like?
- Frontend stability: snapshot stream under high-frequency updates, attention UI with many competing activation scores, animation performance

### January — Production readiness

- Deployment pipeline: build, test, ship to a real device that isn't your laptop
- Persistence hardened — graph storage survives crashes, restarts, upgrades. Migration path for schema changes.
- Monitoring: what does the operator see? Health of the topology, inference costs, traversal depth, memory growth, sync lag
- Error handling: what does the user see when something breaks? Graceful degradation, not a crash.
- Documentation: genome format, act service block contract, frontend component API, deployment guide
- Security review: credential handling, pipe boundary, device data in transit, multi-user isolation
- Alpha user onboarding: first person who isn't you uses it. What breaks.

### What this phase proves

- The system holds under adversarial conditions — not just happy path demos
- Permissions are topological, not bolted on — the architecture IS the access control
- The cold start, the projection, the personal network, the attention UI all survive real usage patterns
- Someone who didn't build it can use it

---

## Timeline summary

| Phase | Window | Duration | What lands |
|---|---|---|---|
| 1 — Alpha | June 1 → July 15 | 6.5 weeks | Multi-agent exchange, activation tuning, places, maze from memory |
| 2 — Bridges | July 16 → August 5 | 3 weeks | Beings build their own adapters at runtime |
| 3 — Projection | August 6 → September 2 | 4 weeks | User's digital brain, cold start from device + conversation |
| 4 — Personal network | September 3 → September 30 | 4 weeks | One being across many devices, topology sync |
| 5 — Attention UI | October 1 → October 31 | 4 weeks | Frontend renders through user's activation landscape |
| 6 — Production | November 1 → January 31 | 3 months | Permissions, testing, failure modes, production readiness |

**Total: June 1 → January 31 — 8 months. Ship-ready by end of January 2026.**

---

## What the full build proves

- Multiple beings can exchange freely in a shared world
- The activation equation governs routing, depth, and attention
- Places are realities with tools and binding fields
- Memory shapes behavior — a being that's been somewhere knows something a fresh being doesn't
- The frontend renders all of this live from the snapshot stream
- The user is a being in the world, not an operator outside it

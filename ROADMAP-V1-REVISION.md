# Roadmap to v.1 — Revision

June 2026 → January 2027. Six phases. Restructured around the core composition loop.

The original roadmap proved the architecture bottom-up — runtime, then bridges, then projection, then network, then UI. Each phase validated a subsystem in isolation. The revision brings the full differentiator loop forward: user intent → runtime activation → bridge selection → result → memory update → UI reorganizes. A crude but integrated version of that loop lands by mid-July, not October.

## The loop that has to prove early

```
user does something
  → runtime scores it (activation)
  → the right thing surfaces (bridge, memory, being)
  → result comes back
  → memory updates (weights shift)
  → UI reorganizes (what's loud changed)
```

Every phase deepens this loop. No phase exists in isolation.

## What's done

- Reality interface, three decay curves, thread binding, four primitives (v.1 branch)
- Exchange routing, think/act layers, memory graph, operators (v.05 on main)
- Frontend scaffold with live snapshot rendering, Zustand + Zod pattern
- Specs: places, binding realities, activation equation, nervous signals, primitives, act service

---

## Phase 1 — June 1 → July 15: Alpha with integrated loop (6.5 weeks)

Multi-agent exchange with the full composition loop working end-to-end. Not polished — but the loop is real.

### Runtime core (weeks 1-3)

- All beings bootstrapped from genome on v.1
- Multi-party thread routing — being A talks to being B, routes to being C
- Activation experiments running — captain's log with real numbers
- Competence thresholds tuned: pass-through / pick-up / propagate
- Memory persistence — flat file, survive restarts
- Frontend connected to real runtime

### One bridge (weeks 3-4)

- Act service pipe wired — JSON in, result back
- One external agent bridge: Claude Code (CLI pipe) or MCP server (stdio)
- Not four. One. Prove the adapter shape, the pipe boundary, the round-trip.
- The being can reach the bridge through its topology — activation decides when it fires

### Primitive attention UI (weeks 4-5)

- Generic activation score per entity in the store — a number, not coupled to the runtime's internal weight format
- Proxy scoring: status (active > idle), exchange entry count, thought history depth, memory count
- Frontend sorts and fades by score — high score is bright/prominent, low score dims
- Click or dwell on a component bumps its proxy score
- Ugly is fine. The loop has to work: something changes in the runtime → the UI reorganizes

### Conversation-seeded user relevance (weeks 5-6)

- Every exchange creates weight in the user's perspective — topics the user engages with score higher
- The frontend renders Michael's view differently from a hypothetical second user based on conversation history
- Filesystem context if there's time — project directory as a signal. Calendar and browser come later.

### Maze as runtime test (week 6)

- Places implemented — engage crosses into them, tools surface
- Tutorial room with tools, being enters and uses them through the act service
- Maze as a validation harness — not the headline demo. Proves memory + place binding compose.

### The headline demo

> User connects Claude Code as a bridge. Talks to Skyra for ten minutes about a codebase. Later asks a question. Claude Code surfaces automatically because the topology learned it's relevant. The UI shows it brighter than things the user hasn't been engaging with.

That's the differentiator in one session.

### What Phase 1 proves

- The full composition loop works end-to-end
- External agents are native citizens through the act service pipe
- The UI reorganizes from activation, not configuration
- Memory shapes what surfaces — what you've engaged with is louder

---

## Phase 2 — July 16 → August 5: Deepen bridges (3 weeks)

Phase 1 proved one bridge. Phase 2 proves the bridge is a pattern, not a special case.

### Weeks 7-8: More bridges

- Second and third bridges: pick from AutoGen (REST), CrewAI (REST), MCP servers (SSE)
- Adapter shape validated across protocols — CLI pipe, REST, stdio all go through the same act service contract
- Being can generate a simple adapter from a description of an external API — tested through the pipe, reinforce or rewrite

### Week 9: Promotion and discovery

- Adapters that prove themselves under use get promoted from generated glue to proven blocks
- Beings share proven adapters through the topology
- doesNotUnderstand seeds new adapters — being reaches for something that doesn't exist, generates the bridge

### What Phase 2 proves

- The bridge pattern generalizes — not one integration, a mechanism
- The being owns its integration surface
- The promotion gradient applies to adapters

---

## Phase 3 — August 6 → September 2: The Projection — cold start (4 weeks)

The user's own digital brain. Same DNA, different leaf. The hard problem is cold start.

### Weeks 10-11: Seeding

- User topology bootstrapped from conversation (real-time) + filesystem (historical)
- Filesystem as binding reality — project structure, recent files, modification patterns
- Every exchange creates nodes in the user's topology, not just the being's
- Start narrow: filesystem + conversation only. No calendar, browser, or contacts yet.

### Weeks 12-13: Surfacing and validation

- The projection runs in parallel — observes exchanges, develops opinion
- Surfacing: projection pushes to the frontend when proxy score crosses threshold
- Iterate the seeding → surfacing loop — how fast does it become useful?
- Cold start validation: new user, filesystem scrape + 30 minutes → projection surfaces something useful unprompted
- Add calendar as second binding reality if filesystem proves the pattern

### What Phase 3 proves

- Cold start solves through filesystem + conversation — no profile, no import
- The projection surfaces relevance without being asked
- Device data becomes binding realities, not data imports

---

## Phase 4 — September 3 → September 30: Personal network — distributed devices (4 weeks)

One being across many devices. Only attempt this after single-device is proven valuable.

### Weeks 14-15: Multi-device world

- Each device is a world with its own components
- User being exists across worlds — same identity, same topology
- Competence routes to the device with attention
- Relation serializes across device boundary — JSON over HTTP/WebSocket

### Weeks 16-17: Topology sync and convergence

- Weight updates propagate across devices
- Conflict resolution through the activation equation — weights merge, not timestamps
- Offline resilience: full topology copy per device, converge on reconnect
- Validation: start on laptop, walk away, pick up phone — projection continues

### What Phase 4 proves

- The network boundary is invisible — same Realize(), the pipe is just longer
- Weights handle distributed convergence without CRDTs
- The personal network is a substrate property, not a feature

---

## Phase 5 — October 1 → October 31: Real attention UI (4 weeks)

The proxy scoring from Phase 1 gets replaced with real activation scores. The runtime's weight structure is stable by now. This is where the UI goes from "works" to "feels alive."

### Weeks 18-19: Activation-driven rendering

- Swap proxy scores for real activation equation output — identity × familiarity × recency × binding fields
- Components scale by real weight — size, brightness, position, presence
- The three decay curves drive visual hierarchy — recency is loud, identity is steady, absence fades
- Animation design — surfacing, fading, reshaping feels organic

### Weeks 20-21: Attention feedback loop and polish

- User interaction feeds back into the topology — click/scroll/dwell bumps real weights
- Per-user rendering from the same snapshot — two users see different UIs
- The higher-order function: user's topology modulates what activation surfaces for them
- Polish: transitions, timing, visual weight

### What Phase 5 proves

- The UI is a sense organ, not a dashboard
- Personalization is a substrate property
- The attention feedback loop is self-reinforcing

---

## Phase 6 — November 1 → January 31: Production hardening (3 months)

### November — Permissions and trust boundaries

- Permission system from the topology — reachable graph IS access control
- User consent model: what can the projection see? What requires human approval?
- Device scrape permissions — filesystem scope, calendar opt-in
- Multi-user isolation
- Act service security audit

### December — Testing and failure modes

- Integration test suite across the full loop
- Chaos testing: kill mid-traversal, corrupt weights, drop network, exhaust budget
- Performance under load: 6 beings, 10 exchanges, continuous traversal
- Cold start regression at scale
- Memory pressure: where does the graph break?

### January — Production readiness

- Deployment pipeline
- Persistence hardened — survives crashes, upgrades, migrations
- Monitoring: topology health, inference costs, traversal depth, sync lag
- Error handling: graceful degradation
- Documentation: genome format, act service contract, component API
- First external user onboarding

---

## Timeline summary

| Phase | Window | Duration | What lands |
|---|---|---|---|
| 1 — Alpha | June 1 → July 15 | 6.5 weeks | Full composition loop: runtime + one bridge + primitive attention UI + user relevance + maze as test |
| 2 — Bridges | July 16 → August 5 | 3 weeks | Bridge pattern generalizes, adapter generation, promotion |
| 3 — Projection | August 6 → September 2 | 4 weeks | User's digital brain, cold start from filesystem + conversation |
| 4 — Personal network | September 3 → September 30 | 4 weeks | Multi-device topology, sync through weights |
| 5 — Attention UI | October 1 → October 31 | 4 weeks | Real activation scores replace proxy, polish, feedback loop |
| 6 — Production | November 1 → January 31 | 3 months | Permissions, testing, failure modes, ship-ready |

**Total: June 1 → January 31 — 8 months.**

## Key change from original roadmap

The original roadmap proved subsystems in isolation — runtime, then bridges, then projection, then network, then UI. The differentiator loop didn't land until Phase 5.

The revision brings a crude but complete loop into Phase 1. The primitive attention UI uses proxy scores (not coupled to the runtime's weight format) so it can ship while the activation equation is still in flux. When Phase 5 arrives and the weights are stable, the proxy swaps for real scores — the rendering pattern doesn't change, only the source of the number.

The headline demo by July 15 is not "beings solve a maze." It's:

> User connects an external agent. Uses the system naturally. The agent and relevant context surface automatically later because the topology learned they matter. The UI shows it.

That's the product. Everything else is depth.

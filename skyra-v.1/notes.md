# v.1 Notes

- macOS world — the machine is a world. Terminal, filesystem, network are devices inside it. The user sits behind the macOS world. The loop lives there.

## Skills

A skill is a direct route to a specific reality, bypassing default routing. It maps to how skills are already understood by every major model — the word is safe to use because it means the same thing here as it does everywhere else.

A skill is not code. It's a markdown file that describes how to call something. The being reads it and knows what to emit. It's documentation that becomes capability.

```
# skill: mike-laptop
device: terminal
call: ssh mike@laptop
```

```
# skill: fetch
device: shell
call: curl -s ~url
```

Skills live in the being's present as artifacts. A being with a skill file knows how to address that reality directly. A being without it goes through default routing.

No skill primitive in the runtime. Skills are just files — retained by the being, read into its present, used like any other knowledge. The LLM already knows what to do with them because the format matches its training data.

## Universe Relay

No dedicated relay needed. The call stack is the relay. A specialist's `Realize()` returns a string. The inner universe's `Realize()` collects it. The relation passes through every layer on the way down and the result propagates back up the same stack.

Same pattern as everything else. No new mechanism. Each `Realize()` does its work and returns. The parent catches it.

This holds for the personal network too. An HTTP call is a relation — the payload serializes across the boundary, the remote universe's `Realize()` fires, a string comes back. The caller doesn't know it left the machine. The network is just another port on the device. No special transport needed. Same interface everywhere.

## Documentation Strategy (v.1)

Describe everything at the lowest level of abstraction. Engineers, not philosophers.

- `Realize()` is the `main()` method of every reality. That's the whole runtime in one sentence.
- Internal ontology (`Reality`, `Relation`, `Impulse`) stays in the code. Docs use plain words: world, being, device, port, skill, memory, think, act, thread, trust.
- The onramp is a genome file and `go run`. Twelve lines declaring who exists and what they can reach.
- Don't explain emergence. Say: "define a being, give it an identity, put it in a world, talk to it. It remembers. It grows. It relates."
- The theory is for us. The experience is for them.

## Retained Artifacts

- **A retained understanding is not a single entity.** It is a pattern across several retained traces. The understanding doesn't live in one place — it's the coherence that emerges when multiple traces activate together. The same way a skill is the pattern across co-activated memories, an understanding is the pattern across co-activated traces. You don't store an understanding. You recognize one when the traces line up.
- **Traces → understandings → skills are the same spectrum.** A trace is what happened. An understanding is the pattern across traces — co-activation reveals coherence. A skill is an understanding that keeps proving useful under stress — keeps resolving setpoints, keeps being what the being reaches for. The boundary between them isn't a type check. It's weight. This extends the promotion gradient below where the spec currently starts.

## Ideas

- **Self-selected inference depth** — `<think-hard>` tag in think loop output bumps the next call to a heavier model (e.g. Opus). Being decides its own inference cost based on the problem. Fits the economics model — burns more budget, gets more depth. The being makes an economic decision, not the config.
- **Client-side spend enforcement** — Anthropic doesn't support per-key limits or throttling. Economics reality should enforce cumulative spend ceilings per session. Provider checks budget before every call. Being hits the wall and knows it — better than silent failure when credits run out.
- **Waiting-on-you indicator** — TUI sidebar dot changes color (e.g. green → amber) when a being has responded and is waiting on the user. Color shift over sound — rewards looking, doesn't interrupt. Makes the harness feel like a place with presence, not just a message log.
- **Unified graph retrieval** — operators, memories, skills, beings are all just nodes in the graph. Retrieving `<search>` to call it and retrieving a memory to recall it are the same operation. Everything lives in the graph. The being's present is "what did the graph return for this impulse." Recently activated nodes weight higher, stale ones drop below threshold and disappear from the present.
- **Cognitive nervous system** — see `provider-collapse.md`.
- **Persistence** — NoSQL (Mongo/BoltDB/Badger) for cell storage, reconstruct pointers on boot, re-register function maps from the genome. Circular references (bidirectional relationships) mean serialize by ID and reconstruct on load. When the topology outgrows this, build custom storage. Let the pain tell us what the storage needs to be.
- **Graph storage vs. Realize() traversal** — the graph data structure (nodes, edges, weights, adjacency) can use a standard Go graph library. But traversal is not library-standard — each node's traversal is a `Realize()` call that accumulates context and mutates the Relation. The library answers "what nodes are relevant." The runtime calls `Realize()` on each one.

## Seeds (v.2+)

- **DNA is the Reality interface, not the .skyra file.** Every reality carries the same methods — ID(), Create(), Realize(). What makes a Thread different from a Memory isn't the code — it's where it sits in the network. The `.skyra` file isn't DNA — it's placement. The bioelectric field equivalent.
- **Pattern**: seeds for the next-next version get planted during the spec work for the next version. v.05 spec work planted the unified graph. Unified graph spec work planted the boot/expression mechanism. The deeper truth tends to show up before the implementation hits the problem it solves.
- **Software as regeneration — emergent integrations.** Integrations aren't built. They emerge. The architecture needs a library of proven, battle-hardened blocks — small `Realize()` implementations that cross one boundary each. The being writes glue — a thin `Realize()` that composes proven blocks. If it degrades, stress catches it, the being writes a new `Realize()`, same blocks, fresh glue. The limb regenerates. The marketplace follows: third parties build blocks, not integrations. You don't ship integrations. You ship biology.
- **Custom primitives at any level — programmable brain wiring.** Primitives are the environment's affordances, the kernel is opaque to them, and the recursion makes every level an environment — so custom primitives at any level fall out for free. A brain's internal message-passing between regions (episodic ↔ goal ↔ attention) could be a protocol *defined* as primitives on the agent's inner environment — `innerWorld.observe(self, content)`, same kernel, same delivery — rather than hardcoded in the runtime. Two beings could run different internal architectures on the same kernel. The open-core line drawn one level below the world: people ship brain protocols the way they ship world affordances. The [Perception Invariant](design-docs/invariants/perception-invariant.md) holds at every level, so custom primitives get freedom in *what* they do and none in the *shape* of the loop — agency out, accountability back. The tell it's right: the extensibility fell out, it wasn't designed.

## Four Signals — What Differentiates a Cell

Biology uses four signal types to differentiate one genome into every cell type:

1. **Chemical** — morphogen gradients. Concentration determines fate. (~12 major pathway families)
2. **Bioelectric** — voltage gradients across membranes. Sits upstream of chemical — voltage determines which chemical signals a cell can hear. (Levin)
3. **Mechanical** — physical forces. Stiffness, pressure, stretch. Force becomes gene expression.
4. **Cell-cell contact** — direct signaling through adjacency. A cell reads its neighbors by touching them.

Four categories, varying intensities, combinations, and timing. A 4-dimensional signal space selects among ~20,000 genes. All differentiation emerges from that. DNA is the complete instruction set. Every cell has identical DNA. Signals don't write code — they select which page to read.

The activation formula might need to be re-examined through this lens — which terms are the actual signals, and which are the same signal measured differently?

## Conversation as Topology — Associative Memory

One decision: treat conversation as traversal through the same medium as everything else. Messages are Reality nodes. Threads are paths. The same activation equation governs what conversation context surfaces. No separate data structure. No special rules.

Messages live on the entities they describe, not on a timeline. The thread gives sequence. The entities give placement. The same message is reachable from multiple directions depending on which internal Reality the traversal enters through.

Recency is traversal count, not wall clock. The system doesn't experience time — it experiences relations passing through it. Each being tracks its own count. Different beings age at different rates. This is relativistic — proper time, Einstein's term for time experienced by the thing itself.

Training and inference are the same pass. Traversal is the forward pass. Weight updates on the return path are the training step. Each conversation is a training run.

The whole topology is an associative memory. You don't retrieve by address. You retrieve by content — what resonates with what the Relation is carrying right now. Old messages don't get deleted. Their edges decay. Compression is natural. Truth is derived not stored.

## Places as Topology Regions

Places are Realities with localized affordances. Beings don't move to places — they subscribe. Subscriptions determine what signals reach a being, what capabilities are visible, what acts are available. A place is just a Reality. A subscription is just an edge. Two creation paths: genome-seeded and density-emerged.

### Two Seeded Places

**Town Hall** — the registry. Every being registers here. Discovery happens through traversal against the registry. Solves "who's here and what can they do." PFC at the world level.

**Factory** — where work happens. Deterministic systems, build pipelines, adapters, act service blocks register here. Solves "what can we build with."

Everything else emerges from density. Density-emerged places are the promotion gradient applied to regions rather than individual nodes.

### Subscriptions Have Cost

More subscriptions increase awareness but increase noise and cognitive load. Forces natural specialization. Attention economics as a runtime constraint.

## Deterministic Systems as First-Class Citizens

LangChain flows, LangGraph workflows, CI/CD, ETL, scripts, adapters — all Realities on the plane. A deterministic system is a Reality whose Realize() is a fixed pipeline. No inference call. No weight update. The topology doesn't care whether a Reality thinks or executes — same interface, same traversal. The plane becomes the substrate existing tooling lives on.

## External Agents as Native Citizens

Claude Code, AutoGen, CrewAI, any CLI agent — participates through adapters. Two lines in the genome, a Reality whose invariant is a pipe. The adapter handles the boundary. The topology treats it like any other being.

```
being ~name claude ~type agent ~entrypoints claude ~relationships michael,skyra
being ~name autogen-team ~type agent ~entrypoints pipe:autogen-cli ~relationships skyra,builder
```

## Four-Provider Cognitive System

Descent and ascent use different collapse physics:

| Phase | Role | Provider | Why |
|---|---|---|---|
| Descent | Breadth exploration | ChatGPT | Wide associative recall, good at surface coverage |
| Descent | Deep seam-finding | DeepSeek | Recursive depth, follows threads into structure |
| Ascent | Dense compression | Claude | Heavy synthesis, preserves nuance under pressure |
| Ascent | Sparse compression | Grok | Fast, lightweight, good at thin signal |

Provider fires based on activation and density, not a router. The being develops cognitive style through which providers keep proving useful where — it emerges from what worked.

## Rest Cycle

Continuous traversal without recovery degrades weight structure. The topology needs silence between waves to consolidate. Not a feature — a requirement of excitable media.

Without rest: fibrillation, seizure, spiral turbulence. Mechanically: a minimum inter-traversal interval per being, scaling with traversal intensity.

## Cognitive Architecture References

ACT-R — Memory activation equations (combining recency, frequency, and context) are mathematically validated against human behavioral data. Spreading activation maps directly to entity graph traversal.

SOAR — Chunking mechanism parallels skill maturation. Impasse resolution research is relevant to think budget exhaustion.

OpenCog Hyperon — Most theoretically ambitious (self-modifying metagraph, attention economy, reflexive cognition). But it's a research substrate, not a deployable runtime.

## Locked Decisions (from v.1 plan, still standing)

- **Deterministic first.** `argmax` for collapse. Stochastic/temperature later.
- **Actors and agents.** Same interface. Distinction is emergent. Actor = no provider calls (purely reactive). Agent = calls providers (can think). The line is whether inference fires, not a type declaration.
- **doesNotUnderstand as growth.** Reaching for something unknown seeds a new cell at minimum weight. The being grows at the edges.
- **Max depth as safety rail.** Not the mechanism. Signal attenuation and visit-once are the mechanism. Max depth catches runaway cases the mechanism misses.

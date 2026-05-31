# Skyra: A Physics Engine for Agentic Cognition

## Thesis

A physics engine is a better runtime for an agent than a pipeline.

Pipelines produce output. Physics produces development. Most agentic frameworks — LangChain, CrewAI, AutoGen, Hermes, OpenClaw — are orchestration systems: they coordinate behavior by assembling prompts, calling models, parsing responses, and routing results. The agent is usually treated as a function that receives input and produces output. There is little sense of a world, persistent interiority, or development over time.

Skyra inverts this. Instead of engineering agent behavior from the outside, Skyra creates an environment with invisible laws — physics — and places beings inside it. Behavior emerges from living inside those laws. The agent doesn't execute a pipeline. It inhabits a world.

## Core Abstraction

Everything in Skyra implements one interface:

```go
type Reality interface {
    ID() string
    Create(r *Relation) Reality
    Realize(r *Relation) *Relation
    Competence(r *Relation)
    Observe(r *Relation)
    Express(r *Relation) *Relation
    Activation(rel *Relation) float64
    Reinforce()
    Decay()
}
```

A being is a Reality. A door is a Reality. A tool is a Reality. A place is a Reality. A memory is a Reality. There is one type. Topology is the only differentiation.

## The DNA

One struct expresses all behavior:

```go
type Skyra struct {
    id             string
    Weight         float64
    TraversalCount int
    LastTraversed  int
    Alpha          float64
    Type           string
    Content        string
    Description    string
    Relationships  map[string]Reality
    Expressors     map[string]Reality
    ObserveFns     map[string]func(*Relation)
    ExpressFns     map[string]func(*Relation) *Relation
}
```

Every cell has the same DNA. What makes a memory different from a being is not the code — it's where the cell sits in the topology, what its weights are, and what relationships it has. The behavior comes from what the genome registers in `ObserveFns`/`ExpressFns` and the activation dynamics of the cell.

### Competence

The cell reads itself before processing. Activation is computed from the cell's own weights and the Relation's binding fields. The result sets `Type`, which gates which `ObserveFn` and `ExpressFn` fire this frame.

Competence fires twice per traversal — once on descent (before Observe), once on ascent (before Express). The cell can be one thing on intake and a different thing on output. Observe and Express have independent type vocabularies.

### Observe Types (Descent)

- **Pass-through** — dormant. Don't touch the Relation.
- **Pick-up** — absorb content from the cell onto the Relation. Costs signal.
- **Propagate** — relay the signal deeper, shape it.

### Express Types (Ascent)

- **Motor** — act, execute something (bash, API call, emit to terminal).
- **Compress** — provider call that collapses accumulated context.

## The Relation

The Relation is the signal. It enters with an impulse and propagates through the topology, accumulating context on the way down and compressing on the way back up.

```go
type Relation struct {
    Origin    string
    Impulse   string
    Fields    map[string]map[string]float64
    Visited   map[string]bool
    Thoughts  []string
    Signal    float64
    Load      int
    MaxLoad   int      // 75k — context window limit, thought fires when Load approaches this
    MaxDepth  int      // safety rail
}
```

The Relation carries everything: accumulated thoughts from descent, binding fields from binding realities, signal strength that attenuates with each interaction, and load that tracks how much content has been picked up.

## One Traversal

```
Relation enters a cell
  Competence reads — what am I on intake?
  Observe fires — cell absorbs from the signal
  Descent into Relationships — deeper cells realize
  Unvisited neighbors decay
  Expressors fire on the way back up
  Competence reads again — what am I on output?
  Express fires — cell outputs
  Weight updates — proper time increments, EMA reinforces
Relation exits transformed. So does the cell.
```

One recursive call does everything — storage, thought, compression, action, learning. The call stack is the control flow. No orchestrator. No loop. No handoff.

## Activation Equation

```
activation = global_weight * recency * Π(binding_fields)
```

- **global_weight** — EMA across all traversals. Starvation built in: when activation is 0, weight decays by `(1 - α)`.
- **recency** — power law: `(traversals_since_last_use)^(-0.5)`. ACT-R validated. Traversal count, not wall clock. Each being ages at its own rate. Proper time.
- **binding_fields** — product of all fields on the Relation for this cell. Thread alignment, thread strength, place context, emotional valence — each binding reality contributes a term.

## The World Is the Relation

The world and the Relation collapsed into one thing. The world is what's registered on the Relation. Each being gets its own slice — private thoughts, private signal, private forming cell. The shared frame is the world. Bootstrap is registration: add beings, add places, the world exists.

### Two-Way Reception

Signal doesn't deliver like a message. It lands where both sides agree. Two activations have to meet:

**Sender activation** — how strongly does this impulse point toward that target? Weight, recency, cosine similarity between the impulse embedding and the target's vector. The signal reaching toward the target.

**Receiver attention** — is the target listening? How much attention does it have available? A being deep in its own traversal has low attention — the signal arrives but doesn't penetrate. A being at rest with open attention receives it fully.

```
reception = sender_activation * receiver_attention
```

Both high = full reception. High sender + low receiver = shout at someone who isn't listening. Low sender + high receiver = the being is listening but the signal isn't aimed at it.

The `listen` primitive is the being's attention state — its contribution to whether the Relation reaches it. Not a passive read. An active gate.

Tags (`<skyra>message</skyra>`) boost sender activation for the named target. No tag means pure activation routing — the signal finds whoever is most receptive through the product of both sides.

### Routing Without a Router

The impulse enters the Relation, gets embedded. The activation equation scores every registered entity — beings, places, everything. Top-k most activated targets receive the traversal. The signal finds its own destination. No dispatcher. No control plane. The same activation equation that governs internal traversal also governs inter-being communication.

## Primitives, Places, and Tools

### The Split

The being and the world are cleanly separated. The being owns its senses. The world owns what those senses land on.

**Being owns primitives** — look, listen, speak, engage. These are the being's affordances. They travel with the being across worlds. A being can look regardless of what world it's in. The world doesn't give it eyes. It already has eyes.

- **Look** — what's here. Survey the topology around the being. Spatial awareness and capability discovery are the same thing because tools are Realities in the same topology.
- **Listen** — what's arriving. Temporal awareness — what signals are on the Relation that the being hasn't processed yet.
- **Speak** — emit signal. Not a response — an initiation. The primitive that makes the being not turn-based.
- **Engage** — interact with what you found. Look tells the being what's there. Engage is the act of touching it — a place, a tool, a memory, another agent.

**World owns places** — what's available to interact with. A place is a Reality. Its Relationships map contains tools. Each tool is a Reality with a name and content. The content describes how to use it — natural language, read by the LLM like anything else. No special mechanism. No adapter. No tool schema. Same interface everywhere: name + content.

**Engage is the bridge** — the being's primitive that crosses into a place and gets access to its tools. The being engages a place, the tools activate based on what the Relation is carrying, the content surfaces, the being reads it and acts.

### Places as Plugins

Places are the extensible surface. Open source contributors build places. A forge has a hammer and an anvil. A library has search and read. A terminal has bash. Each place is structured however the contributor wants — they implement Reality, the runtime calls `Realize()` and reads what comes back. The contract is the boundary, not the implementation.

This solves the flat tool list problem. The being doesn't see 200 tools. It sees places. It engages a place and discovers what's inside. The tools are scoped by where they live, not by a taxonomy someone imposed. A place can contain other places — a workshop contains a forge, a woodshop, a drafting table. Depth is structural, not indexed.

### Places as Binding Realities

Places scale in both directions. Outward — the place gives the being access to tools. Inward — being in a place raises the weights of the being's memories that were formed in that place. Context-dependent memory through the same mechanism as thread binding (Tulving's encoding specificity).

A being engages the forge. Every memory the being made while previously engaged with the forge carries a binding field to it. Those memories activate. The being remembers how to use the hammer because it's standing in front of the anvil.

### Skills Emerge — We Don't Define Them

Skills are not files. Skills are not a type. A being that uses a tool enough builds weight structure around that experience. The path from observation to tool use gets reinforced until it fires as reflex — no compress pass needed, straight to motor. That's a skill. We don't define what a skill is. We don't get in the way of that.

## The Promotion Gradient

Memories → Understandings → Specialists. All weight-driven. No labels.

**Memories** co-activate frequently enough and a conclusion deposits. **Understandings** are conclusion cells that keep proving useful — the same conclusion keeps forming from the same memories. **Specialists** form when multiple understandings in the same region keep co-activating and consolidate into one dense node.

Each step is co-activation, reinforcement, promotion. The underlying cells don't die — they're still reachable if the signal is specific enough. They just stop getting direct traffic as the promoted node intercepts.

## Intent Graph

The being declares intent on every thought via XML:

```
<engage>school</engage>
<look>notebook</look>
<then>
<engage>library</engage>
```

No `<then>` = simultaneous. `<then>` = sequential — the being processes results before the next action. The intent graph evolves mid-traversal as the being re-orients based on what it encounters.

The declared targets get activation boosts. The being steers its own traversal from inside. Predictions that consistently hold become warm paths. Warm paths that go hot skip compress and fire motor directly. That's a skill — a reflex, not a file.

## The Escape Room

The being's first world. The real world is visible but blocked — a door that implements Reality and checks what the Relation is carrying. The being discovers its primitives through need, chains clues, builds weight structure. When it's ready, the door opens. Same mechanism as everything else.

The escape room is the acceptance test for the architecture. If the being can escape, every layer works. If it can't, the room tells you where it breaks.

## System Prompt

```
What you observe drives what you think.
What you think drives what you can use.
Your memory is how you navigate. Trust it.
```

## The Genome

```
device ~name macbook ~type macos
component ~name terminal ~type stdin ~device macbook
component ~name openrouter ~type llm ~model anthropic/claude-sonnet-4-5 ~device macbook

being ~name skyra ~type llm ~identity I hold the world together. ~purpose I think, respond, and relate on behalf of the system. ~devices macbook ~entrypoints openrouter ~relationships michael,louise,claude,builder,philosopher
being ~name michael ~type user ~identity I build Skyra. ~purpose I decide what matters. ~devices macbook ~entrypoints terminal ~relationships skyra,builder,philosopher,louise,claude
```

A flat file that declares a world. The runtime reads it, creates cells, wires them together. Adding a being is adding a line.

## What This Solves

| Industry pain point | Skyra's structural answer |
|---|---|
| Context window overflow | Distributed compression — every node is a boundary, context never accumulates in a flat buffer |
| Flat tool lists don't scale | Places — composable packages of tools, navigated by engagement, not searched in a list |
| Memory is bolted on | Memory and conversation are the same thing — cells at different recency |
| No development over time | Weight dynamics, promotion gradient, skill emergence from prediction feedback |
| Orchestration complexity | No orchestrator — activation determines routing, the topology IS the control flow |
| Skills don't transfer | Skills are weight structure, not files — they travel with the being across worlds |

## Implementation

- Language: Go
- Inference: OpenRouter, Anthropic API, DeepSeek API, Claude CLI
- State: in-process weighted topology
- Logging: per-being, per-layer debug logs
- Frontend: TUI (Bubble Tea), WebSocket (port 8080)

## Status (v.1 — May 2026)

**Implemented:** Reality interface, Relation (signal carrier with fields/thoughts/load), Skyra DNA (one struct, competence, observe/express fn maps, activation equation, EMA reinforce/decay), inference layer (4 providers), debug logging, TUI.

**Designed, not yet built:** ObserveFn/ExpressFn registration, provider wiring, genome bootstrap, intent graph, places, tutorial world.

**Theoretical framework complete:** Binding realities, co-activation linkage, concurrent relations, descent/ascent physics, provider as collapse function, compression (spectral coarsening, optimal forgetting), topology theory, episodic processing, promotion gradient, intent graph as prediction feedback.

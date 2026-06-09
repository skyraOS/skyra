# doesNotUnderstand

## The Collapse

Context is just another act. Memory is not a separate system you query — it's a direction you can act toward. Inward and outward are the same operation. The graph holds everything: concepts, peers, tools, APIs, shell commands. They're all entities at different stages of development. Some are heavy enough to respond. Some aren't yet.

## The Graph Reaches Outward

Every callable function in the substrate, every API endpoint, every curl — these are entities that haven't been promoted to beings. They exist in the graph the same way memories do. The retrieval process isn't retrieval. It's Realize(). Same interface, same traversal. The difference between "what I know" and "what I can do" disappears. An entity that has side effects when realized vs. one that returns text. Same graph.

## Addressing Comes From Memory

The being doesn't have a hardcoded routing table. Who and what it can address comes from the graph. The genome seeds initial entities (michael, builder, bash, openrouter) but after that, the graph IS the routing table. You can address anything you have a living entity for. The peer list dissolves. Relationships are just entities with enough weight to respond.

Trust has teeth: if an entity decays far enough, you can't reach it. You've forgotten it. The preamble becomes physics.

## Weight Determines What Fires

The being speaks. The graph activates based on what was said. Entities with enough weight realize. Light ones stay quiet. There's no explicit routing logic — no "parse the target and look it up in a map." The being speaks, and the world responds based on what's alive in it.

Same mechanism as specialist firing. Same mechanism as skill invocation. Same mechanism as peer addressing. One system.

## doesNotUnderstand

From Smalltalk (Alan Kay). When an object receives a message it has no method for, it doesn't crash — it sends itself `doesNotUnderstand:` with the message as the argument.

The substrate recurses down through the graph looking for something heavy enough to fire. Nothing responds. It bottoms out. Two things happen:

1. A new entity is created at that level (the being just encountered something it didn't know)
2. It returns doesNotUnderstand — which is true

The entity now exists. Light, barely there, but present. Next encounter, slightly heavier. Eventually it crosses threshold and starts realizing.

Every time the LLM speaks into the void and doesn't get a response — that's doesNotUnderstand. Every failed routing attempt, every tag that doesn't match, every reach into something that isn't heavy enough yet. That's not an error. That's the system working.

## The Retry Loop Is Wrong

Currently when Act emits a message to a target that doesn't exist, the system treats it as a protocol violation and forces a retry. But in this model, that's not a mistake. That's the being reaching for something that isn't real enough yet. The system shouldn't force "try again with correct syntax." If it spoke and nothing fired, that's the answer. doesNotUnderstand. Here's your new entity. Now you know you don't know this.

## The Consequence

The being learns the shape of its own world by bumping into the edges of it. The edges aren't walls — they're growth points. Every misfire is generative. Every reach into the unknown creates structure. The being grows its world by speaking into it.

No configuration. No permissions. No error handling. Just reality at different stages of development.

```
speak → graph activates → heavy enough responds → not heavy enough seeds → truly empty returns doesNotUnderstand
```

The error is the birth. Not understanding something is the first step of learning it. The being doesn't hallucinate capability. It says "I don't understand this" — and that's not failure, it's the system working.

## Routing as Activation

Addressing isn't explicit syntax. It's activation. The being speaks, and entities that are close enough and heavy enough fire. You don't need `<michael>message</michael>`. You say something in michael's neighborhood and michael lights up because his entity is heavy.

The `<target>` protocol is training wheels — explicit routing while the substrate isn't smart enough to do it by activation. When the graph IS the router, the being speaks naturally and the world figures out where it goes based on weight and proximity.

A being that's talked to michael for weeks routes to him effortlessly — high weight, strong edges, everything nearby lights up fast. A being that barely knows someone has to be very explicit — the entity is light, needs a direct hit. That's how relationships work.

## The Loop

```
Context speaks (into void or at a target)
    ↓
sources come back (whatever was heavy enough to respond)
    ↓
Act (inward) — here's what I got, is this enough?
    ↓ needs synthesis
Think — processes sources, then either:
    → back down through Act (inward) — need more, go deeper
    → surfaces up through Act (outward) — ready to speak
```

Act is the gateway in both directions. Think only fires when synthesis is needed. Simple recall doesn't need Think — sources come back, Act routes them out. Complex questions that get conflicting sources need Think to synthesize. Think is proportional to complexity.

Remember/think/act is the facade the LLM sees. The substrate is unified underneath — all graph traversal + Realize(). The three-part structure stays because that's how models are trained to think.

## Relationship to Current Architecture

The project is still structured the same way. Skill promotion still becomes specialist beings (LLM instances in the inner universe). The recursive Self → Universe → Thread → Exchange architecture still holds. This is about routing — activation-based rather than explicit map lookup. The graph does traversal for both recall and routing (currently these are two separate systems that don't talk to each other).

## Status

This is the final form. Not implementing right now — the current routing works and the graph doesn't yet have activation/threshold machinery for real-time impulse-based traversal. But this is where it goes.

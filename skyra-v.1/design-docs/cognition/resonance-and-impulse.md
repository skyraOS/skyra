# Resonance & Impulse Reception

## The Problem

The being receives signals from the world. How do those signals land? What determines whether an impulse actually penetrates the Membrane or bounces off?

Two sides to every signal. The sender determines force. The receiver determines resonance. Neither side owns the full equation alone.

## Force Is Energy

Force and energy are the same mechanic. The Relation's Realize traverses inward through the being's brain — energy depletes as it descends the tree. Propagation traverses outward through the world — force depletes as it spreads through the graph. Same depletion curve. Same threshold cutoff. Same "when it runs out you stop."

The Membrane is the pivot point. Inward is cognition (Realize). Outward is expression (Propagate). Same wave, same mechanics, same kernel. The tree just points the other direction.

### Propagation Is Outward Realize

Distance isn't meters — it's activation cost through the graph. Something physically close but disconnected takes more energy to reach than something far but directly connected. The topology IS the physics. Energy flows faster down strong connections. A being you talk to every day is one hop. A stranger across the room is more hops even if they're physically closer.

Each node the energy hits is whatever's there — empty space, an object, another being. The graph is not a geometric radius. It makes a circle because energy depletes uniformly, but the shape of that circle is determined by connection strength, not spatial distance.

Realize and Propagate may not even need to be separate functions. Same DFS. Same energy budget. Same survivor logic. Different root direction — one goes inward from the Membrane, the other goes outward.

### Primitives Return Propagation

The primitive executes and its return value is the propagation call. Motor decides which primitives fire. The primitive does the work. The return value hands the result to the world.

```
func speak(content) {
    signal := process(content)
    return world.Propagate(signal, force)
}
```

Every primitive ends the same way. Whatever it computed, hand it to the world with a force. The primitive doesn't know who receives it. It just produces a result and lets the world carry it. Motor is the decision. The primitive is the action. Propagate is the consequence.

DNA injects the propagation function — defining the physics of how force decays, what the threshold is. The world calls it. No orchestration. No routing table. No target ID. Just physics.

## Impulse Force (Sender Side)

Primitive convergence determines force. Multiple primitives aimed at the same region of the graph compound.

**Primitives are surfaces with direction.** Speak, look, gesture — each is an expressive primitive. The more primitives converging on the same direction, the stronger the force budget for propagation.

- One primitive = base force
- Two primitives converging = stronger (speaking while looking at someone)
- Three = even more forceful (speaking + looking + gesturing)

The being doesn't target anyone. It expresses with a force into the outward graph. What the force reaches is determined by the topology and the energy budget. The receiving being doesn't need to know how many primitives were used — it feels the intensity.

**Some primitives are bidirectional.** Look increases both expressive force (I'm directing attention at you) and receptive resonance (I'm open to signals from you). That's why eye contact is powerful — double-sided convergence.

## Resonance (Receiver Side)

Observe computes resonance. Two independent components that stack:

### 1. Primitive Alignment

Receptive primitives (listen, look) have direction. If listen is active and aimed at the source, resonance is high. If the being isn't listening — occupied, attending elsewhere — resonance drops.

- Listen aimed at speaker + signal from that speaker = high primitive resonance
- Listen aimed elsewhere = low primitive resonance, signal attenuated before it hits prediction

### 2. Goal Relevance (Semantic Resonance)

High-activation goals boost resonance even without primitive alignment. This is the cocktail party effect — you hear your name across a noisy room because the content matches something you're tracking, not because you were listening.

**Two gates, in order:**

1. **String match** — cheap, fast, binary. Do any tokens from high-activation goals appear in the signal? If no, skip embedding. If yes, proceed to cosine.
2. **Cosine similarity** — confirm semantic relationship. Threshold: **0.55**. Below that is noise. Above that boosts resonance proportional to score.

String match and cosine are separate signals that impact each other. String match alone catches exact lexical hits. Cosine alone catches semantic relationship without shared words (e.g., "restaurant" vs "great place"). Either can contribute independently. Both together is strongest confirmation.

### Three Granularity Levels

Incoming signals are scored at multiple granularities. Best score wins.

**Level 1 — Full signal.** The entire incoming message as one embedding. Diluted by filler and noise. Weakest signal but captures overall topic.

**Level 2 — Clause split.** Split on punctuation and conjunctions (comma, period, semicolon, "and", "but", "or"). Each clause embedded independently. Catches the relevant phrase buried in a longer message.

**Level 3 — Keywords.** Individual content words (stopwords removed). Strongest single-concept signal. One word is enough if it's the right word.

Example — goal: "find a restaurant for dinner", signal: "I'm ready, let's go for a drive, maybe grab some dinner on the way"

```
0.60  |  full signal (diluted)
0.74  |  clause: "maybe grab some dinner on the way"
0.77  |  keyword: "dinner"
```

The being half-listening in a noisy room catches "dinner" and the goal lights up. The keyword is the strongest signal because it has no noise around it.

Resonance = max(score across all granularity levels).

### Combined Resonance

```
impulse_strength = force × resonance
resonance = primitive_alignment + goal_boost
goal_boost = max(string_match_boost, cosine_boost) if above threshold else 0
cosine_boost = max(score across full, clauses, keywords) if > 0.55
```

## Surroundings vs Attention

**Surroundings** — everything that propagated to this being's Slice. The raw impulses that reached it based on force and graph distance. World-facing.

**Attention** — what survived resonance and the prediction gate. Being-facing. What Observe produces. What binding realities see.

```
Express → Propagate (outward energy) → Surroundings (what reached the Slice)
  → Resonance (primitive alignment + goal relevance)
    → Prediction Gate (efference/exafference comparison)
      → Attention (filtered set) → Binding Realities
```

Propagation determines what arrives. Resonance determines what penetrates. Prediction determines what gets encoded. Three filters, same energy mechanic at the first stage.

## Primitives

Primitives are permanent equipment on the Membrane. They're always there. But their utility is contextual — what you can do with your primitives is bound by your surroundings.

- The primitive is the capability
- The surrounding is the affordance
- Attention is the intersection — which capabilities have something to act on right now, after prediction filters it

A speak primitive with no one in the room is still there. It just produces nothing useful.

## The Full Loop

Expression and reception are the same mechanic in opposite directions. One being's Express is another being's Observe. The Relation carries both.

```
Motor decides → Primitive executes → return world.Propagate(signal, force)
  → Force depletes through outward graph (same as Realize energy)
    → Reaches Slices within energy budget (surroundings)
      → Receiving Membrane runs resonance (primitive alignment + goal relevance)
        → Prediction gate (efference/exafference comparison)
          → Attention (filtered set)
            → Binding Realities (episodic, motivation, emotion, action)
              → Episodic Encoding (frames, indexes)
                → Motor decides → ...
```

The cycle. Express outward, propagate, receive, resonate, filter, encode, decide, express. Same energy mechanic at the core of both traversal and propagation. The Membrane is where inward becomes outward.

## Open Questions

- Can Realize and Propagate share the same function with a direction parameter, or do the topologies differ enough to warrant separate implementations?
- How does primitive convergence compound force? Additive? Multiplicative?
- Should resonance decay over the course of an episode or stay constant per signal?
- Embedding model selection — nomic-embed-text works but is there a better fit for short phrase similarity?
- How do we handle multimodal primitives? Vision delivers structure, not text — cosine on what?
- What does the outward graph topology actually look like? Is it the same Slice map or a separate structure?
- How does the being's identity activation relate to resonance on hearing its own name?

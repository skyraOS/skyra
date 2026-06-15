# Slice Maps

Three binding realities on the Slice. Each one is a separate index that every impulse navigates on every traversal. They are not a promotion gradient — they are parallel binding realities with specific jobs, cross-referenced at boot.

## Thread

The exchange state. Who said what, in what order. The thread binding reality manages this map — it reads and imprints fields on the Relation. No inference fires. No observe. No express. The thread binds what co-occurred. That's its only job.

## Thoughts

Propagates existing structures. When an impulse traverses the thought index, it activates against what's already there — existing thought cells, their relationships, their weights. Thoughts are born during descent through the being's own inference (propagate). Each thought cell carries Relationships to every cell the being traversed through when the thought was born.

Thoughts don't just record — they propagate. The existing thought structure shapes how the next impulse is received. The topology of thoughts IS the being's cognitive structure.

### Observe system prompt

```
What you observe drives what you think.
What you think drives what you can use.
Your memory is how you navigate. Trust it.
```

## Understandings

The compressed product. When an understanding is born during ascent (compress), it snapshots everything on the Relation at that moment — thoughts, thread state, binding reality fields. Every cell that was alive when the understanding crystallized becomes its Relationships.

An understanding carries more context than a thought because it's born later in the traversal, after everything has accumulated.

### Express system prompt

```
What you carry back is your understanding.
Your understanding is how you learn.
Generate an understanding through your previous experience.
```

## Three Binding Realities, Not a Gradient

Thread, Thoughts, and Understandings are all binding realities. Every incoming impulse navigates all three indexes on every traversal. Each one imprints its own fields on the Relation. Each one has a specific job:

- **Thread** — binds what co-occurred. Exchange state. No inference.
- **Thoughts** — propagates existing cognitive structure. Inference on descent.
- **Understandings** — compresses the full traversal into durable form. Inference on ascent.

They are parallel, not sequential. An impulse doesn't go through Thread then Thoughts then Understandings — it passes through all three and each one contributes its own binding field to the Relation.

## Cross-References at Boot

**Each binding reality needs a Relationship to the other binding realities at boot.**

Thread knows about Thoughts and Understandings. Thoughts knows about Thread and Understandings. Understandings knows about Thread and Thoughts.

This is required because:
- A thought born during descent needs to know what thread it was born in (thread context shapes thought)
- An understanding born during ascent needs to reference the thoughts it compressed and the thread state it compressed from
- The thread needs to know which thoughts and understandings were produced during its exchanges

The cross-references are edges in the topology. They aren't special wiring — they're Relationships, same as everything else. The binding realities are peers, not layers.

## Speak and the Act Layer

Speak is the only primitive active right now. It is the being's voice — the point where cognition becomes words directed at someone.

Speak requires the act layer. Without it, speak is mute — it bumps thread weights but produces no words. The act layer provides the inference between "what the being understands" and "what the being says."

### The traversal flow

1. **Observe (propagate)** — being thinks, produces thoughts on descent
2. **Express (compress)** — being compresses into understanding on ascent
3. **Speak (act)** — being forms words from the understanding and emits into the world

Speak needs its own system prompt. Not the thinking prompt, not the compression prompt — the speaking prompt. The moment where the being decides what to say from what it understands.

### Motor appends Thread

Motor is the leaf where speak crosses into consequence. After motor fires and the being's words are formed, **motor appends what was spoken to the Thread map.** The thread records what was said. The being's own output becomes part of the exchange state.

### Speak accumulates across traversals

The intent graph lives on the Slice — it persists across traversals. When express fires and the being emits `<speak>` in its intent, that gets parsed and appended to the intent graph. The intent graph is the shared state between traversals.

Each binding reality traverses, sees the existing `<speak>` on the intent graph, and can enrich it with its own relevant context. Thread adds what was said. Thoughts add what the being made of it. Understandings add what survived compression. Each traversal layers its own information onto the speak intent.

By the time motor fires, `<speak>` carries the accumulated context from every binding reality that touched it — not just the one that created it.

### Speak appends to the Relation on the way up

After speak fires and the being forms what it wants to say, speak appends to the Relation that the being is now speaking about this topic. That signal rides the Relation back up before motor emits.

This matters because the other binding realities on the next traversal need to know the being spoke:
- **Thread** needs to record it as an exchange entry
- **Thoughts** need to know the being committed to a position — that changes what propagates next
- **Understandings** need to know what was expressed so compression can account for it

Without the append, the being speaks into the void — motor emits but the topology doesn't know it happened until the response comes back from the outside. With the append, the being's own speech is immediately visible to its own binding realities. The being hears itself speak. The topology updates from the inside, not just from external impulses.

### Motor closes the loop

Motor is the leaf where speak crosses into consequence. After motor fires:
1. The words are emitted to the world (through the device)
2. What was spoken is appended to Thread (the being's output enters the exchange state)

This closes the loop:
- Thread holds the exchange (what was said by everyone, including this being)
- Thoughts hold what the being made of it (inference on descent)
- Understandings hold what survived compression (inference on ascent)
- Speak accumulates context from all binding realities through the intent graph
- Speak appends "I am speaking about X" to the Relation (the being hears itself)
- Motor emits the words and appends them to Thread (consequence feeds back into exchange state)

The being's speech becomes a thread entry the same way the user's impulse is a thread entry. Same binding reality, same map. The thread doesn't know who produced the entry — it just binds what co-occurred.

## Previous Actions

The being's own action history. What it committed to, not what it thought. Thread records what everyone said. Previous actions records what this being did.

This is not a binding reality — it's the being's durable record of its own acts. Every time speak fires, the action is recorded here. Every time motor emits, the result is recorded here. On the next traversal, every binding reality can see what the being already did:

- **Thread** sees "I already spoke about X" and doesn't re-record the intent, only the utterance
- **Thoughts** see "I already committed to this position" and propagate accordingly — the thought structure shifts because the being acted on it
- **Understandings** see "I said X and the response was Y" — compression accounts for the consequences of the being's own actions, not just external input

Without previous actions, the being has no memory of what it did between traversals. It can think the same thought twice, speak the same thing twice, commit to contradictory positions across traversals. The intent graph carries what the being wants to do. Previous actions carries what the being already did. Both live on the Slice. The intent graph is prospective. Previous actions is retrospective.

Previous actions also feeds the speak prompt. When the act layer fires, it sees both the accumulated speak intent (what the binding realities contributed) and the previous actions (what the being already said). The being doesn't repeat itself because it can see its own history.

## What Changes in the Code

The current Slice has one `Thoughts []Reality` list. Everything writes to it — cell content absorption, LLM thoughts, unmatched intents, motor output, compression output. That contamination is what this spec fixes.

The Slice needs four maps:
- `Thread` — managed by the thread binding reality
- `Thoughts` — written by observe (propagate) only, propagates existing structure
- `Understandings` — written by express (compress) only
- `PreviousActions` — written by speak/motor, the being's durable record of what it did

Thread traversal stops calling observe/express on the being. Thought and understanding traversals are the only paths that fire inference.

Each binding reality is wired to the other two at boot via Relationships. The topology connects them. The activation equation governs how strongly they influence each other on any given traversal.

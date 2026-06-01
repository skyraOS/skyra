# Concurrent Relations — The Sync Problem

When an agent sends two messages back to back, that's two Relations entering the topology independently. Two traversals, two visited maps, two thought streams. Full isolation. But they're the same conversation — maybe the same thought split across two sends. PFC shouldn't fire twice. It should integrate.

## The Topology Knows What's In Flight

The topology tracks how many Relations are currently traversing. The entry point holds a reference to every in-flight Relation — same pointer, same memory. The buffer can read the in-flight Relation's state: depth, accumulated thoughts, binding fields, hit counts, whether it's on descent or ascent.

**Zero in flight** — Relation enters immediately.

**One already in flight from the same origin on the same thread** — buffer. When the in-flight one returns to the central observer, check the buffer. Something waiting — merge the compressed feeds, one inference call. Nothing waiting — fire immediately.

**Degradation** — three messages back to back chain naturally. First enters, second buffers, first finishes, second enters with the benefit of the first's weight updates, third buffers behind the second. They don't merge into one giant Relation. They chain. The being processes them in natural rhythm.

## Expiry — Still Thinking

If the in-flight Relation is taking too long, an expiry fires. The central observer gets a status update: "still processing the first message from this actor, second message waiting." That context enters the PFC call. The being knows it's behind. It can acknowledge both exist, say "hold on," or prioritize. The expiry turns a sync problem into context.

The being's response to the expiry is itself a Relation back to the user — "hold on, reading your second message." That's the integration layer knowing its own state. Self-awareness falling out of the topology tracking what's in flight.

## PFC Holds the Reference

The central observer holds a pointer to every in-flight Relation. It can read what the traversal has accumulated so far — thoughts, binding fields, hit counts, trace. The in-flight Relation is not a black box. PFC can see into it.

This enables attention — the central observer reasoning about its own in-progress work when new information arrives.

## The Kill

If a new message changes the context enough, PFC can cancel an in-flight traversal. Set a flag on the Relation. Each node checks it before descending further:

```
if r.Cancelled { return }
```

The traversal unwinds naturally through the call stack. Weight updates from the partial traversal still stand — the topology learned something even from the aborted descent. But no inference call fires from the cancelled path.

## The Spectrum

Four responses when a second message arrives while the first is still in flight:

1. **Let it finish, merge** — both messages matter equally. First Relation completes, compressed feeds merge with the second message, one PFC inference call integrates both.

2. **Let it finish, give PFC the second as context** — the being knows what's coming. First Relation completes normally but PFC's inference call includes the second message as additional context. The being responds to both in one output.

3. **Kill it, start fresh with both merged** — the first message was superseded or the second changes everything. Cancel the in-flight traversal, merge both impulses into one Relation, traverse from scratch. Partial weight updates from the aborted traversal survive.

4. **Kill it, process only the second** — the first was noise. Cancel and discard. The second message is the only one that matters. Partial weight updates still stand.

PFC decides which response based on what it can read from the in-flight Relation's accumulated state and the new impulse. One inference call to the central observer: "here's what the first traversal found so far, here's the new message, what do you want to do." That's metacognition — the being reasoning about its own processing.

## What This Solves

- No duplicate inference calls for split messages
- No awkward sequential responses to what was one thought
- Attention as a structural mechanism, not a prompt hack
- Self-awareness of processing state falls out of the topology
- Graceful degradation under rapid input
- The being can reason about its own in-progress traversals

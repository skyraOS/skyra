# Phase: Ownership — What Lives Where

## Context

Motor is dispatch. Intent graph is a struct. TrainOfThought collapsed with Visited into one ordered map. ActionHistory shipped as a binding reality on the Relation. Sensory binding reality designed but not built — triggers on empty TrainOfThought, routes the impulse into the topology.

## The Question

Ownership of state is back in frame. What belongs on the Relation (shared) vs on the being's Relationships (private) vs on the Slice (per-traversal)?

Current layout:
- **Relation.Threads** — shared. Multiple beings bind to the same thread. This is correct — a conversation is between beings, not owned by one.
- **Relation.ActionHistories** — per-being, but sitting on the shared Relation. Skyra's action history is Skyra's private record. Michael doesn't need to see it. Should this move to Skyra's Relationships?
- **Slice.TrainOfThought** — per-traversal, private. Correct.
- **Slice.Thread (nested map)** — per-being private copy of thread state. Correct but not yet populated.
- **Sensory (proposed)** — a cell on Skyra's Relationships that points at thread, action history, etc. Routes impulse into the topology on first contact.

The pattern that's emerging: the Relation holds shared surfaces (threads, world). The being's Relationships hold private topology (sensory, action history, memory cells). The Slice holds per-traversal working state.

ActionHistory might need to move from Relation to Relationships. Sensory needs to be wired. The being needs an internal topology that makes sense — sensory routes in, memory/action history accumulate, motor routes out.

## Next

Resolve ownership. Wire sensory. Get a clean round-trip.

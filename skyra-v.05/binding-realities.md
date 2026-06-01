# Binding Realities

A binding Reality is a Reality that owns a local activation field over the graph. It doesn't store state on every node. It imprints its field onto the Relation during descent, the Relation carries that field during traversal, and the return path updates the binding Reality's own weights from what was actually used.

## The Pattern

1. On descent, the binding Reality's Observe imprints its local field onto the Relation
2. The Relation carries that field — a map of node IDs to weights
3. Activation at each node reads the field to decide what's loud or quiet
4. The Relation records what was actually hit
5. On ascent, the binding Reality's Express updates its durable weights from the hit evidence

Durable state lives in the binding Reality. Traversal-local evidence lives on the Relation. The return path reconciles them.

## The Activation Equation Generalizes

Thread alignment and thread strength were the first two binding fields, added as scalar terms to the activation equation. Binding Realities generalize this — the equation doesn't have a fixed number of terms:

```
activation = global_weight * local_weight * recency * Π(binding_fields)
```

Base activation is three terms that belong to the node and the edge — intrinsic. The product over binding fields is extrinsic — whatever the Relation is carrying from the binding Realities it passed through. The number of terms is however many binding Realities are active on this traversal.

### v.1 Fields semantics

Fields on the Relation: `memoryID → {bindingReality: activation}`. Each memory (cell) has its own slot. The inner map holds the activation values for each topological layer — thread alignment, thread strength, emotional, episodic, etc. The binding reality name is the key, the activation value is the value.

```
Fields[memoryID] = {
    "thread:alignment": 1.0,
    "thread:strength":  1.0,
    "emotional":        0.7,
    ...
}
```

## Binding Realities Are Brain Regions

Each binding Reality interprets the same Relation through its own local field:

| Binding Reality | What it tracks | Brain parallel | Floor |
|---|---|---|---|
| **Thread** | What matters in this conversation right now | Hippocampus | 0.0 — hard gate |
| **Emotion** | What feels similar | Amygdala | >0 — soft amplifier |
| **Causal** | What led to what | Prefrontal cortex | >0 — soft amplifier |
| **Task** | What serves the current objective | Executive function | Variable — hard gate in focused execution, soft in exploration |
| **Relationship** | What matters between two beings | Social cognition | >0 — soft amplifier |

Each one is the same mechanism: a Reality with a local field, imprinted onto the Relation, learned from on the return path. No separate subsystem for memory vs emotion vs task tracking. One primitive.

## Gates vs Amplifiers

Not every binding Reality should be a gate. Thread is a gate — if you're not on this thread, you're zeroed out. But emotional resonance shouldn't gate. A memory that's emotionally relevant to the current state but on a different thread should still be reachable — amplified by emotion, dampened by thread, net activation depends on the product. If emotion were also a gate, anything outside the current emotional state gets zeroed. That's dissociation, not cognition.

Each binding Reality's field has a floor — the minimum value it returns for nodes it has no opinion about:

- **Floor = 0.0** — hard gate. Node must be in this field to activate at all. Thread works this way.
- **Floor > 0** — soft amplifier. Nodes outside the field still pass through at reduced amplitude. Emotion, relationship, causal work this way.
- **Variable floor** — context-dependent. Task region gates hard during focused execution and softens during exploration. The floor is a property of the binding Reality, not of the architecture.

## Thread Strength — What Matters Right Now

Thread alignment answers "is this node part of this conversation." Thread strength answers "what is important in this thread right now."

A thread accumulates weight unevenly. The nodes that matter at exchange 50 are not the same nodes that mattered at exchange 5. Thread alignment treats them equally — both are on this thread, both get 1.0.

Thread strength tracks what the thread has reinforced through use — which nodes keep getting hit, which ones were mentioned once and faded. It's the thread's own opinion about what's load-bearing right now.

```
ThreadReality.Relationships[node_id] -> EdgeReality{Weight = thread_strength}

descent:
  ThreadReality reads its local map
  ThreadReality writes Relation.Fields[node_id]["thread:strength"] = thread_strength

traversal:
  activation reads the field
  Relation records visit / activation evidence

ascent:
  ThreadReality reads hit evidence
  ThreadReality updates its local thread weights
```

The thread is both prior and learner: prior on descent, learner on ascent. This is what makes the cost curve inversion work — the thread teaches the system what matters, and the system gets cheaper as the thread gets longer.

## Unity on the Return Path

Multiple binding Realities produce multiple regional interpretations of the same traversal. The thread says "these nodes are hot." The emotional field says "these nodes resonate." The task field says "these nodes serve the objective."

On the return path, the being's Express receives one Relation carrying all of this. The compression into thought or action integrates across all the regional fields. Unity doesn't need a special integration step. It's what happens when the return path compresses overlapping perspectives into one output. The binding problem is solved structurally — the same Relation, the same ascent, the same compression boundary.

## What This Replaces

The five-term activation equation (`global_weight * local_weight * recency * thread_alignment * thread_strength`) is the special case where thread is the only binding Reality. The generalized form absorbs thread_alignment and thread_strength into the thread binding field and makes room for any number of additional binding Realities without changing the equation.

For v.1: thread is the only binding Reality. The mechanism is general but the implementation starts with one.

# Decision Spaces

Skyra is designed to preserve possibility before it collapses behavior.

Most agent runtimes collapse too early:

```text
input -> classify -> choose tool -> execute -> summarize
```

That shape is easy to engineer, but it strangles the model almost immediately. The model is forced into a narrow command path before it has inhabited the field of memory, context, constraints, goals, and available actions. The runtime becomes an orchestration pipeline instead of a place where cognition can develop.

Skyra aims for a different shape:

```text
input -> open field -> local activations -> binding convergence
      -> partial collapse -> result -> field update -> next collapse
```

The substrate is a possibility engine. It holds a field of experience and collapses those possibilities one at a time.

## Delayed Collapse

A Relation begins as a field of possible observations, memories, goals, and actions. Traversal does not immediately choose one interpretation or one tool. Cells activate locally. Binding realities imprint partial constraints. Motor acts only when enough independent fields converge.

Each action is a collapse:

```text
many possible meanings/actions
  -> convergence on one primitive + target
  -> permission check on the world plane
  -> action/result
  -> experience written back into the field
```

The point is not to remove determinism from the runtime. The physics should be inspectable: activation, fields, weights, thresholds, traversal budgets, and permission gates should all be visible. The point is to avoid premature behavioral determinism.

The substrate can be deterministic while still preserving multiple possible futures until convergence.

## Permissions Are Not Cognition

Permissions belong on the world plane. They constrain enactment, not imagination.

A being may form an intent it cannot execute. Motor checks the world plane before acting. If an action is denied, that denial becomes experience. The being learns the boundary as part of the field instead of being silently prevented from ever approaching it.

```text
substrate: what can be considered
world plane: what can be enacted
```

This keeps safety and capability from fighting at the wrong layer. The substrate remains open enough to explore possibility. The world plane remains responsible for what actually happens.

## Auditable Decision Spaces

The major consequence of delayed collapse is that decision spaces become first-class runtime artifacts.

Most systems audit only the final act:

```text
model called tool X
tool returned Y
model said Z
```

Skyra should audit the field that made an act possible:

```text
what possibilities were active
which cells participated
which binding realities supported each candidate
which fields disagreed
which actions almost fired
which action crossed threshold
what was suppressed by low activation
what was denied by permission
what result rewrote the field
```

This is different from asking a model to explain itself after the fact. The runtime records the decision space before collapse. Explanation becomes structural, not theatrical.

## Motor As Collapse Gate

Motor should not deliberate. Motor detects convergence.

Binding realities express votes in a shared primitive vocabulary:

```text
look(target)
listen(target)
engage(target)
speak(target)
```

Each vote carries source, weight, and metadata. Motor aggregates the votes. When enough independent fields converge on the same primitive and target, the action crosses threshold and collapses into the world.

The audit record should preserve both the winner and the near misses.

```text
candidate action = primitive + target
support = binding fields that voted for it
agreement = convergence across fields
permission = whether the world plane allows enactment
result = what happened after collapse
```

## Inference As Local Repair

Inference should not fire simply because activation is high.

High agreement means the topology already knows what to do. That is reflex. Low activation means the cell is not relevant. Inference belongs in the middle: where activation is relevant but the fields disagree, the setpoint does not match the members, or the runtime cannot resolve closure from weights alone.

```text
low activation -> no participation
high agreement -> reflex
meaningful disagreement -> inference
```

Inference repairs local incoherence on the ascent. A cell compares what it is trying to be with what its members now say. If the error cannot be resolved structurally, inference recognizes the resolution. The result is written back into weights so the same situation can become cheaper next time.

## Vocabulary

These boundaries matter:

- **Relation** — the signal plane; the medium carrying the current wave.
- **Slice** — per-being traversal physics on a Relation.
- **Being** — durable membrane around a private brain.
- **Brain** — the being's persistent topology.
- **Cell** — local Skyra DNA instance with type-specific observe/express behavior.
- **Binding reality** — an index over cells that imprints fields and learns from traversal.
- **Motor** — convergence gate that collapses votes into world action.
- **World plane** — permissioned enactment surface.
- **Decision space** — the active field before collapse.

## Design Principle

Skyra should not merely record what happened.

It should preserve why that action became possible, why alternatives failed, which boundaries prevented enactment, and how the result changed the field.

That is the difference between an agent log and an auditable cognitive runtime.

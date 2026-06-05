# June 4 Improvements

## Hebbian Reinforce

`Reinforce()` should be proportional to the input signal strength (Hebbian), not a flat tick. Stronger activation → stronger weight update. That's how LTP works — the magnitude of the postsynaptic response determines the plasticity.

Currently:
```go
func (s *Skyra) Reinforce() {
    s.IdentityWeight = Identity(s.IdentityWeight, true)
    s.FamiliarityWeight = Familiarity(s.FamiliarityWeight, true)
}
```

The EMAs always receive `1.0` as the activation input. A cell that barely crossed threshold gets the same reinforcement as one that lit up hard. The activation score computed during traversal is discarded before it reaches the weight update.

The fix: pass the activation score into `Reinforce()` and use it as the signal value instead of a flat `1.0`.

```go
func (s *Skyra) Reinforce(activation float64) {
    s.IdentityWeight = Identity(s.IdentityWeight, activation)
    s.FamiliarityWeight = Familiarity(s.FamiliarityWeight, activation)
}
```

This requires `Identity()` and `Familiarity()` to accept a float instead of a bool — the signal strength replaces the binary traversed/not-traversed.

Biology: synaptic weights modulate EPSP amplitude (graded, analog). The magnitude of the postsynaptic response determines the plasticity update. Flat reinforcement is biologically wrong — it treats every firing event as equally important regardless of how strongly the cell was driven.

## Invert Competence — Inference Fires at Binding Field Intersections

The current competence model is backwards. High activation triggers propagate (expensive inference call). Low triggers pass-through. Biology says the opposite.

**Current (wrong):**
```go
switch {
case activation > 0.6:
    s.Type = "propagate"  // inference fires
case activation > 0.2:
    s.Type = "pick-up"    // absorb content, no inference
default:
    s.Type = "pass-through" // dormant
}
```

**Correct model:**
- **High activation** — deep attractor. Weights already know the answer. No inference needed. Hot path. Reflex. Skill.
- **Low activation** — dormant. Not relevant. Pass through.
- **Middle (binding field disagreement)** — inference fires. The goal field says yes, the thread field says no, the cell can't resolve the conflict from weights alone. Bifurcation point. The cell samples, explores, commits.

The trigger for inference is not activation magnitude — it's **disagreement between binding field values**. When the fields align (all high or all low), the answer is clear. When they conflict, inference is needed to break the tie.

**Economics:**
- New being = expensive. Nothing grooved, everything at boundary points, inference fires constantly.
- Mature being = cheap. Deep attractors everywhere, most traversals resolve from weights. Inference only at the edges of what it knows.
- Growth reduces inference cost. The being literally gets cheaper to run as it matures.

This changes `Competence()` from reading one activation score to reading the spread/variance across binding field values for the current node. High variance = disagreement = inference. Low variance = consensus = no inference.

## Future: Reflex Is Metadata Convergence

A reflex isn't just when the primitives agree — it's when the metadata around the same primitive + target agrees across binding realities. The *why* is the same pattern the being has seen before.

High metadata variance across binding realities = novel situation, each one has a different reason = inference fires. Low metadata variance = familiar pattern, everyone agrees on why = reflex, no inference.

This makes the inference trigger concrete: it's not abstract "field variance." It's the metadata around the same vote either matching or not matching. First time the metadata diverges (action has no history). Hundredth time it converges (every binding reality has seen this exact pattern). The transition from inference to reflex is the transition from metadata divergence to metadata convergence.

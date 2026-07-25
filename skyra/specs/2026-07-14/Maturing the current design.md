# Maturing the Current Design

## Note — Where an Expectation Lives

The current distinction is:

> The observation graph answers: What happened in this world?
>
> The expression graph answers: What have I tried before, what did I expect, and how reliably did it work?

This needs to be fleshed out. “What did I expect?” is itself an observation: the being can observe that it formed a particular expectation at a particular time. We need to distinguish the expectation as an event that happened from the expectation's propositional content, its causal attribution, and the learned reliability accumulated from later world outcomes.

Do not assume yet that “what I expected” belongs exclusively to the expression graph. Specify how the observation and expression graphs each represent or reference the expectation without collapsing self-observation into evidence that the expected world state was true.

## Correction — One Durable Graph, Different Node Roles

All three node types must be durable. Durability is not the distinction between them, and beginning a new episode must not mean discarding the observation history.

They are different kinds of nodes in the same graph — three of them: observation, state change, action. Given an observation node, its children are the state-change nodes it was interpreted into; each state change leads to the action nodes it motivated; each action leads to the later observation the world returned:

```text
observation → state change → action → observation
```

Each edge records a relationship: observation→state-change is the interpretation, state-change→action the policy, action→observation the consequence. State changes, actions, and their eventual resolutions must remain addressable within this durable lineage. (Canonical statement of the ontology: [../../design-docs/skyra-dna-v3/README.md](../../design-docs/skyra-dna-v3/README.md).)

An episode may define a scope, root, or active view over the graph, but it must not define which history is durable. The current implementation's fresh per-episode observation graph and persistent expression graph should therefore be treated as a temporary experimental arrangement, not the mature ontology.

## Note — Target-Rooted Kernel Traversal

Planning traversal begins at Skyra's desired target state, not at the current observation. Skyra materializes the desired state, then hands the kernel:

```text
origin:               desired target state
morphological target: current observed state
```

The kernel walks backward through expressions that previously produced the desired state, regressing their expected outcomes into their required predecessor states. Traversal closes when the regressed state is contained in the current observed morphology:

```text
regressed state ⊆ current observed state
```

The resulting lit bridge is discovered from desired back to current and can later be followed from current toward desired. The kernel owns only graph traversal, ordering, admission, termination, and the fold. Skyra owns the target, current morphology, wave purpose, worker selection, inference on an empty frontier, and action scheduling. The scheduler is not part of the kernel.

## Note — State Change Turns the Gears

An observation does not turn the agent's gears merely by arriving. It first has to revise Skyra's represented current state or target state. A revision to either changes the tension between where Skyra represents itself as being and where it is oriented toward going.

The mechanism begins with this question:

> **Given a change to either target or current state, what observation needs to be true given that change?**

That question turns a change in orientation or situation into an observable condition. The condition can then lead into the expressions associated with making it true, and the traversal can continue backward toward the currently grounded morphology.

The distinction to preserve is:

```text
observation arrives → current or target state may change → state tension changes → cognition turns
```

A raw observation that changes neither current nor target state is history, but it is not itself a new turn of the cognitive mechanism.

## Note — The Learning Chain and Its Axes

Keep the pointers distinct:

```text
observation → state change → action → observation
```

The state change is Skyra's interpretation of the observation. The action is Skyra's interpretation of the state change. The observation that follows the action is the world's return, beginning the next turn of the chain.

This exposes two primary learning workers, or learning axes:

1. **Observation → State Change** — learn how an observation should revise Skyra's represented current or target state.
2. **State Change → Action** — learn what action follows from, supports, or responds to that interpreted change.

There is likely a third learning axis:

3. **Action → Observation** — learn what observations return after an action, and therefore how Skyra's actions are changing its environment.

Together they form a complete experiential loop:

```text
interpret the world     interpret the change       learn the consequence
observation ──────────→ state change ────────────→ action ───────────────→ observation
```

Do not collapse these pointers. Each relationship learns a different thing, and preserving them separately keeps interpretation, action selection, and environmental consequence distinguishable.

### The Minimal Learning Set

The three learning questions are:

```text
Given observation Oₜ:
    What is the most likely state change ΔSₜ?

Given state change ΔSₜ:
    What is the most likely action Aₜ?

Given action Aₜ:
    What is the most likely next observation Oₜ₊₁?
```

They close into the experiential loop:

```text
Oₜ → ΔSₜ → Aₜ → Oₜ₊₁ → ΔSₜ₊₁ → Aₜ₊₁
```

These are three recognizable learning mechanisms:

```text
P(ΔS | O)   interpretation model
P(A | ΔS)   action/policy model
P(O′ | A)   environmental consequence model
```

In practice, each relationship inherits context from its node and the morphology in which it occurs:

```text
P(ΔS | O, current state, target state)
P(A | ΔS, current state, target state)
P(O′ | A, current state)
```

That context does not need to complicate the basic mechanism. Observation, state-change, and action nodes can point into the relevant morphology.

Each learning worker should preserve a bounded probability distribution rather than immediately collapsing to one answer:

```text
Observation O
    ├── 0.70 → StateChange ΔS₁
    ├── 0.22 → StateChange ΔS₂
    └── 0.08 → StateChange ΔS₃
```

The kernel can order that frontier and admit the most probable paths until its morphological condition is reached.

The meanings of the three pointers remain distinct:

- **Observation → State Change:** how Skyra interpreted what it perceived.
- **State Change → Action:** how Skyra responded to that interpretation.
- **Action → Observation:** what the environment returned following the action.

This chain describes the learned experiential structure. The kernel may traverse these pointers forward or backward depending on its origin and morphological target, but traversal direction must never change what a pointer means.

Together the three axes provide a minimal learning substrate for perception, orientation, action, and consequence, with each relationship independently learnable and gradeable.

### Do not confuse the learning axes with compute routing

The three axes are the right decomposition of **what is learned**. They are *not* three tiers of
where compute is spent. Compute routes on an orthogonal property — whether a given piece of input is
*resolved cleanly* or *not* — which cuts across all three axes (see
[principle-falsifiable-boundary.md](principle-falsifiable-boundary.md)). An axis-3 return can carry
interpretation; an axis-1 observation can be fully deterministic. Learn along the axes; spend along
the resolved/unresolved boundary. The boundary migrates toward "resolved" as edges converge, and
that migration is what the learning curve measures.

# Workers

**Status:** DESIGN · part of [skyra-dna-v3](README.md) · see [the-wave.md](the-wave.md) for when they fire

Six workers. **Only two call a model.**

---

## 1. Decomposer

```text
TRIGGER    an impulse arrives — user message, world Outcome, perception return
INPUT      the raw impulse + the known-ID table
OUTPUT     observation triples → written to the OBSERVATION graph
MODEL      yes — small
PRESENT    no. it receives the impulse, not a frame.
BLOCKS     NO. async + buffered.
FAILURE    unparseable → drop the line, log it. never fatal.
```

**LAW 1 — the decomposer is the only writer of world state.** Nothing else may write it.

**LAW 3 — exact first.** The string matcher resolves every world-given namespace (`file:`, `proc:`,
`url:`, `coord:`) before the model sees anything. The model decomposes only the residual.

**It must not block the cognitive loop.** Text → triples is slow (model call + entity resolution).
Producer/consumer: the decomposer fills the observed graph, cognition runs on whatever has resolved,
and late edges arrive next cycle.

---

## 2. Aligner

```text
TRIGGER    the exact matcher left a residual — an unresolved reference, or an
           `interpret` expect awaiting an observation
INPUT      the residual + candidate nodes from the graph
OUTPUT     a coreference mapping, or a translated claim
MODEL      yes — small
PRESENT    no. it receives candidates, not a frame.
BLOCKS     no
FAILURE    ambiguous → LOW-WEIGHT CANDIDATES, never a hard link.
```

**LAW 4 — runs ONCE, on the observation side only. May translate. May never judge.**

```text
allowed:     "these two nodes are the same referent"
allowed:     "michael's reply 'yeah I'm good' maps to state=satisfied"
FORBIDDEN:   "the being did well"                ← no source. that is the poison.
```

**Ambiguity is preserved, not resolved.** Forcing a hard link too early corrupts the graph
permanently. Entity resolution is the floor: relevance-correctness ≤ resolution-correctness.

---

## 3. Planner

```text
TRIGGER    a workflow node becomes REALIZABLE — its requires are satisfied by
           OBSERVED state. Also: the target-state change that starts the wave.
INPUT      the present (wave 2, bidirectional, budget = context tokens)
OUTPUT     ONE Expression — framing + acts (payloads included) + expects
MODEL      yes — THE big call. exactly one per invocation.
PRESENT    YES. the only worker that gets one.
BLOCKS     it IS the wave at that node
FAILURE    Validate() rejects any emission where an act projects no expect (LAW 5).
           unparseable lines are dropped, not fatal — fewer predictions is acceptable,
           no action is not.
```

**The planner emits payloads.** The file's contents, the message's text — all of it. This is why the
emission is large, and it is what lets motor be dumb.

**The framing it emits is that node's system prompt**, and it is graded by whether its expects came
true. That is the entire self-authorship mechanism: the model never rewrites a global prompt, it
writes claims, and the world selects among them.

---

## 4. Placer (`parserPayloadNodePlacer`)

```text
TRIGGER    a parsed emission has unplaced nodes (parserPayload > 0)
INPUT      the parsed nodes
OUTPUT     nodes inserted into the workflow graph; the hole set recomputed
MODEL      NO
PRESENT    NO
BLOCKS     no — it is graph insertion
FAILURE    n/a
```

Placement links each act's `expects` to the triple keys they claim and its `requires` to the keys they
need, then recomputes which holes are grounded.

**Drain completely before checking closure.** Extra nodes the plan did not strictly need are still
gradeable bets — discard them and you discard learning signal, and placement costs nothing.

**The placer is where parallel width is born, and width is money.** Throttle it against the spend
account rather than letting the DAG's shape dictate cost.

---

## 5. Motor

```text
TRIGGER    an act's requires are ALL satisfied → Runnable()
INPUT      the act. payload already determined by the planner.
OUTPUT     an invocation into the world → an Outcome
MODEL      NO
PRESENT    NO
BLOCKS     yes — it is doing the thing
FAILURE    Outcome: blocked / denied / stale / unavailable / partial / indeterminate.
           the Outcome is an OBSERVATION and re-enters via the decomposer.
```

> **Motor does not think. It executes.** Which is what motor is in a human.

This deletes v2's motor-collapse inference call entirely.

---

## 6. Grader

```text
TRIGGER    an observation resolves an open expect
INPUT      the expect + the observed triple
OUTPUT     a grade write on the edge (projection → act → observation)
MODEL      NO — it is a set difference
PRESENT    NO
BLOCKS     no
FAILURE    n/a — arithmetic
```

```text
INDIVIDUAL expect fails  → grade the ACT             (the door didn't open)
JOINT      expect fails  → grade the FRAMING that declared it
                           (open door + told her ≠ she walks through)
```

Two grades, two homes, no ambiguity, and no new machinery: the joint was declared by a planner node,
so it lands on that node's framing.

Grade weight is **channel fidelity** (`Mode.Weight()`): `exact` ~1.0, `interpret` ~0.3, no-source 0
and forbidden.

---

## Cost per cycle

```text
decomposer   1 small call   (async, off the critical path)
aligner      0–1 small call (only on the residual)
planner      1 big call     (only when a node becomes realizable)
placer       0
motor        0
grader       0
```

**~2 calls per cycle**, against **95 for six messages** on the first live run.

And the planner's count is the learning curve: `N → 1`.

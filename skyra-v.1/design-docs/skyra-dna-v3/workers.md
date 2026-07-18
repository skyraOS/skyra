# Workers

**Status:** DESIGN · part of [skyra-dna-v3](README.md) · see [the-wave.md](the-wave.md) for the waves they run in

**Six workers, four phases. Two call a model. Only one is expensive.**

| # | worker | trigger | model? | present? |
|---|---|---|---|---|
| **INTAKE** ||||
| 1 | **Decomposer** | an impulse arrives | yes | **yes** |
| **PLAN** ||||
| 2 | **Planner** | a workflow node becomes realizable | **yes (big)** | **yes** |
| 3 | **Placer** | `parserPayload > 0` | no | no |
| **EXECUTE** ||||
| 4 | **Motor** | an act's `requires` are all satisfied | no | no |
| **SETTLE** ||||
| 5 | **Settler** | the tick passes an expect's `MaturesAt` | no | no |
| 6 | **Grader** | an observation resolves an open expect | no | no |

*(The **Aligner** is gone. See "Why there is no aligner" — deleting it is what makes the grading path
model-free.)*

---

## 1. Decomposer

```text
TRIGGER    an impulse arrives — user message, world Outcome, perception return
INPUT      the impulse + a PRESENT (bidirectional: current state ← impulse → expected state)
OUTPUT     observation triples  → observation nodes
           target updates       → the expected state
MODEL      yes
PRESENT    YES — it traverses from current state AND expected state
BLOCKS     NO. async + buffered.
FAILURE    unparseable → drop the line, log it. never fatal.
```

This is the **memory worker**: *"every impulse tasks a memory worker with maintaining both endpoint
graphs."* It is not a text parser. It reads the impulse **in the context of** what is already known
and what is currently wanted, and it updates both.

**It is the `observation → state change` arrow.** The planner is `state change → action`.

```text
DECOMPOSER   what does this MEAN?    → observation triples + target updates
PLANNER      how do I GET THERE?     → acts + expects
```

**LAW 1 — the decomposer is the only writer of observation nodes.**

**LAW 3 — exact first.** The string matcher resolves every world-given namespace (`file:`, `proc:`,
`url:`, `coord:`) before the model sees anything. The model handles only the residual.

**It must not block the cognitive loop.** Text → triples is slow. Producer/consumer: the decomposer
fills in observation nodes, cognition runs on whatever has resolved, late edges arrive next cycle.

### Its writes are the entire reactive surface

```text
observation resolves an open bet          → GRADER fires
observation contradicts a LOAD-BEARING    → KILL the wave, replan
observation is purely additive            → nothing fires. it was news.
the TARGET changed                        → a new WORKFLOW WAVE fires
```

### Why there is no aligner

The aligner was going to do two things: coreference (`{the cup}` → a known node) and translation
(`"yeah I'm good"` → `state satisfied`). **Both happen during decomposition** — the decomposer already
has the entity vocabulary in its present, and resolving references *is* decomposing.

The payoff is large. If the decomposer emits `being:michael state satisfied` as an **observation
triple on the way in**, then at grade time the grader does a **pure exact match**.

> **Interpretation happens at OBSERVATION time, not at GRADING time.**
>
> Which means the model **cannot touch the grading path at all.** LAW 4 stops being a rule anyone has
> to keep and becomes structural: there is no grading-time aligner to misbehave.

### LAW 6 — Observation is produced blind to desire

This is the hole that deleting the aligner opens, and it is a bad one.

If the decomposer's present contains **the open bets**, it knows what answer was hoped for. Michael
says something ambiguous, the decomposer sees `expect: being:michael state satisfied @ 0.4`, and
resolves the ambiguity **toward the bet**. That is the being manufacturing the evidence that grades
it. Same for the target: if it knows the being *wants* Michael happy, ambiguity drifts toward happy.

> **LAW 6 — The pass that turns an impulse into observation triples must not see the open bets, their
> confidences, or the target. It sees only the ENTITY AND PREDICATE VOCABULARY — enough to speak the
> right language, not enough to know what answer was wanted.**

Two passes, structurally separated (possibly one call, hard-split):

```text
PASS 1  TRANSCRIBE   impulse + vocabulary            → observation triples.  BLIND TO DESIRE.
PASS 2  RE-AIM       those triples + current target  → target updates.       DESIRE LIVES HERE.
```

**Pass 1 never sees a bet. Pass 2 never writes an observation.** The whole architecture rests on the
observation channel having no stake in the prediction, and this is the one place we were about to hand
it one.

### OPEN — the interpretation layer has no feedback loop

**The decomposer is the one component nothing grades**, and its output is trusted absolutely.

Worse, it is a **boundary violation**:

```text
RAW           michael said "grab me a glass of water"   ← the world's. unfalsifiable. it happened.
INTERPRETED   being:michael wants water                  ← THE MODEL'S READING of the world's.
```

The second is a model output sitting on the world's side of the boundary, wearing the world's
authority. LAW 1 does not catch it, because it is not an expression.

> **The interpretation layer is the one place model output crosses onto the "world" side. That is
> exactly why it needs its own grounding.**

The eventual shape: an interpreted triple is a **claim**, so it is an edge with a `P`, like everything
else —

```text
"grab me a glass of water"  ──→  being:michael wants water      P(interpretation | impulse kind)
"yeah I'm good"             ──→  being:michael state satisfied
```

— graded by **later contradiction** (bring the water; he says *"no, I meant the report"*). Slow,
sparse, indirect, but real, and it reuses machinery that already exists. A second, diffuse signal: a
systematically wrong decomposition makes *everything* downstream miss, which is detectable as a
common-cause anomaly even when it cannot be localized.

**TABLED.** But three things are required NOW, because they are cheap today and impossible to
retrofit:

1. **Mark every observation triple `raw` or `interpreted`.** If they are indistinguishable, the
   feedback loop can never be added — you will not know which claims were yours.
2. **Interpreted triples carry a confidence and a `DerivedFrom` pointer to the raw evidence.** That
   pointer *is* the edge you will eventually grade.
3. **An interpreted triple never gets `exact` fidelity.** It resolves an expect at ~0.3, same as an
   `interpret`-mode claim — because that is what it is. Letting an inference grade a bet at 1.0 is the
   failure this architecture exists to prevent.

---

## 2. Planner

```text
TRIGGER    a workflow node becomes REALIZABLE — its requires are satisfied by OBSERVED state.
           Also: the target-state change that starts the wave.
INPUT      the present (wave 2, bidirectional, budget = context tokens)
OUTPUT     ONE Expression — framing + acts (payloads included) + expects (each with a `within`)
MODEL      yes — THE big call. exactly one per invocation.
PRESENT    YES
BLOCKS     it IS the wave at that node
FAILURE    Validate() rejects an emission where an act projects no expect, or an expect carries no
           deadline (LAW 5). Unparseable lines are dropped, not fatal — fewer predictions is
           acceptable; no action is not.
```

**The planner emits payloads.** The file's contents, the message's text — all of it. This is why the
emission is large and why motor can be dumb.

**The framing is that node's system prompt**, and it is graded by whether its expects came true. The
model never rewrites a global prompt; it writes claims, and the world selects among them.

Its call count is the learning curve: **`N → 1`**.

---

## 3. Placer

```text
TRIGGER    a parsed emission has unplaced nodes (parserPayload > 0)
INPUT      the parsed nodes
OUTPUT     nodes inserted into the workflow; the hole set recomputed
MODEL      NO
PRESENT    NO
```

**Drain completely before checking closure.** Extra nodes the plan did not strictly need are still
gradeable bets — discard them and you discard learning signal, and placement costs nothing.

**The placer is where parallel width is born, and width is money.** Throttle it against the spend
account rather than letting the DAG's shape dictate cost.

---

## 4. Motor

```text
TRIGGER    an act's requires are ALL satisfied → Runnable()
INPUT      the act. payload already determined by the planner.
OUTPUT     an invocation into the world → an Outcome
MODEL      NO
PRESENT    NO
FAILURE    Outcome: blocked / denied / stale / unavailable / partial / indeterminate.
           the Outcome is an OBSERVATION and re-enters via the decomposer.
```

> **Motor does not think. It executes.** Which is what motor is in a human.

This deletes v2's motor-collapse inference call entirely.

---

## 5. Settler

```text
TRIGGER    the tick passes MaturesAt on an unresolved expect  → Expression.Due(tick)
BEHAVIOR   emit perceive(subject, predicate). MOTOR executes it.
MODEL      NO — the expect already names its own target
PRESENT    NO
```

**This is what makes absence decidable.** A targeted perceive comes back authoritative for exactly the
triple that was bet on, so within it, **absence IS falsity** — and the general closed-world problem
never has to be solved. *The bet manufactures its own evidence.*

It is also what makes the being **curious about its own open bets**, for free, with no prompting and
no inference.

### Three properties

**Settlement outlives the wave.** The wave that placed the bet may be long dead. The bet lives in the
memory graph and the grade is what matters, so **settlement acts belong to no workflow** — they are
standalone. A dead plan's bets still get settled, and that is correct: you still want to know whether
you were right.

**Batch the perceives.** Every matured bet is a world interaction, and interactions cost. Group matured
bets by target scope and emit **one perceive covering many triples** — the scatter-gather the kernel
already does. Otherwise a plan with 40 expects produces 40 perceives.

**Perceives are not inference, but they are not free.** They cost world interaction, latency, and
possibly money (a device call). Budget them against the **spend account**, not the token budget.

---

## 6. Grader

```text
TRIGGER    an observation lands whose Key() matches an open expect
BEHAVIOR   resolve → fold into the edge's EMA → update the framing's Score → record ResolvedBy
MODEL      NO — it is arithmetic
PRESENT    NO
```

Probably **not a worker at all** — a hook on the observation write path, not a scheduled unit with a
lifecycle.

```text
INDIVIDUAL expect  → grades the ACT's edge
JOINT      expect  → grades the SET-SOURCED edge   {E₁, E₂} → W
framing            → accumulates Score (its edge over the base rate)
```

**Joints have a different readiness condition.** An individual expect resolves as soon as its triple is
observed. A joint (`of="a1 a2"`) may only resolve once **all** its acts have completed — otherwise you
would grade `E₁ ∧ E₂ → W` when only `E₁` ran, and the edge would learn garbage.

Grade weight is channel fidelity (`Mode.Weight()`): `exact` ~1.0, `interpret` ~0.3, no-source **0 and
forbidden**.

---

## The job board

A **supervision tree**, keyed by wave.

```text
Row {
    WaveID      the key
    Purpose     workflow | present | perceive       ← what this wave is FOR
    WorkflowID  the wave this belongs to.
                IF Purpose == workflow, it references ITSELF → it is a ROOT.
    Status      running | done | killed
    ReadVersion the endpoint versions it read (staleness detection)
    Spend       tokens + money attributed to this wave
    Trace       append-only
}
```

**Self-reference at the root** is what lets you walk *up* from any wave to the workflow that owns it,
with no special case. That buys three things:

**Cancellation.** Kill the workflow wave → every present and perceive wave beneath it dies with it. The
kill-condition becomes cheap, and it is safe precisely because **waves write nothing durable** (see
[the-wave.md](the-wave.md)) — there is nothing to unwind.

**Accounting.** Spend, tokens, and latency attribute to the **workflow**, so you can ask *"what did
pursuing this goal cost me"* — which is the unit the daily budget actually cares about.

**Tracing.** The whole tree of work for one goal, in one place.

### What is gone versus v2's board

**No join. No fan-out set. No atomic motor guard. No one-reframe rail.** Those were coordination for a
pipeline that no longer exists. The wave owns its own execution, so the row is just identity, purpose,
parentage, status, versions, spend, and trace.

---

## Cost per cycle

```text
decomposer   1 call         (async, off the critical path)
planner      1 big call     ← the learning curve lives here: N → 1
placer       0
motor        0
settler      0 inference    (costs world interaction, not tokens)
grader       0
```

**~2 model calls per cycle**, against **95 for six messages** on the first live run.

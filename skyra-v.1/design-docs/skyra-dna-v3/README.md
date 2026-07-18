# skyra-dna-v3

**Status:** BUILDING — the walking skeleton runs against a live model, tests pass, and the first
learning curves (prediction-diff, retrieval) have appeared. Trace visibility is specced, not built.
The ontology is being matured in [specs/2026-07-14](../../specs/2026-07-14/). v2 will be archived
once v3 stands.
**Derived from:** `scratchpad/2026-07-13/` (the seven working notes).

## The one-line version

> The being commits to a falsifiable expected world state, acts, observes what came back, and the
> **difference is the grade** — computed, never judged.

## Why v3 exists

v2's ceiling was structural, not incidental:

- `Realize`-style free-text predictions could only be graded by an **LLM judge**, so most of the
  learning signal bottomed out in an opinion with no independent source.
- Causality had to be **asserted** by the model (annotation grammar, retain/compose/requery,
  causal-belief expansion), which made the causal graph an opinion too, and cost most of the 95
  inference calls in the first live run.
- The goal was re-derived every turn. The being could not **want** anything for longer than one
  impulse.

v3 removes all three by changing the representation, not by adding machinery.

## Vocabulary — use these names, no others

There is **one durable store: the memory graph.** It holds three kinds of node, and the loop runs
through them. All three are durable.

```text
THE MEMORY GRAPH — one store, three node types

  observation    what crossed the boundary from the world; reality's return.
                 append-only, immutable. the ONLY node type the decomposer writes.

  state change   Skyra's interpretation of an observation — a revision to current or target
                 state. has interior: a partially-specified region (slots), never an atom.

  action         what Skyra emitted back across the boundary; its outcome is a new observation.

LEARNING LIVES ON THE RELATIONSHIPS BETWEEN NODE TYPES — not in a separate store:

  observation → state change   P(ΔS | O)   interpretation
  state change → action        P(A | ΔS)   policy
  action → observation         P(O′ | A)   consequence
  >>> THE GRADED RELATIONSHIPS ARE THE BEING'S MEMORY. <<<

DERIVED  (rebuildable, never authoritative)
  world state    a fold over the observation nodes.
  reverse index  posting lists over the graph.
  workflow       the plan. per-target, disposable, thrown away on kill.
```

Starting an episode scopes a **view** over the graph — it never decides what is durable.

**Migration key** (older sub-docs still use the two-graph words): what they call the *expression
graph* is now the **state-change and action nodes plus their graded relationships** — still the
memory, no longer a separate graph. What they call the *observation graph* is the **derived fold
over the observation nodes** (= world state); the observation nodes themselves are durable in the
one graph. Read every "two graphs" claim through this key until the sub-docs are swept.

### Words that are NOT stores

```text
world state     the fold over the observation nodes. a fine phrase; not a store.

expect          a SINGLE CLAIM in a single instance. RESOLVES ONCE, is scored, is done.
                carries no running economy. IT IS THE MEASUREMENT.

edge            {expressions} → world state.  P = P(W | E), an EMA over counted outcomes.
                IT IS WHAT LEARNS. its source is a SET — cardinality gives you joints free.
                there is no "ledger": the economy is P/N/F on edges.

grade           log(p_claim) − log(p̂_edge), at the outcome that occurred, × channel fidelity.
                the being is scored on BEATING ITS OWN BASE RATE. reciting the edge earns zero.
```

**`A` and `S` are dead.** `A` was *"coherence of node→framing, settle-written"* — a judge's opinion,
which is the thing v3 exists to delete. `P` takes the same slot in the arithmetic but is computed
from **counted outcomes**. `S` belonged to the settle's economy and nothing asks for it back.

### Banned

`two graphs` / `expression graph` / `observation graph` as *separate durable stores* (say **the
memory graph**, with **observation / state change / action** nodes) · `ledger` as a noun for a store
(say **the edges**) · `A` / `alignment` as an edge field (say **P**)

### The boundary runs through the graph, by node type

`MINE` vs `THE WORLD'S` is not a split between two graphs. It is a property of node type:

```text
observation            THE WORLD'S    reality returned it
state change, action   MINE           I generated them
```

A grade is a cross-boundary diff, so it must span from an **action** node (MINE) to the
**observation** it produced (THE WORLD'S). A "grade" that never reaches an observation node is a
type error — this is LAW 2 as a typing rule. Grading a state change against an action is
MINE-against-MINE: forbidden by construction, not by discipline.

The loop is **circular** through this one graph: action nodes shape which observation nodes ever
exist (you only observe what you acted to observe), and observation nodes grade the action nodes
that reached for them. Neither node type is prior; none is disposable.

## Two clocks

They measure different things and must not be confused.

```text
AGENT TICK   SUBJECTIVE. advances on COGNITION.
             → `within` (bet maturity)   "how much have I thought since I bet?"
             → mark decay                "how stale is my own search?"

WORLD TICK   OBJECTIVE.  advances on THE WORLD.
             → world-process timing      "the build takes 30s whether or not I think about it"
```

If the being idles for an hour and thinks nothing, **nothing has happened to it**. A bet does not
mature because time passed; it matures because the being has *thought* and still does not know.

And a mark decays on the clock of **whoever laid it** — a trace records *"I searched here,"* which is
a fact about that agent's history, not the world's. Staleness protection does not come from decay
anyway: **it comes from perception.** If the being needed to know, it looked.

## The cycle

Written from **observation**, because the being does not begin by wanting something — it begins by
being told something.

```text
observation  ->  projection  ->  expression  ->  observation
 (material)        (final)       (efficient)
```

The four causes are **positions in this loop**, not annotations anyone emits. The causal graph is
not a layer over the knowledge graph — it is **the shape of the loop through it**.

Written this way the hard arrow comes first: `observation → state change` is the one edge with the
entire cognitive act inside it. The other two are cheap.

See [the-cycle.md](the-cycle.md).

## The laws

1. **Observation is the only writer of world state.** Expression has no write access. The being
   cannot believe its own output into reality. World state is a *fold over the observation nodes* —
   deterministic, replayable, auditable, fiction-free.

2. **Learning happens only across the boundary.** `MINE` (projections, actions) vs `THE WORLD'S`
   (observations). A grade is by definition a cross-boundary diff, so **grading your own output is a
   type error**, not a rule anyone has to remember.

3. **Exact first, always. The model only ever sees the residual.** Node resolution, the diff,
   retrieval, and impulse handling all run an exact pass first and hand the model only what's left.
   This bounds cost *and* keeps the model out of the trusted path.

4. **Interpretation happens at observation time, never at grading time.** The decomposer translates
   an impulse into triples on the way *in*; the grader then does a pure exact match. **The model
   cannot touch the grading path at all** — there is no grading-time aligner to misbehave.

5. **Every act must be falsifiable, and every bet gets settled.** Every `act` appears in at least one
   `expect`, and every `expect` carries a `within`, or the emission is rejected by the parser. At
   maturity the substrate perceives the claimed triple — *the bet manufactures its own evidence.*
   (Constructive form of
   [prediction-falsifiability-invariant](../laws/invariants/prediction-falsifiability-invariant.md).)

6. **Observation is produced blind to desire.** The pass that turns an impulse into observation
   triples must not see the open bets, their confidences, or the target — only the entity and
   predicate vocabulary. Enough to speak the right language; not enough to know what answer was
   wanted. Otherwise the being resolves ambiguity toward its own bets, and manufactures the evidence
   that grades it.

7. **A mark may be perceived as a fact about itself. It may never be evidence for a claim about
   anything else.** The world holds *facts* (what is true) and *marks* (what agents wrote there —
   traces, targets, notes). *"A wave of mine passed here"* is true and perceiving it is memory. *"So
   food is in the kitchen"* is laundering. A mark can steer where you look; it can never settle what
   you claimed. Enforced mechanically: `Observation.Fidelity()` returns **0** for a mark.
   See [world-agent-v2](../world-agent-v2/README.md).

## The parts

| doc | what it settles |
|---|---|
| [the-cycle.md](the-cycle.md) | one graph / three node types, the write rule, the boundary, what "causal" means now |
| [expression-grammar.md](expression-grammar.md) | what the LLM emits — triples, modes, guards, joint outcomes |
| [closure-and-planning.md](closure-and-planning.md) | when a plan is done, holes, termination, the watcher |
| [retrieval.md](retrieval.md) | the reverse index; how the grade changes what gets recalled |
| [scoring.md](scoring.md) | grading — the edge as P(W|E), the EMA, beat-the-base-rate, the diff |
| [the-wave.md](the-wave.md) | the two waves, the plan loop, the kill condition, what the kernel needs |
| [workers.md](workers.md) | the six workers, their contracts, and when they fire |
| [trace.md](trace.md) | visibility: the event trace (board data model), three tiers. NOT BUILT. |

## The shape of a cycle

```text
impulse → DECOMPOSER → world state (LAW 1)
                          ↓
                   load-bearing triple contradicted?
                          ↓ yes                    ↓ no
                   KILL the wave              plan continues
                          ↓
   ┌──── 1  INDEX     precedent for the holes            FREE
   │      2  PRESENT   bidirectional wave, budget=tokens  FREE
   │      3  PLAN      ONE LLM CALL  ← the only cost
   │      4  DRAIN     placer inserts nodes               FREE
   │      5  INDEX     exact precedent → place            FREE
   └───── 6  closed? → EXPRESS   fixpoint? → UNRESOLVED

EXPRESS:  MOTOR (no model, no present) → Outcome → observation → GRADER
```

**~2 model calls per cycle**, and the planner's count is the learning curve: `N → 1`. The floor is
one — the irreducible call is the **framing**, not the plan. Precedent can supply a plan; only the
model can judge that the precedent applies. Nothing is executed on precedent alone.

*(Deferred: reflexes — a very probable path may eventually skip the framing call. Explicitly out of
scope. The floor stays at 1.)*

## Prior art in this repo — read before writing anything

v3 is not new. Two documents already contain its core, written 2026-06-28/29 and never built:

- [prediction-falsifiability-invariant.md](../laws/invariants/prediction-falsifiability-invariant.md)
  — *"A prediction that can't be wrong is a vibe, not cognition."* **Still binding. Not superseded.**
- [world-assertions-and-policy.md](../cognition/world-assertions-and-policy.md) — the S-R-O triple,
  the observed/expected split, alignment as **graph overlap not an LLM vibe-check**, and the
  action→observation "gradeable hinge." **This is v3's representation, already correct.**

Two constraints from that doc that v3's design did **not** independently rederive, and must honor:

- **Decomposition is async and buffered.** Text → triples is slow (model calls + entity resolution)
  and **must not block the cognitive loop**. Producer/consumer: the decomposer fills the observed
  graph; cognition runs on whatever is resolved; late edges arrive next cycle.
- **Preserve low-confidence ambiguity.** *"Forcing a hard link too early corrupts the graph
  permanently."* The decomposer must keep ambiguous references as **low-weight candidates**, not
  hard-link them. Entity resolution is **the floor** — relevance-correctness ≤
  resolution-correctness.

v3's world-addressable ID namespaces (`file:`, `proc:`, `url:`, `coord:`) are the direct answer to
that floor: identity that comes from the world costs no resolution at all.

## Superseded by v3

- The annotation grammar (ordinals, verb-group braces, edge marks) — causality is no longer asserted.
- `CausalNodes`, `AitiaSet`, `validateStructuralOp`, `expandCausalBeliefs` — the causal layer dissolves.
- The coherence settle as the writer of alignment — alignment falls out of the diff.
- The static workflow DAG, the board's join / fan-out / motor guard / one-reframe rail — coordination
  for a pipeline that no longer exists.

## The measurement — write it down before building

```text
Over N repeated exposures to one task class:
  passes-to-closure                            decreases
  prediction error                             decreases
  fraction of expects served by the index      increases
```

Both curves down is the thesis. **Either alone is a warning, and either failure localizes:**

- passes ↓ but error flat → retrieval is surfacing confident garbage; the grade weighting is wrong.
- error ↓ but passes flat → the world-model learns but the index doesn't.

## Open

- The wave algorithm (next).
- Concurrent descent with a deterministic replay mode — required by rewind
  (`scratchpad/2026-07-13/commit-structure-and-rewind.md`), and the same feature the world's
  fan-out needs (`one-kernel.md`).
- Adaptive shock threshold — a being in a novel world is surprised constantly; a fixed threshold
  pins it in the slow path forever.
- Arbitration between two closed plans competing for the act budget. Probably the only place the
  amplitude/superposition machinery from `scratchpad/2026-07-13/README.md` is needed.

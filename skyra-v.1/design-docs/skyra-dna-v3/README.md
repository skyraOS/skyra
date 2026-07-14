# skyra-dna-v3

**Status:** DESIGN — nothing built yet. v2 will be archived once v3 stands.
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

## The cycle

```text
projection  ->  expression  ->  observation  ->  projection
   (final)      (efficient)      (material)
```

The four causes are **positions in this loop**, not annotations anyone emits. The causal graph is
not a layer over the knowledge graph — it is **the shape of the loop through it**.

See [the-cycle.md](the-cycle.md).

## The laws

1. **Observation is the only writer of world state.** Expression has no write access. The being
   cannot believe its own output into reality. World state is a *fold over the observation log* —
   deterministic, replayable, auditable, fiction-free.

2. **Learning happens only across the boundary.** `MINE` (projections, actions) vs `THE WORLD'S`
   (observations). A grade is by definition a cross-boundary diff, so **grading your own output is a
   type error**, not a rule anyone has to remember.

3. **Exact first, always. The model only ever sees the residual.** Node resolution, the diff,
   retrieval, and impulse handling all run an exact pass first and hand the model only what's left.
   This bounds cost *and* keeps the model out of the trusted path.

4. **The aligner may translate; it may never judge.** It runs **once, on the observation side
   only**. Predictions are literal stated claims and are never interpreted.

5. **Every act must be falsifiable.** Every `act` appears in at least one `expect`, or the emission
   is rejected by the parser. (This is the constructive form of
   [prediction-falsifiability-invariant](../laws/invariants/prediction-falsifiability-invariant.md).)

## The parts

| doc | what it settles |
|---|---|
| [the-cycle.md](the-cycle.md) | two graphs, the write rule, the boundary, what "causal" means now |
| [expression-grammar.md](expression-grammar.md) | what the LLM emits — triples, modes, guards, joint outcomes |
| [closure-and-planning.md](closure-and-planning.md) | when a plan is done, holes, termination, the watcher |
| [retrieval.md](retrieval.md) | the reverse index; how the grade changes what gets recalled |
| [scoring.md](scoring.md) | information gain, contradiction vs incompleteness, surprise shock |
| [the-wave.md](the-wave.md) | the two waves, the plan loop, the kill condition, what the kernel needs |
| [workers.md](workers.md) | the six workers, their contracts, and when they fire |

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
  permanently."* The aligner must keep ambiguous references as **low-weight candidates**, not
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

- passes ↓ but error flat → retrieval is surfacing confident garbage; the ledger weighting is wrong.
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

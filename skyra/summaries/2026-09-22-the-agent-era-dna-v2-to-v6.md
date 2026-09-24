# Skyra: the agent era, from DNA v2 to v6

**Date:** 2026-09-22
**Reviewed:** four folders on main under `skyra/archive/`:
- `user-agent/`
- `distributed-agent/`
- `cognitive-agent/`
- `_skyra-dna-v2/`

For context I also read the interfaces of the kernel they all share,
`skyra/archive/spacetime/`.

**Span:** about 2026-05-31 (first `UserCell`) to 2026-08-19 (the v6 CTU pivot spec).
Everything was archived together in `71a4545` on 2026-08-20 ("archive superseded Skyra
generations and scaffold v7").

**Scope:** about 30k lines of Go:
- v2 is about 9.8k lines;
- v3 is about 11k lines;
- v4, v5 and v6 together are about 7.5k lines;
- the worlds are about 2.9k lines;
- the user cell is 145 lines.

I read the core files in full. For the larger v2 and v3 support files (the settle,
budgets, activation field, prediction fields, TUI) I read signatures and doc comments.
For the `logs/` run output I read only the two `report.txt` files. I skipped the design
docs.

**Health:** **none of this builds as it sits.** Every import points at
`skyra-v1/src/skyra/...` paths that no longer exist. v2 is also in a `_`-prefixed
directory, so the Go toolchain skips it. There are about 250 test functions:

| package | tests |
|---|---|
| v2 | 87 (plus 2 in `cmd/tui`) |
| v3 | 104 |
| v4 | 17 |
| v5 | 21 |
| worlds | 24 |

I did not restore the import paths to run them.

---

## 1. What this era is

Before v.05 and the DNA-series, Skyra was a **physics-shaped runtime with agents mounted
in it**. The organizing idea is in the three archive READMEs: an agent falls into one of
three categories, by **what it owns**.

| folder | category | owns | does not own |
|---|---|---|---|
| `user-agent/` | human | nothing but a faithful record of what crossed the boundary | cognition (the human supplies it) |
| `distributed-agent/` | world | topology and authoritative live state | grading ("the world REPORTS; it never grades") |
| `cognitive-agent/` + `_skyra-dna-v2/` | being | observations, predictions, learned edges | the world it stands in |

The key design principle is that **a world reports and a being grades, and neither may
do the other's job**. Almost every rewrite in this span is an attempt to make that
split airtight.

The genome goes through five generations. Each one exists because the one before hit a
ceiling:

```
v2 (07-03)  LLM genes on a workflow DAG; causal graph asserted by the model; LLM judges grade
  │         └ ceiling: the grade is an opinion with no independent source
v3 (07-14)  expect triples graded by EXACT DIFF against world observation; no judge anywhere
  │         └ ceiling: triples + a filesystem world don't generalize to open-ended asks
v4 (08-03)  "every node is an expectation, every edge is an observation": a filesystem-shaped graph
  │         └ built the graph + handles; behaviors mostly stubbed
v5 (08-11)  kernel/graph split, engrams, two-form grammar (select | commit), credit by citation
  │         └ ran live on Codex: 1/5 then 3/5 asks closed
v6 (08-19)  the CTU pivot: interfaces only. Kernel protocol = instruction set
            └ archived next day; v7 / the DNA-series takes over
```

## 2. The shared kernel (`archive/spacetime/`, about 2.1k lines)

These packages all implement one contract. It is small enough to state:

```go
type Space interface {
    ID() string; Slice() Slice; NewActivation() ActivationRecord
    Observe(origin Space, impulse any)        // descent: do the work
    Express(origin Space, impulse any) any    // ascent: report
    Agents() map[string]Space; Children() map[string]Space; Content() any
}
type Time interface { Realize(origin, traversed Space, impulse any) } // run one wave
```

- **An `Impulse`** is a message with these fields:
  - `Primitive`: perceive, engage, move or wait;
  - `TargetID` and `Payload`;
  - `Source`;
  - `Return`, which marks a reafference (the world answering your own act);
  - `Amplitude` and `ResultMap`.
- **`relation.Relation`** is the wave runner. Amplitude starts at 1.0 and decays by 0.9
  at each level of depth. Because amplitude is also split among children, deep
  breadth-first searches die out fast, and both worlds work around this by listing
  cells within the radius directly.
- **A `Slice`** holds each being's working state: a lock, counters, `Maps()`, and a
  stress value. "Active state lives on the slice as IDs, never pointers" is a
  convention every generation keeps.
- **`traverse.Assemble`** walks a graph under a token budget, giving each child a share
  of the budget proportional to its weight. v2 used it to build prompts.
- **`kernel/`** is the later v4 form: `Perturb` / `Observe(realized)` handles, stacks,
  a runner, and an eviction hook (`EvictOldest` is still a stub that returns 0).

## 3. User agent: `UserCell` (145 lines)

`user-dna/user_cell.go`: the human is represented as a being that does **no
inference**.
- It records every inbound impulse (from the world) and outbound impulse (authored by
  the user), time-stamped by the slice clock.
- It learns which entities exist from the senders it sees.
- It can be seeded at boot (for example with skyra placed next to michael).
- An `OnInbound` hook lets the TUI show replies.

Every other `Space` method is a no-op. The README describes it as "the foundation any
user-modeling layer reads from later". No such user-modeling layer was ever written.

## 4. Distributed agents: the worlds (about 2.9k lines)

The name looks ahead. Both worlds run locally and in-process. They are "distributed"
because they **own authoritative state on the far side of a boundary**.

### `world/`: the flat world
- **Size and range:** a 1000×1000 wrapped grid. Perceive, engage and move all have a
  range of 20.
- **Primitives:** handled in `Observe`:
  - `perceive` sends a wave to every `Point` within range;
  - `engage` is range-gated and returns `EngageBlockedResult` when out of range;
  - `move` can target `"x,y"` or an agent id, and is blocked by walls;
  - `wait` arms a timer that fires a return wave.
- **Results** only ever come back as `Return: true` impulses into the origin's
  `Observe`.
- **Occupancy is a set** and there is a separate identity registry. The comments explain
  why: when one agent could overwrite another on a cell, the overwritten agent became a
  "ghost", present but invisible to perception and delivery. `TestOverlapKeepsAllOccupantsAddressable`
  guards against this.
- **The maze:** a 20×20 recursive-backtracker maze at (478,478), with `entrance` and
  `exit` marker agents. The exit congratulates the agent and sends it back to the
  start.

### `world3d/`: the globe
- **Shape:** a **cube-sphere**, made of six 12×12 faces projected onto a unit sphere.
  Adjacency across face edges is **computed from the geometry** ("step past the edge in
  3D, re-project") rather than from a hand-written table, which avoids the classic
  wrap-around bugs.
- **Movement takes time:** moving is **a walk, not a teleport**. The agent advances one
  cell per `Tick()`, and the return wave fires on arrival, so distance costs time.
- **Task furniture:** a shrine, a sigil and a beacon. The beacon lights only for an
  agent standing on it while carrying the sigil. A failed attempt reports the **first
  missing prerequisite**, which is the same one-hint-at-a-time pattern as v3's hidden
  build recipe.
- **A "snitch":** a catchable agent with an 8-cell reach. The first catch wins,
  guarded by a lock.
- **Events** narrate the world to a frontend. No being ever sees them.

## 5. DNA v2: the genome as a workflow engine (`_skyra-dna-v2/`)

**Model.** Everything is written in terms of biology:
- **the DNA:** one `DNA` cell that holds every durable index plus every "gene";
- **genes:** stateless `ObserveCompetence` functions, keyed `namespace:gene`;
- **mRNA:** each wave runs on a disposable `EphemeralRoot` transcript, so a concurrent
  wave cannot overwrite another wave's state;
- **the membrane:** an orchestrator that routes on a **static DAG**
  (`workflow.go`). A per-impulse **Board** row holds the join, a "one reframe per
  impulse" limit, and a motor guard that fires exactly once.

```
sensory:intake ─external→ final:grade ─held/no-prior→ midlevel:frame ─framed→ ┬ final:predict ─┐
     │                        └ all-preds<0 → reframe (once)                 └ efficient:predict ┴join→ motor:collapse → emit
     ├ reafference → efficient:grade (exec=no → reframe)
     └ perceive-return → sensory:ingest → midlevel:frame
```

**The steps.** Every step below except motor is an LLM call:
1. **`sensory:intake`** records the interaction and **breaks the utterance down into an
   annotated sentence**, such as `michael[1] {asked about}[1→2] {the outage}[2]`. It
   stores the result as an observed `CausalNode`, then grades every active cause
   against it (an LLM grade from −1 to 1).
2. **`midlevel:frame`** chooses `<retain/>`, `<compose>` or `<requery>` over the
   Material/Formal structure. It then expands a full **Aitia set**: one present node
   plus one Efficient and one Final causal edge, each produced by its own LLM worker
   with its own **AIMD token budget**. Finally a **coherence settle** batch-grades the
   episode's nodes against the fixed frame, writes the scores as alignment `A`, prunes
   anything below zero, and repeats until the average stops changing.
3. **`final:predict`** emits a calibrated distribution of what the other party will say
   next. **`efficient:predict`** proposes Skyra's next move. Both are written onto the
   current frame.
4. **`motor:collapse`** turns that proposal into exactly one of
   `<engage>`, `<move>`, `<perceive>` or `<wait>`, retrying on parse failures.

**Learned state.** Three stores sit side by side:
- `ActivationStore` holds A·S·F·N per edge; strength is `rect(A)·S·F·conf(N)`.
- `SignedStore` holds a slow Identity EMA and a fast Familiarity EMA, both signed.
- `QueryBudgetStore` holds one AIMD budget per voice (Efficient and Final).

**Supporting code:**
- `praxis_inspection.go` gives a read-only view of the whole causal structure.
- `trace.go` writes one JSONL file per impulse, recording every inference in full. It
  was added after the first live run (07-13) produced "ten calls, zero recoverable
  outputs".
- `cmd/` holds the harnesses: `tui` (bubbletea, the maze, with michael as a
  `UserCell`), `validate-v2`, and `phase1-ab` / `phase2-accumulate` (the first measured
  learning curve).

**Why it ended.** The README quotes v3's `doc.go`: *"Free-text predictions could only be
graded by an LLM judge, so the learning signal bottomed out in an opinion… Causality had
to be ASSERTED by the model… the goal was re-derived every turn."*

The code agrees:
- Closure grading (`final:grade`, `efficient:grade`) is itself a pair of LLM judges.
- Their link to `A` was **explicitly disconnected on 07-06**, leaving the LLM coherence
  settle as the only thing that writes `A`.
- Cause assessment is a third LLM grader.

Every learned number in v2 traces back to a model's opinion.

## 6. DNA v3: the grade is computed, never judged (`skyra-dna-v3/`)

The strongest generation in this span. `doc.go` sets out six laws:
1. **Only observation writes world state.** World state is a fold over the observation
   log, and expression cannot write it.
2. **Learning happens only across the boundary** between the being and the world, so
   grading your own output is a type error.
3. **Exact matching first.** The model only ever sees what's left over.
4. **Interpretation happens when an observation arrives, never when a bet is graded.**
   The grader is a pure exact match.
5. **Every act must be falsifiable, and every bet must be settled.** An act needs an
   `expect`, and every `expect` needs a `within`.
6. **Observation is produced blind to what the being wants.** The pass that turns an
   impulse into observations can't see the open bets.

**The loop** (`being.go`). `Run(target)` repeats plan → act → observe → re-plan until
the target is *observed* true:
- **Plan.** One model call per pass emits `<framing>`, `<act id target primitive
  requires=…>`, and `<expect of=… > p within subj pred obj </expect>`. Planning ends
  when every `requires` is either observed or promised by another act's `expect`,
  checked **by set membership, with no judge**. A "no progress" check stops the loop if
  the list of gaps stops shrinking.
- **Express.** The being fires every act whose requirements are met, then **Settle**:
  every bet that has come due gets a *targeted perceive*, which is what makes absence
  decidable. Newly observed facts unblock guarded acts, and the cycle repeats (up to 8
  rounds).
- **Grade.** Arithmetic on the observation write path: the being's stated probability
  is scored against the Edge's `P(W|E)` as
  `log p_claim − log p̂_edge`, times the mode's weight (exact 1.0, interpret 0.3).
  **Repeating the edge's own probability earns exactly zero.** An edge nobody has
  measured is graded against 0.5. `TruthUnknown` never moves an EMA.
- **Mode is set by the world.** The `Affordances` registry stamps each claim's Mode from
  what the world can actually measure, so the model can't pick its own grade
  multiplier.
- **The world is a Realm.** v3 ships a sandboxed filesystem with `write`, `exec` and
  `build`. The build has a hidden three-file recipe (`app.src`, `app.conf`,
  `app.lock`) and reports the first missing file. The idea is that a cold being
  discovers the recipe one file at a time, and a warm one closes in a single pass.
  **That is the learning curve being measured.**
- **Metrics:** passes to closure, and **"edge over base rate"**. The report says it
  outright: *"near zero means the model is PARROTING."*

**What grew on top** (07-14 to about 08-01):
- **observe-wave:** resolves an impulse's roots against classes, with the classes derived
  from gap-segmented spans and log(N/df)-weighted co-occurrence (union-find);
- **transient working memory:** `ActivationField`;
- **token-density memory selection** (`SelectMemory`);
- **interpretation claims** with their own separate way of being proven wrong, which
  keeps action success from crediting interpretations backwards;
- **Phase-4F prediction fields:** per-dimension deterministic parsers, blind calibration
  windows, and PMI fusion.

It has 104 tests, including adversarial ones and a stress test.

## 7. DNA v4: the expectation graph (`skyra-dna-v4/graph/`, about 1.5k lines)

The motto was *"Every durable node is an expectation; every edge is an observation."*

- **Nodes are filesystem-shaped.** A node with `Entries` is a **directory**, which makes
  it an expectation. A node with `Act` is a **file**, which makes it an action. The kind
  is derived from what the node contains and never stored.
- **Execution never touches durable nodes directly.** A `Namespace` is the agent's root.
  A `Handle` is an open file description: it pairs a node, a behavior, and the state of
  one particular use. A `Walk` is a frontier plus a fold.
- **Behaviors:**
  - `turn` sequences the phases;
  - `interpret` writes only edges, never nodes;
  - `expect` is the only thing that writes nodes;
  - `pursue` is **stubbed**.

It isn't finished. EXPECT runs ungated and proposes placeholders named after the
impulse. The gate is recorded but not enforced, so it could be calibrated later. The
`Node` comment admits that "an unfalsifiable expectation is unconstructible" is not
enforced yet, because there is no constructor.

## 8. DNA v5: kernel, engrams, and credit by citation (`skyra-dna-v5/`, about 5k lines)

This is the first generation where the **kernel holds everything that is happening and
the graph holds only what happened**. That split is what makes three claims true:
- the kernel writes nothing;
- an engram is only minted once it closes;
- a crash leaves nothing to recover.

**The pieces:**
- **Engram:** a fixed `Expectation` plus the child engrams it owns and an append-only
  history of invocations. **It never reconsolidates.**
- **Possibility:** one ephemeral opening of an engram, run by an **entity**. The
  entities are the Codex CLI, the runtime's built-in discover / by-index / by-recency /
  create, and scripted interpreters.
- **Grammar:** exactly two forms. `<select>` is an invocation group, and how many
  members it has declares the concurrency. `<commit>` is the only durable write.
  **Citation is derived** from the `<obs-N>` refs inside a commit, and credit flows
  along those citations. The kernel checks a select or commit against the frame
  structurally, before anything runs.
- **Root:** a transaction boundary. World effects are **staged** and only happen along
  the chain the accepting commit cited. Anything else is **discarded, and the discard is
  recorded as an event**.
- **Frame:** a possibility sees *down, never up*. It sees what its children concluded
  but not what they cited, and bounds on depth and steps evict runaways.

**The live run.** The logs from 2026-08-11 show five asks run on Codex:

| run | closed | what happened |
|---|---|---|
| first | 1/5 | `summarize-again` couldn't name its own prior engram; `arithmetic`, `decompose` and `impossible` were each evicted after 9 steps without a commit |
| second (after the fix commit "four repairs the model run forced") | 3/5 | `arithmetic` and `impossible` still evicted |

The graph from the first run shows `system:create` invoked 85 times, 82 of which
returned `error:str`.

## 9. DNA v6: the CTU pivot (`skyra-dna-v6/`, about 630 lines, interfaces only)

**This is a specification written as Go types. None of it has an implementation.**
- **`Kernel` has five methods:**
  - `Approach` (the boundary);
  - `Think` (bounded potential that performs nothing);
  - `Decide` (choosing a candidate is binding it; Reach happens inside);
  - `Remember` (receive a return, with Recognize as its first act);
  - `Orient` (live topology only).
- **Engrams** hold `Contracts` for each child position (an ordered, typed `Signature`
  plus a `Return` type), a compiled `Continuation`, and `Compiler` and `Creator`
  **addresses**.
- **"The kernel protocol is the instruction set":** a compiler (an LLM, python, and so
  on) parses its own syntax and calls the kernel.
- **A creator** is reached only when compilation fails.
- **`Horizon`** can only narrow, so a child's horizon is always inside its parent's.

The comments list open holes rather than hiding them:
- `Decide` doesn't say whether it returns when the child is live or when the child
  returns;
- `Credential any` is "a named hole";
- `Approach` authenticates nobody;
- `Branch.Actions` is a map, so the semantically meaningful emission order is lost.

The spec is dated 08-19. The code was archived on 08-20 when v7 was scaffolded.

## 10. Things I found reading the code

I did not run any of this; all of these come from the code alone.

1. **v3's sandbox doesn't hold for `exec`.** `World` claims *"Nothing outside Root is
   readable or writable."* `resolve()` does keep `write`, `perceive` and the exec
   *working directory* inside Root. But `exec` runs `sh -c <payload>`, and the payload
   can use absolute paths or `cd /`, so a model-authored command can read or write
   anywhere the process can.
2. **v2 never escaped the LLM judge.** After closure grading was disconnected on 07-06,
   every number that writes to `A`, `SignedStore` or the budgets comes from an LLM grader
   (the coherence settle and cause assessment). The v2 README says as much; the code
   confirms it.
3. **v2 runs every wave on its own goroutine** (`go o.kernel.Realize`). The CLI waits on
   a flush callback with a 10-minute timeout. If a wave hits a route that isn't in the
   DAG, it just stops without calling `finish`, and the turn hangs until the timeout.
4. **v4 behaviors are mostly scaffolding:** `pursue` is stubbed, `expect` emits
   placeholders, and the node invariant isn't enforced.
5. **In v5, authorization is only half built.** `authorize.go` says the kernel checks
   authority "TWICE, at two boundaries", and only the first check exists. `Ingress` is
   deliberately ungated ("any caller is an ingress channel").
6. **The v5 live runs mostly fail.** The main failure modes are eviction when a
   possibility never commits, and `create` failing repeatedly. Only asks with an obvious
   single-step answer closed.
7. **Nothing here builds anymore.** Restoring any generation means rewriting its
   `skyra-v1/...` imports and pulling `archive/spacetime` back into the module. The v2
   README gives the `git mv` steps for v2 only.

## 11. What carried forward

In rough order of how much weight they carry in the current work:
- **A world reports and a being grades**, and no grade may pass through the being's own
  output (v3 laws 1 and 2). This becomes the DNA-series' authority boundary.
- **Grades come from the world, and they measure how far the being beat its own base
  rate** (v3 `Edge`). This is the answer to v2's LLM-judge ceiling.
- **The kernel holds what is happening and the graph holds what happened.** Commits are
  the only durable write, and world effects are staged and happen only along citations
  (v5). This is the ancestor of the transaction and ownership models in the DNA
  summaries from 08-11 onward.
- **Down, never up:** a possibility sees its children's conclusions but never their
  citations or its own reception (v5 frames, and v6's "observing reception is refused").
- **Some ideas were left behind:** the biology vocabulary (DNA, genes, mRNA) and the
  static workflow DAG from v2, the filesystem-shaped nodes from v4, and the
  hand-built physical worlds, whose job the DNA-series moves behind an external binary
  boundary.

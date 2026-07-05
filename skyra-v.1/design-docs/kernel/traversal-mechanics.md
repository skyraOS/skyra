# Traversal Mechanics

**Status:** design, not built. Evolves `activation-and-traversal.md` — it **removes the
`Activate` seam**, folds activation into `Observe`, and pushes *all* cognitive policy out
of the kernel into the cells. Where the two disagree, this doc wins on the
bookend/activation/traversal contract.

## The move, in one line

Delete `ActivationRecord.Activate`. Make `Observe` **return the cell's onward
distribution** — which edges light up and, relative to each other, how much of the wave's
budget flows down each. The cell owns *all* the cognitive policy (which children, how much
to each, how broad); the kernel becomes **pure transport** — it carries the budget the
cell assigned, checks each node can be afforded, recurses, and folds back. Every cognitive
constant that used to live in the kernel (the split rule, the decay factor, the threshold)
moves into `Observe`.

## Why `Activate` is the wrong seam

Its signature can only see two things:

```go
Activate(visited bool, traversalCount int) float64
```

`visited` and `traversalCount` (plus whatever durable strength sits on the record). It is
**blind to the impulse and to the node's own content.** So it can compute "this edge has
fired a lot, so it's warm" — and *nothing* impulse-relative. But the half of activation
that carries cognition — *which edges light up **for this impulse*** (attention, relevance,
location-based recall) — is exactly the half it can't reach, because the impulse was never
handed to it.

`Observe(origin, impulse)` sees everything: the origin, the impulse, and (it is a method on
the node) the content and the durable edge-strength. So activation belongs *in Observe*.
Put differently: **`Activate` was a crippled `Observe` return** — the seam was cut blind to
the two inputs that make activation mean something.

## The two bookends, as unfold / fold

Making Observe return sharpens the bracket into a clean dual. A wave **expands on the way
down and reduces on the way up**:

- **`Observe` — descent, pre-order — the unfold.** Returns the onward distribution: a map
  of edges with relative weights. One node fans out to many children → a *map* is the
  natural shape. This is the frame forming.
- **`Express` — ascent, post-order — the fold.** Returns the rollup: a reduction over the
  children's expressed results plus the node's own content. Many children collapse into one
  result for the single parent → a *value* is the natural shape.

Down = map (fan-out). Up = value (fold-in). It is a map/reduce — an anamorphism on descent,
a catamorphism on ascent. Symmetric as *duals*, not as "both return maps."

## The kernel carries no cognition (the throughline)

The tell that something was off: the traversal wanted **two models** — a focused deep walk
and a broad recall walk (see "Drill and gather") — and to support both, the *kernel* would
have to know which. But the kernel is the **void kernel: pure transport.** Two models in
the kernel means cognitive priors had been baked into transport. Audit of what the kernel
currently decides:

**Cognitive — moves into `Observe`:**
- the **split rule** (even `/len`, product-weighted, broad, narrow) — a prior about how
  relevance distributes across children.
- the **decay factor** (`amplitudeDecay = 0.9`) — a hardcoded "how fast relevance falls
  with distance." **Deleted outright:** decay now happens because *traversing costs tokens*
  — the budget depletes as it flows, so no separate decay constant is needed. It was pure
  leak.
- the **threshold** — a relevance-floor constant.
- **breadth / order** (drill vs. gather) — cognitive.
- the **units** the budget is denominated in (tokens? distance?) — set by the world's
  *cost function*, not the kernel.
- **conservation** — whether a branch's budget is split among children (bounded) or
  broadcast full to each (unbounded reach) — set by the cell's *weights*, not the kernel.

**Universal — stays in the kernel:**
- **deliver a budget along an edge + recurse** — transport.
- **affordability** (`budget ≥ cost`) — pure arithmetic, no prior.
- the **visited-once rail** — *termination correctness*, not cognition.
- the **fold** (`Express` post-order) — structural.
- **bookkeeping** — the proper-time tick, the double-buffer.

After the move the kernel holds **zero cognitive constants** — and doesn't even know what
the budget *measures* or whether it's *conserved*. All of it is in the cells.

## The interface change

`src/skyra/space.go`, current:

```go
type Space interface {
    ...
    NewActivation() ActivationRecord
    Observe(origin Space, impulse any)
    Express(origin Space, impulse any) any
    Agents() map[string]Space
    Children() map[string]Space
    Content() any
}
type ActivationRecord interface {
    Activate(visited bool, traversalCount int) float64
}
```

Target:

```go
type Space interface {
    ...
    // Observe fires descending. It returns this cell's ONWARD DISTRIBUTION: the
    // edges it lights for this impulse, each with a RELATIVE weight, keyed by edge
    // identity (A→B vs C→B — same key the slice records strength off of). nil = leaf.
    // The cell owns the policy (which edges, weights, how broad); the kernel scales
    // the weights by the budget it's carrying and delivers.
    Observe(origin Space, impulse any) map[string]Alloc
    // Express fires ascending. It returns the fold: this node's rollup of its
    // children's expressions plus its own content.
    Express(origin Space, impulse any) any
    Agents() map[string]Space
    Children() map[string]Space   // structure (see below) — no longer walked by the kernel
    Content() any
    // Cost is the node's weight against the budget, in the budget's units. STAMPED
    // at birth and on mutation (not recomputed per wave). Memory: token count of its
    // content. World: 1 (a Chebyshev step). The kernel subtracts it and checks
    // affordability; it never computes it.
    Cost() int
}

// Alloc is one entry in the onward distribution: a child and its RELATIVE weight
// (a share, not an absolute budget — the cell need not know its own budget).
type Alloc struct {
    Target Space
    Weight float64
}
```

`ActivationRecord.Activate` is **removed**. The kernel's job on the return is the barest
arithmetic: given the budget `B` it carries into a node, subtract the node's own `Cost()`, and
scale each `Alloc` — `child_budget_i = (B − Cost) × weight_i`. **No normalization** — whether
the children's budgets sum back to the parent's (conservation) is decided by the cell's
weights, not the kernel (see "Budget is world-defined"). No constants. `Agents()` and
`Content()` untouched.

## Kill the method, keep the state

Deleting `Activate` does **not** delete activation-as-stored-strength. The accumulated edge
strength — the "engaged-often-strengthens" economy, the lagged `Alignment` grades write
back — is durable and lives **on the slice's edge records**, keyed by edge identity, exactly
as `writeAlignment` already does (`sl.Records()[edgeKey]`). Where that state lives is
unchanged.

What changes is *who reads it*. The cell's `Observe` reads the edge strengths and returns them
as the weights. There is **no separate fresh relevance term** — relevance *is* the edge's
stored alignment (see "Credit assignment & node birth"), and the impulse-dependence is
*structural*: the same node is reached via different edges from different roots, each edge
carrying its own alignment. So `Observe` just reads what's already on the edges it traverses.

Because `Observe` runs **lockless** and now *reads* the strength records, the records must be
**lock-free readable** — store strength as an atomic and accept slight staleness (which is
just read-live-never-snapshot again). This is the enabling move for a genuinely void kernel;
without it, the strength read has to stay kernel-side and a little bias re-leaks. (Ties to
the record-type open decision.)

Same discipline as the dnav2 refactor: durable state on the graph, the traversal a stateless
transform over it, read-live-never-snapshot.

## Activation & the bump

The section above says strength lives on the edge record, read lock-free. This one says
*what* strength is and *how the bump maintains it*. Factors are stored **separately**
(multiply at read — cheap, inspectable, reweightable), grounded in the design already in
`memory_cell.go` (`Intrinsic.Value() = ClosureGrade × StressAtBirth` — alignment × salience).

### The record — four factors

| field | is | updated on |
|---|---|---|
| **A** — alignment | signed running value of the outcome (did it close?) — the existing `ClosureGrade` | outcome |
| **S** — salience | arousal when reinforced — the existing `StressAtBirth` | outcome |
| **F** — frequency/recency | EMA of "was I traversed" (recency via decay) | crossing |
| **N** — count | gradings seen — drives the learning rate *and* confidence | outcome |

**This resolves former open #1.** "Crossing vs. outcome" was a false choice created by treating
strength as one scalar. With factors separate, they update on separate events: **crossing → F**,
**outcome → A, S, N**.

### Read-time weight

```
strength = rect(A) · S · read_F(t) · conf(N)      // durable, edge-owned — this IS the weight
```

`rect(A)` maps the signed value to a non-negative weight — a failure-leading edge gets
near-zero weight, never negative budget. There is **no separate `relevance` multiplier**:
relevance *is* `A` (see "Credit assignment & node birth"), and the same node lights differently
per impulse because different roots reach it via different edges, each with its own `A` — the
impulse-dependence is structural, not a fresh computation.

### Update rules

**F — on crossing, at tick `t`** (lazy timestamped EMA — touch only crossed edges, never the
whole graph):

```
decayed = F · λ^(t − last)
F       = decayed + α_F · (1 − decayed)
last    = t
```

Read (non-mutating): `F · λ^(t − last)`. Recency falls out of the decay — no separate field.

**A, S, N — on outcome** (closure grade `g ∈ [−1,1]` from the comparator, being's current
stress `s`):

```
N = N + 1
α = max(1/N, α_floor)        // running-mean while N is small → EMA once floored
A = A + α · (g − A)
S = <knob 1>
```

`α = max(1/N, α_floor)` makes `N` do double duty: fast averaging early, sustained adaptivity
later, and it *is* the confidence signal at read.

### Bootstrap

A new edge (`N=0` → `conf≈0` → `strength≈0`) would never light and so never learn. The escape
is the **fresh `relevance`** term: a new but impulse-relevant edge still lights, gets crossed
and graded, and *earns* its A/S/F. Relevance is the exploration floor.

### Knobs (defaults flagged)

1. **Salience `S`.** Immutable `StressAtBirth` (default — matches the code) vs. a running
   `max`/EMA of arousal across reinforcements.
2. **Confidence `conf(N)` — the explore/exploit dial.** Penalty (trust proven edges) vs. a
   low-`N` *bonus* (UCB-style — explore uncertain edges). *Default a mild exploration bonus* —
   the accumulation experiment needs exploration.
3. **`rect(A)`.** Hard `max(0, A)` (failed edges go silent) vs. soft `(A+1)/2` (stay barely
   alive, recoverable). *Default soft.*

Ship defaults: `α_floor ≈ 0.1`, `λ ≈ 0.98`/tick, `S` immutable, `conf` = mild explore bonus,
`rect = (A+1)/2`.

### Concurrency

The **F-bump is a write on crossing**, inside the lockless traversal — so `F`/`last` need a
lock-free/atomic write (or a batched deferral), the same discipline as the lock-free strength
*reads* already required. `A`/`S`/`N` update on the grade write-back, which is already locked
(`writeAlignment`).

## Credit assignment & node birth

"Activation & the bump" left one thing open: when a cycle grades, *which* edges get the
alignment, and how much? It falls out of two facts about how nodes are born.

### Relevance is alignment

There is no separate "relevance" signal to invent. **Relevance is the edge-owned alignment.**
An edge is relevant to the current context if arriving through it has *closed goals* before —
which is exactly what `A` measures. The impulse-dependence is **structural**: the same node
carries edges from every root that has reached it, each with its own `A`; a new root reads the
`A` on the edge it arrives by. This finishes the location-based-memory bet — attention is "what
has proven useful arriving here," never flat semantic similarity. It requires only that **every
cycle's closure grade is written back to all edges in that cycle's frame**, not just the graded
prediction (grade-fronts-the-stroke already produces the grade; this distributes it).

### Node birth wires the causal graph

When inference produces a new node (a prediction, a framing, an action), it is born already
wired:
- **its children = the active frame** — the lit nodes that fed the inference. The output points
  at its causes.
- **it is a child of the root** that initiated the cycle.

The edges are *created* here — structure, causal certainty, born ≈ 1. Their *value* alignment is
set later by grading. Keep "the edge exists" (structural) distinct from "the edge is useful"
(`A`, graded): relevance rides on the value.

### Credit = the frame's activation profile

On a closure grade `g ∈ [−1,1]`, distribute it to the frame's edges **weighted by activation**:

```
credit_i = activation_i · g        // each active edge gets its share, scaled by the outcome
```

Free — the frame already computed each node's `activation` (its traversal weight). It's an
eligibility trace: the standard principled approximation to credit assignment, not a hack.

### The graph is the eligibility trace

Credit is *late* — Skyra acts one cycle, the reply grades it the next, after the producing frame
is gone. Normally that needs separate bookkeeping. Here it's free: because **a node's children
are its birth-frame**, grading the node later flows credit straight to what produced it. The
wiring laid down at birth *is* the trace. Thinking builds the causal record; grading walks it
backward.

### Caveat: incidental credit

Activation-weighting credits nodes that were active but *incidental* — an always-lit node (high
`F`) collects credit on every closure, contributing or not. The **signed grade** corrects it:
incidental nodes get debited on bad closures too, so anything present on both good and bad
outcomes washes its `A` toward neutral; only nodes *correlated* with closure accumulate positive
`A`. Correlation does the causal filtering — over many samples, so early graphs are noisy. Fine
for a learning system; just not instant.

## Amplitude is a token budget

Amplitude stops being an abstract [0,1] signal. It **is a token budget** — the knob on
context size. The wave carries a budget in tokens; each cell's content weighs its own token
count; a cell pays that cost when the wave includes it, and distributes the **remainder**
down its edges by the weights its `Observe` returned. When the budget is spent, the wave
stops. The frame is therefore bounded to the context window *by construction* —
working-memory assembly falls out of the walk instead of being a separate compaction pass.
(This is the coherence-managed-context idea, phase-06, as a mechanism; it makes the ~100k
recall ceiling a number you set.) There is no separate decay: **token cost is the decay.**
(Tokens are the *cognition* denomination; "Budget is world-defined" below generalizes the
units and conservation to any world.)

### The four-part model

The old tangled `score` splits into four orthogonal things:

| part | is | owned by |
|---|---|---|
| **candidacy** | which cells are in the running (impulse-relevant) | `Observe` |
| **distribution** | the relative share each branch gets — the `Alloc` weights | `Observe` (reads strength + impulse) |
| **cost** | the node's `Cost()`, stamped at birth/mutation — tokens for memory, 1 for world | the Space |
| **budget** | the amount flowing down, capped at the root | the wave (kernel carries it) |

The first two are the cell's; the kernel only does the arithmetic on the last two.

### Worked example

Root carries 5000. Its `Observe` lit three children with weights derived from
`product(alignment, ema, …)`: A `0.72` (cost 500t), B `0.55` (100t), C `0.86` (400t),
`Σ weight = 2.13`. Root pays its own cost (say 0), remainder 5000:

| cell | weight | share | budget delivered | cost | carries to its kids |
|---|---|---|---|---|---|
| A | 0.72 | 33.8% | 1690 | 500 | 1190 |
| B | 0.55 | 25.8% | 1291 | 100 | 1191 |
| C | 0.86 | 40.4% | 2019 | 400 | 1619 |

C gets the fattest slice and carries the most into its subtree; B the leanest. (Nuance:
*carry-to-kids* is delivered − cost, so a cheap cell keeps more of its share for its
descendants — B nearly matches A despite a smaller slice.)

**The global cap holds by conservation — which the cognition cell chooses.** Its `Observe`
returns weights that normalize to 1 (the shares above), so the children's budgets sum back to
the parent's remainder and total frame tokens can never exceed the root budget — the "global
max knob" is just the root's initial budget. (A *spatial* cell returns un-normalized weights
and does **not** conserve — see "Budget is world-defined.") The kernel drops a child when its
delivered budget < its own cost (**affordability** — arithmetic, not a cognitive threshold).

### The frame is a connected region

Because *traversing through* costs tokens, you pay your way in along paths — so the frame is
a **connected subgraph grown outward from the root**, not a scattered top-k. Consequences of
distributing budget *by branch*: an unaffordable cell **walls off its subtree**, and a
weak-weight parent **starves** even a strong descendant. A subtree that can't spend its share
leaves the frame **under budget** (waste — unused share doesn't flow to siblings). Both are
the price of investing by branch; both are fixable later with **sibling-rebalancing** (unused
budget flows back up and re-splits) — deferred until frames prove too sparse.

### Weight hygiene (when weight = product of factors)

- **AND-semantics.** A product means a cell must score on *every* factor; any near-zero
  factor kills the edge. Stricter than a sum (one strong factor carries) — probably what we
  want (precise, skeptical), just deliberate.
- **Watch zeros / bootstrap.** A fresh edge has alignment ≈ 0 → weight 0 → never lit. Factors
  need a floor or a never-zero base (impulse relevance is the natural base) so new edges can
  enter and *earn* their strength. Otherwise the economy can't bootstrap.

### The fold stays single-phase

Because the distribution *is* a recursive DFS descent, `Express`'s post-order survives
intact: children fold up to parents on the way out, one pass. No two-phase select-then-fold.

## Drill and gather — two policies, one kernel

The two "traversal models" are not two kernels. They are two **distribution policies** a
cell's `Observe` applies, reading the mode off the impulse:

- **drill** = allocate **narrow** — concentrate weight on the top edge → the wave goes deep.
  This is *focused attention / train of thought / exploitation.*
- **gather** = allocate **broad** — spread weight across many edges → the wave goes wide and
  shallow. This is *associative recall / search / exploration* (your attention model's
  `focus` vs `search`).

Same universal kernel; the cell decides narrow vs. broad. The mode rides on the impulse; the
membrane/orchestrator sets it when it launches a wave (a `recall` step fires gather, an
`elaborate` step fires drill).

**What can't be a cell policy — and why we don't want it.** A *true flat-global best-first*
(compare every node across the whole frontier, rescue a high-weight gem three branches away
behind a weak parent) is inherently non-local: no cell can do it. But that operation is
"rank everything by relevance, take top-k, ignore structure" — **flat retrieval, RAG — the
model Skyra rejected** (the location-based-memory pivot). So the un-cell-able strategy is the
one we shouldn't want. Recall done the Skyra way is *broad connected spread* (gather), which
is cell-local. A gem the spread can't reach isn't a kernel failure — it's a signal the
**edge to it is under-learned**, and the fix is the bump (strengthen the edge), not a global
override that papers over the topology.

## Budget is world-defined (units and conservation)

The kernel treats the budget as an abstract scalar. It does not know what it *measures* or
whether it's *conserved* — both are set by the world the wave runs in, never by the kernel.

- **Units** are set by each Space's stamped **`Cost()`**. Cognition: `Cost()` = the token
  count of the content (stamped at birth/write), budget = tokens. The spatial world: `Cost()`
  = 1 (a Chebyshev step), budget = *radius*. The primitive defaults already are this —
  `observePerceive` sets `impulse.Amplitude = DefaultPerceiveRadius` (20). The world has
  always denominated amplitude in reach.
- **Conservation** is set by the **cell's weights**. Cognition returns weights summing to 1 →
  children's budgets sum back to the parent → bounded frame (right: one context window). The
  spatial world returns weights of 1 each → every neighbor inherits the *full* `budget − cost`
  → reach one direction doesn't cost reach in another (right: space isn't conserved across
  directions).

| | cognition | spatial world |
|---|---|---|
| amplitude = | token budget | radius (the primitive default) |
| `Cost()` = | token count of content (stamped) | 1 (a hop) |
| weights = | sum to 1 (conserve → bounded frame) | all 1 (broadcast → flood to radius) |

One kernel, agnostic to both. `child_budget = (parent_budget − Cost) × weight` covers both:
normalized weights conserve; unit weights broadcast.

## World impact

The redesign lands well on the world because **the world is already fighting the bias we're
deleting.** `PerceiveCell.Children()` flat-enumerates every point in radius in *one hop*, and
the comment says why (`world/space.go:547`): *"no recursive BFS — the kernel's
branching-amplitude division kills BFS depth past ~5 hops."* The `/len` split collapsed
amplitude so fast a real grid walk died in ~5 steps, so perception was flattened to dodge it.
And `spaceActivation.Activate` already returns a flat constant — the world never used kernel
activation for falloff. It worked *around* the leak with radius-enumeration + a tuned
`amplitude`/`threshold` so the flat fan clears the floor.

What the world **must absorb** (or it breaks):
- **`Activate` → `Alloc` weights** — the flat baseline becomes weights in `PerceiveCell`/
  `Point` `Observe`.
- **threshold → affordability** — the fragile `amplitude`/`threshold` tuning dissolves.

What it **gains:**
- **Drop the flat-perceive workaround.** With no amplitude-collapse, a real grid spread
  returns: `Point.Observe` returns neighbor `Alloc`s (weight 1), `Point.Cost()` = 1, and the
  **visited-once rail** makes it a bounded **flood-fill to radius = budget**. The radius falls
  out of budget-depletion; no enumeration hack, no decay constant.
- **Honest distance falloff** if wanted: weight neighbors by proximity/direction instead of a
  flat 1, and perception thins with distance natively.

Blast radius: every world `Space` — `World`, `Point`, `PerceiveCell`, `EngageCell`, `MoveCell`
— gets the new `Observe`. Most are boundary/control cells returning `nil` or their enumerated
children as `Alloc`s; only `PerceiveCell`/`Point` carry real weight policy. Argues for
migration option (b): prove it in the dnav2 cognition path first; the world rework is its own
project.

## Division of labor — the line that must stay clean

- **The cell owns all cognition** (`Observe`): which edges light up, their weights, how broad
  the distribution — impulse-dependent, content-dependent, strength-informed.
- **The kernel owns transport only:** scale weights by the carried budget, subtract cost,
  check affordability, deliver, recurse; the visited rail; the fold; bookkeeping. **No
  constants, no policy.**

**The failure mode to guard:** any cognitive constant creeping back into `relation.go` — a
decay factor, a threshold, a default breadth. If the kernel ever needs to "know" something
about cognition to walk, the line has broken. The kernel should be explicable without
reference to what the cells *mean*.

## Kernel impact (`relation.go`)

The kernel gets **simpler**, not more complex — it sheds policy. The DFS and the fold survive.

**1. `scoreChildren` dissolves into arithmetic transport.** It no longer activates,
attenuates, or thresholds. It reads the `Alloc` map `Observe` returned, and for each child
computes `child_budget = (parent_budget − parent_cost) × weight_i`. No normalization —
conservation is the cell's weights, not the kernel's. `amplitudeDecay` (`:12`, `:204`) is
**deleted**. The candidate set is the `Alloc` map, not `Children()`.

**2. Threshold → affordability.** A child is descended iff `child_budget ≥ child.Cost()` —
pure arithmetic, replacing `score > threshold` (`:207`) and its constant. The DFS stack
(`:96`,`:98`) and descent are the same shape.

**3. `Activate(visited, count)` → gone; strength reads move into `Observe`.** The two kernel
call sites (`:168`, `:197`) disappear. The cell reads strength (lock-free) to compute its
weights. Edge-keying is preserved: `recordFor` (`:226`) still keys off the map key, and the
`Alloc` keys are those same edge identities.

**4. `Observe` runs lockless and now reads records.** Selection + weighting stays in the
lockless phase (`:90`, `:110`); the only lock is the kernel's arithmetic bookkeeping. Requires
lock-free strength reads (see "Kill the method, keep the state").

**5. The bump relocates** (unchanged): the strengthening moves to the crossing or the grade
write-back — the live open decision.

Untouched: the DFS stack + descent, the `onPath` visited-once rail (`:88`/`:105`), the **fold**
(`Express`/`recordResult` inline on stack-pop — single-phase), the double-buffer, proper-time,
lock granularity, the (suspended) return-perception path. `Children()` is **no longer read by
the kernel**.

## `Children()` vs. the frame

`Children()` does **not** die — it *diverges* from the `Observe` return:

- **`Children()` = structure** — every edge that exists. Retained for imperative reads and
  inspection (the grouping index reading its predictions, motor reading a proposal — reach-in,
  not traverse-through).
- **`Observe`'s return = the frame** — the weighted subset lit for *this* impulse.

A grouping may have 100 predictions attached and light up 3. That gap — structure vs. lit
distribution — **is** "the frame comes from traversal," made literal.

## Enforcement

- **Type — enforced, free.** Put the return in the interface and every implementer must
  return `map[string]Alloc` or it won't compile. Leaves return `nil`.
- **Semantics — not enforceable.** Real neighbors, sane weights, honest breadth — runtime
  discipline, not a compile check.

## Per-kind bookends — the next pass

Each live `MemoryCell` kind needs its `Observe` (what it lights + weights) and `Express` (what
it folds). Provisional — **next working session, not decided:**

| Kind | `Observe` lights (with weights) | `Express` folds to |
|---|---|---|
| `Entity` | interaction children, weighted by their edge alignment (`A`, contextual) | the entity's current state/summary |
| `Grouping` | the active episode: relevant predictions + the gist | the grouping's current understanding |
| `Prediction` (dist. root) | its branch children | the bet's current grade / status |
| `Prediction` (branch) | leaf (nil) | the branch's text + grade |
| `Response` | leaf (nil) | the interaction record |
| `Action` | leaf (nil) | the action record |

Baseline today (`memory_cell.go:97,101`): `Observe` no-op, `Express` returns `content`. Every
row is currently empty bracket.

## Open decisions

1. **The bump knobs** (rule is settled — see "Activation & the bump"). Three tunable defaults
   remain: salience immutable vs. running; confidence as penalty vs. exploration bonus;
   `rect(A)` hard vs. soft. Ship the flagged defaults, tune from experiment. (Former "where the
   bump lives" is resolved: crossing → F, outcome → A/S/N.)
2. **The record type / lock-free.** The record's *fields* are decided (A, S, F+`last`, N). Open:
   the exact lock-free/atomic mechanism for the F-bump write and the strength reads, and who
   mints the record (`NewActivation` survives vs. creation moves to `writeAlignment`/nucleus).
3. **`Express` as fold.** Ratify: `Express` returns a *fold over children's expressions +
   content*, still `any` (not a map). Proposed.
4. **`Alloc` weight = raw product or product-per-token?** Density (`product/tokens`) packs
   more small cells; raw product protects big important ones. A cell-policy knob now, not a
   kernel one. Low priority.
5. **Sibling-rebalancing.** Unused subtree budget flows back up and re-splits (fixes
   starvation/under-fill) or we accept branch-committed budget. Deferred.
6. **Should `Observe` receive the remaining budget?** Weight-only keeps the cell budget-blind
   and the kernel purest; but budget-aware breadth policies (adapt spread to how much is left)
   would need it. Start weight-only.
*Settled this pass:* the kernel carries **no cognition** — and doesn't know the budget's
**units** or whether it **conserves**. Split rule, decay factor, threshold, units, and
conservation all leave the kernel; `amplitudeDecay` is **deleted** (cost is the decay).
`Observe` returns a weighted onward distribution (`map[string]Alloc`); the kernel does the
barest arithmetic — `child_budget = (budget − Cost) × weight`, **no normalization**.
Units come from each Space's stamped **`Cost()`** (token count for memory, 1 for world),
conservation from the cell's *weights* (sum-to-1 vs. all-1). **Cost lives on the Space,
stamped at birth/mutation** (resolves former #7). Drill/gather are cell policies; true
flat-global best-first is flat retrieval and is **out**. DFS + single-phase fold survive.
The world already fights the old bias (flat-perceive workaround) — the change frees it to
flood-fill. **Activation is four factors stored separately** (A/S/F/N), maintained by the
bump — crossing updates F, outcome updates A/S/N (resolves former #1); `strength = rect(A) ·
S · F · conf(N)`, which **is** the weight. **Relevance = alignment** — no separate fresh term;
impulse-dependence is structural (which edge you arrive by). **Credit assignment falls out of
node birth:** a node's children = its birth-frame, and the closure grade distributes to the
frame's edges as `credit = activation · g` — the graph is the eligibility trace.

## Migration / blast radius — the real cost

This changes **`skyra.Space`, THE interface** — `Observe` gains a return and a new `Cost()`
method lands on every implementer (`MemoryCell`, `EphemeralRoot`, `membraneRoot`, the world
`Point`/grid, archive stubs), even leaves (`Observe` returns `nil`; `Cost()` returns the
stamped size or 1). `NewActivation`/`Activate` change on record types. Repo-wide diff.

Staging (decide before touching `space.go`):

- **(a) Change the shared interface directly** — cleanest end-state, widest blast radius.
- **(b) Prove it in the dnav2 path first** — the return on the nodes the dnav2 kernel walks
  (membrane, ephemeral, the five live `MemoryCell` kinds), world/archive on a shim, propagate
  once the per-kind bookends are validated.

Given the per-kind bookends are still unspecified, **(b)** likely — don't ripple the whole
interface until the node behaviors that justify the return are proven where the waves run.

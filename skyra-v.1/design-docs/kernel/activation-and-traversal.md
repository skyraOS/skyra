# Activation & Traversal Mechanics

## The Problem

The cell sees what's in front of it. The Slice keeps score. The Relation carries energy. Who does what, when, and where does state live?

## The Architecture

Two contracts. The wave and the inhabitant.

**IRelation** — the wave + shared substrate. Drives traversal, holds per-being state. This is the kernel. It does not know how scoring works.

**Reality** — the inhabitant. Cells implement this. The kernel traverses these. Each cell type defines its own activation struct and scoring equation.

They meet at the interface but they don't own each other. The Relation is mechanism. The Reality is identity.

## The Loop

1. **Cell.Observe(Relation)** — cell receives the Relation, does internal work
2. **Kernel reads Children()** — gets the cell's visible neighbors by reference
3. **Kernel calls Activate(true)** — on each present child's activation record
4. **Kernel calls Activate(false)** — on absent children (previously seen, not in current report)
5. **Kernel filters by threshold** — above threshold = survivor (the wave's amplitude is already folded into the score, so a quiet wave fails here)
6. **Kernel applies the rail** — a node already on the current path is skipped (visited-once)
7. **Survivors get pushed onto the stack** — iterative DFS continues

The cell is the sensor. The activation record is the scorer. The kernel is the driver.

**Competence brackets the cell's own work.** When a wave carries passed competence (a parent cell throwing its method down — e.g. global coherence localizing), the cell runs it **first on Observe, last on Express**: context primes before the cell scores, the check audits after the cell votes. This is a cell-execution contract, not a kernel mechanic — the kernel never reads the payload, it just delivers the wave. See [comparator.md](../cognition/comparator.md#the-cell-execution-contract).

**Note:** Attention is a separate concept from activation records. Attention is the being's aperture — heading and width — one field the membrane reads from both directions: sensory gates incoming signal against it, motor fires through it (see [world-kernel-coupling.md](world-kernel-coupling.md#attention-is-one-field-read-twice)). Activation records are the scoring system. They are not the same thing.

## The Interfaces

Three interfaces. Two contracts and one extension point.

### IRelation — the wave

```go
type IRelation interface {
    ID() string
    Realize(beingID string, brID string, root Reality)
    GetSlice(beingID string) *Slice
}
```

- **Realize** — drives the full iterative DFS. The one entry point.
- **GetSlice** — the being reads its own cognition. Attention, Prior trees, Records.

The Relation is both the wave that moves and the substrate that remembers. It carries energy for one traversal but holds Slices that persist across all traversals. The relationship between the world and the being.

### Reality — the inhabitant

```go
type Reality interface {
    ID() string
    NewActivation() ActivationRecord
    Observe(r *Relation)
    Express(r *Relation)
    Children() map[string]Reality
}
```

- **NewActivation** — factory. Returns whatever scoring struct this cell type's DNA defines. The kernel stores it, calls it, never downcasts.
- **Observe** — called on descent. The cell does its internal work.
- **Express** — called on ascent. The cell does its output work.
- **Children** — exposes visible neighbors to the kernel.

### ActivationRecord — the scoring contract

```go
type ActivationRecord interface {
    Activate(visited bool, traversalCount int) float64
}
```

One method. The kernel calls it. The cell defines what it means. DNA injects the factory via `NewActivation()` on the Reality interface — the genome wires in the activation at boot, the cell carries it, the kernel scores it.

Each cell type implements `NewActivation()` — a factory that produces whatever scoring struct that cell type needs. The kernel stores it, calls it, reads the score. Never downcasts.

## BaseActivation

A default implementation for embedding:

```go
type BaseActivation struct {
    Identity       float64  // EMA α=0.02 — the groove (long tail)
    Familiarity    float64  // EMA α=0.1  — recent relationship
    LocalTraversal int      // per-cell visit count
    LastTraversed  int      // global count when last bumped

    // Per-edge inputs — written by the cell (in Observe or Express) before the
    // kernel calls Activate. The kernel never writes these; the cell owns them.
    IndexScore  float64  // surroundings overlap (one-shot per cycle)
    Alignment   float64  // alignment spectrum value (-1 to 1), lagged from Frame 2
    // NOTE: stress is NOT here — it is global (read from the slice at Activate
    // time). Intrinsic excitability is NOT here either — it is per-node
    // (MemoryCell.BirthAlignment). See the bump equation below.
}

func (b *BaseActivation) Activate(visited bool, traversalCount int) float64 {
    // sample = visited ? (IndexScore × Alignment) × stress_modulator × intrinsic : 0
    // Identity    += αI · (sample − Identity)    // αI = 0.02
    // Familiarity += αF · (sample − Familiarity) // αF = 0.1
    // return Identity + Familiarity              // groove + recency (tunable blend)
    // The bump IS the sample fed into the EMAs — no fixed +1 constant. Signed:
    // a negative Alignment gives a negative sample → EMAs (and the result) go
    // negative → threshold drops it (below-0 eviction).
}
```

### The bump equation

```
bump = ( IndexScore × Alignment ) × stress_modulator × intrinsic
       └── per-edge, relational ──┘   └── global ──┘   └─ per-node ─┘
```

Three loci, three different homes — this is the load-bearing distinction:

- **`IndexScore`, `Alignment`** — **per-edge** (relational). On the activation record, keyed by the edge. "How much does *this connection* overlap the surroundings / how aligned was *this connection*." Synaptic weights.
- **`stress_modulator`** — **global** (the being's state). One value on the slice, read at `Activate` time. A neuromodulator (cortisol / norepinephrine): diffuse, brain-wide gain, **not** stored per-edge. (See "Stress is global" below.)
- **`intrinsic`** — **per-node** (`MemoryCell.Intrinsic`, a struct; its composed `Value()`). Cell-intrinsic excitability, set **once at crystallization** from `{ClosureGrade, StressAtBirth}`. The memory's born confidence/precision, distinct from any synapse. (See "Birth alignment" below.)

Both EMAs receive the same composite bump — they differ only in α:

| EMA | α | Character |
|---|---|---|
| Identity | 0.02 | Hard to shift; represents what is persistently true about this cell |
| Familiarity | 0.1 | Moves faster; reflects the recent picture |

A spike in all three inputs bumps Familiarity noticeably and Identity barely. Only
consistent signal over many cycles deepens the Identity groove. One-off spikes
show in Familiarity but don't move Identity.

**`IndexScore` is the surroundings overlap, computed once per cycle.** It is *how
many of the currently active surroundings entities connect to this cell* — the
overlap that gives high-overlap setpoints priority. Compute it **once per cycle**
as a one-shot count (gather `entity.Children()` across the surroundings, or let the
entity-waves converge) and feed it into a **single** bump. Do **not** let a cell
take N separate bumps because N entity-waves reached it in one cycle — that would
over-groove the durable EMAs and conflate "relevant to a lot *right now*" with
"consistently relevant *over time*." Overlap is the this-cycle IndexScore
multiplier; the EMAs groove once per cycle. (See
[entities-and-surroundings.md](../cognition/entities-and-surroundings.md).)

### Ownership: the cell writes its inputs before the kernel calls Activate

The `Activate(visited, traversalCount)` signature is unchanged — the kernel stays
clean. The cell updates the **per-edge** inputs (`IndexScore`, `Alignment`) on its own
activation record before the kernel calls `Activate`; the record reads **global stress**
off the slice and the **node's intrinsic** birth value at compute time. By the time the
kernel calls, everything it needs is reachable.

```
1. Cell.Observe / Cell.Express → writes IndexScore, Alignment onto its (per-edge) record
2. Kernel calls Activate(visited, traversalCount)
3. Record reads global stress (slice) + node intrinsic (BirthAlignment)
4. Record computes bump = (IndexScore × Alignment) × stress_modulator × intrinsic
5. Record updates Identity EMA (α=0.02) and Familiarity EMA (α=0.1)
6. Returns score → kernel gates by threshold
```

No signature change. Per-edge state is the cell's; stress is the being's (slice); intrinsic is the node's.

### Stress is global, not per-edge

Stress is a **neuromodulator** — cortisol (HPA axis, blood-borne) and norepinephrine
(locus coeruleus, diffuse cortical projection). Neuromodulators are *defined* by being
broadcast, brain-wide, **not** point-to-point. So stress is **one value for the being**,
held on the being's substrate (the concrete slice) next to the surroundings map and WM
buffer — read as a global gain in the bump, never stored per-edge. **It is NOT on the
`skyra.Slice` (kernel) interface** — the kernel is mechanism and must not know about a
cognition signal; cells read it via a capability assertion, not the kernel contract.
It rises with cascade / prediction-error and decays.
One global stress drives several gates at once (the conflict-monitor compression): the
bump modulator, episode-segmentation sensitivity (`threshold / (1 + stress)`), and the
cascade-stop threshold. (Likely one of a small set of global modulators — stress, and
later valence/emotion — so leave room for an affective-state vector, but one scalar to
start.)

### Birth alignment — cell-intrinsic excitability (per-node)

Memories aren't born equal. Beyond the synaptic (edge) weights, a memory carries an
**intrinsic** strength set **once at crystallization** — the brain's *memory-allocation*
mechanism (cell-intrinsic excitability, CREB; Silva/Josselyn): cells more excitable at
encoding are recruited into the engram and carry a baseline strength independent of any
synapse, stamped by the global state at that instant (why high-arousal moments form
"flashbulb" memories). In predictive-coding terms it's the memory's **precision** — its
born confidence, the prior on how much it pulls before relational evidence accumulates.

It is kept as a **struct**, not a single float, so the constituents stay separable —
room to reweight or add fields (valence/emotion-at-birth, novelty) later without losing
the raw inputs (same philosophy as keeping IndexScore/Alignment distinct):

```
Intrinsic { ClosureGrade, StressAtBirth }        // stamped ONCE, at completion
BirthAlignment = Intrinsic.Value() = ClosureGrade × StressAtBirth   // the composed value used in the bump
```

It is **distinct from the EMAs**: Identity/Familiarity are *accumulated* activation
(groove/recency, per-edge); BirthAlignment is a *one-time intrinsic stamp* (per-node). A
memory can be born strong yet rarely revisited, or born marginal yet grooved over time —
independent axes, as in the biology. Where it earns its keep:

1. **Cold-start confidence** — a fresh setpoint has no edge history; its BirthAlignment is
   the baseline that lets it participate before edges accumulate (else it's invisible).
2. **Solid ground (cascade-stop)** — a high-BirthAlignment node is *intrinsically solid*;
   a cascade should halt on it. BirthAlignment is a natural input to the solid-ground
   threshold ([error-recovery-loop.md](../cognition/error-recovery-loop.md)).
3. **Precision-weighting** — the node's pull scales by it: `node_score ≈ intrinsic × relational`.

**Alignment is a lagged value.** The `Alignment` field on the activation record is
written from the *previous* Frame 2 output for this cell — not computed fresh during
the kernel traversal. The kernel traversal runs first; Frame 2 scoring runs after;
the resulting alignment scores are written back onto the activation records before
the *next* traversal. The cycle is:

```
kernel traversal (reads Alignment from last cycle)
  → Frame 1 (one LLM call → N predictions)
  → Frame 2 (N concurrent LLM calls → N alignment scores)
  → scores written back onto activation records
  → next kernel traversal reads updated Alignment
```

This makes the kernel traversal fully self-contained (no LLM calls inside it) and
the alignment feedback a natural part of the episode cycle.

### Activation is edge-owned, not node-owned (the synaptic model)

A record belongs to a **connection**, not a cell. The same shared content cell B
can be excited differently by A and by C — A→B and C→B are independent weights.

This needs **no new interface and no edge-node objects.** Records still live on
the slice; the kernel just keys them off the **map key** the cell provides, not
the child's own `ID()`:

```
recordFor(key, cell, sl) → records[key]
```

A cell achieves edge-ownership by handing edge-scoped map keys (`"A→B"` vs
`"C→B"`) in its `Children()` map — the records then differ automatically. The
kernel stays oblivious: it keys a record off whatever identity it was handed.
The world keeps node-keying for free (its map keys equal its point IDs).

**Ownership split:** the cell owns its `Children()` *adjacency* (which edges
exist); the slice owns the edge-keyed *weights* (persistent across traversals).
Footprint is `O(actual edges)` — keys exist only for edges a cell declares.

**Transitive reach is a path, not a stored edge.** A declares edges only to its
direct neighbors. Its excitation of a 2-hop cell is the path composition
(`ab ⊗ bc`) — computed at traversal time by amplitude attenuation
(`activation × amplitude-that-arrived`), not stored. A-specific because the
amplitude started at A. Spend-then-spill is emergent, not a kernel mechanic:
direct edges spend the loud amplitude, leftover attenuates to potentials.

**Path → edge crystallization (Hebbian).** A path-value (a *potential*) that
grades above the keep-threshold (the same `0.1` that gates adding stimulus to a
setpoint's children) crystallizes into a direct edge: the source cell adds the
target to its own `Children()` adjacency (DNA-side — the kernel doesn't do this),
born at the value that triggered it, then grooving or decaying like any edge.
**History** = edge-keyed records with no adjacency (scored potentials that didn't
crystallize, incl. negatives): they persist with their value but the kernel never
walks them. Edges that stop earning grades decay below threshold and fall away —
the graph forms *and* prunes, the structural-learning channel.

**Decay trends to 0 from both directions.** A fading relationship returns to
neutral (0), not to an extreme — multiplicative decay does this symmetrically
(positive shrinks toward 0, negative rises toward 0). Bumps are **signed**
(negative `Alignment` → negative bump → EMA moves negative), so a
consistently-and-recently-negative relationship holds its negative value while a
negative-once-long-ago one drifts back to ~0 (effectively unknown again) — the
two-EMA timescales encode that distinction (Familiarity = recent, Identity =
persistent). If negatives should be *stickier* than positives (aversive memory),
let **stress** carry it (a high-stress negative grooves harder / decays slower),
keeping stickiness state-dependent rather than a hardcoded sign-asymmetry.

Binding realities embed BaseActivation and extend with domain-specific fields:

```go
type EpisodicActivation struct {
    BaseActivation
    Overlap float64
}

func (e *EpisodicActivation) Activate(visited bool, traversalCount int) float64 {
    base := e.BaseActivation.Activate(visited, traversalCount)
    return base * e.Overlap
}
```

Each binding reality defines what activation means for its own cells. No shared BindingFields map. No external writes. The cell owns its scoring.

## Bump and Decay

The kernel calls `Activate(true)` for present children, `Activate(false)` for absent ones. What happens inside is the activation record's business.

- **Visited = true** → compute `bump = IndexScore × Alignment × stress_modulator`; advance Identity and Familiarity EMAs toward bump; increment LocalTraversal; update LastTraversed
- **Visited = false** → decay Identity EMA toward 0.0, decay Familiarity EMA toward 0.0

Stress modulates both bump magnitude and decay rate (open — exact stress_modulator function TBD).

Other implementations can do whatever they want. The kernel doesn't care.

## Threshold

The Slice holds a threshold value (DNA-defined). After the kernel calls Activate on all children, only those returning a score above threshold become survivors for descent.

## Dependencies

What depends on what. If something accumulates across traversals, the Slice owns it. If something exists only for the current wave, the Relation owns it.

### What the Slice owns:
- **TraversalCount** — monotonic global clock, increments once per BeginTraversal
- **Records** — `map[string]ActivationRecord` — persistent scoring state, **keyed by the map key the cell hands the kernel** (an edge key like `"A→B"` for DNA cells, a point ID for the world). Under edge-owned activation this becomes per-edge rather than per-node, but it stays on the slice — only the key changed (`recordFor(key, …)`), not the home.
- **LastChildren** — `map[string][]string` — what children each cell reported last time (for decay)
- **Attention** — what's around the being (placeholder)
- **Threshold** — DNA-defined minimum score to descend
- **Stack** — DFS state machine (ephemeral per traversal)
- **Trees** — `map[string]*TraversalPair` — current/prior trees per binding reality

### What the Relation owns:
- **Amplitude** — the wave's signal strength, set at spawn (the agent's investment). Folds into activation (`activation = base × amplitude`) and attenuates as the wave descends, so a quiet wave simply fails threshold. Not a step-budget — visited-once is the rail. Attenuates across crossings too, which bounds the goroutine tree.
- **Slices** — `map[string]*Slice` — one per being

### What the ActivationRecord owns:
- Whatever internal state it needs for scoring (EMAs, counters, custom fields)
- Updated only through the `Activate` method
- Persists across traversals via the Slice's Records map

### Traversal itself
- **Depends on:** score > threshold (amplitude is already a factor in the score) AND the node is not already on the current path (visited-once)
- **Therefore:** the Slice's score gates descent; amplitude rides inside the score, and visited-once guarantees termination

### Order of operations per cell:
1. Kernel calls cell.Observe(Relation)
2. Kernel reads cell.Children()
3. Kernel looks up or creates ActivationRecord for each child
4. Kernel calls Activate(true) on present, Activate(false) on absent
5. Kernel filters by threshold (amplitude folded in)
6. Kernel skips nodes already on the path (visited-once)
7. Survivors pushed to stack

## The Node

The kernel's internal traversal frame:

```go
type Node struct {
    CellID    string
    Cell      Reality
    Record    ActivationRecord
    Score     float64
    Children  []*Node    // all children — above and below threshold
    Survivors []Reality  // above threshold — queued for descent
    NextChild int        // index into Survivors for iterative DFS
}
```

This is ephemeral traversal state. The Node exists for one traversal. The Record inside it persists across traversals (stored in Slice.Records).

## Iterative DFS

No recursion. No suspended stack frames. The kernel drives the loop explicitly.

```
push root
loop:
    top = stack peek
    if top has unvisited survivors:        # survivors already passed threshold (amplitude folded in) and visited-once
        child = next survivor
        child.Observe(r)
        attenuate amplitude
        score children (amplitude folds into the score), push child node
    else:
        pop top
        top.Cell.Express(r)
        append top to parent's Children (or to tree root if stack empty)
```

The Relation sees all the cells' methods. The Relation calls those methods in a specific order that traverses the interior. The cell doesn't drive traversal — the kernel does. If the cell's implementation doesn't traverse correctly, that's the cell's problem. Like Linux — no concessions.

## Tree Reconstruction

The tree builds on ascent. When a node pops, its parent is always the current top of stack (guaranteed by sequential DFS — no interleaving).

```
[A]         → A sees B, C. Pushes B first
[A, B]      → B sees D, E. Pushes D
[A, B, D]   → D has no survivors. Express pops D, appends to B
[A, B]      → B pushes E
[A, B, E]   → E pops, appends to B
[A, B]      → B done. Express pops B (with children D,E), appends to A
[A, C]      → C sees F, G, H. Pushes F
[A, C, F]   → F pops, appends to C
[A, C, G]   → G pops, appends to C
[A, C, H]   → H pops, appends to C
[A, C]      → C done. Express pops C (with children F,G,H), appends to A
[A]         → A has full tree
```

### Double Buffer (TraversalPair)

Each binding reality gets its own current/prior tree pair:

```go
type TraversalPair struct {
    Current []*Node
    Prior   []*Node
}

Trees map[string]*TraversalPair  // keyed by binding reality ID
```

At the start of every traversal: `Prior = Current`, `Current = nil`, stack cleared.

Nothing reads a half-built tree. Binding realities read Prior (last complete traversal). Current builds without interference.

### Complexity

O(n) push + O(n) pop = O(n). Theoretical minimum for tree construction.

### Parallel Future

If traversal ever goes parallel, the stack model breaks — interleaving destroys parent lookup. Would need separate stacks per subtree. Not now.

## Two Recursions

This is the single hardest thing to hold about the kernel, and once you see it the whole system reads cleanly: the kernel runs **two different recursions, using two different mechanisms.** Keep them separate and it makes sense; conflate them and nothing does.

**1. In-plane traversal — the explicit stack, synchronous.** A wave moving across *one* actor's topology (its own binding realities, memories — reached through `Children()`) is the iterative DFS above, driven by a manual stack (`[]*Node`). No recursion here, no goroutines — one wave, one stack. The kernel owns every frame so it can apply policy at each step (visit-once, energy, threshold, scoring) and build the actor's result on the ascent. This is the wave spreading through a graph, and it is sequential.

**2. Cross-actor crossing — goroutine spawn, asynchronous.** When the wave hits a cell whose `Content` is itself a Reality, that is not a sub-structure of the current computation — it is **another actor.** The tell is in the code: the crossing looks up *that reality's own slice* in `ActiveState`, and a separate slice means a separate frame, a separate proper-time clock, a separate being. Actors are concurrent, so the crossing **spawns a goroutine** (`go realize(inner, inner, impulse)` — a new wave whose origin is the inner actor) — it does not call down the Go stack. The parent wave's `OriginReality` is not threaded through the boundary. The inner actor runs its own in-plane DFS on its own goroutine — concurrent and non-blocking. The outer wave does not wait, and it does not receive the inner wave's result.

So the shape is: **an explicit stack-machine inside each actor, and a goroutine per actor.** The goroutine tree *is* the actor tree. Within an actor, one stack walks the topology sequentially; cross an actor boundary and a new goroutine opens with its own stack — concurrently.

### Why the split

Each recursion is matched to the mechanism that fits it:

- **In-plane is large, policy-heavy, and one being's own thought** → explicit stack, sequential. A plane can hold a huge topology, so you want a heap-allocated stack; and the kernel must own the frame state (survivors, NextChild, score) to gate and score each step.
- **Cross-actor is a boundary between concurrent beings** → goroutine spawn, async. Crossing into another actor synchronously would mean one being blocking inside another's interior and ripping out its response — wrong. The other being is its own goroutine; it processes on its own clock and answers when it answers.

### Delivery to origin is the relay

There is **no return bubbling up a Go stack across actors.** Each goroutine delivers its result to *its own* origin — that is the [Perception Invariant](../laws/invariants/perception-invariant.md), and it is the only relay. The explicit stack delivers in-plane (it builds the actor's own result on the ascent); the spawned goroutine delivers to the actor that owns that spawned wave. A response from one being to another is not a return value — it is a **separate wave back**, another goroutine. Async message passing between concurrent actors.

That this is async is not a compromise, it is more correct: it respects being boundaries (no blocking inside another being), it gives the concurrency temporal binding wants (many staggered waves over the shared substrate), and it matches the physics — active sensing has propagation delay, so a perception returning a cycle later *is* the reafference loop, not a bug.

### The two-word vocabulary

- **traverse** a plane → the explicit stack (one actor, one topology, sequential).
- **cross** an actor boundary → a goroutine spawn (async; the inner actor delivers to its own origin).

Sync exists only *inside* an actor; crossing an actor boundary is always a spawn. Whether something is sync or async is therefore a **placement decision** — make it a `Child` (in-plane, part of my thought, sequential) or put it behind `Content` as its own actor (cross, concurrent). The genome decides by where it puts the reality. (The "Parallel Future" caveat above still holds: a *single* plane's DFS stays sequential — the concurrency is across actors, never within one stack.)

## What the Kernel Does NOT Know

- What scoring means
- How many EMAs exist or what their alphas are
- What recency is
- What binding fields are
- How decay works internally
- What emotion, episodic overlap, or motivation relevance are
- What a wave's shape is — it carries the aperture on the Relation and traverses the survivors the cell returns; it never interprets shape or resolves direction. Shape is attention's, not the kernel's (see [world-kernel-coupling.md](world-kernel-coupling.md#wave-shape--ownership))

The kernel is mechanism. The cell is policy. DNA is the distro.

# Durable Structure vs. Ephemeral Execution

**Date:** 2026-08-03
**Status: FOUNDATION.** The separation is settled. The graph's own edge semantics are not —
see *Open* below, and note that the arithmetic depends on them.

> **The runtime separates what the agent knows from what the agent is currently doing.**

---

## The two structures

```text
DURABLE     EXPECTATION GRAPH    accumulated experience, expectations, impulses, and
                                 their causal relationships. persists across invocations.
                                 evolves ONLY when learning occurs.

EPHEMERAL   EXECUTION STACK      a traversal over that graph. its top frame identifies the
                                 CURRENT EXPECTATION. begins, pauses, resumes, terminates
                                 without altering the causal structure.
```

Execution is not represented by the graph. The graph is a place; execution is a walk.

```text
Durable                          Ephemeral
────────                         ─────────
Expectation Graph                Execution Stack

Maintain Agent                   Maintain Agent
├── Respond                          ↓
│   ├── Retrieve Context          Respond
│   └── Generate Response             ↓
└── Learn                         Retrieve Context
```

## The anchor

> **The active expectation is the runtime's execution anchor — a process's current working
> directory.**

All observation, affordance discovery, realization, and context retrieval resolve **relative to
this position**. The runtime does not reason globally over the graph; it reasons locally from
where it is standing.

This is the filesystem relationship exactly. The filesystem exists independently of any process.
A process does not become the filesystem — it traverses it, holds a position in it, and resolves
names against that position.

The analogy is load-bearing rather than decorative, and it predicts:

```text
relative resolution    names resolve against the anchor, not the root
the parent hop         `..` is the frame beneath the top one
local affordances      what is executable HERE, not everywhere
retrieval as reading   context comes from the neighbourhood, not a global scan
```

## What this buys, structurally

**Framing stops being a search problem.** [2026-07-27/workers.md](../2026-07-27/workers.md) built
the observation worker's frame out of activation, trust ranking, and a budget cut — three
mechanisms deciding what to look at. Position does that job for free: the frame is **derived
from where you are standing** rather than searched for.

### Handle and frame are different things

Two established meanings were briefly collapsed here and should not be. They are related by
derivation, not identity:

```text
HANDLE   a POSITION — where execution is standing. one stack entry.
FRAME    the NEIGHBOURHOOD that position exposes — what is in view from here.
```

A handle's position **determines** the attention frame. It is not one. `workers.md` keeps
"frame," and the execution object is a **handle** — Unix's open file description, coupling a
durable node to a behavior plus the transient state of this particular use.

**And it satisfies Toy 5's weakened requirement by construction.** The toy's partial answer
reduced the runtime guarantee to:

> *guarantee `o*` is scored against SOME frame. any frame.*

A stack always has a top. **There is always an active expectation, so there is always a frame** —
the guarantee is structural rather than something the runtime must be careful to provide.

That is not the whole of Toy 5. An impulse arriving while the stack sits somewhere unrelated
still finds no home. But the failure changes character in a way that matters:

```text
BEFORE   attention suppressed it        invisible. the controller narrowed and nothing recorded why.
NOW      we were standing elsewhere     IN THE TRANSCRIPT. auditable, and recoverable by moving.
```

The self-sealing loop needed the suppression to be invisible. Position is never invisible.

## Why the graph may be walked while it is learned

A process holding a working directory raises the obvious question: what happens when the
structure changes underneath it. Unix answers with handles; this set already answered it
differently, and better:

```text
ACTION EDGE      write-once
SUPPORT VALUE    immutable once written
EVIDENCE SET     append-only
```

Nothing is ever removed or rewritten, so a stack cannot be standing on something that ceases to
exist. Learning **adds**. A traversal in flight sees either the new material or not, and both are
consistent. This is the same property that let the kernel drop its lock, arriving from the other
direction.

## Where the kernel already agrees

[kernel.md](kernel.md) derived, from law 1 alone, that a wave's transcript is a **chain** —
`depth == index`, `parent == Hops[i-1]`. That was not designed toward this. It is the same
constraint surfacing twice:

> **The transcript IS an execution stack.** `Hops[i-1]` is the frame beneath the top one.

So the substrate needs nothing added to carry a position. A wave already walks one path and
already records it as a chain of frames.

## Space creation is a coupling

> **Space creation is the coupling of data to methods, in a context specific to the agent
> being created.**

A Space is neither half on its own. The durable node is data with no behavior; a behavior is
methods with nothing to act on. The Space is the **join**, made for one walk and thrown away.

```text
NODE        durable data, agent-neutral, no methods that execute
BEHAVIOR    methods, supplied by whoever is walking
SPACE       the coupling. ephemeral. dies with the wave.
```

This is dnav2's `Transcribe` generalized. That selected which genes to express for one wave;
this selects which methods are coupled to this data for this walk.

**It generalizes many-processes-one-filesystem into many-PROGRAMS-one-filesystem.** The
cognitive agent, a world, and a user agent can traverse the same durable graph and do entirely
different things over it, because none of them is doing anything the graph itself specifies.

### The rail

```text
THE COUPLING IS CHOSEN BY THE AGENT. NEVER BY THE DATA.
```

A node that could influence which behavior gets coupled to it would be **selecting its own
transformer** — the guarantee `2026-07-26/definitions.md` §0a exists to hold, and the epistemic
form of the kernel's own discipline. Nothing reachable from a node may reach that choice.

### It puts BehaviorID back in the transcript

`2026-07-27/kernel-alteration.md` §10 listed `BehaviorID + version` as transcript content, and
[kernel](kernel.md) stripped it on entry along with the rest of the superseded vocabulary,
because nothing in the loop asked for it.

Coupling asks for it. **Two waves over the same node with different behavior are different
events**, and a transcript carrying only `SpaceID` cannot tell them apart — which means it can
no longer answer *what actually ran*, the one thing it exists for.

That is the strip-down rule working correctly rather than a reversal. It comes back because
something now requires it, not because it was always coherent. See [questions](questions.md) Q7.

## ONE STACK — 08-06

> **One subgraph is active at a time. It is connected, entered at a root. There is one stack.**

### Why this is safe now, and would not have been

Stack-per-impulse was adopted (08-04) to stop goal-directed and stimulus-driven work competing
for a single pointer. That competition only matters if **an execution can run long** — blocked on
the model, blocked on the world.

Actions-as-residual removed both. A wave is graph reads and graph writes and nothing else: no
model call, no world call, no network. It runs to completion and ends. There is nothing
long-running for an arriving impulse to compete with, so the premise of the earlier argument is
gone.

### The property that has to survive the change

Several stacks were carrying Toy 5's ingress guarantee — *admission always succeeds; at capacity,
evict rather than reject*. Under one stack that has to be preserved a different way:

```text
SAFE    impulses queue in ARRIVAL ORDER. every one is processed. nothing is ranked.
SEALS   the queue is ordered, or dropped, by relevance · trust · activation.
```

The property was never "many stacks." It is:

> **Admission must not consult attention.**

A FIFO queue gives that as well as N stacks did, and more simply. What must never appear is a
queue that decides which impulse matters — that is the sealing failure relocated into the intake.

### Enter at the standing root

*"Joined at a root"* raises which root, and under one stack the answer carries more weight than it
used to: whatever is entered is the agent's **entire** attention for that turn.

```text
THE STANDING ROOT   always the entry. the active subgraph is the whole tree, BOUNDED BY
                    activation and closure. the bound is derived; nobody chooses it.

A SUBGRAPH ROOT     enter partway down. cheaper — and whoever picks the entry point has
                    chosen what the agent attends to, by hand.
```

Entering at the standing root every time keeps that authority derived. `chroot` remains possible
([Namespace.Root](../../src/skyra/cognitive-agent/skyra-dna-v4/graph/namespace.go)) and is how a
resumed execution re-enters, but it should not be the ordinary path.

### What this makes vestigial

`kernel/runner.go` was built for concurrent long-running stacks. Under this rule its capacity,
eviction, and round-robin are unreachable. **Kept, not deleted** — the concurrency deferral names
*many processes, one filesystem* as the eventual target and this is that machinery — but the file
says so, so its tests are not read as describing live behavior.

## Three lifetimes, not two

The separation above names two, but *"begin, pause, resume, or terminate"* requires a third. A
stack that survives a pause outlives any single wave:

```text
DURABLE     expectation graph        across invocations         append-only
SESSION     execution stack          across waves, within work  resumable
EPHEMERAL   the wave, and everything
            derived during it        one traversal              dies at the end
```

The kernel currently has only the first and third — it holds no state between waves, which is
precisely why concurrency dissolved. **The session tier is new, and it needs an owner.** It
cannot go on the kernel without giving back the property that made concurrency go away.

The natural placement is that the stack is a value the *caller* holds and threads into each
wave, exactly as a shell holds a working directory rather than the kernel holding it for it. Not
decided here.

---

## Open

Carried in [questions](questions.md) — Q1 (what the child edge means), Q3 (wave granularity),
Q4 (what owns the session tier), Q5 (unhomed impulse).

**Settled since this doc was written:** every expectation is graded directly ([graph](graph.md)),
which retires the depth-and-closure concern this section previously raised — closure now has a
channel that does not inherit path decay. Concurrency is deferred, single-threaded for now, with
many-processes-one-filesystem as the target shape (Q6).

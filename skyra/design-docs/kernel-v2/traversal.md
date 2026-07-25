# Traversal

**Status:** DESIGN · part of [kernel-v2](README.md)
**Replaces:** `transport.md` — the word was wrong. There is no budget to transport.

## The kernel's job, in one line

> **The kernel moves from one state to the next state.**

## Nodes are STATES. Edges are EXPRESSIONS.

```text
desired ←──[E₃]── S₂ ←──[E₂]── S₁ ←──[E₁]── current
```

A **state** is a set of triples. An **expression** transforms one state into another:

```text
FORWARD    S' = S ∪ expects(E)                     applicable when requires(E) ⊆ S
BACKWARD   S  = (S' \ expects(E)) ∪ requires(E)    ← REGRESSION. this is what the wave does.
```

That is goal regression — classical preconditions and add-effects. The kernel's atomic operation is
**one state transition**; a wave is a chain of them.

## Origin and target are the INVERSE of the agent's

```text
THE AGENT'S TARGET   the DESIRED state.  where it wants to be.
                     ↓ handed to the substrate, which MATERIALIZES the nodes (bootstrap.md)
THE WAVE'S ORIGIN    the DESIRED state.  ← the wave STARTS here.
THE WAVE'S TARGET    the CURRENT state.  ← the wave is trying to REACH here.
```

The agent's *goal* is the wave's *starting point*. The agent's *now* is the wave's *destination*.

## `Done()` — amplitude is a stop condition, not a budget

**Budget does not exist.** No `budget − cost`, no affordability gate, no proportional split. There is
one polymorphic termination predicate, and the cell computes it as it walks:

```text
RETRIEVAL wave    Done() = the regressed state ⊆ the CURRENT state         ← closure
PRESENT           Done() = the assembled present is full   tokenPresent: 100k
PERCEIVE          Done() = the boundary is reached          radius: 3  |  500ms
```

Same kernel. Same walk. **The condition is what the wave is aiming at.**

### And reaching the target IS the trigger for Express

```text
DESCENT   Observe walks backward from desired, guided by P
REACHED   the regressed state ⊆ current  →  THE WAVE HAS ARRIVED
ASCENT    Express walks it forward, executes, and reality reports
```

Descent creates final causes. Ascent creates efficient causes. **Arrival is the hinge.**

## Weights ORDER; they do not PORTION

The old model split budget proportionally across children. That is wrong now, for a reason that is
easy to miss:

> **A context window is GLOBAL.** It is not *"this branch gets 30k and that branch gets 70k."* It is
> *"the whole present must fit in 100k."* **There is nothing to split.**

So:

```text
OLD   weight ALLOCATES   child_budget = remainder × weight    proportional split
NEW   weight ORDERS      expand the most probable first, fill until Done()
```

**Best-first expansion.** And it is strictly better: under a budgeted DFS you can burn the whole
window on one deep irrelevant branch before ever seeing a shallow relevant one.

## Weight = PROBABILITY. Say it precisely.

Not "relevance." Not "priority." Not "salience." **Probability** — and being sloppy about *which*
probability is how a judge sneaks into the kernel.

> **The kernel asks, at every step: what is the probability that this reaches my target state?**

And here is why backward chaining is right and not merely elegant:

```text
P(W | E)         the learned edge.  "does this expression produce this state?"      CAUSAL
P(reach target)  the search weight. "does this edge get me where I'm going?"        GUIDANCE
```

**In a backward search these coincide.** You start at the desired state; a state's frontier is the
expressions that produced it, weighted by how reliably they did. `P(W|E)` **is** the search
heuristic. There is no second quantity to invent, and **no relevance model to smuggle in.**

### The present is where a judge could still get in

For the present, *"is this node relevant to my target"* sounds like a similarity judgment — and the
moment it becomes one, an embedding model or an LLM is now steering the kernel and the grounding
story is dead.

**It does not have to be.** `P(useful | target)` is **countable**:

> *"When I was aiming at a target like this before, what was in the present, and did it work out?"*

Same posting lists. Same edges. Weighted by outcome. **No embeddings, no cosine similarity, no
judge.** If we ever cannot make a kernel weight countable, that is the alarm.

## The kernel, derived

```text
push the ORIGIN (the desired state)
while !env.Done() and the queue is not empty:
    pop the highest-P node
    if !env.Admit(node): continue          // it does not fit
    push node.Observe(env)                 // each child with its P(→ target)
fold post-order  →  the tree IS the bridge, and the lit subgraph IS the present
```

The kernel holds **a queue, a float it does not interpret, and two booleans it does not interpret:**

```go
Done()  bool          // has the condition tripped?
Admit(s Space) bool   // does this fit? (and account for it)
```

No arithmetic. No affordability. No normalization. No notion of AND or OR.

## Conserve vs broadcast is still AND vs OR — but the cell owns it

```text
weights SUM TO 1   CONSERVE    alternatives compete                → an OR node
weights ALL 1      BROADCAST   every child is needed               → an AND node
```

**The CELL normalizes if it wants to conserve. The kernel NEVER divides.**

This is the correction to `proportional-amplitude-allocation.md`: its formula
(`child = available × claim/Σclaim`) is right as a description of *what an OR cell does to its own
weights*, and wrong as a thing the kernel does. If the kernel normalized, **broadcast would be
impossible and AND-nodes could not exist** — and nobody would have known why.

## The one hard thing, not hand-waved

**Best-first over an AND/OR graph is AO\*, not A\*.**

An AND node is not satisfied until **all** its children are, so its cost is the *sum* of its subtree,
not the best single path. You cannot greedily commit to the best-looking child: a cheap-looking AND
node may hide an expensive sibling requirement. Naive best-first will pick bridges that look good at
the frontier and turn out to be dead ends.

**v0 sidesteps it, deliberately:**

```text
OR nodes   best-first by P.   ← this is where the real choice is. greedy is fine here.
AND nodes  expand ALL children immediately. They are CHEAP — `requires` is a short list, and
           expanding them costs an index lookup, not an inference call.
```

The expensive resource is **model calls**, and those only happen at OR nodes that come back empty. So
spend the search intelligence where the money is, and let AND nodes be dumb and eager.

**If bridges start coming back as dead ends, that is when you upgrade to real AO\*.** Same discipline
as everywhere: build the simple thing, and let the number tell you when it stops working.

## States carry no negations

You would never *observe* `skyra not-helping user`. There is no perceive that returns it, and
**absence is not falsity** (a law).

You do not need it. Closure is a **subset check**:

```text
desired  = { skyra helping user }
current  = { …whatever was observed… }        `helping` is simply NOT IN IT
Done()   = regressed ⊆ current                → false. THE GAP IS THE ABSENCE.
```

Putting `not-X` into a state re-introduces closed-world assumptions through the back door.

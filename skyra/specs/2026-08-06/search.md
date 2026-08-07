# Search

**Date:** 2026-08-06
**Status: SPEC.** Carries `2026-07-22/tokenizer.md` §5 with the vocabulary updated — that doc says
*beliefs* and *settlement*, both superseded, but its index structure and its guard survived intact.
The codec law in §7 and the measured probe in §8 are still authoritative and are not restated here.

> **One pass tokenizes. It produces the index entry and the token count.**

---

## The index

```text
TokenID  →  occurrences  →  the node that owns them
```

```go
type TokenOccurrence struct {
    NodeID   NodeID
    Position int
}
```

**Impulses and expectations are both tokenized**, which makes the index symmetric:

```text
FORWARD   an impulse arrives → tokenize → candidate impulses sharing tokens
                             → walk SUPPORT EDGES → the expectations they bear on

REVERSE   start from an expectation → its tokens → impulses that share them
```

Same structure read from either end. Nothing extra is stored for the second direction.

The forward walk is over **support edges**, which is why this works without an interpretation node:
what comes back is *expectations that had similar evidence*, not expectations that mean something
similar. That is a weaker claim and an honest one.

## Ranking

### Rare tokens discriminate. Common ones do not.

A token appearing in nearly every node separates nothing, so raw overlap ranks by stopword count and
returns the graph. Weight each token by how **narrowly** it occurs:

```text
CONTRIBUTES MOST   a token occurring in few nodes
CONTRIBUTES LEAST  a token occurring in most of them
```

### Activation breaks ties. It never multiplies.

> **Term match ranks. Activation may only order what already matched.**

Reading activation is not heating, so a multiplier would not violate the letter of *retrieval heats
nothing*. It reproduces the pathology anyway, through the ranking function instead of the write:

```text
what is active ranks higher → is retrieved more → enters more plans → stays active
```

And the bias falls exactly the wrong way. **Contradictory evidence is usually about something not
currently active** — that is what makes it contradictory. Under a multiplier the seatbelt
observation ranks beneath whatever the agent is already fixated on, which is Toy 5's sealing failure
relocated into the ranker.

So activation orders equals. It may never suppress a strong lexical match.

### The guard, inherited verbatim

> Approximate similarity may decide what gets **considered**. It may not decide what an observation
> **meant**, and it may never **settle** a prediction.

Token overlap selects candidates for INTERPRET. It does not grade, and grading is what writes
support edges.

## Cost falls out of the same pass

This answers [questions.md](questions.md) Q14 as a side effect rather than as separate work.

```text
tokenize once  →  occurrences go to the index
               →  len(tokens) IS the cost
```

**Derived, not stored on the Node.** `Nothing is stored that a walk can derive`, and the index is
already declared derived, rebuildable, deletable, and cheap to be wrong about — which is the exact
tolerance a cached token count wants. Cost on a Node would be durable state able to disagree with
the content it describes. Cost in the index structurally cannot, because it is thrown away and
rebuilt.

It also makes the codec law survivable: `2026-07-22/tokenizer.md` §7 forbids retokenizing durable
observations, and a new codec version becomes **a rebuild of a derived index**, never a migration of
stored state.

### What is measured

```text
A NODE'S COST      its OWN content. one number, from one tokenization.
A SUBTREE'S COST   not stored. the assembly walk accumulates it on descent, which is
                   what the walk is already for.
```

The descend rule reads per-node and the budget split reads recursive; they are not in tension once
the recursion belongs to the walk rather than to the node.

### Approximate, and deliberately so

Cost is model-specific and [questions.md](questions.md) Q12 commits to **two runners**, so an exact
per-runner count would make a node's cost two numbers that drift apart.

Budgets are heuristics, not contracts. One codec's count is wrong for the other runner by an amount
that only changes descend decisions which were already marginal — and a marginal descend decided
either way is not a defect.

## Why this is cheap enough to do on every impulse

From `2026-07-22/tokenizer.md` §8, measured:

```text
cached vocabulary load    ~202 ms   once per process
steady-state encode        2.2 µs   mean of 100,000 short encodes
```

Tokenization is not a material part of any latency budget here. The expensive thing was always the
boundary call, and this path never makes one.

---

## What is not decided

```text
WEIGHT              behaviors.md's descend rule reads
                    child_budget = (parent_budget − parent_cost) × weight
                    and `weight` has never been defined. COST does not supply it.

HOW MANY CANDIDATES the index returns a ranked list and nothing bounds its length.

PARAPHRASE          this is LEXICAL. "seatbelt is tight" and "still restrained" share no
                    tokens. cmd/semantic-sim exists because that limit was already known;
                    nomic-embed-text runs locally and is the obvious upgrade path.
                    exact-first is still right — measure the lexical path before adding a
                    similarity index that can only decide what gets CONSIDERED anyway.

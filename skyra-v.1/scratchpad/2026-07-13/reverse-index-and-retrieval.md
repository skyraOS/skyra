# The Reverse Index — And The Number

**Date:** 2026-07-14
**Status:** SCRATCHPAD — worked out jointly. This closes the last open arrow.
**Companions:** `expression-syntax.md` (what gets indexed), `two-graphs.md` (the cycle),
`self-authored-loop.md`, `one-kernel.md`, `commit-structure-and-rewind.md`, `README.md`

Every note before this one left the same arrow open: **`observation → projection`** — the one
edge with the entire cognitive act inside it. This is the mechanism inside that arrow.

---

# Part I — Michael

## Each Expression Node Points At A Possible Future

Each node in the expression graph carries a pointer to a possible future *perceived* world state —
its expects.

## Known And Unknown Nodes

```text
michael   engage   skyra   "want to go to the mall"
   ↑         ↑        ↑              ↑
 known     known    known      UNKNOWN nodes
```

The LLM should emit an **engage graph** — the utterance itself parsed into a subgraph — so the
interpreter can resolve those unknown nodes against the existing graph.

## The Selectivity Argument

Each node in the expression graph carries **many** world-state expects.

But *all together*, they carry a **very limited possibility space of previous observations.**

**This is where the reverse index comes in.**

---

# Part II — Claude

## It's An Inverted Index, And The Intersection Is The Retrieval

A single triple is low-information — `door:front state open` matches a thousand past observations.
The **conjunction** of an expression's expects is enormously selective. Intersect the posting lists
and the space collapses to a handful.

So retrieval is a **set operation**:

```text
posting list per triple  →  intersect  →  candidate prior episodes
```

No vector search. No similarity heuristic. No judge. The same property that made the diff free —
world-addressable IDs over exact triples — makes retrieval free.

**Honest caveat:** `exact` triples index exactly; `interpret` triples cannot be posting-list keys.
So the index is **hybrid** — the exact layer carries the load and narrows the space, and the
interpretive layer rides on top of what it returns. Another reason the crisp world stuff earns its
keep even though the fuzzy stuff is the product.

## The Engage Graph Is The Coreference Operation You Already Have

An incoming utterance becomes a subgraph:

```text
being:michael  wants  ???:go-to-mall
???:mall       is-a   place
```

Unknown nodes are **resolved against the existing graph, or minted new**. That is *exactly* the
coreference step from the diff (`{the cup}[7]` resolving to a known node).

> **One mechanism, two uses:** resolving incoming observations, and aligning the diff.

Same code, same aligner, same discipline — it may **translate**, it may never **judge**.

And a newly-minted node has **no history**, so it cannot be reverse-indexed. The first time Michael
wants something genuinely novel there is no precedent and the bridge must be built from scratch.
That is the expensive path. The second time, retrieval fires. **That asymmetry is the learning
curve** (see below).

## The Residual Is The Query

The unification worth banking:

> **The ungrounded `requires` are simultaneously (a) the residual shown to the model and (b) the
> reverse-index key.**

```text
"nothing produces  file:/src/auth.go exists true"
        ↓ as residual                    ↓ as query
   shown to the model            "how have I made this true before?"
                                          ↓
                                    posting list
                                          ↓
                              prior acts that produced it
                                          ↓
                                surfaced as precedent
```

So the multi-pass descent is **not** the model inventing bridges. It is the model **adapting
retrieved precedent**, generating from scratch only where the index comes back empty.

One object, two jobs.

## Retrieval Is Partial-Overlap, And That Is Where The Grade Lands

Exact conjunction matching returns **zero** on anything novel, even when a partial match would have
been useful. So **rank, don't filter**:

```text
score(prior) = Σ over matching triples ( overlap × ledger_weight(that edge) )
```

Top-k by how many target triples it satisfied, **weighted by how well that precedent's predictions
actually held up.** A prior episode that predicted confidently and was right gets surfaced. One that
predicted confidently and was wrong decays out of retrieval.

> **This is where the grade changes what gets recalled.**

Not "the model remembers better." **The ledger reranks the posting lists.** That is the entire
learning mechanism, stated mechanically, with no model in the loop.

And it closes the arrow:

```text
observation → projection
   = resolve incoming into the graph
   → take the target/residual triples as the query
   → intersect posting lists, rank by overlap × ledger weight
   → surface graded precedent as the framing
   → the model adapts it into the next expression
```

Nothing in that path tells the model how to think, and nothing in it bottoms out in an opinion.

---

## THE NUMBER

**Passes-to-closure is the learning curve.**

```text
first encounter    no precedent → N passes to close. expensive.
tenth encounter    index hits, precedent covers most of the plan → 1-2 passes.
```

It is free to instrument. It is already a first-class quantity in the runtime. It falls out of the
architecture rather than being bolted onto it. **The planner IS the evaluation harness.**

And it is **independent** of the other curve, which is why both are required:

```text
passes-to-closure ↓     retrieval is working — planning is getting cheaper
prediction error  ↓     the world-model is working — the bets are getting better
```

You can have the first without the second: confidently retrieving precedent that is confidently
wrong. Fast, cheap, and learning nothing.

> **Both curves down is the thesis. Either one alone is a warning.**

That is the experiment. It was derived from the retrieval mechanism, not bolted on afterward.

## The Prediction To Write Down Before Building

Per `self-authored-loop.md` — a vague plan that turns out right teaches nothing. So state it
falsifiably, in advance:

```text
Over N repeated exposures to one task class:
  passes-to-closure  decreases
  prediction error   decreases
  and the fraction of expects served by the reverse index  increases
```

If all three move, the thesis has evidence for the first time.
If passes drop but error doesn't, retrieval is surfacing confident garbage — the ledger weighting
is wrong.
If error drops but passes don't, the world-model is learning but retrieval isn't — the index is
wrong.

Either failure **localizes**. Which is the whole point of having built it this way.

## Open

- Posting-list storage: on the nucleus, or a separate index structure? (It is derived state — it can
  be rebuilt from the observation log, which the `deterministic` rule guarantees.)
- Decay in the index: do stale precedents fall out of posting lists, or only lose weight?
- How wide is top-k, and is it budgeted (the traversal budget again)?
- The `interpret` layer's fuzzy match — embeddings, or the aligner over the top-k the exact layer
  returned? (Lean: the latter. Keeps the model out of the retrieval path.)

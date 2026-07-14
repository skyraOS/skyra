# Retrieval — The Reverse Index

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

This is the mechanism inside `observation → projection` — the one arrow with the entire cognitive act
in it.

## The selectivity argument

Each node in an expression graph carries **many** world-state expects. But *all together*, they
carry a **very limited possibility space of previous observations.**

A single triple is low-information — `door:front state open` matches a thousand past observations.
The **conjunction** of an expression's expects is enormously selective.

```text
posting list per triple  →  intersect  →  candidate prior episodes
```

**Retrieval is a set operation.** No vector search, no similarity heuristic, no judge. The same
property that makes the diff free — world-addressable IDs over exact triples — makes retrieval free.

**Caveat:** `exact` triples index exactly; `interpret` triples cannot be posting-list keys. The
index is **hybrid** — the exact layer carries the load and narrows the space; the interpretive layer
rides on whatever it returns. (Another reason the crisp world stuff earns its keep even though the
fuzzy stuff is the product.)

## The residual is the query

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

So multi-pass descent is **not** the model inventing bridges. It is the model **adapting retrieved
precedent**, generating from scratch only where the index comes back empty.

One object, two jobs.

## Ranking — where the grade lands

Exact conjunction matching returns **zero** on anything novel, even when a partial match would have
been useful. So **rank, don't filter**:

```text
score(prior) = Σ over matching triples ( overlap × ledger_weight(that edge) )
```

Top-k by how many target triples it satisfied, **weighted by how well that precedent's predictions
actually held up.** A prior episode that predicted confidently and was right gets surfaced. One that
predicted confidently and was wrong decays out of retrieval.

> **This is where the grade changes what gets recalled.**
>
> Not "the model remembers better." **The ledger reranks the posting lists.**

## The arrow, closed

```text
observation → projection
   = resolve incoming into the graph        (exact first; aligner on the residual)
   → take the target/residual triples as the query
   → intersect posting lists, rank by overlap × ledger weight
   → surface graded precedent as the framing
   → the model adapts it into the next expression
```

Nothing in that path tells the model how to think, and nothing in it bottoms out in an opinion.

## The engage graph

An incoming utterance must become a subgraph before it can be indexed or diffed:

```text
michael  engage  skyra   "want to go to the mall"
   ↑        ↑       ↑              ↑
 known    known   known      UNKNOWN nodes
```

```text
being:michael  wants  ???:go-to-mall
???:mall       is-a   place
```

Unknown nodes are resolved against the existing graph — **as low-weight candidates when ambiguous,
never hard-linked** (see [expression-grammar.md](expression-grammar.md), the aligner discipline).

> **One mechanism, two uses:** resolving incoming observations, and aligning the diff.

Same code, same aligner, same discipline.

**Decomposition is async and buffered.** Per
[world-assertions-and-policy.md](../cognition/world-assertions-and-policy.md): text → triples is slow
(model calls + entity resolution) and **must not block the cognitive loop**. The decomposer fills the
observed graph; cognition runs on whatever is resolved; late edges arrive next cycle.

## The novelty asymmetry — and the learning curve

A newly-minted node has **no history**, so it cannot be reverse-indexed. The first time Michael wants
something genuinely novel there is no precedent and the bridge must be built from scratch. The second
time, retrieval fires.

**That asymmetry is the learning curve.**

```text
first encounter    no precedent → N passes to close. expensive.
tenth encounter    index hits, precedent covers most of the plan → 1-2 passes.
```

> **Passes-to-closure is the learning curve.**

Free to instrument. Already a first-class quantity in the runtime. It falls out of the architecture
rather than being bolted on: **the planner is its own evaluation harness.**

And it is **independent** of prediction error, which is why both are required:

```text
passes-to-closure ↓     retrieval is working — planning is getting cheaper
prediction error  ↓     the world-model is working — the bets are getting better
```

You can have the first without the second: confidently retrieving precedent that is confidently
wrong. Fast, cheap, and learning nothing.

**Both curves down is the thesis. Either one alone is a warning.**

## Open

- Posting-list storage: on the nucleus, or a separate index? It is **derived state** — rebuildable
  from the observation log, which LAW 1 guarantees.
- Decay: do stale precedents fall out of posting lists, or only lose weight?
- Top-k width — budgeted by the traversal budget?
- The `interpret` layer's fuzzy match — embeddings, or the aligner over the top-k the exact layer
  returned? **Lean: the latter.** Keeps the model out of the retrieval path.

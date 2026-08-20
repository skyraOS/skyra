# Credit

**Date:** 2026-08-09
**Status: SPEC.** Not built. The largest doc in the set, and the one carrying the design's central
claim: that topological learning is sufficient.

> **Errors stay local. Acceptance walks down by hops. Everything lands on an engram.**


> **REVISED BY `2026-08-11`.** Two things below no longer hold, and the whole set carries them:
>
> ```text
> COMMIT DOES NOT CLOSE   a possibility returns and stays live. only the ROOT's commit — CLOSURE
>                         — is durable. see semantics/open-until-commit.md.
> ROOT IS RETIRED         what it named is a BRANCH. a namespace is an ENGRAM.
>                         see semantics/terminology.md, which is authoritative on words.
> ```

---

## The write

> **State is written by the parent, at the parent's commit. It is the only write in the system.**

It must be the commit, because that is the first moment `used` exists — during exploration a
possibility does not yet know what it will compose from.

**Selects write nothing.** Exploration is pure gathering. The commit emits one observation upward and
one set of writes downward.

```text
EVICTION SAFETY, FOR FREE   a possibility cut off mid-exploration has written nothing, so
                            nothing happened. the phase rule engineered this with a permission
                            split; here it is just where the information lives.
```

The write covers the possibility's **whole life at once** — which children were invoked, what was
passed to each, which returns were used. An evicted possibility leaves no trace at all, which is what
*nothing happened* should mean.

**That is a claim about the graph, and it needs the world to hold too.** A leaf that spawned a
process which wrote a file has left a trace, and the graph's silence would make the record say
nothing happened — the silent-wrong-answer shape this design refuses. What closes it is staging
world effects until the root accepts: [kernel.md](kernel.md), *the chain is the transaction*.

> **Every durable write in the system is caused by a commit.** Credit fires when an ancestor commits;
> mint-on-closure is at commit; an inbound observation is not recorded until the possibility that
> used it commits. Eviction safety stops being a property you verify and becomes one you cannot
> violate.

## Invocation registers nothing. Being used does.

```text
CITED BY ANY EMISSION   → counts as USED.      the efficiency metric.
CITED BY THE COMMIT     → becomes a CHILD.     durable structure.
```

The two thresholds are deliberately different. If citation-by-select made children, `discover` would
become a permanent child of everything that ever searched and the tree would be mostly scaffolding.

> **The durable graph records what WORKED, not what was tried.**

Exploration costs nothing permanent — a possibility may try five things, use one, and only the one
attaches. Append-only stops being a liability because the filter runs **before** the write instead of
never.

## Everything durable is owned by the parent

A child engram sits in exactly one parent's map, and its `History` is the record of **that parent
invoking it**. Two parents structurally cannot share one.

> **A never writes into B's memory. A writes into the child engram A owns.**

No shared mutable state, no contention. **Locality becomes structural instead of argued**, and the
no-transfer property follows rather than being conceded: a bad expectation is independently
rediscovered as bad by every parent, because there is no "it," only copies.

### The local positive is structural. Everything else is a record.

```text
CITED BY THE PARENT      THE CHILD EXISTS. the attachment IS the signal — nothing is written.
ACCEPTED AT RANGE        an appended event, carrying its hop count.
NOT CITED / REJECTED     an appended record.
```

The first costs nothing to store because existence carries it. The cut is between **local** and **at
range**, not between good and bad.

### What it commits to: synthesis does not learn

If B's failures live on the child engram A owns, B's *own* tree never sees them. What improves is A's
selection — *stop invoking B in these conditions* — and, when that is not enough, A grows a better
sibling. That is *refinement grows the asker, not the answerer* applied to credit, and it means **all
learning is topological.**

## History is one record per invocation

Nothing collapses at write time.

```text
Invocation   span · observations · entity · host · cite
             the return, and whether the parent cited it
             acceptances, appended later by the walk
```

**Store-on-change was tried and dropped.** It was introduced to bound growth on the positive channel,
and it does not: credit records are already one per invocation, so collapsing identical returns saves
duplicate *content* without reducing the record count. It bought little and cost the per-occurrence
cite, span, and host.

So *"stable, n=10"* is a **fold**, not a field — the same rule as `nothing is stored that a walk can
derive`. Changing what counts as the same question invalidates nothing already written.

### The history partitions, and the diff carries the signal

Two returns are comparable only if they answered the same question **and came back in the same
shape**:

```text
{tests: 480} vs {tests: 500}      comparable. THE WORLD MOVED.
{tests: 480} vs {error: timeout}  not comparable. SOMETHING STRUCTURAL HAPPENED.
```

A shape change is its own event, distinct from a value change and usually more important. Errors
partition away from successes without anyone modelling errors.

**Within a partition, the DIFF is the informative object** — not the returns. `480 → 500` is the
signal; two absolute numbers are not.

**Both folds are read-time**, so this adds no field and no migration. And **nothing declares the
shape**: it is a property of neither side alone — the same expectation under Claude and under
python3.11 returns different shapes — so it belongs to the pairing and can only be observed. The
partitioned history *is* the knowledge of what this pairing returns.

Same posture as dropping the language tag, one level up: do not assert what can only be found out.

### The stored return is history. It is never a cache.

> **A stored return may be read as WHAT B SAID BEFORE. It may never substitute for INVOKING B NOW.**

*Has the door usually been open* is cheap; *is the door open right now* requires looking. A history
carrying B's last return makes reading the answer instead of asking trivially available — the stored
present the design refuses, arriving as an optimization.

The rule wants to be structural rather than remembered, and nobody has proposed how.

---

## Dispersal

> **Credit assignment is a possibility, not a mechanism.**

Same move as discovery: don't add machinery, use the primitive. The credit pass is recorded,
inspectable, and evictable like anything else.

**Nothing is ever paid to a possibility.** Possibilities are ephemeral and hold nothing durable, so
none can receive credit — including a credit possibility. No regress to stop, no exemption needed.

### Every commit disperses. The root is not special.

A possibility closes by synthesizing, and synthesizing **is** rendering verdicts on its children.

```text
ERRORS        stay LOCAL.         the immediate parent's verdict, on the child it owns.
ACCEPTANCE    WALKS DOWN by hops. an ephemeral credit possibility walks the dependency chain.
```

**A chain that never reaches its root is not ungraded — it is locally graded only.** What is missing
is the longest-range grounding, not all grounding, so the system runs on partial global results.

### Why the asymmetry is not arbitrary

```text
FAILURE   "somewhere along here it broke."   DOES NOT DISTRIBUTE. at most one did it, and
                                             marking all of them asserts more than is known.
SUCCESS   "this whole chain worked."         DISTRIBUTES. every member genuinely participated
                                             in something that worked.
```

Descending a *rejection* would be Toy 2 in the credit channel — one evidential origin becoming N
records — and the design already refuses that shape.

There is a consequence argument underneath. A false positive is self-limiting: a useless child
picking up credit from a chain that worked anyway will also sit in chains that fail. **A false
negative is not.** Something punished for another's failure gets selected less, and *being unselected
is absorbing* — nothing earns its way back if it is never tried.

**What a rejection means** is *"somewhere along this chain it failed"* — not *"E is bad."* Which is
exactly what licenses the parent to discover an alternative rather than to blame anyone.

### The walk

```text
OBSERVATION → CITES      reverse causality. where to descend.
OBSERVATION → PRODUCER   which engram to descend into.
```

```text
walk(observation, hop):
    for each c in observation.Cites:
        inv := engramFor(c.Producer).ByReturn[c.ID]
        inv.Acceptances = append(inv.Acceptances, Acceptance{Span, Hop: hop})
        walk(c, hop+1)
```

The engram set for a span is already the **causal skeleton**, because a child only exists if a commit
cited it. So the walk traverses only what actually contributed, with no filtering step of its own.

### Span is a label, not a bound

Span filtering was a hard constraint while the walk descended **durable child structure**, which
holds everything an engram ever did. Descending **constituents** does not have that problem:
observations are *instances*, so the walk lands only on invocations that actually produced this
result. There is nothing unrelated to stray into.

**One less thing that has to be right.** Span survives for reading and grouping; it is no longer what
keeps the walk correct.

### Records are events, not counters

> **An engram records the accepted event. Magnitude is derived by counting.**

Append-only without qualification — a counter is a mutation, an appended event is not.

**Trajectory comes free**, because the log is ordered: *recent* versus *ever* is a comparison over the
same structure, with no windowing scheme. A bare counter would have destroyed that — an engram that
worked for a year and has failed for a month would read identically to one always mediocre.

**And set membership survives.** With events, `{A,B,C} succeeded` and `{A,B,D} failed` are both
recoverable, so a successful subgraph can be rebuilt after the fact. That does not solve sibling
blindness; it keeps it **approachable rather than permanently foreclosed** — the set differences
counterfactual methods run on are retained. Counts would have thrown them away.

### Hop profiles are the detector

The repeated visiting is not overhead — it is one fact per accepting ancestor:

```text
{hop1, hop2, hop3}   validated locally AND at range.
{hop1, hop2, — }     good nearby, never confirmed further out.
{hop1, — , — }       cited by its parent, and nothing else has ever agreed.
```

The third profile is the self-sustaining subtree, and it reads off directly. **Not one signal
weakening with distance — signals at every scale.**

## Credit on selections is exact

A select member cites **one** observation. A commit cites **many**:

```text
SELECTION     ONE cause.    credit assignment is EXACT.
COMPOSITION   MANY causes.  the unsolved sibling-blindness problem.
```

Arity, not cleverness. And the citation supplies the condition an attachment lacks: the child says
*A used P*; the citation says *A used P **because of O***. Typed through O's producer, that is
`(Q → P)`, which recurs and carries statistics where a raw observation pair never could.

> **The pairwise correlation table is not a separate mechanism. It is credit on selections**, derived
> from the citation already in the grammar.

### Credit follows data, not decisions

> Passing an observation to a child creates a **dependency**. Being *prompted* by one without passing
> it is influence with no data flow — exploration, and it earns nothing.

**Enforced rather than agreed:** citation is derived from what appears in the commit, and you can
only carry forward what you were given.

```text
PASSED AND CITED     in the constituents. the walk finds it. CREDIT.
PASSED, NOT CITED    the child had it and didn't use it. correct that it earns nothing.
NOT PASSED           no data flow. exploratory. not creditable and not reachable.
```

Which makes `Cite` a **grouping key only**, carrying no credit of its own.

## Cardinality is a second selection signal

Nothing else measures whether the **shape** of the work was well chosen — only whether the pieces
were useful:

```text
DEPTH   the dependencies believed to be sequential.
WIDTH   the work believed safe to proceed together.
CITES   which declared dependencies were actually exercised.

A then B, and B cites A         the serialization was EARNED.
A then B, but B never cites A   possibly unnecessary serialization.
```

Available at the commit, costing no model call, landing on the parent's own record — selection
learning, in the existing channel.

## Pruning is behavioural, not structural

Append-only forbids deletion, so nothing is removed — the selection simply stops being made. And it
is finer than the child: `(Q → P)` can be pruned while `(R → P)` is kept.

## Three constraints

```text
PROPAGATION MUST FOLLOW   constituents, not durable structure. instances are unique to one
CONSTITUENTS              execution, which is what bounds the recursion.

VOLUME IS THE DATA        an engram deep in a chain is visited once per ACCEPTING ANCESTOR, so
                          the work is O(n · depth). that is not overhead — it POPULATES THE HOP
                          PROFILE. and the walk costs NO MODEL CALL.

RANKING MAY NOT           index over engrams for reach; credit for ORDERING. order what already
REPLACE THE INDEX         matched, never multiply it.
```

## What this addresses

A child attaches when cited, and citation is usage. Follow the chain up and **usage would be the only
internal signal in the system** — so a subtree deep enough that no boundary signal reached it would
sustain itself on being followed, internally coherent and undetectable from inside. Not weak
correction at depth: an active reinforcement process that does not require correctness.

Long-range records break it. Accuracy still enters at only two places — an entity crossing and a
human keeping a change — but the acceptance walk carries those inward, and **the hop profile is what
makes their absence visible.**

---

## Selection is the correction mechanism

Fabrication is not prevented. It is **starved**:

```text
WHICH ENGRAM         what gets to run
WHICH OBSERVATIONS   what it sees — and therefore what it can name
WHICH ENTITY         who interprets it
WHICH HOST           where it lands, when the host is the target rather than the interpreter
```

A child authors none of these and does not decide whether it is invoked again. **Selection happens at
the call site, not at the return.**

```text
STARVATION REQUIRES ALTERNATIVES   a parent with exactly one option for a need is not
                                   selecting, it is calling.
```

Competition selects among alternatives; **it cannot run on a population of one.** The only things
that grow the population are placement failure and ingress, so variation and selection are two
mechanisms and neither substitutes for the other.

### The root closes — acceptance is citation

> **The human's synthesis is acceptance. Accepting a change IS citation; rejecting is not-citing.**

The human is simply the outermost parent, running the same operation with a different surface.

**And it is the same event that materializes.** Acceptance stops being something the runtime watches
for after the fact and becomes the act that realizes the chain's staged world effects — which is a
better grounding for the one signal the whole pressure model rests on. *The root is not special* for
**dispersal**, which every commit performs identically; it is special for **materialization**, which
happens once per chain. Two axes, easy to conflate. See [kernel.md](kernel.md).

```text
COMPUTED    the change was kept. the diff was not reverted. the file still stands.
NOT ASKED   nobody prompts the human to grade anything.
```

Which keeps the entry point consistent with *a self-report is not grounding*, and means the whole
pressure model rests on something a runtime can actually see.

### The trade, stated plainly

```text
PREVENTION   failures are impossible.  bought with complexity.
SELECTION    failures are transient.   bought with dependence on pressure actually arriving.
```

`used/seen` is a local discriminator working at every level with no model call, and hop profiles
record acceptance at every range, so depth changes which part of the profile is populated rather than
attenuating the signal.

**The limit that remains:** something at the terminus of accepted chains has no pressure at all —
nothing above it can reject, and nothing fails to place. Correction there is external and discrete.
See [questions.md](questions.md).

---

## Rejected

```text
NODE-OWNED RECORDS        a cell written by every parent that invoked it. shared mutable state,
                          and positives lose their conditions.

SPAN-OWNED VERDICTS       one record, N joining against it. a possibility sees everything on its
                          engram's children and nothing else, so a join would make long-range
                          standing a global read during gather.

NODE TAGS                 avoids walking the chain, pays in lost context. the walk is bounded by
                          depth and costs no model call.

A TAG READ AT SELECTION   a global goodness prior. what is used gets selected, gets more
TIME                      universality, gets selected more — and it suppresses exactly the
                          unproven material that contradiction arrives as.

ELIGIBILITY TRACES        the record written during the span IS the trace. nothing transient.

HELD-OPEN SPACES ·        the signal lands on durable structure, which persists. the
FORWARD-FACING WRITERS    possibility's mortality was never the problem — it was the wrong
                          object to watch.

COUNTERS INSTEAD OF       destroys trajectory and set membership, and forecloses counterfactual
EVENTS                    credit permanently.

STORE-ON-CHANGE           bounded nothing, because credit records are already one per
                          invocation. cost the per-occurrence cite, span, and host.

DESCENDING REJECTIONS     one evidential origin becoming N records. Toy 2 in the credit channel.
```

# The Build

**Date:** 2026-08-11
**Status: BUILT.** `skyra-dna-v5` exists, runs, and has been driven by a model.

> **The substrate holds. The harness that tested it did not.**

*Revised 2026-08-11 after review. The first draft's headline finding — that nothing in the design
produces decomposition — is withdrawn: the tasks were all things a model already knows, so routing
was never necessary and nothing about it was measured. See* ***What this run did not test****.*

---

## What this is

`questions.md §1` is the only item on the critical path:

```text
FIRST BUILD   the learning substrate with everything else stubbed — hand-seeded
              possibilities, deterministic bodies, one entity, no model, no create,
              no ingress. it tests the claim everything rests on and dodges every
              other open item.
```

That was built, and then the step past it: ingress open, `discover` and `create` live, every
possibility manned by `codex`. 4,100 lines across twelve packages, plus 1,100 lines of tests, with the
deterministic substrate and the model run as two separate configurations so their results do not
contaminate each other.

```text
ids · content · graph · grammar · frame · entity · entity/{codex,runtime,scripted} · host · kernel · trace
```

**Read the code before the findings.** Every claim below is anchored to a test that asserts it or a
trace that shows it, and both are in the tree.

---

## What this run did not test, and why

**Two errors in the harness. They are recorded first because they invalidate part of what follows,
and reading the findings without them would be misleading.**

```text
THE TASKS WERE WRONG   all five asks — summarize a write-ahead log, explain a bloom filter,
                       multiply 17×23, name Python and Go's test runner and formatter — are
                       NATIVE MODEL KNOWLEDGE. a model answers them from weights.

                       so nothing about decomposition was tested. an entity one-shotting a task
                       it already knows is CORRECT, not a defect, and "why did it not decompose"
                       was never a fair question to ask of it. it had nothing to decompose FOR.

THE MODEL WAS PUT      the harness told codex it was "the OUTERMOST possibility in a chain."
AT THE ROOT            that collapses two roles world.md keeps apart. THE ROOT IS THE ASKER —
                       the human — and "you cannot invoke the thing that will judge your
                       return." a model at the root is grader and interpreter at once.

                       which is also why acceptance here was computed from the chain's own
                       output. that is not a discovery about the design. it is what the harness
                       was built to do.
```

**What a real test needs.** Asks the model *cannot* answer from weights, so routing is the only path
that closes, and **a human at the root**, so acceptance is a second instrument rather than the same
one read twice.

```text
UNAFFECTED   everything under WHAT HELD. those are the deterministic substrate tests — hand-seeded
             possibilities, scripted bodies, NO MODEL ANYWHERE. task choice cannot touch them.

NOT TESTED   §1 decomposition, §1's termination half, and the cardinality result. all three are
             consequences of asking a model to route around knowledge it already had.

STANDS       §3 · §4 · §5 · §6 · §7 · §9 · §10 — structural, and would read the same on any task
             set. §2 is real but arose in a thin index and wants re-running.
```

---

## What held

These were not obvious in advance. Each is a test in `kernel/substrate_test.go` or
`grammar/parse_test.go`, and each asserts the claim rather than covering the code.

```text
CREDIT ACCUMULATES         a leaf two levels down carries {hop1:1, hop2:1} — its own parent's
AS DESCRIBED               commit AND the root's, from one traversal per accepting ancestor.
                           acceptance walks down by hops. it works.

SELECTION MEASURABLY       a parent with two children, one of which degrades after three uses,
IMPROVES                   moved: veteran×5 → rookie×3. no weights, no rewriting, nothing but
                           what the parent SEES on its children. THE CENTRAL CLAIM, DEMONSTRATED.

PRUNING IS BEHAVIOURAL     the degraded child is still a child, still selectable, still holding
                           its whole history. it is starved, not deleted.

THE SELF-SUSTAINING        a child cited by its parent but never carried further shows
SUBTREE READS OFF          {hop1} and nothing beyond. the profile distinguishes "locally good"
                           from "confirmed at range" exactly as claimed.

EXPLORATION COSTS          five invoked, one cited, ONE attached. all five have history on the
NOTHING PERMANENT          engram the parent owns. discovery sees attempts; selection sees
                           successes.

EVICTION SAFETY            a possibility that never commits leaves ZERO records, having invoked
                           a child repeatedly. every durable write is caused by a commit, so
                           there is nothing to roll back and no recovery protocol to write.

THE CHAIN IS THE           A wrote, B was not cited, the root accepted. A's file landed, B's did
TRANSACTION                not, and B's discard is a RECORDED EVENT. materialization follows the
                           credit walk — one traversal, two jobs.

THE ORIGINATION            an orchestrator that selected and then tried to author content was
FORFEIT                    refused. fires at the first select, permanent thereafter.

MONITOR IN EMITTED         under a rig where the FIRST-emitted member returns LAST, monitor still
ORDER                      saw them in emitted order. the barrier is real and replay is stable.

TWO FORMS AND NO THIRD     17 cases. prose, a tool call, an invented attribute, a member missing
                           cite, a group-level anything — all refused. incapacity as a parsing
                           property survives contact with a model that would rather be helpful.

SHAPE PARTITIONS ERRORS    {tests:480} and {tests:500} share a partition; {error:timeout} does
FROM SUCCESSES             not. nobody modelled an error anywhere.

THE TWO ROLES ARE          the runtime-as-entity package holds an `Index` interface with one
SEPARATED BY A TYPE        read-only method. there is no *graph.Graph in it, so there is no
                           Apply, no Mint, no Disperse to reach. the honest-witness split is a
                           compile error, not a promise.
```

**Crash semantics were never tested, because there is nothing to test.** No code path writes
anything durable outside `graph.Apply`, and `Apply` is only reachable from a commit. That is the
strongest structural result in the set: the property is not verified, it is unavailable.

---

## What broke

Ordered by how much rests on them.

### 1 · Decomposition — NOT TESTED. The questions were wrong.

**This was first written up as the largest finding in the set. It is not a finding, and why it is not
is worth more than the claim was.**

Run the five asks with a bare root expectation and a capable model, and the model answers every one
of them directly. The root emits `<commit>{...}</commit>` on its first step. It never selects, so it
is a **leaf**, and a leaf originating freely is entirely legal.

> **Because every ask was something the model already knew.** Summarizing a write-ahead log does not
> need a possibility discovered for it. There was nothing to route to that the entity did not already
> contain, so declining to route was correct, and the experiment measured nothing.

```text
WHAT HAPPENS   the root one-shots the ask.
WHAT FOLLOWS   discover is never invoked. create never runs. no engram is formed.
               no child attaches. no credit disperses. no hop profile exists.
               THE ENTIRE LEARNING APPARATUS IS INERT, and nothing is wrong.
```

```text
WHAT WAS SHOWN   an entity that can answer a need from its own priors will answer it. that is not
                 a defect and it is not news.

WHAT WAS NOT     whether an entity FACED WITH SOMETHING IT CANNOT DO reaches for discover. that
SHOWN            is the actual question, and no ask in this run posed it.
```

The run then supplied the routing pressure by hand, in the root's expectation. Legitimate — an
expectation is content — but it means every model result below was produced *under an instruction to
route*, which is a different experiment from watching whether routing emerges.

**The question is still open and now has a shape.** It needs asks with no native answer:

```text
read THIS file and report what is in it       the weights do not contain it
run THIS command and use its actual output    the weights cannot produce it
write something, then read it back            requires state across invocations
answer something past the training cutoff     the weights are stale by construction
```

Under those, *did it decompose* becomes a real question, because one-shotting is no longer available.
Until then nothing here says anything about it either way.

#### The termination half — same confound

Recorded because the behaviour is worth remembering, not because it concluded anything. It showed up
only once the routing pressure had been supplied by hand, on a task the entity could have answered
directly, so what it demonstrates is unclear.

A root selected `discover`, got a good engram, invoked it, and received **a correct answer to its
actual need**. It then declined to commit — re-invoking the same child with an extra argument, going
back to `discover`, invoking again — and was evicted holding a perfectly good answer the whole time.

```text
step 1  select discover                    → reach
step 2  select the child                   → A CORRECT ANSWER, in frame
step 3  select the same child again, +1 arg
step 4  select discover again
step 5  select something else
step 6  select the first child again        → EVICTED
```

**A possibility decides when it is done, and nothing tells it.** There is no notion of *good enough*
anywhere in the set — no satisfaction signal, no cost that rises with steps, nothing a frame carries
that says stop. The design's only answer is eviction, which is a **bound, not a signal**: it fires
after the work is wasted and reports the same way whether the possibility was stuck or merely slow.

Decomposition and termination are the two ends of the same missing thing. The set specifies precisely
what an emission *may* be and says nothing about when either kind should be emitted, which is the
entire content of *"selection is the skill."*

#### Which makes the same-question fold a distribution, not a fact

Worth recording beside it: **the same ask, same code, same expectation ran three times and did three
different things** — accepted cleanly once, evicted while holding a good answer once, and livelocked
into a wrong pairing once.

`world.md` half-anticipates this: for entities we do not control, hidden state is a fact and *"the
comparison degrades — but it degrades visibly."* It does degrade visibly. What the set does not draw
out is the consequence for the folds: a partition over a model entity holds **samples from a
distribution**, so a diff inside it (`480 → 500`) may be the world moving or may be the entity
wobbling, and nothing distinguishes them. Every read-time fold in `credit.md` assumes the first.

### 2 · A wrong pairing between two natural-language artifacts fails silently

`interfaces.md` argues that no pairing check is needed, and prices cold start:

> **COLD START IS THE WHOLE COST** — send Python to a TV, it fails immediately, one record, no
> model call. compare a wrong declaration, silently believed forever.

**Observed, in the second ask of the model run:**

```text
THE NEED            "summarize what a bloom filter is and why caches use one"
THE EXPECTATION     "Explain what a write-ahead log is, how it records changes before they
                    are applied to the main database, and why databases use it…"
WHAT CAME BACK      a fluent, correct, well-structured explanation OF A WRITE-AHEAD LOG.
                    {Definition, Operation, Benefits}. no error. no complaint.
```

The argument holds only when **one side is a parser.** A compiler handed natural language errors; a
TV with no speaker fails at dispatch. Between two natural-language artifacts there is nothing that
can fail. The expectation is what the entity attends to, the argument is context, and a model given a
mismatched pair produces a confident answer to the expectation and quietly ignores the need.

**And every downstream mechanism reads it as a success.** The return is well-shaped, so it partitions
with the correct answers. It has no `error` field, so it never partitions away. If the parent cites
it — and in `bare` framing the parent is told to name what came back — it earns `Cited=true`, a
hop-1 acceptance from its parent and a hop-2 from the root.

> **A perfect credit record, for an answer to a question nobody asked.**

This is the silent-wrong-answer class the design refuses everywhere else — the projector dropping a
modality, the grant the platform layer does not realize, the declaration nobody verified — arriving
through the one door the design deliberately left open.

```text
IT IS NOT THE TERMINUS CASE   questions.md §4 is one instrument read twice. this is TWO artifacts
                              that never made contact, and the mismatch is invisible IN FLIGHT from
                              every angle the design has: shape says fine, structure says fine,
                              citation says fine.

WHAT WOULD CATCH IT           only something that reads the need and the answer TOGETHER and
                              judges. that is a grader, one level up, and the design's answer is
                              the root — which in this run was reading the same output.
```

**But the record does show it afterwards, and that is worth crediting.** The WAL engram's history
after the final run:

```text
eng-0006  Explain what a write-ahead log is…
     invoked 3 · cited 1 · hops map[1:1]
     · Benefits:str Definition:str Operation:str ×1 (used 1)
     · Benefits:str Definition:str Operation:str ×1 (used 0)
     · Benefits:str Definition:str Operation:str ×1 (used 0)
```

Three partitions with an **identical shape** and three different questions — one used, two not. The
same-question fold separated the right pairing from the two wrong ones without anything modelling a
mismatch, purely because the questions differed and the parent declined to cite. `credit.md`'s claim
that *"the partitioned history IS the knowledge of what this pairing returns"* is doing real work
here.

So the signal exists; it is just **strictly after the fact**, and it needs a parent that already
noticed. The design's detection of a wrong pairing is downstream of a human-or-model catching it, not
upstream of it.

### 3 · `discover` can never earn credit — two documents contradict each other

**The cleanest defect in the set, and it reads straight off the final run:**

```text
eng-0001  system:discover
     invoked 7 · cited 0 · hops map[]
```

Seven invocations. Every one of them produced the name that made the next selection possible. **Zero
citations, and an empty hop profile** — which is precisely the record of a useless engram, belonging
to the engram without which nothing in any chain could have been invoked at all.

It is not an implementation slip. Two documents in this set say incompatible things:

```text
discover.md   "the select CITES THAT OBSERVATION, so discover earns credit exactly when its
               answer was used, through the mechanism that already exists. NO SPECIAL CASE."

credit.md     "Which makes `Cite` a GROUPING KEY ONLY, carrying no credit of its own."
```

A parent selects `discover`, gets back a name, then selects that name with `cite=<discover's
return>`. That citation is the *select's* cite — the grouping key. The parent's **commit** cites the
work's return, because that is what it composed from. So discover's observation is cited by an
emission that carries no credit, and never by the one that does.

**And `grammar.md` saw this coming and thought `cite` was the fix:**

> An observation that supplied the *name* is never bound as an argument and never composed, so
> without a citation it falls through as unused — **and that is the observation that made the
> invocation possible at all.** `cite` fixes it in the grammar rather than in the metric.

It fixes it in the grammar. `credit.md` then makes the grammar's fix carry nothing, and the repair
is silently undone one document later.

**The consequence is not cosmetic**, because starvation runs on exactly this signal:

```text
DISCOVER'S RECORD    invoked often, never used, never accepted at range.
WHAT THAT MEANS      a parent reading its own children sees discover as the WORST thing it has.
                     with a second discover strategy in play, selection moves AWAY from whichever
                     one is working, because neither can ever show a positive.
```

Competition among discover engrams is the entire payoff of making discovery an engram rather than a
function — `discover.md` rests on it, and `kernel.md` reuses the same argument for permissions. That
competition cannot run on a signal that is structurally pinned at zero.

Note what *does* work, which sharpens it: `by-index` shows `invoked 23 · cited 17 · hops {1:17}`,
because `discover` is an orchestrator and commits by **naming** its resolver's return — a real
citation, in a commit. **The resolvers earn credit and the thing that chose among them cannot.**

**Flagged for the same pass:** discovery is currently a pure lexical index walk, and it should not be.
It wants to be **a search AND a judge** — the search narrows the index, and the judge decides whether
to *carry over* an existing expectation or *create* a new one. That also gives the constructed claim
`discover.md` asks for — *"there is nothing here that does X"* — a place to come from, which a ranker
can never produce.

**And that is not a discovery-specific design.** `possibilities.md` already names the shape:

```text
GATHER      the sequence of selects.     SELECTION is the skill.
SYNTHESIZE  the terminal composition.    SYNTHESIS is the skill.
```

Search **is** gather. Judge **is** synthesize. So "search + judge" is just the generic internal-node
shape applied to discovery, and `discover.md` frames discover-as-an-orchestrator as a choice when it
is really only what any non-leaf already is.

Which makes the change smaller than it sounds. It is not *adding a model call to discovery* — it is
**choosing a different entity to man the synthesize half**, which the design already permits per
invocation.

**The judge is another child, not a call inside the adapter.** The build puts both halves in the
runtime entity, so the judge is code — `len(reach) > 0`. Fixing that by having the runtime entity
reach a model internally would bury an inference call in an adapter, which is §2 of the questions
list arriving again. Make it a child:

```text
discover
  ├─ by-index          code.   candidates
  ├─ by-recency        code.   candidates
  ├─ judge             MODEL.  carry over, or create
  └─ create            model.  authors
```

Then the judge is selectable, gradeable and competitive like everything else, and its model call is a
real invocation with a real record rather than something hidden one layer down.

> **One wrinkle to settle before it is built: reach and credit travel separately.** `EngramRef` is
> deliberately not a `Ref`, so it never enters `Cites` and the walk cannot descend through it. A
> judge that picks candidate 2 over candidate 1 is therefore **invisible to credit** — its return
> carries reach but no data flow. For a judge to earn or lose anything, its return must carry a `Ref`
> to the observation that offered the candidate *alongside* the `EngramRef`. Otherwise it inherits
> the exact defect described above: the thing making the decision has a record showing it never
> contributed.

Not built; direction recorded.

```text
THE FIX IS A CHOICE, NOT A PATCH

  EITHER   a select's cite carries credit, and credit.md's "grouping key only" is wrong —
           but then selection credit stops being exact, which is the one place the design
           has sibling blindness solved.

  OR       an observation used to NAME is a distinct consumption from one used to COMPOSE,
           and both count. possibilities.md ALREADY SAYS THIS about used/seen — "two
           consumptions, and both count" — but only for the efficiency metric, never for
           the child-forming threshold or the walk.

The second is smaller and the set already contains half of it.
```

### 4 · A chain can discover itself, and the design forbids the obvious fix

Two sentences from different documents, each fine alone:

```text
possibilities.md   a possibility mints its durable engram ON CLOSURE.
discover.md        by-index walks THE INDEX — everything that has accumulated.
```

Together: **every engram that ever closed is a search result, including roots.** And a root's
expectation *is* the ask, so it is the single best lexical match for its own chain's need.

Observed on every ask in the first model run: `by-index` returned the root as the resolution of the
root's own need, the chain re-entered itself, and it recursed until the depth bound caught it.

The kernel can refuse the binding — it holds ancestry, and refusing is execution topology rather than
semantics, so nothing reads content. **But it cannot tell discovery why**, and that is the
interesting half:

```text
ONLY THE KERNEL KNOWS ANCESTRY     a possibility sees down; no entity spans invocations.
A POSSIBILITY MAY NEVER LOOK UP    it is what keeps the citation signal ungameable.
THE GRAMMAR HAS NO THIRD CHANNEL   monitor is void, by constraint. there is nowhere to put
                                   "not that one, it is above you."
```

So the refusal is a **dead end rather than a redirect**: discovery proposes the ancestor, the kernel
refuses the name, the parent cannot select it, and the chain spends a step learning nothing. Observed
directly — three cycles of propose-refuse-fail, then eviction.

**The cut that works is upstream of all of it.**

> **A root is an ask, not a capability.**

Mint-on-closure exists so that things which have proven they can close become discoverable. A root
has nothing to offer a later chain except the whole chain. Taking roots out of the discoverable index
resolves it without touching the grammar, without a new channel, and without a possibility ever
looking upward. The root still persists, is still reused, still owns its children, still accumulates
all of the learning — it is simply not a search result.

**The spec has to say this.** As written, the two sentences above compose into a system where every
chain can rediscover itself.

### 5 · The refinement ladder is unreachable, because every rung is amnesiac

`discover.md` gives the cost ladder and says the failure *"is cited as the prompt for a new
discovery, which may create a new possibility."*

**A possibility is one opening.** Each `discover` invocation gets a fresh frame and cannot know that
its own previous answer was handed back. So:

```text
root selects discover      → by-index returns the write-ahead-log engram
root invokes it            → gets a WAL answer for a bloom-filter need
root REFUSES to name it    → correct! it noticed.
root selects discover again → by-index returns THE SAME ENGRAM. it is deterministic and has
                              no memory.
… three times, then EVICTED.
```

The root did the right thing at every step and still starved. **Rung 3 of the ladder — grow a new
possibility — is unreachable while rungs 1 and 2 cannot remember failing.**

The repair needs nothing new. An observation carries its `Producer`, so a caller that passes back
what failed has already said which engram failed — no rejection type, no new field, no memory
anywhere.

**But getting it right took three attempts, and each failure was a different missing sentence.** That
is the useful part, because each one is a place the spec reads as complete and is not:

```text
1  THE CALLER MUST PASS      "the failure is cited as the prompt for a new discovery" — but a
   THE FAILURE BACK          CITE is a grouping key carrying no credit, and no resolver reads it.
                             what has to travel is an ARGUMENT, and the set never says so.

2  AN ORCHESTRATOR MUST      discover forwarded only its FIRST argument to its resolvers and
   FORWARD WHAT IT WAS       dropped the failure on the floor. narrowing what you hand down is
   HANDED                    legitimate scoping — and a resolver cannot exclude what it never saw.
                             THE INFORMATION EXISTED AND NOTHING ROUTED IT, which is the same
                             shape as the finding it was meant to fix.

3  "WHAT FAILED" IS          the caller may hand back THE WORK THAT WAS WRONG (producer = the
   AMBIGUOUS                 engram that did it) or THE SUGGESTION IT DISLIKED (producer =
                             discover; the engram to skip is in its REACH). the run took the
                             second reading, having never invoked anything. both are legitimate
                             and the set distinguishes neither.
```

Note carefully what this is not: it is not credit. The exclusion is scoped to one discovery, does not
walk, and lands on nobody's record. *Descending a rejection* remains refused.

> **DIRECTION, recorded not built.** Rejection wants a **short-term memory**, held for the life of
> the asking rather than in the graph, and discovery should run **with the rejections as a
> parameter**. That names what the four repairs below were each groping at separately: there is no
> rejection type, no durable record, and nothing descends — only a set of "not these," handed to the
> next discovery and discarded when the ask closes.
>
> It sits naturally beside the discovery change above: a search bounded by the rejections, then a
> model call deciding carry-over versus create.

**With all three, the ladder ran end to end for the first time** — by-index missed, by-recency
missed, `create` grew a new possibility, and the chain was accepted. That is `discover.md`'s cost
ladder working exactly as written, and it took three separate repairs to reach rung 3.

**And a fourth thing is still missing, which the third ask exposed.** The exclusion set has to be
**cumulative**, and the only place it can accumulate is the caller's frame:

```text
candidate 1 rejected  → caller passes it back  → discover excludes it, offers candidate 2
candidate 2 rejected  → caller passes back… ONLY THE SECOND  → discover excludes only that,
                        and offers candidate 1 again.
```

Observed: with two bad candidates in the index, the chain alternated between them until eviction.
Every rejection is in the caller's `Returned` — the memory exists — but the caller must remember to
forward **all of it, every time**, and nothing says so. With N bad candidates that is N growing
arguments carried by hand, and a single lapse re-offers something already refused.

> **The design has no accumulating record of what a chain has already tried.** Its only memory is
> what a possibility happens to still be holding, and its only discipline is that the possibility
> remembers to pass it on.

#### And nothing says when a resolver should give up

Related, and left entirely to a resolver's body. `discover.md` is emphatic that the valuable claim is
the constructed one:

> *"There is nothing here that does X"* is a CONSTRUCTED claim — it requires knowing what was sought
> and judging that nothing met it. A function returns a list. **Only a possibility can return a
> diagnosis.**

But nothing says how the judgement is made, and a ranking resolver will almost never make it. The
lexical one built here returns its best match however bad it is, and reports `found: false` only when
overlap is literally zero — so *"compute a numeric answer"* resolved to an explanation of
write-ahead logs, on one shared token.

A threshold is exactly the kind of magic number the set refuses elsewhere, and there is no obvious
principled one. But **without it, the diagnosis is unreachable in practice**, and every miss becomes
a wrong answer instead of a "nothing here."

### 6 · The root is minted fresh per ask, so the top of a chain can never learn

Not stated anywhere in the set, and it silently disables the central claim.

All learning is topological, and topology accumulates on a **parent's** children. If every ask mints
a fresh root, the root's child map is empty on arrival every time — so the outermost possibility, the
one whose selections matter most, is the only thing in the system that can never learn. Eight
identical asks make eight unrelated one-shot trees.

Caught only because `TestSelectionImproves` failed: with a fresh root per ask the parent chose the
degraded child eight times out of eight. With reuse-by-expectation it moved after five.

```text
BEFORE   veteran veteran veteran veteran veteran veteran veteran veteran
AFTER    veteran veteran veteran veteran veteran   rookie  rookie  rookie
```

Reuse is the same mechanism `by-index` already implements, applied at ingress. Nothing new — but it
has to be **said**, because an implementation that misses it produces a system that looks like it is
running and is learning nothing.

### 7 · Citation is lost when a model writes naturally

`interfaces.md` anticipates blended composition and predicts the credit will be **coarse**:

> BLENDED   prose merging A and B into one field. credit is coarse — because attribution genuinely
> is not there.

**It is not coarse. It is zero.** The very first real model return put its reference at the end of a
sentence:

```text
"…without being flushed to disk after every change. <obs-0001>"
```

A reference inside a prose string is not a `Ref` unless something looks inside strings. Matching only
whole-string references made that commit cite nothing — a chain that genuinely used its input
recorded no data flow at all, and *credit follows data* silently stopped being true.

Scanning inside strings fixes it and costs nothing: it widens what is **derived** and changes nothing
about what is **authorized**, because the scope check at the graph boundary already refuses any
reference the emitter never held.

> **Deriving citation from content means deciding what "appearing in the content" means, and prose
> is the common case rather than the exotic one.**

### 8 · A system expectation paired with a general entity improvises instead of failing

`interfaces.md` flags exactly this and files it as a thing to watch:

> **WATCH** — a code expectation reading a missing field fails LOUDLY; a model given the same frame
> just does worse QUIETLY. same missing input, very different credit records.

Confirmed, and worse than "does worse." A caller named `discover` with `entity=codex`; codex was
handed the opaque token `system:discover`, and rather than failing, it **improvised a plausible
emission** — selecting `discover` again. Six levels deep before the depth bound stopped it.

The runtime entity handed the same mismatch says `runtime entity cannot man expectation "…"` and
stops. **Loud where it is code, recursive where it is a model** — the same pairing, the same frame,
two failure modes with nothing in common.

Type safety being replaced by learning assumes wrong pairings are **cheap and legible**. A model
turns a wrong pairing into an expensive, well-formed, recursive one.

### 9 · The composition tax is immediate, not eventual

`possibilities.md` predicts the *"spawn a leaf to add a sentence"* tax *"will feel absurd the first
time someone hits it."*

It was hit on the third test written, doing the most ordinary thing an orchestrator does: combining
three children into one answer. An orchestrator that selected A, B and C cannot emit
`{a: <A>, b: <B>, c: <C>}` — that is origination, and the forfeit does not care that every field is a
reference.

The cost is real and the prediction was accurate. Worth noting only that it arrives immediately
rather than eventually, and that the workaround — invoke a fourth possibility whose entire job is to
hold three references — is exactly the plumbing the singular-commit rule exists to prevent.

### 10 · Three encodings of one fact

`credit.md` says the local positive is structural and costs nothing to store:

> **CITED BY THE PARENT — THE CHILD EXISTS. the attachment IS the signal — nothing is written.**

But the walk's own pseudocode records `Acceptance{Hop: 1}` on the direct cites of every commit, and
the invocation carries `Cited bool` as the local verdict. So *the parent cited this* is written three
times: the attachment, the boolean, and the hop-1 event.

All three were implemented, because all three are specified. They cannot disagree today. They are
three places for one fact, which is the shape this set refuses everywhere else, and the hop-1 entry
is the one to question — the profile `{hop1, hop2, hop3}` reads better with hop1 present, so the
redundancy may be deliberate. It is not stated either way.

---

## Two methodology notes

Both are mistakes made while building, kept because each is a trap the next implementation will hit.

```text
THE EXAMPLE IS A VALID    the prompt's illustration of a naming commit was
EMISSION                  <commit><obs-0007></commit>. the model emitted it VERBATIM, and the
                          kernel correctly refused it as out of frame. the grammar is so small
                          that an instructional example is indistinguishable from an emission.
                          use ids that cannot exist.

AN INVARIANT CHECKED      interfaces.md says of assemble/monitor: BUILD BOTH AND COMPARE. the
AGAINST A RENDERING       first version compared the two PROMPT TEXTS and reported a divergence
REPORTS ON THE RENDERER   on every step — the culprit was the literal obs-0007 in the fixed
                          instructions. the comparison must be over the FRAME. once it was, the
                          invariant HELD on every step of every run.
```

On that second one, the real result: **the monitor invariant holds trivially, and the reason is worth
knowing.** Nothing else in a frame can change mid-possibility — child histories only update at the
parent's commit — so `Returned` is the only thing that moves, and monitor pushes exactly that. The
invariant is not a live constraint today. It becomes one the moment impulses are built, which is
where it should be re-tested.

---

## Divergences from the spec, recorded rather than hidden

```text
Step RETURNS AN ERROR     interfaces.md writes `Step(x, f) Emission`. a response that parses as
                          neither form has to be representable, and the set itself says it becomes
                          "an ordinary failed invocation, observed and recorded." something must
                          carry it across. returning it keeps the failure OUTSIDE the two forms
                          rather than inventing a third to hold it.

SHAPE IS FIELD PATHS ×    the set calls this "THE ONLY REAL DECISION" and leaves it open. settled
VALUE KINDS               as the sorted set of paths paired with kinds. a nested Ref is a shape
                          ELEMENT of kind `ref`, not dereferenced. list length is a value change,
                          not a structural one.

SAME-QUESTION IS A KEY    the set leans toward "a comparison function rather than a hashed key, so
                          it stays soft." a key is what a map needs, and the fold is read-time, so
                          it can be replaced without migration. narrower than intended; flagged.

THE ROOT'S EXPECTATION    the set offers two options and picks neither. chosen: whatever was typed,
IS THE ASK                which makes the human an entity reading natural language and needs no
                          special case anywhere else.

ROOTS ARE NOT IN THE      see §4. the set implies the opposite by composition; this is the change
DISCOVERABLE INDEX        with the strongest claim on the docs.

CITES SCANS INSIDE        see §7.
STRINGS
```

---

## What the model run actually did

Seven runs, five asks each, delegating framing, `codex` manning every non-system possibility. The
first four runs are the findings above; the last is the one where the repairs are in.

**The final run, ask by ask:**

```text
1  summarize        index missed → create authored `explain-write-ahead-logs` → invoked
                    → ACCEPTED, effect materialized

2  summarize-again  index offered the WAL engram → invoked → WAL ANSWER FOR A BLOOM FILTER NEED
                    → root refused to name it, passed it back
                    → by-index EXCLUDED it, by-recency EXCLUDED it, create authored
                      `explain-bloom-filters` → ACCEPTED
                    the full cost ladder, rungs 1→2→3, for the first time.

3  arithmetic       two bad candidates in the index, and the root passed back only the most
                    recent one each time. discover alternated between them. EVICTED.

4  decompose        index missed → create authored `identify-test-and-format-tools`
                    → ACCEPTED. and it did NOT emit a group — see below.

5  impossible       cycling between the two summary engrams at the step bound.
```

**Three of five accepted, on a cold graph, with structure grown from nothing.** Ask 2 is the one
worth reading a trace of: it contains the wrong pairing, the root catching it, the refinement, and
the recovery, in one chain.

**And the recovery has a catch worth stating.** What caught the wrong answer was the root — a model,
reading the need and the answer together and judging. That is exactly the *second instrument* §2 says
is missing, except it is the same model that produced the error one level down. It worked here. There
is no structural reason it must, and the design does not claim there is: `world.md`'s *"you cannot
invoke the thing that will judge your return"* constrains the relationship, not the competence.

### Cardinality > 1 was never once emitted — but NOT TESTED

**Confounded by the same task error.** Ask 4 was *"for both Python and Go: name the standard test
runner and the standard formatter"* — which a model answers in one breath from weights. It was built
to invite a group and instead invited a one-shot, so the absence of groups says nothing about whether
cardinality is reachable.

The count is recorded because the number is worth having as a baseline, not because it concludes
anything. A real test needs two halves that each require the world — two files to read, two commands
to run — where neither can be answered without an invocation.

Across every run, counting the response half of every trace:

```text
432   selects emitted by the model
  0   with more than one member
```

Ask 4 exists to invite a group — *"for both Python and Go: name the standard test runner and the
standard formatter"* — two independent halves, neither needing the other's answer. The grammar
document explains cardinality in the fixed instructions every entity receives. The model chose a
sequence of single-member selects every single time.

**Cardinality is the concurrency declaration, and the declaration was never made.** That is not a
failure — every group awaits, so a sequence is always *correct* — but it means the claim that
*"arity is the whole declaration"* is **untested against a real interpreter**, and the group barrier
that `questions.md §9` prices so carefully has not yet been paid by anything.

It also suggests the reason, which is the same one as §1: a sequence is what you emit when you are
deciding one step at a time, and deciding one step at a time is what an entity does when nothing
rewards planning the frontier. Emitting a group requires committing to a claim — *these two do not
need each other* — before seeing either result, and nothing makes that claim worth making.

---

## Rejected while building

```text
TEACHING THE MODEL THE    tried, and it is the wrong shape. a possibility that knows what a hop
CREDIT MODEL              profile is can optimise for one. the frame shows history as facts —
                          invoked N, used N, this is what it said — and never as a score.

A `system:` TAG ON THE    would be a declaration, and the set refuses those. the runtime entity
EXPECTATION READ BY       reads its OWN expressions, which is an entity understanding what it
THE KERNEL                authored, not the kernel knowing about kinds.

FILTERING ANCESTORS       discovery is a possibility and cannot see ancestry. doing it there
INSIDE DISCOVERY          means either the kernel feeds a possibility upward information, or a
                          possibility looks up. both break something load-bearing. the cut moved
                          upstream instead — §4.

MAKING THE ROOT RETRY     eviction after N steps is the honest outcome. a retry budget hides
AUTOMATICALLY             §5 rather than fixing it, and the livelock is the finding.

AN `error` FIELD          nothing models errors. a failure is a return with a different SHAPE,
CONVENTION                and the partition falls out. this survived contact and is one of the
                          nicer results in the set.
```

---

## Against `questions.md`

The set's rule is that a question leaves that doc when it is answered. Nothing there has been edited
— this is the proposed disposition, so the edit is a decision rather than a side effect.

```text
§1   NOTHING HAS BEEN BUILT       ANSWERED. it is built and it has run. the substrate claims hold;
                                  see What Held. this question closes and the ones it was blocking
                                  are now askable — with the harness corrected first.

§2   THE UNATTRIBUTED TERM        UNCHANGED, and now concrete: `create` reaches a model to author
                                  an expectation, and that expression appears in no citation. it is
                                  the largest unattributed term in any chain that uses it.

§3   SIBLING BLINDNESS            UNCHANGED for credit. but the COST of the singular commit is
                                  worse than priced — see §9 below. an orchestrator cannot combine
                                  its children at all.

§4   THE TERMINUS HAS NO          UNCHANGED, and joined by a second route to the same place: two
     VARIATION                    natural-language artifacts that never made contact produce a
                                  confident wrong answer with a perfect credit record. §2 above.

§5   NOTHING GATES INGRESS        UNCHANGED. still open, still ungated, and now also the place a
                                  root's own expectation enters the index — which turned out to
                                  matter for a different reason. see §4 above.

§6   HISTORY MUST NOT BECOME      UNCHANGED. asserted as an open hole in terminus_test.go rather
     A CACHE                      than papered over.

§7   EVIDENTIAL BLOAT IN          CONFIRMED as the right place to watch. system engrams are by far
     SYSTEM SPACE                 the most-invoked — discover ran 3–9× per ask against 1–2 for any
                                  work engram — exactly as predicted.

§8   LEAVES LEARN NOTHING         UNCHANGED, and now measurable: every work engram in every model
                                  run was a leaf. if leaves do all the work, "leaves learn nothing"
                                  means the system learns only about ROUTING.

§9   THE GROUP BARRIER            UNTESTED. cardinality > 1 was never emitted in 432 selects, but
                                  the ask meant to provoke it was answerable from weights, so this
                                  is a bad measurement rather than a result.

§10  GRANTS                       UNBUILT. staging works and is gated by nothing.

§11  WHO MAY SELECT CREATE        ANSWERED by construction: create is a child of discover and is
                                  nameable nowhere else, so scoping does it. no rule was needed.

§13  SHOULD EVERY KERNEL          UNTESTED — no kernel method was made an engram, because none was
     METHOD BE AN ENGRAM          needed before grants.
```

**And three that were not on the list:** decomposition has no cause (§1 above), termination has no
signal, and reach has no upward bound.

---

## What to build next

```text
0  FIX THE HARNESS       asks the model cannot answer from weights, and a HUMAN AT THE ROOT.
                         nothing about decomposition, termination or cardinality can be measured
                         until both are true, and three of the write-ups above are void for
                         exactly that reason.

1  RE-ASK §1             is decomposition emergent or imposed? still open, still the question
                         everything else is downstream of — and now askable, which it was not.

2  A SECOND INSTRUMENT   §2 has no defence inside the design, and the run could not test it
                         honestly because the root was a model reading its own subtree's output.
                         with a human at the root: does the credit record ever disagree with the
                         person, and how long does it take.

3  CARDINALITY > 1       never exercised by a real interpreter. until it is, the concurrency
   AGAINST A MODEL       claim and the group barrier are both untested.

4  IMPULSES              the monitor invariant is trivially true until something can change a
                         frame mid-possibility. that is the test it was written for.

5  GRANTS                not built at all. the world boundary is entirely open: staging works and
                         is gated by nothing, and `introspect:self` is still the first grant the
                         design genuinely requires.
```

---

## The one-line version

**The learning substrate works.** Credit accumulates, hop profiles appear at range, selection
measurably improves with no weights, exploration costs nothing permanent, and eviction safety and
crash semantics are unavailable rather than verified — which is stronger. With four repairs, a cold
graph grows real structure and three of five asks close: discover misses, create authors, the work
runs, the root catches a wrong answer, refinement excludes it, and new structure gets built.

**What is genuinely broken is smaller than the first draft of this doc claimed.** `discover` cannot
earn credit — two documents contradict each other and the numbers print it. A chain could rediscover
itself. The refinement ladder needs a rejection memory it does not have. Those are real, structural,
and independent of what was asked.

**What was claimed and is not established:** that nothing produces decomposition, that nothing
produces termination, that cardinality is unreachable. All three came from asking a model to route
around things it already knew, with the model sitting in the human's chair at the root. They are open
questions, not results.

> **The hard part is right.** Credit assignment without weights, topological learning, eviction
> safety falling out of where information lives rather than being engineered — all of it survives a
> real model doing unhelpful things to it.

The lesson from the run is less about the design than about testing it: **a substrate for routing
work cannot be measured on work that needs no routing.** Every confounded finding here traces back to
that one mistake, and the corrected experiment is cheap.

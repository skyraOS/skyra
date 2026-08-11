# Skyra — The First Build, and the Surface Collapse

**Date:** 2026-08-11, later the same day.
**Scope:** what happened when the `specs/2026-08-09` set was built and driven by a model, and the
kernel surface that came out of it.
**Status: BUILT.** `skyra-dna-v5` exists, runs, and has been driven by `codex`. `skyra-dna-v4` still
stays and runs; nothing was migrated.

> **Continues [`2026-08-11-the-kernel-boundary-and-the-transaction-model`](2026-08-11-the-kernel-boundary-and-the-transaction-model.md).**
> That document ends by describing what the first build should be. This is what happened when it was
> built. **Nothing there is superseded** — the vocabulary, the mechanisms and the ten-document map
> are all still correct, and this assumes them rather than repeating them. Read that one first;
> this is the delta.

---

## What changed in one paragraph

The set was built — twelve packages, ~2,200 lines of live code — and the learning substrate holds.
Credit accumulates, hop profiles populate at range, and selection measurably improves with no
weights. Then it was driven by a model, which found four repairs and three structural defects, and
the kernel surface it actually needed turned out to be **much smaller** than the specs describe:
there is no control plane, because `Drive` is a select and `Accept` is a commit. Two things claimed
in the first report were **withdrawn** on review: the tasks used to test decomposition were all
things a model already knows, and the harness put the model in the human's chair at the root.

New documents: `specs/2026-08-09/build.md` (the report), `specs/2026-08-11/kernel-revision.md` (the
surface), and `questions.md §15`.

---

# 1 · What was built

Nothing was reused from `v3` or `v4`. The fork is by version, as the set specifies, and every
mechanism is derived from the 08-09 docs rather than from running code.

```text
ids        identity types and the one minting mechanism
content    Value · Ref · WorldRef · EngramRef · Cites derivation · shape · same-question fold
graph      Engram · Observation · Invocation · Acceptance · the witness · the credit walk
grammar    the two forms, and the parser that admits nothing else
frame      what a possibility sees · name resolution · the scope check
entity     the Interpreter boundary and the registry
  codex      marshalling a frame into a prompt; the Author bootstrap
  runtime    discover and its resolvers, manned by code
  scripted   deterministic bodies, for the substrate tests
host       capability classes — Execution · Storage — and the local platform layer
kernel     the loop · authorization · the root as transaction · staging
trace      per-run logs. NOT part of the design, and says so in its package comment.
```

```text
LIVE CODE       2,200 lines across 12 packages     (3,855 with comments — 43% comment ratio)
TESTS             812 lines
THE HARNESS       207 lines   cmd/v5, driving it against codex
```

The kernel is 540 code lines, of which **245 is the load-bearing part** — the loop plus
authorization. Against v4's 220 for the same job that is parity; the growth is staging and
materialization at 106 lines, which `kernel.md` predicted would be the largest thing it added.

---

# 2 · What held

These are the deterministic substrate tests — hand-seeded engrams, scripted bodies, **no model
anywhere**. They are the answer to `questions.md §1`, and nothing about task choice or harness design
touches them.

```text
CREDIT ACCUMULATES         a leaf two levels down carries {hop1:1, hop2:1} — its own parent's
                           commit AND the root's. acceptance walks down by hops.

SELECTION MEASURABLY       a parent with two children, one degrading after three uses, moved
IMPROVES                   veteran×5 → rookie×3. no weights, no rewriting, nothing but what the
                           parent SEES on its children. THE CENTRAL BET, DEMONSTRATED.

PRUNING IS BEHAVIOURAL     the degraded child is still a child, still selectable, still holding
                           its whole history. starved, not deleted.

THE SELF-SUSTAINING        a child cited by its parent but never carried further shows {hop1} and
SUBTREE READS OFF          nothing beyond, exactly as the hop-profile argument claims.

EXPLORATION COSTS          five invoked, one cited, ONE attached — and all five have history on
NOTHING PERMANENT          the engram the parent owns.

EVICTION SAFETY            a possibility that never commits leaves ZERO records, having invoked a
                           child repeatedly.

THE CHAIN IS THE           A wrote, B was uncited, the root accepted. A's file landed, B's did
TRANSACTION                not, and B's discard is a RECORDED EVENT.

THE ORIGINATION FORFEIT    an orchestrator that selected and then tried to author was refused.

MONITOR IN EMITTED ORDER   under a rig where the FIRST-emitted member returns LAST.

TWO FORMS AND NO THIRD     17 cases — prose, a tool call, an invented attribute, a member missing
                           cite. all refused.

SHAPE PARTITIONS ERRORS    {tests:480} and {tests:500} share a partition; {error:timeout} does
FROM SUCCESSES             not. nobody modelled an error anywhere.

THE TWO ROLES ARE          the runtime-as-entity package holds a one-method read-only interface
SEPARATED BY A TYPE        and no *graph.Graph. the write path is ABSENT, not restrained.
```

**Crash semantics were never tested, because there is nothing to test.** No code path writes durably
outside `graph.Apply`, and `Apply` is reachable only from a commit. The property is not verified — it
is **unavailable**, which is the stronger result and the clearest vindication of the active/durable
split.

> **The central bet is no longer only argued.** *All learning is topological* was the one claim the
> whole set rests on and had never been observed. It has now been observed, in the smallest possible
> setting, with no model involved.

---

# 3 · What was measured badly

**Two harness errors, found on review, that void part of the first report.** Recorded before the
findings because reading them without this is misleading.

```text
THE TASKS WERE WRONG   all five asks — summarize a write-ahead log, explain a bloom filter,
                       multiply 17×23, name Python and Go's test runner — are NATIVE MODEL
                       KNOWLEDGE. so nothing about decomposition was tested. an entity
                       one-shotting a task it already knows is CORRECT, not a defect, and it had
                       nothing to decompose FOR.

THE MODEL WAS PUT      the harness told codex it was "the OUTERMOST possibility in a chain,"
AT THE ROOT            collapsing two roles world.md keeps apart. the root is the ASKER — the
                       human — and you cannot invoke the thing that will judge your return. that
                       is also why acceptance was computed from the chain's own output.
```

**Withdrawn as NOT TESTED:** that nothing produces decomposition, that nothing produces termination,
and that cardinality > 1 is unreachable. All three came from asking a model to route around things it
already knew, with the model sitting in the human's chair.

A real test needs asks the model **cannot** answer from weights — read this file, run this command,
carry state across two invocations, answer something past the cutoff — and **a human at the root**, so
acceptance is a second instrument rather than the same one read twice.

---

# 4 · What broke, and still stands

Structural, and would read the same on any task set.

### `discover` can never earn credit — now `questions.md §15`

```text
eng-0001  system:discover
     invoked 7 · cited 0 · hops map[]
```

The record of a useless engram, belonging to the thing without which nothing in any chain could have
been invoked. An observation can be used two ways and only one earns anything:

```text
USED TO NAME       a select's cite. what made the invocation POSSIBLE.    EARNS NOTHING.
USED TO COMPOSE    a reference in the commit.                             earns credit.
```

Three documents disagree. `discover.md` says the select's cite earns discover credit *"through the
mechanism that already exists, no special case."* `credit.md` says a cite is *"a grouping key only,
carrying no credit of its own."* `grammar.md` identifies the exact failure — *"an observation that
supplied the NAME… falls through as unused, and that is the observation that made the invocation
possible at all"* — and believes `cite` repairs it. **The repair is undone one document later.**

It is not about `discover`. It is any possibility whose contribution is a **name** rather than
**data**, so it breaks discovery competition, the approval-strategy competition `kernel.md` reuses
the same argument for, and any future judge.

### A chain could rediscover itself

Mint-on-closure puts every engram that closed into the index; `by-index` walks the index; a root's
expectation **is** the ask, so it is the best lexical match for its own chain's need. Observed: the
root returned as the resolution of its own need, recursing until the depth bound.

The kernel can **refuse** the binding — it holds ancestry — but cannot **redirect** the search, because
a possibility may never look up and the grammar has no third channel. So the cut moves upstream:
**a root is an ask, not a capability**, and is held apart from what discovery searches.

> Mint-on-closure reads as one rule and has **two call sites**. Closing only one leaves roots in the
> index.

### The refinement ladder is unreachable while every rung is amnesiac

A possibility is one opening, so a fresh `discover` cannot know its own last answer was rejected. The
root rejected a bad candidate, asked again, got the same one, and starved. Reaching rung 3 needed
four things no doc says: the failure must travel as an **argument**; an orchestrator must **forward**
what it was handed; *"what failed"* is **ambiguous** between the work that was wrong and the
suggestion that was disliked; and the exclusion set must be **cumulative**, with the caller's frame as
the only memory.

With all four, the ladder ran end to end for the first time — index missed, recency missed, `create`
grew new structure, accepted.

### The others

```text
THE ROOT WAS MINTED     all learning is topological and topology accumulates on a PARENT's
FRESH PER ASK           children. a fresh root per ask means the outermost possibility can never
                        learn. measured: 8/8 wrong choices before, moved after 5 with reuse.

CITATION IS LOST IN     a model puts its reference at the end of a sentence. interfaces.md
PROSE                   predicts credit will be COARSE there; it was ZERO.

A SYSTEM EXPECTATION    codex handed `system:discover` improvised a plausible emission and
PAIRED WITH A GENERAL   recursed six levels. the runtime entity handed the same thing says so and
ENTITY IMPROVISES       stops. interfaces.md's own WATCH note, confirmed: loud where it is code,
                        recursive where it is a model.

THE COMPOSITION TAX     hit on the third test written, combining three children into one answer.
IS IMMEDIATE            the "spawn a leaf to add a sentence" tax arrives at once, not eventually.

THREE ENCODINGS OF      the child attachment, the Cited bool, and the hop-1 acceptance all say
ONE FACT                "the parent cited it."
```

### A wrong pairing between two natural-language artifacts fails silently

A bloom-filter need, invoked against a write-ahead-log expectation, returned a fluent, well-formed,
structurally valid explanation **of write-ahead logs**. No error anywhere.

`interfaces.md` prices cold start as cheap — *"send Python to a TV, it fails immediately, one record,
no model call"* — and that holds only when **one side is a parser**. Between two natural-language
artifacts nothing can fail.

The record does show it afterwards: three partitions with an identical shape and three different
questions, one used and two not. The same-question fold separated the right pairing from the wrong
ones without anything modelling a mismatch. But it is **strictly after the fact**, and it needs a
parent that already noticed.

---

# 5 · The surface collapse

**There is no control plane.** The build exported nine methods; seven should not exist, because they
were already in the design under other names.

### Drive is a select

`Drive` and a select member are **the same function** in the build — `k.open`, once with a parent and
once without. Drive is a select of cardinality 1, at depth 0. The only difference is who emits it.

### Accept is a commit

`credit.md` already said so: *"the human is simply the outermost parent, running the same operation
with a different surface."* So `Accept` and `Reject` are not two operations and not one with a flag —
they are **one commit that either carries a reference or does not**, and materialization follows
citation exactly as it does everywhere else. This retracts the earlier argument for keeping them
apart.

### And the rest were never operations

`Ask` is a composition. `rootFor` is internal. `Spiral` is a read. **`Hold` is an implementation leak**
— it exists only because mint-on-closure needs somewhere to park an engram with identity that has not
run, which is a detail of a rule, and it got exported.

### The surface, entire

```text
INGRESS       permission to enter the system.                    A CHANNEL.

SELECT        choose.                                            THE TWO FORMS.
COMMIT        internalize.

discover      find engrams · entities · observations · hosts.    TWO NAMES,
introspect    what am I doing.                                   reached BY the forms.
```

**`discover` and `introspect` are one kind of thing** — the system asking about itself — differing
only in which half they read, which is the same split everything else rests on:

```text
discover      reads the GRAPH     what HAPPENED       engrams · entities · observations · hosts
introspect    reads the KERNEL    what is HAPPENING   roots · heads · ancestry · bindings
```

So `introspect` becomes a **kernel engram**: an engram whose entity reads kernel state. `kernel.md`
already said this and the build got it wrong by making it a method.

> **An engram is not an inference call.** The entity decides whether anything infers. Measured: 29 of
> 32 system-engram invocations in the run were pure code — `by-index` 23, `by-recency` 6, zero model
> calls between them. The cost of an engram is a function call plus a history append.

### Against Linux

The syscall analogue is the grammar, and it is **two**.

```text
UNIX v1 (1971)        ~33 system calls
Linux 0.01 (1991)      67
Linux 6.x x86-64      ~350

SKYRA                    2            SELECT · COMMIT
```

The embedder's side is one channel and has no Linux analogue worth drawing.

### Ingress is root access, and it is an unscoped select

Since drive is a select, **ingress plus drive is a select with no frame to check against** — the
root's engram is in nobody's frame. That is exactly why opening a root is the maximum-privilege
operation and exactly why `questions.md §5` has no gate for it.

---

# 6 · Every entity has two faces

The runtime/kernel split is not special. It is one instance of a rule the set states three times, in
three documents, without noticing it is one rule.

```text
ENTITY      SELECTABLE FACE                      PRIVILEGED FACE
──────────────────────────────────────────────────────────────────────────────────────────
runtime     mans discover · introspect ·         THE KERNEL. pairs, drives, applies writes.
            create.
human       an interpreter, invoked DOWN.        THE GRADER AT THE ROOT. ingress, grants,
                                                 acceptance, materialization.
a model     interprets an expectation.           ITS ADAPTER. spawns, holds credentials.
```

`kernel.md` calls the human wearing both hats *"the first place the two roles meet inside one
chain."* **It is not the first place. It is every place** — the human is just the case where one
being visibly wears two hats and the runtime reads as plumbing.

**The nouns collapse; the boundary does not.** One identity with a selectable face and a privileged
face — and a possibility that can select the privileged face can influence what is recorded about
itself, which is the only safety property in the design.

And it needs no new mechanism, because the incapacity is already there:

> **The privileged face is never in any frame.**

That is capability-as-scoping applied to the entity rather than to an observation, and `world.md`
already derives it for the human: *"the grader is unreachable because nobody put it in scope, not
because anything forbids it."*

---

# 7 · What this converged on

Two faces, one of which is never in a frame, is **object-capability** — E, Capsicum, seL4. Worth
naming, because the model has decades of catalogued failure modes and checking them is cheaper than
rediscovering them.

It is **not** public/private, and the difference is the point:

```text
PUBLIC / PRIVATE   DECLARED, per-type, static.    protects against MISTAKES. defeated by
                                                  reflection — the reference exists, merely hidden.

TWO FACES          NEVER GIVEN, per-frame,        protects against ADVERSARIES. nothing to reflect
                   narrowing down a chain.        on — THE REFERENCE IS NOT THERE.
```

Arriving here from *"the citation signal must be ungameable from inside"* — a credit requirement with
nothing to do with security — is the set's own standard for a real result.

```text
CONFUSED DEPUTY      SOLVED, INDEPENDENTLY. kernel.md's "a kernel-rooted chain runs under the
                     grants of the chain that provoked it" is the textbook fix, derived sideways.

AMBIENT AUTHORITY    PRESENT AND DELIBERATE. `discover` is nameable in EVERY frame, `here` is an
                     axiom. both load-bearing, neither an accident — and both should be NAMED as
                     ambient, because that is where ocap systems break.

REVOCATION           A REAL HOLE with no home in questions.md. credit.md's "pruning is
                     behavioural, nothing is removed, the selection simply stops being made" is
                     correct for credit and insufficient as capability: a child engram formed on
                     citation is permanently nameable. starvation means nobody CHOOSES it;
                     nothing stops it being INVOKED.
```

Append-only forbids deletion and every other constraint here is an incapacity rather than an
operation, so *"add a revoke"* is the wrong shape on its face. Either reach becomes **leased rather
than given**, or compromise is outside the threat model. Not decided.

---

# 8 · What is open now

Carrying forward everything in the previous summary's *What is open*, with these changes:

```text
§1   NOTHING HAS BEEN BUILT     ANSWERED. it is built, it ran, the substrate claims hold.

§4   THE TERMINUS               UNCHANGED, and joined by a second route to the same place —
                                two natural-language artifacts that never made contact.

§7   EVIDENTIAL BLOAT IN        CONFIRMED as the right place to watch. discover ran 3–9× per ask
     SYSTEM SPACE               against 1–2 for any work engram, exactly as predicted.

§9   THE GROUP BARRIER          UNTESTED, and it was a BAD MEASUREMENT rather than a result —
                                432 selects emitted, 0 with cardinality > 1, and the ask meant to
                                provoke a group was answerable from weights.

§11  WHO MAY SELECT CREATE      ANSWERED by construction. create is a child of discover and
                                nameable nowhere else, so scoping does it. no rule needed.

§13  EVERY KERNEL METHOD        PARTLY SETTLED. there is no indirection to pay for — a code-manned
     AN ENGRAM                  engram IS the function, wrapped in a record.

§15  CREDIT SEES ONE OF TWO     NEW. see above.
     USES
```

And three that are not in `questions.md` at all:

```text
REVOCATION              reach is given and never taken back.
THE ROOTLESS COMMIT     if accept IS a commit it has nowhere to land — "the outermost parent" has
                        no engram, so nothing owns the root's Cited bool. either the human gets an
                        engram and acceptance is literally an ordinary commit, or it stays a
                        kernel-side event and the uniformity is approximate. IT IS THE LAST THING
                        BETWEEN THIS SURFACE AND ONE VOCABULARY.
DECOMPOSITION           still open, and now askable for the first time — the previous attempt
                        measured nothing.
```

---

## What to attack

```text
0  FIX THE HARNESS       asks the model cannot answer from weights, and a HUMAN AT THE ROOT.
                        three write-ups are void until both are true, and it is cheap.

1  §15                   the credit contradiction. it is a fork in the docs with two stated
                        options, and discovery competition cannot run until it is picked.

2  DISCOVERY AS          search + judge, where the judge is a CHILD rather than a call inside an
   SEARCH + JUDGE        adapter. and it is not a new pattern — search IS gather, judge IS
                        synthesize, so it is the generic internal-node shape applied to discovery.

3  REJECTION MEMORY      short-term, held for the life of the asking, passed to discovery as a
                        parameter. not credit, not durable, dies with the ask.

4  RE-ASK DECOMPOSITION  with tasks that need it. everything else is downstream of the answer.

5  GRANTS                the world boundary is entirely open. staging works and nothing checks it.
```

---

# 9 · Where it stands

**The hard part is right.** Credit assignment without weights, topological learning, and eviction
safety falling out of where information lives rather than being engineered — all of it survived a
real model doing unhelpful things to it, and the central bet has now been observed rather than
argued.

**The surface is smaller than the specs describe.** One channel, two forms, two names, and the two
names are engrams like anything else. That is the strongest structural result after the substrate:
the design's claim to be small is not rhetoric.

**What is genuinely broken is narrow and mostly decidable.** `discover` cannot earn credit — a fork
in the docs with two stated options. A chain could rediscover itself — fixed, and the spec has to say
it. Refinement has no rejection memory — four missing sentences and a direction already chosen.
Revocation does not exist — the only one that is properly open.

**And the lesson from the run is about testing more than design:** a substrate for routing work
cannot be measured on work that needs no routing. Every confounded finding traced back to that single
mistake, and the corrected experiment costs an afternoon.

```text
BUILT      the substrate, the grammar, the graph, the credit walk, the transaction model.
UNBUILT    grants · Spawn · impulses · monitor as anything but a stub · partial completion.
UNTESTED   cardinality > 1 · decomposition · the group barrier · anything with a human at the root.
```

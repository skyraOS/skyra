# Skyra — The Expectation Substrate and the Turn Toward Local Learning

**Date:** 2026-08-08
**Scope:** Everything after the 2026-07-21 snapshot: the archival of dnav2, the package
restructure, the 07-24 / 07-27 / 08-03 / 08-06 spec sets, the kernel and v4 graph as built,
and the undecided re-derivation opened 08-07.
**Status:** Project snapshot. The kernel and the v4 graph are implemented and tested. The
2026-08-06 spec set is the live design. The simplification opened on 08-07 — one traversal
method, a meta recording layer, and learning from recurrence rather than from graded judgement —
is **undecided** and nothing has been built from it.

## Executive summary

The eighteen days since the last snapshot inverted the project's centre of gravity. The July
work was implementation-led: span ownership, activation fields, and interpretation learning were
built and then described. This window is design-led. Twenty-three commits, of which roughly
twenty are specs, and two packages that between them run a single turn.

That is a real imbalance and it is recorded here as one, not smoothed over.

What the design work produced is a much smaller system. Four consolidations ran in sequence —
`2026-07-24`, `2026-07-27`, `2026-08-03`, `2026-08-06` — each governed by one inherited rule:

> **Strip complexity, do not add it. Nothing carries across by default. A doc earns entry the
> same way a mechanism does — by being required, not by being coherent with the rest of the set.**

Applied to mechanisms it deleted the interpretation node, the eight-method kernel, the result-tree
apparatus, stack-per-impulse, the world object, and the dispatch table. Applied to the docs it
turned nine into six.

The shape that survived:

```text
an impulse arrives
    ↓
one stack, entered at the standing root
    ↓
INTERPRET   grade what matches a perceive, accumulate the rest as residual
    ↓
EXPECT      what would close this node — writes nodes, the only writer of state
    ↓
the model is called HERE, outside any wave
    ↓
ACT         accumulate sends
    ↓
the exchange fires them
    ↓
effects arrive as raw impulses  →  a new wave
```

Three structures do all the work. A **kernel** that executes a Space, records that it did, and
returns the record — and which by construction cannot enumerate, compare, or interpret anything.
A **graph** where expectations are directories and actions are files, kind derived from
composition rather than declared. And **four behaviors** that hold all the arithmetic, because
the kernel holds none.

One thing blocks everything else and has since 08-03: **nothing spends the outbox.** No model call
exists, so no support edge has ever been written, the fold has never run, and calibration has never
been computed. Every quantity downstream of grading is currently authored and unchecked.

Then on 08-07 the design turned again, and further than the previous consolidations. The open
question is no longer how to build the grading substrate. It is whether to build one at all.

```text
THE PROPOSAL   collapse Perturb/Observe into one Realize, with a kernel-side loop.
               put bookkeeping in a meta recording layer above the kernel.
               drop the [-1,1] support fold entirely. keep ±1 closure.
               learn from RECURRENCE over time, not from model judgement.
```

The argument behind it is that a model asked to grade its own material has no way not to
self-confirm, and that `behaviors.md` already conceded the point by logging a detector it admits
is ambiguous. Taste is temporal; it cannot be produced by an intermediate judgement. The signal
is that an expectation had to be **recalled** — which is a mechanical fact the recorder already
sees, and which cannot exist at the moment of the call, so it cannot be gamed by the thing that
would want to.

None of that is decided. It is captured in `specs/2026-08-06/locality.md`, marked UNDECIDED and
deliberately not entered into the set's README.

## Review window

The previous summary is
[2026-07-21-span-owned-activation-and-interpretation-learning.md](2026-07-21-span-owned-activation-and-interpretation-learning.md).

| Commit | Date | Result |
|---|---|---|
| `04202da` | 2026-07-21 | dnav3: interpretation learns through the world, not through itself |
| `6031821` | 2026-07-21 | Thesis: proposal scaffold — one swapped part, four cells |
| `d5d2646` | 2026-07-21 | Thesis: drop the judge arm — the ablation was always the experiment |
| `9693df3` | 2026-07-21 | Archive dnav2 and everything that ran it out of `src/` |
| `015f530` | 2026-07-24 | Restructure Skyra packages; record the v4 traversal design |
| `b0e2b6e` | 2026-07-24 | Measurement law, formal definitions, the 2026-07-24 live set |
| `baed0fc` | 2026-07-24 | Coherence sweep and staged emission credit rule |
| `0572ec2` | 2026-07-25 | Rename `skyra-v.1` to `skyra` |
| `1957f24` | 2026-07-25 | Definitions: the loop unrolled, arcs mapped to A/p/g/ΔB |
| `851eb39` | 2026-07-28 | The 07-27 strip-down lands; the kernel is admitted |
| `2bc9cbe` | 2026-08-06 | 2026-08-03: the loop as subgraph execution; kernel and graph built |
| `d771d91` | 2026-08-06 | Program/process, and differentiation from overload |
| `696dd8f` | 2026-08-06 | Consolidate into 2026-08-06; retire 2026-08-03 |
| `c35d92f` | 2026-08-06 | The syscall surface, and the worktree gets a lifetime |
| `ce15dd6` | 2026-08-06 | Retrieval is a private expectation; the confined unit is a clone |
| `5635514` | 2026-08-06 | Runtime goes inside the VM; SDK spawn surface checked |
| `7424383` | 2026-08-06 | Two runners, and confinement varies per limb |
| `0ab9980` | 2026-08-06 | Affordance discovery — retrieve, find, build |
| `a021c9b` | 2026-08-06 | Descriptions are expectations; the gate is at admission |
| `900976c` | 2026-08-06 | Flat world, system/user space, COST gets a number |
| `17fe55b` | 2026-08-06 | The search model, and COST comes free with it |
| `c4429b4` | 2026-08-06 | No semantic embeddings — relatedness is the graph |
| `2ecd3e5` | 2026-08-06 | The boundary is two things — name the exchange |

**Uncommitted at time of writing**, on `slice-based-dispatch`:

```text
MODIFIED    specs/2026-07-27/  FLAGS · README · actions · toy-runthroughs · trust · truth
UNTRACKED   specs/2026-07-27/  expectations.md · workers.md
UNTRACKED   specs/2026-08-06/  locality.md      ← written 08-07, UNDECIDED
```

The `2026-07-27` set remains live reasoning and has not been re-derived into `2026-08-06`. Its
`truth.md` is superseded by `closure.md` and its `expectations.md` by `shape.md`; the rest —
trust, actions, workers, dynamics, and the five toys — is still cited.

## What Skyra is

The 2026-07-08 thesis still holds and has been sharpened rather than replaced:

> Skyra is a long-lived agentic substrate, not an agent pipeline.

Most agent systems put the model at the centre and build orchestration around it. Skyra puts the
model outside the loop entirely — it is called once per turn, at a boundary, on the residual that
exact machinery could not resolve. What is at the centre is a durable graph of expectations and a
kernel that walks it.

The question the project is answering has narrowed usefully since July:

```text
JULY     what is the smallest runtime physics that lets an AI being inhabit continuity?
NOW      what is the smallest substrate in which a claim about the world cannot be made
         without also being falsifiable?
```

Everything in the current design descends from one failure mode: **a system grading its own
work.** The design's response is consistent and structural rather than procedural — it does not
forbid the failure, it removes the ability to commit it.

```text
the kernel WITNESSES        because a Space recording itself can omit itself
exit codes DON'T CLOSE      because that hands the grading channel to the graded
the interpretation node      because a stored reading is a claim about the world, and a
  IS DELETED                 claim about the world is state
every expectation CARRIES    because there must be no unfalsifiable expectation to detect —
  ITS OWN PERCEIVE           there is no way to construct one
confidence and support       because authored-vs-assigned is the only calibration signal,
  ARE SEPARATE               and separation is what makes the comparison possible
```

The recurring method, stated in the kernel doc and applied everywhere since:

> **Do not enforce a rule. Remove the ability to break it.**

### The loop, unchanged since 07-27

```text
an impulse comes in
    │
    ▼
it is graded against an expectation
    │
    ▼
given that expectation, what is the next ACTION
    │
    ▼
given that action, what is the next EXPECTATION
    │
    └──────────────────────────▶ (an impulse comes in)
```

## What is built

Verified 2026-08-08. Both packages compile and their tests pass.

```text
skyra/src/skyra/spacetime/kernel                     882 lines with tests
  kernel.go · space.go · stack.go · runner.go        the substrate
  kernel_test.go · runner_test.go

skyra/src/skyra/cognitive-agent/skyra-dna-v4/graph  1519 lines with tests
  namespace.go · node.go · handle.go · walk.go       the graph
  behaviors.go · behaviors_test.go                   four behaviors, PURSUE stubbed
  namespace_test.go
```

**The kernel.** Three methods on `Space` — `Perturb`, `Observe`, `ID` — both traversal methods
closed over `Space`, so descent and ascent never leave the algebra. Frames live in a `Stack`
value rather than on Go's stack, so an execution can be inspected, stepped, and dropped.
`Perturb` iterates: a frame is asked repeatedly and pops only when it promotes nothing, which is
what lets one frame run a pipeline. The transcript is a tree of hops carrying identity and
mechanism, never content — every field a string or a bool.

Its tests are claims rather than coverage:

```text
descent then post-order ascent            the return path IS the tree
nil Perturb is base case, still observes  termination is a fixed point, not a verdict
transcript is the spine                   depth and parentage
nil realization still records the hop     the two artifacts keep different authority
grouped members never cross the kernel    3 members run · 1 hop recorded
only the boundary reaches the concrete    the invariant, as method reachability
concurrent waves share no state           no shared transcript to corrupt
```

**The graph.** Namespace, handles, walks, and the four behaviors. One turn runs.
`TestPursueIsStubbed` asserts the stub is still a stub — the tree grows and nothing fires, which
is what makes the current state safe to run with EXPECT ungated.

**Three states that were rules are now unconstructible:** an expectation without a perceive, an
edge without a grade, an action grown into an expectation.

## The 2026-08-06 set

Eleven docs. `kernel.md` is the only one that matches running code line for line.

```text
kernel           BUILT     the substrate, and what the kernel cannot do
shape            SETTLED   one family of state, three structural states, two edge kinds
execution        FOUNDATION durable vs ephemeral, ONE STACK, the worktree's lifetime
syscalls         SPEC      two traps out, one signal in, no grading verb
phases           RULE      Perturb READS. Observe WRITES.
behaviors        SPEC      the four, the one question every node asks, the assembly walk
search           SPEC      the reverse index, ranking, and COST falling out of one pass
closure          RESOLVED  closed is +1, this execution. P is retrieval.
differentiation  WORKING TOWARD  structure from overload
questions        LIVE      decisions nobody has made
locality         UNDECIDED opened 08-07. not entered in the README.
```

The load-bearing decisions, compressed:

**Composition decides kind.** Expectations are directories, actions are files, raw impulses are
empty files. Not even a mode bit is stored, so kind cannot disagree with structure. No empty
directories, because every expectation carries a perceive and a perceive is an action. Directories
are never executed. A file never becomes a directory.

**EXPECT is the only writer of state.** INTERPRET writes edges and may create nothing. What the
model may produce is a scalar on an edge or a proposal that becomes an expectation — neither is
an assertion, so every claim about the world is falsifiable by construction.

**Closed is +1, this execution.** No threshold. `P` is a lifetime fold that ranks retrieval and
settles nothing. `−1` is a finding, not an absence; `0` and absent mean nobody looked. The durable
graph holds structure and history and **no claim about the present** — *is the door open right
now* has no stored answer.

**ONE STACK limits attention, not limbs.** N processes may be alive; one subgraph is attended.
Admission is FIFO and must never consult attention.

**Route and record, never filter.** Interception makes the agent society auditable. It must not
become admission control.

**The worktree is a third lifetime.** Exit freezes it, never drops it — dropping is the scrap
branch, decided after a perceive. The confined unit is a git *clone*, not a worktree, and the
runtime goes inside the VM.

**No semantic embeddings.** A pretrained embedding is someone else's opinion about what is
similar; relatedness has to move with the person. Relatedness is the graph, and it is grounded
because every edge in it was graded.

## The critical path

**Q1, the exchange.** Nothing spends the outbox.

```text
no support values flow      →  the fold has never run
T · G · P · coherence       →  written, never exercised
calibration                 →  never computed
supplier calibration        →  nothing to compare against
the re-trap perceive        →  its grade has nowhere to land
co-occurrence               →  derived from support edges that do not exist
```

The last is the sharp version: the alternative to embeddings was designed against a substrate that
has never been written to once. The set is explicit that this makes confidence edges
"unfalsifiable decoration."

A naming collision blocks building it cleanly and is flagged but unfixed. Three things are called
*the boundary*: the BOUNDARY behavior (accumulates, inside a wave), the NAMESPACE (addressing),
and THE EXCHANGE (transport, runs after a wave, fires what accumulated). Sentences like *"ACT
accumulates the send, the boundary fires it"* mean the third. **Rename before building.**

The set also records that the exchange **can be fake before it is real** — a stub returning a
fixed grade starts support flowing and lets the fold run. That has not been done.

## The 08-07 turn — undecided

Captured in `specs/2026-08-06/locality.md`. Recorded here because it would change the shape of
everything above, and because it arrived from the interface side and the learning side
independently and met in the middle.

### One traversal method, and a loop that drives the kernel

Collapse `Perturb`/`Observe` into a single `Realize`, with the loop kernel-side rather than
inside the node. The loop produces one ephemeral Space per iteration and invokes the kernel on it.

The loop turns out to **be** the exchange. A wave ends when it promotes nothing, so a second
invocation only makes sense if something new arrived, and something new only arrives if something
fired — so a loop that does not fire has nothing to loop over. `questions.md` had already
described the exchange as *"not a Space, not a behavior — it DRIVES the kernel rather than being
driven by it."*

```text
FITS       RealizeConcrete already named "the one caller entitled to the root's concrete
           return" before anything existed to be it.
           firing between invocations makes "nothing fires inside a wave" geometric.
           the loop serializes, so ONE STACK holds.
           kernel pops on promoting nothing; loop exits on accumulating nothing. same
           fixed point, twice.

COSTS      the Reader/Writer phase split, which currently makes the read/write rule
           unbreakable rather than remembered. its stated payoff is eviction safety, and
           runner.go is currently unreachable under ONE STACK — so the split may be
           protecting a mechanism that does not run.

MUST NOT   reassemble the tree (that is Law 7 inverted, and re-inherits the deleted
           result-tree apparatus), or choose the entry (the fourth recurrence of the
           sealing failure — after queue admission, interception, and the ranker).
```

### A meta recording layer

Bookkeeping above the kernel, wrapping both levels. The argument is externality, not economy:
§4 already holds that a thing recording its own execution can omit itself, so kernel-side
bookkeeping is safe only because the kernel is trusted. The constraint that has to bind it is
**strings and bools at both levels** — it sees the accumulation going out and the results coming
back, which makes it the best-positioned component in the system to become a results channel.

### Learning from recurrence, not from judgement

The largest of the three. Drop the `[-1,1]` support fold entirely; keep `±1` closure.

The case against grading is that a model asked to score its own material will self-confirm, and
`behaviors.md` already conceded this by logging a detector it calls **ambiguous** — *"revision
rate falling → converging, OR self-confirming."* A signal whose own doc cannot distinguish
learning from delusion is not a signal.

The replacement is already in the set, scoped to one node. `behaviors.md` grades retrieval
competence by *"did the process trap AGAIN for the same condition?"* — recurrence, not
assessment. The proposal is that this is not a special case but the general one.

```text
recurrence CANNOT be computed at call time — it does not exist yet.
so there is no grader to bend, and nothing to bend it with.
```

This is the incapacity discipline reaching learning, which is the one part of the system it had
never reached.

**What it takes with it:** `P`, the normalizer question, the unpinned scale, the breadth-punishing
mean, which-edge-folds, Q6's child edge, calibration, and Toy 2's inflated fan-in. **What it
keeps:** the fold, folding over *events* rather than opinions — closure events and re-entries are
mechanical facts the recorder already sees, and a frequency over definite values has no scale
question. `shape.md`'s two model outputs collapse to one: **the model proposes, time disposes.**

**What must not go with it:** `±1` closure. It is what makes descent advance, what makes
sequential execution fall out with no `requires` field, and what fires EXPECT via *children
closed, self not closed*. Remove it and there is no frontier and no loop.

### Time, and where age lives

The kernel is time in this system — a local claim, not a physical one. Nothing changes except
through a kernel invocation, so invocations are the only events and their ordering is the only
before/after there is. A clock measures time; the thing that produces irreversible change is time.

**Append-only is the arrow.** Nothing removed, nothing rewritten, so the substrate is strictly
monotone, and *"again"* is therefore a fact rather than an inference. Recurrence-as-learning is
only possible because of it — in a mutable substrate the record of the first visit could have been
overwritten by the second.

Age is **not stored**. The biology it was checked against says the same: epigenetic clocks read a
distributed methylation pattern that exists to regulate expression, not to keep time. Age is drift
in machinery that exists for other reasons, and the number is computed by a third party who knows
what undrifted looks like.

```text
node                              the key
the record                        the value — the pattern, append-only
tick · recurrence · reliability   PROJECTIONS of it. derived, never stored.
```

Two consequences carried forward. **A byproduct cannot be gamed** — nothing writes methylation in
order to signal age, and the recurrence signal keeps that immunity only while nothing writes it
deliberately. And **interpretation needs a population**: drift is visible within one individual,
but *"nine re-entries is a lot"* requires a reference distribution. The agent's own graph is that
population, which sharpens cold start — you need N nodes with history before any single node's
history means anything.

Resetting has a shape too. The marker resets; time does not. A reset must be **a new append**
declaring an epoch, never a deletion — and it must be partial, since a total reset loses what the
node learned. `differentiation.md` may already be this mechanism: a child born of a split has a
clean recurrence record while the parent's history is preserved and frozen.

## Honest assessment

**The design has outrun the code, and the gap is now the main risk.** Twenty of twenty-three
commits in this window are specs. The critical-path item is, by the set's own admission, a stub
away — and it has not been built. Meanwhile the docs elaborated VM placement, UID mapping,
two SDK runners, a syscall surface, and an immune system, all of which sit downstream of a grading
substrate that has never been written to.

**The spec set is subject to its own central claim.** Its rule is that self-report is not
grounding and a perceive must fire. Its own perceive has never fired. That is not a criticism of
the reasoning, which is unusually careful; it is the observation that internal coherence is
exactly what the design says is insufficient.

**The 08-07 turn is a genuine simplification, not a truncation.** The signature is that it kills a
whole tier of open questions at once rather than deferring them. But it is also the second large
re-derivation in five days, and the pattern to watch is consolidation becoming the work.

**Cold start is now two-dimensional and unpriced.** No relatedness, plus taste that accrues only
over ticks, plus a reference distribution that needs N nodes before it means anything — and the
set already says early graphs are disposable. The honest question is whether enough signal
accumulates before the graph is thrown away.

## Next

```text
1  BUILD THE FAKE EXCHANGE     a stub returning a fixed answer, no model. it converts the
                              set from an argument into an experiment and tells you within
                              a day whether losing the fold costs anything.

2  RENAME THE BOUNDARY         cheap, already flagged, and blocking clean work on Q1.

3  SETTLE THE 08-07 TURN       one method or two, and grading or recurrence. these change
                              what the kernel carries, so they precede building on it.

4  A FUEL COUNTER ON THE LOOP  the one place that sees every wave. bounds unbounded
                              re-execution (Q7), the mount cycle, and a Space that never
                              returns nil — in one place.

5  BehaviorID IN THE TRANSCRIPT  a defect in built code: four hops, one identity, nothing
                              distinguishing them.
```

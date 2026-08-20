# Questions

**Date:** 2026-08-06

Decisions not yet made, carried across sessions. A question leaves this doc when it is answered,
and the answer goes to the doc that owns it.

Distinct from `2026-07-27/FLAGS.md`, which records *defects found in the design*. These are
choices nobody has made.

Renumbered on entry to this set. Six of the previous ten had been answered and were still listed
as open.

---

## 1 · The exchange

**The keystone, and the only thing left on the critical path.**

Nothing spends the outbox. No model call exists, so:

```text
no support values flow      →  the fold has never run
T · G · P · coherence       →  written, never exercised
calibration                 →  never computed — and shape.md says that is the ONLY thing
                               making confidence edges safe. right now they are decoration.
```

### Three more dependents, all added 2026-08-06

```text
SUPPLIER CALIBRATION   descriptions-as-expectations catches an overclaiming catalog only
                       by comparing authored confidence to assigned support.

THE RE-TRAP PERCEIVE   retrieval-as-expectation is graded by whether a process traps
                       again. that grade has nowhere to land.

CO-OCCURRENCE          search.md's personal relatedness model is derived from SUPPORT
                       EDGES. none exist, because nothing grades.
```

The last one is the sharp version: the alternative to embeddings was designed against a substrate
that has never been written to once.

### It is not the Namespace, and not the BOUNDARY behavior

A naming collision, caught 2026-08-06 and not yet fixed in the prose.

```text
BOUNDARY the BEHAVIOR   crossings to other agents. resolves a mount name and ACCUMULATES.
                        a Space, inside a wave, reached through Perturb/Observe.

NAMESPACE               ADDRESSING. what is reachable and by what name. also inside.

THE EXCHANGE            TRANSPORT. runs AFTER a wave, fires what was accumulated, calls
                        the model, feeds results back as impulses. not a Space, not a
                        behavior — it DRIVES the kernel rather than being driven by it.
```

`behaviors.md` already relies on the split without naming it — *"ACT: accumulate the send. **the
boundary fires it**"* — and one object cannot both accumulate and fire, because **nothing fires
inside a wave**. That is what makes eviction safe.

> Rename before building. Someone will otherwise put firing logic in the BOUNDARY behavior on the
> strength of a sentence saying the boundary fires.

### Undecided: is the model a mount?

```text
YES   asking and crossing stop being two kinds of leaving. a model call is a crossing,
      crossings are new waves correlated by WaveID, one mechanism covers both. consistent
      with what is already written.

NO    a mount points at a PEER with its own graph, namespace and expectations. the model
      has none — it is a limb, not a peer. calling it a mount imports a relationship that
      is not there.
```

Leaning **no**: keep the model as the exchange's own capability and reserve mounts for agents. It
changes what gets built, so it should be settled first.

```text
NEEDED   something that takes the outbox, does the outside work, and returns results as
         impulses. IT CAN BE FAKE BEFORE IT IS REAL — a stub returning a fixed grade
         starts support flowing and lets the fold run.
```

## 2 · What a closure event carries

**From closure.md.**

Closure is per-execution; the *event* is durable and is what `P` folds over and what the
*"this closed recently"* search reads.

```text
NEEDED   at minimum a node, a wave, and a time. enough for P to fold, and enough for
         re-execution to find.
```

## 3 · `BehaviorID` in the transcript

**A defect in built code, not a future concern.**

A turn opens the *same node* with four behaviors, so the transcript reads:

```text
0  depth=0  skyra:/     ← turn
1  depth=1  skyra:/     ← interpret
5  depth=1  skyra:/     ← expect
6  depth=1  skyra:/     ← pursue
```

Four hops, one identity, nothing distinguishing them. The transcript's one job is answering *what
actually ran*.

The constraint: **`Hop` must not learn what a node is.** The kernel cannot grow
`NodeID()`/`BehaviorID()` accessors without learning that a Space decomposes into data and methods.

```text
EITHER   the Space reports one opaque identity the minter composes
OR       the kernel gets a second read-only accessor, and we accept it
```

## 4 · Unhomed impulse

**The remaining half of Toy 5.**

An impulse arrives while the stack is standing somewhere it does not bear on.

```text
DROPPED   loses contradictory evidence, which is the toy's failure exactly
QUEUED    needs an aging rule and something that decides when to revisit
MOVES THE STACK   powerful, and the obvious way to build an attention hijack the being
                  cannot refuse
```

Direct grading means such an impulse *can* attach once the stack is in the right place. Nothing
says how it gets there.

## 5 · What structured impulses carry

Only **structured** impulses — world returns and boundary answers — need content comparable to a
perceive. Raw payloads are residual by definition and need none.

`Node.Label` may be enough, or it may need something comparable rather than merely printable. The
gap test is *"is there a node for this?"*, and that question licenses EXPECT.

## 6 · What the child edge asserts

**Retrieval only.** It no longer blocks execution — closure is one observation from the current
execution, and `P` ranks. A wrong reading produces worse retrieval order, not wrong closure.

**Candidate:** the child edge is the AUTHORED CONFIDENCE, and support multiplies through it. A
child grounded at `0.8` beneath a `0.3` confidence contributes `0.24` — *how much of this child's
grounding is mine to claim.*

## 7 · Re-execution has no bound

*"Cannot close → search history → run the prerequisite again"* could chain a long way back through
stale prerequisites. Nothing stops it.

## 8 · Differentiation's shape

**From differentiation.md.** COST now exists — [search.md](search.md) — so this is live.

```text
TWO CHILDREN OR N       forcing two means the tree records split HISTORY rather than structure
THE PARENT'S PERCEIVE   it keeps one, but is now grounded through children — redundancy, or a
                        genuinely different question?
WHAT PARTITIONS         dnav3 has clustering machinery built against the old model
DEPTH AND ATTENUATION   a heavily-used region differentiates, gets deep, and depth attenuates —
                        so the more a region is used, the quieter each part becomes
```

## 9 · Where round-robin lives

**Deferred, not forgotten.** `kernel/runner.go` is currently unreachable under ONE STACK. Whether
scheduling belongs in the kernel package at all is undecided — it is a policy, and the kernel's
value is having none.

## 10 · Retrieval behind `blocked`

**From syscalls.md, and the largest thing behind that surface.**

CONTINUE and RE-ASSEMBLE both mean *walk the graph for a value a process needs*. A retrieval walk,
heating nothing, over the active subgraph first.

**Settled:** retrieval is a private expectation per agent, in its own mount, perceived by whether
the process re-traps on the same condition — behaviors.md. What it ranks by is
[search.md](search.md): term match ranks, activation only breaks ties.

**Open:** what bounds the candidate list, and how a value gets handed back into a blocked process.

## 11 · Confinement — mostly answered, one thing left

**Checked 2026-08-06 — execution.md.** Linux namespaces hide unprivileged. macOS needs a container,
and both Docker Desktop and `apple/container` provide one; the confined unit is a **clone, not a
worktree**, and it is copied in rather than bind-mounted.

**The runtime goes inside the VM**, which turns the macOS problem into the Linux one: it owns the
clone on native filesystem and namespaces its own children. One VM, N confined processes.

```text
OPEN   which macOS runtime. apple/container's per-container VM is the better isolation
       story; Docker Desktop is the mature one. not urgent — the runtime-in-VM approach
       works on either, which is why this stopped being blocking.
```

## 12 · Which runner, and how much of it to depend on

**Checked 2026-08-06 — execution.md.** Both SDKs provide spawn, stream, resume, and
function-registered tools. The Claude SDK can remove **every** built-in (`tools: []`); Codex can
only override built-ins by name.

**Decided: support both, and let confinement vary per limb** — execution.md. Designing to the
weakest common floor would discard `tools: []` for nothing. The runtime knows which runner a limb
has and shifts how much grading leans on the diff.

```text
OPEN   nothing structural. what remains is operational: the two runners' subscription
       stories are not symmetric — Anthropic sanctions SDK use on a plan through a named
       monthly credit, while OpenAI's subscription login lives at the CLI and its guidance
       steers programmatic use toward API keys.
```

## 13 · Binding an already-admitted irrevocable tool

**Mostly answered — behaviors.md.** A description enters as an expectation and is checked
immediately after firing; revertable mismatches revert. The gate on new capability is at **tool
admission**, once per tool, not per invocation.

An isolated describing context was considered and rejected: it would have required something to
safely read hostile text, where the admission gate triggers on structure instead. **Structural beats
semantic.**

```text
OPEN   admission covers new capability, not misuse of granted capability. so: does the
       FIRST BINDING of an irrevocable tool to a new expectation deserve its own gate?
       more than once, far less than always.

ALSO   nothing decides who may enter the catalog. a trusted aggregator is not a vetted
       catalog, and find-before-build makes the catalog load-bearing.
```

## 14 · `weight` in the descend rule

**What is left of COST.** COST itself is answered — [search.md](search.md). One tokenization pass
produces the index entry and the token count; cost is derived, cached in the rebuildable index,
per-node for its own content, and deliberately approximate because Q12 commits to two runners.

What did not come with it:

```text
child_budget = (parent_budget − parent_cost) × weight
```

`weight` has never been defined anywhere in the set. It is what decides how a parent's remaining
budget is divided among children, so it is the difference between a descent that spends everything
on the first child and one that spreads.

```text
CANDIDATES   authored confidence · P · uniform · something the parent declares
```

Nothing rules any of them in. It is small compared to what COST was, and the assembly walk cannot
run without it.

---

## Answered, and where the answer went

```text
08-06  A SECOND GRADING OF THE SAME    both count. STORE TWO EDGES, FOLD ONE — the second
       IMPULSE                         is a record about the GRADER, not a second piece of
                                       evidence about the world. divergence between them is
                                       the learning signal.               → shape.md

08-06  IS THERE A CLOSURE THRESHOLD    no. closed is +1 — the most recent grade in this
                                       execution.                          → closure.md

08-06  DOES P MEASURE CLOSURE          no. P ranks for retrieval; closure is one
                                       observation.                        → closure.md

08-06  CONCURRENT STACKS               → ONE STACK.                        → execution.md

08-06  WHAT OWNS THE SESSION TIER      nothing. execution spans waves through the GRAPH,
                                       not through held state.             → execution.md

08-06  WHERE DOES INTERPRETATION LIVE  nowhere. the node is killed; EXPECT is the only
                                       writer of state.                    → shape.md

08-06  WHAT IS ASSEMBLE                a WALK with a budget, built by EXPECT.
                                                                           → behaviors.md

08-05  IS A WAVE ONE DESCENT OR STEP   neither. Perturb ITERATES; a frame is asked
                                       repeatedly and pops when it promotes nothing.
                                                                           → kernel.md

08-05  DOES A WAVE STOP AT THE AGENT   yes. crossings are new waves, correlated by WaveID.
                                                                           → execution.md

08-05  CAN A FILE BECOME A DIRECTORY   no. the runtime refuses the growth.  → shape.md

08-05  WHERE DO ACTIONS LIVE / Q       expectations are directories, actions are files,
                                       derived by composition. the entry IS the edge Q
                                       lives on.                           → shape.md

08-04  MANY ENTRIES                    a frontier is a chain, not a fan-out.
                                                                           → behaviors.md

07-23  §13 Q3 GROUPING BEHAVIOR        a Space does not encode sequential/parallel — it
                                       reads the ORDERING POLICY off the parent.
                                                                           → shape.md
```

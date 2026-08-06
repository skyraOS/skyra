# Questions

**Date:** 2026-08-03

Open questions carried forward across sessions. A question leaves this doc when it is answered,
and the answer goes to the doc that owns it — not here.

Distinct from `2026-07-27/FLAGS.md`, which records *defects found in the design*. These are
*decisions not yet made*.

---

## 1 · What does the child edge mean

**Raised 08-03. DOWNGRADED 08-06** — it no longer blocks execution, only retrieval ranking.
See [closure.md](closure.md): closure is one observation from the current execution, and `P` is a
retrieval quantity. A wrong reading of the child edge produces worse retrieval order, not wrong
closure.

**Candidate, 08-06:** the child edge is the AUTHORED CONFIDENCE — *"I am x% sure this child
completes me"* — and support multiplies through it. A child grounded at `0.8` beneath a `0.3`
confidence contributes `0.24`: *how much of this child's grounding is mine to claim.*

[graph](graph.md) said every edge is an observation (now: every edge carries a grade —
[families.md](families.md)). [execution](execution.md)'s worked example is a task decomposition:

```text
Respond
├── Retrieve Context
└── Generate Response
```

Subtask and evidential ground are different relations. `2026-07-27/trust.md` **multiplies along
the child edge**, so which relation it carries decides what `T`, `G`, and `P` compute.

One reading that makes both true, recorded not adopted: the child edge is the observation that
**licensed creating the child**. The runtime, standing at `Respond`, observed something that made
it create `Retrieve Context`, and that observation is the edge. The graph then records not only
what the agent expects but what made it expect it — which is
`2026-07-27/expectations.md`'s *motivation is derived* with a place to live.

If that is right, decomposition and grounding are the same relation seen from two ends, and
nothing needs a second edge type.

```text
NEEDED    a statement of what the edge asserts, in one sentence, that survives both the
          task-tree reading and the multiplication in trust.md
```

## 2 · Can a file ever become a directory

**Raised 08-03, replacing the previous Q2 (now answered — see below).**

Actions are files and files are leaves. Unix cannot convert a file into a directory. If a
`perceive` turns out to need decomposing, the analogy says you cannot grow entries under it —
you must build an expectation instead and let the action stay atomic.

```text
IF ATOMIC IS RIGHT    an action that needs decomposing was never an action. no change needed,
                      but the runtime must reject the growth rather than silently allow it.

IF NOT                composition-derived typing means a node with entries IS a directory, so
                      adding one entry to an action silently promotes it. that is either the
                      feature or the bug, and it is currently unspecified which.
```

The second is the live risk: nothing in the model stops it, because the type is derived from
composition rather than declared.

## 3 · Is a wave one descent or one step

**Carried from 08-03. Sharpened 08-04 — it is one question, not two.**

A session is a call stack. So is a wave: `execute` recurses, Go's stack holds the frames, and
the transcript is a chain that mirrors it. What separates the two tiers is not shape but
**whether the stack is a value you can hold**:

```text
the wave's stack      implicit, in the Go runtime, dies when Realize returns
the session's stack   REIFIED — suspendable, resumable, inspectable
```

An implicit stack cannot be paused. So Q3 and Q4 collapse into one discriminator:

```text
PAUSE WITHIN A PROCESS    block a goroutine. free. recursion stands, no kernel change.
PAUSE ACROSS A RESTART    the stack must be data. the kernel becomes a TRAMPOLINE, and
                          pending Observe calls have to be held as values rather than as
                          stack frames — which is the hard part, since the fold happens
                          on ascent.
```

`2026-07-27/kernel-alteration.md` §2 listed *"Interpreter / trampoline — the kernel repeatedly
executes returned Spaces"* among its patterns. The built kernel chose recursion. That choice IS
this question, and nothing needs revisiting unless sessions must outlive processes.

## 8 · Does a wave stop at the agent

**Raised 08-04.**

The proposal: one agent's `Observe` becomes another agent's `Perturb`, so a single wave — and a
single session — spans both.

`Observe` runs on **ascent**, which forces a choice:

```text
BATON      the handoff is at A's ROOT Observe. A is finished, B begins, the wave continues
           forward and never returns. one chain, no results channel, invariant intact.

NESTED     B descends inside A's ascent and returns INTO A. both agents' frames live at
           once — which is what "keeps a session open for either agent" describes, and
           which makes A the collector of its own send. see below.
```

**Nested collides with the Perception Invariant.** `spacetime/space.go` — *"a return value would
be a second, chosen results channel — the caller collecting the outcome of its own send."* If
B's realization flows back into A's remaining frames, that is the forbidden shape wearing
different clothes. It may still be the right thing to do; it cannot be done without deciding
that the invariant does not cover cross-agent calls, and saying why.

**Either reading needs a kernel change, for a non-obvious reason.** The transcript follows
DESCENT — a hop is recorded when `Perturb` returns a continuation. Anything reached through an
`Observe` return cannot appear in that chain. So a single transcript spanning both agents
requires the kernel to re-enter on ascent, which it structurally cannot do. Today the only
available shape is two waves with two transcripts, which is the correlation problem the proposal
exists to remove.

```text
NEEDED    whether cross-agent flow is baton or nested, and if nested, what the Perception
          Invariant means at an agent boundary
```

## 4 · What owns the session tier

**Carried from 08-03. Folded into Q3 on 08-04** — the session IS a call stack, so "what owns it"
and "is it reified" are the same question.

The stack outlives a wave and does not persist to the graph. It cannot live on the kernel without
giving back the property that dissolved concurrency (`kernel.md` §6 — the kernel holds no state
between waves).

Working assumption, not decided: the caller holds it and threads it into each wave, the way a
shell holds a working directory rather than the kernel holding it. Under the goroutine answer to
Q3 the caller is a goroutine and this needs nothing built.

## 5 · Unhomed observation

**Carried from 08-03. The remaining half of Toy 5.**

An observation arrives while the stack is standing somewhere it does not bear on.

```text
DROPPED   loses contradictory evidence, which is the toy's failure exactly
QUEUED    needs an aging rule and something that decides when to revisit
MOVES THE STACK   the observation relocates execution. powerful, and the obvious way
                  to build an attention hijack the being cannot refuse.
```

[graph](graph.md)'s direct grading means such an observation *can* now attach once the stack is in
the right place. It does not say how the stack gets there.

## 6 · Concurrent stacks

**Deferred 08-03. Narrowed to ONE STACK on 08-06** — see [execution.md](execution.md). Not merely
single-threaded: one subgraph active at a time, connected, entered at a root.

Safe because actions-as-residual made waves non-blocking; there is no long-running execution for
an arriving impulse to compete with. The ingress guarantee that several stacks were carrying moves
to a FIFO impulse queue, and the rule that matters is unchanged: **admission must not consult
attention.**

Target shape is still many processes over one filesystem. `kernel/runner.go` is the machinery for
it and is currently unreachable.

## 10 · What does an INTERPRETATION carry

**Raised and deferred 08-05. Re-framed and un-deferred the same day** — the first version asked
about the edge, which was the wrong object, and it was written before [families.md](families.md)
split the impulse plane out.

**What is settled.** `Node` is reused across both families — impulse and expectation — and
composition differentiates within each, exactly as it does for expectation/action:

```text
entries beneath, no Act    a COMPOSITE     expectation  or  interpretation
no entries, an Act         an ACTION       expectation family, leaf
no entries, no Act         a RAW IMPULSE   impulse family, leaf
```

Three structural states, no type tag. **Which family a composite belongs to is POSITION** — the
impulse plane runs across, the expectation plane runs down — not a stored marker. A marker would
be the type tag that composition-derived typing exists to avoid.

Edges need no content either: they carry `Support` or `Confidence`, and the content lives on the
nodes at each end. That part of the original struct was already right.

**What is open.** An interpretation must name something checkable against the expectation tree,
because the gap test is *"is there a node for this?"* — see the two gap kinds in
[behaviors.md](behaviors.md). `Node.Label` may be enough, or it may need something comparable
rather than merely printable.

```text
NEEDED   what makes "is there already a node saying this?" answerable. that question is
         what licenses EXPECT, so nothing structural can be built until it has a shape.
```

## 9 · Should round-robin live in the kernel

**Raised 08-05. Deferred by decision, not forgotten.**

`Runner` sits in the kernel package but is not the kernel — it holds stacks, admits, evicts, and
orders. The question is whether that belongs there at all.

```text
IN THE KERNEL     one place executes things. Runner already lives beside Stack and reads
                  nothing off a Space but Done(), so it breaks no rail.

ABOVE THE KERNEL  scheduling is a POLICY (capacity, eviction, fairness) and the kernel's whole
                  value is having no policy. a Runner in the kernel invites a smarter Runner
                  in the kernel later.
```

The second risk is the real one: nothing in the kernel today can be tuned, and a scheduler is
the obvious first thing someone would tune.

## 7 · What identifies a behavior in the transcript

**Raised 08-03. A defect in built code, not a future concern.**

[execution](execution.md) makes behavior a coupling chosen per agent. So the same node walked
with different methods is a different event, and `kernel.Hop` currently carries only
`SpaceID` — it cannot tell those apart, which breaks the one thing a transcript is for.

`2026-07-27/kernel-alteration.md` §10 asked for `BehaviorID + version`. The ID is clearly
needed. The version is the question:

```text
ID ONLY        cheap, and enough to say WHICH program ran.
ID + VERSION   enough to say the program ran AND that it was this build of it. required if
               a transcript is ever replayed against changed behavior — and replay is the
               stated reason the transcript exists.
```

`Hop` currently holds a string and two bools. Adding an ID keeps it within the rail that the
transcript carries identity and mechanism, never content. A version does too. Neither is a
results channel.

**Until this lands, the kernel's transcript is incomplete for any agent with more than one
behavior.** Single-behavior walks are unaffected, which is why it is not urgent today.

---

## Answered

```text
08-03   WHERE DO ACTIONS LIVE / WHAT HOLDS Q
        expectations are directories, actions are files, and the distinction is DERIVED
        FROM COMPOSITION rather than declared. the directory entry is the expectation →
        action edge, so Q(E,A) has a home with no carve-out and no second node type.
        licensed by 2026-07-27/expectations.md's own deferral condition: the collapse was
        held back only by concurrency, and concurrency is deferred.
        → graph.md

08-03   IS THE FALSIFIABILITY INVARIANT CHECKED OR CONSTITUTIVE
        constitutive. an expectation cannot be constructed without its perceive.
        → graph.md

08-03   ARE THERE STILL TWO NODE TYPES
        no. every durable node is an expectation, every edge is an observation.
        → graph.md

08-03   MINIMAL TRANSCRIPT / SLICE CONCURRENCY / IS NIL SUFFICIENT
        → kernel.md, and 2026-07-27/kernel-alteration.md §13 Q2, Q4, Q5
```

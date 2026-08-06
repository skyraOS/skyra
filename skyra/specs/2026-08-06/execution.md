# Execution

**Date:** 2026-08-06
**Status: FOUNDATION.** Carries `2026-08-03/execution.md` with the rules separated from the
direction — two of them applied to what is built today and were sitting under a *working toward*
heading.

> **The runtime separates what the agent knows from what the agent is currently doing.**

---

## The two structures

```text
DURABLE     THE GRAPH          structure and history. persists. evolves only when learning
                               occurs. holds NO claim about the present.

EPHEMERAL   EXECUTION STACK    a traversal over it. its top frame is the current expectation.
                               begins, pauses, resumes, terminates without altering the graph.
```

Execution is not represented by the graph. The graph is a place; execution is a walk. A process
does not become the filesystem — it traverses it, holds a position, and resolves names against
that position.

## The anchor

> **The active expectation is the execution anchor — a process's current working directory.**

Everything resolves **relative to that position**. The runtime does not reason globally; it reasons
locally from where it is standing. Which is what keeps context load small: it only gets large in
one place, *"what would close me,"* and that is a retrieval-and-plan task where a large context is
earning something.

### Handle and frame are different things

```text
HANDLE   a POSITION — where execution is standing. one stack entry.
FRAME    the NEIGHBOURHOOD that position exposes — what is in view from here.
```

A handle's position **determines** the attention frame. It is not one. `2026-07-27/workers.md`
keeps "frame"; the execution object is a **handle** — Unix's open file description, coupling a
durable node to a behavior plus the transient state of this use.

### Nothing is stored that a walk can derive

The pattern that resolved most of this set:

```text
where are we                  the frontier: where a descent stops
what is the active subgraph   what the walk touched
where does the plan live      it IS the subtree
what has fired                the transcript, queried
how do phases get ordered     nesting depth
```

Each started as *"we need a field for this"* and ended as a query. That is a reasonable prior for
the next thing that feels like it needs a home.

## ONE STACK

> **One subgraph is active at a time. It is connected, entered at a root. There is one stack.**

**This limits ATTENTION, not limbs.**

```text
ATTENTION   ONE stack. one subgraph. what the runtime is THINKING ABOUT.
LIMBS       N processes. what is happening IN THE WORLD on its behalf.
```

> **Just because one subgraph is active does not mean one process is alive.**

They were never the same resource. Reading ONE STACK as a limit on concurrent action is the
misreading to avoid; each in-flight process is addressed by its own `WaveID`, so several can be
outstanding without ambiguity about which returning effect belongs to which.

### Why one stack is safe

Stack-per-impulse existed to stop goal-directed and stimulus-driven work competing for one
pointer. That competition only matters if an execution can **run long** — blocked on the model,
blocked on the world.

Actions-as-residual removed both. A wave is graph reads and graph writes and nothing else: it runs
to completion and ends. There is nothing long-running for an arriving impulse to compete with, so
the premise of the earlier argument is gone.

### The property that has to survive

Several stacks were carrying Toy 5's ingress guarantee. Under one stack it moves to the queue:

```text
SAFE    impulses queue in ARRIVAL ORDER. every one is processed. nothing is ranked.
SEALS   the queue is ordered, or dropped, by relevance · trust · activation.
```

> **Admission must not consult attention.**

The property was never "many stacks." A FIFO queue gives it more simply. What must never appear is
a queue that decides which impulse matters — that is the sealing failure relocated into the
intake.

### Enter at the standing root

Under one stack, whatever is entered is the agent's **entire** attention for that turn.

```text
THE STANDING ROOT   always the entry. the active subgraph is the whole tree, BOUNDED BY
                    activation and closure. the bound is derived; nobody chooses it.
A SUBGRAPH ROOT     enter partway down. cheaper — and whoever picks the entry has chosen
                    what the agent attends to, by hand.
```

`chroot` remains possible and is how a resumed execution re-enters, but it should not be the
ordinary path.

## Interception: route and record, never filter

Cross-agent traffic passes through the runtime rather than going directly. The payoff is larger
than routing: **it makes the agent society auditable.** Direct channels would leave the runtime
blind to its own crossings.

```text
INTERCEPT TO ROUTE and RECORD     legitimate. which agent, and that it happened.
INTERCEPT TO FILTER by relevance  the sealing failure, at the front door.
```

*"The runtime is in charge of attention"* is true for **ordering and delivery**. It must not become
true for **admission**.

## The perceive is the interface

```text
NOT THIS RUNTIME'S CONCERN   how a delegated agent edits files behind its harness.
THIS RUNTIME'S CONCERN       what is alive within it, and what came back.
```

The runtime never inspects a process. It observes effects.

> **A process's self-report is not grounding. A perceive still has to fire.**

An agent saying *"I wrote the file"* is an expression, not an observation — dnav3's LAW 1 exactly:
*"world state does not contain it until something came back and said so."* Delegation makes this
**more** important: the more capable and opaque the actor, the more the only thing under control is
whether you looked. Letting an exit code close an expectation hands the grading channel to the
thing being graded.

**Worktree and diff is a perceive**, because the *runtime* computes it and the process does not
report it. Hold it as the fallback: a targeted perceive is cheaper and sharper, and a full diff is
what you reach for when you **cannot name what you expect** — the same condition as unplaceable
residual.

## Why the graph may be walked while it is learned

Nothing is removed or rewritten, so a stack cannot be standing on something that ceases to exist.
Learning **adds**. A traversal in flight sees the new material or not, and both are consistent —
the same property that let the kernel drop its lock, arriving from the other direction.

---

## PROGRAM and PROCESS — a direction

**Not built.** Recorded because it is where this is heading, and because the vocabulary names the
durable/ephemeral split in words that carry their own intuitions.

```text
PROGRAM   the expectation.  durable. on disk. the thing you can run again.
PROCESS   one running of it. ephemeral. has a pid, a lifecycle, an exit.
```

An expectation is not a process any more than `/bin/ls` is. It is the thing you **exec**.

```text
WaveID          the pid
the transcript  the process trace
ack = +1        it STARTED. not that it worked.
the perceive    checking the world yourself instead of trusting the exit code
in-flight       running vs exited-nonzero
kill            SIGTERM — and "can it kill itself" is just exit
sequential      `&&` versus `&`
ONE STACK       a shell with one foreground job
```

### Actions delegated to an external agent

An expectation spawns a process, and that process may be an off-the-shelf coding agent rather than
anything this runtime implements. `2026-07-27/actions.md` argued for this without saying so:

> The action worker no longer needs to speak the target's language — it needs only to know that
> running this tends to move this expectation.

The action layer improves on its own. What this runtime is for is the expectation and the
observation.

**It does not break ONE STACK.** A live process is a crossing that *streams*: the wave spawns and
returns, the process runs beyond the boundary, results arrive as impulses. The wave still never
blocks. What is new is that a crossing becomes an open channel with a lifecycle rather than a send
with a receipt.

### Interpretation is the interrupt handler

```text
CONTINUOUS   interpretation polls a running process        does not work
EVENT        the process throws when it needs something;   an interrupt IS an impulse
             interpretation answers
```

Tool-use shape, needing nothing that does not already exist.

### It is octopus-shaped

One brain holding expectation and attention; limbs with real local competence acting without
central command. The center learns *whether reaching that way works*, never how the arm does it.

### Before building any of it

What the chosen SDK exposes for spawning, streaming, interrupting, and killing a long-running
agent has to be checked rather than assumed. The shapes above are what the design wants; whether
they are available is a separate question.

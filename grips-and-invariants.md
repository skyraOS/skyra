# Why Skyra: the grips, the invariants, the thesis

Date: 2026-09-27 (first written 2026-09-01)  
Status: motivation, stable. This is the *why* the rest of the repo assumes.
Every other document names a mechanism; this one names the pain each mechanism
pays down, so the build stays pointed at a wound and not at an aesthetic. It
is written from long, first-hand time inside Codex and Claude Code: the
failures below were observed, not theorized.

> **The thesis.** Operating systems assumed one actor per user. That is no
> longer true. A user now runs several semi-autonomous agents at once, and the
> OS's principal, the process, owned by one trusted human, has no answer for
> it. Skyra's principal is the **bounded, auditable agent**, not the process.
> That is the whole move; everything below follows from it.

**The move is coupling, not invention.** Permissions exist. Plans and specs
exist. Audit logs exist. Skyra invents none of them. It makes them first class
and welds them *into* the execution model, because until now they were bolted
on beside it: a permissions prompt here, a plan-mode checklist the model can
ignore there, a log someone chose to write. None of them is an edge in the
graph that actually runs. Coupling them to execution is the contribution.
Claiming less makes the claim truer.

## Six grips, three invariants

| # | Grip: the observed failure | Invariant it forces |
|---|---|---|
| 1 | **Skills do too much.** A skill is a markdown document re-read every turn, loosely wired to its tools. The pairing of tool and prompt is never a durable object, so the whole interface is re-engaged every time | **I. Cognition is compiled and remembered, not re-derived** |
| 2 | **The same work is paid for again and again.** The same workflow is re-derived, the same long documents re-sent, every step. It costs real money | I |
| 3 | **The plan is not first class.** Plan mode is a checklist the model can silently skip. A failed plan should be an event, not a line quietly dropped | **II. Every action is a typed claim; deviation is a fault** |
| 4 | **Shell access is velocity bought with ambient authority.** One `rm -rf` from disaster, and no legible record of what ran or why | **III. Authority is one list that only narrows, enforced by the OS; provenance is structural** |
| 5 | **Five agents cannot be watched by hand.** Reading all the code is impossible at scale, and losing fine-grained control is fine. But you must keep a high-level view of what each agent is doing and, above all, what it *cannot* touch | III |
| 6 | **Tool acceptance fatigue.** Every effect asks: allow this command, this tool, this file? Nobody reads the fortieth prompt. People approve by reflex, then switch to full auto, and the prompts become ambient authority. An agent that writes its own tools makes it worse: every new tool is new code to approve | III, and the approval moves: **the human grants the Horizon once; anything built inside it, agent-written tools included, runs without asking** |

## The invariants, made concrete

### I. Cognition is compiled and remembered, not re-derived

The unit of work is a durable pairing of a capability and how to use it, not
a document re-read every turn. A workflow that ran once, and succeeded, should
run again warm: without re-deriving it, and without paying the model to
rediscover it.

- **In the design:** memoized procedures keyed by input contract
  ([`design/durable-storage.md`](design/durable-storage.md)); a resolve
  record's completion hash names a whole finished sub-computation and is the
  natural memo key ([thought 13](design/thoughts/13-dependencies.md)).
- **Built:** not yet. Deliberately deferred past the December evaluation.

### II. Every action is a typed claim; deviation is a fault

Nothing an agent emits is trusted as a statement about itself. It supplies
operands, which capability and what payload, and the runtime checks them
against a typed contract. The plan cannot be ignored because it is not a
checklist beside the execution: it is the type the execution must fit. Drift
is not forbidden, which would kill adaptivity; drift that fits no contract is
a fault, never silence.

- **Built:** targets never say "invoke", "return" or "fault". `interpret`
  derives the state from operands ([thought 14](next/14-interpreted-not-stated.md)).
  Every capability has an input contract and a return contract, and a payload
  that does not fit is a typed `contract` fault, committed and replayable.
  The emitter is never the checker.
- **Not built:** negotiating the contract with the human *before* acting
  (the plan as a proposal the human accepts), and learning contracts for
  outputs nothing has seen before (novelty interrupts).

### III. Authority is one list that only narrows, enforced by the OS; provenance is structural

An agent fires nothing without a Horizon handed to it by its parent. There is
exactly one list, the Horizon: what an agent *cannot* touch is its complement,
read off the same list for free. A second list is the exact overload this
project exists to delete. Across many agents, the only composed question a
human asks is where two Horizons overlap: set intersection over one
primitive, not a new artifact. And the audit trail is a consequence of how
the system runs, not a log bolted on after.

- **Built:** a child's Horizon is always a subset of its parent's; a
  capability a parent does not hold cannot be delegated. Each invocation runs
  under a Seatbelt sandbox generated from its Horizon, and the runtime's own
  state is always denied. Every step is a content-addressed, hash-linked
  record, and replay re-derives every one of them from its stored input:
  truth is derived, not stored.
- **Not built:** preventing (not only detecting) what Codex runs natively
  inside its own sandbox. Self-extension: an agent registering a tool it wrote
  as a leaf under a narrower Horizon, and running it without approval. Today
  the registry is read once, at genesis.

## The loop that couples all three

```text
PROPOSE  the human and a model agree what is wanted and the Horizon it may touch: a typed contract
ACT      the model supplies operands; the runtime derives every step, checks every contract, and runs
         every tool as its own sandboxed leaf. The model is present only where judgment is needed
ASK      if it needs more than its Horizon grants, it does not widen: an authority fault climbs to the
         human, who grants a new, narrower Horizon or refuses. Escalation is renegotiation, never a grab
FAULT    the world returns something no contract expects: a typed fault, back to the one who can fix it
REMEMBER what worked is kept, so the next run starts warm (invariant I)
```

The model is present at the cold start and whenever something new trips a
contract. A remembered run executes on its contracts, until the world
surprises it.

## Recurrence is the resource

Remembering is the one mechanism that makes "you did this before" both cheaper
and safer, two faces of one coin:

| | Cost | Safety |
|---|---|---|
| **First run (cold)** | expensive: everything is derived | loosely guarded: the contract is still being discovered, which is exactly when a human most wants to watch |
| **Nth run (warm)** | cheap: nothing is re-derived | tightly guarded: the remembered contract is a structural type, and any violation faults sharply |

Cost and safety are not two features. They are one property that accrues
with repetition.

## Scope honesty: what is pain-grounded, and what is a bet

Every grip above is single-machine and multi-agent. None of them needs
distribution to hurt.

| | What | Rule |
|---|---|---|
| **Pain-grounded** | the single-machine substrate: bounded authority with one list, typed contracts, structural provenance, and compiled cognition | protect this; it is what ships |
| **A bet** | the distributed tier: capabilities verified across hosts, records replicated between machines | wanted, not driven by any grip yet. It may lag; it must not front-run the core |

Ranking, not cutting: keep the whole vision, but when time is short, the
pain-grounded core is what ships.

## Where the code stands

The thesis leads and the code follows. The risk was never a thesis ahead of
the code; it was a thesis that kept moving, so the code never had a still
target. In September it stopped moving, and the code has been converging on
it since:

| Invariant | Status |
|---|---|
| **III.** One narrowing list, enforced by the OS; structural provenance | **built** in the stateless controller, and running real Codex conversations |
| **II.** Every action a typed claim; deviation a fault | **built** at the level of each action; the negotiated plan is next |
| **I.** Compiled, remembered cognition | **designed**, deferred until the authority model is evaluated |

The next revision of this document should come from the code pushing back,
not from a nicer ontology. See [`README.md`](README.md) for what runs today,
and [`next/research-question.md`](next/research-question.md) for how it will
be measured.

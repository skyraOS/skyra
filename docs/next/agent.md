# An agent: think, provision, act

Date: 2026-10-24  
Status: proposed, not built. Replaces the pairing of [think](think.md) and
[act](act.md) as written on 2026-09-28; both are rewritten to match. Builds
on the kernel/interpreter split in
[interpreter primitives](../../ideas/interpreter-primitives.md) (PR #43) and
on [delegation](../design/DELEGATION.md).

## The claim

> An agent is not a model. It is a composition of procedures under Horizons,
> some of which invoke a model. The model proposes; the system provisions;
> the human puts a proposal into motion.

```text
think ──proposal──▶ human ──sets in motion──▶ provision ──starts──▶ act (a tree)
  ▲                                                                   │
  └──────────── a fault past the tree's root is a new proposal ───────┘
```

| Part | What it is | Model? | Authority |
|---|---|---|---|
| **think** | reads, converses, returns a proposal | yes | read: its parent's Horizon, and any read it is granted |
| **proposal** | an observation: the arguments to provision's contract | — | none: it is data |
| **provision** | checks the proposal and builds the tree it describes | **never** | narrow and delegate, within the human's grant |
| **act** | the tree of delegations provision built, now running | yes, per node | each node's own Horizon, set by its position |

## One rule underneath: the proposer is never the provisioner

This is the kernel's first invariant (the emitter is never the checker)
applied one level up.

| Kernel | Interpreter |
|---|---|
| a target emits bytes | think emits a proposal |
| intake checks them and derives the state | provision checks it and builds the tree |
| nothing emitted can put the system into a state | nothing proposed can create authority |

Think holds no authority to pass on: it cannot narrow, delegate or extend its
own Horizon. It can only produce arguments. Creating Horizons and delegations
is provision's job, done by a procedure, deterministically, before any model
in the tree runs.

## A proposal is an observation

A proposal is a set of arguments meant to fill one contract: provision's
input. Its fundamental components are therefore exactly what that contract
requires, and nothing else. A first version:

```json
{
  "root": {
    "actor": "claude",
    "goal": "…",
    "reach": ["repo.write", "test"],
    "authority": { "write": [{ "path": "/ws/pkg/foo", "scope": "subpath" }] },
    "leaves": 4,
    "children": [
      { "actor": "codex", "goal": "…", "reach": ["repo.write"], "children": [] }
    ]
  }
}
```

| Per node | Binding? |
|---|---|
| **position** (its parent) | yes: it decides who can grant it more |
| **actor** | yes: chosen from what think read in its parent's Horizon |
| **Horizon** (reach and authority) | yes: rendered to the sandbox |
| **leaf budget** | yes: counted by the runtime ([act](act.md#the-leaf-budget)) |
| **goal** | **no**: the prompt think wrote for the node's model |

**The goal is advisory, the Horizon is binding.** Think reads untrusted
content and writes every goal, so the goal is the path an injection takes. An
injected goal can make a node *try* something; only the Horizon decides what
can succeed. Nothing in a goal is ever read as authority.

**The proposal declares the tree.** Provision builds the nodes the proposal
names and no others. Inside a node, nothing else is fixed: no order, no
control flow, only the Horizon attached to that node.

## Approval is about intent, not authority

Provision's Horizon is the human's grant, and every edge of a proposal can
only narrow it. So any proposal provision accepts is within authority by
construction: the runtime proves that, before anything runs. A proposal that
reaches too far is an ordinary horizon or contract fault at provision, cheap
and recorded, with no model started.

What is left for the human is the question a human is good at: *is this the
work I want?* That is review, not approval of each act, and it is the only
decision in the loop.

## Exits

A node in the tree has five ways to end: return, or one of four typed exits.
They are the agent experience channels: what a node does instead of
improvising when the proposal does not decide something.

| Exit | Means | Goes to |
|---|---|---|
| **ask** | the goal is ambiguous | its parent; past the root, think and the human |
| **missing observation** | it needs something its Horizon cannot read | its parent; past the root, think, which can read |
| **give up** | it cannot do this within its Horizon, and says why | its parent, with the node's strand as the record of what it tried |
| **authority** | it needs more than its position holds | its parent, which may grant from its own Horizon; otherwise it climbs |

Authority depends on position. A parent can only grant from what it holds,
so a request climbs until it reaches a holder. Provision has no model and
decides nothing, so a request that climbs past the tree's root becomes a new
proposal: think revises, the human sets it in motion, provision builds again.
Authority above the tree changes only through think and the human.

Where these exits live (interpreter syscalls, typed returns, or a parked
frame on a host) is open. None of them needs the kernel: each is an invoke,
a return or a park, so each is expressible as kernel guarantees plus commit,
the minimality test in PR #43.

## What this changes

- **think** no longer writes a spec and contract pair for the human to read
  and act to derive a Horizon from. It returns the proposal, and the Horizons
  in it are the ones that will be enforced: there is no derivation step
  between what is reviewed and what runs. [think](think.md).
- **act** is no longer write-only, one LLM invocation per function, under a
  limit of k continuations per node. It is provision plus the tree it starts;
  the limit moves from continuations to leaves. [act](act.md).
- **The interrupt transition in `core/machine`** that act.md required is not
  needed. Exits are interpreter-level.

## Open

1. **Provision's contract.** The schema above is a first guess. Does a node
   name a contract it must return, or only a goal ([DELEGATION](../design/DELEGATION.md)
   says the caller supplies no completion test today)?
2. **Can a parent re-delegate at runtime?** A Horizon never widens, so
   granting a child more means ending it and starting a new child within the
   parent's Horizon, citing the old strand. Is that allowed inside a declared
   tree, or only through a new proposal?
3. **Standing grants.** A proposal that fits a standing grant could run with
   no one looking. Probably never when think read untrusted content.
4. **Nesting in the kernel.** `system.narrow` today returns a reference into
   the caller's own reach, so references are flat. Provision needs to narrow
   into a reference it is building, with each child checked against its
   parent.

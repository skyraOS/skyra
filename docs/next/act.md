# Act: provision, then the tree

Date: 2026-09-28, rewritten 2026-10-24  
Status: proposed, not built. Part of [an agent](agent.md):
[think](think.md) proposes, provision builds, act runs. Builds on
[delegation](../design/DELEGATION.md), [24 — the room](24-the-room.md) and
[21 — observation passing](21-observation-passing.md).

## What it is

Act takes one proposal and makes it run. It is two parts, and only the
second involves a model:

```text
user ──proposal──▶ provision          a procedure: check every edge, narrow, delegate
                      │
                      └─▶ root        the tree starts; each node is a delegation
                           ├─ node A  its own Horizon, its own goal, its own model
                           │   └─ …
                           └─ node B
```

The system provisions; the model works inside what was provisioned. No model
chooses its own Horizon.

## Provision

Provision is an interpreter procedure with no model. Its input contract is
the shape of a proposal ([an agent](agent.md#a-proposal-is-an-observation));
its Horizon is the human's grant. It:

1. **Checks every edge.** The root must fit provision's Horizon, and every
   child must fit its parent's. A node that names anything its parent does
   not hold is a horizon fault, of the whole proposal, before anything runs.
2. **Builds the tree.** For each node, a narrowing to its Horizon and a
   delegation binding that Horizon to its actor and goal. Nothing in the tree
   runs yet.
3. **Starts the root.**

Provision is a pure function of the proposal and its own Horizon, so replay
derives the same tree again. It decides nothing the proposal did not say.

Today `system.narrow` returns a reference into the caller's own reach, so
references are flat. Provision needs to narrow into a reference it is
building, with each child checked against its parent: the same check
`machine.Narrow` makes at runtime, made at provision.

## The tree

Each node is a delegation, as [DELEGATION](../design/DELEGATION.md) defines
it: a fresh child whose only task context is its goal and the observations
the proposal cites for it. A node's Horizon is the one provision built for its
position. Its model may do anything that Horizon allows, in any order: the
proposal fixes the Horizons, not the control flow.

A node may narrow further inside its own Horizon, which needs no approval
because narrowing only drops. It may not start a child the proposal did not
name ([open](#open-design-choices)).

## The leaf budget

The old act limited each node to k continuation invocations, to force
composition. That makes a long, flat pipeline impossible: one root over n
continuations is a good shape, and a limit on continuations forbids it.

The limit moves to **leaves**: the invocations that touch the world.
Continuations are structure, and structure is what this should encourage;
syscalls touch only the Horizon. A node that can touch the world only k times
has to be about one small thing, so composition stays tight without making
any pipeline impossible.

- **Counted per frame, across every round.** A continuation resumes once per
  round; a budget per round would let a node loop forever, one leaf a round.
  Gather already walks a frame's rounds, so the count is cheap.
- **Set in the proposal, capped by the parent.** A child's budget cannot
  exceed what its parent may still spend.
- **Possibly by kind.** Reversible leaves (a read, a write to scratch) are
  cheaper than irreversible ones ([interpreter primitives](../../ideas/interpreter-primitives.md#undo-reversibility-as-a-derived-fact)).
  Whether the budget counts them separately is open.

## Exits

A node never improvises past its goal. When the goal does not decide
something, it exits to its parent, typed:

| Exit | Means | Resolved by |
|---|---|---|
| **ask** | the goal is ambiguous | the parent; past the root, think and the user |
| **missing observation** | it needs something its Horizon cannot read | the parent; past the root, think, which reads it and cites it in vN+1 |
| **give up** | it cannot do this within its Horizon | the parent, with the node's strand as the record of what was tried; the node adds only its diagnosis |
| **authority** | it needs more than its position holds | the parent, from its own Horizon; otherwise it climbs |

**Authority depends on position.** Nothing widens a Horizon, so a parent
grants more by ending the child and starting a new one under a wider
narrowing of its own Horizon, citing the old child's strand so the work goes
on. A parent that does not hold what is asked exits upward itself. Past the
root, provision decides nothing, so the request becomes a new proposal
through [think](think.md#exits-come-back-here).

The old act raised interrupts through a new transition in `core/machine`.
Exits need none: each is a return, a fault or a parked frame, so each is
expressible as kernel guarantees plus commit.

## What act returns

Each node returns the observations its leaves authored, as
[21](21-observation-passing.md) requires; parents pass them up, selected and
ordered but not rewritten. The root returns them to the user, each naming
the leaf that wrote it and the proposal version it served.

## Cases the design must decide

1. **A node invokes something outside its Horizon:** a horizon fault, which
   the node may follow with an authority exit.
2. **A node spends more leaves than its budget:** refused at intake, before
   the leaf runs.
3. **A proposal edge where the child holds more than its parent:** refused at
   provision, before the root starts.
4. **A goal that tells a node to reach further:** the Horizon refuses it. A
   goal is never authority.
5. **A node returns code it wrote itself:** refused under 21. Only a leaf's
   observation goes up.

## Open design choices

- **Re-delegation inside a declared tree.** Is a parent starting a wider
  replacement child allowed, or must every change to the tree's shape be a
  new proposal?
- **Reads in the tree.** The old act was write-only, so a node saw only what
  the user approved. A node may now hold reads if the proposal gives them,
  which makes verification possible (a build, a test) and lets untrusted
  content into a node with effects. Is the default no reads but check leaves?
- **A node's return contract.** Delegation today returns observations, with
  no completion test from the caller. Does a proposal node name what it must
  return?
- **Where exits live:** interpreter syscalls, typed returns, or a park.

## What this needs in the controller

- Provision as an interpreter procedure, with its input contract.
- Narrowing into a reference under construction, checked edge by edge.
- The leaf budget, counted at intake across a frame's rounds.
- Typed exits in the interpreter.
- Observation IDs, the per-frame register and the return check
  ([21](21-observation-passing.md)).

## Done when

- A proposal whose edges all narrow builds a tree that matches it, and one
  that widens anywhere is refused before any model runs.
- No node exceeds its leaf budget, counted across its rounds.
- Each exit reaches the right holder, and one past the root ends in a new
  proposal.
- Replay re-derives provision, the tree, every budget check and every exit.

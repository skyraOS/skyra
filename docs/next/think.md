# Think: the conversation that proposes

Date: 2026-09-28, rewritten 2026-10-24  
Status: proposed, not built. Part of [an agent](agent.md): think proposes,
provision builds, [act](act.md) runs. Builds on
[21 — observation passing](21-observation-passing.md). The think procedure
that exists today ([38](proposal/38-think-procedure-proposal.md),
[39](proposal/39-unified-think-procedure-proposal.md)) is the conversation
part; the proposal is not built.

## What it is

Think is the conversation between the user and a continuation that can read
and cannot act. It talks with the user, reads, and returns one kind of
result: a **proposal**. Think changes nothing, and it creates nothing that
can act.

```text
user ⇄ think          conversation: ordinary back and forth
think → proposal      an observation: the arguments to provision's contract
user → provision      the user sets the proposal in motion: the one decision
```

## What think holds

| Holds | Does not hold |
|---|---|
| a read of its parent's Horizon: the actors and contracts it may propose against | `narrow`, `delegate`, or any way to create a Horizon |
| any other read it is granted, named explicitly | write, or any leaf with an effect outside its scratch |
| a return: the proposal | `extend-reach` on its own Horizon |

Think's authority is to read and to answer. It never passes on authority,
because it holds none to pass. Two reads are kept apart on purpose:

- **Its parent's Horizon** is kernel state: the catalog of what a proposal
  can name. It is trusted.
- **Everything else it reads** (files, pages, tool output) is where untrusted
  content enters. Each such read is granted explicitly, so the human can see
  whether a proposal was made from untrusted content.

## The proposal

A proposal is an observation: a set of arguments meant to fill provision's
contract. It is a tree of delegations that do not run yet, one node per
delegation, each with its position, actor, Horizon, leaf budget and goal.
[An agent](agent.md#a-proposal-is-an-observation) has the shape.

Think writes the **goal** for every node: the prompt that node's model will
start from. The goal is advisory. The Horizon is binding. Think may write
anything as a goal; nothing in it is read as authority.

A proposal cites the observations it was made from by ID, as the old spec
did: the contract headers, the files, the listings. A node's model receives
its goal and the observations cited for it, not think's conversation.

## The conversation stays

Think is still a conversation, and that is not something to design away.
When it gets long, it compacts, when

- **the user asks for it**, or
- **the conversation passes a token threshold** (configurable).

Compaction does not summarize. It writes the current proposal from the
conversation. After compaction the conversation continues from the proposal,
the observations it cites, and the turns since. Nothing from before survives
unless the proposal cited it.

Under [21](21-observation-passing.md), think cannot return an observation it
authored itself, so compaction is a **leaf**: invoked with the IDs of the
turns and reads, it authors proposal vN and names its inputs. Think passes it
up unchanged.

## The user's move

Proposal vN comes back to the user. The user may:

- **set it in motion.** The user invokes provision with it. This is the
  approval, and it is about intent: provision proves the authority before
  anything runs ([an agent](agent.md#approval-is-about-intent-not-authority)).
- **pass it back to think** with improvements of their own. The user is the
  root, so they may author on the way down; the next compaction is vN+1.
- **keep talking.** The proposal stays in the lineage either way.

## Exits come back here

When a node in the tree meets something its goal does not decide, it exits
to its parent ([act: exits](act.md#exits)). An exit that climbs past the
tree's root reaches think and the user, because provision has no model and
decides nothing. Answering it is a think turn, and it ends in proposal vN+1.

## Cases the design must decide

1. **Think tries to write:** a horizon fault. Nothing that writes is in reach.
2. **Think tries to narrow or delegate:** a horizon fault. Neither is in reach.
3. **Think returns its own text as the proposal:** refused under 21. The
   proposal is authored by the compaction leaf.
4. **The proposal names an actor not in think's parent's Horizon:** a
   contract or horizon fault at provision, before anything runs.
5. **The proposal cites an observation think never received:** refused before
   compaction commits.

## Open design choices

- **Who invokes compaction at the threshold.** The runtime, or think,
  instructed by its adapter? A threshold crossed mid-turn needs a defined
  point where compaction runs.
- **What compaction reads.** The whole conversation, or the previous proposal
  plus the turns since it?
- **Prose beside the tree.** Does a proposal carry a human-readable summary
  for the user's review, and if so, is it advisory like a goal?

## Done when

- Think runs under a Horizon with read of its parent's Horizon and the reads
  it was granted, and nothing else; a write, narrow or delegate is a horizon
  fault.
- Compaction is a leaf that authors a proposal fitting provision's contract
  and citing observations by ID.
- The user can set a proposal in motion, or pass it back with their own
  observations.
- Replay re-derives every proposal version, its citations and its author.

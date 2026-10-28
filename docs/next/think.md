# Think: the conversation that proposes

Proposal: **In the proposal.** Its mechanism, a read-only Horizon made through `system.narrow`, was built by [38](proposal/38-think-procedure-proposal.md); think's own per-document proposal was removed on 2026-10-28 (in git history).

Date: 2026-09-28  
Status: proposed, not built. Pairs with [act](act.md). Builds on
[21 — observation passing](21-observation-passing.md) and the interrupt
transition drafted in the agenda.

## What it is

Think is the conversation between the user and a read-only continuation. It
sees the whole picture: it talks with the user, reads the workspace, and
proposes. It produces one thing, a **spec** and the **contract proposal** that
spec pins. [Act](act.md) implements it. Think changes nothing.

```text
user ⇄ think          conversation: ordinary back and forth, read-only
think → compaction   spec vN + contract vN, citing what think read
user → act           the user passes spec vN down: the one decision
act ⇢ user           an interrupt comes back; answering it is a think turn
```

## One truth of state

The contract proposal is the state. Think and act both reason over it, and
nothing else carries context between them: no transcript handed to act, no
memory kept beside the lineage. Every piece is an observation in the lineage,
content-addressed, with its author ([21](21-observation-passing.md)).

| | Is | Author |
|---|---|---|
| **Turn** | one message in the conversation | the user, or think |
| **Read** | a file, a contract header, a listing | the read leaf that observed it |
| **Spec vN** | what compaction produces from the conversation | the compaction leaf |
| **Contract vN** | the interface or struct proposal, one to one with spec vN | the compaction leaf |
| **Improvement** | an observation the user adds when passing a spec back | the user |

A spec version and its contract version are one pair. Changing either makes a
new version of both.

## Read-only

Think's Horizon reaches read leaves and nothing that writes. What it reads
comes back as an observation authored by the leaf and enters think's register.
The spec **cites** those observations by ID: the contract headers a function
must match, the file it replaces, the listing it was planned against.

So a read happens once, in think, and is already in the frame when act runs.
Act never reads. What act knows about the world is exactly what the spec
cites, and the user saw the spec before passing it on.

## The conversation stays

Think is still a conversation, with ordinary back and forth, and that is not
something to design away. What changes is what happens when it gets long:
compaction runs when

- **the user asks for it**, or
- **the conversation passes a token threshold** (configurable).

## Compaction is spec generation

Compaction does not summarize. It writes the spec and the contract proposal
from the conversation.

Under [21](21-observation-passing.md), think cannot return an observation it
authored itself, so compaction is a **leaf**. It is invoked with the IDs of
the conversation's turns and reads, it authors spec vN and contract vN, and
it names the inputs it used. Think passes the pair up to the user unchanged.
This settles 21's first open question, how a model mints an observation
without declaring its own hash: the turns are the authoring, and compaction
is the mint.

After compaction the conversation continues from the spec. Think's context is
spec vN, the observations it cites, and the turns since. Nothing from before
the compaction survives unless the spec cited it.

## The contract proposal

The contract proposal is an interface or struct: the functions that should
exist, their input and output types (the JSON-types contracts of
[`core/contract`](../../controller/core/contract/)), and the structure that
composes them.

The structure is a **tree**, and think proposes it with the user as its root.
Act limits how many continuations each node may invoke
([act: the tool limit](act.md#the-tool-limit-forces-composition)), so a
contract with more functions than one node can hold must group them. Think
proposes that grouping, and act grows it. Composition is part of the
proposal, not something act improvises.

## The user's move

Spec vN comes back to the user as an observation. The user may:

- **pass it to act.** This is the approval: the one human decision. Act's
  Horizon is derived from the contract it receives.
- **pass it back to think** as a parameter, with improvement observations of
  their own. The user is the root, so they may author on the way down. Think
  continues, and the next compaction produces spec vN+1.
- **keep talking.** The spec stays in the lineage either way.

## Interrupts come back here

When act meets something the spec does not decide, it raises an interrupt
([act: interrupts](act.md#interrupts)). The interrupt reaches the user, and
answering it is a think turn. The answer changes the spec and the contract,
a new version is compacted, and act runs again against it.

## Cases the design must decide

1. **Think tries to write:** a horizon fault. Nothing that writes is in reach.
2. **Think returns its own text as the spec:** refused under 21. The spec is
   authored by the compaction leaf.
3. **The spec cites an observation think never received:** refused before
   compaction commits; the ID is not in think's register.
4. **The user passes spec vN back with an improvement:** accepted. The
   improvement is authored by the user and cited by vN+1.
5. **A spec with no contract, or a contract with no spec:** cannot be formed.
   They are one pair.

## Open design choices

- **Who invokes compaction at the threshold.** The runtime, as an interrupt
  to think? Or think, instructed by its adapter? A threshold crossed in the
  middle of a turn needs a defined point where compaction runs.
- **The spec's shape.** Prose plus the contract, or the contract with
  annotations? The contract format needs function signatures and the tree,
  which `core/contract` does not express today.
- **What compaction reads.** The whole conversation, or the previous spec plus
  the turns since it?
- **The token threshold's default,** and whether it is per conversation.

## Done when

- Think runs under a Horizon with no write in reach, and a write attempt is a
  horizon fault.
- Compaction is a leaf, triggered by the user or the threshold, and authors a
  spec and contract pair that cites observations by ID.
- The user can pass a spec to act, or back to think with their own
  observations.
- Replay re-derives every spec version, its citations and its author.

# Act: the tree that implements

Proposal: **In the proposal.** Narrowing per node and authority as a climbing fault are proposed in [proposal/act-proposal](proposal/act-proposal.md).

Date: 2026-09-28  
Status: proposed, not built. Pairs with [think](think.md). Builds on
[21 — observation passing](21-observation-passing.md) and the interrupt
transition drafted in the agenda.

## What it is

Act implements one approved spec version. The user invokes it with spec vN,
contract vN, and the observations the spec cites. Act is **write-only**: it
sees what the user passed it and what its own children return, and nothing
else from the world.

```text
user
 └─ act: contract vN                      the root of the implementation
     ├─ act: module A                     groups functions, invokes ≤ k continuations
     │   ├─ act: A.parse    → write leaf  one LLM invocation per function
     │   └─ act: A.render   → write leaf
     └─ act: B.handle       → write leaf
```

## Write-only

Act's Horizon reaches write leaves and the continuations the contract
proposes, and no read leaf. Everything act needs was read in think and cited
in the spec, so it is already in the frame. If act needs something the spec
does not cite, it cannot go and look: it raises an interrupt.

That split is a security property. The node that reads untrusted content
(think) cannot write, and the node that writes (act) sees only what the user
approved. A prompt injection in a file think read reaches act only if the
spec cites it, and the user saw the spec.

## One LLM invocation per function

Each function in the contract is implemented by **one** continuation
invocation. It receives that function's contract and the observations cited
for it, and nothing about its siblings.

Under [21](21-observation-passing.md) a continuation cannot return code it
wrote itself. It authors the implementation on the way down, as the operand of
a write leaf; the write leaf performs the write and authors the observation of
what was written. The continuation passes that observation up unchanged.

## The tool limit forces composition

Each continuation may invoke at most **k** continuations. `reply` does not
count, and neither do leaves: only continuations do.

The limit is the point. A contract with more functions than k cannot be
implemented by one node calling everything. It has to be grouped into
modules, each a node with at most k children, so the work takes the shape of
a tree. Composition comes from grown structure, not from one agent holding a
long list of tools. Think proposes that tree in the contract; act grows it,
node by node, each under a narrower Horizon.

This is also what makes self-extension happen: a node that needs more than k
continuations must create structure, and every node it creates runs under a
Horizon narrowed from its own, with no approval.

## The Horizon is derived from the contract

Act's Horizon comes from the contract it was passed:

- **writes:** only the places the contract names.
- **reach:** only the continuations the contract's tree proposes, at most k
  per node.
- **narrowing:** each child's Horizon covers its own subtree, checked by
  `machine.Narrow` like any other edge.

Passing the spec is how the user approves the boundary: they approve a plan
they can read, and the Horizon follows from it.

## Interrupts

Act never improvises past the spec. When the spec does not decide something,
act raises an interrupt, and the interrupt's type decides what happens next.

| Interrupt | Means | Resolved by |
|---|---|---|
| **clarification** | the spec is ambiguous | the user answers in think; spec vN+1; act runs again |
| **missing observation** | act needs something the spec does not cite | think reads it; spec vN+1 cites it |
| **contract misfit** | the function cannot be implemented to its interface | think revises the contract; vN+1 |
| **authority** | act needs to write outside what the contract grants | the user decides; vN+1 grants it, or not |

Every interrupt ends in a new spec and contract version, never in context
passed to act on the side. The system has one truth of state, and an
interrupt changes it through think.

The interrupt transition is drafted in the agenda but not built:
`core/machine` has no interrupt today.

## What act returns

Each function's node returns the observation its write leaf authored. Module
nodes pass their children's observations up, selected and ordered but not
rewritten. The root returns the set to the user, and every one names the
leaf that wrote it and the spec version it implemented.

## Cases the design must decide

1. **Act invokes a read leaf:** a horizon fault. No read is in reach.
2. **A node invokes more than k continuations:** refused at intake, before
   the child runs.
3. **A function node returns code it wrote:** refused under 21. Only the write
   leaf's observation can go up.
4. **Act writes outside the contract's paths:** denied by the rendered
   sandbox; asking for it is an authority interrupt.
5. **A node needs an uncited file:** a missing-observation interrupt, not a
   read.

## Open design choices

- **Verification.** Can act invoke a check leaf (build, type check, tests)
  and receive its result? The result is a child's observation, which 21
  allows, but the check reads the world, which write-only forbids. One answer:
  checks are leaves act may invoke, whose output is only pass or fail against
  the contract.
- **Resume or re-run.** The agenda's interrupt keeps the original slot open
  and resumes. Here the answer changes the spec, which suggests re-running
  act. Which one, and at what granularity?
- **Re-running only what changed.** A function whose contract and cited
  observations did not change between versions has the same inputs, so its
  implementation could be reused, keyed by that hash (grips I, compiled
  cognition).
- **Where k is set:** globally, per conversation, or per node in the
  contract.
- **Files a function modifies.** Assumed: think cites the current contents,
  and act never reads them itself.

## What this needs in the controller

- Observation IDs, the per-frame register and the return check
  ([21](21-observation-passing.md)).
- A write leaf, and the compaction leaf ([think](think.md)).
- A contract format with function signatures and the tree.
- The tool limit, counted at intake.
- The interrupt transition in `core/machine`, with typed interrupts.
- A Horizon derived from a contract.

## Done when

- Act runs under a Horizon derived from the contract, with no read in reach.
- Each function is one continuation invocation, and no node invokes more than
  k continuations.
- Every returned observation was authored by a write leaf.
- Each interrupt type reaches the user, and answering it produces a new spec
  version that act runs against.
- Replay re-derives the tree, every limit check and every interrupt.

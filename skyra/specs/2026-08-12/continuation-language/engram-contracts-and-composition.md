# Engram contracts and composition

**Date:** 2026-08-13
**Status: PROPOSAL.** This defines the fundamental semantic shape of an engram, the ownership of
contracts between parent and child, typing by known provenance, and the boundary between possible,
live, and remembered composition. It does not propose concrete types or an implementation.

> **An engram is a compiled continuation plus the contracts required for its children to
> participate in that continuation. The parent owns those contracts because the parent is the
> behavior whose future depends on their returns.**

---

## The fundamental shape

```text
ENGRAM

1  CHILD CONTRACTS
   for each child position in the continuation:
   ├── expectation contract     what behavior is being requested
   ├── input contract           which observations may be given to it
   └── return contract          which observations the continuation can consume

2  COMPILED CONTINUATION
   ├── which child positions may be invoked
   ├── invocation order
   ├── dependencies among invocations
   ├── transitions determined by returned observations
   └── terminal return behavior
```

In one line:

```text
ENGRAM = CHILD CONTRACTS + COMPILED CONTINUATION
```

The continuation is compiled in the broad sense: it is behavior whose possible next steps are
already structured. It may branch on observations returned by children, invoke independent
children concurrently, repeat a position, stop, or produce a terminal observation.

It does not contain the future observations themselves. Those arrive only through live execution.

---

## The parent owns the child contract

If Parent invokes Child, Parent holds the contract for that child position:

```text
Parent
│
├── expects Child(input contract) → Observation<X>
│
└── compiled continuation
       │
       ├── invoke the child position
       ├── receive an observation
       ├── test whether it satisfies Observation<X>
       └── continue only through a transition valid for that observation
```

This ownership follows dependency:

> **The return contract belongs to the parent because the return determines the parent's future
> behavior.**

The child does not declare what the parent must accept. It runs under its own continuation and
returns an observation. Compatibility is judged at the boundary where that observation attempts to
enter the parent's continuation.

Recursively:

```text
Parent owns contracts for Parent's child positions.
Child owns contracts for Child's child positions.
Grandchild owns contracts for Grandchild's child positions.
```

No engram owns the contracts by which every possible caller in the system must use it. The same
engram may participate under different parent-owned contracts in different compositions.

---

## A child contract describes a position, not a permanent edge

A contract says what may occupy one position in the parent's continuation. It does not create a
durable relationship to a particular child.

```text
CHILD CONTRACT
    behavior of this kind
    receiving observations of these known kinds
    returning an observation consumable here
```

Before composition, the position is potential. It may be filled by an existing engram, a
thought-derived candidate expectation, or another compatible behavior the parent can Decide upon.

This means the structure does not need a persistent type edge:

```text
Parent ── type ──▶ Child
```

Compatibility is evaluated when the composition is proposed and selected. A stored edge would say
only that two things might compose, would become stale when either side changed, and would be easy
to mistake for evidence that they had composed successfully.

The durable graph edge has a stronger meaning and should remain reserved for it:

> **This composition happened, its observations participated in the retained causal cone, and it
> survived Remember.**

---

## Typing is known observation provenance

Typing in this system is not a value declaring a tag about itself.

```text
NOT SUFFICIENT
    {"type": "reach"}

LOAD-BEARING
    this observation is known by the substrate to have been produced by the behavior whose
    observations this contract accepts
```

An observation carries content. The substrate carries its provenance. The parent carries the
contract which determines whether that known observation may be consumed at a particular position.

These are separate facts:

```text
PROVENANCE
    Which engram and invocation produced this observation?

CONTRACT SATISFACTION
    May this parent continuation consume that observation at this child position?
```

In notation:

```text
Y ⊨ X

Y     the returned observation with substrate-known provenance
X     the parent-owned return contract at this position
⊨     Y is admissible to the continuation governed by X
```

The producer is not self-reported and cannot be forged by placing an ID or type name in content.
The contract is not a universal classification of the producer. It is a contextual statement by
the consumer:

> **I know observations originating through this behavior, and I have a continuation for them
> here.**

This preserves both halves of typing:

```text
NOMINAL GROUNDING      the observation came from known behavior.
CONTEXTUAL USE         this parent accepts it in this position.
```

---

## Input contracts

The parent also owns what may be handed into the child position.

```text
INPUT CONTRACT
    which observations from the parent's live causal state may be bound to this child
```

This does two jobs:

```text
COMPATIBILITY     the proposed child composition must be meaningful for these inputs.
CAUSALITY         the substrate records the exact observations actually supplied.
```

The input contract does not grant the child everything in the parent's frame. It identifies what
may cross this boundary. The actual invocation may receive a permitted subset selected by the
compiled continuation.

Horizon remains separate:

```text
INPUT CONTRACT    what observational material may be supplied.
HORIZON           how far outward the child may see or act.
```

Treating authority as an input observation would make it caller-authored semantic data. Treating
input observations as Horizon would create a second frame. The two bounds travel together but do
not mean the same thing.

---

## Return contracts

The return contract is part of the parent's compiled control flow.

```text
Child position → Observation<X>
                          │
                          ▼
                 continuation for X
```

The contract is necessary because a returned observation is not merely output to display. It is an
input to the parent's next transition.

```text
RETURN SATISFIES X
    the parent has a valid continuation for this observation.

RETURN DOES NOT SATISFY X
    the parent has no valid continuation through this child position for this observation.
```

The return contract therefore does not primarily document the child. It defines the domain over
which the parent's future behavior exists.

---

## Working note: the return contract is the child's parser

Because the parent owns the domain of its own future behavior, it also owns the parser for each
child return. That parser has one local result:

```text
PARENT PARSES CHILD RETURN

1    this observation satisfies the contract; the compiled continuation may proceed.
0    this observation does not satisfy the contract; this parent has no next state for it.
```

The child cannot own this parser. The child knows what it produced, but only the parent knows which
observations it can continue from. The parser is therefore the executable boundary of the
parent-owned return contract.

This repeats recursively:

```text
leaf observation
    ↓ parsed by its parent
parent observation
    ↓ parsed by its parent
outer observation
    ↓ parsed by its parent
```

No single parser must understand the entire causal tail. Each parent owns one boundary. The nested
chain of parent-owned parsers establishes the type of the whole composition. In that precise sense,
the recursion is the parser.

---

## Think produces contract candidates

[`think-interface-proposal.md`](think-interface-proposal.md) defines Think as the interface to
potential:

```text
Think(entity[], expectation[], observation[])
    within From and Horizon
    → bounded ranked expectation-observations
```

Those candidate observations may be considered for a child position only when they are compatible
with the parent-owned contracts there.

```text
THINK
    proposes potential behavior.

CHILD CONTRACT
    determines whether that potential can occupy this position.

DECIDE
    selects one compatible composition.
```

Think does not create a type edge and does not edit the parent. It returns an observation carrying
potential behavior. If the parent uses that candidate, it Decides on the observation itself so the
thought which produced it remains a constituent of the execution.

---

## Decide composes executable and destination

The proposed decision surface is:

```text
Decide(engram, reachObservation)
```

It composes:

```text
ENGRAM              executable behavior proposed for the current child position.
REACH OBSERVATION   destination potential produced by known reach-thinking behavior.
```

The reach observation already identifies the entity. A second entity operand would create two
places for one fact and permit contradiction:

```text
Decide(child, entity A, reachObservation naming entity B)
```

Removing it makes that disagreement inexpressible.

The current position in the parent's compiled continuation supplies the child contract. The
decision does not need to repeat the input contract, return contract, or continuation order. Those
belong to the parent which is making the decision.

If several child positions are simultaneously available, each decision still occurs at one known
position in the compiled continuation. Position is causal context, not another semantic operand
the entity must declare.

---

## Reach realizes the decision

The reach observation is typed potential, not an address and not a connection.

```text
think-reach
    → reach observation naming destination semantics and entity

Decide(engram, reach observation)
    → one completely chosen composition

hidden Reach
    → possibility | bounded impossibility observation
```

Reach resolves volatile mechanism behind the kernel boundary. It may establish the exact
composition Decide chose. It may not replace the engram, entity, inputs, child position, or Horizon.
Any substitution would create a different possibility.

This produces the composition lifecycle:

```text
CONTRACTS                    can compose here.
THINK                        proposes executable or destination potential.
DECIDE                       chooses one complete composition.
REACH                        establishes that composition as a live possibility.
RETURN SATISFIES CONTRACT    the parent's continuation can proceed.
REMEMBER                     the retained composition becomes durable causal history.
```

Strictly:

```text
Decide requests composition.
Reach success means it did compose.
Remember means the composition survives.
```

---

## Failure requires no universal Error object

Failure is inability to cross a required boundary. It does not fundamentally require a semantic
object called `Error`.

### 1 · Incompatible potential

```text
The proposed engram or reach observation does not satisfy the current child contract.
```

Decide cannot form that composition. No child possibility exists.

### 2 · Unrealizable decision

```text
The composition is valid and completely chosen, but Reach cannot establish it now or the narrowed
Horizon does not permit it.
```

The result is `beyond current reach` or `beyond your horizon`. No target possibility opened and
failed; the selected potential never became live.

### 3 · Unsatisfied return contract

The child did run and returned `Observation<Y>`, while the parent position requires
`Observation<X>`:

```text
Y ⊭ X
```

The parent's normal continuation cannot proceed through that position. This is the substrate-level
failure.

An explanatory observation may accompany any of these:

```text
{"explanation": "credential expired"}
{"explanation": "returned a collection where one item was required"}
```

But the explanation is evidence about the failure. It is not what makes the composition fail.

The substrate must preserve the structural fact at the appropriate boundary:

```text
what composition was proposed
which contract rejected it, if any
whether Reach established it
what observation actually returned
which return contract it did not satisfy
where the continuation became unable to proceed
```

Without that record, failure would work operationally but teach nothing.

---

## Working note: a failed parser creates a narrower problem

A parser result of `0` wakes the parent that owns the failed child contract. The mismatch is already
localized:

```text
OWNING PARENT
FAILED CHILD POSITION
EXPECTED RETURN CONTRACT
ACTUAL RETURNED OBSERVATION
```

The parent may then be given a frame bounded to that specific dependency. Within it, the parent can
inspect, Think, Decide, and Reach for behavior which rectifies the mismatch. It does not receive a
general-purpose error frame or wider access to the graph.

If the corrective behavior produces a deeper mismatch, its owning parent receives a still more
focused frame. Resolution therefore zooms down the causal dependency chain until a satisfying
observation is produced or the narrowed Horizon admits no resolution:

```text
failed child A
    ↓ bounded resolution frame A
failed child B
    ↓ bounded resolution frame B
failed child C
    ↓ resolved
resume suspended continuations upward
```

This gives operational force to the ownership rule:

> **Define a child contract, own its parser, and own the behavior invoked when that parser returns
> `0`.**

The mismatch itself remains the failure. The bounded frame is the opportunity to compose a repair,
not an `Error` object standing in for the failure.

---

## Leaves remain engrams

If an engram were only child contracts plus child invocation order, a leaf would be empty. The
compiled continuation therefore includes terminal return behavior.

```text
LEAF
    CHILD CONTRACTS          none
    COMPILED CONTINUATION    produce a terminal observation

ORCHESTRATOR
    CHILD CONTRACTS          one or more
    COMPILED CONTINUATION    invoke, consume, branch, and eventually return
```

Leaf and orchestrator remain structural positions, not declared types. The same engram may behave
as a leaf along one path and invoke children along another if its compiled continuation permits
both.

A leaf grounds the recursive parser chain. Once an activated and successfully reached flow arrives
at a leaf path, its terminal behavior fires: there is no deeper child return on which it can wait.
This does not mean every leaf runs continuously or globally; it means a reached leaf cannot defer
to another child before originating its observation.

---

## The graph records history, not compatibility

The three layers must remain separate:

```text
CONTRACTS
    this composition is admissible as potential.

LIVE CAUSAL STRUCTURE
    Decide selected it, Reach established it, and observations crossed its boundaries.

DURABLE GRAPH
    Remember retained this composition as part of an accepted causal cone.
```

Conflating them produces three different errors:

```text
CONTRACT AS GRAPH EDGE
    potential looks like experience.

LIVE ATTEMPT AS DURABLE EDGE
    exploration looks accepted before anyone remembers it.

DURABLE EDGE AS TYPE
    what worked once becomes the only thing considered compatible later.
```

The clean invariant is:

> **Contracts say what can compose. Decide and Reach say what did compose. Remember says which
> composition survives.**

---

## Boundaries

The parent may:

```text
define the child positions its continuation needs
define the observations each position may receive
define the observations for which it has a continuation
Think for compatible executable and destination potential
Decide on one compatible engram and reach observation
continue only from returns its contract accepts
```

The parent may not:

```text
rewrite the child's internal continuation
make the child self-certify compatibility
forge the producer of a returned observation
treat a compatibility contract as evidence that composition occurred
treat a live attempt as durable before Remember
```

The child may:

```text
execute its compiled continuation
invoke positions governed by its own child contracts
return an observation with substrate-known provenance
```

The child may not:

```text
declare which observations its parent must accept
rewrite the parent's continuation
claim that an incompatible return satisfies the parent's contract
decide whether the parent or outer caller remembers it
```

---

## What remains open

```text
the vocabulary by which a parent contract identifies accepted observation provenance
whether contracts admit one producer, a bounded set, or a composable relation among producers
how structural content requirements combine with provenance requirements
how a thought-derived provisional expectation is compared with an expectation contract
whether one child position may accept several return contracts with different continuations
how contract revision behaves while a possibility remains live
how explanatory observations attach to structural failures without becoming mandatory Error values
what exact bounded frame is supplied when a parent-owned child parser returns 0
how suspended continuations resume after a deeper resolution returns a satisfying observation
```

None changes the proposed ownership and lifecycle:

```text
PARENT OWNS THE CHILD CONTRACT.
OBSERVATION PROVENANCE IS SUBSTRATE-KNOWN.
COMPATIBILITY IS CHECKED AT COMPOSITION AND RETURN BOUNDARIES.
THE GRAPH RECORDS ONLY COMPOSITION WHICH SURVIVED REMEMBER.
```

---

## Acceptance test

The proposal holds only if:

```text
1  Every child input and return contract is owned by the parent whose continuation uses it.
2  A child cannot declare or forge compatibility with its parent.
3  Typing depends on substrate-known observation provenance, not a content tag.
4  Contract satisfaction remains contextual to one parent position.
5  No durable edge is required merely to represent compatibility.
6  Decide composes an executable engram with a typed reach observation.
7  Entity is not duplicated outside the reach observation in Decide's semantic signature.
8  Reach realizes exactly what Decide selected or returns bounded impossibility.
9  A mismatched return blocks the continuation without requiring a universal Error object.
10 Structural failure remains observable even when no explanatory observation exists.
11 A leaf remains expressible through terminal behavior in its compiled continuation.
12 Potential, live composition, and remembered history remain three distinct states.
13 Executing a thought-derived expectation preserves the candidate observation as a cause.
14 Every child return is parsed by the parent which owns that child position's return contract.
15 A parser result of 0 wakes that parent without widening its frame or Horizon.
```

If any one fails, contracts, execution, or memory have begun to stand in for one another, and the
system can no longer say whether something merely fit, actually happened, or was retained.

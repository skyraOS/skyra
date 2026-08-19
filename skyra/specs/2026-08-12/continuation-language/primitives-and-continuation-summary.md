# Primitives and continuation language

**Date:** 2026-08-13
**Status: WORKING SUMMARY.** This document relates the continuation language directly to the
semantic primitives. It records the current directional execution model and the consequences of a
child return failing its parent-owned type.

---

## Core claim

The system is a recursive flow with two directions:

> **The continuation language composes behavior on the way down. Typed observations realize that
> behavior on the way up.**

```text
                         DOWN

              Think → Decide → Reach
                         ↓
              composed child behavior
                         ↓
              narrower dependencies
                         ↓
                        LEAF
                         ↓
                          UP

                  child observation
                         ↓
                parent-owned parser
                    1          0
                    ↓          ↓
              continuation    wake parent
                    ↓
                 Remember
                    ↓
              parent observation
```

The downward path establishes what will be attempted. The upward path establishes whether the
attempt actually produced observations for which each parent has a continuation.

---

## Engrams are recursive typed continuations

An engram contains:

```text
ENGRAM

1  CHILD CONTRACTS
   the types this parent expects from each child position

2  COMPILED CONTINUATION
   the behavior this parent performs when those types return
```

The parent owns both parts because its future behavior depends on the child's return.

The return contract for a child position is executable as a binary parser:

```text
parse(child observation)

1    the observation satisfies this child contract;
     the parent's compiled continuation has a next state.

0    the observation does not satisfy this child contract;
     the parent's compiled continuation has no next state for it.
```

The child does not own this parser. The child knows what it produced; only the parent knows what it
can consume.

Each parent parses only its direct child. If it succeeds, that parent executes its continuation and
produces an observation for its own parent:

```text
leaf produces A
    ↓ Parent 1 parses A and produces B
    ↓ Parent 2 parses B and produces C
    ↓ Parent 3 parses C
```

No global parser needs to interpret the complete causal tail. The recursive chain of local parsers
is the parser for the whole composition.

---

## The primitives on the way down

The downward phase discovers, chooses, and establishes behavior.

### Think

```text
QUESTION     What bounded behavior might satisfy the current expectation?
RESULT       Ranked expectation-bearing observations representing potential.
```

Think operates inside runtime-supplied `From` and `Horizon`. During ordinary discovery it may
consider the frame's available possibilities. During failure resolution it receives a smaller
frame centered on one failed child dependency.

Think creates potential. It does not select, execute, or certify that a candidate will return the
required type.

### Decide

```text
QUESTION     Which proposed composition will this engram attempt?
RESULT       One selected and completely bound composition.
```

Decide makes the parent responsible for the selected child composition. The decision includes the
engram and the reach observation which already carries its entity and destination semantics.

### Reach

```text
QUESTION     Can the selected composition be made live within this Horizon?
RESULT       A live possibility, beyond current reach, or beyond your Horizon.
```

Reach realizes exactly what Decide selected. It does not determine whether the eventual child
return satisfies the parent's type. That fact can exist only after the possibility runs and an
observation travels upward.

### Approach

Approach remains the outside-in admission boundary. It establishes the initial namespace, `From`,
and Horizon before the first downward composition begins. It is not the parser for an ordinary
child return.

---

## The turn at the leaf

A leaf grounds the recursion.

```text
LEAF
    child contracts          none
    terminal behavior        produce an observation
```

Once an activated flow successfully reaches a leaf path, that terminal behavior fires. The leaf
has no deeper child observation on which it can wait.

This does not mean every leaf executes continuously. It means a reached leaf is the point where
downward composition turns into upward observation.

---

## The primitives on the way up

The upward phase realizes the accepted flow through actual observations.

### Parent-owned type check

The first operation at each returning boundary is the parser owned by that parent for that child
position:

```text
Observation<Y> returns
        ↓
does Y satisfy the parent's expected Observation<X> here?
        ↓
      1 or 0
```

Typing is not a content label asserted by the child. It is the parent's binary ability or inability
to continue from an observation with substrate-known provenance.

### Remember

When the parser returns `1`, the child observation may be bound into the parent's continuation.
That binding is local Remember:

```text
typed assignment
    = remember the accepted child observation under a parent-owned name
```

The parent then executes the rest of its compiled continuation. Its return is outward Remember:

```text
return
    = expose accepted observations according to the parent's own return contract
```

The same operation repeats upward. At the outer boundary, an unbroken chain of remembrance can
make the accepted causal cone durable.

### Recognize and Orient are kernel protocols

The kernel owns both methods. `Recognize` coordinates the return boundary and calls entity and
engram methods whose composition performs the semantic work. `Orient` reports bounded live
topology directly because that topology already belongs to the kernel. Neither independently opens
a child possibility.

> **The kernel owns the protocol. Entities and engrams compose under it.**

---

## What a type failure means

A type failure occurs on the way up:

```text
Parent expects Child → Observation<X>
Child returns          Observation<Y>

Y ⊭ X
```

The parser returns `0`. This is already a complete structural failure:

```text
the failed parent is known
the failed child position is known
the expected contract is known
the returned observation and its provenance are known
```

No universal `Error` object is required. An explanatory observation may accompany the mismatch,
but it does not create the failure.

The parent wakes because it owns the parser which returned `0` and the continuation which can no
longer proceed.

> **Define a child contract, own its parser, and own the behavior used when that parser returns
> `0`.**

---

## Failure resolution is progressive causal zoom

When a parent wakes, the runtime may give it a frame focused on the exact failed dependency:

```text
FOCUSED FRAME

requesting engram    the parent which owns the failed contract
expected             the child return contract it requires
received             the actual observation which failed
failed child         the exact child position which returned it
available            only relevant observations and behaviors
From                 preserved by the runtime
Horizon              inherited and no wider than before
```

The parent may then use the same downward vocabulary to rectify the failure:

```text
inspect the failed dependency
        ↓
Think within the focused frame
        ↓
Decide on corrective behavior
        ↓
Reach
        ↓
new child observation rises
        ↓
parent parser returns 1 or 0
```

If the corrective behavior fails at a deeper child contract, the owner of that contract wakes into
another, narrower frame:

```text
type failure at child A
    ↓ focus on A
type failure at child B
    ↓ focus on B
type failure at child C
    ↓ focus on C
resolved observation
    ↓ unwind upward
```

This is not a global syscall fallback. It is the same language operating with progressively smaller
instruments. Every zoom must remain causally attached to the failure which created it and must not
gain unrelated authority, observations, or children.

When a satisfying observation finally returns, suspended continuations can resume upward. When no
resolution exists inside the narrowed Horizon, the owning continuation remains structurally unable
to proceed.

---

## What the continuation language means

The surface language can remain ordinary and typed:

```text
METHOD SIGNATURE
    observations already available in this frame

CHILD SIGNATURES
    parent-owned contracts for available child positions

METHOD BODY
    the downward dependency and invocation program

TYPED ASSIGNMENT
    local Remember after a child parser returns 1

RETURN CONTRACT
    the observations this engram must send upward
```

For example:

```text
Destination destination = getDestination(request);
```

expresses both directions:

```text
DOWN
    choose and invoke the getDestination child position with request

UP
    parse the returned observation as Destination
    1 → remember it locally as destination and continue
    0 → do not create destination; wake this parent at getDestination
```

The language does not need a generic `Error` value or ambient inspection syntax to express this.
The missing typed assignment is the stopped continuation. The failed child position determines the
focus of subsequent inspection and repair.

---

## Whole-system rhythm

```text
APPROACH
    ↓ establish initial From and Horizon

THINK
    ↓ propose bounded potential

DECIDE
    ↓ select and bind one composition

REACH
    ↓ make it live

LEAF FIRES
    ↓ originate observation

PARENT PARSES
    ├── 1 → continuation → Remember → rise
    └── 0 → wake → focused frame
                         ↓
                   Think → Decide → Reach
                         ↓
                    attempt resolution
```

Accepted and type-stable flows spend their meaningful work on the way up: each engram consumes a
child observation and produces the next observation. Engrams change on the way down when discovery
or repair requires a different composition.

---

## Current invariants

```text
1  A parent owns the contracts for each of its child positions.
2  Each parent-owned return contract acts as a binary parser for that child.
3  Runtime return mismatch is discovered on the way up.
4  Parser result 1 advances the compiled continuation.
5  Parser result 0 wakes the parent which owns the failed contract.
6  A reached leaf fires terminal behavior and grounds the upward flow.
7  A failure frame is centered on one failed dependency.
8  Recursive repair frames may narrow but never widen Horizon.
9  Inspection follows the failed dependency; it is not ambient graph access.
10 Resolution uses the same Think → Decide → Reach vocabulary as ordinary composition.
11 Successful resolution resumes suspended continuations upward.
12 No universal Error object or global parser is required.
```

---

## Still open

```text
the exact language form, if any, for behavior owned when a child parser returns 0
whether inspection is the first required behavior in every focused failure frame
what proves that each recursive frame is strictly narrower than its parent
how many alternate continuations may be attempted before the parent enters focused discovery
how suspension and upward resumption are represented causally
how an irreducible failure is observed when no correction exists within Horizon
```

These questions affect the expression and scheduling of recovery. They do not change the central
shape:

> **Program downward. Fire at the leaf. Parse and continue upward. On `0`, wake the parser's owner
> and zoom into the failed dependency.**

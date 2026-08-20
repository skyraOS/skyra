# Space Authority Over Spans

**Date:** 2026-08-09  
**Status: FUTURE POSSIBILITY — NOT CANON.** This is downstream of unsettled work. It should not be
built from, cited, or merged into the standing set until the meaning of entity and the execution
model are resolved.

> **A span is a capability owned by the space that created it.**

## The idea

A selection initiates work. The running space that made that selection retains authority over the
resulting spans rather than surrendering their control to a global traversal mechanism.

That authority may eventually let the space:

```text
LIST     the spans it created.
LOCATE   the current head of each span.
WALK     its own dependency train.
ACT      through syscalls permitted in its frame.
CLOSE    work it owns by return, cancellation, or another terminal operation.
```

These operations expose locally owned execution state, not the global graph. The runtime maintains
span identity and enforces authority; it does not decide what a span means or how the owner should
use it.

## Authority belongs to a running space

A durable possibility may be invoked more than once and may be executing concurrently. Span
authority therefore cannot belong to the possibility definition as shared global state. It belongs
to the particular running space produced by an invocation.

```text
possibility P
├── space P1    owns only spans created by P1
└── space P2    owns only spans created by P2
```

P1 cannot inspect or manipulate P2's spans merely because both instantiate the same possibility.
The durable possibility may learn from both histories later; that does not merge their authority
while they run.

## Nested locality

Authority follows initiation:

```text
A selects B       A owns the span by which it invoked B.
B selects C       B owns the span by which it invoked C.
```

Whether A may walk through B to the head at C, and what A is allowed to observe while doing so, is
not settled. The minimum constraint is that no operation turns local span authority into permission
to enumerate or control unrelated work.

## Operators

The authority model may be common to every space even when different kinds of operators man them.
A model, human, or program could each inspect and act on the spans of the space it is currently
operating through a surface appropriate to that operator.

```text
OPERATOR   determines how a space is operated.
SPACE      owns the work it initiated.
RUNTIME    records spans and enforces ownership.
```

This would make a human walking a dependency train an instance of the ordinary execution model, not
a human-specific exception. It must remain separate from the human acting as an upstream grader: an
invokable operator in one relationship does not make the grader reachable from below.

## Relation to invocation cardinality

Invocation cardinality describes how a space opens work: one invocation establishes a narrow step;
several in one group expose a concurrent frontier. Span authority describes what the space may do
with the work it opened afterward.

The two ideas do not yet settle whether every invocation group is a barrier, whether completed
members can be inspected before the whole group returns, or whether a space can continue emitting
while owned spans remain active.

## Deferred questions

```text
ENTITY              its meaning must be cleaned up before operator-specific behavior is assigned.
SPAN BOUNDARY       whether ownership covers one edge, one chain, or some explicitly delegated set.
WALK AUTHORITY      what an owner may see or change beyond the directly invoked space.
RETURN              what happens to active spans when their owning space returns.
DELEGATION          whether span authority can move, and how that transfer is cited.
SYSCALL SURFACE     which operations exist and how their results enter the frame.
HUMAN EXECUTION     how a human operates a space without becoming its grader or global observer.
```

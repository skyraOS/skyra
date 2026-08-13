# Primitives

**Date:** 2026-08-13  
**Status: WORKING NOTE.** Seven kernel-backed primitives. The first six form the execution and
learning cycle. The seventh observes live execution.

---

## The seven

```text
APPROACH      Should this be let in at all?
THINK         What behavior might address this need?
ENGAGE        Choose behavior and bind what it will receive.
REACH         Resolve and contact the entity that will interpret it.
REMEMBER      Receive its result and retain the causal path chosen by its parent.
RECOGNIZE     Translate what was remembered and match it against prior experience.
ORIENT        Observe where you are and what is happening in live execution.
```

These are kernel-backed because each depends on something only the kernel can safely coordinate:
scope, live topology, routing, causal order, or the transition from ephemeral to durable state.
They do not all place semantic judgement inside the kernel. Where meaning must be interpreted, the
kernel opens behavior that performs that interpretation.

---

## The primary cycle

```text
APPROACH
    ↓
THINK
    ↓
ENGAGE
    ↓
REACH
    ↓
REMEMBER
    ↓
RECOGNIZE
```

The cycle may recurse after recognition:

```text
approach → think → engage → reach → remember
                         ↓
                     recognize
                         ↓
                engage known behavior
                         ↓
                   reach → remember
```

`Orient` is adjacent to this cycle rather than another step in it. It can inspect the live
execution at any point without deciding what that execution means.

---

## 1 · Approach

```text
QUESTION     Should this be let in at all?
READS        the boundary — who is reaching, and with what.
RETURNS      an admitted root, or a recorded avoidance.
```

Approach is **always at a root, and always outside-in**. It is the only way anything enters, which
puts the trust boundary and the gate on the same object — a gate anywhere else would be guarding a
door with no wall attached.

```text
APPROACH     admitted. a root opens, and the observation enters with an empty producer because
             no possibility produced it.
AVOID        refused, AND RECORDED. "<caller> was avoided" is an event with a subject, not an
             absence — the same property that makes "reached for and couldn't" worth having.
```

This is where the maximum-privilege operation in the design sits, and `questions.md` §5 is the
record of it being ungated: *a root originates freely, holds the grants, is the grader, and
materializes.* Naming the boundary does not gate it — but it establishes that there is exactly one
place a gate can go.

**Approach is recursive, for the same reason reach is.** It applies at every boundary where
something outside contacts something inside:

```text
a human       → approaches the runtime
the runtime   → approaches an entity it composes with
that entity   → is a whole system from inside, one party from outside
```

**And it is where bootstrap becomes expressible.** `discover.md` names bootstrap as the
maximally-nil case — every blank empty. `Engage` cannot express that: it is a parent choosing, and
the first engagement has no parent. An approach has no parent by construction.

---

## 2 · Think

```text
QUESTION     What behavior might address this need?
READS        durable graph state — what has happened.
RETURNS      candidates that may be engaged.
```

Thinking searches and reasons over known engrams, entities, observations, or hosts. It is learnable
behavior, not a hardcoded ranking table: different thinking strategies can accumulate history and
compete.

The regress stops at a fixed kernel-known name. `think` is always nameable; thinking about how to
think is not required before the first thought can run.

The kernel supplies access to durable structure. The thinking engram interprets the need and ranks
candidates. The kernel does not decide which candidate is semantically appropriate.

---

## 3 · Engage

```text
QUESTION     Which behavior will run, with which observations and entity?
READS        the current frame and names in scope.
RETURNS      one or more authorized invocations.
```

Engagement is the choice to open an engram. It binds:

```text
ENGRAM       the behavior being opened.
ENTITY       who will interpret it.
INPUTS       observations passed into its frame.
CITATION     which visible need prompted the choice.
ACCESS       whatever reach and grants are explicitly provisioned downward.
```

Engagement precedes reach. The parent first chooses what and who; the kernel then determines how to
reach that entity.

For a concurrent engagement group, emitted member order is meaningful and completion order is not.
The kernel preserves emitted order while allowing the engaged work to execute concurrently.

---

## 4 · Reach

```text
QUESTION     How can this engaged entity be contacted now?
READS        volatile host endpoints and prior reach history.
RETURNS      a route, delivery, or loud failure.
```

An entity is an identity and does not carry its endpoints. Hosts hold endpoints. Reach resolves an
entity to an address and transport, then obtains an opaque handle or delivery path.

```text
ENTITY       stable identity; keyed in history.
HOST         a place from which endpoints are available.
ADDRESS      the name being resolved.
TRANSPORT    the mechanism used to contact it.
```

The kernel routes, but learnable reach behavior ranks alternatives. A failed endpoint and a working
fallback are observations from which reach can improve.

Like `think`, `reach` has a fixed kernel-known name so resolution does not recursively require
another unresolved resolver.

Reach happens before the possibility opens. Failure must be immediate and loud: an undeliverable
engagement is a failed invocation, not a possibility silently opened with nowhere to run.

---

## 5 · Remember

```text
QUESTION     What did this invocation produce, and does its parent keep it?
READS        the entity's emission after it crosses back.
WRITES       a causal remembrance into its parent; the outermost remembrance becomes durable.
```

`Remember` is the single operation previously split between handing a result upward and `Accept`.
A child does not merely hand an observation to its parent: the parent remembers that observation as
part of its own live causal history.

Remembering at an inner boundary is still free in the durable sense:

```text
ENTITY EMITS
    ↓
PARENT REMEMBERS the observation and its causal position ephemerally
    ↓
PARENT may use, translate, revise after, or ignore it
```

The same operation repeats at every level. A parent remembers a child's observation; its parent may
then remember the composition; eventually the outermost caller remembers a result. That outermost
remembrance closes the live execution and makes the recursively remembered causal cone durable.

```text
leaf observation
      ↓ remembered by its parent
parent observation
      ↓ remembered by its parent
outer observation
      ↓ remembered by the outermost caller
durable causal cone
```

An observation is not durable merely because an inner parent remembered it into a live frame. It
becomes durable only when an unbroken chain of remembrance reaches the outer boundary.

Remember must preserve:

```text
SEQUENCE             order within its possibility.
CAUSE                which invocation produced it.
EMITTED POSITION     its place in a concurrent engagement group.
PROVENANCE           the typed distinction between data and delegated reach.
```

At the outer boundary, `Remember` walks backward from the chosen observation:

```text
chosen observation
        ↓ backward causal walk
translations · engagements · remembrances · expectations that contributed
        ↓ projection
durable graph records and materialized reached effects
```

The kernel does not interpret which content was good. Each parent chose what to remember, and the
kernel follows those structural causal edges. Work outside that cone evaporates when live execution
closes.

Remember is selective projection, not deletion from an already-written graph:

```text
EPHEMERAL CAUSAL RECORD
          ↓ outermost Remember projects one retained cone
DURABLE GRAPH
```

This is also where provisional compiled behavior may receive durable identity, ownership, and
history as an engram. The exact installation contract remains to be specified.

Remember decides which staged effects materialize: effects belonging to reached producers may
land; every other staged effect must be explicitly discarded.

---

## 6 · Recognize

```text
QUESTION     Have I seen this observation, or something like it, before?
READS        an observation just remembered into the live frame, and durable remembered experience.
RETURNS      a comparable representation and applicable known behavior.
```

Recognition occurs while execution is still ephemeral:

```text
remembered observation
        ↓
translator behavior
        ↓
canonical or class key
        ↓
structural lookup
        ↓
known automatic behavior
```

The division is strict:

```text
TRANSLATE     an engram interprets the observation and makes likeness explicit.
MATCH         the kernel compares explicit keys and retrieves remembered associations.
DISPATCH      the kernel engages or opens the associated behavior through the ordinary cycle.
```

The kernel may compare exact keys. It may not decide from content that two observations mean the
same thing. “Something like this” remains semantic until translator behavior produces a key that
makes the claimed likeness structural.

Recognition has several possible strengths:

```text
IDENTICAL     the same canonical material has been seen.
SIMILAR       the same explicit shape or partition has been seen.
FAMILIAR      translator behavior assigns a known class key.
```

Recognition chooses behavior to run now. It must never substitute an old observation for a fresh
invocation merely because the inputs look familiar.

Anything recognition produces joins live causal state through the same `Remember` operation. It
remains ephemeral unless its chain of remembrance reaches the outer boundary.

---

## 7 · Orient

```text
QUESTION     Where am I, and what is happening in live execution right now?
READS        ephemeral kernel topology.
RETURNS      structural descriptions of live work.
```

Orientation pairs with thinking:

```text
THINK          reads the graph       what has happened, and what might address this need.
ORIENT         reads the kernel      where am I, and what is happening.
```

It reports topology, not meaning:

```text
spans · engram IDs · entity IDs · host IDs · depth · live heads · remembered leaves
```

It must not expose expectations, observation content, or reception history. Handles may be supplied
for a caller to resolve against authority it already holds.

Orientation is bounded downward by the live tree. It cannot reveal ancestors or siblings merely
because they share the same kernel. That structural bound preserves the rule that a possibility
cannot observe its own reception or optimize directly for being cited.

`orient` is also reached through a fixed kernel-known name. It is reflective behavior over
kernel topology, not a separate control plane.

---

## Words absorbed by the primitives

Several apparent operations reduce to these six:

```text
CREATE       remember provisional behavior into the live frame; retain it only if the outer chain does.
COMPILE      recognize/translate, execute the candidate, then remember it if retained.
ACCEPT       renamed Remember.
INGRESS      renamed Approach — and the rename carries the gate, because approach has an
             opposite and ingress did not.
REJECT       close without Remember; ephemeral work evaporates and effects are discarded.
COMMIT       Remember.
RESOLVE      the central operation inside Reach.
MATCH        structural comparison inside Recognize.
ROUTE        kernel coordination inside Reach.
FORGET       unnecessary for provisional work; do not Remember it.
```

---

## Boundary

The primitives divide responsibility without turning the kernel into a semantic interpreter:

```text
KERNEL
    holds scope, live topology, causal order, routes, exact keys, and closure

ENGRAMS + ENTITIES
    interpret needs, rank candidates, translate observations, recognize meaning, and produce results

GRAPH
    witnesses only what Remember retained
```

In one sentence:

> **Approach decides what is let in at all; Think finds and reasons over possibilities; Engage chooses one; Reach contacts its interpreter;
> Remember brings evidence back and retains its causal path; Recognize connects it to experience;
> Orient locates the caller in the live structure throughout.**

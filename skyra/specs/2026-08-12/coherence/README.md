# Coherence map — 2026-08-13

**Status: READING ORDER AND AUTHORITY MAP.** This document does not add a primitive or mechanism.
It states which documents control when the August 12–13 notes use older surfaces or overlapping
vocabulary.

---

## Current center

Read these first, in order:

1. [`primitives-and-continuation-summary.md`](../continuation-language/primitives-and-continuation-summary.md)
   — the current directional synthesis.
2. [`primitives.md`](../continuation-language/primitives.md) — the seven semantic distinctions and
   their surface roles.
3. [`think-interface-proposal.md`](../continuation-language/think-interface-proposal.md) — Think as
   the universal possibility-facing interface.
4. [`engram-contracts-and-composition.md`](../continuation-language/engram-contracts-and-composition.md)
   — parent-owned child contracts, parsers, and continuation ownership.
5. [`continuation-language.md`](../continuation-language/continuation-language.md) — the typed
   LLM-facing language and focused failure frames.
6. [`distributed-system-summary.md`](../../../summaries/distributed-system-summary.md) — the same
   model across runtime and authority boundaries.

When one of these conflicts with a carried semantic note, code snapshot, worked-example outcome, or
earlier Reach proposal, this list controls.

---

## Current coherent shape

```text
SEVEN SEMANTIC DISTINCTIONS

Approach     outside-in admission boundary
Think        universal possibility-facing interface
Decide       downward causal form
Reach        hidden realization of one decided composition
Remember     upward causal form
Recognize    specialized behavior beneath Think
Orient       specialized behavior beneath Think
```

The seven meanings are not seven peer methods.

```text
DOWN
    Think proposes bounded potential
    Decide selects and binds one composition
    Reach attempts to make that exact composition live
    the continuation composes child dependencies

TURN
    a reached leaf fires terminal behavior

UP
    a child returns an observation
    its parent-owned parser returns 1 or 0
    1 advances the continuation and Remember carries the result upward
    0 wakes the parent which owns the failed child contract
```

After `0`, inspection follows the exact failed dependency. The runtime may supply the owning parent
a narrower frame containing the expected contract, actual observation, failed child position, and
only relevant behavior. Further failures zoom into still smaller frames. `From` is preserved and
Horizon never widens.

---

## Stable invariants

```text
1  An engram is child contracts plus compiled continuation.
2  A parent owns the input and return contracts for each child position it invokes.
3  A parent-owned return contract is the parser for that child return.
4  The child cannot self-certify compatibility by writing a type label into content.
5  Compile-time source incompatibility, unrealizable Reach, and return mismatch are distinct.
6  A reached leaf fires because it has no deeper child return on which to wait.
7  Think creates bounded potential; it does not silently execute a candidate.
8  Decide chooses one complete composition.
9  Reach realizes exactly that composition or returns bounded impossibility.
10 Remember determines which causal path travels upward and ultimately becomes durable.
11 The durable graph records accepted causal history, not compatibility or live reachability.
12 Address and transport remain behind the kernel boundary.
13 Local and remote are relations to runtime-supplied From, not entity properties.
14 Horizon is authority, is separate from the frame, and can only narrow.
15 Structural failure does not require a universal Error object.
```

---

## Reach vocabulary

The original [`reach-boundaries-proposal.md`](../reach/reach-boundaries-proposal.md) combined two
stages. Read its ranked-list sections as reach-planning beneath Think:

```text
think-reach-entity    returns bounded ranked relationship potential through Think
Decide                selects one member
Reach                 attempts to establish that member as a live possibility
```

The following parts of that proposal remain current for actual Reach:

```text
entity preservation
runtime-supplied From
monotonically narrowing Horizon
local and remote strategies behind one abstraction
endpoint mechanism hidden below the semantic boundary
beyond current reach distinct from beyond your Horizon
application work beginning only after the relationship exists
```

[`reach-abstraction.md`](../reach/reach-abstraction.md) is an operating-system reference, not the
public Skyra signature. [`distance.md`](../energy/distance.md) is a working measurement proposal.

---

## Document roles

### Continuation language

The documents in [`continuation-language/`](../continuation-language/) are the active proposals.
The summary states the shared shape; the longer documents own their narrower subjects.

### Causality

[`observable-causality.md`](../causality/observable-causality.md) controls claims about what the
substrate can prove. [`causality-and-recognition.md`](../causality/causality-and-recognition.md) is
the broader exploratory note about learning, cost, and recognition.

### Distributed system

[`proprietary-behavior-and-open-protocol.md`](../distributed-system/proprietary-behavior-and-open-protocol.md)
defines the open seam around potentially opaque behavior. [`gravitate.md`](../distributed-system/gravitate.md)
names a possible emergent pattern, not an additional primitive.

### Carried semantics

[`semantics/`](../semantics/) preserves reasoning from August 11. Its causal distinctions and
no-control-plane argument remain useful. Its seven-peer-method ruling is superseded. Each relevant
document now marks that boundary.

### Structs and interfaces

[`structs and interfaces/`](../structs%20and%20interfaces/) is primarily a snapshot of code which
existed at the time. A snapshot may legitimately disagree with the proposed semantic surface.
Documents under `interface changes/` are proposals; the seven-peer-method proposal is now marked
partially superseded.

### Syscall read example

[`syscall-read-example/`](../syscall-read-example/) is evidence and decision history which led to
the new abstractions. Its `outcomes/` are Michael-owned and remain untouched. Direct syscall or
peer-method claims there are historical when the current center says otherwise.

### Work orders and research

[`overnightBuild/`](../overnightBuild/) is a bounded historical work order, not current semantic
authority. [`future-research/`](../future-research/) and [`misc/`](../misc/) are non-authoritative
research and reference material.

---

## Failure ownership

The failure model is local to the boundary which failed:

```text
INVALID SOURCE PROGRAM
    owner: continuation compiler
    result: no decision

UNREALIZABLE DECISION
    owner: Reach boundary
    result: beyond current reach or beyond your Horizon

UNSATISFIED CHILD RETURN
    owner: parent which defined that child contract
    result: parser 0, parent wakes, focused repair frame
```

This prevents transport failure, type failure, and source-language failure from collapsing into one
generic error channel.

---

## Open rather than contradictory

The following remain intentionally unsettled:

```text
the concrete continuation syntax and parser implementation
the exact representation of observation contracts
the exact inspect observation and whether every focused frame begins with it
branching, retry, concurrency, suspension, and resumption syntax
the proof that each recursive repair frame is strictly narrower
leaf origination syntax
the concrete From and Horizon representations
remote causal envelopes, retry identity, and effect disposition
which provisional failure evidence becomes durable
```

These are open details inside the current boundaries. They should not be resolved by reviving an
older surface or collapsing two distinct failure phases.

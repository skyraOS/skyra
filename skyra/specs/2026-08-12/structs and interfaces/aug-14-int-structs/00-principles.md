# 00 · Principles

**Status: DERIVED.** Not new claims. These are the constraints the source documents already make,
restated as tests a type surface can fail.

> **Every rule below is stated so that violating it is a compile error or a structural
> impossibility, not a policy someone must remember.**

---

## The constraints

Consolidated from `../../coherence/README.md`'s fifteen invariants and the acceptance tests in
`think-interface-proposal.md`, `engram-contracts-and-composition.md`, and
`continuation-language.md`. Numbered here for citation inside this directory.

```text
P1   AN ENGRAM IS CHILD CONTRACTS PLUS COMPILED CONTINUATION.
     nothing else is part of what an engram IS.

     ⚠ CONTESTED 08-15. creator-and-compiler.md gives an engram a CREATOR SLOT and a COMPILER
     SLOT, fixed at solidification. Either P1 is now wrong, or the slots are not part of what an
     engram IS but of how it is REACHED — the way Entity was per-invocation rather than a field.
     The second reading preserves P1 and is probably right, since the compiler is determined by
     the body's syntax rather than chosen, but it was not ruled. NOT RESOLVED.

P2   A PARENT OWNS THE INPUT AND RETURN CONTRACTS for each child position it invokes.
     no engram owns the contracts by which every caller must use it.

P3   THE PARENT-OWNED RETURN CONTRACT DECIDES whether that child return may continue.
     binary. 1 continues the continuation, 0 wakes the parent.

     REVISED 08-15. the contract is a TYPE NAME, and the check is performed BY THE KERNEL — a
     comparison of parent.Contracts[pos].Return against child.Returns. The parent still owns what
     is expected; it does not own the act of checking, because checking is causal. See P19.

P4   A CHILD CANNOT SELF-CERTIFY COMPATIBILITY by writing a type label into content.

     REVISED 08-15. typing is TWO SUBSTRATE-HELD DECLARATIONS compared by the kernel — what the
     parent expects, and what the child engram exposes. Both live on ENGRAMS, which are durable and
     substrate-held; neither is in observation content. A child cannot forge what it does not
     author, which is what this principle is actually about.

P5   THREE FAILURES STAY DISTINCT, AND EACH HAS A PARTY WHO ACTS.
     REVISED 08-15 — the three are different three, and each now names an actor rather than a
     boundary. see coherence.md §2.2 and unreachability-and-escalation.md §3.

         1  UNREALIZABLE DECISION   detected at REACH    → acted on by THE NAMESPACE
         2  FAILURE TO COMPILE      detected on compile  → acted on by THE CREATOR
         3  UNSATISFIED RETURN      detected by parser   → acted on by THE PARENT

     no universal Error object.

     WAS: compile-time incompatibility · unrealizable Reach · return mismatch. The first was
     failure kind 1, checked before Reach; it is removed, since Recognize runs on the return path
     only. Kind 2 above is a DIFFERENT failure with a confusingly similar name — the child's own
     body failing to compile, which the old set had no slot for.

P6   A REACHED LEAF FIRES because it has no deeper child return to wait on.
     leaf and orchestrator are positions, never declared kinds.

P7   THINK CREATES BOUNDED POTENTIAL and does not execute a candidate.

P8   DECIDE CHOOSES ONE COMPLETE COMPOSITION.

P9   REACH REALIZES EXACTLY THAT COMPOSITION or returns bounded impossibility.
     it may not substitute entity, expectation, input, position, or Horizon.

P10  REMEMBER DETERMINES which causal path travels upward and ultimately becomes durable.

P11  THE DURABLE GRAPH RECORDS ACCEPTED CAUSAL HISTORY, not compatibility and not live
     reachability.

P12  ADDRESS AND TRANSPORT REMAIN BEHIND THE KERNEL BOUNDARY.

P13  LOCAL AND REMOTE ARE RELATIONS to runtime-supplied `From`, never entity properties.

P14  HORIZON IS AUTHORITY, is separate from the frame, and can only narrow.

P15  STRUCTURAL FAILURE DOES NOT REQUIRE a universal Error object.
```

Four more the type surface needs and the invariant list does not state as such. P16 and P17 are
taken verbatim in substance from their own documents; P18 and P19 were added 2026-08-15.

```text
P16  THE KERNEL MAY HOLD WHAT IT MAY NOT READ.
     `../../semantics/the-crossing.md`. holding an opaque blob is fine; reaching into it is not.

P17  REMEMBERED ⊆ OCCURRED.
     `../../causality/observable-causality.md`. something can occur without being remembered;
     nothing can be remembered without having occurred.

P18  DO NOT AUDIT WHAT CAN BE INFINITELY GENERATED. ENFORCE BELOW THE BOUNDARY.
     added 2026-08-15. an AI generates behavior without bound, so any scheme requiring each
     artifact to be reviewed before it may run does not scale and will be abandoned under load.
     The enforceable surface is the boundary AROUND a process, not the contents of one.
```

```text
P19  THE KERNEL OWNS CAUSALITY — AND TYPE CHECKING IS CAUSAL.
     added 2026-08-15. Decide and Remember were already the two causal forms. Recognize belongs
     with them, because a check that returns 0 STOPS THE CHAIN. That is not a judgment about a
     value; it is an act that determines whether causal structure extends.

         DECIDE      extends causal structure downward.
         RECOGNIZE   determines whether it continues upward.
         REMEMBER    determines what survives.

     ALL THREE ARE KERNEL METHODS. Nothing outside the kernel may hold one.
```

### P19 demotes the compiler to a parser

The corollary is the whole shape of the compiler slot:

> **The causal syntax lives at an address, and all it is is a parser that calls the kernel's
> methods.**

Which is how every language works — source compiles to syscalls. A compiler that held `Decide` or
`Remember` would be extending causal structure from outside the kernel, and `primitives.md` reserves
that: *"Decide and Remember... are the only two primitives that change the live causal structure."*

```text
COMPILERS OWN   their syntax. nothing else.
THE KERNEL OWNS the methods, and scope · topology · causal order · routing · closure.
```

**This was got wrong twice in one session** — a `Compiler` interface carrying the three methods was
proposed, written down, and reverted. The failure both times was treating a type check as a
predicate rather than as a causal event. Recorded here so the third attempt does not happen.

### P18 completes a family

The design already makes this move twice. Stating it as a principle keeps someone from re-deriving
the discarded alternative later:

```text
DON'T ASSERT WHAT CAN ONLY BE FOUND OUT     → shape belongs to the pairing and is observed, never
                                              declared. 2026-08-09/interfaces.md.

DON'T ASK AN ENTITY WHAT IT BELIEVES        → capability declarations dissolved. the system records
IT CAN DO                                     what happened when an entity was tried.

DON'T AUDIT WHAT CAN BE INFINITELY          → sandbox the process; the language, the compiler, and
GENERATED                                     the contents are all irrelevant to enforcement.
```

Each replaces a judgment that does not scale with a structure that does.

**The concrete consequence, from the 08-15 sandbox session:** enforcement lives at the process
boundary, not in the language. Every effect a process can have on the world passes through a
syscall, so the syscall boundary is not *a* chokepoint for outbound effect — it is *the* chokepoint.
Filter there and what runs above it genuinely cannot matter. This is why a restricted language
subset was rejected: it solves the problem at a layer that has to be re-solved per language, and
never covers the next one.

```text
COROLLARY   arbitrary software is just an entity. an LLM, a python process, and a browser all have
            opaque interiors and mediated effects, which is the entity contract already. the causal
            record never claimed to see inside anything — it records exposure, choice, order,
            output, remembrance, and acceptance, and ignores hidden activation and private
            reasoning alike.
```

---

## How each one becomes structural

A principle enforced by discipline is not enforced. This is where each lands in the type surface.

```text
P1   Engram has exactly two semantic fields.          durable/01-engram.md
P2   a contract is reached through a Position that    contracts/02-child-contract.md
     names its owning parent. there is no
     contract-on-child accessor to call.
P3   Child.Return is a string, and only the Kernel    contracts/02-child-contract.md
     compares it. no type outside the kernel has a
     method that could perform the check.
P4   both type labels live on Engrams, never in       durable/01-engram.md
     content. content cannot express one, so a
     child cannot write one.
P5   three distinct result types, no shared           kernel/05-reach.md, creator-and-compiler.md §4,
     Error interface between them.                    contracts/03-parser.md
P6   no Kind field anywhere. a leaf is an engram       durable/01-engram.md
     whose Contracts map is empty.
P7   Think returns []Candidate, and Candidate has      kernel/03-think.md
     no method that runs anything.
P8   Decide takes exactly one candidate.               kernel/04-decide.md
P9   Reach receives a Decision and cannot construct    kernel/05-reach.md
     one; the type has no exported fields to set.
P10  Remember is the only path to the graph.           kernel/06-remember.md
P11  Graph has no contract type and no liveness        durable/02-graph.md
     field. it structurally cannot record either.
P12  no Address or Transport type is reachable         kernel/05-reach.md
     from any possibility-facing signature.
P13  From is runtime-supplied and has no setter.       identity/02-from.md
P14  Horizon is opaque and its only operation          identity/03-horizon.md
     narrows.
P15  see P5.
P16  Rendering and Delivery are opaque interfaces      durable/03-crossing.md
     with unexported marker methods.
P17  the graph is written only by Remember, which      durable/02-graph.md
     projects a cone of witnessed events.
P18  no type anywhere describes what a compiler        creator-and-compiler.md
     is permitted to contain. the enforcement
     surface is the endpoint set in Horizon, and
     the sandbox around the process.
P19  Decide, Recognize and Remember appear on the      kernel/01-protocol.md
     Kernel and nowhere else. Engram.Compiler is
     an Address, not an interface — there is no
     type through which a compiler could hold one.
```

---

## Deliberately deferred to a later version

Recorded so they are not mistaken for oversights.

```text
STRICTER TYPING   a contract is a type NAME, and Recognize is a string comparison between two
                  substrate-held declarations. There is NO STRUCTURAL CHECK anywhere — two engrams
                  both declaring "Destination" match regardless of what they return.

                  This is consistent with correctness being Remember's job rather than the type
                  system's (observable-causality.md §4, "correctness is local to acceptance"), and
                  it is a real narrowing from the structural/nominal split proposed earlier in
                  primitives-as-engrams.md §5.

                  ACCEPTED FOR v6. "We have to cut it somewhere." Stricter typing can be added
                  later without moving the boundary, because the boundary — the kernel compares,
                  and never reads content — does not depend on how rich the comparison is.
```

---

## The two tests this surface must pass

Both are borrowed, and both are mechanical.

**Desugaring**, from `surfaces-and-solidification.md`:

```text
EVERY LANGUAGE CONSTRUCT   has a primitive sequence it compiles to.
EVERY PRIMITIVE SEQUENCE   has a language form, or is DISCOVER / ORIGINATE and deliberately
                           has none.
```

If the type surface lets the continuation language express something the seven primitives cannot,
or leaves a primitive composition with no surface form, there are two systems wearing one name.

**Three-layer separation**, from `engram-contracts-and-composition.md`:

```text
CONTRACTS               say what CAN compose.
DECIDE and REACH        say what DID compose.
REMEMBER                says which composition SURVIVES.
```

Each layer gets its own types in this outline and they share none. A type appearing in two layers
is the defect that document names three ways: potential looking like experience, exploration looking
accepted, and what worked once becoming the only thing considered compatible later.

---

## What is deliberately not a principle

```text
NO GLOBAL TYPE HIERARCHY      a type name is a contract meaningful in the continuation that owns
                              it. the same observation may satisfy a contract under one parent and
                              not another.

NO PERSISTENT TYPE EDGE       Parent ──type──▶ Child is rejected. it would go stale and would be
                              mistaken for evidence that composition succeeded.

NO CAPABILITY DECLARATION     an entity is an identity and nothing else. the system records what
                              happened when an entity was tried.

NO AGE AS A SELECTION INPUT   `../../reach/age.md`: age is a resolution coordinate. reading it as
                              quality restores the global goodness prior wearing a version number.
```

# aug-14 — the type surface, outlined

**Date:** 2026-08-14
**Status: OUTLINE.** A proposed type surface for the seven-primitive model. This is not a snapshot
of code and does not describe `skyra-dna-v5` as built.

> **This outlines the types the current semantic center implies. It is derived from
> `continuation-language/`, not from the runtime. Where the runtime already has a type doing a
> related job, that is recorded as a reconciliation note, never as authority.**

---

## What this is, and what the sibling directory is

```text
../01-ids.md … ../09-trace.md       LIVE SNAPSHOT of skyra-dna-v5. accurate as a record of code.
                                    NOT a design target. its own README lists what to cut.

THIS DIRECTORY                      the surface the seven primitives imply. a proposal.
                                    nothing here exists in the tree.
```

The two disagree on purpose. `../README.md` says the source tree is authoritative for *what is
built*; `../../coherence/README.md` says the continuation-language documents control *what is
intended*, and that "a snapshot may legitimately disagree with the proposed semantic surface." This
directory is written against the second.

## Authority

Everything here derives from, and defers to:

```text
../../continuation-language/primitives-and-continuation-summary.md    the directional model
../../continuation-language/primitives.md                             the seven
../../continuation-language/think-interface-proposal.md               Think's boundary
../../continuation-language/engram-contracts-and-composition.md       parent-owned contracts
../../continuation-language/continuation-language.md                  the typed surface
../../continuation-language/surfaces-and-solidification.md            two surfaces, one semantics
../../semantics/terminology.md                                        AUTHORITATIVE on vocabulary
../../semantics/the-crossing.md                                       hold vs read
../../causality/observable-causality.md                               what the substrate can prove
../../reach/age.md                                                    versioning, positional
```

**If a type here contradicts one of those, the document wins and the type is wrong.**

## The documents

| doc | holds |
|---|---|
| [00-principles](00-principles.md) | the constraints every type below must satisfy, and how to fail one |
| [identity/01-ids](identity/01-ids.md) | the identity types, and the one new one |
| [identity/02-from](identity/02-from.md) | `From` — namespace × requesting engram × runtime |
| [identity/03-horizon](identity/03-horizon.md) | `Horizon` — authority that only narrows |
| [identity/04-position](identity/04-position.md) | a child position, and age as a coordinate |
| [contracts/01-observation](contracts/01-observation.md) | the observation, and substrate-known provenance |
| [contracts/02-child-contract](contracts/02-child-contract.md) | expectation, input, return — all parent-owned |
| [contracts/03-parser](contracts/03-parser.md) | the return contract as a binary parser, **and where it can run** |
| [contracts/04-continuation](contracts/04-continuation.md) | the compiled continuation, and the frame |
| [kernel/01-protocol](kernel/01-protocol.md) | the seven as one protocol, and what is not a peer method |
| [kernel/02-approach](kernel/02-approach.md) | the admission boundary |
| [kernel/03-think](kernel/03-think.md) | the universal possibility-facing interface |
| [kernel/04-decide](kernel/04-decide.md) | the choice that is the binding |
| [kernel/05-reach](kernel/05-reach.md) | hidden realization, and the version coordinate |
| [kernel/06-remember](kernel/06-remember.md) | two boundaries, one operation |
| [kernel/07-recognize-orient](kernel/07-recognize-orient.md) | the two remaining kernel protocols |
| [durable/01-engram](durable/01-engram.md) | contracts plus continuation, plus the version chain |
| [durable/02-graph](durable/02-graph.md) | what Remember projects, and what it may not hold |
| [durable/03-crossing](durable/03-crossing.md) | the interpreter boundary, held and unread |
| [open-questions](open-questions.md) | every hole, what depends on it, and what constrains it |

**Loose, not part of the structure above:**

| doc | holds |
|---|---|
| [primitives-as-engrams](primitives-as-engrams.md) | **CONVERSATION RECORD, UNRATIFIED.** Primitives as leaf engrams, the provenance floor, Recognize as compile. If ratified it substantially resolves H1 and supersedes `contracts/03-parser.md` §3. Parked while H16 is worked. |
| [creator-and-compiler](creator-and-compiler.md) | **CONVERSATION RECORD, UNRATIFIED.** `Entity` dissolves into a creator slot and a compiler slot, both fixed at solidification. Recognize is return-path only; failure kind 1 is removed; compile failure belongs to the child. Answers H16 and makes `questions.md §2` blocking. |

| [approach-and-root](approach-and-root.md) | **CONVERSATION RECORD, UNRATIFIED.** One gate, two uses: Approach admits from outside and elevates from inside — **sudo is Approach**. Sign-up is interior provisioning. `root` = the namespace with no provisioner, dormant and never in the call path. Places H2, carries Linux/AWS evidence bearing on H3, and contradicts `kernel/02-approach.md` §1. |

| [unreachability-and-escalation](unreachability-and-escalation.md) | **CONVERSATION RECORD, UNRATIFIED.** Failure kind 1 had no actor; the namespace is it. `Think` is the kernel's namespace, which is where its hidden behaviors live. |

| [diffs-and-the-possibility](diffs-and-the-possibility.md) | **CONVERSATION RECORD, UNRATIFIED.** A pairing does not return; a possibility's process does, while the possibility stays open. The diff is owned by the possibility and reached through the span an observation already carries. Deletes `Partition` and dissolves H16's consequence 2. |

Read the first four loose docs in that order — each came out of the one before. The fifth is
independent of that thread and came out of reading [`durable/02-graph.md`](durable/02-graph.md) §4.
All five revise documents in the table above; none has been checked against the code.
**[`coherence.md`](coherence.md) states what they change and what should be cut.**

## Reading order

```text
1  00-principles       what the surface has to satisfy
2  contracts/          the observation and the contract are the load-bearing pair
3  kernel/             the protocol over them
4  identity/           the bounds every call carries
5  durable/            what survives
6  open-questions      what is deliberately not decided here
```

`contracts/` before `kernel/` is deliberate. The seven primitives are readable as a protocol only
once observation provenance and parent-owned contracts are in hand; read in the other order, Decide
and Recognize look like they need operands they do not need.

## What this outline does NOT do

```text
DOES NOT DECIDE     the From representation, the Horizon representation, the observation
                    contract's concrete form, or the authenticate gate. see open-questions.md.

DOES NOT PROPOSE    a parser implementation, a serialization, a storage layout, a transport,
                    or an implementation order.

DOES NOT DESCRIBE   any code that exists. every `go` block here is a shape, not a declaration.
```

The rule from `../../overnightBuild/README.md` applies to a design pass as much as to a build:

> *"A wrong decision here is worse than an unfinished item, because it will be made differently
> later and the code will encode the wrong one."*

Holes are marked **HOLE** and left open. A hole with its constraints written down is a result.

## Three findings this pass produced

Recorded here because each one changes a type, and none was the goal going in.

```text
1  PROVENANCE TYPING DISSOLVES EngramRef, and with it the E1/E2 stop.
   see contracts/01-observation.md §4.

2  THE PARSER CANNOT RUN WHOLLY IN THE KERNEL without breaching NO SEMANTICS.
   the provenance half can; the structural half cannot. see contracts/03-parser.md §3.
   THIS IS THE SHARPEST UNRESOLVED TENSION IN THE SURFACE.

3  A NAMESPACE HAS NO POSITION, so the age rule has nothing to attach to at the outer edge.
   age.md flags it; it lands on the type surface as a missing coordinate. see
   identity/04-position.md §4.
```

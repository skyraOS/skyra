# The primitive list

**Date:** 2026-08-19
**Status: WORKING.** Eight, in four groups. The grouping is the claim — what a primitive
belongs to turns out to say more about it than what it does.

---

## At a glance

```text
PROTOCOL                     what crosses, and what receives
  MOVE                       custody transfer. TWO CAUSAL FORMS — request and response.
  APPROACH                   the port a REQUEST arrives at.
  RECOGNIZE                  the port a RESPONSE arrives at.

DISCOVERY                    what makes a MOVE eligible to fire
  THINK                      READ. maintains the frontier of potential. what is AVAILABLE.
  BOUND                      creates a horizon.                        what is PERMITTED.

LIFECYCLE                    the durable ↔ ephemeral boundary. AN INVERSE PAIR.
  ACTIVATE                   reads an engram, creates a possibility.       LOAD.
  REMEMBER                   creates engrams, tears down possibilities.    STORE.

UNPLACED
  ORIENT                     still needed. mechanism unknown.
```

**The line that keeps THINK and ACTIVATE from overlapping:**

```text
THINK      produces POTENTIAL.   plural. nothing opened, nothing live.
ACTIVATE   produces POSSIBILITY. singular. one engram, in memory, running.
```

---

## PROTOCOL

### MOVE — and why there are exactly two ports

MOVE carries a two-valued header, request or response. The set of things that can arrive
anywhere is therefore exhaustive at two, and so is **the set of receiving ports.**

```text
        REQUEST arrives   →   APPROACH    admits WORK.
        RESPONSE arrives  →   RECOGNIZE   admits a RESULT.
```

Not a convenience of the grouping — forced. A third port would require a third header value,
and the header is capped at two.

**ALL TRANSPORT IS MOVE.** There is no second channel — THINK's own resolution travels the
same way, under the same envelope, subject to the same refusal. A discovery path that bypassed
MOVE would be an unbounded, unrefusable one.

### The two ports are the same kind of thing

Both are inbound admission. Both may refuse. They differ only in what they check:

```text
APPROACH    "MAY YOU SEND ME WORK?"        checks AUTHORITY.
RECOGNIZE   "IS THIS WHAT I ASKED FOR?"    checks CONTRACT, against a frame-resident
                                           expectation at the correlated position.
```

**Whose frame? THE PARENT'S.** The parent owns the return contract, so the parent is the only
party that can check it — which is also WHY the port sits there. One cause, two facts. It
agrees with the independent argument that the party bearing the consequence of a bad match
should be the one performing the check.


### Every decision returns a possibility — the machine is TOTAL

```text
FOR EVERY MOVE, THE SYSTEM IS REQUIRED TO RETURN A POSSIBILITY.
```

Which forecloses something worth foreclosing: **there is no stuck state.** Every state has a
successor, even when the successor is a failure state. A CPU has the same property — there is
always a next instruction, a fault is a TRANSITION to a handler rather than an absence of one,
and even `HLT` is an instruction.

So "nothing came back" is never a legal machine state. It is only ever a legal PAYLOAD.

### APPROACH-as-port closes the bootstrap

**A port is not a call.** It exists before anything arrives and does not need a connection in
order to be arrived at — which is the shape every machine uses to solve this from outside its
own instruction set: a reset vector, a boot address, a root capability handed in.

```text
THE CIRCLE WAS AN ARTEFACT of treating APPROACH as something a possibility INVOKES. It is
something a possibility IS ARRIVED AT THROUGH.
```

Not fully closed — what authority a port grants on first contact, and who provisioned the port
— but no longer circular.

---

## DISCOVERY

### THINK is READ. The execution pointer is WRITE.

```text
THINK               reads and refines the possibility space. Speculative. Nothing it
                    produces has happened.
EXECUTION POINTER   commits. Where work has actually been retired up to.
```

A reader and a writer over one structure — and in hardware this pair has a name. **THINK is
the prefetcher; the pointer is the retire point.** A CPU fetches ahead of where it is
executing, predicts branches, and keeps a bounded window of decoded work ready. Speculative
work that never retires is discarded and costs nothing but the fetch.

That is the same three-part story this design now has end to end:

```text
THINK       speculates, within a bounded window ahead of the pointer
POINTER     retires. this is what actually happened.
REMEMBER    commits only what the outcome DEPENDED ON. everything else goes.
```

### The window is a VARIABLE, and its default is +1

Depth `D` is a tunable, and **the default is 1 — speculation sits exactly one hop ahead of
execution.**

```text
D = 1     THINK resolves the NEXT hop and nothing further.
```

What that buys is the exponent:

```text
D HOPS DEEP     b^D candidates to resolve. exponential in depth.
D = 1           b candidates. LINEAR — only the branches at the immediate decision.
```

Fan-out at the current decision is irreducible: to offer a choice among `b` options, all `b`
must be resolved. What `D = 1` removes is everything beyond that, which was the entire
exponential term.

It also maximises the two things speculation is worst at:

```text
ACCURACY   prediction decays with depth. At one hop it is at its best.
WASTE      a misprediction discards ONE hop of resolution, never a tree of it.
```

### What D actually trades, stated honestly

`D = 1` does not remove latency exposure — **it minimises speculation cost at the price of the
least possible latency tolerance.** The two are inversely related and D is the knob between
them:

```text
DEPTH IS HOW MUCH THINK-LATENCY THE MACHINE CAN HIDE. At D = 1 it hides exactly one
inter-MOVE interval — the time a body spends between one decision and the next. If THINK
takes longer than that, the next MOVE arrives before its candidates do.
```

Which is why it is a variable rather than a constant. The right D is set by one ratio:

```text
                    THINK latency
        D  ≈  ───────────────────────
              inter-MOVE interval

LOCAL, ALREADY INDEXED    THINK is fast. D = 1 is plenty.
REMOTE DISCOVERY,         THINK is seconds. D = 1 stalls on every decision, and depth is the
MODEL CALLS               only thing that hides it.
```

Real hardware does exactly this — stream prefetchers adjust their distance from observed
latency and accuracy rather than shipping a fixed number.

### A stall at D = 1 is not a correctness problem

Worth stating, because it makes the default safe: if a MOVE arrives before THINK has
returned, the machine does not violate totality. It returns the third outcome `move.md`
already defines:

```text
SUCCESS   it crossed.
FAILURE   it did not.
NOT YET   operands unresolved. NOTHING HAPPENED AND NOTHING WENT WRONG.
```

**D cannot break correctness. It can only cost throughput.** That is what makes 1 the right
default and the ratio above the only reason to raise it.

The window slides regardless of size — as the pointer advances, frontier behind it becomes
irrelevant and new frontier ahead must be resolved.

### What folded in, and why it is a unification rather than a bundle

`PERCEIVE` and discovery are inside THINK now. Three sub-operations:

```text
DEPENDENCY WALK     what is reachable through what has already been written.
BOUNDARY INDEX      what exists but is suspended — found by indexing, not by connection.
PERCEIVE            what is live right now.
```

The test that matters, because `REACH` failed it: **do they share an irreducible act, or only
a consumer?**

```text
REACH's THREE   constrain / report / cross. DIFFERENT PRIVILEGE, DIFFERENT PARTIES. A bundle.
                It had to split, and did.
THINK's THREE   same party, same privilege, same output type, same consumer. All three produce
                candidates for the same chooser. A UNIFICATION.
```

The thing to watch is `faults:`. If the three fail in ways one fault list cannot express, they
are still three.

### THINK is FIRED, not polled — and its trigger is MOVE

An earlier draft called THINK a background daemon keeping a frontier current. It is not. **A
prefetcher does not run constantly; it RE-AIMS when a path resolves.**

```text
MOVE retires          the machine has committed to a direction
      │
      ▼
CHECK THE WINDOW      are dependencies solid D hops out?  (D defaults to 1)
      │
      ├── YES  →  fire nothing. A SOLIDIFIED PATH COSTS NOTHING.
      └── NO   →  THINK fires AHEAD from the new position, resolving dependencies and
                  returning potential for the MOVEs that follow.
```

That is branch-redirect prefetch, and it is demand-driven rather than polling.

**THINK is the resolver.** Not merely a discoverer — given an unresolved dependency, it
produces something concrete that satisfies it. Which is the dynamic linker:

```text
CALL an unresolved symbol   → PLT stub → resolver → finds it, PATCHES THE GOT
EVERY SUBSEQUENT CALL       → direct. no stub, no resolver.
```

So the two halves of solidification are one thing after all: **what THINK resolves stays
resolved, and a body whose dependencies are all resolved never fires a THINK at all.**

### Program order and execution order — BOTH, and they are not the same thing

The execution pointer is **a watermark, not a sequencer.** Nothing consults it to decide what
runs; it records how far commitment has swept.

```text
SEQUENCER   the machine LOOKS AT it to choose what to do next. control flow comes FROM it.
            RIP. von Neumann. NOT THIS.
WATERMARK   nothing consults it. Work runs because its operands resolved. The pointer marks
            the boundary between committed and speculative. THE RETIREMENT POINTER IN AN
            OUT-OF-ORDER MACHINE.
```

The test: two MOVEs with no dependency between them. A sequencer visits them in an order and
that order is part of the machine's behaviour. **A watermark has no opinion about which went
first** — both fire when ready, and it sweeps past both.

Real out-of-order machines carry both orders without contradiction, and so does this one:

```text
PROGRAM ORDER      the continuation language says A; B; C.       A SPECIFICATION.
EXECUTION ORDER    the machine runs them in any order the        DATAFLOW. fires on readiness.
                   dependencies allow.
RETIREMENT         the watermark sweeps forward IN ORDER.        restores the illusion.
```

### The sharp edge: sequence is a DEPENDENCY, not a schedule

```text
WRITING `A; B` DOES NOT ORDER THEM. B follows A only if B DEPENDS ON A.
```

This is where the continuation language will differ from every imperative language someone
brings habits from. In C or Python, `;` means *after*. Here it means *next in the text*, and
the machine is free to run both at once if neither reads the other.

Ordering for effect alone therefore has to be written as a dependency — there is no other way
to say it, and this is deliberate: it closes the hole recorded as "ordering is no longer free"
by putting ordering where it can be seen rather than inferred from layout.

### What stays OUT of MOVE because of this

THINK-ahead does not belong in MOVE's `writes:`. On a CPU the branch instruction's semantics
do not include redirecting the front end — the front end REACTS to a resolved branch. Same
here: THINK-ahead is mechanism triggered by MOVE's retirement, below the ISA, not part of the
instruction.

Keeping it out is what preserves MOVE's minimality. Folding it in would make the one transport
primitive quietly contain a discovery engine.

### Why BOUND does not fold in as well

```text
THINK   what is AVAILABLE.  DISCOVERED by an endpoint. unprivileged.
BOUND   what is PERMITTED.  ENFORCED by the fabric. privileged — and the endpoint is the
        party it is enforced AGAINST.
```

Different party performs them, and one exists to constrain the other. That is a principled
floor at two, not an arbitrary stopping point.

### What a POTENTIAL is

```text
SHAPE      A METHOD SIGNATURE. The name carries intent; the types carry shape.
SIZE       capped at 180 tokens.
FIELDS     NO PROSE FIELD.
```

The design already refuses self-report — `Expression` carries no language tag, because a tag
would be a claim about oneself. **A prose field is a self-report with more words.**

Fixed shape is what keeps two expansions comparable. Truncation and summarization destroy the
evidence, so neither is available: a potential either fits the shape or is decomposed until it
does.

**THINK hands up A NAME AND A TOP-LEVEL CONTRACT. Nothing else.** It does not negotiate, and
it reports what a candidate CLAIMS — never that the claim will be satisfied. That is what
keeps RECOGNIZE from being redundant; if THINK certified satisfaction there would be nothing
left to verify on arrival.

### Potentials are CONTRACT-shaped, never VALUE-shaped

Children exist before the parent executes — THINK runs ahead of the pointer, so a potential
cannot reference anything the parent has produced. **Binding happens at ACTIVATE and nowhere
earlier.**

Which upgrades a claim made loosely elsewhere in this document:

```text
WEAK      "nothing THINK produces has happened."     merely unretired.
STRONG    a potential HOLDS NO VALUES AT ALL.        UNBINDABLE.
```

And that is what makes speculation safe in a stronger sense than `D = 1` alone gives.
A misprediction cannot corrupt anything, because **there is nothing stale to carry** — not
discarded state, no state.

### Five slots — a CAP, not a quota

```text
b ≤ 5. THINK EMITS AS MANY AS APPLY, UP TO FIVE. It is never required to produce five.
```

Two shapes this has to serve, and it serves both without either being a defined path:

```text
TWO BRANCHING PATHS, and THINK does not know which     → emit 2.
A LARGE TASK needing five pieces to begin decomposing  → emit 5.
```

**Execution is not obliged to follow all of them.** A slot is an offer, not a commitment.

The cap is not lossy, because the potential tree is **closed under itself** — a potential
decomposes into potentials. Anything not representable in five slots at this level is
represented one level down. So the cap does not truncate the space; **it forces
decomposition**, which is the behaviour wanted anyway.

**The cap is enforced AT THINK CREATION, not at the graph or possibility boundary.** It bounds
one firing of THINK, not how many children a possibility may ever have — THINK fires on every
MOVE that leaves resolved ground, so a possibility accumulates children across many firings.

```text
BOUNDED       one THINK. at most five potentials. THIS IS THE SPECULATION COST BOUND.
NOT BOUNDED   how many children a possibility ends up with. That is the work's shape, and
              nothing here constrains it.
```

Which is what closes the speculation hole: at `D = 1` and `b ≤ 5`, **a decision** costs at most
five resolutions.

### The overflow rule — SEQUENTIAL, AND A HARD CUTOFF

```text
FIVE SLOTS, FILLED IN ORDER OF ARRIVAL. A SIXTH IS NOT INCLUDED.
```

The model is told five. If it produces six, **the last one is dropped** — not scored, not
compared, not chosen against. That is what makes the cap an enumeration bound rather than a
selection, and it is what keeps RANK evicted in fact rather than only on paper.

What a slot therefore does NOT carry:

```text
NO CONFIDENCE     a score is a self-report, and self-reports are refused everywhere else.
NO ORDERING       position in the list means arrival, never quality.
NO IMPLEMENTATION nothing about how it would be done. THAT IS LEAKAGE.
NO PROSE          already established — a prose field is a self-report with more words.

WHAT REMAINS: FIVE CREATE STATEMENTS. CONTRACT HEADERS AND NOTHING ELSE.
```

A hard cutoff is uncomfortable and it is the point: any rule for choosing WHICH to drop would
be ranking, wearing a different name.

### ALTERNATIVES or PARTS — DEFERRED, and it does not need answering yet

The two scenarios differ in a way the slots do not mark:

```text
SCENARIO 1   two candidates, EITHER would do.        alternatives.  OR.
SCENARIO 2   five pieces, ALL are needed.            parts.         AND.
```

**AND/OR gates belong to the continuation language, not to this list.** Nothing here needs
them, because the current rule is smaller and it settles the question after the fact:

```text
IT IS LITERALLY JUST "WHAT DID A CHILD RETURN TO A PARENT."

If the remembered observation has no dependency on a potential, THAT POTENTIAL IS NOT
RECORDED AS AN ENGRAM DURING REMEMBER.
```

So the distinction resolves itself **retroactively, by trace**, and never has to be declared:

```text
ALL FIVE CONTRIBUTED     all five are in the chain and all five are written.  AND, in effect.
ONE CONTRIBUTED          one is in the chain. the other four leave nothing.   OR, in effect.
```

Nothing marks the intent because nothing needs to. A gate would be a declaration ABOUT the
work; the trace is a record OF it, and only one of those can be wrong.

### Containment moves here

Earlier drafts put containment on eager loading of a closure. It belongs to the frontier
instead, and improves: **you can only ACTIVATE what THINK surfaced, and THINK is maintained
under a horizon.** You cannot address what was never on the frontier — the MMU property,
without needing to predict a closure in advance.

---

## LIFECYCLE — ACTIVATE and REMEMBER

```text
ACTIVATE    engram → possibility     durable → ephemeral     LOAD
REMEMBER    possibility → engram     ephemeral → durable     STORE
```

### ACTIVATE is required, not convenient

`state-and-storage.md` records the rule in its strongest form:

> **THE CPU CANNOT EXECUTE AN INSTRUCTION AGAINST PERSISTENT STORAGE.** There is no addressing
> mode for it and no way to phrase the attempt. Data must be brought into memory first.

Engrams are persistent storage. Without ACTIVATE there is **no legal way to run anything at
all** — not a missing convenience, a missing precondition. It makes "a possibility cannot be
opened without representing an engram" true by construction rather than by rule.

`ACT` was renamed because it collides with "action" and reads as *doing*, when the operation is
*loading*.

### What ACTIVATE does

```text
READS      the engram: its contracts, and its continuation if it has one
           the admitted payload: the inputs
           the horizon admitted at the port
WRITES     a frame — CODE AND CONTRACTS LOADED IN, inputs bound at offsets
IT DOES    which engram. THINK surfaced it; something chose it. ACTIVATE loads what it is
NOT CHOOSE handed.
PRIVILEGE  reading persistent storage is device I/O, which is privileged. THIS IS WHERE THE
           "frame allocation is assumed cheap" question lands — raised against DECIDE, and it
           belongs here.
```

A leaf is not a special case: no continuation, so ACTIVATE loads contracts and no code, and
the observation comes from the crossing.

### REMEMBER is available ONLY AT THE ROOT — a position, not a permission

```text
A POSSIBILITY WITH NO PARENT IS THE ROOT, AND ROOT IS WHAT MAKES REMEMBER AVAILABLE.
```

Not a capability anyone grants or checks. The same idiom the design uses for `Leaf` — a
POSITION, not a kind — and for `Branch`, which is a position rather than a type.

It follows from the arena being the root: REMEMBER tears down the whole arena, so a child
performing one would tear down its own parent. And nothing needs to forbid that, because a
non-root **names no arena** and therefore has no operand to hand it. Inexpressible rather than
refused, which is the form this design prefers everywhere it can get it.

### To remember an engram is to remember its CONE

```text
WRITING AN ENGRAM WRITES EVERYTHING IT DEPENDS ON. There is no half of a graph.
```

Forced rather than chosen: a durable engram must be RE-ACTIVATABLE later, and activating it
requires its children. An engram whose children were not written is a dangling reference — a
behaviour that cannot be run.

This is the same statement as the trace rule, read from the other direction:

```text
FROM THE OBSERVATION   write the dependency chain of what was remembered.
FROM THE ENGRAM        write the cone of what was remembered.
                       THE CHAIN IS THE CONE.
```

### REMEMBER is bulk, and traced

```text
TEARS DOWN   EVERY possibility in the arena. Not one frame — all of them.
WRITES       ONLY the engrams in the remembered observation's dependency chain.
```

Which turns an invariant into an arithmetic fact:

> `REMEMBERED ⊆ OCCURRED` stops being a principle anyone maintains. **The subset is
> COMPUTED, by reachability from a root.** No operation could violate it, because writing
> something unreachable would require reaching it.

### A potential becomes an engram only by contributing

New durable behaviour enters the system here and nowhere else. A potential that was depended
on is written; one that was not leaves nothing behind, whether it was never taken, taken and
failed, or taken and ignored.

```text
THERE IS NO "DECIDE TO SAVE THIS." Contribution is the criterion, and it is COMPUTED.
```

**Three levels, one rule.** New TYPES enter the vocabulary the same way: a signature that
survives to be remembered crystallizes into it, and one that does not leaves nothing behind.

```text
POTENTIAL   → ENGRAM     by contributing.
COMPOSITION → CONTRACT   by contributing.
SIGNATURE   → TYPE       by contributing.
```

**Consequence worth stating plainly: the durable record is a DEPENDENCY GRAPH, not a
history.** Work that occurred and did not matter is not retained — which is exactly what
"what the outcome depended on" means, and exactly what an audit log is not.

> **HISTORY IS DEFERRED, NOT DROPPED.** It will be part of this. The goal right now is a small
> working system, and the dependency graph is what that needs. Recorded here so the gap is a
> decision rather than an oversight: as written, "what did this system actually do" is not
> recoverable from the engram graph.

It also dissolves a question rather than answering it: **a failed possibility is not torn down
specially.** It contributed nothing to the dependency chain and it goes when everything goes.
Failure and success have identical teardown, which is why the unwind path never resolved —
that was a question about a machine with per-frame lifetimes.

---

## UNPLACED — ORIENT

```text
CHANGES STATE?        no.  → not an execution primitive.
REQUIRES PRIVILEGE?   yes. → reads what an endpoint cannot see for itself.
CAN A PROGRAM         yes. → nameable, so not below the ISA.
NAME IT?
```

It neither transports, discovers, nor bounds a lifetime — **the only primitive that fits no
group**, and the only one flagged as having no known mechanism. Probably the same fact.

It did not fold into THINK, and the reason is direction: **THINK looks outward at what could
be reached. ORIENT looks inward at your own topology.** Under strict request/response an
endpoint already knows its own outstanding set, which is topology it never had to ask anyone
for — so ORIENT may be purely local, and not a protocol act at all.

---

## What is NOT on this list, and why

```text
PERCEIVE   folded into THINK. It was the "what is live" third of discovery, and separating it
           only made its answer look more durable than it was.
ACT        renamed ACTIVATE. What ALSO lived under that name — creating potential — went to
           THINK, which is why it can be said to have disappeared and survived at once.
RANK       which candidate is best. SEMANTIC, always was entity work. Endpoints resolve their
           own payloads, so ranking is computation inside a CTU — not a primitive.
COMPOSE    arranging several children's contracts into one's own — the case where no single
           child offers what is needed. ENDPOINT WORK, with RANK, below the protocol layer.
           THINK OFFERS · THE PARENT COMPOSES · RECOGNIZE VERIFIES.
DECIDE     a HEADER VALUE. request.
RETURN     a HEADER VALUE. response.
REACH      dissolved into three: BOUND constrains the walk, THINK reports, MOVE crosses. One
           word for three relationships to one axis, which is why it never sat still.
PROJECT    absorbed. REMEMBER writes the durable record; there is no second durable mutator.
```

---

## Open

```text
IS THE POINTER A         RESOLVED — A WATERMARK, AND BOTH ORDERS EXIST. See "Program order
WATERMARK OR A           and execution order" under DISCOVERY.
SEQUENCER?

WHAT IS THE ARENA?       RESOLVED — THE ROOT. One arena per root, many roots per namespace.
                         A namespace scopes RESOLUTION (what names mean); a root scopes
                         LIFETIME (what lives and dies together). REMEMBER tears down one
                         root's possibilities, and concurrent roots in a namespace are
                         possible. Matches Linux, where a namespace is a private view of a
                         name→object mapping and its members' lifetimes determine ITS
                         lifetime rather than the reverse.

ONE ENGRAM OR A CONE?    RESOLVED — A CONE. Writing an engram writes everything it depends
                         on, because a durable engram must be re-activatable and activating it
                         requires its children.

CONTRACTS ACCUMULATE     RESOLVED, AND IT DISSOLVED RATHER THAN BEING DECIDED. The question
                         was whether a written engram keeps contracts for positions that were
                         never filled. It cannot: contracts come FROM composition, and the
                         parent composes only over potentials it actually used. A position
                         never filled was never composed with, so there is no contract to keep.

                             ENGRAMS GROW CONTRACTS ACROSS INVOCATIONS. Re-activate on a path
                             needing a new position and THINK fires, a potential is produced,
                             the parent composes, and the engram gains a contract it did not
                             have.

                         Behaviour that ACCUMULATES rather than being authored complete, as a
                         mechanism rather than a slogan. It retires the older claim that "a
                         contract may exist for a position nothing has occupied" — contracts
                         cannot precede occupation, because composition is what creates them.

                         Contracts and engrams both come into existence by CONTRIBUTING. Same
                         rule, two levels.

WHAT IS A HOP?           RESOLVED. A LEVEL OF DECOMPOSITION in the potential tree — not a
                         connection, not a contract-graph edge. It follows from the tree being
                         closed under itself.

SPECULATION COST —       CLOSED. D defaults to 1, so cost is b rather than b^D. b is capped
CLOSED                   at 5. Potentials are contract-shaped, so a misprediction has nothing
                         to corrupt. And the overflow rule is now fixed — see below.

BRANCH PREDICTION        still absent. D controls how much is speculated, never which. A
                         machine that could bias toward the likely candidate would resolve
                         fewer than b.

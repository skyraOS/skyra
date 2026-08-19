# The CTU pivot and the eight primitives

**Date:** 2026-08-19  
**Scope:** whole project state after `specs/2026-08-18/`, which substantially replaced the
protocol model summarized in `2026-08-18-primitive-behavior-summaries.md` one day earlier.  
**Status:** current picture, including what was retired, what survived, and what is open. The
code in `src/skyra/cognitive-agent/skyra-dna-v6/` predates all of it and is now stale.

> **The kernel was being asked to execute everything. A kernel that executes everything is not
> a kernel — it is an interpreter wearing the name.**
>
> That correction is the whole of what changed, and everything below follows from it.

---

## 1 · What changed, in one picture

```text
BEFORE                                   AFTER
seven protocol meanings                  EIGHT PRIMITIVES in four groups
five caller-facing kernel methods        ONE transport primitive: MOVE
the kernel executes                      TWO UNITS — the CTU executes, the fabric transports
Reach hidden inside Decide               REACH DISSOLVED into three
Decide and Return are methods            they are the TWO HEADER VALUES of MOVE
a possibility is a stack frame           a possibility is an ARENA MEMBER, torn down in bulk
expectations are drafts until closure    contracts CRYSTALLIZE at REMEMBER
```

The prior model was not wrong so much as **conflated**: "the kernel holds everything happening"
was carrying two claims at once — *the kernel is the authority* and *the kernel is the
machine*. Only the first survives. Every attempt to shrink the kernel surface had been stalling
at the same floor, and the floor was that conflation.

---

## 2 · Where the frame came from

Hardware, read for its split rather than its speed.

An ISA is a CPU's vocabulary of instructions — `MOV`, `ADD`, `CMP`, `JMP`, `CALL`, `RET`,
`SYSCALL`. They are all members of the ISA; `MOV`/`ADD`/`JMP` are not a smaller set the others
reduce to. Below the ISA are micro-operations, which no program can name.

Alongside that sits a privilege split, and the load-bearing observation is this:

```text
THE KERNEL DOES NOT INSPECT EVERY INSTRUCTION.
```

It is not interposed. It is **trapped into**. It configures the enforcement machinery once,
exits, and is re-entered only on privilege, fault, or interrupt. The MMU checks every memory
access in hardware and the kernel is consulted on none of them — it only wrote the page tables.

```text
INTERPOSED        every step passes through the kernel, which inspects and forwards.
                  THE MODEL THIS PROJECT HAD BEEN ASSUMING.
CONFIGURE-EXIT    the kernel establishes what is permitted and steps out.
                  THE MODEL HARDWARE ACTUALLY USES.
```

**Solidification already described the second** — primitive mode crosses per step, a compiled
continuation does not cross on the happy path. That had been treated as an optimization. It is
the architecture, and the optimization reading was the mistake. A frame that retrodicts
something you derived independently, and explains it better than your own derivation did, is
telling you something true about the structure.

---

## 3 · Two units, not one

```text
CTU       EXECUTES. one per endpoint. resolves its own payloads, runs its own body,
          performs RECOGNIZE on arrival, decides what to retain.
FABRIC    TRANSPORTS between CTUs. ONE INSTRUCTION — MOVE. reads the envelope, refuses on
          authority, and never looks inside.
```

**MOVE is the act of writing an instruction somewhere. It is not the act of performing one.**
The computation still has to happen, and the CTU is what performs it.

This is not an invention. Tagged-token dataflow machines are built exactly this way: a token
queue and matching store that route, separate processing elements that compute.

---

## 4 · The eight primitives

```text
PROTOCOL                     what crosses, and what receives
  MOVE                       custody transfer. TWO CAUSAL FORMS — request and response.
  APPROACH                   the port a REQUEST arrives at.  checks AUTHORITY.
  RECOGNIZE                  the port a RESPONSE arrives at. checks CONTRACT.

DISCOVERY                    what makes a MOVE eligible to fire
  THINK                      READ. maintains the frontier of potential.  what is AVAILABLE.
  BOUND                      creates a horizon.                          what is PERMITTED.

LIFECYCLE                    the durable ↔ ephemeral boundary. AN INVERSE PAIR.
  ACTIVATE                   reads an engram, creates a possibility.     LOAD.
  REMEMBER                   creates engrams, tears down possibilities.  STORE.

UNPLACED
  ORIENT                     still needed. mechanism unknown.
```

Three tests sort them, and running one test against two columns was what produced spurious
question marks in the earlier model:

```text
DOES IT CHANGE STATE?        sorts EXECUTION membership.
DOES IT REQUIRE PRIVILEGE?   sorts what the fabric must refuse rather than an endpoint perform.
CAN A PROGRAM NAME IT?       sorts what is an instruction at all. Caches fail it. So did REACH.
```

---

## 5 · MOVE, and why there are exactly two ports

```text
state = { target, horizon, payload }

target    ENVELOPE. the fabric reads it — it must, to route.
horizon   ENVELOPE. the fabric reads it — it must, to refuse.
payload   OPAQUE. direction header, correlation, body. ALL ENDPOINT BUSINESS.
```

The header is two-valued, so the set of things that can arrive anywhere is exhaustive at two,
and therefore **so is the set of receiving ports.** A third port would require a third header
value. Not a convenience of the grouping — forced.

**All transport is MOVE.** There is no second channel; THINK's own resolution travels the same
way, under the same envelope, subject to the same refusal. A discovery path that bypassed MOVE
would be an unbounded, unrefusable one.

Running MOVE through the decomposition form produced two things no CPU instruction has:

```text
fires when:    NOT-READY IS NOT A FAULT. A third outcome the success/failure code cannot
               carry: SUCCESS · FAILURE · NOT YET. Collapsing the third into failure would
               make an unresolved dependency indistinguishable from a refused one.
payload is     an operand that appears in `operands:` and in no `reads:`. The only precedent
CARRIED,       is block copy — `REP MOVSB` moves bytes it never interprets.
NEVER READ     THIS IS WHAT MAKES THE END-TO-END PRINCIPLE STRUCTURAL: the fabric cannot read
               the payload because nothing in MOVE is defined over it. Inexpressible beats
               forbidden.
```

**Custody is the load-bearing word.** It is single-holder, which gives blocking a structural
justification rather than a scheduling one: a parent cannot continue because *it no longer
holds the work*. And strict request/response — one request, exactly one response — makes it
countable: **the number of outstanding requests is exactly the number of pieces of work you do
not hold.**

Foreclosed by construction, deliberately: no fire-and-forget, no streaming, no unsolicited
response.

---

## 6 · APPROACH-as-port closed the bootstrap

The circularity was: every MOVE needs a connection from THINK, which needs a horizon from
BOUND, which needs an authority APPROACH establishes.

**A port is not a call.** It exists before anything arrives and does not need a connection in
order to be arrived at — which is the shape every machine uses to solve this from outside its
own instruction set: a reset vector, a boot address, a root capability handed in.

```text
THE CIRCLE WAS AN ARTEFACT of treating APPROACH as something a possibility INVOKES.
It is something a possibility IS ARRIVED AT THROUGH.
```

Not fully closed — what authority a port grants on first contact, and who provisioned the port,
remain open.

---

## 7 · THINK is READ, the execution pointer is WRITE

```text
THINK               reads and refines the possibility space. Speculative.
EXECUTION POINTER   commits. Where work has actually been retired up to.
```

In hardware this pair has a name: **THINK is the prefetcher, the pointer is the retire point.**
Which makes the pipeline consistent end to end:

```text
THINK       speculates, in a bounded window ahead of the pointer
POINTER     retires. this is what actually happened.
REMEMBER    commits only what the outcome DEPENDED ON. everything else goes.
```

Speculative work costing nothing unless it retires and REMEMBER writing only the dependency
chain turn out to be the same fact, not two design choices.

### THINK is fired, not polled

```text
MOVE retires  →  are dependencies solid D hops out?
                     YES → fire nothing. A SOLIDIFIED PATH COSTS NOTHING.
                     NO  → THINK fires AHEAD from the new position.
```

Branch-redirect prefetch, demand-driven. And **THINK is the resolver** — given an unresolved
dependency it produces something concrete, which is the dynamic linker. `CALL` an unresolved
symbol → resolver → patches the GOT → every subsequent call goes direct. So the two halves of
solidification are one mechanism: what THINK resolves stays resolved.

### D is a variable, defaulting to 1

```text
LOWER BOUND   PIPELINE SLACK. at D = 0 every MOVE stalls waiting for a THINK.
UPPER BOUND   FAN-OUT. b^D candidates. exponential.
DEFAULT 1     cost is b rather than b^D. accuracy is best at one hop. a misprediction
              discards one hop, never a tree.
```

**A stall at D = 1 is not a correctness problem.** If a MOVE arrives before THINK returns, the
machine returns `NOT YET`, which is already legal. D cannot break correctness — only throughput.

### What folded in

`PERCEIVE` and discovery are inside THINK: dependency-walk, boundary-index, and what-is-live.
This is the first *consolidation* after a week of decompositions, and the test that catches a
bad one is whether the parts share an irreducible act or merely a consumer. REACH's three had
different privilege and different parties — a bundle, and it split. THINK's three have the same
party, privilege, output type, and consumer — a unification.

`BOUND` does not fold in, and the reason is principled: THINK is what is **available**,
discovered by an endpoint, unprivileged. BOUND is what is **permitted**, enforced by the fabric,
and the endpoint is the party it is enforced against.

---

## 8 · Potentials, slots, and the type language

```text
SHAPE      A METHOD SIGNATURE.  ReturnType name-stating-what-this-will-cause(InputType x, ...)
SIZE       at most 180 tokens.
FIELDS     NO PROSE. NO CONFIDENCE. NO ORDERING. NO IMPLEMENTATION.
```

A prose field is a self-report with more words, and the design refuses self-report everywhere —
which is also why `Expression` carries no language tag. Fixed shape is what keeps two expansions
comparable; truncation and summarization destroy the evidence.

**Potentials are contract-shaped, never value-shaped.** THINK runs ahead of the pointer, so a
potential cannot reference anything the parent produced. Which upgrades a claim made loosely
elsewhere: not merely *unretired* but **unbindable** — and that is why a misprediction cannot
corrupt anything. There is nothing stale to carry, because there is no state at all.

### Five slots — a cap, not a quota

```text
b ≤ 5. EMIT AS MANY AS APPLY. Never required to produce five.
OVERFLOW: SEQUENTIAL, HARD CUTOFF. a sixth is dropped — not scored, not compared.
```

A hard cutoff is uncomfortable and that is the point: any rule for choosing *which* to drop
would be ranking wearing a different name.

The cap is not lossy, because the potential tree is **closed under itself** — a potential
decomposes into potentials, so anything not representable in five slots at this level is
represented one level down. **The cap does not truncate the space; it forces decomposition.**
That also defines the hop: *a hop is a level of decomposition*, not a connection or a graph edge.

Enforcement is **at THINK creation**, not at the graph or possibility boundary. One firing is
bounded at five; how many children a possibility accumulates across many firings is the work's
shape and is not bounded.

### Three authorships

```text
SIGNATURE        authored by THINK.       NEITHER THE CALLER NOR THE CALLEE.
BODY             authored by the CHILD.   it fills a shape it did not choose.
RETURN CONTRACT  NEGOTIATED.              crystallized at REMEMBER.
```

The first is unusual. In normal languages either the callee declares its signature and callers
conform, or the caller declares an interface and callees implement. Here **a third thing
proposes the shape and both sides work against it.** THINK is a broker, not a directory.

The third makes the durable contract **evidence rather than declaration** — which gives
`closure.md`'s *"expectations are drafts until closure"* a mechanism and a moment. It also
means two different things are currently both called the return contract and need separating:
the **expectation** (pre-run, from THINK's signature, what RECOGNIZE checks against) and the
**contract** (post-run, settled, what gets written).

A consequence falls out: if the contract crystallizes *from a match*, then on the recorded path
it matches by construction. **Verification joins the list of things that cost nothing once the
path is real** — alongside a compiled continuation that never faults, a warm GOT that never
resolves, and a resident target that never walks.

---

## 9 · ACTIVATE and REMEMBER

```text
ACTIVATE    engram → possibility     durable → ephemeral     LOAD
REMEMBER    possibility → engram     ephemeral → durable     STORE
```

**ACTIVATE is required, not convenient.** The CPU cannot execute an instruction against
persistent storage — there is no addressing mode for it and no way to phrase the attempt.
Engrams are persistent storage, so without ACTIVATE there is no legal way to run anything at
all. It makes *"a possibility cannot be opened without representing an engram"* true by
construction rather than by rule.

The word is apt in its theatrical sense: **an actor takes on a role from a script.** The engram
is the script — durable, reusable, unchanged by any performance. The possibility is one
performance of it.

**REMEMBER is available only at the root** — a position, not a permission, the same idiom as
`Leaf` and `Branch`. It follows from the arena being the root: a child performing one would tear
down its own parent, and nothing has to forbid that because a non-root **names no arena** and so
has no operand to hand it.

```text
THE ARENA IS THE ROOT.   one arena per root, many roots per namespace.
                         NAMESPACE scopes RESOLUTION — what names mean.
                         ROOT scopes LIFETIME — what lives and dies together.
```

That matches Linux, where a namespace is a private view of a name→object mapping and its
members' lifetimes determine *its* lifetime rather than the reverse.

**REMEMBER is bulk and traced.** It tears down every possibility in the arena and writes only
the cone of the remembered observation's dependencies. Which turns an invariant into an
arithmetic fact:

> `REMEMBERED ⊆ OCCURRED` stops being a principle anyone maintains. **The subset is COMPUTED,
> by reachability from a root.** No operation could violate it, because writing something
> unreachable would require reaching it.

It also dissolves a question rather than answering it. **A failed possibility is not torn down
specially** — it contributed nothing and goes when everything goes. Failure and success have
identical teardown, which is why the unwind path never resolved: that was a question about a
machine with per-frame lifetimes.

---

## 10 · The contribution rule

New things enter the durable layer one way only:

```text
POTENTIAL   → ENGRAM     by contributing.
COMPOSITION → CONTRACT   by contributing.
SIGNATURE   → TYPE       by contributing.
```

There is no "decide to save this." Contribution is the criterion and it is computed.

**Engrams therefore grow contracts across invocations.** A contract cannot precede occupation,
because composition is what creates them — which retires the older claim that *"a contract may
exist for a position nothing has occupied."* Re-activate on a path needing a new position and
THINK fires, a potential is produced, the parent composes, and the engram gains a contract it
did not have. Behaviour that accumulates rather than being authored complete, as a mechanism
rather than a slogan.

**Consequence to be deliberate about: the durable record is a DEPENDENCY GRAPH, not a history.**
Work that occurred and did not matter is not retained. History is deferred, not dropped — but as
written, *"what did this system actually do"* is not recoverable from the engram graph.

---

## 11 · Both orders

The execution pointer is **a watermark, not a sequencer.** Nothing consults it to decide what
runs; it records how far commitment has swept. The test is two MOVEs with no dependency between
them: a sequencer visits them in an order and that order is part of the machine's behaviour; a
watermark has no opinion about which went first.

```text
PROGRAM ORDER      the continuation language says A; B; C.    A SPECIFICATION.
EXECUTION ORDER    any order the dependencies allow.          DATAFLOW.
RETIREMENT         the watermark sweeps forward in order.     restores the illusion.
```

**The sharp edge: writing `A; B` does not order them.** B follows A only if B depends on A.
Ordering for effect alone must be written as a dependency — deliberately, because that puts
ordering where it can be seen rather than inferred from layout.

This closed four items at once: the pointer question, "ordering is no longer free," concurrency
needing a group form, and the instruction-position problem — the last of which never needed
solving, because a watermark has no position to save.

---

## 12 · What survived from the earlier design

```text
ENGRAM / POSSIBILITY         durable behaviour vs one performance of it. Untouched, and
                             ACTIVATE made it structural.
P1 — NOTHING FROM NOWHERE    now enforced at generation: input types only from observations in
                             scope, and no immediate operand form.
NO SELF-REPORT               no language tag, no prose field, no confidence score.
LEAF AND THE FORFEIT         a position, not a kind. still governs origination.
THE PARENT OWNS THE          RECOGNIZE checks against the parent's frame, because the parent
RETURN CONTRACT              owns the contract. One cause, two facts.
SOLIDIFICATION               promoted from optimization to architecture.
REMEMBERED ⊆ OCCURRED        promoted from principle to computed fact.
```

## 13 · What was retired

```text
THE FIVE-METHOD KERNEL       Approach/Think/Decide/Remember/Orient as a caller-facing surface.
DECIDE, RETURN               as opcodes. They are header values.
REACH                        as a primitive. Split three ways — BOUND constrains the walk,
                             THINK reports, MOVE crosses. One word for three relationships to
                             one axis, which is why it never sat still.
PERCEIVE                     folded into THINK.
PROJECT                      absorbed into REMEMBER.
RANK                         endpoint computation, not a primitive.
REALIZATION                  the three-way Reach conclusion type.
PER-FRAME TEARDOWN           replaced by bulk arena collection.
AMBIENT AUTHORITY            replaced by capability passing — horizon rides in the envelope,
                             because ambient authority does not survive leaving one box.
```

---

## 14 · Code status

`src/skyra/cognitive-agent/skyra-dna-v6/` **predates all of this.** It holds the pre-pivot
model: a five-method `Kernel` interface with `Decide` returning a `Realization`, a
`Possibility` with `ChildBranches` and `Leaf`, `Engram` with `Contracts`/`Continuation`, and
`Compiler`/`Creator` registries. It builds and vets clean and is now a snapshot of a superseded
design.

Nothing in `specs/2026-08-18/` has been implemented. The next build is not a revision of v6.

---

## 15 · Open, ranked

```text
1  CONTRACTS                 the named next thread. RECOGNIZE checks against the parent's,
                             ACTIVATE loads them into the frame, a potential IS a signature,
                             and ordering is now expressed as dependency. Four things lean on
                             one undefined object.

2  WHAT "NEGOTIATED" MEANS   given strict request/response with exactly one reply. Within a
                             call it is settlement; across calls it is accumulation. If both,
                             REMEMBER is merging rather than writing, and merge rules are
                             needed.

3  CONTAINMENT               a CPU may let user code run unaccompanied because the MMU
                             CONTAINS it, physically. Endpoints here are addresses, possibly
                             python, on machines nobody controls. Capability passing is the
                             right model; it is not yet a mechanism. THIS HOLE HAS NOW APPEARED
                             INDEPENDENTLY FOUR TIMES.

4  BRANCH PREDICTION         D controls how much is speculated, never which. Speculation is
                             cheap on a CPU and expensive here — a mispredicted THINK may cost
                             round-trips or model calls. THE FIRST PLACE THE CPU FRAME MISLED
                             RATHER THAN CLARIFIED.

5  ORIENT                    privileged, nameable, and a member of no group. Live candidate:
                             under strict request/response an endpoint already knows its own
                             outstanding set, so ORIENT may be purely local and not a protocol
                             act at all.

6  CONVERGENCE               what makes the type vocabulary shared across frames, in the
                             machine rather than as prompt pressure. Stated to exist; not
                             written down.

7  THE OVERFLOW RESIDUAL     the cap is an enumeration bound, but nothing yet ENFORCES
                             decomposing an overflowing space rather than sampling it.

8  bound.md · think.md       neither decomposed yet.
```

---

## Source map

```text
specs/2026-08-18/
  CPU-Research/
    cpu-structure.md                 four layers; architectural state; what is not uniformly
                                     writable and why RIP's rule matters.
    kernel-entry-point.md            the IDT; EVENT vs REQUEST entry; the two mechanisms.
    state-and-storage.md             registers · memory · persistent; the rule that nothing
                                     executes against storage.
    cpu-to-ctu-correspondence.md     the correspondence table and the bucketing — one register
                                     file, the possibility as save area and as memory.
  primitive-decomposition/causal-transition-unit/
    README.md                        the frame, the form, two units, where the original eight
                                     went.
    primitives.md                    THE WORKING DOCUMENT. eight primitives, four groups.
    move.md                          the only transport primitive, run through the form.
  sysprompts/
    think.md                         THINK's system prompt, revised, with what it commits to
                                     that the primitive list did not have.
```

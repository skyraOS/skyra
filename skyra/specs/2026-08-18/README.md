# 2026-08-18 — the CTU pivot

**Status: THE CURRENT DESIGN SET.** Start here. This directory supersedes the protocol model in
every earlier spec directory.

> **The kernel was being asked to execute everything. A kernel that executes everything is not a
> kernel — it is an interpreter wearing the name.**
>
> That correction is the whole of what changed. Everything in this directory follows from it.

---

## Read in this order

```text
1  README.md                 (this file) — what is live, what is dead, and the house rules.
2  the-quadrant.md           THE FOUNDATION. system/user × privilege, the four cells, and the one
                             firm rule (no diagonal). the frame the boundary docs instantiate.
3  primitive-decomposition/causal-transition-unit/README.md
                             the frame, the decomposition form, and the two units.
4  primitive-decomposition/causal-transition-unit/primitive.md
                             THE ONE-LINE INDEX. primitive · owner · what it does. If you read
                             only one file, read this one.
5  primitive-decomposition/causal-transition-unit/primitive-details.md
                             THE WORKING DOCUMENT. the same eight in four groups, with the
                             reasoning, the model talk, and the arguments.
6  continuation-language.md  what a body actually says, and what each surface form desugars to.
7  network/traffic-routing.md one APPROACH egress; local and external ingress; normalization before
                             GREET / RECOGNIZE; host-kernel routing.
8  kernel-ctu-boundary/horizon-and-root.md
                             the token = horizon; the pin, account, and root — and the token as a
                             SIGNATURE CHAIN, not an unsigned {root, caveats}. (foregrounded here
                             after a build stubbed it unsigned for being off the path.)
9  carry-forward-from-08-12.md
                             which of the 104 older documents still hold.
```

The full narrative picture — including what was retired and why — is
`skyra/summaries/2026-08-19-the-ctu-pivot-and-the-eight-primitives.md`.

---

## What is in here

```text
the-quadrant.md              WORKING, FOUNDATIONAL. system/user × privilege; the four cells; the
                             no-diagonal rule. the hub the boundary folders instantiate.

CPU-Research/                RESEARCH. hardware read for its split rather than its speed. This
                             is where the pivot came from, and it is source material, not design.
  cpu-structure.md           four layers; architectural state; why RIP is not assignable.
  kernel-entry-point.md      the IDT; EVENT vs REQUEST entry.
  state-and-storage.md       registers · memory · persistent, and the rule that nothing executes
                             against storage.
  cpu-to-ctu-correspondence.md
                             the correspondence table, and the bucketing that makes it work.

primitive-decomposition/causal-transition-unit/
  README.md                  frame · form · two units · where the original eight primitives went.
  primitive.md               WORKING. the one-line index — primitive · owner · what it does.
  primitive-details.md       WORKING. the live document. everything settles here first.
  move.md                    PROPOSAL. the only transport primitive, run through the form.

sysprompts/
  think.md                   RESEARCH. THINK's system prompt, with what it commits to that the
                             primitive list does not.

continuation-language.md     WORKING PROPOSAL. supersedes the 08-12 language document.
network/
  traffic-routing.md         WORKING. one logical egress, two ingress paths, custody milestones.

kernel-ctu-boundary/         WORKING. the PRIVILEGE line — the two crossings.
  boundaries.md              the contract; load/store; BOUND's DEFINE/ATTENUATE split.
  durable-storage.md         the two crossings as the shared driver abstraction.
  horizon-and-root.md        token = horizon; the pin, account, and root.
  implementation.md          content-addressed resolution — one way to fill the boundary.
entity-ctu-boundary/         WORKING. the TRUST line — CTU ↔ entity execution.
  decode-and-run.md          the execution interior: decode → schedule → run (the out-of-order
                             pipeline). scheduler owns where/when; local-vs-distributed is open.

carry-forward-from-08-12.md  ASSESSMENT. 104 documents judged against this model.
```

---

## Two mistakes available to a fresh reader

### 1 · `specs/2026-08-12/` is mostly superseded

It describes **five caller-facing kernel methods** and **seven primitives**. Neither survives.
Do not reason from it without checking `carry-forward-from-08-12.md` first, which sorts every
document into stands · stands-with-edits · superseded.

The short version of the pattern:

```text
DIED    every document whose subject was A SURFACE — which methods exist, what they return.
LIVED   every document whose subject was A CONSTRAINT — what may not be forged, what may not
        appear from nowhere, what the substrate can prove, what the kernel may not read.
```

### 2 · `src/skyra/cognitive-agent/skyra-dna-v6/` is stale

It builds and vets clean and is a snapshot of the **pre-pivot** model: a five-method `Kernel`
with `Decide` returning a `Realization`, `Possibility` with `ChildBranches`, `Compiler` and
`Creator` registries.

**Nothing in this directory is implemented. The next build is not a revision of v6.**

---

## The design in twelve lines

```text
TWO UNITS         the CTU executes, one per endpoint. the FABRIC transports between them.
ONE TRANSPORT     MOVE. custody transfer. { target, horizon, payload }.
                  target and horizon are ENVELOPE; payload is OPAQUE and never read.
ONE EGRESS        APPROACH socket. every local CTU submits a resolved connection here; the host routes.
TWO INGRESSES     local and external. separate trust paths, normalized to one arrival.
TWO HEADERS       request and response. Which is why there are exactly two receiving ports.
EIGHT PRIMITIVES  MOVE · GREET · RECOGNIZE | THINK · BOUND | ACTIVATE · REMEMBER | ORIENT
DATAFLOW          a MOVE fires when its operands resolve, not when a pointer reaches it.
                  THE POINTER IS A WATERMARK, NOT A SEQUENCER.
ORDER             sequence is a DEPENDENCY, not a schedule. `A; B` does not order them.
SPECULATION       THINK runs D hops ahead (default 1), emitting AT MOST five potentials.
                  Potentials are CONTRACT-SHAPED, so a misprediction cannot corrupt anything.
THE ARENA         is the ROOT. REMEMBER is root-only, bulk, and traced.
CONTRIBUTION      potentials become engrams, compositions become contracts, and signatures
                  become types — ALL BY CONTRIBUTING, and only through REMEMBER.
```

---

## House rules for adding to this directory

These are how the set was written and why it is readable. Keep them.

```text
STATUS ON EVERY DOC       RESEARCH · MEASURED · WORKING · PROPOSAL · ASSESSMENT · DECIDED.
                          A reader must know whether they are holding evidence or a choice.

RECORD CONTRADICTIONS,    when a document disagrees with another, NAME THE OTHER DOCUMENT. Two
NEVER SMOOTH THEM         files quietly disagreeing is the failure mode this set is written
                          against.

SAY WHAT IS OPEN          every document ends with what it did not settle. An unstated hole
                          gets rediscovered; a stated one gets closed.

DERIVE, DO NOT DECLARE    privilege comes from what an operation WRITES. Classification is a
                          CONSEQUENCE of a decomposition, never an input to it.

PREFER INEXPRESSIBLE      a rule nobody can break beats a rule nobody may break. RIP is not
TO FORBIDDEN              assignable; the fabric cannot read a payload; a non-root names no
                          arena. None of those need enforcement.

ARGUE THE OTHER SIDE      a classification that survives being argued against is worth
                          something. One never tested is a guess with a box drawn round it.
```

---

## The four things most likely to change next

```text
CONTRACTS               the named next thread. RECOGNIZE checks against the parent's, ACTIVATE
                        loads them into the frame, a potential IS a signature. Four things lean
                        on one undefined object.
ESCALATION              the 08-12 model was "progressive causal zoom" — a failed return wakes
                        the parent. Bulk teardown removed the mechanism and nothing replaced it.
                        THE LARGEST GAP IN THE SET. First lead from the mechanism side:
                        `../2026-08-22/ctu/orient.md` — a stalled subtree shows as a hole in the
                        propagated ORIENT; the ack-up-the-tree is where "wake the parent" could hang.
CONTAINMENT             capability passing is the right model; it is not yet a mechanism. This
                        hole has now appeared independently four times.
BRANCH PREDICTION       D controls how much is speculated, never which. Speculation is cheap on
                        a CPU and expensive here.
```

**And hold `move.md` and `primitive-details.md` loosely.** They are surface documents, and
`carry-forward-from-08-12.md` records what happened to the last set of those.

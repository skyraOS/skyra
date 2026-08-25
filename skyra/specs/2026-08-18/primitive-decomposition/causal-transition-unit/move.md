# MOVE

**Date:** 2026-08-19
**Status: PROPOSAL.** MOVE is the primitive. `decide.md` was deleted rather than kept —
DECIDE is a header value now, not an opcode, and leaving a superseded decomposition beside
this one would give two answers to the same question.

---

## What it is

> **MOVE is custody transfer. It is the only transition, and it carries a payload it never
> reads.**

Two families, both real machines:

```text
PACKET SWITCHING   request and response are ONE primitive with different headers. there is
                   no "call" and no "return" on a fabric — there are only packets.
DATAFLOW           an operation fires when its operands are resolved, NOT when a program
                   counter arrives at it. there is no program counter.
```

The second is why the instruction-position question disappears rather than getting answered.
It failed three times — RIP vs RSP, RETURN's resumption point, an in-flight DECIDE needing to
be resumable — because it was a question about a machine this is not.

---

## The decomposition

```text
MOVE

operands:
    target
        allowed forms:
            live connection      from PERCEIVE. ENVELOPE — the fabric reads it, because it
                                 must, in order to route.
                                 NO IMMEDIATE. a target from nowhere has no provenance.

    horizon
        allowed forms:
            capability           from BOUND. ENVELOPE — the fabric reads it, because it must,
                                 in order to refuse.

    payload
        allowed forms:
            opaque               CARRIED, NEVER READ. contains the direction header
                                 (request | response), correlation, and the body. ALL OF IT
                                 IS ENDPOINT BUSINESS.

    sys-instruction
        allowed forms:
            optional constraint  ENVELOPE — read by the RECEIVING CTU, never the fabric and never
                                 the entity. ABSENT BY DEFAULT. when present it NARROWS the eligible
                                 set (placement), ⊆ horizon — it cannot pin beyond authority.

fires when:
    target      resolved         PERCEIVE has produced a live connection
    horizon     resolved         BOUND has produced a capability
    payload     present          the body produced one

    ALL THREE, AND NOT BEFORE. NOT-READY IS NOT A FAULT — it is ineligibility, and it is a
    third outcome the S/F code cannot express.

reads:
    value(target)                DEREFERENCE. can fault — the connection may no longer be live.
    horizon                      compared against current authority. must narrow.
    ── payload IS NOT READ ──    it appears in operands and in no read. see §2.

transition:
    custody(payload) : here → target

writes:
    custody(payload)             RELEASED. the sender no longer holds it.
    outstanding                  +1 on request, −1 on response. ENDPOINT-LOCAL BOOKKEEPING.
    flags                        S/F — did custody cross.

faults:
    connection not live          PERCEIVE said live and it is not. THE TOCTOU WINDOW, and
                                 closing it is this fault's whole job.
    horizon refuses target       the fabric declines. envelope-level, no payload inspected.
    delivery failed              CUSTODY DID NOT CROSS — the sender still holds the work and
                                 may retry, repair, or choose again. RECOVERABLE IN PLACE.
    unsolicited response         a response with no outstanding request. protocol violation.
```

---

## Two instructions, one per space

A MOVE carries two instructions — one to each side of the trust line — under one authority:

```text
USER-INSTRUCTION   target + payload.   instructs USER space: which entity, and what to run.
SYS-INSTRUCTION    sys-instruction.    instructs SYSTEM space: where/how to run in the system block.
HORIZON            the authority that bounds BOTH. neither instruction may exceed it.
```

The two spaces each keep ONE private field — a clean mirror across the trust line:

```text
payload           read ONLY by the ENTITY.  user-space content. OPAQUE to fabric and CTU.
sys-instruction   read ONLY by the CTU.      system-space instruction. the entity never sees it.
```

The opacity is not incidental. The payload is opaque BECAUSE user space is untrusted and
proprietary; the sys-instruction is transparent to the CTU BECAUSE system space is the machine's
own business. **The envelope/payload split IS the system/user split, at the field level.**

---

## What running the form exposed

### 1. `fires when:` is a section the CPU form does not have

On a von Neumann machine an instruction executes when reached; there is no waiting and no
eligibility. Here there is, and it produces a **third outcome** that the S/F code cannot
carry:

```text
SUCCESS      custody crossed.
FAILURE      custody did not cross.
NOT YET      the operands are not resolved. NOTHING HAPPENED, AND NOTHING WENT WRONG.
```

Collapsing the third into failure would make an unresolved dependency indistinguishable from a
refused one — a distinction the whole design has insisted on elsewhere.

### 2. An operand that is never read

`payload` appears under `operands:` and appears in no `reads:` entry. **No CPU instruction
does this.** Every operand a CPU names is read or written, because a CPU computes over its
operands.

There is one precedent, and it is the right one: **block copy.** `REP MOVSB` moves bytes it
never interprets. That is transport, and transport is what MOVE does.

This is the end-to-end principle expressed as a hole in the form:

> The kernel having no policy and reading no content stops being a RULE anyone maintains.
> **MOVE cannot read the payload because nothing in MOVE is defined over it.** Inexpressible
> beats forbidden.

### 3. Privilege derives, and splits by party

```text
THE SENDER   writes only endpoint-local state: its custody ledger and a flag.
             ISSUING A MOVE IS UNPRIVILEGED.
THE FABRIC   reads horizon and refuses. THE REFUSAL IS THE PRIVILEGED ACT, and it happens
             without any payload being inspected.
```

Same pattern the CPU work produced: unprivileged at the top, privileged at the crossing.

### 4. Custody becomes countable

Strict request/response — one request, exactly one response — means:

> **The number of outstanding requests is exactly the number of pieces of work you do not
> hold.**

No traversal, no reconciliation. It also forecloses three things, by construction and
deliberately: **no fire-and-forget, no streaming, no unsolicited response.**

---

## Where the other primitives went

```text
DECIDE       → the REQUEST header value.
RETURN       → the RESPONSE header value.
REACH        → split. PERCEIVE resolves the target operand; the fabric does the routing.
PERCEIVE     → resolves `target`. no longer a peer of MOVE — it is what makes one eligible.
BOUND        → resolves `horizon`. same.
RECOGNIZE    → AT THE RECEIVER, on arrival, against a frame-resident contract. the party that
               bears the consequence of a bad match performs the check, which is
               self-enforcing in a way boundary-side checking never was.
REMEMBER     → an endpoint act on receipt. payload business.
THINK        → PERCEIVE (substrate: what is live) + RANK (semantic: which is best, and always
               was entity work).
ORIENT       → an endpoint reading its own outstanding set.
GREET     → UNPLACED. see below.
```

---

## Open

```text
GREET CANNOT BE       every MOVE requires a live connection from PERCEIVE, and PERCEIVE
A MOVE                   requires a Horizon from BOUND, and BOUND requires an authority that
                         Greet is what establishes. THE BOOTSTRAP IS CIRCULAR and this
                         does not close it.

RESOLVED — TWO UNITS     an earlier draft of this item said the CTU had become a fabric and
                         needed renaming. It had not. MOVE IS THE ACT OF WRITING AN
                         INSTRUCTION SOMEWHERE, NOT OF PERFORMING ONE — the computation still
                         has to happen, and the CTU is what performs it, one per endpoint.
                         The fabric carries; the CTU executes. See README.

ORDERING IS NO LONGER    dataflow fires on dependency alone, so independent MOVEs may fire in
FREE                     any order. That is what makes concurrency native — and the design
                         states elsewhere that EMITTED ORDER IS SEMANTIC. Ordering is still
                         expressible by threading a dependency, but it is now explicit and
                         costly rather than automatic.

ALL CONTENT VALIDATION   the fabric rejects on target and horizon, never on payload. A
IS LATE                  malformed payload is discovered at the endpoint, after transport was
                         spent. The standard price of end-to-end; the design's fail-fast
                         instincts now apply to the envelope only.

WHAT CORRELATES A        response matching is endpoint business, so a correlation lives in the
RESPONSE                 payload. Whether it is an ID, a position, or the connection itself is
                         unwritten — and the connection-derived version is the unforgeable one.
```

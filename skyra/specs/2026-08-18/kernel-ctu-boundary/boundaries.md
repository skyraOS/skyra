# The kernel/CTU boundary

**Date:** 2026-08-20
**Status: WORKING.** The boundaries settled in conversation on 2026-08-20, recorded before they
drift. This document is about WHERE the lines are, not how they are built — the mechanism I would
put underneath them is `implementation.md`, kept separate on purpose. Everything here either
agrees with the 08-18 set or is named as disagreeing with it.

> **The kernel and the CTU meet at exactly two crossings. That is the entire contract between
> them, and stating it as exactly two tells you the kernel's whole surface.**

---

## Three parties, and the line this document draws

```text
KERNEL    owns the DURABLE GRAPH and the root/lifetime structure. name → meaning.
CTU       owns the LIVE possibility tree. executes MOVEs within it. one per endpoint.
FABRIC    transports between CTUs. one instruction — MOVE. reads the envelope, refuses, never
          looks inside.
```

The 08-18 set already draws CTU↔FABRIC. **The line this document draws is KERNEL↔CTU**, and its
whole claim is that the line is crossed in exactly two places.

---

## The contract is exactly two crossings

```text
ACTIVATE   LOAD    CTU traps → "give me this engram as a possibility."
                   kernel reads durable storage, hands back a possibility.
                   the CTU now owns it and runs it. KERNEL STEPS OUT.

REMEMBER   STORE   CTU traps AT ROOT → "commit the cone of this contribution."
                   kernel writes durable storage; the arena tears down. KERNEL STEPS OUT.
```

Between those two traps the kernel is not consulted. The CTU builds MOVEs, routes them, resolves
payloads, and walks up and down the possibility tree with no kernel involvement.

### Why exactly two — this is a load/store architecture

```text
DURABLE GRAPH        the slow durable tier.     MEMORY.
POSSIBILITY TREE     the fast live tier.        REGISTERS.
ACTIVATE / REMEMBER  the only tier-crossing.    LOAD / STORE.
```

A load/store (RISC) ISA is *defined* by the rule that only two instruction classes touch memory
and everything else is register-to-register. "Exactly two crossings" is the same statement as
"the kernel/CTU boundary is load/store." That is why it is two and not three — not a count that
happened to come out small, but the defining property of a tractable machine. It is also what
lets the contract be stated as a surface: two instructions, and the kernel needs no other door.

---

## The non-crossings — what looks like it should cross and does not

The contract only holds if the following are NOT crossings. Three of them are clean. The fourth
is a seam.

```text
FABRIC REFUSAL   the fabric reads `horizon` and refuses on every MOVE. This is the MMU checking
                 every access in hardware — CONFIGURED at a crossing, CONSULTED without one.
                 NOT A TRAP. Most MOVEs never enter kernel mode.

THINK (propose)  proposes names at the frontier. unprivileged, endpoint-run, read-only against a
                 cached view of the namespace. effects nothing. NOT A CROSSING.

RECOGNIZE        checks a return against a FRAME-RESIDENT contract. A register compare over live
                 state. writes only a causal condition. NOT A CROSSING.

BOUND            — SEAM 2. the two-crossing contract REQUIRES this to be a non-crossing; the
                 written 08-18 set files it as a syscall. Unresolved here; see the seams.
```

---

## THINK is the namespace; resolution is the privileged half of the two crossings

```text
THINK IS A NAMESPACE   the kernel owns what the tree's proposed names MEAN. resolution scope.
                       this is the 08-18 arena answer — "a namespace scopes RESOLUTION" —
                       named to its owner.

THINK THE ACT          proposes names. unprivileged, endpoint, effectless. one door up from
                       "nothing from nowhere": it may only propose names that CAN resolve.

RESOLUTION             binding a name to a meaning. PRIVILEGED — and it is the privileged HALF of
                       the two crossings already listed, not a third crossing:
                         ACTIVATE   resolve name → EXISTING meaning, then load.  READ-resolution.
                         REMEMBER   install a surviving name → NEW meaning.      WRITE-resolution.

CONVERGENCE            the type vocabulary is shared because ALL binding goes through ONE
                       kernel-owned namespace — not because a prompt preferred recurring types.
                       This closes the convergence hole `sysprompts/think.md` left open.
```

The correction to the 08-18 set: `primitive-details.md` says flatly "THINK is the resolver." That fuses
two acts. **Search is endpoint and free; binding is kernel and owned.** The mess is the fusion;
this document splits it.

---

## Nothing from nowhere, at two tiers

```text
VALUE TIER   reference BOUND values,      produce a binding,   lands at RECOGNIZE.   live.
NAME TIER    reference RESOLVABLE types,  produce a name,      lands at REMEMBER.    durable.
```

Same discipline one door up, and the landing tier matches the thing's lifetime: an ephemeral
value lands live, a durable name lands durable. This is why type-birth is at REMEMBER and
value-birth is at RECOGNIZE — not a choice, a consequence of which tier each thing lives in.

---

## The spawn boundary — what must be stated to hand work across the line

The two boundaries this conversation cared about are the two arguments to a spawn:

```text
HORIZON    what it may touch       capability, narrowing-only.        ENVELOPE.
CONTRACT   what it must return     signature, checked at RECOGNIZE.   FRAME-RESIDENT.
CUSTODY    who is blocked on whom  parent link + exactly one exit.    MOVE's S/F + `outstanding`.
```

Under the process reading below, these are precisely what you state to `clone` a process across
a trust line: its capability set, its return ABI, and its parent/exit channel.

---

## Where MOVE lands — one process, two views

Recorded because it is the placement we agreed and it grounds the spawn boundary:

```text
A MOVE IS NOT A PROCESS. A correlated REQUEST/RESPONSE PAIR is the causal shadow of one — the
future the fabric holds over work it cannot see. THE PROCESS is the possibility between them.

  request MOVE    spawn / clone            the causal birth of work across the line
  possibility     the process (task_struct) the running body, address space, save area
  response MOVE   exit(status)             exactly one status home
  outstanding     the un-reaped children   the wait-set

ONE process, TWO views — possibility is the inside (endpoint), the MOVE-pair is the outside
(fabric). Not two processes.
```

---

## The seams

Each is a place where two boundaries stated above collide. None is a worry about a mechanism —
each is forced by the lines themselves.

```text
SEAM 1 — NO PARTIAL SPILL
  BOUNDARIES    load/store (spill a register to memory to free it)  ×  REMEMBER is root-only and
                BULK (the only STORE is the terminal one).
  COLLISION     a load/store machine spills individual registers; this one cannot spill a subtree
                to durable storage before the root commits. The entire live possibility tree must
                fit for the whole arena's lifetime.
  OPEN          is an arena bounded by construction (fine, and "no spill" is then a feature), or
                can it outgrow live capacity (then it needs a partial STORE — which would be a
                THIRD crossing and break the contract)? The load-bearing assumption under
                "exactly two" is: an arena always fits live.

SEAM 2 — BOUND
  BOUNDARIES    "exactly two crossings" (BOUND must not cross)  ×  the written 08-18 set
                (`causal-transition-unit/README.md`: BOUND is "still [a] syscall, because the map
                lives where the process cannot write it").
  COLLISION     direct contradiction. If narrowing the horizon edits kernel-held state, every
                BOUND traps and the contract is three crossings, not two.
  RESOLVED ONLY IF   horizons are attenuable tokens the endpoint can narrow locally and cannot
                widen by construction — then narrowing needs no kernel and enforcement is the
                fabric refusal (a non-crossing). This is doc-vs-doc: the 08-18 pivot to capability
                passing already picked that branch; the "still a syscall" line is a seccomp
                leftover that the pivot overturns. Until the token commitment is written down,
                the boundary is contradictory. See `implementation.md`.

SEAM 3 — GLOBAL IDENTITY MEETS LOCAL SATISFACTION
  BOUNDARIES    the namespace makes NAMES converge globally  ×  RECOGNIZE decides SATISFACTION
                locally, in the parent's frame, parent-relative.
  COLLISION     identity is global; satisfaction is local. It is therefore possible for an
                observation whose type IS `Destination` (globally the same name) to fail to
                satisfy `Destination` at a given position. When that happens, name-identity and
                contract-satisfaction have come apart, and the boundary between "same type" and
                "acceptable here" is unstated.
  OPEN          does global name-identity constrain local satisfaction at all, or are they fully
                independent axes? If independent, "type" is doing two unrelated jobs under one
                word.

SEAM 4 — DURABLE-ONLY BINDING MEETS INTRA-ARENA RESOLUTION
  BOUNDARIES    new names install ONLY at REMEMBER (durable, terminal, at root)  ×  ACTIVATE
                resolves names THROUGHOUT the arena, long before any REMEMBER.
  COLLISION     a name proposed inside this arena has no durable meaning until the arena commits.
                So within one arena, whether a sibling can resolve a name a sibling just proposed
                is undefined — resolution is specified against the kernel-owned durable map, and
                the new name is not in it yet.
  OPEN          is there a live, arena-scoped resolution scope beneath the durable namespace
                (convergence only across arenas), or can proposals resolve intra-arena before
                commit (then something holds provisional bindings the two-crossing contract does
                not name)?
```

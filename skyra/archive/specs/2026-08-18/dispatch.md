# Dispatch — two tables, and the kernel is named

**Date:** 2026-08-21
**Status: WORKING.** How the CTU routes a MOVE to a target. This realizes `the-quadrant.md`'s
TRUST axis concretely. The privilege realization (the kernel is *named*, not *flagged*) and the
CPU-not-process framing are settled; where the kernel entry physically lives is open.

_Instantiates `the-quadrant.md` — the concrete routing behind the four cells._

> **The CTU is a CPU routing instructions to hardware layers. It keeps two lookup tables — trusted
> and untrusted hardware — and privilege is NOT a flag on the instruction. It is whether the
> instruction NAMES THE KERNEL.**

---

## First, what this is NOT — CPU → hardware, not OS → processes

The easy wrong turn, said plainly so it stays closed:

```text
NOT A PROCESS TABLE   the entity table is not a table of Skyra "processes." entities are not our
                      processes.
IT IS CPU ROUTING     the CTU (a CPU) routes an instruction (a MOVE) to a HARDWARE LAYER. drivers
                      are trusted hardware; entities are untrusted hardware.
PROCESS IS DEFERRED   Skyra will define "process" as its own thing, later. it is not this, and it
                      is kept out until then so it cannot smuggle OS semantics in early.
```

---

## Two tables realize TRUST

The CTU owns two lookup tables, and which one answers *is* the trust decision:

```text
HOST TABLE     trusted hardware — drivers that ship with the underlying OS.   output BELIEVED.
ENTITY TABLE   untrusted hardware — foreign code (e.g. a python interpreter).  output STAMPED +
                                                                               CHECKED.
```

The line is **provenance**:

```text
SHIPS WITH THE OS  → HOST TABLE   (trusted).
DOES NOT           → ENTITY TABLE (untrusted).
```

Trust is never asserted per call; it is *where the thing came from*, fixed at registration into
which table accepts it. This is the quadrant's vertical (trust) line, made structural — one split,
two tables.

---

## Dispatch is two steps — eligibility, then selection

A MOVE is not routed to one target by lookup. Several resources may be able to run it (several
hosts, several drivers), exactly as several ALUs can add. So dispatch splits, the out-of-order way:

```text
DECODE     reads the tables → the ELIGIBLE SET. "which registered resources CAN run this?"
SCHEDULE   picks one, and picks WHEN. "which free eligible resource, and is it ready?"
```

Decode filters by capability AND by the quadrant — a privileged MOVE is eligible only on the kernel
(named), a compute MOVE on the drivers/workers that can do it — so the eligible set never spans the
forbidden diagonal. The scheduler then chooses *within* a cell. The full mechanism (the operand
wait, the resource gate where backpressure lives, and placement) is in
`entity-ctu-boundary/decode-and-run.md`. **Whether the scheduler is local or distributed is open,
and is the largest question there.**

---

## Privilege is NOT a flag — the kernel is named

The correction that matters most: **there is no privilege bit on a MOVE.**

```text
PRIVILEGE = THE MOVE NAMES THE KERNEL. named outright, like any other target. a MOVE that names a
            compute driver is unprivileged; a MOVE that names the kernel is privileged — because of
            WHO it names, not a flag it carries.
```

And there is exactly one kernel. **Only the kernel executes privileged operations** — the durable
writes, ACTIVATE / REMEMBER. Everyone who needs privilege NAMES the kernel. There is no second
privileged executor and no way to be privileged except to *be*, or to *name*, the kernel.

So the account is **not "a privileged entity."** It is an untrusted entity that HOLDS A TOKEN and
NAMES THE KERNEL. The privilege is the kernel's; the token is what lets the account name it
successfully. **The account never crosses; the kernel does.**

---

## The example — the four combinations

```text
HOST, unprivileged        a compute driver.              → plain trusted call. no trap, no check.
THE KERNEL (named)        the one privileged executor.   → TRAP. the two crossings (ACTIVATE / REMEMBER).
ENTITY, unprivileged      foreign code.                  → validated dispatch. STAMP + contract-CHECK.
ENTITY names the kernel   the ACCOUNT, holding a token.  → the kernel crosses; the account never does.
```

Routing, decoded:

```text
host driver named   → call it, trust the result.
kernel named        → cross (trap).
entity named        → call it, then STAMP + CHECK the result.
account names kernel → route to the kernel with the token; the KERNEL crosses.
```

Only *naming the kernel* crosses. Everything else is a call; entity calls are the ones that get
checked.

---

## Why "named, not flagged" is the right shape

Derive-don't-declare, one more time. A flag is a claim a MOVE makes about itself, and this design
refuses self-claims everywhere. **A name is not a claim — it is a target**, resolved against a
table (or, for the kernel, a fixed well-known entry). Privilege falls out of WHO you route to,
exactly as trust falls out of WHICH TABLE answers. Neither is ever asserted; both are consequences
of the address.

---

## Open

```text
KERNEL ENTRY    whether the kernel is a fixed well-known target (named directly, like a syscall
                entry point or a reset vector) or a distinguished entry inside the host table. It
                "does not map cleanly to a driver" — recorded, not resolved.

PROCESS         Skyra's own definition of process is deferred and is NOT the entity table. Named
                here so the gap is a decision, not an omission.

TOKEN ROUTING   the mechanics of an account naming the kernel with its token live in
                `kernel-ctu-boundary/horizon-and-root.md`; the wire form is open.
```

# acquire — THINK's capability-acquisition node (discover · test · admit)

**Date:** 2026-08-24
**Status: SHAPE, converging — BOUNDED SO FAR.** `acquire` is the grouping node under THINK that holds
its first settled set of children — the **acquisition pipeline** that turns a capability we *could* use
into one we *have* (find → characterize → register into userspace). This is NOT all of THINK: a
`compose` sibling of `acquire` (turn registered capabilities into the executor tree) is expected later,
so `THINK = { acquire, compose, … }`. Ties together `discover.md`, `../potentials/test.md`,
`../potentials/user-as-privilege-boundary.md`, and `../research/decisions.md` D1–D2.

---

## The structure

```text
THINK  (read-only horizon of ADMITTED reads; holds NO ask_permission by default)
└── acquire                              the capability-acquisition node (this doc)
    ├── discover → RegistrationProposal  FIND         search + resolve
    ├── test     → ConnectionContract    CHARACTERIZE reach + handshake + challenge + describe
    └── admit    → registration (MINT)   REGISTER     select fitness-slice + package + request the write
```

`acquire` groups three siblings into one pipeline: **find → characterize → register-for-reuse.**

---

## The pipeline

```text
STAGE         CHILD      IN                         OUT                       KIND
FIND          discover   obs (intent)               RegistrationProposal      read
CHARACTERIZE  test       one candidate endpoint     ConnectionContract        read (safe-probe)
REGISTER      admit      endpoint + contract        registration into CTU     write (MINT, delegated)
```

The output of each stage feeds the next: candidates → the endpoint to probe → the contract to register.

---

## The children

```text
discover   FIND. search (intent → names) + resolve (name → endpoints). read-only, via ADMITTED
           resolvers. returns a bounded RegistrationProposal of candidate (capability + endpoint)
           records tagged with novelty verdicts.  → discover.md

test       CHARACTERIZE. four safe-probe leaves — reach · handshake · challenge · describe — that
           provoke the endpoint to STATE its contract, joined into a ConnectionContract (transport,
           identity, auth, shape). a READ; privilege is conditional (direct if admitted, faults to the
           privilege boundary for NOVEL first-contact).  → ../potentials/test.md

admit      REGISTER. selects the fitness-matched slice of test's contract, packages (endpoint +
           ConnectionContract), and REQUESTS the write into the userspace (entity) block so the endpoint
           becomes KNOWN for reuse. a WRITE — the actual MINT is performed at the privilege boundary,
           not by admit.  (detail doc TBD)
```

---

## Two reads and one write — and how the write happens

```text
discover, test   READS   sit inside THINK's read-only horizon.
admit            WRITE   the MINT — mutation of the userspace registry.
```

None of the three holds `ask_permission` (it is **scarce**, D2). So privileged acts do not happen
inside THINK — they **fault to the system-administered privilege boundary**, which checks the
registration, applies the user's policy/tolerance, and grants (known → auto, novel → wake human):

```text
test  faults   for NOVEL first-contact (permission to poke).
admit faults   for the REGISTER / MINT (permission to write the capability into userspace).
```

`discover`/`test`/`admit` are **orchestrators** — they say *what* is needed; the privilege entity
**authorizes and writes**. This is what lets a write-shaped step (`admit`) live under a read-only THINK.

---

## Dataflow and the two loops

```text
discover ──▶ test ──▶ admit          the forward pipeline (RegistrationProposal → ConnectionContract → MINT)
   ▲            │
   └── refine ◀─┘                     the search↔test feedback loop: a failed/unfit probe refines search
```

The **fitness match** (does an operation serve the purpose?) lives *above* `test`'s leaves; it both
feeds the `search ↔ test` loop (no fit → refine) and selects the slice `admit` registers.

---

## The two grants (from the poke-vs-register split)

Because `admit` is separate from `test`, there are two distinct privilege moments:

```text
POKE-GRANT      before test    "may I make first contact to learn its contract?"    narrow
REGISTER-GRANT  at admit        "may I register this endpoint + contract for reuse?"  the MINT
```

So the user sees the **full ConnectionContract before committing to register** — poke-first, then
decide. (Whether the two collapse in the known/fast-path case is scope-of-grant, still open.)

---

## Bounded — what is NOT here yet

```text
compose        turn registered capabilities into the executor tree / the actual request graph. THINK's
               likely FOURTH child; the composition half, distinct from this acquisition half.
system-admin   ask_permission's grant may be a system possibility, not a model call. (open, parked.)
addressing     how the privilege boundary is reached from any depth. (open.)
scope-of-grant poke-grant vs register-grant collapse; re-ask on new purpose. (open.)
```

These three siblings are the settled floor, not the ceiling.

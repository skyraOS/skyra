# interface-rationale — v11

**Date:** 2026-08-24
Every type, the term it comes from, and the seams already known. Fresh build — no prior-gen import.

## The types and their terms

```text
TYPE                   TERM / SOURCE                                            ROLE
Outcome (+FAULT)       move.md's three + FAULT = unprovisioned reach            total step result
Body                   cpu-execution-model (compiled artifact) + ENGRAM         the compiled move graph (static)
Context                cpu-execution-model (process) + Possibility              the live run (mutable, per-run)
MoveTemplate / Move    live-possibility, think-act-signature                    static node / live node
Transfer               move-and-the-kernel-barrier                             the wire form across the barrier
Capability             authorization-and-registration (macaroon)               a grant; user-held; offline-verifiable
Horizon                capability-table-and-horizon (D4)                        the provisioned view = []Capability
Entity / Host / Address bootstrap A (entity→host→address)                       WHO / WHERE / volatile locator
Block (Host/Entity)    dispatch.md — believed / checked                         trust axis; privilege is the trap (D5)
```

## The boundaries, and how they cross

```text
COLD   request → kernel.Handle → think.Compile → Body → Table.Provision + Mint → Materialize → Drive →
       Remember → returns a held Capability. the user now holds it.
HOT    request + held Capability → ctu.Move → Verify (offline) → Materialize(cached Body, horizon) →
       Drive. NO kernel.
FAULT  a move to an unprovisioned target → FAULT → kernel provisions (page-table map) or denies → re-drive.
BARRIER egress parks a cut-ID (fd) in Context.InFlight; the kernel socket table holds routing/custody;
       ingress Response pops by fd, Request establishes a return route (ReturnTo, no fd).
LOADER Materialize: Body + Horizon → Context. kernel on cold, CTU on hot.
```

## Seams already known (detail in claude-ideas.md)

```text
S1  Body/Context — separate types vs template+state-array.                        (idea 1)
S2  Capability — whole Horizon per Transfer vs one per target; what a Caveat is.   (idea 2)  ← check
S3  hot path — where the cached Body lives (CTU cache vs content-addressed fetch). (idea 3)  ← most likely to bite
S4  Materialize — same op cold and hot, or cold also provisions/mints.            (idea 4)
S5  Move Header — does "invoke a remembered engram" fit Request, or a 3rd shape.   (idea 5)
S6  fd table (Context) vs socket table (kernel) — split holds on Response.         (idea 6)
S7  host(entity) — where it lives (Host / kernel / driver). bootstrap A leaning.   (idea 8)
S8  THINK — a compiler call vs a possibility-tree (acquire/compose as moves).      (idea 9)
S9  REMEMBER — content-address the Body, the request→Body map, or both; matching.  (idea 10)
```

## What is deliberately NOT modelled

```text
· the distributed kernel (kernel-to-kernel) — one kernel per host (bootstrap B). single-host only.
· authentication internals — the principal is self-auth; foreign is policy-gated; both STUBBED.
· the acquire pipeline internals — reuse the v10 result; here it is a named THINK child + the handshake.
· real drivers / real network — mock the OS backend (the driver seam). structure + seam, not sockets.
```

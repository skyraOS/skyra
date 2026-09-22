# interface-rationale — v10

**Date:** 2026-08-24
Every type, the term it comes from, and the seams already known. Fresh build — no v9.

## The types and their terms

```text
TYPE                    TERM / SOURCE                                              ROLE
Outcome (+FAULT)        move.md's three + a fourth for name-miss on unadmitted     total step result
Possibility             live-possibility.md — reorder buffer                       the running context
Move / Register / Reg…  live-possibility.md, possibility-graph.md                  graph node + register file
Horizon(.Grants)        horizon docs + decisions D2 (scarcity)                     issued, endpoint/purpose/action scopes
Endpoint / Triple       pipeline.md (socket triple + address)                      resolved locator (a verdict)
Capability              acquire.md — what admit registers                          endpoint + contract, Entity block
RegistrationRequest     acquire.md — admit packages before requesting the write    non-authoritative mint input
Candidate / Proposal    discover.md, pipeline.md (addrinfo-shaped)                 discover's bounded output
NoveltyVerdict          risk model — "remembered, not tested"                      known? + history
ConnectionContract      test.md — what a successful conversation requires          test's output, possibly partial
Grant / GrantKind       user-as-privilege-boundary.md, D2                          poke vs register authorization
Block (Host/Entity)     dispatch.md — believed / checked                           the trust axis; privilege is a trap
```

## The boundaries, and how they cross

```text
INTENT → discover         read-only. search (intent→names) + resolve (name→endpoints, per-namespace).
discover → test           a Candidate. test recovers the contract by four safe probes.
test → FAULT → privilege   novel first-contact has no admitted resource → Interpret FAULTs → the
                          privilege PRINCIPAL grants (Poke) or the run stalls with a diagnosis.
admit → privilege → kernel registration is a mint → admit cannot do it → FAULT → privilege DECIDES
                          (novelty + user policy) → TRAPS to kernel → kernel.Mint. two authorities, one
                          crossing: FAULT and TRAP are the same "name resolves to no peer" move.
privilege → engine         a waiting grant PARKS (Cut/InFlight); Recognize lands it and re-drives.
Registry → executor        an admitted Capability is an Entity → runLocal CHECKS its replies vs the contract.
```

The build's final shape keeps `RegistrationRequest` distinct from `Capability`: acquire may package
the former, but only the kernel's private mint can produce and register the latter.

## Seams already known (detail in claude-ideas.md)

```text
S1  acquire as interfaces vs as a Possibility graph.        (idea 1)
S2  FAULT as a bare Outcome vs a typed value with target.    (idea 2)
S3  AskPermission bool vs a resource-scoped capability.      (idea 3)  ← most likely to break
S4  parked grant vs parked cut — one InFlight index or two.  (idea 5)
S5  grant timeout — a human who never answers = Failure?     (idea 5)
S6  poke/register grants collapsing on the fast path.        (idea 8)
S7  the search↔test refine loop — needed in v10, or later.   (idea 10)
S8  principal (privilege decides) vs mechanism (kernel mints) — SETTLED (D3): stay SPLIT, always. a
    collapsed boundary reads as NO boundary; it is real, keep it visible. NOT a seam to explore. (idea 4)
```

## What is deliberately NOT modelled

```text
· real network — search/resolve/probe are mocked. this run is structure + seam, not sockets.
· credential acquisition + live-verify — the executor's, at dispatch. test only recovers the contract.
· compose — THINK's other (future) child. v10 is the acquisition half only.
· the ENGRAM BODY / how a possibility materializes — re-use the 2026-08-23 answer if needed; not the focus.
```

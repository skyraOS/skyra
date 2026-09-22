# The primitives

**Date:** 2026-08-20
**Status: WORKING.** The one-line index. Primitive · owner · what it does. The reasoning, the
model talk, and the arguments are in `primitive-details.md`. A `?` marks genuine ambiguity, not
brevity.

```text
OWNERS
  CTU       an endpoint. executes; unprivileged.
  FABRIC    transport between endpoints.
  KERNEL    owns durable storage and name-meaning; privileged.
  DRIVER    OS-facing routing, ingress, and persistence mechanisms.
  ROOT      a position (a possibility with no parent), not a party.
```

---

## PROTOCOL

```text
MOVE        FABRIC (issued by CTU)   transfers custody of an opaque payload to a target.
                                     request or response. never reads the payload.
GREET    CTU (request port)       admits a normalized REQUEST; checks authority.
RECOGNIZE   CTU (parent's frame)     admits a normalized RESPONSE; checks the return against the
                                     parent's frame-resident contract.
```

## DISCOVERY

```text
THINK       CTU proposes / KERNEL    proposes child signatures at the frontier (what is
            resolves                 AVAILABLE). the act is unprivileged; resolving the names it
                                     proposes is the kernel's.
BOUND       ?                        creates / narrows a horizon (what is PERMITTED).
                                     owner unresolved — CTU-local if horizons are tokens, KERNEL
                                     if a held map. see boundaries SEAM 2.
```

## LIFECYCLE

```text
ACTIVATE    KERNEL                   loads an engram into a possibility. durable → live. the LOAD
                                     crossing.
REMEMBER    KERNEL (at ROOT)         commits the cone of a contribution; tears down the arena.
                                     live → durable. the STORE crossing.
```

## UNPLACED

```text
ORIENT      CTU (endpoint) ?         reads your own outstanding set / topology. purely local, and
                                     possibly not a primitive at all — mechanism unknown.
```

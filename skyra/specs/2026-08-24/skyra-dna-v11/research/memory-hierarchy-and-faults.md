# The memory hierarchy — cache, disk, compile

**Date:** 2026-08-26
**Status: CONVERGING.** How a possibility is resolved when work arrives — the demand-paging hierarchy
and the two-fault cascade. Builds on `think-act-paths.md` (cold/hot), `move-and-the-kernel-barrier.md`
(the seed docks), `capability-table-and-horizon.md` (page faults), and the Body/Context split.

---

## The three tiers

A crossing carries a **seed** (`BodyID` + horizon + inputs). The far CTU resolves it against a
three-tier store:

```text
TIER            HOLDS                          MISS →
──────────────  ─────────────────────────────  ────────────────────────────────────────────
CACHE (RAM)     resident live possibilities    PAGE FAULT → go to disk.       HOT  (already here)
ENGRAM DISK     persisted compiled bodies       FAULT AGAIN → go to the kernel. WARM (page-in)
KERNEL / THINK  (nothing to load)               COMPILE fresh + provision.     COLD (create anew)
```

The three request states we called **cold / warm / hot ARE the three tiers**, and the two faults are
the boundaries between them.

---

## The two-fault cascade

```text
resolve(BodyID):
  cache HIT   → materialize from RAM.                                     HOT
  cache MISS  → PAGE FAULT → the kernel pulls the ENGRAM from disk
     disk HIT  → page-in: load the engram, materialize a Context.         WARM
     disk MISS → the kernel COMPILES a fresh body (THINK) + provisions
                 → hands back a fresh Context.                            COLD
```

So the **cold path is the deepest page fault** — nothing cached, nothing on disk, so the kernel creates
it anew. THINK-compile isn't a separate mechanism; it's the fault that reaches all the way down.

---

## The CTU sees only TWO faults — MATERIALIZE and REMEMBER

The CTU never issues ACTIVATE / MINT / PROVISION. It calls **MATERIALIZE** and receives a Context it can
operate on — oblivious to whether the body was cached, paged in from disk (ACTIVATE), or compiled fresh
(MINT), or whether its horizon needed PROVISION. All of that is **internal** to the kernel's Materialize.

```text
MATERIALIZE   the READ fault — bring a runnable Context INTO the cache.
              hides: ACTIVATE (page-in) · MINT (compile fresh via THINK) · PROVISION (map the horizon) · consent.
REMEMBER      the WRITE fault — persist an engram OUT to disk (page-out).
```

**Symmetric: materialize = in, remember = out.** These are the ONLY two verbs a CTU names on the kernel —
a fault is `move(target=kernel, action=materialize|remember)`. ACTIVATE / MINT / PROVISION are
kernel-internal steps of Materialize, chosen by the tier the resolve reaches (cache / disk / compile).
This keeps the CTU nearly-stateless and tier-oblivious. (v11 kernel: `Faults{Materialize, Remember}`.)

---

## What this unifies

```text
REMEMBER          = writing the engram to disk (+ keeping it in cache). it pushes a compiled body DOWN
                    into the persistent tier — what makes a future run WARM/HOT instead of COLD.
PER-HOST          cache = this host's RAM; disk = this host's engram store (one kernel per host).
                  so warm/cold is host-local — "never been here before" = a cache (or disk) miss here.
BodyID = the KEY  a content address; if the disk is content-addressed, BodyID is its natural key (a CAS).
```

---

## The deepest fault has two flavors

```text
MIGRATED possibility   arrives with a BodyID (a specific hash). a cache+disk miss means "FETCH it by
                       content address" from the sender / a peer. this is the DISTRIBUTED tier —
                       deferred (bootstrap B). for single-host, treat as unreachable.
NEW request            no body yet (unstructured intent). the kernel COMPILES fresh (THINK). this is
                       the "fresh Context" case — the cold path.
```

Single-host now: `cache → disk → compile`. The peer-fetch tier lands with the distributed kernel later.

---

## Open — the engram disk is UN-COMBED (next area)

```text
· the ENGRAM format on disk — what a persisted compiled body actually is.
· how REMEMBER writes it (and when; the persistence trigger).
· eviction / GC — cache tier and disk tier both need it.
· content-addressed? (BodyID = disk key → a CAS; dedup + verify by hash for free.)
· the boundary between a live Context (RAM) and a persisted engram (disk) — Materialize bridges them,
  but what exactly persists (the Body only, or a checkpointed Context?).
```

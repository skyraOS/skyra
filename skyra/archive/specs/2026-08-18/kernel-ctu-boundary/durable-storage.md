# Durable storage

**Date:** 2026-08-20
**Status: WORKING.** How the kernel reaches durable storage, kept deliberately small. The
mechanism below the interface — content-addressing, journaling, block transport — is out of
scope on purpose; see `implementation.md` for one way to fill it, but the point of this document
is that the kernel does not have to know.

> **The shared abstraction is the two crossings. ACTIVATE and REMEMBER are the interface, and
> everything below them is a pluggable OS driver.**

---

## The layering

```text
CTU        issues ACTIVATE / REMEMBER.
  │   ← THE SHARED ABSTRACTION = the two crossings, unchanged. load / store.
OS DRIVER  the filesystem subsystem lives HERE. content-addressing, retention, journaling,
  │        block transport — all internal, all hidden.
BACKEND    whatever actually stores. swappable — a different OS hides beneath the same interface.
```

This is the driver-is-opaque pattern (`blk_mq_ops`) lifted one level: the whole OS, filesystem
included, is the opaque thing. VFS-shaped, drawn above the filesystem instead of below it. The
kernel stays thin.

---

## What the interface is

Two operations, and they are the two crossings this folder already established — not a new
boundary:

```text
ACTIVATE   name → possibility.   LOAD.   "give me this engram as a possibility."
REMEMBER   cone → durable.       STORE.  "here is a cone; make it durable and tell me when."
```

Between them the kernel is not consulted, and beneath them the kernel sees no hash, no sector,
no journal.

---

## The one line the kernel cannot delegate

Retention has two halves. Only one goes down the interface.

```text
WHAT persists   the cone — reachability from a root.   KERNEL. it owns roots and lifetime, so it
                                                        computes the cone and hands it down.
HOW it persists atomically · addressed · durable.       DRIVER. content-addressing, journal, flush.
```

**Kernel owns lifetime; driver owns persistence.** This is the only durable-storage decision that
cannot hide beneath the abstraction, because "what is still reachable" is a fact about roots, and
roots are the kernel's.

---

## Why keep it this small

The three seams from the block-layer read — address-vs-content, the retention *mechanism*, and the
durability/journal race — all move INSIDE the OS driver. They stop being kernel problems, the
kernel/CTU boundary stays at exactly two crossings, and none of content-addressing, journaling, or
log-structure has to be decided now. The fewer promises the shared abstraction makes, the more
backends can honor it.

---

## Open

```text
COMPLETION       REMEMBER is "make it durable and tell me when." Whether the CTU blocks on that
                 completion or is signalled asynchronously is the same submit→complete question
                 durable I/O always has, and it is not answered here.

CONE HANDOFF     the kernel computes the cone and hands it down — but in what form the driver
                 receives it (the whole object set, or a root to walk) is a shape decision left
                 to whoever writes the interface.
```

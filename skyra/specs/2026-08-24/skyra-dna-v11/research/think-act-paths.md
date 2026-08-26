# THINK / ACT — the cold and hot paths

**Date:** 2026-08-24
**Status: DIRECTION — developing.** The system-level THINK/ACT split — the biggest open question of
execution. To be made concrete next (trace a request cold, then hot). Builds on
`think-act-signature.md`, `capability-table-and-horizon.md`, `authorization-and-registration.md`.

---

## The split

```text
NEW request        → the KERNEL (THINK). compile the request into a possibility, provision, MINT a
                     capability. the COLD path.   move(entity=kernel, host=thismachine, "some request").
REMEMBERED engram  → the CTU directly (ACT). the body is already compiled; execute it. the HOT path.
                     the user holds the capability and hits the CTU — no kernel round-trip.
REMEMBER           the BRIDGE: caches THINK's output (compiled engram + provisioning) so the next
                     identical request skips THINK and goes straight to ACT.
```

So **THINK = the kernel path (compile + provision + mint); ACT = the CTU path (execute).** The move
target flips and a held capability appears exactly when a request graduates from new to remembered:

```text
move(entity=kernel, ...)        NEW        user holds NO capability → goes to the kernel.
move(entity=engram, +cap)       REMEMBERED user HOLDS the capability → goes direct to the CTU.
```

---

## Dynamic-linker shape

```text
NEW request        an UNRESOLVED symbol → through the resolver (the KERNEL). slow.
REMEMBERED         a RESOLVED symbol → a DIRECT call (the CTU). fast.
REMEMBER           patching the GOT — what resolved stays resolved.
```

THINK compiles → ACT executes → REMEMBER caches → next time ACT directly.

---

## The capability makes the hot path kernel-free

No-ambient-authority: the CTU can't hold its own horizon between calls. So **the user holds the
capability, not the CTU.**

```text
NEW        user holds no cap → the KERNEL THINKs (compiles) and MINTS a capability (a macaroon = the
           horizon) → executes → REMEMBERs. now the USER holds the capability.
REMEMBERED user HOLDS the cap → presents it WITH the move → the CTU verifies it OFFLINE (macaroon
           signature, no kernel) → ACT.
```

This is **OAuth + macaroons at the request level**: first call → the auth server (kernel) issues a
token; every later call → present the token straight to the resource (CTU), verified offline. The
kernel is only ever in the loop for the **cold path** (minting a new capability). It resolves the
tension cleanly — the CTU stays stateless; the *user* carries the horizon for their remembered engrams.

---

## The loop

```text
THINK (compile) → ACT (execute) → REMEMBER (mint + cache) → next request: ACT directly.
```

---

## Open / next — make concrete

```text
· TRACE cold then hot   one request through the kernel (minting the cap), then the same request direct
                        to the CTU (with the held cap).
· WHAT "COMPILE" IS      request → engram body. connects to think-act-signature (THINK's runnable half).
· WHAT THE CAP CARRIES   the macaroon = the provisioned horizon entries + caveats.
· MATCHING               how "the same request" maps to a remembered engram (content-address? the intent?).
· REVOCATION             a remembered engram whose horizon was revoked (view-coherence) — the held cap
                        must fail closed.
```

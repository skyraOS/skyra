# Bootstrap — single-host sequence

**Date:** 2026-08-24
**Status: SCOPING — decisions forming.** The assembled single-host bootstrap, and the A/B/C calls.
Ties together `kernel.md` and `ctu.md`. B is decided; C is decided as a direction; A is leaning.

---

## The sequence (single host)

```text
1. HUMAN seeds the kernel out-of-band: root user-entity (userspace) + root policy.
2. KERNEL loads and comes up — like an OS kernel from a bootloader. its legitimacy comes from the seed,
   not from anything running (this breaks the loop: the kernel is LOADED, not provisioned).
3. CTU comes up with drivers → abstracts the machine. NO authority; inherent = the handshake channel only.
4. KERNEL ACQUIRES the CTU — the handshake IS acquire: discover (CTU advertises) → test (mutual auth +
   read drivers) → admit (drivers → sysblock).
5. KERNEL provisions horizons.
6. the root user-entity issues the root THINK → a possibility → the CTU executes.
```

---

## main — the bootloader — DECIDED (exit)

`main` is a thin, **out-of-band launcher** (≈ firmware / GRUB). It runs steps 1–4, then closes out:

```text
1. SEED + start the KERNEL   reads the human's config (root identity + policy), hands it to the kernel.
2. start the CTU             with its drivers.
3. let them HANDSHAKE         the kernel acquires the CTU.
→ EXIT                        hand off and die; kernel + CTU run independently (like GRUB → kernel).
```

**MECHANICAL, NOT AN AUTHORITY:** `main` spawns the kernel but does not *grant* it authority — the
kernel's authority comes from the SEED (the human's config), not from `main`. Chain: human → seed →
kernel. `main` is outside the system entirely — not an entity, not a CTU, not the kernel.

**DECIDED — EXIT:** `main` fully exits after handoff (simplest; keeps it mechanical). Supervision /
crash-restart is the host OS's job, or a later deliberate step (promote `main` to a PID-1-style
supervisor) — not now.

---

## A · Addressing / seeding — LEANING

```text
entity → host → address       ENTITY (userspace) = WHO, holds a pointer to its home HOST. HOST (sysblock)
                              resolves the entity to an ADDRESS locally. address is volatile, host-local,
                              NEVER stored on the entity.
host(entity)                  the resolution primitive — the home-registry / HLR pattern (like user@host,
                              or an HLR resolving a subscriber). the local sibling of acquire's `resolve`.
the human is an ENTITY        a userspace entity with a home host — structurally UNIFORM, special ONLY in
                              authority (it's the root). the SEED = root user-entity + root policy.
```

## B · One kernel per host — DECIDED

Simplest first: **one kernel per host.** Bootstrap is purely **local** — a host is its own kernel + its
own CTU, self-contained, no distributed coordination to design yet. The **distributed kernel**
(kernel-to-kernel across hosts) is **DEFERRED** — "the kernel is the distributed part of the system,"
but later.

## C · The handshake = acquire — DECIDED (direction)

The handshake is the **kernel running acquire on the CTU:**

```text
DISCOVER   the CTU connects and advertises itself.
TEST       mutual auth + READ the CTU's drivers (enumeration). the CTU's "ConnectionContract" =
           { its driver set, how to reach it }.
ADMIT      register the CTU's drivers into the sysblock.
```

So **acquire is the universal onboarding primitive**: THINK acquires *endpoints*, the kernel acquires
*CTUs* — one pipeline, two uses. The kernel runs it with its **own seed authority** (already up), not a
provisioned horizon — a privileged bootstrap op, so no loop.

---

## Open

```text
· A's exact shape — is an entity single-homed? how host(entity) resolves; where the pointer lives.
· C's flush-out — the acquire-for-CTU `test` step (mutual-auth mechanism); stubbed like other auth for now.
· the minimal HANDSHAKE CHANNEL — co-located it is local; remote is deferred with the distributed kernel.
```

# The capability table and the horizon — the kernel as MMU

**Date:** 2026-08-24
**Status: CONVERGING.** Who owns the capability table, and what the horizon actually is. Builds on
`../ownership.md`, `move-and-the-kernel-barrier.md`, and D4 (`Horizon = []Grant`). Two prior open pins
are now DECIDED (view, and provision-vs-Bound).

---

## Why the kernel owns identity AND the full table

Authorization is **identity × capability** — "is *this principal* allowed *this capability*." You can't
answer that from a slice of the table, and **a self-owned table is self-authorization** (the CTU could
extend its own reach). So the moment the kernel owns identity, it must own the **full table** — both
the entity/userspace and host/systemspace capabilities.

```text
KERNEL owns   IDENTITY (who is the principal) + the FULL TABLE (all entity + host capabilities).
```

**The motivation is the inbound case.** An unsolicited request has no fd, so it is authorized not by
solicitation but by **identity** (see `move-and-the-kernel-barrier.md` §7):

```text
RESPONSE (has fd)          authorized by SOLICITATION.
REQUEST from the PRINCIPAL authorized by IDENTITY — self-authorizing; it SEEDS the root horizon (it is
  (the user)               the authority SOURCE, not a consumer). the check for it is AUTHENTICATION.
REQUEST from a FOREIGN     authorized by POLICY (the stubbed permission check at the barrier).
  entity
```

To decide any of that, the kernel must own identity — and therefore the full table.

---

## The horizon IS the provisioned subset

The kernel provisions a subset of its full table to a possibility "to move over." **That subset is the
horizon.** D4 said `Horizon = []Grant`; now we know what a grant *is* — **one provisioned entry from
the kernel's full table.**

```text
HORIZON = the subset of the kernel's full capability table PROVISIONED to this possibility.
          []Grant = the provisioned entries. the CTU can only "move over" what is in its horizon.
```

---

## The unification: kernel = MMU, horizon = address space

```text
OS / CPU                              SKYRA
────────────────────────────────     ─────────────────────────────────────────────
kernel owns all physical memory       kernel owns the full capability table
the process's mapped address space    the HORIZON (provisioned subset)
a page-table entry (a mapping)        a GRANT
access to an UNMAPPED page → PAGE      a move to an UNPROVISIONED capability → FAULT → kernel provisions
  FAULT → kernel maps or segfaults       (ask_permission / mint) or DENIES
the TLB (fast local view)             the horizon as the CTU reads it to dispatch
the MMU                               the kernel at the barrier
```

So a novel-capability fault **is a page fault**: the CTU reached for something not mapped into its
horizon → trap to the kernel (the MMU) → **provision it** (map it in, under the human's policy =
`ask_permission` / mint) or **deny** (segfault). Dispatch reads the horizon; a miss faults to the full
table.

---

## What this relocates

In v9/v10 the registries lived **on the CTU** (`Machine.hosts`, `Machine.entities`). This moves them:
**the kernel owns the full table; the CTU holds a provisioned VIEW (its horizon).** The CTU's registry
becomes a cache/view — a TLB over the kernel's page table. Dispatch reads the view; a miss faults.

And it completes a pattern seen twice now:

```text
kernel owns the SOCKET TABLE (in-flight)   → possibility holds FDs (cut-IDs).
kernel owns the CAPABILITY TABLE (full)    → possibility holds the HORIZON (provisioned view).
```

Same shape: **the kernel owns the authoritative state; the possibility holds a provisioned handle/view.**
That is why the CTU stays nearly-stateless — everything authoritative is the kernel's.

---

## Decided

```text
VIEW, NOT COPY       the horizon is a LIVE view of the kernel table — a kernel revocation is reflected
                     (TLB-shootdown-like), not a materialize-time snapshot. revocation matters.
PROVISION vs BOUND   only the KERNEL provisions (WIDENS) — a page-fault-and-map under policy. `Bound`
                     only NARROWS (D4). widening is the kernel's; narrowing is the possibility's.
```

## Open

```text
VIEW COHERENCE    the mechanism by which a kernel revocation propagates to a live horizon view.
IDENTITY          what authentication concretely is — stubbed for now (move-and-the-kernel-barrier §7).
```
